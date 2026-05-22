"use client";

import * as React from "react";
import type { StepSchema } from "./step-list";
import type { VisualBlockType } from "./block-types";
import { EditorSection } from "./block-editors/common";
import { ChoiceEditor } from "./block-editors/choice-editor";
import { ConceptMatchEditor } from "./block-editors/concept-match-editor";
import { ConditionMarkEditor } from "./block-editors/condition-mark-editor";
import { FillBlankEditor } from "./block-editors/fill-blank-editor";
import { FormulaDragEditor } from "./block-editors/formula-drag-editor";
import { ParameterChartEditor } from "./block-editors/parameter-chart-editor";
import { ReasoningSortEditor } from "./block-editors/reasoning-sort-editor";

export function VisualBlockInspector({
  blockType,
  step,
  onChange,
}: {
  blockType: VisualBlockType;
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  if (blockType === "formula_drag") {
    return <FormulaDragEditor step={step} onChange={onChange} />;
  }

  if (blockType === "parameter_chart") {
    return <ParameterChartEditor step={step} onChange={onChange} />;
  }

  if (blockType === "reasoning_sort") {
    return <ReasoningSortEditor step={step} onChange={onChange} />;
  }

  if (blockType === "concept_match") {
    return <ConceptMatchEditor step={step} onChange={onChange} />;
  }

  if (blockType === "condition_mark") {
    return <ConditionMarkEditor step={step} onChange={onChange} />;
  }

  if (blockType === "choice") {
    return <ChoiceEditor step={step} onChange={onChange} />;
  }

  if (blockType === "fill_blank") {
    return <FillBlankEditor step={step} onChange={onChange} />;
  }

  return (
    <EditorSection
      title="可视化配置"
      description="交互块将逐步从 JSON 编辑切换到结构化表单。当前阶段先使用块类型分发骨架。"
    >
      <div className="space-y-2 text-sm text-muted-foreground">
        <p>当前块类型：{blockType}</p>
        <p>下一阶段会在这里渲染块专属配置表单。</p>
      </div>
      <button
        type="button"
        className="hidden"
        onClick={() => onChange(step)}
      >
        noop
      </button>
    </EditorSection>
  );
}
