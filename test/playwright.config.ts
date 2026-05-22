import { defineConfig, devices } from "@playwright/test"

const headed = process.env.E2E_HEADLESS !== "1"
const slowMo = Number.parseInt(process.env.E2E_SLOW_MO_MS ?? "60", 10)

export default defineConfig({
  testDir: "./e2e",
  outputDir: "./artifacts/test-results",
  timeout: 30 * 60 * 1000,
  expect: {
    timeout: 20 * 1000,
  },
  fullyParallel: false,
  workers: 1,
  retries: Number.parseInt(process.env.E2E_RETRIES ?? "0", 10),
  reporter: [
    ["list"],
    ["html", { outputFolder: "artifacts/html-report", open: "never" }],
    ["json", { outputFile: "artifacts/results.json" }],
  ],
  use: {
    ...devices["Desktop Chrome"],
    headless: !headed,
    actionTimeout: 20 * 1000,
    navigationTimeout: 45 * 1000,
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
    video: "on",
    launchOptions: {
      slowMo: Number.isFinite(slowMo) ? slowMo : 60,
    },
  },
  projects: [
    {
      name: "edge",
      use: {
        channel: process.env.E2E_BROWSER_CHANNEL ?? "msedge",
        viewport: { width: 1440, height: 1000 },
      },
    },
  ],
})
