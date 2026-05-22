"use client";

import type { StepSchema } from "../step-list";

export function OrderingMatchingPreview({ step }: { step: StepSchema }) {
  const content = step.content ?? {};
  const items: Array<{ id: string; text: string }> =
    Array.isArray(content.items)
      ? content.items.map((item: any, index: number) => ({
          id: item?.id ?? String(index),
          text:
            typeof item === "string"
              ? item
              : item?.text ?? item?.label ?? String(item ?? ""),
        }))
      : [];
  const pairs: Array<{ left: string; right: string }> = Array.isArray(
    content.pairs,
  )
    ? content.pairs
    : [];

  if (items.length === 0 && pairs.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        在 content 中配置 items 或 pairs
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {items.length > 0 && (
        <div className="space-y-1.5">
          <p className="text-xs font-medium text-muted-foreground">排序项</p>
          {items.map((item, i) => (
            <div
              key={item.id ?? i}
              className="rounded-md border bg-muted/30 px-3 py-2 text-sm"
            >
              {item.text ?? `项 ${i + 1}`}
            </div>
          ))}
        </div>
      )}
      {pairs.length > 0 && (
        <div className="space-y-1.5">
          <p className="text-xs font-medium text-muted-foreground">匹配对</p>
          {pairs.map((pair, i) => (
            <div
              key={i}
              className="flex items-center gap-2 rounded-md border bg-muted/30 px-3 py-2 text-sm"
            >
              <span>{pair.left}</span>
              <span className="text-muted-foreground">↔</span>
              <span>{pair.right}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
