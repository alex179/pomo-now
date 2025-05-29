# 🍅 Pomo-Now - 现代化番茄钟专注应用

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/react-18-blue.svg)](https://reactjs.org/)
[![Docker](https://img.shields.io/badge/docker-supported-blue.svg)](https://www.docker.com/)

一个现代化的番茄钟专注应用，帮助您提高工作效率和专注力。采用 Go + React 技术栈，支持Docker容器化部署。

## ✨ 核心功能

### 🍅 番茄钟核心
- ⏱️ 经典25分钟专注，5分钟短休息，15分钟长休息
- 🔔 智能通知系统，工作和休息时间提醒
- ⚙️ 完全可自定义的时间设置
- 🎵 多种提示音选择

### 📝 任务管理
- ✅ 创建、编辑、删除、完成任务
- 🎯 任务状态跟踪（待办、进行中、已完成）
- 📊 任务完成进度统计
- 🏷️ 任务分类和标记

### 📊 数据统计
- 📈 每日、每周、每月专注数据分析
- ⏰ 每小时专注分布图表
- 🎯 专注目标设定和追踪
- 📋 完成任务统计

### 👤 用户系统
- 🔐 安全的用户注册和登录
- 👤 个人资料管理
- 🔒 密码和邮箱修改
- 🖼️ 头像上传支持

## 🚀 快速开始

### Docker 部署（推荐）

```bash
# 克隆仓库
git clone https://github.com/your-username/pomo-now.git
cd pomo-now

# 一键启动（包含数据库）
./docker-start.sh
```

访问应用：
- 🌐 前端界面: http://localhost:3000
- 🔧 后端API: http://localhost:8080

### 传统部署

#### 系统要求
- **Go**: 1.21+
- **Node.js**: 18+
- **MySQL**: 5.7+ 或 8.0+

#### 安装步骤

1. **后端设置**
```bash
cd api
go mod download
go build -o pomo-now-server .
```

2. **前端设置**
```bash
cd web
npm install
npm run build
```

3. **数据库设置**
```sql
CREATE DATABASE pomo_now CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

4. **启动应用**
```bash
# 启动后端
./api/pomo-now-server &

# 启动前端
cd web && npm start
```

## 🛠️ 技术架构

### 后端技术栈
- **Go 1.21+** - 高性能后端服务
- **标准库HTTP** - 原生HTTP服务器和路由
- **MySQL** - 可靠的数据存储
- **JWT** - 安全的身份认证

### 前端技术栈
- **React 18** - 现代化用户界面框架
- **TypeScript** - 类型安全开发
- **Material-UI** - 优雅的组件库
- **响应式设计** - 支持各种设备

### DevOps
- **Docker** - 容器化部署
- **Docker Compose** - 服务编排
- **Nginx** - 反向代理和静态文件服务
- **Multi-stage Build** - 优化镜像大小

## 📊 API 文档

### 认证端点
```
POST /api/auth/register  # 用户注册
POST /api/auth/login     # 用户登录
GET  /api/auth/me        # 获取用户信息
PUT  /api/auth/me        # 更新用户信息
```

### 任务管理
```
GET    /api/tasks        # 获取任务列表
POST   /api/tasks        # 创建新任务
GET    /api/tasks/{id}   # 获取单个任务
PUT    /api/tasks/{id}   # 更新任务
DELETE /api/tasks/{id}   # 删除任务
```

### 统计数据
```
GET /api/stats?period=today   # 今日统计
GET /api/stats?period=week    # 本周统计
GET /api/stats?period=month   # 本月统计
GET /api/stats/hourly         # 每小时统计
```

### 设置管理
```
GET /api/settings/pomodoro      # 获取番茄钟设置
PUT /api/settings/pomodoro      # 更新番茄钟设置
GET /api/settings/notification # 获取通知设置
PUT /api/settings/notification # 更新通知设置
```

## 🔧 配置选项

### 环境变量
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=pomo_now

# 服务器配置
PORT=8080
FRONTEND_PORT=3000

# 安全配置
JWT_SECRET=your-secret-key
```

### Docker配置
创建 `.env` 文件：
```bash
cp docker.env.example .env
# 编辑 .env 文件设置您的配置
```

## 🧪 测试

### API测试
```bash
# 启动应用后运行API测试
./test_all_apis.sh
```

### 单元测试
```bash
# 后端测试
cd api && go test ./...

# 前端测试  
cd web && npm test
```

## 📦 部署

### Docker生产部署
```bash
# 构建生产镜像
docker-compose -f docker-compose.prod.yml up -d

# 查看日志
docker-compose logs -f

# 扩展服务
docker-compose up -d --scale backend=3
```

### 传统部署
详见 [部署文档](./docs/deployment.md)

## 🤝 贡献指南

我们欢迎所有形式的贡献！

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 开发环境设置
```bash
# 后端开发
cd api
go mod download
go run main.go

# 前端开发
cd web  
npm install
npm start
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 支持

### 问题报告
如果您遇到bug或有功能建议，请 [创建Issue](https://github.com/your-username/pomo-now/issues)。

### 常见问题

<details>
<summary>Docker启动失败怎么办？</summary>

1. 确保Docker服务已启动
2. 检查端口是否被占用
3. 查看容器日志：`docker-compose logs`
</details>

<details>
<summary>数据库连接失败？</summary>

1. 检查MySQL服务状态
2. 验证数据库配置
3. 确保数据库已创建
</details>

<details>
<summary>如何备份数据？</summary>

```bash
# 备份数据库
docker-compose exec database mysqldump -u root -p pomo_now > backup.sql

# 恢复数据库  
docker-compose exec -T database mysql -u root -p pomo_now < backup.sql
```
</details>

## 🎯 路线图

- [ ] 📱 PWA支持
- [ ] 🌙 深色模式
- [ ] 👥 团队协作
- [ ] 📊 高级统计
- [ ] 🔄 数据同步
- [ ] 🎮 成就系统

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

**🍅 开始您的专注之旅，提高工作效率！**

如果这个项目对您有帮助，请给我们一个 ⭐ Star！ 