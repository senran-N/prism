"use client";

import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import { type Locale, type Key, t } from "@/app/i18n";

interface LocaleCtx {
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: (key: Key) => string;
}

const Ctx = createContext<LocaleCtx>({
  locale: "en",
  setLocale: () => {},
  t: (k) => k,
});

export function LocaleProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("prism_locale");
      if (saved === "zh" || saved === "en") return saved;
      return navigator.language.startsWith("zh") ? "zh" : "en";
    }
    return "en";
  });

  const setLocale = useCallback((l: Locale) => {
    setLocaleState(l);
    if (typeof window !== "undefined") localStorage.setItem("prism_locale", l);
  }, []);

  const translate = useCallback((key: Key) => t(key, locale), [locale]);

  return <Ctx.Provider value={{ locale, setLocale, t: translate }}>{children}</Ctx.Provider>;
}

export function useLocale() {
  return useContext(Ctx);
}

export function LocaleSwitch() {
  const { locale, setLocale } = useLocale();
  return (
    <button
      onClick={() => setLocale(locale === "en" ? "zh" : "en")}
      className="text-[12px] text-[#8792a2] hover:text-[#635bff] transition-colors border border-[#e3e8ee] rounded px-1.5 py-0.5"
      title="Switch language"
    >
      {locale === "en" ? "中文" : "EN"}
    </button>
  );
}
