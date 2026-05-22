const major = Number.parseInt(process.versions.node.split(".")[0], 10)

if (major < 18) {
  console.error(`Node.js 18+ is required. Current runtime: ${process.version}`)
  process.exit(1)
}

console.log(`Node runtime OK: ${process.version}`)
console.log("Install dependencies with: npm install")
console.log("Run Edge E2E with: npm run test:e2e")
