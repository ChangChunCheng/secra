#!/bin/bash
set -e

# 切換 Go 版本（視開發環境需要）
source ~/.gvm/scripts/gvm
gvm use go1.23.8

# 設定環境變數（如果不是靠 .env）
export YEAR=$(date +%Y)
export SOURCE=nvd-cve

# 切到專案根目錄
cd "$(dirname "$0")/.."

# 執行 make 任務
make import-nvd