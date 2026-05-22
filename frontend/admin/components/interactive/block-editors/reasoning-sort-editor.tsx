"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

function normalizeItems(step: StepSchema): string[] {
  const raw = Array.isArray(step.content?.items) ? step.content.items : [];
  return raw.map((item: any) =>
    typeof item === "string" ? item : item?.text ?? item?.label ?? String(item ?? ""),
  );
}

export function ReasoningSortEditor({
  step,
  onChange,
}: {
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  const items = normalizeItems(step);

  function updateItems(nextItems: string[]) {
    onChange({
      ...step,
      content: { ...step.content, items: nextItems },
      evaluation_config: {
        ...step.evaluation_config,
        correct_order: nextItems,
      },
    });
  }

  return (
    <EditorSection title="推理链配置" description="定义需要用户排序的推理步骤。">
      <div className="space-y-3">
        <div className="space-y-2">
          <Label>推理步骤</Label>
          {items.map((item, index) => (
            <div key={index} className="flex items-center gap-2">
              <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
                {index + 1}
              </span>
              <Input
                value={item}
                onChange={(event) => {
                  const next = [...items];
                  next[index] = event.target.value;
                  updateItems(next);
                }}
                placeholder="输入一条推理步骤"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => updateItems(items.filter((_, i) => i !== index))}
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          ))}
        </div>

        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() =>
            updateItems([...items, ""])
          }
        >
          <Plus className="mr-1 size-4" />
          添加推理步骤
        </Button>
      </div>
    </EditorSection>
  );
}
