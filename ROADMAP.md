# Prism Roadmap

## ✅ 已完成

### Phase 1: 协议逆向 & 验证
- [x] SC 注册协议 (反机器人: spinner + honeypot + 3s 延迟)
- [x] SC 登录协议
- [x] GitHub 登录 + TOTP 2FA
- [x] SC GitHub OAuth 连接 (POST /auth/github → 3步重定向)
- [x] SC 项目创建 + 仓库绑定
- [x] SC Ticket 创建 (25个模型 ID)
- [x] SC GitHub 解绑 (DELETE /identities/{id})
- [x] 临时邮箱 (YYDS Mail API)
- [x] Python 协议库 (sc_proto.py) + 操作手册

### Phase 2: Go 后端核心
- [x] SC 协议引擎 Go 重写 (internal/scproto)
- [x] 浏览器指纹随机化 (UA/TLS/Sec-CH-UA/时间抖动)
- [x] 账号池管理 (internal/account)
- [x] 自动调度/换号 (internal/scheduler)
- [x] GitHub 登录模块 (internal/github)
- [x] REST API (internal/api)
- [x] E2E 测试通过 (27s 全链路)
- [x] GitHub OAuth App (Prism Agent Platform)
- [x] 用户 OAuth 登录 → 自动添加 collaborator

### Phase 3: Next.js 前端
- [x] 模型选择器 (4组 12个模型)
- [x] 任务输入 + Run Agent
- [x] GitHub OAuth 按钮 (用户只看到 "Prism", 不暴露 SC)
- [x] 仓库选择列表
- [x] 任务列表 + 状态轮询
- [x] 前后端 API 代理对接

---

## 📋 待开发

### Phase 4: LinuxDo 接入
> 优先级: 中 | 依赖: LinuxDo OAuth App + Credit 商户 API Key

#### 4.1 LinuxDo OAuth 登录
- [ ] 申请 OAuth App: https://connect.linux.do/dash/sso
- [ ] 接入端点:
  - 授权: `https://connect.linux.do/oauth2/authorize`
  - Token: `https://connect.linux.do/oauth2/token`
  - 用户信息: `https://connect.linux.do/api/user`
- [ ] 用户字段: id, username, name, avatar_template, trust_level
- [ ] Go 后端: `internal/api/linuxdo_oauth.go`
- [ ] 前端: 登录页添加 "Sign in with LinuxDo" 按钮
- [ ] 用户表存储 linuxdo_id + trust_level

#### 4.2 LinuxDo Credit 计费
- [ ] 注册商户: https://credit.linux.do
- [ ] 获取 Merchant API Key (client_id + client_secret)
- [ ] 接入支付 API:
  - 创建订单: `POST /pay/submit.php` (跳转收银台)
  - 查询订单: `GET /api.php?act=order&pid=&key=&out_trade_no=`
  - 退款: `POST /api.php`
  - 商户分发: `POST /pay/distribute` (Basic Auth)
- [ ] 计费逻辑: 按模型/任务时长计费, 从用户 LDC 余额扣除
- [ ] Go 后端: `internal/billing/linuxdo_credit.go`
- [ ] 前端: 余额显示 + 充值入口 + 消费记录

### Phase 5: 生产化
- [ ] PostgreSQL 持久化 (账号池、用户、任务记录)
- [ ] 用户系统 (注册/登录/Session)
- [ ] 多用户隔离 (每用户独立账号池 or 共享池)
- [ ] 任务历史 + 状态持久化
- [ ] SC 页面反代完善 (cookie 注入 + URL 重写)
- [ ] Docker Compose 部署
- [ ] 环境变量 .env 配置

### Phase 6: 增强
- [ ] 代理池支持 (HTTP/SOCKS5 代理轮转)
- [ ] 多 GitHub 服务账号池
- [ ] 任务结果通知 (WebSocket 推送)
- [ ] PR 自动提交到用户原始仓库
- [ ] Agent 对话记录存储 + 展示
- [ ] 额度预警 + 自动预注册账号
- [ ] 仪表盘统计 (使用量、成本、模型分布)

---

## 📐 架构

```
用户浏览器
  │
  ├─ LinuxDo OAuth 登录 (Phase 4)
  ├─ GitHub OAuth 绑定仓库
  │
  ▼
Next.js (:3001) ──proxy──▶ Go API (:8080)
                              │
                    ┌─────────┼─────────┐
                    ▼         ▼         ▼
               调度器     账号池    SC协议引擎
                 │         │         │
                 │    ┌────┴────┐    │
                 │    │ SC #1   │    │
                 └───▶│ SC #2   │◀───┘
                      │ SC #N   │
                      └────┬────┘
                           │
                    Superconductor
                      (Agent执行)
                           │
                      用户的 GitHub 仓库
```

## 🔑 所需凭据

| 凭据 | 用途 | 获取方式 |
|------|------|---------|
| YYDS Mail API Key | 创建临时邮箱 | 已有: `<YYDS_API_KEY>` |
| GitHub 服务账号 | SC 绑定 + collaborator | 已有: `<GITHUB_SERVICE_USER>` |
| GitHub OAuth App | 用户登录 | 已有: `<GITHUB_CLIENT_ID>` |
| LinuxDo OAuth App | 用户登录 | 待申请: connect.linux.do |
| LinuxDo Credit Key | 计费扣款 | 待申请: credit.linux.do |
