"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

type FormulaSlot = { id: string; label: string };
type FormulaVariable = { id: string; label: string; value?: string };

function normalizeSlots(step: StepSchema): Array<FormulaSlot & { options?: string[] }> {
  const raw = Array.isArray(step.content?.slots) ? step.content.slots : [];
  return raw.map((slot: any, index: number) => ({
    id: slot?.id ?? slot?.key ?? String(index),
    label: slot?.label ?? slot?.name ?? "",
    options: Array.isArray(slot?.options) ? slot.options.map((option: any) => String(option)) : undefined,
  }));
}

function normalizeVariables(step: StepSchema): FormulaVariable[] {
  if (Array.isArray(step.content?.variable_bank) && step.content.variable_bank.length > 0) {
    return step.content.variable_bank.map((item: any, index: number) => ({
      id: item?.id ?? String(index),
      label: item?.label ?? item?.name ?? "",
      value:
        item?.value !== undefined
          ? String(item.value)
          : item?.display_value !== undefined
            ? String(item.display_value)
            : undefined,
    }));
  }

  const slots = normalizeSlots(step);
  return slots.flatMap((slot) =>
    Array.isArray(slot.options)
      ? slot.options.map((option) => ({
          id: option,
          label: option,
          value: option,
        }))
      : [],
  );
}

export function FormulaDragEditor({
  step,
  onChange,
}: {
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  const formulaTemplate =
    (step.content?.formula_template as string | undefined) ??
    (step.content?.answer as string | undefined) ??
    (step.content?.formula as string | undefined) ??
    "";
  const slots = normalizeSlots(step);
  const variableBank = normalizeVariables(step);

  function updateContent(nextContent: Record<string, any>) {
    const nextSlots = Array.isArray(nextContent.slots) ? nextContent.slots : slots;
    const nextVariableBank = Array.isArray(nextContent.variable_bank)
      ? nextContent.variable_bank
      : variableBank;
    const syncedSlots = nextSlots.map((slot: any) => {
      if (!Array.isArray(slot?.options)) {
        return slot;
      }
      return {
        ...slot,
        options: nextVariableBank.map((item: FormulaVariable) => item.value ?? item.label),
      };
    });
    onChange({
      ...step,
      content: {
        ...step.content,
        ...nextContent,
        slots: syncedSlots,
        variable_bank: nextVariableBank,
        answer: nextContent.formula_template ?? step.content?.answer ?? formulaTemplate,
        formula: nextContent.formula_template ?? step.content?.formula ?? formulaTemplate,
      },
    });
  }

  return (
    <div className="space-y-4">
      <EditorSection title="公式模板" description="配置用户需要拼装的目标公式。">
        <div className="space-y-2">
          <Label>公式表达式</Label>
          <Input
            value={formulaTemplate}
            onChange={(event) => updateContent({ formula_template: event.target.value })}
            placeholder="如: PV = FV / (1 + r)^n"
          />
        </div>
      </EditorSection>

      <EditorSection title="变量槽位" description="定义公式中需要用户放置变量的位置。">
        <div className="space-y-2">
          {slots.map((slot, index) => (
            <div key={slot.id || index} className="flex items-center gap-2">
              <Input
                value={slot.label}
                onChange={(event) => {
                  const next = [...slots];
                  next[index] = { ...slot, label: event.target.value };
                  updateContent({ slots: next });
                }}
                placeholder="槽位名，如 PV"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => {
                  const next = slots.filter((_, i) => i !== index);
                  updateContent({ slots: next });
                }}
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
              updateContent({
                slots: [...slots, { id: `slot-${slots.length + 1}`, label: "" }],
              })
            }
          >
            <Plus className="mr-1 size-4" />
            添加槽位
          </Button>
        </div>
      </EditorSection>

      <EditorSection title="候选变量" description="配置可拖拽的变量池。">
        <div className="space-y-2">
          {variableBank.map((item, index) => (
            <div key={item.id || index} className="grid grid-cols-[1fr_1fr_auto] gap-2">
              <Input
                value={item.label}
                onChange={(event) => {
                  const next = [...variableBank];
                  next[index] = { ...item, label: event.target.value };
                  updateContent({ variable_bank: next });
                }}
                placeholder="变量名，如 FV"
              />
              <Input
                value={item.value ?? ""}
                onChange={(event) => {
                  const next = [...variableBank];
                  next[index] = { ...item, value: event.target.value };
                  updateContent({ variable_bank: next });
                }}
                placeholder="显示值，可选"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => {
                  const next = variableBank.filter((_, i) => i !== index);
                  updateContent({ variable_bank: next });
                }}
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
              updateContent({
                variable_bank: [
                  ...variableBank,
                  { id: `var-${variableBank.length + 1}`, label: "", value: "" },
                ],
              })
            }
          >
            <Plus className="mr-1 size-4" />
            添加变量
          </Button>
        </div>
      </EditorSection>
    </div>
  );
}
