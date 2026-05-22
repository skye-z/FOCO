import { execSync } from "node:child_process"
import { e2eConfig } from "./config"

async function sleep(ms: number) {
  await new Promise((resolve) => setTimeout(resolve, ms))
}

export async function waitForBackendHealth() {
  const healthUrl = new URL("/api/v1/health", e2eConfig.apiUrl).toString()
  const deadline = Date.now() + 60_000
  let lastError = ""

  while (Date.now() < deadline) {
    try {
      const response = await fetch(healthUrl)
      if (response.ok) return
      lastError = `${response.status} ${response.statusText}`
    } catch (error) {
      lastError = error instanceof Error ? error.message : String(error)
    }
    await sleep(1000)
  }

  throw new Error(`Backend health check timed out: ${healthUrl}; last error: ${lastError}`)
}

export async function seedDefaultAdmin() {
  const seedUrl = new URL("/api/v1/seed/admin", e2eConfig.apiUrl).toString()
  const response = await fetch(seedUrl, { method: "POST" })
  if (!response.ok) {
    throw new Error(`Seed default admin failed: ${response.status} ${await response.text()}`)
  }
}

export function resetDataIfConfigured() {
  if (!e2eConfig.resetCommand) {
    if (e2eConfig.requireReset) {
      throw new Error("E2E_REQUIRE_RESET is enabled, but E2E_RESET_COMMAND is empty.")
    }
    console.warn("E2E_RESET_COMMAND is empty; data reset is skipped.")
    return
  }

  execSync(e2eConfig.resetCommand, {
    cwd: process.cwd(),
    stdio: "inherit",
    shell:
      process.platform === "win32"
        ? process.env.ComSpec ?? "cmd.exe"
        : process.env.SHELL ?? "/bin/sh",
    env: process.env,
  })
}
