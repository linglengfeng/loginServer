#!/usr/bin/env bash

# 切到脚本所在目录，确保使用本目录下的二进制
cd "$(dirname "$0")" || exit 1

# Initialize database
mysql -hlocalhost -P3306 -uroot -p123456 -e "source ../sql/server.sql"
if [ $? -ne 0 ]; then
    echo "Database initialization failed!"
    exit 1
fi
echo "Database initialization successful!"
echo ""

# 检查并设置 loginServer 执行权限（从 Windows 复制过来的文件可能没有执行权限）
if [ -f "./loginServer" ]; then
    if [ ! -x "./loginServer" ]; then
        echo "Setting execute permission for loginServer..."
        chmod +x ./loginServer
    fi
else
    echo "Error: loginServer not found in $(pwd)"
    exit 1
fi

# 后台运行，将输出写入 loginServer.log
nohup ./loginServer > ./loginServer.log 2>&1 &

echo "loginServer started in background. log: $(pwd)/loginServer.log"

