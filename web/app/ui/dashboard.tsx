"use client";

import { useState } from "react";
import { useLocale } from "./locale-context";
import GitHubPanel from "./github-panel";
import RedeemPanel from "./redeem-panel";
import ModelSelector from "./model-selector";
import TaskPanel from "./task-panel";
import TaskList from "./task-list";

type SideTab = "tasks" | "github" | "credits";

export default function Dashboard() {
  const { t } = useLocale();
  const [tab, setTab] = useState<SideTab>("tasks");
  const [selectedModel, setSelectedModel] = useState("claude_code_claude_opus_4_8");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [activeTask, setActiveTask] = useState<string | null>(null);

  return (
    <div className="flex h-[calc(100vh-53px)]">
      <aside className="w-[280px] bg-white border-r border-[#e3e8ee] flex flex-col">
        <nav className="flex border-b border-[#e3e8ee]">
          {(["tasks", "github", "credits"] as SideTab[]).map((tb) => (
            <button key={tb} onClick={() => setTab(tb)}
              className={`flex-1 py-2.5 text-[13px] font-medium transition-colors ${tab === tb ? "text-[#635bff] border-b-2 border-[#635bff]" : "text-[#697386] hover:text-[#0a2540]"}`}>
              {tb === "credits" ? (t as any)("credits") || "Credits" : t(tb as any)}
            </button>
          ))}
        </nav>
        {tab === "tasks" && <TaskList tasks={tasks} activeTask={activeTask} onSelect={setActiveTask} />}
        {tab === "github" && <GitHubPanel />}
        {tab === "credits" && <RedeemPanel />}
      </aside>
      <div className="flex-1 flex flex-col bg-[#f6f9fc]">
        <div className="bg-white border-b border-[#e3e8ee] px-5 py-2.5">
          <ModelSelector value={selectedModel} onChange={setSelectedModel} />
        </div>
        <TaskPanel model={selectedModel} activeTask={activeTask} onTaskCreated={(task) => { setTasks(prev => [task, ...prev]); setActiveTask(task.id); }} />
      </div>
    </div>
  );
}

export interface Task {
  id: string; description: string; model: string; modelName: string;
  status: string; createdAt: string; viewUrl?: string;
}
