#!/usr/bin/env bash
# ────────────────────────────────────────────────────────────────
# arch-check.sh — 分层架构合规性静态检查（Semgrep + Bash 混合工具链）
#
# 工具链分层：
#   Layer 1: Semgrep — 架构/鉴权/事务/代码质量模式匹配（scripts/semgrep-rules/*.yml）
#   Layer 2: Bash    — Semgrep 不擅长的结构性检查（返回值计数、函数-文件名匹配、
#                      多写操作无事务、文件行数、鉴权覆盖度表）
#   Layer 3: go vet  — 编译器级检查（由 Makefile lint 目标覆盖）
#
# 输出：docs/arch-report.md（AI 易读的结构化 Markdown）
# ────────────────────────────────────────────────────────────────
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
REPORT="$ROOT_DIR/docs/arch-report.md"
RULES_DIR="$ROOT_DIR/scripts/semgrep-rules"
SEMGREP_JSON=$(mktemp)
trap 'rm -f "$SEMGREP_JSON"' EXIT

# ── 依赖检查 ──────────────────────────────────────────────────
for cmd in semgrep jq; do
	if ! command -v "$cmd" >/dev/null 2>&1; then
		echo "错误：$cmd 未安装。请安装后重试。" >&2
		echo "  semgrep: pip install semgrep" >&2
		echo "  jq:      sudo apt install jq" >&2
		exit 1
	fi
done

# ── 颜色 ──────────────────────────────────────────────────────
RED='\033[31m'; GREEN='\033[32m'; YELLOW='\033[33m'; CYAN='\033[36m'; RESET='\033[0m'

# ── 辅助函数 ──────────────────────────────────────────────────
md_escape() {
	sed 's/[\\`*_{}/#<>|\[\]]/\\&/g'
}

# 根据规则 ID 获取类别（A/B/D/E/F）
get_category() {
	case "$1" in
		controller-no-require-role|controller-no-role-level-compare|controller-no-role-constant-compare|controller-no-forbidden|service-no-echo-context|repository-no-auth-usercontext|repository-no-require-role|repository-no-forbidden|middleware-no-require-role|middleware-no-forbidden|router-no-require-role|router-no-auth-role)
			echo "A" ;;
		service-write-method-missing-auth|service-guest-silent-return-nil|controller-guest-silent-json-return)
			echo "B" ;;
		service-tx-not-using-tx-var)
			echo "D" ;;
		service-log-error-no-return|dead-code-var-underscore)
			echo "E" ;;
		*)
			echo "F" ;;
	esac
}

# 严重级别映射
sev_emoji() {
	case "$1" in
		ERROR)   echo "❌" ;;
		WARNING) echo "⚠️" ;;
		INFO)    echo "ℹ️" ;;
		*)       echo "❓" ;;
	esac
}

# ════════════════════════════════════════════════════════════════
# Phase 1: 运行 Semgrep
# ════════════════════════════════════════════════════════════════
printf "${CYAN}  运行 Semgrep 规则...${RESET}\n" >&2
semgrep --config "$RULES_DIR" --json --timeout 60 > "$SEMGREP_JSON" 2>"$ROOT_DIR/.semgrep-stderr.log" || true

if [ -s "$ROOT_DIR/.semgrep-stderr.log" ]; then
	printf "${YELLOW}  ⚠ Semgrep stderr 有输出，详见 .semgrep-stderr.log${RESET}\n" >&2
fi

SEMGREP_TOTAL=$(jq '.results | length' "$SEMGREP_JSON")
SEMGREP_ERROR_COUNT=$(jq '.errors | length' "$SEMGREP_JSON")

if [ "${SEMGREP_ERROR_COUNT:-0}" -gt 0 ] 2>/dev/null; then
	printf "${YELLOW}  ⚠ Semgrep 报告了 ${SEMGREP_ERROR_COUNT} 个规则解析错误（已跳过这些规则）${RESET}\n" >&2
fi

# 解析 Semgrep 结果到分类数组
# 格式: severity|rule_id|file|line|message
declare -a FINDINGS_A=() FINDINGS_B=() FINDINGS_D=() FINDINGS_E=() FINDINGS_F=()

if [ "${SEMGREP_TOTAL:-0}" -gt 0 ] 2>/dev/null; then
	while IFS=$'\t' read -r rid sev path line msg; do
		[ -z "$rid" ] && continue
		short_id="${rid##*.}"
		file_rel="${path#$ROOT_DIR/}"
		cat=$(get_category "$short_id")
		finding="${sev}|${short_id}|${file_rel}|${line}|${msg}"
		case "$cat" in
			A) FINDINGS_A+=("$finding") ;;
			B) FINDINGS_B+=("$finding") ;;
			D) FINDINGS_D+=("$finding") ;;
			E) FINDINGS_E+=("$finding") ;;
			F) FINDINGS_F+=("$finding") ;;
		esac
	done < <(jq -r '.results[] | [.check_id, .extra.severity, .path, (.start.line|tostring), .extra.message] | @tsv' "$SEMGREP_JSON")
fi

count_a=${#FINDINGS_A[@]}
count_b=${#FINDINGS_B[@]}
count_d=${#FINDINGS_D[@]}
count_e=${#FINDINGS_E[@]}
count_f=${#FINDINGS_F[@]}
semgrep_total=$((count_a + count_b + count_d + count_e + count_f))

# ════════════════════════════════════════════════════════════════
# Phase 2: Bash 结构性检查（Semgrep 不擅长的部分）
# ════════════════════════════════════════════════════════════════
printf "${CYAN}  运行 Bash 结构性检查...${RESET}\n" >&2

SERVICE_FILES=$(find "$ROOT_DIR/internal/service" -name '*.go' -not -name '*_test.go' | sort)
CONTROLLER_FILES=$(find "$ROOT_DIR/internal/controller" -name '*.go' -not -name '*_test.go' -not -path '*/param/*' | sort)

# ── C1: 函数返回值过多（> 5）──────────────────────────────────
declare -a MANY_RETURNS=()
RETURN_THRESHOLD=5
for f in $SERVICE_FILES $CONTROLLER_FILES; do
	while IFS= read -r fline; do
		[ -z "$fline" ] && continue
		lineno=$(echo "$fline" | cut -d: -f1)
		code_raw=$(echo "$fline" | cut -d: -f2-)
		echo "$code_raw" | grep -q ') (' || continue
		ret_section=$(echo "$code_raw" | sed 's/.*) (//' | sed 's/).*//')
		comma_count=$(echo "$ret_section" | tr -cd ',' | wc -c)
		ret_count=$((comma_count + 1))
		if [ "$ret_count" -gt "$RETURN_THRESHOLD" ]; then
			code_escaped=$(echo "$code_raw" | sed 's/^[[:space:]]*//' | md_escape)
			file_rel=$(echo "$f" | sed "s|$ROOT_DIR/||")
			MANY_RETURNS+=("${file_rel}§${lineno}§${ret_count}§${code_escaped}")
		fi
	done < <(grep -n '^func ' "$f" 2>/dev/null || true)
done
count_returns=${#MANY_RETURNS[@]}

# ── C2: 函数-文件名不匹配 ─────────────────────────────────────
declare -a STRUCT_ISSUES=()
STRUCT_SKIP='helpers\|transaction\|validate\|_input\|_test\|routes\|router\|bootstrap\|init\|error\|context\|guest\|role'

for layer_dir in internal/controller internal/service; do
	layer_name=$(echo "$layer_dir" | sed 's|internal/||' | sed 's/\b\(.\)/\u\1/')
	# 收集该层所有 struct 定义（跨文件），用于识别同结构体方法分散在多文件的情况
	all_structs=$(grep -rhoP 'type \K[A-Z][a-zA-Z0-9]*(?= struct)' "$ROOT_DIR/$layer_dir" --include='*.go' 2>/dev/null | sort -u || true)
	while IFS= read -r f; do
		fname=$(basename "$f" .go)
		echo "$fname" | grep -qE "($STRUCT_SKIP)" && continue
		keyword=$(echo "$fname" | sed 's/_controller//;s/_service//;s/_repo_gorm//;s/_repo//;s/_like//;s/_moderator//' | cut -d_ -f1)
		[ -z "$keyword" ] && continue

		struct_names=$(grep -oP 'type \K[A-Z][a-zA-Z0-9]*(?= struct)' "$f" 2>/dev/null || true)
		all_keywords="$keyword"
		if [ -n "$struct_names" ]; then
			while IFS= read -r sname; do
				words=$(echo "$sname" | sed 's/[A-Z]/_\l\&/g' | sed 's/^_//' | tr '_' '\n' | grep -v '^$')
				all_keywords="$all_keywords $words"
			done <<< "$struct_names"
		fi

		methods=$(grep -n '^func ([^)]*) [A-Z]' "$f" 2>/dev/null || true)
		[ -z "$methods" ] && continue
		while IFS= read -r mline; do
			method_name=$(echo "$mline" | sed -n 's/.*func [^)]*) \([A-Z][a-zA-Z0-9]*\).*/\1/p')
			[ -z "$method_name" ] && continue
			case "$method_name" in New*) continue ;; esac

			receiver_type=$(echo "$mline" | sed -n 's/.*func [^)]*\* *\([A-Z][a-zA-Z0-9]*\).*/\1/p')
			if [ -n "$receiver_type" ] && echo "$all_structs" | grep -qxF "$receiver_type"; then
				continue
			fi

			matched=false
			for kw in $all_keywords; do
				[ ${#kw} -lt 3 ] && continue
				echo "$method_name" | grep -qi "$kw" && { matched=true; break; }
			done
			$matched && continue

			lineno=$(echo "$mline" | cut -d: -f1)
			is_called=$(grep -cE "\b${method_name}\b" "$f" 2>/dev/null || echo "0")
			if [ "$is_called" -le 1 ]; then
				file_rel=$(echo "$f" | sed "s|$ROOT_DIR/||")
				STRUCT_ISSUES+=("🔄§${layer_name}§${file_rel}§${lineno}§${method_name}§方法名与文件领域关键词 '${keyword}' 无关且未被文件内其他函数调用")
			fi
		done <<< "$methods"
	done < <(find "$ROOT_DIR/$layer_dir" -name '*.go' -not -name '*_test.go' -not -path '*/param/*' | sort)
done
count_struct=${#STRUCT_ISSUES[@]}

# ── D2: 多写操作无事务包裹 ────────────────────────────────────
declare -a MULTI_WRITE=()
for f in $SERVICE_FILES; do
	methods=$(grep -n '^func ([^)]*) [A-Z]' "$f" 2>/dev/null || true)
	[ -z "$methods" ] && continue
	while IFS= read -r mline; do
		method_name=$(echo "$mline" | sed -n 's/.*func [^)]*) \([A-Z][a-zA-Z0-9]*\).*/\1/p')
		[ -z "$method_name" ] && continue
		case "$method_name" in
			New*|List*|Get*|Find*|Count*|Resolve*|Seed*|build*|Default*|Parse*|Send*|Verify*|Run*|extract*|generate*) continue ;;
		esac

		lineno=$(echo "$mline" | cut -d: -f1)
		remaining=$(tail -n +"$lineno" "$f")
		method_body=$(echo "$remaining" | sed '/^func /{1d;q}' | head -150)

		write_count=$(echo "$method_body" | grep -cE '\.(Create|Update|Delete|CreateBatch|DeleteByUserID)\(' 2>/dev/null || true)
		write_count=${write_count:-0}
		has_tx=$(echo "$method_body" | grep -c 'RunInTransaction' 2>/dev/null || true)
		has_tx=${has_tx:-0}

		if [ "$write_count" -ge 2 ] && [ "$has_tx" -eq 0 ]; then
			file_rel=$(echo "$f" | sed "s|$ROOT_DIR/||")
			MULTI_WRITE+=("⚠️§Service§${file_rel}§${lineno}§${method_name}§方法含 ${write_count} 次写操作但无 RunInTransaction 事务包裹，部分失败可能导致数据不一致")
		fi
	done <<< "$methods"
done
count_multi_write=${#MULTI_WRITE[@]}

# ── E1: 文件行数检查 ──────────────────────────────────────────
declare -a LARGE_FILES=()
declare -A LAYER_THRESHOLDS=(
	[controller]=300
	[service]=400
	[repository]=200
	[middleware]=200
	[router]=200
	[model]=200
)
for layer_dir in internal/controller internal/service internal/middleware internal/router internal/model; do
	layer_key=$(echo "$layer_dir" | sed 's|internal/||')
	threshold="${LAYER_THRESHOLDS[$layer_key]:-300}"
	while IFS= read -r f; do
		lines=$(wc -l < "$f")
		if [ "$lines" -gt "$threshold" ]; then
			file_rel=$(echo "$f" | sed "s|$ROOT_DIR/||")
			LARGE_FILES+=("⚠️§${layer_key}§${file_rel}§${lines}§${threshold}")
		fi
	done < <(find "$ROOT_DIR/$layer_dir" -name '*.go' -not -name '*_test.go' -not -path '*/param/*' | sort)
done
count_large=${#LARGE_FILES[@]}

bash_total=$((count_returns + count_struct + count_multi_write + count_large))
grand_total=$((semgrep_total + bash_total))

# ════════════════════════════════════════════════════════════════
# Phase 3: 生成报告
# ════════════════════════════════════════════════════════════════
printf "${CYAN}  生成报告...${RESET}\n" >&2

cat > "$REPORT" << 'HEADER'
# 分层架构合规性检查报告

> 此报告由 `scripts/arch-check.sh`（Semgrep + Bash 混合工具链）自动生成。
> 规则来源：`docs/role-system.md` §6 — Service 层鉴权模式。
>
> **核心原则**：中间件负责"你是谁"，Service 负责"你能看什么/做什么"。
>
> ⚠️ **免责声明**：本工具仅做模式匹配和结构性检查，会存在漏报和误报，不能替代更仔细的代码 review。

HEADER

echo "生成时间：$(date '+%Y-%m-%d %H:%M:%S')" >> "$REPORT"
echo "" >> "$REPORT"

# ── 1. 汇总 ────────────────────────────────────────────────────
echo "## 1. 汇总" >> "$REPORT"
echo "" >> "$REPORT"
echo "| 工具 | 检查类别 | 规则来源 | 发现数 |" >> "$REPORT"
echo "|---|---|---|---|" >> "$REPORT"
echo "| Semgrep | A — 分层架构违规 | \`arch-layers.yml\` | $count_a |" >> "$REPORT"
echo "| Semgrep | B — 鉴权模式检查 | \`auth-patterns.yml\` | $count_b |" >> "$REPORT"
echo "| Semgrep | D1 — 事务内未使用 tx | \`code-quality.yml\` | $count_d |" >> "$REPORT"
echo "| Semgrep | E2/E3 — 代码质量 | \`code-quality.yml\` | $count_e |" >> "$REPORT"
echo "| Semgrep | F — 扩展规则 | 多个 yml | $count_f |" >> "$REPORT"
echo "| Bash | C1 — 函数返回值过多 | 本脚本 | $count_returns |" >> "$REPORT"
echo "| Bash | C2 — 函数-文件名不匹配 | 本脚本 | $count_struct |" >> "$REPORT"
echo "| Bash | D2 — 多写操作无事务 | 本脚本 | $count_multi_write |" >> "$REPORT"
echo "| Bash | E1 — 文件行数过多 | 本脚本 | $count_large |" >> "$REPORT"
echo "" >> "$REPORT"
echo "**总发现数：${grand_total}**（Semgrep ${semgrep_total} + Bash ${bash_total}）" >> "$REPORT"
echo "" >> "$REPORT"

# ── 2. 架构层级图 ──────────────────────────────────────────────
echo "## 2. 架构层级图" >> "$REPORT"
echo "" >> "$REPORT"
echo '```mermaid' >> "$REPORT"
cat >> "$REPORT" << 'MERMAID'
graph TD
    subgraph 请求处理链
        direction TB
        R["🔧 Router<br/><small>internal/router</small>"]
        MW["🔐 Middleware<br/><small>身份解析 + 限流</small>"]
        C["📋 Controller<br/><small>参数绑定 + 调用 Service</small>"]
        S["⚙️ Service<br/><small>鉴权 + 业务逻辑</small>"]
        Repo["💾 Repository<br/><small>数据访问</small>"]
        M["📊 Model<br/><small>数据结构</small>"]
    end

    R --> MW --> C --> S --> Repo --> M
MERMAID
echo '```' >> "$REPORT"
echo "" >> "$REPORT"

# ── 3. Semgrep 违规详情 ────────────────────────────────────────
echo "## 3. Semgrep 违规详情" >> "$REPORT"
echo "" >> "$REPORT"

# 辅助函数：输出一个类别下的所有发现
emit_findings() {
	local title="$1" findings_array_name="$2"
	local -n findings_ref="$findings_array_name"
	local count=${#findings_ref[@]}

	if [ "$count" -eq 0 ]; then
		echo "✅ 未发现${title}问题。" >> "$REPORT"
		return
	fi

	echo "共发现 ${count} 个${title}问题：" >> "$REPORT"
	echo "" >> "$REPORT"
	echo "| 级别 | 规则 | 文件 | 行号 | 说明 |" >> "$REPORT"
	echo "|---|---|---|---|---|" >> "$REPORT"
	for f in "${findings_ref[@]}"; do
		IFS='|' read -r sev rid file line msg <<< "$f"
		emoji=$(sev_emoji "$sev")
		echo "| ${emoji} | \`${rid}\` | \`${file}\`:L${line} | ${line} | ${msg} |" >> "$REPORT"
	done
	echo "" >> "$REPORT"
}

echo "### 3a. A 类 — 分层架构违规" >> "$REPORT"
echo "" >> "$REPORT"
emit_findings "分层架构" FINDINGS_A

echo "### 3b. B 类 — 鉴权模式检查" >> "$REPORT"
echo "" >> "$REPORT"
emit_findings "鉴权模式" FINDINGS_B

echo "### 3c. D1 — 事务内未使用 tx" >> "$REPORT"
echo "" >> "$REPORT"
emit_findings "事务一致性" FINDINGS_D

echo "### 3d. E2/E3 — 代码质量" >> "$REPORT"
echo "" >> "$REPORT"
emit_findings "代码质量" FINDINGS_E

echo "### 3e. F 类 — 扩展规则" >> "$REPORT"
echo "" >> "$REPORT"
emit_findings "扩展规则" FINDINGS_F

# ── 4. Service 层鉴权覆盖分析 ─────────────────────────────────
echo "## 4. Service 层鉴权覆盖分析" >> "$REPORT"
echo "" >> "$REPORT"
echo "以下列出 Service 层每个公开方法是否包含 \`RequireRole\` 或其他鉴权逻辑。" >> "$REPORT"
echo "" >> "$REPORT"
echo "| 文件 | 方法 | 有鉴权 |" >> "$REPORT"
echo "|---|---|---|" >> "$REPORT"

for f in $SERVICE_FILES; do
	fname=$(basename "$f" .go)
	methods=$(grep -n '^func ([^)]*) [A-Z]' "$f" 2>/dev/null || true)
	[ -z "$methods" ] && continue
	while IFS= read -r mline; do
		method_name=$(echo "$mline" | sed -n 's/.*func [^)]*) \([A-Z][a-zA-Z0-9]*\).*/\1/p')
		[ -z "$method_name" ] && continue
		case "$method_name" in New*) continue ;; esac

		lineno=$(echo "$mline" | cut -d: -f1)
		remaining=$(tail -n +"$lineno" "$f")
		method_body=$(echo "$remaining" | sed '/^func /{1d;q}' | head -150)

		has_auth=""
		if echo "$method_body" | grep -qE 'RequireRole|uc\.Role\.Level|uc\.Role ==|apperror\.Forbidden'; then
			has_auth="✅"
		elif echo "$method_body" | grep -qE 'uc\.IsGuest\(\)|uc\.IsAuthenticated\(\)'; then
			has_auth="🔄 条件判断"
		else
			case "$method_name" in
				Login*|Register*|Check*|Send*|Verify*|Parse*|Run*) has_auth="📖 公开/基础设施" ;;
				List*|Get*|Find*|Count*|Resolve*|Default*|Seed*|build*) has_auth="📖 公开读取" ;;
				*) has_auth="⚠️ 无" ;;
			esac
		fi
		echo "| \`${fname}\` | \`${method_name}\` | ${has_auth} |" >> "$REPORT"
	done <<< "$methods"
done
echo "" >> "$REPORT"

# ── 5. Bash 结构性检查 ─────────────────────────────────────────
echo "## 5. Bash 结构性检查" >> "$REPORT"
echo "" >> "$REPORT"

# 5a. C1 返回值过多
echo "### 5a. C1 — 函数返回值过多（> ${RETURN_THRESHOLD}）" >> "$REPORT"
echo "" >> "$REPORT"
if [ "$count_returns" -eq 0 ]; then
	echo "✅ 未发现返回值过多问题。" >> "$REPORT"
else
	echo "| 文件 | 行号 | 返回值数 | 签名 |" >> "$REPORT"
	echo "|---|---|---|---|" >> "$REPORT"
	for mr in "${MANY_RETURNS[@]}"; do
		IFS='§' read -r mr_file mr_lineno mr_count mr_code <<< "$mr"
		echo "| \`${mr_file}\`:L${mr_lineno} | ${mr_lineno} | ${mr_count} | \`${mr_code}\` |" >> "$REPORT"
	done
fi
echo "" >> "$REPORT"

# 5b. C2 函数-文件名不匹配
echo "### 5b. C2 — 函数-文件名不匹配" >> "$REPORT"
echo "" >> "$REPORT"
if [ "$count_struct" -eq 0 ]; then
	echo "✅ 未发现函数-文件名不匹配问题。" >> "$REPORT"
else
	echo "| 级别 | 层 | 文件 | 行号 | 方法名 | 说明 |" >> "$REPORT"
	echo "|---|---|---|---|---|---|" >> "$REPORT"
	for si in "${STRUCT_ISSUES[@]}"; do
		IFS='§' read -r si_level si_layer si_file si_lineno si_code si_desc <<< "$si"
		echo "| ${si_level} | ${si_layer} | \`${si_file}\`:L${si_lineno} | ${si_lineno} | \`${si_code}\` | ${si_desc} |" >> "$REPORT"
	done
fi
echo "" >> "$REPORT"

# 5c. D2 多写操作无事务
echo "### 5c. D2 — 多写操作无事务包裹" >> "$REPORT"
echo "" >> "$REPORT"
if [ "$count_multi_write" -eq 0 ]; then
	echo "✅ 未发现多写操作无事务问题。" >> "$REPORT"
else
	echo "| 级别 | 层 | 文件 | 行号 | 方法 | 说明 |" >> "$REPORT"
	echo "|---|---|---|---|---|---|" >> "$REPORT"
	for mw in "${MULTI_WRITE[@]}"; do
		IFS='§' read -r mw_level mw_layer mw_file mw_lineno mw_code mw_desc <<< "$mw"
		echo "| ${mw_level} | ${mw_layer} | \`${mw_file}\`:L${mw_lineno} | ${mw_lineno} | \`${mw_code}\` | ${mw_desc} |" >> "$REPORT"
	done
fi
echo "" >> "$REPORT"

# 5d. E1 文件行数
echo "### 5d. E1 — 文件行数过多" >> "$REPORT"
echo "" >> "$REPORT"
if [ "$count_large" -eq 0 ]; then
	echo "✅ 未发现文件行数过多问题。" >> "$REPORT"
else
	echo "| 级别 | 层 | 文件 | 行数 | 阈值 |" >> "$REPORT"
	echo "|---|---|---|---|---|" >> "$REPORT"
	for lf in "${LARGE_FILES[@]}"; do
		IFS='§' read -r lf_level lf_layer lf_file lf_lines lf_threshold <<< "$lf"
		echo "| ${lf_level} | ${lf_layer} | \`${lf_file}\` | ${lf_lines} | ${lf_threshold} |" >> "$REPORT"
	done
fi
echo "" >> "$REPORT"

# ── 6. 附录 ────────────────────────────────────────────────────
cat >> "$REPORT" << 'APPENDIX'
## 6. 检查规则说明

### 工具链分层

| 层 | 工具 | 职责 | 规则文件 |
|---|---|---|---|
| Layer 1 | **Semgrep** | AST 级模式匹配 — 架构违规、鉴权模式、事务检查、代码质量 | `scripts/semgrep-rules/*.yml` |
| Layer 2 | **Bash** | 结构性检查 — 返回值计数、文件名匹配、多写事务、文件行数、鉴权覆盖度表 | 本脚本 |
| Layer 3 | **go vet** | 编译器级检查 | Makefile `lint` 目标 |

### 各层职责

| 层 | 允许 | 禁止 |
|---|---|---|
| **Router** | 路由注册 | 引用 auth 包、角色判断 |
| **Middleware** | 身份解析（构造 UserContext）、限流 | RequireRole()、返回 Forbidden |
| **Controller** | GetUserContext(c) 取身份、参数绑定、调用 Service、格式化响应 | RequireRole()、uc.Role 比较、apperror.Forbidden()、业务逻辑 |
| **Service** | 鉴权（RequireRole）、业务逻辑、数据组装 | 引用 echo.Context |
| **Repository** | 数据库 CRUD | 引用 auth.UserContext、RequireRole()、apperror.Forbidden() |

### Semgrep 规则分类

| 类别 | 规则文件 | 检查内容 |
|---|---|---|
| A — 分层架构 | `arch-layers.yml` | Controller/Service/Repository/Middleware/Router 层边界违规 |
| B — 鉴权模式 | `auth-patterns.yml` | 写操作缺鉴权、Guest 静默返回 |
| D1 — 事务 | `code-quality.yml` | RunInTransaction 闭包内用 repo(ctx,...) 而非 tx |
| E2/E3 — 质量 | `code-quality.yml` | log.Error 后未返回、var _ = 死代码 |
| F — 扩展 | 多个 yml | net/http 依赖、fmt.Errorf、GORM 泄漏、context.Background |
APPENDIX

echo "" >> "$REPORT"

# ════════════════════════════════════════════════════════════════
# Phase 4: 控制台输出
# ════════════════════════════════════════════════════════════════
echo ""
printf "${CYAN}═══════════════════════════════════════════════════${RESET}\n"
printf "${CYAN}  分层架构合规性检查（Semgrep + Bash）${RESET}\n"
printf "${CYAN}═══════════════════════════════════════════════════${RESET}\n"
echo ""

printf "  ${CYAN}── Semgrep 发现 ──${RESET}\n"
printf "  %-30s %s\n" "A — 分层架构违规" "$([ $count_a -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${RED}❌ $count_a${RESET}")"
printf "  %-30s %s\n" "B — 鉴权模式" "$([ $count_b -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${RED}❌ $count_b${RESET}")"
printf "  %-30s %s\n" "D1 — 事务内未使用 tx" "$([ $count_d -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${RED}❌ $count_d${RESET}")"
printf "  %-30s %s\n" "E2/E3 — 代码质量" "$([ $count_e -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${YELLOW}⚠️ $count_e${RESET}")"
printf "  %-30s %s\n" "F — 扩展规则" "$([ $count_f -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${YELLOW}⚠️ $count_f${RESET}")"
echo ""

printf "  ${CYAN}── Bash 结构性检查 ──${RESET}\n"
printf "  %-30s %s\n" "C1 — 函数返回值过多" "$([ $count_returns -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${YELLOW}⚠️ $count_returns${RESET}")"
printf "  %-30s %s\n" "C2 — 函数-文件名不匹配" "$([ $count_struct -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${YELLOW}⚠️ $count_struct${RESET}")"
printf "  %-30s %s\n" "D2 — 多写操作无事务" "$([ $count_multi_write -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${YELLOW}⚠️ $count_multi_write${RESET}")"
printf "  %-30s %s\n" "E1 — 文件行数过多" "$([ $count_large -eq 0 ] && printf "${GREEN}✅ 0${RESET}" || printf "${YELLOW}⚠️ $count_large${RESET}")"
echo ""

if [ "$grand_total" -eq 0 ]; then
	printf "${GREEN}  ✅ 全部通过！所有检查项均合规。${RESET}\n"
else
	printf "${YELLOW}  ⚠ 共发现 ${grand_total} 个问题（详见 ${REPORT#$ROOT_DIR/}）${RESET}\n"
fi
echo ""