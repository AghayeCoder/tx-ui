# TX-UI Panel v2 frontend

Vue 3 + Tailwind + shadcn-vue migration workspace.

## Commands

```bash
npm install
npm run dev
npm run build
```

`npm run build` emits production assets into `internal/web/assets/panel-v2`.

## Route

The Go server serves the v2 app as the main panel routes:
- `/panel/`
- `/panel/inbounds`
- `/panel/settings`
- `/panel/xray`

Legacy Ant/Vue panel templates are removed from the active runtime.
