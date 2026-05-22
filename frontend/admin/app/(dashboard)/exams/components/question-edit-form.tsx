"use client";

import * as React from "react";
import { LoaderCircle, Send } from "lucide-react";
import { authFetch } from "@/lib/auth-session";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { KpMultiSelect } from "./kp-multi-select";
import {
  type TreeNode,
  type KnowledgePoint,
  type QuestionVersionSummary,
  type VersionDetail,
  DIFFICULTY_OPTIONS,
  statusLabel,
  typeLabel,
  difficultyLabel,
  jsonField,
  jsonToOptions,
  showActionError,
} from "../types";

export function QuestionEditForm({
  detail: initial,
  tree,
  knowledgePoints,
  token,
  versions,
  historyLoading,
  onSave,
  onCancel,
  onOpenVersion,
  onRestoreVersion,
  onPublish,
  publishing,
}: {
  detail: VersionDetail;
  tree: TreeNode[];
  knowledgePoints: KnowledgePoint[];
  token: string | null;
  versions: QuestionVersionSummary[];
  historyLoading: boolean;
  onSave: (detail: VersionDetail) => void | Promise<void>;
  onCancel: () => void;
  onOpenVersion: (versionId: string) => void | Promise<void>;
  onRestoreVersion: (versionId: string) => void | Promise<void>;
  onPublish: () => void;
  publishing: boolean;
}) {
  const NONE_CHAPTER_VALUE = "__none__";
  const [saving, setSaving] = React.useState(false);
  const [stem, setStem] = React.useState(() => jsonField(initial.stem, "text"));
  const [options, setOptions] = React.useState(() =>
    jsonToOptions(initial.options),
  );
  const [correctAnswer, setCorrectAnswer] = React.useState(() =>
    jsonField(initial.correct_answer, "answer"),
  );
  const [explanation, setExplanation] = React.useState(() =>
    jsonField(initial.explanation, "text"),
  );
  const [difficulty, setDifficulty] = React.useState(
    String(initial.difficulty),
  );
  const [subjectId, setSubjectId] = React.useState(initial.subject_id);
  const [chapterId, setChapterId] = React.useState(
    initial.chapter_id ?? NONE_CHAPTER_VALUE,
  );
  const [selectedKPs, setSelectedKPs] = React.useState<Set<string>>(
    new Set(initial.knowledge_point_ids),
  );

  React.useEffect(() => {
    setStem(jsonField(initial.stem, "text"));
    setOptions(jsonToOptions(initial.options));
    setCorrectAnswer(jsonField(initial.correct_answer, "answer"));
    setExplanation(jsonField(initial.explanation, "text"));
    setDifficulty(String(initial.difficulty));
    setSubjectId(initial.subject_id);
    setChapterId(initial.chapter_id ?? NONE_CHAPTER_VALUE);
    setSelectedKPs(new Set(initial.knowledge_point_ids));
  }, [initial]);

  const exam =
    tree.find(
      (e) => e.children?.some((s) => s.id === initial.subject_id) || false,
    ) ?? tree[0];
  const subjects = exam?.children ?? [];
  const selectedSubject = subjects.find((s) => s.id === subjectId);
  const chapters = selectedSubject?.children ?? [];
  const examKPs = knowledgePoints;

  function toggleKp(id: string) {
    setSelectedKPs((prev) => {
      const n = new Set(prev);
      if (n.has(id)) n.delete(id);
      else n.add(id);
      return n;
    });
  }

  async function handleSave() {
    if (!token) return;
    setSaving(true);
    try {
      const optsObj: Record<string, string> = {};
      options.forEach((o, i) => {
        if (o.trim()) optsObj[String.fromCharCode(65 + i)] = o.trim();
      });
      const res = await authFetch(
        `/api/v1/admin/question-versions/${initial.version_id}`,
        {
          method: "PATCH",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            stem: { text: stem },
            options: optsObj,
            correct_answer: { answer: correctAnswer },
            explanation: { text: explanation },
            difficulty: parseInt(difficulty),
            subject_id: subjectId,
            chapter_id: chapterId === NONE_CHAPTER_VALUE ? null : chapterId,
            knowledge_point_ids: Array.from(selectedKPs),
          }),
        },
      );
      if (!res.ok) {
        showActionError("保存失败", "题目修改未保存成功，请检查内容后重试。");
        return;
      }
      const payload = await res.json();
      toast.success("题目已保存");
      await onSave(payload.data as VersionDetail);
    } catch {
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
      <div className="max-h-[72vh] overflow-y-auto pr-4">
        <div className="space-y-5 pb-2">
          <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
            <span>{typeLabel(initial.question_type)}</span>
            <span>·</span>
            <Badge
              variant="outline"
              className={`text-xs ${initial.status === "draft" ? "bg-gray-100 text-gray-600" : "bg-blue-50 text-blue-700"}`}
            >
              {statusLabel(initial.status)}
            </Badge>
            <span>·</span>
            <span>v{initial.version_no}</span>
            {initial.status === "published" ? (
              <Badge variant="outline" className="bg-blue-50 text-blue-700">
                当前查看的是正式版，保存将生成新草稿
              </Badge>
            ) : (
              <Badge variant="outline" className="bg-gray-100 text-gray-700">
                当前正在编辑草稿
              </Badge>
            )}
          </div>

          <div className="space-y-2">
            <Label className="text-sm font-medium">题干</Label>
            <Textarea
              value={stem}
              onChange={(e) => setStem(e.target.value)}
              className="min-h-[80px]"
              placeholder="请输入题干内容"
            />
          </div>

          {(initial.question_type === "single_choice" ||
            initial.question_type === "multiple_choice") && (
            <div className="space-y-2">
              <Label className="text-sm font-medium">选项</Label>
              {options.map((opt, i) => (
                <div key={i} className="flex items-center gap-2">
                  <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
                    {String.fromCharCode(65 + i)}
                  </span>
                  <Input
                    value={opt}
                    onChange={(e) => {
                      const n = [...options];
                      n[i] = e.target.value;
                      setOptions(n);
                    }}
                    placeholder={`选项 ${String.fromCharCode(65 + i)}`}
                    className="flex-1"
                  />
                </div>
              ))}
              <div className="space-y-2 pt-1">
                <Label className="text-sm font-medium">正确答案</Label>
                <Select
                  value={correctAnswer}
                  onValueChange={(v) => setCorrectAnswer(v ?? "A")}
                >
                  <SelectTrigger className="w-24">
                    {correctAnswer}
                  </SelectTrigger>
                  <SelectContent>
                    {options.map((_, i) => (
                      <SelectItem key={i} value={String.fromCharCode(65 + i)}>
                        {String.fromCharCode(65 + i)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}

          {initial.question_type === "judgment" && (
            <div className="space-y-2">
              <Label className="text-sm font-medium">正确答案</Label>
              <Select
                value={correctAnswer}
                onValueChange={(v) => setCorrectAnswer(v ?? "true")}
              >
                <SelectTrigger className="w-24">
                  {correctAnswer === "true" ? "正确" : "错误"}
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="true">正确</SelectItem>
                  <SelectItem value="false">错误</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-2">
            <Label className="text-sm font-medium">解析</Label>
            <Textarea
              value={explanation}
              onChange={(e) => setExplanation(e.target.value)}
              className="min-h-[60px]"
              placeholder="请输入题目解析"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label className="text-sm font-medium">难度</Label>
              <Select
                value={difficulty}
                onValueChange={(v) => setDifficulty(v ?? "3")}
              >
                <SelectTrigger>
                  {difficultyLabel(parseInt(difficulty))}
                </SelectTrigger>
                <SelectContent>
                  {DIFFICULTY_OPTIONS.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label className="text-sm font-medium">科目</Label>
              <Select
                value={subjectId}
                onValueChange={(v) => {
                  setSubjectId(v ?? "");
                  setChapterId(NONE_CHAPTER_VALUE);
                }}
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
              <Label className="text-sm font-medium">章节</Label>
              <Select
                value={chapterId}
                onValueChange={(v) => setChapterId(v ?? NONE_CHAPTER_VALUE)}
              >
                <SelectTrigger>
                  {chapterId === NONE_CHAPTER_VALUE
                    ? "无章节"
                    : (chapters.find((c) => c.id === chapterId)?.name ??
                      "选择章节")}
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={NONE_CHAPTER_VALUE}>无章节</SelectItem>
                  {chapters.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label className="text-sm font-medium">知识点（多选）</Label>
            <KpMultiSelect
              knowledgePoints={examKPs}
              selected={selectedKPs}
              onToggle={toggleKp}
            />
          </div>

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button
              variant="outline"
              type="button"
              onClick={onCancel}
              disabled={saving || publishing}
            >
              取消
            </Button>
            <Button
              type="button"
              onClick={handleSave}
              disabled={saving || publishing}
            >
              {saving ? (
                <LoaderCircle className="mr-1 size-4 animate-spin" />
              ) : null}
              {saving ? "保存中..." : "保存"}
            </Button>
            {initial.status === "draft" && (
              <Button
                type="button"
                className="bg-emerald-600 hover:bg-emerald-700"
                onClick={onPublish}
                disabled={saving || publishing}
              >
                {publishing ? (
                  <LoaderCircle className="mr-1 size-4 animate-spin" />
                ) : (
                  <Send className="mr-1 size-4" />
                )}
                {publishing ? "发布中..." : "发布"}
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="rounded-xl border bg-muted/20 p-4">
        <div className="mb-4 flex items-center justify-between">
          <div>
            <h3 className="text-sm font-semibold text-foreground">历史版本</h3>
            <p className="text-xs text-muted-foreground">
              查看、切换或从旧版本恢复新草稿
            </p>
          </div>
        </div>

        {historyLoading ? (
          <div className="flex items-center justify-center py-10">
            <LoaderCircle className="size-5 animate-spin text-muted-foreground" />
          </div>
        ) : versions.length === 0 ? (
          <div className="rounded-lg border border-dashed px-3 py-8 text-center text-sm text-muted-foreground">
            暂无历史版本
          </div>
        ) : (
          <div className="max-h-[64vh] overflow-y-auto pr-2">
            <div className="space-y-3">
              {versions.map((version) => (
                <div
                  key={version.version_id}
                  className={`w-full rounded-xl border p-3 text-left transition-colors hover:bg-muted ${version.version_id === initial.version_id ? "border-primary bg-primary/5" : "border-border bg-background"}`}
                >
                  <div className="mb-2 flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-foreground">
                        v{version.version_no}
                      </span>
                      <Badge
                        variant="outline"
                        className={
                          version.status === "published"
                            ? "bg-blue-50 text-blue-700"
                            : "bg-gray-100 text-gray-700"
                        }
                      >
                        {statusLabel(version.status)}
                      </Badge>
                      {version.is_current ? (
                        <Badge
                          variant="outline"
                          className="bg-emerald-50 text-emerald-700"
                        >
                          当前正式版
                        </Badge>
                      ) : null}
                    </div>
                  </div>
                  <p className="mb-2 text-xs text-muted-foreground">
                    {version.published_at
                      ? `发布时间：${new Date(version.published_at).toLocaleString("zh-CN")}`
                      : `更新时间：${new Date(version.updated_at).toLocaleString("zh-CN")}`}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        onOpenVersion(version.version_id);
                      }}
                    >
                      查看版本
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        onRestoreVersion(version.version_id);
                      }}
                    >
                      恢复为草稿
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
