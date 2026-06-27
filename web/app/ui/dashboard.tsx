"use client";

import { useState } from "react";
import GitHubPanel from "./github-panel";
import ModelSelector from "./model-selector";
import TaskPanel from "./task-panel";
import TaskList from "./task-list";

type Tab = "tasks" | "github";

export default function Dashboard() {
  const [tab, setTab] = useState<Tab>("tasks");
  const [selectedModel, setSelectedModel] = useState("claude_code_claude_opus_4_8");
  const [tasks, setTasks] = useState<Task[]>([]);
  const [activeTask, setActiveTask] = useState<string | null>(null);

  function addTask(task: Task) {
    setTasks((prev) => [task, ...prev]);
    setActiveTask(task.id);
  }

  return (
    <div className="flex h-[calc(100vh-53px)]">
      {/* Sidebar */}
      <aside className="w-[280px] bg-white border-r border-[#e3e8ee] flex flex-col">
        <nav className="flex border-b border-[#e3e8ee]">
          {(["tasks", "github"] as Tab[]).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`flex-1 py-2.5 text-[13px] font-medium capitalize transition-colors ${
                tab === t
                  ? "text-[#635bff] border-b-2 border-[#635bff]"
                  : "text-[#697386] hover:text-[#0a2540]"
              }`}
            >
              {t === "tasks" ? "Tasks" : "GitHub"}
            </button>
          ))}
        </nav>

        {tab === "tasks" ? (
          <TaskList tasks={tasks} activeTask={activeTask} onSelect={setActiveTask} />
        ) : (
          <GitHubPanel />
        )}
      </aside>

      {/* Main */}
      <div className="flex-1 flex flex-col bg-[#f6f9fc]">
        <div className="bg-white border-b border-[#e3e8ee] px-5 py-2.5">
          <ModelSelector value={selectedModel} onChange={setSelectedModel} />
        </div>
        <TaskPanel model={selectedModel} activeTask={activeTask} onTaskCreated={addTask} />
      </div>
    </div>
  );
}

export interface Task {
  id: string;
  description: string;
  model: string;
  modelName: string;
  status: string;
  createdAt: string;
  viewUrl?: string;
}
