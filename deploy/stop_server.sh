#!/usr/bin/env bash

# 切到脚本所在目录，确保在 deploy/linux 下执行
cd "$(dirname "$0")" || exit 1

# 优雅停止 gvaServer（若需强制可改用 pkill -9）
if pkill -f loginServer; then
  echo "loginServer stopped."
else
  echo "no loginServer process found."
fi

