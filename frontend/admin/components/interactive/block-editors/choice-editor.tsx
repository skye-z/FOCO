"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

type ChoiceOption = { id: string; text: string; correct?: boolean };

function normalizeOptions(step: StepSchema): ChoiceOption[] {
  const raw = Array.isArray(step.content?.options) ? step.content.options : [];
  const correctSelections = Array.isArray(step.evaluation_config?.correct_selections)
    ? step.evaluation_config.correct_selections.map((item: any) => String(item))
    : [];
  const correctOptionIds = Array.isArray(step.evaluation_config?.correct_option_ids)
    ? step.evaluation_config.correct_option_ids.map((item: any) => String(item))
    : [];
  const correctOptionId = step.evaluation_config?.correct_option_id
    ? String(step.evaluation_config.correct_option_id)
    : "";

  return raw.map((option: any, index: number) => {
    const text =
      typeof option === "string"
        ? option
        : option?.text ?? option?.label ?? String(option ?? "");
    const id = option?.id ?? String(index);
    const correct =
      Boolean(option?.correct) ||
      correctSelections.includes(text) ||
      correctOptionIds.includes(String(id)) ||
      correctOptionId === String(id);
    return { id: String(id), text, correct };
  });
}

export function ChoiceEditor({
  step,
  onChange,
}: {
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  const prompt =
    (step.content?.prompt as string | undefined) ??
    (step.content?.text as string | undefined) ??
    (step.content?.instruction as string | undefined) ??
    "";
  const options = normalizeOptions(step);

  function updateOptions(nextOptions: ChoiceOption[]) {
    const correctSelections = nextOptions.filter((item) => item.correct).map((item) => item.text);
    onChange({
      ...step,
      content: { ...step.content, options: nextOptions },
      evaluation_config: {
        ...step.evaluation_config,
        correct_answers: correctSelections,
        correct_selections: correctSelections,
        correct_option_ids: nextOptions.filter((item) => item.correct).map((item) => item.id),
        correct_option_id: nextOptions.find((item) => item.correct)?.id ?? "",
        mode:
          step.evaluation_config?.mode ??
          (correctSelections.length > 1 ? "multi_choice" : "single"),
      },
    });
  }

  return (
    <div className="space-y-4">
      <EditorSection title="选择题题干" description="配置题干和可选项。">
        <div className="space-y-2">
          <Label>题干</Label>
          <Textarea
            value={prompt}
            onChange={(event) =>
              onChange({ ...step, content: { ...step.content, prompt: event.target.value } })
            }
            className="min-h-[100px]"
            placeholder="输入选择题题干"
          />
        </div>
      </EditorSection>

      <EditorSection title="选项配置" description="勾选正确答案。">
        <div className="space-y-2">
          {options.map((option, index) => (
            <div key={option.id || index} className="grid grid-cols-[auto_1fr_auto] items-center gap-2">
              <input
                type="checkbox"
                checked={Boolean(option.correct)}
                onChange={(event) => {
                  const next = [...options];
                  next[index] = { ...option, correct: event.target.checked };
                  updateOptions(next);
                }}
              />
              <Input
                value={option.text}
                onChange={(event) => {
                  const next = [...options];
                  next[index] = { ...option, text: event.target.value };
                  updateOptions(next);
                }}
                placeholder={`选项 ${index + 1}`}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => updateOptions(options.filter((_, i) => i !== index))}
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() =>
              updateOptions([
                ...options,
                { id: `option-${options.length + 1}`, text: "", correct: false },
              ])
            }
          >
            <Plus className="mr-1 size-4" />
            添加选项
          </Button>
        </div>
      </EditorSection>
    </div>
  );
}
