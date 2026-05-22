"use client";

import * as React from "react";
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { widgetTypeLabel } from "@/app/(dashboard)/exams/types";
import type { StepReadiness } from "./editor-status";

export type StepSchema = {
  id?: string;
  widget_type: string;
  content: Record<string, any>;
  initial_state: Record<string, any>;
  allowed_actions: Record<string, any>;
  evaluation_config: Record<string, any>;
  feedback_map: Record<string, any>;
  hint_policy: Record<string, any>;
  knowledge_point_ids?: string[];
  knowledge_point_tags?: string[];
};

function SortableItem({
  step,
  index,
  isSelected,
  status,
  onSelect,
  onDelete,
}: {
  step: StepSchema;
  index: number;
  isSelected: boolean;
  status?: StepReadiness;
  onSelect: () => void;
  onDelete: (e: React.MouseEvent) => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: step.id ?? `step-${index}` });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`group flex items-center gap-2 rounded-md border px-3 py-2.5 text-sm transition-colors ${
        isSelected
          ? "border-primary bg-primary/5"
          : "border-transparent hover:bg-muted/50"
      }`}
      onClick={onSelect}
    >
      <button
        className="cursor-grab text-muted-foreground hover:text-foreground"
        {...attributes}
        {...listeners}
      >
        <GripVertical className="size-4" />
      </button>
      <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
        {index + 1}
      </span>
      <Badge variant="outline" className="text-xs">
        {widgetTypeLabel(step.widget_type)}
      </Badge>
      {status ? (
        <Badge
          variant="outline"
          className={`text-xs ${status.complete ? "text-emerald-700" : "text-amber-700"}`}
        >
          {status.complete ? "已配置" : "待完善"}
        </Badge>
      ) : null}
      <button
        className="ml-auto rounded p-0.5 text-muted-foreground opacity-0 transition-opacity hover:text-destructive group-hover:opacity-100"
        onClick={onDelete}
      >
        <Trash2 className="size-3.5" />
      </button>
    </div>
  );
}

export function StepList({
  steps,
  selectedIndex,
  stepStatuses,
  onSelect,
  onReorder,
  onDelete,
}: {
  steps: StepSchema[];
  selectedIndex: number;
  stepStatuses?: StepReadiness[];
  onSelect: (index: number) => void;
  onReorder: (oldIndex: number, newIndex: number) => void;
  onDelete: (index: number) => void;
}) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id) return;

    const oldIndex = steps.findIndex(
      (s, i) => (s.id ?? `step-${i}`) === active.id,
    );
    const newIndex = steps.findIndex(
      (s, i) => (s.id ?? `step-${i}`) === over.id,
    );
    if (oldIndex !== -1 && newIndex !== -1) {
      onReorder(oldIndex, newIndex);
    }
  }

  return (
    <div className="flex flex-1 flex-col">
      <div className="flex items-center justify-between border-b px-3 py-2">
        <span className="text-sm font-medium">步骤目录</span>
      </div>
      <div className="flex-1 overflow-y-auto p-2">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={steps.map((s, i) => s.id ?? `step-${i}`)}
            strategy={verticalListSortingStrategy}
          >
            <div className="flex flex-col gap-1">
              {steps.map((step, index) => (
                <SortableItem
                  key={step.id ?? `step-${index}`}
                    step={step}
                    index={index}
                    isSelected={index === selectedIndex}
                    status={stepStatuses?.[index]}
                    onSelect={() => onSelect(index)}
                    onDelete={(e) => {
                      e.stopPropagation();
                    onDelete(index);
                  }}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>
        {steps.length === 0 && (
          <p className="py-8 text-center text-sm text-muted-foreground">
            暂无步骤，点击上方添加
          </p>
        )}
      </div>
    </div>
  );
}
