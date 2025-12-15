import {defineConfig} from "vite";
import react from "@vitejs/plugin-react";
import {visualizer} from 'rollup-plugin-visualizer';
// https://vitejs.dev/config/
const ReactCompilerConfig = { /* ... */};
export default defineConfig({
    plugins: [
        react({
            babel: {
                plugins: [
                    ["babel-plugin-react-compiler", ReactCompilerConfig],
                ],
            },
        }), visualizer({
            gzipSize: true,
            brotliSize: true,
            emitFile: false,
            filename: "stat.html",
        })
    ],
    server: {
        proxy: {

            "/api": {
                target: "http://127.0.0.1:8081",
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/api/, ''),
                configure: (proxy, options) => {
                    proxy.on('proxyRes', (proxyRes) => {
                        // 不要让 Vite 加 no-cache
                        delete proxyRes.headers['cache-control']
                        delete proxyRes.headers['pragma'];
                        proxyRes.headers['cache-control'] = 'public, max-age=0, stale-while-revalidate=30'
                    });
                }
            },
            "/api/status": {
                target: "ws://127.0.0.1:8081",
                ws: true,
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/api/, '')
            }

        },
        port: 5174,
        allowedHosts: ['live-dev.ikun.dev','8d3caf.ikun.dev']
    },
    build: {
        sourcemap: true,
        rollupOptions: {
            output: {
                manualChunks(id) {
                    // 按依赖路径分组
                    if (id.includes('node_modules')) {
                        if (id.includes('react') || id.includes('heroui') || id.includes('rechart')) {
                            return 'react-vendor'
                        }
                    }
                },
            }



        }
    },
    esbuild: {
        sourcemap: true,
    },
});