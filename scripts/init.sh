#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

generate() {
    openssl rand -hex "$1"
}

echo "==> 生成安全随机配置..."

DB_USER="app_$(generate 6)"
DB_PASSWORD="$(generate 32)"
DB_NAME="app_$(generate 6)"
JWT_SECRET="$(generate 32)"
REDIS_PASSWORD="$(generate 32)"

echo "    DB_USER        = $DB_USER"
echo "    DB_PASSWORD    = $DB_PASSWORD"
echo "    DB_NAME        = $DB_NAME"
echo "    JWT_SECRET     = $JWT_SECRET"
echo "    REDIS_PASSWORD = $REDIS_PASSWORD"

sed \
    -e "s|CHANGE_ME_DB_USER|$DB_USER|g" \
    -e "s|CHANGE_ME_DB_PASSWORD|$DB_PASSWORD|g" \
    -e "s|CHANGE_ME_DB_NAME|$DB_NAME|g" \
    "$SCRIPT_DIR/../configs/config.yaml.example" > "$SCRIPT_DIR/../configs/config.yaml"

sed \
    -e "s|CHANGE_ME_DB_USER|$DB_USER|g" \
    -e "s|CHANGE_ME_DB_PASSWORD|$DB_PASSWORD|g" \
    -e "s|CHANGE_ME_DB_NAME|$DB_NAME|g" \
    "$SCRIPT_DIR/../deployments/docker-compose.yml.example" > "$SCRIPT_DIR/../deployments/docker-compose.yml"

sed -i.bak "s|CHANGE_ME_JWT_SECRET|$JWT_SECRET|g" "$SCRIPT_DIR/../configs/config.yaml"
sed -i.bak2 "s|CHANGE_ME_REDIS_PASSWORD|$REDIS_PASSWORD|g" "$SCRIPT_DIR/../configs/config.yaml"
sed -i.bak "s|CHANGE_ME_REDIS_PASSWORD|$REDIS_PASSWORD|g" "$SCRIPT_DIR/../deployments/docker-compose.yml"
rm -f "$SCRIPT_DIR/../configs/config.yaml.bak" "$SCRIPT_DIR/../configs/config.yaml.bak2" "$SCRIPT_DIR/../deployments/docker-compose.yml.bak"

echo "==> 配置文件已生成：configs/config.yaml, deployments/docker-compose.yml"
