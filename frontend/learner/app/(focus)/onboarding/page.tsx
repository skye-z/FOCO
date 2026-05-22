"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { CheckCircle2, Loader2, GraduationCap } from "lucide-react"
import { toast } from "sonner"

import { authFetch, readBrowserAccessToken, writeActiveExam } from "@/lib/auth-session"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

type Exam = {
  id: string
  code: string
  name: string
  description: string
}

const STEPS = ["选择目标", "学习计划", "能力诊断"]

const EXAM_ICONS: Record<string, string> = {
  "grad-school": "🎓",
  ielts: "🌍",
  toefl: "📝",
  cfa: "📊",
  gre: "📐",
  gmat: "💼",
  cpas: "💰",
  default: "📚",
}

function getExamIcon(code: string) {
  return EXAM_ICONS[code] || EXAM_ICONS.default
}

export default function OnboardingPage() {
  const router = useRouter()
  const [exams, setExams] = useState<Exam[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const token = readBrowserAccessToken()
        const res = await authFetch("/api/v1/learner/exams", {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (!res.ok) throw new Error("Failed to load exams")
        const json = await res.json()
        setExams(json.data ?? [])
      } catch (err) {
        toast.error("无法加载考试列表，请稍后重试")
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  async function handleContinue() {
    if (!selectedId) return
    setSubmitting(true)
    try {
      const token = readBrowserAccessToken()
      const res = await authFetch("/api/v1/learner/exam-enrollments", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ exam_id: selectedId }),
      })
      if (!res.ok) throw new Error("Enrollment failed")
      const activeExam = exams.find((exam) => exam.id === selectedId)
      if (activeExam) {
        writeActiveExam({ id: activeExam.id, name: activeExam.name })
      }
      router.replace(`/diagnostic?exam_id=${encodeURIComponent(selectedId)}`)
    } catch {
      toast.error("报名失败，请稍后重试")
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center bg-[var(--surface)] px-6 py-12">
      <div className="w-full max-w-3xl">
        <div className="flex items-center justify-center gap-2 mb-10">
          {STEPS.map((label, i) => (
            <div key={label} className="flex items-center gap-2">
              <div
                className={cn(
                  "flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium transition-colors",
                  i === 0
                    ? "bg-[var(--secondary)] text-white"
                    : "bg-[var(--surface-container-highest)] text-[var(--on-surface-variant)]"
                )}
              >
                <span
                  className={cn(
                    "flex size-5 items-center justify-center rounded-full text-[10px] font-bold",
                    i === 0
                      ? "bg-white/20"
                      : "bg-[var(--surface-container)]"
                  )}
                >
                  {i + 1}
                </span>
                {label}
              </div>
              {i < STEPS.length - 1 && (
                <div className="h-px w-6 bg-[var(--outline-variant)]" />
              )}
            </div>
          ))}
        </div>

        <div className="text-center mb-10">
          <h1 className="text-2xl font-bold text-[var(--text-main)] mb-2" style={{ fontFamily: "var(--font-heading)" }}>
            开启你的学习旅程
          </h1>
          <p className="text-[var(--on-surface-variant)] text-sm">
            请选择你当前正在准备的考试
          </p>
        </div>

        {loading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="size-8 animate-spin text-[var(--secondary)]" />
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {exams.map((exam) => {
                const selected = selectedId === exam.id
                return (
                  <button
                    key={exam.id}
                    type="button"
                    onClick={() => setSelectedId(exam.id)}
                    className={cn(
                      "group relative flex flex-col items-center gap-3 rounded-2xl border-2 bg-white p-6 text-center transition-all hover:shadow-md",
                      selected
                        ? "border-[var(--secondary)] shadow-md"
                        : "border-transparent ring-1 ring-foreground/10 hover:border-[var(--outline-variant)]"
                    )}
                  >
                    {selected && (
                      <div className="absolute top-3 right-3">
                        <CheckCircle2 className="size-5 text-[var(--secondary)]" />
                      </div>
                    )}
                    <div
                      className={cn(
                        "flex size-14 items-center justify-center rounded-full text-2xl transition-colors",
                        selected
                          ? "bg-[var(--secondary-fixed)]"
                          : "bg-[var(--surface-container-highest)]"
                      )}
                    >
                      {getExamIcon(exam.code)}
                    </div>
                    <div className="text-base font-semibold text-[var(--text-main)]">
                      {exam.name}
                    </div>
                    <div className="text-xs text-[var(--on-surface-variant)] leading-relaxed">
                      {exam.description}
                    </div>
                  </button>
                )
              })}
            </div>

            {exams.length === 0 && (
              <div className="flex flex-col items-center gap-3 py-16 text-[var(--on-surface-variant)]">
                <GraduationCap className="size-12 opacity-40" />
                <p className="text-sm">暂无可用的考试</p>
              </div>
            )}

            <div className="mt-10 flex justify-center">
              <Button
                size="lg"
                disabled={!selectedId || submitting}
                onClick={handleContinue}
                className="min-w-[160px]"
              >
                {submitting && <Loader2 className="size-4 animate-spin" />}
                继续
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
