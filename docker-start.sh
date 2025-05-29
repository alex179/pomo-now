#!/bin/bash

# Pomo-Now Docker 一键启动脚本
echo "🍅 Pomo-Now Docker 部署启动中..."
echo "================================="

# 检查Docker和docker-compose是否安装
check_requirements() {
    echo "🔍 检查系统要求..."
    
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker 未安装"
        echo "   请访问 https://docs.docker.com/get-docker/ 安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo "❌ Docker Compose 未安装"
        echo "   请访问 https://docs.docker.com/compose/install/ 安装Docker Compose"
        exit 1
    fi
    
    echo "✅ Docker 已安装"
    echo "✅ Docker Compose 已安装"
    
    # 检查Docker服务是否运行
    if ! docker info &> /dev/null; then
        echo "❌ Docker 服务未运行，请先启动Docker"
        exit 1
    fi
    
    echo "✅ Docker 服务正在运行"
}

# 创建环境文件
setup_env() {
    if [ ! -f .env ]; then
        echo "📝 创建环境配置文件..."
        cp docker.env.example .env
        echo "✅ 已创建 .env 文件，您可以根据需要修改配置"
    else
        echo "✅ 环境配置文件已存在"
    fi
}

# 构建并启动服务
start_services() {
    echo ""
    echo "🏗️  构建Docker镜像..."
    docker-compose build --no-cache
    
    if [ $? -ne 0 ]; then
        echo "❌ Docker镜像构建失败"
        exit 1
    fi
    
    echo ""
    echo "🚀 启动服务..."
    docker-compose up -d
    
    if [ $? -ne 0 ]; then
        echo "❌ 服务启动失败"
        exit 1
    fi
    
    echo "✅ 服务启动成功"
}

# 等待服务就绪
wait_for_services() {
    echo ""
    echo "⏳ 等待服务启动..."
    
    # 等待数据库就绪
    echo "等待数据库启动..."
    while ! docker-compose exec -T database mysqladmin ping -h localhost --silent; do
        sleep 2
        echo -n "."
    done
    echo " ✅ 数据库已就绪"
    
    # 等待后端服务就绪
    echo "等待后端服务启动..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo " ✅ 后端服务已就绪"
            break
        fi
        sleep 2
        echo -n "."
        if [ $i -eq 30 ]; then
            echo " ⚠️  后端服务启动超时，但可能仍在启动中"
        fi
    done
    
    # 等待前端服务就绪
    echo "等待前端服务启动..."
    for i in {1..20}; do
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            echo " ✅ 前端服务已就绪"
            break
        fi
        sleep 2
        echo -n "."
        if [ $i -eq 20 ]; then
            echo " ⚠️  前端服务启动超时，但可能仍在启动中"
        fi
    done
}

# 显示访问信息
show_info() {
    echo ""
    echo "🎉 Pomo-Now 应用部署完成！"
    echo "=========================="
    echo ""
    echo "📱 应用地址:"
    echo "   - 前端界面: http://localhost:3000"
    echo "   - 后端API: http://localhost:8080"
    echo "   - 健康检查: http://localhost:8080/health"
    echo ""
    echo "🗄️  数据库:"
    echo "   - 主机: localhost:3306"
    echo "   - 数据库名: pomo_now" 
    echo "   - 用户名: pomo_user"
    echo ""
    echo "🔧 管理命令:"
    echo "   - 查看日志: docker-compose logs -f"
    echo "   - 停止应用: docker-compose down"
    echo "   - 重启应用: docker-compose restart"
    echo "   - 查看状态: docker-compose ps"
    echo ""
    echo "💡 使用提示:"
    echo "   1. 打开浏览器访问 http://localhost:3000"
    echo "   2. 注册新用户账户"
    echo "   3. 开始您的番茄钟专注之旅！"
    echo ""
    echo "🛑 停止应用: ./docker-stop.sh 或 docker-compose down"
}

# 主函数
main() {
    check_requirements
    setup_env
    start_services
    wait_for_services
    show_info
}

# 执行主函数
main "$@" 