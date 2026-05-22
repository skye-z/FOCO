"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

type ParameterItem = {
  id: string;
  label: string;
  min?: string;
  max?: string;
  step?: string;
  default_value?: string;
};

function normalizeParameters(step: StepSchema): ParameterItem[] {
  const raw = Array.isArray(step.content?.parameters)
    ? step.content.parameters
    : Array.isArray(step.content?.params)
      ? step.content.params
      : [];
  return raw.map((parameter: any, index: number) => ({
    id: parameter.id ?? parameter.key ?? String(index),
    label: parameter.label ?? parameter.name ?? "",
    min: parameter.min !== undefined ? String(parameter.min) : "",
    max: parameter.max !== undefined ? String(parameter.max) : "",
    step: parameter.step !== undefined ? String(parameter.step) : "",
    default_value:
      parameter.default_value !== undefined
        ? String(parameter.default_value)
        : parameter.default !== undefined
          ? String(parameter.default)
          : "",
  }));
}

export function ParameterChartEditor({
  step,
  onChange,
}: {
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  const chartType = (step.content?.chart_type as string | undefined) ?? "line";
  const targetMetric =
    (step.content?.target_metric as string | undefined) ??
    (step.content?.formula as string | undefined) ??
    (step.content?.quiz_question as string | undefined) ??
    "";
  const parameters = normalizeParameters(step);

  function updateContent(nextContent: Record<string, any>) {
    const nextParameters = Array.isArray(nextContent.parameters)
      ? nextContent.parameters
      : parameters;
    onChange({
      ...step,
      content: {
        ...step.content,
        ...nextContent,
        parameters: nextParameters,
        params: nextParameters.map((parameter: ParameterItem) => ({
          name: parameter.label,
          label: parameter.label,
          min: parameter.min === "" ? undefined : Number(parameter.min),
          max: parameter.max === "" ? undefined : Number(parameter.max),
          step: parameter.step === "" ? undefined : Number(parameter.step),
          default: parameter.default_value === "" ? undefined : Number(parameter.default_value),
        })),
      },
    });
  }

  return (
    <div className="space-y-4">
      <EditorSection title="图表设置" description="定义参数实验的可视化表现形式。">
        <div className="space-y-2">
          <Label>图表类型</Label>
          <Select value={chartType} onValueChange={(value) => updateContent({ chart_type: value })}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="line">折线图</SelectItem>
              <SelectItem value="bar">柱状图</SelectItem>
              <SelectItem value="area">面积图</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>目标指标</Label>
          <Input
            value={targetMetric}
            onChange={(event) => updateContent({ target_metric: event.target.value })}
            placeholder="如: 净现值 / 久期 / 比率变化"
          />
        </div>
      </EditorSection>

      <EditorSection title="可调参数" description="定义用户可以滑动或修改的计算参数。">
        <div className="space-y-3">
          {parameters.map((parameter, index) => (
            <div key={parameter.id || index} className="rounded-md border p-3">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-medium">参数 {index + 1}</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => updateContent({ parameters: parameters.filter((_, i) => i !== index) })}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
              <div className="grid grid-cols-2 gap-2">
                <Input
                  value={parameter.label}
                  onChange={(event) => {
                    const next = [...parameters];
                    next[index] = { ...parameter, label: event.target.value };
                    updateContent({ parameters: next });
                  }}
                  placeholder="参数名，如 折现率"
                />
                <Input
                  value={parameter.default_value ?? ""}
                  onChange={(event) => {
                    const next = [...parameters];
                    next[index] = { ...parameter, default_value: event.target.value };
                    updateContent({ parameters: next });
                  }}
                  placeholder="默认值"
                />
                <Input
                  value={parameter.min ?? ""}
                  onChange={(event) => {
                    const next = [...parameters];
                    next[index] = { ...parameter, min: event.target.value };
                    updateContent({ parameters: next });
                  }}
                  placeholder="最小值"
                />
                <Input
                  value={parameter.max ?? ""}
                  onChange={(event) => {
                    const next = [...parameters];
                    next[index] = { ...parameter, max: event.target.value };
                    updateContent({ parameters: next });
                  }}
                  placeholder="最大值"
                />
                <Input
                  value={parameter.step ?? ""}
                  onChange={(event) => {
                    const next = [...parameters];
                    next[index] = { ...parameter, step: event.target.value };
                    updateContent({ parameters: next });
                  }}
                  placeholder="步长"
                />
              </div>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() =>
              updateContent({
                parameters: [
                  ...parameters,
                  { id: `param-${parameters.length + 1}`, label: "", min: "", max: "", step: "", default_value: "" },
                ],
              })
            }
          >
            <Plus className="mr-1 size-4" />
            添加参数
          </Button>
        </div>
      </EditorSection>
    </div>
  );
}
