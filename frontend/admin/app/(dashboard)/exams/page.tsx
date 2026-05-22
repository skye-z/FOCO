"use client";

import * as React from "react";
import {
  AlertTriangle,
  BookOpen,
  Filter,
  FlaskConical,
  Layers,
  LoaderCircle,
  Menu,
  Plus,
  Trash2,
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import {
  readBrowserAccessToken,
  clearStoredSession,
  authFetch,
} from "@/lib/auth-session";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogMedia,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { KnowledgeGraphModal } from "@/components/knowledge-graph-modal";
import { toast } from "sonner";

import {
  type TreeNode,
  type KnowledgePoint,
  type QuestionCard,
  type QuestionVersionSummary,
  type VersionDetail,
  type InteractiveUnitSummary,
  DIFFICULTY_OPTIONS,
  difficultyLabel,
  difficultyColor,
  statusLabel,
  typeLabel,
  typeColor,
  questionStatusLabel,
  questionStatusColor,
  showActionError,
} from "./types";
import { TreeLevel } from "./components/tree-level";
import { QuestionEditForm } from "./components/question-edit-form";
import { CreateNodeDialog } from "./components/create-node-dialog";
import { EditNodeDialog } from "./components/edit-node-dialog";
import { CreateQuestionDialog } from "./components/create-question-dialog";
import { InteractiveUnitCard } from "./components/interactive-unit-card";
import { CreateInteractiveUnitDialog } from "./components/create-interactive-unit-dialog";

export default function ExamsPage() {
  const router = useRouter();
  const [tree, setTree] = React.useState<TreeNode[]>([]);
  const [questions, setQuestions] = React.useState<QuestionCard[]>([]);
  const [knowledgePoints, setKnowledgePoints] = React.useState<
    KnowledgePoint[]
  >([]);
  const [loading, setLoading] = React.useState(true);
  const [questionsLoading, setQuestionsLoading] = React.useState(false);
  const [expandedIds, setExpandedIds] = React.useState<Set<string>>(new Set());
  const [selectedNode, setSelectedNode] = React.useState<{
    type: string;
    id: string;
  } | null>(null);
  const [filterDifficulty, setFilterDifficulty] = React.useState("all");
  const [filterStatus, setFilterStatus] = React.useState("all");
  const [filterKP, setFilterKP] = React.useState("all");
  const [sidebarOpen, setSidebarOpen] = React.useState(true);

  const [editOpen, setEditOpen] = React.useState(false);
  const [detailLoading, setDetailLoading] = React.useState(false);
  const [detail, setDetail] = React.useState<VersionDetail | null>(null);
  const [versionHistory, setVersionHistory] = React.useState<
    QuestionVersionSummary[]
  >([]);
  const [historyLoading, setHistoryLoading] = React.useState(false);
  const [graphOpen, setGraphOpen] = React.useState(false);
  const [questionPage, setQuestionPage] = React.useState(1);

  const [createDialogType, setCreateDialogType] = React.useState<
    "exam" | "subject" | "chapter" | null
  >(null);
  const [createParentId, setCreateParentId] = React.useState("");
  const [editNode, setEditNode] = React.useState<{
    type: "exam" | "subject" | "chapter";
    id: string;
    name: string;
  } | null>(null);
  const [createQuestionOpen, setCreateQuestionOpen] = React.useState(false);
  const [createInteractiveUnitOpen, setCreateInteractiveUnitOpen] =
    React.useState(false);

  const [deleteConfirm, setDeleteConfirm] = React.useState<{
    type: "exam" | "subject" | "chapter" | "question" | "interactive";
    id: string;
    name: string;
  } | null>(null);
  const [deleting, setDeleting] = React.useState(false);
  const [publishingVersionId, setPublishingVersionId] = React.useState<
    string | null
  >(null);

  const [interactiveUnits, setInteractiveUnits] = React.useState<
    InteractiveUnitSummary[]
  >([]);
  const [interactiveLoading, setInteractiveLoading] = React.useState(false);
  const [filterType, setFilterType] = React.useState<
    "all" | "question" | "interactive"
  >("all");

  const tokenRef = React.useRef<string | null>(null);

  function refreshTree() {
    if (!tokenRef.current) return;
    authFetch("/api/v1/admin/exam-tree", {
      headers: { Authorization: `Bearer ${tokenRef.current}` },
      cache: "no-store",
    })
      .then((r) => r.json())
      .then((p) => setTree(p.data ?? []))
      .catch(() => {});
  }

  function refreshQuestions() {
    if (!tokenRef.current) return;
    setQuestionsLoading(true);
    const params = new URLSearchParams();
    if (selectedNode?.type === "exam") params.set("exam_id", selectedNode.id);
    if (selectedNode?.type === "subject")
      params.set("subject_id", selectedNode.id);
    if (selectedNode?.type === "chapter")
      params.set("chapter_id", selectedNode.id);
    if (filterDifficulty !== "all") params.set("difficulty", filterDifficulty);
    if (filterStatus !== "all") params.set("status", filterStatus);
    if (filterKP !== "all") params.set("knowledge_point_id", filterKP);
    const qs = params.toString();
    authFetch(`/api/v1/admin/questions${qs ? `?${qs}` : ""}`, {
      headers: { Authorization: `Bearer ${tokenRef.current}` },
      cache: "no-store",
    })
      .then((r) => r.json())
      .then((p) => {
        setQuestions(p.data ?? []);
        setQuestionsLoading(false);
      })
      .catch(() => {
        setQuestions([]);
        setQuestionsLoading(false);
      });
  }

  function refreshInteractiveUnits() {
    if (!tokenRef.current) return;
    setInteractiveLoading(true);
    const params = new URLSearchParams();
    if (selectedNode?.type === "exam") params.set("exam_id", selectedNode.id);
    if (selectedNode?.type === "subject")
      params.set("subject_id", selectedNode.id);
    const qs = params.toString();
    authFetch(`/api/v1/admin/interactive-units${qs ? `?${qs}` : ""}`, {
      headers: { Authorization: `Bearer ${tokenRef.current}` },
      cache: "no-store",
    })
      .then((r) => r.json())
      .then((p) => {
        setInteractiveUnits(p.data ?? []);
        setInteractiveLoading(false);
      })
      .catch(() => {
        setInteractiveUnits([]);
        setInteractiveLoading(false);
      });
  }

  async function refreshDetail(versionId: string) {
    if (!tokenRef.current) return;
    setDetailLoading(true);
    setDetail(null);
    try {
      const res = await authFetch(
        `/api/v1/admin/question-versions/${versionId}`,
        {
          headers: { Authorization: `Bearer ${tokenRef.current}` },
          cache: "no-store",
        },
      );
      if (res.ok) {
        const p = await res.json();
        setDetail(p.data);
      } else {
        showActionError(
          "题目详情加载失败",
          "当前版本详情读取失败，请稍后重试。",
        );
      }
    } finally {
      setDetailLoading(false);
    }
  }

  async function refreshVersionHistory(questionId: string) {
    if (!tokenRef.current) return;
    setHistoryLoading(true);
    try {
      const res = await authFetch(
        `/api/v1/admin/questions/${questionId}/versions`,
        {
          headers: { Authorization: `Bearer ${tokenRef.current}` },
          cache: "no-store",
        },
      );
      if (res.ok) {
        const payload = await res.json();
        setVersionHistory((payload.data ?? []) as QuestionVersionSummary[]);
      }
    } finally {
      setHistoryLoading(false);
    }
  }

  React.useEffect(() => {
    let cancelled = false;
    async function init() {
      const token = readBrowserAccessToken();
      if (!token) {
        clearStoredSession();
        router.replace("/");
        return;
      }
      tokenRef.current = token;
      try {
        const [treeRes, kpRes] = await Promise.all([
          authFetch("/api/v1/admin/exam-tree", {
            headers: { Authorization: `Bearer ${token}` },
            cache: "no-store",
          }),
          authFetch("/api/v1/admin/knowledge-points", {
            headers: { Authorization: `Bearer ${token}` },
            cache: "no-store",
          }),
        ]);
        if (!cancelled && treeRes.ok) {
          const p = await treeRes.json();
          setTree(p.data ?? []);
        }
        if (!cancelled && kpRes.ok) {
          const p = await kpRes.json();
          setKnowledgePoints(p.data ?? []);
        }
      } catch {
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void init();
    return () => {
      cancelled = true;
    };
  }, [router]);

  React.useEffect(() => {
    refreshQuestions();
    refreshInteractiveUnits();
  }, [selectedNode, filterDifficulty, filterStatus, filterKP]);

  React.useEffect(() => {
    setQuestionPage(1);
  }, [selectedNode, filterDifficulty, filterStatus, filterKP]);

  const pageSize = 20;

  type WaterfallItem =
    | { kind: "question"; data: QuestionCard }
    | { kind: "interactive"; data: InteractiveUnitSummary };

  const totalQuestionPages = Math.max(
    1,
    Math.ceil(questions.length / pageSize),
  );
  const pagedQuestions = React.useMemo(() => {
    const start = (questionPage - 1) * pageSize;
    return questions.slice(start, start + pageSize);
  }, [questionPage, questions]);
  const visibleQuestionPages = React.useMemo(() => {
    const maxButtons = 5;
    let start = Math.max(1, questionPage - 2);
    let end = Math.min(totalQuestionPages, start + maxButtons - 1);
    start = Math.max(1, end - maxButtons + 1);
    return Array.from({ length: end - start + 1 }, (_, index) => start + index);
  }, [questionPage, totalQuestionPages]);

  const waterfallItems = React.useMemo<WaterfallItem[]>(() => {
    const items: WaterfallItem[] = [
      ...interactiveUnits.map(
        (u) => ({ kind: "interactive" as const, data: u }) satisfies WaterfallItem,
      ),
      ...pagedQuestions.map(
        (q) => ({ kind: "question" as const, data: q }) satisfies WaterfallItem,
      ),
    ];
    if (filterType === "question") return items.filter((i) => i.kind === "question");
    if (filterType === "interactive")
      return items.filter((i) => i.kind === "interactive");
    return items;
  }, [pagedQuestions, interactiveUnits, filterType]);

  function toggleExpand(id: string) {
    setExpandedIds((prev) => {
      const n = new Set(prev);
      if (n.has(id)) n.delete(id);
      else n.add(id);
      return n;
    });
  }

  async function openDetail(card: QuestionCard) {
    if (!tokenRef.current) return;
    setDetailLoading(true);
    setEditOpen(true);
    try {
      const res = await authFetch(
        `/api/v1/admin/question-versions/${card.version_id}`,
        {
          headers: { Authorization: `Bearer ${tokenRef.current}` },
        },
      );
      if (res.ok) {
        const p = await res.json();
        setDetail(p.data);
        await refreshVersionHistory(p.data.question_id);
      }
    } catch {
    } finally {
      setDetailLoading(false);
    }
  }

  async function handlePublish(versionId: string) {
    if (!tokenRef.current) return;
    setPublishingVersionId(versionId);
    try {
      const res = await authFetch(
        `/api/v1/admin/question-versions/${versionId}/publish`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${tokenRef.current}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ publish_note: "通过管理端发布" }),
        },
      );
      if (!res.ok) {
        showActionError("发布失败", "版本发布未成功，请稍后重试。");
        return;
      }
      setEditOpen(false);
      toast.success("版本已发布");
      refreshQuestions();
      if (detail) {
        await refreshVersionHistory(detail.question_id);
      }
    } catch {
      showActionError("发布失败", "版本发布未成功，请稍后重试。");
    } finally {
      setPublishingVersionId(null);
    }
  }

  async function handleOpenVersion(versionId: string) {
    await refreshDetail(versionId);
  }

  async function handleRestoreVersion(versionId: string) {
    if (!tokenRef.current) return;
    try {
      const res = await authFetch(
        `/api/v1/admin/question-versions/${versionId}/restore`,
        {
          method: "POST",
          headers: { Authorization: `Bearer ${tokenRef.current}` },
        },
      );
      if (!res.ok) {
        showActionError("恢复草稿失败", "历史版本恢复未完成，请稍后重试。");
        return;
      }
      const payload = await res.json();
      setDetail(payload.data as VersionDetail);
      if (payload.data?.question_id) {
        await refreshVersionHistory(payload.data.question_id as string);
      }
      await refreshQuestions();
      toast.success("已恢复为新草稿");
    } catch {
      showActionError("恢复草稿失败", "历史版本恢复未完成，请稍后重试。");
    }
  }

  function shouldResetSelectionOnDelete(target: {
    type: "exam" | "subject" | "chapter" | "question" | "interactive";
    id: string;
  }) {
    if (!selectedNode) return false;
    if (selectedNode.id === target.id) return true;
    if (target.type === "exam") {
      const exam = tree.find((node) => node.id === target.id);
      if (!exam) return false;
      return (
        exam.children?.some(
          (subject) =>
            subject.id === selectedNode.id ||
            subject.children?.some((chapter) => chapter.id === selectedNode.id),
        ) ?? false
      );
    }
    if (target.type === "subject") {
      const subject = tree
        .flatMap((node) => node.children ?? [])
        .find((node) => node.id === target.id);
      if (!subject) return false;
      return (
        subject.children?.some((chapter) => chapter.id === selectedNode.id) ??
        false
      );
    }
    return false;
  }

  async function handleDelete() {
    if (!tokenRef.current || !deleteConfirm) return;
    setDeleting(true);
    const endpoint =
      deleteConfirm.type === "exam"
        ? `/api/v1/admin/exams/${deleteConfirm.id}`
        : deleteConfirm.type === "subject"
          ? `/api/v1/admin/subjects/${deleteConfirm.id}`
          : deleteConfirm.type === "chapter"
            ? `/api/v1/admin/chapters/${deleteConfirm.id}`
            : deleteConfirm.type === "interactive"
              ? `/api/v1/admin/interactive-units/${deleteConfirm.id}`
              : `/api/v1/admin/questions/${deleteConfirm.id}`;
    try {
      const res = await authFetch(endpoint, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${tokenRef.current}` },
      });
      if (res.status === 409) {
        const data = await res.json();
        showActionError(
          "无法删除",
          data.error || "存在关联数据，请先清理关联题目后再删除。",
        );
      } else if (!res.ok) {
        showActionError("删除失败", "删除操作未完成，请稍后重试。");
      } else {
        if (
          deleteConfirm.type === "question" &&
          detail?.question_id === deleteConfirm.id
        ) {
          setEditOpen(false);
          setDetail(null);
        }
        if (shouldResetSelectionOnDelete(deleteConfirm)) {
          setSelectedNode(null);
        }
        if (deleteConfirm.type === "question") refreshQuestions();
        else if (deleteConfirm.type === "interactive") refreshInteractiveUnits();
        else refreshTree();
        toast.success("删除成功");
      }
    } catch {
      showActionError("删除失败", "删除操作未完成，请稍后重试。");
    } finally {
      setDeleting(false);
      setDeleteConfirm(null);
    }
  }

  function openCreateDialog(
    type: "exam" | "subject" | "chapter",
    parentId?: string,
  ) {
    setCreateDialogType(type);
    setCreateParentId(parentId ?? "");
  }

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="relative flex">
      <aside
        className={`fixed top-14 bottom-0 left-0 z-30 w-72 border-r bg-background transition-transform lg:translate-x-0 ${sidebarOpen ? "translate-x-0" : "-translate-x-full"}`}
      >
        <div className="flex items-center justify-between p-4">
          <h2 className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
            <Layers className="size-4" />
            知识体系
          </h2>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 px-2 text-xs"
              onClick={() => setGraphOpen(true)}
            >
              可视化
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-7"
              title="添加考试"
              onClick={() => openCreateDialog("exam")}
            >
              <Plus className="size-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-7 lg:hidden"
              onClick={() => setSidebarOpen(false)}
            >
              <X className="size-4" />
            </Button>
          </div>
        </div>
        <ScrollArea className="h-[calc(100vh-7.5rem)]">
          <div className="px-2 pb-4">
            {tree.length === 0 ? (
              <p className="px-3 py-8 text-center text-sm text-muted-foreground">
                暂无考试数据
              </p>
            ) : (
              tree.map((exam) => (
                <TreeLevel
                  key={exam.id}
                  node={exam}
                  expandedIds={expandedIds}
                  selectedId={selectedNode?.id}
                  depth={0}
                  onToggle={toggleExpand}
                  onSelect={(n) => setSelectedNode({ type: n.type, id: n.id })}
                  onAdd={(type, parentId) => openCreateDialog(type, parentId)}
                  onEdit={(node) =>
                    setEditNode({ type: node.type, id: node.id, name: node.name })
                  }
                  onDelete={(type, id, name) =>
                    setDeleteConfirm({
                      type: type as "exam" | "subject" | "chapter",
                      id,
                      name,
                    })
                  }
                />
              ))
            )}
          </div>
        </ScrollArea>
      </aside>

      <main
        className={`min-w-0 flex-1 transition-all lg:ml-72 ${sidebarOpen ? "ml-72" : "ml-0"}`}
      >
        <div className="mx-auto max-w-5xl p-6">
          <div className="mb-6 flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold tracking-tight">题库管理</h1>
              <p className="mt-1 text-sm text-muted-foreground">
                {selectedNode
                  ? `当前选择: ${selectedNode.type === "exam" ? "考试" : selectedNode.type === "subject" ? "科目" : "章节"}`
                  : "选择左侧节点筛选题目，或浏览全部题目"}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                className="lg:hidden"
                onClick={() => setSidebarOpen(true)}
              >
                <Menu className="mr-1 size-4" />
                目录
              </Button>
              <Button size="sm" onClick={() => setCreateQuestionOpen(true)}>
                <Plus className="mr-1 size-4" />
                新建题目
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setCreateInteractiveUnitOpen(true)}
              >
                <FlaskConical className="mr-1 size-4" />
                新建交互单元
              </Button>
            </div>
          </div>

          <div className="mb-6 flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
              <Filter className="size-4" />
              筛选
            </div>
            <Select
              value={filterDifficulty}
              onValueChange={(v) => setFilterDifficulty(v ?? "all")}
            >
              <SelectTrigger className="w-28">
                {filterDifficulty === "all"
                  ? "全部难度"
                  : difficultyLabel(parseInt(filterDifficulty))}
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部难度</SelectItem>
                {DIFFICULTY_OPTIONS.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={filterStatus}
              onValueChange={(v) => setFilterStatus(v ?? "all")}
            >
              <SelectTrigger className="w-28">
                {filterStatus === "all"
                  ? "全部状态"
                  : statusLabel(filterStatus)}
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="draft">草稿</SelectItem>
                <SelectItem value="published">已发布</SelectItem>
                <SelectItem value="archived">已归档</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={filterKP}
              onValueChange={(v) => setFilterKP(v ?? "all")}
            >
              <SelectTrigger className="w-44">
                {filterKP === "all"
                  ? "全部知识点"
                  : (knowledgePoints.find((kp) => kp.id === filterKP)?.name ??
                    "全部知识点")}
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部知识点</SelectItem>
                {knowledgePoints.map((kp) => (
                  <SelectItem key={kp.id} value={kp.id}>
                    {kp.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={filterType}
              onValueChange={(v) => setFilterType((v ?? "all") as "all" | "question" | "interactive")}
            >
              <SelectTrigger className="w-28">
                {filterType === "all"
                  ? "全部类型"
                  : filterType === "question"
                    ? "题目"
                    : "交互单元"}
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部类型</SelectItem>
                <SelectItem value="question">题目</SelectItem>
                <SelectItem value="interactive">交互单元</SelectItem>
              </SelectContent>
            </Select>
            {(selectedNode ||
              filterDifficulty !== "all" ||
              filterStatus !== "all" ||
              filterKP !== "all" ||
              filterType !== "all") && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setSelectedNode(null);
                  setFilterDifficulty("all");
                  setFilterStatus("all");
                  setFilterKP("all");
                  setFilterType("all");
                }}
              >
                清除筛选
              </Button>
            )}
          </div>

          {questionsLoading || interactiveLoading ? (
            <div className="flex items-center justify-center py-20">
              <div className="flex flex-col items-center gap-3">
                <LoaderCircle className="size-8 animate-spin text-primary" />
                <p className="text-sm text-muted-foreground">加载中…</p>
              </div>
            </div>
          ) : waterfallItems.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
              <FlaskConical className="mb-3 size-10" />
              <p className="text-sm">暂无数据</p>
            </div>
          ) : (
            <>
              <div className="columns-1 gap-4 sm:columns-2 lg:columns-3">
                {waterfallItems.map((item) =>
                  item.kind === "question" ? (
                    <Card
                      key={item.data.id}
                      className="group mb-4 cursor-pointer break-inside-avoid transition-all hover:shadow-md hover:-translate-y-0.5"
                      onClick={() => openDetail(item.data)}
                    >
                      <CardContent>
                        <div className="mb-3 flex items-center justify-between gap-2">
                          <div className="flex flex-wrap items-center gap-1.5">
                            <Badge
                              variant="outline"
                              className={`text-xs ${typeColor(item.data.question_type)}`}
                            >
                              {typeLabel(item.data.question_type)}
                            </Badge>
                            <Badge
                              variant="outline"
                              className={`text-xs ${questionStatusColor(item.data)}`}
                            >
                              {questionStatusLabel(item.data)}
                            </Badge>
                            <Badge
                              variant="outline"
                              className={`text-xs ${difficultyColor(item.data.difficulty)}`}
                            >
                              {difficultyLabel(item.data.difficulty)}
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
                                      onClick={(
                                        event: React.MouseEvent<HTMLButtonElement>,
                                      ) => event.stopPropagation()}
                                    />
                                  }
                                >
                                  <Menu className="size-3.5" />
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                  align="end"
                                  side="bottom"
                                  sideOffset={6}
                                >
                                  <DropdownMenuGroup>
                                    <DropdownMenuItem
                                      variant="destructive"
                                      onClick={(event) => {
                                        event.stopPropagation();
                                        setDeleteConfirm({
                                          type: "question",
                                          id: item.data.id,
                                          name: item.data.stem_preview.slice(0, 30),
                                        });
                                      }}
                                    >
                                      <Trash2 />
                                      删除题目
                                    </DropdownMenuItem>
                                  </DropdownMenuGroup>
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </div>
                          </div>
                        </div>
                        <p className="mb-3 line-clamp-3 text-sm leading-relaxed">
                          {item.data.stem_preview || "（无题干内容）"}
                        </p>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <BookOpen className="size-3" />
                          <span>{item.data.subject_name}</span>
                          {item.data.chapter_name && (
                            <>
                              <span>/</span>
                              <span>{item.data.chapter_name}</span>
                            </>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  ) : (
                    <InteractiveUnitCard
                      key={item.data.id}
                      unit={item.data}
                      onDelete={(u) =>
                        setDeleteConfirm({
                          type: "interactive",
                          id: u.id,
                          name: u.title?.slice(0, 30) ?? "交互单元",
                        })
                      }
                    />
                  ),
                )}
              </div>
              <div className="mt-6 flex items-center justify-between gap-3 border-t pt-4 text-sm text-muted-foreground">
                <span>
                  第 {questionPage} / {totalQuestionPages} 页，共{" "}
                  {waterfallItems.length} 项
                </span>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setQuestionPage((page) => Math.max(1, page - 1))
                    }
                    disabled={questionPage === 1}
                  >
                    上一页
                  </Button>
                  <div className="flex items-center gap-1">
                    {visibleQuestionPages.map((page) => (
                      <Button
                        key={page}
                        variant={page === questionPage ? "default" : "outline"}
                        size="sm"
                        className="min-w-9"
                        onClick={() => setQuestionPage(page)}
                      >
                        {page}
                      </Button>
                    ))}
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      setQuestionPage((page) =>
                        Math.min(totalQuestionPages, page + 1),
                      )
                    }
                    disabled={questionPage === totalQuestionPages}
                  >
                    下一页
                  </Button>
                </div>
              </div>
            </>
          )}
        </div>
      </main>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="max-h-[92vh] w-[96vw] max-w-[96vw] sm:!max-w-7xl">
          <DialogHeader>
            <DialogTitle>题目详情</DialogTitle>
          </DialogHeader>
          {detailLoading ? (
            <div className="flex items-center justify-center py-12">
              <LoaderCircle className="size-6 animate-spin text-primary" />
            </div>
          ) : detail ? (
            <QuestionEditForm
              detail={detail}
              tree={tree}
              knowledgePoints={knowledgePoints}
              token={tokenRef.current}
              versions={versionHistory}
              historyLoading={historyLoading}
              onSave={async (savedDetail) => {
                setDetail(savedDetail);
                await refreshQuestions();
                await refreshVersionHistory(savedDetail.question_id);
              }}
              onCancel={() => setEditOpen(false)}
              onOpenVersion={handleOpenVersion}
              onRestoreVersion={handleRestoreVersion}
              onPublish={() => handlePublish(detail.version_id)}
              publishing={publishingVersionId === detail.version_id}
            />
          ) : (
            <div className="rounded-lg border border-destructive/20 bg-destructive/5 px-4 py-6 text-sm text-destructive">
              题目详情未成功加载，请关闭后重试。
            </div>
          )}
        </DialogContent>
      </Dialog>

      <KnowledgeGraphModal
        open={graphOpen}
        onClose={() => setGraphOpen(false)}
      />

      {/* Create Node Dialog */}
      <CreateNodeDialog
        type={createDialogType}
        parentId={createParentId}
        tree={tree}
        token={tokenRef.current}
        onClose={() => setCreateDialogType(null)}
        onCreated={() => {
          refreshTree();
          setCreateDialogType(null);
        }}
      />

      <EditNodeDialog
        open={editNode !== null}
        node={editNode}
        token={tokenRef.current}
        onClose={() => setEditNode(null)}
        onSaved={() => {
          refreshTree();
          setEditNode(null);
        }}
      />

      {/* Create Question Dialog */}
      <CreateQuestionDialog
        open={createQuestionOpen}
        tree={tree}
        token={tokenRef.current}
        onClose={() => setCreateQuestionOpen(false)}
        onCreated={(card: QuestionCard) => {
          setCreateQuestionOpen(false);
          openDetail(card);
        }}
      />

      {/* Create Interactive Unit Dialog */}
      <CreateInteractiveUnitDialog
        open={createInteractiveUnitOpen}
        tree={tree}
        token={tokenRef.current}
        onClose={() => setCreateInteractiveUnitOpen(false)}
      />

      {/* Delete Confirm Dialog */}
      <AlertDialog
        open={deleteConfirm !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteConfirm(null);
        }}
      >
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogMedia>
              <AlertTriangle className="size-5 text-amber-500" />
            </AlertDialogMedia>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              <span className="text-foreground">
                确定要删除「{deleteConfirm?.name}」吗？
              </span>
              {deleteConfirm?.type !== "question" &&
              deleteConfirm?.type !== "interactive"
                ? " 如果该节点下仍有关联题目，需要先删除题目。"
                : " 该操作不可撤销。"}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleting}
              variant="destructive"
            >
              {deleting ? "删除中..." : "确认删除"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
