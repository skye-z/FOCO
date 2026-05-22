import path from "node:path"

export type E2EConfig = {
  adminUrl: string
  learnerUrl: string
  apiUrl: string
  adminEmail: string
  adminPassword: string
  contentPackagePath: string
  resetCommand: string
  requireReset: boolean
}

function boolEnv(name: string, fallback = false) {
  const value = process.env[name]
  if (value == null || value === "") return fallback
  return ["1", "true", "yes", "on"].includes(value.toLowerCase())
}

export const e2eConfig: E2EConfig = {
  adminUrl: process.env.E2E_ADMIN_URL ?? "http://localhost:3001",
  learnerUrl: process.env.E2E_LEARNER_URL ?? "http://localhost:3000",
  apiUrl: process.env.E2E_API_URL ?? "http://localhost:8080",
  adminEmail: process.env.E2E_ADMIN_EMAIL ?? "skai-zhang@hotmail.com",
  adminPassword: process.env.E2E_ADMIN_PASSWORD ?? "DevAdmin@2026",
  contentPackagePath: path.resolve(
    process.cwd(),
    process.env.E2E_CONTENT_PACKAGE ?? "../cfa.json",
  ),
  resetCommand: process.env.E2E_RESET_COMMAND ?? "",
  requireReset: boolEnv("E2E_REQUIRE_RESET", false),
}
