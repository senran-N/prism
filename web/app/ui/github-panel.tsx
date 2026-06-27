"use client";

import { useState, useEffect } from "react";
import { useLocale } from "./locale-context";

interface Repo { full_name: string; private: boolean; }
interface GitHubState { connected: boolean; user: string; repos: Repo[]; selectedRepo: string; }

export default function GitHubPanel() {
  const { t } = useLocale();
  const [state, setState] = useState<GitHubState>({ connected: false, user: "", repos: [], selectedRepo: "" });
  const [loading, setLoading] = useState(true);
  const [selecting, setSelecting] = useState(false);

  useEffect(() => { fetch("/api/github/status").then(r => r.json()).then(setState).finally(() => setLoading(false)); }, []);

  async function handleSelectRepo(repo: string) {
    setSelecting(true);
    try {
      const res = await fetch("/api/github/select-repo", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ repo }) });
      if (res.ok) setState(prev => ({ ...prev, selectedRepo: repo }));
      else { const err = await res.json(); alert(err.error); }
    } finally { setSelecting(false); }
  }

  if (loading) return <div className="flex-1 flex items-center justify-center"><span className="w-5 h-5 border-2 border-[#e3e8ee] border-t-[#635bff] rounded-full animate-spin" /></div>;

  if (!state.connected) return (
    <div className="flex-1 flex flex-col items-center justify-center p-6">
      <div className="w-10 h-10 rounded-xl bg-[#f6f9fc] border border-[#e3e8ee] flex items-center justify-center mb-3">
        <svg className="w-5 h-5 text-[#0a2540]" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
      </div>
      <p className="text-[13px] text-[#697386] text-center mb-4 leading-5 whitespace-pre-line">{t("connect_github_desc")}</p>
      <a href="/api/github/login" className="flex items-center gap-2 bg-[#0a2540] hover:bg-[#1a3550] text-white rounded-lg px-4 py-2 text-[13px] font-medium transition-colors" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.12)" }}>
        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
        {t("connect_github")}
      </a>
      <p className="text-[11px] text-[#8792a2] text-center mt-3">{t("only_repo_access")}</p>
    </div>
  );

  return (
    <div className="flex-1 flex flex-col p-4 gap-3 overflow-y-auto">
      <div className="flex items-center gap-2 pb-3 border-b border-[#e3e8ee]">
        <svg className="w-4 h-4 text-[#0a2540]" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
        <span className="text-[13px] font-medium text-[#0a2540]">{state.user}</span>
        <span className="text-[11px] font-medium text-[#0caf60] bg-[#e6f9f0] rounded px-1.5 py-0.5">{t("connected")}</span>
      </div>
      {state.selectedRepo && (
        <div className="bg-[#f0efff] border border-[#d9d6ff] rounded-lg px-3 py-2.5">
          <div className="text-[11px] text-[#635bff] font-medium uppercase tracking-wider">{t("active_repo")}</div>
          <div className="text-[13px] font-medium text-[#0a2540] mt-0.5">{state.selectedRepo}</div>
        </div>
      )}
      <div className="text-[11px] font-medium text-[#8792a2] uppercase tracking-wider mt-1">{state.selectedRepo ? t("switch_repo") : t("select_repo")}</div>
      <div className="flex flex-col gap-0.5">
        {state.repos.map(repo => (
          <button key={repo.full_name} onClick={() => handleSelectRepo(repo.full_name)} disabled={selecting || repo.full_name === state.selectedRepo}
            className={`text-left px-3 py-2 rounded-md text-[13px] transition-colors ${repo.full_name === state.selectedRepo ? "bg-[#f0efff] text-[#635bff] font-medium" : "hover:bg-[#f6f9fc] text-[#0a2540]"} disabled:opacity-50`}>
            <div className="flex items-center gap-1.5">
              <span className="truncate">{repo.full_name}</span>
              {repo.private && <span className="text-[10px] text-[#8792a2] border border-[#e3e8ee] rounded px-1 py-px shrink-0">{t("private")}</span>}
            </div>
          </button>
        ))}
      </div>
      <button onClick={async () => { await fetch("/api/github/disconnect", { method: "POST" }); setState({ connected: false, user: "", repos: [], selectedRepo: "" }); }}
        className="text-[12px] text-[#df1b41] hover:text-[#ff4d6a] transition-colors mt-auto pt-3 border-t border-[#e3e8ee]">{t("disconnect_github")}</button>
    </div>
  );
}
