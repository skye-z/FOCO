"use client";

import * as React from "react";
import { Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { StepSchema } from "../step-list";
import { EditorSection } from "./common";

type ConceptPair = { left: string; right: string };

export function ConceptMatchEditor({
  step,
  onChange,
}: {
  step: StepSchema;
  onChange: (step: StepSchema) => void;
}) {
  const pairs = (Array.isArray(step.content?.pairs) ? step.content.pairs : []) as ConceptPair[];

  function updatePairs(nextPairs: ConceptPair[]) {
    onChange({
      ...step,
      content: { ...step.content, pairs: nextPairs },
      evaluation_config: { ...step.evaluation_config, correct_pairs: nextPairs },
    });
  }

  return (
    <EditorSection title="概念配对" description="配置概念与定义或案例的匹配关系。">
      <div className="space-y-2">
        <Label>配对项</Label>
        {pairs.map((pair, index) => (
          <div key={index} className="grid grid-cols-[1fr_1fr_auto] gap-2">
            <Input
              value={pair.left}
              onChange={(event) => {
                const next = [...pairs];
                next[index] = { ...pair, left: event.target.value };
                updatePairs(next);
              }}
              placeholder="概念"
            />
            <Input
              value={pair.right}
              onChange={(event) => {
                const next = [...pairs];
                next[index] = { ...pair, right: event.target.value };
                updatePairs(next);
              }}
              placeholder="定义 / 案例"
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => updatePairs(pairs.filter((_, i) => i !== index))}
            >
              <Trash2 className="size-4" />
            </Button>
          </div>
        ))}
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => updatePairs([...pairs, { left: "", right: "" }])}
        >
          <Plus className="mr-1 size-4" />
          添加配对
        </Button>
      </div>
    </EditorSection>
  );
}
