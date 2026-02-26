import { createApp } from "vue";
import App from "./App.vue";
import { router } from "./router";
import "./style.css";
import { applyLocaleDirection, loadI18n } from "./lib/i18n";

applyLocaleDirection();

void loadI18n([
  "menu.dashboard",
  "menu.inbounds",
  "menu.settings",
  "menu.xray",
  "menu.logout",
  "menu.link",
  "pages.index.title",
  "pages.inbounds.title",
  "pages.settings.title",
  "pages.xray.title",
  "pages.settings.actions",
  "pages.settings.panelSettings",
  "pages.inbounds.operate",
  "pages.inbounds.remark",
  "pages.inbounds.protocol",
  "pages.inbounds.port",
  "pages.inbounds.traffic",
  "pages.inbounds.expireDate",
  "pages.inbounds.details",
  "pages.inbounds.export",
  "pages.inbounds.addInbound",
  "pages.inbounds.generalActions",
  "pages.inbounds.resetTraffic",
  "pages.inbounds.resetAllTraffic",
  "pages.inbounds.resetAllClientTraffics",
  "pages.inbounds.delDepletedClients",
  "pages.inbounds.enable",
  "pages.inbounds.clients",
  "username",
  "password",
  "login",
  "loading",
  "close",
  "copy",
  "edit",
  "delete",
  "enabled",
  "disabled",
  "online",
  "depleted",
  "moveUp",
  "moveDown",
  "search",
  "status",
  "comment",
  "success",
  "fail",
  "default",
  "unlimited",
  "none",
  "info",
  "qrCode"
]);

createApp(App).use(router).mount("#app");
