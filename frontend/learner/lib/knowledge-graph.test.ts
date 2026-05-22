import { describe, expect, it } from "vitest"

import {
  buildKnowledgePointMasteryMap,
  colorForKnowledgePointNode,
  type LearnerKnowledgeGraphNode,
  type KnowledgePointScore,
} from "./knowledge-graph"

describe("knowledge-graph helpers", () => {
  it("prefers node mastery and falls back to weak/strong point scores", () => {
    const nodes: LearnerKnowledgeGraphNode[] = [
      {
        id: "kp-1",
        type: "knowledge_point",
        ref_id: "kp-1",
        label: "Point 1",
        attempts: 3,
        mastery: 33,
      },
      {
        id: "kp-2",
        type: "knowledge_point",
        ref_id: "kp-2",
        label: "Point 2",
        attempts: 2,
      },
    ]

    const weakPoints: KnowledgePointScore[] = [
      { knowledge_point_id: "kp-2", knowledge_point_name: "Point 2", attempts: 5, correct_count: 2, accuracy: 40 },
    ]

    const masteryMap = buildKnowledgePointMasteryMap(nodes, weakPoints, [])

    expect(masteryMap.get("kp-1")).toBe(33)
    expect(masteryMap.get("kp-2")).toBe(40)
    expect(colorForKnowledgePointNode(nodes[0], masteryMap)).toBe("#ef4444")
    expect(colorForKnowledgePointNode(nodes[1], masteryMap)).toBe("#ef4444")
  })

  it("colors mastered points green", () => {
    const node: LearnerKnowledgeGraphNode = {
      id: "kp-3",
      type: "knowledge_point",
      ref_id: "kp-3",
      label: "Point 3",
      attempts: 4,
    }

    const masteryMap = new Map<string, number>([["kp-3", 87]])
    expect(colorForKnowledgePointNode(node, masteryMap)).toBe("#22c55e")
  })
})
