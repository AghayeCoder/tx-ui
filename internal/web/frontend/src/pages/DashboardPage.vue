<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { MoreVertical } from "lucide-vue-next";
import { UiButton } from "@/components/ui/button";
import { UiCard, UiCardContent, UiCardDescription, UiCardHeader, UiCardTitle } from "@/components/ui/card";
import { t, loadI18n } from "@/lib/i18n";

type ApiResponse<T> = {
  success: boolean;
  obj: T;
  msg?: string;
};

type ServerStatus = {
  cpu?: number;
  mem?: { current?: number; total?: number };
  swap?: { current?: number; total?: number };
  disk?: { current?: number; total?: number };
  xray?: { state?: string; version?: string; errorMsg?: string };
  hostname?: string;
  publicIP?: { ipv4?: string; ipv6?: string };
  cpuCores?: number;
  logicalPro?: number;
  cpuSpeedMhz?: number;
  uptime?: number;
  loads?: number[];
  tcpCount?: number;
  udpCount?: number;
  netIO?: { up?: number; down?: number };
  netTraffic?: { sent?: number; recv?: number };
  geoFiles?: {
    geoip?: { exists?: boolean; size?: number; updatedAt?: number; version?: string };
    geosite?: { exists?: boolean; size?: number; updatedAt?: number; version?: string };
  };
  appStats?: {
    mem?: number;
    threads?: number;
    uptime?: number;
    tx_update?: string;
    panel_version?: string;
  };
};

const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";

const loading = ref(false);
const busy = ref(false);
const error = ref("");
const success = ref("");
const confirmModalOpen = ref(false);
const confirmModalTitle = ref(t("confirm", "Confirm"));
const confirmModalMessage = ref("");
const status = ref<ServerStatus | null>(null);
const configModalOpen = ref(false);
const configJson = ref("");
const logsModalOpen = ref(false);
const logsTitle = ref("Logs");
const logsContent = ref("");
const logsType = ref<"panel" | "xray">("panel");
const panelLogRows = ref("100");
const panelLogLevel = ref("debug");
const panelLogSyslog = ref(false);
const xrayLogRows = ref("100");
const xrayLogFilter = ref("");
const xrayLogShowDirect = ref(true);
const xrayLogShowBlocked = ref(true);
const xrayLogShowProxy = ref(true);
const importDbInput = ref<HTMLInputElement | null>(null);
const xrayVersionModalOpen = ref(false);
const xrayVersions = ref<string[]>([]);
const selectedXrayVersion = ref("");
let refreshTimer: number | null = null;
let statusRequestInFlight = false;
let notifyTimer: number | null = null;
let confirmResolver: ((value: boolean) => void) | null = null;

type ParsedLogRow = {
  ts: string;
  level: string;
  message: string;
  raw: string;
};

type ParsedXrayLogTableRow = {
  dateTime: string;
  fromAddress: string;
  toAddress: string;
  inbound: string;
  outbound: string;
  email: string;
  raw: string;
};

function escapeHtml(input: string): string {
  return input
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}

function highlightJsonForHtml(jsonText: string): string {
  const source = escapeHtml(jsonText);
  return source.replace(
    /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(\s*:)?|\btrue\b|\bfalse\b|\bnull\b|-?\d+(?:\.\d+)?(?:[eE][+\-]?\d+)?)/g,
    (match) => {
      let cls = "json-number";
      if (match.startsWith('"')) {
        cls = match.endsWith(":") ? "json-key" : "json-string";
      } else if (match === "true" || match === "false") {
        cls = "json-boolean";
      } else if (match === "null") {
        cls = "json-null";
      }
      return `<span class="${cls}">${match}</span>`;
    }
  );
}

function splitLogLines(text: string): string[] {
  return String(text || "")
    .split(/\r?\n/)
    .map((line) => line.trimEnd())
    .filter((line) => line.length > 0);
}

function normalizeLogLevel(level: string): string {
  const l = String(level || "").trim().toLowerCase();
  if (l === "warn") return "warning";
  if (l === "err") return "error";
  return l;
}

function detectLevel(input: string): string {
  const m = input.match(/\b(debug|info|notice|warning|warn|error|err)\b/i);
  return normalizeLogLevel(m?.[1] || "");
}

function parsePanelLogLine(line: string): ParsedLogRow {
  const source = unwrapLogArrayLine(line);
  if (!source) {
    return { ts: "", level: "", message: "", raw: "" };
  }

  const canonical = source.match(/^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+([A-Za-z]+)\s*-\s*(.*)$/);
  if (canonical) {
    return {
      ts: String(canonical[1] || ""),
      level: normalizeLogLevel(String(canonical[2] || "")),
      message: String(canonical[3] || ""),
      raw: source
    };
  }

  const bracketed = source.match(/^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[([A-Za-z]+)\]\s*(.*)$/);
  if (bracketed) {
    return {
      ts: String(bracketed[1] || ""),
      level: normalizeLogLevel(String(bracketed[2] || "")),
      message: String(bracketed[3] || ""),
      raw: source
    };
  }

  const spaceLevel = source.match(/^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+([A-Za-z]+)\s+(.*)$/);
  if (spaceLevel) {
    return {
      ts: String(spaceLevel[1] || ""),
      level: normalizeLogLevel(String(spaceLevel[2] || "")),
      message: String(spaceLevel[3] || ""),
      raw: source
    };
  }

  const m = source.match(/^(\d{4}[-/]\d{2}[-/]\d{2}[ T]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?)\s*(.*)$/);
  if (m) {
    const ts = m[1];
    const rest = String(m[2] || "").trim();
    const level = detectLevel(rest);
    return { ts, level, message: rest || source, raw: source };
  }
  return { ts: "", level: detectLevel(source), message: source, raw: source };
}

function parseXrayLogLine(line: string): ParsedLogRow {
  const m = line.match(/^(\d{4}\/\d{2}\/\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[([^\]]+)\]\s*(.*)$/);
  if (m) {
    return {
      ts: String(m[1] || ""),
      level: normalizeLogLevel(String(m[2] || "")),
      message: String(m[3] || ""),
      raw: line
    };
  }
  return { ts: "", level: detectLevel(line), message: line, raw: line };
}

function unwrapLogArrayLine(line: string): string {
  let text = String(line || "").trim();
  if (!text || text === "[" || text === "]") return "";
  if (text.endsWith(",")) {
    text = text.slice(0, -1).trim();
  }
  if (!text) return "";
  if (text.startsWith('"') && text.endsWith('"')) {
    try {
      const parsed = JSON.parse(text);
      if (typeof parsed === "string") {
        return parsed.trim();
      }
    } catch {
      // keep raw fallback
    }
  }
  return text;
}

function parsePanelLogRows(text: string): ParsedLogRow[] {
  const source = String(text || "").trim();
  if (!source) return [];
  try {
    const parsed = JSON.parse(source) as unknown;
    if (Array.isArray(parsed)) {
      return parsed
        .filter((item) => typeof item === "string")
        .map((item) => parsePanelLogLine(String(item)))
        .filter((row) => row.raw || row.message);
    }
  } catch {
    // fallback to line parsing
  }
  return splitLogLines(source)
    .map((line) => parsePanelLogLine(line))
    .filter((row) => row.raw || row.message);
}

function formatLogTime(input: string): string {
  const raw = String(input || "").trim();
  if (!raw) return "-";

  // Panel format: 2026/02/23 19:25:16 (treated as local time).
  const panel = raw.match(/^(\d{4})\/(\d{2})\/(\d{2})\s+(\d{2}):(\d{2}):(\d{2})$/);
  if (panel) {
    const dt = new Date(
      Number(panel[1]),
      Number(panel[2]) - 1,
      Number(panel[3]),
      Number(panel[4]),
      Number(panel[5]),
      Number(panel[6])
    );
    if (!Number.isNaN(dt.getTime())) {
      return dt.toLocaleString();
    }
  }

  // Xray format is usually ISO-8601 (e.g. 2025-09-26T12:20:16.14159Z).
  const iso = new Date(raw);
  if (!Number.isNaN(iso.getTime())) {
    return iso.toLocaleString();
  }

  return raw;
}

function toXrayTableRowFromObject(obj: Record<string, unknown>): ParsedXrayLogTableRow {
  const dateTime = String(obj.DateTime ?? obj.dateTime ?? "");
  const fromAddress = String(obj.FromAddress ?? obj.fromAddress ?? "");
  const toAddress = String(obj.ToAddress ?? obj.toAddress ?? "");
  const inbound = String(obj.Inbound ?? obj.inbound ?? "");
  const outbound = String(obj.Outbound ?? obj.outbound ?? "");
  const email = String(obj.Email ?? obj.email ?? "");
  return {
    dateTime,
    fromAddress,
    toAddress,
    inbound,
    outbound,
    email,
    raw: JSON.stringify(obj)
  };
}

function extractJsonObjectBlocks(text: string): string[] {
  const blocks: string[] = [];
  let depth = 0;
  let start = -1;
  for (let i = 0; i < text.length; i += 1) {
    const ch = text[i];
    if (ch === "{") {
      if (depth === 0) start = i;
      depth += 1;
    } else if (ch === "}") {
      if (depth > 0) {
        depth -= 1;
        if (depth === 0 && start >= 0) {
          blocks.push(text.slice(start, i + 1));
          start = -1;
        }
      }
    }
  }
  return blocks;
}

function parseXrayLogTableRows(text: string): ParsedXrayLogTableRow[] {
  const source = String(text || "").trim();
  if (!source) return [];

  // Best case: whole payload is valid JSON.
  try {
    const parsed = JSON.parse(source) as unknown;
    if (Array.isArray(parsed)) {
      return parsed
        .filter((x) => x && typeof x === "object")
        .map((x) => toXrayTableRowFromObject(x as Record<string, unknown>));
    }
    if (parsed && typeof parsed === "object") {
      return [toXrayTableRowFromObject(parsed as Record<string, unknown>)];
    }
  } catch {
    // Continue with block extraction / line parsing fallback.
  }

  // Extract and parse embedded JSON object blocks.
  const rowsFromBlocks: ParsedXrayLogTableRow[] = [];
  for (const block of extractJsonObjectBlocks(source)) {
    try {
      const parsed = JSON.parse(block) as Record<string, unknown>;
      rowsFromBlocks.push(toXrayTableRowFromObject(parsed));
    } catch {
      // Ignore malformed block.
    }
  }
  if (rowsFromBlocks.length > 0) {
    return rowsFromBlocks;
  }

  // Fallback to plain text xray log lines.
  return splitLogLines(source).map((line) => {
    const parsed = parseXrayLogLine(line);
    return {
      dateTime: parsed.ts,
      fromAddress: "",
      toAddress: "",
      inbound: "",
      outbound: "",
      email: "",
      raw: parsed.raw
    };
  });
}

function formatLogMessage(input: string): string {
  const msg = unwrapLogArrayLine(String(input || "")).replace(/\t/g, "  ").trim();
  if (!msg) return "-";
  return msg;
}

function logLevelBadgeClass(level: string): string {
  const l = normalizeLogLevel(level);
  if (l === "debug") return "bg-slate-100 text-slate-700";
  if (l === "info") return "bg-sky-100 text-sky-700";
  if (l === "notice") return "bg-indigo-100 text-indigo-700";
  if (l === "warning") return "bg-amber-100 text-amber-800";
  if (l === "error") return "bg-rose-100 text-rose-700";
  return "bg-muted text-muted-foreground";
}

function sizeFormat(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 2)} ${units[i]}`;
}

function formatCpuFrequency(value?: number): string {
  const raw = Number(value || 0);
  if (!Number.isFinite(raw) || raw <= 0) return "-";
  if (raw < 20) return `${raw.toFixed(2)} GHz`;
  if (raw >= 1000) return `${(raw / 1000).toFixed(2)} GHz`;
  return `${Math.round(raw)} MHz`;
}

function formatSeconds(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds) || totalSeconds <= 0) return "0s";
  const s = Math.floor(totalSeconds);
  const days = Math.floor(s / 86400);
  const hours = Math.floor((s % 86400) / 3600);
  const mins = Math.floor((s % 3600) / 60);
  const secs = s % 60;
  return `${days}d ${hours}h ${mins}m ${secs}s`;
}

function formatUnixTime(ts?: number): string {
  if (!ts || ts <= 0) return "-";
  try {
    return new Date(ts * 1000).toLocaleString();
  } catch {
    return "-";
  }
}

async function getJson<T = unknown>(path: string): Promise<ApiResponse<T>> {
  const res = await fetch(`${basePath}${path}`, {
    method: "GET",
    credentials: "same-origin"
  });
  return (await res.json()) as ApiResponse<T>;
}

async function postJson<T = unknown>(path: string, body?: unknown): Promise<ApiResponse<T>> {
  const res = await fetch(`${basePath}${path}`, {
    method: "POST",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json"
    },
    body: body === undefined ? "{}" : JSON.stringify(body)
  });
  return (await res.json()) as ApiResponse<T>;
}

async function postForm<T = unknown>(path: string, form: Record<string, string>): Promise<ApiResponse<T>> {
  const body = new URLSearchParams();
  for (const [k, v] of Object.entries(form)) {
    body.append(k, v);
  }
  const res = await fetch(`${basePath}${path}`, {
    method: "POST",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded"
    },
    body: body.toString()
  });
  return (await res.json()) as ApiResponse<T>;
}

async function loadStatus(showLoading = true): Promise<void> {
  if (statusRequestInFlight) {
    return;
  }
  statusRequestInFlight = true;
  if (showLoading) {
    loading.value = true;
  }
  error.value = "";
  try {
    const msg = await getJson<ServerStatus>("panel/api/server/status");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load server status");
    }
    status.value = msg.obj;
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    statusRequestInFlight = false;
    if (showLoading) {
      loading.value = false;
    }
  }
}

function stopAutoRefresh(): void {
  if (refreshTimer != null) {
    window.clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

function startAutoRefresh(): void {
  stopAutoRefresh();
  refreshTimer = window.setInterval(() => {
    if (document.visibilityState !== "visible") {
      return;
    }
    if (!busy.value) {
      void loadStatus(false);
    }
  }, 1000);
}

async function performAction(action: () => Promise<void>): Promise<void> {
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    await action();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    busy.value = false;
  }
}

async function restartXrayService(): Promise<void> {
  await performAction(async () => {
    const msg = await postJson("panel/api/server/restartXrayService");
    if (!msg.success) {
      throw new Error(msg.msg || "Restart xray failed");
    }
    success.value = "Xray restart requested";
    await loadStatus();
  });
}

async function stopXrayService(): Promise<void> {
  await performAction(async () => {
    const msg = await postJson("panel/api/server/stopXrayService");
    if (!msg.success) {
      throw new Error(msg.msg || "Stop xray failed");
    }
    success.value = "Xray stop requested";
    await loadStatus();
  });
}

async function restartPanel(): Promise<void> {
  if (!(await openConfirmModal("Restart panel now?", "Restart Panel"))) {
    return;
  }
  await performAction(async () => {
    const msg = await postJson("panel/setting/restartPanel");
    if (!msg.success) {
      throw new Error(msg.msg || "Panel restart failed");
    }
    success.value = "Panel restart requested";
  });
}

async function openConfig(): Promise<void> {
  await performAction(async () => {
    const msg = await getJson<unknown>("panel/api/server/getConfigJson");
    if (!msg.success) {
      throw new Error(msg.msg || "Load config failed");
    }
    configJson.value = JSON.stringify(msg.obj, null, 2);
    configModalOpen.value = true;
  });
}

function openBackupDownload(): void {
  window.location.href = `${basePath}panel/api/server/getDb`;
}

async function openPanelLogs(): Promise<void> {
  await performAction(async () => {
    const msg = await postForm<unknown>(`panel/api/server/logs/${panelLogRows.value}`, {
      level: panelLogLevel.value,
      syslog: panelLogSyslog.value ? "true" : "false"
    });
    if (!msg.success) {
      throw new Error(msg.msg || "Load panel logs failed");
    }
    logsType.value = "panel";
    logsTitle.value = "Panel Logs";
    logsContent.value = typeof msg.obj === "string" ? msg.obj : JSON.stringify(msg.obj, null, 2);
    logsModalOpen.value = true;
  });
}

async function openXrayLogs(): Promise<void> {
  await performAction(async () => {
    const msg = await postForm<unknown>(`panel/api/server/xraylogs/${xrayLogRows.value}`, {
      filter: xrayLogFilter.value.trim(),
      showDirect: xrayLogShowDirect.value ? "true" : "false",
      showBlocked: xrayLogShowBlocked.value ? "true" : "false",
      showProxy: xrayLogShowProxy.value ? "true" : "false"
    });
    if (!msg.success) {
      throw new Error(msg.msg || "Load xray logs failed");
    }
    logsType.value = "xray";
    logsTitle.value = "Xray Logs";
    logsContent.value = typeof msg.obj === "string" ? msg.obj : JSON.stringify(msg.obj, null, 2);
    logsModalOpen.value = true;
  });
}

async function installPanelUpdate(): Promise<void> {
  if (!(await openConfirmModal("Install latest panel update now?", "Install Update"))) {
    return;
  }
  await performAction(async () => {
    const msg = await postJson("panel/api/server/installPanel");
    if (!msg.success) {
      throw new Error(msg.msg || "Panel update failed");
    }
    const restartMsg = await postJson("panel/setting/restartPanel");
    if (!restartMsg.success) {
      throw new Error(restartMsg.msg || "Panel updated but restart failed");
    }
    success.value = "Panel updated and restart requested";
  });
}

async function openXrayVersionPicker(): Promise<void> {
  await performAction(async () => {
    const versions = await getJson<string[]>("panel/api/server/getXrayVersion");
    if (!versions.success || !Array.isArray(versions.obj) || versions.obj.length === 0) {
      throw new Error(versions.msg || "No xray versions available");
    }
    xrayVersions.value = versions.obj;
    selectedXrayVersion.value = versions.obj[0] || "";
    xrayVersionModalOpen.value = true;
  });
}

async function installSelectedXrayVersion(): Promise<void> {
  if (!selectedXrayVersion.value) {
    error.value = "Please select an Xray version";
    return;
  }
  await performAction(async () => {
    const msg = await postJson(`panel/api/server/installXray/${encodeURIComponent(selectedXrayVersion.value)}`);
    if (!msg.success) {
      throw new Error(msg.msg || "Xray install failed");
    }
    success.value = `Xray install requested: ${selectedXrayVersion.value}`;
    xrayVersionModalOpen.value = false;
    await loadStatus();
  });
}

async function updateGeoFiles(): Promise<void> {
  await performAction(async () => {
    const msg = await postJson("panel/api/server/updateGeoFiles");
    if (!msg.success) {
      throw new Error(msg.msg || "Geo update failed");
    }
    success.value = "Geo files update requested";
  });
}

async function backupToTelegram(): Promise<void> {
  await performAction(async () => {
    const msg = await getJson("panel/api/backuptotgbot");
    if (!msg.success) {
      throw new Error(msg.msg || "Telegram backup trigger failed");
    }
    success.value = "Telegram backup requested";
  });
}

async function onXrayActionMenu(action: string): Promise<void> {
  if (!action) return;
  if (action === "restartXray") {
    await restartXrayService();
    return;
  }
  if (action === "stopXray") {
    await stopXrayService();
    return;
  }
  if (action === "installVersion") {
    await openXrayVersionPicker();
    return;
  }
  if (action === "updateGeo") {
    await updateGeoFiles();
  }
}

async function onPanelActionMenu(action: string): Promise<void> {
  if (!action) return;
  if (action === "viewConfig") {
    await openConfig();
    return;
  }
  if (action === "downloadBackup") {
    openBackupDownload();
    return;
  }
  if (action === "importDb") {
    triggerImportDb();
    return;
  }
  if (action === "backupTg") {
    await backupToTelegram();
    return;
  }
  if (action === "panelLogs") {
    await openPanelLogs();
    return;
  }
  if (action === "xrayLogs") {
    await openXrayLogs();
    return;
  }
  if (action === "installUpdate") {
    await installPanelUpdate();
    return;
  }
  if (action === "restartPanel") {
    await restartPanel();
  }
}

function triggerImportDb(): void {
  importDbInput.value?.click();
}

async function onImportDbChange(event: Event): Promise<void> {
  const target = event.target as HTMLInputElement | null;
  const file = target?.files?.[0];
  if (!file) return;
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const form = new FormData();
    form.append("db", file);
    const res = await fetch(`${basePath}panel/api/server/importDB`, {
      method: "POST",
      credentials: "same-origin",
      body: form
    });
    const msg = (await res.json()) as ApiResponse<unknown>;
    if (!msg.success) {
      throw new Error(msg.msg || "DB import failed");
    }
    const restartMsg = await postJson("panel/setting/restartPanel");
    if (!restartMsg.success) {
      throw new Error(restartMsg.msg || "DB imported but panel restart failed");
    }
    success.value = "DB import completed, panel restart requested";
    await loadStatus();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "DB import failed";
  } finally {
    busy.value = false;
    if (target) target.value = "";
  }
}

function downloadLogsFile(): void {
  const blob = new Blob([logsContent.value || ""], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = logsType.value === "xray" ? "xray.log" : "tx-ui.log";
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

const memText = computed(() => {
  if (!status.value?.mem) return "-";
  return `${sizeFormat(status.value.mem.current || 0)} / ${sizeFormat(status.value.mem.total || 0)}`;
});

const swapText = computed(() => {
  if (!status.value?.swap) return "-";
  return `${sizeFormat(status.value.swap.current || 0)} / ${sizeFormat(status.value.swap.total || 0)}`;
});

const diskText = computed(() => {
  if (!status.value?.disk) return "-";
  return `${sizeFormat(status.value.disk.current || 0)} / ${sizeFormat(status.value.disk.total || 0)}`;
});

const loadText = computed(() => {
  const loads = status.value?.loads || [];
  if (loads.length < 3) return "-";
  const n0 = Number(loads[0] || 0).toFixed(2);
  const n1 = Number(loads[1] || 0).toFixed(2);
  const n2 = Number(loads[2] || 0).toFixed(2);
  return `${n0} | ${n1} | ${n2}`;
});

const cpuPercent = computed(() => Math.round(Number(status.value?.cpu || 0)));
const memPercent = computed(() => {
  const current = Number(status.value?.mem?.current || 0);
  const total = Number(status.value?.mem?.total || 0);
  if (!total) return 0;
  return Math.round((current / total) * 10000) / 100;
});
const swapPercent = computed(() => {
  const current = Number(status.value?.swap?.current || 0);
  const total = Number(status.value?.swap?.total || 0);
  if (!total) return 0;
  return Math.round((current / total) * 10000) / 100;
});
const diskPercent = computed(() => {
  const current = Number(status.value?.disk?.current || 0);
  const total = Number(status.value?.disk?.total || 0);
  if (!total) return 0;
  return Math.round((current / total) * 10000) / 100;
});

const floatingNotice = computed(() => {
  if (error.value) return { type: "error" as const, text: error.value };
  if (success.value) return { type: "success" as const, text: success.value };
  return null;
});

watch([success, error], ([nextSuccess, nextError]) => {
  const hasNotice = Boolean(nextSuccess || nextError);
  if (!hasNotice) {
    if (notifyTimer !== null) {
      window.clearTimeout(notifyTimer);
      notifyTimer = null;
    }
    return;
  }
  if (notifyTimer !== null) {
    window.clearTimeout(notifyTimer);
  }
  notifyTimer = window.setTimeout(() => {
    success.value = "";
    error.value = "";
    notifyTimer = null;
  }, 3500);
});

function openConfirmModal(message: string, title = "Confirm"): Promise<boolean> {
  confirmModalMessage.value = message;
  confirmModalTitle.value = title;
  confirmModalOpen.value = true;
  return new Promise((resolve) => {
    confirmResolver = resolve;
  });
}

function resolveConfirmModal(value: boolean): void {
  confirmModalOpen.value = false;
  const resolver = confirmResolver;
  confirmResolver = null;
  resolver?.(value);
}
const highlightedConfigJsonHtml = computed(() => highlightJsonForHtml(configJson.value || "{}"));
const panelLogTableRows = computed(() => parsePanelLogRows(logsContent.value));
const xrayLogTableRows = computed(() => parseXrayLogTableRows(logsContent.value));

onMounted(async () => {
  void loadI18n([
    "pages.index.title",
    "pages.index.memory",
    "pages.index.hard",
    "pages.index.xrayStatus",
    "pages.index.systemLoadDesc",
    "pages.index.systemHost",
    "pages.index.operationHours",
    "pages.index.systemLoad",
    "pages.index.connectionCount",
    "pages.index.upSpeed",
    "pages.index.downSpeed",
    "pages.index.totalSent",
    "pages.index.totalReceive",
    "pages.index.geoData",
    "pages.index.geoVersion",
    "pages.index.geoSize",
    "pages.index.geoUpdated",
    "pages.index.exportDatabase",
    "pages.index.importDatabase",
    "pages.index.backup",
    "pages.dashboard.xrayDesc",
    "pages.dashboard.panelRuntimeDesc",
    "pages.dashboard.panelRuntime",
    "pages.dashboard.panelLogs",
    "pages.dashboard.xrayLogs",
    "pages.settings.title",
    "pages.xray.cpu",
    "pages.xray.logical",
    "pages.xray.threads",
    "pages.index.logs",
    "pages.index.config",
    "pages.settings.actions",
    "pages.index.restartXray",
    "pages.index.stopXray",
    "pages.index.xraySwitch",
    "pages.index.geoUpdate",
    "pages.settings.restartPanel",
    "close",
    "usage",
    "traffic",
    "none",
    "download",
    "filter",
    "status",
    "email",
    "inbound",
    "outbound"
  ]);
  await loadStatus();
  startAutoRefresh();
});

onUnmounted(() => {
  stopAutoRefresh();
  if (notifyTimer !== null) {
    window.clearTimeout(notifyTimer);
    notifyTimer = null;
  }
  if (confirmResolver) {
    confirmResolver(false);
    confirmResolver = null;
  }
});
</script>

<template>
  <div class="space-y-4">
    <UiCard>
      <UiCardHeader>
        <div class="flex items-center justify-between gap-3">
          <div>
            <UiCardTitle>{{ t("pages.index.title", "System Overview") }}</UiCardTitle>
            <UiCardDescription>{{ t("pages.index.systemLoadDesc", "Real-time server resources, network activity, and service health") }}</UiCardDescription>
          </div>
        </div>
      </UiCardHeader>
      <UiCardContent>
        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <UiCard>
            <UiCardHeader>
              <UiCardTitle>CPU</UiCardTitle>
              <UiCardDescription>{{ t("pages.xray.cpu", "CPU") }}: {{ status?.cpuCores ?? "-" }} | {{ t("pages.xray.logical", "Logical") }}: {{ status?.logicalPro ?? "-" }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="text-3xl font-semibold">{{ cpuPercent }}%</div>
              <div class="text-sm text-muted-foreground">{{ formatCpuFrequency(status?.cpuSpeedMhz) }}</div>
            </UiCardContent>
          </UiCard>

          <UiCard>
            <UiCardHeader>
              <UiCardTitle>{{ t("pages.index.memory", "Memory") }}</UiCardTitle>
              <UiCardDescription>{{ t("usage", "Usage") }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="text-3xl font-semibold">{{ memPercent }}%</div>
              <div class="text-sm text-muted-foreground">{{ memText }}</div>
            </UiCardContent>
          </UiCard>

          <UiCard>
            <UiCardHeader>
              <UiCardTitle>Swap</UiCardTitle>
              <UiCardDescription>{{ t("usage", "Usage") }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="text-3xl font-semibold">{{ swapPercent }}%</div>
              <div class="text-sm text-muted-foreground">{{ swapText }}</div>
            </UiCardContent>
          </UiCard>

          <UiCard>
            <UiCardHeader>
              <UiCardTitle>{{ t("pages.index.hard", "Disk") }}</UiCardTitle>
              <UiCardDescription>{{ t("usage", "Usage") }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="text-3xl font-semibold">{{ diskPercent }}%</div>
              <div class="text-sm text-muted-foreground">{{ diskText }}</div>
            </UiCardContent>
          </UiCard>
        </div>

        <div class="mt-4 grid gap-4 md:grid-cols-2">
          <UiCard>
            <UiCardHeader>
              <UiCardTitle>{{ t("pages.index.xrayStatus", "Xray Service") }}</UiCardTitle>
              <UiCardDescription>{{ t("pages.dashboard.xrayDesc", "Service state, version, and operational controls") }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="text-sm text-muted-foreground">{{ t("status", "Status") }}: {{ status?.xray?.state || t("none", "Unknown") }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.xraySwitch", "Version") }}: {{ status?.xray?.version || "-" }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.systemHost", "Host") }}: {{ status?.hostname || "-" }}</div>
              <div class="text-sm text-muted-foreground">IPv4: {{ status?.publicIP?.ipv4 || "-" }}</div>
              <div class="text-sm text-muted-foreground">IPv6: {{ status?.publicIP?.ipv6 || "-" }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.operationHours", "Uptime") }}: {{ formatSeconds(status?.appStats?.uptime || 0) }}</div>
              <div class="text-sm text-muted-foreground">
                {{ t("pages.index.geoData", "Geo Data") }}: GeoIP {{ status?.geoFiles?.geoip?.version || "-" }} | GeoSite {{ status?.geoFiles?.geosite?.version || "-" }}
              </div>
              <div class="text-sm text-muted-foreground">
                GeoIP {{ t("pages.index.geoSize", "Size") }}/{{ t("pages.index.geoUpdated", "Updated") }}: {{ sizeFormat(status?.geoFiles?.geoip?.size || 0) }} / {{ formatUnixTime(status?.geoFiles?.geoip?.updatedAt) }}
              </div>
              <div class="text-sm text-muted-foreground">
                GeoSite {{ t("pages.index.geoSize", "Size") }}/{{ t("pages.index.geoUpdated", "Updated") }}: {{ sizeFormat(status?.geoFiles?.geosite?.size || 0) }} / {{ formatUnixTime(status?.geoFiles?.geosite?.updatedAt) }}
              </div>
              <div class="mt-3">
                <details class="relative inline-block">
                  <summary class="action-trigger">
                    <MoreVertical class="h-4 w-4" />
                    {{ t("pages.settings.actions", "Actions") }}
                  </summary>
                  <div class="absolute left-0 bottom-full z-20 mb-1 max-h-[60vh] min-w-[220px] overflow-y-auto rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onXrayActionMenu('restartXray')">{{ t("pages.index.restartXray", "Restart Xray") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onXrayActionMenu('stopXray')">{{ t("pages.index.stopXray", "Stop Xray") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onXrayActionMenu('installVersion')">{{ t("pages.index.xraySwitch", "Install Xray Version") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onXrayActionMenu('updateGeo')">{{ t("pages.index.geoUpdate", "Update Geo Files") }}</button>
                  </div>
                </details>
              </div>
              <pre v-if="status?.xray?.errorMsg" class="mt-3 max-h-32 overflow-auto whitespace-pre-wrap rounded-md border bg-muted/30 p-2 text-xs">{{ status?.xray?.errorMsg }}</pre>
            </UiCardContent>
          </UiCard>

          <UiCard>
            <UiCardHeader>
              <UiCardTitle>{{ t("pages.dashboard.panelRuntime", "Panel Runtime") }}</UiCardTitle>
              <UiCardDescription>{{ t("pages.dashboard.panelRuntimeDesc", "Runtime metrics, backups, logs, and maintenance actions") }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.operationHours", "System uptime") }}: {{ formatSeconds(status?.uptime || 0) }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.systemLoad", "System load") }}: {{ loadText }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.memory", "Panel memory") }}: {{ sizeFormat(status?.appStats?.mem || 0) }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.xray.threads", "Panel threads") }}: {{ status?.appStats?.threads ?? 0 }}</div>
              <div class="text-sm text-muted-foreground">TCP / UDP: {{ status?.tcpCount ?? 0 }} / {{ status?.udpCount ?? 0 }}</div>
              <div class="text-sm text-muted-foreground">{{ t("pages.index.connectionCount", "Net IO") }}: {{ t("pages.index.upSpeed", "Up") }} {{ sizeFormat(status?.netIO?.up || 0) }}/s, {{ t("pages.index.downSpeed", "Down") }} {{ sizeFormat(status?.netIO?.down || 0) }}/s</div>
              <div class="text-sm text-muted-foreground">{{ t("traffic", "Net Traffic") }}: {{ t("pages.index.totalSent", "Out") }} {{ sizeFormat(status?.netTraffic?.sent || 0) }}, {{ t("pages.index.totalReceive", "In") }} {{ sizeFormat(status?.netTraffic?.recv || 0) }}</div>
              <div class="mt-3">
                <details class="relative inline-block">
                  <summary class="action-trigger">
                    <MoreVertical class="h-4 w-4" />
                    {{ t("pages.settings.actions", "Operations") }}
                  </summary>
                  <div class="absolute left-0 bottom-full z-20 mb-1 max-h-[60vh] min-w-[260px] overflow-y-auto rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onPanelActionMenu('viewConfig')">{{ t("pages.index.config", "Config") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onPanelActionMenu('downloadBackup')">{{ t("pages.index.exportDatabase", "Download DB Backup") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onPanelActionMenu('importDb')">{{ t("pages.index.importDatabase", "Import DB") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onPanelActionMenu('backupTg')">{{ t("pages.index.backup", "Backup To Telegram") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onPanelActionMenu('panelLogs')">{{ t("pages.dashboard.panelLogs", "Panel Logs") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onPanelActionMenu('xrayLogs')">{{ t("pages.dashboard.xrayLogs", "Xray Logs") }}</button>
                    <button
                      data-menu-action="1"
                      v-if="status?.appStats?.tx_update && status?.appStats?.tx_update !== status?.appStats?.panel_version"
                      class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted"
                      :disabled="busy"
                      @click="onPanelActionMenu('installUpdate')"
                    >
                      Install Panel Update ({{ status?.appStats?.tx_update }})
                    </button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm text-red-700 hover:bg-red-50" :disabled="busy" @click="onPanelActionMenu('restartPanel')">{{ t("pages.settings.restartPanel", "Restart Panel") }}</button>
                  </div>
                </details>
              </div>
            </UiCardContent>
          </UiCard>
        </div>
      </UiCardContent>
    </UiCard>

    <div v-if="configModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="max-h-[92vh] w-[min(96vw,1700px)] max-w-none overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ t("pages.index.config", "Current Config JSON") }}</h2>
          <UiButton size="sm" variant="outline" @click="configModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
        <pre
          class="json-codeblock h-[560px] overflow-auto whitespace-pre-wrap break-words p-3 text-xs"
          v-html="highlightedConfigJsonHtml"
        />
      </div>
    </div>
    <input ref="importDbInput" class="hidden" type="file" accept=".db,application/octet-stream" @change="onImportDbChange" />

    <div v-if="logsModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-2 sm:p-4">
      <div class="ltr-ui max-h-[92vh] w-[min(96vw,1900px)] max-w-none overflow-y-auto rounded-xl border bg-card p-3 shadow-lg sm:p-4">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ logsTitle }}</h2>
          <UiButton size="sm" variant="outline" @click="logsModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
        <div class="mb-3 grid gap-2 rounded-md border bg-muted/30 p-3 text-sm md:grid-cols-2">
          <template v-if="logsType === 'panel'">
            <label class="flex items-center gap-2">
              <span class="min-w-16 text-muted-foreground">{{ t("row", "Rows") }}</span>
              <select v-model="panelLogRows" class="w-full rounded-md border bg-background px-2 py-1">
                <option value="10">10</option>
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="500">500</option>
              </select>
            </label>
            <label class="flex items-center gap-2">
              <span class="min-w-16 text-muted-foreground">{{ t("pages.xray.logLevel", "Level") }}</span>
              <select v-model="panelLogLevel" class="w-full rounded-md border bg-background px-2 py-1">
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="notice">Notice</option>
                <option value="warning">Warning</option>
                <option value="err">Error</option>
              </select>
            </label>
            <label class="flex items-center gap-2 md:col-span-2">
              <input v-model="panelLogSyslog" type="checkbox" />
              <span>Include SysLog</span>
            </label>
          </template>
          <template v-else>
            <label class="flex items-center gap-2">
              <span class="min-w-16 text-muted-foreground">{{ t("row", "Rows") }}</span>
              <select v-model="xrayLogRows" class="w-full rounded-md border bg-background px-2 py-1">
                <option value="10">10</option>
                <option value="20">20</option>
                <option value="50">50</option>
                <option value="100">100</option>
                <option value="500">500</option>
              </select>
            </label>
            <label class="flex items-center gap-2">
              <span class="min-w-16 text-muted-foreground">{{ t("filter", "Filter") }}</span>
              <input v-model="xrayLogFilter" class="w-full rounded-md border bg-background px-2 py-1" type="text" />
            </label>
            <div class="flex items-center gap-4 md:col-span-2">
              <label class="flex items-center gap-2"><input v-model="xrayLogShowDirect" type="checkbox" />Direct</label>
              <label class="flex items-center gap-2"><input v-model="xrayLogShowBlocked" type="checkbox" />Blocked</label>
              <label class="flex items-center gap-2"><input v-model="xrayLogShowProxy" type="checkbox" />Proxy</label>
            </div>
          </template>
          <div class="flex gap-2 md:col-span-2">
            <UiButton size="sm" variant="outline" @click="logsType === 'xray' ? openXrayLogs() : openPanelLogs()">{{ t("reload", "Reload") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="downloadLogsFile">{{ t("download", "Download") }}</UiButton>
          </div>
        </div>
        <div v-if="logsType === 'xray'" class="overflow-x-auto rounded-md border">
          <table class="tx-table-center w-full min-w-[1320px] border-collapse text-xs">
            <thead>
              <tr class="border-b bg-muted/30 text-left text-muted-foreground">
                <th class="px-3 py-2 font-medium">DateTime</th>
                <th class="px-3 py-2 font-medium">FromAddress</th>
                <th class="px-3 py-2 font-medium">ToAddress</th>
                <th class="px-3 py-2 font-medium">{{ t("inbound", "Inbound") }}</th>
                <th class="px-3 py-2 font-medium">{{ t("outbound", "Outbound") }}</th>
                <th class="px-3 py-2 font-medium">{{ t("email", "Email") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in xrayLogTableRows" :key="`xlog-${idx}`" class="border-b align-top">
                <td class="whitespace-nowrap px-3 py-2 font-mono text-[11px]">{{ formatLogTime(row.dateTime) }}</td>
                <td class="px-3 py-2 font-mono text-[11px]">{{ row.fromAddress || "-" }}</td>
                <td class="px-3 py-2 font-mono text-[11px]">{{ row.toAddress || "-" }}</td>
                <td class="px-3 py-2 font-mono text-[11px]">{{ row.inbound || "-" }}</td>
                <td class="px-3 py-2 font-mono text-[11px]">{{ row.outbound || "-" }}</td>
                <td class="px-3 py-2 font-mono text-[11px]">{{ row.email || "-" }}</td>
              </tr>
              <tr v-if="xrayLogTableRows.length === 0">
                <td class="px-3 py-6 text-center text-muted-foreground" colspan="6">No Xray logs</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="overflow-x-auto rounded-md border">
          <table class="tx-table-center w-full min-w-[980px] border-collapse text-xs">
            <thead>
              <tr class="border-b bg-muted/30 text-left text-muted-foreground">
                <th class="px-3 py-2 font-medium">Time</th>
                <th class="px-3 py-2 font-medium">{{ t("pages.xray.logLevel", "Level") }}</th>
                <th class="px-3 py-2 font-medium">Message</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in panelLogTableRows" :key="`plog-${idx}`" class="border-b align-top">
                <td class="whitespace-nowrap px-3 py-2 font-mono text-[11px]">{{ formatLogTime(row.ts) }}</td>
                <td class="px-3 py-2">
                  <span :class="['inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium', logLevelBadgeClass(row.level)]">
                    {{ row.level ? row.level.toUpperCase() : "-" }}
                  </span>
                </td>
                <td class="px-3 py-2 font-mono text-[11px] whitespace-pre-wrap break-words">{{ formatLogMessage(row.message || row.raw) }}</td>
              </tr>
              <tr v-if="panelLogTableRows.length === 0">
                <td class="px-3 py-6 text-center text-muted-foreground" colspan="3">No panel logs</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-if="xrayVersionModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div class="w-full max-w-md rounded-xl border bg-card p-4 shadow-lg">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">Select Xray Version</h2>
          <UiButton size="sm" variant="outline" @click="xrayVersionModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
        <select v-model="selectedXrayVersion" class="w-full rounded-md border bg-background px-3 py-2 text-sm">
          <option v-for="version in xrayVersions" :key="version" :value="version">{{ version }}</option>
        </select>
        <div class="mt-3 flex justify-end gap-2">
          <UiButton size="sm" variant="outline" @click="xrayVersionModalOpen = false">{{ t("cancel", "Cancel") }}</UiButton>
          <UiButton :disabled="busy || !selectedXrayVersion" size="sm" @click="installSelectedXrayVersion">{{ t("install", "Install") }}</UiButton>
        </div>
      </div>
    </div>
    <div
      v-if="floatingNotice"
      class="pointer-events-none fixed right-4 top-4 z-[90] max-w-md rounded-lg border px-4 py-3 text-sm shadow-2xl backdrop-blur-sm"
      :class="floatingNotice.type === 'error' ? 'border-red-300 bg-red-50/95 text-red-700' : 'border-emerald-300 bg-emerald-50/95 text-emerald-700'"
    >
      {{ floatingNotice.text }}
    </div>
    <div v-if="confirmModalOpen" class="fixed inset-0 z-[95] flex items-center justify-center bg-black/50 p-4">
      <div class="w-full max-w-md rounded-xl border bg-card p-4 shadow-2xl">
        <h2 class="text-base font-semibold">{{ confirmModalTitle }}</h2>
        <p class="mt-2 text-sm text-muted-foreground">{{ confirmModalMessage }}</p>
        <div class="mt-4 flex justify-end gap-2">
          <UiButton size="sm" variant="outline" @click="resolveConfirmModal(false)">{{ t("cancel", "Cancel") }}</UiButton>
          <UiButton size="sm" @click="resolveConfirmModal(true)">{{ t("confirm", "Confirm") }}</UiButton>
        </div>
      </div>
    </div>
  </div>
</template>
