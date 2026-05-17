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

echo "    DB_USER     = $DB_USER"
echo "    DB_PASSWORD = $DB_PASSWORD"
echo "    DB_NAME     = $DB_NAME"
echo "    JWT_SECRET  = $JWT_SECRET"

sed \
    -e "s|CHANGE_ME_DB_USER|$DB_USER|g" \
    -e "s|CHANGE_ME_DB_PASSWORD|$DB_PASSWORD|g" \
    -e "s|CHANGE_ME_DB_NAME|$DB_NAME|g" \
    "$SCRIPT_DIR/config.yaml.example" > "$SCRIPT_DIR/config.yaml"

sed \
    -e "s|CHANGE_ME_DB_USER|$DB_USER|g" \
    -e "s|CHANGE_ME_DB_PASSWORD|$DB_PASSWORD|g" \
    -e "s|CHANGE_ME_DB_NAME|$DB_NAME|g" \
    "$SCRIPT_DIR/docker-compose.yml.example" > "$SCRIPT_DIR/docker-compose.yml"

sed -i.bak "s|CHANGE_ME_JWT_SECRET|$JWT_SECRET|g" "$SCRIPT_DIR/config.yaml"
rm -f "$SCRIPT_DIR/config.yaml.bak"

echo "==> 配置文件已生成：config.yaml, docker-compose.yml"