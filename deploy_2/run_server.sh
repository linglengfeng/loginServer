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

# 检查并设置 loginServer2 执行权限（从 Windows 复制过来的文件可能没有执行权限）
if [ -f "./loginServer2" ]; then
    if [ ! -x "./loginServer2" ]; then
        echo "Setting execute permission for loginServer2..."
        chmod +x ./loginServer2
    fi
else
    echo "Error: loginServer2 not found in $(pwd)"
    exit 1
fi

# 后台运行，将输出写入 loginServer2.log
nohup ./loginServer2 > ./loginServer2.log 2>&1 &

echo "loginServer2 started in background. log: $(pwd)/loginServer2.log"

