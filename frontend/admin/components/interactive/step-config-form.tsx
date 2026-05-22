"use client";

import * as React from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import type { StepSchema } from "./step-list";
import {
  VISUAL_BLOCKS,
  type VisualBlockType,
  getBlockDefinition,
  getStepTitle,
  getStepVisualBlockType,
} from "./block-types";
import { VisualBlockInspector } from "./visual-block-inspector";

const ERROR_TYPES = [
  { value: "concept_confusion", label: "概念混淆" },
  { value: "formula_mapping_error", label: "公式映射错误" },
  { value: "parameter_misread", label: "参数读取错误" },
  { value: "reasoning_order_error", label: "推理顺序错误" },
  { value: "condition_missed", label: "关键条件遗漏" },
  { value: "careless_calculation", label: "粗心计算错误" },
];

export function StepConfigForm({
  step,
  onChange,
}: {
  step: StepSchema | null;
  onChange: (step: StepSchema) => void;
}) {
  if (!step) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-muted-foreground">
        选择左侧步骤以编辑配置
      </div>
    );
  }

  const currentBlockType = getStepVisualBlockType(step);
  const blockDefinition = getBlockDefinition(currentBlockType);

  return (
    <ScrollArea className="flex-1">
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label>交互块类型</Label>
          <Select
            value={currentBlockType}
            onValueChange={(v) =>
              onChange({
                ...step,
                content: {
                  ...step.content,
                  visual_block_type: v ?? currentBlockType,
                },
              })
            }
          >
            <SelectTrigger>{blockDefinition?.label ?? "选择交互块"}</SelectTrigger>
            <SelectContent>
              {VISUAL_BLOCKS.map((block) => (
                <SelectItem key={block.type} value={block.type}>
                  {block.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2 rounded-lg border bg-muted/20 p-3">
          <Label>块说明</Label>
          <p className="text-sm text-muted-foreground">
            {blockDefinition?.description ?? "未定义块描述"}
          </p>
          <div className="flex flex-wrap gap-2">
            <Badge variant="outline">widget: {step.widget_type}</Badge>
          </div>
        </div>

        <div className="space-y-2">
          <Label>块标题</Label>
          <Input
            value={(step.content?.title as string | undefined) ?? getStepTitle(step)}
            onChange={(event) =>
              onChange({
                ...step,
                content: { ...step.content, title: event.target.value },
              })
            }
            placeholder="输入块标题"
          />
        </div>

        <VisualBlockInspector
          blockType={currentBlockType}
          step={step}
          onChange={onChange}
        />

        <div className="space-y-4 rounded-lg border bg-muted/10 p-4">
          <div>
            <Label className="mb-2 block">主要错误类型</Label>
            <Select
              value={(step.feedback_map?.error_type as string | undefined) ?? "concept_confusion"}
              onValueChange={(value) =>
                onChange({
                  ...step,
                  feedback_map: { ...step.feedback_map, error_type: value },
                })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ERROR_TYPES.map((item) => (
                  <SelectItem key={item.value} value={item.value}>
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>正确反馈</Label>
            <Textarea
              value={(step.feedback_map?.success_message as string | undefined) ?? ""}
              onChange={(event) =>
                onChange({
                  ...step,
                  feedback_map: {
                    ...step.feedback_map,
                    success_message: event.target.value,
                  },
                })
              }
              className="min-h-[80px]"
              placeholder="用户答对后显示的反馈"
            />
          </div>

          <div className="space-y-2">
            <Label>错误反馈</Label>
            <Textarea
              value={(step.feedback_map?.failure_message as string | undefined) ?? ""}
              onChange={(event) =>
                onChange({
                  ...step,
                  feedback_map: {
                    ...step.feedback_map,
                    failure_message: event.target.value,
                  },
                })
              }
              className="min-h-[80px]"
              placeholder="用户答错后显示的反馈"
            />
          </div>

          <div className="space-y-2">
            <Label>提示策略</Label>
            <Textarea
              value={(step.hint_policy?.default_hint as string | undefined) ?? ""}
              onChange={(event) =>
                onChange({
                  ...step,
                  hint_policy: {
                    ...step.hint_policy,
                    default_hint: event.target.value,
                  },
                })
              }
              className="min-h-[80px]"
              placeholder="当用户卡住时提供的提示"
            />
          </div>
        </div>
      </div>
    </ScrollArea>
  );
}
