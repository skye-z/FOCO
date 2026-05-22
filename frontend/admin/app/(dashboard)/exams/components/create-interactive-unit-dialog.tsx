"use client";

import * as React from "react";
import { LoaderCircle } from "lucide-react";
import { useRouter } from "next/navigation";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { toast } from "sonner";
import type { TreeNode } from "../types";
import { showActionError } from "../types";

export function CreateInteractiveUnitDialog({
  open,
  tree,
  token,
  onClose,
}: {
  open: boolean;
  tree: TreeNode[];
  token: string | null;
  onClose: () => void;
}) {
  const router = useRouter();
  const [examId, setExamId] = React.useState("");
  const [subjectId, setSubjectId] = React.useState("");
  const [title, setTitle] = React.useState("");
  const [saving, setSaving] = React.useState(false);

  const selectedExam = tree.find((e) => e.id === examId);
  const subjects = selectedExam?.children ?? [];

  React.useEffect(() => {
    if (!open) {
      setExamId("");
      setSubjectId("");
      setTitle("");
    }
  }, [open]);

  async function handleCreate() {
    if (!token || !examId || !subjectId || !title.trim()) return;
    setSaving(true);
    try {
      const res = await authFetch("/api/v1/admin/interactive-units", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          exam_id: examId,
          subject_id: subjectId,
          title: title.trim(),
        }),
      });
      if (!res.ok) {
        showActionError("创建失败", "交互单元创建未完成，请稍后重试。");
        return;
      }
      const p = await res.json();
      toast.success("交互单元已创建");
      onClose();
      if (!p.data?.id) {
        showActionError("跳转失败", "交互单元已创建，但未返回详情 ID。");
        return;
      }
      router.push(`/interactive-units/${p.data.id}?exam_id=${encodeURIComponent(examId)}`);
    } catch {
      showActionError("创建失败", "交互单元创建未完成，请稍后重试。");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        if (!o) onClose();
      }}
    >
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>新建交互单元</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label>考试</Label>
            <Select
              value={examId}
              onValueChange={(v) => {
                setExamId(v ?? "");
                setSubjectId("");
              }}
            >
              <SelectTrigger>
                {selectedExam?.name ?? "选择考试"}
              </SelectTrigger>
              <SelectContent>
                {tree.map((e) => (
                  <SelectItem key={e.id} value={e.id}>
                    {e.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>科目</Label>
            <Select
              value={subjectId}
              onValueChange={(v) => setSubjectId(v ?? "")}
              disabled={!examId}
            >
              <SelectTrigger>
                {subjects.find((s) => s.id === subjectId)?.name ?? "选择科目"}
              </SelectTrigger>
              <SelectContent>
                {subjects.map((s) => (
                  <SelectItem key={s.id} value={s.id}>
                    {s.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>标题</Label>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="输入交互单元标题"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            取消
          </Button>
          <Button
            onClick={handleCreate}
            disabled={saving || !examId || !subjectId || !title.trim()}
          >
            {saving ? "创建中..." : "创建"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
