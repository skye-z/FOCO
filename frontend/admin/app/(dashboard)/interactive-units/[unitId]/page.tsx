"use client";

import * as React from "react";
import { LoaderCircle } from "lucide-react";
import { useRouter, useParams, useSearchParams } from "next/navigation";
import {
  authFetch,
  readBrowserAccessToken,
  clearStoredSession,
} from "@/lib/auth-session";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { KpMultiSelect } from "@/app/(dashboard)/exams/components/kp-multi-select";
import { statusLabel, type KnowledgePoint } from "@/app/(dashboard)/exams/types";
import { StepList, type StepSchema } from "@/components/interactive/step-list";
import { StepConfigForm } from "@/components/interactive/step-config-form";
import { StepPreview } from "@/components/interactive/step-preview";
import { FlowCanvas } from "@/components/interactive/flow-canvas";
import { createDefaultVisualStep, type VisualBlockType } from "@/components/interactive/block-types";
import { ResultSummaryPanel } from "@/components/interactive/result-model";
import {
  createEditorSnapshot,
  hasUnsavedChanges,
  summarizePublishReadiness,
} from "@/components/interactive/editor-status";

type AdminVersionDetail = {
  version_id: string;
  unit_id: string;
  version_no: number;
  status: string;
  title: string;
  steps: StepSchema[];
};

type AdminVersionSummary = {
  version_id: string;
  unit_id: string;
  version_no: number;
  status: string;
  published_at?: string | null;
  updated_at: string;
};

function stripQuotes(s: string): string {
  if (s.length >= 2 && s.startsWith('"') && s.endsWith('"')) {
    return s.slice(1, -1);
  }
  return s;
}

export default function InteractiveUnitEditorPage() {
  const router = useRouter();
  const params = useParams<{ unitId: string }>();
  const searchParams = useSearchParams();
  const unitId = params.unitId;
  const examId = searchParams.get("exam_id") ?? "";

  const [loading, setLoading] = React.useState(true);
  const [versionDetail, setVersionDetail] =
    React.useState<AdminVersionDetail | null>(null);
  const [versions, setVersions] = React.useState<AdminVersionSummary[]>([]);
  const [localSteps, setLocalSteps] = React.useState<StepSchema[]>([]);
  const [title, setTitle] = React.useState("");
  const [selectedStepIndex, setSelectedStepIndex] = React.useState(0);
  const [saving, setSaving] = React.useState(false);
  const [publishing, setPublishing] = React.useState(false);
  const [knowledgePoints, setKnowledgePoints] = React.useState<KnowledgePoint[]>([]);
  const [loadedSnapshot, setLoadedSnapshot] = React.useState(() =>
    createEditorSnapshot("", []),
  );

  const tokenRef = React.useRef<string | null>(null);
  const publishReadiness = React.useMemo(
    () => summarizePublishReadiness(localSteps, title),
    [localSteps, title],
  );
  const isDirty = React.useMemo(
    () => hasUnsavedChanges(loadedSnapshot.title, loadedSnapshot.steps, title, localSteps),
    [loadedSnapshot, localSteps, title],
  );
  const isReadOnly = versionDetail?.status === "published";

  async function loadVersions() {
    if (!tokenRef.current) return;
    try {
      const res = await authFetch(
        `/api/v1/admin/interactive-units/${unitId}/versions`,
        {
          headers: { Authorization: `Bearer ${tokenRef.current}` },
          cache: "no-store",
        },
      );
      if (res.ok) {
        const p = await res.json();
        const vers: AdminVersionSummary[] = p.data ?? [];
        setVersions(vers);
        return vers;
      }
    } catch {}
    return [];
  }

  async function loadKnowledgePoints() {
    if (!tokenRef.current || !examId) return;
    try {
      const res = await authFetch(`/api/v1/admin/knowledge-points?exam_id=${encodeURIComponent(examId)}`, {
        headers: { Authorization: `Bearer ${tokenRef.current}` },
        cache: "no-store",
      });
      if (res.ok) {
        const p = await res.json();
        setKnowledgePoints((p.data ?? []) as KnowledgePoint[]);
      }
    } catch {}
  }

  async function loadVersionDetail(versionId: string) {
    if (!tokenRef.current || !versionId) return null;
    try {
      const res = await authFetch(
        `/api/v1/admin/interactive-unit-versions/${versionId}`,
        {
          headers: { Authorization: `Bearer ${tokenRef.current}` },
          cache: "no-store",
        },
      );
      if (res.ok) {
        const p = await res.json();
        const detail: AdminVersionDetail = p.data;
        const cleanTitle = stripQuotes(detail.title ?? "");
        setVersionDetail(detail);
        setLocalSteps(detail.steps ?? []);
        setTitle(cleanTitle);
        setLoadedSnapshot(createEditorSnapshot(cleanTitle, detail.steps ?? []));
        return detail;
      }
    } catch {}
    return null;
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
      await loadKnowledgePoints();

      const vers = await loadVersions();
      if (cancelled || !vers) return;

      const draft = vers.find((v) => v.status === "draft");
      const target = draft ?? vers[0];
      if (target?.version_id) {
        await loadVersionDetail(target.version_id);
      } else if (target) {
        toast.error("交互单元版本加载失败", {
          description: "当前版本缺少有效 ID，请刷新后重试。",
        });
      }
      if (!cancelled) setLoading(false);
    }
    void init();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router, unitId]);

  React.useEffect(() => {
    function handleBeforeUnload(event: BeforeUnloadEvent) {
      if (!isDirty) return;
      event.preventDefault();
      event.returnValue = "";
    }

    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => window.removeEventListener("beforeunload", handleBeforeUnload);
  }, [isDirty]);

  function confirmDiscardChanges(message: string) {
    if (!isDirty) return true;
    return window.confirm(message);
  }

  function handleReorder(oldIndex: number, newIndex: number) {
    if (isReadOnly) return;
    setLocalSteps((prev) => {
      const next = [...prev];
      const [moved] = next.splice(oldIndex, 1);
      next.splice(newIndex, 0, moved);
      return next;
    });
    setSelectedStepIndex(newIndex);
  }

  function handleAddBlock(blockType: VisualBlockType, insertAt?: number) {
    if (isReadOnly) return;
    const newStep: StepSchema = createDefaultVisualStep(blockType);
    setLocalSteps((prev) => {
      if (insertAt === undefined || insertAt < 0 || insertAt > prev.length) {
        return [...prev, newStep];
      }
      return [...prev.slice(0, insertAt), newStep, ...prev.slice(insertAt)];
    });
    setSelectedStepIndex(insertAt ?? localSteps.length);
  }

  function handleDeleteStep(index: number) {
    if (isReadOnly) return;
    setLocalSteps((prev) => prev.filter((_, i) => i !== index));
    setSelectedStepIndex((prev) =>
      prev >= localSteps.length - 1 ? Math.max(0, localSteps.length - 2) : prev,
    );
  }

  function handleStepChange(updated: StepSchema) {
    if (isReadOnly) return;
    setLocalSteps((prev) =>
      prev.map((s, i) => (i === selectedStepIndex ? updated : s)),
    );
  }

  function toggleKnowledgePoint(id: string) {
    if (!currentStep || isReadOnly) return;
    const current = new Set(currentStep.knowledge_point_ids ?? []);
    if (current.has(id)) current.delete(id);
    else current.add(id);
    const nextIds = Array.from(current);
    const nextTags = nextIds
      .map((kpId) => knowledgePoints.find((item) => item.id === kpId)?.name)
      .filter((value): value is string => Boolean(value));
    handleStepChange({
      ...currentStep,
      knowledge_point_ids: nextIds,
      knowledge_point_tags: nextTags,
    });
  }

  async function handleSave() {
    if (!tokenRef.current || !versionDetail || isReadOnly || !isDirty) return;
    setSaving(true);
    try {
      const res = await authFetch(
        `/api/v1/admin/interactive-unit-versions/${versionDetail.version_id}`,
        {
          method: "PATCH",
          headers: {
            Authorization: `Bearer ${tokenRef.current}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ title, steps: localSteps }),
        },
      );
      if (res.ok) {
        const p = await res.json();
        setVersionDetail(p.data);
        setLocalSteps(p.data.steps ?? []);
        setLoadedSnapshot(
          createEditorSnapshot(p.data.title ?? "", p.data.steps ?? []),
        );
        toast.success("已保存");
      } else {
        toast.error("保存失败");
      }
    } catch {
      toast.error("保存失败");
    } finally {
      setSaving(false);
    }
  }

  async function handlePublish() {
    if (!tokenRef.current || !versionDetail) return;
    if (!publishReadiness.canPublish) {
      toast.error("当前版本暂不可发布", {
        description:
          publishReadiness.missing[0] ??
          (publishReadiness.incompleteSteps[0]
            ? `请先完善步骤 ${publishReadiness.incompleteSteps[0].index + 1}`
            : "请补全必填信息后重试。"),
      });
      return;
    }
    setPublishing(true);
    try {
      const res = await authFetch(
        `/api/v1/admin/interactive-unit-versions/${versionDetail.version_id}/publish`,
        {
          method: "POST",
          headers: { Authorization: `Bearer ${tokenRef.current}` },
        },
      );
      if (res.ok) {
        toast.success("已发布");
        await loadVersions();
      } else {
        toast.error("发布失败");
      }
    } catch {
      toast.error("发布失败");
    } finally {
      setPublishing(false);
    }
  }

  async function handleCreateVersion() {
    if (!tokenRef.current) return;
    if (
      !confirmDiscardChanges(
        "当前有未保存更改，创建新版本前将放弃这些修改。是否继续？",
      )
    ) {
      return;
    }
    try {
      const res = await authFetch(
        `/api/v1/admin/interactive-units/${unitId}/versions`,
        {
          method: "POST",
          headers: { Authorization: `Bearer ${tokenRef.current}` },
        },
      );
      if (!res.ok) {
        toast.error("新建版本失败");
        return;
      }
      const p = await res.json();
      toast.success("已创建新版本");
      await loadVersions();
      if (p.data?.version_id) {
        await loadVersionDetail(p.data.version_id);
        setSelectedStepIndex(0);
      }
    } catch {
      toast.error("新建版本失败");
    }
  }

  async function loadVersion(versionId: string) {
    if (!versionId) {
      toast.error("交互单元版本加载失败", {
        description: "当前版本缺少有效 ID，请刷新后重试。",
      });
      return;
    }
    if (
      versionId !== versionDetail?.version_id &&
      !confirmDiscardChanges("当前有未保存更改，切换版本会丢失本地修改。是否继续？")
    ) {
      return;
    }
    await loadVersionDetail(versionId);
    setSelectedStepIndex(0);
  }

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const currentStep =
    selectedStepIndex >= 0 && selectedStepIndex < localSteps.length
      ? localSteps[selectedStepIndex]
      : null;

  return (
    <div className="flex h-[calc(100vh-3.5rem)] flex-col">
      <div className="flex items-center justify-between border-b px-6 py-3">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              if (
                !confirmDiscardChanges(
                  "当前有未保存更改，返回题库会丢失本地修改。是否继续？",
                )
              ) {
                return;
              }
              router.push("/exams");
            }}
          >
            ← 返回题库
          </Button>
          <Input
            value={title}
            onChange={(e) => {
              if (!isReadOnly) setTitle(e.target.value);
            }}
            className="max-w-sm font-medium"
            placeholder="交互单元标题"
            disabled={isReadOnly}
          />
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">
            {isReadOnly
              ? "已发布版本只读"
              : isDirty
                ? "未保存更改"
                : "已保存"}
          </span>
          <Button variant="outline" onClick={handleCreateVersion}>
            新建版本
          </Button>
          <Button
            variant="outline"
            onClick={handleSave}
            disabled={saving || isReadOnly || !isDirty}
          >
            {saving ? "保存中..." : "保存"}
          </Button>
          {versionDetail?.status === "draft" && (
            <Button
              onClick={handlePublish}
              disabled={publishing || !publishReadiness.canPublish}
              className="bg-emerald-600 hover:bg-emerald-700"
            >
              {publishing ? "发布中..." : "发布"}
            </Button>
          )}
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex w-[260px] shrink-0 flex-col border-r bg-background">
          <StepList
            steps={localSteps}
            selectedIndex={selectedStepIndex}
            stepStatuses={publishReadiness.stepStatuses}
            onSelect={setSelectedStepIndex}
            onReorder={handleReorder}
            onDelete={handleDeleteStep}
          />
        </div>
        <FlowCanvas
          steps={localSteps}
          selectedIndex={selectedStepIndex}
          onSelect={setSelectedStepIndex}
          onInsertBlock={handleAddBlock}
        />
        <div className="flex w-[420px] shrink-0 flex-col overflow-y-auto border-l bg-background">
          <StepConfigForm
            step={currentStep}
            onChange={handleStepChange}
          />
          {currentStep ? (
            <div className="px-4 pb-4">
              <div className="space-y-2 rounded-lg border bg-muted/10 p-4">
                <p className="text-sm font-medium">知识点绑定</p>
                <KpMultiSelect
                  knowledgePoints={knowledgePoints}
                  selected={new Set(currentStep.knowledge_point_ids ?? [])}
                  onToggle={toggleKnowledgePoint}
                  disabled={knowledgePoints.length === 0 || isReadOnly}
                />
              </div>
            </div>
          ) : null}
          <StepPreview step={currentStep} />
          <ResultSummaryPanel steps={localSteps} readiness={publishReadiness} />
        </div>
      </div>

      <div className="border-t px-6 py-2">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>版本历史:</span>
          {versions.length === 0 && (
            <span className="text-xs">暂无版本</span>
          )}
          {versions.map((v) => (
            <button
              key={v.version_id}
              onClick={() => loadVersion(v.version_id)}
              className={`rounded px-2 py-0.5 text-xs transition-colors hover:bg-muted ${
                v.version_id === versionDetail?.version_id
                  ? "bg-primary/10 font-medium text-primary"
                  : ""
              }`}
            >
              v{v.version_no} {statusLabel(v.status)}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
