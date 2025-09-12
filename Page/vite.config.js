import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { visualizer } from 'rollup-plugin-visualizer';

// https://vitejs.dev/config/
const ReactCompilerConfig = { /* ... */ };
export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: [
          ["babel-plugin-react-compiler", ReactCompilerConfig],
        ],
      },
    }),    visualizer({
      gzipSize: true,
      brotliSize: true,
      emitFile: false,
      filename: "stat.html",
    }),
  ],
  server: {
    proxy: {

      "/api": {
        target: "http://127.0.0.1:8081",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },

    },
    port:5174,
    allowedHosts:['live-dev.ikun.dev']
  },
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {

        manualChunks: undefined,
        inlineDynamicImports: true


        /*
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
        },
      },external: ['react', 'react-dom']

         */
    }
  }},
  esbuild: {
    sourcemap: true,
  },
});