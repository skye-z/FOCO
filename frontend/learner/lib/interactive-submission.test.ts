import { describe, expect, it } from "vitest"

import {
  buildFormulaSubmission,
  buildHighlightSubmission,
  selectedTextsFromSegments,
  type HighlightSegment,
} from "./interactive-submission"

describe("interactive submission helpers", () => {
  it("extracts grouped selected highlight texts from character segments", () => {
    const segments: HighlightSegment[] = [
      { id: "char-0", text: "接", selectable: true },
      { id: "char-1", text: "受", selectable: true },
      { id: "char-2", text: "。", selectable: false },
      { id: "char-3", text: "礼", selectable: true },
      { id: "char-4", text: "物", selectable: true },
    ]

    expect(selectedTextsFromSegments(segments, ["char-0", "char-1", "char-3", "char-4"])).toEqual([
      "接受",
      "礼物",
    ])
  })

  it("builds highlight payloads with both ids and grouped texts", () => {
    const segments: HighlightSegment[] = [
      { id: "char-0", text: "减", selectable: true },
      { id: "char-1", text: "值", selectable: true },
      { id: "char-2", text: "损", selectable: true },
      { id: "char-3", text: "失", selectable: true },
    ]

    expect(buildHighlightSubmission(segments, ["char-0", "char-1", "char-2", "char-3"])).toEqual({
      marked_ids: ["char-0", "char-1", "char-2", "char-3"],
      marked_texts: ["减值损失"],
    })
  })

  it("includes formula_text in the formula submission payload", () => {
    expect(
      buildFormulaSubmission("(1+r/m)^n - 1", {
        left: "(1+r/m)^n",
        right: "-1",
      }),
    ).toEqual({
      formula_text: "(1+r/m)^n - 1",
      slot_values: {
        left: "(1+r/m)^n",
        right: "-1",
      },
    })
  })
})
