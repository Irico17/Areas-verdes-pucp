import { cpSync, createReadStream, mkdirSync } from "node:fs"
import path from "node:path"
import react from "@vitejs/plugin-react"
import { defineConfig, type Plugin } from "vite"
import { VitePWA } from "vite-plugin-pwa"

function maplibreWorker(): Plugin {
  const files = ["maplibre-gl-worker.mjs", "maplibre-gl-shared.mjs"]
  let root = ""
  const copyInto = (dir: string) => {
    mkdirSync(dir, { recursive: true })
    for (const name of files) {
      cpSync(path.join(root, "node_modules/maplibre-gl/dist", name), path.join(dir, name))
    }
  }
  return {
    name: "maplibre-worker",
    // Vite fotografía public/ al crear el servidor, antes de buildStart.
    // Si el worker no está en esa lista, /maplibre-gl-worker.mjs responde
    // index.html, el worker no arranca y MapLibre no dibuja ninguna geometría.
    configResolved(config) {
      root = config.root
      if (config.command === "serve") copyInto(path.join(root, "public"))
    },
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const url = (req.url ?? "").split("?")[0]
        const name = files.find((file) => url === `/${file}`)
        if (!name) {
          next()
          return
        }
        res.setHeader("Content-Type", "application/javascript; charset=utf-8")
        const stream = createReadStream(path.join(server.config.root, "public", name))
        stream.on("error", () => next())
        stream.pipe(res)
      })
    },
    buildStart() {
      copyInto(path.join(root, "public"))
    },
    closeBundle() {
      copyInto(path.join(root, "dist"))
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
      "/areas-verdes": { target: process.env.VITE_DEV_API ?? "http://127.0.0.1:8091", changeOrigin: true },
      "/api": { target: process.env.VITE_DEV_API ?? "http://127.0.0.1:8091", changeOrigin: true },
      "/health": { target: process.env.VITE_DEV_API ?? "http://127.0.0.1:8091", changeOrigin: true },
    },
  },
  preview: {
    host: "127.0.0.1",
    port: 4317,
    strictPort: true,
  },
})
