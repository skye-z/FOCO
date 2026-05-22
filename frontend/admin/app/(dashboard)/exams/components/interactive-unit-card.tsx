"use client";

import * as React from "react";
import { FlaskConical, Menu, Trash2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { InteractiveUnitSummary } from "../types";

export function InteractiveUnitCard({
  unit,
  onDelete,
}: {
  unit: InteractiveUnitSummary;
  onDelete: (unit: InteractiveUnitSummary) => void;
}) {
  const router = useRouter();

  function statusLabel(s: string) {
    if (s === "published" || s === "active") return "已发布";
    if (s === "draft") return "草稿";
    return s;
  }

  function statusColor(s: string) {
    if (s === "published" || s === "active") return "bg-blue-50 text-blue-700";
    return "bg-gray-100 text-gray-700";
  }

  function computedStatus(): string {
    if (unit.published_version_no || unit.status === "published") return "已发布";
    if (unit.has_unpublished_draft) return "草稿";
    return statusLabel(unit.status);
  }

  function computedStatusColor(): string {
    if (unit.published_version_no || unit.status === "published")
      return "bg-blue-50 text-blue-700";
    if (unit.has_unpublished_draft) return "bg-gray-100 text-gray-700";
    return statusColor(unit.status);
  }

  function handleOpen() {
    if (!unit.id) {
      return;
    }
    const qs = unit.exam_id ? `?exam_id=${encodeURIComponent(unit.exam_id)}` : "";
    router.push(`/interactive-units/${unit.id}${qs}`);
  }

  return (
    <Card
      className="group mb-4 cursor-pointer break-inside-avoid transition-all hover:shadow-md hover:-translate-y-0.5"
      onClick={handleOpen}
    >
      <CardContent>
        <div className="mb-3 flex items-center justify-between gap-2">
          <div className="flex flex-wrap items-center gap-1.5">
            <Badge
              variant="outline"
              className="bg-teal-50 text-xs text-teal-700"
            >
              <FlaskConical className="mr-1 size-3" />
              交互单元
            </Badge>
            <Badge
              variant="outline"
              className={`text-xs ${computedStatusColor()}`}
            >
              {computedStatus()}
            </Badge>
            <Badge variant="outline" className="text-xs text-muted-foreground">
              {unit.step_count} 步
            </Badge>
          </div>
          <div className="flex items-center gap-1">
            <div className="opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
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
                    <DropdownMenuItem
                      variant="destructive"
                      onClick={(event) => {
                        event.stopPropagation();
                        onDelete(unit);
                      }}
                    >
                      <Trash2 />
                      删除交互单元
                    </DropdownMenuItem>
                  </DropdownMenuGroup>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>
        <p className="mb-3 line-clamp-2 text-sm font-medium leading-relaxed">
          {unit.title || "（无标题）"}
        </p>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          {unit.subject_name && <span>{unit.subject_name}</span>}
        </div>
      </CardContent>
    </Card>
  );
}
