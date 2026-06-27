export type Locale = "en" | "zh";

const dict = {
  // Header
  "agent_platform": { en: "Agent Platform", zh: "智能体平台" },
  "online": { en: "Online", zh: "在线" },
  "admin": { en: "Admin", zh: "管理" },
  "back_to_app": { en: "← Back to app", zh: "← 返回应用" },

  // Tabs
  "tasks": { en: "Tasks", zh: "任务" },
  "github": { en: "GitHub", zh: "GitHub" },

  // Model selector
  "model": { en: "Model", zh: "模型" },
  "most_capable": { en: "Most capable", zh: "最强" },
  "fast": { en: "Fast", zh: "快速" },
  "economy": { en: "Economy", zh: "经济" },

  // Task panel
  "ready_to_work": { en: "Ready to work", zh: "准备就绪" },
  "ready_desc": { en: "Describe what you want the AI agent to do.\nIt reads your code, makes changes, and opens a PR.", zh: "描述你希望 AI 代理做什么。\n它会阅读代码、修改并提交 PR。" },
  "input_placeholder": { en: "Describe what you want the AI to build, fix, or improve…", zh: "描述你想让 AI 构建、修复或改进的内容…" },
  "run_agent": { en: "Run Agent", zh: "运行" },
  "to_send": { en: "to send", zh: "发送" },
  "credits_auto_rotate": { en: "Credits auto-rotate", zh: "额度自动轮转" },

  // Task list
  "no_tasks": { en: "No tasks yet.\nDescribe what you need and hit Run Agent.", zh: "暂无任务。\n描述需求后点击运行。" },
  "running": { en: "Running", zh: "运行中" },
  "waiting": { en: "Waiting", zh: "等待中" },
  "completed": { en: "Done", zh: "完成" },
  "failed": { en: "Failed", zh: "失败" },
  "starting": { en: "Starting", zh: "启动中" },

  // GitHub panel
  "connect_github": { en: "Connect GitHub", zh: "连接 GitHub" },
  "connect_github_desc": { en: "Connect GitHub to let\nagents work on your repos", zh: "连接 GitHub 让\nAI 代理操作你的仓库" },
  "only_repo_access": { en: "Only requests repo access", zh: "仅请求仓库权限" },
  "connected": { en: "Connected", zh: "已连接" },
  "active_repo": { en: "Active repo", zh: "当前仓库" },
  "select_repo": { en: "Select a repository", zh: "选择仓库" },
  "switch_repo": { en: "Switch repo", zh: "切换仓库" },
  "disconnect_github": { en: "Disconnect GitHub", zh: "断开 GitHub" },
  "private": { en: "private", zh: "私有" },

  // Admin
  "overview": { en: "Overview", zh: "概览" },
  "accounts": { en: "Accounts", zh: "账号" },
  "users": { en: "Users", zh: "用户" },
  "config": { en: "Config", zh: "配置" },
  "accounts_ready": { en: "Accounts Ready", zh: "可用账号" },
  "credits_available": { en: "Credits Available", zh: "可用额度" },
  "total_tasks": { en: "Total Tasks", zh: "总任务数" },
  "total_users": { en: "Total Users", zh: "总用户数" },
  "active_now": { en: "Active Now", zh: "当前活跃" },
  "accounts_in_use": { en: "accounts in use", zh: "账号使用中" },
  "uptime": { en: "Uptime", zh: "运行时间" },
  "goroutines": { en: "Goroutines", zh: "协程数" },
  "memory": { en: "Memory", zh: "内存" },
  "no_accounts": { en: "No accounts in pool", zh: "账号池为空" },
  "no_tasks_yet": { en: "No tasks yet", zh: "暂无任务" },
  "no_users_yet": { en: "No users yet", zh: "暂无用户" },
  "current_config": { en: "Current Configuration", zh: "当前配置" },
  "update_config": { en: "Update Configuration", zh: "更新配置" },
  "leave_empty": { en: "Leave fields empty to keep current values.", zh: "留空保持当前值。" },
  "save_config": { en: "Save Configuration", zh: "保存配置" },
  "saving": { en: "Saving...", zh: "保存中..." },
  "saved": { en: "✓ Saved", zh: "✓ 已保存" },
  "email": { en: "Email", zh: "邮箱" },
  "status": { en: "Status", zh: "状态" },
  "credits": { en: "Credits", zh: "额度" },
  "workspace": { en: "Workspace", zh: "工作区" },
  "task": { en: "Task", zh: "任务" },
  "user": { en: "User", zh: "用户" },
  "cost": { en: "Cost", zh: "花费" },
  "created": { en: "Created", zh: "创建时间" },
  "repository": { en: "Repository", zh: "仓库" },
  "joined": { en: "Joined", zh: "加入时间" },
  "github_service_user": { en: "GitHub Service User", zh: "GitHub 服务账号" },
  "github_password": { en: "GitHub Password", zh: "GitHub 密码" },
  "github_totp_secret": { en: "GitHub TOTP Secret", zh: "GitHub TOTP 密钥" },
  "yyds_api_key": { en: "YYDS API Key", zh: "YYDS API 密钥" },
  "default_repo_id": { en: "Default Repo ID", zh: "默认仓库 ID" },
  "exhausted": { en: "exhausted", zh: "已耗尽" },
  "total": { en: "total", zh: "总计" },
} as const;

export type Key = keyof typeof dict;

export function t(key: Key, locale: Locale): string {
  return dict[key]?.[locale] ?? key;
}

export function createT(locale: Locale) {
  return (key: Key) => t(key, locale);
}
