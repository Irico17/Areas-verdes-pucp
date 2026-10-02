import { defineConfig } from "@playwright/test"

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  expect: { timeout: 15_000 },
  use: {
    baseURL: "http://127.0.0.1:4317",
    viewport: { width: 1280, height: 900 },
  },
  reporter: [["list"]],
})
