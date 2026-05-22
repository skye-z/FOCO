export type LearnerKnowledgeGraphNode = {
  id: string
  type: string
  ref_id: string
  label: string
  description?: string
  group?: string
  mastery?: number
  attempts: number
}

export type KnowledgePointScore = {
  knowledge_point_id: string
  knowledge_point_name: string
  attempts: number
  correct_count: number
  accuracy: number
}

export function buildKnowledgePointMasteryMap(
  nodes: LearnerKnowledgeGraphNode[],
  weakPoints: KnowledgePointScore[],
  strongPoints: KnowledgePointScore[],
) {
  const mastery = new Map<string, number>()

  for (const node of nodes) {
    if (node.type !== "knowledge_point" || node.mastery == null) continue
    mastery.set(node.ref_id, node.mastery)
  }

  for (const point of [...weakPoints, ...strongPoints]) {
    if (!mastery.has(point.knowledge_point_id)) {
      mastery.set(point.knowledge_point_id, point.accuracy)
    }
  }

  return mastery
}

export function colorForKnowledgePointNode(
  node: Pick<LearnerKnowledgeGraphNode, "type" | "ref_id" | "mastery">,
  masteryMap: Map<string, number>,
) {
  if (node.type === "exam") return "#f59e0b"
  if (node.type === "subject") return "#7c3aed"
  if (node.type === "chapter") return "#0f766e"

  const mastery = node.mastery ?? masteryMap.get(node.ref_id)
  if (mastery == null) return "#94a3b8"
  if (mastery < 50) return "#ef4444"
  if (mastery < 80) return "#f59e0b"
  return "#22c55e"
}
