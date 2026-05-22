"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { widgetTypeLabel } from "@/app/(dashboard)/exams/types";
import type { StepSchema } from "./step-list";
import { OrderingMatchingPreview } from "./widget-preview/ordering-matching";
import { FormulaBuilderPreview } from "./widget-preview/formula-builder";
import { ParameterLabPreview } from "./widget-preview/parameter-lab";
import { HighlightMarkingPreview } from "./widget-preview/highlight-marking";
import { ChoiceClozePreview } from "./widget-preview/choice-cloze";

function WidgetRenderer({ step }: { step: StepSchema }) {
  switch (step.widget_type) {
    case "ordering_matching":
      return <OrderingMatchingPreview step={step} />;
    case "formula_builder":
      return <FormulaBuilderPreview step={step} />;
    case "parameter_lab":
      return <ParameterLabPreview step={step} />;
    case "highlight_marking":
      return <HighlightMarkingPreview step={step} />;
    case "choice_cloze":
      return <ChoiceClozePreview step={step} />;
    default:
      return (
        <p className="text-sm text-muted-foreground">
          未知组件类型: {step.widget_type}
        </p>
      );
  }
}

export function StepPreview({ step }: { step: StepSchema | null }) {
  return (
    <div className="flex w-[360px] shrink-0 flex-col border-l">
      <div className="border-b px-4 py-2">
        <span className="text-sm font-medium">预览</span>
      </div>
      <div className="flex-1 overflow-y-auto p-4">
        {!step ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            选择步骤预览
          </div>
        ) : (
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">
                预览: {widgetTypeLabel(step.widget_type)}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <WidgetRenderer step={step} />
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
