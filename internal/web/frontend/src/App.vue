<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { ExternalLink, LayoutDashboard, LogOut, Menu, Monitor, Moon, Network, Settings, Sun, Wrench } from "lucide-vue-next";
import { t, loadI18n } from "@/lib/i18n";

const route = useRoute();
const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";
const panelVersionRaw = (window as Window & { __TXUI_PANEL_VERSION__?: string }).__TXUI_PANEL_VERSION__ || "";
const panelVersion = computed(() => {
  const normalized = String(panelVersionRaw || "").trim();
  return normalized || "latest";
});

const navItems = [
  { to: "/", key: "menu.dashboard", fallback: "Dashboard", icon: LayoutDashboard },
  { to: "/inbounds", key: "menu.inbounds", fallback: "Inbounds", icon: Network },
  { to: "/settings", key: "menu.settings", fallback: "Settings", icon: Settings },
  { to: "/xray", key: "menu.xray", fallback: "Xray", icon: Wrench }
];

const title = computed(() => {
  if (route.name === "dashboard") return t("pages.index.title", "Dashboard");
  if (route.name === "inbounds") return t("pages.inbounds.title", "Inbounds");
  if (route.name === "settings") return t("pages.settings.title", "Settings");
  if (route.name === "xray") return t("pages.xray.title", "Xray");
  return t("pages.index.title", "Panel");
});
const themeMode = ref<"light" | "dark" | "system">("system");
let themeMedia: MediaQueryList | null = null;

function closeClosestDetails(target: EventTarget | null): void {
  const el = target instanceof Element ? target : null;
  if (!el) return;
  const actionEl = el.closest("[data-menu-action='1']");
  if (!actionEl) return;
  const details = actionEl.closest("details");
  if (details instanceof HTMLDetailsElement) {
    details.open = false;
  }
}

function closeAllOpenDetails(): void {
  const openMenus = document.querySelectorAll("details[open]");
  openMenus.forEach((item) => {
    if (item instanceof HTMLDetailsElement) {
      item.open = false;
    }
  });
}

function onGlobalClick(event: MouseEvent): void {
  const el = event.target instanceof Element ? event.target : null;
  if (!el) return;

  closeClosestDetails(el);

  if (el.closest("summary")) {
    return;
  }
  if (el.closest("details")) {
    return;
  }
  closeAllOpenDetails();
}

function onDetailsToggle(event: Event): void {
  const target = event.target;
  if (!(target instanceof HTMLDetailsElement) || !target.open) {
    return;
  }
  const openMenus = document.querySelectorAll("details[open]");
  openMenus.forEach((item) => {
    if (item instanceof HTMLDetailsElement && item !== target) {
      item.open = false;
    }
  });
}

function onGlobalKeydown(event: KeyboardEvent): void {
  if (event.key === "Escape") {
    closeAllOpenDetails();
  }
}

function systemPrefersDark(): boolean {
  return !!(window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches);
}

function resolvedTheme(mode: "light" | "dark" | "system"): "light" | "dark" {
  if (mode === "system") {
    return systemPrefersDark() ? "dark" : "light";
  }
  return mode;
}

function applyTheme(mode: "light" | "dark" | "system"): void {
  themeMode.value = mode;
  const active = resolvedTheme(mode);
  document.documentElement.classList.toggle("dark", active === "dark");
  try {
    localStorage.setItem("panelThemeMode", mode);
  } catch {
    // ignore storage failures
  }
}

function initTheme(): void {
  let mode: "light" | "dark" | "system" = "system";
  try {
    const saved = localStorage.getItem("panelThemeMode");
    if (saved === "light" || saved === "dark" || saved === "system") {
      mode = saved;
    }
  } catch {
    // ignore storage failures
  }
  applyTheme(mode);
}

function cycleTheme(): void {
  const order: Array<"light" | "dark" | "system"> = ["light", "dark", "system"];
  const idx = order.indexOf(themeMode.value);
  const next = order[(idx + 1) % order.length] || "system";
  applyTheme(next);
}

function onSystemThemeChange(): void {
  if (themeMode.value === "system") {
    applyTheme("system");
  }
}

onMounted(() => {
  void loadI18n(["menu.dashboard", "menu.inbounds", "menu.settings", "menu.xray", "menu.logout", "menu.link", "pages.index.title"]);
  initTheme();
  if (window.matchMedia) {
    themeMedia = window.matchMedia("(prefers-color-scheme: dark)");
    if (typeof themeMedia.addEventListener === "function") {
      themeMedia.addEventListener("change", onSystemThemeChange);
    } else if (typeof themeMedia.addListener === "function") {
      themeMedia.addListener(onSystemThemeChange);
    }
  }
  document.addEventListener("click", onGlobalClick);
  document.addEventListener("keydown", onGlobalKeydown);
  document.addEventListener("toggle", onDetailsToggle, true);
});

onUnmounted(() => {
  if (themeMedia) {
    if (typeof themeMedia.removeEventListener === "function") {
      themeMedia.removeEventListener("change", onSystemThemeChange);
    } else if (typeof themeMedia.removeListener === "function") {
      themeMedia.removeListener(onSystemThemeChange);
    }
  }
  document.removeEventListener("click", onGlobalClick);
  document.removeEventListener("keydown", onGlobalKeydown);
  document.removeEventListener("toggle", onDetailsToggle, true);
});
</script>

<template>
  <div class="min-h-screen overflow-x-hidden bg-background text-foreground">
    <div class="mx-auto grid min-h-screen w-full max-w-[1440px] grid-cols-1 lg:grid-cols-[240px_1fr]">
      <aside class="hidden border-r border-border bg-muted/40 p-4 lg:block">
        <div class="mb-4 rounded-lg border border-border bg-card p-4">
          <p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{{ t("pages.settings.title", "Control Panel") }}</p>
          <p class="text-xl font-bold">{{ t("menu.settings", "Panel") }} v{{ panelVersion }}</p>
        </div>
        <nav class="space-y-2">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors"
            :class="route.path === item.to ? 'bg-card border border-border shadow-sm' : 'hover:bg-muted'"
          >
            <component :is="item.icon" class="h-4 w-4" />
            <span>{{ t(item.key, item.fallback) }}</span>
          </RouterLink>
        </nav>
        <div class="mt-4 space-y-2 border-t border-border pt-4">
          <a
            href="https://github.com/AghayeCoder/TX-ThemeHub"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors hover:bg-muted"
          >
            <ExternalLink class="h-4 w-4" />
            <span>{{ t("menu.link", "Manage") }}</span>
          </a>
          <a
            :href="`${basePath}logout`"
            class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors hover:bg-muted"
          >
            <LogOut class="h-4 w-4" />
            <span>{{ t("menu.logout", "Log Out") }}</span>
          </a>
        </div>
      </aside>

      <main class="min-w-0 p-3 sm:p-4 md:p-5">
        <header class="mb-3 rounded-lg border border-border bg-card px-3 py-2.5 sm:mb-4 sm:px-4 sm:py-3">
          <div class="flex items-center justify-between gap-2">
            <div>
              <h1 class="text-lg font-semibold sm:text-xl">{{ title }}</h1>
              <p class="text-xs text-muted-foreground">TX-UI</p>
            </div>
            <div class="flex items-center gap-2">
              <button class="rounded-md border bg-background px-2.5 py-1.5 text-sm hover:bg-muted" type="button" @click="cycleTheme">
                <span class="inline-flex items-center gap-2">
                  <Sun v-if="themeMode === 'light'" class="h-4 w-4" />
                  <Moon v-else-if="themeMode === 'dark'" class="h-4 w-4" />
                  <Monitor v-else class="h-4 w-4" />
                  {{ themeMode === "system" ? t("default", "System") : themeMode === "dark" ? t("dark", "Dark") : t("light", "Light") }}
                </span>
              </button>
              <details class="relative lg:hidden">
                <summary class="action-trigger">
                  <Menu class="h-4 w-4" />
                  {{ t("pages.settings.actions", "Actions") }}
                </summary>
                <div class="menu-inline-end absolute z-20 mt-2 w-56 rounded-xl border border-border bg-card/95 p-2 shadow-2xl backdrop-blur-sm menu-popover">
                  <nav class="space-y-1">
                    <RouterLink
                      v-for="item in navItems"
                      :key="`mobile-${item.to}`"
                      :to="item.to"
                      data-menu-action="1"
                      class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors"
                      :class="route.path === item.to ? 'bg-muted font-medium' : 'hover:bg-muted/70'"
                    >
                      <component :is="item.icon" class="h-4 w-4" />
                      <span>{{ t(item.key, item.fallback) }}</span>
                    </RouterLink>
                  </nav>
                  <div class="mt-2 space-y-1 border-t border-border pt-2">
                    <a
                      href="https://github.com/AghayeCoder/TX-ThemeHub"
                      target="_blank"
                      rel="noopener noreferrer"
                      data-menu-action="1"
                      class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors hover:bg-muted/70"
                    >
                      <ExternalLink class="h-4 w-4" />
                      <span>{{ t("menu.link", "Manage") }}</span>
                    </a>
                    <a
                      :href="`${basePath}logout`"
                      data-menu-action="1"
                      class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors hover:bg-muted/70"
                    >
                      <LogOut class="h-4 w-4" />
                      <span>{{ t("menu.logout", "Log Out") }}</span>
                    </a>
                  </div>
                </div>
              </details>
            </div>
          </div>
        </header>
        <RouterView />
      </main>
    </div>
  </div>
</template>
