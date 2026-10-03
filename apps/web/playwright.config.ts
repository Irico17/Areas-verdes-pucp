import { defineConfig } from "@playwright/test"

/**
 * La web sale de `npm run dev` (Vite en 127.0.0.1:4317).
 * VITE_DEV_API es el destino del proxy `/areas-verdes` (por defecto http://127.0.0.1:8091).
 * E2E_BASE_URL cambia la base de Playwright si la web no está en el puerto de desarrollo.
 */
const baseURL = process.env.E2E_BASE_URL ?? "http://127.0.0.1:4317"
const api = process.env.VITE_DEV_API ?? "http://127.0.0.1:8091"

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  expect: { timeout: 15_000 },
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  // La suite comparte una sola base. Un worker evita que un spec pise a otro.
  workers: 1,
  reporter: [
    ["list"],
    ["html", { open: "never", outputFolder: "playwright-report" }],
  ],
  use: {
    baseURL,
    viewport: { width: 1280, height: 900 },
    trace: "retain-on-failure",
  },
  webServer: {
    command: "npm run dev",
    url: baseURL,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    env: {
      VITE_DEV_API: api,
    },
  },
})
