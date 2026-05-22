import {
  VISUAL_BLOCKS,
  createDefaultVisualStep,
  evaluateBlockCompleteness,
  getBlockDefinition,
  type VisualBlockType,
} from "./block-types";

const REQUIRED_BLOCKS: VisualBlockType[] = [
  "formula_drag",
  "parameter_chart",
  "reasoning_sort",
  "concept_match",
  "condition_mark",
  "choice",
  "fill_blank",
];

for (const blockType of REQUIRED_BLOCKS) {
  const definition = getBlockDefinition(blockType);
  if (!definition) {
    throw new Error(`Missing block definition for ${blockType}`);
  }
  if (!definition.label || !definition.description) {
    throw new Error(`Block ${blockType} must have label and description`);
  }

  const step = createDefaultVisualStep(blockType);
  if (!step.widget_type) {
    throw new Error(`Block ${blockType} must map to widget_type`);
  }
  if (!step.content || !step.initial_state || !step.allowed_actions) {
    throw new Error(`Block ${blockType} must initialize config objects`);
  }
  if (!step.evaluation_config || !step.feedback_map || !step.hint_policy) {
    throw new Error(`Block ${blockType} must initialize evaluation objects`);
  }

  const completeness = evaluateBlockCompleteness(step);
  if (completeness.complete) {
    throw new Error(`Default block ${blockType} should start incomplete`);
  }
  if (completeness.missing.length === 0) {
    throw new Error(`Default block ${blockType} should report missing fields`);
  }
}

if (VISUAL_BLOCKS.length !== REQUIRED_BLOCKS.length) {
  throw new Error("Visual block registry must contain all required blocks");
}
