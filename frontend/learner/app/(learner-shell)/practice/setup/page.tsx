"use client"

import { useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import {
  BookOpen,
  Brain,
  CheckCircle2,
  Layers3,
  Loader2,
  Search,
  Sparkles,
} from "lucide-react"
import { toast } from "sonner"

import {
  authFetch,
  readBrowserAccessToken,
  readStoredSession,
  type ActiveEnrollmentSnapshot,
  writeActiveExam,
} from "@/lib/auth-session"
import { buildDiagnosticSummaryText } from "@/lib/diagnostic"
import {
  buildPracticeSessionRequest,
  filterVisibleChapters,
  isSmartPracticeLocked,
  type ContentTreeSubject,
  type PracticeDifficulty,
  type PracticeQuestionType,
} from "@/lib/practice-setup"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"

const API_BASE = "/api/v1"

type ActiveExam = {
  id: string
  name: string
}

type ContentTree = {
  exam: { id: string; name: string }
  subjects: ContentTreeSubject[]
}

type MePayload = {
  user: {
    id: string
    email: string
    roles: string[]
  }
  active_exam_enrollment: ActiveEnrollmentSnapshot
}

type DiagnosticSummary = {
  has_completed: boolean
  summary_text: string
  recommended_difficulty: PracticeDifficulty
  recommended_subject_names: string[]
  recommended_chapter_names: string[]
  recommended_knowledge_point_names: string[]
}

type DiagnosticPayload = {
  status: "pending" | "completed"
  attempt_id?: string
  summary?: DiagnosticSummary
}

const QUESTION_TYPE_OPTIONS: { value: PracticeQuestionType; label: string }[] = [
  { value: "single_choice", label: "单选题" },
  { value: "multiple_choice", label: "多选题" },
  { value: "judgment", label: "判断题" },
]

const DIFFICULTY_OPTIONS: { value: PracticeDifficulty; label: string }[] = [
  { value: "easy", label: "简单" },
  { value: "medium", label: "中等" },
  { value: "hard", label: "困难" },
]

function authHeaders(): Record<string, string> {
  const token = readBrowserAccessToken()
  const headers: Record<string, string> = { "Content-Type": "application/json" }
  if (token) headers.Authorization = `Bearer ${token}`
  return headers
}

async function fetchMe(): Promise<MePayload | null> {
  const response = await authFetch(`${API_BASE}/me`, {
    headers: authHeaders(),
    cache: "no-store",
  })
  if (!response.ok) return null
  const payload = (await response.json()) as { data?: MePayload }
  return payload.data ?? null
}

function difficultyLabel(value: PracticeDifficulty) {
  return DIFFICULTY_OPTIONS.find((item) => item.value === value)?.label ?? value
}

export default function PracticeSetupPage() {
  const router = useRouter()

  const [mode, setMode] = useState<"intelligent" | "manual">("manual")
  const [activeExam, setActiveExam] = useState<ActiveExam | null>(readStoredSession()?.activeExam ?? null)
  const [contentTree, setContentTree] = useState<ContentTree | null>(null)
  const [diagnostic, setDiagnostic] = useState<DiagnosticPayload | null>(null)
  const [selectedSubjectIds, setSelectedSubjectIds] = useState<Set<string>>(new Set())
  const [selectedChapterIds, setSelectedChapterIds] = useState<Set<string>>(new Set())
  const [questionTypes, setQuestionTypes] = useState<PracticeQuestionType[]>([
    "single_choice",
    "multiple_choice",
  ])
  const [difficulty, setDifficulty] = useState<PracticeDifficulty>("medium")
  const [count, setCount] = useState(15)
  const [chapterQuery, setChapterQuery] = useState("")

  const [loadingIdentity, setLoadingIdentity] = useState(true)
  const [loadingTree, setLoadingTree] = useState(false)
  const [loadingDiagnostic, setLoadingDiagnostic] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadIdentity() {
      try {
        const me = await fetchMe()
        if (cancelled || !me) return

        const enrollment = me.active_exam_enrollment
        if (enrollment) {
          const nextActiveExam = {
            id: enrollment.exam_id,
            name: enrollment.exam_name,
          }
          writeActiveExam(nextActiveExam)
          setActiveExam(nextActiveExam)
        } else if (!readStoredSession()?.activeExam) {
          setActiveExam(null)
        }
      } catch {
        if (!cancelled) {
          setError("无法确认当前激活考试，请重新登录后重试。")
        }
      } finally {
        if (!cancelled) {
          setLoadingIdentity(false)
        }
      }
    }

    void loadIdentity()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!activeExam?.id) {
      setContentTree(null)
      setDiagnostic(null)
      return
    }

    let cancelled = false
    const examId = activeExam.id

    async function loadResources() {
      setLoadingTree(true)
      setLoadingDiagnostic(true)
      setError(null)

      try {
        const [treeResponse, diagnosticResponse] = await Promise.all([
          authFetch(`${API_BASE}/learner/exams/${examId}/content-tree`, {
            headers: authHeaders(),
          }),
          authFetch(`${API_BASE}/learner/diagnostic/current?exam_id=${encodeURIComponent(examId)}`, {
            headers: authHeaders(),
            cache: "no-store",
          }),
        ])

        if (!treeResponse.ok) throw new Error("Failed to load content tree")
        if (!diagnosticResponse.ok) throw new Error("Failed to load diagnostic")

        const treePayload = (await treeResponse.json()) as { data?: ContentTree }
        const diagnosticPayload = (await diagnosticResponse.json()) as { data?: DiagnosticPayload }

        if (!cancelled) {
          const diagnosticData = diagnosticPayload.data ?? null
          setContentTree(treePayload.data ?? null)
          setDiagnostic(diagnosticData)
          setMode(isSmartPracticeLocked(diagnosticData?.summary ?? null) ? "manual" : "intelligent")
        }
      } catch {
        if (!cancelled) {
          setError("无法加载练习配置，请稍后重试。")
        }
      } finally {
        if (!cancelled) {
          setLoadingTree(false)
          setLoadingDiagnostic(false)
        }
      }
    }

    void loadResources()
    return () => {
      cancelled = true
    }
  }, [activeExam?.id])

  const smartLocked = isSmartPracticeLocked(diagnostic?.summary ?? null)

  const visibleChapters = useMemo(
    () =>
      filterVisibleChapters(
        contentTree?.subjects ?? [],
        selectedSubjectIds,
        chapterQuery,
      ),
    [chapterQuery, contentTree?.subjects, selectedSubjectIds],
  )

  const selectedSubjects = useMemo(
    () => (contentTree?.subjects ?? []).filter((subject) => selectedSubjectIds.has(subject.id)),
    [contentTree?.subjects, selectedSubjectIds],
  )

  const selectedChapters = useMemo(
    () => visibleChapters.filter(({ chapter }) => selectedChapterIds.has(chapter.id)).map(({ chapter }) => chapter),
    [selectedChapterIds, visibleChapters],
  )

  function toggleQuestionType(type: PracticeQuestionType) {
    setQuestionTypes((current) =>
      current.includes(type) ? current.filter((item) => item !== type) : [...current, type],
    )
  }

  function handleModeCardKeyDown(
    event: React.KeyboardEvent<HTMLDivElement>,
    nextMode: "intelligent" | "manual",
  ) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault()
      setMode(nextMode)
    }
  }

  function toggleSubject(subjectId: string) {
    setSelectedSubjectIds((current) => {
      const next = new Set(current)
      if (next.has(subjectId)) next.delete(subjectId)
      else next.add(subjectId)
      return next
    })

    setSelectedChapterIds((current) => {
      if (!contentTree) return current
      const stillSelected = !selectedSubjectIds.has(subjectId)
      if (stillSelected) return current

      const removedChapterIds = new Set(
        contentTree.subjects.find((subject) => subject.id === subjectId)?.chapters.map((chapter) => chapter.id) ?? [],
      )
      const next = new Set(current)
      for (const chapterId of removedChapterIds) next.delete(chapterId)
      return next
    })
  }

  function toggleChapter(chapterId: string) {
    setSelectedChapterIds((current) => {
      const next = new Set(current)
      if (next.has(chapterId)) next.delete(chapterId)
      else next.add(chapterId)
      return next
    })
  }

  async function createPracticeSession(requestBody: Record<string, unknown>) {
    const response = await authFetch(`${API_BASE}/learner/practice-sessions`, {
      method: "POST",
      headers: authHeaders(),
      body: JSON.stringify(requestBody),
    })

    if (!response.ok) {
      const payload = (await response.json().catch(() => null)) as { error?: string; message?: string } | null
      throw new Error(payload?.error ?? payload?.message ?? "创建练习会话失败")
    }

    const payload = (await response.json()) as {
      data?: { session_id?: string; id?: string }
      session_id?: string
    }
    return payload.data?.session_id ?? payload.data?.id ?? payload.session_id
  }

  async function handleStartSmartPractice() {
    if (!activeExam?.id) return
    if (smartLocked) {
      router.push(`/diagnostic?exam_id=${encodeURIComponent(activeExam.id)}`)
      return
    }

    setSubmitting(true)
    setError(null)
    try {
      const sessionId = await createPracticeSession({
        exam_id: activeExam.id,
        mode: "intelligent",
        question_types: QUESTION_TYPE_OPTIONS.map((item) => item.value),
        count: 15,
      })
      if (!sessionId) throw new Error("创建练习会话失败")
      router.push(`/practice/${sessionId}`)
    } catch (err) {
      const message = err instanceof Error ? err.message : "创建练习会话失败，请稍后重试。"
      setError(message)
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  async function handleStartManualPractice() {
    if (!activeExam?.id) return
    if (questionTypes.length === 0) {
      toast.error("请至少选择一种题型。")
      return
    }

    setSubmitting(true)
    setError(null)
    try {
      const sessionId = await createPracticeSession(
        buildPracticeSessionRequest({
          examId: activeExam.id,
          mode: "manual",
          questionTypes,
          difficulty,
          count,
          selectedSubjectIds: Array.from(selectedSubjectIds),
          selectedChapterIds: Array.from(selectedChapterIds),
        }),
      )
      if (!sessionId) throw new Error("创建练习会话失败")
      router.push(`/practice/${sessionId}`)
    } catch (err) {
      const message = err instanceof Error ? err.message : "创建练习会话失败，请稍后重试。"
      setError(message)
      toast.error(message)
    } finally {
      setSubmitting(false)
    }
  }

  if (loadingIdentity) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="flex items-center gap-3 text-sm text-[var(--text-muted)]">
          <Loader2 className="size-5 animate-spin text-[var(--secondary)]" />
          正在确认当前激活考试...
        </div>
      </div>
    )
  }

  if (!activeExam) {
    return (
      <div className="mx-auto max-w-2xl px-[var(--margin-mobile)] py-8 md:px-[var(--margin-desktop)]">
        <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
          <CardHeader>
            <CardTitle>还没有激活考试</CardTitle>
            <CardDescription>请先完成新用户引导并选择考试。</CardDescription>
          </CardHeader>
          <CardContent>
            <Button className="rounded-full" onClick={() => router.push("/onboarding")}>
              前往引导
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-[var(--container-max-width)] px-[var(--margin-mobile)] py-6 md:px-[var(--margin-desktop)]">
      <div className="mb-8 space-y-2">
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="font-heading text-3xl font-bold tracking-tight text-[var(--text-main)]">
            选择练习方式
          </h1>
          <Badge className="rounded-full bg-[var(--secondary)]/10 px-3 py-1 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
            当前激活考试
          </Badge>
        </div>
        <p className="text-base text-[var(--text-muted)]">
          当前练习默认使用你的激活考试：
          <span className="ml-1 font-medium text-[var(--text-main)]">{activeExam.name}</span>
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        <div className="flex flex-col gap-6 lg:col-span-8">
          <div className="grid gap-4 md:grid-cols-2">
            <Card
              className={cn(
                "cursor-pointer rounded-2xl border-0 shadow-sm transition-all hover:shadow-md",
                mode === "intelligent" ? "bg-[var(--secondary)]/5 ring-2 ring-[var(--secondary)]/30" : "bg-[var(--surface-container-lowest)]",
              )}
              role="button"
              tabIndex={0}
              aria-pressed={mode === "intelligent"}
              onClick={() => setMode("intelligent")}
              onKeyDown={(event) => handleModeCardKeyDown(event, "intelligent")}
            >
              <CardHeader>
                <div className="flex items-center gap-2">
                  <Sparkles className="size-5 text-[var(--secondary)]" />
                  <CardTitle>智能练习</CardTitle>
                </div>
                <CardDescription>
                  根据引导测验结果自动选择薄弱的难度、科目、章节和知识点。
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {loadingDiagnostic ? (
                  <div className="flex items-center gap-2 text-sm text-[var(--text-muted)]">
                    <Loader2 className="size-4 animate-spin" />
                    正在读取测验结果...
                  </div>
                ) : smartLocked ? (
                  <div className="space-y-3 rounded-2xl border border-dashed border-[var(--outline-variant)] bg-[var(--surface)] p-4">
                    <p className="text-sm text-[var(--text-main)]">
                      智能练习需要先完成引导测验。
                    </p>
                    <Button
                      variant="outline"
                      className="rounded-full"
                      onClick={() => router.push(`/diagnostic?exam_id=${encodeURIComponent(activeExam.id)}`)}
                    >
                      去完成测验
                    </Button>
                  </div>
                ) : diagnostic?.summary ? (
                  <>
                    <div className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-4 text-sm leading-6 text-[var(--text-main)]">
                      {buildDiagnosticSummaryText(diagnostic.summary)}
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {diagnostic.summary.recommended_subject_names.map((name) => (
                        <Badge key={name} className="rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
                          {name}
                        </Badge>
                      ))}
                      {diagnostic.summary.recommended_chapter_names.map((name) => (
                        <Badge key={name} variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                          {name}
                        </Badge>
                      ))}
                    </div>
                  </>
                ) : null}
              </CardContent>
            </Card>

            <Card
              className={cn(
                "cursor-pointer rounded-2xl border-0 shadow-sm transition-all hover:shadow-md",
                mode === "manual" ? "bg-[var(--primary)]/5 ring-2 ring-[var(--primary)]/20" : "bg-[var(--surface-container-lowest)]",
              )}
              role="button"
              tabIndex={0}
              aria-pressed={mode === "manual"}
              onClick={() => setMode("manual")}
              onKeyDown={(event) => handleModeCardKeyDown(event, "manual")}
            >
              <CardHeader>
                <div className="flex items-center gap-2">
                  <BookOpen className="size-5 text-[var(--primary)]" />
                  <CardTitle>手动练习</CardTitle>
                </div>
                <CardDescription>
                  自己选择科目、章节、题型和难度，适合有明确训练目标时使用。
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-4 text-sm text-[var(--text-muted)]">
                  手动练习不依赖诊断结果，你可以直接按范围自定义本次练习。
                </div>
              </CardContent>
            </Card>
          </div>

          {mode === "manual" ? (
            <>
              <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
                <CardHeader className="gap-2">
                  <CardTitle className="flex items-center gap-2">
                    <BookOpen className="size-5 text-[var(--secondary)]" />
                    科目
                  </CardTitle>
                  <CardDescription>默认按当前激活考试出题，可进一步缩小到一个或多个科目。</CardDescription>
                </CardHeader>
                <CardContent>
                  {loadingTree ? (
                    <div className="flex items-center gap-2 py-3 text-sm text-[var(--text-muted)]">
                      <Loader2 className="size-4 animate-spin" />
                      正在加载科目...
                    </div>
                  ) : (
                    <div className="flex flex-wrap gap-3">
                      {(contentTree?.subjects ?? []).map((subject) => {
                        const selected = selectedSubjectIds.has(subject.id)
                        return (
                          <Button
                            key={subject.id}
                            type="button"
                            variant={selected ? "default" : "outline"}
                            className={cn(
                              "rounded-full px-5",
                              selected
                                ? "bg-[var(--secondary)] text-white hover:bg-[var(--secondary)]/90"
                                : "border-[var(--outline-variant)] bg-[var(--surface-container)] text-[var(--on-surface-variant)] hover:bg-[var(--surface-container-highest)]",
                            )}
                            onClick={() => toggleSubject(subject.id)}
                          >
                            {subject.name}
                          </Button>
                        )
                      })}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
                <CardHeader className="gap-3">
                  <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                    <div className="space-y-2">
                      <CardTitle className="flex items-center gap-2">
                        <Layers3 className="size-5 text-[var(--secondary)]" />
                        章节
                      </CardTitle>
                      <CardDescription>未选科目时会展示当前考试下全部章节。</CardDescription>
                    </div>
                    <div className="relative w-full md:w-64">
                      <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-[var(--text-muted)]" />
                      <Input
                        value={chapterQuery}
                        onChange={(event) => setChapterQuery(event.target.value)}
                        placeholder="搜索章节"
                        className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)] pl-9"
                      />
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  {loadingTree ? (
                    <div className="flex items-center gap-2 py-3 text-sm text-[var(--text-muted)]">
                      <Loader2 className="size-4 animate-spin" />
                      正在加载章节...
                    </div>
                  ) : visibleChapters.length === 0 ? (
                    <p className="text-sm text-[var(--text-muted)]">没有匹配的章节，请调整筛选条件。</p>
                  ) : (
                    <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                      {visibleChapters.map(({ chapter, subjectName }) => {
                        const selected = selectedChapterIds.has(chapter.id)
                        return (
                          <Button
                            key={chapter.id}
                            type="button"
                            variant={selected ? "default" : "outline"}
                            onClick={() => toggleChapter(chapter.id)}
                            className={cn(
                              "h-auto items-start justify-start rounded-2xl px-4 py-4 text-left",
                              selected
                                ? "bg-[var(--secondary)] text-white hover:bg-[var(--secondary)]/90"
                                : "border-[var(--outline-variant)] bg-white text-[var(--text-main)] hover:bg-[var(--surface-container-low)]",
                            )}
                          >
                            <div className="flex w-full items-start gap-3">
                              <div className={cn("mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border", selected ? "border-white bg-white/15 text-white" : "border-[var(--outline-variant)] text-transparent")}>
                                <CheckCircle2 className="size-3.5" />
                              </div>
                              <div className="min-w-0 space-y-1">
                                <p className="text-sm font-medium leading-5">{chapter.name}</p>
                                <p className={cn("text-xs", selected ? "text-white/80" : "text-[var(--text-muted)]")}>
                                  {subjectName}
                                </p>
                              </div>
                            </div>
                          </Button>
                        )
                      })}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
                <CardHeader className="gap-2">
                  <CardTitle>手动练习设置</CardTitle>
                </CardHeader>
                <CardContent className="space-y-5">
                  <div className="space-y-3">
                    <Label className="text-sm font-semibold">题型</Label>
                    <div className="flex flex-wrap gap-3">
                      {QUESTION_TYPE_OPTIONS.map((option) => {
                        const selected = questionTypes.includes(option.value)
                        return (
                          <Button
                            key={option.value}
                            type="button"
                            variant={selected ? "default" : "outline"}
                            className={cn(
                              "rounded-full px-5",
                              selected
                                ? "bg-[var(--secondary)] text-white hover:bg-[var(--secondary)]/90"
                                : "border-[var(--outline-variant)] bg-[var(--surface-container)] text-[var(--on-surface-variant)] hover:bg-[var(--surface-container-highest)]",
                            )}
                            onClick={() => toggleQuestionType(option.value)}
                          >
                            {option.label}
                          </Button>
                        )
                      })}
                    </div>
                  </div>

                  <div className="space-y-3">
                    <Label className="text-sm font-semibold">难度</Label>
                    <div className="flex flex-wrap gap-3">
                      {DIFFICULTY_OPTIONS.map((option) => {
                        const selected = difficulty === option.value
                        return (
                          <Button
                            key={option.value}
                            type="button"
                            variant={selected ? "default" : "outline"}
                            className={cn(
                              "rounded-full px-5",
                              selected
                                ? "bg-[var(--secondary)] text-white hover:bg-[var(--secondary)]/90"
                                : "border-[var(--outline-variant)] bg-[var(--surface-container)] text-[var(--on-surface-variant)] hover:bg-[var(--surface-container-highest)]",
                            )}
                            onClick={() => setDifficulty(option.value)}
                          >
                            {option.label}
                          </Button>
                        )
                      })}
                    </div>
                  </div>

                  <div className="space-y-3">
                    <Label className="text-sm font-semibold">题目数量</Label>
                    <Input
                      type="number"
                      min={5}
                      max={50}
                      value={count}
                      onChange={(event) => {
                        const next = Number.parseInt(event.target.value, 10)
                        if (!Number.isNaN(next)) setCount(Math.min(50, Math.max(5, next)))
                      }}
                      className="w-28 rounded-full border-[var(--outline-variant)] bg-[var(--surface)]"
                    />
                  </div>
                </CardContent>
              </Card>
            </>
          ) : null}
        </div>

        <div className="lg:col-span-4">
          <Card className="sticky top-24 rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
            <CardHeader className="gap-2">
              <CardTitle>{mode === "intelligent" ? "智能练习预览" : "手动练习预览"}</CardTitle>
              <CardDescription>
                {mode === "intelligent"
                  ? "系统会根据引导测验和最近表现自动给出最适合的练习范围。"
                  : "确认范围后即可直接进入手动练习。"}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <div className="rounded-2xl bg-[var(--surface-container-low)] p-4">
                <div className="space-y-3 text-sm">
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-[var(--text-muted)]">当前考试</span>
                    <span className="font-medium text-[var(--text-main)]">{activeExam.name}</span>
                  </div>
                  {mode === "intelligent" && diagnostic?.summary ? (
                    <>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-[var(--text-muted)]">建议难度</span>
                        <span className="font-medium text-[var(--text-main)]">
                          {difficultyLabel(diagnostic.summary.recommended_difficulty)}
                        </span>
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {diagnostic.summary.recommended_subject_names.map((name) => (
                          <Badge key={name} className="rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
                            {name}
                          </Badge>
                        ))}
                        {diagnostic.summary.recommended_chapter_names.map((name) => (
                          <Badge key={name} variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                            {name}
                          </Badge>
                        ))}
                      </div>
                    </>
                  ) : (
                    <>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-[var(--text-muted)]">已选科目</span>
                        <span className="font-medium text-[var(--text-main)]">{selectedSubjectIds.size || "全部"}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-[var(--text-muted)]">已选章节</span>
                        <span className="font-medium text-[var(--text-main)]">{selectedChapterIds.size || "全部"}</span>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-[var(--text-muted)]">难度</span>
                        <span className="font-medium text-[var(--text-main)]">{difficultyLabel(difficulty)}</span>
                      </div>
                    </>
                  )}
                </div>
              </div>

              {mode === "manual" && selectedSubjects.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {selectedSubjects.map((subject) => (
                    <Badge key={subject.id} variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                      {subject.name}
                    </Badge>
                  ))}
                </div>
              ) : null}

              {mode === "manual" && selectedChapters.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {selectedChapters.slice(0, 6).map((chapter) => (
                    <Badge key={chapter.id} className="rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
                      {chapter.name}
                    </Badge>
                  ))}
                </div>
              ) : null}

              {error ? <p className="text-sm text-[var(--error)]">{error}</p> : null}

              <div className="flex flex-col gap-3">
                <Button
                  size="lg"
                  className="w-full rounded-full bg-[var(--secondary)] text-white hover:bg-[var(--secondary)]/90"
                  disabled={submitting || loadingTree || loadingDiagnostic}
                  onClick={mode === "intelligent" ? handleStartSmartPractice : handleStartManualPractice}
                >
                  {submitting ? (
                    <>
                      <Loader2 className="size-4 animate-spin" />
                      正在创建练习...
                    </>
                  ) : mode === "intelligent" ? (
                    "开始智能练习"
                  ) : (
                    "开始手动练习"
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
