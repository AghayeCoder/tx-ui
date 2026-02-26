import { createRouter, createWebHistory } from "vue-router";
import DashboardPage from "@/pages/DashboardPage.vue";
import InboundsPage from "@/pages/InboundsPage.vue";
import SettingsPage from "@/pages/SettingsPage.vue";
import XrayPage from "@/pages/XrayPage.vue";

const basePath = (window as Window & { __TXUI_BASE_PATH__?: string }).__TXUI_BASE_PATH__ || "/";
const normalizedBase = basePath.endsWith("/") ? basePath : `${basePath}/`;
const routerBase = `${normalizedBase}panel/`;

export const router = createRouter({
  history: createWebHistory(routerBase),
  routes: [
    { path: "/", name: "dashboard", component: DashboardPage, meta: { title: "Dashboard" } },
    { path: "/inbounds", name: "inbounds", component: InboundsPage, meta: { title: "Inbounds" } },
    { path: "/settings", name: "settings", component: SettingsPage, meta: { title: "Settings" } },
    { path: "/xray", name: "xray", component: XrayPage, meta: { title: "Xray" } },
    { path: "/:pathMatch(.*)*", redirect: "/" }
  ]
});
