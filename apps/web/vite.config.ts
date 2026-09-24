import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"
import { VitePWA } from "vite-plugin-pwa"

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["favicon.svg"],
      manifest: {
        name: "Campus Verde",
        short_name: "Campus Verde",
        description: "Supervisión de áreas verdes del campus PUCP Pando",
        lang: "es-PE",
        start_url: "/",
        display: "standalone",
        background_color: "#e4ebe4",
        theme_color: "#0e3b2c",
        icons: [{ src: "favicon.svg", sizes: "any", type: "image/svg+xml" }],
      },
      devOptions: { enabled: false },
      workbox: {
        navigateFallback: "index.html",
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/tile\.openstreetmap\.org\/.*/i,
            handler: "CacheFirst",
            options: {
              cacheName: "osm-tiles",
              expiration: { maxEntries: 400, maxAgeSeconds: 60 * 60 * 24 * 14 },
            },
          },
        ],
      },
    }),
  ],
  server: {
    host: "127.0.0.1",
    port: 4317,
    strictPort: true,
    proxy: {
      "/api": { target: "http://127.0.0.1:8091", changeOrigin: true },
      "/health": { target: "http://127.0.0.1:8091", changeOrigin: true },
    },
  },
  preview: {
    host: "127.0.0.1",
    port: 4317,
    strictPort: true,
  },
})
