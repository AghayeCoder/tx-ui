<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { MoreVertical, RefreshCw } from "lucide-vue-next";
import { UiButton } from "@/components/ui/button";
import { UiCard, UiCardContent, UiCardDescription, UiCardHeader, UiCardTitle } from "@/components/ui/card";
import { t, loadI18n } from "@/lib/i18n";

type ApiResponse<T> = {
  success: boolean;
  obj: T;
  msg?: string;
};

type XrayPayload = {
  xraySetting: unknown;
};

type OutboundTraffic = {
  tag: string;
  up: number;
  down: number;
  total: number;
};

type RoutingRuleRow = {
  type: "field";
  outboundTag: string;
  balancerTag: string;
  domain: string;
  ip: string;
  port: string;
  network: string;
  source: string;
  sourcePort: string;
  user: string;
  protocol: string;
  inboundTag: string;
  attrs: string;
};

type FakeDnsRow = {
  ipPool: string;
  poolSize: number;
};

type OutboundFormRow = {
  raw: Record<string, unknown>;
  tag: string;
  protocol: string;
  sendThrough: string;
  domainStrategy: string;
};

type BalancerFormRow = {
  tag: string;
  selector: string;
  strategy: string;
};

type InboundListItem = {
  tag?: string;
};

const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";

const loading = ref(false);
const busy = ref(false);
const success = ref("");
const error = ref("");
const confirmModalOpen = ref(false);
const confirmModalTitle = ref(t("confirm", "Confirm"));
const confirmModalMessage = ref("");
const basicModalOpen = ref(false);
const dnsRoutingModalOpen = ref(false);
const addRoutingRuleModalOpen = ref(false);
const addOutboundModalOpen = ref(false);
const outboundsModalOpen = ref(false);
const warpModalOpen = ref(false);
const tunnelModalOpen = ref(false);

const xraySettingRaw = ref("{}");
const originalXraySettingRaw = ref("{}");
const xrayResult = ref("");
const outbounds = ref<OutboundTraffic[]>([]);
const dnsServersText = ref("");
const outboundTagOptions = ref<string[]>(["direct", "block"]);
const inboundTagOptions = ref<string[]>([]);
const routingRuleRows = ref<RoutingRuleRow[]>([]);
const routingExtraRules = ref<unknown[]>([]);
const dnsQueryStrategy = ref("");
const dnsClientIp = ref("");
const fakeDnsRows = ref<FakeDnsRow[]>([]);
const outboundRows = ref<OutboundFormRow[]>([]);
const balancerRows = ref<BalancerFormRow[]>([]);
const xrayImportInput = ref<HTMLInputElement | null>(null);
const addOutboundTab = ref<"form" | "json">("form");
const addOutboundLink = ref("");
const addOutboundJson = ref("{}");
const selectedInboundTagDraft = ref("");
const routingRuleModalMode = ref<"add" | "edit">("add");
const routingRuleEditIndex = ref(-1);
const editOutboundModalOpen = ref(false);
const editOutboundIndex = ref(-1);
const editOutboundJson = ref("{}");
const outboundDomainStrategies = ["AsIs", "UseIP", "UseIPv4", "UseIPv6"] as const;
const routingDomainStrategies = ["AsIs", "IPIfNonMatch", "IPOnDemand"] as const;
const logLevels = ["debug", "info", "warning", "error", "none"] as const;

const freedomStrategy = ref<(typeof outboundDomainStrategies)[number]>("AsIs");
const routingStrategy = ref<(typeof routingDomainStrategies)[number]>("AsIs");
const logLevel = ref<(typeof logLevels)[number]>("warning");
const accessLogPath = ref("");
const errorLogPath = ref("");
const maskAddressLog = ref("");
const dnsLogEnabled = ref(false);
const outboundTrafficEnabled = ref(false);
const blockedDomainCsv = ref("");
const blockedIpCsv = ref("");
const blockedProtocolCsv = ref("");
const directDomainCsv = ref("");
const directIpCsv = ref("");
const ipv4DomainCsv = ref("");
const warpDomainCsv = ref("");

const warp = reactive({
  privateKey: "",
  publicKey: "",
  license: "",
  data: "",
  config: ""
});

const tunnel = reactive({
  ip: "",
  port: "",
  user: "",
  password: ""
});
const addOutboundForm = reactive({
  protocol: "vless",
  tag: "",
  sendThrough: "",
  vmessVlessAddress: "",
  vmessVlessPort: "443",
  vmessVlessUuid: "",
  vmessEncryption: "",
  vlessFlow: "",
  trojanAddress: "",
  trojanPort: "443",
  trojanPassword: "",
  trojanFlow: "",
  ssAddress: "",
  ssPort: "443",
  ssMethod: "aes-128-gcm",
  ssPassword: "",
  socksAddress: "",
  socksPort: "1080",
  socksUsername: "",
  socksPassword: "",
  socksVersion: "5",
  httpAddress: "",
  httpPort: "8080",
  httpUsername: "",
  httpPassword: "",
  hysteriaAddress: "",
  hysteriaPort: "443",
  hysteriaAuth: "",
  hysteriaObfsType: "",
  hysteriaObfsPassword: "",
  hysteriaSni: "",
  hysteriaAlpnCsv: "",
  hysteriaUpMbps: "",
  hysteriaDownMbps: "",
  wireguardSecretKey: "",
  wireguardPublicKey: "",
  wireguardPreSharedKey: "",
  wireguardEndpointHost: "",
  wireguardEndpointPort: "51820",
  wireguardLocalAddressCsv: "172.16.0.2/32",
  wireguardAllowedIpsCsv: "0.0.0.0/0,::/0",
  wireguardMtu: "",
  dnsAddress: "8.8.8.8",
  dnsPort: "53",
  dnsNetwork: "",
  freedomDomainStrategy: "",
  freedomRedirect: "",
  blackholeResponseType: "none",
  transmission: "tcp",
  security: "none",
  tlsServerName: "",
  tlsAlpnCsv: "",
  tlsMinVersion: "",
  tlsMaxVersion: "",
  tlsCipherSuitesCsv: "",
  tlsFingerprint: "",
  tlsAllowInsecure: false,
  tlsRejectUnknownSni: false,
  tlsCertFile: "",
  tlsKeyFile: "",
  realityShow: false,
  realityServerName: "",
  realityFingerprint: "chrome",
  realityPublicKey: "",
  realityShortId: "",
  realitySpiderX: "/",
  realityMldsa65Verify: "",
  httpObfs: false,
  httpObfsHostCsv: "",
  httpObfsPathCsv: "/",
  udpMasks: "",
  sockopt: false,
  sockoptJson: "{}",
  mux: false,
  muxConcurrency: "",
  muxXudpConcurrency: "",
  muxXudpProxyUDP443: "",
  customSettingsJson: "{}",
  customStreamSettingsJson: "{}"
});
const routingRuleDraft = reactive<RoutingRuleRow>({
  type: "field",
  outboundTag: "direct",
  balancerTag: "",
  domain: "",
  ip: "",
  port: "",
  network: "",
  source: "",
  sourcePort: "",
  user: "",
  protocol: "",
  inboundTag: "",
  attrs: ""
});
let notifyTimer: number | null = null;
let confirmResolver: ((value: boolean) => void) | null = null;

function escapeHtml(input: string): string {
  return input.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
}

function highlightJsonForHtml(jsonText: string): string {
  const source = escapeHtml(jsonText);
  return source
    .replace(/("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*")(?=\s*:)/g, '<span class="json-key">$1</span>')
    .replace(/:\s*("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*")/g, ': <span class="json-string">$1</span>')
    .replace(/\b-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?\b/g, '<span class="json-number">$&</span>')
    .replace(/\b(true|false)\b/g, '<span class="json-boolean">$1</span>')
    .replace(/\bnull\b/g, '<span class="json-null">null</span>');
}

const highlightedJsonEditorHtml = computed(() => highlightJsonForHtml(xraySettingRaw.value || "{}"));
const addOutboundProtocolOptions = [
  "freedom",
  "blackhole",
  "dns",
  "vmess",
  "vless",
  "http",
  "trojan",
  "socks",
  "hysteria",
  "shadowsocks",
  "wireguard"
] as const;
const streamCapableProtocols = new Set(["vmess", "vless", "trojan", "http", "socks", "hysteria", "shadowsocks"]);
const selectedAddOutboundProtocol = computed(() => addOutboundForm.protocol.trim().toLowerCase() || "freedom");
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

function refreshOutboundTagOptions(extraTags: string[] = []): void {
  const fromRows = outboundRows.value.map((row) => String(row.tag || "").trim()).filter(Boolean);
  const fromTraffic = outbounds.value.map((row) => String(row.tag || "").trim()).filter(Boolean);
  const fromCurrent = outboundTagOptions.value.map((tag) => String(tag || "").trim()).filter(Boolean);
  outboundTagOptions.value = Array.from(new Set([...fromRows, ...fromTraffic, ...extraTags, ...fromCurrent, "direct", "block"]));
  if (!routingRuleDraft.outboundTag.trim()) {
    routingRuleDraft.outboundTag = outboundTagOptions.value[0] || "direct";
  }
}

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

function sizeFormat(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const idx = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, idx)).toFixed(idx === 0 ? 0 : 2)} ${units[idx]}`;
}

function outboundAddressLabel(row: OutboundFormRow): string {
  const raw = row.raw || {};
  const settings = (raw.settings && typeof raw.settings === "object") ? raw.settings as Record<string, unknown> : {};
  const protocol = (row.protocol || "").trim().toLowerCase();

  if ((protocol === "vmess" || protocol === "vless") && Array.isArray(settings.vnext) && settings.vnext.length > 0) {
    const first = settings.vnext[0];
    if (first && typeof first === "object") {
      const target = first as Record<string, unknown>;
      const address = typeof target.address === "string" ? target.address : "";
      const port = typeof target.port === "number" || typeof target.port === "string" ? String(target.port) : "";
      if (address && port) return `${address}:${port}`;
      if (address) return address;
    }
  }

  if ((protocol === "trojan" || protocol === "shadowsocks" || protocol === "socks" || protocol === "http" || protocol === "hysteria") && Array.isArray(settings.servers) && settings.servers.length > 0) {
    const first = settings.servers[0];
    if (first && typeof first === "object") {
      const target = first as Record<string, unknown>;
      const address = typeof target.address === "string" ? target.address : "";
      const port = typeof target.port === "number" || typeof target.port === "string" ? String(target.port) : "";
      if (address && port) return `${address}:${port}`;
      if (address) return address;
    }
  }

  if (protocol === "wireguard") {
    const endpoint = settings.endpoint;
    if (typeof endpoint === "string" && endpoint.trim()) return endpoint;
    if (Array.isArray(settings.peers) && settings.peers.length > 0) {
      const first = settings.peers[0];
      if (first && typeof first === "object") {
        const target = first as Record<string, unknown>;
        const endpointFirst = target.endpoint;
        if (typeof endpointFirst === "string" && endpointFirst.trim()) return endpointFirst;
      }
    }
  }

  if (protocol === "dns") {
    const address = typeof settings.address === "string" ? settings.address : "";
    const port = typeof settings.port === "number" || typeof settings.port === "string" ? String(settings.port) : "";
    if (address && port) return `${address}:${port}`;
    if (address) return address;
  }

  if (protocol === "freedom") {
    const redirect = typeof settings.redirect === "string" ? settings.redirect : "";
    if (redirect.trim()) return redirect;
  }

  return "-";
}

function outboundTrafficLabel(tag: string): string {
  const item = outbounds.value.find((entry) => entry.tag === tag);
  if (!item) return "0 B / 0 B";
  return `${sizeFormat(item.down || 0)} / ${sizeFormat(item.total || 0)}`;
}

function isValidJSON(text: string): boolean {
  try {
    JSON.parse(text);
    return true;
  } catch {
    return false;
  }
}

function downloadXrayJson(): void {
  const blob = new Blob([xraySettingRaw.value], { type: "application/json;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "xray-config.json";
  a.click();
  URL.revokeObjectURL(url);
}

function triggerImportXrayJson(): void {
  xrayImportInput.value?.click();
}

async function onImportXrayJson(event: Event): Promise<void> {
  const target = event.target as HTMLInputElement | null;
  const file = target?.files?.[0];
  if (!file) return;
  error.value = "";
  success.value = "";
  try {
    const content = await file.text();
    if (!isValidJSON(content)) {
      throw new Error(t("pages.xray.ui.selectedFileInvalidJson", "Selected file is not valid JSON"));
    }
    xraySettingRaw.value = JSON.stringify(JSON.parse(content), null, 2);
    syncStructuredFromJson();
    success.value = t("pages.xray.ui.xrayJsonLoaded", "Xray JSON loaded from file");
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.importXrayJsonFailed", "Failed to import Xray JSON");
  } finally {
    if (target) target.value = "";
  }
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

async function getJson<T = unknown>(path: string): Promise<ApiResponse<T>> {
  const res = await fetch(`${basePath}${path}`, {
    method: "GET",
    credentials: "same-origin"
  });
  return (await res.json()) as ApiResponse<T>;
}

async function loadXraySetting(): Promise<void> {
  loading.value = true;
  error.value = "";
  try {
    const msg = await postJson<string>("panel/xray/");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load xray config");
    }
    const parsed = JSON.parse(msg.obj) as XrayPayload;
    xraySettingRaw.value = JSON.stringify(parsed.xraySetting, null, 2);
    originalXraySettingRaw.value = xraySettingRaw.value;
    syncStructuredFromJson();
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    loading.value = false;
  }
}

function safeParseXrayRaw(): Record<string, unknown> | null {
  if (!isValidJSON(xraySettingRaw.value)) return null;
  return JSON.parse(xraySettingRaw.value) as Record<string, unknown>;
}

function parseCSV(csv: string): string[] {
  return csv
    .split(",")
    .map((x) => x.trim())
    .filter(Boolean);
}

function toCSV(value: unknown): string {
  return Array.isArray(value) ? value.map((x) => String(x)).filter(Boolean).join(",") : "";
}

function routingRuleObjectToRow(ruleObj: Record<string, unknown>): RoutingRuleRow | null {
  if (String(ruleObj.type ?? "field") !== "field") return null;
  const protocol = Array.isArray(ruleObj.protocol) ? (ruleObj.protocol as string[]).join(",") : String(ruleObj.protocol ?? "");
  const inboundTag = Array.isArray(ruleObj.inboundTag) ? (ruleObj.inboundTag as string[]).join(",") : String(ruleObj.inboundTag ?? "");
  const source = Array.isArray(ruleObj.source) ? (ruleObj.source as string[]).join(",") : String(ruleObj.source ?? ruleObj.sourceIP ?? "");
  const user = Array.isArray(ruleObj.user) ? (ruleObj.user as string[]).join(",") : String(ruleObj.user ?? "");
  const attrs = ruleObj.attrs && typeof ruleObj.attrs === "object" ? JSON.stringify(ruleObj.attrs) : "";
  const domain = Array.isArray(ruleObj.domain) ? (ruleObj.domain as string[]).join(",") : "";
  const ip = Array.isArray(ruleObj.ip) ? (ruleObj.ip as string[]).join(",") : "";
  const network = Array.isArray(ruleObj.network) ? (ruleObj.network as string[]).join(",") : String(ruleObj.network ?? "");
  return {
    type: "field",
    outboundTag: String(ruleObj.outboundTag ?? "direct"),
    balancerTag: String(ruleObj.balancerTag ?? ""),
    domain,
    ip,
    port: String(ruleObj.port ?? ""),
    network,
    source,
    sourcePort: String(ruleObj.sourcePort ?? ""),
    user,
    protocol,
    inboundTag,
    attrs
  };
}

function routingRowToRuleObject(row: RoutingRuleRow): Record<string, unknown> {
  const domain = row.domain.split(",").map((x) => x.trim()).filter(Boolean);
  const ip = row.ip.split(",").map((x) => x.trim()).filter(Boolean);
  const source = row.source.split(",").map((x) => x.trim()).filter(Boolean);
  const user = row.user.split(",").map((x) => x.trim()).filter(Boolean);
  const protocol = row.protocol.split(",").map((x) => x.trim()).filter(Boolean);
  const inboundTag = row.inboundTag.split(",").map((x) => x.trim()).filter(Boolean);
  const rule: Record<string, unknown> = { type: "field" };
  if (row.outboundTag.trim()) rule.outboundTag = row.outboundTag.trim();
  if (row.balancerTag.trim()) rule.balancerTag = row.balancerTag.trim();
  if (domain.length > 0) rule.domain = domain;
  if (ip.length > 0) rule.ip = ip;
  if (row.port.trim()) rule.port = row.port.trim();
  if (source.length > 0) rule.source = source;
  if (row.sourcePort.trim()) rule.sourcePort = row.sourcePort.trim();
  if (user.length > 0) rule.user = user;
  if (protocol.length > 0) rule.protocol = protocol;
  if (inboundTag.length > 0) rule.inboundTag = inboundTag;
  if (row.attrs.trim()) {
    try {
      const parsed = JSON.parse(row.attrs);
      if (parsed && typeof parsed === "object") {
        rule.attrs = parsed;
      }
    } catch {
      // keep current row values; attrs validation handled elsewhere on apply
    }
  }
  if (row.network.trim()) {
    const nets = row.network.split(",").map((x) => x.trim()).filter(Boolean);
    if (nets.length > 0) rule.network = nets;
  }
  return rule;
}

function outboundRawToRow(rawOutbound: Record<string, unknown>): OutboundFormRow {
  const settings = ((rawOutbound.settings as Record<string, unknown>) || {}) as Record<string, unknown>;
  return {
    raw: { ...rawOutbound },
    tag: String(rawOutbound.tag ?? ""),
    protocol: String(rawOutbound.protocol ?? "freedom"),
    sendThrough: String(rawOutbound.sendThrough ?? ""),
    domainStrategy: String(settings.domainStrategy ?? "")
  };
}

function getRulesList(root: Record<string, unknown>): Record<string, unknown>[] {
  const routing = ((root.routing as Record<string, unknown>) || {}) as Record<string, unknown>;
  if (!Array.isArray(routing.rules)) routing.rules = [];
  root.routing = routing;
  return routing.rules as Record<string, unknown>[];
}

function getRuleByOutbound(root: Record<string, unknown>, outboundTag: string): Record<string, unknown> | null {
  const rules = getRulesList(root);
  return (
    rules.find((rule) => String(rule.type ?? "") === "field" && String(rule.outboundTag ?? "") === outboundTag) || null
  );
}

function getOrCreateRuleByOutbound(root: Record<string, unknown>, outboundTag: string): Record<string, unknown> {
  const rules = getRulesList(root);
  let rule = rules.find((r) => String(r.type ?? "") === "field" && String(r.outboundTag ?? "") === outboundTag);
  if (!rule) {
    rule = { type: "field", outboundTag };
    rules.push(rule);
  }
  return rule;
}

function removeRulePropIfEmpty(rule: Record<string, unknown>, key: string): void {
  const value = rule[key];
  if (!Array.isArray(value) || value.length === 0) {
    delete rule[key];
  }
}

function getActiveBlockedTag(root: Record<string, unknown>): "blocked" | "block" {
  if (getRuleByOutbound(root, "blocked")) return "blocked";
  return "block";
}

function syncBasicTemplateFromJson(root: Record<string, unknown>): void {
  const outboundsObj = Array.isArray(root.outbounds) ? (root.outbounds as Record<string, unknown>[]) : [];
  const directOutbound = outboundsObj.find((o) => String(o.tag ?? "") === "direct" && String(o.protocol ?? "") === "freedom");
  const streamSettings = ((directOutbound?.settings as Record<string, unknown>) || {}) as Record<string, unknown>;
  const directDomainStrategy = String(streamSettings.domainStrategy ?? "AsIs");
  freedomStrategy.value = outboundDomainStrategies.includes(directDomainStrategy as (typeof outboundDomainStrategies)[number])
    ? (directDomainStrategy as (typeof outboundDomainStrategies)[number])
    : "AsIs";

  const routingObj = ((root.routing as Record<string, unknown>) || {}) as Record<string, unknown>;
  const routingDomainStrategy = String(routingObj.domainStrategy ?? "AsIs");
  routingStrategy.value = routingDomainStrategies.includes(routingDomainStrategy as (typeof routingDomainStrategies)[number])
    ? (routingDomainStrategy as (typeof routingDomainStrategies)[number])
    : "AsIs";

  const logObj = ((root.log as Record<string, unknown>) || {}) as Record<string, unknown>;
  const lvl = String(logObj.loglevel ?? "warning");
  logLevel.value = logLevels.includes(lvl as (typeof logLevels)[number]) ? (lvl as (typeof logLevels)[number]) : "warning";
  accessLogPath.value = String(logObj.access ?? "");
  errorLogPath.value = String(logObj.error ?? "");
  maskAddressLog.value = String(logObj.maskAddress ?? "");
  dnsLogEnabled.value = Boolean(logObj.dnsLog ?? false);

  const policyObj = ((root.policy as Record<string, unknown>) || {}) as Record<string, unknown>;
  const systemPolicy = ((policyObj.system as Record<string, unknown>) || {}) as Record<string, unknown>;
  outboundTrafficEnabled.value = Boolean(systemPolicy.statsOutboundDownlink && systemPolicy.statsOutboundUplink);

  const blockedTag = getActiveBlockedTag(root);
  const blockedRule = getRuleByOutbound(root, blockedTag);
  blockedDomainCsv.value = toCSV(blockedRule?.domain);
  blockedIpCsv.value = toCSV(blockedRule?.ip);
  blockedProtocolCsv.value = toCSV(blockedRule?.protocol);

  const directRule = getRuleByOutbound(root, "direct");
  directDomainCsv.value = toCSV(directRule?.domain);
  directIpCsv.value = toCSV(directRule?.ip);

  const ipv4Rule = getRuleByOutbound(root, "IPv4");
  ipv4DomainCsv.value = toCSV(ipv4Rule?.domain);

  const warpRule = getRuleByOutbound(root, "warp");
  warpDomainCsv.value = toCSV(warpRule?.domain);
}

function applyBasicTemplateToJson(root: Record<string, unknown>): void {
  let outboundsObj = Array.isArray(root.outbounds) ? (root.outbounds as Record<string, unknown>[]) : [];
  let directOutbound = outboundsObj.find((o) => String(o.tag ?? "") === "direct" && String(o.protocol ?? "") === "freedom");
  if (!directOutbound) {
    directOutbound = { tag: "direct", protocol: "freedom", settings: {} };
    outboundsObj = [...outboundsObj, directOutbound];
    root.outbounds = outboundsObj;
  }
  const directSettings = ((directOutbound.settings as Record<string, unknown>) || {}) as Record<string, unknown>;
  directSettings.domainStrategy = freedomStrategy.value;
  directOutbound.settings = directSettings;

  const routingObj = ((root.routing as Record<string, unknown>) || {}) as Record<string, unknown>;
  routingObj.domainStrategy = routingStrategy.value;
  root.routing = routingObj;

  const logObj = ((root.log as Record<string, unknown>) || {}) as Record<string, unknown>;
  logObj.loglevel = logLevel.value;
  logObj.access = accessLogPath.value.trim();
  logObj.error = errorLogPath.value.trim();
  logObj.maskAddress = maskAddressLog.value.trim();
  logObj.dnsLog = dnsLogEnabled.value;
  root.log = logObj;

  const policyObj = ((root.policy as Record<string, unknown>) || {}) as Record<string, unknown>;
  const systemPolicy = ((policyObj.system as Record<string, unknown>) || {}) as Record<string, unknown>;
  if (outboundTrafficEnabled.value) {
    systemPolicy.statsOutboundUplink = true;
    systemPolicy.statsOutboundDownlink = true;
  } else {
    delete systemPolicy.statsOutboundUplink;
    delete systemPolicy.statsOutboundDownlink;
  }
  policyObj.system = systemPolicy;
  root.policy = policyObj;

  const blockedTag = getActiveBlockedTag(root);
  const blockedRule = getRuleByOutbound(root, blockedTag);
  if (blockedRule) {
    blockedRule.domain = parseCSV(blockedDomainCsv.value);
    blockedRule.ip = parseCSV(blockedIpCsv.value);
    blockedRule.protocol = parseCSV(blockedProtocolCsv.value);
    removeRulePropIfEmpty(blockedRule, "domain");
    removeRulePropIfEmpty(blockedRule, "ip");
    removeRulePropIfEmpty(blockedRule, "protocol");
  }

  const directRule = getRuleByOutbound(root, "direct");
  if (directRule) {
    directRule.domain = parseCSV(directDomainCsv.value);
    directRule.ip = parseCSV(directIpCsv.value);
    removeRulePropIfEmpty(directRule, "domain");
    removeRulePropIfEmpty(directRule, "ip");
  }

  const ipv4Rule = getRuleByOutbound(root, "IPv4");
  if (ipv4Rule) {
    ipv4Rule.domain = parseCSV(ipv4DomainCsv.value);
    removeRulePropIfEmpty(ipv4Rule, "domain");
  }

  const warpRule = getRuleByOutbound(root, "warp");
  if (warpRule) {
    warpRule.domain = parseCSV(warpDomainCsv.value);
    removeRulePropIfEmpty(warpRule, "domain");
  }
}

function applyBasicTemplateOnly(): void {
  const root = safeParseXrayRaw();
  if (!root) {
    error.value = t("pages.xray.ui.xraySettingsJsonInvalid", "Xray settings JSON is invalid");
    return;
  }
  applyBasicTemplateToJson(root);
  xraySettingRaw.value = JSON.stringify(root, null, 2);
  syncStructuredFromJson();
  success.value = t("pages.xray.ui.basicTemplateApplied", "Basic template settings applied to JSON");
}

function syncStructuredFromJson(): void {
  const root = safeParseXrayRaw();
  if (!root) return;

  const dns = (root.dns as Record<string, unknown>) || {};
  const dnsServers = Array.isArray(dns.servers) ? dns.servers : [];
  dnsServersText.value = dnsServers
    .map((x) => (typeof x === "string" ? x : JSON.stringify(x)))
    .join("\n");
  dnsQueryStrategy.value = String(dns.queryStrategy ?? "");
  dnsClientIp.value = String(dns.clientIp ?? "");
  const fakeDnsObj = Array.isArray(root.fakedns) ? root.fakedns : [];
  fakeDnsRows.value = fakeDnsObj
    .map((item) => {
      const row = (item as Record<string, unknown>) || {};
      return {
        ipPool: String(row.ipPool ?? ""),
        poolSize: Number(row.poolSize ?? 0)
      };
    })
    .filter((x) => x.ipPool.trim().length > 0);

  const routing = (root.routing as Record<string, unknown>) || {};
  const rules = Array.isArray(routing.rules) ? routing.rules : [];
  routingRuleRows.value = rules
    .map((rule) => {
      const r = rule as Record<string, unknown>;
      if (String(r.type ?? "") !== "field") return null;
      const protocol = Array.isArray(r.protocol) ? (r.protocol as string[]).join(",") : String(r.protocol ?? "");
      const inboundTag = Array.isArray(r.inboundTag) ? (r.inboundTag as string[]).join(",") : String(r.inboundTag ?? "");
      const source = Array.isArray(r.source) ? (r.source as string[]).join(",") : String(r.source ?? r.sourceIP ?? "");
      const user = Array.isArray(r.user) ? (r.user as string[]).join(",") : String(r.user ?? "");
      const attrs = r.attrs && typeof r.attrs === "object" ? JSON.stringify(r.attrs) : "";
      const hasUnknownField = Object.keys(r).some(
        (k) =>
          ![
            "type",
            "outboundTag",
            "balancerTag",
            "domain",
            "ip",
            "port",
            "network",
            "source",
            "sourceIP",
            "sourcePort",
            "user",
            "protocol",
            "inboundTag",
            "attrs"
          ].includes(k)
      );
      if (hasUnknownField) return null;
      const domain = Array.isArray(r.domain) ? (r.domain as string[]).join(",") : "";
      const ip = Array.isArray(r.ip) ? (r.ip as string[]).join(",") : "";
      const network = Array.isArray(r.network) ? (r.network as string[]).join(",") : String(r.network ?? "");
      return {
        type: "field" as const,
        outboundTag: String(r.outboundTag ?? "direct"),
        balancerTag: String(r.balancerTag ?? ""),
        domain,
        ip,
        port: String(r.port ?? ""),
        network,
        source,
        sourcePort: String(r.sourcePort ?? ""),
        user,
        protocol,
        inboundTag,
        attrs
      };
    })
    .filter((x): x is RoutingRuleRow => Boolean(x));
  routingExtraRules.value = rules.filter((rule) => {
      const r = rule as Record<string, unknown>;
      if (String(r.type ?? "") !== "field") return true;
      const keys = Object.keys(r);
      const known = [
        "type",
        "outboundTag",
        "balancerTag",
        "domain",
        "ip",
        "port",
        "network",
        "source",
        "sourceIP",
        "sourcePort",
        "user",
        "protocol",
        "inboundTag",
        "attrs"
      ];
      return keys.some((k) => !known.includes(k));
    });

  const outboundsObj = Array.isArray(root.outbounds) ? (root.outbounds as Record<string, unknown>[]) : [];
  const tags = outboundsObj.map((o) => String(o.tag ?? "")).filter(Boolean);
  refreshOutboundTagOptions(tags);
  const inboundsObj = Array.isArray(root.inbounds) ? (root.inbounds as Record<string, unknown>[]) : [];
  const fromXrayInbounds = inboundsObj.map((inbound) => String(inbound.tag ?? "").trim()).filter(Boolean);
  inboundTagOptions.value = Array.from(new Set([...inboundTagOptions.value, ...fromXrayInbounds]));
  outboundRows.value = outboundsObj.map((outbound) => outboundRawToRow(outbound));

  const balancers = Array.isArray(routing.balancers) ? routing.balancers : [];
  balancerRows.value = balancers
    .map((entry) => {
      const row = (entry as Record<string, unknown>) || {};
      const selector = Array.isArray(row.selector) ? (row.selector as unknown[]).map((x) => String(x)).join(",") : "";
      return {
        tag: String(row.tag ?? ""),
        selector,
        strategy: String(row.strategy ?? "")
      };
    })
    .filter((x) => x.tag.trim().length > 0);
  syncBasicTemplateFromJson(root);
}

function applyStructuredToJson(): void {
  const root = safeParseXrayRaw();
  if (!root) return;
  let parseError = "";

  const dnsServers = dnsServersText.value
    .split("\n")
    .map((x) => x.trim())
    .filter(Boolean)
    .map((line) => {
      if (line.startsWith("{") || line.startsWith("[")) {
        try {
          return JSON.parse(line);
        } catch {
          return line;
        }
      }
      return line;
    });
  const dnsObj = (root.dns as Record<string, unknown>) || {};
  dnsObj.servers = dnsServers;
  if (dnsQueryStrategy.value.trim()) {
    dnsObj.queryStrategy = dnsQueryStrategy.value.trim();
  } else {
    delete dnsObj.queryStrategy;
  }
  if (dnsClientIp.value.trim()) {
    dnsObj.clientIp = dnsClientIp.value.trim();
  } else {
    delete dnsObj.clientIp;
  }
  root.dns = dnsObj;
  const nextFakeDns = fakeDnsRows.value
    .map((row) => ({
      ipPool: row.ipPool.trim(),
      poolSize: Number.isFinite(row.poolSize) ? Number(row.poolSize) : 65535
    }))
    .filter((row) => row.ipPool.length > 0);
  if (nextFakeDns.length > 0) {
    root.fakedns = nextFakeDns;
  } else {
    delete root.fakedns;
  }

  const rowRules = routingRuleRows.value
  .map((r) => {
    const domain = r.domain
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const ip = r.ip
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const source = r.source
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const user = r.user
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const protocol = r.protocol
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const inboundTag = r.inboundTag
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const rule: Record<string, unknown> = {
      type: "field"
    };
    if (r.outboundTag.trim()) {
      rule.outboundTag = r.outboundTag.trim();
    }
    if (r.balancerTag.trim()) {
      rule.balancerTag = r.balancerTag.trim();
    }
    if (domain.length > 0) rule.domain = domain;
    if (ip.length > 0) rule.ip = ip;
    if (r.port.trim()) rule.port = r.port.trim();
    if (source.length > 0) rule.source = source;
    if (r.sourcePort.trim()) rule.sourcePort = r.sourcePort.trim();
    if (user.length > 0) rule.user = user;
    if (protocol.length > 0) rule.protocol = protocol;
    if (inboundTag.length > 0) rule.inboundTag = inboundTag;
    if (r.attrs.trim()) {
      try {
        const parsed = JSON.parse(r.attrs);
        if (parsed && typeof parsed === "object") {
          rule.attrs = parsed;
        }
      } catch {
        parseError = t("pages.xray.ui.routingAttrsJsonInvalid", "Routing row attrs must be valid JSON object");
        return rule;
      }
    }
    if (r.network.trim()) {
      const nets = r.network
        .split(",")
        .map((x) => x.trim())
        .filter(Boolean);
      if (nets.length > 0) rule.network = nets;
    }
    return rule;
  })
  .filter((rule) => {
    const hasEffectiveField =
      Array.isArray(rule.domain) ||
      Array.isArray(rule.ip) ||
      Array.isArray(rule.source) ||
      Array.isArray(rule.user) ||
      Array.isArray(rule.protocol) ||
      Array.isArray(rule.inboundTag) ||
      Array.isArray(rule.network) ||
      Boolean(rule.port) ||
      Boolean(rule.sourcePort) ||
      Boolean(rule.attrs);
    return hasEffectiveField;
  });

  const rules = [...rowRules];
  if (parseError) {
    error.value = parseError;
    return;
  }
  const routingObj = (root.routing as Record<string, unknown>) || {};
  routingObj.rules = rules;
  const nextBalancers = balancerRows.value
    .map((row) => {
      const selector = row.selector
        .split(",")
        .map((x) => x.trim())
        .filter(Boolean);
      const balancer: Record<string, unknown> = {
        tag: row.tag.trim()
      };
      if (selector.length > 0) {
        balancer.selector = selector;
      }
      if (row.strategy.trim()) {
        balancer.strategy = row.strategy.trim();
      }
      return balancer;
    })
    .filter((row) => String(row.tag ?? "").length > 0);
  if (nextBalancers.length > 0) {
    routingObj.balancers = nextBalancers;
  } else {
    delete routingObj.balancers;
  }
  root.routing = routingObj;

  const nextOutbounds = outboundRows.value
    .map((row) => {
      const outbound: Record<string, unknown> = { ...(row.raw || {}) };
      outbound.tag = row.tag.trim();
      outbound.protocol = row.protocol.trim() || "freedom";
      if (row.sendThrough.trim()) {
        outbound.sendThrough = row.sendThrough.trim();
      } else {
        delete outbound.sendThrough;
      }
      if (row.domainStrategy.trim()) {
        const settings = ((outbound.settings as Record<string, unknown>) || {}) as Record<string, unknown>;
        settings.domainStrategy = row.domainStrategy.trim();
        outbound.settings = settings;
      }
      return outbound;
    })
    .filter((row) => String(row.tag ?? "").length > 0);
  if (nextOutbounds.length > 0) {
    root.outbounds = nextOutbounds;
  }

  applyBasicTemplateToJson(root);

  xraySettingRaw.value = JSON.stringify(root, null, 2);
  syncStructuredFromJson();
}

async function applyDnsRoutingModal(): Promise<void> {
  await saveXraySetting();
}

function addRoutingRuleRow(): void {
  routingRuleRows.value.push({
    type: "field",
    outboundTag: outboundTagOptions.value[0] || "direct",
    balancerTag: "",
    domain: "",
    ip: "",
    port: "",
    network: "",
    source: "",
    sourcePort: "",
    user: "",
    protocol: "",
    inboundTag: "",
    attrs: ""
  });
}

function resetRoutingRuleDraft(): void {
  routingRuleDraft.type = "field";
  routingRuleDraft.outboundTag = outboundTagOptions.value[0] || "direct";
  routingRuleDraft.balancerTag = "";
  routingRuleDraft.domain = "";
  routingRuleDraft.ip = "";
  routingRuleDraft.port = "";
  routingRuleDraft.network = "";
  routingRuleDraft.source = "";
  routingRuleDraft.sourcePort = "";
  routingRuleDraft.user = "";
  routingRuleDraft.protocol = "";
  routingRuleDraft.inboundTag = "";
  routingRuleDraft.attrs = "";
}

function openAddRoutingRuleModal(): void {
  resetRoutingRuleDraft();
  selectedInboundTagDraft.value = "";
  routingRuleModalMode.value = "add";
  routingRuleEditIndex.value = -1;
  addRoutingRuleModalOpen.value = true;
}

function addRoutingRuleFromModal(): void {
  const nextRow: RoutingRuleRow = {
    type: "field",
    outboundTag: routingRuleDraft.outboundTag.trim() || (outboundTagOptions.value[0] || "direct"),
    balancerTag: routingRuleDraft.balancerTag.trim(),
    domain: routingRuleDraft.domain.trim(),
    ip: routingRuleDraft.ip.trim(),
    port: routingRuleDraft.port.trim(),
    network: routingRuleDraft.network.trim(),
    source: routingRuleDraft.source.trim(),
    sourcePort: routingRuleDraft.sourcePort.trim(),
    user: routingRuleDraft.user.trim(),
    protocol: routingRuleDraft.protocol.trim(),
    inboundTag: routingRuleDraft.inboundTag.trim(),
    attrs: routingRuleDraft.attrs.trim()
  };
  if (routingRuleModalMode.value === "edit" && routingRuleEditIndex.value >= 0 && routingRuleEditIndex.value < routingRuleRows.value.length) {
    routingRuleRows.value.splice(routingRuleEditIndex.value, 1, nextRow);
  } else {
    routingRuleRows.value.push(nextRow);
  }
  routingRuleModalMode.value = "add";
  routingRuleEditIndex.value = -1;
  addRoutingRuleModalOpen.value = false;
}

function openEditRoutingRuleModal(index: number): void {
  if (index < 0 || index >= routingRuleRows.value.length) return;
  const row = routingRuleRows.value[index];
  routingRuleDraft.type = "field";
  routingRuleDraft.outboundTag = row.outboundTag || (outboundTagOptions.value[0] || "direct");
  routingRuleDraft.balancerTag = row.balancerTag || "";
  routingRuleDraft.domain = row.domain || "";
  routingRuleDraft.ip = row.ip || "";
  routingRuleDraft.port = row.port || "";
  routingRuleDraft.network = row.network || "";
  routingRuleDraft.source = row.source || "";
  routingRuleDraft.sourcePort = row.sourcePort || "";
  routingRuleDraft.user = row.user || "";
  routingRuleDraft.protocol = row.protocol || "";
  routingRuleDraft.inboundTag = row.inboundTag || "";
  routingRuleDraft.attrs = row.attrs || "";
  routingRuleModalMode.value = "edit";
  routingRuleEditIndex.value = index;
  selectedInboundTagDraft.value = "";
  addRoutingRuleModalOpen.value = true;
}

function addInboundTagToRoutingDraft(tag: string): void {
  const nextTag = String(tag || "").trim();
  if (!nextTag) return;
  const existing = parseCSV(routingRuleDraft.inboundTag);
  if (existing.includes(nextTag)) return;
  routingRuleDraft.inboundTag = [...existing, nextTag].join(",");
}

function removeInboundTagFromRoutingDraft(tag: string): void {
  const nextTag = String(tag || "").trim();
  if (!nextTag) return;
  routingRuleDraft.inboundTag = parseCSV(routingRuleDraft.inboundTag)
    .filter((x) => x !== nextTag)
    .join(",");
}

function removeRoutingRuleRow(index: number): void {
  routingRuleRows.value.splice(index, 1);
}

function addDnsServer(): void {
  const lines = dnsServersText.value
    .split("\n")
    .map((x) => x.trim())
    .filter(Boolean);
  lines.push("1.1.1.1");
  dnsServersText.value = lines.join("\n");
}

function addFakeDnsRow(): void {
  fakeDnsRows.value.push({ ipPool: "198.18.0.0/15", poolSize: 65535 });
}

function removeFakeDnsRow(index: number): void {
  fakeDnsRows.value.splice(index, 1);
}

function defaultOutboundTag(): string {
  const base = `out-${outboundRows.value.length + 1}`;
  const used = new Set(outboundRows.value.map((x) => x.tag.trim()).filter(Boolean));
  if (!used.has(base)) return base;
  let idx = outboundRows.value.length + 2;
  while (used.has(`out-${idx}`)) idx += 1;
  return `out-${idx}`;
}

function resetAddOutboundForm(): void {
  addOutboundForm.protocol = "vless";
  addOutboundForm.tag = defaultOutboundTag();
  addOutboundForm.sendThrough = "";
  addOutboundForm.vmessVlessAddress = "";
  addOutboundForm.vmessVlessPort = "443";
  addOutboundForm.vmessVlessUuid = "";
  addOutboundForm.vmessEncryption = "";
  addOutboundForm.vlessFlow = "";
  addOutboundForm.trojanAddress = "";
  addOutboundForm.trojanPort = "443";
  addOutboundForm.trojanPassword = "";
  addOutboundForm.trojanFlow = "";
  addOutboundForm.ssAddress = "";
  addOutboundForm.ssPort = "443";
  addOutboundForm.ssMethod = "aes-128-gcm";
  addOutboundForm.ssPassword = "";
  addOutboundForm.socksAddress = "";
  addOutboundForm.socksPort = "1080";
  addOutboundForm.socksUsername = "";
  addOutboundForm.socksPassword = "";
  addOutboundForm.socksVersion = "5";
  addOutboundForm.httpAddress = "";
  addOutboundForm.httpPort = "8080";
  addOutboundForm.httpUsername = "";
  addOutboundForm.httpPassword = "";
  addOutboundForm.hysteriaAddress = "";
  addOutboundForm.hysteriaPort = "443";
  addOutboundForm.hysteriaAuth = "";
  addOutboundForm.hysteriaObfsType = "";
  addOutboundForm.hysteriaObfsPassword = "";
  addOutboundForm.hysteriaSni = "";
  addOutboundForm.hysteriaAlpnCsv = "";
  addOutboundForm.hysteriaUpMbps = "";
  addOutboundForm.hysteriaDownMbps = "";
  addOutboundForm.wireguardSecretKey = "";
  addOutboundForm.wireguardPublicKey = "";
  addOutboundForm.wireguardPreSharedKey = "";
  addOutboundForm.wireguardEndpointHost = "";
  addOutboundForm.wireguardEndpointPort = "51820";
  addOutboundForm.wireguardLocalAddressCsv = "172.16.0.2/32";
  addOutboundForm.wireguardAllowedIpsCsv = "0.0.0.0/0,::/0";
  addOutboundForm.wireguardMtu = "";
  addOutboundForm.dnsAddress = "8.8.8.8";
  addOutboundForm.dnsPort = "53";
  addOutboundForm.dnsNetwork = "";
  addOutboundForm.freedomDomainStrategy = "";
  addOutboundForm.freedomRedirect = "";
  addOutboundForm.blackholeResponseType = "none";
  addOutboundForm.transmission = "tcp";
  addOutboundForm.security = "none";
  addOutboundForm.tlsServerName = "";
  addOutboundForm.tlsAlpnCsv = "";
  addOutboundForm.tlsMinVersion = "";
  addOutboundForm.tlsMaxVersion = "";
  addOutboundForm.tlsCipherSuitesCsv = "";
  addOutboundForm.tlsFingerprint = "";
  addOutboundForm.tlsAllowInsecure = false;
  addOutboundForm.tlsRejectUnknownSni = false;
  addOutboundForm.tlsCertFile = "";
  addOutboundForm.tlsKeyFile = "";
  addOutboundForm.realityShow = false;
  addOutboundForm.realityServerName = "";
  addOutboundForm.realityFingerprint = "chrome";
  addOutboundForm.realityPublicKey = "";
  addOutboundForm.realityShortId = "";
  addOutboundForm.realitySpiderX = "/";
  addOutboundForm.realityMldsa65Verify = "";
  addOutboundForm.httpObfs = false;
  addOutboundForm.httpObfsHostCsv = "";
  addOutboundForm.httpObfsPathCsv = "/";
  addOutboundForm.udpMasks = "";
  addOutboundForm.sockopt = false;
  addOutboundForm.sockoptJson = "{}";
  addOutboundForm.mux = false;
  addOutboundForm.muxConcurrency = "";
  addOutboundForm.muxXudpConcurrency = "";
  addOutboundForm.muxXudpProxyUDP443 = "";
  addOutboundForm.customSettingsJson = "{}";
  addOutboundForm.customStreamSettingsJson = "{}";
}

function parseOptionalJsonObject(value: string, fieldName: string): Record<string, unknown> | null {
  const raw = value.trim();
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      error.value = `${fieldName} must be a JSON object`;
      return null;
    }
    return parsed as Record<string, unknown>;
  } catch {
    error.value = `${fieldName} must be valid JSON`;
    return null;
  }
}

function buildOutboundRawFromForm(): Record<string, unknown> | null {
  const protocol = addOutboundForm.protocol.trim().toLowerCase() || "freedom";
  const customSettings = parseOptionalJsonObject(addOutboundForm.customSettingsJson, "Settings JSON");
  if (!customSettings) return null;
  const customStreamSettings = parseOptionalJsonObject(addOutboundForm.customStreamSettingsJson, "Stream Settings JSON");
  if (!customStreamSettings) return null;
  const sockoptSettings = parseOptionalJsonObject(addOutboundForm.sockoptJson, "Sockopt JSON");
  if (!sockoptSettings) return null;

  const outbound: Record<string, unknown> = {
    tag: addOutboundForm.tag.trim() || defaultOutboundTag(),
    protocol
  };
  if (addOutboundForm.sendThrough.trim()) {
    outbound.sendThrough = addOutboundForm.sendThrough.trim();
  }
  if (streamCapableProtocols.has(protocol)) {
    const streamSettings: Record<string, unknown> = {
      network: addOutboundForm.transmission.trim() || "tcp",
      security: addOutboundForm.security.trim() || "none",
      ...customStreamSettings
    };
    if (addOutboundForm.httpObfs) {
      const obfsHosts = parseCSV(addOutboundForm.httpObfsHostCsv);
      const obfsPaths = parseCSV(addOutboundForm.httpObfsPathCsv);
      streamSettings.tcpSettings = {
        header: {
          type: "http",
          request: {
            path: obfsPaths.length > 0 ? obfsPaths : ["/"],
            headers: obfsHosts.length > 0 ? { Host: obfsHosts } : {}
          }
        }
      };
    }
    if (addOutboundForm.sockopt) {
      streamSettings.sockopt = { ...sockoptSettings };
    }
    if (addOutboundForm.security === "tls") {
      const tlsAlpn = parseCSV(addOutboundForm.tlsAlpnCsv);
      const tlsCipherSuites = parseCSV(addOutboundForm.tlsCipherSuitesCsv);
      const certDefined = addOutboundForm.tlsCertFile.trim() || addOutboundForm.tlsKeyFile.trim();
      streamSettings.tlsSettings = {
        serverName: addOutboundForm.tlsServerName.trim() || undefined,
        alpn: tlsAlpn.length > 0 ? tlsAlpn : undefined,
        minVersion: addOutboundForm.tlsMinVersion.trim() || undefined,
        maxVersion: addOutboundForm.tlsMaxVersion.trim() || undefined,
        cipherSuites: tlsCipherSuites.length > 0 ? tlsCipherSuites : undefined,
        fingerprint: addOutboundForm.tlsFingerprint.trim() || undefined,
        allowInsecure: addOutboundForm.tlsAllowInsecure,
        rejectUnknownSni: addOutboundForm.tlsRejectUnknownSni,
        certificates: certDefined
          ? [
              {
                certificateFile: addOutboundForm.tlsCertFile.trim(),
                keyFile: addOutboundForm.tlsKeyFile.trim()
              }
            ]
          : undefined
      };
      delete streamSettings.realitySettings;
    } else if (addOutboundForm.security === "reality") {
      streamSettings.realitySettings = {
        show: addOutboundForm.realityShow,
        serverName: addOutboundForm.realityServerName.trim() || undefined,
        fingerprint: addOutboundForm.realityFingerprint.trim() || "chrome",
        publicKey: addOutboundForm.realityPublicKey.trim() || undefined,
        shortId: addOutboundForm.realityShortId.trim() || undefined,
        spiderX: addOutboundForm.realitySpiderX.trim() || "/",
        mldsa65Verify: addOutboundForm.realityMldsa65Verify.trim() || undefined
      };
      delete streamSettings.tlsSettings;
    } else {
      delete streamSettings.tlsSettings;
      delete streamSettings.realitySettings;
    }
    outbound.streamSettings = streamSettings;
  } else if (Object.keys(customStreamSettings).length > 0) {
    outbound.streamSettings = customStreamSettings;
  }

  const normalizedProtocol = protocol === "ss" ? "shadowsocks" : protocol;
  outbound.protocol = normalizedProtocol;
  if (normalizedProtocol === "vless" || normalizedProtocol === "vmess") {
    const address = addOutboundForm.vmessVlessAddress.trim();
    const port = Number(addOutboundForm.vmessVlessPort || 443);
    const id = addOutboundForm.vmessVlessUuid.trim();
    const encryption = addOutboundForm.vmessEncryption.trim();
    outbound.settings = {
      vnext: [
        {
          address,
          port: Number.isFinite(port) && port > 0 ? port : 443,
          users: [
            {
              id,
              encryption: encryption || (normalizedProtocol === "vless" ? "none" : "auto")
            }
          ]
        }
      ],
      ...customSettings
    };
    if (normalizedProtocol === "vless" && addOutboundForm.vlessFlow.trim()) {
      const settings = outbound.settings as Record<string, unknown>;
      const vnext = (settings.vnext as Record<string, unknown>[]) || [];
      const first = (vnext[0] || {}) as Record<string, unknown>;
      const users = (first.users as Record<string, unknown>[]) || [];
      if (users[0]) {
        users[0].flow = addOutboundForm.vlessFlow.trim();
      }
    }
  } else if (normalizedProtocol === "trojan") {
    const address = addOutboundForm.trojanAddress.trim();
    const port = Number(addOutboundForm.trojanPort || 443);
    const password = addOutboundForm.trojanPassword.trim();
    outbound.settings = {
      servers: [
        {
          address,
          port: Number.isFinite(port) && port > 0 ? port : 443,
          password,
          flow: addOutboundForm.trojanFlow.trim() || undefined
        }
      ],
      ...customSettings
    };
  } else if (normalizedProtocol === "shadowsocks") {
    const address = addOutboundForm.ssAddress.trim();
    const port = Number(addOutboundForm.ssPort || 443);
    outbound.settings = {
      servers: [
        {
          address,
          port: Number.isFinite(port) && port > 0 ? port : 443,
          method: addOutboundForm.ssMethod.trim() || "aes-128-gcm",
          password: addOutboundForm.ssPassword.trim()
        }
      ],
      ...customSettings
    };
  } else if (normalizedProtocol === "socks" || normalizedProtocol === "http") {
    const address = normalizedProtocol === "socks" ? addOutboundForm.socksAddress.trim() : addOutboundForm.httpAddress.trim();
    const port = Number(normalizedProtocol === "socks" ? addOutboundForm.socksPort : addOutboundForm.httpPort);
    const username = normalizedProtocol === "socks" ? addOutboundForm.socksUsername.trim() : addOutboundForm.httpUsername.trim();
    const password = normalizedProtocol === "socks" ? addOutboundForm.socksPassword.trim() : addOutboundForm.httpPassword.trim();
    outbound.settings = {
      servers: [
        {
          address,
          port: Number.isFinite(port) && port > 0 ? port : 1080,
          users: username ? [{ user: username, pass: password }] : undefined
        }
      ],
      version: normalizedProtocol === "socks" ? addOutboundForm.socksVersion.trim() || "5" : undefined,
      ...customSettings
    };
  } else if (normalizedProtocol === "dns") {
    const port = Number(addOutboundForm.dnsPort || 53);
    outbound.settings = {
      address: addOutboundForm.dnsAddress.trim() || "8.8.8.8",
      port: Number.isFinite(port) && port > 0 ? port : 53,
      network: addOutboundForm.dnsNetwork.trim() || undefined,
      ...customSettings
    };
  } else if (normalizedProtocol === "hysteria") {
    const address = addOutboundForm.hysteriaAddress.trim();
    const port = Number(addOutboundForm.hysteriaPort || 443);
    const alpn = parseCSV(addOutboundForm.hysteriaAlpnCsv);
    const up = Number(addOutboundForm.hysteriaUpMbps || 0);
    const down = Number(addOutboundForm.hysteriaDownMbps || 0);
    outbound.settings = {
      servers: [
        {
          address,
          port: Number.isFinite(port) && port > 0 ? port : 443,
          users: addOutboundForm.hysteriaAuth.trim() ? [{ auth: addOutboundForm.hysteriaAuth.trim() }] : undefined
        }
      ],
      obfs: addOutboundForm.hysteriaObfsPassword.trim()
        ? {
            type: addOutboundForm.hysteriaObfsType.trim() || "salamander",
            password: addOutboundForm.hysteriaObfsPassword.trim()
          }
        : undefined,
      up_mbps: Number.isFinite(up) && up > 0 ? up : undefined,
      down_mbps: Number.isFinite(down) && down > 0 ? down : undefined,
      serverName: addOutboundForm.hysteriaSni.trim() || undefined,
      alpn: alpn.length > 0 ? alpn : undefined,
      ...customSettings
    };
  } else if (normalizedProtocol === "wireguard") {
    const localAddrs = parseCSV(addOutboundForm.wireguardLocalAddressCsv);
    const allowedIps = parseCSV(addOutboundForm.wireguardAllowedIpsCsv);
    const endpointHost = addOutboundForm.wireguardEndpointHost.trim() || "127.0.0.1";
    const endpointPort = Number(addOutboundForm.wireguardEndpointPort || 51820);
    const mtu = Number(addOutboundForm.wireguardMtu || 0);
    outbound.settings = {
      secretKey: addOutboundForm.wireguardSecretKey.trim(),
      address: localAddrs.length > 0 ? localAddrs : ["172.16.0.2/32"],
      peers: [
        {
          publicKey: addOutboundForm.wireguardPublicKey.trim(),
          preSharedKey: addOutboundForm.wireguardPreSharedKey.trim() || undefined,
          endpoint: `${endpointHost}:${Number.isFinite(endpointPort) && endpointPort > 0 ? endpointPort : 51820}`,
          allowedIPs: allowedIps.length > 0 ? allowedIps : ["0.0.0.0/0", "::/0"]
        }
      ],
      mtu: Number.isFinite(mtu) && mtu > 0 ? mtu : undefined,
      ...customSettings
    };
  } else if (normalizedProtocol === "blackhole") {
    outbound.settings = {
      response: { type: addOutboundForm.blackholeResponseType.trim() || "none" },
      ...customSettings
    };
  } else if (normalizedProtocol === "freedom") {
    outbound.settings = {
      domainStrategy: addOutboundForm.freedomDomainStrategy.trim() || undefined,
      redirect: addOutboundForm.freedomRedirect.trim() || undefined,
      ...customSettings
    };
  } else {
    outbound.settings = { ...customSettings };
  }
  if (addOutboundForm.udpMasks.trim()) {
    outbound.xudpProxyUDP443 = addOutboundForm.udpMasks.trim();
  }
  if (addOutboundForm.mux) {
    const muxConcurrency = Number(addOutboundForm.muxConcurrency || 0);
    const muxXudpConcurrency = Number(addOutboundForm.muxXudpConcurrency || 0);
    outbound.mux = {
      enabled: true,
      concurrency: Number.isFinite(muxConcurrency) && muxConcurrency > 0 ? muxConcurrency : undefined,
      xudpConcurrency: Number.isFinite(muxXudpConcurrency) && muxXudpConcurrency > 0 ? muxXudpConcurrency : undefined,
      xudpProxyUDP443: addOutboundForm.muxXudpProxyUDP443.trim() || undefined
    };
  }
  return outbound;
}

function syncAddOutboundJsonFromForm(): void {
  const built = buildOutboundRawFromForm();
  if (!built) return;
  addOutboundJson.value = JSON.stringify(built, null, 2);
  error.value = "";
}

function openAddOutboundModal(): void {
  resetAddOutboundForm();
  addOutboundTab.value = "form";
  addOutboundLink.value = "";
  syncAddOutboundJsonFromForm();
  addOutboundModalOpen.value = true;
}

function switchAddOutboundTab(tab: "form" | "json"): void {
  if (tab === "json") {
    syncAddOutboundJsonFromForm();
  }
  addOutboundTab.value = tab;
}

function decodeBase64Utf8(input: string): string {
  const normalized = input.replace(/-/g, "+").replace(/_/g, "/");
  const padLen = (4 - (normalized.length % 4)) % 4;
  const padded = normalized + "=".repeat(padLen);
  try {
    return decodeURIComponent(escape(atob(padded)));
  } catch {
    return atob(padded);
  }
}

function parsePortMaybe(value: string, fallback: number): number {
  const n = Number(value || 0);
  return Number.isFinite(n) && n > 0 ? n : fallback;
}

function applyCommonStreamSettingsFromParams(
  protocol: string,
  params: URLSearchParams,
  streamSettings: Record<string, unknown>,
  user0?: Record<string, unknown>
): void {
  const network = params.get("type") || params.get("net") || "tcp";
  const tlsFlag = (params.get("tls") || "").toLowerCase();
  const security = params.get("security") || (tlsFlag === "tls" || tlsFlag === "1" ? "tls" : "none");
  streamSettings.network = network;
  streamSettings.security = security;

  if (network === "ws") {
    const host = params.get("host") || params.get("wsHost") || "";
    streamSettings.wsSettings = {
      path: params.get("path") || "/",
      headers: host ? { Host: host } : {}
    };
  } else if (network === "grpc") {
    const serviceName = params.get("serviceName") || params.get("service") || params.get("path") || "";
    streamSettings.grpcSettings = {
      serviceName: serviceName.replace(/^\//, ""),
      authority: params.get("authority") || ""
    };
  } else if (network === "httpupgrade") {
    streamSettings.httpupgradeSettings = {
      host: params.get("host") || "",
      path: params.get("path") || "/"
    };
  } else if (network === "tcp" && (params.get("headerType") || "").toLowerCase() === "http") {
    const host = params.get("host") || "";
    streamSettings.tcpSettings = {
      header: {
        type: "http",
        request: {
          path: [params.get("path") || "/"],
          headers: host ? { Host: host.split(",").map((x) => x.trim()).filter(Boolean) } : {}
        }
      }
    };
  }

  const sni = params.get("sni") || params.get("servername") || "";
  const alpn = parseCSV(params.get("alpn") || "");
  const fp = params.get("fp") || params.get("fingerprint") || "";
  const allowInsecure = (params.get("allowInsecure") || params.get("insecure") || "").toLowerCase();

  if (security === "tls") {
    streamSettings.tlsSettings = {
      serverName: sni || undefined,
      alpn: alpn.length > 0 ? alpn : undefined,
      fingerprint: fp || undefined,
      allowInsecure: allowInsecure === "1" || allowInsecure === "true"
    };
    delete streamSettings.realitySettings;
  } else if (security === "reality") {
    streamSettings.realitySettings = {
      serverName: sni || undefined,
      fingerprint: fp || "chrome",
      publicKey: params.get("pbk") || params.get("publicKey") || undefined,
      shortId: params.get("sid") || params.get("shortId") || undefined,
      spiderX: params.get("spx") || params.get("spiderX") || "/"
    };
    delete streamSettings.tlsSettings;
  }

  const flow = params.get("flow") || "";
  if (flow && user0 && (protocol === "vless" || protocol === "trojan")) {
    user0.flow = flow;
  }
}

function parseSsOutboundLink(raw: string): Record<string, unknown> {
  const full = raw.replace(/^ss:\/\//, "");
  const hashIndex = full.indexOf("#");
  const beforeHash = hashIndex >= 0 ? full.slice(0, hashIndex) : full;
  const tag = hashIndex >= 0 ? decodeURIComponent(full.slice(hashIndex + 1)) : defaultOutboundTag();
  const queryIndex = beforeHash.indexOf("?");
  const body = queryIndex >= 0 ? beforeHash.slice(0, queryIndex) : beforeHash;
  const query = queryIndex >= 0 ? beforeHash.slice(queryIndex + 1) : "";
  const params = new URLSearchParams(query);

  let method = "";
  let password = "";
  let address = "";
  let port = 443;

  if (body.includes("@")) {
    const [credRaw, hostPortRaw] = body.split("@");
    const credDecoded = credRaw.includes(":") ? credRaw : decodeBase64Utf8(credRaw);
    const [m, p] = credDecoded.split(":");
    method = decodeURIComponent(m || "");
    password = decodeURIComponent((p || "").trim());
    const hp = hostPortRaw.trim();
    if (hp.startsWith("[")) {
      const close = hp.indexOf("]");
      address = hp.slice(1, close);
      port = parsePortMaybe(hp.slice(close + 2), 443);
    } else {
      const idx = hp.lastIndexOf(":");
      address = idx > 0 ? hp.slice(0, idx) : hp;
      port = parsePortMaybe(idx > 0 ? hp.slice(idx + 1) : "", 443);
    }
  } else {
    const decoded = decodeBase64Utf8(body);
    const atIdx = decoded.lastIndexOf("@");
    const creds = decoded.slice(0, atIdx);
    const hostPort = decoded.slice(atIdx + 1);
    const [m, p] = creds.split(":");
    method = m || "";
    password = p || "";
    const idx = hostPort.lastIndexOf(":");
    address = idx > 0 ? hostPort.slice(0, idx) : hostPort;
    port = parsePortMaybe(idx > 0 ? hostPort.slice(idx + 1) : "", 443);
  }

  return {
    tag: tag || defaultOutboundTag(),
    protocol: "shadowsocks",
    settings: {
      servers: [
        {
          address,
          port,
          method: method || "aes-128-gcm",
          password
        }
      ]
    },
    plugin: params.get("plugin") || undefined
  };
}

function importOutboundLink(): void {
  const raw = addOutboundLink.value.trim();
  if (!raw) return;
  try {
    if (raw.startsWith("vmess://")) {
      const payload = raw.slice("vmess://".length).trim();
      const decoded = decodeBase64Utf8(payload);
      const obj = JSON.parse(decoded) as Record<string, unknown>;
      const user0: Record<string, unknown> = {
        id: String(obj.id || ""),
        security: String(obj.scy || "auto")
      };
      const aid = Number(obj.aid || 0);
      if (Number.isFinite(aid) && aid > 0) user0.alterId = aid;
      const streamSettings: Record<string, unknown> = {};
      const params = new URLSearchParams();
      const setIf = (k: string, v: unknown): void => {
        const s = String(v ?? "");
        if (s) params.set(k, s);
      };
      setIf("type", obj.net);
      setIf("path", obj.path);
      setIf("host", obj.host);
      setIf("sni", obj.sni);
      setIf("alpn", obj.alpn);
      setIf("fp", obj.fp);
      setIf("security", obj.tls === "tls" || obj.tls === true ? "tls" : obj.security);
      applyCommonStreamSettingsFromParams("vmess", params, streamSettings, user0);
      const outbound = {
        tag: String(obj.ps || defaultOutboundTag()),
        protocol: "vmess",
        settings: {
          vnext: [
            {
              address: String(obj.add || ""),
              port: parsePortMaybe(String(obj.port || ""), 443),
              users: [user0]
            }
          ]
        },
        streamSettings
      };
      addOutboundJson.value = JSON.stringify(outbound, null, 2);
      addOutboundTab.value = "json";
      error.value = "";
      return;
    }
    if (raw.startsWith("hy2://") || raw.startsWith("hysteria2://")) {
      const normalizedRaw = raw.startsWith("hy2://") ? raw.replace(/^hy2:\/\//, "hysteria2://") : raw;
      const u = new URL(normalizedRaw);
      const obfsType = u.searchParams.get("obfs") || "";
      const obfsPassword = u.searchParams.get("obfs-password") || u.searchParams.get("obfsPassword") || "";
      const sni = u.searchParams.get("sni") || u.searchParams.get("peer") || "";
      const insecureRaw = (u.searchParams.get("insecure") || "").toLowerCase();
      const alpn = parseCSV(u.searchParams.get("alpn") || "");
      const upMbps = Number(u.searchParams.get("upmbps") || u.searchParams.get("up") || 0);
      const downMbps = Number(u.searchParams.get("downmbps") || u.searchParams.get("down") || 0);
      const outbound = {
        tag: decodeURIComponent(u.hash.replace(/^#/, "")) || defaultOutboundTag(),
        protocol: "hysteria2",
        settings: {
          servers: [
            {
              address: u.hostname,
              port: parsePortMaybe(u.port, 443)
            }
          ],
          auth: decodeURIComponent(u.username || ""),
          obfs: obfsPassword
            ? {
                type: obfsType || "salamander",
                password: obfsPassword
              }
            : undefined,
          tls: {
            serverName: sni || undefined,
            insecure: insecureRaw === "1" || insecureRaw === "true",
            alpn: alpn.length > 0 ? alpn : undefined
          },
          up_mbps: Number.isFinite(upMbps) && upMbps > 0 ? upMbps : undefined,
          down_mbps: Number.isFinite(downMbps) && downMbps > 0 ? downMbps : undefined
        }
      };
      addOutboundJson.value = JSON.stringify(outbound, null, 2);
      addOutboundTab.value = "json";
      error.value = "";
      return;
    }
    if (raw.startsWith("vless://") || raw.startsWith("trojan://") || raw.startsWith("ss://")) {
      if (raw.startsWith("ss://")) {
        const outbound = parseSsOutboundLink(raw);
        addOutboundJson.value = JSON.stringify(outbound, null, 2);
        addOutboundTab.value = "json";
        error.value = "";
        return;
      }
      const u = new URL(raw);
      const protocol = u.protocol.replace(":", "");
      const user0: Record<string, unknown> = protocol === "trojan"
        ? { password: decodeURIComponent(u.username || "") }
        : { id: decodeURIComponent(u.username || ""), encryption: u.searchParams.get("encryption") || "none" };
      const streamSettings: Record<string, unknown> = {};
      applyCommonStreamSettingsFromParams(protocol, u.searchParams, streamSettings, user0);
      const outbound = {
        tag: decodeURIComponent(u.hash.replace(/^#/, "")) || defaultOutboundTag(),
        protocol,
        settings: protocol === "trojan"
          ? {
              servers: [
                {
                  address: u.hostname,
                  port: parsePortMaybe(u.port, 443),
                  ...user0
                }
              ]
            }
          : {
              vnext: [
                {
                  address: u.hostname,
                  port: parsePortMaybe(u.port, 443),
                  users: [user0]
                }
              ]
            },
        streamSettings
      };
      addOutboundJson.value = JSON.stringify(outbound, null, 2);
      addOutboundTab.value = "json";
      error.value = "";
      return;
    }
    error.value = t("pages.xray.ui.unsupportedImportLink", "Unsupported link format for import");
  } catch {
    error.value = t("pages.xray.ui.parseOutboundLinkFailed", "Failed to parse outbound link");
  }
}

function addOutboundFromModal(): void {
  let raw: Record<string, unknown>;
  if (addOutboundTab.value === "json") {
    try {
      const parsed = JSON.parse(addOutboundJson.value) as Record<string, unknown>;
      if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
        error.value = t("pages.xray.ui.outboundJsonObject", "Outbound JSON must be an object");
        return;
      }
      raw = parsed;
    } catch {
      error.value = t("pages.xray.ui.addOutboundJsonInvalid", "Add outbound JSON is invalid");
      return;
    }
  } else {
    const built = buildOutboundRawFromForm();
    if (!built) return;
    raw = built;
  }
  const tag = String(raw.tag ?? "").trim();
  if (!tag) {
    error.value = t("pages.xray.ui.outboundTagRequired", "Outbound tag is required");
    return;
  }
  if (outboundRows.value.some((row) => row.tag.trim() === tag)) {
    error.value = t("pages.xray.ui.outboundTagUnique", "Outbound tag must be unique");
    return;
  }
  const protocol = String(raw.protocol ?? "freedom").trim() || "freedom";
  const sendThrough = String(raw.sendThrough ?? "").trim();
  outboundRows.value.push(outboundRawToRow({
    ...raw,
    tag,
    protocol,
    sendThrough
  }));
  addOutboundModalOpen.value = false;
  error.value = "";
}

function openEditOutboundModal(index: number): void {
  if (index < 0 || index >= outboundRows.value.length) return;
  editOutboundIndex.value = index;
  editOutboundJson.value = JSON.stringify(outboundRows.value[index].raw || {}, null, 2);
  editOutboundModalOpen.value = true;
}

function saveEditOutboundModal(): void {
  if (editOutboundIndex.value < 0 || editOutboundIndex.value >= outboundRows.value.length) return;
  try {
    const parsed = JSON.parse(editOutboundJson.value) as Record<string, unknown>;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      error.value = t("pages.xray.ui.outboundJsonObject", "Outbound JSON must be an object");
      return;
    }
    const tag = String(parsed.tag ?? "").trim();
    if (!tag) {
      error.value = t("pages.xray.ui.outboundTagRequired", "Outbound tag is required");
      return;
    }
    const duplicate = outboundRows.value.some((row, idx) => idx !== editOutboundIndex.value && row.tag.trim() === tag);
    if (duplicate) {
      error.value = t("pages.xray.ui.outboundTagUnique", "Outbound tag must be unique");
      return;
    }
    outboundRows.value.splice(editOutboundIndex.value, 1, outboundRawToRow(parsed));
    editOutboundModalOpen.value = false;
    error.value = "";
  } catch {
    error.value = t("pages.xray.ui.outboundJsonInvalid", "Outbound JSON is invalid");
  }
}

async function loadInboundTagsFromPanel(): Promise<void> {
  try {
    const msg = await getJson<InboundListItem[]>("panel/api/inbounds/list");
    if (!msg.success || !Array.isArray(msg.obj)) return;
    const tags = msg.obj.map((x) => String(x?.tag ?? "").trim()).filter(Boolean);
    inboundTagOptions.value = Array.from(new Set([...inboundTagOptions.value, ...tags]));
  } catch {
    // best-effort only
  }
}

function removeOutboundRow(index: number): void {
  outboundRows.value.splice(index, 1);
}

function addBalancerRow(): void {
  balancerRows.value.push({
    tag: `balancer-${balancerRows.value.length + 1}`,
    selector: "direct",
    strategy: ""
  });
}

function removeBalancerRow(index: number): void {
  balancerRows.value.splice(index, 1);
}

async function saveXraySetting(): Promise<void> {
  error.value = "";
  success.value = "";
  applyStructuredToJson();
  if (!isValidJSON(xraySettingRaw.value)) {
    error.value = t("pages.xray.ui.xraySettingsJsonInvalid", "Xray settings JSON is invalid");
    return;
  }
  busy.value = true;
  try {
    const msg = await postForm("panel/xray/update", {
      xraySetting: xraySettingRaw.value
    });
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.updateXrayConfigFailed", "Failed to update xray config"));
    }
    success.value = t("pages.xray.ui.xraySettingsUpdated", "Xray settings updated");
    await loadXraySetting();
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function loadDefaultXrayConfig(): Promise<void> {
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await getJson<unknown>("panel/xray/getDefaultJsonConfig");
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.loadDefaultXrayConfigFailed", "Failed to load default xray config"));
    }
    xraySettingRaw.value = JSON.stringify(msg.obj, null, 2);
    syncStructuredFromJson();
    success.value = t("pages.xray.ui.defaultXrayConfigLoaded", "Default xray config loaded into editor");
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function getXrayRuntimeResult(): Promise<void> {
  busy.value = true;
  error.value = "";
  try {
    const msg = await getJson<string>("panel/xray/getXrayResult");
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.fetchRuntimeResultFailed", "Failed to fetch xray runtime result"));
    }
    xrayResult.value = typeof msg.obj === "string" ? msg.obj : JSON.stringify(msg.obj, null, 2);
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function restartXrayService(): Promise<void> {
  if (!(await openConfirmModal(t("pages.xray.ui.restartNow", "Restart Xray service now?"), t("pages.xray.restart", "Restart Xray")))) {
    return;
  }
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await postJson("panel/api/server/restartXrayService");
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.restartXrayFailed", "Failed to restart xray service"));
    }
    success.value = t("pages.xray.ui.xrayRestartRequested", "Xray restart requested");
    await getXrayRuntimeResult();
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function stopXrayService(): Promise<void> {
  if (!(await openConfirmModal(t("pages.xray.ui.stopNow", "Stop Xray service now?"), t("pages.index.stopXray", "Stop Xray")))) {
    return;
  }
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await postJson("panel/api/server/stopXrayService");
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.stopXrayFailed", "Failed to stop xray service"));
    }
    success.value = t("pages.xray.ui.xrayStopRequested", "Xray stop requested");
    await getXrayRuntimeResult();
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function loadOutboundsTraffic(): Promise<void> {
  busy.value = true;
  error.value = "";
  try {
    const msg = await getJson<OutboundTraffic[]>("panel/xray/getOutboundsTraffic");
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.loadOutboundTrafficFailed", "Failed to load outbound traffic"));
    }
    outbounds.value = msg.obj || [];
    refreshOutboundTagOptions();
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function resetOutboundTraffic(tag: string): Promise<void> {
  if (!(await openConfirmModal(t("pages.xray.ui.resetOutboundTrafficConfirm", `Reset traffic for outbound tag '${tag}'?`), t("pages.xray.ui.resetOutboundTraffic", "Reset Outbound Traffic")))) {
    return;
  }
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const msg = await postForm("panel/xray/resetOutboundsTraffic", { tag });
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.resetOutboundTrafficFailed", "Failed to reset outbound traffic"));
    }
    success.value = `Outbound traffic reset: ${tag}`;
    await loadOutboundsTraffic();
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

function setFirstOutbound(tag: string): void {
  error.value = "";
  try {
    const idx = outboundRows.value.findIndex((o) => o.tag === tag);
    if (idx < 0) {
      throw new Error(t("pages.xray.ui.outboundTagNotFound", `Outbound tag not found: ${tag}`));
    }
    if (idx === 0) {
      success.value = t("pages.xray.ui.outboundAlreadyFirst", `Outbound ${tag} is already first`);
      return;
    }
    const next = [...outboundRows.value];
    const [target] = next.splice(idx, 1);
    next.unshift(target);
    outboundRows.value = next;
    applyStructuredToJson();
    success.value = t("pages.xray.ui.outboundMovedFirst", `Moved outbound ${tag} to first position`);
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.reorderOutboundsFailed", "Failed to reorder outbounds");
  }
}

function openJsonEditor(): void {
  const target = document.getElementById("xray-json-editor");
  target?.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function warpAction(
  action: "data" | "config" | "del" | "reg" | "license",
  options?: { silentSuccess?: boolean }
): Promise<void> {
  const silentSuccess = options?.silentSuccess === true;
  busy.value = true;
  error.value = "";
  if (!silentSuccess) {
    success.value = "";
  }
  try {
    let msg: ApiResponse<string>;
    if (action === "reg") {
      msg = await postForm(`panel/xray/warp/${action}`, {
        privateKey: warp.privateKey,
        publicKey: warp.publicKey
      });
    } else if (action === "license") {
      msg = await postForm(`panel/xray/warp/${action}`, {
        license: warp.license
      });
    } else {
      msg = await postForm(`panel/xray/warp/${action}`, {});
    }

    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.warpActionFailed", `WARP ${action} failed`));
    }

    if (action === "data") {
      warp.data = typeof msg.obj === "string" ? msg.obj : JSON.stringify(msg.obj, null, 2);
      if (!silentSuccess) {
        success.value = t("pages.xray.ui.warpDataLoaded", "WARP data loaded");
      }
    } else if (action === "config") {
      warp.config = typeof msg.obj === "string" ? msg.obj : JSON.stringify(msg.obj, null, 2);
      if (!silentSuccess) {
        success.value = t("pages.xray.ui.warpConfigLoaded", "WARP config loaded");
      }
    } else if (action === "del") {
      warp.data = "";
      warp.config = "";
      if (!silentSuccess) {
        success.value = t("pages.xray.ui.warpDataDeleted", "WARP data deleted");
      }
    } else if (action === "reg") {
      warp.data = typeof msg.obj === "string" ? msg.obj : JSON.stringify(msg.obj, null, 2);
      if (!silentSuccess) {
        success.value = t("pages.xray.ui.warpRegistrationComplete", "WARP registration complete");
      }
    } else if (action === "license") {
      if (!silentSuccess) {
        success.value = typeof msg.obj === "string" ? msg.obj : t("pages.xray.ui.warpLicenseUpdated", "WARP license updated");
      }
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

async function onMainActionMenu(action: string): Promise<void> {
  if (!action) return;
  if (action === "save") {
    await saveXraySetting();
    return;
  }
  if (action === "revert") {
    xraySettingRaw.value = originalXraySettingRaw.value;
    syncStructuredFromJson();
    return;
  }
  if (action === "export") {
    downloadXrayJson();
    return;
  }
  if (action === "import") {
    triggerImportXrayJson();
    return;
  }
  if (action === "default") {
    await loadDefaultXrayConfig();
    return;
  }
  if (action === "jsonEditor") {
    openJsonEditor();
    return;
  }
  if (action === "restart") {
    await restartXrayService();
    return;
  }
  if (action === "stop") {
    await stopXrayService();
    return;
  }
  if (action === "runtime") {
    await getXrayRuntimeResult();
  }
}

async function onOutboundRowMenu(action: string, tag: string): Promise<void> {
  if (!action) return;
  if (action === "setFirst") {
    setFirstOutbound(tag);
    return;
  }
  if (action === "reset") {
    await resetOutboundTraffic(tag);
  }
}

async function onWarpActionMenu(action: string): Promise<void> {
  if (!action) return;
  if (action === "data" || action === "config" || action === "reg" || action === "license" || action === "del") {
    await warpAction(action);
  }
}

async function applyTunnel(): Promise<void> {
  busy.value = true;
  error.value = "";
  success.value = "";
  try {
    const ip = encodeURIComponent(tunnel.ip || "");
    const port = encodeURIComponent(tunnel.port || "");
    const user = encodeURIComponent(tunnel.user || "");
    const password = encodeURIComponent(tunnel.password || "");
    const msg = await postJson(`panel/api/server/setTunnel/${ip}/${port}/${user}/${password}`);
    if (!msg.success) {
      throw new Error(msg.msg || t("pages.xray.ui.applyTunnelFailed", "Failed to apply tunnel"));
    }
    success.value = t("pages.xray.ui.tunnelSettingsApplied", "Tunnel settings applied");
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("pages.xray.ui.unexpectedError", "Unexpected error");
  } finally {
    busy.value = false;
  }
}

onMounted(async () => {
  void loadI18n([
    "pages.xray.title",
    "pages.settings.actions",
    "pages.xray.save",
    "pages.xray.restart",
    "pages.xray.ui.revert",
    "pages.xray.ui.exportJson",
    "pages.xray.ui.importJson",
    "pages.xray.ui.runtimeRefresh",
    "pages.xray.ui.runtimeResult",
    "pages.xray.ui.basicTemplateControls",
    "pages.xray.ui.basicTemplateControlsDesc",
    "pages.xray.ui.applyToJson",
    "pages.xray.ui.structuredDnsServers",
    "pages.xray.ui.structuredRoutingRules",
    "pages.xray.ui.outboundsSection",
    "pages.xray.ui.routingBalancersSection",
    "pages.xray.ui.reverseSection",
    "pages.xray.ui.observatorySection",
    "pages.xray.ui.burstObservatorySection",
    "pages.xray.ui.noOutboundRows",
    "pages.xray.outBoundTraffic",
    "pages.xray.warpRouting",
    "pages.xray.ui.warpDesc",
    "pages.xray.ui.warpActions",
    "pages.xray.ui.loadData",
    "pages.xray.ui.loadConfig",
    "pages.xray.ui.register",
    "pages.xray.ui.setLicense",
    "pages.xray.ui.deleteWarp",
    "pages.xray.ui.warpData",
    "pages.xray.ui.warpConfig",
    "pages.xray.ui.tunnel",
    "pages.xray.ui.tunnelDesc",
    "pages.xray.ui.privateKey",
    "pages.xray.ui.publicKey",
    "pages.xray.ui.licenseKey",
    "pages.xray.ui.tunnelHost",
    "pages.xray.ui.tunnelPort",
    "pages.xray.ui.tunnelUser",
    "pages.xray.ui.tunnelPassword",
    "pages.xray.ui.applyTunnel",
    "reload",
    "close"
  ]);
  await loadXraySetting();
  await loadInboundTagsFromPanel();
  await loadOutboundsTraffic();
  await getXrayRuntimeResult();
  await warpAction("data", { silentSuccess: true });
});

watch(outboundRows, () => refreshOutboundTagOptions(), { deep: true });

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
  <div class="space-y-4">
    <UiCard>
      <UiCardHeader>
        <div class="flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
          <div>
            <UiCardTitle>{{ t("pages.xray.title", "Xray Configuration") }}</UiCardTitle>
            <UiCardDescription>{{ t("pages.xray.TemplateDesc", "Edit Xray configuration, apply templates, and manage runtime operations.") }}</UiCardDescription>
          </div>
          <div class="flex w-full flex-wrap gap-2 sm:w-auto">
            <UiButton class="w-full sm:w-auto" :disabled="loading || busy" variant="outline" @click="loadXraySetting">
              <RefreshCw class="mr-2 h-4 w-4" />
              {{ t("reload", "Reload") }}
            </UiButton>
            <details class="relative w-full sm:w-auto">
              <summary class="action-trigger w-full justify-center sm:w-auto sm:justify-start">
                <MoreVertical class="h-4 w-4" />
                {{ t("pages.settings.actions", "Actions") }}
              </summary>
              <div class="absolute left-0 z-20 mt-1 min-w-[240px] rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onMainActionMenu('save')">{{ t("pages.xray.save", "Save Xray Config") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onMainActionMenu('revert')">{{ t("pages.xray.ui.revert", "Revert Changes") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onMainActionMenu('export')">{{ t("pages.xray.ui.exportJson", "Export JSON") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onMainActionMenu('import')">{{ t("pages.xray.ui.importJson", "Import JSON") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onMainActionMenu('default')">{{ t("default", "Load Default") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onMainActionMenu('runtime')">{{ t("pages.xray.ui.runtimeRefresh", "Refresh Runtime Result") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm text-red-700 hover:bg-red-50" :disabled="busy" @click="onMainActionMenu('restart')">{{ t("pages.xray.restart", "Restart Xray") }}</button>
                <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm text-red-700 hover:bg-red-50" :disabled="busy" @click="onMainActionMenu('stop')">{{ t("pages.index.stopXray", "Stop Xray") }}</button>
              </div>
            </details>
          </div>
        </div>
      </UiCardHeader>
      <UiCardContent>
        <input ref="xrayImportInput" class="hidden" type="file" accept=".json,application/json" @change="onImportXrayJson" />

        <div class="mt-4 rounded-lg border p-3">
          <h3 class="mb-2 text-sm font-semibold">{{ t("pages.xray.ui.runtimeResult", "Runtime Result") }}</h3>
          <pre class="max-h-40 overflow-auto whitespace-pre-wrap rounded-md border bg-muted/30 p-3 text-xs">{{ xrayResult || "(empty)" }}</pre>
        </div>

        <div class="mt-4 flex flex-wrap gap-2">
          <UiButton :disabled="busy" size="sm" variant="outline" @click="basicModalOpen = true">{{ t("pages.xray.ui.basicSettings", "Basic Settings") }}</UiButton>
          <UiButton :disabled="busy" size="sm" variant="outline" @click="dnsRoutingModalOpen = true">{{ t("pages.xray.ui.dnsAndRouting", "DNS and Routing") }}</UiButton>
          <UiButton :disabled="busy" size="sm" variant="outline" @click="outboundsModalOpen = true">{{ t("pages.xray.ui.outboundsAndBalancers", "Outbounds and Balancers") }}</UiButton>
          <UiButton :disabled="busy" size="sm" variant="outline" @click="tunnelModalOpen = true">{{ t("pages.xray.ui.tunnel", "Tunnel") }}</UiButton>
        </div>
      </UiCardContent>
    </UiCard>

    <div v-if="warpModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <UiCard class="max-h-[92vh] w-[min(96vw,1400px)] max-w-none overflow-y-auto">
        <UiCardHeader>
        <UiCardTitle>{{ t("pages.xray.warpRouting", "WARP") }}</UiCardTitle>
        <UiCardDescription>{{ t("pages.xray.ui.warpDesc", "Manage WARP registration, licensing, and generated connection data.") }}</UiCardDescription>
        </UiCardHeader>
        <UiCardContent>
        <div class="mb-3">
          <UiButton size="sm" variant="outline" @click="warpModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
          <input v-model="warp.privateKey" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.privateKey', 'Private key')" type="text" />
          <input v-model="warp.publicKey" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.publicKey', 'Public key')" type="text" />
          <input v-model="warp.license" class="rounded-md border bg-background px-3 py-2 text-sm md:col-span-2" :placeholder="t('pages.xray.ui.licenseKey', 'License key')" type="text" />
        </div>
        <div class="mt-3">
          <details class="relative inline-block">
            <summary class="action-trigger">
              <MoreVertical class="h-4 w-4" />
              {{ t("pages.xray.ui.warpActions", "WARP Actions") }}
            </summary>
            <div class="absolute left-0 z-20 mt-1 min-w-[200px] rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onWarpActionMenu('data')">{{ t("pages.xray.ui.loadData", "Load Data") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onWarpActionMenu('config')">{{ t("pages.xray.ui.loadConfig", "Load Config") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onWarpActionMenu('reg')">{{ t("pages.xray.ui.register", "Register") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="busy" @click="onWarpActionMenu('license')">{{ t("pages.xray.ui.setLicense", "Set License") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm text-red-700 hover:bg-red-50" :disabled="busy" @click="onWarpActionMenu('del')">{{ t("pages.xray.ui.deleteWarp", "Delete WARP") }}</button>
            </div>
          </details>
        </div>
        <div class="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-2">
          <div>
            <div class="mb-1 text-sm font-medium">{{ t("pages.xray.ui.warpData", "WARP Data") }}</div>
            <textarea v-model="warp.data" class="h-48 w-full json-field px-3 py-2 text-xs" readonly />
          </div>
          <div>
            <div class="mb-1 text-sm font-medium">{{ t("pages.xray.ui.warpConfig", "WARP Config") }}</div>
            <textarea v-model="warp.config" class="h-48 w-full json-field px-3 py-2 text-xs" readonly />
          </div>
        </div>
        </UiCardContent>
      </UiCard>
    </div>

    <div v-if="tunnelModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <UiCard class="max-h-[92vh] w-[min(96vw,1200px)] max-w-none overflow-y-auto">
        <UiCardHeader>
        <UiCardTitle>{{ t("pages.xray.ui.tunnel", "Tunnel") }}</UiCardTitle>
        <UiCardDescription>{{ t("pages.xray.ui.tunnelDesc", "Set upstream tunnel credentials used by the Xray connection path.") }}</UiCardDescription>
        </UiCardHeader>
        <UiCardContent>
        <div class="mb-3">
          <UiButton size="sm" variant="outline" @click="tunnelModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
        <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
          <input v-model="tunnel.ip" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.tunnelHost', 'Tunnel IP / Host')" type="text" />
          <input v-model="tunnel.port" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.tunnelPort', 'Tunnel Port')" type="text" />
          <input v-model="tunnel.user" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.tunnelUser', 'Tunnel User')" type="text" />
          <input v-model="tunnel.password" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.tunnelPassword', 'Tunnel Password')" type="password" />
        </div>
        <div class="mt-3">
          <UiButton :disabled="busy" size="sm" variant="outline" @click="applyTunnel">{{ t("pages.xray.ui.applyTunnel", "Apply Tunnel") }}</UiButton>
        </div>
        </UiCardContent>
      </UiCard>
    </div>

    <UiCard id="xray-json-editor" class="mt-4">
      <UiCardHeader>
        <UiCardTitle>{{ t("pages.xray.ui.fullJsonEditor", "Full Xray JSON Editor") }}</UiCardTitle>
      </UiCardHeader>
      <UiCardContent>
        <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
          <textarea v-model="xraySettingRaw" class="json-field h-[68vh] w-full px-3 py-2 text-xs" />
          <pre class="json-codeblock h-[68vh] overflow-auto whitespace-pre-wrap break-words p-3 text-xs" v-html="highlightedJsonEditorHtml" />
        </div>
      </UiCardContent>
    </UiCard>

    <div v-if="basicModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <div class="w-full max-w-5xl rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-6">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ t("pages.xray.ui.basicSettings", "Basic Settings") }}</h2>
          <UiButton size="sm" variant="outline" @click="basicModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
        <div class="grid max-h-[72vh] grid-cols-1 gap-5 overflow-y-auto lg:grid-cols-2">
          <div class="space-y-4 rounded-lg border bg-muted/20 p-4 sm:p-5">
            <label class="grid gap-1.5 text-xs">
              <span class="font-medium">{{ t("pages.xray.ui.freedomDomainStrategy", "Freedom Domain Strategy (direct)") }}</span>
              <select v-model="freedomStrategy" class="rounded-md border bg-background px-3 py-2 text-sm">
                <option v-for="opt in outboundDomainStrategies" :key="`f-${opt}`" :value="opt">{{ opt }}</option>
              </select>
            </label>
            <label class="grid gap-1.5 text-xs">
              <span class="font-medium">{{ t("pages.xray.ui.routingDomainStrategy", "Routing Domain Strategy") }}</span>
              <select v-model="routingStrategy" class="rounded-md border bg-background px-3 py-2 text-sm">
                <option v-for="opt in routingDomainStrategies" :key="`r-${opt}`" :value="opt">{{ opt }}</option>
              </select>
            </label>
            <label class="grid gap-1.5 text-xs">
              <span class="font-medium">{{ t("pages.xray.logLevel", "Log Level") }}</span>
              <select v-model="logLevel" class="rounded-md border bg-background px-3 py-2 text-sm">
                <option v-for="opt in logLevels" :key="`l-${opt}`" :value="opt">{{ opt }}</option>
              </select>
            </label>
            <input v-model="accessLogPath" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.accessLogPath', 'Log access path')" type="text" />
            <input v-model="errorLogPath" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.errorLogPath', 'Log error path')" type="text" />
            <input v-model="maskAddressLog" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.maskAddressLog', 'Mask address log mode')" type="text" />
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="dnsLogEnabled" type="checkbox" />
              <span>{{ t("pages.xray.dnsLog", "Enable DNS log") }}</span>
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="outboundTrafficEnabled" type="checkbox" />
              <span>{{ t("pages.xray.outBoundTraffic", "Enable outbound traffic stats") }}</span>
            </label>
          </div>
          <div class="rounded-lg border bg-muted/20 p-4 sm:p-5">
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <input v-model="blockedDomainCsv" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.blockedDomainsCsv', 'Blocked domains (csv)')" type="text" />
              <input v-model="blockedIpCsv" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.blockedIpsCsv', 'Blocked IPs (csv)')" type="text" />
              <input v-model="blockedProtocolCsv" class="rounded-md border bg-background px-3 py-2 text-sm sm:col-span-2" :placeholder="t('pages.xray.ui.blockedProtocolsCsv', 'Blocked protocols (csv)')" type="text" />
              <input v-model="directDomainCsv" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.directDomainsCsv', 'Direct domains (csv)')" type="text" />
              <input v-model="directIpCsv" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.directIpsCsv', 'Direct IPs (csv)')" type="text" />
              <input v-model="ipv4DomainCsv" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.ipv4DomainsCsv', 'IPv4 domains (csv)')" type="text" />
              <input v-model="warpDomainCsv" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.warpDomainsCsv', 'WARP domains (csv)')" type="text" />
            </div>
          </div>
        </div>
        <div class="mt-4 flex flex-wrap gap-2 border-t pt-4">
          <UiButton :disabled="busy" size="sm" @click="applyBasicTemplateOnly">{{ t("pages.xray.tunnel.apply", "Apply") }}</UiButton>
        </div>
      </div>
    </div>

    <div v-if="dnsRoutingModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <div class="w-full max-w-6xl rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-5">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ t("pages.xray.ui.dnsAndRouting", "DNS and Routing") }}</h2>
        </div>
        <div class="grid max-h-[70vh] grid-cols-1 gap-3 overflow-y-auto lg:grid-cols-2">
          <div class="rounded-lg border p-3">
            <div class="mb-2 grid grid-cols-1 gap-2 md:grid-cols-2">
              <input v-model="dnsQueryStrategy" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.queryStrategyHint', 'Query strategy (e.g. UseIPv4)')" type="text" />
              <input v-model="dnsClientIp" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.clientIpOptional', 'Client IP (optional)')" type="text" />
            </div>
            <p class="mb-2 text-xs text-muted-foreground">{{ t("pages.xray.ui.dnsServersPerLine", "DNS servers (one per line)") }}</p>
            <textarea v-model="dnsServersText" class="h-28 w-full rounded-md border bg-background px-3 py-2 text-xs" />
            <div class="mt-2">
              <UiButton :disabled="busy" size="sm" variant="outline" @click="addDnsServer">{{ t("pages.xray.dns.add", "Add DNS Server") }}</UiButton>
            </div>
            <p class="mb-2 mt-3 text-xs text-muted-foreground">{{ t("pages.xray.ui.fakeDnsPools", "FakeDNS pools") }}</p>
            <div class="space-y-2">
              <div v-for="(row, idx) in fakeDnsRows" :key="`fdm-${idx}`" class="grid grid-cols-1 gap-2 md:grid-cols-[1fr_150px_auto]">
                <input v-model="row.ipPool" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.ipPoolCidr', 'IP Pool (CIDR)')" type="text" />
                <input v-model.number="row.poolSize" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.poolSize', 'Pool size')" type="number" />
                <UiButton :disabled="busy" size="sm" variant="outline" @click="removeFakeDnsRow(idx)">{{ t("pages.xray.ui.remove", "Remove") }}</UiButton>
              </div>
            </div>
            <div class="mt-2">
              <UiButton :disabled="busy" size="sm" variant="outline" @click="addFakeDnsRow">{{ t("pages.xray.fakedns.add", "Add FakeDNS Pool") }}</UiButton>
            </div>
          </div>
          <div class="rounded-lg border p-3">
            <div class="mb-2 overflow-x-auto">
              <table class="tx-table-center w-full min-w-[760px] border-collapse text-xs">
                <thead>
                  <tr class="border-b text-left text-muted-foreground">
                    <th class="px-2 py-1">{{ t("pages.xray.rules.outbound", "Outbound") }}</th>
                    <th class="px-2 py-1">{{ t("pages.xray.rules.balancer", "Balancer") }}</th>
                    <th class="px-2 py-1">{{ t("pages.xray.ui.inboundTags", "Inbound Tags") }}</th>
                    <th class="px-2 py-1">{{ t("protocol", "Protocol") }}</th>
                    <th class="px-2 py-1">{{ t("pages.xray.outbound.domain", "Domain") }}</th>
                    <th class="px-2 py-1">{{ t("ip", "IP") }}</th>
                    <th class="px-2 py-1">{{ t("pages.xray.rules.source", "Source") }}</th>
                    <th class="px-2 py-1">{{ t("pages.xray.ui.sourcePort", "Source Port") }}</th>
                    <th class="px-2 py-1">{{ t("user", "User") }}</th>
                    <th class="px-2 py-1">{{ t("port", "Port") }}</th>
                    <th class="px-2 py-1">{{ t("pages.inbounds.network", "Network") }}</th>
                    <th class="px-2 py-1">{{ t("pages.xray.ui.attrsJson", "Attrs JSON") }}</th>
                    <th class="px-2 py-1">{{ t("pages.settings.actions", "Actions") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in routingRuleRows" :key="`rrm-${idx}`" class="border-b">
                    <td class="px-2 py-1"><span class="inline-block rounded border px-2 py-0.5">{{ row.outboundTag || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[130px] truncate align-middle" :title="row.balancerTag || '-'">{{ row.balancerTag || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[130px] truncate align-middle" :title="row.inboundTag || '-'">{{ row.inboundTag || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[120px] truncate align-middle" :title="row.protocol || '-'">{{ row.protocol || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[160px] truncate align-middle" :title="row.domain || '-'">{{ row.domain || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[120px] truncate align-middle" :title="row.ip || '-'">{{ row.ip || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[120px] truncate align-middle" :title="row.source || '-'">{{ row.source || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[90px] truncate align-middle" :title="row.sourcePort || '-'">{{ row.sourcePort || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[90px] truncate align-middle" :title="row.user || '-'">{{ row.user || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[90px] truncate align-middle" :title="row.port || '-'">{{ row.port || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[100px] truncate align-middle" :title="row.network || '-'">{{ row.network || "-" }}</span></td>
                    <td class="px-2 py-1"><span class="inline-block max-w-[130px] truncate align-middle" :title="row.attrs || '-'">{{ row.attrs || "-" }}</span></td>
                    <td class="px-2 py-1">
                      <div class="flex items-center gap-1">
                        <UiButton :disabled="busy" size="sm" variant="outline" @click="openEditRoutingRuleModal(idx)">{{ t("pages.xray.ui.editJson", "Edit JSON") }}</UiButton>
                        <UiButton :disabled="busy" size="sm" variant="outline" @click="removeRoutingRuleRow(idx)">{{ t("pages.xray.ui.remove", "Remove") }}</UiButton>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="routingRuleRows.length === 0">
                    <td class="px-2 py-4 text-center text-muted-foreground" colspan="13">No routing rules</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <UiButton :disabled="busy" size="sm" variant="outline" @click="openAddRoutingRuleModal">Add Rule</UiButton>
          </div>
        </div>
        <div class="mt-4 flex flex-wrap justify-end gap-2 border-t pt-4">
          <UiButton size="sm" variant="outline" @click="dnsRoutingModalOpen = false">{{ t("close", "Close") }}</UiButton>
          <UiButton :disabled="busy" size="sm" @click="applyDnsRoutingModal">{{ t("pages.xray.tunnel.apply", "Apply") }}</UiButton>
        </div>
      </div>
    </div>
    <div v-if="addRoutingRuleModalOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <div class="w-full max-w-xl rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-5">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ routingRuleModalMode === "edit" ? t("edit", "Edit Rule") : t("pages.xray.rules.add", "Add Rule") }}</h2>
        </div>
        <div class="grid max-h-[70vh] grid-cols-1 gap-2 overflow-y-auto">
          <input v-model="routingRuleDraft.source" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.sourceIpsCsv', 'Source IPs (csv)')" type="text" />
          <input v-model="routingRuleDraft.sourcePort" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.sourcePortExample', 'Source Port (e.g. 53,443,1000-2000)')" type="text" />
          <input v-model="routingRuleDraft.port" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.portExample', 'Port (e.g. 53,443,1000-2000)')" type="text" />
          <select v-model="routingRuleDraft.network" class="rounded-md border bg-background px-3 py-2 text-sm">
            <option value="">{{ t("pages.xray.ui.networkAny", "Network (any)") }}</option>
            <option value="tcp">tcp</option>
            <option value="udp">udp</option>
            <option value="tcp,udp">tcp,udp</option>
          </select>
          <input v-model="routingRuleDraft.protocol" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.protocolCsv', 'Protocol (csv)')" type="text" />
          <input v-model="routingRuleDraft.attrs" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder='Attributes JSON (e.g. {"k":"v"})' type="text" />
          <input v-model="routingRuleDraft.ip" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.ipCsv', 'IP (csv)')" type="text" />
          <input v-model="routingRuleDraft.domain" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.domainCsv', 'Domain (csv)')" type="text" />
          <input v-model="routingRuleDraft.user" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.userCsv', 'User (csv)')" type="text" />
          <div class="grid gap-2">
            <input v-model="routingRuleDraft.inboundTag" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.ui.inboundTagsCsv', 'Inbound Tags (csv)')" type="text" />
            <div class="flex items-center gap-2">
              <select v-model="selectedInboundTagDraft" class="flex-1 rounded-md border bg-background px-3 py-2 text-sm">
                <option value="">{{ t("pages.xray.ui.selectInboundTag", "Select inbound tag") }}</option>
                <option v-for="tag in inboundTagOptions" :key="`new-it-${tag}`" :value="tag">{{ tag }}</option>
              </select>
              <UiButton size="sm" variant="outline" :disabled="!selectedInboundTagDraft" @click="addInboundTagToRoutingDraft(selectedInboundTagDraft)">{{ t("pages.xray.rules.add", "Add") }}</UiButton>
            </div>
            <div v-if="parseCSV(routingRuleDraft.inboundTag).length > 0" class="flex flex-wrap gap-1.5">
              <button
                v-for="tag in parseCSV(routingRuleDraft.inboundTag)"
                :key="`draft-it-chip-${tag}`"
                class="rounded-full border px-2 py-0.5 text-xs hover:bg-muted"
                type="button"
                @click="removeInboundTagFromRoutingDraft(tag)"
              >
                {{ tag }} ×
              </button>
            </div>
          </div>
          <select v-model="routingRuleDraft.outboundTag" class="rounded-md border bg-background px-3 py-2 text-sm">
            <option v-for="tag in outboundTagOptions" :key="`new-rt-${tag}`" :value="tag">{{ tag }}</option>
          </select>
          <select v-model="routingRuleDraft.balancerTag" class="rounded-md border bg-background px-3 py-2 text-sm">
            <option value="">{{ t("pages.xray.ui.balancerTagOptional", "Balancer Tag (optional)") }}</option>
            <option v-for="row in balancerRows" :key="`new-bl-${row.tag}`" :value="row.tag">{{ row.tag }}</option>
          </select>
        </div>
        <div class="mt-4 flex justify-end gap-2">
          <UiButton size="sm" variant="outline" @click="addRoutingRuleModalOpen = false">{{ t("close", "Close") }}</UiButton>
          <UiButton :disabled="busy" size="sm" @click="addRoutingRuleFromModal">{{ routingRuleModalMode === "edit" ? t("save", "Save Rule") : t("pages.xray.rules.add", "Add Rule") }}</UiButton>
        </div>
      </div>
    </div>

    <div v-if="outboundsModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <div class="w-full max-w-6xl rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-5">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ t("pages.xray.ui.outboundsAndBalancers", "Outbounds and Balancers") }}</h2>
          <div class="flex items-center gap-2">
            <UiButton :disabled="busy" size="sm" variant="outline" @click="openAddOutboundModal">{{ t("pages.xray.outbound.addOutbound", "Add Outbound") }}</UiButton>
            <UiButton :disabled="busy" size="sm" variant="outline" @click="warpModalOpen = true">{{ t("pages.xray.warpRouting", "WARP") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="outboundsModalOpen = false">{{ t("close", "Close") }}</UiButton>
          </div>
        </div>
        <div class="max-h-[70vh] space-y-3 overflow-y-auto">
          <div class="rounded-lg border">
            <div class="overflow-x-auto">
              <table class="tx-table-center w-full min-w-[880px] border-collapse text-sm">
                <thead>
                  <tr class="border-b text-left text-muted-foreground">
                    <th class="w-16 px-3 py-2 font-medium">#</th>
                    <th class="px-3 py-2 font-medium">{{ t("pages.xray.outbound.tag", "Tag") }}</th>
                    <th class="px-3 py-2 font-medium">{{ t("protocol", "Protocol") }}</th>
                    <th class="px-3 py-2 font-medium">{{ t("address", "Address") }}</th>
                    <th class="px-3 py-2 font-medium">{{ t("traffic", "Traffic") }}</th>
                    <th class="px-3 py-2 font-medium">{{ t("pages.settings.actions", "Actions") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in outboundRows" :key="`obm-${idx}`" class="border-b">
                    <td class="px-3 py-2">{{ idx + 1 }}</td>
                    <td class="px-3 py-2">
                      <input v-model="row.tag" class="w-full rounded-md border bg-background px-2 py-1.5 text-sm" :placeholder="t('pages.xray.outbound.tag', 'Tag')" type="text" />
                    </td>
                    <td class="px-3 py-2">
                      <input v-model="row.protocol" class="w-full rounded-md border bg-background px-2 py-1.5 text-sm" :placeholder="t('protocol', 'Protocol')" type="text" />
                    </td>
                    <td class="px-3 py-2">
                      <span class="inline-block max-w-[280px] truncate align-middle" :title="outboundAddressLabel(row)">{{ outboundAddressLabel(row) }}</span>
                    </td>
                    <td class="px-3 py-2">
                      <span class="inline-flex rounded-full border px-2 py-0.5 text-xs">{{ outboundTrafficLabel(row.tag) }}</span>
                    </td>
                    <td class="px-3 py-2">
                      <div class="flex flex-wrap items-center gap-1.5">
                        <UiButton :disabled="busy" size="sm" variant="outline" @click="openEditOutboundModal(idx)">{{ t("pages.xray.ui.editJson", "Edit JSON") }}</UiButton>
                        <UiButton :disabled="busy" size="sm" variant="outline" @click="onOutboundRowMenu('setFirst', row.tag)">{{ t("pages.xray.ui.setFirst", "Set First") }}</UiButton>
                        <UiButton :disabled="busy || !row.tag.trim()" size="sm" variant="outline" @click="onOutboundRowMenu('reset', row.tag)">{{ t("reset", "Reset") }}</UiButton>
                        <UiButton :disabled="busy" size="sm" variant="outline" @click="removeOutboundRow(idx)">{{ t("pages.xray.ui.removeOutbound", "Remove Outbound") }}</UiButton>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="outboundRows.length === 0">
                    <td class="px-3 py-6 text-center text-muted-foreground" colspan="6">{{ t("pages.xray.ui.noOutboundRows", "No outbound rows") }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="rounded-lg border p-3">
            <div class="mb-2 text-sm font-medium">{{ t("pages.xray.ui.routingBalancersSection", "Balancers") }}</div>
            <div class="space-y-2">
              <div v-for="(row, idx) in balancerRows" :key="`blm-${idx}`" class="grid grid-cols-1 gap-2 rounded-md border p-2 md:grid-cols-2">
                <input v-model="row.tag" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.balancer.tag', 'Tag')" type="text" />
                <input v-model="row.strategy" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="t('pages.xray.balancer.balancerStrategy', 'Strategy (optional)')" type="text" />
                <input v-model="row.selector" class="rounded-md border bg-background px-3 py-2 text-sm md:col-span-2" :placeholder="t('pages.xray.ui.selectorTagsCsv', 'Selector tags (csv)')" type="text" />
                <div class="md:col-span-2">
                  <UiButton :disabled="busy" size="sm" variant="outline" @click="removeBalancerRow(idx)">Remove Balancer</UiButton>
                </div>
              </div>
            </div>
            <div class="mt-2">
              <UiButton :disabled="busy" size="sm" variant="outline" @click="addBalancerRow">Add Balancer</UiButton>
            </div>
          </div>
        </div>
        <div class="mt-4 flex gap-2">
        </div>
      </div>
    </div>
    <div v-if="addOutboundModalOpen" class="fixed inset-0 z-[80] flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <div class="w-full max-w-2xl rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-5">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ t("pages.xray.outbound.addOutbound", "Add Outbound") }}</h2>
        </div>
        <div class="mb-3 flex gap-2 border-b pb-2">
          <UiButton size="sm" :variant="addOutboundTab === 'form' ? 'default' : 'outline'" @click="switchAddOutboundTab('form')">{{ t("pages.xray.ui.form", "Form") }}</UiButton>
          <UiButton size="sm" :variant="addOutboundTab === 'json' ? 'default' : 'outline'" @click="switchAddOutboundTab('json')">{{ t("pages.xray.ui.json", "JSON") }}</UiButton>
        </div>

        <div v-if="addOutboundTab === 'form'" class="grid max-h-[65vh] grid-cols-1 gap-3 overflow-y-auto">
          <label class="grid gap-1 text-sm">
            <span class="text-muted-foreground">{{ t("protocol", "Protocol") }}</span>
            <select v-model="addOutboundForm.protocol" class="rounded-md border bg-background px-3 py-2" @change="syncAddOutboundJsonFromForm">
              <option v-for="opt in addOutboundProtocolOptions" :key="`proto-${opt}`" :value="opt">{{ opt }}</option>
            </select>
          </label>
          <input v-model="addOutboundForm.tag" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Unique Tag" type="text" @input="syncAddOutboundJsonFromForm" />
          <input v-model="addOutboundForm.sendThrough" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Send Through" type="text" @input="syncAddOutboundJsonFromForm" />
          <div v-if="selectedAddOutboundProtocol === 'vmess' || selectedAddOutboundProtocol === 'vless'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.vmessVlessAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.vmessVlessPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.vmessVlessUuid" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="UUID" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.vmessEncryption" class="rounded-md border bg-background px-3 py-2 text-sm" :placeholder="selectedAddOutboundProtocol === 'vmess' ? 'Security (auto/none...)' : 'Encryption (none)'" type="text" @input="syncAddOutboundJsonFromForm" />
            <input
              v-if="selectedAddOutboundProtocol === 'vless'"
              v-model="addOutboundForm.vlessFlow"
              class="rounded-md border bg-background px-3 py-2 text-sm"
              placeholder="Flow (optional)"
              type="text"
              @input="syncAddOutboundJsonFromForm"
            />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'trojan'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.trojanAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.trojanPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.trojanPassword" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Password" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.trojanFlow" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Flow (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'shadowsocks'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.ssAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.ssPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.ssMethod" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Method (e.g. chacha20-ietf-poly1305)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.ssPassword" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Password" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'socks'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.socksAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.socksPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.socksUsername" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Username (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.socksPassword" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Password (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <select v-model="addOutboundForm.socksVersion" class="rounded-md border bg-background px-3 py-2" @change="syncAddOutboundJsonFromForm">
              <option value="5">SOCKS5</option>
              <option value="4">SOCKS4</option>
            </select>
          </div>
          <div v-if="selectedAddOutboundProtocol === 'http'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.httpAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.httpPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.httpUsername" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Username (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.httpPassword" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Password (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'hysteria'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.hysteriaAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaAuth" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Auth" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaObfsType" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Obfs type (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaObfsPassword" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Obfs password (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaSni" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="SNI (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaAlpnCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="ALPN CSV (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaUpMbps" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Up Mbps (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.hysteriaDownMbps" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Down Mbps (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'wireguard'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.wireguardSecretKey" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Secret key" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardPublicKey" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Peer public key" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardPreSharedKey" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Pre-shared key (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardEndpointHost" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Endpoint host" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardEndpointPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Endpoint port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardLocalAddressCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Local addresses CSV" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardAllowedIpsCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Allowed IPs CSV" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.wireguardMtu" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="MTU (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'dns'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.dnsAddress" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="DNS server address" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.dnsPort" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="DNS server port" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.dnsNetwork" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Network (optional: udp/tcp)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'freedom'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <input v-model="addOutboundForm.freedomDomainStrategy" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Domain strategy (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.freedomRedirect" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Redirect (optional host:port)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div v-if="selectedAddOutboundProtocol === 'blackhole'" class="grid grid-cols-1 gap-2 rounded-md border p-2">
            <select v-model="addOutboundForm.blackholeResponseType" class="rounded-md border bg-background px-3 py-2" @change="syncAddOutboundJsonFromForm">
              <option value="none">none</option>
              <option value="http">http</option>
            </select>
          </div>
          <label v-if="streamCapableProtocols.has(selectedAddOutboundProtocol)" class="grid gap-1 text-sm">
            <span class="text-muted-foreground">{{ t("transmission", "Transmission") }}</span>
            <select v-model="addOutboundForm.transmission" class="rounded-md border bg-background px-3 py-2" @change="syncAddOutboundJsonFromForm">
              <option value="tcp">TCP (RAW)</option>
              <option value="udp">UDP</option>
              <option value="ws">WebSocket</option>
              <option value="grpc">gRPC</option>
              <option value="kcp">KCP</option>
              <option value="httpupgrade">HTTPUpgrade</option>
              <option value="quic">QUIC</option>
            </select>
          </label>
          <label v-if="streamCapableProtocols.has(selectedAddOutboundProtocol)" class="grid gap-1 text-sm">
            <span class="text-muted-foreground">{{ t("security", "Security") }}</span>
            <div class="inline-flex w-fit overflow-hidden rounded-md border bg-background">
              <button
                type="button"
                class="px-3 py-1.5 text-xs transition-colors"
                :class="addOutboundForm.security === 'none' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
                @click="addOutboundForm.security = 'none'; syncAddOutboundJsonFromForm()"
              >
                None
              </button>
              <button
                type="button"
                class="border-l px-3 py-1.5 text-xs transition-colors"
                :class="addOutboundForm.security === 'tls' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
                @click="addOutboundForm.security = 'tls'; syncAddOutboundJsonFromForm()"
              >
                TLS
              </button>
              <button
                type="button"
                class="border-l px-3 py-1.5 text-xs transition-colors"
                :class="addOutboundForm.security === 'reality' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
                @click="addOutboundForm.security = 'reality'; syncAddOutboundJsonFromForm()"
              >
                Reality
              </button>
            </div>
          </label>
          <div v-if="streamCapableProtocols.has(selectedAddOutboundProtocol) && addOutboundForm.security === 'tls'" class="grid grid-cols-1 gap-2 rounded-md border p-2.5">
            <div class="text-xs font-medium text-muted-foreground">TLS Settings</div>
            <input v-model="addOutboundForm.tlsServerName" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="SNI / Server Name" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.tlsAlpnCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="ALPN CSV (e.g. h2,http/1.1)" type="text" @input="syncAddOutboundJsonFromForm" />
            <div class="grid grid-cols-2 gap-2">
              <input v-model="addOutboundForm.tlsMinVersion" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Min version (1.2)" type="text" @input="syncAddOutboundJsonFromForm" />
              <input v-model="addOutboundForm.tlsMaxVersion" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Max version (1.3)" type="text" @input="syncAddOutboundJsonFromForm" />
            </div>
            <input v-model="addOutboundForm.tlsCipherSuitesCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Cipher suites CSV (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.tlsFingerprint" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="uTLS fingerprint (chrome/firefox/...)" type="text" @input="syncAddOutboundJsonFromForm" />
            <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
              <input v-model="addOutboundForm.tlsCertFile" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Cert file path (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
              <input v-model="addOutboundForm.tlsKeyFile" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Key file path (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            </div>
            <label class="flex items-center gap-2 text-sm">
              <input v-model="addOutboundForm.tlsRejectUnknownSni" type="checkbox" @change="syncAddOutboundJsonFromForm" />
              <span>Reject Unknown SNI</span>
            </label>
            <label class="flex items-center gap-2 text-sm">
              <input v-model="addOutboundForm.tlsAllowInsecure" type="checkbox" @change="syncAddOutboundJsonFromForm" />
              <span>Allow Insecure</span>
            </label>
          </div>
          <div v-if="streamCapableProtocols.has(selectedAddOutboundProtocol) && addOutboundForm.security === 'reality'" class="grid grid-cols-1 gap-2 rounded-md border p-2.5">
            <div class="text-xs font-medium text-muted-foreground">Reality Settings</div>
            <label class="flex items-center gap-2 text-sm">
              <input v-model="addOutboundForm.realityShow" type="checkbox" @change="syncAddOutboundJsonFromForm" />
              <span>Show</span>
            </label>
            <input v-model="addOutboundForm.realityServerName" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Target / Server Name" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.realityFingerprint" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="uTLS fingerprint (chrome/firefox/...)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.realityPublicKey" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Public key" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.realityShortId" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Short ID" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.realitySpiderX" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="SpiderX (/)" type="text" @input="syncAddOutboundJsonFromForm" />
            <input v-model="addOutboundForm.realityMldsa65Verify" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="mldsa65Verify (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
          </div>
          <div class="grid grid-cols-1 gap-2 rounded-md border p-2.5">
            <div class="text-xs font-medium text-muted-foreground">Advanced transport</div>
            <input v-model="addOutboundForm.udpMasks" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="xudpProxyUDP443 (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            <label v-if="streamCapableProtocols.has(selectedAddOutboundProtocol)" class="flex items-center gap-2 text-sm">
              <input v-model="addOutboundForm.httpObfs" type="checkbox" @change="syncAddOutboundJsonFromForm" />
              <span>{{ t("camouflage", "HTTP Obfuscation") }}</span>
            </label>
            <div v-if="streamCapableProtocols.has(selectedAddOutboundProtocol) && addOutboundForm.httpObfs" class="grid grid-cols-1 gap-2">
              <input v-model="addOutboundForm.httpObfsHostCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="HTTP obfs Host CSV (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
              <input v-model="addOutboundForm.httpObfsPathCsv" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="HTTP obfs path CSV (default /)" type="text" @input="syncAddOutboundJsonFromForm" />
            </div>
            <label v-if="streamCapableProtocols.has(selectedAddOutboundProtocol)" class="flex items-center gap-2 text-sm">
              <input v-model="addOutboundForm.sockopt" type="checkbox" @change="syncAddOutboundJsonFromForm" />
              <span>{{ t("pages.xray.ui.sockopts", "Sockopts") }}</span>
            </label>
            <label v-if="streamCapableProtocols.has(selectedAddOutboundProtocol) && addOutboundForm.sockopt" class="grid gap-1 text-sm">
              <span class="text-muted-foreground">Sockopt JSON</span>
              <textarea v-model="addOutboundForm.sockoptJson" class="h-20 w-full json-field px-3 py-2 text-xs" @input="syncAddOutboundJsonFromForm" />
            </label>
            <label class="flex items-center gap-2 text-sm">
              <input v-model="addOutboundForm.mux" type="checkbox" @change="syncAddOutboundJsonFromForm" />
              <span>{{ t("mux", "Mux") }}</span>
            </label>
            <div v-if="addOutboundForm.mux" class="grid grid-cols-1 gap-2">
              <input v-model="addOutboundForm.muxConcurrency" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Mux concurrency (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
              <input v-model="addOutboundForm.muxXudpConcurrency" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Mux xudpConcurrency (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
              <input v-model="addOutboundForm.muxXudpProxyUDP443" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Mux xudpProxyUDP443 (optional)" type="text" @input="syncAddOutboundJsonFromForm" />
            </div>
          </div>
          <label class="grid gap-1 text-sm">
            <span class="text-muted-foreground">Settings JSON (protocol-specific)</span>
            <textarea v-model="addOutboundForm.customSettingsJson" class="h-28 w-full json-field px-3 py-2 text-xs" @input="syncAddOutboundJsonFromForm" />
          </label>
          <label class="grid gap-1 text-sm">
            <span class="text-muted-foreground">Stream Settings JSON (optional)</span>
            <textarea v-model="addOutboundForm.customStreamSettingsJson" class="h-24 w-full json-field px-3 py-2 text-xs" @input="syncAddOutboundJsonFromForm" />
          </label>
        </div>

        <div v-else class="space-y-2">
          <div class="flex items-center gap-2">
            <input v-model="addOutboundLink" class="flex-1 rounded-md border bg-background px-3 py-2 text-sm" placeholder="Link: vmess:// vless:// trojan:// ..." type="text" />
            <UiButton :disabled="busy" size="sm" variant="outline" @click="importOutboundLink">{{ t("import", "Import Link") }}</UiButton>
          </div>
          <textarea v-model="addOutboundJson" class="h-[46vh] w-full json-field px-3 py-2 text-xs" />
        </div>

        <div class="mt-4 flex justify-end gap-2">
          <UiButton size="sm" variant="outline" @click="addOutboundModalOpen = false">{{ t("close", "Close") }}</UiButton>
          <UiButton :disabled="busy" size="sm" @click="addOutboundFromModal">{{ t("pages.xray.outbound.addOutbound", "Add Outbound") }}</UiButton>
        </div>
      </div>
    </div>
    <div v-if="editOutboundModalOpen" class="fixed inset-0 z-[85] flex items-center justify-center bg-black/50 p-3 sm:p-4">
      <div class="w-full max-w-2xl rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-5">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold">{{ t("pages.xray.ui.editOutboundJson", "Edit Outbound JSON") }}</h2>
        </div>
        <textarea v-model="editOutboundJson" class="h-[52vh] w-full json-field px-3 py-2 text-xs" />
        <div class="mt-4 flex justify-end gap-2">
          <UiButton size="sm" variant="outline" @click="editOutboundModalOpen = false">{{ t("close", "Close") }}</UiButton>
          <UiButton :disabled="busy" size="sm" @click="saveEditOutboundModal">{{ t("save", "Save Outbound") }}</UiButton>
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
