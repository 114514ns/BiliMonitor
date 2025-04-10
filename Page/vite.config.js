import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
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