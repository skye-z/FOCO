import { execFileSync } from "node:child_process"

try {
  const output = execFileSync("gofmt", ["-l", "backend"], { encoding: "utf8" }).trim()
  if (output) {
    console.error("The following Go files need formatting:")
    console.error(output)
    process.exit(1)
  }
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error))
  process.exit(1)
}
