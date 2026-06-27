"use client";

import { useState, useEffect } from "react";

interface Stats {
  pool: { ready: number; active: number; exhausted: number; total_credits: number; total: number };
  tasks: number; users: number; uptime: string; go_routines: number; mem_mb: number;
}
interface Account { id: string; email: string; workspace_id: string; project_id: string; credits: number; status: string; github_bound: boolean; created_at: string; }
interface TaskRow { id: string; ticket_id: string; description: string; model: string; status: string; cost: number; created_at: string; user_login: string; }
interface UserRow { id: number; github_login: string; avatar_url: string; selected_repo: string; linuxdo_username: string; linuxdo_name: string; trust_level: number; is_banned: boolean; ban_reason: string; created_at: string; task_count: number; }
interface Config { github_user: string; github_pass: string; github_totp: string; yyds_api_key: string; repo_id: string; github_client_id: string; base_url: string; }

type Tab = "overview" | "accounts" | "tasks" | "users" | "config";

export default function AdminPage() {
  const [tab, setTab] = useState<Tab>("overview");

  return (
    <div className="min-h-screen bg-[#f6f9fc]">
      <header className="bg-white border-b border-[#e3e8ee] px-6 py-3 flex items-center justify-between" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
        <div className="flex items-center gap-3">
          <a href="/" className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-md bg-[#635bff] flex items-center justify-center">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2" />
                <line x1="12" y1="22" x2="12" y2="15.5" />
                <polyline points="22 8.5 12 15.5 2 8.5" />
              </svg>
            </div>
            <span className="text-[15px] font-semibold text-[#0a2540]">Prism</span>
          </a>
          <span className="text-[12px] text-[#8792a2] border border-[#e3e8ee] rounded px-1.5 py-0.5">Admin</span>
        </div>
        <a href="/" className="text-[13px] text-[#697386] hover:text-[#0a2540]">← Back to app</a>
      </header>

      <div className="max-w-6xl mx-auto px-6 py-6">
        <nav className="flex gap-1 mb-6 bg-white rounded-lg border border-[#e3e8ee] p-1" style={{ boxShadow: "var(--shadow-sm)" }}>
          {(["overview", "accounts", "tasks", "users", "config"] as Tab[]).map(t => (
            <button key={t} onClick={() => setTab(t)} className={`px-4 py-2 rounded-md text-[13px] font-medium capitalize transition-colors ${tab === t ? "bg-[#635bff] text-white" : "text-[#697386] hover:text-[#0a2540] hover:bg-[#f6f9fc]"}`}>
              {t}
            </button>
          ))}
        </nav>

        {tab === "overview" && <OverviewTab />}
        {tab === "accounts" && <AccountsTab />}
        {tab === "tasks" && <TasksTab />}
        {tab === "users" && <UsersTab />}
        {tab === "config" && <ConfigTab />}
      </div>
    </div>
  );
}

function Card({ title, value, sub }: { title: string; value: string | number; sub?: string }) {
  return (
    <div className="bg-white rounded-lg border border-[#e3e8ee] p-5" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
      <div className="text-[12px] text-[#8792a2] uppercase tracking-wider font-medium">{title}</div>
      <div className="text-[28px] font-semibold text-[#0a2540] mt-1">{value}</div>
      {sub && <div className="text-[12px] text-[#697386] mt-1">{sub}</div>}
    </div>
  );
}

function OverviewTab() {
  const [stats, setStats] = useState<Stats | null>(null);
  useEffect(() => {
    fetch("/api/admin/stats").then(r => r.json()).then(setStats);
    const id = setInterval(() => fetch("/api/admin/stats").then(r => r.json()).then(setStats), 10000);
    return () => clearInterval(id);
  }, []);
  if (!stats) return <Loading />;

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <Card title="Accounts Ready" value={stats.pool.ready} sub={`${stats.pool.total} total`} />
      <Card title="Credits Available" value={`$${stats.pool.total_credits.toFixed(2)}`} sub={`${stats.pool.exhausted} exhausted`} />
      <Card title="Total Tasks" value={stats.tasks} />
      <Card title="Total Users" value={stats.users} />
      <Card title="Active Now" value={stats.pool.active} sub="accounts in use" />
      <Card title="Uptime" value={stats.uptime.split(".")[0]} />
      <Card title="Goroutines" value={stats.go_routines} />
      <Card title="Memory" value={`${stats.mem_mb} MB`} />
    </div>
  );
}

function AccountsTab() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  useEffect(() => { fetch("/api/admin/accounts").then(r => r.json()).then(d => setAccounts(Array.isArray(d) ? d : [])); }, []);

  const statusColor: Record<string, string> = { ready: "text-[#0caf60] bg-[#e6f9f0]", active: "text-[#635bff] bg-[#e8e6ff]", exhausted: "text-[#df1b41] bg-[#fde8ed]", error: "text-[#df1b41] bg-[#fde8ed]" };

  return (
    <div className="bg-white rounded-lg border border-[#e3e8ee] overflow-hidden" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
      <table className="w-full text-[13px]">
        <thead><tr className="border-b border-[#e3e8ee] bg-[#f6f9fc]">
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Email</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Status</th>
          <th className="text-right px-4 py-3 text-[#8792a2] font-medium">Credits</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">GitHub</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Workspace</th>
        </tr></thead>
        <tbody>
          {accounts.length === 0 && <tr><td colSpan={5} className="px-4 py-8 text-center text-[#8792a2]">No accounts in pool</td></tr>}
          {accounts.map(a => (
            <tr key={a.id} className="border-b border-[#e3e8ee] hover:bg-[#f6f9fc]">
              <td className="px-4 py-3 font-mono text-[12px]">{a.email}</td>
              <td className="px-4 py-3"><span className={`text-[11px] font-medium px-2 py-0.5 rounded ${statusColor[a.status] || "text-[#697386] bg-[#f6f9fc]"}`}>{a.status}</span></td>
              <td className="px-4 py-3 text-right font-mono">${a.credits.toFixed(2)}</td>
              <td className="px-4 py-3">{a.github_bound ? <span className="text-[#0caf60]">●</span> : <span className="text-[#8792a2]">○</span>}</td>
              <td className="px-4 py-3 font-mono text-[12px] text-[#8792a2]">{a.workspace_id}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TasksTab() {
  const [tasks, setTasks] = useState<TaskRow[]>([]);
  useEffect(() => { fetch("/api/admin/tasks").then(r => r.json()).then(d => setTasks(Array.isArray(d) ? d : [])); }, []);

  const statusColor: Record<string, string> = { Running: "text-[#635bff] bg-[#e8e6ff]", Waiting: "text-[#f5a623] bg-[#fef6e6]", Completed: "text-[#0caf60] bg-[#e6f9f0]", Failed: "text-[#df1b41] bg-[#fde8ed]", created: "text-[#635bff] bg-[#e8e6ff]" };

  return (
    <div className="bg-white rounded-lg border border-[#e3e8ee] overflow-hidden" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
      <table className="w-full text-[13px]">
        <thead><tr className="border-b border-[#e3e8ee] bg-[#f6f9fc]">
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Task</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Model</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">User</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Status</th>
          <th className="text-right px-4 py-3 text-[#8792a2] font-medium">Cost</th>
          <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Created</th>
        </tr></thead>
        <tbody>
          {tasks.length === 0 && <tr><td colSpan={6} className="px-4 py-8 text-center text-[#8792a2]">No tasks yet</td></tr>}
          {tasks.map(t => (
            <tr key={t.id} className="border-b border-[#e3e8ee] hover:bg-[#f6f9fc]">
              <td className="px-4 py-3 max-w-[300px] truncate">{t.description}</td>
              <td className="px-4 py-3 text-[12px] text-[#697386]">{t.model.replace(/_/g, " ")}</td>
              <td className="px-4 py-3 text-[12px]">{t.user_login || "—"}</td>
              <td className="px-4 py-3"><span className={`text-[11px] font-medium px-2 py-0.5 rounded ${statusColor[t.status] || ""}`}>{t.status}</span></td>
              <td className="px-4 py-3 text-right font-mono">{t.cost > 0 ? `$${t.cost.toFixed(2)}` : "—"}</td>
              <td className="px-4 py-3 text-[12px] text-[#8792a2]">{new Date(t.created_at).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function UsersTab() {
  const [users, setUsers] = useState<UserRow[]>([]);
  const [banModal, setBanModal] = useState<{id: number; name: string} | null>(null);
  const [banReason, setBanReason] = useState("");

  useEffect(() => { fetch("/api/admin/users").then(r => r.json()).then(d => setUsers(Array.isArray(d) ? d : [])); }, []);

  async function handleBan(id: number) {
    await fetch(`/api/admin/users/${id}/ban`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ reason: banReason }) });
    setUsers(users.map(u => u.id === id ? { ...u, is_banned: true, ban_reason: banReason } : u));
    setBanModal(null); setBanReason("");
  }

  async function handleUnban(id: number) {
    await fetch(`/api/admin/users/${id}/unban`, { method: "POST" });
    setUsers(users.map(u => u.id === id ? { ...u, is_banned: false, ban_reason: "" } : u));
  }

  return (
    <>
      <div className="bg-white rounded-lg border border-[#e3e8ee] overflow-hidden" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
        <table className="w-full text-[13px]">
          <thead><tr className="border-b border-[#e3e8ee] bg-[#f6f9fc]">
            <th className="text-left px-4 py-3 text-[#8792a2] font-medium">User</th>
            <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Trust</th>
            <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Repository</th>
            <th className="text-right px-4 py-3 text-[#8792a2] font-medium">Tasks</th>
            <th className="text-left px-4 py-3 text-[#8792a2] font-medium">Status</th>
            <th className="text-right px-4 py-3 text-[#8792a2] font-medium">Actions</th>
          </tr></thead>
          <tbody>
            {users.length === 0 && <tr><td colSpan={6} className="px-4 py-8 text-center text-[#8792a2]">No users yet</td></tr>}
            {users.map(u => (
              <tr key={u.id} className={`border-b border-[#e3e8ee] hover:bg-[#f6f9fc] ${u.is_banned ? "opacity-60" : ""}`}>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    {u.avatar_url && <img src={u.avatar_url} className="w-6 h-6 rounded-full" alt="" />}
                    <div>
                      <div className="font-medium">{u.linuxdo_name || u.linuxdo_username || u.github_login}</div>
                      {u.linuxdo_username && <div className="text-[11px] text-[#8792a2]">@{u.linuxdo_username}</div>}
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">
                  <span className={`text-[11px] font-medium px-1.5 py-0.5 rounded ${u.trust_level >= 3 ? "text-[#635bff] bg-[#e8e6ff]" : u.trust_level >= 1 ? "text-[#0caf60] bg-[#e6f9f0]" : "text-[#8792a2] bg-[#f6f9fc]"}`}>
                    TL{u.trust_level}
                  </span>
                </td>
                <td className="px-4 py-3 font-mono text-[12px] text-[#697386]">{u.selected_repo || "—"}</td>
                <td className="px-4 py-3 text-right">{u.task_count}</td>
                <td className="px-4 py-3">
                  {u.is_banned ? (
                    <span className="text-[11px] font-medium text-[#df1b41] bg-[#fde8ed] px-2 py-0.5 rounded" title={u.ban_reason}>Banned</span>
                  ) : (
                    <span className="text-[11px] font-medium text-[#0caf60] bg-[#e6f9f0] px-2 py-0.5 rounded">Active</span>
                  )}
                </td>
                <td className="px-4 py-3 text-right">
                  {u.is_banned ? (
                    <button onClick={() => handleUnban(u.id)} className="text-[12px] text-[#0caf60] hover:text-[#0a9050] font-medium">Unban</button>
                  ) : (
                    <button onClick={() => setBanModal({ id: u.id, name: u.linuxdo_name || u.linuxdo_username || u.github_login })} className="text-[12px] text-[#df1b41] hover:text-[#ff4d6a] font-medium">Ban</button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Ban Modal */}
      {banModal && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setBanModal(null)}>
          <div className="bg-white rounded-xl border border-[#e3e8ee] p-6 w-full max-w-md" style={{ boxShadow: "0 10px 40px rgba(0,0,0,0.12)" }} onClick={e => e.stopPropagation()}>
            <h3 className="text-[16px] font-semibold text-[#0a2540] mb-1">Ban User</h3>
            <p className="text-[13px] text-[#697386] mb-4">Ban <strong>{banModal.name}</strong> from using Prism.</p>
            <input type="text" value={banReason} onChange={e => setBanReason(e.target.value)} placeholder="Reason for ban (optional)"
              className="w-full bg-[#f6f9fc] border border-[#e3e8ee] rounded-lg px-3 py-2 text-[13px] mb-4 focus:outline-none focus:ring-2 focus:ring-[#df1b41]/20 focus:border-[#df1b41]" />
            <div className="flex gap-2 justify-end">
              <button onClick={() => setBanModal(null)} className="px-4 py-2 text-[13px] text-[#697386] hover:text-[#0a2540]">Cancel</button>
              <button onClick={() => handleBan(banModal.id)} className="px-4 py-2 text-[13px] font-medium text-white bg-[#df1b41] hover:bg-[#c4162f] rounded-lg">Confirm Ban</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

function ConfigTab() {
  const [config, setConfig] = useState<Config | null>(null);
  const [form, setForm] = useState({ github_user: "", github_pass: "", github_totp: "", yyds_api_key: "", repo_id: "" });
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => { fetch("/api/admin/config").then(r => r.json()).then(setConfig); }, []);

  async function handleSave() {
    setSaving(true);
    const body: Record<string, string> = {};
    if (form.github_user) body.github_user = form.github_user;
    if (form.github_pass) body.github_pass = form.github_pass;
    if (form.github_totp) body.github_totp = form.github_totp;
    if (form.yyds_api_key) body.yyds_api_key = form.yyds_api_key;
    if (form.repo_id) body.repo_id = form.repo_id;
    await fetch("/api/admin/config", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(body) });
    setSaving(false);
    setSaved(true);
    setForm({ github_user: "", github_pass: "", github_totp: "", yyds_api_key: "", repo_id: "" });
    fetch("/api/admin/config").then(r => r.json()).then(setConfig);
    setTimeout(() => setSaved(false), 3000);
  }

  if (!config) return <Loading />;

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Current config */}
      <div className="bg-white rounded-lg border border-[#e3e8ee] p-6" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
        <h3 className="text-[14px] font-semibold text-[#0a2540] mb-4">Current Configuration</h3>
        <div className="space-y-3">
          {Object.entries(config).map(([k, v]) => (
            <div key={k} className="flex justify-between text-[13px]">
              <span className="text-[#697386]">{k.replace(/_/g, " ")}</span>
              <code className="text-[12px] bg-[#f6f9fc] px-2 py-0.5 rounded text-[#0a2540]">{v || "—"}</code>
            </div>
          ))}
        </div>
      </div>

      {/* Update form */}
      <div className="bg-white rounded-lg border border-[#e3e8ee] p-6" style={{ boxShadow: "0 1px 3px rgba(0,0,0,0.04)" }}>
        <h3 className="text-[14px] font-semibold text-[#0a2540] mb-4">Update Configuration</h3>
        <p className="text-[12px] text-[#8792a2] mb-4">Leave fields empty to keep current values.</p>
        <div className="space-y-3">
          {[
            { key: "github_user", label: "GitHub Service User", type: "text" },
            { key: "github_pass", label: "GitHub Password", type: "password" },
            { key: "github_totp", label: "GitHub TOTP Secret", type: "password" },
            { key: "yyds_api_key", label: "YYDS API Key", type: "password" },
            { key: "repo_id", label: "Default Repo ID", type: "text" },
          ].map(({ key, label, type }) => (
            <div key={key}>
              <label className="text-[12px] text-[#697386] block mb-1">{label}</label>
              <input
                type={type}
                value={form[key as keyof typeof form]}
                onChange={e => setForm({ ...form, [key]: e.target.value })}
                placeholder={`Enter ${label.toLowerCase()}`}
                className="w-full bg-[#f6f9fc] border border-[#e3e8ee] rounded-md px-3 py-2 text-[13px] focus:outline-none focus:ring-2 focus:ring-[#635bff]/20 focus:border-[#635bff]"
              />
            </div>
          ))}
          <button
            onClick={handleSave}
            disabled={saving}
            className="bg-[#635bff] hover:bg-[#7a73ff] text-white rounded-md px-4 py-2 text-[13px] font-medium transition-colors disabled:opacity-50 w-full"
          >
            {saving ? "Saving..." : saved ? "✓ Saved" : "Save Configuration"}
          </button>
        </div>
      </div>
    </div>
  );
}

function Loading() {
  return (
    <div className="flex items-center justify-center py-20">
      <span className="w-6 h-6 border-2 border-[#e3e8ee] border-t-[#635bff] rounded-full animate-spin" />
    </div>
  );
}
