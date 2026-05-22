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
import { showActionError } from "../types";

export function EditNodeDialog({
  open,
  node,
  token,
  onClose,
  onSaved,
}: {
  open: boolean;
  node: { type: "exam" | "subject" | "chapter"; id: string; name: string } | null;
  token: string | null;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [name, setName] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  React.useEffect(() => {
    setName(node?.name ?? "");
  }, [node]);

  if (!node) return null;

  const activeNode = node;

  const title =
    activeNode.type === "exam"
      ? "编辑考试"
      : activeNode.type === "subject"
        ? "编辑科目"
        : "编辑章节";

  async function handleSave() {
    if (!token || !name.trim()) return;
    setSaving(true);
    try {
      const endpoint =
        activeNode.type === "exam"
          ? `/api/v1/admin/exams/${activeNode.id}`
          : activeNode.type === "subject"
            ? `/api/v1/admin/subjects/${activeNode.id}`
            : `/api/v1/admin/chapters/${activeNode.id}`;
      const res = await authFetch(endpoint, {
        method: "PATCH",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: name.trim() }),
      });
      if (!res.ok) {
        showActionError("编辑失败", "名称更新未完成，请稍后重试。");
        return;
      }
      toast.success("名称已更新");
      onSaved();
    } catch {
      showActionError("编辑失败", "名称更新未完成，请稍后重试。");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) onClose();
      }}
    >
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <div className="space-y-2 py-2">
          <Label>名称</Label>
          <Input value={name} onChange={(event) => setName(event.target.value)} />
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button type="button" onClick={handleSave} disabled={saving || !name.trim()}>
            {saving ? "保存中..." : "保存"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
