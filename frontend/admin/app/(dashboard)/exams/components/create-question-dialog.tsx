"use client";

import * as React from "react";
import { LoaderCircle } from "lucide-react";
import { authFetch } from "@/lib/auth-session";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import type { TreeNode, QuestionCard } from "../types";
import {
  DIFFICULTY_OPTIONS,
  typeLabel,
  difficultyLabel,
  showActionError,
} from "../types";

export function CreateQuestionDialog({
  open,
  tree,
  token,
  onClose,
  onCreated,
}: {
  open: boolean;
  tree: TreeNode[];
  token: string | null;
  onClose: () => void;
  onCreated: (card: QuestionCard) => void;
}) {
  const NONE_CHAPTER_VALUE = "__none__";
  const [step, setStep] = React.useState<"select" | "edit">("select");
  const [examId, setExamId] = React.useState("");
  const [subjectId, setSubjectId] = React.useState("");
  const [chapterId, setChapterId] = React.useState(NONE_CHAPTER_VALUE);
  const [questionType, setQuestionType] = React.useState("single_choice");
  const [saving, setSaving] = React.useState(false);

  const [stem, setStem] = React.useState("");
  const [opts, setOpts] = React.useState(["", "", "", ""]);
  const [correctAnswer, setCorrectAnswer] = React.useState("A");
  const [explanation, setExplanation] = React.useState("");
  const [difficulty, setDifficulty] = React.useState("3");

  const selectedExam = tree.find((e) => e.id === examId);
  const subjects = selectedExam?.children ?? [];
  const selectedSubject = subjects.find((s) => s.id === subjectId);
  const chapters = selectedSubject?.children ?? [];

  React.useEffect(() => {
    if (!open) {
      setStep("select");
      setExamId("");
      setSubjectId("");
      setChapterId(NONE_CHAPTER_VALUE);
      setQuestionType("single_choice");
      setStem("");
      setOpts(["", "", "", ""]);
      setCorrectAnswer("A");
      setExplanation("");
      setDifficulty("3");
    }
  }, [open]);

  async function handleCreate() {
    if (!token || !examId || !subjectId || !stem.trim()) return;
    setSaving(true);
    try {
      const chapId = chapterId === NONE_CHAPTER_VALUE ? null : chapterId;
      const qRes = await authFetch("/api/v1/admin/questions", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          exam_id: examId,
          subject_id: subjectId,
          chapter_id: chapId,
        }),
      });
      if (!qRes.ok) {
        setSaving(false);
        return;
      }
      const qData = await qRes.json();
      const questionId = qData.data.id;

      const optsObj: Record<string, string> = {};
      if (
        questionType === "single_choice" ||
        questionType === "multiple_choice"
      ) {
        opts.forEach((o, i) => {
          if (o.trim()) optsObj[String.fromCharCode(65 + i)] = o.trim();
        });
      }

      const vRes = await authFetch(
        `/api/v1/admin/questions/${questionId}/versions`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            question_type: questionType,
            difficulty: parseInt(difficulty),
            stem: { text: stem },
            options: optsObj,
            correct_answer: { answer: correctAnswer },
            explanation: { text: explanation },
          }),
        },
      );
      if (!vRes.ok) {
        setSaving(false);
        return;
      }
      const vData = await vRes.json();
      toast.success("题目已创建");
      onCreated({
        id: questionId,
        exam_id: examId,
        subject_id: subjectId,
        subject_name: selectedSubject?.name ?? "",
        chapter_id: chapId,
        chapter_name: chapId
          ? (chapters.find((c) => c.id === chapId)?.name ?? null)
          : null,
        status: "draft",
        question_type: questionType,
        difficulty: parseInt(difficulty),
        version_no: 1,
        version_id: vData.data.id,
        stem_preview: stem.slice(0, 120),
      });
    } catch {
      showActionError("创建失败", "题目创建未完成，请检查内容后重试。");
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
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>新建题目</DialogTitle>
        </DialogHeader>
        {step === "select" ? (
          <div className="space-y-5 py-2">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>考试</Label>
                <Select
                  value={examId}
                  onValueChange={(v) => {
                    setExamId(v ?? "");
                    setSubjectId("");
                    setChapterId(NONE_CHAPTER_VALUE);
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
                <Label>题型</Label>
                <Select
                  value={questionType}
                  onValueChange={(v) => setQuestionType(v ?? "single_choice")}
                >
                  <SelectTrigger>{typeLabel(questionType)}</SelectTrigger>
                  <SelectContent>
                    <SelectItem value="single_choice">单选题</SelectItem>
                    <SelectItem value="multiple_choice">多选题</SelectItem>
                    <SelectItem value="judgment">判断题</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>科目</Label>
                <Select
                  value={subjectId}
                  onValueChange={(v) => {
                    setSubjectId(v ?? "");
                    setChapterId(NONE_CHAPTER_VALUE);
                  }}
                  disabled={!examId}
                >
                  <SelectTrigger>
                    {selectedSubject?.name ?? "选择科目"}
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
                <Label>章节（可选）</Label>
                <Select
                  value={chapterId}
                  onValueChange={(v) => setChapterId(v ?? NONE_CHAPTER_VALUE)}
                  disabled={!subjectId}
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
            <div className="flex justify-end gap-3 pt-2">
              <Button
                variant="outline"
                type="button"
                onClick={() => onClose()}
              >
                取消
              </Button>
              <Button
                type="button"
                onClick={() => setStep("edit")}
                disabled={!examId || !subjectId}
              >
                下一步
              </Button>
            </div>
          </div>
        ) : (
          <ScrollArea className="max-h-[60vh]">
            <div className="space-y-5 px-1 py-2">
              <div className="space-y-2">
                <Label>题干</Label>
                <Textarea
                  value={stem}
                  onChange={(e) => setStem(e.target.value)}
                  className="min-h-[80px]"
                  placeholder="请输入题干内容"
                />
              </div>
              {(questionType === "single_choice" ||
                questionType === "multiple_choice") && (
                <div className="space-y-2">
                  <Label>选项</Label>
                  {opts.map((opt, i) => (
                    <div key={i} className="flex items-center gap-2">
                      <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium">
                        {String.fromCharCode(65 + i)}
                      </span>
                      <Input
                        value={opt}
                        onChange={(e) => {
                          const n = [...opts];
                          n[i] = e.target.value;
                          setOpts(n);
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
                        {opts.map((_, i) => (
                          <SelectItem
                            key={i}
                            value={String.fromCharCode(65 + i)}
                          >
                            {String.fromCharCode(65 + i)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              )}
              {questionType === "judgment" && (
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
              <div className="flex items-center justify-end gap-3 pt-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setStep("select")}
                >
                  上一步
                </Button>
                <Button
                  type="button"
                  onClick={handleCreate}
                  disabled={saving || !stem.trim()}
                >
                  {saving ? "创建中..." : "创建题目"}
                </Button>
              </div>
            </div>
          </ScrollArea>
        )}
      </DialogContent>
    </Dialog>
  );
}
