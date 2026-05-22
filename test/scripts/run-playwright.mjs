import { spawn } from "node:child_process"
import path from "node:path"
import process from "node:process"

const mode = process.argv[2] ?? "test"
const extraArgs = process.argv.slice(3)
const cliPath = path.resolve(process.cwd(), "node_modules", "@playwright", "test", "cli.js")
const env = { ...process.env }

let cliArgs

switch (mode) {
  case "test":
    cliArgs = ["test", "--project=edge", ...extraArgs]
    break
  case "test-headed":
    env.E2E_HEADLESS = "0"
    cliArgs = ["test", "--project=edge", ...extraArgs]
    break
  case "test-headless":
    env.E2E_HEADLESS = "1"
    cliArgs = ["test", "--project=edge", ...extraArgs]
    break
  case "report":
    cliArgs = ["show-report", "artifacts/html-report", ...extraArgs]
    break
  default:
    console.error(`Unknown mode: ${mode}`)
    process.exit(1)
}

const child = spawn(process.execPath, [cliPath, ...cliArgs], {
  cwd: process.cwd(),
  env,
  stdio: "inherit",
})

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal)
    return
  }
  process.exit(code ?? 1)
})
