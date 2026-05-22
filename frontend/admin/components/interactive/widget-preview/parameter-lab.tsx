"use client";

import type { StepSchema } from "../step-list";

export function ParameterLabPreview({ step }: { step: StepSchema }) {
  const content = step.content ?? {};
  const parameters: Array<{
    name: string;
    label: string;
    value: string | number;
    default_value?: string | number;
    unit?: string;
  }> = Array.isArray(content.parameters)
    ? content.parameters
    : Array.isArray(content.params)
      ? content.params.map((param: any) => ({
          name: param.name ?? param.label ?? "",
          label: param.label ?? param.name ?? "",
          value: param.value ?? param.default ?? param.default_value ?? "",
          default_value: param.default_value ?? param.default ?? "",
          unit: param.unit,
        }))
      : [];

  if (parameters.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        在 content.parameters 中配置参数
      </p>
    );
  }

  return (
    <div className="space-y-2">
      <p className="text-xs font-medium text-muted-foreground">参数字段</p>
      <div className="space-y-1.5">
        {parameters.map((param, i) => (
          <div
            key={i}
            className="flex items-center justify-between rounded-md border bg-muted/30 px-3 py-2 text-sm"
          >
            <span className="text-muted-foreground">
              {param.label ?? param.name ?? `参数${i + 1}`}
            </span>
            <span className="font-mono">
              {param.value ?? param.default_value ?? "—"}
              {param.unit ? ` ${param.unit}` : ""}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
