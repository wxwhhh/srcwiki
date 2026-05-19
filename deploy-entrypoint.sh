#!/bin/sh
set -e

# 每次启动时，将镜像中的前端文件复制到共享目录
# 这样 rebuild 后前端文件总是最新的
if [ -d /app/frontend-dist ]; then
    # 清空旧文件（保留目录结构）
    rm -rf /srv/admin/*
    cp -r /app/frontend-dist/* /srv/admin/ 2>/dev/null || true
    echo "[entrypoint] 前端文件已同步到 /srv/admin"
fi

# 启动后端
exec /app/litewiki
