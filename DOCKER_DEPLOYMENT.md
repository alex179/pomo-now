# 🐳 Pomo-Now Docker 部署指南

## 🚀 快速部署

### 1. 克隆仓库
```bash
git clone https://github.com/your-username/pomo-now.git
cd pomo-now
```

### 2. 一键启动
```bash
./docker-start.sh
```

### 3. 访问应用
- 前端: http://localhost:3000
- 后端API: http://localhost:8080
- 健康检查: http://localhost:8080/health

## 📋 系统要求

- Docker 20.10+
- Docker Compose v2.0+
- 4GB+ 可用内存
- 2GB+ 可用磁盘空间

## ⚙️ 配置

### 环境变量
复制并编辑环境配置：
```bash
cp docker.env.example .env
```

重要配置项：
```bash
# 数据库
DB_ROOT_PASSWORD=rootpassword123
DB_NAME=pomo_now
DB_USER=pomo_user
DB_PASSWORD=password123

# 安全
JWT_SECRET=your-super-secret-jwt-key

# 端口
FRONTEND_PORT=3000
```

## 🏗️ 手动部署

### 1. 构建镜像
```bash
# 构建后端
docker build -f Dockerfile.backend -t pomo-now-backend .

# 构建前端
docker build -f Dockerfile.frontend -t pomo-now-frontend .
```

### 2. 启动服务
```bash
docker-compose up -d
```

### 3. 查看状态
```bash
docker-compose ps
docker-compose logs -f
```

## 🔧 管理命令

### 服务管理
```bash
# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 停止服务
docker-compose down

# 完全清理
./docker-stop.sh
```

### 数据库管理
```bash
# 进入数据库
docker-compose exec database mysql -u root -p

# 备份数据
docker-compose exec database mysqldump -u root -p pomo_now > backup.sql

# 恢复数据
docker-compose exec -T database mysql -u root -p pomo_now < backup.sql
```

## 🐛 故障排除

### 常见问题

#### 1. 端口被占用
```bash
# 检查端口占用
lsof -i :3000
lsof -i :8080

# 修改端口（在.env文件中）
FRONTEND_PORT=3001
```

#### 2. 镜像拉取失败
```bash
# 配置镜像源
echo '{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com"
  ]
}' | sudo tee /etc/docker/daemon.json

sudo systemctl restart docker
```

#### 3. 内存不足
```bash
# 检查Docker资源
docker system df
docker system prune -f

# 调整内存限制（docker-compose.yml）
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 512M
```

#### 4. 数据库连接失败
```bash
# 检查数据库状态
docker-compose exec database mysqladmin ping -h localhost -u root -p

# 重置数据库
docker-compose down -v
docker-compose up -d database
```

## 🔒 生产环境配置

### 1. 安全配置
```bash
# 强密码
DB_ROOT_PASSWORD=$(openssl rand -hex 32)
DB_PASSWORD=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 64)

# 限制网络访问
networks:
  pomo-network:
    driver: bridge
    internal: true
```

### 2. 性能优化
```bash
# 数据库优化
services:
  database:
    command: >
      --innodb_buffer_pool_size=512M
      --max_connections=100
      --query_cache_type=1
      --query_cache_size=64M
```

### 3. 监控配置
```bash
# 健康检查
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

## 📊 监控和日志

### 应用监控
```bash
# 实时日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 容器状态
docker stats

# 健康检查
curl http://localhost:8080/health
```

### 数据库监控
```bash
# 连接数
docker-compose exec database mysql -u root -p -e "SHOW STATUS LIKE 'Threads_connected';"

# 查询性能
docker-compose exec database mysql -u root -p -e "SHOW PROCESSLIST;"
```

## 🔄 更新部署

### 应用更新
```bash
# 拉取最新代码
git pull origin main

# 重新构建并部署
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### 零停机更新
```bash
# 滚动更新
docker-compose up -d --no-deps backend
docker-compose up -d --no-deps frontend
```

## 📦 镜像管理

### 镜像标记
```bash
# 标记版本
docker tag pomo-now-backend:latest pomo-now-backend:v1.0.0
docker tag pomo-now-frontend:latest pomo-now-frontend:v1.0.0
```

### 推送到registry
```bash
# 推送到Docker Hub
docker push your-username/pomo-now-backend:v1.0.0
docker push your-username/pomo-now-frontend:v1.0.0
```

## 🎯 最佳实践

### 1. 开发环境
- 使用volume mount代码目录
- 启用热重载
- 使用开发模式配置

### 2. 测试环境
- 独立的数据库实例
- 模拟生产配置
- 自动化测试集成

### 3. 生产环境
- 使用编译后的镜像
- 配置资源限制
- 启用健康检查
- 设置重启策略

---

如果遇到问题，请查看 [故障排除文档](./TROUBLESHOOTING.md) 或提交 [Issue](https://github.com/your-username/pomo-now/issues)。 