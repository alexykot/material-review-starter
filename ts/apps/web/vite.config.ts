import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// WEB_PORT and API_PORT let several checkouts run side by side: tools/wt
// gives each worktree environment its own pair. The defaults are the primary
// checkout's ports.
const webPort = Number(process.env.WEB_PORT ?? 5180);
const apiPort = Number(process.env.API_PORT ?? 8787);

export default defineConfig({
  plugins: [react()],
  server: {
    port: webPort,
    strictPort: true,
    proxy: {
      "/api": `http://localhost:${apiPort}`,
    },
  },
});
