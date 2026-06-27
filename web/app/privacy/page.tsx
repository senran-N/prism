"use client";

import { useLocale, LocaleSwitch } from "@/app/ui/locale-context";

export default function PrivacyPage() {
  const { locale } = useLocale();
  const isZh = locale === "zh";

  return (
    <div className="min-h-screen bg-[#f6f9fc]">
      <header className="bg-white border-b border-[#e3e8ee] px-6 py-3 flex items-center justify-between" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
        <a href="/" className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-md bg-[#635bff] flex items-center justify-center">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2" />
            </svg>
          </div>
          <span className="text-[15px] font-semibold text-[#0a2540]">Prism</span>
        </a>
        <LocaleSwitch />
      </header>

      <article className="max-w-2xl mx-auto px-6 py-12">
        <h1 className="text-[28px] font-semibold text-[#0a2540] mb-2">{isZh ? "隐私政策" : "Privacy Policy"}</h1>
        <p className="text-[13px] text-[#8792a2] mb-8">{isZh ? "最后更新：2026 年 6 月 28 日" : "Last updated: June 28, 2026"}</p>

        <div className="prose prose-sm max-w-none text-[#425466] leading-relaxed space-y-6">
          {isZh ? (
            <>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">1. 信息收集</h2>
                <p>我们通过 LinuxDo OAuth 收集以下信息：</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li><strong>LinuxDo 账号信息：</strong>用户 ID、用户名、昵称、头像、信任等级</li>
                  <li><strong>GitHub 仓库信息：</strong>仓库名称和访问权限（仅在用户主动连接时）</li>
                  <li><strong>使用数据：</strong>任务描述、模型选择、任务状态和花费</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">2. 信息使用</h2>
                <p>我们使用收集的信息用于：</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>提供和维护平台服务</li>
                  <li>用户身份验证和会话管理</li>
                  <li>任务执行和状态追踪</li>
                  <li>服务改进和问题排查</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">3. 信息存储</h2>
                <p>用户数据存储在安全的服务器上。我们使用加密技术保护敏感信息。会话数据通过签名 Cookie 管理，不可篡改。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">4. 信息共享</h2>
                <p>我们不会出售或出租您的个人信息。仅在以下情况下可能共享：</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>执行您的任务时与第三方 AI 服务交互（仅传递任务描述，不传递个人信息）</li>
                  <li>法律要求或执法机构合法请求</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">5. 代码数据</h2>
                <p>您提交的任务描述会被发送给 AI 模型处理。AI 对您仓库的操作通过您授权的 GitHub 权限进行。我们不会永久存储您的源代码。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">6. 用户权利</h2>
                <p>您有权：</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>查看我们持有的您的数据</li>
                  <li>要求删除您的账号和相关数据</li>
                  <li>断开 GitHub 连接以撤销仓库访问权限</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">7. 联系方式</h2>
                <p>如有隐私相关问题，请通过 LinuxDo 社区联系我们。</p>
              </section>
            </>
          ) : (
            <>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">1. Information Collection</h2>
                <p>We collect the following information through LinuxDo OAuth:</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li><strong>LinuxDo account info:</strong> User ID, username, display name, avatar, trust level</li>
                  <li><strong>GitHub repository info:</strong> Repository names and access permissions (only when you connect)</li>
                  <li><strong>Usage data:</strong> Task descriptions, model selections, task status, and costs</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">2. Information Use</h2>
                <p>We use collected information to:</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>Provide and maintain platform services</li>
                  <li>User authentication and session management</li>
                  <li>Task execution and status tracking</li>
                  <li>Service improvement and troubleshooting</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">3. Information Storage</h2>
                <p>User data is stored on secure servers. We use encryption to protect sensitive information. Session data is managed through signed cookies that cannot be tampered with.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">4. Information Sharing</h2>
                <p>We do not sell or rent your personal information. We may share data only when:</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>Interacting with third-party AI services to execute your tasks (only task descriptions, not personal info)</li>
                  <li>Required by law or lawful requests from law enforcement</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">5. Code Data</h2>
                <p>Task descriptions you submit are sent to AI models for processing. AI operates on your repositories through GitHub permissions you authorized. We do not permanently store your source code.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">6. User Rights</h2>
                <p>You have the right to:</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>View the data we hold about you</li>
                  <li>Request deletion of your account and associated data</li>
                  <li>Disconnect GitHub to revoke repository access</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">7. Contact</h2>
                <p>For privacy-related inquiries, please contact us through the LinuxDo community.</p>
              </section>
            </>
          )}
        </div>
      </article>
    </div>
  );
}
