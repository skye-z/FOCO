"use client";

import type { StepSchema } from "../step-list";

export function FormulaBuilderPreview({ step }: { step: StepSchema }) {
  const content = step.content ?? {};
  const formulaTemplate: string = content.formula_template ?? content.answer ?? content.formula ?? "";
  const slots: Array<{ label: string; value: string }> = Array.isArray(
    content.slots,
  )
    ? content.slots
    : [];

  if (slots.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        在 content.slots 中配置公式槽位
      </p>
    );
  }

  return (
    <div className="space-y-2">
      {formulaTemplate ? (
        <div className="rounded-md border bg-background px-3 py-2 text-sm font-mono">
          {formulaTemplate}
        </div>
      ) : null}
      <p className="text-xs font-medium text-muted-foreground">公式槽位</p>
      <div className="flex flex-wrap items-center gap-2 rounded-md border bg-muted/30 p-3">
        {slots.map((slot, i) => (
          <div key={i} className="flex items-center gap-1">
            <span className="text-xs text-muted-foreground">
              {slot.label ?? `槽${i + 1}`}:
            </span>
            <span className="rounded border bg-background px-2 py-0.5 text-sm font-mono">
              {slot.value ?? "___"}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
