export type HighlightSegment = {
  id: string
  text: string
  selectable: boolean
}

export function selectedTextsFromSegments(segments: HighlightSegment[], markedIds: string[]) {
  const selected = new Set(markedIds)
  const groups: string[] = []
  let buffer = ""

  for (const segment of segments) {
    if (segment.selectable && selected.has(segment.id)) {
      buffer += segment.text
      continue
    }
    if (!segment.selectable && /\s/.test(segment.text) && buffer) {
      buffer += segment.text
      continue
    }
    if (buffer.trim()) groups.push(buffer.trim())
    buffer = ""
  }

  if (buffer.trim()) groups.push(buffer.trim())

  return groups
}

export function buildHighlightSubmission(segments: HighlightSegment[], markedIds: string[]) {
  return {
    marked_ids: markedIds,
    marked_texts: selectedTextsFromSegments(segments, markedIds),
  }
}

export function buildFormulaSubmission(formulaText: string, slotValues: Record<string, string>) {
  return {
    formula_text: formulaText.trim(),
    slot_values: slotValues,
  }
}
