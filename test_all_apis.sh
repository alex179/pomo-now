#!/bin/bash

BASE_URL="http://localhost:8080"
API_BASE="${BASE_URL}/api"

echo "🧪 Pomo-Now API 接口全面测试"
echo "=================================="

# 1. 测试用户注册
echo "📝 测试用户注册..."
REGISTER_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"api-test@example.com","username":"apitest","password":"password123"}')

echo "注册响应: $REGISTER_RESPONSE"

# 提取token（如果注册成功）
TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 如果注册失败（用户已存在），尝试登录
if [ -z "$TOKEN" ]; then
  echo "🔑 用户已存在，尝试登录..."
  LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"api-test@example.com","password":"password123"}')
  
  echo "登录响应: $LOGIN_RESPONSE"
  TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$TOKEN" ]; then
  echo "❌ 无法获取认证token，停止测试"
  exit 1
fi

echo "✅ 获得认证token: ${TOKEN:0:20}..."

# 2. 测试获取用户信息
echo ""
echo "👤 测试获取用户信息..."
USER_INFO=$(curl -s -X GET "${API_BASE}/auth/me" \
  -H "Authorization: Bearer $TOKEN")
echo "用户信息: $USER_INFO"

# 3. 测试任务API
echo ""
echo "📋 测试任务API..."

# 获取任务列表
echo "获取任务列表..."
TASKS_RESPONSE=$(curl -s -X GET "${API_BASE}/tasks" \
  -H "Authorization: Bearer $TOKEN")
echo "任务列表: $TASKS_RESPONSE"

# 创建任务
echo "创建新任务..."
CREATE_TASK_RESPONSE=$(curl -s -X POST "${API_BASE}/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试任务","description":"这是一个测试任务","estimated_pomodoros":4,"status":"pending"}')
echo "创建任务响应: $CREATE_TASK_RESPONSE"

# 4. 测试统计API
echo ""
echo "📊 测试统计API..."

# 今日统计
echo "获取今日统计..."
TODAY_STATS=$(curl -s -X GET "${API_BASE}/stats?period=today" \
  -H "Authorization: Bearer $TOKEN")
echo "今日统计: $TODAY_STATS"

# 本周统计
echo "获取本周统计..."
WEEK_STATS=$(curl -s -X GET "${API_BASE}/stats?period=week" \
  -H "Authorization: Bearer $TOKEN")
echo "本周统计: $WEEK_STATS"

# 本月统计
echo "获取本月统计..."
MONTH_STATS=$(curl -s -X GET "${API_BASE}/stats?period=month" \
  -H "Authorization: Bearer $TOKEN")
echo "本月统计: $MONTH_STATS"

# 每小时统计
echo "获取每小时统计..."
HOURLY_STATS=$(curl -s -X GET "${API_BASE}/stats/hourly" \
  -H "Authorization: Bearer $TOKEN")
echo "每小时统计: $HOURLY_STATS"

# 已完成任务
echo "获取已完成任务..."
COMPLETED_TASKS=$(curl -s -X GET "${API_BASE}/stats/completed-tasks" \
  -H "Authorization: Bearer $TOKEN")
echo "已完成任务: $COMPLETED_TASKS"

# 5. 测试设置API
echo ""
echo "⚙️ 测试设置API..."

# 获取番茄钟设置
echo "获取番茄钟设置..."
POMODORO_SETTINGS=$(curl -s -X GET "${API_BASE}/settings/pomodoro" \
  -H "Authorization: Bearer $TOKEN")
echo "番茄钟设置: $POMODORO_SETTINGS"

# 更新番茄钟设置
echo "更新番茄钟设置..."
UPDATE_POMODORO=$(curl -s -X PUT "${API_BASE}/settings/pomodoro" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"work_minutes":25,"short_break_minutes":5,"long_break_minutes":15,"long_break_interval":4}')
echo "更新番茄钟设置响应: $UPDATE_POMODORO"

# 获取通知设置
echo "获取通知设置..."
NOTIFICATION_SETTINGS=$(curl -s -X GET "${API_BASE}/settings/notification" \
  -H "Authorization: Bearer $TOKEN")
echo "通知设置: $NOTIFICATION_SETTINGS"

# 6. 测试用户管理API
echo ""
echo "👥 测试用户管理API..."

# 获取用户资料
echo "获取用户资料..."
USER_PROFILE=$(curl -s -X GET "${API_BASE}/user/profile" \
  -H "Authorization: Bearer $TOKEN")
echo "用户资料: $USER_PROFILE"

echo ""
echo "🎉 API测试完成！"
echo "=================================="

# 检查所有响应是否包含错误
ERROR_COUNT=0

if echo "$REGISTER_RESPONSE$LOGIN_RESPONSE" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if echo "$USER_INFO" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if echo "$TASKS_RESPONSE" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if echo "$CREATE_TASK_RESPONSE" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if echo "$TODAY_STATS$WEEK_STATS$MONTH_STATS$HOURLY_STATS$COMPLETED_TASKS" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if echo "$POMODORO_SETTINGS$UPDATE_POMODORO$NOTIFICATION_SETTINGS" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if echo "$USER_PROFILE" | grep -q "error"; then
  ((ERROR_COUNT++))
fi

if [ $ERROR_COUNT -eq 0 ]; then
  echo "✅ 所有API接口测试通过！"
else
  echo "⚠️  发现 $ERROR_COUNT 个API接口错误"
fi 