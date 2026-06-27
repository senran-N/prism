"use client";

import Dashboard from "@/app/ui/dashboard";
import { useLocale, LocaleSwitch } from "@/app/ui/locale-context";

export default function Home() {
  const { t } = useLocale();

  return (
    <div className="flex flex-col h-full min-h-screen">
      <header className="bg-white border-b border-[#e3e8ee] px-6 py-3 flex items-center justify-between sticky top-0 z-50" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
        <div className="flex items-center gap-3">
          <div className="w-7 h-7 rounded-md bg-[#635bff] flex items-center justify-center">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2" />
              <line x1="12" y1="22" x2="12" y2="15.5" />
              <polyline points="22 8.5 12 15.5 2 8.5" />
            </svg>
          </div>
          <span className="text-[15px] font-semibold text-[#0a2540] tracking-tight">Prism</span>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 text-[13px] text-[#697386]">
            <span className="w-1.5 h-1.5 rounded-full bg-[#0caf60]" />
            {t("online")}
          </div>
          <LocaleSwitch />
          <a href="/admin" className="text-[12px] text-[#8792a2] hover:text-[#635bff] transition-colors">{t("admin")}</a>
          <div className="w-7 h-7 rounded-full bg-[#e3e8ee] flex items-center justify-center text-xs font-medium text-[#697386]">U</div>
        </div>
      </header>
      <main className="flex-1">
        <Dashboard />
      </main>
    </div>
  );
}
