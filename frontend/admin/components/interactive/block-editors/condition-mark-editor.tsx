"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

type MarkItem = { text: string };

function normalizeItems(step: StepSchema): MarkItem[] {
  const raw = Array.isArray(step.content?.items)
    ? step.content.items
    : Array.isArray(step.evaluation_config?.expected_highlights)
      ? step.evaluation_config.expected_highlights
      : [];
  return raw.map((item: any) => ({
    text: typeof item === "string" ? item : item?.text ?? item?.label ?? String(item ?? ""),
  }));
}

export function ConditionMarkEditor({
  step,
  onChange,
}: {
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  const passage =
    (step.content?.passage as string | undefined) ??
    (step.content?.text as string | undefined) ??
    (step.content?.instruction as string | undefined) ??
    "";
  const items = normalizeItems(step);

  function updateItems(nextItems: MarkItem[]) {
    onChange({
      ...step,
      content: { ...step.content, items: nextItems },
      evaluation_config: {
        ...step.evaluation_config,
        correct_marks: nextItems.map((item) => item.text),
        expected_highlights: nextItems.map((item) => item.text),
      },
    });
  }

  return (
    <div className="space-y-4">
      <EditorSection title="题干文本" description="输入用户要阅读并标注的文本。">
        <div className="space-y-2">
          <Label>题干</Label>
          <Textarea
            value={passage}
            onChange={(event) =>
              onChange({ ...step, content: { ...step.content, passage: event.target.value } })
            }
            className="min-h-[140px]"
            placeholder="输入需要标记关键条件的题干文本"
          />
        </div>
      </EditorSection>

      <EditorSection title="关键条件" description="配置需要用户标出的关键条件片段。">
        <div className="space-y-2">
          {items.map((item, index) => (
            <div key={index} className="flex items-center gap-2">
              <Input
                value={item.text}
                onChange={(event) => {
                  const next = [...items];
                  next[index] = { text: event.target.value };
                  updateItems(next);
                }}
                placeholder="关键条件文本"
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
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => updateItems([...items, { text: "" }])}
          >
            <Plus className="mr-1 size-4" />
            添加关键条件
          </Button>
        </div>
      </EditorSection>
    </div>
  );
}
