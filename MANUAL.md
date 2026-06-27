# Superconductor 协议化操作手册

## 概述

本工具包实现了 Superconductor 平台的完整协议化操作，包括账号注册、GitHub OAuth 连接、项目创建、任务分配等全流程自动化。无需浏览器，纯 Python 标准库实现。

## 文件说明

| 文件 | 用途 |
|------|------|
| `sc_proto.py` | 核心协议库 — 所有 API 封装 |
| `sc_register.py` | 批量注册脚本 (调用 sc_proto) |
| `create_ticket.py` | 单独创建 Ticket 脚本 |
| `sc_accounts.json` | 已注册账号记录 |

## 前置条件

- Python 3.10+ (仅标准库，无需 pip install)
- YYDS Mail API Key (`AC-` 前缀)
- GitHub 账号 (用户名 + 密码 + TOTP 密钥)

## 关键发现 (反检测机制)

| 机制 | 说明 |
|------|------|
| **蜜罐字段** | 注册/登录表单包含随机命名的 `type="text"` 隐藏字段，必须提交且值为空 |
| **Spinner** | 隐藏的 `spinner` 字段，服务端生成的反机器人 token |
| **时间检测** | GET 表单页与 POST 提交之间必须等待 **≥3 秒**，否则静默拒绝 |
| **字段顺序** | 使用 `OrderedDict` 保持与浏览器一致的提交顺序 |
| **GitHub 限制** | 一个 GitHub 账号只能绑定一个 SC 账号 |

## 可用模型 ID

```
python3 sc_proto.py models
```

### Codex (OpenAI)
- `codex_gpt_5_5_medium` — GPT-5.5 (Medium)
- `codex_gpt_5_5_high` — GPT-5.5 (High)
- `codex_gpt_5_5_xhigh` — GPT-5.5 (Xhigh)

### Claude Code
- `claude_code_claude_opus_4_8` — Opus 4.8 ⭐ 最强
- `claude_code_claude_opus_4_7` — Opus 4.7
- `claude_code_claude_opus_4_6` — Opus 4.6
- `claude_code_claude_opus_4_5` — Opus 4.5
- `claude_code_claude_sonnet_4_6` — Sonnet 4.6 (性价比)

### OpenCode
- `opencode_opus_4_8` / `opencode_sonnet_4_6` — Claude
- `opencode_gemini_3_1_pro` / `opencode_gemini_3_flash` — Gemini
- `opencode_gpt_5_5` / `opencode_gpt_5_5_pro` — GPT
- `opencode_glm_5_2` / `opencode_kimi_k2_6` — 国产模型

### Pi
- `pi_deepseek_v4_pro` / `pi_deepseek_v4_flash` — DeepSeek
- `pi_glm_5_2` / `pi_minimax_m2_7` / `pi_minimax_m3`

---

## 操作流程

### 一、生成 TOTP 验证码

```bash
python3 sc_proto.py totp <GITHUB_TOTP>
```

### 二、创建临时邮箱

```bash
python3 sc_proto.py mail <YYDS_API_KEY> my-prefix
```

### 三、完整流程 (注册 + OAuth + 项目 + Ticket)

```python
from sc_proto import full_pipeline

result = full_pipeline(
    yyds_api_key="<YYDS_API_KEY>",
    github_user="<GITHUB_SERVICE_USER>",
    github_pass="<GITHUB_PASS>",
    github_totp_secret="<GITHUB_TOTP>",
    repo_id="<REPO_ID>",  # <GITHUB_SERVICE_USER>/SRapi
    ticket_description="重构文档系统，重写 AGENTS.md",
    model="claude_code_claude_opus_4_8",
)
print(result)
```

### 四、用已有账号创建 Ticket

```python
from sc_proto import login_and_create_ticket

result = login_and_create_ticket(
    email="<EMAIL>",
    password="<PASSWORD>",
    project_id="brQP7zgfjzC6",  # SRapi 项目
    ticket_description="你的任务描述",
    model="claude_code_claude_opus_4_8",
)
```

### 五、分步操作

```python
from sc_proto import GitHubSession, SuperconductorSession, YYDSMail

# 1. 创建临时邮箱
mail = YYDSMail("<YYDS_API_KEY>")
email_data = mail.create("my-prefix")

# 2. 登录 GitHub
gh = GitHubSession()
gh.login("<GITHUB_SERVICE_USER>", "<GITHUB_PASS>", "<GITHUB_TOTP>")

# 3. 注册 SC
sc = SuperconductorSession()
sc.register(email_data["address"], "MyPass#2026!", "My Name")

# 4. 连接 GitHub OAuth
sc.connect_github(gh)

# 5. 创建项目绑定仓库
project_id = sc.create_project("<REPO_ID>")

# 6. 创建 Ticket
ticket_id = sc.create_ticket(project_id, "任务描述", "claude_code_claude_opus_4_8")

# 7. 查看状态
status = sc.get_ticket_status(ticket_id)
print(status)

# 8. 发送追问
sc.send_message("conversation_id", "追问内容")
```

### 六、批量注册

```bash
python3 sc_register.py 5   # 注册 5 个账号
```

### 七、查看 Ticket 状态

```bash
python3 sc_proto.py status <ticket_id>
```

---

## API 端点参考

### Superconductor

| 操作 | 方法 | 路径 |
|------|------|------|
| 注册 | POST | `/sign_up` |
| 登录 | POST | `/log_in` |
| GitHub OAuth | POST | `/auth/github` → 302 → GitHub → 302 → `/auth/github/callback` |
| 创建项目 | POST | `/workspaces/{wid}/projects` |
| 创建 Ticket | POST | `/projects/{pid}/tickets` |
| 发送消息 | POST | `/conversations/{cid}/messages` |
| 标记已读 | PATCH | `/conversations/{cid}/mark_seen` |
| 查看 Ticket | GET | `/tickets/{tid}` |
| 查看实现 | GET | `/tickets/{tid}/implementations/{iid}` |

### YYDS Mail

| 操作 | 方法 | 路径 |
|------|------|------|
| 创建邮箱 | POST | `/v1/accounts` |
| 查看消息 | GET | `/v1/messages?address=xxx` |
| 消息详情 | GET | `/v1/messages/{id}?address=xxx` |

### GitHub

| 操作 | 方法 | 路径 |
|------|------|------|
| 获取登录页 | GET | `/login` |
| 登录 | POST | `/session` |
| 2FA | POST | `/sessions/two-factor` |
| OAuth 授权 | GET | `/login/oauth/authorize?client_id=...` |

---

## 已注册账号

| 邮箱 | 密码 | Workspace | GitHub | 用途 |
|------|------|-----------|--------|------|
| `<EMAIL>` | `<PASSWORD>` | `6tGk8QBF9gTT` | ✅ <GITHUB_SERVICE_USER> | 主账号 (已绑 SRapi) |
| `<EMAIL>` | `<PASSWORD>` | `8NfdDDpnnkjp` | ❌ (冲突) | — |
| `<EMAIL>` | `<PASSWORD>` | `kGDQrFWTLr7q` | ✅ | crack 项目 |
| `<EMAIL>` | `<PASSWORD>` | `8gHnHRwtQKth` | ❌ (冲突) | — |

## 项目 ID

| 项目 | ID | 仓库 |
|------|-----|------|
| crack | `J8g89HQpCRDr` | <GITHUB_SERVICE_USER>/crack |
| SRapi | `brQP7zgfjzC6` | <GITHUB_SERVICE_USER>/SRapi |

## 快速换号 (额度耗尽时)

当一个账号的 $20 额度用完后，执行换号流程 (~15 秒):

```python
from sc_proto import rotate_account

new_sc, info = rotate_account(
    old_email="<EMAIL>",       # 旧账号
    old_password="<PASSWORD>",
    yyds_api_key="<YYDS_API_KEY>",
    github_user="<GITHUB_SERVICE_USER>",
    github_pass="<GITHUB_PASS>",
    github_totp_secret="<GITHUB_TOTP>",
    repo_id="<REPO_ID>",                  # <GITHUB_SERVICE_USER>/SRapi
)

# new_sc 已登录，直接创建 ticket
new_sc.create_ticket(info["project_id"], "任务描述", "claude_code_claude_opus_4_8")
```

**流程 (~15s):**
1. 登录旧账号 (3s 反机器人)
2. `DELETE /identities/{id}` 解绑 GitHub
3. 登录 GitHub + TOTP
4. 创建临时邮箱
5. 注册新 SC 账号 (3s 反机器人)
6. `POST /auth/github` OAuth 绑定
7. `POST /workspaces/{wid}/projects` 绑定仓库
8. 新账号就绪，$20 额度可用

**也可以单独解绑:**
```python
from sc_proto import SuperconductorSession

sc = SuperconductorSession()
sc.login("old@email.com", "password")
sc.disconnect_github()  # DELETE /identities/{identity_id}
```

---

## 重要限制

1. **一个 GitHub 账号只能绑一个 SC 账号** — 批量注册需要多个 GitHub 账号
2. **新项目需等待 Environment Setup** — 创建项目后约 30s Agent 才能工作
3. **注册/登录需等待 3 秒** — 服务端反机器人时间检测
4. **Agent 回复通过 WebSocket 推送** — 无法通过 HTTP 轮询实时获取，需刷新页面
5. **每个账号 $20 免费额度** — Opus 4.8 约 $0.40/ticket，可用约 50 次
