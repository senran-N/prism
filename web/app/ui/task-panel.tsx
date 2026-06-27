"use client";

import { useState, useRef, useEffect } from "react";
import type { Task } from "./dashboard";

interface Props {
  model: string;
  activeTask: string | null;
  onTaskCreated: (task: Task) => void;
}

const MODEL_NAMES: Record<string, string> = {
  claude_code_claude_opus_4_8: "Claude Code · Opus 4.8",
  claude_code_claude_opus_4_7: "Claude Code · Opus 4.7",
  claude_code_claude_opus_4_6: "Claude Code · Opus 4.6",
  claude_code_claude_sonnet_4_6: "Claude Code · Sonnet 4.6",
  codex_gpt_5_5_high: "Codex · GPT-5.5 High",
  codex_gpt_5_5_medium: "Codex · GPT-5.5 Medium",
  codex_gpt_5_5_xhigh: "Codex · GPT-5.5 Xhigh",
  opencode_gemini_3_1_pro: "OpenCode · Gemini 3.1 Pro",
  opencode_gpt_5_5: "OpenCode · GPT-5.5",
  opencode_kimi_k2_6: "OpenCode · Kimi K2.6",
  pi_deepseek_v4_pro: "Pi · DeepSeek V4 Pro",
  pi_deepseek_v4_flash: "Pi · DeepSeek V4 Flash",
};

export default function TaskPanel({ model, activeTask, onTaskCreated }: Props) {
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [viewUrl, setViewUrl] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (!activeTask) { setViewUrl(null); setStatus(null); }
  }, [activeTask]);

  useEffect(() => {
    if (!activeTask) return;
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/tasks/${activeTask}/status`);
        if (res.ok) {
          const data = await res.json();
          setStatus(data.status);
          if (data.status === "Completed" || data.status === "Failed") clearInterval(interval);
        }
      } catch {}
    }, 10000);
    return () => clearInterval(interval);
  }, [activeTask]);

  async function handleSubmit() {
    if (!input.trim()) return;
    setLoading(true);
    try {
      const res = await fetch("/api/tasks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ description: input.trim(), model }),
      });
      if (!res.ok) {
        const err = await res.json();
        alert(err.error || "Failed to create task");
        return;
      }
      const data = await res.json();
      onTaskCreated({
        id: data.task_id,
        description: input.trim(),
        model,
        modelName: MODEL_NAMES[model] || model,
        status: "created",
        createdAt: new Date().toISOString(),
        viewUrl: data.view_url,
      });
      setViewUrl(data.view_url);
      setStatus("Running");
      setInput("");
    } finally {
      setLoading(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) { e.preventDefault(); handleSubmit(); }
  }

  useEffect(() => {
    const ta = textareaRef.current;
    if (ta) { ta.style.height = "auto"; ta.style.height = Math.min(ta.scrollHeight, 200) + "px"; }
  }, [input]);

  return (
    <div className="flex-1 flex flex-col">
      {/* Agent view */}
      <div className="flex-1 relative">
        {viewUrl ? (
          <iframe src={viewUrl} className="w-full h-full border-0" sandbox="allow-same-origin allow-scripts allow-forms" />
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center max-w-sm">
              <div className="w-14 h-14 rounded-2xl bg-white border border-[#e3e8ee] flex items-center justify-center mx-auto mb-4" style={{ boxShadow: "0 2px 8px rgba(99,91,255,0.08)" }}>
                <svg className="w-7 h-7 text-[#635bff]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456z" />
                </svg>
              </div>
              <h2 className="text-[17px] font-semibold text-[#0a2540] mb-1.5">Ready to work</h2>
              <p className="text-[13px] text-[#697386] leading-5">
                Describe what you want the AI agent to do.<br />
                It reads your code, makes changes, and opens a PR.
              </p>
            </div>
          </div>
        )}

        {status && (
          <div className="absolute top-3 right-3 flex items-center gap-2 bg-white border border-[#e3e8ee] rounded-lg px-3 py-1.5 text-[13px] text-[#0a2540]" style={{ boxShadow: "var(--shadow-sm)" }}>
            <span className={`w-1.5 h-1.5 rounded-full ${
              status === "Running" ? "bg-[#635bff] animate-pulse" :
              status === "Waiting" ? "bg-[#f5a623]" :
              status === "Completed" ? "bg-[#0caf60]" :
              status === "Failed" ? "bg-[#df1b41]" : "bg-[#8792a2]"
            }`} />
            {status}
          </div>
        )}
      </div>

      {/* Input */}
      <div className="bg-white border-t border-[#e3e8ee] px-5 py-4">
        <div className="flex gap-3 items-end">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Describe what you want the AI to build, fix, or improve…"
            rows={1}
            className="flex-1 bg-[#f6f9fc] border border-[#e3e8ee] rounded-lg px-4 py-2.5 text-[14px] text-[#0a2540] resize-none focus:outline-none focus:ring-2 focus:ring-[#635bff]/20 focus:border-[#635bff] placeholder:text-[#8792a2] transition-all"
          />
          <button
            onClick={handleSubmit}
            disabled={loading || !input.trim()}
            className="bg-[#635bff] hover:bg-[#7a73ff] text-white rounded-lg px-5 py-2.5 text-[14px] font-medium transition-all disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2 shrink-0"
            style={{ boxShadow: "0 1px 3px rgba(99,91,255,0.3)" }}
          >
            {loading ? (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" />
              </svg>
            )}
            Run Agent
          </button>
        </div>
        <div className="flex items-center justify-between mt-2">
          <p className="text-[12px] text-[#8792a2]">
            <span className="text-[#635bff] font-medium">{MODEL_NAMES[model] || model}</span>
            {" · "}
            <kbd className="text-[11px] bg-[#f6f9fc] border border-[#e3e8ee] rounded px-1 py-0.5 font-mono">⌘↵</kbd> to send
          </p>
          <p className="text-[12px] text-[#8792a2]">Credits auto-rotate</p>
        </div>
      </div>
    </div>
  );
}
