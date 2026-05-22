"use client";

import type { StepSchema } from "../step-list";

export function ChoiceClozePreview({ step }: { step: StepSchema }) {
  const content = step.content ?? {};
  const evalConfig = step.evaluation_config ?? {};
  const mode: string = content.mode ?? evalConfig.mode ?? "single";
  const prompt: string = content.prompt ?? content.text ?? content.instruction ?? "";
  const options: Array<{ id: string; text: string }> = Array.isArray(content.options)
    ? content.options.map((option: any, index: number) => ({
        id: option?.id ?? String(index),
        text:
          typeof option === "string"
            ? option
            : option?.text ?? option?.label ?? String(option ?? ""),
      }))
    : [];
  const blanks: Array<{
    id: string;
    text: string;
    options?: string[];
    answer?: string;
  }> = Array.isArray(content.blanks)
    ? content.blanks
    : typeof content.blanks === "number" && content.blanks > 0
      ? Array.from({ length: content.blanks }, (_, index) => ({
          id: String(index),
          text: `空 ${index + 1}`,
        }))
      : [];

  if (blanks.length === 0 && options.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        在可视化编辑器中配置题干、选项或填空项
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {prompt ? <p className="text-sm font-medium">{prompt}</p> : null}
      <p className="text-xs font-medium text-muted-foreground">
        模式: {mode === "single" || mode === "single_choice" ? "单选" : mode === "multi_choice" || mode === "multiple" ? "多选" : "填空"}
      </p>
      {options.length > 0 ? (
        <div className="space-y-1">
          {options.map((option, index) => (
            <label
              key={option.id ?? index}
              className="flex items-center gap-2 rounded-md border bg-muted/30 px-3 py-1.5 text-sm"
            >
              <input
                type={mode === "multiple" ? "checkbox" : "radio"}
                name="visual-choice-preview"
                readOnly
                className="size-3.5"
              />
              <span>{option.text}</span>
            </label>
          ))}
        </div>
      ) : null}
      {blanks.map((blank, i) => (
        <div key={blank.id ?? i} className="space-y-1.5">
          <p className="text-sm font-medium">{blank.text ?? `空 ${i + 1}`}</p>
          {Array.isArray(blank.options) && blank.options.length > 0 ? (
            <div className="space-y-1">
              {blank.options.map((opt, j) => (
                <label
                  key={j}
                  className="flex items-center gap-2 rounded-md border bg-muted/30 px-3 py-1.5 text-sm"
                >
                  <input
                    type={mode === "multiple" ? "checkbox" : "radio"}
                    name={`blank-${i}`}
                    readOnly
                    className="size-3.5"
                  />
                  <span>{opt}</span>
                </label>
              ))}
            </div>
          ) : (
            <div className="rounded-md border bg-muted/30 px-3 py-1.5 text-sm text-muted-foreground">
              {blank.answer ?? "（待填入）"}
            </div>
          )}
        </div>
      ))}
    </div>
  );
}
