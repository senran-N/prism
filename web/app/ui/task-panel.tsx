"use client";

import { useState, useRef, useEffect, useCallback } from "react";
import { useLocale } from "./locale-context";
import type { Task } from "./dashboard";

function BalanceBadge() {
  const [bal, setBal] = useState<{balance:number;can_rotate:boolean}|null>(null);
  const { locale } = useLocale();
  useEffect(() => { fetch("/api/balance").then(r=>r.ok?r.json():null).then(setBal).catch(()=>{}); }, []);
  if (!bal) return <p className="text-[12px] text-[#8792a2]">{locale==="zh"?"额度自动轮转":"Credits auto-rotate"}</p>;
  return (
    <p className="text-[12px]">
      <span className={bal.can_rotate?"text-[#0caf60]":"text-[#f5a623]"}>●</span>{" "}
      <span className="text-[#8792a2]">{bal.balance.toFixed(0)} {locale==="zh"?"积分":"credits"}</span>
    </p>
  );
}

interface AgentMsg { id: string; author: string; content: string; }

function AgentChat({ ticketId, status, locale }: { ticketId: string; status: string | null; locale: string }) {
  const [messages, setMessages] = useState<AgentMsg[]>([]);
  const [followUp, setFollowUp] = useState("");
  const [sending, setSending] = useState(false);
  const chatEndRef = useRef<HTMLDivElement>(null);
  const isZh = locale === "zh";

  useEffect(() => {
    fetchMessages();
    const interval = setInterval(fetchMessages, 8000);
    return () => clearInterval(interval);
  }, [ticketId]);

  useEffect(() => { chatEndRef.current?.scrollIntoView({ behavior: "smooth" }); }, [messages]);

  async function fetchMessages() {
    try {
      const res = await fetch(`/api/tasks/${ticketId}/messages`);
      if (res.ok) {
        const data = await res.json();
        if (data.messages) setMessages(data.messages);
      }
    } catch {}
  }

  async function handleSend() {
    if (!followUp.trim() || sending) return;
    setSending(true);
    try {
      const res = await fetch(`/api/tasks/${ticketId}/message`, {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: followUp.trim() }),
      });
      if (res.ok) {
        setFollowUp("");
        setTimeout(fetchMessages, 2000);
      }
    } finally { setSending(false); }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Status bar */}
      <div className={`px-4 py-2 text-[13px] font-medium flex items-center gap-2 ${
        status === "Failed" ? "bg-[#fde8ed] text-[#df1b41]" :
        status === "Completed" ? "bg-[#e6f9f0] text-[#0caf60]" :
        status === "Waiting" ? "bg-[#fef6e6] text-[#f5a623]" :
        "bg-[#e8e6ff] text-[#635bff]"
      }`}>
        <span className={`w-2 h-2 rounded-full ${
          status === "Failed" ? "bg-[#df1b41]" :
          status === "Completed" ? "bg-[#0caf60]" :
          status === "Running" ? "bg-[#635bff] animate-pulse" :
          "bg-[#f5a623]"
        }`} />
        {status === "Failed" ? (isZh ? "任务失败" : "Failed") :
         status === "Completed" ? (isZh ? "已完成" : "Completed") :
         status === "Waiting" ? (isZh ? "等待回复" : "Waiting") :
         (isZh ? "运行中" : "Running")}
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-4">
        {messages.length === 0 && (
          <div className="flex items-center justify-center h-full text-[13px] text-[#8792a2]">
            {isZh ? "等待 Agent 响应..." : "Waiting for agent response..."}
          </div>
        )}
        {messages.map((msg) => {
          const isAgent = msg.author !== "Evolink Reg" && msg.author !== "User" && !msg.author.includes("@");
          return (
            <div key={msg.id} className={`flex ${isAgent ? "justify-start" : "justify-end"}`}>
              <div className={`max-w-[80%] rounded-xl px-4 py-2.5 text-[13px] leading-relaxed ${
                isAgent
                  ? "bg-white border border-[#e3e8ee] text-[#0a2540]"
                  : "bg-[#635bff] text-white"
              }`} style={isAgent ? { boxShadow: "0 1px 3px rgba(0,0,0,0.04)" } : {}}>
                <div className={`text-[11px] font-medium mb-1 ${isAgent ? "text-[#635bff]" : "text-white/70"}`}>
                  {msg.author}
                </div>
                <div className="whitespace-pre-wrap break-words">{msg.content}</div>
              </div>
            </div>
          );
        })}
        <div ref={chatEndRef} />
      </div>

      {/* Follow-up input */}
      {status !== "Completed" && status !== "Failed" && (
        <div className="border-t border-[#e3e8ee] px-4 py-3 flex gap-2">
          <input
            type="text" value={followUp}
            onChange={(e) => setFollowUp(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSend()}
            placeholder={isZh ? "追问 Agent..." : "Follow up with agent..."}
            className="flex-1 bg-[#f6f9fc] border border-[#e3e8ee] rounded-lg px-3 py-2 text-[13px] focus:outline-none focus:ring-2 focus:ring-[#635bff]/20 focus:border-[#635bff] placeholder:text-[#8792a2]"
          />
          <button onClick={handleSend} disabled={sending || !followUp.trim()}
            className="bg-[#635bff] hover:bg-[#7a73ff] text-white rounded-lg px-4 py-2 text-[13px] font-medium disabled:opacity-40 shrink-0">
            {sending ? "..." : (isZh ? "发送" : "Send")}
          </button>
        </div>
      )}
    </div>
  );
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

interface Props { model: string; activeTask: string | null; onTaskCreated: (task: Task) => void; onStatusChange?: (taskId: string, status: string) => void; }

export default function TaskPanel({ model, activeTask, onTaskCreated, onStatusChange }: Props) {
  const { t, locale } = useLocale();
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [viewUrl, setViewUrl] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => { if (!activeTask) { setViewUrl(null); setStatus(null); } }, [activeTask]);

  useEffect(() => {
    if (!activeTask) return;
    const interval = setInterval(async () => {
      try {
        const res = await fetch(`/api/tasks/${activeTask}/status`);
        if (res.ok) {
          const d = await res.json();
          setStatus(d.status);
          if (onStatusChange) onStatusChange(activeTask, d.status);
          if (d.status === "Completed" || d.status === "Failed") clearInterval(interval);
        }
      } catch {}
    }, 10000);
    return () => clearInterval(interval);
  }, [activeTask]);

  async function handleSubmit() {
    if (!input.trim()) return;
    setLoading(true);
    try {
      const res = await fetch("/api/tasks", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ description: input.trim(), model }) });
      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: `HTTP ${res.status}` }));
        setError(err.error || `Request failed (${res.status})`);
        return;
      }
      const data = await res.json();
      onTaskCreated({ id: data.task_id, description: input.trim(), model, modelName: MODEL_NAMES[model] || model, status: "created", createdAt: new Date().toISOString(), viewUrl: data.view_url });
      setViewUrl(data.view_url); setStatus("Running"); setInput("");
    } finally { setLoading(false); }
  }

  useEffect(() => { const ta = textareaRef.current; if (ta) { ta.style.height = "auto"; ta.style.height = Math.min(ta.scrollHeight, 200) + "px"; } }, [input]);

  return (
    <div className="flex-1 flex flex-col">
      <div className="flex-1 relative">
        {viewUrl ? (
          <iframe src={viewUrl} className="w-full h-full border-0" />
        ) : (
          <div className="flex items-center justify-center h-full">
            <div className="text-center max-w-sm">
              <div className="w-14 h-14 rounded-2xl bg-white border border-[#e3e8ee] flex items-center justify-center mx-auto mb-4" style={{ boxShadow: "0 2px 8px rgba(99,91,255,0.08)" }}>
                <svg className="w-7 h-7 text-[#635bff]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456z" />
                </svg>
              </div>
              <h2 className="text-[17px] font-semibold text-[#0a2540] mb-1.5">{t("ready_to_work")}</h2>
              <p className="text-[13px] text-[#697386] leading-5 whitespace-pre-line">{t("ready_desc")}</p>
            </div>
          </div>
        )}
        {error && (
          <div className="absolute top-3 left-3 right-3 bg-[#fde8ed] border border-[#f5c6cb] rounded-lg px-4 py-3 text-[13px] text-[#df1b41] flex justify-between items-start" style={{ zIndex: 10 }}>
            <div>
              <strong>{locale === "zh" ? "错误" : "Error"}:</strong> {error}
            </div>
            <button onClick={() => setError(null)} className="text-[#df1b41] hover:text-[#ff4d6a] ml-3 shrink-0">✕</button>
          </div>
        )}

        {status && (
          <div className="absolute top-3 right-3 flex items-center gap-2 bg-white border border-[#e3e8ee] rounded-lg px-3 py-1.5 text-[13px] text-[#0a2540]" style={{ boxShadow: "var(--shadow-sm)" }}>
            <span className={`w-1.5 h-1.5 rounded-full ${status === "Running" ? "bg-[#635bff] animate-pulse" : status === "Waiting" ? "bg-[#f5a623]" : status === "Completed" ? "bg-[#0caf60]" : status === "Failed" ? "bg-[#df1b41]" : "bg-[#8792a2]"}`} />
            {status}
          </div>
        )}
      </div>
      <div className="bg-white border-t border-[#e3e8ee] px-5 py-4">
        <div className="flex gap-3 items-end">
          <textarea ref={textareaRef} value={input} onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) { e.preventDefault(); handleSubmit(); } }}
            placeholder={t("input_placeholder")} rows={1}
            className="flex-1 bg-[#f6f9fc] border border-[#e3e8ee] rounded-lg px-4 py-2.5 text-[14px] text-[#0a2540] resize-none focus:outline-none focus:ring-2 focus:ring-[#635bff]/20 focus:border-[#635bff] placeholder:text-[#8792a2] transition-all" />
          <button onClick={handleSubmit} disabled={loading || !input.trim()}
            className="bg-[#635bff] hover:bg-[#7a73ff] text-white rounded-lg px-5 py-2.5 text-[14px] font-medium transition-all disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2 shrink-0"
            style={{ boxShadow: "0 1px 3px rgba(99,91,255,0.3)" }}>
            {loading ? <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" /> :
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M6 12L3.269 3.126A59.768 59.768 0 0121.485 12 59.77 59.77 0 013.27 20.876L5.999 12zm0 0h7.5" /></svg>}
            {t("run_agent")}
          </button>
        </div>
        <div className="flex items-center justify-between mt-2">
          <p className="text-[12px] text-[#8792a2]">
            <span className="text-[#635bff] font-medium">{MODEL_NAMES[model] || model}</span>
            {" · "}<kbd className="text-[11px] bg-[#f6f9fc] border border-[#e3e8ee] rounded px-1 py-0.5 font-mono">⌘↵</kbd> {t("to_send")}
          </p>
          <BalanceBadge />
        </div>
      </div>
    </div>
  );
}
