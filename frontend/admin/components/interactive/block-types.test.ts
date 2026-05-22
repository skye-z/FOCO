import { describe, expect, it } from "vitest";

import {
  VISUAL_BLOCKS,
  createDefaultVisualStep,
  evaluateBlockCompleteness,
  getBlockDefinition,
  type VisualBlockType,
} from "./block-types";

describe("block-types", () => {
  const REQUIRED_BLOCKS: VisualBlockType[] = [
    "formula_drag",
    "parameter_chart",
    "reasoning_sort",
    "concept_match",
    "condition_mark",
    "choice",
    "fill_blank",
  ];

  it("registers every required visual block", () => {
    expect(VISUAL_BLOCKS).toHaveLength(REQUIRED_BLOCKS.length);
  });

  it.each(REQUIRED_BLOCKS)("defines %s with a complete scaffold", (blockType) => {
    const definition = getBlockDefinition(blockType);
    expect(definition).toBeTruthy();
    expect(definition?.label).toBeTruthy();
    expect(definition?.description).toBeTruthy();

    const step = createDefaultVisualStep(blockType);
    expect(step.widget_type).toBeTruthy();
    expect(step.content).toBeTruthy();
    expect(step.initial_state).toBeTruthy();
    expect(step.allowed_actions).toBeTruthy();
    expect(step.evaluation_config).toBeTruthy();
    expect(step.feedback_map).toBeTruthy();
    expect(step.hint_policy).toBeTruthy();

    const completeness = evaluateBlockCompleteness(step);
    expect(completeness.complete).toBe(false);
    expect(completeness.missing.length).toBeGreaterThan(0);
  });
});
