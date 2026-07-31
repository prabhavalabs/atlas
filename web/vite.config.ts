import { createReadStream, readFileSync } from "node:fs"
import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import type { Plugin } from "vite"
import { defineConfig } from "vitest/config"

const mapLibreAssets = [
  "maplibre-gl-worker.mjs",
  "maplibre-gl-shared.mjs",
] as const

function mapLibreWorkerAssets(): Plugin {
  const assetDirectory = path.resolve(
    import.meta.dirname,
    "node_modules/maplibre-gl/dist",
  )

  return {
    name: "atlas-maplibre-worker-assets",
    configureServer(server) {
      server.middlewares.use((request, response, next) => {
        const asset = mapLibreAssets.find(
          (name) => request.url === `/assets/${name}`,
        )
        if (!asset) {
          next()
          return
        }
        response.setHeader("Content-Type", "text/javascript; charset=utf-8")
        createReadStream(path.join(assetDirectory, asset)).pipe(response)
      })
    },
    generateBundle() {
      for (const asset of mapLibreAssets) {
        this.emitFile({
          type: "asset",
          fileName: `assets/${asset}`,
          source: readFileSync(path.join(assetDirectory, asset)),
        })
      }
    },
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss(), mapLibreWorkerAssets()],
  build: {
    chunkSizeWarningLimit: 1000,
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    css: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
})
