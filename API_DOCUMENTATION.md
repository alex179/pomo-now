# 番茄时光 API 文档

## 概述
番茄时光后端API提供完整的番茄钟应用功能支持，包括用户认证、任务管理、会话管理、统计数据和设置管理。

## 基础信息
- **Base URL**: `http://localhost:8080/api`
- **认证方式**: Bearer Token (JWT)
- **响应格式**: JSON

## 标准响应格式

### 成功响应
```json
{
  "status": "success",
  "data": {
    // 具体数据
  }
}
```

### 错误响应
```json
{
  "status": "error",
  "error": {
    "code": "ERROR_CODE",
    "message": "用户友好的错误信息"
  }
}
```

## 认证接口

### 1. 用户注册
**POST** `/auth/register`

#### 请求体
```json
{
  "email": "user@example.com",
  "username": "用户名",
  "password": "password123"
}
```

#### 响应
```json
{
  "status": "success",
  "data": {
    "token": "jwt_token_here",
    "user": {
      "id": "user_id",
      "email": "user@example.com",
      "username": "用户名",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 错误码
- `VALIDATION_ERROR`: 请求格式错误
- `CONFLICT`: 邮箱已存在

### 2. 用户登录
**POST** `/auth/login`

#### 请求体
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

#### 响应
同注册接口

#### 错误码
- `VALIDATION_ERROR`: 邮箱或密码错误
- `NOT_FOUND`: 用户不存在

### 3. Apple登录
**POST** `/auth/apple`

#### 请求体
```json
{
  "identity_token": "apple_identity_token",
  "fullName": {
    "givenName": "名",
    "familyName": "姓"
  }
}
```

#### 响应
同注册接口

## 任务管理接口

### 4. 获取任务列表
**GET** `/tasks`

#### Headers
```
Authorization: Bearer <token>
```

#### 响应
```json
{
  "status": "success",
  "data": [
    {
      "id": "task_id",
      "user_id": "user_id",
      "title": "任务标题",
      "description": "任务描述",
      "estimated_pomodoros": 3,
      "status": "pending",
      "color": "red",
      "notes": "备注",
      "created_at": "2024-01-01T00:00:00Z",
      "completed_at": null
    }
  ]
}
```

### 5. 创建任务
**POST** `/tasks`

#### 请求体
```json
{
  "title": "任务标题",
  "description": "任务描述",
  "estimated_pomodoros": 3,
  "color": "red",
  "notes": "备注"
}
```

#### 响应
返回创建的任务对象

### 6. 获取单个任务
**GET** `/tasks/{id}`

#### 响应
返回任务对象

### 7. 更新任务
**PUT** `/tasks/{id}`

#### 请求体
```json
{
  "title": "新标题",
  "estimated_pomodoros": 5,
  "status": "completed",
  "color": "blue",
  "notes": "更新的备注"
}
```

### 8. 删除任务
**DELETE** `/tasks/{id}`

#### 响应
```json
{
  "status": "success",
  "data": {
    "message": "任务删除成功"
  }
}
```

## 会话管理接口

### 9. 获取会话列表
**GET** `/sessions`

#### 响应
```json
{
  "status": "success",
  "data": [
    {
      "id": "session_id",
      "user_id": "user_id",
      "task_id": "task_id",
      "type": "focus",
      "duration_minutes": 25,
      "started_at": "2024-01-01T00:00:00Z",
      "ended_at": "2024-01-01T00:25:00Z"
    }
  ]
}
```

### 10. 创建会话
**POST** `/sessions`

#### 请求体
```json
{
  "task_id": "task_id",
  "type": "focus",
  "duration_minutes": 25
}
```

### 11. 获取当前会话
**GET** `/sessions/current`

#### 响应
返回当前进行中的会话，如果没有则返回null

### 12. 结束当前会话
**PUT** `/sessions/current`

#### 请求体
```json
{
  "session_id": "session_id"
}
```

## 统计接口

### 13. 获取统计数据
**GET** `/stats?period={today|week|month}`

#### 参数
- `period`: 统计周期，可选值：today, week, month

#### 响应
```json
{
  "status": "success",
  "data": {
    "total_pomodoros": 15,
    "total_minutes": 375,
    "completed_tasks": 5,
    "daily_average": 2.5,
    "streak_days": 7,
    "completion_rate": 0.85
  }
}
```

## 设置管理接口

### 14. 获取番茄钟设置
**GET** `/settings/pomodoro`

#### 响应
```json
{
  "status": "success",
  "data": {
    "id": "settings_id",
    "user_id": "user_id",
    "work_minutes": 25,
    "short_break_minutes": 5,
    "long_break_minutes": 15,
    "long_break_interval": 4,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 15. 更新番茄钟设置
**PUT** `/settings/pomodoro`

#### 请求体
```json
{
  "work_minutes": 30,
  "short_break_minutes": 10,
  "long_break_minutes": 20,
  "long_break_interval": 3
}
```

### 16. 获取通知设置
**GET** `/settings/notification`

#### 响应
```json
{
  "status": "success",
  "data": {
    "id": "settings_id",
    "user_id": "user_id",
    "work_start_notification": true,
    "work_end_notification": true,
    "break_start_notification": true,
    "break_end_notification": true,
    "sound_enabled": true,
    "selected_sound": "default",
    "volume": 70,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 17. 更新通知设置
**PUT** `/settings/notification`

#### 请求体
```json
{
  "work_start_notification": false,
  "work_end_notification": true,
  "break_start_notification": true,
  "break_end_notification": true,
  "sound_enabled": false,
  "selected_sound": "bell",
  "volume": 50
}
```

## 用户管理接口

### 18. 获取用户资料
**GET** `/user/profile`

#### 响应
```json
{
  "status": "success",
  "data": {
    "id": "user_id",
    "email": "user@example.com",
    "username": "用户名",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "profile": {
      "id": "profile_id",
      "user_id": "user_id",
      "username": "显示名称",
      "avatar": "avatar_url",
      "bio": "个人简介",
      "timezone": "Asia/Shanghai",
      "language": "zh-CN",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### 19. 更新用户资料
**PUT** `/user/profile`

#### 请求体
```json
{
  "username": "新用户名",
  "bio": "新的个人简介",
  "timezone": "America/New_York",
  "language": "en-US"
}
```

### 20. 修改密码
**POST** `/user/change-password`

#### 请求体
```json
{
  "current_password": "旧密码",
  "new_password": "新密码"
}
```

#### 响应
```json
{
  "status": "success",
  "data": {
    "message": "密码修改成功"
  }
}
```

### 21. 修改邮箱
**POST** `/user/change-email`

#### 请求体
```json
{
  "new_email": "new@example.com",
  "password": "确认密码"
}
```

## 前端需要实现的功能清单

### 高优先级（必须立即修复）
1. **番茄钟设置API对接** - PomodoroSettingsPage
   - 接口：GET/PUT `/settings/pomodoro`
   - 需要修改：`handleSavePomodoroSettings` 方法

2. **通知设置API对接** - NotificationSettingsPage
   - 接口：GET/PUT `/settings/notification`
   - 需要修改：`handleSaveNotificationSettings` 方法

3. **修改密码功能** - ChangePasswordPage
   - 接口：POST `/user/change-password`
   - 需要修改：`handleChangePassword` 方法

4. **用户资料管理** - AccountSecurityPage
   - 接口：GET/PUT `/user/profile`，POST `/user/change-email`
   - 需要实现：用户名修改、邮箱修改功能

### 中优先级（增强功能）
5. **实时计时器状态同步**
   - 接口：GET/POST/PUT `/sessions`，GET/PUT `/sessions/current`
   - 需要在TimerPage中实现会话管理

6. **任务详情扩展**
   - 添加任务描述、颜色、备注字段的编辑功能
   - 修改CreateTaskPage和TaskListPage

### 低优先级（优化功能）
7. **头像上传功能**
8. **关联账户管理**
9. **忘记密码功能**

## 错误码对照表
- `VALIDATION_ERROR`: 请求格式错误或验证失败
- `UNAUTHORIZED`: 未登录或token无效
- `FORBIDDEN`: 权限不足
- `NOT_FOUND`: 资源不存在
- `CONFLICT`: 资源冲突（如邮箱已存在）
- `INTERNAL_ERROR`: 服务器内部错误

## 测试建议
1. 使用Postman或类似工具测试所有API接口
2. 验证错误处理和状态码
3. 测试认证中间件
4. 验证数据验证规则
5. 测试并发访问和性能 