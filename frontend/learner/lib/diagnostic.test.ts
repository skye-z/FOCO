import { describe, expect, it } from "vitest"

import {
  buildDiagnosticSummaryText,
  shouldShowVolatilityAlert,
} from "./diagnostic"

describe("diagnostic helpers", () => {
  it("shows volatility alert when recent ten practice accuracies swing by at least 20 points", () => {
    expect(shouldShowVolatilityAlert([88, 82, 84, 67])).toBe(true)
    expect(shouldShowVolatilityAlert([88, 82, 84, 72])).toBe(false)
  })

  it("builds readable diagnostic summary copy", () => {
    expect(
      buildDiagnosticSummaryText({
        summary_text: "固定收益和收益率曲线需要优先补强。",
        recommended_difficulty: "medium",
      }),
    ).toContain("建议难度：中等")
  })
})
