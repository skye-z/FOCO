"use client"

import { useEffect, useMemo, useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"

import {
  authFetch,
  readBrowserAccessToken,
  readStoredSession,
} from "@/lib/auth-session"
import { buildDiagnosticSummaryText } from "@/lib/diagnostic"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import {
  AlertCircle,
  Award,
  BarChart3,
  BookOpen,
  CheckCircle2,
  ChevronRight,
  Clock,
  Coins,
  Flame,
  Loader2,
  Sparkles,
  Target,
  TrendingUp,
} from "lucide-react"

const API_BASE = "/api/v1"

interface StudyTask {
  id: string
  title: string
  type: string
  typeLabel: string
  duration: string
  status: string
  progressPercent: number
  xpRewardPreview: number
  coinRewardPreview: number
  reason: string
  actionHref: string
  actionTarget?: Record<string, string | number | boolean | null | undefined>
}

interface WeeklyActivity {
  day: string
  date: string
  minutes: number
  completedTasks: number
  questionCount: number
  xpEarned: number
  keptStreak: boolean
}

interface WeakPoint {
  id: string
  name: string
  mastery: number
  confidenceScore: number
  attempts: number
  correctCount: number
  forgettingDueAt?: string
  lastEvidenceAt?: string
  reviewStage: string
  reviewDue: boolean
  intervalDays: number
  source: string
}

interface RecommendationReason {
  reason_code: string
  reason_text: string
  evidence: Record<string, string | number | boolean | null | undefined>
}

interface Recommendation {
  id: string
  task_type: string
  task_type_label: string
  title: string
  estimated_minutes: number
  priority_score: number
  status: string
  action_href: string
  reasons: RecommendationReason[]
}

interface ProgressStats {
  completedQuestions: number
  accuracy: number
  lastStudiedAt?: string
  wrongCount: number
  totalXP: number
  completedSessions: number
}

interface LearningReport {
  periodLabel: string
  coreMetrics: Array<{
    label: string
    value: string
    delta?: string
  }>
  weakSummary: string
  trendSummary: string
  nextActions: string[]
  generatedAt: string
}

interface DiagnosticSummary {
  has_completed: boolean
  summary_text: string
  recommended_difficulty: "easy" | "medium" | "hard"
  recommended_subject_names: string[]
  recommended_chapter_names: string[]
  recommended_knowledge_point_names: string[]
}

interface VolatilityAlert {
  shouldRetest: boolean
  message: string
  minAccuracy: number
  maxAccuracy: number
}

interface DashboardData {
  examName: string
  nextExamDate?: string
  nextNextExamDate?: string
  countdownDays?: number
  level: number
  levelTitle: string
  coins: number
  streak: number
  xpCurrent: number
  xpTarget: number
  diagnosticSummary?: DiagnosticSummary
  volatilityAlert?: VolatilityAlert
  todayTasks: StudyTask[]
  weeklyActivity: WeeklyActivity[]
  weakPoints: WeakPoint[]
  recommendations: Recommendation[]
  progressStats: ProgressStats
  learningReport: LearningReport
}

type MePayload = {
  active_exam_enrollment: null | {
    exam_id: string
    exam_name: string
  }
}

function authHeaders(): Record<string, string> {
  const token = readBrowserAccessToken()
  const headers: Record<string, string> = {}
  if (token) headers.Authorization = `Bearer ${token}`
  return headers
}

function taskIcon(type: string) {
  switch (type) {
    case "continue":
      return <Clock className="size-4" />
    case "review":
    case "spaced_review":
      return <Target className="size-4" />
    case "sprint":
      return <TrendingUp className="size-4" />
    default:
      return <BookOpen className="size-4" />
  }
}

function taskTypeColor(type: string) {
  switch (type) {
    case "continue":
      return "bg-[var(--primary)] text-white"
    case "review":
    case "spaced_review":
      return "bg-[var(--error)] text-white"
    case "sprint":
      return "bg-[var(--secondary)] text-white"
    default:
      return "bg-[var(--tertiary-fixed)] text-[var(--tertiary)]"
  }
}

function taskStatusLabel(status: string) {
  switch (status) {
    case "completed":
      return "已完成"
    case "in_progress":
      return "进行中"
    default:
      return "待开始"
  }
}

function isCompletedTask(task: StudyTask) {
  return task.status === "completed" || task.progressPercent >= 100
}

function formatShortDate(dateLike?: string) {
  if (!dateLike) return "暂无"
  const date = new Date(dateLike)
  if (Number.isNaN(date.getTime())) return dateLike
  return date.toLocaleDateString("zh-CN", {
    month: "numeric",
    day: "numeric",
  })
}

function formatExamDate(dateLike?: string) {
  if (!dateLike) return "待设置"
  const date = new Date(dateLike)
  if (Number.isNaN(date.getTime())) return dateLike
  return date.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "numeric",
    day: "numeric",
  })
}

function formatRelativeTime(dateLike?: string) {
  if (!dateLike) return "暂无记录"
  const date = new Date(dateLike)
  if (Number.isNaN(date.getTime())) return dateLike

  const diffMinutes = Math.round((Date.now() - date.getTime()) / 60000)
  const absMinutes = Math.abs(diffMinutes)
  const suffix = diffMinutes >= 0 ? "前" : "后"

  if (absMinutes < 1) return "刚刚"
  if (absMinutes < 60) return `${absMinutes}分钟${suffix}`

  const absHours = Math.round(absMinutes / 60)
  if (absHours < 24) return `${absHours}小时${suffix}`

  const absDays = Math.round(absHours / 24)
  return `${absDays}天${suffix}`
}

async function resolveExamID() {
  const storedExamID = readStoredSession()?.activeExam?.id
  if (storedExamID) return storedExamID

  const response = await authFetch(`${API_BASE}/me`, {
    headers: authHeaders(),
    cache: "no-store",
  })
  if (!response.ok) {
    return ""
  }

  const payload = (await response.json()) as { data?: MePayload }
  return payload.data?.active_exam_enrollment?.exam_id ?? ""
}

async function loadDashboard(): Promise<DashboardData | null> {
  const examID = await resolveExamID()
  if (!examID) {
    return null
  }

  const response = await authFetch(`${API_BASE}/learner/home?exam_id=${encodeURIComponent(examID)}`, {
    headers: authHeaders(),
    cache: "no-store",
  })
  if (!response.ok) {
    throw new Error("Failed to load dashboard")
  }

  const payload = (await response.json()) as { data?: DashboardData }
  return payload.data ?? null
}

function WeeklyBarChart({ data }: { data: WeeklyActivity[] }) {
  const maxMinutes = Math.max(...data.map((d) => d.minutes), 1)
  const totalMinutes = data.reduce((sum, item) => sum + item.minutes, 0)
  const activeDays = data.filter((item) => item.minutes > 0).length
  const averageMinutes = data.length > 0 ? Math.round(totalMinutes / data.length) : 0

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-7 gap-2">
        {data.map((item) => {
          const heightPercent = (item.minutes / maxMinutes) * 100
          return (
            <div key={item.day} className="flex min-w-0 flex-col items-center gap-1">
              <span className="text-[10px] tabular-nums text-[var(--text-muted)]">{item.minutes}m</span>
              <div className="flex h-20 w-full items-end justify-center rounded-full bg-[var(--surface-container-low)] px-1 py-1">
                <div
                  className={cn(
                    "w-full max-w-[16px] rounded-full transition-all",
                    item.minutes > 0 ? "bg-[var(--secondary)]" : "bg-[var(--outline-variant)]",
                  )}
                  style={{ height: `${Math.max(heightPercent, item.minutes > 0 ? 12 : 4)}%` }}
                />
              </div>
              <span className="text-[10px] text-[var(--on-surface-variant)]">{item.day}</span>
            </div>
          )
        })}
      </div>
      <div className="grid grid-cols-2 gap-2 text-xs text-[var(--text-muted)]">
        <div className="rounded-xl bg-[var(--surface)] px-3 py-2">
          <p className="text-[10px] text-[var(--text-muted)]">最近 7 天合计</p>
          <p className="mt-1 text-sm font-semibold text-[var(--text-main)]">{totalMinutes} 分钟</p>
        </div>
        <div className="rounded-xl bg-[var(--surface)] px-3 py-2">
          <p className="text-[10px] text-[var(--text-muted)]">活跃天数</p>
          <p className="mt-1 text-sm font-semibold text-[var(--text-main)]">
            {activeDays} 天 · {averageMinutes} 分钟/天
          </p>
        </div>
      </div>
    </div>
  )
}

export default function HomePage() {
  const router = useRouter()
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [startingTaskId, setStartingTaskId] = useState<string | null>(null)
  const [taskStartError, setTaskStartError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    loadDashboard()
      .then((payload) => {
        if (!cancelled) {
          setData(payload)
          setError(null)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setError("无法加载首页数据，请稍后重试。")
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [])

  const xpPercent = useMemo(() => {
    if (!data || data.xpTarget <= 0) return 0
    return Math.round((data.xpCurrent / data.xpTarget) * 100)
  }, [data])

  async function createPracticeSession(requestBody: Record<string, unknown>) {
    const response = await authFetch(`${API_BASE}/learner/practice-sessions`, {
      method: "POST",
      headers: {
        ...authHeaders(),
        "Content-Type": "application/json",
      },
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
    return payload.data?.session_id ?? payload.data?.id ?? payload.session_id ?? ""
  }

  function isPracticeCreationTask(task: StudyTask) {
    return task.actionHref === "/practice/setup" || ["spaced_review", "sprint", "mixed"].includes(task.type)
  }

  async function handleStartTask(task: StudyTask) {
    if (isCompletedTask(task) || startingTaskId) return

    setTaskStartError(null)

    if (task.actionHref?.startsWith("/practice/") && task.actionHref !== "/practice/setup") {
      router.push(task.actionHref)
      return
    }

    if (!isPracticeCreationTask(task)) {
      router.push(task.actionHref || "/practice/setup")
      return
    }

    setStartingTaskId(task.id)
    try {
      const examId = await resolveExamID()
      if (!examId) throw new Error("还没有激活考试，请先完成引导。")

      const knowledgePointId = task.actionTarget?.knowledge_point_id
      const hasKnowledgePointTarget = typeof knowledgePointId === "string" && knowledgePointId.trim() !== ""
      const count = task.type === "sprint" ? 20 : task.type === "spaced_review" ? 10 : 15
      const sessionId = await createPracticeSession({
        exam_id: examId,
        mode: hasKnowledgePointTarget ? "manual" : "intelligent",
        question_types: ["single_choice", "multiple_choice", "judgment"],
        count,
        ...(hasKnowledgePointTarget ? { knowledge_point_ids: [knowledgePointId] } : {}),
      })

      if (!sessionId) throw new Error("创建练习会话失败")
      router.push(`/practice/${sessionId}`)
    } catch (err) {
      setTaskStartError(err instanceof Error ? err.message : "创建练习会话失败，请稍后重试。")
    } finally {
      setStartingTaskId(null)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="flex items-center gap-3 text-sm text-[var(--text-muted)]">
          <Loader2 className="size-5 animate-spin text-[var(--secondary)]" />
          正在加载首页...
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="space-y-3 text-center">
          <p className="text-sm text-[var(--error)]">{error}</p>
          <Link href="/practice/setup">
            <Button className="rounded-full">前往练习</Button>
          </Link>
        </div>
      </div>
    )
  }

  if (!data) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="space-y-3 text-center">
          <p className="text-sm text-[var(--text-muted)]">还没有激活考试，请先完成引导。</p>
          <Link href="/onboarding">
            <Button className="rounded-full">前往引导</Button>
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-[var(--container-max-width)] px-[var(--margin-mobile)] py-6 md:px-[var(--margin-desktop)]">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-12">
        <div className="flex flex-col gap-4 md:col-span-4">
          <div className="flex flex-col gap-5 rounded-2xl bg-[var(--surface-container-lowest)] p-6 shadow-sm">
            <div className="relative mx-auto">
              <div className="flex size-24 items-center justify-center rounded-full bg-[var(--secondary-container)]">
                <Award className="size-10 text-[var(--secondary)]" />
              </div>
            </div>

            <div className="space-y-2 text-center">
              <p className="text-sm text-[var(--text-muted)]">{data.examName}</p>
              <p className="text-lg font-semibold text-[var(--text-main)]">
                Level {data.level} {data.levelTitle}
              </p>
            </div>

            <div className="flex flex-wrap items-center justify-center gap-3">
              <Badge className="gap-1 rounded-full border-0 bg-[var(--tertiary-fixed)] px-3 py-1 text-sm text-[var(--tertiary)]">
                <Coins className="size-3.5" />
                {data.coins.toLocaleString()}
              </Badge>
              <Badge className="gap-1 rounded-full border-0 bg-[var(--error-container)] px-3 py-1 text-sm text-[var(--error)]">
                <Flame className="size-3.5" />
                {data.streak} 天
              </Badge>
            </div>

            <div className="rounded-2xl bg-[var(--surface)] p-4 text-center">
              <p className="text-xs text-[var(--text-muted)]">下次考试</p>
              <p className="mt-1 text-sm font-medium text-[var(--text-main)]">{formatExamDate(data.nextExamDate)}</p>
              <p className="mt-2 text-2xl font-semibold text-[var(--secondary)]">
                {typeof data.countdownDays === "number" ? `${data.countdownDays} 天` : "待定"}
              </p>
              {data.nextNextExamDate ? (
                <p className="mt-2 text-xs text-[var(--text-muted)]">下下次考试：{formatExamDate(data.nextNextExamDate)}</p>
              ) : null}
            </div>

            <div className="w-full space-y-2">
              <div className="flex items-center justify-between text-xs text-[var(--on-surface-variant)]">
                <span>经验值</span>
                <span>
                  {data.xpCurrent} / {data.xpTarget} XP
                </span>
              </div>
              <Progress value={xpPercent} className="h-2.5 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]" />
            </div>

            <Separator className="bg-[var(--outline-variant)]" />

            <Link href="/practice/setup">
              <Button className="h-10 w-full rounded-xl bg-[var(--secondary)] text-white hover:bg-[var(--secondary)]/90">
                开始今日练习
                <ChevronRight className="size-4" />
              </Button>
            </Link>
          </div>

          <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm ring-0">
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base font-semibold text-[var(--text-main)]">
                <TrendingUp className="size-4 text-[var(--secondary)]" />
                本周学习节奏
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <WeeklyBarChart data={data.weeklyActivity} />
            </CardContent>
          </Card>

          {data.diagnosticSummary ? (
            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm ring-0">
              <CardHeader className="pb-3">
                <CardTitle className="flex items-center gap-2 text-base font-semibold text-[var(--text-main)]">
                  <Sparkles className="size-4 text-[var(--secondary)]" />
                  学业水平
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 pt-0">
                <div className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-4 text-sm leading-6 text-[var(--text-main)]">
                  {buildDiagnosticSummaryText(data.diagnosticSummary)}
                </div>
                <div className="flex flex-wrap gap-2">
                  {data.diagnosticSummary.recommended_subject_names.map((name) => (
                    <Badge key={name} className="rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
                      {name}
                    </Badge>
                  ))}
                  {data.diagnosticSummary.recommended_chapter_names.map((name) => (
                    <Badge key={name} variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                      {name}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          ) : (
            <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm ring-0">
              <CardHeader className="pb-3">
                <CardTitle className="flex items-center gap-2 text-base font-semibold text-[var(--text-main)]">
                  <AlertCircle className="size-4 text-[var(--error)]" />
                  学业水平
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4 pt-0">
                <p className="text-sm text-[var(--text-muted)]">
                  完成一次引导测验后，系统才能为你生成强弱项分析，并解锁智能练习。
                </p>
                <Link href={`/diagnostic?exam_id=${encodeURIComponent(readStoredSession()?.activeExam?.id ?? "")}`}>
                  <Button className="rounded-full">去完成测验</Button>
                </Link>
              </CardContent>
            </Card>
          )}
        </div>

        <div className="flex flex-col gap-4 md:col-span-8">
          {data.volatilityAlert?.shouldRetest ? (
            <Card className="rounded-2xl border border-[var(--error)]/20 bg-[var(--error-container)]/35 shadow-sm ring-0">
              <CardContent className="flex items-start gap-3 py-5">
                <AlertCircle className="mt-0.5 size-5 text-[var(--error)]" />
                <div className="space-y-2">
                  <p className="font-medium text-[var(--text-main)]">{data.volatilityAlert.message}</p>
                  <p className="text-sm text-[var(--text-muted)]">
                    最近 10 次练习正确率区间：{data.volatilityAlert.minAccuracy}% - {data.volatilityAlert.maxAccuracy}%
                  </p>
                  <Link href={`/diagnostic?exam_id=${encodeURIComponent(readStoredSession()?.activeExam?.id ?? "")}`}>
                    <Button variant="outline" className="rounded-full">
                      重新测验
                    </Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          ) : null}

          <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {[
              {
                label: "已完成题数",
                value: data.progressStats.completedQuestions,
                icon: CheckCircle2,
              },
              {
                label: "累计正确率",
                value: `${data.progressStats.accuracy}%`,
                icon: BarChart3,
              },
              {
                label: "最近学习",
                value: formatRelativeTime(data.progressStats.lastStudiedAt),
                icon: Clock,
              },
              {
                label: "错题知识点",
                value: data.progressStats.wrongCount,
                icon: Target,
              },
            ].map((item) => (
              <Card key={item.label} className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm ring-0">
                <CardContent className="flex items-center gap-3 py-4">
                  <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-[var(--secondary)]/10 text-[var(--secondary)]">
                    <item.icon className="size-5" />
                  </div>
                  <div className="min-w-0">
                    <p className="truncate text-xl font-semibold text-[var(--text-main)]">{item.value}</p>
                    <p className="text-xs text-[var(--text-muted)]">{item.label}</p>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm ring-0">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base font-semibold text-[var(--text-main)]">
                <Clock className="size-4 text-[var(--secondary)]" />
                今日学习路径
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {taskStartError ? (
                <div className="rounded-xl border border-[var(--error)]/25 bg-[var(--error-container)]/40 px-4 py-3 text-sm text-[var(--error)]">
                  {taskStartError}
                </div>
              ) : null}
              {data.todayTasks.map((task, idx) => (
                <div key={task.id}>
                  <div
                    className={cn(
                      "grid gap-3 rounded-2xl border border-transparent px-3 py-3 transition-colors md:grid-cols-[auto_1fr_auto] md:items-center",
                      isCompletedTask(task)
                        ? "border-green-200 bg-green-50/80"
                        : "bg-transparent",
                    )}
                  >
                    <div
                      className={cn(
                        "flex size-8 shrink-0 items-center justify-center rounded-lg",
                        isCompletedTask(task)
                          ? "bg-green-600 text-white"
                          : "bg-[var(--surface-container)] text-[var(--secondary)]",
                      )}
                    >
                      {isCompletedTask(task) ? <CheckCircle2 className="size-4" /> : taskIcon(task.type)}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <p
                          className={cn(
                            "text-sm font-medium text-[var(--text-main)]",
                            isCompletedTask(task) && "text-green-800 line-through decoration-green-500/70",
                          )}
                        >
                          {task.title}
                        </p>
                        <span className={cn("inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-medium leading-none", taskTypeColor(task.type))}>
                          {task.typeLabel}
                        </span>
                      </div>
                      <p className="mt-1 text-xs leading-5 text-[var(--text-muted)]">{task.reason}</p>
                      <div className="mt-2 flex flex-wrap gap-2 text-[10px] text-[var(--on-surface-variant)]">
                        <span>{task.duration}</span>
                        <span>XP +{task.xpRewardPreview}</span>
                        <span>金币 +{task.coinRewardPreview}</span>
                        <span>{task.status === "in_progress" ? `进度 ${task.progressPercent}%` : taskStatusLabel(task.status)}</span>
                      </div>
                    </div>
                    <div className="md:justify-self-end">
                      {isCompletedTask(task) ? (
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-9 rounded-full border-[var(--secondary)]/20 bg-white text-[var(--secondary)] gap-2"
                          disabled
                        >
                          <CheckCircle2 className="size-4" />
                          已完成
                        </Button>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-9 rounded-full"
                          disabled={startingTaskId !== null}
                          onClick={() => handleStartTask(task)}
                        >
                          {startingTaskId === task.id ? (
                            <>
                              <Loader2 className="size-3.5 animate-spin" />
                              创建中...
                            </>
                          ) : task.status === "in_progress" ? (
                            "继续"
                          ) : (
                            "开始"
                          )}
                        </Button>
                      )}
                    </div>
                  </div>
                  {idx < data.todayTasks.length - 1 ? <Separator className="bg-[var(--outline-variant)] opacity-50" /> : null}
                </div>
              ))}
            </CardContent>
          </Card>

          <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm ring-0">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base font-semibold text-[var(--text-main)]">
                <Target className="size-4 text-[var(--secondary)]" />
                薄弱知识点
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              {data.weakPoints.length === 0 ? (
                <p className="text-sm text-[var(--text-muted)]">当前还没有足够的真实答题数据来识别薄弱点。</p>
              ) : (
                data.weakPoints.map((point) => (
                  <div key={point.id} className="space-y-1.5">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium text-[var(--text-main)]">{point.name}</span>
                      <span className={cn("text-xs font-medium", point.mastery < 50 ? "text-[var(--error)]" : "text-[var(--on-surface-variant)]")}>
                        {point.mastery}%
                      </span>
                    </div>
                    <Progress
                      value={point.mastery}
                      className={cn(
                        "h-1.5 bg-[var(--surface-container-high)]",
                        point.mastery < 50
                          ? "[&>[data-slot=progress-indicator]]:bg-[var(--error)]"
                          : "[&>[data-slot=progress-indicator]]:bg-[var(--secondary)]",
                      )}
                    />
                    <div className="flex flex-wrap gap-3 text-[10px] text-[var(--text-muted)]">
                      <span>置信度 {point.confidenceScore}%</span>
                      <span>证据 {point.correctCount}/{point.attempts}</span>
                      <span>{point.reviewStage || "间隔复习"}</span>
                      {point.lastEvidenceAt ? <span>最近证据 {formatShortDate(point.lastEvidenceAt)}</span> : null}
                      {point.forgettingDueAt ? <span>{point.reviewDue ? "今日到期" : "复习窗口"} {formatShortDate(point.forgettingDueAt)}</span> : null}
                    </div>
                  </div>
                ))
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
