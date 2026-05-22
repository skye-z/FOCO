import { evaluateBlockCompleteness } from "./block-types";
import type { StepSchema } from "./step-list";

export type StepReadiness = {
  index: number;
  complete: boolean;
  missing: string[];
};

export type PublishReadiness = {
  canPublish: boolean;
  missing: string[];
  stepStatuses: StepReadiness[];
  incompleteSteps: StepReadiness[];
};

type EditorSnapshot = {
  title: string;
  steps: StepSchema[];
};

export function createEditorSnapshot(title: string, steps: StepSchema[]): EditorSnapshot {
  return {
    title: title.trim(),
    steps: steps.map((step) => ({
      id: step.id ?? "",
      widget_type: step.widget_type,
      content: step.content ?? {},
      initial_state: step.initial_state ?? {},
      allowed_actions: step.allowed_actions ?? {},
      evaluation_config: step.evaluation_config ?? {},
      feedback_map: step.feedback_map ?? {},
      hint_policy: step.hint_policy ?? {},
      knowledge_point_ids: [...(step.knowledge_point_ids ?? [])],
      knowledge_point_tags: [...(step.knowledge_point_tags ?? [])],
    })),
  };
}

export function hasUnsavedChanges(
  initialTitle: string,
  initialSteps: StepSchema[],
  currentTitle: string,
  currentSteps: StepSchema[],
): boolean {
  return (
    JSON.stringify(createEditorSnapshot(initialTitle, initialSteps)) !==
    JSON.stringify(createEditorSnapshot(currentTitle, currentSteps))
  );
}

export function summarizePublishReadiness(
  steps: StepSchema[],
  title: string,
): PublishReadiness {
  const missing: string[] = [];
  if (!title.trim()) missing.push("单元标题");
  if (steps.length === 0) missing.push("至少 1 个交互步骤");

  const stepStatuses = steps.map((step, index) => {
    const completeness = evaluateBlockCompleteness(step);
    return {
      index,
      complete: completeness.complete,
      missing: completeness.missing,
    } satisfies StepReadiness;
  });

  const incompleteSteps = stepStatuses.filter((item) => !item.complete);

  return {
    canPublish: missing.length === 0 && incompleteSteps.length === 0,
    missing,
    stepStatuses,
    incompleteSteps,
  };
}
