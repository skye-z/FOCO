import { execFileSync } from "node:child_process"

try {
  execFileSync("gofmt", ["-w", "backend"], { stdio: "inherit" })
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error))
  process.exit(1)
}
