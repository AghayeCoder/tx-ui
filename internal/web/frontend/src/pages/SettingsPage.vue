<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { RefreshCw } from "lucide-vue-next";
import { UiButton } from "@/components/ui/button";
import { UiCard, UiCardContent, UiCardDescription, UiCardHeader, UiCardTitle } from "@/components/ui/card";
import { applyLocaleDirection, t, loadI18n } from "@/lib/i18n";

type ApiResponse<T> = {
  success: boolean;
  obj: T;
  msg?: string;
};

type UserSecret = {
  loginSecret?: string;
};

type SettingsForm = {
  webListen: string;
  webPort: number;
  webBasePath: string;
  webDomain: string;
  webCertFile: string;
  webKeyFile: string;
  sessionMaxAge: number;
  pageSize: number;
  expireDiff: number;
  trafficDiff: number;
  xrayCronJob: number;
  autoDeleteDay: number;
  timeLocation: string;
  datepicker: string;
  remarkModel: string;
  tgBotEnable: boolean;
  tgBotToken: string;
  tgBotProxy: string;
  tgBotAPIServer: string;
  tgBotChatId: string;
  tgRunTime: string;
  tgBotBackup: boolean;
  tgBotLoginNotify: boolean;
  tgCpu: number;
  tgLang: string;
  xrayTemplateConfig: string;
  secretEnable: boolean;
  subEnable: boolean;
  subTitle: string;
  subSupportUrl: string;
  subProfileUrl: string;
  subAnnounce: string;
  subListen: string;
  subPath: string;
  subPort: number;
  subDomain: string;
  subCertFile: string;
  subKeyFile: string;
  subUpdates: number;
  syncClients: boolean;
  subURI: string;
  subJsonURI: string;
  subJsonPath: string;
  subJsonFragment: string;
  subJsonNoises: string;
  subJsonMux: string;
  subJsonRules: string;
  subEncrypt: boolean;
  subCustomUI: boolean;
  subShowInfo: boolean;
};

const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";

const loading = ref(false);
const busy = ref(false);
const success = ref("");
const error = ref("");
const confirmModalOpen = ref(false);
const confirmModalTitle = ref(t("confirm", "Confirm"));
const confirmModalMessage = ref("");
type SettingsTab = "panel" | "subscription" | "telegram" | "security" | "json";
const activeTab = ref<SettingsTab>("panel");
const selectedWebLang = ref("en_US");

const allSettingRaw = ref("{}");
const defaultSettingRaw = ref("{}");
const allSettingObj = ref<Record<string, unknown>>({});

const oldUsername = ref("");
const oldPassword = ref("");
const newUsername = ref("");
const newPassword = ref("");
const loginSecret = ref("");

const form = reactive<SettingsForm>({
  webListen: "",
  webPort: 54321,
  webBasePath: "/",
  webDomain: "",
  webCertFile: "",
  webKeyFile: "",
  sessionMaxAge: 60,
  pageSize: 50,
  expireDiff: 3,
  trafficDiff: 5,
  xrayCronJob: 10,
  autoDeleteDay: 30,
  timeLocation: "UTC",
  datepicker: "gregorian",
  remarkModel: "-ieo",
  tgBotEnable: false,
  tgBotToken: "",
  tgBotProxy: "",
  tgBotAPIServer: "",
  tgBotChatId: "",
  tgRunTime: "0 0 * * *",
  tgBotBackup: false,
  tgBotLoginNotify: false,
  tgCpu: 80,
  tgLang: "en_US",
  xrayTemplateConfig: "",
  secretEnable: false,
  subEnable: false,
  subTitle: "",
  subSupportUrl: "",
  subProfileUrl: "",
  subAnnounce: "",
  subListen: "",
  subPath: "/sub/",
  subPort: 2096,
  subDomain: "",
  subCertFile: "",
  subKeyFile: "",
  subUpdates: 6,
  syncClients: false,
  subURI: "",
  subJsonURI: "",
  subJsonPath: "/json/",
  subJsonFragment: "",
  subJsonNoises: "",
  subJsonMux: "",
  subJsonRules: "",
  subEncrypt: false,
  subCustomUI: false,
  subShowInfo: false
});
let notifyTimer: number | null = null;
let confirmResolver: ((value: boolean) => void) | null = null;

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

function openConfirmModal(message: string, title = t("confirm", "Confirm")): Promise<boolean> {
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

type NoiseItem = {
  type: string;
  packet: string;
  delay: string;
  applyTo?: string;
};

const defaultFragment = {
  packets: "tlshello",
  length: "100-200",
  interval: "10-20",
  maxSplit: "300-400"
};

const defaultNoises: NoiseItem[] = [{ type: "rand", packet: "10-20", delay: "10-16", applyTo: "ip" }];
const defaultMux = { enabled: true, concurrency: 8, xudpConcurrency: 16, xudpProxyUDP443: "reject" };
const defaultRules = [
  { type: "field", outboundTag: "direct", domain: ["geosite:category-ir"] },
  { type: "field", outboundTag: "direct", ip: ["geoip:private", "geoip:ir"] }
];

const directIPsOptions = [
  "geoip:private",
  "geoip:ir",
  "geoip:cn",
  "geoip:ru",
  "geoip:vn",
  "geoip:es",
  "geoip:id",
  "geoip:ua",
  "geoip:tr",
  "geoip:br"
];

const directDomainsOptions = [
  "geosite:private",
  "geosite:category-ir",
  "geosite:cn",
  "geosite:category-ru",
  "geosite:apple",
  "geosite:meta",
  "geosite:google"
];

const webLangOptions = [
  { value: "en_US", label: "English", emoji: "🇺🇸" },
  { value: "fa_IR", label: "Persian", emoji: "🇮🇷" },
  { value: "ar_EG", label: "Arabic", emoji: "🇪🇬" },
  { value: "es_ES", label: "Spanish", emoji: "🇪🇸" },
  { value: "id_ID", label: "Indonesian", emoji: "🇮🇩" },
  { value: "ja_JP", label: "Japanese", emoji: "🇯🇵" },
  { value: "pt_BR", label: "Portuguese (BR)", emoji: "🇧🇷" },
  { value: "ru_RU", label: "Russian", emoji: "🇷🇺" },
  { value: "tr_TR", label: "Turkish", emoji: "🇹🇷" },
  { value: "uk_UA", label: "Ukrainian", emoji: "🇺🇦" },
  { value: "vi_VN", label: "Vietnamese", emoji: "🇻🇳" },
  { value: "zh_CN", label: "Chinese (Simplified)", emoji: "🇨🇳" },
  { value: "zh_TW", label: "Chinese (Traditional)", emoji: "🇹🇼" }
] as const;
const tgLangOptions = webLangOptions;

const remarkModelOptions = [
  { value: "i", label: "Inbound" },
  { value: "e", label: "Email" },
  { value: "o", label: "Other" }
] as const;
const remarkSeparatorOptions = [" ", "-", "_", "@", ":", "~", "|", ",", ".", "/"] as const;
const remarkModelParts = ref<string[]>(["i", "e", "o"]);
const remarkSeparator = ref("-");

const fragmentEnabled = ref(false);
const noisesEnabled = ref(false);
const muxEnabled = ref(false);
const directEnabled = ref(false);
const fragmentPackets = ref(defaultFragment.packets);
const fragmentLength = ref(defaultFragment.length);
const fragmentInterval = ref(defaultFragment.interval);
const fragmentMaxSplit = ref(defaultFragment.maxSplit);
const noisesArray = ref<NoiseItem[]>([]);
const muxConcurrency = ref<number>(defaultMux.concurrency);
const muxXudpConcurrency = ref<number>(defaultMux.xudpConcurrency);
const muxXudpProxyUDP443 = ref<string>(defaultMux.xudpProxyUDP443);
const directIPsInput = ref("");
const directDomainsInput = ref("");

function parseJSONString<T>(text: string, fallback: T): T {
  if (!text || !text.trim()) return fallback;
  try {
    return JSON.parse(text) as T;
  } catch {
    return fallback;
  }
}

function arrayToCsv(items: string[]): string {
  return items.join(", ");
}

function csvToArray(text: string): string[] {
  return text
    .split(",")
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
}

function hydrateSubJsonEditors(): void {
  const fragment = parseJSONString<Record<string, string>>(form.subJsonFragment, {} as Record<string, string>);
  fragmentEnabled.value = Boolean(form.subJsonFragment && form.subJsonFragment.trim());
  fragmentPackets.value = fragment.packets || defaultFragment.packets;
  fragmentLength.value = fragment.length || defaultFragment.length;
  fragmentInterval.value = fragment.interval || defaultFragment.interval;
  fragmentMaxSplit.value = fragment.maxSplit || defaultFragment.maxSplit;

  noisesEnabled.value = Boolean(form.subJsonNoises && form.subJsonNoises.trim());
  const parsedNoises = parseJSONString<NoiseItem[]>(form.subJsonNoises, []);
  noisesArray.value = parsedNoises.length > 0 ? parsedNoises : [...defaultNoises];

  muxEnabled.value = Boolean(form.subJsonMux && form.subJsonMux.trim());
  const mux = parseJSONString<Record<string, unknown>>(form.subJsonMux, defaultMux);
  muxConcurrency.value = Number(mux.concurrency ?? defaultMux.concurrency);
  muxXudpConcurrency.value = Number(mux.xudpConcurrency ?? defaultMux.xudpConcurrency);
  muxXudpProxyUDP443.value = String(mux.xudpProxyUDP443 ?? defaultMux.xudpProxyUDP443);

  directEnabled.value = Boolean(form.subJsonRules && form.subJsonRules.trim());
  const rules = parseJSONString<Array<Record<string, unknown>>>(form.subJsonRules, []);
  const ipRule = rules.find((r) => Array.isArray(r.ip));
  const domainRule = rules.find((r) => Array.isArray(r.domain));
  directIPsInput.value = arrayToCsv((ipRule?.ip as string[] | undefined) || []);
  directDomainsInput.value = arrayToCsv((domainRule?.domain as string[] | undefined) || []);
}

function syncSubJsonEditorsToForm(): void {
  if (fragmentEnabled.value) {
    form.subJsonFragment = JSON.stringify(
      {
        packets: fragmentPackets.value || defaultFragment.packets,
        length: fragmentLength.value || defaultFragment.length,
        interval: fragmentInterval.value || defaultFragment.interval,
        maxSplit: fragmentMaxSplit.value || defaultFragment.maxSplit
      },
      null,
      2
    );
  } else {
    form.subJsonFragment = "";
  }

  if (noisesEnabled.value) {
    form.subJsonNoises = JSON.stringify(noisesArray.value.length > 0 ? noisesArray.value : defaultNoises, null, 2);
  } else {
    form.subJsonNoises = "";
  }

  if (muxEnabled.value) {
    form.subJsonMux = JSON.stringify(
      {
        enabled: true,
        concurrency: Number.isFinite(muxConcurrency.value) ? muxConcurrency.value : defaultMux.concurrency,
        xudpConcurrency: Number.isFinite(muxXudpConcurrency.value) ? muxXudpConcurrency.value : defaultMux.xudpConcurrency,
        xudpProxyUDP443: muxXudpProxyUDP443.value || defaultMux.xudpProxyUDP443
      },
      null,
      2
    );
  } else {
    form.subJsonMux = "";
  }

  if (directEnabled.value) {
    const domain = csvToArray(directDomainsInput.value);
    const ip = csvToArray(directIPsInput.value);
    const rules: Array<Record<string, unknown>> = [];
    if (domain.length > 0) {
      rules.push({ ...defaultRules[0], domain });
    }
    if (ip.length > 0) {
      rules.push({ ...defaultRules[1], ip });
    }
    form.subJsonRules = rules.length > 0 ? JSON.stringify(rules, null, 2) : "";
  } else {
    form.subJsonRules = "";
  }
}

function addNoise(): void {
  noisesArray.value = [...noisesArray.value, { type: "rand", packet: "10-20", delay: "10-16", applyTo: "ip" }];
}

function removeNoise(index: number): void {
  const next = [...noisesArray.value];
  next.splice(index, 1);
  noisesArray.value = next;
}

function addDirectIPToken(token: string): void {
  const items = new Set(csvToArray(directIPsInput.value));
  items.add(token);
  directIPsInput.value = arrayToCsv([...items]);
}

function addDirectDomainToken(token: string): void {
  const items = new Set(csvToArray(directDomainsInput.value));
  items.add(token);
  directDomainsInput.value = arrayToCsv([...items]);
}

function safePrettyJSON(value: unknown): string {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return "{}";
  }
}

function isValidJSON(text: string): boolean {
  try {
    JSON.parse(text);
    return true;
  } catch {
    return false;
  }
}

function randomSecret(length = 64): string {
  const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
  let out = "";
  for (let i = 0; i < length; i += 1) {
    out += chars[Math.floor(Math.random() * chars.length)];
  }
  return out;
}

function parseRemarkModel(value: string): void {
  const model = String(value || "-ieo").replace(/\s+$/, "");
  if (!model) {
    remarkSeparator.value = "-";
    remarkModelParts.value = ["i", "e", "o"];
    return;
  }
  const first = model.charAt(0);
  const hasKnownSeparator = remarkSeparatorOptions.includes(first as (typeof remarkSeparatorOptions)[number]);
  remarkSeparator.value = hasKnownSeparator ? first : "-";
  const partsSource = hasKnownSeparator ? model.slice(1) : model;
  const parts = partsSource
    .split("")
    .filter((x): x is "i" | "e" | "o" => x === "i" || x === "e" || x === "o");
  remarkModelParts.value = parts.length > 0 ? parts : ["i", "e", "o"];
}

function buildRemarkModel(): void {
  const uniqueOrdered = Array.from(new Set(remarkModelParts.value.filter((x) => x === "i" || x === "e" || x === "o")));
  const parts = uniqueOrdered.length > 0 ? uniqueOrdered : ["i", "e", "o"];
  const sep = remarkSeparator.value || "-";
  form.remarkModel = `${sep}${parts.join("")}`;
}

function toggleRemarkModelPart(part: "i" | "e" | "o"): void {
  if (remarkModelParts.value.includes(part)) {
    const next = remarkModelParts.value.filter((x) => x !== part);
    // Keep at least one part selected.
    if (next.length === 0) return;
    remarkModelParts.value = next;
  } else {
    remarkModelParts.value = [...remarkModelParts.value, part];
  }
  buildRemarkModel();
}

function readCookie(name: string): string {
  const needle = `${name}=`;
  const parts = document.cookie.split(";");
  for (const part of parts) {
    const item = part.trim();
    if (item.startsWith(needle)) {
      return decodeURIComponent(item.substring(needle.length));
    }
  }
  return "";
}

function applyWebLanguage(): void {
  const nextLang = selectedWebLang.value || "en_US";
  document.cookie = `lang=${encodeURIComponent(nextLang)}; path=/; max-age=31536000; samesite=lax`;
  applyLocaleDirection(nextLang);
  success.value = `Web language changed to ${nextLang}. Reloading...`;
  setTimeout(() => {
    window.location.reload();
  }, 150);
}

async function onSecretToggleChange(): Promise<void> {
  if (form.secretEnable) {
    if (!loginSecret.value.trim()) {
      loginSecret.value = randomSecret();
    }
    return;
  }
  const previous = loginSecret.value;
  loginSecret.value = "";
  busy.value = true;
  error.value = "";
  try {
    const msg = await postJson("panel/setting/updateUserSecret", {
      loginSecret: ""
    });
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to clear login secret");
    }
    success.value = t("pages.settings.ui.loginSecretCleared", "Login secret cleared");
  } catch (e) {
    loginSecret.value = previous;
    form.secretEnable = true;
    error.value = e instanceof Error ? e.message : "Failed to clear login secret";
  } finally {
    busy.value = false;
  }
}

const configAlerts = computed(() => {
  const alerts: string[] = [];
  if (window.location.protocol !== "https:") {
    alerts.push("Panel is not running on HTTPS.");
  }
  if (form.webPort === 2053) {
    alerts.push("Panel port is set to 2053.");
  }
  const pathSegments = window.location.pathname.split("/").filter(Boolean);
  if (pathSegments.length < 3 && form.webBasePath.trim() === "/") {
    alerts.push("Panel base path is `/` on a short URL path.");
  }
  if (form.secretEnable && !loginSecret.value.trim()) {
    alerts.push("Secret login is enabled but login secret is empty.");
  }
  if (form.subEnable) {
    try {
      const subPath = form.subURI.trim().length > 0 ? new URL(form.subURI).pathname : form.subPath;
      if (subPath === "/sub/") {
        alerts.push("Subscription path is still default `/sub/`.");
      }
    } catch {
      alerts.push("Subscription URI is invalid.");
    }
    try {
      const subJsonPath = form.subJsonURI.trim().length > 0 ? new URL(form.subJsonURI).pathname : form.subJsonPath;
      if (subJsonPath === "/json/") {
        alerts.push("Subscription JSON path is still default `/json/`.");
      }
    } catch {
      alerts.push("Subscription JSON URI is invalid.");
    }
  }
  return alerts;
});

const remarkSample = computed(() => {
  const map: Record<string, string> = { i: "Inbound", e: "Email", o: "Other" };
  const labels = remarkModelParts.value.map((x) => map[x]).filter(Boolean);
  return labels.length > 0 ? labels.join(remarkSeparator.value || "-") : "";
});

function syncFormFromObj(obj: Record<string, unknown>): void {
  form.webListen = String(obj.webListen ?? form.webListen);
  form.webPort = Number(obj.webPort ?? form.webPort);
  form.webBasePath = String(obj.webBasePath ?? form.webBasePath);
  form.webDomain = String(obj.webDomain ?? "");
  form.webCertFile = String(obj.webCertFile ?? "");
  form.webKeyFile = String(obj.webKeyFile ?? "");
  form.sessionMaxAge = Number(obj.sessionMaxAge ?? form.sessionMaxAge);
  form.pageSize = Number(obj.pageSize ?? form.pageSize);
  form.expireDiff = Number(obj.expireDiff ?? form.expireDiff);
  form.trafficDiff = Number(obj.trafficDiff ?? form.trafficDiff);
  form.xrayCronJob = Number(obj.xrayCronJob ?? form.xrayCronJob);
  form.autoDeleteDay = Number(obj.autoDeleteDay ?? form.autoDeleteDay);
  form.timeLocation = String(obj.timeLocation ?? form.timeLocation);
  form.datepicker = String(obj.datepicker ?? form.datepicker);
  form.remarkModel = String(obj.remarkModel ?? form.remarkModel);
  parseRemarkModel(form.remarkModel);
  form.tgBotEnable = Boolean(obj.tgBotEnable ?? form.tgBotEnable);
  form.tgBotToken = String(obj.tgBotToken ?? "");
  form.tgBotProxy = String(obj.tgBotProxy ?? "");
  form.tgBotAPIServer = String(obj.tgBotAPIServer ?? "");
  form.tgBotChatId = String(obj.tgBotChatId ?? "");
  form.tgRunTime = String(obj.tgRunTime ?? form.tgRunTime);
  form.tgBotBackup = Boolean(obj.tgBotBackup ?? form.tgBotBackup);
  form.tgBotLoginNotify = Boolean(obj.tgBotLoginNotify ?? form.tgBotLoginNotify);
  form.tgCpu = Number(obj.tgCpu ?? form.tgCpu);
  form.tgLang = String(obj.tgLang ?? form.tgLang);
  form.xrayTemplateConfig = typeof obj.xrayTemplateConfig === "string" ? String(obj.xrayTemplateConfig) : safePrettyJSON(obj.xrayTemplateConfig ?? {});
  form.secretEnable = Boolean(obj.secretEnable ?? form.secretEnable);
  form.subEnable = Boolean(obj.subEnable ?? form.subEnable);
  form.subTitle = String(obj.subTitle ?? "");
  form.subSupportUrl = String(obj.subSupportUrl ?? "");
  form.subProfileUrl = String(obj.subProfileUrl ?? "");
  form.subAnnounce = String(obj.subAnnounce ?? "");
  form.subListen = String(obj.subListen ?? "");
  form.subPath = String(obj.subPath ?? form.subPath);
  form.subPort = Number(obj.subPort ?? form.subPort);
  form.subDomain = String(obj.subDomain ?? "");
  form.subCertFile = String(obj.subCertFile ?? "");
  form.subKeyFile = String(obj.subKeyFile ?? "");
  form.subUpdates = Number(obj.subUpdates ?? form.subUpdates);
  form.syncClients = Boolean(obj.syncClients ?? form.syncClients);
  form.subURI = String(obj.subURI ?? "");
  form.subJsonURI = String(obj.subJsonURI ?? "");
  form.subJsonPath = String(obj.subJsonPath ?? form.subJsonPath);
  form.subJsonFragment = String(obj.subJsonFragment ?? "");
  form.subJsonNoises = String(obj.subJsonNoises ?? "");
  form.subJsonMux = String(obj.subJsonMux ?? "");
  form.subJsonRules = String(obj.subJsonRules ?? "");
  form.subEncrypt = Boolean(obj.subEncrypt ?? form.subEncrypt);
  form.subCustomUI = Boolean(obj.subCustomUI ?? form.subCustomUI);
  form.subShowInfo = Boolean(obj.subShowInfo ?? form.subShowInfo);
  hydrateSubJsonEditors();
}

function applyFormToObj(): void {
  syncSubJsonEditorsToForm();
  // Reflect raw remark model edits in preview controls without rewriting the raw value.
  parseRemarkModel(form.remarkModel);
  const next = { ...allSettingObj.value };
  next.webListen = form.webListen;
  next.webPort = form.webPort;
  next.webBasePath = form.webBasePath;
  next.webDomain = form.webDomain;
  next.webCertFile = form.webCertFile;
  next.webKeyFile = form.webKeyFile;
  next.sessionMaxAge = form.sessionMaxAge;
  next.pageSize = form.pageSize;
  next.expireDiff = form.expireDiff;
  next.trafficDiff = form.trafficDiff;
  next.xrayCronJob = form.xrayCronJob;
  next.autoDeleteDay = form.autoDeleteDay;
  next.timeLocation = form.timeLocation;
  next.datepicker = form.datepicker;
  next.remarkModel = form.remarkModel;
  next.tgBotEnable = form.tgBotEnable;
  next.tgBotToken = form.tgBotToken;
  next.tgBotProxy = form.tgBotProxy;
  next.tgBotAPIServer = form.tgBotAPIServer;
  next.tgBotChatId = form.tgBotChatId;
  next.tgRunTime = form.tgRunTime;
  next.tgBotBackup = form.tgBotBackup;
  next.tgBotLoginNotify = form.tgBotLoginNotify;
  next.tgCpu = form.tgCpu;
  next.tgLang = form.tgLang;
  next.xrayTemplateConfig = form.xrayTemplateConfig;
  next.secretEnable = form.secretEnable;
  next.subEnable = form.subEnable;
  next.subTitle = form.subTitle;
  next.subSupportUrl = form.subSupportUrl;
  next.subProfileUrl = form.subProfileUrl;
  next.subAnnounce = form.subAnnounce;
  next.subListen = form.subListen;
  next.subPath = form.subPath;
  next.subPort = form.subPort;
  next.subDomain = form.subDomain;
  next.subCertFile = form.subCertFile;
  next.subKeyFile = form.subKeyFile;
  next.subUpdates = form.subUpdates;
  next.syncClients = form.syncClients;
  next.subURI = form.subURI;
  next.subJsonURI = form.subJsonURI;
  next.subJsonPath = form.subJsonPath;
  next.subJsonFragment = form.subJsonFragment;
  next.subJsonNoises = form.subJsonNoises;
  next.subJsonMux = form.subJsonMux;
  next.subJsonRules = form.subJsonRules;
  next.subEncrypt = form.subEncrypt;
  next.subCustomUI = form.subCustomUI;
  next.subShowInfo = form.subShowInfo;
  allSettingObj.value = next;
  allSettingRaw.value = safePrettyJSON(next);
}

function applyRawToForm(): void {
  if (!isValidJSON(allSettingRaw.value)) {
  error.value = t("pages.settings.ui.settingsJsonInvalid", "Settings JSON is invalid");
    return;
  }
  const parsed = JSON.parse(allSettingRaw.value) as Record<string, unknown>;
  allSettingObj.value = parsed;
  syncFormFromObj(parsed);
  success.value = t("pages.settings.ui.formUpdatedFromJson", "Form updated from JSON");
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

async function loadAllSettings(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const msg = await postJson<Record<string, unknown>>("panel/setting/all");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load settings");
    }
    allSettingObj.value = msg.obj || {};
    allSettingRaw.value = safePrettyJSON(allSettingObj.value);
    syncFormFromObj(allSettingObj.value);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    loading.value = false;
  }
}

async function loadDefaultSettings(): Promise<void> {
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await postJson<Record<string, unknown>>("panel/setting/defaultSettings");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load default settings");
    }
    defaultSettingRaw.value = safePrettyJSON(msg.obj);
    success.value = t("pages.settings.ui.defaultSettingsLoaded", "Default settings loaded");
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    busy.value = false;
  }
}

function applyDefaultsToEditor(): void {
  if (!defaultSettingRaw.value.trim()) return;
  allSettingRaw.value = defaultSettingRaw.value;
  applyRawToForm();
  success.value = t("pages.settings.ui.defaultSettingsCopied", "Default settings copied to editor");
}

async function saveAllSettings(): Promise<void> {
  error.value = "";
  success.value = "";
  applyFormToObj();

  busy.value = true;
  try {
    const msg = await postJson("panel/setting/update", allSettingObj.value);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to save settings");
    }
    success.value = t("pages.settings.ui.settingsSaved", "Settings saved");
    await loadAllSettings();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    busy.value = false;
  }
}

async function updateUserCredentials(): Promise<void> {
  error.value = "";
  success.value = "";
  if (!oldUsername.value || !oldPassword.value || !newUsername.value || !newPassword.value) {
    error.value = t("pages.settings.ui.allUserFieldsRequired", "All user fields are required");
    return;
  }

  busy.value = true;
  try {
    const msg = await postJson("panel/setting/updateUser", {
      oldUsername: oldUsername.value,
      oldPassword: oldPassword.value,
      newUsername: newUsername.value,
      newPassword: newPassword.value
    });
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to update user");
    }
    success.value = t("pages.settings.ui.userCredentialsUpdated", "User credentials updated. Redirecting to login...");
    oldPassword.value = "";
    newPassword.value = "";
    setTimeout(() => {
      window.location.href = `${basePath}logout`;
    }, 400);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    busy.value = false;
  }
}

async function fetchUserSecret(): Promise<void> {
  busy.value = true;
  error.value = "";
  try {
    const msg = await postJson<UserSecret>("panel/setting/getUserSecret");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to fetch login secret");
    }
    loginSecret.value = msg.obj?.loginSecret || "";
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    busy.value = false;
  }
}

async function saveUserSecret(): Promise<void> {
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await postJson("panel/setting/updateUserSecret", {
      loginSecret: loginSecret.value
    });
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to update login secret");
    }
    success.value = t("pages.settings.ui.loginSecretUpdated", "Login secret updated");
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    busy.value = false;
  }
}

async function restartPanel(): Promise<void> {
  if (!(await openConfirmModal(t("pages.settings.ui.restartPanelNow", "Restart panel now?"), t("pages.settings.restartPanel", "Restart Panel")))) {
    return;
  }

  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await postJson("panel/setting/restartPanel");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to restart panel");
    }
    success.value = t("pages.settings.ui.restartRequested", "Restart requested. Redirecting to panel settings when service is back...");
    const targetProtocol = form.webCertFile.trim() || form.webKeyFile.trim() ? "https:" : window.location.protocol;
    const targetHost = form.webDomain.trim() || window.location.hostname;
    const targetPort = Number.isFinite(form.webPort) && form.webPort > 0 ? String(form.webPort) : window.location.port;
    const base = form.webBasePath.trim() || "/";
    const normalizedBase = base.startsWith("/") ? base : `/${base}`;
    const finalBase = normalizedBase.endsWith("/") ? normalizedBase : `${normalizedBase}/`;
    const portPart = targetPort ? `:${targetPort}` : "";
    const redirectUrl = `${targetProtocol}//${targetHost}${portPart}${finalBase}panel/settings`;
    setTimeout(() => {
      window.location.href = redirectUrl;
    }, 5000);
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    busy.value = false;
  }
}

onMounted(async () => {
  void loadI18n([
    "pages.settings.title",
    "pages.settings.actions",
    "pages.settings.panelSettings",
    "pages.settings.TGBotSettings",
    "pages.settings.subSettings",
    "pages.settings.securitySettings",
    "pages.settings.save",
    "pages.settings.restartPanel",
    "close",
    "loading"
  ]);
  const cookieLang = readCookie("lang");
  if (cookieLang && webLangOptions.some((x) => x.value === cookieLang)) {
    selectedWebLang.value = cookieLang;
  }
  await loadAllSettings();
  await fetchUserSecret();
});

onUnmounted(() => {
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
  <div class="settings-compact space-y-3">
    <UiCard>
      <UiCardHeader>
        <div class="flex flex-col items-start justify-between gap-2 sm:flex-row sm:items-center">
          <div>
            <UiCardTitle>{{ t("pages.settings.title", "Panel Settings") }}</UiCardTitle>
            <UiCardDescription>{{ t("pages.settings.infoDesc", "Configure web, subscription, Telegram, and security settings with JSON fallback.") }}</UiCardDescription>
          </div>
          <div class="flex w-full flex-wrap gap-2 sm:w-auto">
            <UiButton class="w-full sm:w-auto" :disabled="loading || busy" variant="outline" @click="loadAllSettings">
              <RefreshCw class="mr-2 h-4 w-4" />
              {{ t("reload", "Reload") }}
            </UiButton>
            <UiButton class="w-full sm:w-auto" :disabled="busy" variant="outline" @click="saveAllSettings">{{ t("pages.settings.save", "Save Settings") }}</UiButton>
            <UiButton class="w-full sm:w-auto border-red-300 text-red-700 hover:bg-red-50" :disabled="busy" variant="outline" @click="restartPanel">{{ t("pages.settings.restartPanel", "Restart Panel") }}</UiButton>
          </div>
        </div>
      </UiCardHeader>
      <UiCardContent>
        <div v-if="configAlerts.length > 0" class="mb-2 rounded-md border border-amber-300 bg-amber-50 p-2 text-sm text-amber-800">
          <div class="mb-1 font-semibold">{{ t("pages.settings.ui.configurationAlerts", "Configuration Alerts") }}</div>
          <ul class="list-disc pl-5">
            <li v-for="(alert, idx) in configAlerts" :key="idx">{{ alert }}</li>
          </ul>
        </div>

        <div class="mb-3 flex flex-wrap gap-2">
          <UiButton :disabled="busy" size="sm" :variant="activeTab === 'panel' ? 'default' : 'outline'" @click="activeTab = 'panel'">
            {{ t("pages.settings.panelSettings", "Panel") }}
          </UiButton>
          <UiButton :disabled="busy" size="sm" :variant="activeTab === 'subscription' ? 'default' : 'outline'" @click="activeTab = 'subscription'">
            {{ t("pages.settings.subSettings", "Subscription") }}
          </UiButton>
          <UiButton :disabled="busy" size="sm" :variant="activeTab === 'telegram' ? 'default' : 'outline'" @click="activeTab = 'telegram'">
            {{ t("pages.settings.TGBotSettings", "Telegram") }}
          </UiButton>
          <UiButton :disabled="busy" size="sm" :variant="activeTab === 'security' ? 'default' : 'outline'" @click="activeTab = 'security'">
            {{ t("pages.settings.securitySettings", "Security") }}
          </UiButton>
          <UiButton :disabled="busy" size="sm" :variant="activeTab === 'json' ? 'default' : 'outline'" @click="activeTab = 'json'">
            {{ t("pages.settings.ui.json", "JSON") }}
          </UiButton>
        </div>

        <div class="space-y-3 rounded-lg border p-2.5">
          <template v-if="activeTab === 'panel'">
            <h3 class="text-sm font-semibold">{{ t("pages.settings.panelSettings", "Core Panel") }}</h3>
            <div class="grid grid-cols-1 gap-1.5 md:grid-cols-2">
              <input v-model="form.webListen" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.panelListeningIP', 'Web listen')" type="text" />
              <input v-model.number="form.webPort" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.panelPort', 'Web port')" type="number" />
              <input v-model="form.webBasePath" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.panelUrlPath', 'Web base path')" type="text" />
              <input v-model="form.webDomain" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.panelListeningDomain', 'Web domain')" type="text" />
              <input v-model="form.webCertFile" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.publicKeyPath', 'Web cert file')" type="text" />
              <input v-model="form.webKeyFile" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.privateKeyPath', 'Web key file')" type="text" />
              <input v-model.number="form.sessionMaxAge" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.sessionMaxAge', 'Session max age')" type="number" />
              <input v-model.number="form.pageSize" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.pageSize', 'Page size')" type="number" />
              <input v-model="form.timeLocation" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.timeZone', 'Time location')" type="text" />
              <div class="flex items-center gap-2 md:col-span-2">
                <select v-model="selectedWebLang" class="min-w-0 flex-1 rounded-md border bg-background px-3 py-2 text-sm">
                  <option v-for="opt in webLangOptions" :key="opt.value" :value="opt.value">{{ opt.emoji }} {{ opt.label }} ({{ opt.value }})</option>
                </select>
                <UiButton :disabled="busy" size="sm" variant="outline" @click="applyWebLanguage">{{ t("pages.settings.ui.applyWebLanguage", "Apply Web Language") }}</UiButton>
              </div>
              <div class="grid grid-cols-1 gap-1.5 md:col-span-2 md:grid-cols-[88px_minmax(0,1fr)_minmax(220px,auto)] md:items-center">
                <select v-model="remarkSeparator" class="rounded-md border bg-background px-3 py-2 text-sm" @change="buildRemarkModel">
                  <option v-for="sep in remarkSeparatorOptions" :key="`sep-${sep}`" :value="sep">{{ sep === " " ? "[space]" : sep }}</option>
                </select>
                <div class="flex flex-wrap gap-1.5 rounded-md border bg-background px-2 py-1.5">
                  <button
                    v-for="opt in remarkModelOptions"
                    :key="opt.value"
                    type="button"
                    class="rounded-md border px-2 py-1 text-xs transition-colors"
                    :class="remarkModelParts.includes(opt.value) ? 'border-primary bg-primary text-primary-foreground' : 'border-border bg-background hover:bg-muted'"
                    @click="toggleRemarkModelPart(opt.value)"
                  >
                    {{ opt.label }}
                  </button>
                </div>
                <div class="flex items-center rounded-md border bg-muted/30 px-3 py-2 text-sm">
                  <span class="font-medium">{{ t("pages.settings.sampleRemark", "Remark sample:") }}</span>&nbsp;{{ remarkSample || "-" }}
                </div>
              </div>
            </div>

            <h3 class="pt-1 text-sm font-semibold">{{ t("pages.settings.ui.automationPolicies", "Automation / Policies") }}</h3>
            <div class="grid grid-cols-1 gap-1.5 md:grid-cols-2">
              <input v-model.number="form.expireDiff" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.expireTimeDiff', 'Expire diff (days)')" type="number" />
              <input v-model.number="form.trafficDiff" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.trafficDiff', 'Traffic diff (GB)')" type="number" />
              <input v-model.number="form.xrayCronJob" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.xrayCronJob', 'Xray cron (seconds)')" type="number" />
              <input v-model.number="form.autoDeleteDay" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.autoDeleteDay', 'Auto delete day')" type="number" />
              <select v-model="form.datepicker" class="rounded-md border bg-background px-3 py-2 text-sm">
                <option value="gregorian">gregorian</option>
                <option value="persian">persian</option>
              </select>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.secretEnable" type="checkbox" @change="onSecretToggleChange" /> {{ t("pages.settings.security.loginSecurity", "Secret Login") }}
              </label>
            </div>
          </template>

          <template v-else-if="activeTab === 'subscription'">
            <h3 class="pt-1 text-sm font-semibold">{{ t("pages.settings.subSettings", "Subscription") }}</h3>
            <div class="grid grid-cols-1 gap-1.5 md:grid-cols-2">
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
                <input v-model="form.subEnable" type="checkbox" /> {{ t("pages.settings.subEnable", "Enable Sub") }}
              </label>
              <input v-model="form.subTitle" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subTitle', 'Sub title')" type="text" />
              <input v-model="form.subDomain" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subDomain', 'Sub domain')" type="text" />
              <input v-model="form.subListen" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subListen', 'Sub listen')" type="text" />
              <input v-model.number="form.subPort" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subPort', 'Sub port')" type="number" />
              <input v-model="form.subPath" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subPath', 'Sub path')" type="text" />
              <input v-model="form.subURI" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subURI', 'Sub URI')" type="text" />
              <input v-model="form.subSupportUrl" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subSupportUrl', 'Sub support URL')" type="text" />
              <input v-model="form.subProfileUrl" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subProfileUrl', 'Sub profile URL')" type="text" />
              <input v-model="form.subAnnounce" class="rounded-md border bg-background px-3 py-2 text-sm md:col-span-2" :placeholder="t('pages.settings.subAnnounce', 'Sub announce')" type="text" />
              <input v-model="form.subJsonPath" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.subJsonPath', 'Sub json path')" type="text" />
              <input v-model="form.subJsonURI" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.subJsonURI', 'Sub JSON URI')" type="text" />
              <input v-model.number="form.subUpdates" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subUpdates', 'Sub updates')" type="number" />
              <input v-model="form.subCertFile" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subCertPath', 'Sub cert file')" type="text" />
              <input v-model="form.subKeyFile" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.subKeyPath', 'Sub key file')" type="text" />
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.subEncrypt" type="checkbox" /> {{ t("pages.settings.subEncrypt", "Sub Encrypt") }}
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.subCustomUI" type="checkbox" /> {{ t("pages.settings.subUI", "Sub Custom UI") }}
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.subShowInfo" type="checkbox" /> {{ t("pages.settings.subShowInfo", "Sub Show Info") }}
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.syncClients" type="checkbox" /> {{ t("pages.settings.clientSynchronization", "Sync Clients") }}
              </label>
            </div>

            <div class="space-y-2 rounded-md border p-2">
              <div class="flex flex-wrap items-center gap-2">
                <p class="text-sm font-semibold">{{ t("pages.settings.ui.structuredSubJsonHelpers", "Structured Sub JSON Helpers") }}</p>
              </div>

              <div class="space-y-1.5 rounded-md border p-1.5">
                <label class="flex items-center gap-2 text-sm font-medium">
                  <input v-model="fragmentEnabled" type="checkbox" /> Fragment
                </label>
                <div v-if="fragmentEnabled" class="grid grid-cols-1 gap-2 md:grid-cols-2">
                  <input v-model="fragmentPackets" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.fragmentPackets', 'packets (e.g. tlshello)')" type="text" />
                  <input v-model="fragmentLength" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.fragmentLength', 'length (e.g. 100-200)')" type="text" />
                  <input v-model="fragmentInterval" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.fragmentInterval', 'interval (e.g. 10-20)')" type="text" />
                  <input v-model="fragmentMaxSplit" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.fragmentMaxSplit', 'maxSplit (e.g. 300-400)')" type="text" />
                </div>
              </div>

              <div class="space-y-1.5 rounded-md border p-1.5">
                <label class="flex items-center gap-2 text-sm font-medium">
                  <input v-model="noisesEnabled" type="checkbox" /> Noises
                </label>
                <div v-if="noisesEnabled" class="space-y-1.5">
                  <div v-for="(noise, idx) in noisesArray" :key="`noise-${idx}`" class="grid grid-cols-1 gap-1.5 md:grid-cols-5">
                    <select v-model="noise.type" class="rounded-md border bg-background px-3 py-2 text-sm">
                      <option value="rand">rand</option>
                      <option value="str">str</option>
                      <option value="base64">base64</option>
                    </select>
                    <select v-model="noise.applyTo" class="rounded-md border bg-background px-3 py-2 text-sm">
                      <option value="ip">ip</option>
                      <option value="ipv4">ipv4</option>
                      <option value="ipv6">ipv6</option>
                    </select>
                    <input v-model="noise.packet" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.noisePacket', 'packet 5-10')" type="text" />
                    <input v-model="noise.delay" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.noiseDelay', 'delay 10-20')" type="text" />
                    <UiButton :disabled="noisesArray.length <= 1" class="border-red-300 text-red-700 hover:bg-red-50" size="sm" variant="outline" @click="removeNoise(idx)">{{ t("pages.settings.ui.remove", "Remove") }}</UiButton>
                  </div>
                  <UiButton size="sm" variant="outline" @click="addNoise">{{ t("pages.settings.ui.addNoise", "Add Noise") }}</UiButton>
                </div>
              </div>

              <div class="space-y-1.5 rounded-md border p-1.5">
                <label class="flex items-center gap-2 text-sm font-medium">
                  <input v-model="muxEnabled" type="checkbox" /> Mux
                </label>
                <div v-if="muxEnabled" class="grid grid-cols-1 gap-2 md:grid-cols-3">
                  <input v-model.number="muxConcurrency" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.concurrency', 'Concurrency')" type="number" />
                  <input v-model.number="muxXudpConcurrency" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.xudpConcurrency', 'XUDP Concurrency')" type="number" />
                  <select v-model="muxXudpProxyUDP443" class="rounded-md border bg-background px-3 py-2 text-sm">
                    <option value="reject">reject</option>
                    <option value="skip">skip</option>
                    <option value="allow">allow</option>
                  </select>
                </div>
              </div>

              <div class="space-y-1.5 rounded-md border p-1.5">
                <label class="flex items-center gap-2 text-sm font-medium">
                  <input v-model="directEnabled" type="checkbox" /> Direct Rules
                </label>
                <div v-if="directEnabled" class="space-y-1.5">
                  <input v-model="directDomainsInput" class="w-full rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.directDomainsComma', 'Direct domains (comma separated)')" type="text" />
                  <div class="flex flex-wrap gap-2">
                    <UiButton v-for="token in directDomainsOptions" :key="`dd-${token}`" size="sm" variant="outline" @click="addDirectDomainToken(token)">{{ token }}</UiButton>
                  </div>
                  <input v-model="directIPsInput" class="w-full rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.ui.directIpsComma', 'Direct IP rules (comma separated)')" type="text" />
                  <div class="flex flex-wrap gap-2">
                    <UiButton v-for="token in directIPsOptions" :key="`di-${token}`" size="sm" variant="outline" @click="addDirectIPToken(token)">{{ token }}</UiButton>
                  </div>
                </div>
              </div>

              <div class="grid grid-cols-1 gap-1.5 md:grid-cols-2">
                <textarea v-model="form.subJsonFragment" class="h-24 json-field px-3 py-2 text-xs" :placeholder="t('pages.settings.ui.rawSubJsonFragment', 'Raw subJsonFragment JSON')" />
                <textarea v-model="form.subJsonNoises" class="h-24 json-field px-3 py-2 text-xs" :placeholder="t('pages.settings.ui.rawSubJsonNoises', 'Raw subJsonNoises JSON')" />
                <textarea v-model="form.subJsonMux" class="h-24 json-field px-3 py-2 text-xs" :placeholder="t('pages.settings.ui.rawSubJsonMux', 'Raw subJsonMux JSON')" />
                <textarea v-model="form.subJsonRules" class="h-24 json-field px-3 py-2 text-xs" :placeholder="t('pages.settings.ui.rawSubJsonRules', 'Raw subJsonRules JSON')" />
              </div>
            </div>
          </template>

          <template v-else-if="activeTab === 'telegram'">
            <h3 class="pt-1 text-sm font-semibold">{{ t("pages.settings.TGBotSettings", "Telegram Bot") }}</h3>
            <div class="grid grid-cols-1 gap-1.5 md:grid-cols-2">
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
                  <input v-model="form.tgBotEnable" type="checkbox" /> {{ t("pages.settings.telegramBotEnable", "Enable Telegram Bot") }}
              </label>
              <input v-model="form.tgBotToken" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.telegramToken', 'Bot token')" type="text" />
              <input v-model="form.tgBotProxy" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.telegramProxy', 'Bot proxy')" type="text" />
              <input v-model="form.tgBotAPIServer" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.telegramAPIServer', 'Bot API server')" type="text" />
              <input v-model="form.tgBotChatId" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.telegramChatId', 'Chat ID')" type="text" />
              <input v-model="form.tgRunTime" class="rounded-md border bg-background px-3 py-2 text-sm md:col-span-2" :placeholder="t('pages.settings.telegramNotifyTime', 'Run time cron')" type="text" />
              <input v-model.number="form.tgCpu" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.tgNotifyCpu', 'CPU threshold')" type="number" />
              <select v-model="form.tgLang" class="rounded-md border bg-background px-3 py-2 text-sm">
                <option v-for="opt in tgLangOptions" :key="`tg-${opt.value}`" :value="opt.value">{{ opt.emoji }} {{ opt.label }} ({{ opt.value }})</option>
              </select>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.tgBotBackup" type="checkbox" /> {{ t("pages.settings.tgNotifyBackup", "TG Backup") }}
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="form.tgBotLoginNotify" type="checkbox" /> {{ t("pages.settings.tgNotifyLogin", "TG Login Notify") }}
              </label>
            </div>
          </template>

          <template v-else-if="activeTab === 'security'">
            <div class="space-y-3">
              <div class="rounded-lg border p-2.5">
                <h3 class="mb-2 text-sm font-semibold">{{ t("pages.settings.securitySettings", "User Credentials") }}</h3>
                <div class="grid grid-cols-1 gap-2">
                  <input v-model="oldUsername" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.oldUsername', 'Old username')" type="text" />
                  <input v-model="oldPassword" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.currentPassword', 'Old password')" type="password" />
                  <input v-model="newUsername" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.newUsername', 'New username')" type="text" />
                  <input v-model="newPassword" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.settings.newPassword', 'New password')" type="password" />
                </div>
                <div class="mt-2">
                  <UiButton :disabled="busy" size="sm" variant="outline" @click="updateUserCredentials">{{ t("pages.settings.ui.updateUser", "Update User") }}</UiButton>
                </div>
              </div>

              <div class="rounded-lg border p-2.5">
                <h3 class="mb-2 text-sm font-semibold">{{ t("secretToken", "Login Secret") }}</h3>
                <textarea v-model="loginSecret" class="h-20 w-full json-field px-3 py-2 text-xs" />
                <div class="mt-2 flex flex-wrap gap-2">
                  <UiButton :disabled="busy" size="sm" variant="outline" @click="loginSecret = randomSecret()">{{ t("pages.settings.ui.generateSecret", "Generate Secret") }}</UiButton>
                  <UiButton :disabled="busy" size="sm" variant="outline" @click="saveUserSecret">{{ t("pages.settings.ui.saveSecret", "Save Secret") }}</UiButton>
                </div>
              </div>

            </div>
          </template>

          <template v-else-if="activeTab === 'json'">
            <div class="space-y-3">
              <div class="rounded-lg border p-2.5">
                <h3 class="mb-2 text-sm font-semibold">{{ t("pages.settings.title", "Current Settings JSON") }}</h3>
                <textarea v-model="allSettingRaw" class="h-[300px] w-full json-field px-3 py-2 text-xs" />
                <div class="mt-2 flex flex-wrap gap-2">
                  <UiButton :disabled="busy" size="sm" variant="outline" @click="loadAllSettings">{{ t("pages.settings.ui.resetJson", "Reset JSON") }}</UiButton>
                  <UiButton :disabled="busy" size="sm" variant="outline" @click="loadDefaultSettings">{{ t("pages.settings.ui.loadDefaults", "Load Defaults") }}</UiButton>
                  <UiButton :disabled="busy || !defaultSettingRaw.trim()" size="sm" variant="outline" @click="applyDefaultsToEditor">{{ t("pages.settings.ui.useLoadedDefaults", "Use Loaded Defaults") }}</UiButton>
                </div>
              </div>
              <div v-if="defaultSettingRaw.trim()" class="rounded-lg border p-2.5">
                <h3 class="mb-2 text-sm font-semibold">{{ t("pages.settings.ui.defaultSettingsJson", "Default Settings JSON") }}</h3>
                <textarea v-model="defaultSettingRaw" class="h-48 w-full json-field px-3 py-2 text-xs" />
              </div>
            </div>
          </template>
        </div>

      </UiCardContent>
    </UiCard>
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

<style scoped>
.settings-compact :deep(input:not([type="checkbox"])),
.settings-compact :deep(select),
.settings-compact :deep(textarea) {
  min-height: 2rem;
  padding: 0.375rem 0.625rem;
}
</style>
