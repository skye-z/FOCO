"use client";

import * as React from "react";
import {
  BookOpen,
  ChevronDown,
  ChevronRight,
  Pencil,
  GraduationCap,
  Layers,
  Menu,
  Plus,
  Trash2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { TreeNode } from "../types";

export function TreeLevel({
  node,
  expandedIds,
  selectedId,
  depth,
  onToggle,
  onSelect,
  onAdd,
  onEdit,
  onDelete,
}: {
  node: TreeNode;
  expandedIds: Set<string>;
  selectedId?: string;
  depth: number;
  onToggle: (id: string) => void;
  onSelect: (node: TreeNode) => void;
  onAdd: (type: "subject" | "chapter", parentId: string) => void;
  onEdit: (node: TreeNode) => void;
  onDelete: (type: string, id: string, name: string) => void;
}) {
  const hasChildren = node.children && node.children.length > 0;
  const expanded = expandedIds.has(node.id);
  const selected = selectedId === node.id;
  const indent = depth * 16;
  const icon =
    node.type === "exam" ? (
      <GraduationCap className="size-4" />
    ) : node.type === "subject" ? (
      <BookOpen className="size-4" />
    ) : (
      <Layers className="size-3.5" />
    );

  return (
    <div>
      <div className="relative group/tree-item">
        <button
          onClick={() => {
            onSelect(node);
            if (hasChildren) onToggle(node.id);
          }}
          className={`flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-muted ${selected ? "bg-primary/15 font-semibold text-primary ring-1 ring-primary/30" : "text-foreground"}`}
          style={{ paddingLeft: `${indent + 12}px` }}
        >
          {hasChildren ? (
            expanded ? (
              <ChevronDown className="size-3.5 shrink-0" />
            ) : (
              <ChevronRight className="size-3.5 shrink-0" />
            )
          ) : (
            <span className="w-3.5" />
          )}
          <span className="shrink-0 text-muted-foreground">{icon}</span>
          <span className="truncate">{node.name}</span>
          {node.type === "exam" && node.countdown_days != null && (
            <span className="ml-auto shrink-0 text-xs text-muted-foreground">
              {node.countdown_days > 0
                ? `${node.countdown_days}天后`
                : "已到考试日"}
            </span>
          )}
        </button>
        <div className="absolute right-1 top-1/2 -translate-y-1/2 opacity-0 transition-opacity group-hover/tree-item:opacity-100 focus-within:opacity-100">
          <DropdownMenu>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-7"
                  onClick={(event: React.MouseEvent<HTMLButtonElement>) =>
                    event.stopPropagation()
                  }
                />
              }
            >
              <Menu className="size-3.5" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" side="bottom" sideOffset={6}>
              <DropdownMenuGroup>
                {node.type === "exam" && (
                  <DropdownMenuItem onClick={() => onAdd("subject", node.id)}>
                    <Plus />
                    添加科目
                  </DropdownMenuItem>
                )}
                {node.type === "subject" && (
                  <DropdownMenuItem onClick={() => onAdd("chapter", node.id)}>
                    <Plus />
                    添加章节
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem
                  onClick={(event) => {
                    event.stopPropagation();
                    onEdit(node);
                  }}
                >
                  <Pencil />
                  编辑名称
                </DropdownMenuItem>
                <DropdownMenuItem
                  variant="destructive"
                  onClick={(event) => {
                    event.stopPropagation();
                    onDelete(node.type, node.id, node.name);
                  }}
                >
                  <Trash2 />
                  删除
                </DropdownMenuItem>
              </DropdownMenuGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
      {hasChildren && expanded && (
        <div>
          {node.children!.map((child) => (
            <TreeLevel
              key={child.id}
              node={child}
              expandedIds={expandedIds}
              selectedId={selectedId}
              depth={depth + 1}
              onToggle={onToggle}
              onSelect={onSelect}
              onAdd={onAdd}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}
