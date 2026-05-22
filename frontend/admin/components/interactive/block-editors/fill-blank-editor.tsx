"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

type BlankItem = { id: string; answer: string; hint?: string };

function normalizeBlanks(step: StepSchema): BlankItem[] {
  const raw = Array.isArray(step.content?.blanks)
    ? step.content.blanks
    : Array.isArray(step.evaluation_config?.correct_selections)
      ? step.evaluation_config.correct_selections
      : [];
  return raw.map((item: any, index: number) => ({
    id: item?.id ?? String(index),
    answer:
      typeof item === "string"
        ? item
        : item?.answer ?? item?.text ?? item?.label ?? String(item ?? ""),
    hint: item?.hint,
  }));
}

export function FillBlankEditor({
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
  const blanks = normalizeBlanks(step);

  function updateBlanks(nextBlanks: BlankItem[]) {
    onChange({
      ...step,
      content: { ...step.content, blanks: nextBlanks },
      evaluation_config: {
        ...step.evaluation_config,
        correct_answers: nextBlanks.map((item) => item.answer),
        correct_selections: nextBlanks.map((item) => item.answer),
      },
    });
  }

  return (
    <div className="space-y-4">
      <EditorSection title="填空题题干" description="输入题干和填空位说明。">
        <div className="space-y-2">
          <Label>题干</Label>
          <Textarea
            value={prompt}
            onChange={(event) =>
              onChange({ ...step, content: { ...step.content, prompt: event.target.value } })
            }
            className="min-h-[100px]"
            placeholder="输入填空题题干"
          />
        </div>
      </EditorSection>

      <EditorSection title="填空项" description="配置每个空的正确答案和提示。">
        <div className="space-y-2">
          {blanks.map((blank, index) => (
            <div key={blank.id || index} className="grid grid-cols-2 gap-2 rounded-md border p-3">
              <Input
                value={blank.answer}
                onChange={(event) => {
                  const next = [...blanks];
                  next[index] = { ...blank, answer: event.target.value };
                  updateBlanks(next);
                }}
                placeholder="正确答案"
              />
              <div className="flex gap-2">
                <Input
                  value={blank.hint ?? ""}
                  onChange={(event) => {
                    const next = [...blanks];
                    next[index] = { ...blank, hint: event.target.value };
                    updateBlanks(next);
                  }}
                  placeholder="提示，可选"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => updateBlanks(blanks.filter((_, i) => i !== index))}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() =>
              updateBlanks([
                ...blanks,
                { id: `blank-${blanks.length + 1}`, answer: "", hint: "" },
              ])
            }
          >
            <Plus className="mr-1 size-4" />
            添加填空
          </Button>
        </div>
      </EditorSection>
    </div>
  );
}
