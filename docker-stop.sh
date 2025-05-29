#!/bin/bash

# Pomo-Now Docker 停止脚本
echo "🛑 停止 Pomo-Now 应用..."

# 停止并移除容器
docker-compose down

echo "✅ 应用已停止"

# 询问是否要清理数据
read -p "是否要删除数据库数据？(这将删除所有用户数据) [y/N]: " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  清理数据库数据..."
    docker-compose down -v
    docker volume rm pomo-now_mysql_data 2>/dev/null || true
    echo "✅ 数据已清理"
fi

# 询问是否要清理镜像
read -p "是否要删除Docker镜像？ [y/N]: " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️  清理Docker镜像..."
    docker rmi pomo-now_backend pomo-now_frontend 2>/dev/null || true
    docker image prune -f
    echo "✅ 镜像已清理"
fi

echo "👋 Pomo-Now 应用已完全停止" 