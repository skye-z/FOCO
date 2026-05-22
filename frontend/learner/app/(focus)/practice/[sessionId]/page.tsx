"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { useRouter, useParams } from "next/navigation"
import { toast } from "sonner"
import {
  ArrowLeft,
  Check,
  X,
  Clock,
  ChevronRight,
  Loader2,
  AlertCircle,
  Trophy,
} from "lucide-react"

import { authFetch, readBrowserAccessToken } from "@/lib/auth-session"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Progress } from "@/components/ui/progress"

const API_BASE = "/api/v1"

type QuestionOption = {
  label: string
  text: string
}

type QuestionContent = {
  stem: string
  options: QuestionOption[]
}

type PracticeItem = {
  item_id: string
  question_version_id: string
  question_type: "single_choice" | "multiple_choice" | "judgment"
  score: number
  content: QuestionContent
  submitted: boolean
  is_correct?: boolean
  user_answer?: string[]
}

type PracticeSession = {
  session_id: string
  exam_name: string
  total_count: number
  items: PracticeItem[]
}

type SubmitResult = {
  is_correct: boolean
  correct_answer: string | string[]
  explanation: string
  knowledge_points: { id: string; name: string }[]
  xp_earned: number
}

type QuestionState = {
  selectedAnswers: string[]
  submitted: boolean
  result: SubmitResult | null
}

function authHeaders(): Record<string, string> {
  const token = readBrowserAccessToken()
  const headers: Record<string, string> = { "Content-Type": "application/json" }
  if (token) headers["Authorization"] = `Bearer ${token}`
  return headers
}

const TYPE_LABELS: Record<string, string> = {
  single_choice: "单选题",
  multiple_choice: "多选题",
  judgment: "判断题",
}

function isSingleChoiceLike(questionType: string) {
  return questionType === "single_choice" || questionType === "judgment"
}

function formatElapsed(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`
}

export default function PracticeSessionPage() {
  const router = useRouter()
  const params = useParams<{ sessionId: string }>()
  const sessionId = params.sessionId

  const [session, setSession] = useState<PracticeSession | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [currentIndex, setCurrentIndex] = useState(0)
  const [questionStates, setQuestionStates] = useState<QuestionState[]>([])
  const [submitting, setSubmitting] = useState(false)

  const [elapsedSeconds, setElapsedSeconds] = useState(0)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const questionStartRef = useRef<number>(Date.now())

  const [totalSeconds, setTotalSeconds] = useState(0)
  const totalTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    let cancelled = false
    async function loadSession() {
      try {
        const res = await authFetch(
          `${API_BASE}/learner/practice-sessions/${sessionId}`,
          { headers: authHeaders() }
        )
        if (!res.ok) throw new Error("Failed to load session")
        const json = await res.json()
        const data: PracticeSession = json.data ?? json
        if (!cancelled) {
          setSession(data)
          let firstUnanswered = 0
          setQuestionStates(
            data.items.map((item, idx) => {
              if (item.submitted) {
                if (firstUnanswered === idx) firstUnanswered = idx + 1
                return {
                  selectedAnswers: item.user_answer ?? [],
                  submitted: true,
                  result: item.is_correct != null
                    ? {
                        is_correct: item.is_correct,
                        correct_answer: [],
                        explanation: "",
                        knowledge_points: [],
                        xp_earned: 0,
                      }
                    : null,
                }
              }
              return {
                selectedAnswers: [],
                submitted: false,
                result: null,
              }
            })
          )
          setCurrentIndex(firstUnanswered < data.items.length ? firstUnanswered : data.items.length - 1)
        }
      } catch {
        if (!cancelled) setError("无法加载练习会话，请稍后重试")
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    loadSession()
    return () => {
      cancelled = true
    }
  }, [sessionId])

  useEffect(() => {
    totalTimerRef.current = setInterval(() => {
      setTotalSeconds((s) => s + 1)
    }, 1000)
    return () => {
      if (totalTimerRef.current) clearInterval(totalTimerRef.current)
    }
  }, [])

  useEffect(() => {
    questionStartRef.current = Date.now()
    setElapsedSeconds(0)
    timerRef.current = setInterval(() => {
      setElapsedSeconds(
        Math.floor((Date.now() - questionStartRef.current) / 1000)
      )
    }, 1000)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [currentIndex])

  const currentItem = session?.items[currentIndex]
  const currentQState = questionStates[currentIndex]

  const answeredCount = questionStates.filter(
    (qs) => qs.submitted
  ).length

  const correctCount = questionStates.filter(
    (qs) => qs.submitted && qs.result?.is_correct
  ).length

  const allAnswered =
    session && answeredCount === session.total_count

  const toggleAnswer = useCallback(
    (label: string) => {
      if (!currentQState || currentQState.submitted) return
      const item = currentItem
      if (!item) return

      setQuestionStates((prev) => {
        const next = [...prev]
        const state = { ...next[currentIndex] }

        if (isSingleChoiceLike(item.question_type)) {
          state.selectedAnswers = state.selectedAnswers.includes(label)
            ? []
            : [label]
        } else {
          if (state.selectedAnswers.includes(label)) {
            state.selectedAnswers = state.selectedAnswers.filter(
              (a) => a !== label
            )
          } else {
            state.selectedAnswers = [...state.selectedAnswers, label]
          }
        }

        next[currentIndex] = state
        return next
      })
    },
    [currentIndex, currentItem, currentQState]
  )

  const handleSubmit = useCallback(async () => {
    if (!currentItem || !currentQState) return
    if (currentQState.selectedAnswers.length === 0) {
      toast.error("请先选择答案")
      return
    }

    const durationSec = Math.floor(
      (Date.now() - questionStartRef.current) / 1000
    )
    setSubmitting(true)

    try {
      const answer =
        isSingleChoiceLike(currentItem.question_type)
          ? currentQState.selectedAnswers[0]
          : currentQState.selectedAnswers

      const res = await authFetch(
        `${API_BASE}/learner/practice-sessions/${sessionId}/items/${currentItem.item_id}/submit`,
        {
          method: "POST",
          headers: authHeaders(),
          body: JSON.stringify({
            answer,
            duration_seconds: durationSec,
          }),
        }
      )

      if (res.status === 409) {
        toast.error("该题目已作答")
        setQuestionStates((prev) => {
          const next = [...prev]
          next[currentIndex] = {
            ...next[currentIndex],
            submitted: true,
            result: null,
          }
          return next
        })
        return
      }
      if (!res.ok) throw new Error("Submit failed")
      const json = await res.json()
      const result: SubmitResult = json.data ?? json

      setQuestionStates((prev) => {
        const next = [...prev]
        next[currentIndex] = {
          ...next[currentIndex],
          submitted: true,
          result,
        }
        return next
      })

      if (timerRef.current) clearInterval(timerRef.current)

      if (result.xp_earned > 0) {
        toast.success(`+${result.xp_earned} XP`, {
          description: result.is_correct ? "回答正确！" : "继续加油！",
        })
      }
    } catch {
      toast.error("提交失败，请重试")
    } finally {
      setSubmitting(false)
    }
  }, [currentItem, currentQState, sessionId, currentIndex])

  const handleNext = useCallback(() => {
    if (!session) return
    if (currentIndex < session.total_count - 1) {
      setCurrentIndex(currentIndex + 1)
    }
  }, [currentIndex, session])

  const goToQuestion = useCallback((index: number) => {
    setCurrentIndex(index)
  }, [])

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="size-8 animate-spin text-[var(--secondary)]" />
          <span className="text-sm text-[var(--text-muted)]">加载练习...</span>
        </div>
      </div>
    )
  }

  if (error || !session) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="flex flex-col items-center gap-4 text-center">
          <AlertCircle className="size-10 text-[var(--error)]" />
          <p className="text-sm text-[var(--text-muted)]">
            {error ?? "无法加载练习数据"}
          </p>
          <Button variant="outline" onClick={() => router.push("/home")}>
            返回首页
          </Button>
        </div>
      </div>
    )
  }

  const correctAnswerArr: string[] = currentQState?.result?.correct_answer
    ? Array.isArray(currentQState.result.correct_answer)
      ? currentQState.result.correct_answer
      : [currentQState.result.correct_answer]
    : []

  return (
    <div className="flex min-h-screen flex-col">
      {/* Sticky top bar */}
      <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] px-4 md:px-6">
        <Button
          variant="ghost"
          size="sm"
          className="gap-1.5 text-[var(--text-muted)]"
          onClick={() => router.push("/home")}
        >
          <ArrowLeft className="size-4" />
          退出练习
        </Button>

        <span className="hidden font-heading text-sm font-semibold text-[var(--text-main)] md:block truncate max-w-xs">
          {session.exam_name}
        </span>

        <span className="text-sm font-medium tabular-nums text-[var(--on-surface-variant)]">
          {answeredCount} / {session.total_count}
        </span>
      </header>

      {/* 3-column body */}
      <div className="mx-auto flex w-full max-w-[var(--container-max-width)] flex-1 gap-4 px-[var(--margin-mobile)] py-6 md:px-[var(--margin-desktop)]">
        {/* Left — Answer Sheet */}
        <aside className="hidden w-[220px] shrink-0 lg:block">
          <div className="sticky top-20">
            <Card className="bg-[var(--surface-container-lowest)] ring-0 border-0 rounded-xl">
              <CardContent className="pt-4 pb-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-semibold text-[var(--text-main)]">
                    答题卡
                  </h3>
                  <span className="text-xs text-[var(--text-muted)] tabular-nums">
                    {correctCount}/{answeredCount}
                  </span>
                </div>
                <div className="grid grid-cols-5 gap-2">
                  {session.items.map((item, idx) => {
                    const qs = questionStates[idx]
                    let circleClass =
                      "border-2 border-[var(--outline-variant)] text-[var(--on-surface-variant)] bg-transparent"

                    if (qs?.submitted && qs.result) {
                      if (qs.result.is_correct) {
                        circleClass =
                          "bg-[var(--secondary)] text-white border-0"
                      } else {
                        circleClass =
                          "bg-[var(--error)] text-white border-0"
                      }
                    } else if (idx === currentIndex) {
                      circleClass =
                        "border-2 border-[var(--secondary)] text-[var(--secondary)] bg-transparent"
                    }

                    return (
                      <button
                        key={item.item_id}
                        type="button"
                        onClick={() => goToQuestion(idx)}
                        className={cn(
                          "flex size-9 items-center justify-center rounded-full text-xs font-medium transition-colors hover:opacity-80",
                          circleClass
                        )}
                        aria-label={`第 ${idx + 1} 题`}
                      >
                        {idx + 1}
                      </button>
                    )
                  })}
                </div>

                <Separator className="my-3 bg-[var(--outline-variant)]" />

                <div className="flex flex-col gap-1.5 text-xs text-[var(--text-muted)]">
                  <div className="flex items-center gap-2">
                    <span className="size-3 rounded-full bg-[var(--secondary)]" />
                    正确
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="size-3 rounded-full bg-[var(--error)]" />
                    错误
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="size-3 rounded-full border-2 border-[var(--outline-variant)]" />
                    未答
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </aside>

        {/* Center — Question */}
        <main className="flex-1 min-w-0">
          {currentItem && currentQState && (
            <Card className="bg-[var(--surface-container-lowest)] ring-0 border-0 rounded-xl">
              <CardContent className="pt-5 pb-6 space-y-5">
                {/* Question header */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary" className="rounded-md text-xs">
                      {TYPE_LABELS[currentItem.question_type] ?? "题目"}
                    </Badge>
                    <span className="text-xs text-[var(--text-muted)]">
                      {currentItem.score} 分
                    </span>
                  </div>
                  <span className="text-xs text-[var(--text-muted)]">
                    第 {currentIndex + 1} / {session.total_count} 题
                  </span>
                </div>

                {/* Progress bar */}
                <Progress
                  value={((currentIndex + 1) / session.total_count) * 100}
                  className="h-1 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]"
                />

                {/* Stem */}
                <div className="text-base leading-relaxed text-[var(--text-main)] whitespace-pre-wrap">
                  {currentItem.content.stem}
                </div>

                {/* Options */}
                <div className="flex flex-col gap-2.5">
                  {currentItem.content.options.map((opt) => {
                    const isSelected = currentQState.selectedAnswers.includes(
                      opt.label
                    )
                    const isCorrect = correctAnswerArr.includes(opt.label)
                    const isWrong =
                      currentQState.submitted &&
                      isSelected &&
                      !isCorrect

                    let optionClass =
                      "border border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] hover:border-[var(--secondary)]"

                    if (currentQState.submitted && currentQState.result) {
                      if (isCorrect) {
                        optionClass =
                          "border-2 border-[var(--secondary)] bg-[var(--secondary)]/10"
                      } else if (isWrong) {
                        optionClass =
                          "border-2 border-[var(--error)] bg-[var(--error)]/10"
                      }
                    } else if (isSelected) {
                      optionClass =
                        "border-2 border-[var(--secondary)] bg-[var(--secondary)]/10"
                    }

                    return (
                      <button
                        key={opt.label}
                        type="button"
                        disabled={currentQState.submitted}
                        onClick={() => toggleAnswer(opt.label)}
                        className={cn(
                          "flex items-start gap-3 rounded-xl px-4 py-3 text-left transition-all disabled:cursor-default",
                          optionClass
                        )}
                      >
                        <span
                          className={cn(
                            "flex size-6 shrink-0 items-center justify-center rounded-full border text-xs font-medium",
                            currentQState.submitted && isCorrect
                              ? "border-[var(--secondary)] bg-[var(--secondary)] text-white"
                              : currentQState.submitted && isWrong
                                ? "border-[var(--error)] bg-[var(--error)] text-white"
                                : isSelected
                                  ? "border-[var(--secondary)] bg-[var(--secondary)] text-white"
                                  : "border-[var(--outline-variant)] text-[var(--on-surface-variant)]"
                          )}
                        >
                          {currentQState.submitted && isCorrect ? (
                            <Check className="size-3.5" />
                          ) : currentQState.submitted && isWrong ? (
                            <X className="size-3.5" />
                          ) : (
                            opt.label
                          )}
                        </span>
                        <span
                          className={cn(
                            "text-sm leading-relaxed pt-0.5",
                            currentQState.submitted && isCorrect
                              ? "text-[var(--secondary)] font-medium"
                              : currentQState.submitted && isWrong
                                ? "text-[var(--error)] font-medium"
                                : "text-[var(--text-main)]"
                          )}
                        >
                          {opt.text}
                        </span>
                      </button>
                    )
                  })}
                </div>

                {/* Explanation */}
                {currentQState.submitted && currentQState.result && (
                  <div
                    className={cn(
                      "rounded-xl p-4 text-sm leading-relaxed",
                      currentQState.result.is_correct
                        ? "bg-[var(--secondary)]/5 border border-[var(--secondary)]/20"
                        : "bg-[var(--error)]/5 border border-[var(--error)]/20"
                    )}
                  >
                    <p
                      className={cn(
                        "mb-2 text-xs font-semibold",
                        currentQState.result.is_correct
                          ? "text-[var(--secondary)]"
                          : "text-[var(--error)]"
                      )}
                    >
                      {currentQState.result.is_correct
                        ? "回答正确！"
                        : "回答错误"}
                    </p>
                    {currentQState.result.explanation && (
                      <p className="text-[var(--on-surface-variant)]">
                        {currentQState.result.explanation}
                      </p>
                    )}
                    {currentQState.result.knowledge_points &&
                      currentQState.result.knowledge_points.length > 0 && (
                        <div className="mt-2 flex flex-wrap gap-1.5">
                          {currentQState.result.knowledge_points.map((kp) => (
                            <Badge
                              key={kp.id}
                              variant="outline"
                              className="text-[10px] rounded-md"
                            >
                              {kp.name}
                            </Badge>
                          ))}
                        </div>
                      )}
                  </div>
                )}

                {/* Actions */}
                <div className="flex items-center justify-between pt-2">
                  <div className="text-xs text-[var(--text-muted)]">
                    {currentItem.question_type === "multiple_choice"
                      ? "可多选"
                      : ""}
                  </div>

                  <div className="flex items-center gap-2">
                    {!currentQState.submitted && (
                      <Button
                        onClick={handleSubmit}
                        disabled={
                          submitting ||
                          currentQState.selectedAnswers.length === 0
                        }
                        className="bg-[var(--secondary)] hover:bg-[var(--secondary)]/90 text-white rounded-xl"
                      >
                        {submitting ? (
                          <>
                            <Loader2 className="size-4 animate-spin" />
                            提交中...
                          </>
                        ) : (
                          "提交答案"
                        )}
                      </Button>
                    )}

                    {currentQState.submitted && !allAnswered && (
                      <Button
                        onClick={handleNext}
                        disabled={currentIndex === session.total_count - 1}
                        className="bg-[var(--secondary)] hover:bg-[var(--secondary)]/90 text-white rounded-xl gap-1"
                      >
                        下一题
                        <ChevronRight className="size-4" />
                      </Button>
                    )}

                    {currentQState.submitted && allAnswered && (
                      <Button
                        className="bg-[var(--secondary)] hover:bg-[var(--secondary)]/90 text-white rounded-xl gap-1"
                        onClick={() => router.push(`/practice/${sessionId}/complete`)}
                      >
                        <Trophy className="size-4" />
                        查看结果
                      </Button>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </main>

        {/* Right — Timer & Info */}
        <aside className="hidden w-[220px] shrink-0 lg:block">
          <div className="sticky top-20 space-y-3">
            <Card className="bg-[var(--surface-container-lowest)] ring-0 border-0 rounded-xl">
              <CardContent className="pt-4 pb-4 space-y-4">
                <div className="flex items-center gap-2 text-[var(--text-muted)]">
                  <Clock className="size-4" />
                  <span className="text-xs font-medium">当前用时</span>
                </div>
                <p className="text-center text-2xl font-heading font-bold tabular-nums text-[var(--text-main)]">
                  {formatElapsed(elapsedSeconds)}
                </p>
                <Separator className="bg-[var(--outline-variant)]" />
                <div className="flex items-center gap-2 text-[var(--text-muted)]">
                  <Clock className="size-4" />
                  <span className="text-xs font-medium">总用时</span>
                </div>
                <p className="text-center text-lg font-heading font-semibold tabular-nums text-[var(--on-surface-variant)]">
                  {formatElapsed(totalSeconds)}
                </p>
              </CardContent>
            </Card>

            <Card className="bg-[var(--surface-container-lowest)] ring-0 border-0 rounded-xl">
              <CardContent className="pt-4 pb-4 space-y-3">
                <h3 className="text-xs font-semibold text-[var(--text-muted)]">
                  本轮进度
                </h3>
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-[var(--text-muted)]">已完成</span>
                    <span className="font-medium tabular-nums text-[var(--text-main)]">
                      {answeredCount}/{session.total_count}
                    </span>
                  </div>
                  <Progress
                    value={
                      session.total_count > 0
                        ? (answeredCount / session.total_count) * 100
                        : 0
                    }
                    className="h-1.5 bg-[var(--surface-container-high)] [&>[data-slot=progress-indicator]]:bg-[var(--secondary)]"
                  />
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-[var(--secondary)] font-medium">
                    正确 {correctCount}
                  </span>
                  <span className="text-[var(--error)] font-medium">
                    错误 {answeredCount - correctCount}
                  </span>
                </div>
              </CardContent>
            </Card>

            {/* Mobile answer sheet shortcut (visible on < lg) handled by scrollable dots */}
            <div className="lg:hidden">
              <Card className="bg-[var(--surface-container-lowest)] ring-0 border-0 rounded-xl">
                <CardContent className="pt-3 pb-3">
                  <div className="flex flex-wrap gap-1.5">
                    {session.items.map((item, idx) => {
                      const qs = questionStates[idx]
                      let dotClass =
                        "bg-[var(--outline-variant)] text-white"
                      if (qs?.submitted && qs.result) {
                        dotClass = qs.result.is_correct
                          ? "bg-[var(--secondary)] text-white"
                          : "bg-[var(--error)] text-white"
                      } else if (idx === currentIndex) {
                        dotClass =
                          "bg-transparent border-2 border-[var(--secondary)] text-[var(--secondary)]"
                      }
                      return (
                        <button
                          key={item.item_id}
                          type="button"
                          onClick={() => goToQuestion(idx)}
                          className={cn(
                            "flex size-7 items-center justify-center rounded-full text-[10px] font-medium transition-colors",
                            dotClass
                          )}
                        >
                          {idx + 1}
                        </button>
                      )
                    })}
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </aside>
      </div>

      {/* Mobile bottom bar with answer dots */}
      <div className="sticky bottom-0 z-30 border-t border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] p-3 lg:hidden">
        <div className="flex items-center gap-2 overflow-x-auto">
          <span className="shrink-0 text-xs text-[var(--text-muted)] tabular-nums">
            {answeredCount}/{session.total_count}
          </span>
          <Separator orientation="vertical" className="h-5 shrink-0" />
          <div className="flex gap-1.5 overflow-x-auto py-0.5">
            {session.items.map((item, idx) => {
              const qs = questionStates[idx]
              let dotClass =
                "bg-transparent border-2 border-[var(--outline-variant)] text-[var(--on-surface-variant)]"
              if (qs?.submitted && qs.result) {
                dotClass = qs.result.is_correct
                  ? "bg-[var(--secondary)] text-white border-0"
                  : "bg-[var(--error)] text-white border-0"
              } else if (idx === currentIndex) {
                dotClass =
                  "bg-transparent border-2 border-[var(--secondary)] text-[var(--secondary)]"
              }
              return (
                <button
                  key={item.item_id}
                  type="button"
                  onClick={() => goToQuestion(idx)}
                  className={cn(
                    "flex size-7 shrink-0 items-center justify-center rounded-full text-[10px] font-medium transition-colors",
                    dotClass
                  )}
                >
                  {idx + 1}
                </button>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
