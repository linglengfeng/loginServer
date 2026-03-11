#!/usr/bin/env bash

# 切到脚本所在目录，确保在 deploy/linux 下执行
cd "$(dirname "$0")" || exit 1

# 优雅停止 loginServer2（若需强制可改用 pkill -9）
if pkill -f loginServer2; then
  echo "loginServer2 stopped."
else
  echo "no loginServer2 process found."
fi

