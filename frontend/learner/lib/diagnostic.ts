export function shouldShowVolatilityAlert(accuracies: number[]) {
  if (accuracies.length < 2) {
    return false
  }

  let min = Number.POSITIVE_INFINITY
  let max = Number.NEGATIVE_INFINITY
  for (const accuracy of accuracies.slice(0, 10)) {
    if (accuracy < min) min = accuracy
    if (accuracy > max) max = accuracy
  }

  return max - min >= 20
}

export function buildDiagnosticSummaryText(summary: {
  summary_text: string
  recommended_difficulty: "easy" | "medium" | "hard" | string
}) {
  const labelMap: Record<string, string> = {
    easy: "简单",
    medium: "中等",
    hard: "困难",
  }

  return `${summary.summary_text} 建议难度：${labelMap[summary.recommended_difficulty] ?? summary.recommended_difficulty}。`
}
