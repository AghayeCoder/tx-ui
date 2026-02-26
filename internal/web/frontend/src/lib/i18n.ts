import { ref } from "vue";

type ApiResponse<T> = {
  success: boolean;
  obj: T;
  msg?: string;
};

const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";
const messages = ref<Record<string, string>>({});
const loadedKeys = new Set<string>();

function normalizeLocale(input: string): string {
  return String(input || "")
    .trim()
    .replaceAll("_", "-");
}

function readCookie(name: string): string {
  if (typeof document === "undefined") return "";
  const needle = `${name}=`;
  for (const part of document.cookie.split(";")) {
    const item = part.trim();
    if (item.startsWith(needle)) {
      try {
        return decodeURIComponent(item.slice(needle.length));
      } catch {
        return item.slice(needle.length);
      }
    }
  }
  return "";
}

export function isRtlLocale(locale: string): boolean {
  void locale;
  return false;
}

export function applyLocaleDirection(locale?: string): void {
  if (typeof document === "undefined") return;
  const html = document.documentElement;
  const candidate =
    normalizeLocale(locale || "") ||
    normalizeLocale(readCookie("lang")) ||
    normalizeLocale(html.getAttribute("lang") || "") ||
    normalizeLocale(navigator.language || "en-US");
  const normalized = candidate || "en-US";
  html.setAttribute("lang", normalized);
  html.setAttribute("dir", "ltr");
  html.classList.remove("rtl");
}

export function setLocale(locale: string): void {
  applyLocaleDirection(locale);
}

export function t(key: string, fallback?: string): string {
  const value = messages.value[key];
  if (value && value.trim()) {
    return value;
  }
  return fallback ?? key;
}

export async function loadI18n(keys: string[]): Promise<void> {
  const unique = [...new Set(keys.map((k) => String(k || "").trim()).filter(Boolean))];
  const needed = unique.filter((k) => !loadedKeys.has(k));
  if (!needed.length) return;

  try {
    const res = await fetch(`${basePath}panel/api/i18n`, {
      method: "POST",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ keys: needed })
    });
    const data = (await res.json()) as ApiResponse<Record<string, string>>;
    if (!data.success || !data.obj) return;
    messages.value = { ...messages.value, ...data.obj };
    for (const k of needed) {
      loadedKeys.add(k);
    }
  } catch {
    // best-effort translation loading
  }
}
