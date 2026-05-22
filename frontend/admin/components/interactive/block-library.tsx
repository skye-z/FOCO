"use client";

import * as React from "react";
import { Plus } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { VISUAL_BLOCKS, type VisualBlockType } from "./block-types";

export function BlockLibrary({
  onAdd,
}: {
  onAdd: (type: VisualBlockType) => void;
}) {
  return (
    <Card className="border-0 shadow-none">
      <CardHeader className="px-3 pb-2 pt-3">
        <CardTitle className="text-sm">块库</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 px-3 pb-3">
        {VISUAL_BLOCKS.map((block) => (
          <button
            key={block.type}
            type="button"
            onClick={() => onAdd(block.type)}
            className="flex w-full items-start justify-between rounded-lg border bg-background px-3 py-2 text-left transition-colors hover:bg-muted/50"
          >
            <div className="space-y-1">
              <div className="text-sm font-medium">{block.label}</div>
              <p className="text-xs text-muted-foreground">{block.description}</p>
            </div>
            <span className="flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground">
              <Plus className="size-4" />
            </span>
          </button>
        ))}
      </CardContent>
    </Card>
  );
}
