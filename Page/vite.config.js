import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {

      "/api": {
        target: "http://127.0.0.1:8084",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },

    },
    port:5173
  },
  build: {
    sourcemap: true,
  },
  esbuild: {
    sourcemap: true,
  },
});