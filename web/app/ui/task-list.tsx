"use client";

import { useLocale } from "./locale-context";
import type { Task } from "./dashboard";

interface Props { tasks: Task[]; activeTask: string | null; onSelect: (id: string) => void; }

export default function TaskList({ tasks, activeTask, onSelect }: Props) {
  const { t } = useLocale();

  const STATUS = {
    Running:   { color: "bg-[#635bff]", pulse: true, label: t("running") },
    Waiting:   { color: "bg-[#f5a623]", pulse: false, label: t("waiting") },
    Completed: { color: "bg-[#0caf60]", pulse: false, label: t("completed") },
    Failed:    { color: "bg-[#df1b41]", pulse: false, label: t("failed") },
    created:   { color: "bg-[#635bff]", pulse: true, label: t("starting") },
  } as const;

  if (tasks.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-6">
        <div className="w-10 h-10 rounded-xl bg-[#f6f9fc] border border-[#e3e8ee] flex items-center justify-center mb-3">
          <svg className="w-5 h-5 text-[#8792a2]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
          </svg>
        </div>
        <p className="text-[13px] text-[#697386] text-center leading-5 whitespace-pre-line">{t("no_tasks")}</p>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto">
      {tasks.map((task) => {
        const s = STATUS[task.status as keyof typeof STATUS] || STATUS.created;
        return (
          <button key={task.id} onClick={() => onSelect(task.id)}
            className={`w-full text-left px-4 py-3 border-b border-[#e3e8ee] transition-colors ${activeTask === task.id ? "bg-[#f0efff]" : "hover:bg-[#f6f9fc]"}`}>
            <div className="flex items-center justify-between mb-1">
              <code className="text-[11px] text-[#8792a2] font-mono">#{task.id.slice(0, 8)}</code>
              <div className="flex items-center gap-1.5">
                <span className={`w-1.5 h-1.5 rounded-full ${s.color} ${s.pulse ? "animate-pulse" : ""}`} />
                <span className="text-[11px] text-[#697386]">{s.label}</span>
              </div>
            </div>
            <p className="text-[13px] text-[#0a2540] truncate leading-5">{task.description}</p>
            <p className="text-[11px] text-[#8792a2] mt-0.5">{task.modelName}</p>
          </button>
        );
      })}
    </div>
  );
}
