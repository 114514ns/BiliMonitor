import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {

      "/api": {
        target: "http://127.0.0.1:8081",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },

    },
    port:5173,
    allowedHosts:['live-dev.ikun.dev']
  },
  build: {
    sourcemap: true,
  },
  esbuild: {
    sourcemap: true,
  },
});