"use client"

import { useEffect, useMemo, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  AlertCircle,
  Brain,
  CheckCircle2,
  Loader2,
  RefreshCw,
} from "lucide-react"

import { authFetch, readBrowserAccessToken, readStoredSession } from "@/lib/auth-session"
import { buildDiagnosticSummaryText } from "@/lib/diagnostic"
import { cn } from "@/lib/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"

const API_BASE = "/api/v1"

type DiagnosticOption = {
  label: string
  text: string
}

type DiagnosticItem = {
  id: string
  question_version_id: string
  subject_id: string
  subject_name: string
  chapter_id: string
  chapter_name: string
  question_type: "single_choice" | "multiple_choice"
  stem: string
  options: DiagnosticOption[]
  knowledge_points: { id: string; name: string }[]
}

type DiagnosticSummary = {
  has_completed: boolean
  completed_at?: string
  overall_accuracy: number
  summary_text: string
  recommended_difficulty: "easy" | "medium" | "hard"
  recommended_subject_ids: string[]
  recommended_subject_names: string[]
  recommended_chapter_ids: string[]
  recommended_chapter_names: string[]
  recommended_knowledge_point_ids: string[]
  recommended_knowledge_point_names: string[]
  knowledge_points: Array<{
    knowledge_point_id: string
    knowledge_point_name: string
    mastery_score: number
    confidence_score: number
    forgetting_due_at: string
  }>
}

type DiagnosticPayload = {
  status: "pending" | "completed"
  attempt_id?: string
  items?: DiagnosticItem[]
  summary?: DiagnosticSummary
}

function authHeaders(): Record<string, string> {
  const token = readBrowserAccessToken()
  const headers: Record<string, string> = { "Content-Type": "application/json" }
  if (token) headers.Authorization = `Bearer ${token}`
  return headers
}

function difficultyLabel(value: string) {
  switch (value) {
    case "easy":
      return "简单"
    case "medium":
      return "中等"
    case "hard":
      return "困难"
    default:
      return value
  }
}

export default function DiagnosticPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const storedExamId = readStoredSession()?.activeExam?.id ?? ""
  const examId = searchParams.get("exam_id") ?? storedExamId

  const [payload, setPayload] = useState<DiagnosticPayload | null>(null)
  const [answers, setAnswers] = useState<Record<string, string[]>>({})
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadCurrent() {
      if (!examId) {
        setError("缺少考试信息，请先完成考试选择。")
        setLoading(false)
        return
      }

      try {
        const response = await authFetch(
          `${API_BASE}/learner/diagnostic/current?exam_id=${encodeURIComponent(examId)}`,
          { headers: authHeaders(), cache: "no-store" },
        )
        if (!response.ok) throw new Error("Failed to load diagnostic")
        const json = (await response.json()) as { data?: DiagnosticPayload }
        if (!cancelled) {
          setPayload(json.data ?? null)
          if (json.data?.attempt_id) {
            try {
              const saved = window.localStorage.getItem(`foco.diagnostic.${json.data.attempt_id}`)
              if (saved) {
                const parsed = JSON.parse(saved) as Record<string, string[]>
                if (typeof parsed === "object" && parsed !== null) {
                  setAnswers(parsed)
                }
              }
            } catch {}
          }
        }
      } catch {
        if (!cancelled) {
          setError("无法加载引导测验，请稍后重试。")
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadCurrent()
    return () => {
      cancelled = true
    }
  }, [examId])

  useEffect(() => {
    if (!payload?.attempt_id) return
    if (Object.keys(answers).length === 0) return
    try {
      window.localStorage.setItem(`foco.diagnostic.${payload.attempt_id}`, JSON.stringify(answers))
    } catch {}
  }, [answers, payload?.attempt_id])

  const completionRatio = useMemo(() => {
    if (!payload?.items?.length) return 0
    const answered = payload.items.filter((item) => (answers[item.id] ?? []).length > 0).length
    return Math.round((answered / payload.items.length) * 100)
  }, [answers, payload?.items])

  function toggleAnswer(item: DiagnosticItem, label: string) {
    setAnswers((current) => {
      const selected = current[item.id] ?? []
      if (item.question_type === "single_choice") {
        return { ...current, [item.id]: selected.includes(label) ? [] : [label] }
      }
      return {
        ...current,
        [item.id]: selected.includes(label)
          ? selected.filter((value) => value !== label)
          : [...selected, label],
      }
    })
  }

  async function handleRestart() {
    if (!examId) return
    setSubmitting(true)
    setError(null)
    try {
      const response = await authFetch(`${API_BASE}/learner/diagnostic/restart`, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({ exam_id: examId }),
      })
      if (!response.ok) throw new Error("restart failed")
      const json = (await response.json()) as { data?: DiagnosticPayload }
      setAnswers({})
      setPayload(json.data ?? null)
    } catch {
      setError("重新开始测验失败，请稍后重试。")
    } finally {
      setSubmitting(false)
    }
  }

  async function handleSubmit() {
    if (!payload?.attempt_id || !payload.items) return
    const missing = payload.items.some((item) => (answers[item.id] ?? []).length === 0)
    if (missing) {
      setError("请先完成所有测验题目。")
      return
    }

    setSubmitting(true)
    setError(null)
    try {
      const response = await authFetch(`${API_BASE}/learner/diagnostic/${payload.attempt_id}/submit`, {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({ answers }),
      })
      if (!response.ok) throw new Error("submit failed")
      const json = (await response.json()) as { data?: DiagnosticSummary }
      if (payload.attempt_id) {
        try { window.localStorage.removeItem(`foco.diagnostic.${payload.attempt_id}`) } catch {}
      }
      setPayload({ status: "completed", summary: json.data })
    } catch {
      setError("提交测验失败，请稍后重试。")
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--surface)] px-6">
        <div className="flex items-center gap-3 text-sm text-[var(--text-muted)]">
          <Loader2 className="size-5 animate-spin text-[var(--secondary)]" />
          正在准备引导测验...
        </div>
      </div>
    )
  }

  if (payload?.status === "completed" && payload.summary) {
    const summary = payload.summary
    return (
      <div className="min-h-screen bg-[var(--surface)] px-6 py-10">
        <div className="mx-auto max-w-3xl space-y-6">
          <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
            <CardHeader className="gap-3">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="size-5 text-[var(--secondary)]" />
                <CardTitle>引导测验完成</CardTitle>
              </div>
              <CardDescription>已生成初始强弱项分析，后续智能练习会优先围绕这些结果进行推荐。</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid gap-4 sm:grid-cols-3">
                <div className="rounded-2xl bg-[var(--surface)] p-4 text-center">
                  <p className="text-xs text-[var(--text-muted)]">测验正确率</p>
                  <p className="mt-2 text-3xl font-semibold text-[var(--text-main)]">{summary.overall_accuracy}%</p>
                </div>
                <div className="rounded-2xl bg-[var(--surface)] p-4 text-center">
                  <p className="text-xs text-[var(--text-muted)]">建议难度</p>
                  <p className="mt-2 text-2xl font-semibold text-[var(--text-main)]">{difficultyLabel(summary.recommended_difficulty)}</p>
                </div>
                <div className="rounded-2xl bg-[var(--surface)] p-4 text-center">
                  <p className="text-xs text-[var(--text-muted)]">完成时间</p>
                  <p className="mt-2 text-sm font-medium text-[var(--text-main)]">
                    {summary.completed_at ? new Date(summary.completed_at).toLocaleDateString("zh-CN") : "刚刚"}
                  </p>
                </div>
              </div>

              <div className="rounded-2xl border border-[var(--outline-variant)] bg-[var(--surface)] p-4 text-sm leading-6 text-[var(--text-main)]">
                {buildDiagnosticSummaryText(summary)}
              </div>

              <div className="space-y-2">
                <p className="text-sm font-semibold text-[var(--text-main)]">推荐优先补强</p>
                <div className="flex flex-wrap gap-2">
                  {summary.recommended_subject_names.map((name) => (
                    <Badge key={name} className="rounded-full bg-[var(--secondary)]/10 text-[var(--secondary)] hover:bg-[var(--secondary)]/10">
                      {name}
                    </Badge>
                  ))}
                  {summary.recommended_chapter_names.map((name) => (
                    <Badge key={name} variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                      {name}
                    </Badge>
                  ))}
                </div>
              </div>

              <div className="space-y-3">
                {summary.knowledge_points.slice(0, 5).map((item) => (
                  <div key={item.knowledge_point_id} className="space-y-1.5">
                    <div className="flex items-center justify-between text-sm">
                      <span className="font-medium text-[var(--text-main)]">{item.knowledge_point_name}</span>
                      <span className="text-[var(--text-muted)]">{item.mastery_score}%</span>
                    </div>
                    <Progress
                      value={item.mastery_score}
                      className="h-2 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]"
                    />
                  </div>
                ))}
              </div>

              {error ? <p className="text-sm text-[var(--error)]">{error}</p> : null}

              <div className="flex flex-col gap-3 sm:flex-row">
                <Button className="flex-1 rounded-full" onClick={() => router.push("/home")}>
                  进入首页
                </Button>
                <Button variant="outline" className="flex-1 rounded-full" disabled={submitting} onClick={handleRestart}>
                  <RefreshCw className={cn("size-4", submitting && "animate-spin")} />
                  重新测验
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  const items = payload?.items ?? []

  return (
    <div className="min-h-screen bg-[var(--surface)] px-6 py-10">
      <div className="mx-auto max-w-3xl space-y-6">
        <div className="space-y-2 text-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-[var(--secondary)]/10 px-4 py-1.5 text-sm text-[var(--secondary)]">
            <Brain className="size-4" />
            引导测验
          </div>
          <h1 className="text-3xl font-semibold text-[var(--text-main)]">先完成一次入门诊断</h1>
          <p className="text-sm text-[var(--text-muted)]">测验结果会生成初始强弱项分析，并驱动后续智能练习推荐。</p>
        </div>

        <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
          <CardHeader>
            <CardTitle>完成进度</CardTitle>
            <CardDescription>请回答全部题目后再提交诊断结果。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-[var(--text-muted)]">当前完成度</span>
              <span className="font-medium text-[var(--text-main)]">{completionRatio}%</span>
            </div>
            <Progress value={completionRatio} className="h-2.5 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]" />
          </CardContent>
        </Card>

        <div className="space-y-4">
          {items.map((item, index) => {
            const selected = answers[item.id] ?? []
            return (
              <Card key={item.id} className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
                <CardHeader className="gap-3">
                  <div className="flex items-center justify-between gap-3">
                    <CardTitle className="text-base">第 {index + 1} 题</CardTitle>
                    <Badge variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                      {item.subject_name}
                    </Badge>
                  </div>
                  <CardDescription>{item.chapter_name}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <p className="text-sm leading-6 text-[var(--text-main)]">{item.stem}</p>
                  <div className="space-y-2">
                    {item.options.map((option) => {
                      const active = selected.includes(option.label)
                      return (
                        <Button
                          key={option.label}
                          type="button"
                          variant="outline"
                          onClick={() => toggleAnswer(item, option.label)}
                          className={cn(
                            "h-auto w-full justify-start rounded-2xl px-4 py-3 text-left",
                            active
                              ? "border-[var(--secondary)] bg-[var(--secondary)]/10 text-[var(--secondary)]"
                              : "border-[var(--outline-variant)] bg-[var(--surface)] text-[var(--text-main)]",
                          )}
                        >
                          <div className="flex items-start gap-3">
                            <div className="mt-0.5 flex size-5 items-center justify-center rounded-full border text-xs">
                              {option.label}
                            </div>
                            <span className="text-sm leading-6">{option.text}</span>
                          </div>
                        </Button>
                      )
                    })}
                  </div>
                  {item.knowledge_points.length > 0 ? (
                    <>
                      <Separator />
                      <div className="flex flex-wrap gap-2">
                        {item.knowledge_points.map((kp) => (
                          <Badge key={kp.id} variant="outline" className="rounded-full border-[var(--outline-variant)] bg-[var(--surface)]">
                            {kp.name}
                          </Badge>
                        ))}
                      </div>
                    </>
                  ) : null}
                </CardContent>
              </Card>
            )
          })}
        </div>

        {error ? <p className="text-center text-sm text-[var(--error)]">{error}</p> : null}

        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button className="rounded-full" disabled={submitting || items.length === 0} onClick={handleSubmit}>
            {submitting ? (
              <>
                <Loader2 className="size-4 animate-spin" />
                正在生成结果...
              </>
            ) : (
              "提交测验"
            )}
          </Button>
          <Button variant="outline" className="rounded-full" disabled={submitting} onClick={handleRestart}>
            <RefreshCw className={cn("size-4", submitting && "animate-spin")} />
            重新抽题
          </Button>
          <Button variant="ghost" className="rounded-full" onClick={() => router.push("/home")}>
            稍后再说
          </Button>
        </div>
      </div>
    </div>
  )
}
