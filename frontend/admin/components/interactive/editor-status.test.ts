import { describe, expect, it } from "vitest";

import { createDefaultVisualStep } from "./block-types";
import { hasUnsavedChanges, summarizePublishReadiness } from "./editor-status";

describe("editor-status", () => {
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

  it("prevents publishing when title or required block fields are missing", () => {
    const blankReadiness = summarizePublishReadiness([], "");
    expect(blankReadiness.canPublish).toBe(false);
    expect(blankReadiness.missing).toContain("单元标题");

    const incompleteReadiness = summarizePublishReadiness([incompleteStep], "Draft");
    expect(incompleteReadiness.canPublish).toBe(false);
    expect(incompleteReadiness.incompleteSteps).toHaveLength(1);
  });

  it("allows publishing when the editor state is complete", () => {
    const completeReadiness = summarizePublishReadiness([completeStep], "Unit title");
    expect(completeReadiness.canPublish).toBe(true);
  });

  it("tracks unsaved changes from title updates", () => {
    expect(hasUnsavedChanges("Title", [completeStep], "Title", [completeStep])).toBe(false);
    expect(hasUnsavedChanges("Title", [completeStep], "New Title", [completeStep])).toBe(true);
  });
});
