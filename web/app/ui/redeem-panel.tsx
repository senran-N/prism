"use client";

import { useState, useEffect } from "react";
import { useLocale } from "./locale-context";

interface BalanceInfo {
  balance: number;
  rotation_cost: number;
  can_rotate: boolean;
  total_rotations: number;
}

export default function RedeemPanel() {
  const { locale } = useLocale();
  const isZh = locale === "zh";
  const [balance, setBalance] = useState<BalanceInfo | null>(null);
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    fetch("/api/balance").then(r => r.json()).then(setBalance).catch(() => {});
  }, []);

  async function handleRedeem() {
    if (!code.trim()) return;
    setLoading(true);
    setMessage(null);
    try {
      const res = await fetch("/api/redeem", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code: code.trim() }),
      });
      const data = await res.json();
      if (res.ok) {
        setMessage({
          type: "success",
          text: isZh
            ? `兑换成功！获得 ${data.rotations} 次轮转（${data.credits} 积分）`
            : `Redeemed! Got ${data.rotations} rotation${data.rotations > 1 ? "s" : ""} (${data.credits} credits)`,
        });
        setBalance((prev) => prev ? { ...prev, balance: data.balance, can_rotate: data.balance >= (prev.rotation_cost || 20) } : prev);
        setCode("");
      } else {
        setMessage({ type: "error", text: data.error || "Redemption failed" });
      }
    } catch {
      setMessage({ type: "error", text: "Network error" });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex-1 flex flex-col p-4 gap-4">
      {/* Balance card */}
      <div className="bg-[#f0efff] border border-[#d9d6ff] rounded-lg p-4">
        <div className="text-[11px] text-[#635bff] font-medium uppercase tracking-wider">
          {isZh ? "当前余额" : "Balance"}
        </div>
        <div className="text-[28px] font-semibold text-[#0a2540] mt-1">
          {balance ? balance.balance.toFixed(0) : "—"}
          <span className="text-[14px] text-[#697386] ml-1">{isZh ? "积分" : "credits"}</span>
        </div>
        <div className="flex items-center gap-3 mt-2 text-[12px] text-[#697386]">
          <span>{isZh ? "每次轮转" : "Per rotation"}: {balance?.rotation_cost || 20}</span>
          <span>·</span>
          <span>{isZh ? "可用" : "Available"}: {balance ? Math.floor(balance.balance / (balance.rotation_cost || 20)) : 0} {isZh ? "次" : "times"}</span>
        </div>
      </div>

      {/* Redeem form */}
      <div>
        <h3 className="text-[12px] font-medium text-[#8792a2] uppercase tracking-wider mb-2">
          {isZh ? "兑换码" : "Redeem Code"}
        </h3>
        <div className="flex gap-2">
          <input
            type="text"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleRedeem()}
            placeholder={isZh ? "输入兑换码 PRISM-XXXX" : "Enter code PRISM-XXXX"}
            className="flex-1 bg-[#f6f9fc] border border-[#e3e8ee] rounded-lg px-3 py-2 text-[13px] focus:outline-none focus:ring-2 focus:ring-[#635bff]/20 focus:border-[#635bff] placeholder:text-[#8792a2]"
          />
          <button
            onClick={handleRedeem}
            disabled={loading || !code.trim()}
            className="bg-[#635bff] hover:bg-[#7a73ff] text-white rounded-lg px-4 py-2 text-[13px] font-medium transition-colors disabled:opacity-50 shrink-0"
          >
            {loading ? "..." : isZh ? "兑换" : "Redeem"}
          </button>
        </div>
      </div>

      {/* Message */}
      {message && (
        <div className={`rounded-lg px-3 py-2 text-[13px] ${
          message.type === "success" ? "bg-[#e6f9f0] text-[#0caf60]" : "bg-[#fde8ed] text-[#df1b41]"
        }`}>
          {message.text}
        </div>
      )}

      {/* LDC Credit 充值 */}
      <div>
        <h3 className="text-[12px] font-medium text-[#8792a2] uppercase tracking-wider mb-2">
          {isZh ? "积分充值" : "Buy Credits"}
        </h3>
        <button
          onClick={async () => {
            const res = await fetch("/api/credit/pay", {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ amount: balance?.rotation_cost || 20, description: "Prism 1 Rotation" }),
            });
            if (res.ok) {
              const data = await res.json();
              if (data.pay_url) window.location.href = data.pay_url;
            } else {
              const err = await res.json();
              setMessage({ type: "error", text: err.error || "Payment failed" });
            }
          }}
          className="w-full flex items-center justify-center gap-2 bg-[#f39c12] hover:bg-[#e67e22] text-white rounded-lg px-4 py-2.5 text-[13px] font-medium transition-colors"
        >
          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>
          {isZh ? `支付 ${balance?.rotation_cost || 20} LDC = 1 次轮转` : `Pay ${balance?.rotation_cost || 20} LDC = 1 rotation`}
        </button>
        <p className="text-[11px] text-[#8792a2] mt-1 text-center">{isZh ? "通过 LinuxDo Credit 支付" : "Pay via LinuxDo Credit"}</p>
      </div>

      {/* Stats */}
      {balance && (
        <div className="mt-auto pt-4 border-t border-[#e3e8ee] text-[12px] text-[#8792a2]">
          <div className="flex justify-between mb-1">
            <span>{isZh ? "总轮转次数" : "Total rotations"}</span>
            <span>{balance.total_rotations}</span>
          </div>
          <div className="flex justify-between">
            <span>{isZh ? "状态" : "Status"}</span>
            <span className={balance.can_rotate ? "text-[#0caf60]" : "text-[#f5a623]"}>
              {balance.can_rotate ? (isZh ? "可用" : "Ready") : (isZh ? "余额不足" : "Low balance")}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
