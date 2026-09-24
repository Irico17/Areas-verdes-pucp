import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"
import { VitePWA } from "vite-plugin-pwa"

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["favicon.svg", "icon-192.png", "icon-512.png"],
      manifest: {
        name: "Campus Verde",
        short_name: "Campus Verde",
        description: "Supervisión de áreas verdes del campus PUCP Pando",
        lang: "es",
        start_url: "/",
        display: "standalone",
        background_color: "#e7ece6",
        theme_color: "#145c3e",
        icons: [
          { src: "icon-192.png", sizes: "192x192", type: "image/png" },
          { src: "icon-512.png", sizes: "512x512", type: "image/png" },
        ],
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
