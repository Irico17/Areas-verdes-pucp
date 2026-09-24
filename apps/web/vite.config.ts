import { cpSync } from "node:fs"
import react from "@vitejs/plugin-react"
import { defineConfig, type Plugin } from "vite"
import { VitePWA } from "vite-plugin-pwa"

function maplibreWorker(): Plugin {
  const files = ["maplibre-gl-worker.mjs", "maplibre-gl-shared.mjs"]
  const copy = (dir: string) => {
    for (const name of files) {
      cpSync(`node_modules/maplibre-gl/dist/${name}`, `${dir}/${name}`)
    }
  }
  return {
    name: "maplibre-worker",
    buildStart() {
      copy("public")
    },
    closeBundle() {
      copy("dist")
    },
  }
}

export default defineConfig({
  plugins: [
    react(),
    maplibreWorker(),
    VitePWA({
      registerType: "autoUpdate",
      includeAssets: ["favicon.svg", "apple-touch-icon.png"],
      manifest: {
        name: "VerdePUCP",
        short_name: "VerdePUCP",
        description: "Gestión de Áreas Verdes",
        lang: "es-PE",
        start_url: "/",
        display: "standalone",
        background_color: "#e7eeeb",
        theme_color: "#083465",
        icons: [
          { src: "favicon.svg", sizes: "any", type: "image/svg+xml", purpose: "any" },
          { src: "icon-192.png", sizes: "192x192", type: "image/png", purpose: "any" },
          { src: "icon-512.png", sizes: "512x512", type: "image/png", purpose: "any" },
          { src: "icon-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
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
