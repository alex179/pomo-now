# OAuth 登录集成指南

## 概述

本项目已完成 Apple ID 和 Google 账户登录的基础集成，包含：

1. 后端 OAuth 服务实现
2. 前端 OAuth 组件集成
3. 模拟登录功能（用于开发测试）

## 当前实现状态

### ✅ 已完成功能

1. **后端 OAuth 基础架构**
   - OAuth 服务抽象层 (`api/internal/service/oauth.go`)
   - OAuth 配置管理 (`api/internal/config/oauth.go`)
   - Apple/Google 登录 API 端点
   - 用户创建和认证流程

2. **前端 OAuth 集成**
   - OAuth hooks (`web/src/hooks/useOAuth.ts`)
   - 登录/注册表单 OAuth 按钮
   - API 客户端 OAuth 方法
   - 错误处理和加载状态

3. **模拟登录功能**
   - 开发环境下可测试的模拟 Apple/Google 登录
   - 自动创建测试用户
   - 完整的认证流程

### 🚧 需要配置的功能

1. **Google OAuth 配置**
   - 需要在 Google Cloud Console 创建项目
   - 获取 Google Client ID 和 Client Secret
   - 配置回调 URL

2. **Apple Sign In 配置**
   - 需要在 Apple Developer 账户中配置
   - 获取 Apple Client ID、Team ID、Key ID
   - 配置私钥文件

## 配置步骤

### Google OAuth 配置

1. **创建 Google Cloud 项目**
   ```
   https://console.cloud.google.com/
   ```

2. **启用 Google+ API**
   - 在 API 库中搜索并启用 "Google+ API"
   - 启用 "Google OAuth2 API"

3. **创建 OAuth 2.0 凭据**
   - 转到 "凭据" → "创建凭据" → "OAuth 客户端 ID"
   - 应用类型：Web 应用
   - 授权重定向 URI：`http://localhost:3000/auth/google/callback`

4. **配置环境变量**
   ```bash
   # 后端环境变量
   export GOOGLE_CLIENT_ID="your-google-client-id"
   export GOOGLE_CLIENT_SECRET="your-google-client-secret"
   export OAUTH_REDIRECT_URL="http://localhost:3000/auth/callback"
   
   # 前端环境变量（web/.env.local）
   REACT_APP_GOOGLE_CLIENT_ID=your-google-client-id
   ```

### Apple Sign In 配置

1. **Apple Developer 配置**
   ```
   https://developer.apple.com/account/
   ```

2. **创建 App ID**
   - 启用 "Sign In with Apple"
   - 配置域名和回调 URL

3. **创建服务 ID**
   - 用于 Web 认证
   - 配置域名：`localhost:3000`（开发环境）
   - 回调 URL：`http://localhost:3000/auth/apple/callback`

4. **创建私钥**
   - 下载 .p8 私钥文件
   - 记录 Key ID 和 Team ID

5. **配置环境变量**
   ```bash
   # 后端环境变量
   export APPLE_CLIENT_ID="com.yourcompany.pomonoow"
   export APPLE_TEAM_ID="your-team-id"
   export APPLE_KEY_ID="your-key-id"
   export APPLE_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
   
   # 前端环境变量（web/.env.local）
   REACT_APP_APPLE_CLIENT_ID=com.yourcompany.pomonoow
   ```

## 使用方法

### 开发环境（模拟登录）

当前实现包含模拟登录功能，无需真实的 OAuth 配置即可测试：

```typescript
// 模拟 Apple 登录
const authResponse = await apiClient.appleLogin(
  'mock_apple_token_' + Date.now(),
  { givenName: 'Apple用户', familyName: '' }
);

// 模拟 Google 登录
const authResponse = await apiClient.mockGoogleLogin(
  'mock_google_token_' + Date.now(),
  { givenName: 'Google用户', familyName: '' }
);
```

### 生产环境（真实 OAuth）

替换 `web/src/hooks/useOAuth.ts` 中的模拟实现：

```typescript
// Google 真实登录实现
const googleLogin = useCallback(async () => {
  try {
    if (!window.google) {
      throw new Error('Google Sign-In SDK 未加载');
    }

    // 初始化 Google Sign-In
    window.google.accounts.id.initialize({
      client_id: process.env.REACT_APP_GOOGLE_CLIENT_ID,
      callback: async (response: any) => {
        const authResponse = await apiClient.googleLogin(response.credential);
        return authResponse;
      }
    });

    // 显示登录弹窗
    window.google.accounts.id.prompt();
  } catch (error) {
    setError('Google 登录失败');
  }
}, []);

// Apple 真实登录实现
const appleLogin = useCallback(async () => {
  try {
    if (!window.AppleID) {
      throw new Error('Apple Sign In SDK 未加载');
    }

    await window.AppleID.auth.init({
      clientId: process.env.REACT_APP_APPLE_CLIENT_ID,
      scope: 'name email',
      redirectURI: window.location.origin + '/auth/apple/callback',
      usePopup: true,
    });

    const data = await window.AppleID.auth.signIn();
    const authResponse = await apiClient.appleLogin(
      data.authorization.id_token,
      data.user?.name
    );
    return authResponse;
  } catch (error) {
    setError('Apple 登录失败');
  }
}, []);
```

## API 端点

### 后端 OAuth API

```
POST /api/auth/google         # Google 登录
POST /api/auth/apple          # Apple 登录
GET  /api/auth/google/url     # 获取 Google 授权 URL
GET  /api/auth/apple/url      # 获取 Apple 授权 URL
```

### 请求格式

```json
// Google 登录
{
  "code": "google-authorization-code"
}

// Apple 登录
{
  "identity_token": "apple-id-token",
  "fullName": {
    "givenName": "用户名",
    "familyName": "姓氏"
  }
}
```

## 测试指南

### 1. 启动服务

```bash
# 启动后端
cd api && go run main.go

# 启动前端
cd web && npm start
```

### 2. 测试模拟登录

1. 访问 `http://localhost:3000`
2. 点击 "Apple ID 登录" 或 "Google 账户登录"
3. 系统会自动创建模拟用户并完成登录

### 3. 测试真实 OAuth（需要配置）

1. 完成上述配置步骤
2. 替换模拟实现为真实实现
3. 测试真实的 OAuth 登录流程

## 注意事项

1. **开发环境**：使用模拟登录，无需配置真实 OAuth
2. **生产环境**：必须配置真实的 Google/Apple OAuth
3. **安全性**：生产环境中务必妥善保管私钥和密钥
4. **HTTPS**：Apple Sign In 在生产环境要求 HTTPS
5. **域名**：OAuth 回调 URL 必须与配置的域名匹配

## 故障排除

### 常见问题

1. **Google 登录失败**
   - 检查 Client ID 配置
   - 确认回调 URL 配置正确
   - 验证 API 是否启用

2. **Apple 登录失败**
   - 检查 Client ID、Team ID、Key ID
   - 验证私钥格式是否正确
   - 确认域名和回调 URL 配置

3. **SDK 加载失败**
   - 检查网络连接
   - 确认 HTML 中的 SDK 引用正确
   - 验证环境变量配置

### 调试方法

```typescript
// 启用详细错误日志
console.log('OAuth Error:', error);
console.log('Window.google:', window.google);
console.log('Window.AppleID:', window.AppleID);
```

## 下一步开发

1. 实现真实的 Google OAuth 集成
2. 实现真实的 Apple Sign In 集成
3. 添加 OAuth 用户信息同步
4. 实现账户绑定功能
5. 添加第三方登录的安全机制 