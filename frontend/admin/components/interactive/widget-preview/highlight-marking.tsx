"use client";

import type { StepSchema } from "../step-list";

export function HighlightMarkingPreview({ step }: { step: StepSchema }) {
  const content = step.content ?? {};
  const passage: string = content.passage ?? content.text ?? "";
  const items: Array<{ id: string; text: string; markable?: boolean }> =
    Array.isArray(content.items)
      ? content.items.map((item: any, index: number) => ({
          id: item?.id ?? String(index),
          text:
            typeof item === "string"
              ? item
              : item?.text ?? item?.label ?? String(item ?? ""),
          markable: item?.markable,
        }))
      : Array.isArray(step.evaluation_config?.expected_highlights)
        ? step.evaluation_config.expected_highlights.map((text: any, index: number) => ({
            id: String(index),
            text: String(text ?? ""),
            markable: true,
          }))
        : [];

  if (items.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        在 content.items 中配置标记项
      </p>
    );
  }

  return (
    <div className="space-y-2">
      {passage ? (
        <div className="rounded-md border bg-background px-3 py-2 text-sm leading-relaxed">
          {passage}
        </div>
      ) : null}
      <p className="text-xs font-medium text-muted-foreground">标记项</p>
      <div className="space-y-1.5">
        {items.map((item, i) => (
          <label
            key={item.id ?? i}
            className="flex items-center gap-2 rounded-md border bg-muted/30 px-3 py-2 text-sm"
          >
            <input
              type="checkbox"
              checked={!!item.markable}
              readOnly
              className="size-4 rounded border"
            />
            <span>{item.text ?? `项 ${i + 1}`}</span>
          </label>
        ))}
      </div>
    </div>
  );
}
