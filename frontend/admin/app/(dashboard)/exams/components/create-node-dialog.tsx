"use client";

import * as React from "react";
import { authFetch } from "@/lib/auth-session";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import type { TreeNode } from "../types";
import { showActionError } from "../types";

export function CreateNodeDialog({
  type,
  parentId,
  tree,
  token,
  onClose,
  onCreated,
}: {
  type: "exam" | "subject" | "chapter" | null;
  parentId: string;
  tree: TreeNode[];
  token: string | null;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [code, setCode] = React.useState("");
  const [name, setName] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  if (!type) return null;

  const title =
    type === "exam" ? "添加考试" : type === "subject" ? "添加科目" : "添加章节";
  const parentLabel =
    type === "subject"
      ? (tree.find((e) => e.id === parentId)?.name ?? "")
      : type === "chapter"
        ? (tree.flatMap((e) => e.children ?? []).find((s) => s.id === parentId)
            ?.name ?? "")
        : "";

  async function handleCreate() {
    if (!token || !name.trim()) return;
    setSaving(true);
    try {
      const autoCode = code || name.trim().replace(/\s+/g, "_").toLowerCase();
      if (type === "exam") {
        await authFetch("/api/v1/admin/exams", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ code: autoCode, name: name.trim() }),
        });
      } else if (type === "subject") {
        await authFetch("/api/v1/admin/subjects", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            exam_id: parentId,
            code: autoCode,
            name: name.trim(),
            sort_order: 0,
          }),
        });
      } else if (type === "chapter") {
        await authFetch("/api/v1/admin/chapters", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            subject_id: parentId,
            code: autoCode,
            name: name.trim(),
            sort_order: 0,
          }),
        });
      }
      setCode("");
      setName("");
      onCreated();
      toast.success("创建成功");
    } catch {
      showActionError("创建失败", "新增节点未完成，请检查输入后重试。");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog
      open={type !== null}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          {parentLabel && (
            <div className="rounded-lg bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
              上级: {parentLabel}
            </div>
          )}
          <div className="space-y-2">
            <Label>名称</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={
                type === "exam"
                  ? "如: CFA Level I"
                  : type === "subject"
                    ? "如: 定量方法"
                    : "如: 货币时间价值"
              }
            />
          </div>
          <div className="space-y-2">
            <Label>编码（可选）</Label>
            <Input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="自动生成，或手动输入"
            />
          </div>
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button
            type="button"
            onClick={handleCreate}
            disabled={saving || !name.trim()}
          >
            {saving ? "创建中..." : "创建"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
