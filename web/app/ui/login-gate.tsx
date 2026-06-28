"use client";

import { useState, useEffect } from "react";
import { useLocale } from "./locale-context";

interface UserInfo {
  logged_in: boolean;
  id?: number;
  github_login?: string;
  linuxdo_username?: string;
  linuxdo_name?: string;
  avatar_url?: string;
  trust_level?: number;
  selected_repo?: string;
  is_banned?: boolean;
  ban_reason?: string;
  is_admin?: boolean;
}

export function useUser() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/me").then(r => r.json()).then(setUser).catch(() => setUser({ logged_in: false })).finally(() => setLoading(false));
  }, []);

  return { user, loading, refresh: () => fetch("/api/me").then(r => r.json()).then(setUser) };
}

export function LoginGate({ children }: { children: React.ReactNode }) {
  const { user, loading } = useUser();
  const { t, locale } = useLocale();

  if (loading) return (
    <div className="min-h-screen bg-[#f6f9fc] flex items-center justify-center">
      <span className="w-6 h-6 border-2 border-[#e3e8ee] border-t-[#635bff] rounded-full animate-spin" />
    </div>
  );

  if (!user?.logged_in) {
    // Show URL errors from OAuth callbacks
    const urlError = typeof window !== "undefined" ? new URLSearchParams(window.location.search).get("error") : null;
    return <LoginPage urlError={urlError} />;
  }

  if (user?.is_banned) return (
    <div className="min-h-screen bg-[#f6f9fc] flex items-center justify-center">
      <div className="bg-white rounded-2xl border border-[#e3e8ee] p-8 w-full max-w-sm text-center" style={{ boxShadow: "0 4px 24px rgba(0,0,0,0.06)" }}>
        <div className="w-12 h-12 rounded-xl bg-[#fde8ed] flex items-center justify-center mx-auto mb-4">
          <svg className="w-6 h-6 text-[#df1b41]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
          </svg>
        </div>
        <h2 className="text-[18px] font-semibold text-[#0a2540] mb-2">{locale === "zh" ? "账号已被封禁" : "Account Suspended"}</h2>
        <p className="text-[13px] text-[#697386] mb-2">{locale === "zh" ? "您的账号已被管理员封禁。" : "Your account has been suspended by an administrator."}</p>
        {user.ban_reason && <p className="text-[13px] text-[#df1b41] bg-[#fde8ed] rounded-lg px-3 py-2">{user.ban_reason}</p>}
      </div>
    </div>
  );

  return <>{children}</>;
}

function LoginPage({ urlError }: { urlError?: string | null }) {
  const { locale } = useLocale();
  const isZh = locale === "zh";

  const errorMessages: Record<string, string> = {
    state_expired: isZh ? "登录超时，请重试" : "Login expired, please try again",
    token_exchange: isZh ? "认证失败，请重试" : "Authentication failed, please retry",
    user_fetch: isZh ? "获取用户信息失败" : "Failed to get user info",
    github_link_failed: isZh ? "GitHub 关联失败，请重试" : "GitHub linking failed, please retry",
    github_save_failed: isZh ? "GitHub 保存失败" : "GitHub save failed",
    github_missing_code: isZh ? "GitHub 授权失败" : "GitHub authorization failed",
    github_token_failed: isZh ? "GitHub 认证失败，请重试" : "GitHub auth failed, please retry",
    github_user_failed: isZh ? "获取 GitHub 用户信息失败" : "Failed to get GitHub user info",
  };

  return (
    <div className="min-h-screen bg-[#f6f9fc] flex items-center justify-center">
      <div className="bg-white rounded-2xl border border-[#e3e8ee] p-8 w-full max-w-sm text-center" style={{ boxShadow: "0 4px 24px rgba(0,0,0,0.06)" }}>
        {/* Logo */}
        <div className="w-12 h-12 rounded-xl bg-[#635bff] flex items-center justify-center mx-auto mb-4">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2" />
            <line x1="12" y1="22" x2="12" y2="15.5" />
            <polyline points="22 8.5 12 15.5 2 8.5" />
          </svg>
        </div>

        <h1 className="text-[22px] font-semibold text-[#0a2540] mb-1">Prism</h1>
        <p className="text-[14px] text-[#697386] mb-6">
          {isZh ? "AI 智能体平台 — 连接仓库，选择模型，发布代码" : "AI Agent Platform — connect repos, pick models, ship code"}
        </p>

        {urlError && (
          <div className="bg-[#fde8ed] border border-[#f5c6cb] rounded-lg px-3 py-2 text-[13px] text-[#df1b41] mb-3">
            {errorMessages[urlError] || urlError}
          </div>
        )}

        {/* LinuxDo Login */}
        <a href="/api/linuxdo/login"
          className="flex items-center justify-center gap-2.5 w-full bg-[#f39c12] hover:bg-[#e67e22] text-white rounded-lg px-4 py-3 text-[14px] font-medium transition-colors"
          style={{ boxShadow: "0 1px 3px rgba(243,156,18,0.3)" }}>
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/>
          </svg>
          {isZh ? "使用 LinuxDo 登录" : "Sign in with LinuxDo"}
        </a>

        <p className="text-[12px] text-[#8792a2] mt-5">
          {isZh ? "登录即表示同意" : "By signing in you agree to our "}
          <a href="/terms" className="text-[#635bff] hover:underline">{isZh ? "服务条款" : "Terms"}</a>
          {isZh ? "和" : " and "}
          <a href="/privacy" className="text-[#635bff] hover:underline">{isZh ? "隐私政策" : "Privacy Policy"}</a>
        </p>
      </div>
    </div>
  );
}
