"use client";

import * as React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { StepSchema } from "./step-list";
import type { PublishReadiness } from "./editor-status";

export type BlockResult = {
  block_id: string;
  widget_type: string;
  knowledge_point_ids: string[];
  knowledge_point_tags: string[];
  user_input: Record<string, any>;
  expected_answer: Record<string, any>;
  is_correct: boolean;
  score: number;
  feedback_message: string;
  error_type: string;
  time_spent_ms: number;
  hint_used: boolean;
  attempt_count: number;
};

export type SessionSummary = {
  completed_blocks: number;
  correct_blocks: number;
  overall_score: number;
  knowledge_point_breakdown: Record<string, number>;
  error_type_breakdown: Record<string, number>;
};

export function buildDraftBlockResults(steps: StepSchema[]): BlockResult[] {
  return steps.map((step, index) => ({
    block_id: step.id ?? `draft-step-${index + 1}`,
    widget_type: step.widget_type,
    knowledge_point_ids: Array.isArray(step.knowledge_point_ids)
      ? step.knowledge_point_ids
      : [],
    knowledge_point_tags: Array.isArray(step.knowledge_point_tags)
      ? step.knowledge_point_tags
      : [],
    user_input: {},
    expected_answer: step.evaluation_config ?? {},
    is_correct: false,
    score: 0,
    feedback_message:
      (step.feedback_map?.failure_message as string | undefined) ?? "待配置错误反馈",
    error_type:
      (step.feedback_map?.error_type as string | undefined) ?? "unattempted",
    time_spent_ms: 0,
    hint_used: Boolean(step.hint_policy?.default_hint),
    attempt_count: 0,
  }));
}

export function buildDraftSessionSummary(results: BlockResult[]): SessionSummary {
  const knowledge_point_breakdown: Record<string, number> = {};
  const error_type_breakdown: Record<string, number> = {};

  for (const result of results) {
    for (const tag of result.knowledge_point_tags) {
      knowledge_point_breakdown[tag] = (knowledge_point_breakdown[tag] ?? 0) + 1;
    }
    error_type_breakdown[result.error_type] =
      (error_type_breakdown[result.error_type] ?? 0) + 1;
  }

  return {
    completed_blocks: results.length,
    correct_blocks: results.filter((item) => item.is_correct).length,
    overall_score: results.length === 0 ? 0 : 0,
    knowledge_point_breakdown,
    error_type_breakdown,
  };
}

export function ResultSummaryPanel({
  steps,
  readiness,
}: {
  steps: StepSchema[];
  readiness?: PublishReadiness;
}) {
  const results = React.useMemo(() => buildDraftBlockResults(steps), [steps]);
  const summary = React.useMemo(() => buildDraftSessionSummary(results), [results]);

  return (
    <Card className="m-4 mt-0">
      <CardHeader className="pb-3">
        <CardTitle className="text-sm">结果汇总预览</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm text-muted-foreground">
        <p>已配置块数：{summary.completed_blocks}</p>
        <p>知识点标签数：{Object.keys(summary.knowledge_point_breakdown).length}</p>
        <p>错误类型维度：{Object.keys(summary.error_type_breakdown).join(" / ") || "无"}</p>
        {readiness ? (
          <>
            <p>发布状态：{readiness.canPublish ? "可以发布" : "仍有阻塞项"}</p>
            <p>待完善步骤数：{readiness.incompleteSteps.length}</p>
            {readiness.missing.length > 0 ? (
              <p>基础缺项：{readiness.missing.join("、")}</p>
            ) : null}
            {readiness.incompleteSteps.length > 0 ? (
              <div className="space-y-1 text-amber-700">
                {readiness.incompleteSteps.map((item) => (
                  <p key={`readiness-${item.index}`}>
                    步骤 {item.index + 1} 缺少：{item.missing.join("、")}
                  </p>
                ))}
              </div>
            ) : null}
          </>
        ) : null}
        <p>
          已配置提示策略块数：
          {results.filter((item) => item.hint_used).length}
        </p>
        <p>
          已配置错误反馈块数：
          {results.filter((item) => item.feedback_message !== "待配置错误反馈").length}
        </p>
      </CardContent>
    </Card>
  );
}
