"use client";

import * as React from "react";
import { ArrowDown, Plus } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { StepSchema } from "./step-list";
import { widgetTypeLabel } from "@/app/(dashboard)/exams/types";
import {
  VISUAL_BLOCKS,
  evaluateBlockCompleteness,
  getStepSummary,
  getStepTitle,
  type VisualBlockType,
} from "./block-types";

export function FlowCanvas({
  steps,
  selectedIndex,
  onSelect,
  onInsertBlock,
}: {
  steps: StepSchema[];
  selectedIndex: number;
  onSelect: (index: number) => void;
  onInsertBlock: (type: VisualBlockType, insertAt: number) => void;
}) {
  return (
    <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
      <div className="mx-auto max-w-3xl space-y-3">
        <div className="mb-4">
          <h2 className="text-base font-semibold">线性流程</h2>
          <p className="text-sm text-muted-foreground">
            交互单元按线性步骤执行。点击任意块以编辑其配置与预览。
          </p>
        </div>

        {steps.length === 0 ? (
          <div className="rounded-xl border border-dashed bg-background px-6 py-12 text-center text-sm text-muted-foreground">
            先从左侧块库添加一个交互块，开始搭建学习流程。
          </div>
        ) : (
          steps.map((step, index) => (
            <React.Fragment key={step.id ?? `flow-step-${index}`}>
              <InsertBlockButton onInsert={(type) => onInsertBlock(type, index)} />
              <button
                type="button"
                onClick={() => onSelect(index)}
                className="block w-full text-left"
              >
                <Card className={selectedIndex === index ? "border-primary shadow-sm" : "hover:border-primary/40"}>
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <CardTitle className="text-sm">步骤 {index + 1}</CardTitle>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {getStepTitle(step)}
                        </p>
                      </div>
                      <Badge variant="outline">{widgetTypeLabel(step.widget_type)}</Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-2">
                    <div className="text-xs text-muted-foreground">
                      {Array.isArray(step.knowledge_point_tags) && step.knowledge_point_tags.length > 0
                        ? `知识点: ${step.knowledge_point_tags.join(" / ")}`
                        : "未绑定知识点"}
                    </div>
                    <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                      <Badge variant="secondary" className="text-[11px]">
                        {computeStepStatus(step).label}
                      </Badge>
                      <span>{getStepSummary(step)}</span>
                    </div>
                    {computeStepStatus(step).missing.length > 0 ? (
                      <div className="text-[11px] text-amber-700">
                        缺少: {computeStepStatus(step).missing.join("、")}
                      </div>
                    ) : null}
                  </CardContent>
                </Card>
              </button>
              {index < steps.length - 1 ? (
                <div className="flex justify-center py-1 text-muted-foreground">
                  <ArrowDown className="size-4" />
                </div>
              ) : null}
            </React.Fragment>
          ))
        )}
        {steps.length > 0 ? (
          <InsertBlockButton onInsert={(type) => onInsertBlock(type, steps.length)} />
        ) : null}
      </div>
    </div>
  );
}

function InsertBlockButton({ onInsert }: { onInsert: (type: VisualBlockType) => void }) {
  return (
    <div className="flex justify-center py-1">
      <DropdownMenu>
        <DropdownMenuTrigger
          className="inline-flex items-center justify-center gap-1.5 rounded-full border bg-background px-3 h-8 text-xs font-medium hover:bg-muted transition-colors"
        >
          <Plus className="size-3.5" />
          插入块
        </DropdownMenuTrigger>
        <DropdownMenuContent align="center" side="bottom" sideOffset={6} className="w-64">
          {VISUAL_BLOCKS.map((block) => (
            <DropdownMenuItem key={block.type} onClick={() => onInsert(block.type)}>
              <div className="space-y-0.5">
                <div className="text-sm font-medium">{block.label}</div>
                <div className="text-xs text-muted-foreground">{block.description}</div>
              </div>
            </DropdownMenuItem>
          ))}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

function computeStepStatus(step: StepSchema) {
  const completeness = evaluateBlockCompleteness(step);
  return {
    label: completeness.complete ? "已配置" : "待完善",
    missing: completeness.missing,
  };
}
