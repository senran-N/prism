"use client";

import { useLocale } from "./locale-context";

import type { Key } from "@/app/i18n";

type ModelItem = { id: string; name: string; tagKey?: Key };
type ModelGroup = { group: string; icon: string; items: ModelItem[] };

const MODELS: ModelGroup[] = [
  { group: "Claude Code", icon: "✦", items: [
    { id: "claude_code_claude_opus_4_8", name: "Opus 4.8", tagKey: "most_capable" },
    { id: "claude_code_claude_opus_4_7", name: "Opus 4.7" },
    { id: "claude_code_claude_opus_4_6", name: "Opus 4.6" },
    { id: "claude_code_claude_sonnet_4_6", name: "Sonnet 4.6", tagKey: "fast" },
  ]},
  { group: "Codex", icon: "◎", items: [
    { id: "codex_gpt_5_5_high", name: "GPT-5.5 High" },
    { id: "codex_gpt_5_5_medium", name: "GPT-5.5 Medium", tagKey: "economy" },
    { id: "codex_gpt_5_5_xhigh", name: "GPT-5.5 Xhigh" },
  ]},
  { group: "OpenCode", icon: "◈", items: [
    { id: "opencode_gemini_3_1_pro", name: "Gemini 3.1 Pro" },
    { id: "opencode_gpt_5_5", name: "GPT-5.5" },
    { id: "opencode_kimi_k2_6", name: "Kimi K2.6" },
  ]},
  { group: "Pi", icon: "◉", items: [
    { id: "pi_deepseek_v4_pro", name: "DeepSeek V4 Pro" },
    { id: "pi_deepseek_v4_flash", name: "DeepSeek V4 Flash", tagKey: "fast" },
  ]},
];

export default function ModelSelector({ value, onChange }: { value: string; onChange: (id: string) => void }) {
  const { t } = useLocale();
  const current = MODELS.flatMap(g => g.items).find(m => m.id === value);
  const currentGroup = MODELS.find(g => g.items.some(m => m.id === value));

  return (
    <div className="flex items-center gap-3">
      <span className="text-[13px] text-[#697386]">{t("model")}</span>
      <div className="relative">
        <select value={value} onChange={(e) => onChange(e.target.value)}
          className="appearance-none bg-[#f6f9fc] border border-[#e3e8ee] rounded-md px-3 py-1.5 pr-8 text-[13px] font-medium text-[#0a2540] focus:outline-none focus:ring-2 focus:ring-[#635bff]/20 focus:border-[#635bff] cursor-pointer transition-all hover:border-[#c1c9d2]">
          {MODELS.map((group) => (
            <optgroup key={group.group} label={`${group.icon} ${group.group}`}>
              {group.items.map((model) => (
                <option key={model.id} value={model.id}>
                  {model.name}{model.tagKey ? ` · ${t(model.tagKey)}` : ""}
                </option>
              ))}
            </optgroup>
          ))}
        </select>
        <svg className="absolute right-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-[#697386] pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
        </svg>
      </div>
      {currentGroup && <span className="inline-flex items-center gap-1 text-[12px] text-[#697386] bg-[#f6f9fc] border border-[#e3e8ee] rounded px-1.5 py-0.5">{currentGroup.icon} {currentGroup.group}</span>}
      {current?.tagKey && <span className="text-[11px] font-medium text-[#635bff] bg-[#e8e6ff] rounded px-1.5 py-0.5">{t(current.tagKey)}</span>}
    </div>
  );
}
