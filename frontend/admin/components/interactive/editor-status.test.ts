import { createDefaultVisualStep } from "./block-types";
import { hasUnsavedChanges, summarizePublishReadiness } from "./editor-status";

const incompleteStep = createDefaultVisualStep("choice");
const completeStep = {
  ...incompleteStep,
  content: {
    ...incompleteStep.content,
    title: "选择正确结论",
    prompt: "Which option is correct?",
    options: [{ id: "a", label: "A" }],
  },
  knowledge_point_ids: ["kp_1"],
};

const blankReadiness = summarizePublishReadiness([], "");
if (blankReadiness.canPublish) {
  throw new Error("blank editor must not be publishable");
}
if (!blankReadiness.missing.includes("单元标题")) {
  throw new Error("blank editor must require title");
}

const incompleteReadiness = summarizePublishReadiness([incompleteStep], "Draft");
if (incompleteReadiness.canPublish) {
  throw new Error("incomplete block must not be publishable");
}
if (incompleteReadiness.incompleteSteps.length !== 1) {
  throw new Error("incomplete block must be reported");
}

const completeReadiness = summarizePublishReadiness([completeStep], "Unit title");
if (!completeReadiness.canPublish) {
  throw new Error("complete editor state must be publishable");
}

if (hasUnsavedChanges("Title", [completeStep], "Title", [completeStep])) {
  throw new Error("identical snapshots must not be dirty");
}
if (!hasUnsavedChanges("Title", [completeStep], "New Title", [completeStep])) {
  throw new Error("title change must mark editor dirty");
}
