<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { MoreVertical, Search } from "lucide-vue-next";
import { UiButton } from "@/components/ui/button";
import { UiCard, UiCardContent, UiCardDescription, UiCardHeader, UiCardTitle } from "@/components/ui/card";
import { t, loadI18n } from "@/lib/i18n";

type ApiResponse<T> = {
  success: boolean;
  obj: T;
  msg?: string;
};

type InboundClient = {
  id?: string;
  email?: string;
  password?: string;
  flow?: string;
  security?: string;
  enable?: boolean;
  totalGB?: number;
  expiryTime?: number;
  limitIp?: number;
  subId?: string;
  comment?: string;
  reset?: number;
  tgId?: number;
};

type ClientTraffic = {
  email?: string;
  up?: number;
  down?: number;
  total?: number;
  inboundId?: number;
};

type RawInbound = {
  id: number;
  remark?: string;
  enable?: boolean;
  protocol?: string;
  listen?: string;
  port?: number;
  up?: number;
  down?: number;
  total?: number;
  expiryTime?: number;
  settings?: string;
  streamSettings?: string;
  sniffing?: string;
  tag?: string;
  sort?: number;
};

type InboundClientTrafficRow = {
  email?: string;
  up?: number;
  down?: number;
  total?: number;
  expiryTime?: number;
  enable?: boolean;
  inboundId?: number;
};

type InboundRow = {
  id: number;
  remark: string;
  enable: boolean;
  protocol: string;
  listen: string;
  port: number;
  up: number;
  down: number;
  total: number;
  expiryTime: number;
  settings: string;
  streamSettings: string;
  sniffing: string;
  tag: string;
  sort: number;
  clients: InboundClient[];
};

type InboundModalState = {
  open: boolean;
  mode: "add" | "edit";
  id: number | null;
  remark: string;
  protocol: string;
  listen: string;
  port: string;
  enable: boolean;
  total: string;
  expiryTime: string;
  settings: string;
  streamSettings: string;
  sniffing: string;
  showAdvancedJson: boolean;
};

type ImportModalState = {
  open: boolean;
  data: string;
};

type ClientsModalState = {
  open: boolean;
  inboundId: number | null;
  inboundRemark: string;
  protocol: string;
  clients: InboundClient[];
  originalClients: InboundClient[];
};

type ClientEditorState = {
  open: boolean;
  mode: "add" | "edit";
  index: number;
  email: string;
  id: string;
  password: string;
  flow: string;
  security: string;
  enable: boolean;
  totalGB: string;
  delayedStart: boolean;
  delayedExpireDays: string;
  expiryTime: string;
  limitIp: string;
  subId: string;
  comment: string;
  reset: string;
  tgId: string;
};

type ClientBulkState = {
  open: boolean;
  method: number;
  quantity: string;
  firstNum: string;
  lastNum: string;
  prefix: string;
  postfix: string;
  security: string;
  flow: string;
  subId: string;
  tgId: string;
  limitIp: string;
  totalGB: string;
  delayedStart: boolean;
  delayedExpireDays: string;
  expiryTime: string;
  reset: string;
};

const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";

const loading = ref(false);
const refreshing = ref(false);
const error = ref("");
const actionError = ref("");
const actionSuccess = ref("");
const search = ref("");
const rows = ref<InboundRow[]>([]);
const onlineClients = ref<string[]>([]);
const onlineModalOpen = ref(false);
const actionBusyId = ref<number | null>(null);
const globalBusy = ref(false);
const modalBusy = ref(false);
const importBusy = ref(false);
const clientIpLogs = ref("");
const clientIpLogsEmail = ref("");
const clientIpLogsModalOpen = ref(false);
const realityPublicKeyPreview = ref("");
const mlkemClientPreview = ref("");
const mldsaVerifyPreview = ref("");
const echServerKeysPreview = ref("");
const echConfigListPreview = ref("");
const echSniInput = ref("");
const depletedModalOpen = ref(false);
const disabledModalOpen = ref(false);
const depletedClients = ref<Array<Record<string, unknown>>>([]);
const disabledClients = ref<Array<Record<string, unknown>>>([]);
const depletedIndex = ref<Record<number, number>>({});
const disabledIndex = ref<Record<number, number>>({});
const clientTrafficMap = ref<Record<string, ClientTraffic>>({});
const detailModalOpen = ref(false);
const detailLoading = ref(false);
const detailInboundJson = ref("{}");
const detailClientTraffics = ref<InboundClientTrafficRow[]>([]);
const clientSortBy = ref<"default" | "email" | "traffic" | "expiry">("default");
const clientSearch = ref("");
const subEnabled = ref(false);
const subBaseUri = ref("");
const remarkModel = ref("-ieo");
const clientInfoModalOpen = ref(false);
const clientInfoData = ref<InboundClient | null>(null);
const clientQrModalOpen = ref(false);
const clientQrValue = ref("");
const clientQrImageSrc = ref("");
const clientConfigsModalOpen = ref(false);
const clientConfigsText = ref("");
const clientConfigsQrValue = ref("");
const clientConfigsQrImageSrc = ref("");
const datepickerMode = ref<"gregorian" | "persian">("gregorian");
const inboundExpiryInput = ref("");
const clientBulkExpiryInput = ref("");
const clientEditorExpiryInput = ref("");
const confirmModalOpen = ref(false);
const confirmModalTitle = ref(t("confirm", "Confirm"));
const confirmModalMessage = ref("");
const firstLoadDone = ref(false);
const autoRefresh = ref(true);
const autoRefreshSeconds = ref(5);
let qriousLoader: Promise<void> | null = null;
let confirmResolver: ((value: boolean) => void) | null = null;
let refreshTimer: number | null = null;
let inboundsRequestInFlight = false;
let notifyTimer: number | null = null;
const clientActionsFloating = reactive<{
  open: boolean;
  client: InboundClient | null;
  left: number;
  top: number;
  place: "top" | "bottom";
}>({
  open: false,
  client: null,
  left: 0,
  top: 0,
  place: "bottom"
});

const highlightedClientConfigsHtml = computed(() => {
  return escapeHtml(clientConfigsText.value || "");
});

const highlightedDetailInboundHtml = computed(() => {
  return highlightJsonForHtml(detailInboundJson.value || "{}");
});

const highlightedClientInfoHtml = computed(() => {
  return highlightJsonForHtml(JSON.stringify(clientInfoData.value || {}, null, 2));
});

const clientSummary = computed(() => {
  let totalClients = 0;
  let onlineClientsCount = 0;
  for (const row of rows.value) {
    totalClients += row.clients.length;
    onlineClientsCount += getOnlineCount(row.clients);
  }
  const disabledClientsCount = Object.values(disabledIndex.value).reduce((a, b) => a + b, 0);
  const depletedClientsCount = Object.values(depletedIndex.value).reduce((a, b) => a + b, 0);
  return {
    totalClients,
    onlineClientsCount,
    disabledClientsCount,
    depletedClientsCount
  };
});

const inboundTrafficSummary = computed(() => {
  let up = 0;
  let down = 0;
  for (const row of rows.value) {
    up += Number(row.up || 0);
    down += Number(row.down || 0);
  }
  return {
    up,
    down,
    total: up + down
  };
});

const floatingNotice = computed(() => {
  if (actionError.value) {
    return { type: "error" as const, text: actionError.value };
  }
  if (actionSuccess.value) {
    return { type: "success" as const, text: actionSuccess.value };
  }
  return null;
});

watch([actionSuccess, actionError], ([nextSuccess, nextError]) => {
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
    actionSuccess.value = "";
    actionError.value = "";
    notifyTimer = null;
  }, 3500);
});

const inboundModal = reactive<InboundModalState>({
  open: false,
  mode: "add",
  id: null,
  remark: "",
  protocol: "vless",
  listen: "",
  port: "",
  enable: true,
  total: "0",
  expiryTime: "0",
  settings: '{"clients":[]}',
  streamSettings: "{}",
  sniffing: "{}",
  showAdvancedJson: false
});

const inboundStructured = reactive({
  decryption: "none",
  vlessEncryption: "",
  vlessSelectedAuth: "",
  vlessEncOptions: [] as Array<{ label: string; decryption?: string; encryption?: string }>,
  vlessFallbacks: [] as Array<{
    name: string;
    alpn: string;
    path: string;
    dest: string;
    xver: number;
  }>,
  trojanFallbacks: [] as Array<{
    name: string;
    alpn: string;
    path: string;
    dest: string;
    xver: number;
  }>,
  ssMethod: "aes-128-gcm",
  ssPassword: "",
  ssNetwork: "tcp,udp",
  ssIvCheck: false,
  httpTimeout: 300,
  httpAllowTransparent: false,
  httpAccounts: [{ user: "", pass: "" }] as Array<{ user: string; pass: string }>,
  mixedTimeout: 300,
  mixedAuth: "password",
  mixedUdp: false,
  mixedIp: "127.0.0.1",
  mixedAccounts: [{ user: "", pass: "" }] as Array<{ user: string; pass: string }>,
  streamNetwork: "tcp",
  streamSecurity: "none",
  tcpAcceptProxyProtocol: false,
  tcpHeaderType: "none",
  tcpRequestMethod: "GET",
  tcpRequestPath: "/",
  tcpRequestHeaders: [{ name: "", value: "" }] as Array<{ name: string; value: string }>,
  tcpResponseStatus: "200",
  tcpResponseReason: "OK",
  tcpResponseHeaders: [{ name: "", value: "" }] as Array<{ name: string; value: string }>,
  wsAcceptProxyProtocol: false,
  wsPath: "/",
  wsHost: "",
  wsHeaders: [{ name: "", value: "" }] as Array<{ name: string; value: string }>,
  wsHeartbeatPeriod: 0,
  grpcServiceName: "",
  grpcAuthority: "",
  grpcMultiMode: false,
  tlsServerName: "",
  tlsAlpn: "",
  tlsRejectUnknownSni: false,
  tlsMinVersion: "",
  tlsMaxVersion: "",
  tlsCipherSuites: "",
  tlsFingerprint: "",
  tlsAllowInsecure: false,
  tlsCertFile: "",
  tlsKeyFile: "",
  realityShow: false,
  realityDest: "",
  realityServerNames: "",
  realityPrivateKey: "",
  realityShortIds: "",
  realityMaxTimeDiff: 0,
  realityMinClient: "",
  realityMaxClient: "",
  sockoptEnabled: false,
  sockoptAcceptProxyProtocol: false,
  sockoptTcpFastOpen: false,
  sockoptTcpKeepAliveInterval: 0,
  sockoptMark: 0,
  sockoptDomainStrategy: "",
  sockoptDialerProxy: "",
  externalProxyEnabled: false,
  externalProxyAddress: "",
  externalProxyPort: "",
  externalProxyRemark: "",
  sniffEnabled: true,
  sniffHttp: true,
  sniffTls: true,
  sniffQuic: false,
  sniffFakeDns: false,
  sniffMetadataOnly: false,
  sniffRouteOnly: false,
  tunnelAddress: "",
  tunnelPort: 0,
  tunnelPortMap: [{ from: "", to: "" }] as Array<{ from: string; to: string }>,
  tunnelNetwork: "tcp,udp",
  tunnelFollowRedirect: false,
  tunName: "xray0",
  tunMtu: 1500,
  tunUserLevel: 0,
  wgSecretKey: "",
  wgPubKey: "",
  wgMtu: 1420,
  wgNoKernelTun: false,
  wgPeerPrivateKey: "",
  wgPeerPublicKey: "",
  wgPeerPsk: "",
  wgAllowedIPs: "0.0.0.0/0,::/0",
  wgKeepAlive: 0,
  kcpMtu: 1350,
  kcpTti: 50,
  kcpUpCap: 12,
  kcpDownCap: 100,
  kcpCongestion: false,
  kcpReadBuffer: 2,
  kcpWriteBuffer: 2,
  huAcceptProxyProtocol: false,
  huHost: "",
  huPath: "/",
  huHeaders: [] as Array<{ name: string; value: string }>,
  xhHost: "",
  xhPath: "/",
  xhMode: "auto",
  xhPaddingBytes: "",
  xhNoSSEHeader: false,
  xhHeaders: [] as Array<{ name: string; value: string }>,
  xhUplinkHTTPMethod: "",
  xhSessionPlacement: "",
  xhSessionKey: "",
  xhSeqPlacement: "",
  xhSeqKey: "",
  xhUplinkDataPlacement: "",
  xhUplinkDataKey: "",
  xhUplinkChunkSize: 0,
  xhPaddingObfsMode: false,
  xhPaddingKey: "",
  xhPaddingHeader: "",
  xhPaddingPlacement: "",
  xhPaddingMethod: "",
  xhScMaxBufferedPosts: 0,
  xhScMaxEachPostBytes: "",
  xhScStreamUpServerSecs: ""
});

const vmessSecurityOptions = ["auto", "none", "aes-128-gcm", "chacha20-poly1305", "zero"];

const clientsModal = reactive<ClientsModalState>({
  open: false,
  inboundId: null,
  inboundRemark: "",
  protocol: "",
  clients: [],
  originalClients: []
});

const importModal = reactive<ImportModalState>({
  open: false,
  data: ""
});

const clientEditor = reactive<ClientEditorState>({
  open: false,
  mode: "add",
  index: -1,
  email: "",
  id: "",
  password: "",
  flow: "",
  security: "",
  enable: true,
  totalGB: "0",
  delayedStart: false,
  delayedExpireDays: "0",
  expiryTime: "0",
  limitIp: "0",
  subId: "",
  comment: "",
  reset: "0",
  tgId: "0"
});

const clientBulk = reactive<ClientBulkState>({
  open: false,
  method: 0,
  quantity: "1",
  firstNum: "1",
  lastNum: "1",
  prefix: "",
  postfix: "",
  security: "auto",
  flow: "",
  subId: "",
  tgId: "0",
  limitIp: "0",
  totalGB: "0",
  delayedStart: false,
  delayedExpireDays: "0",
  expiryTime: "0",
  reset: "0"
});

function sizeFormat(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const idx = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, idx)).toFixed(idx === 0 ? 0 : 2)} ${units[idx]}`;
}

function trafficText(row: InboundRow): string {
  const used = (row.up || 0) + (row.down || 0);
  if (!row.total || row.total <= 0) {
    return `${sizeFormat(used)} / Unlimited`;
  }
  return `${sizeFormat(used)} / ${sizeFormat(row.total)}`;
}

function formatTotalGB(totalGB?: number): string {
  const value = Number(totalGB || 0);
  if (!Number.isFinite(value) || value <= 0) return "Unlimited";
  const gbValue = value / (1024 * 1024 * 1024);
  const normalized = gbValue.toFixed(2).replace(/\.?0+$/, "");
  return `${normalized} GB`;
}

function parseSettingsObject(settingsText?: string): Record<string, unknown> {
  if (!settingsText) return { clients: [] };
  try {
    const parsed = JSON.parse(settingsText);
    if (parsed && typeof parsed === "object") {
      return parsed as Record<string, unknown>;
    }
  } catch {
    return { clients: [] };
  }
  return { clients: [] };
}

function parseClients(settingsText?: string): InboundClient[] {
  const settingsObj = parseSettingsObject(settingsText);
  const clients = settingsObj.clients;
  if (!Array.isArray(clients)) return [];
  return clients as InboundClient[];
}

function clientIdentity(client: InboundClient, protocol: string): string {
  if (protocol === "trojan") {
    return String(client.password || "").trim();
  }
  if (protocol === "shadowsocks") {
    return String(client.email || "").trim();
  }
  return String(client.id || "").trim();
}

function normalizeClient(client: InboundClient): InboundClient {
  return {
    email: String(client.email || "").trim(),
    id: String(client.id || "").trim(),
    password: String(client.password || "").trim(),
    flow: String(client.flow || "").trim(),
    security: String(client.security || "").trim(),
    enable: client.enable ?? true,
    totalGB: Number(client.totalGB || 0),
    expiryTime: Number(client.expiryTime || 0),
    limitIp: Number(client.limitIp || 0),
    subId: String(client.subId || "").trim(),
    comment: String(client.comment || "").trim(),
    reset: Number(client.reset || 0),
    tgId: Number(client.tgId || 0)
  };
}

function clientEqual(a: InboundClient, b: InboundClient): boolean {
  return JSON.stringify(normalizeClient(a)) === JSON.stringify(normalizeClient(b));
}

function normalizeRow(raw: RawInbound): InboundRow {
  return {
    id: raw.id,
    remark: raw.remark || `Inbound #${raw.id}`,
    enable: raw.enable ?? true,
    protocol: raw.protocol || "unknown",
    listen: raw.listen || "",
    port: raw.port || 0,
    up: raw.up || 0,
    down: raw.down || 0,
    total: raw.total || 0,
    expiryTime: raw.expiryTime || 0,
    settings: raw.settings || "{}",
    streamSettings: raw.streamSettings || "{}",
    sniffing: raw.sniffing || "{}",
    tag: raw.tag || "",
    sort: raw.sort || 0,
    clients: parseClients(raw.settings)
  };
}

function isClientOnline(email?: string): boolean {
  const key = String(email || "").trim().toLowerCase();
  if (!key) return false;
  return onlineClients.value.some((x) => String(x || "").trim().toLowerCase() === key);
}

function getOnlineCount(clients: InboundClient[]): number {
  let count = 0;
  for (const c of clients) {
    if (isClientOnline(c.email)) {
      count += 1;
    }
  }
  return count;
}

function formatExpiry(unixMillis: number): string {
  if (!unixMillis || unixMillis <= 0) return "Never";
  const dt = new Date(unixMillis);
  if (Number.isNaN(dt.getTime())) return "-";
  if (datepickerMode.value === "persian") {
    return new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
      dateStyle: "medium",
      timeStyle: "short"
    }).format(dt);
  }
  return dt.toLocaleString();
}

function setDatepickerMode(raw: unknown): void {
  datepickerMode.value = String(raw || "").toLowerCase() === "persian" ? "persian" : "gregorian";
}

function unixMillisToDateTimeInput(value: string | number): string {
  const unixMillis = Number(value || 0);
  if (!Number.isFinite(unixMillis) || unixMillis <= 0) return "";
  const dt = new Date(unixMillis);
  if (Number.isNaN(dt.getTime())) return "";
  const tzOffsetMs = dt.getTimezoneOffset() * 60000;
  return new Date(dt.getTime() - tzOffsetMs).toISOString().slice(0, 16);
}

function dateTimeInputToUnixMillis(value: string): string {
  const text = String(value || "").trim();
  if (!text) return "0";
  const dt = new Date(text);
  if (Number.isNaN(dt.getTime())) return "0";
  return String(dt.getTime());
}

function syncInboundExpiryInputFromModel(): void {
  inboundExpiryInput.value = unixMillisToDateTimeInput(inboundModal.expiryTime);
}

function onInboundExpiryInputChange(): void {
  inboundModal.expiryTime = dateTimeInputToUnixMillis(inboundExpiryInput.value);
}

function syncClientBulkExpiryInputFromModel(): void {
  clientBulkExpiryInput.value = unixMillisToDateTimeInput(clientBulk.expiryTime);
}

function onClientBulkExpiryInputChange(): void {
  clientBulk.expiryTime = dateTimeInputToUnixMillis(clientBulkExpiryInput.value);
}

function syncClientEditorExpiryInputFromModel(): void {
  clientEditorExpiryInput.value = unixMillisToDateTimeInput(clientEditor.expiryTime);
}

function onClientEditorExpiryInputChange(): void {
  clientEditor.expiryTime = dateTimeInputToUnixMillis(clientEditorExpiryInput.value);
}

function openConfirmModal(message: string, title = "Confirm Action"): Promise<boolean> {
  confirmModalTitle.value = title;
  confirmModalMessage.value = message;
  confirmModalOpen.value = true;
  return new Promise<boolean>((resolve) => {
    confirmResolver = resolve;
  });
}

function resolveConfirmModal(value: boolean): void {
  confirmModalOpen.value = false;
  const resolver = confirmResolver;
  confirmResolver = null;
  if (resolver) {
    resolver(value);
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
  if (!autoRefresh.value) return;
  const sec = Math.max(2, Number(autoRefreshSeconds.value || 5));
  refreshTimer = window.setInterval(() => {
    if (document.visibilityState !== "visible") {
      return;
    }
    if (!loading.value && !refreshing.value && !modalBusy.value && !importBusy.value) {
      void refresh();
    }
  }, sec * 1000);
}

function onAutoRefreshChange(): void {
  try {
    localStorage.setItem("inboundsAutoRefresh", autoRefresh.value ? "1" : "0");
    localStorage.setItem("inboundsAutoRefreshSeconds", String(autoRefreshSeconds.value));
  } catch {
    // ignore storage failures
  }
  startAutoRefresh();
}

function jsonPrettyOrDefault(input: string, fallback: string): string {
  try {
    return JSON.stringify(JSON.parse(input), null, 2);
  } catch {
    return fallback;
  }
}

function numberFromString(value: string, fallback = 0): number {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) return fallback;
  return parsed;
}

function validJSON(input: string): boolean {
  try {
    JSON.parse(input);
    return true;
  } catch {
    return false;
  }
}

function generateId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

function protocolDefaultSettings(protocol: string): string {
  if (protocol === "vless") {
    return JSON.stringify({ decryption: "none", clients: [] }, null, 2);
  }
  if (protocol === "vmess") {
    return JSON.stringify({ clients: [] }, null, 2);
  }
  if (protocol === "trojan") {
    return JSON.stringify({ clients: [] }, null, 2);
  }
  if (protocol === "shadowsocks") {
    return JSON.stringify({ method: "aes-128-gcm", password: "", network: "tcp,udp", clients: [] }, null, 2);
  }
  if (protocol === "http") {
    return JSON.stringify({ timeout: 300, accounts: [] }, null, 2);
  }
  if (protocol === "mixed") {
    return JSON.stringify({ timeout: 300 }, null, 2);
  }
  if (protocol === "wireguard") {
    return JSON.stringify(
      {
        secretKey: "",
        pubKey: "",
        mtu: 1420,
        noKernelTun: false,
        peers: [
          {
            privateKey: "",
            publicKey: "",
            psk: "",
            allowedIPs: ["0.0.0.0/0", "::/0"],
            keepAlive: 0
          }
        ]
      },
      null,
      2
    );
  }
  if (protocol === "tunnel") {
    return JSON.stringify(
      {
        address: "",
        port: 0,
        network: "tcp,udp",
        followRedirect: false,
        portMap: []
      },
      null,
      2
    );
  }
  if (protocol === "tun") {
    return JSON.stringify(
      {
        name: "xray0",
        mtu: 1500,
        userLevel: 0
      },
      null,
      2
    );
  }
  return JSON.stringify({ clients: [] }, null, 2);
}

function protocolDefaultStreamSettings(): string {
  return JSON.stringify({ network: "tcp", security: "none", tcpSettings: {} }, null, 2);
}

function protocolDefaultSniffing(): string {
  return JSON.stringify({ enabled: true, destOverride: ["http", "tls"] }, null, 2);
}

function applyProtocolTemplate(protocol: string): void {
  inboundModal.settings = protocolDefaultSettings(protocol);
  inboundModal.streamSettings = protocolDefaultStreamSettings();
  inboundModal.sniffing = protocolDefaultSniffing();
  hydrateStructuredFromJson();
}

function hydrateStructuredFromJson(): void {
  const settings = parseSettingsObject(inboundModal.settings);
  inboundStructured.decryption = String(settings.decryption ?? "none");
  inboundStructured.vlessEncryption = String(settings.encryption ?? "");
  inboundStructured.vlessSelectedAuth = String(settings.selectedAuth ?? "");
  const vlessFallbacks = Array.isArray(settings.fallbacks) ? (settings.fallbacks as Record<string, unknown>[]) : [];
  inboundStructured.vlessFallbacks = vlessFallbacks.map((f) => ({
    name: String(f.name ?? ""),
    alpn: String(f.alpn ?? ""),
    path: String(f.path ?? ""),
    dest: String(f.dest ?? ""),
    xver: Number(f.xver ?? 0)
  }));
  const trojanFallbacks = Array.isArray(settings.fallbacks) ? (settings.fallbacks as Record<string, unknown>[]) : [];
  inboundStructured.trojanFallbacks = trojanFallbacks.map((f) => ({
    name: String(f.name ?? ""),
    alpn: String(f.alpn ?? ""),
    path: String(f.path ?? ""),
    dest: String(f.dest ?? ""),
    xver: Number(f.xver ?? 0)
  }));
  inboundStructured.ssMethod = String(settings.method ?? "aes-128-gcm");
  inboundStructured.ssPassword = String(settings.password ?? "");
  inboundStructured.ssNetwork = String(settings.network ?? "tcp,udp");
  inboundStructured.ssIvCheck = Boolean(settings.ivCheck ?? false);
  inboundStructured.httpTimeout = Number(settings.timeout ?? 300);
  inboundStructured.httpAllowTransparent = Boolean(settings.allowTransparent ?? false);
  const httpAccounts = Array.isArray(settings.accounts) ? (settings.accounts as Record<string, unknown>[]) : [];
  inboundStructured.httpAccounts = httpAccounts.map((a) => ({
    user: String(a.user ?? ""),
    pass: String(a.pass ?? "")
  }));
  if (inboundStructured.httpAccounts.length === 0) {
    inboundStructured.httpAccounts = [{ user: "", pass: "" }];
  }
  inboundStructured.mixedTimeout = Number(settings.timeout ?? 300);
  inboundStructured.mixedAuth = String(settings.auth ?? "password");
  inboundStructured.mixedUdp = Boolean(settings.udp ?? false);
  inboundStructured.mixedIp = String(settings.ip ?? "127.0.0.1");
  const mixedAccounts = Array.isArray(settings.accounts) ? (settings.accounts as Record<string, unknown>[]) : [];
  inboundStructured.mixedAccounts = mixedAccounts.map((a) => ({
    user: String(a.user ?? ""),
    pass: String(a.pass ?? "")
  }));
  if (inboundStructured.mixedAccounts.length === 0) {
    inboundStructured.mixedAccounts = [{ user: "", pass: "" }];
  }

  const stream = parseSettingsObject(inboundModal.streamSettings);
  inboundStructured.streamNetwork = String(stream.network ?? "tcp");
  inboundStructured.streamSecurity = String(stream.security ?? "none");
  const tcp = (stream.tcpSettings as Record<string, unknown>) || {};
  inboundStructured.tcpAcceptProxyProtocol = Boolean(tcp.acceptProxyProtocol ?? false);
  const tcpHeader = (tcp.header as Record<string, unknown>) || {};
  inboundStructured.tcpHeaderType = String(tcpHeader.type ?? "none");
  const tcpReq = (tcpHeader.request as Record<string, unknown>) || {};
  inboundStructured.tcpRequestMethod = String(tcpReq.method ?? "GET");
  const reqPath = Array.isArray(tcpReq.path) ? (tcpReq.path as string[]) : ["/"];
  inboundStructured.tcpRequestPath = reqPath.length > 0 ? String(reqPath[0] ?? "/") : "/";
  const reqHeaders = (tcpReq.headers as Record<string, unknown>) || {};
  inboundStructured.tcpRequestHeaders = Object.entries(reqHeaders).flatMap(([k, v]) => {
    if (Array.isArray(v)) return (v as unknown[]).map((x) => ({ name: k, value: String(x ?? "") }));
    return [{ name: k, value: String(v ?? "") }];
  });
  if (inboundStructured.tcpRequestHeaders.length === 0) {
    inboundStructured.tcpRequestHeaders = [{ name: "", value: "" }];
  }
  const tcpResp = (tcpHeader.response as Record<string, unknown>) || {};
  inboundStructured.tcpResponseStatus = String(tcpResp.status ?? "200");
  inboundStructured.tcpResponseReason = String(tcpResp.reason ?? "OK");
  const respHeaders = (tcpResp.headers as Record<string, unknown>) || {};
  inboundStructured.tcpResponseHeaders = Object.entries(respHeaders).flatMap(([k, v]) => {
    if (Array.isArray(v)) return (v as unknown[]).map((x) => ({ name: k, value: String(x ?? "") }));
    return [{ name: k, value: String(v ?? "") }];
  });
  if (inboundStructured.tcpResponseHeaders.length === 0) {
    inboundStructured.tcpResponseHeaders = [{ name: "", value: "" }];
  }
  const ws = (stream.wsSettings as Record<string, unknown>) || {};
  inboundStructured.wsAcceptProxyProtocol = Boolean(ws.acceptProxyProtocol ?? false);
  inboundStructured.wsPath = String(ws.path ?? "/");
  const wsHeadersObj = (ws.headers as Record<string, unknown>) || {};
  inboundStructured.wsHost = String(ws.host ?? wsHeadersObj.Host ?? "");
  inboundStructured.wsHeaders = Object.entries(wsHeadersObj).flatMap(([k, v]) => {
    if (Array.isArray(v)) return (v as unknown[]).map((x) => ({ name: k, value: String(x ?? "") }));
    return [{ name: k, value: String(v ?? "") }];
  });
  if (inboundStructured.wsHeaders.length === 0) {
    inboundStructured.wsHeaders = [{ name: "", value: "" }];
  }
  inboundStructured.wsHeartbeatPeriod = Number(ws.heartbeatPeriod ?? 0);
  const grpc = (stream.grpcSettings as Record<string, unknown>) || {};
  inboundStructured.grpcServiceName = String(grpc.serviceName ?? "");
  inboundStructured.grpcAuthority = String(grpc.authority ?? "");
  inboundStructured.grpcMultiMode = Boolean(grpc.multiMode ?? false);
  const tls = (stream.tlsSettings as Record<string, unknown>) || {};
  inboundStructured.tlsServerName = String(tls.serverName ?? "");
  const alpn = Array.isArray(tls.alpn) ? (tls.alpn as string[]) : [];
  inboundStructured.tlsAlpn = alpn.join(",");
  inboundStructured.tlsRejectUnknownSni = Boolean(tls.rejectUnknownSni ?? false);
  inboundStructured.tlsMinVersion = String(tls.minVersion ?? "");
  inboundStructured.tlsMaxVersion = String(tls.maxVersion ?? "");
  const ciphers = Array.isArray(tls.cipherSuites) ? (tls.cipherSuites as string[]) : [];
  inboundStructured.tlsCipherSuites = ciphers.join(",");
  inboundStructured.tlsFingerprint = String(tls.fingerprint ?? "");
  inboundStructured.tlsAllowInsecure = Boolean(tls.allowInsecure ?? false);
  const certs = Array.isArray(tls.certificates) ? (tls.certificates as Record<string, unknown>[]) : [];
  const cert0 = certs.length > 0 ? certs[0] : {};
  inboundStructured.tlsCertFile = String(cert0.certificateFile ?? "");
  inboundStructured.tlsKeyFile = String(cert0.keyFile ?? "");
  const reality = (stream.realitySettings as Record<string, unknown>) || {};
  inboundStructured.realityShow = Boolean(reality.show ?? false);
  inboundStructured.realityDest = String(reality.dest ?? "");
  const srvNames = Array.isArray(reality.serverNames) ? (reality.serverNames as string[]) : [];
  inboundStructured.realityServerNames = srvNames.join(",");
  inboundStructured.realityPrivateKey = String(reality.privateKey ?? "");
  const shortIds = Array.isArray(reality.shortIds) ? (reality.shortIds as string[]) : [];
  inboundStructured.realityShortIds = shortIds.join(",");
  inboundStructured.realityMaxTimeDiff = Number(reality.maxTimeDiff ?? 0);
  inboundStructured.realityMinClient = String(reality.minClient ?? "");
  inboundStructured.realityMaxClient = String(reality.maxClient ?? "");
  const sockopt = (stream.sockopt as Record<string, unknown>) || {};
  inboundStructured.sockoptEnabled = Object.keys(sockopt).length > 0;
  inboundStructured.sockoptAcceptProxyProtocol = Boolean(sockopt.acceptProxyProtocol ?? false);
  inboundStructured.sockoptTcpFastOpen = Boolean(sockopt.tcpFastOpen ?? false);
  inboundStructured.sockoptTcpKeepAliveInterval = Number(sockopt.tcpKeepAliveInterval ?? 0);
  inboundStructured.sockoptMark = Number(sockopt.mark ?? 0);
  inboundStructured.sockoptDomainStrategy = String(sockopt.domainStrategy ?? "");
  inboundStructured.sockoptDialerProxy = String(sockopt.dialerProxy ?? "");
  const dialerProxyRaw = inboundStructured.sockoptDialerProxy.trim();
  inboundStructured.externalProxyEnabled = dialerProxyRaw.length > 0;
  inboundStructured.externalProxyAddress = "";
  inboundStructured.externalProxyPort = "";
  inboundStructured.externalProxyRemark = "";
  if (dialerProxyRaw) {
    const [addrPort, remark] = dialerProxyRaw.split("#");
    const idx = addrPort.lastIndexOf(":");
    inboundStructured.externalProxyAddress = idx > 0 ? addrPort.slice(0, idx) : addrPort;
    inboundStructured.externalProxyPort = idx > 0 ? addrPort.slice(idx + 1) : "";
    inboundStructured.externalProxyRemark = remark || "";
  }
  const kcp = (stream.kcpSettings as Record<string, unknown>) || {};
  inboundStructured.kcpMtu = Number(kcp.mtu ?? 1350);
  inboundStructured.kcpTti = Number(kcp.tti ?? 50);
  inboundStructured.kcpUpCap = Number(kcp.uplinkCapacity ?? 12);
  inboundStructured.kcpDownCap = Number(kcp.downlinkCapacity ?? 100);
  inboundStructured.kcpCongestion = Boolean(kcp.congestion ?? false);
  inboundStructured.kcpReadBuffer = Number(kcp.readBufferSize ?? 2);
  inboundStructured.kcpWriteBuffer = Number(kcp.writeBufferSize ?? 2);
  const hu = (stream.httpupgradeSettings as Record<string, unknown>) || {};
  inboundStructured.huAcceptProxyProtocol = Boolean(hu.acceptProxyProtocol ?? false);
  inboundStructured.huHost = String(hu.host ?? "");
  inboundStructured.huPath = String(hu.path ?? "/");
  const huHeaders = Array.isArray(hu.headers) ? (hu.headers as Record<string, unknown>[]) : [];
  inboundStructured.huHeaders = huHeaders.map((h) => ({
    name: String(h.name ?? ""),
    value: String(h.value ?? "")
  }));
  if (inboundStructured.huHeaders.length === 0) {
    inboundStructured.huHeaders = [{ name: "", value: "" }];
  }
  const xh = (stream.xhttpSettings as Record<string, unknown>) || {};
  inboundStructured.xhHost = String(xh.host ?? "");
  inboundStructured.xhPath = String(xh.path ?? "/");
  inboundStructured.xhMode = String(xh.mode ?? "auto");
  inboundStructured.xhPaddingBytes = String(xh.xPaddingBytes ?? "");
  inboundStructured.xhNoSSEHeader = Boolean(xh.noSSEHeader ?? false);
  const xhHeaders = Array.isArray(xh.headers) ? (xh.headers as Record<string, unknown>[]) : [];
  inboundStructured.xhHeaders = xhHeaders.map((h) => ({
    name: String(h.name ?? ""),
    value: String(h.value ?? "")
  }));
  if (inboundStructured.xhHeaders.length === 0) {
    inboundStructured.xhHeaders = [{ name: "", value: "" }];
  }
  inboundStructured.xhUplinkHTTPMethod = String(xh.uplinkHTTPMethod ?? "");
  inboundStructured.xhSessionPlacement = String(xh.sessionPlacement ?? "");
  inboundStructured.xhSessionKey = String(xh.sessionKey ?? "");
  inboundStructured.xhSeqPlacement = String(xh.seqPlacement ?? "");
  inboundStructured.xhSeqKey = String(xh.seqKey ?? "");
  inboundStructured.xhUplinkDataPlacement = String(xh.uplinkDataPlacement ?? "");
  inboundStructured.xhUplinkDataKey = String(xh.uplinkDataKey ?? "");
  inboundStructured.xhUplinkChunkSize = Number(xh.uplinkChunkSize ?? 0);
  inboundStructured.xhPaddingObfsMode = Boolean(xh.xPaddingObfsMode ?? false);
  inboundStructured.xhPaddingKey = String(xh.xPaddingKey ?? "");
  inboundStructured.xhPaddingHeader = String(xh.xPaddingHeader ?? "");
  inboundStructured.xhPaddingPlacement = String(xh.xPaddingPlacement ?? "");
  inboundStructured.xhPaddingMethod = String(xh.xPaddingMethod ?? "");
  inboundStructured.xhScMaxBufferedPosts = Number(xh.scMaxBufferedPosts ?? 0);
  inboundStructured.xhScMaxEachPostBytes = String(xh.scMaxEachPostBytes ?? "");
  inboundStructured.xhScStreamUpServerSecs = String(xh.scStreamUpServerSecs ?? "");

  const sniff = parseSettingsObject(inboundModal.sniffing);
  inboundStructured.sniffEnabled = Boolean(sniff.enabled ?? true);
  const destOverride = Array.isArray(sniff.destOverride) ? (sniff.destOverride as string[]) : [];
  inboundStructured.sniffHttp = destOverride.includes("http");
  inboundStructured.sniffTls = destOverride.includes("tls");
  inboundStructured.sniffQuic = destOverride.includes("quic");
  inboundStructured.sniffFakeDns = destOverride.includes("fakedns");
  inboundStructured.sniffMetadataOnly = Boolean(sniff.metadataOnly ?? false);
  inboundStructured.sniffRouteOnly = Boolean(sniff.routeOnly ?? false);

  inboundStructured.tunnelAddress = String(settings.address ?? "");
  inboundStructured.tunnelPort = Number(settings.port ?? 0);
  const portMap = Array.isArray(settings.portMap) ? (settings.portMap as Record<string, unknown>[]) : [];
  inboundStructured.tunnelPortMap = portMap.map((m) => ({
    from: String(m.name ?? ""),
    to: String(m.value ?? "")
  }));
  if (inboundStructured.tunnelPortMap.length === 0) {
    inboundStructured.tunnelPortMap = [{ from: "", to: "" }];
  }
  inboundStructured.tunnelNetwork = String(settings.network ?? "tcp,udp");
  inboundStructured.tunnelFollowRedirect = Boolean(settings.followRedirect ?? false);

  inboundStructured.tunName = String(settings.name ?? "xray0");
  inboundStructured.tunMtu = Number(settings.mtu ?? 1500);
  inboundStructured.tunUserLevel = Number(settings.userLevel ?? 0);

  inboundStructured.wgSecretKey = String(settings.secretKey ?? "");
  inboundStructured.wgPubKey = String(settings.pubKey ?? "");
  inboundStructured.wgMtu = Number(settings.mtu ?? 1420);
  inboundStructured.wgNoKernelTun = Boolean(settings.noKernelTun ?? false);
  const peers = Array.isArray(settings.peers) ? (settings.peers as Record<string, unknown>[]) : [];
  const peer = peers.length > 0 ? peers[0] : {};
  inboundStructured.wgPeerPrivateKey = String(peer.privateKey ?? "");
  inboundStructured.wgPeerPublicKey = String(peer.publicKey ?? "");
  inboundStructured.wgPeerPsk = String(peer.psk ?? "");
  const allowedIPs = Array.isArray(peer.allowedIPs) ? (peer.allowedIPs as string[]) : ["0.0.0.0/0", "::/0"];
  inboundStructured.wgAllowedIPs = allowedIPs.join(",");
  inboundStructured.wgKeepAlive = Number(peer.keepAlive ?? 0);
}

function applyStructuredToJson(): void {
  const settings = parseSettingsObject(inboundModal.settings);
  if (inboundModal.protocol === "vless") {
    settings.decryption = inboundStructured.decryption || "none";
    settings.encryption = inboundStructured.vlessEncryption || undefined;
    settings.selectedAuth = inboundStructured.vlessSelectedAuth || undefined;
    if (inboundStructured.streamNetwork === "tcp" && !inboundStructured.vlessSelectedAuth) {
      const fallbacks = inboundStructured.vlessFallbacks
        .map((f) => ({
          name: f.name.trim(),
          alpn: f.alpn.trim(),
          path: f.path.trim(),
          dest: f.dest.trim(),
          xver: Number(f.xver || 0)
        }))
        .filter((f) => f.name || f.alpn || f.path || f.dest);
      if (fallbacks.length > 0) {
        settings.fallbacks = fallbacks;
      } else {
        delete settings.fallbacks;
      }
    } else {
      delete settings.fallbacks;
    }
  }
  if (inboundModal.protocol === "trojan") {
    if (inboundStructured.streamNetwork === "tcp") {
      const fallbacks = inboundStructured.trojanFallbacks
        .map((f) => ({
          name: f.name.trim(),
          alpn: f.alpn.trim(),
          path: f.path.trim(),
          dest: f.dest.trim(),
          xver: Number(f.xver || 0)
        }))
        .filter((f) => f.name || f.alpn || f.path || f.dest);
      if (fallbacks.length > 0) {
        settings.fallbacks = fallbacks;
      } else {
        delete settings.fallbacks;
      }
    } else {
      delete settings.fallbacks;
    }
  }
  if (inboundModal.protocol === "shadowsocks") {
    settings.method = inboundStructured.ssMethod;
    settings.password = inboundStructured.ssPassword;
    settings.network = inboundStructured.ssNetwork;
    settings.ivCheck = inboundStructured.ssIvCheck;
  }
  if (inboundModal.protocol === "http") {
    settings.timeout = Number(inboundStructured.httpTimeout || 300);
    settings.allowTransparent = inboundStructured.httpAllowTransparent;
    settings.accounts = inboundStructured.httpAccounts
      .map((a) => ({ user: a.user.trim(), pass: a.pass.trim() }))
      .filter((a) => a.user || a.pass);
  }
  if (inboundModal.protocol === "mixed") {
    settings.timeout = Number(inboundStructured.mixedTimeout || 300);
    settings.auth = inboundStructured.mixedAuth || "password";
    settings.udp = inboundStructured.mixedUdp;
    settings.ip = inboundStructured.mixedIp || "127.0.0.1";
    settings.accounts =
      settings.auth === "password"
        ? inboundStructured.mixedAccounts
            .map((a) => ({ user: a.user.trim(), pass: a.pass.trim() }))
            .filter((a) => a.user || a.pass)
        : [];
  }
  if (inboundModal.protocol === "tunnel") {
    settings.address = inboundStructured.tunnelAddress;
    settings.port = Number(inboundStructured.tunnelPort || 0);
    settings.portMap = inboundStructured.tunnelPortMap
      .map((m) => ({ name: m.from.trim(), value: m.to.trim() }))
      .filter((m) => m.name || m.value);
    settings.network = inboundStructured.tunnelNetwork || "tcp,udp";
    settings.followRedirect = inboundStructured.tunnelFollowRedirect;
  }
  if (inboundModal.protocol === "tun") {
    settings.name = inboundStructured.tunName || "xray0";
    settings.mtu = Number(inboundStructured.tunMtu || 1500);
    settings.userLevel = Number(inboundStructured.tunUserLevel || 0);
  }
  if (inboundModal.protocol === "wireguard") {
    settings.secretKey = inboundStructured.wgSecretKey;
    settings.pubKey = inboundStructured.wgPubKey;
    settings.mtu = Number(inboundStructured.wgMtu || 1420);
    settings.noKernelTun = inboundStructured.wgNoKernelTun;
    settings.peers = [
      {
        privateKey: inboundStructured.wgPeerPrivateKey,
        publicKey: inboundStructured.wgPeerPublicKey,
        psk: inboundStructured.wgPeerPsk,
        allowedIPs: inboundStructured.wgAllowedIPs
          .split(",")
          .map((x) => x.trim())
          .filter(Boolean),
        keepAlive: Number(inboundStructured.wgKeepAlive || 0)
      }
    ];
  }
  if (!Array.isArray(settings.clients)) {
    settings.clients = [];
  }
  inboundModal.settings = JSON.stringify(settings, null, 2);

  const stream = parseSettingsObject(inboundModal.streamSettings);
  stream.network = inboundStructured.streamNetwork || "tcp";
  stream.security = inboundStructured.streamSecurity || "none";
  if (inboundStructured.streamNetwork === "tcp") {
    const tcpReqHeaders = inboundStructured.tcpRequestHeaders
      .map((h) => ({ name: h.name.trim(), value: h.value.trim() }))
      .filter((h) => h.name || h.value);
    const tcpRespHeaders = inboundStructured.tcpResponseHeaders
      .map((h) => ({ name: h.name.trim(), value: h.value.trim() }))
      .filter((h) => h.name || h.value);
    const reqHeaderObj = tcpReqHeaders.reduce<Record<string, string[]>>((acc, h) => {
      if (!acc[h.name]) acc[h.name] = [];
      acc[h.name].push(h.value);
      return acc;
    }, {});
    const respHeaderObj = tcpRespHeaders.reduce<Record<string, string[]>>((acc, h) => {
      if (!acc[h.name]) acc[h.name] = [];
      acc[h.name].push(h.value);
      return acc;
    }, {});
    const isHttpHeader = inboundStructured.tcpHeaderType === "http";
    stream.tcpSettings = {
      acceptProxyProtocol: inboundStructured.tcpAcceptProxyProtocol,
      header: {
        type: inboundStructured.tcpHeaderType || "none",
        request: isHttpHeader
          ? {
              method: inboundStructured.tcpRequestMethod || "GET",
              path: [inboundStructured.tcpRequestPath || "/"],
              headers: reqHeaderObj
            }
          : undefined,
        response: isHttpHeader
          ? {
              status: inboundStructured.tcpResponseStatus || "200",
              reason: inboundStructured.tcpResponseReason || "OK",
              headers: respHeaderObj
            }
          : undefined
      }
    };
  }
  if (inboundStructured.streamNetwork === "ws") {
    const wsHeaders = inboundStructured.wsHeaders
      .map((h) => ({ name: h.name.trim(), value: h.value.trim() }))
      .filter((h) => h.name || h.value);
    const wsHeaderObj = wsHeaders.reduce<Record<string, string>>((acc, h) => {
      acc[h.name] = h.value;
      return acc;
    }, {});
    if (inboundStructured.wsHost) {
      wsHeaderObj.Host = inboundStructured.wsHost;
    }
    stream.wsSettings = {
      acceptProxyProtocol: inboundStructured.wsAcceptProxyProtocol,
      path: inboundStructured.wsPath || "/",
      host: inboundStructured.wsHost || "",
      headers: wsHeaderObj,
      heartbeatPeriod: Number(inboundStructured.wsHeartbeatPeriod || 0)
    };
  }
  if (inboundStructured.streamNetwork === "grpc") {
    stream.grpcSettings = {
      serviceName: inboundStructured.grpcServiceName || "",
      authority: inboundStructured.grpcAuthority || "",
      multiMode: inboundStructured.grpcMultiMode
    };
  }
  if (inboundStructured.streamNetwork === "kcp") {
    stream.kcpSettings = {
      mtu: Number(inboundStructured.kcpMtu || 1350),
      tti: Number(inboundStructured.kcpTti || 50),
      uplinkCapacity: Number(inboundStructured.kcpUpCap || 12),
      downlinkCapacity: Number(inboundStructured.kcpDownCap || 100),
      congestion: inboundStructured.kcpCongestion,
      readBufferSize: Number(inboundStructured.kcpReadBuffer || 2),
      writeBufferSize: Number(inboundStructured.kcpWriteBuffer || 2)
    };
  }
  if (inboundStructured.streamNetwork === "httpupgrade") {
    const huHeaders = inboundStructured.huHeaders
      .map((h) => ({ name: h.name.trim(), value: h.value.trim() }))
      .filter((h) => h.name || h.value);
    stream.httpupgradeSettings = {
      acceptProxyProtocol: inboundStructured.huAcceptProxyProtocol,
      host: inboundStructured.huHost || "",
      path: inboundStructured.huPath || "/",
      headers: huHeaders
    };
  }
  if (inboundStructured.streamNetwork === "xhttp") {
    const xhHeaders = inboundStructured.xhHeaders
      .map((h) => ({ name: h.name.trim(), value: h.value.trim() }))
      .filter((h) => h.name || h.value);
    stream.xhttpSettings = {
      host: inboundStructured.xhHost || "",
      path: inboundStructured.xhPath || "/",
      mode: inboundStructured.xhMode || "auto",
      xPaddingBytes: inboundStructured.xhPaddingBytes || "",
      noSSEHeader: inboundStructured.xhNoSSEHeader,
      headers: xhHeaders,
      xPaddingObfsMode: inboundStructured.xhPaddingObfsMode,
      xPaddingKey: inboundStructured.xhPaddingKey || undefined,
      xPaddingHeader: inboundStructured.xhPaddingHeader || undefined,
      xPaddingPlacement: inboundStructured.xhPaddingPlacement || undefined,
      xPaddingMethod: inboundStructured.xhPaddingMethod || undefined,
      scMaxBufferedPosts: Number(inboundStructured.xhScMaxBufferedPosts || 0),
      scMaxEachPostBytes: inboundStructured.xhScMaxEachPostBytes || undefined,
      scStreamUpServerSecs: inboundStructured.xhScStreamUpServerSecs || undefined,
      uplinkHTTPMethod: inboundStructured.xhUplinkHTTPMethod || undefined,
      sessionPlacement: inboundStructured.xhSessionPlacement || undefined,
      sessionKey: inboundStructured.xhSessionKey || undefined,
      seqPlacement: inboundStructured.xhSeqPlacement || undefined,
      seqKey: inboundStructured.xhSeqKey || undefined,
      uplinkDataPlacement: inboundStructured.xhUplinkDataPlacement || undefined,
      uplinkDataKey: inboundStructured.xhUplinkDataKey || undefined,
      uplinkChunkSize: Number(inboundStructured.xhUplinkChunkSize || 0)
    };
  }
  if (inboundStructured.streamSecurity === "tls") {
    const alpn = inboundStructured.tlsAlpn
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const cipherSuites = inboundStructured.tlsCipherSuites
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const certificates =
      inboundStructured.tlsCertFile || inboundStructured.tlsKeyFile
        ? [
            {
              certificateFile: inboundStructured.tlsCertFile || "",
              keyFile: inboundStructured.tlsKeyFile || ""
            }
          ]
        : [];
    stream.tlsSettings = {
      serverName: inboundStructured.tlsServerName || "",
      alpn,
      rejectUnknownSni: inboundStructured.tlsRejectUnknownSni,
      minVersion: inboundStructured.tlsMinVersion || undefined,
      maxVersion: inboundStructured.tlsMaxVersion || undefined,
      cipherSuites,
      fingerprint: inboundStructured.tlsFingerprint || undefined,
      allowInsecure: inboundStructured.tlsAllowInsecure,
      certificates
    };
    delete stream.realitySettings;
  } else if (inboundStructured.streamSecurity === "reality") {
    const serverNames = inboundStructured.realityServerNames
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    const shortIds = inboundStructured.realityShortIds
      .split(",")
      .map((x) => x.trim())
      .filter(Boolean);
    stream.realitySettings = {
      show: inboundStructured.realityShow,
      dest: inboundStructured.realityDest || "",
      serverNames,
      privateKey: inboundStructured.realityPrivateKey || "",
      shortIds,
      maxTimeDiff: Number(inboundStructured.realityMaxTimeDiff || 0),
      minClient: inboundStructured.realityMinClient || undefined,
      maxClient: inboundStructured.realityMaxClient || undefined
    };
    delete stream.tlsSettings;
  } else {
    delete stream.tlsSettings;
    delete stream.realitySettings;
  }
  const externalProxyAddress = inboundStructured.externalProxyAddress.trim();
  const externalProxyPort = inboundStructured.externalProxyPort.trim();
  const externalProxyRemark = inboundStructured.externalProxyRemark.trim();
  if (inboundStructured.externalProxyEnabled) {
    if (externalProxyAddress && externalProxyPort) {
      inboundStructured.sockoptDialerProxy = `${externalProxyAddress}:${externalProxyPort}${externalProxyRemark ? `#${externalProxyRemark}` : ""}`;
    } else {
      inboundStructured.sockoptDialerProxy = "";
    }
  } else {
    inboundStructured.sockoptDialerProxy = "";
  }
  if (inboundStructured.sockoptEnabled || inboundStructured.externalProxyEnabled) {
    stream.sockopt = {
      acceptProxyProtocol: inboundStructured.sockoptAcceptProxyProtocol,
      tcpFastOpen: inboundStructured.sockoptTcpFastOpen,
      tcpKeepAliveInterval: Number(inboundStructured.sockoptTcpKeepAliveInterval || 0),
      mark: Number(inboundStructured.sockoptMark || 0),
      domainStrategy: inboundStructured.sockoptDomainStrategy || undefined,
      dialerProxy: inboundStructured.sockoptDialerProxy || undefined
    };
  } else {
    delete stream.sockopt;
  }
  inboundModal.streamSettings = JSON.stringify(stream, null, 2);

  const sniffDest: string[] = [];
  if (inboundStructured.sniffHttp) sniffDest.push("http");
  if (inboundStructured.sniffTls) sniffDest.push("tls");
  if (inboundStructured.sniffQuic) sniffDest.push("quic");
  if (inboundStructured.sniffFakeDns) sniffDest.push("fakedns");
  inboundModal.sniffing = JSON.stringify(
    {
      enabled: inboundStructured.sniffEnabled,
      destOverride: sniffDest,
      metadataOnly: inboundStructured.sniffMetadataOnly,
      routeOnly: inboundStructured.sniffRouteOnly
    },
    null,
    2
  );
}

function addHuHeader(): void {
  inboundStructured.huHeaders.push({ name: "", value: "" });
}

function removeHuHeader(index: number): void {
  inboundStructured.huHeaders.splice(index, 1);
  if (inboundStructured.huHeaders.length === 0) {
    inboundStructured.huHeaders.push({ name: "", value: "" });
  }
}

function addXhHeader(): void {
  inboundStructured.xhHeaders.push({ name: "", value: "" });
}

function removeXhHeader(index: number): void {
  inboundStructured.xhHeaders.splice(index, 1);
  if (inboundStructured.xhHeaders.length === 0) {
    inboundStructured.xhHeaders.push({ name: "", value: "" });
  }
}

function addTcpRequestHeader(): void {
  inboundStructured.tcpRequestHeaders.push({ name: "", value: "" });
}

function removeTcpRequestHeader(index: number): void {
  inboundStructured.tcpRequestHeaders.splice(index, 1);
  if (inboundStructured.tcpRequestHeaders.length === 0) {
    inboundStructured.tcpRequestHeaders.push({ name: "", value: "" });
  }
}

function addTcpResponseHeader(): void {
  inboundStructured.tcpResponseHeaders.push({ name: "", value: "" });
}

function removeTcpResponseHeader(index: number): void {
  inboundStructured.tcpResponseHeaders.splice(index, 1);
  if (inboundStructured.tcpResponseHeaders.length === 0) {
    inboundStructured.tcpResponseHeaders.push({ name: "", value: "" });
  }
}

function addWsHeader(): void {
  inboundStructured.wsHeaders.push({ name: "", value: "" });
}

function removeWsHeader(index: number): void {
  inboundStructured.wsHeaders.splice(index, 1);
  if (inboundStructured.wsHeaders.length === 0) {
    inboundStructured.wsHeaders.push({ name: "", value: "" });
  }
}

function addHttpAccount(): void {
  inboundStructured.httpAccounts.push({ user: "", pass: "" });
}

function removeHttpAccount(index: number): void {
  inboundStructured.httpAccounts.splice(index, 1);
  if (inboundStructured.httpAccounts.length === 0) {
    inboundStructured.httpAccounts.push({ user: "", pass: "" });
  }
}

function addMixedAccount(): void {
  inboundStructured.mixedAccounts.push({ user: "", pass: "" });
}

function removeMixedAccount(index: number): void {
  inboundStructured.mixedAccounts.splice(index, 1);
  if (inboundStructured.mixedAccounts.length === 0) {
    inboundStructured.mixedAccounts.push({ user: "", pass: "" });
  }
}

function addTunnelPortMap(): void {
  inboundStructured.tunnelPortMap.push({ from: "", to: "" });
}

function removeTunnelPortMap(index: number): void {
  inboundStructured.tunnelPortMap.splice(index, 1);
  if (inboundStructured.tunnelPortMap.length === 0) {
    inboundStructured.tunnelPortMap.push({ from: "", to: "" });
  }
}

function addVlessFallback(): void {
  inboundStructured.vlessFallbacks.push({ name: "", alpn: "", path: "", dest: "", xver: 0 });
}

function removeVlessFallback(index: number): void {
  inboundStructured.vlessFallbacks.splice(index, 1);
}

function addTrojanFallback(): void {
  inboundStructured.trojanFallbacks.push({ name: "", alpn: "", path: "", dest: "", xver: 0 });
}

function removeTrojanFallback(index: number): void {
  inboundStructured.trojanFallbacks.splice(index, 1);
}

async function getNewVlessEncOptions(): Promise<void> {
  actionError.value = "";
  try {
    const res = await fetch(`${basePath}panel/api/server/getNewVlessEnc`, {
      method: "GET",
      credentials: "same-origin"
    });
    const msg = (await res.json()) as ApiResponse<{ auths?: Array<{ label?: string; decryption?: string; encryption?: string }> }>;
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load VLESS encryption options");
    }
    const auths = Array.isArray(msg.obj?.auths) ? msg.obj.auths : [];
    inboundStructured.vlessEncOptions = auths.map((a) => ({
      label: String(a.label || ""),
      decryption: a.decryption ? String(a.decryption) : "",
      encryption: a.encryption ? String(a.encryption) : ""
    }));
    if (!inboundStructured.vlessSelectedAuth && inboundStructured.vlessEncOptions.length > 0) {
      inboundStructured.vlessSelectedAuth = inboundStructured.vlessEncOptions[0].label;
      onVlessAuthChange();
    }
    actionSuccess.value = t("pages.inbounds.ui.vlessEncOptionsLoaded", "VLESS encryption options loaded");
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to load VLESS encryption options";
  }
}

function onVlessAuthChange(): void {
  if (!inboundStructured.vlessSelectedAuth) return;
  const found = inboundStructured.vlessEncOptions.find((x) => x.label === inboundStructured.vlessSelectedAuth);
  if (!found) return;
  if (found.decryption) inboundStructured.decryption = found.decryption;
  if (found.encryption) inboundStructured.vlessEncryption = found.encryption;
}

async function fetchServerUUID(): Promise<string> {
  const res = await fetch(`${basePath}panel/api/server/getNewUUID`, {
    method: "GET",
    credentials: "same-origin"
  });
  const data = (await res.json()) as ApiResponse<{ uuid?: string }>;
  if (!data.success) {
    throw new Error(data.msg || "UUID generation failed");
  }
  return data.obj?.uuid || generateId();
}

async function generateClientUuid(): Promise<void> {
  try {
    const id = await fetchServerUUID();
    clientEditor.id = id;
    actionSuccess.value = t("pages.inbounds.ui.newUuidGenerated", "New UUID generated");
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "UUID generation failed";
  }
}

function generateClientPassword(): void {
  clientEditor.password = generateId();
}

function generateClientSubId(): void {
  clientEditor.subId = Math.random().toString(36).slice(2, 18);
}

async function generateRealityKeypair(): Promise<void> {
  actionError.value = "";
  try {
    const res = await fetch(`${basePath}panel/api/server/getNewX25519Cert`, {
      method: "GET",
      credentials: "same-origin"
    });
    const msg = (await res.json()) as ApiResponse<{ privateKey?: string; publicKey?: string }>;
    if (!msg.success) {
      throw new Error(msg.msg || "Generate x25519 failed");
    }
    inboundStructured.realityPrivateKey = msg.obj?.privateKey || "";
    realityPublicKeyPreview.value = msg.obj?.publicKey || "";
    actionSuccess.value = t("pages.inbounds.ui.realityX25519Generated", "Reality x25519 keypair generated");
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Generate x25519 failed";
  }
}

async function generateMlkem768(): Promise<void> {
  actionError.value = "";
  try {
    const res = await fetch(`${basePath}panel/api/server/getNewmlkem768`, {
      method: "GET",
      credentials: "same-origin"
    });
    const msg = (await res.json()) as ApiResponse<{ seed?: string; client?: string }>;
    if (!msg.success) {
      throw new Error(msg.msg || "Generate mlkem768 failed");
    }
    inboundStructured.vlessEncryption = String(msg.obj?.seed || "");
    mlkemClientPreview.value = String(msg.obj?.client || "");
    inboundStructured.vlessSelectedAuth = "ML-KEM-768, Post-Quantum";
    actionSuccess.value = t("pages.inbounds.ui.mlkemGenerated", "ML-KEM-768 keys generated");
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Generate mlkem768 failed";
  }
}

async function generateMldsa65(): Promise<void> {
  actionError.value = "";
  try {
    const res = await fetch(`${basePath}panel/api/server/getNewmldsa65`, {
      method: "GET",
      credentials: "same-origin"
    });
    const msg = (await res.json()) as ApiResponse<{ seed?: string; verify?: string }>;
    if (!msg.success) {
      throw new Error(msg.msg || "Generate mldsa65 failed");
    }
    inboundStructured.decryption = String(msg.obj?.seed || inboundStructured.decryption);
    mldsaVerifyPreview.value = String(msg.obj?.verify || "");
    actionSuccess.value = t("pages.inbounds.ui.mldsaGenerated", "ML-DSA-65 keys generated");
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Generate mldsa65 failed";
  }
}

async function generateEchCert(): Promise<void> {
  actionError.value = "";
  try {
    const form = new URLSearchParams();
    form.append("sni", echSniInput.value.trim());
    const res = await fetch(`${basePath}panel/api/server/getNewEchCert`, {
      method: "POST",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded"
      },
      body: form.toString()
    });
    const msg = (await res.json()) as ApiResponse<{ echServerKeys?: string; echConfigList?: string }>;
    if (!msg.success) {
      throw new Error(msg.msg || "Generate ECH cert failed");
    }
    echServerKeysPreview.value = String(msg.obj?.echServerKeys || "");
    echConfigListPreview.value = String(msg.obj?.echConfigList || "");
    actionSuccess.value = t("pages.inbounds.ui.echCertGenerated", "ECH cert generated");
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Generate ECH cert failed";
  }
}

async function openDepletedClients(): Promise<void> {
  actionError.value = "";
  try {
    const msg = await postJson<Array<string | Record<string, unknown>>>("panel/api/inbounds/depleted");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load depleted clients");
    }
    depletedClients.value = normalizeClientStateItems(msg.obj);
    depletedModalOpen.value = true;
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to load depleted clients";
  }
}

async function openDisabledClients(): Promise<void> {
  actionError.value = "";
  try {
    const msg = await postJson<Array<string | Record<string, unknown>>>("panel/api/inbounds/disabled");
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to load disabled clients");
    }
    disabledClients.value = normalizeClientStateItems(msg.obj);
    disabledModalOpen.value = true;
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to load disabled clients";
  }
}

async function deleteClientByEmail(inboundId: number, email: string): Promise<void> {
  actionError.value = "";
  try {
    const msg = await postJson(`panel/api/inbounds/${inboundId}/delClientByEmail/${encodeURIComponent(email)}`);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to delete client");
    }
    actionSuccess.value = `Client deleted: ${email}`;
    await refresh();
    if (depletedModalOpen.value) {
      await openDepletedClients();
    }
    if (disabledModalOpen.value) {
      await openDisabledClients();
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to delete client";
  }
}

function clientInboundId(item: Record<string, unknown>): number {
  const keys = ["inboundId", "inbound_id", "inbound", "id", "InboundId", "inboundID", "inbound_id"];
  for (const k of keys) {
    const n = Number(item[k]);
    if (Number.isFinite(n) && n > 0) return n;
  }
  return 0;
}

function toInboundIndex(items: Array<Record<string, unknown>>): Record<number, number> {
  const out: Record<number, number> = {};
  for (const item of items) {
    const id = clientInboundId(item);
    if (!id) continue;
    out[id] = (out[id] || 0) + 1;
  }
  return out;
}

function normalizeEmail(text: unknown): string {
  return String(text || "").trim().toLowerCase();
}

function emailInboundIdsMap(): Record<string, number[]> {
  const map: Record<string, number[]> = {};
  for (const row of rows.value) {
    for (const client of row.clients) {
      const key = normalizeEmail(client.email);
      if (!key) continue;
      if (!map[key]) {
        map[key] = [];
      }
      if (!map[key].includes(row.id)) {
        map[key].push(row.id);
      }
    }
  }
  return map;
}

function normalizeClientStateItems(raw: unknown): Array<Record<string, unknown>> {
  if (!Array.isArray(raw)) return [];
  const map = emailInboundIdsMap();
  const out: Array<Record<string, unknown>> = [];
  for (const item of raw) {
    if (typeof item === "string") {
      const email = String(item || "").trim();
      if (!email) continue;
      const ids = map[normalizeEmail(email)] || [0];
      for (const inboundId of ids) {
        out.push({ inboundId, email });
      }
      continue;
    }
    if (item && typeof item === "object") {
      out.push(item as Record<string, unknown>);
    }
  }
  return out;
}

async function refreshClientStateIndexes(): Promise<void> {
  try {
    const dep = await postJson<Array<string | Record<string, unknown>>>("panel/api/inbounds/depleted");
    depletedIndex.value = dep.success ? toInboundIndex(normalizeClientStateItems(dep.obj)) : {};
  } catch {
    depletedIndex.value = {};
  }
  try {
    const dis = await postJson<Array<string | Record<string, unknown>>>("panel/api/inbounds/disabled");
    disabledIndex.value = dis.success ? toInboundIndex(normalizeClientStateItems(dis.obj)) : {};
  } catch {
    disabledIndex.value = {};
  }
}

function clientEmail(item: Record<string, unknown>): string {
  const keys = ["email", "clientEmail", "client_email"];
  for (const k of keys) {
    const v = String(item[k] ?? "");
    if (v) return v;
  }
  return "";
}

async function fetchClientIps(email?: string): Promise<void> {
  if (!email) return;
  actionError.value = "";
  try {
    const msg = await postJson<string>(`panel/api/inbounds/clientIps/${encodeURIComponent(email)}`);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to get client IP logs");
    }
    clientIpLogsEmail.value = email;
    clientIpLogs.value = String(msg.obj || "");
    clientIpLogsModalOpen.value = true;
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to get client IP logs";
  }
}

async function clearClientIps(email?: string): Promise<void> {
  if (!email) return;
  actionError.value = "";
  try {
    const msg = await postJson(`panel/api/inbounds/clearClientIps/${encodeURIComponent(email)}`);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to clear client IP logs");
    }
    actionSuccess.value = `IP logs cleared for ${email}`;
    if (clientIpLogsEmail.value === email) {
      clientIpLogs.value = "";
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to clear client IP logs";
  }
}

async function resetClientTrafficFromModal(email?: string): Promise<void> {
  if (!email || clientsModal.inboundId == null) return;
  if (!(await openConfirmModal(`Reset traffic for client '${email}'?`, "Reset Client Traffic"))) return;
  actionError.value = "";
  try {
    const msg = await postJson(`panel/api/inbounds/${clientsModal.inboundId}/resetClientTraffic/${encodeURIComponent(email)}`);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to reset client traffic");
    }
    actionSuccess.value = `Client traffic reset for ${email}`;
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to reset client traffic";
  }
}

async function loadClientsTraffic(): Promise<void> {
  const emails = clientsModal.clients
    .map((c) => (c.email || "").trim())
    .filter(Boolean);
  if (emails.length === 0) return;
  const next: Record<string, ClientTraffic> = {};
  for (const email of emails) {
    try {
      const res = await fetch(`${basePath}panel/api/inbounds/getClientTraffics/${encodeURIComponent(email)}`, {
        method: "GET",
        credentials: "same-origin"
      });
      const msg = (await res.json()) as ApiResponse<ClientTraffic>;
      if (msg.success && msg.obj) {
        next[email.toLowerCase()] = msg.obj;
      }
    } catch {
      // best-effort per email
    }
  }
  clientTrafficMap.value = next;
}

function clientTraffic(client: InboundClient): ClientTraffic | undefined {
  const email = (client.email || "").toLowerCase();
  if (!email) return undefined;
  return clientTrafficMap.value[email];
}

function clientUsageText(client: InboundClient): string {
  const t = clientTraffic(client);
  if (!t) return "-";
  return `${sizeFormat(t.up || 0)} / ${sizeFormat(t.down || 0)} (${sizeFormat((t.up || 0) + (t.down || 0))})`;
}

function clientTrafficUsage(client: InboundClient): number {
  const t = clientTraffic(client);
  return Number((t?.up || 0) + (t?.down || 0));
}

const sortedModalClients = computed(() => {
  const list = [...clientsModal.clients];
  if (clientSortBy.value === "email") {
    return list.sort((a, b) => String(a.email || "").localeCompare(String(b.email || "")));
  }
  if (clientSortBy.value === "traffic") {
    return list.sort((a, b) => clientTrafficUsage(b) - clientTrafficUsage(a));
  }
  if (clientSortBy.value === "expiry") {
    return list.sort((a, b) => Number(a.expiryTime || 0) - Number(b.expiryTime || 0));
  }
  return list;
});

const filteredModalClients = computed(() => {
  const keyword = clientSearch.value.trim().toLowerCase();
  if (!keyword) return sortedModalClients.value;
  return sortedModalClients.value.filter((client) => {
    const email = String(client.email || "").toLowerCase();
    const comment = String(client.comment || "").toLowerCase();
    const subId = String(client.subId || "").toLowerCase();
    return email.includes(keyword) || comment.includes(keyword) || subId.includes(keyword);
  });
});

async function getOnlineUsers(): Promise<void> {
  const res = await fetch(`${basePath}panel/api/inbounds/onlines`, {
    method: "POST",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json"
    },
    body: "{}"
  });
  const data = (await res.json()) as ApiResponse<string[] | null>;
  if (!data.success) {
    throw new Error(data.msg || "Failed to load online clients");
  }
  onlineClients.value = data.obj || [];
}

async function loadSubSettings(): Promise<void> {
  try {
    const msg = await postJson<Record<string, unknown>>("panel/setting/all");
    if (!msg.success || !msg.obj) {
      return;
    }
    setDatepickerMode(msg.obj.datepicker);
    subEnabled.value = Boolean(msg.obj.subEnable);
    remarkModel.value = String(msg.obj.remarkModel ?? "-ieo");
    const subOrigin = configuredSubOrigin(msg.obj);
    const explicit = String(msg.obj.subURI ?? "").trim();
    if (explicit) {
      const absolute = resolveSubBaseUri(explicit, subOrigin);
      subBaseUri.value = absolute.endsWith("/") ? absolute : `${absolute}/`;
      return;
    }
    const rawPath = String(msg.obj.subPath ?? "/sub/");
    const finalPath = normalizeUrlPath(rawPath, "/sub/");
    subBaseUri.value = `${subOrigin}${finalPath}`;
  } catch {
    // best-effort only
  }
}

async function loadInbounds(): Promise<void> {
  if (inboundsRequestInFlight) {
    return;
  }
  inboundsRequestInFlight = true;
  const showBlockingLoader = !refreshing.value && !firstLoadDone.value;
  if (showBlockingLoader) {
    loading.value = true;
  }
  if (!refreshing.value) {
    error.value = "";
  }
  try {
    const res = await fetch(`${basePath}panel/api/inbounds/list`, {
      method: "GET",
      credentials: "same-origin"
    });
    const data = (await res.json()) as ApiResponse<RawInbound[]>;
    if (!data.success) {
      throw new Error(data.msg || "Failed to load inbounds");
    }
    await loadSubSettings();
    await getOnlineUsers();
    rows.value = (data.obj || []).map(normalizeRow);
    await refreshClientStateIndexes();
    firstLoadDone.value = true;
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Unexpected error";
  } finally {
    inboundsRequestInFlight = false;
    if (showBlockingLoader) {
      loading.value = false;
    }
  }
}

async function refresh(): Promise<void> {
  if (inboundsRequestInFlight) {
    return;
  }
  refreshing.value = true;
  await loadInbounds();
  refreshing.value = false;
}

function openOnlineClients(): void {
  onlineModalOpen.value = true;
}

async function refreshOnlineClients(): Promise<void> {
  try {
    await getOnlineUsers();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to refresh online clients";
  }
}

function buildUpdatePayload(row: InboundRow, patch?: Partial<InboundRow>): Record<string, unknown> {
  const next = { ...row, ...(patch || {}) };
  return {
    up: next.up,
    down: next.down,
    total: next.total,
    remark: next.remark,
    enable: next.enable,
    expiryTime: next.expiryTime,
    listen: next.listen,
    port: next.port,
    protocol: next.protocol,
    settings: next.settings,
    streamSettings: next.streamSettings,
    sniffing: next.sniffing,
    tag: next.tag,
    sort: next.sort
  };
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

async function mutateInbound(row: InboundRow, patch: Partial<InboundRow>, successMsg: string): Promise<void> {
  actionBusyId.value = row.id;
  actionError.value = "";
  actionSuccess.value = "";
  try {
    const payload = buildUpdatePayload(row, patch);
    const msg = await postJson(`panel/api/inbounds/update/${row.id}`, payload);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to update inbound");
    }
    actionSuccess.value = successMsg;
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    actionBusyId.value = null;
  }
}

async function toggleEnabled(row: InboundRow): Promise<void> {
  await mutateInbound(row, { enable: !row.enable }, `Inbound #${row.id} updated`);
}

async function resetInboundTraffic(row: InboundRow): Promise<void> {
  if (!(await openConfirmModal(`Reset traffic for inbound #${row.id}?`, "Reset Inbound Traffic"))) {
    return;
  }
  await mutateInbound(row, { up: 0, down: 0 }, `Traffic reset for inbound #${row.id}`);
}

async function deleteInbound(row: InboundRow): Promise<void> {
  if (!(await openConfirmModal(`Delete inbound #${row.id}? This cannot be undone.`, "Delete Inbound"))) {
    return;
  }
  actionBusyId.value = row.id;
  actionError.value = "";
  actionSuccess.value = "";
  try {
    const msg = await postJson(`panel/api/inbounds/del/${row.id}`);
    if (!msg.success) {
      throw new Error(msg.msg || "Failed to delete inbound");
    }
    actionSuccess.value = `Inbound #${row.id} deleted`;
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    actionBusyId.value = null;
  }
}

async function openInboundDetails(row: InboundRow): Promise<void> {
  detailModalOpen.value = true;
  detailLoading.value = true;
  detailInboundJson.value = "{}";
  detailClientTraffics.value = [];
  try {
    const inboundRes = await fetch(`${basePath}panel/api/inbounds/get/${row.id}`, {
      method: "GET",
      credentials: "same-origin"
    });
    const inboundMsg = (await inboundRes.json()) as ApiResponse<Record<string, unknown>>;
    if (!inboundMsg.success) {
      throw new Error(inboundMsg.msg || "Failed to load inbound details");
    }
    detailInboundJson.value = JSON.stringify(inboundMsg.obj || {}, null, 2);

    const trafficRes = await fetch(`${basePath}panel/api/inbounds/getClientTrafficsById/${row.id}`, {
      method: "GET",
      credentials: "same-origin"
    });
    const trafficMsg = (await trafficRes.json()) as ApiResponse<InboundClientTrafficRow[]>;
    if (trafficMsg.success && Array.isArray(trafficMsg.obj)) {
      detailClientTraffics.value = trafficMsg.obj;
    }
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to load inbound details";
  } finally {
    detailLoading.value = false;
  }
}

async function runGlobalAction(confirmText: string, path: string, successMsg: string): Promise<void> {
  if (!(await openConfirmModal(confirmText, "Confirm Operation"))) {
    return;
  }
  globalBusy.value = true;
  actionError.value = "";
  actionSuccess.value = "";
  try {
    const msg = await postJson(path);
    if (!msg.success) {
      throw new Error(msg.msg || "Action failed");
    }
    actionSuccess.value = successMsg;
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    globalBusy.value = false;
  }
}

function openAddInboundModal(): void {
  inboundModal.open = true;
  inboundModal.mode = "add";
  inboundModal.id = null;
  inboundModal.remark = "";
  inboundModal.protocol = "vless";
  inboundModal.listen = "";
  inboundModal.port = "";
  inboundModal.enable = true;
  inboundModal.total = "0";
  inboundModal.expiryTime = "0";
  syncInboundExpiryInputFromModel();
  inboundModal.showAdvancedJson = false;
  realityPublicKeyPreview.value = "";
  mlkemClientPreview.value = "";
  mldsaVerifyPreview.value = "";
  echConfigListPreview.value = "";
  echServerKeysPreview.value = "";
  echSniInput.value = "";
  applyProtocolTemplate(inboundModal.protocol);
}

function openEditInboundModal(row: InboundRow): void {
  inboundModal.open = true;
  inboundModal.mode = "edit";
  inboundModal.id = row.id;
  inboundModal.remark = row.remark;
  inboundModal.protocol = row.protocol;
  inboundModal.listen = row.listen;
  inboundModal.port = String(row.port);
  inboundModal.enable = row.enable;
  inboundModal.total = String(row.total);
  inboundModal.expiryTime = String(row.expiryTime);
  syncInboundExpiryInputFromModel();
  inboundModal.settings = jsonPrettyOrDefault(row.settings, "{}");
  inboundModal.streamSettings = jsonPrettyOrDefault(row.streamSettings, "{}");
  inboundModal.sniffing = jsonPrettyOrDefault(row.sniffing, "{}");
  inboundModal.showAdvancedJson = false;
  realityPublicKeyPreview.value = "";
  mlkemClientPreview.value = "";
  mldsaVerifyPreview.value = "";
  echConfigListPreview.value = "";
  echServerKeysPreview.value = "";
  echSniInput.value = "";
  hydrateStructuredFromJson();
}

function closeInboundModal(): void {
  inboundModal.open = false;
  realityPublicKeyPreview.value = "";
  mlkemClientPreview.value = "";
  mldsaVerifyPreview.value = "";
  echConfigListPreview.value = "";
  echServerKeysPreview.value = "";
  echSniInput.value = "";
}

function onInboundProtocolChange(): void {
  if (inboundModal.mode === "add") {
    applyProtocolTemplate(inboundModal.protocol);
  } else {
    hydrateStructuredFromJson();
  }
  if (inboundModal.protocol === "vless" && inboundStructured.vlessEncOptions.length === 0) {
    getNewVlessEncOptions();
  }
}

async function saveInboundModal(): Promise<void> {
  actionError.value = "";
  actionSuccess.value = "";
  applyStructuredToJson();

  if (!inboundModal.remark.trim()) {
    actionError.value = t("pages.inbounds.ui.inboundRemarkRequired", "Inbound remark is required");
    return;
  }

  const port = numberFromString(inboundModal.port, 0);
  if (port <= 0 || port > 65535) {
    actionError.value = t("pages.inbounds.ui.portRangeInvalid", "Port must be between 1 and 65535");
    return;
  }

  if (!validJSON(inboundModal.settings)) {
    actionError.value = t("pages.inbounds.ui.settingsJsonInvalid", "Settings must be valid JSON");
    return;
  }
  if (!validJSON(inboundModal.streamSettings)) {
    actionError.value = t("pages.inbounds.ui.streamSettingsJsonInvalid", "Stream settings must be valid JSON");
    return;
  }
  if (!validJSON(inboundModal.sniffing)) {
    actionError.value = t("pages.inbounds.ui.sniffingJsonInvalid", "Sniffing must be valid JSON");
    return;
  }

  modalBusy.value = true;
  try {
    onInboundExpiryInputChange();
    const payload = {
      up: 0,
      down: 0,
      total: numberFromString(inboundModal.total, 0),
      remark: inboundModal.remark.trim(),
      enable: inboundModal.enable,
      expiryTime: numberFromString(inboundModal.expiryTime, 0),
      listen: inboundModal.listen.trim(),
      port,
      protocol: inboundModal.protocol,
      settings: jsonPrettyOrDefault(inboundModal.settings, "{}"),
      streamSettings: jsonPrettyOrDefault(inboundModal.streamSettings, "{}"),
      sniffing: jsonPrettyOrDefault(inboundModal.sniffing, "{}")
    };

    let msg: ApiResponse<unknown>;
    if (inboundModal.mode === "add") {
      msg = await postJson("panel/api/inbounds/add", payload);
    } else {
      msg = await postJson(`panel/api/inbounds/update/${inboundModal.id}`, payload);
    }

    if (!msg.success) {
      throw new Error(msg.msg || "Failed to save inbound");
    }

    actionSuccess.value = inboundModal.mode === "add" ? "Inbound created" : `Inbound #${inboundModal.id} updated`;
    closeInboundModal();
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    modalBusy.value = false;
  }
}

function openClientsModal(row: InboundRow): void {
  clientsModal.open = true;
  clientsModal.inboundId = row.id;
  clientsModal.inboundRemark = row.remark;
  clientsModal.protocol = row.protocol;
  clientSearch.value = "";
  clientsModal.clients = [...row.clients].map((c) => ({ ...c }));
  clientsModal.originalClients = [...row.clients].map((c) => ({ ...c }));
  clientTrafficMap.value = {};
  closeClientEditor();
  loadClientsTraffic();
}

function closeClientsModal(): void {
  clientsModal.open = false;
  closeClientActionsFloating();
  clientsModal.inboundId = null;
  clientsModal.inboundRemark = "";
  clientsModal.protocol = "";
  clientSearch.value = "";
  clientsModal.clients = [];
  clientsModal.originalClients = [];
  clientTrafficMap.value = {};
  clientIpLogs.value = "";
  clientIpLogsEmail.value = "";
  clientIpLogsModalOpen.value = false;
  clientQrModalOpen.value = false;
  clientQrValue.value = "";
  clientQrImageSrc.value = "";
  clientConfigsModalOpen.value = false;
  clientConfigsText.value = "";
  closeClientEditor();
  closeClientBulk();
}

function closeClientActionsFloating(): void {
  clientActionsFloating.open = false;
  clientActionsFloating.client = null;
}

function openClientActionsFloating(event: MouseEvent, client: InboundClient): void {
  const anchor = event.currentTarget instanceof HTMLElement ? event.currentTarget : null;
  if (!anchor) return;
  const rect = anchor.getBoundingClientRect();
  const menuWidth = 190;
  const space = 8;
  const left = Math.max(space, Math.min(rect.right - menuWidth, window.innerWidth - menuWidth - space));
  const roomBelow = window.innerHeight - rect.bottom;
  const roomAbove = rect.top;
  const place: "top" | "bottom" = roomBelow < 250 && roomAbove > roomBelow ? "top" : "bottom";
  clientActionsFloating.open = true;
  clientActionsFloating.client = client;
  clientActionsFloating.left = left;
  clientActionsFloating.top = place === "top" ? rect.top - space : rect.bottom + space;
  clientActionsFloating.place = place;
}

function onClientActionsFloatingRootClick(event: MouseEvent): void {
  const el = event.target instanceof Element ? event.target : null;
  if (!el) return;
  if (el.closest("[data-client-floating-menu='1']")) return;
  closeClientActionsFloating();
}

function onClientActionsFloatingViewportChange(): void {
  if (!clientActionsFloating.open) return;
  closeClientActionsFloating();
}

function saveClientSortPreference(): void {
  try {
    localStorage.setItem("clientSortBy", clientSortBy.value);
  } catch {
    // ignore
  }
}

function loadClientSortPreference(): void {
  try {
    const stored = localStorage.getItem("clientSortBy");
    if (stored === "default" || stored === "email" || stored === "traffic" || stored === "expiry") {
      clientSortBy.value = stored;
    }
  } catch {
    // ignore
  }
}

function closeClientEditor(): void {
  clientEditor.open = false;
  clientEditor.mode = "add";
  clientEditor.index = -1;
  clientEditor.email = "";
  clientEditor.id = "";
  clientEditor.password = "";
  clientEditor.flow = "";
  clientEditor.security = "";
  clientEditor.enable = true;
  clientEditor.totalGB = "0";
  clientEditor.delayedStart = false;
  clientEditor.delayedExpireDays = "0";
  clientEditor.expiryTime = "0";
  syncClientEditorExpiryInputFromModel();
  clientEditor.limitIp = "0";
  clientEditor.subId = "";
  clientEditor.comment = "";
  clientEditor.reset = "0";
  clientEditor.tgId = "0";
  clientIpLogs.value = "";
  clientIpLogsEmail.value = "";
  clientIpLogsModalOpen.value = false;
}

function openClientBulk(): void {
  closeClientEditor();
  clientBulk.open = true;
  clientBulk.method = 0;
  clientBulk.quantity = "1";
  clientBulk.firstNum = "1";
  clientBulk.lastNum = "1";
  clientBulk.prefix = "";
  clientBulk.postfix = "";
  clientBulk.security = "auto";
  clientBulk.flow = "";
  clientBulk.subId = "";
  clientBulk.tgId = "0";
  clientBulk.limitIp = "0";
  clientBulk.totalGB = "0";
  clientBulk.delayedStart = false;
  clientBulk.delayedExpireDays = "0";
  clientBulk.expiryTime = "0";
  syncClientBulkExpiryInputFromModel();
  clientBulk.reset = "0";
}

function closeClientBulk(): void {
  clientBulk.open = false;
}

function generateBulkEmail(index: number): string {
  const prefix = clientBulk.prefix.trim();
  const postfix = clientBulk.postfix.trim();
  const rand = Math.random().toString(36).slice(2, 10);
  switch (clientBulk.method) {
    case 1:
      return `${prefix}${rand}`;
    case 2:
      return `${prefix}${index}`;
    case 3:
      return `${prefix}${index}${postfix}`;
    case 4:
      return `${prefix}${index}${postfix}`;
    case 0:
    default:
      return rand;
  }
}

function applyBulkClients(): void {
  actionError.value = "";
  if (!clientBulk.delayedStart) {
    onClientBulkExpiryInputChange();
  }
  const protocol = clientsModal.protocol;
  const list: InboundClient[] = [];
  const method = Number(clientBulk.method || 0);

  let start = 0;
  let endExclusive = Math.max(1, numberFromString(clientBulk.quantity, 1));
  if (method > 1) {
    start = Math.max(1, numberFromString(clientBulk.firstNum, 1));
    endExclusive = Math.max(start, numberFromString(clientBulk.lastNum, start)) + 1;
  }

  let expiryTime = numberFromString(clientBulk.expiryTime, 0);
  if (clientBulk.delayedStart) {
    const days = Math.max(0, numberFromString(clientBulk.delayedExpireDays, 0));
    expiryTime = -1 * days * 86400000;
  }

  for (let i = start; i < endExclusive; i += 1) {
    const email = generateBulkEmail(i);
    const next: InboundClient = {
      email,
      id: protocol === "trojan" || protocol === "shadowsocks" ? "" : generateId(),
      password: protocol === "trojan" || protocol === "shadowsocks" ? generateId() : "",
      flow: protocol === "vless" ? clientBulk.flow.trim() : "",
      security: protocol === "vmess" ? (clientBulk.security.trim() || "auto") : "",
      enable: true,
      totalGB: numberFromString(clientBulk.totalGB, 0),
      expiryTime,
      limitIp: numberFromString(clientBulk.limitIp, 0),
      subId: clientBulk.subId.trim(),
      comment: "",
      reset: numberFromString(clientBulk.reset, 0),
      tgId: numberFromString(clientBulk.tgId, 0)
    };
    list.push(next);
  }

  if (list.length === 0) {
    actionError.value = t("pages.inbounds.ui.noClientsGenerated", "No clients generated");
    return;
  }

  const existing = new Set(clientsModal.clients.map((c) => (c.email || "").toLowerCase()));
  let appended = 0;
  for (const c of list) {
    const key = (c.email || "").toLowerCase();
    if (!key || existing.has(key)) {
      continue;
    }
    existing.add(key);
    clientsModal.clients.push(c);
    appended += 1;
  }

  if (appended === 0) {
    actionError.value = t("pages.inbounds.ui.generatedClientsDuplicate", "All generated clients were duplicates");
    return;
  }

  actionSuccess.value = `${appended} clients added to editor list`;
  closeClientBulk();
}

function openAddClientEditor(): void {
  closeClientBulk();
  closeClientEditor();
  clientEditor.open = true;
  clientEditor.mode = "add";
  if (clientsModal.protocol === "trojan") {
    clientEditor.password = generateId();
  } else if (clientsModal.protocol !== "shadowsocks") {
    clientEditor.id = generateId();
    if (clientsModal.protocol === "vmess") {
      clientEditor.security = "auto";
    }
  } else {
    clientEditor.password = generateId();
  }
  syncClientEditorExpiryInputFromModel();
}

function openEditClientEditor(client: InboundClient, index: number): void {
  clientEditor.open = true;
  clientEditor.mode = "edit";
  clientEditor.index = index;
  clientEditor.email = client.email || "";
  clientEditor.id = client.id || "";
  clientEditor.password = client.password || "";
  clientEditor.flow = client.flow || "";
  clientEditor.security = client.security || "";
  clientEditor.enable = client.enable ?? true;
  clientEditor.totalGB = String(client.totalGB || 0);
  const rawExpiry = Number(client.expiryTime || 0);
  if (rawExpiry < 0) {
    clientEditor.delayedStart = true;
    clientEditor.delayedExpireDays = String(Math.floor(Math.abs(rawExpiry) / 86400000));
    clientEditor.expiryTime = "0";
  } else {
    clientEditor.delayedStart = false;
    clientEditor.delayedExpireDays = "0";
    clientEditor.expiryTime = String(rawExpiry);
  }
  syncClientEditorExpiryInputFromModel();
  clientEditor.limitIp = String(client.limitIp || 0);
  clientEditor.subId = client.subId || "";
  clientEditor.comment = client.comment || "";
  clientEditor.reset = String(client.reset || 0);
  clientEditor.tgId = String(client.tgId || 0);
}

function upsertClientFromEditor(): void {
  actionError.value = "";
  if (!clientEditor.delayedStart) {
    onClientEditorExpiryInputChange();
  }

  if (!clientEditor.email.trim()) {
    actionError.value = t("pages.inbounds.ui.clientEmailRequired", "Client email is required");
    return;
  }

  const protocol = clientsModal.protocol;
  if (protocol === "trojan" && !clientEditor.password.trim()) {
    actionError.value = t("pages.inbounds.ui.trojanPasswordRequired", "Trojan client password is required");
    return;
  }
  if (protocol !== "trojan" && protocol !== "shadowsocks" && !clientEditor.id.trim()) {
    actionError.value = t("pages.inbounds.ui.clientIdRequired", "Client ID is required for this protocol");
    return;
  }
  if (protocol === "vmess" && !clientEditor.security.trim()) {
    clientEditor.security = "auto";
  }
  if (protocol === "vless" && !clientEditor.flow.trim()) {
    clientEditor.flow = "";
  }

  let expiryTime = numberFromString(clientEditor.expiryTime, 0);
  if (clientEditor.delayedStart) {
    const days = Math.max(0, numberFromString(clientEditor.delayedExpireDays, 0));
    expiryTime = -1 * days * 86400000;
  }

  const next: InboundClient = {
    email: clientEditor.email.trim(),
    id: clientEditor.id.trim(),
    password: clientEditor.password.trim(),
    flow: clientEditor.flow.trim(),
    security: clientEditor.security.trim(),
    enable: clientEditor.enable,
    totalGB: numberFromString(clientEditor.totalGB, 0),
    expiryTime,
    limitIp: numberFromString(clientEditor.limitIp, 0),
    subId: clientEditor.subId.trim(),
    comment: clientEditor.comment.trim(),
    reset: numberFromString(clientEditor.reset, 0),
    tgId: numberFromString(clientEditor.tgId, 0)
  };

  const emailLower = next.email?.toLowerCase() || "";
  const duplicate = clientsModal.clients.findIndex((c, idx) => {
    if (clientEditor.mode === "edit" && idx === clientEditor.index) {
      return false;
    }
    return (c.email || "").toLowerCase() === emailLower;
  });
  if (duplicate >= 0) {
    actionError.value = t("pages.inbounds.ui.clientEmailUnique", "Client email must be unique in this inbound");
    return;
  }

  if (clientEditor.mode === "add") {
    clientsModal.clients.push(next);
  } else if (clientEditor.index >= 0 && clientEditor.index < clientsModal.clients.length) {
    clientsModal.clients.splice(clientEditor.index, 1, next);
  }

  closeClientEditor();
}

function removeClient(index: number): void {
  clientsModal.clients.splice(index, 1);
}

function findClientIndex(client: InboundClient): number {
  const keyEmail = (client.email || "").toLowerCase();
  const keyId = String(client.id || "");
  const keyPassword = String(client.password || "");
  return clientsModal.clients.findIndex((c) => {
    if (keyEmail && (c.email || "").toLowerCase() === keyEmail) return true;
    if (keyId && (c.id || "") === keyId) return true;
    if (keyPassword && (c.password || "") === keyPassword) return true;
    return false;
  });
}

function openEditClientEditorFromRow(client: InboundClient): void {
  const idx = findClientIndex(client);
  if (idx < 0) {
    actionError.value = t("pages.inbounds.ui.clientNotFound", "Client not found");
    return;
  }
  openEditClientEditor(clientsModal.clients[idx], idx);
}

function removeClientFromRow(client: InboundClient): void {
  const idx = findClientIndex(client);
  if (idx < 0) {
    actionError.value = t("pages.inbounds.ui.clientNotFound", "Client not found");
    return;
  }
  removeClient(idx);
}

function toggleClientEnabledFromRow(client: InboundClient): void {
  const idx = findClientIndex(client);
  if (idx < 0) {
    actionError.value = t("pages.inbounds.ui.clientNotFound", "Client not found");
    return;
  }
  const current = clientsModal.clients[idx];
  clientsModal.clients[idx] = {
    ...current,
    enable: !(current.enable ?? true)
  };
}

async function saveClientsModal(): Promise<void> {
  if (clientsModal.inboundId == null) {
    return;
  }

  modalBusy.value = true;
  actionError.value = "";
  actionSuccess.value = "";
  try {
    const inboundId = clientsModal.inboundId;
    const protocol = clientsModal.protocol;
    const supportedClientApi = ["vless", "vmess", "trojan", "shadowsocks"].includes(protocol);

    if (!supportedClientApi) {
      const row = rows.value.find((r) => r.id === inboundId);
      if (!row) {
        throw new Error("Inbound not found in current list");
      }
      const settingsObj = parseSettingsObject(row.settings);
      settingsObj.clients = clientsModal.clients;
      const payload = buildUpdatePayload(row, { settings: JSON.stringify(settingsObj, null, 2) });
      const msg = await postJson(`panel/api/inbounds/update/${row.id}`, payload);
      if (!msg.success) {
        throw new Error(msg.msg || "Failed to update clients");
      }
      actionSuccess.value = `Clients updated for inbound #${row.id}`;
      closeClientsModal();
      await refresh();
      return;
    }

    const oldMap = new Map<string, InboundClient>();
    const newMap = new Map<string, InboundClient>();

    for (const c of clientsModal.originalClients) {
      const key = clientIdentity(c, protocol);
      if (key) {
        oldMap.set(key, c);
      }
    }
    for (const c of clientsModal.clients) {
      const key = clientIdentity(c, protocol);
      if (!key) {
        throw new Error("Client identity is missing");
      }
      if (newMap.has(key)) {
        throw new Error(`Duplicate client identity: ${key}`);
      }
      newMap.set(key, c);
    }

    const toDelete = [...oldMap.keys()].filter((k) => !newMap.has(k));
    const toAdd = [...newMap.entries()].filter(([k]) => !oldMap.has(k));
    const toUpdate = [...newMap.entries()].filter(([k, client]) => {
      const old = oldMap.get(k);
      return old ? !clientEqual(old, client) : false;
    });

    for (const key of toDelete) {
      const msg = await postJson(`panel/api/inbounds/${inboundId}/delClient/${encodeURIComponent(key)}`);
      if (!msg.success) {
        throw new Error(msg.msg || `Failed to delete client: ${key}`);
      }
    }

    for (const [, client] of toAdd) {
      const msg = await postForm("panel/api/inbounds/addClient", {
        id: String(inboundId),
        settings: JSON.stringify({ clients: [normalizeClient(client)] })
      });
      if (!msg.success) {
        throw new Error(msg.msg || `Failed to add client: ${client.email || clientIdentity(client, protocol)}`);
      }
    }

    for (const [key, client] of toUpdate) {
      const msg = await postForm(`panel/api/inbounds/updateClient/${encodeURIComponent(key)}`, {
        id: String(inboundId),
        settings: JSON.stringify({ clients: [normalizeClient(client)] })
      });
      if (!msg.success) {
        throw new Error(msg.msg || `Failed to update client: ${client.email || key}`);
      }
    }

    if (toDelete.length === 0 && toAdd.length === 0 && toUpdate.length === 0) {
      actionSuccess.value = t("pages.inbounds.ui.noClientChanges", "No client changes to apply");
    } else {
      actionSuccess.value = `Clients synced for inbound #${inboundId}`;
    }
    closeClientsModal();
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    modalBusy.value = false;
  }
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

function rowAsImportJSON(row: InboundRow): string {
  return JSON.stringify(
    {
      up: row.up,
      down: row.down,
      total: row.total,
      remark: row.remark,
      enable: row.enable,
      expiryTime: row.expiryTime,
      listen: row.listen,
      port: row.port,
      protocol: row.protocol,
      settings: row.settings,
      streamSettings: row.streamSettings,
      sniffing: row.sniffing
    },
    null,
    2
  );
}

function downloadText(filename: string, content: string, mime = "application/json;charset=utf-8"): void {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function exportInbound(row: InboundRow): void {
  downloadText(`inbound-${row.id}.json`, rowAsImportJSON(row));
}

function exportAllInbounds(): void {
  const payload = rows.value.map((row) => ({
    up: row.up,
    down: row.down,
    total: row.total,
    remark: row.remark,
    enable: row.enable,
    expiryTime: row.expiryTime,
    listen: row.listen,
    port: row.port,
    protocol: row.protocol,
    settings: row.settings,
    streamSettings: row.streamSettings,
    sniffing: row.sniffing
  }));
  downloadText("inbounds-all.json", JSON.stringify(payload, null, 2));
}

function collectInboundSubLinks(row: InboundRow): string[] {
  if (!subEnabled.value || !subBaseUri.value) return [];
  const links: string[] = [];
  const seen = new Set<string>();
  for (const client of row.clients) {
    const link = clientSubLink(client);
    if (!link) continue;
    if (!seen.has(link)) {
      seen.add(link);
      links.push(link);
    }
  }
  return links;
}

function exportInboundSubs(row: InboundRow): void {
  const links = collectInboundSubLinks(row);
  if (links.length === 0) {
    actionError.value = `No subscription links found for inbound #${row.id}`;
    return;
  }
  downloadText(`subs-inbound-${row.id}.txt`, links.join("\n"), "text/plain;charset=utf-8");
}

function exportAllSubs(): void {
  const links: string[] = [];
  const seen = new Set<string>();
  for (const row of rows.value) {
    for (const link of collectInboundSubLinks(row)) {
      if (!seen.has(link)) {
        seen.add(link);
        links.push(link);
      }
    }
  }
  if (links.length === 0) {
    actionError.value = t("pages.inbounds.ui.noSubscriptionLinks", "No subscription links found");
    return;
  }
  downloadText("subs-all.txt", links.join("\n"), "text/plain;charset=utf-8");
}

function clientSubLink(client: InboundClient): string {
  if (!subEnabled.value || !subBaseUri.value) return "";
  const subId = String(client.subId || "").trim();
  if (!subId) return "";
  const base = subBaseUri.value;
  if (base.endsWith("/")) return `${base}${subId}`;
  return `${base}/${subId}`;
}

function resolveSubBaseUri(raw: string, subOrigin: string): string {
  const value = String(raw || "").trim();
  if (!value) return "";
  if (/^https?:\/\//i.test(value)) {
    return value;
  }
  return `${subOrigin}${normalizeUrlPath(value, "/sub/")}`;
}

function normalizeUrlPath(pathValue: string, fallback: string): string {
  const base = String(pathValue || "").trim() || fallback;
  const withLeadingSlash = base.startsWith("/") ? base : `/${base}`;
  return withLeadingSlash.endsWith("/") ? withLeadingSlash : `${withLeadingSlash}/`;
}

function configuredSubOrigin(settings: Record<string, unknown>): string {
  const cert = String(settings.subCertFile ?? "").trim();
  const key = String(settings.subKeyFile ?? "").trim();
  const tls = cert.length > 0 && key.length > 0;
  const scheme = tls ? "https" : "http";
  const domain = String(settings.subDomain ?? "").trim() || window.location.hostname;
  const port = Number(settings.subPort ?? 0);
  const isDefaultPort = (tls && port === 443) || (!tls && port === 80);
  if (Number.isFinite(port) && port > 0 && !isDefaultPort) {
    return `${scheme}://${domain}:${port}`;
  }
  return `${scheme}://${domain}`;
}

type QRiousInstance = {
  toDataURL: (mime?: string) => string;
};

type QRiousConstructor = new (opts: {
  value: string;
  size: number;
  level?: string;
  background?: string;
  foreground?: string;
}) => QRiousInstance;

function ensureQRiousLoaded(): Promise<void> {
  const win = window as Window & { QRious?: QRiousConstructor };
  if (typeof win.QRious === "function") {
    return Promise.resolve();
  }
  if (qriousLoader) {
    return qriousLoader;
  }
  qriousLoader = new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.src = `${basePath}assets/qrcode/qrious2.min.js`;
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("Failed to load local QR generator"));
    document.head.appendChild(script);
  });
  return qriousLoader;
}

async function generateQrDataUrl(payload: string): Promise<string> {
  await ensureQRiousLoaded();
  const win = window as Window & { QRious?: QRiousConstructor };
  const QRious = win.QRious;
  if (typeof QRious !== "function") {
    throw new Error("Local QR generator unavailable");
  }
  const qr = new QRious({
    value: payload,
    size: 320,
    level: "M",
    background: "white",
    foreground: "black"
  });
  return qr.toDataURL("image/png");
}

function safeParseJSONText(input: string): unknown {
  try {
    return JSON.parse(input);
  } catch {
    return input;
  }
}

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

function inboundAddress(row: InboundRow | null): string {
  if (!row) return window.location.hostname;
  const listen = String(row.listen || "").trim();
  if (!listen || listen === "0.0.0.0" || listen === "::" || listen === "::0") {
    return window.location.hostname;
  }
  return listen;
}

function encodeBase64Utf8(input: string): string {
  try {
    return btoa(unescape(encodeURIComponent(input)));
  } catch {
    return btoa(input);
  }
}

function encodeQuery(params: Record<string, string | undefined>): string {
  const usp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v == null || v === "") continue;
    usp.set(k, v);
  }
  return usp.toString();
}

function externalProxyTarget(streamSettings: Record<string, unknown>): { host: string; port: string } | null {
  const sockopt = streamSettings.sockopt && typeof streamSettings.sockopt === "object"
    ? (streamSettings.sockopt as Record<string, unknown>)
    : null;
  if (!sockopt) return null;
  const raw = String(sockopt.dialerProxy ?? "").trim();
  if (!raw || raw === "same" || raw.startsWith("same#")) return null;
  const addrPort = raw.split("#")[0] || "";
  const idx = addrPort.lastIndexOf(":");
  if (idx <= 0) return null;
  const host = addrPort.slice(0, idx).trim();
  const port = addrPort.slice(idx + 1).trim();
  if (!host || !port) return null;
  return { host, port };
}

function firstString(value: unknown): string {
  if (typeof value === "string") return value.trim();
  if (Array.isArray(value)) {
    for (const v of value) {
      if (typeof v === "string" && v.trim()) return v.trim();
    }
  }
  return "";
}

function headerHost(headers: Record<string, unknown>): string {
  for (const [k, v] of Object.entries(headers)) {
    if (k.toLowerCase() === "host") {
      return firstString(v);
    }
  }
  return "";
}

function clientRemark(row: InboundRow, client: InboundClient): string {
  const model = String(remarkModel.value || "").trim() || "-ieo";
  const sep = model.length > 0 ? model[0] : "-";
  const order = model.length > 1 ? model.slice(1) : "ieo";
  const values: Record<string, string> = {
    i: String(row.remark || "").trim(),
    e: String(client.email || "").trim(),
    o: ""
  };
  const parts: string[] = [];
  for (const ch of order) {
    const val = values[ch] || "";
    if (val) parts.push(val);
  }
  return parts.join(sep) || String(client.email || row.remark || "client");
}

function buildClientUri(row: InboundRow | null, client: InboundClient): string {
  if (!row) return "";
  const protocol = String(row.protocol || "").toLowerCase();
  const remark = clientRemark(row, client);
  const tag = encodeURIComponent(remark);
  const inboundSettings = safeParseJSONText(row.settings);
  const st = inboundSettings && typeof inboundSettings === "object" ? (inboundSettings as Record<string, unknown>) : {};
  const stream = safeParseJSONText(row.streamSettings);
  const s = stream && typeof stream === "object" ? (stream as Record<string, unknown>) : {};
  const external = externalProxyTarget(s);
  const host = external?.host || inboundAddress(row);
  const port = external?.port || String(row.port || 0);
  const network = String(s.network ?? "tcp");
  const security = String(s.security ?? "none");
  const tlsSettings = (s.tlsSettings && typeof s.tlsSettings === "object" ? (s.tlsSettings as Record<string, unknown>) : {}) as Record<string, unknown>;
  const realitySettings = (s.realitySettings && typeof s.realitySettings === "object"
    ? (s.realitySettings as Record<string, unknown>)
    : {}) as Record<string, unknown>;
  const wsSettings = (s.wsSettings && typeof s.wsSettings === "object" ? (s.wsSettings as Record<string, unknown>) : {}) as Record<string, unknown>;
  const grpcSettings = (s.grpcSettings && typeof s.grpcSettings === "object"
    ? (s.grpcSettings as Record<string, unknown>)
    : {}) as Record<string, unknown>;
  const tcpSettings = (s.tcpSettings && typeof s.tcpSettings === "object" ? (s.tcpSettings as Record<string, unknown>) : {}) as Record<string, unknown>;

  const serverName = String(tlsSettings.serverName ?? realitySettings.serverName ?? "");
  const fingerprint = String(realitySettings.fingerprint ?? "");
  const pbk = String(realitySettings.publicKey ?? "");
  const sid = String(realitySettings.shortId ?? "");
  const spx = String(realitySettings.spiderX ?? "");
  const tcpHeaderObj = tcpSettings.header && typeof tcpSettings.header === "object"
    ? (tcpSettings.header as Record<string, unknown>)
    : {};
  const tcpHeader = String(tcpHeaderObj.type ?? "");
  const tcpRequest = tcpHeaderObj.request && typeof tcpHeaderObj.request === "object"
    ? (tcpHeaderObj.request as Record<string, unknown>)
    : {};
  const tcpRequestHeaders = tcpRequest.headers && typeof tcpRequest.headers === "object"
    ? (tcpRequest.headers as Record<string, unknown>)
    : {};
  const tcpRequestPaths = Array.isArray(tcpRequest.path) ? (tcpRequest.path as Array<unknown>) : [];

  const wsHeaders = wsSettings.headers && typeof wsSettings.headers === "object" ? (wsSettings.headers as Record<string, unknown>) : {};
  const wsPath = String(wsSettings.path ?? "");
  const wsHost = String(wsSettings.host ?? "").trim() || headerHost(wsHeaders);
  const grpcServiceName = String(grpcSettings.serviceName ?? "");
  const grpcAuthority = String(grpcSettings.authority ?? "");
  const grpcMode = grpcSettings.multiMode === true ? "multi" : "";
  const kcpSettings = s.kcpSettings && typeof s.kcpSettings === "object" ? (s.kcpSettings as Record<string, unknown>) : {};
  const kcpHeader = kcpSettings.header && typeof kcpSettings.header === "object" ? (kcpSettings.header as Record<string, unknown>) : {};
  const httpUpgradeSettings = s.httpupgradeSettings && typeof s.httpupgradeSettings === "object" ? (s.httpupgradeSettings as Record<string, unknown>) : {};
  const httpUpgradeHeaders = httpUpgradeSettings.headers && typeof httpUpgradeSettings.headers === "object"
    ? (httpUpgradeSettings.headers as Record<string, unknown>)
    : {};
  const xhttpSettings = s.xhttpSettings && typeof s.xhttpSettings === "object" ? (s.xhttpSettings as Record<string, unknown>) : {};
  const xhttpHeaders = xhttpSettings.headers && typeof xhttpSettings.headers === "object" ? (xhttpSettings.headers as Record<string, unknown>) : {};

  const tcpHost = headerHost(tcpRequestHeaders);
  const tcpPath = firstString(tcpRequestPaths);
  const huHost = String(httpUpgradeSettings.host ?? "").trim() || headerHost(httpUpgradeHeaders);
  const huPath = String(httpUpgradeSettings.path ?? "").trim();
  const xhttpHost = String(xhttpSettings.host ?? "").trim() || headerHost(xhttpHeaders);
  const xhttpPath = String(xhttpSettings.path ?? "").trim();

  if (protocol === "vless") {
    const id = String(client.id || "").trim();
    if (!id) return "";
    const vlessEncryption = String(st.encryption ?? "").trim();
    const query = encodeQuery({
      encryption: vlessEncryption || undefined,
      flow: String(client.flow || ""),
      type: network,
      security,
      sni: serverName,
      fp: fingerprint,
      pbk,
      sid,
      spx,
      host: network === "ws"
        ? wsHost
        : network === "tcp" && tcpHeader === "http"
          ? tcpHost
          : network === "httpupgrade"
            ? huHost
            : network === "xhttp"
              ? xhttpHost
              : "",
      path: network === "ws"
        ? wsPath
        : network === "tcp" && tcpHeader === "http"
          ? tcpPath
          : network === "httpupgrade"
            ? huPath
            : network === "xhttp"
              ? xhttpPath
              : "",
      serviceName: grpcServiceName,
      authority: grpcAuthority,
      mode: network === "grpc" ? grpcMode : network === "xhttp" ? String(xhttpSettings.mode ?? "") : "",
      headerType: tcpHeader
        || (network === "kcp" ? String(kcpHeader.type ?? "") : ""),
      seed: network === "kcp" ? String(kcpSettings.seed ?? "") : ""
    });
    return `vless://${id}@${host}:${port}${query ? `?${query}` : ""}#${tag}`;
  }

  if (protocol === "vmess") {
    const id = String(client.id || "").trim();
    if (!id) return "";
    const vmessObject: Record<string, string> = {
      v: "2",
      ps: remark,
      add: host,
      port,
      id,
      aid: "0",
      scy: String(client.security || "auto"),
      net: network,
      type: tcpHeader || "none",
      host: network === "ws"
        ? wsHost
        : network === "tcp" && tcpHeader === "http"
          ? tcpHost
          : "",
      path: network === "ws"
        ? wsPath
        : network === "tcp" && tcpHeader === "http"
          ? tcpPath
          : "",
      tls: security === "none" ? "" : "tls",
      sni: serverName,
      alpn: "",
      fp: fingerprint
    };
    return `vmess://${encodeBase64Utf8(JSON.stringify(vmessObject))}`;
  }

  if (protocol === "trojan") {
    const password = String(client.password || "").trim();
    if (!password) return "";
    const query = encodeQuery({
      type: network,
      security,
      sni: serverName,
      fp: fingerprint,
      host: network === "ws"
        ? wsHost
        : network === "tcp" && tcpHeader === "http"
          ? tcpHost
          : network === "httpupgrade"
            ? huHost
            : network === "xhttp"
              ? xhttpHost
              : "",
      path: network === "ws"
        ? wsPath
        : network === "tcp" && tcpHeader === "http"
          ? tcpPath
          : network === "httpupgrade"
            ? huPath
            : network === "xhttp"
              ? xhttpPath
              : "",
      serviceName: grpcServiceName
        || undefined,
      authority: grpcAuthority || undefined,
      mode: network === "grpc" ? grpcMode : network === "xhttp" ? String(xhttpSettings.mode ?? "") : "",
      headerType: tcpHeader
        || (network === "kcp" ? String(kcpHeader.type ?? "") : ""),
      seed: network === "kcp" ? String(kcpSettings.seed ?? "") : ""
    });
    return `trojan://${encodeURIComponent(password)}@${host}:${port}${query ? `?${query}` : ""}#${tag}`;
  }

  if (protocol === "shadowsocks") {
    const settings = safeParseJSONText(row.settings);
    const st = settings && typeof settings === "object" ? (settings as Record<string, unknown>) : {};
    const method = String(st.method ?? "aes-128-gcm");
    const password = String(client.password || st.password || "").trim();
    if (!password) return "";
    const userInfo = encodeBase64Utf8(`${method}:${password}`);
    return `ss://${userInfo}@${host}:${port}#${tag}`;
  }

  return "";
}

function openClientInfo(client: InboundClient): void {
  clientInfoData.value = { ...client };
  clientInfoModalOpen.value = true;
}

async function openClientQr(client: InboundClient): Promise<void> {
  const subLink = clientSubLink(client).trim();
  const payload = subLink || JSON.stringify(normalizeClient(client));
  if (!payload) {
    actionError.value = t("pages.inbounds.ui.noQrData", "No data available for QR");
    return;
  }
  actionError.value = "";
  clientQrValue.value = payload;
  clientQrImageSrc.value = "";
  try {
    clientQrImageSrc.value = await generateQrDataUrl(payload);
    clientQrModalOpen.value = true;
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Failed to generate QR code";
  }
}

async function openClientConfigs(client: InboundClient): Promise<void> {
  const row = rows.value.find((r) => r.id === clientsModal.inboundId) || null;
  const uri = buildClientUri(row, client);
  const sub = clientSubLink(client);
  const lines: string[] = [];
  if (uri) lines.push(uri);
  if (sub) lines.push(sub);
  if (!lines.length) {
    actionError.value = t("pages.inbounds.ui.noClientUri", "No URI available for this client");
    return;
  }
  clientConfigsText.value = lines.join("\n");
  clientConfigsQrValue.value = lines[0] || "";
  clientConfigsQrImageSrc.value = "";
  if (clientConfigsQrValue.value) {
    try {
      clientConfigsQrImageSrc.value = await generateQrDataUrl(clientConfigsQrValue.value);
    } catch {
      clientConfigsQrImageSrc.value = "";
    }
  }
  clientConfigsModalOpen.value = true;
}

async function onGlobalActionMenu(action: string): Promise<void> {
  if (!action) return;
  if (action === "addInbound") {
    openAddInboundModal();
    return;
  }
  if (action === "importInbound") {
    openImportModal();
    return;
  }
  if (action === "exportAll") {
    exportAllInbounds();
    return;
  }
  if (action === "exportAllSubs") {
    exportAllSubs();
    return;
  }
  if (action === "resetAllTraffic") {
    await runGlobalAction("Reset all inbound traffic?", "panel/api/inbounds/resetAllTraffics", "All inbound traffics reset");
    return;
  }
  if (action === "resetAllClients") {
    await runGlobalAction("Reset all client traffics?", "panel/api/inbounds/resetAllClientTraffics/-1", "All client traffics reset");
    return;
  }
  if (action === "deleteAllDepleted") {
    await runGlobalAction(
      "Delete depleted clients for all inbounds?",
      "panel/api/inbounds/delDepletedClients/-1",
      "Depleted clients deleted"
    );
    return;
  }
  if (action === "viewDepleted") {
    await openDepletedClients();
    return;
  }
  if (action === "viewOnline") {
    openOnlineClients();
    return;
  }
  if (action === "viewDisabled") {
    await openDisabledClients();
    return;
  }
  if (action === "refreshAll") {
    await refresh();
  }
}

async function onInboundActionMenu(action: string, row: InboundRow): Promise<void> {
  if (!action) return;
  if (action === "details") {
    await openInboundDetails(row);
    return;
  }
  if (action === "edit") {
    openEditInboundModal(row);
    return;
  }
  if (action === "clients") {
    openClientsModal(row);
    return;
  }
  if (action === "up") {
    await moveInbound(row, "up");
    return;
  }
  if (action === "down") {
    await moveInbound(row, "down");
    return;
  }
  if (action === "export") {
    exportInbound(row);
    return;
  }
  if (action === "exportSubs") {
    exportInboundSubs(row);
    return;
  }
  if (action === "toggle") {
    await toggleEnabled(row);
    return;
  }
  if (action === "resetTraffic") {
    await resetInboundTraffic(row);
    return;
  }
  if (action === "resetClients") {
    await runGlobalAction(
      `Reset all client traffics for inbound #${row.id}?`,
      `panel/api/inbounds/resetAllClientTraffics/${row.id}`,
      `All client traffics reset for inbound #${row.id}`
    );
    return;
  }
  if (action === "deleteDepleted") {
    await runGlobalAction(
      `Delete depleted clients for inbound #${row.id}?`,
      `panel/api/inbounds/delDepletedClients/${row.id}`,
      `Depleted clients deleted for inbound #${row.id}`
    );
    return;
  }
  if (action === "delete") {
    await deleteInbound(row);
  }
}

async function onClientActionMenu(action: string, client: InboundClient): Promise<void> {
  if (!action) return;
  closeClientActionsFloating();
  if (action === "info") {
    openClientInfo(client);
    return;
  }
  if (action === "qr") {
    await openClientQr(client);
    return;
  }
  if (action === "configs") {
    await openClientConfigs(client);
    return;
  }
  if (action === "edit") {
    openEditClientEditorFromRow(client);
    return;
  }
  if (action === "toggleEnabled") {
    toggleClientEnabledFromRow(client);
    return;
  }
  if (action === "reset") {
    await resetClientTrafficFromModal(client.email);
    return;
  }
  if (action === "ipLogs") {
    await fetchClientIps(client.email);
    return;
  }
  if (action === "clearIps") {
    await clearClientIps(client.email);
    return;
  }
  if (action === "delete") {
    removeClientFromRow(client);
  }
}

async function copyToClipboard(text: string): Promise<void> {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    actionSuccess.value = t("pages.inbounds.ui.copiedToClipboard", "Copied to clipboard");
  } catch {
    actionError.value = t("pages.inbounds.ui.copyToClipboardFailed", "Failed to copy to clipboard");
  }
}

function openImportModal(): void {
  importModal.open = true;
  importModal.data = "";
}

function closeImportModal(): void {
  importModal.open = false;
}

async function importInboundFromJSON(): Promise<void> {
  actionError.value = "";
  actionSuccess.value = "";
  if (!importModal.data.trim()) {
    actionError.value = t("pages.inbounds.ui.importDataRequired", "Import data is required");
    return;
  }
  if (!validJSON(importModal.data)) {
    actionError.value = t("pages.inbounds.ui.importDataJsonInvalid", "Import data must be valid JSON");
    return;
  }
  importBusy.value = true;
  try {
    const msg = await postForm("panel/api/inbounds/import", { data: importModal.data.trim() });
    if (!msg.success) {
      throw new Error(msg.msg || "Import failed");
    }
    actionSuccess.value = t("pages.inbounds.ui.inboundImported", "Inbound imported");
    closeImportModal();
    await refresh();
  } catch (e) {
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    importBusy.value = false;
  }
}

async function moveInbound(row: InboundRow, direction: "up" | "down"): Promise<void> {
  if (globalBusy.value || refreshing.value) return;
  const idx = rows.value.findIndex((x) => x.id === row.id);
  if (idx < 0) return;
  const nextIndex = direction === "up" ? idx - 1 : idx + 1;
  if (nextIndex < 0 || nextIndex >= rows.value.length) return;

  const previous = [...rows.value];
  const tmp = rows.value[idx];
  rows.value[idx] = rows.value[nextIndex];
  rows.value[nextIndex] = tmp;

  globalBusy.value = true;
  actionError.value = "";
  try {
    const ids = rows.value.map((r) => r.id);
    const msg = await postJson("panel/api/inbounds/reorder", { ids });
    if (!msg.success) {
      throw new Error(msg.msg || "Reorder failed");
    }
    actionSuccess.value = t("pages.inbounds.ui.inboundOrderUpdated", "Inbound order updated");
  } catch (e) {
    rows.value = previous;
    actionError.value = e instanceof Error ? e.message : "Unexpected action error";
  } finally {
    globalBusy.value = false;
  }
}

const filteredRows = computed(() => {
  const keyword = search.value.trim().toLowerCase();
  const base = rows.value;

  if (!keyword) return base;

  return base.filter((row) => {
    if (String(row.id).includes(keyword)) return true;
    if (row.remark.toLowerCase().includes(keyword)) return true;
    if (row.protocol.toLowerCase().includes(keyword)) return true;
    if (String(row.port).includes(keyword)) return true;

    for (const client of row.clients) {
      if ((client.email || "").toLowerCase().includes(keyword)) {
        return true;
      }
    }
    return false;
  });
});

onMounted(async () => {
  void loadI18n([
    "pages.inbounds.title",
    "pages.inbounds.generalActions",
    "pages.inbounds.addInbound",
    "pages.inbounds.importInbound",
    "pages.inbounds.export",
    "pages.inbounds.resetAllTraffic",
    "pages.inbounds.resetAllClientTraffics",
    "pages.inbounds.delDepletedClients",
    "pages.inbounds.operate",
    "pages.inbounds.remark",
    "pages.inbounds.protocol",
    "pages.inbounds.port",
    "pages.inbounds.traffic",
    "pages.inbounds.expireDate",
    "pages.inbounds.details",
    "pages.inbounds.clients",
    "status",
    "enabled",
    "disabled",
    "online",
    "depleted",
    "search",
    "loading",
    "moveUp",
    "moveDown",
    "edit",
    "delete",
    "close",
    "copy",
    "info",
    "qrCode",
    "reset",
    "import",
    "upload",
    "download",
    "enable",
    "inbound",
    "clearIPs",
    "pages.inbounds.ui.importPlaceholder",
    "pages.inbounds.ui.clientTrafficRows",
    "pages.inbounds.ui.noClientTrafficRows",
    "pages.inbounds.ui.noDepletedClients",
    "pages.inbounds.ui.subscriptionLink",
    "pages.inbounds.ui.rawClientJson",
    "pages.inbounds.ui.qrContent",
    "pages.inbounds.ui.configQr",
    "pages.inbounds.ui.copyUris",
    "pages.inbounds.ui.ipLogs",
    "pages.inbounds.ui.noIpLogRecord"
  ]);
  loadClientSortPreference();
  try {
    const auto = localStorage.getItem("inboundsAutoRefresh");
    const sec = localStorage.getItem("inboundsAutoRefreshSeconds");
    if (auto === "0" || auto === "1") {
      autoRefresh.value = auto === "1";
    }
    if (sec) {
      const parsed = Number(sec);
      if (Number.isFinite(parsed) && parsed > 0) {
        autoRefreshSeconds.value = parsed;
      }
    }
  } catch {
    // ignore storage failures
  }
  document.addEventListener("click", onClientActionsFloatingRootClick);
  document.addEventListener("scroll", onClientActionsFloatingViewportChange, true);
  window.addEventListener("resize", onClientActionsFloatingViewportChange);
  await loadInbounds();
  startAutoRefresh();
});

onUnmounted(() => {
  document.removeEventListener("click", onClientActionsFloatingRootClick);
  document.removeEventListener("scroll", onClientActionsFloatingViewportChange, true);
  window.removeEventListener("resize", onClientActionsFloatingViewportChange);
  stopAutoRefresh();
  if (notifyTimer !== null) {
    window.clearTimeout(notifyTimer);
    notifyTimer = null;
  }
});
</script>

<template>
  <UiCard>
    <UiCardHeader>
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <UiCardTitle>{{ t("pages.inbounds.title", "Inbounds") }}</UiCardTitle>
          <UiCardDescription>{{ t("pages.inbounds.generalActions", "General Actions") }}</UiCardDescription>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <label class="inline-flex items-center gap-2 rounded-md border bg-background px-2 py-2 text-xs">
            <input v-model="autoRefresh" type="checkbox" @change="onAutoRefreshChange" />
            {{ t("default", "Auto") }}
          </label>
          <select v-model.number="autoRefreshSeconds" class="rounded-md border bg-background px-2 py-2 text-xs" @change="onAutoRefreshChange">
            <option :value="2">2s</option>
            <option :value="5">5s</option>
            <option :value="10">10s</option>
            <option :value="30">30s</option>
          </select>
          <details class="relative inline-block">
            <summary class="action-trigger">
              <MoreVertical class="h-4 w-4" />
              {{ t("pages.inbounds.operate", "Menu") }}
            </summary>
            <div class="menu-inline-end absolute z-20 mt-1 w-60 rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('addInbound')">{{ t("pages.inbounds.addInbound", "Add Inbound") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('importInbound')">{{ t("pages.inbounds.importInbound", "Import Inbound") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('exportAll')">{{ t("pages.inbounds.export", "Export") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted disabled:opacity-50" :disabled="globalBusy || refreshing || !subEnabled" @click="onGlobalActionMenu('exportAllSubs')">{{ t("pages.inbounds.export", "Export") }} {{ t("pages.settings.subSettings", "Subs") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('resetAllTraffic')">{{ t("pages.inbounds.resetAllTraffic", "Reset All Inbounds Traffic") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('resetAllClients')">{{ t("pages.inbounds.resetAllClientTraffics", "Reset All Clients Traffic") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('deleteAllDepleted')">{{ t("pages.inbounds.delDepletedClients", "Delete Depleted Clients") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('viewDepleted')">{{ t("depleted", "Depleted") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('viewOnline')">{{ t("online", "Online") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('viewDisabled')">{{ t("disabled", "Disabled") }}</button>
              <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-sm hover:bg-muted" :disabled="globalBusy || refreshing" @click="onGlobalActionMenu('refreshAll')">{{ t("reset", "Refresh") }}</button>
            </div>
          </details>
        </div>
      </div>
    </UiCardHeader>
    <UiCardContent>
      <div class="mb-4 flex items-center gap-2 rounded-md border bg-muted/40 px-3 py-2">
        <Search class="h-4 w-4 text-muted-foreground" />
        <input
          v-model="search"
          class="w-full border-none bg-transparent text-sm outline-none"
          :placeholder="t('search', 'Search') + '...'"
          type="text"
        />
      </div>
      <div class="mb-4 grid grid-cols-2 gap-2 md:grid-cols-6">
        <div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
          <div class="text-muted-foreground">{{ t("pages.inbounds.clients", "Clients") }}</div>
          <div class="text-base font-semibold">{{ clientSummary.totalClients }}</div>
        </div>
        <div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
          <div class="text-muted-foreground">{{ t("online", "Online") }}</div>
          <div class="text-base font-semibold">{{ clientSummary.onlineClientsCount }}</div>
        </div>
        <div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
          <div class="text-muted-foreground">{{ t("disabled", "Disabled") }}</div>
          <div class="text-base font-semibold">{{ clientSummary.disabledClientsCount }}</div>
        </div>
        <div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
          <div class="text-muted-foreground">{{ t("depleted", "Depleted") }}</div>
          <div class="text-base font-semibold">{{ clientSummary.depletedClientsCount }}</div>
        </div>
        <div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
          <div class="text-muted-foreground">{{ t("upload", "Upload") }}</div>
          <div class="text-base font-semibold">{{ sizeFormat(inboundTrafficSummary.up) }}</div>
        </div>
        <div class="rounded-md border bg-muted/20 px-3 py-2 text-sm">
          <div class="text-muted-foreground">{{ t("download", "Download") }}</div>
          <div class="text-base font-semibold">{{ sizeFormat(inboundTrafficSummary.down) }}</div>
        </div>
      </div>
      <div v-if="loading" class="py-8 text-sm text-muted-foreground">{{ t("loading", "Loading...") }}</div>
      <div v-else>
        <div v-if="error" class="mb-3 rounded-md border border-red-300 bg-red-50 p-3 text-sm text-red-700">
          {{ error }}
        </div>
        <div class="overflow-visible pb-24">
        <div class="space-y-2 md:hidden">
          <div
            v-for="row in filteredRows"
            :key="`mobile-${row.id}`"
            class="rounded-lg border border-border/80 bg-card/40 p-3"
          >
            <div class="mb-2 flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="truncate text-sm font-semibold" :title="row.remark">{{ row.remark }}</div>
                <div class="text-xs text-muted-foreground">#{{ row.id }} · {{ row.protocol.toUpperCase() }} · {{ row.port }}</div>
              </div>
              <details class="relative shrink-0">
                <summary class="action-trigger action-trigger-sm">
                  <MoreVertical class="h-4 w-4" />
                  {{ t("pages.inbounds.operate", "Menu") }}
                </summary>
                <div class="menu-inline-end absolute z-[70] mt-1 w-44 rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('details', row)">{{ t("pages.inbounds.details", "Details") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('edit', row)">{{ t("edit", "Edit") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('clients', row)">{{ t("clients", "Clients") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="refreshing || globalBusy" @click="onInboundActionMenu('up', row)">{{ t("moveUp", "Move Up") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="refreshing || globalBusy" @click="onInboundActionMenu('down', row)">{{ t("moveDown", "Move Down") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="refreshing || globalBusy" @click="onInboundActionMenu('export', row)">{{ t("pages.inbounds.export", "Export") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted disabled:opacity-50" :disabled="refreshing || globalBusy || !subEnabled" @click="onInboundActionMenu('exportSubs', row)">{{ t("pages.inbounds.export", "Export") }} {{ t("pages.settings.subSettings", "Subs") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('toggle', row)">{{ row.enable ? t("disabled", "Disable") : t("enabled", "Enable") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('resetTraffic', row)">{{ t("pages.inbounds.resetTraffic", "Reset Traffic") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('resetClients', row)">{{ t("pages.inbounds.resetInboundClientTraffics", "Reset Clients Traffic") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('deleteDepleted', row)">{{ t("pages.inbounds.delDepletedClients", "Delete Depleted Clients") }}</button>
                  <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs text-red-700 hover:bg-red-50" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('delete', row)">{{ t("delete", "Delete") }}</button>
                </div>
              </details>
            </div>
            <div class="mb-2 flex flex-wrap gap-1 text-[11px]">
              <span class="rounded bg-indigo-100 px-1.5 py-0.5 text-indigo-700">{{ row.clients.length }}</span>
              <span class="rounded bg-sky-100 px-1.5 py-0.5 text-sky-700">{{ getOnlineCount(row.clients) }}</span>
              <span class="rounded bg-amber-100 px-1.5 py-0.5 text-amber-800">{{ disabledIndex[row.id] || 0 }}</span>
              <span class="rounded bg-rose-100 px-1.5 py-0.5 text-rose-700">{{ depletedIndex[row.id] || 0 }}</span>
            </div>
            <div class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs">
              <div class="text-muted-foreground">{{ t("status", "Status") }}</div>
              <div>
                <span
                  :class="[
                    'inline-flex rounded-full px-2 py-0.5 text-[11px] font-medium',
                    row.enable ? 'bg-emerald-100 text-emerald-700' : 'bg-rose-100 text-rose-700'
                  ]"
                >
                  {{ row.enable ? t("enabled", "Enabled") : t("disabled", "Disabled") }}
                </span>
              </div>
              <div class="text-muted-foreground">{{ t("pages.inbounds.traffic", "Traffic") }}</div>
              <div class="truncate" :title="trafficText(row)">{{ trafficText(row) }}</div>
              <div class="text-muted-foreground">{{ t("pages.inbounds.expireDate", "Expiry") }}</div>
              <div class="truncate" :title="formatExpiry(row.expiryTime)">{{ formatExpiry(row.expiryTime) }}</div>
            </div>
          </div>
          <div v-if="filteredRows.length === 0" class="rounded-md border px-3 py-6 text-center text-sm text-muted-foreground">
            {{ t("none", "None") }}
          </div>
        </div>
        <div class="hidden overflow-x-auto md:block md:overflow-visible">
        <table class="tx-table-center min-w-[980px] md:min-w-full border-collapse text-xs">
          <thead>
            <tr class="border-b text-left text-muted-foreground">
              <th class="px-2 py-2 font-medium">ID</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.remark", "Remark") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.protocol", "Protocol") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.port", "Port") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("status", "Status") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.clients", "Clients") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.traffic", "Traffic") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.expireDate", "Expiry") }}</th>
              <th class="px-2 py-2 font-medium">{{ t("pages.inbounds.operate", "Menu") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in filteredRows" :key="row.id" class="border-b">
              <td class="whitespace-nowrap px-2 py-2">{{ row.id }}</td>
              <td class="max-w-[200px] truncate px-2 py-2 font-medium" :title="row.remark">{{ row.remark }}</td>
              <td class="whitespace-nowrap px-2 py-2 uppercase">{{ row.protocol }}</td>
              <td class="whitespace-nowrap px-2 py-2">{{ row.port }}</td>
              <td class="whitespace-nowrap px-2 py-2">
                <span
                  :class="[
                    'inline-flex rounded-full px-2 py-0.5 text-xs font-medium',
                    row.enable ? 'bg-emerald-100 text-emerald-700' : 'bg-rose-100 text-rose-700'
                  ]"
                >
                  {{ row.enable ? t("enabled", "Enabled") : t("disabled", "Disabled") }}
                </span>
              </td>
              <td class="px-2 py-2">
                <div class="flex flex-wrap gap-1 text-[11px]">
                  <span class="rounded bg-indigo-100 px-1.5 py-0.5 text-indigo-700">{{ row.clients.length }}</span>
                  <span class="rounded bg-sky-100 px-1.5 py-0.5 text-sky-700">{{ getOnlineCount(row.clients) }}</span>
                  <span class="rounded bg-amber-100 px-1.5 py-0.5 text-amber-800">{{ disabledIndex[row.id] || 0 }}</span>
                  <span class="rounded bg-rose-100 px-1.5 py-0.5 text-rose-700">{{ depletedIndex[row.id] || 0 }}</span>
                </div>
              </td>
              <td class="max-w-[170px] truncate px-2 py-2" :title="trafficText(row)">{{ trafficText(row) }}</td>
              <td class="max-w-[180px] truncate px-2 py-2" :title="formatExpiry(row.expiryTime)">{{ formatExpiry(row.expiryTime) }}</td>
              <td class="relative overflow-visible px-2 py-2">
                <details class="relative inline-block">
                  <summary class="action-trigger action-trigger-sm">
                    <MoreVertical class="h-4 w-4" />
                    {{ t("pages.inbounds.operate", "Menu") }}
                  </summary>
                  <div class="menu-inline-end absolute bottom-full z-[70] mb-1 w-48 overflow-y-auto max-h-[70vh] md:max-h-none md:overflow-visible rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover">
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('details', row)">{{ t("pages.inbounds.details", "Details") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('edit', row)">{{ t("edit", "Edit") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('clients', row)">{{ t("clients", "Clients") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="refreshing || globalBusy" @click="onInboundActionMenu('up', row)">{{ t("moveUp", "Move Up") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="refreshing || globalBusy" @click="onInboundActionMenu('down', row)">{{ t("moveDown", "Move Down") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="refreshing || globalBusy" @click="onInboundActionMenu('export', row)">{{ t("pages.inbounds.export", "Export") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted disabled:opacity-50" :disabled="refreshing || globalBusy || !subEnabled" @click="onInboundActionMenu('exportSubs', row)">{{ t("pages.inbounds.export", "Export") }} {{ t("pages.settings.subSettings", "Subs") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('toggle', row)">{{ row.enable ? t("disabled", "Disable") : t("enabled", "Enable") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('resetTraffic', row)">{{ t("pages.inbounds.resetTraffic", "Reset Traffic") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('resetClients', row)">{{ t("pages.inbounds.resetInboundClientTraffics", "Reset Clients Traffic") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('deleteDepleted', row)">{{ t("pages.inbounds.delDepletedClients", "Delete Depleted Clients") }}</button>
                    <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs text-red-700 hover:bg-red-50" :disabled="actionBusyId === row.id || refreshing || globalBusy" @click="onInboundActionMenu('delete', row)">{{ t("delete", "Delete") }}</button>
                  </div>
                </details>
              </td>
            </tr>
            <tr v-if="filteredRows.length === 0">
              <td class="px-2 py-6 text-center text-muted-foreground" colspan="9">{{ t("none", "None") }}</td>
            </tr>
          </tbody>
        </table>
        </div>
      </div>
      </div>
    </UiCardContent>
  </UiCard>

  <div v-if="inboundModal.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="max-h-[92vh] w-[min(96vw,1650px)] max-w-none overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
      <h2 class="mb-3 text-lg font-semibold">{{ inboundModal.mode === "add" ? t("pages.inbounds.addInbound", "Add Inbound") : `${t("pages.inbounds.modifyInbound", "Edit Inbound")} #${inboundModal.id}` }}</h2>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("remark", "Remark") }}</div>
          <input v-model="inboundModal.remark" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("protocol", "Protocol") }}</div>
          <select v-model="inboundModal.protocol" class="w-full rounded-md border bg-background px-3 py-2" @change="onInboundProtocolChange">
            <option value="vmess">vmess</option>
            <option value="vless">vless</option>
            <option value="trojan">trojan</option>
            <option value="shadowsocks">shadowsocks</option>
            <option value="http">http</option>
            <option value="mixed">mixed</option>
            <option value="wireguard">wireguard</option>
            <option value="tunnel">tunnel</option>
            <option value="tun">tun</option>
          </select>
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.address", "Listen") }}</div>
          <input v-model="inboundModal.listen" class="w-full rounded-md border bg-background px-3 py-2" :placeholder="t('pages.inbounds.ui.listenHint', '0.0.0.0 or empty')" type="text" />
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.port", "Port") }}</div>
          <input v-model="inboundModal.port" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.totalBytesUnlimited", "Total (bytes, 0 = unlimited)") }}</div>
          <input v-model="inboundModal.total" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">
            Expiry Time ({{ datepickerMode === "persian" ? "Persian Calendar" : "Gregorian Calendar" }}, empty = never)
          </div>
          <input v-model="inboundExpiryInput" class="w-full rounded-md border bg-background px-3 py-2" type="datetime-local" @input="onInboundExpiryInputChange" />
          <div class="mt-1 text-xs text-muted-foreground">
            {{ formatExpiry(numberFromString(inboundModal.expiryTime, 0)) }} | unix ms: {{ numberFromString(inboundModal.expiryTime, 0) || 0 }}
          </div>
        </label>
        <label class="text-sm md:col-span-2">
          <div class="mb-1 text-muted-foreground">{{ t("enable", "Enable") }}</div>
          <input v-model="inboundModal.enable" type="checkbox" />
        </label>
      </div>

      <div class="mt-3 rounded-lg border p-3">
        <h3 class="mb-2 text-sm font-semibold">{{ t("pages.inbounds.ui.protocolStreamSniffing", "Protocol / Stream / Sniffing") }}</h3>
        <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
          <template v-if="inboundModal.protocol === 'vless'">
            <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.authentication", "Authentication") }}</div>
              <select v-model="inboundStructured.vlessSelectedAuth" class="w-full rounded-md border bg-background px-3 py-2" @change="onVlessAuthChange">
                <option value="">none</option>
                <option v-for="opt in inboundStructured.vlessEncOptions" :key="opt.label" :value="opt.label">{{ opt.label }}</option>
              </select>
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">VLESS Decryption</div>
              <input v-model="inboundStructured.decryption" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">VLESS Encryption</div>
              <input v-model="inboundStructured.vlessEncryption" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.vlessHelpers", "VLESS Helpers") }}</div>
              <div class="flex gap-2 rounded-md border px-2 py-2">
                <UiButton size="sm" variant="outline" @click="getNewVlessEncOptions">{{ t("pages.inbounds.ui.getNewKeys", "Get New Keys") }}</UiButton>
                <UiButton size="sm" variant="outline" @click="inboundStructured.vlessSelectedAuth = ''; inboundStructured.vlessEncryption = ''; inboundStructured.decryption = 'none'">{{ t("pages.inbounds.ui.clear", "Clear") }}</UiButton>
              </div>
            </label>
            <label class="text-sm md:col-span-2" v-if="inboundStructured.streamNetwork === 'tcp' && !inboundStructured.vlessSelectedAuth">
              <div class="mb-1 flex items-center justify-between text-muted-foreground">
                <span>{{ t("pages.inbounds.ui.vlessFallbacks", "VLESS Fallbacks") }}</span>
                <UiButton size="sm" variant="outline" @click="addVlessFallback">{{ t("pages.xray.rules.add", "Add") }}</UiButton>
              </div>
              <div class="space-y-2">
                <div v-for="(fb, idx) in inboundStructured.vlessFallbacks" :key="`vless-fb-${idx}`" class="grid grid-cols-[1fr_1fr_1fr_1fr_100px_auto] gap-2">
                  <input v-model="fb.name" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="sni" type="text" />
                  <input v-model="fb.alpn" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="alpn" type="text" />
                  <input v-model="fb.path" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="path" type="text" />
                  <input v-model="fb.dest" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="dest" type="text" />
                  <input v-model.number="fb.xver" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="xver" type="number" />
                  <UiButton size="sm" variant="outline" @click="removeVlessFallback(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
              </div>
            </label>
          </template>

          <label class="text-sm md:col-span-2" v-if="inboundModal.protocol === 'trojan' && inboundStructured.streamNetwork === 'tcp'">
            <div class="mb-1 flex items-center justify-between text-muted-foreground">
              <span>{{ t("pages.inbounds.ui.trojanFallbacks", "Trojan Fallbacks") }}</span>
              <UiButton size="sm" variant="outline" @click="addTrojanFallback">{{ t("pages.xray.rules.add", "Add") }}</UiButton>
            </div>
            <div class="space-y-2">
              <div v-for="(fb, idx) in inboundStructured.trojanFallbacks" :key="`trojan-fb-${idx}`" class="grid grid-cols-[1fr_1fr_1fr_1fr_100px_auto] gap-2">
                <input v-model="fb.name" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="sni" type="text" />
                <input v-model="fb.alpn" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="alpn" type="text" />
                <input v-model="fb.path" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="path" type="text" />
                <input v-model="fb.dest" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="dest" type="text" />
                <input v-model.number="fb.xver" class="rounded-md border bg-background px-2 py-2 text-xs" placeholder="xver" type="number" />
                <UiButton size="sm" variant="outline" @click="removeTrojanFallback(idx)">{{ t("delete", "Delete") }}</UiButton>
              </div>
            </div>
          </label>

          <template v-if="inboundModal.protocol === 'shadowsocks'">
            <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.ssMethod", "SS Method") }}</div>
              <input v-model="inboundStructured.ssMethod" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.ssPassword", "SS Password") }}</div>
              <input v-model="inboundStructured.ssPassword" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm md:col-span-2">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.ssNetwork", "SS Network") }}</div>
              <input v-model="inboundStructured.ssNetwork" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
              <input v-model="inboundStructured.ssIvCheck" type="checkbox" /> SS IV Check
            </label>
          </template>

          <template v-if="inboundModal.protocol === 'http'">
            <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.httpTimeout", "HTTP Timeout") }}</div>
              <input v-model.number="inboundStructured.httpTimeout" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.httpAllowTransparent" type="checkbox" /> {{ t("pages.inbounds.ui.allowTransparent", "Allow Transparent") }}
            </label>
            <label class="text-sm md:col-span-2">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.httpAccounts", "HTTP Accounts") }}</div>
              <div class="space-y-2">
                <div v-for="(acc, idx) in inboundStructured.httpAccounts" :key="`http-acc-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                  <input v-model="acc.user" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="user" type="text" />
                  <input v-model="acc.pass" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="pass" type="text" />
                  <UiButton size="sm" variant="outline" @click="removeHttpAccount(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
                <UiButton size="sm" variant="outline" @click="addHttpAccount">{{ t("pages.inbounds.ui.addAccount", "Add Account") }}</UiButton>
              </div>
            </label>
          </template>
          <template v-if="inboundModal.protocol === 'mixed'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.mixedTimeout", "Mixed Timeout") }}</div>
              <input v-model.number="inboundStructured.mixedTimeout" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.auth", "Auth") }}</div>
              <select v-model="inboundStructured.mixedAuth" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="password">password</option>
                <option value="noauth">noauth</option>
              </select>
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.bindIp", "Bind IP") }}</div>
              <input v-model="inboundStructured.mixedIp" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.mixedUdp" type="checkbox" /> UDP
            </label>
            <label class="text-sm md:col-span-2" v-if="inboundStructured.mixedAuth === 'password'">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.mixedAccounts", "Mixed Accounts") }}</div>
              <div class="space-y-2">
                <div v-for="(acc, idx) in inboundStructured.mixedAccounts" :key="`mixed-acc-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                  <input v-model="acc.user" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="user" type="text" />
                  <input v-model="acc.pass" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="pass" type="text" />
                  <UiButton size="sm" variant="outline" @click="removeMixedAccount(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
                <UiButton size="sm" variant="outline" @click="addMixedAccount">{{ t("pages.inbounds.ui.addAccount", "Add Account") }}</UiButton>
              </div>
            </label>
          </template>

          <template v-if="inboundModal.protocol === 'tunnel'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.targetAddress", "Target Address") }}</div>
              <input v-model="inboundStructured.tunnelAddress" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.destinationPort", "Destination Port") }}</div>
              <input v-model.number="inboundStructured.tunnelPort" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Network</div>
              <select v-model="inboundStructured.tunnelNetwork" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="tcp,udp">tcp,udp</option>
                <option value="tcp">tcp</option>
                <option value="udp">udp</option>
              </select>
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.tunnelFollowRedirect" type="checkbox" /> Follow Redirect
            </label>
            <label class="text-sm md:col-span-2">
              <div class="mb-1 text-muted-foreground">Tunnel Port Map</div>
              <div class="space-y-2">
                <div v-for="(map, idx) in inboundStructured.tunnelPortMap" :key="`tunnel-map-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                  <input v-model="map.from" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="from port" type="text" />
                  <input v-model="map.to" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="to port" type="text" />
                  <UiButton size="sm" variant="outline" @click="removeTunnelPortMap(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
                <UiButton size="sm" variant="outline" @click="addTunnelPortMap">Add Port Map</UiButton>
              </div>
            </label>
          </template>

          <template v-if="inboundModal.protocol === 'tun'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Interface Name</div>
              <input v-model="inboundStructured.tunName" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">MTU</div>
              <input v-model.number="inboundStructured.tunMtu" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">User Level</div>
              <input v-model.number="inboundStructured.tunUserLevel" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
          </template>

          <template v-if="inboundModal.protocol === 'wireguard'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Secret Key</div>
              <input v-model="inboundStructured.wgSecretKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Public Key</div>
              <input v-model="inboundStructured.wgPubKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">WireGuard MTU</div>
              <input v-model.number="inboundStructured.wgMtu" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.wgNoKernelTun" type="checkbox" /> No Kernel Tun
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Peer Private Key</div>
              <input v-model="inboundStructured.wgPeerPrivateKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Peer Public Key</div>
              <input v-model="inboundStructured.wgPeerPublicKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Peer PSK</div>
              <input v-model="inboundStructured.wgPeerPsk" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Allowed IPs (comma-separated)</div>
              <input v-model="inboundStructured.wgAllowedIPs" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Keep Alive</div>
              <input v-model.number="inboundStructured.wgKeepAlive" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
          </template>

          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">Stream Network</div>
            <select v-model="inboundStructured.streamNetwork" class="w-full rounded-md border bg-background px-3 py-2">
              <option value="tcp">tcp</option>
              <option value="ws">ws</option>
              <option value="grpc">grpc</option>
              <option value="kcp">kcp</option>
              <option value="httpupgrade">httpupgrade</option>
              <option value="xhttp">xhttp</option>
            </select>
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">Stream Security</div>
            <div class="inline-flex w-fit overflow-hidden rounded-md border bg-background">
              <button
                type="button"
                class="px-3 py-1.5 text-xs transition-colors"
                :class="inboundStructured.streamSecurity === 'none' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
                @click="inboundStructured.streamSecurity = 'none'"
              >
                None
              </button>
              <button
                type="button"
                class="border-l px-3 py-1.5 text-xs transition-colors"
                :class="inboundStructured.streamSecurity === 'reality' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
                @click="inboundStructured.streamSecurity = 'reality'"
              >
                Reality
              </button>
              <button
                type="button"
                class="border-l px-3 py-1.5 text-xs transition-colors"
                :class="inboundStructured.streamSecurity === 'tls' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
                @click="inboundStructured.streamSecurity = 'tls'"
              >
                TLS
              </button>
            </div>
          </label>

          <template v-if="inboundStructured.streamNetwork === 'tcp'">
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.tcpAcceptProxyProtocol" type="checkbox" /> TCP Accept Proxy Protocol
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">TCP Header Type</div>
              <select v-model="inboundStructured.tcpHeaderType" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="none">none</option>
                <option value="http">http</option>
              </select>
            </label>
            <template v-if="inboundStructured.tcpHeaderType === 'http'">
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Request Method</div>
                <input v-model="inboundStructured.tcpRequestMethod" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Request Path</div>
                <input v-model="inboundStructured.tcpRequestPath" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm md:col-span-2">
                <div class="mb-1 text-muted-foreground">Request Headers</div>
                <div class="space-y-2">
                  <div v-for="(h, idx) in inboundStructured.tcpRequestHeaders" :key="`tcp-req-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                    <input v-model="h.name" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="name" type="text" />
                    <input v-model="h.value" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="value" type="text" />
                    <UiButton size="sm" variant="outline" @click="removeTcpRequestHeader(idx)">{{ t("delete", "Delete") }}</UiButton>
                  </div>
                  <UiButton size="sm" variant="outline" @click="addTcpRequestHeader">Add Header</UiButton>
                </div>
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Response Status</div>
                <input v-model="inboundStructured.tcpResponseStatus" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Response Reason</div>
                <input v-model="inboundStructured.tcpResponseReason" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm md:col-span-2">
                <div class="mb-1 text-muted-foreground">Response Headers</div>
                <div class="space-y-2">
                  <div v-for="(h, idx) in inboundStructured.tcpResponseHeaders" :key="`tcp-resp-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                    <input v-model="h.name" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="name" type="text" />
                    <input v-model="h.value" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="value" type="text" />
                    <UiButton size="sm" variant="outline" @click="removeTcpResponseHeader(idx)">{{ t("delete", "Delete") }}</UiButton>
                  </div>
                  <UiButton size="sm" variant="outline" @click="addTcpResponseHeader">Add Header</UiButton>
                </div>
              </label>
            </template>
          </template>

          <template v-if="inboundStructured.streamNetwork === 'ws'">
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.wsAcceptProxyProtocol" type="checkbox" /> WS Accept Proxy Protocol
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">WS Path</div>
              <input v-model="inboundStructured.wsPath" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">WS Host</div>
              <input v-model="inboundStructured.wsHost" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">WS Heartbeat Period</div>
              <input v-model.number="inboundStructured.wsHeartbeatPeriod" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm md:col-span-2">
              <div class="mb-1 text-muted-foreground">WS Headers</div>
              <div class="space-y-2">
                <div v-for="(h, idx) in inboundStructured.wsHeaders" :key="`ws-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                  <input v-model="h.name" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="name" type="text" />
                  <input v-model="h.value" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="value" type="text" />
                  <UiButton size="sm" variant="outline" @click="removeWsHeader(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
                <UiButton size="sm" variant="outline" @click="addWsHeader">Add Header</UiButton>
              </div>
            </label>
          </template>
          <template v-if="inboundStructured.streamNetwork === 'grpc'">
            <label class="text-sm md:col-span-2">
              <div class="mb-1 text-muted-foreground">gRPC Service Name</div>
              <input v-model="inboundStructured.grpcServiceName" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">gRPC Authority</div>
              <input v-model="inboundStructured.grpcAuthority" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.grpcMultiMode" type="checkbox" /> gRPC Multi Mode
            </label>
          </template>
          <template v-if="inboundStructured.streamNetwork === 'kcp'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">KCP MTU</div>
              <input v-model.number="inboundStructured.kcpMtu" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">KCP TTI</div>
              <input v-model.number="inboundStructured.kcpTti" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">KCP Uplink Capacity</div>
              <input v-model.number="inboundStructured.kcpUpCap" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">KCP Downlink Capacity</div>
              <input v-model.number="inboundStructured.kcpDownCap" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">KCP Read Buffer</div>
              <input v-model.number="inboundStructured.kcpReadBuffer" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">KCP Write Buffer</div>
              <input v-model.number="inboundStructured.kcpWriteBuffer" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
              <input v-model="inboundStructured.kcpCongestion" type="checkbox" /> KCP Congestion
            </label>
          </template>
          <template v-if="inboundStructured.streamNetwork === 'httpupgrade'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">HTTPUpgrade Host</div>
              <input v-model="inboundStructured.huHost" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">HTTPUpgrade Path</div>
              <input v-model="inboundStructured.huPath" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
              <input v-model="inboundStructured.huAcceptProxyProtocol" type="checkbox" /> Accept Proxy Protocol
            </label>
            <label class="text-sm md:col-span-2">
              <div class="mb-1 text-muted-foreground">HTTPUpgrade Headers</div>
              <div class="space-y-2">
                <div v-for="(h, idx) in inboundStructured.huHeaders" :key="`hu-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                  <input v-model="h.name" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="name" type="text" />
                  <input v-model="h.value" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="value" type="text" />
                  <UiButton size="sm" variant="outline" @click="removeHuHeader(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
                <UiButton size="sm" variant="outline" @click="addHuHeader">Add Header</UiButton>
              </div>
            </label>
          </template>
          <template v-if="inboundStructured.streamNetwork === 'xhttp'">
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">XHTTP Host</div>
              <input v-model="inboundStructured.xhHost" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">XHTTP Path</div>
              <input v-model="inboundStructured.xhPath" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">XHTTP Mode</div>
              <select v-model="inboundStructured.xhMode" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="auto">auto</option>
                <option value="packet-up">packet-up</option>
                <option value="stream-up">stream-up</option>
              </select>
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">XHTTP Padding Bytes</div>
              <input v-model="inboundStructured.xhPaddingBytes" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
              <input v-model="inboundStructured.xhNoSSEHeader" type="checkbox" /> XHTTP No SSE Header
            </label>
            <label class="text-sm md:col-span-2">
              <div class="mb-1 text-muted-foreground">XHTTP Headers</div>
              <div class="space-y-2">
                <div v-for="(h, idx) in inboundStructured.xhHeaders" :key="`xh-${idx}`" class="grid grid-cols-[1fr_1fr_auto] gap-2">
                  <input v-model="h.name" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="name" type="text" />
                  <input v-model="h.value" class="rounded-md border bg-background px-3 py-2 text-xs" placeholder="value" type="text" />
                  <UiButton size="sm" variant="outline" @click="removeXhHeader(idx)">{{ t("delete", "Delete") }}</UiButton>
                </div>
                <UiButton size="sm" variant="outline" @click="addXhHeader">Add Header</UiButton>
              </div>
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Uplink HTTP Method</div>
              <select v-model="inboundStructured.xhUplinkHTTPMethod" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="">default</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="GET">GET</option>
              </select>
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Session Placement</div>
              <select v-model="inboundStructured.xhSessionPlacement" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="">default</option>
                <option value="path">path</option>
                <option value="header">header</option>
                <option value="cookie">cookie</option>
                <option value="query">query</option>
              </select>
            </label>
            <label class="text-sm" v-if="inboundStructured.xhSessionPlacement && inboundStructured.xhSessionPlacement !== 'path'">
              <div class="mb-1 text-muted-foreground">Session Key</div>
              <input v-model="inboundStructured.xhSessionKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Sequence Placement</div>
              <select v-model="inboundStructured.xhSeqPlacement" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="">default</option>
                <option value="path">path</option>
                <option value="header">header</option>
                <option value="cookie">cookie</option>
                <option value="query">query</option>
              </select>
            </label>
            <label class="text-sm" v-if="inboundStructured.xhSeqPlacement && inboundStructured.xhSeqPlacement !== 'path'">
              <div class="mb-1 text-muted-foreground">Sequence Key</div>
              <input v-model="inboundStructured.xhSeqKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm" v-if="inboundStructured.xhMode === 'packet-up'">
              <div class="mb-1 text-muted-foreground">Uplink Data Placement</div>
              <select v-model="inboundStructured.xhUplinkDataPlacement" class="w-full rounded-md border bg-background px-3 py-2">
                <option value="">default</option>
                <option value="body">body</option>
                <option value="header">header</option>
                <option value="query">query</option>
              </select>
            </label>
            <label class="text-sm" v-if="inboundStructured.xhMode === 'packet-up' && inboundStructured.xhUplinkDataPlacement && inboundStructured.xhUplinkDataPlacement !== 'body'">
              <div class="mb-1 text-muted-foreground">Uplink Data Key</div>
              <input v-model="inboundStructured.xhUplinkDataKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm" v-if="inboundStructured.xhMode === 'packet-up' && inboundStructured.xhUplinkDataPlacement && inboundStructured.xhUplinkDataPlacement !== 'body'">
              <div class="mb-1 text-muted-foreground">Uplink Chunk Size</div>
              <input v-model.number="inboundStructured.xhUplinkChunkSize" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm" v-if="inboundStructured.xhMode === 'packet-up'">
              <div class="mb-1 text-muted-foreground">Max Buffered Upload</div>
              <input v-model.number="inboundStructured.xhScMaxBufferedPosts" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm" v-if="inboundStructured.xhMode === 'packet-up'">
              <div class="mb-1 text-muted-foreground">Max Upload Size (Byte)</div>
              <input v-model="inboundStructured.xhScMaxEachPostBytes" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm" v-if="inboundStructured.xhMode === 'stream-up'">
              <div class="mb-1 text-muted-foreground">Stream-Up Server</div>
              <input v-model="inboundStructured.xhScStreamUpServerSecs" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
              <input v-model="inboundStructured.xhPaddingObfsMode" type="checkbox" /> Padding Obfs Mode
            </label>
            <template v-if="inboundStructured.xhPaddingObfsMode">
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Padding Key</div>
                <input v-model="inboundStructured.xhPaddingKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Padding Header</div>
                <input v-model="inboundStructured.xhPaddingHeader" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Padding Placement</div>
                <select v-model="inboundStructured.xhPaddingPlacement" class="w-full rounded-md border bg-background px-3 py-2">
                  <option value="">default</option>
                  <option value="queryInHeader">queryInHeader</option>
                  <option value="header">header</option>
                </select>
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Padding Method</div>
                <select v-model="inboundStructured.xhPaddingMethod" class="w-full rounded-md border bg-background px-3 py-2">
                  <option value="">default</option>
                  <option value="repeat-x">repeat-x</option>
                  <option value="tokenish">tokenish</option>
                </select>
              </label>
            </template>
          </template>

          <template v-if="inboundStructured.streamSecurity === 'tls'">
            <div class="grid grid-cols-1 gap-2 rounded-lg border bg-muted/20 p-3 md:col-span-2 md:grid-cols-2">
              <div class="text-xs font-medium text-muted-foreground md:col-span-2">TLS Settings</div>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS Server Name</div>
                <input v-model="inboundStructured.tlsServerName" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS ALPN (comma-separated)</div>
                <input v-model="inboundStructured.tlsAlpn" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS Min Version</div>
                <input v-model="inboundStructured.tlsMinVersion" class="w-full rounded-md border bg-background px-3 py-2" placeholder="1.2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS Max Version</div>
                <input v-model="inboundStructured.tlsMaxVersion" class="w-full rounded-md border bg-background px-3 py-2" placeholder="1.3" type="text" />
              </label>
              <label class="text-sm md:col-span-2">
                <div class="mb-1 text-muted-foreground">TLS Cipher Suites (comma-separated)</div>
                <input v-model="inboundStructured.tlsCipherSuites" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS Fingerprint</div>
                <input v-model="inboundStructured.tlsFingerprint" class="w-full rounded-md border bg-background px-3 py-2" placeholder="chrome/firefox/..." type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS Certificate File</div>
                <input v-model="inboundStructured.tlsCertFile" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">TLS Key File</div>
                <input v-model="inboundStructured.tlsKeyFile" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.tlsRejectUnknownSni" type="checkbox" /> Reject Unknown SNI
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.tlsAllowInsecure" type="checkbox" /> Allow Insecure
              </label>
              <label class="text-sm md:col-span-2">
                <div class="mb-1 text-muted-foreground">ECH Helper</div>
                <div class="grid grid-cols-[1fr_auto] gap-2">
                  <input v-model="echSniInput" class="rounded-md border bg-background px-3 py-2" placeholder="SNI for ECH (e.g. example.com)" type="text" />
                  <UiButton size="sm" variant="outline" @click="generateEchCert">Generate ECH</UiButton>
                </div>
                <textarea v-if="echConfigListPreview" v-model="echConfigListPreview" class="mt-2 h-16 w-full json-field px-3 py-2 text-xs" readonly />
                <textarea v-if="echServerKeysPreview" v-model="echServerKeysPreview" class="mt-2 h-16 w-full json-field px-3 py-2 text-xs" readonly />
              </label>
            </div>
          </template>

          <template v-if="inboundStructured.streamSecurity === 'reality'">
            <div class="grid grid-cols-1 gap-2 rounded-lg border bg-muted/20 p-3 md:col-span-2 md:grid-cols-2">
              <div class="text-xs font-medium text-muted-foreground md:col-span-2">Reality Settings</div>
              <label class="text-sm md:col-span-2">
                <div class="mb-1 text-muted-foreground">Reality Helper</div>
                <div class="flex flex-wrap items-center gap-2 rounded-md border px-3 py-2">
                  <UiButton size="sm" variant="outline" @click="generateRealityKeypair">Generate X25519</UiButton>
                  <UiButton size="sm" variant="outline" @click="generateMlkem768">Generate ML-KEM-768</UiButton>
                  <UiButton size="sm" variant="outline" @click="generateMldsa65">Generate ML-DSA-65</UiButton>
                  <span v-if="realityPublicKeyPreview" class="text-xs text-muted-foreground">Public: {{ realityPublicKeyPreview }}</span>
                </div>
                <textarea v-if="mlkemClientPreview" v-model="mlkemClientPreview" class="mt-2 h-14 w-full json-field px-3 py-2 text-xs" readonly />
                <textarea v-if="mldsaVerifyPreview" v-model="mldsaVerifyPreview" class="mt-2 h-14 w-full json-field px-3 py-2 text-xs" readonly />
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.realityShow" type="checkbox" /> Reality Show
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Reality Dest</div>
                <input v-model="inboundStructured.realityDest" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm md:col-span-2">
                <div class="mb-1 text-muted-foreground">Reality Server Names (comma-separated)</div>
                <input v-model="inboundStructured.realityServerNames" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Reality Private Key</div>
                <input v-model="inboundStructured.realityPrivateKey" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Reality Short IDs (comma-separated)</div>
                <input v-model="inboundStructured.realityShortIds" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Reality Max Time Diff</div>
                <input v-model.number="inboundStructured.realityMaxTimeDiff" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Reality Min Client</div>
                <input v-model="inboundStructured.realityMinClient" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
              <label class="text-sm">
                <div class="mb-1 text-muted-foreground">Reality Max Client</div>
                <input v-model="inboundStructured.realityMaxClient" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              </label>
            </div>
          </template>

          <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm md:col-span-2">
            <input v-model="inboundStructured.sockoptEnabled" type="checkbox" /> Enable Sockopt
          </label>
          <template v-if="inboundStructured.sockoptEnabled">
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.sockoptAcceptProxyProtocol" type="checkbox" /> Sockopt Accept Proxy Protocol
            </label>
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.sockoptTcpFastOpen" type="checkbox" /> Sockopt TCP Fast Open
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Sockopt KeepAlive Interval</div>
              <input v-model.number="inboundStructured.sockoptTcpKeepAliveInterval" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Sockopt Mark</div>
              <input v-model.number="inboundStructured.sockoptMark" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Sockopt Domain Strategy</div>
              <input v-model="inboundStructured.sockoptDomainStrategy" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
            <label class="text-sm">
              <div class="mb-1 text-muted-foreground">Sockopt Dialer Proxy</div>
              <input v-model="inboundStructured.sockoptDialerProxy" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
            </label>
          </template>
        </div>

        <div class="mt-3 rounded-lg border bg-muted/20 p-3">
          <div class="mb-3 text-sm font-medium">External Proxy</div>
          <div class="space-y-2">
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.externalProxyEnabled" type="checkbox" /> Enabled
            </label>
            <div v-if="inboundStructured.externalProxyEnabled" class="grid grid-cols-1 gap-2 md:grid-cols-3">
              <input
                v-model="inboundStructured.externalProxyAddress"
                class="rounded-md border bg-background px-3 py-2 text-sm"
                placeholder="Host / Address"
                type="text"
              />
              <input
                v-model="inboundStructured.externalProxyPort"
                class="rounded-md border bg-background px-3 py-2 text-sm"
                placeholder="Port"
                type="text"
              />
              <input v-model="inboundStructured.externalProxyRemark" class="rounded-md border bg-background px-3 py-2 text-sm" placeholder="Remark (optional)" type="text" />
            </div>
          </div>
        </div>

        <div class="mt-3 rounded-lg border bg-muted/20 p-3">
          <div class="mb-3 text-sm font-medium">Sniffing</div>
          <div class="space-y-2">
            <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input v-model="inboundStructured.sniffEnabled" type="checkbox" /> Enabled
            </label>
            <div class="grid grid-cols-2 gap-2 md:grid-cols-4">
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.sniffHttp" type="checkbox" /> HTTP
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.sniffTls" type="checkbox" /> TLS
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.sniffQuic" type="checkbox" /> QUIC
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.sniffFakeDns" type="checkbox" /> FAKEDNS
              </label>
            </div>
            <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.sniffMetadataOnly" type="checkbox" /> Metadata Only
              </label>
              <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
                <input v-model="inboundStructured.sniffRouteOnly" type="checkbox" /> Route Only
              </label>
            </div>
          </div>
        </div>
      </div>

      <div class="mt-4 flex justify-end gap-2">
        <UiButton :disabled="modalBusy" size="sm" variant="outline" @click="closeInboundModal">{{ t("cancel", "Cancel") }}</UiButton>
        <UiButton :disabled="modalBusy" size="sm" @click="saveInboundModal">{{ inboundModal.mode === "add" ? t("pages.inbounds.create", "Create") : t("pages.settings.save", "Save") }}</UiButton>
      </div>
    </div>
  </div>

  <div v-if="clientsModal.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-2 sm:p-4">
    <div class="max-h-[96vh] w-[min(96vw,1700px)] max-w-none overflow-y-auto rounded-xl border bg-card p-3 shadow-lg sm:p-4">
      <div class="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h2 class="text-lg font-semibold">{{ t("clients", "Clients") }} - {{ clientsModal.inboundRemark }} (#{{ clientsModal.inboundId }})</h2>
        <div class="flex flex-wrap items-center gap-2">
          <select v-model="clientSortBy" class="w-full rounded-md border bg-background px-2 py-1 text-sm sm:w-auto" @change="saveClientSortPreference">
            <option value="default">{{ t("default", "Sort: Default") }}</option>
            <option value="email">{{ t("email", "Sort: Email") }}</option>
            <option value="traffic">{{ t("pages.inbounds.traffic", "Sort: Traffic") }}</option>
            <option value="expiry">{{ t("pages.inbounds.expireDate", "Sort: Expiry") }}</option>
          </select>
          <UiButton class="w-full sm:w-auto" size="sm" variant="outline" @click="openAddClientEditor">{{ t("pages.client.add", "Add Client") }}</UiButton>
          <UiButton class="w-full sm:w-auto" size="sm" variant="outline" @click="openClientBulk">{{ t("pages.client.bulk", "Bulk Add") }}</UiButton>
        </div>
      </div>
      <div class="mb-3 flex items-center gap-2 rounded-md border bg-muted/40 px-3 py-2">
        <Search class="h-4 w-4 text-muted-foreground" />
        <input
          v-model="clientSearch"
          class="w-full border-none bg-transparent text-sm outline-none"
          :placeholder="t('search', 'Search') + '...'"
          type="text"
        />
      </div>

      <div class="space-y-2 pb-4 sm:hidden">
        <div v-for="(client, index) in filteredModalClients" :key="`mclient-${client.email || client.id || client.password || 'client'}-${index}`" class="rounded-lg border border-border/80 bg-card/40 p-3">
          <div class="mb-2 truncate text-sm font-semibold">{{ client.email || "-" }}</div>
          <div class="grid grid-cols-2 gap-x-3 gap-y-1 text-xs">
            <div class="text-muted-foreground">{{ t("enable", "Enable") }}</div>
            <div>{{ client.enable ?? true ? "yes" : "no" }}</div>
            <div class="text-muted-foreground">{{ t("pages.inbounds.ui.totalGb", "Total GB") }}</div>
            <div>{{ formatTotalGB(client.totalGB) }}</div>
            <div class="text-muted-foreground">{{ t("usage", "Usage") }}</div>
            <div class="truncate" :title="clientUsageText(client)">{{ clientUsageText(client) }}</div>
            <div class="text-muted-foreground">{{ t("pages.inbounds.expireDate", "Expiry") }}</div>
            <div class="truncate" :title="formatExpiry(client.expiryTime || 0)">{{ formatExpiry(client.expiryTime || 0) }}</div>
          </div>
          <div class="mt-3 grid grid-cols-2 gap-2">
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('info', client)">{{ t("info", "Info") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('qr', client)">{{ t("qrCode", "QR Code") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('configs', client)">{{ t("pages.inbounds.export", "Configs") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('edit', client)">{{ t("edit", "Edit") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('reset', client)">{{ t("pages.inbounds.resetTraffic", "Reset Traffic") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('ipLogs', client)">{{ t("pages.inbounds.ui.ipLogs", "IP Logs") }}</UiButton>
            <UiButton size="sm" variant="outline" @click="onClientActionMenu('clearIps', client)">{{ t("reset", "Clear IPs") }}</UiButton>
            <UiButton size="sm" variant="outline" class="text-red-700 hover:bg-red-50" @click="onClientActionMenu('delete', client)">{{ t("delete", "Delete") }}</UiButton>
          </div>
        </div>
        <div v-if="filteredModalClients.length === 0" class="rounded-md border px-3 py-6 text-center text-sm text-muted-foreground">
          {{ t("none", "No clients in this inbound") }}
        </div>
      </div>

      <div class="relative hidden overflow-x-auto [overflow-y:visible] pb-16 sm:block">
        <table class="tx-table-center w-full min-w-[720px] border-collapse text-sm">
          <thead>
            <tr class="border-b text-left text-muted-foreground">
              <th class="px-3 py-2 font-medium">{{ t("email", "Email") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("enable", "Enable") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.inbounds.ui.totalGb", "Total GB") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("usage", "Usage") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.inbounds.expireDate", "Expiry") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.settings.actions", "Actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(client, index) in filteredModalClients" :key="`${client.email || client.id || client.password || 'client'}-${index}`" class="border-b">
              <td class="px-3 py-2">{{ client.email || "-" }}</td>
              <td class="px-3 py-2">{{ client.enable ?? true ? "yes" : "no" }}</td>
              <td class="px-3 py-2">{{ formatTotalGB(client.totalGB) }}</td>
              <td class="px-3 py-2">{{ clientUsageText(client) }}</td>
              <td class="px-3 py-2">{{ formatExpiry(client.expiryTime || 0) }}</td>
              <td class="relative overflow-visible px-3 py-2">
                <button class="action-trigger action-trigger-sm" type="button" @click.stop="openClientActionsFloating($event, client)">
                  <MoreVertical class="h-4 w-4" />
                  {{ t("pages.settings.actions", "Actions") }}
                </button>
              </td>
            </tr>
            <tr v-if="filteredModalClients.length === 0">
              <td class="px-3 py-6 text-center text-muted-foreground" colspan="6">{{ t("none", "No clients in this inbound") }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div
        v-if="clientActionsFloating.open && clientActionsFloating.client"
        data-client-floating-menu="1"
        :class="clientActionsFloating.place === 'top' ? '-translate-y-full' : ''"
        :style="{ left: `${clientActionsFloating.left}px`, top: `${clientActionsFloating.top}px` }"
        class="fixed z-[140] min-w-[170px] rounded-xl border border-border bg-card/95 p-1.5 shadow-2xl backdrop-blur-sm menu-popover"
        @click.stop
      >
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('info', clientActionsFloating.client)">{{ t("info", "Info") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('qr', clientActionsFloating.client)">{{ t("qrCode", "QR Code") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('configs', clientActionsFloating.client)">{{ t("pages.inbounds.export", "Configs") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('edit', clientActionsFloating.client)">{{ t("edit", "Edit") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('toggleEnabled', clientActionsFloating.client)">{{ (clientActionsFloating.client.enable ?? true) ? t("disabled", "Disable") : t("enabled", "Enable") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('reset', clientActionsFloating.client)">{{ t("pages.inbounds.resetTraffic", "Reset Traffic") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('ipLogs', clientActionsFloating.client)">{{ t("pages.inbounds.ui.ipLogs", "IP Logs") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs hover:bg-muted" @click="onClientActionMenu('clearIps', clientActionsFloating.client)">{{ t("reset", "Clear IPs") }}</button>
        <button data-menu-action="1" class="w-full rounded px-2 py-2 text-left text-xs text-red-700 hover:bg-red-50" @click="onClientActionMenu('delete', clientActionsFloating.client)">{{ t("delete", "Delete") }}</button>
      </div>
      <div class="mt-2">
        <UiButton size="sm" variant="outline" @click="loadClientsTraffic">{{ t("pages.inbounds.ui.refreshUsage", "Refresh Usage") }}</UiButton>
      </div>

      <div v-if="clientBulk.open" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 p-2 sm:p-4">
        <div class="max-h-[92vh] w-[min(96vw,1400px)] max-w-none overflow-y-auto rounded-xl border bg-card p-3 shadow-lg sm:p-4">
          <h3 class="mb-2 text-sm font-semibold">{{ t("pages.client.bulk", "Bulk Add Clients") }}</h3>
          <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.client.method", "Method") }}</div>
            <select v-model.number="clientBulk.method" class="w-full rounded-md border bg-background px-3 py-2">
              <option :value="0">Random</option>
              <option :value="1">Random + Prefix</option>
              <option :value="2">Prefix + Num</option>
              <option :value="3">Prefix + Num + Postfix</option>
              <option :value="4">Prefix + Num + Postfix (strict)</option>
            </select>
          </label>
          <label class="text-sm" v-if="clientBulk.method < 2">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.quantity", "Quantity") }}</div>
            <input v-model="clientBulk.quantity" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="1" max="100" />
          </label>
          <label class="text-sm" v-if="clientBulk.method > 1">
            <div class="mb-1 text-muted-foreground">{{ t("pages.client.first", "First Num") }}</div>
            <input v-model="clientBulk.firstNum" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="1" />
          </label>
          <label class="text-sm" v-if="clientBulk.method > 1">
            <div class="mb-1 text-muted-foreground">{{ t("pages.client.last", "Last Num") }}</div>
            <input v-model="clientBulk.lastNum" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="1" />
          </label>
          <label class="text-sm" v-if="clientBulk.method > 0">
            <div class="mb-1 text-muted-foreground">{{ t("pages.client.prefix", "Prefix") }}</div>
            <input v-model="clientBulk.prefix" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
          </label>
          <label class="text-sm" v-if="clientBulk.method > 2">
            <div class="mb-1 text-muted-foreground">{{ t("pages.client.postfix", "Postfix") }}</div>
            <input v-model="clientBulk.postfix" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
          </label>
          <label class="text-sm" v-if="clientsModal.protocol === 'vmess'">
            <div class="mb-1 text-muted-foreground">{{ t("security", "Security") }}</div>
            <select v-model="clientBulk.security" class="w-full rounded-md border bg-background px-3 py-2">
              <option v-for="s in vmessSecurityOptions" :key="s" :value="s">{{ s }}</option>
            </select>
          </label>
          <label class="text-sm" v-if="clientsModal.protocol === 'vless'">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.flow", "Flow") }}</div>
            <input v-model="clientBulk.flow" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.subId", "Sub ID") }}</div>
            <input v-model="clientBulk.subId" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.telegramId", "Telegram ID") }}</div>
            <input v-model="clientBulk.tgId" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="0" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.IPLimit", "Limit IP") }}</div>
            <input v-model="clientBulk.limitIp" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="0" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.totalGb", "Total GB") }}</div>
            <input v-model="clientBulk.totalGB" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="0" />
          </label>
          <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
            <input v-model="clientBulk.delayedStart" type="checkbox" /> Delayed Start
          </label>
          <label class="text-sm" v-if="clientBulk.delayedStart">
            <div class="mb-1 text-muted-foreground">{{ t("pages.client.expireDays", "Expire Days") }}</div>
            <input v-model="clientBulk.delayedExpireDays" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="0" />
          </label>
          <label class="text-sm" v-else>
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.expiryDateTime", "Expiry Date & Time") }}</div>
            <input v-model="clientBulkExpiryInput" class="w-full rounded-md border bg-background px-3 py-2" type="datetime-local" @input="onClientBulkExpiryInputChange" />
            <div class="mt-1 text-xs text-muted-foreground">
              {{ formatExpiry(numberFromString(clientBulk.expiryTime, 0)) }} | unix ms: {{ numberFromString(clientBulk.expiryTime, 0) || 0 }}
            </div>
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("reset", "Reset") }}</div>
            <input v-model="clientBulk.reset" class="w-full rounded-md border bg-background px-3 py-2" type="number" min="0" />
          </label>
          </div>
          <div class="mt-3 flex justify-end gap-2">
            <UiButton size="sm" variant="outline" @click="closeClientBulk">{{ t("cancel", "Cancel") }}</UiButton>
            <UiButton size="sm" @click="applyBulkClients">{{ t("confirm", "Apply Bulk") }}</UiButton>
          </div>
        </div>
      </div>

      <div v-if="clientEditor.open" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 p-2 sm:p-4">
        <div class="max-h-[92vh] w-[min(96vw,1450px)] max-w-none overflow-y-auto rounded-xl border bg-card p-3 shadow-lg sm:p-4">
          <h3 class="mb-2 text-sm font-semibold">{{ clientEditor.mode === "add" ? t("pages.client.add", "Add Client") : t("pages.client.edit", "Edit Client") }}</h3>
          <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("email", "Email") }}</div>
            <input v-model="clientEditor.email" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.id", "ID") }}</div>
            <div class="flex gap-2">
              <input :disabled="clientsModal.protocol === 'trojan' || clientsModal.protocol === 'shadowsocks'" v-model="clientEditor.id" class="w-full rounded-md border bg-background px-3 py-2 disabled:opacity-60" type="text" />
              <UiButton v-if="clientsModal.protocol !== 'trojan' && clientsModal.protocol !== 'shadowsocks'" size="sm" variant="outline" @click="generateClientUuid">{{ t("pages.inbounds.ui.generate", "Gen") }}</UiButton>
            </div>
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("password", "Password") }}</div>
            <div class="flex gap-2">
              <input :disabled="clientsModal.protocol !== 'trojan' && clientsModal.protocol !== 'shadowsocks'" v-model="clientEditor.password" class="w-full rounded-md border bg-background px-3 py-2 disabled:opacity-60" type="text" />
              <UiButton v-if="clientsModal.protocol === 'trojan' || clientsModal.protocol === 'shadowsocks'" size="sm" variant="outline" @click="generateClientPassword">{{ t("pages.inbounds.ui.generate", "Gen") }}</UiButton>
            </div>
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.flow", "Flow") }}</div>
            <input :disabled="clientsModal.protocol !== 'vless'" v-model="clientEditor.flow" class="w-full rounded-md border bg-background px-3 py-2 disabled:opacity-60" type="text" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("security", "Security") }}</div>
            <select v-if="clientsModal.protocol === 'vmess'" v-model="clientEditor.security" class="w-full rounded-md border bg-background px-3 py-2">
              <option v-for="s in vmessSecurityOptions" :key="s" :value="s">{{ s }}</option>
            </select>
            <input v-else :disabled="true" v-model="clientEditor.security" class="w-full rounded-md border bg-background px-3 py-2 disabled:opacity-60" type="text" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("enable", "Enable") }}</div>
            <input v-model="clientEditor.enable" type="checkbox" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.totalGb", "Total GB") }}</div>
            <input v-model="clientEditor.totalGB" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
          </label>
          <label class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
            <input v-model="clientEditor.delayedStart" type="checkbox" /> Delayed Start
          </label>
          <label class="text-sm" v-if="clientEditor.delayedStart">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.expireDaysAfterFirstUse", "Expire Days (after first use)") }}</div>
            <input v-model="clientEditor.delayedExpireDays" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.expiryDateTime", "Expiry Date & Time") }}</div>
            <input
              :disabled="clientEditor.delayedStart"
              v-model="clientEditorExpiryInput"
              class="w-full rounded-md border bg-background px-3 py-2 disabled:opacity-60"
              type="datetime-local"
              @input="onClientEditorExpiryInputChange"
            />
            <div class="mt-1 text-xs text-muted-foreground">
              {{ formatExpiry(numberFromString(clientEditor.expiryTime, 0)) }} | unix ms: {{ numberFromString(clientEditor.expiryTime, 0) || 0 }}
            </div>
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.IPLimit", "Limit IP") }}</div>
            <input v-model="clientEditor.limitIp" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.subId", "Sub ID") }}</div>
            <div class="flex gap-2">
              <input v-model="clientEditor.subId" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
              <UiButton size="sm" variant="outline" @click="generateClientSubId">{{ t("pages.inbounds.ui.generate", "Gen") }}</UiButton>
            </div>
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("comment", "Comment") }}</div>
            <input v-model="clientEditor.comment" class="w-full rounded-md border bg-background px-3 py-2" type="text" />
          </label>
          <label class="text-sm">
            <div class="mb-1 text-muted-foreground">{{ t("reset", "Reset") }}</div>
            <input v-model="clientEditor.reset" class="w-full rounded-md border bg-background px-3 py-2" type="number" />
          </label>
          </div>
          <div class="mt-3 flex justify-end gap-2">
            <UiButton size="sm" variant="outline" @click="closeClientEditor">{{ t("cancel", "Cancel") }}</UiButton>
            <UiButton size="sm" @click="upsertClientFromEditor">{{ t("confirm", "Apply") }}</UiButton>
          </div>
        </div>
      </div>

      <div class="mt-4 flex justify-end gap-2">
        <UiButton :disabled="modalBusy" size="sm" variant="outline" @click="closeClientsModal">{{ t("close", "Close") }}</UiButton>
        <UiButton :disabled="modalBusy" size="sm" @click="saveClientsModal">{{ t("pages.settings.save", "Save Clients") }}</UiButton>
      </div>
    </div>
  </div>

  <div v-if="importModal.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="w-[min(96vw,1400px)] max-w-none rounded-xl border bg-card p-4 shadow-lg">
      <h2 class="mb-3 text-lg font-semibold">{{ t("pages.inbounds.importInbound", "Import Inbound JSON") }}</h2>
      <textarea
        v-model="importModal.data"
        class="h-72 w-full json-field px-3 py-2 text-xs"
        :placeholder="t('pages.inbounds.ui.importPlaceholder', 'Paste exported inbound JSON here')"
      />
      <div class="mt-4 flex justify-end gap-2">
        <UiButton :disabled="importBusy" size="sm" variant="outline" @click="closeImportModal">{{ t("cancel", "Cancel") }}</UiButton>
        <UiButton :disabled="importBusy" size="sm" @click="importInboundFromJSON">{{ t("import", "Import") }}</UiButton>
      </div>
    </div>
  </div>

  <div v-if="detailModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="max-h-[92vh] w-[min(96vw,1700px)] max-w-none overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-lg font-semibold">{{ t("pages.inbounds.details", "Inbound Details") }}</h2>
        <UiButton size="sm" variant="outline" @click="detailModalOpen = false">{{ t("close", "Close") }}</UiButton>
      </div>
      <div v-if="detailLoading" class="py-8 text-sm text-muted-foreground">{{ t("loading", "Loading details...") }}</div>
      <div v-else class="grid grid-cols-1 gap-3 lg:grid-cols-2">
        <div class="rounded-lg border p-3">
          <div class="mb-2 text-sm font-semibold">{{ t("pages.inbounds.exportInbound", "Inbound JSON") }}</div>
          <pre
            class="json-codeblock h-[420px] overflow-auto whitespace-pre-wrap break-words p-3 text-xs"
            v-html="highlightedDetailInboundHtml"
          />
        </div>
        <div class="rounded-lg border p-3">
          <div class="mb-2 text-sm font-semibold">{{ t("pages.inbounds.ui.clientTrafficRows", "Client Traffic Rows") }}</div>
          <div class="overflow-x-auto">
            <table class="tx-table-center w-full min-w-[640px] border-collapse text-xs">
              <thead>
                <tr class="border-b text-left text-muted-foreground">
                  <th class="px-2 py-1">{{ t("email", "Email") }}</th>
                  <th class="px-2 py-1">{{ t("upload", "Upload") }}</th>
                  <th class="px-2 py-1">{{ t("download", "Download") }}</th>
                  <th class="px-2 py-1">{{ t("traffic", "Total") }}</th>
                  <th class="px-2 py-1">{{ t("pages.inbounds.expireDate", "Expiry") }}</th>
                  <th class="px-2 py-1">{{ t("enable", "Enable") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(t, idx) in detailClientTraffics" :key="`dt-${idx}`" class="border-b">
                  <td class="px-2 py-1">{{ t.email || "-" }}</td>
                  <td class="px-2 py-1">{{ sizeFormat(t.up || 0) }}</td>
                  <td class="px-2 py-1">{{ sizeFormat(t.down || 0) }}</td>
                  <td class="px-2 py-1">{{ sizeFormat(t.total || 0) }}</td>
                  <td class="px-2 py-1">{{ formatExpiry(t.expiryTime || 0) }}</td>
                  <td class="px-2 py-1">{{ t.enable ?? true ? "yes" : "no" }}</td>
                </tr>
                <tr v-if="detailClientTraffics.length === 0">
                  <td class="px-2 py-4 text-center text-muted-foreground" colspan="6">{{ t("pages.inbounds.ui.noClientTrafficRows", "No client traffic rows") }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div v-if="depletedModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="max-h-[92vh] w-[min(96vw,1600px)] max-w-none overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-lg font-semibold">{{ t("depleted", "Depleted Clients") }}</h2>
        <div class="flex gap-2">
          <UiButton size="sm" variant="outline" @click="openDepletedClients">{{ t("reset", "Refresh") }}</UiButton>
          <UiButton size="sm" variant="outline" @click="depletedModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="tx-table-center w-full min-w-[860px] border-collapse text-sm">
          <thead>
            <tr class="border-b text-left text-muted-foreground">
              <th class="px-3 py-2 font-medium">{{ t("inbound", "Inbound") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("email", "Email") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.xray.settings", "Raw") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.settings.actions", "Actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in depletedClients" :key="`dep-${idx}`" class="border-b">
              <td class="px-3 py-2">{{ clientInboundId(item) || "-" }}</td>
              <td class="px-3 py-2">{{ clientEmail(item) || "-" }}</td>
              <td class="px-3 py-2 font-mono text-xs">{{ JSON.stringify(item) }}</td>
              <td class="px-3 py-2">
                <UiButton
                  :disabled="!clientInboundId(item) || !clientEmail(item)"
                  size="sm"
                  variant="outline"
                  @click="deleteClientByEmail(clientInboundId(item), clientEmail(item))"
                >
                  {{ t("delete", "Delete") }}
                </UiButton>
              </td>
            </tr>
            <tr v-if="depletedClients.length === 0">
              <td class="px-3 py-6 text-center text-muted-foreground" colspan="4">{{ t("pages.inbounds.ui.noDepletedClients", "No depleted clients") }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <div v-if="onlineModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-lg font-semibold">{{ t("online", "Online Clients") }}</h2>
        <div class="flex gap-2">
          <UiButton size="sm" variant="outline" @click="refreshOnlineClients">{{ t("reset", "Refresh") }}</UiButton>
          <UiButton size="sm" variant="outline" @click="onlineModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="tx-table-center w-full min-w-[420px] border-collapse text-sm">
          <thead>
            <tr class="border-b text-left text-muted-foreground">
              <th class="px-3 py-2 font-medium">#</th>
              <th class="px-3 py-2 font-medium">{{ t("email", "Email") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(email, idx) in onlineClients" :key="`online-${email}-${idx}`" class="border-b">
              <td class="px-3 py-2">{{ idx + 1 }}</td>
              <td class="px-3 py-2">{{ email }}</td>
            </tr>
            <tr v-if="onlineClients.length === 0">
              <td class="px-3 py-6 text-center text-muted-foreground" colspan="2">{{ t("none", "No online clients") }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <div v-if="clientInfoModalOpen && clientInfoData" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-lg font-semibold">{{ t("pages.inbounds.info", "Client Info") }}</h2>
        <UiButton size="sm" variant="outline" @click="clientInfoModalOpen = false">{{ t("close", "Close") }}</UiButton>
      </div>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("email", "Email") }}</div>
          <input :value="clientInfoData.email || ''" class="w-full rounded-md border bg-background px-3 py-2" readonly type="text" />
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.idLabel", "ID") }}</div>
          <div class="flex gap-2">
            <input :value="clientInfoData.id || ''" class="w-full rounded-md border bg-background px-3 py-2" readonly type="text" />
            <UiButton size="sm" variant="outline" @click="copyToClipboard(clientInfoData.id || '')">{{ t("copy", "Copy") }}</UiButton>
          </div>
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("password", "Password") }}</div>
          <div class="flex gap-2">
            <input :value="clientInfoData.password || ''" class="w-full rounded-md border bg-background px-3 py-2" readonly type="text" />
            <UiButton size="sm" variant="outline" @click="copyToClipboard(clientInfoData.password || '')">{{ t("copy", "Copy") }}</UiButton>
          </div>
        </label>
        <label class="text-sm">
          <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.subIdLabel", "Sub ID") }}</div>
          <div class="flex gap-2">
            <input :value="clientInfoData.subId || ''" class="w-full rounded-md border bg-background px-3 py-2" readonly type="text" />
            <UiButton size="sm" variant="outline" @click="copyToClipboard(clientInfoData.subId || '')">{{ t("copy", "Copy") }}</UiButton>
          </div>
        </label>
      </div>
      <label class="mt-3 block text-sm">
        <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.subscriptionLink", "Subscription Link") }}</div>
        <div class="flex gap-2">
          <input :value="clientSubLink(clientInfoData)" class="w-full rounded-md border bg-background px-3 py-2" readonly type="text" />
          <UiButton size="sm" variant="outline" @click="copyToClipboard(clientSubLink(clientInfoData))">{{ t("copy", "Copy") }}</UiButton>
        </div>
      </label>
      <label class="mt-3 block text-sm">
        <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.rawClientJson", "Raw Client JSON") }}</div>
        <pre
          class="json-codeblock h-40 overflow-auto whitespace-pre-wrap break-words p-3 text-xs"
          v-html="highlightedClientInfoHtml"
        />
      </label>
    </div>
  </div>

  <div v-if="clientQrModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-2 sm:p-4">
    <div class="w-full max-w-md rounded-xl border bg-card p-3 shadow-lg sm:p-4">
      <div class="mb-3 flex items-center justify-between gap-2">
        <h2 class="text-lg font-semibold">{{ t("qrCode", "Client QR Code") }}</h2>
        <UiButton size="sm" variant="outline" @click="clientQrModalOpen = false">{{ t("close", "Close") }}</UiButton>
      </div>
      <div class="flex justify-center">
        <img v-if="clientQrImageSrc" :src="clientQrImageSrc" alt="Client QR" class="h-56 w-56 rounded-md border bg-white p-2 sm:h-72 sm:w-72" />
      </div>
      <label class="mt-3 block text-sm">
        <div class="mb-1 text-muted-foreground">{{ t("pages.inbounds.ui.qrContent", "QR Content") }}</div>
        <div class="flex flex-col gap-2 sm:flex-row">
          <input :value="clientQrValue" class="w-full rounded-md border bg-background px-3 py-2" readonly type="text" />
          <UiButton class="w-full sm:w-auto" size="sm" variant="outline" @click="copyToClipboard(clientQrValue)">{{ t("copy", "Copy") }}</UiButton>
        </div>
      </label>
    </div>
  </div>

  <div v-if="clientConfigsModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-2 sm:p-4">
    <div class="max-h-[90vh] w-[min(94vw,760px)] overflow-y-auto rounded-xl border bg-card p-2.5 shadow-lg sm:p-3">
      <div class="mb-2 flex items-center justify-between gap-2">
        <h2 class="text-base font-semibold">{{ t("pages.inbounds.export", "Client URIs") }}</h2>
        <UiButton size="sm" variant="outline" @click="clientConfigsModalOpen = false">{{ t("close", "Close") }}</UiButton>
      </div>
      <div class="mb-2 rounded-lg border p-2">
        <div class="mb-1 text-xs font-medium text-muted-foreground">{{ t("pages.inbounds.ui.configQr", "Config QR") }}</div>
        <div class="flex justify-center">
          <img
            v-if="clientConfigsQrImageSrc"
            :src="clientConfigsQrImageSrc"
            alt="Client Config QR"
            class="h-44 w-44 rounded-md border bg-white p-2 sm:h-52 sm:w-52"
          />
          <div v-else class="flex h-44 w-44 items-center justify-center rounded-md border bg-muted/30 text-[10px] text-muted-foreground sm:h-52 sm:w-52">
            QR unavailable
          </div>
        </div>
      </div>
      <pre
        class="max-h-[56vh] overflow-auto whitespace-pre-wrap break-all rounded-md border bg-background p-2 font-mono text-[10px] leading-4 sm:text-[11px] sm:leading-5"
        v-html="highlightedClientConfigsHtml"
      />
      <div class="mt-2 flex flex-col justify-end gap-2 sm:flex-row">
        <UiButton class="w-full sm:w-auto" size="sm" variant="outline" @click="copyToClipboard(clientConfigsText)">{{ t("pages.inbounds.ui.copyUris", "Copy URIs") }}</UiButton>
      </div>
    </div>
  </div>

  <div v-if="clientIpLogsModalOpen" class="fixed inset-0 z-[65] flex items-center justify-center bg-black/50 p-2 sm:p-4">
    <div class="max-h-[90vh] w-[min(96vw,900px)] overflow-y-auto rounded-xl border bg-card p-3 shadow-lg sm:p-4">
      <div class="mb-3 flex items-center justify-between gap-2">
        <h2 class="text-lg font-semibold">{{ t("pages.inbounds.ui.ipLogs", "IP Logs") }} - {{ clientIpLogsEmail || "-" }}</h2>
        <div class="flex gap-2">
          <UiButton size="sm" variant="outline" @click="fetchClientIps(clientIpLogsEmail)">{{ t("reset", "Refresh") }}</UiButton>
          <UiButton size="sm" variant="outline" @click="clearClientIps(clientIpLogsEmail)">{{ t("clearIPs", "Clear") }}</UiButton>
          <UiButton size="sm" variant="outline" @click="clientIpLogsModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
      </div>
      <textarea
        v-model="clientIpLogs"
        class="h-[360px] w-full json-field px-3 py-2 text-xs"
        :placeholder="t('pages.inbounds.ui.noIpLogRecord', 'No IP log record')"
        readonly
      />
    </div>
  </div>

  <div v-if="disabledModalOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
    <div class="max-h-[92vh] w-[min(96vw,1600px)] max-w-none overflow-y-auto rounded-xl border bg-card p-4 shadow-lg">
      <div class="mb-3 flex items-center justify-between">
        <h2 class="text-lg font-semibold">{{ t("disabled", "Disabled Clients") }}</h2>
        <div class="flex gap-2">
          <UiButton size="sm" variant="outline" @click="openDisabledClients">{{ t("reset", "Refresh") }}</UiButton>
          <UiButton size="sm" variant="outline" @click="disabledModalOpen = false">{{ t("close", "Close") }}</UiButton>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="tx-table-center w-full min-w-[860px] border-collapse text-sm">
          <thead>
            <tr class="border-b text-left text-muted-foreground">
              <th class="px-3 py-2 font-medium">{{ t("inbound", "Inbound") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("email", "Email") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.xray.settings", "Raw") }}</th>
              <th class="px-3 py-2 font-medium">{{ t("pages.settings.actions", "Actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in disabledClients" :key="`dis-${idx}`" class="border-b">
              <td class="px-3 py-2">{{ clientInboundId(item) || "-" }}</td>
              <td class="px-3 py-2">{{ clientEmail(item) || "-" }}</td>
              <td class="px-3 py-2 font-mono text-xs">{{ JSON.stringify(item) }}</td>
              <td class="px-3 py-2">
                <UiButton
                  :disabled="!clientInboundId(item) || !clientEmail(item)"
                  size="sm"
                  variant="outline"
                  @click="deleteClientByEmail(clientInboundId(item), clientEmail(item))"
                >
                  {{ t("delete", "Delete") }}
                </UiButton>
              </td>
            </tr>
            <tr v-if="disabledClients.length === 0">
              <td class="px-3 py-6 text-center text-muted-foreground" colspan="4">{{ t("none", "No disabled clients") }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <div v-if="confirmModalOpen" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 p-4">
    <div class="w-[min(96vw,560px)] max-w-none rounded-xl border bg-card p-4 shadow-lg">
      <h3 class="text-base font-semibold">{{ confirmModalTitle }}</h3>
      <p class="mt-2 text-sm text-muted-foreground">{{ confirmModalMessage }}</p>
      <div class="mt-4 flex justify-end gap-2">
        <UiButton size="sm" variant="outline" @click="resolveConfirmModal(false)">{{ t("cancel", "Cancel") }}</UiButton>
        <UiButton size="sm" @click="resolveConfirmModal(true)">{{ t("confirm", "Confirm") }}</UiButton>
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
</template>
