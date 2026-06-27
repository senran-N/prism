"use client";

import { useLocale, LocaleSwitch } from "@/app/ui/locale-context";

export default function TermsPage() {
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
        <h1 className="text-[28px] font-semibold text-[#0a2540] mb-2">{isZh ? "服务条款" : "Terms of Service"}</h1>
        <p className="text-[13px] text-[#8792a2] mb-8">{isZh ? "最后更新：2026 年 6 月 28 日" : "Last updated: June 28, 2026"}</p>

        <div className="prose prose-sm max-w-none text-[#425466] leading-relaxed space-y-6">
          {isZh ? (
            <>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">1. 服务说明</h2>
                <p>Prism（以下简称"本平台"）是一个 AI 代码智能体平台，允许用户通过选择 AI 模型，自动完成代码编写、修复、重构等任务。本平台通过第三方 AI 服务提供代码生成能力。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">2. 账号与认证</h2>
                <p>用户通过 LinuxDo 账号登录本平台。您对通过您的账号进行的所有活动负责。如发现未授权使用，请立即联系我们。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">3. 使用规范</h2>
                <p>用户不得利用本平台：</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>生成违反法律法规的内容</li>
                  <li>滥用平台资源（包括但不限于恶意消耗计算额度）</li>
                  <li>尝试攻击、逆向工程或干扰平台运行</li>
                  <li>侵犯他人知识产权</li>
                  <li>将平台用于任何非法目的</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">4. 知识产权</h2>
                <p>通过本平台生成的代码归用户所有。本平台的软件、设计、商标等知识产权归 Prism 运营方所有。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">5. 免责声明</h2>
                <p>本平台按"现状"提供服务，不对 AI 生成的代码的准确性、安全性或适用性做任何保证。用户应在使用生成代码前自行审查。本平台不对因使用生成代码造成的任何损失承担责任。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">6. 服务中断与终止</h2>
                <p>我们保留在不另行通知的情况下暂停或终止服务的权利。违反本条款的用户账号可能被暂停或永久封禁。</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">7. 条款变更</h2>
                <p>我们可能不定期修改本条款。修改后继续使用本平台即表示您接受修改后的条款。</p>
              </section>
            </>
          ) : (
            <>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">1. Service Description</h2>
                <p>Prism ("the Platform") is an AI code agent platform that allows users to select AI models to automatically write, fix, and refactor code. The Platform provides code generation capabilities through third-party AI services.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">2. Accounts & Authentication</h2>
                <p>Users log in via their LinuxDo account. You are responsible for all activities conducted through your account. Contact us immediately if you discover unauthorized use.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">3. Acceptable Use</h2>
                <p>You may not use the Platform to:</p>
                <ul className="list-disc pl-5 space-y-1">
                  <li>Generate content that violates any laws or regulations</li>
                  <li>Abuse platform resources (including excessive credit consumption)</li>
                  <li>Attack, reverse-engineer, or interfere with platform operations</li>
                  <li>Infringe on others' intellectual property rights</li>
                  <li>Use the Platform for any illegal purpose</li>
                </ul>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">4. Intellectual Property</h2>
                <p>Code generated through the Platform belongs to the user. The Platform's software, design, and trademarks are owned by Prism's operators.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">5. Disclaimer</h2>
                <p>The Platform is provided "as is." We make no warranties regarding the accuracy, security, or suitability of AI-generated code. Users should review generated code before use. We are not liable for any damages arising from the use of generated code.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">6. Service Interruption & Termination</h2>
                <p>We reserve the right to suspend or terminate service without notice. Accounts violating these terms may be suspended or permanently banned.</p>
              </section>
              <section>
                <h2 className="text-[18px] font-semibold text-[#0a2540] mt-8 mb-3">7. Changes to Terms</h2>
                <p>We may modify these terms from time to time. Continued use of the Platform after modifications constitutes acceptance of the updated terms.</p>
              </section>
            </>
          )}
        </div>
      </article>
    </div>
  );
}
