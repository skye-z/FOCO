"use client";

import * as React from "react";
import { Check, ChevronUp, X } from "lucide-react";
import type { KnowledgePoint } from "../types";

export function KpMultiSelect({
  knowledgePoints,
  selected,
  onToggle,
  disabled,
}: {
  knowledgePoints: KnowledgePoint[];
  selected: Set<string>;
  onToggle: (id: string) => void;
  disabled?: boolean;
}) {
  const [open, setOpen] = React.useState(false);
  const ref = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (!open) return;
    function handler(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node))
        setOpen(false);
    }
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => {
          if (!disabled) setOpen(!open);
        }}
        className={`flex w-full min-h-[38px] flex-wrap items-center gap-1 rounded-lg border bg-background px-3 py-2 text-left text-sm ${disabled ? "opacity-60 cursor-not-allowed" : ""}`}
        disabled={disabled}
      >
        {selected.size === 0 ? (
          <span className="text-muted-foreground">选择知识点</span>
        ) : (
          Array.from(selected).map((id) => {
            const kp = knowledgePoints.find((k) => k.id === id);
            return kp ? (
              <span
                key={id}
                className="inline-flex items-center gap-1 rounded bg-primary/10 px-2 py-0.5 text-xs text-primary"
              >
                {kp.name}
                {!disabled && (
                  <span
                    role="button"
                    tabIndex={0}
                    onClick={(e) => {
                      e.stopPropagation();
                      onToggle(id);
                    }}
                    className="ml-0.5 hover:text-destructive"
                  >
                    <X className="size-3" />
                  </span>
                )}
              </span>
            ) : null;
          })
        )}
        <span className="ml-auto shrink-0 text-muted-foreground">
          <ChevronUp className="size-3.5" />
        </span>
      </button>
      {open && !disabled && (
        <div className="absolute bottom-full z-50 mb-1 max-h-48 w-full overflow-auto rounded-lg border bg-background shadow-lg">
          {knowledgePoints.length === 0 ? (
            <p className="px-3 py-4 text-center text-xs text-muted-foreground">
              暂无知识点
            </p>
          ) : (
            knowledgePoints.map((kp) => (
              <button
                key={kp.id}
                type="button"
                onClick={() => onToggle(kp.id)}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-muted"
              >
                <span
                  className={`flex size-4 shrink-0 items-center justify-center rounded border ${selected.has(kp.id) ? "border-primary bg-primary text-primary-foreground" : "border-muted-foreground/30"}`}
                >
                  {selected.has(kp.id) && <Check className="size-3" />}
                </span>
                {kp.name}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
