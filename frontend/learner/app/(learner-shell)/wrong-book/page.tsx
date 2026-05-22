"use client"

import { useEffect, useMemo, useState } from "react"
import { cn } from "@/lib/utils"
import { authFetch, readBrowserAccessToken, readStoredSession } from "@/lib/auth-session"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select"
import {
  BookX,
  ChevronDown,
  ChevronRight,
  Clock,
  Filter,
  Layers3,
  Loader2,
  Shuffle,
  X,
  Check,
  CheckCircle2,
} from "lucide-react"

type WrongItemStatus = "open" | "reviewing" | "mastered"

type WrongItem = {
  id: string
  question_type: string
  stem: string
  options: Array<{ label: string; text: string }>
  user_answer: string
  correct_answer: string
  explanation: string
  error_count: number
  fix_count: number
  first_error_at: string
  last_error_at: string
  status: WrongItemStatus
  subject_id: string
  subject_name: string
  chapter_id: string
  chapter_name: string
  knowledge_points: Array<{ id: string; name: string }>
}

const API_BASE = "/api/v1"

const QUESTION_TYPE_LABELS: Record<string, string> = {
  single_choice: "单选题",
  multiple_choice: "多选题",
  judgment: "判断题",
  true_false: "判断题",
  fill_blank: "填空题",
  essay: "简答题",
  short_answer: "简答题",
}

const STATUS_CONFIG: Record<
  WrongItemStatus,
  { label: string; variant: "destructive" | "secondary" | "default"; dotColor: string }
> = {
  open: { label: "未订正", variant: "destructive", dotColor: "bg-red-500" },
  reviewing: { label: "订正中", variant: "secondary", dotColor: "bg-yellow-500" },
  mastered: { label: "已掌握", variant: "default", dotColor: "bg-green-500" },
}

function filterLabel(value: string, fallback: string, options: Array<{ id: string; name: string }>) {
  if (value === "all") return fallback
  return options.find((item) => item.id === value)?.name ?? fallback
}

function splitLabels(value: string) {
  return value
    .split(/[，,、\s]+/)
    .map((label) => label.trim())
    .filter(Boolean)
}

function answerChipClass(selected: boolean, correct: boolean) {
  if (correct && selected) return "border-[var(--secondary)] bg-[var(--secondary)]/10 text-[var(--secondary)]"
  if (correct) return "border-[var(--secondary)]/40 bg-[var(--secondary)]/6 text-[var(--secondary)]"
  if (selected) return "border-[var(--error)] bg-[var(--error)]/10 text-[var(--error)]"
  return "border-[var(--outline-variant)] bg-[var(--surface)] text-[var(--text-main)]"
}

function WrongQuestionCard({ item }: { item: WrongItem }) {
  const [expanded, setExpanded] = useState(false)
  const statusCfg = STATUS_CONFIG[item.status]
  const userLabels = splitLabels(item.user_answer)
  const correctLabels = splitLabels(item.correct_answer)

  return (
    <Card className="rounded-2xl border-0 bg-[var(--surface-container-lowest)] shadow-sm">
      <CardContent className="flex flex-col gap-4 py-5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant="outline">{QUESTION_TYPE_LABELS[item.question_type] ?? item.question_type}</Badge>
            <Badge variant={statusCfg.variant}>
              <span className={cn("inline-block size-1.5 rounded-full", statusCfg.dotColor)} />
              {statusCfg.label}
            </Badge>
            <span className="text-xs text-muted-foreground">
              {[item.subject_name, item.chapter_name].filter(Boolean).join(" · ")}
            </span>
          </div>
          <div className="flex items-center gap-1 text-xs whitespace-nowrap text-muted-foreground">
            <Clock className="size-3" />
            {item.last_error_at}
          </div>
        </div>

        <p className="text-sm leading-relaxed text-[var(--text-main)]">{item.stem}</p>

        {item.options.length > 0 ? (
          <div className="grid gap-2">
            {item.options.map((option) => {
              const selected = userLabels.includes(option.label)
              const correct = correctLabels.includes(option.label)
              return (
                <div
                  key={option.label}
                  className={cn(
                    "flex items-start gap-3 rounded-xl border px-4 py-3 text-sm transition-colors",
                    answerChipClass(selected, correct),
                  )}
                >
                  <span className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full border text-xs font-medium">
                    {option.label}
                  </span>
                  <span className="flex-1 leading-relaxed">{option.text}</span>
                  {correct ? <CheckCircle2 className="mt-0.5 size-4" /> : null}
                </div>
              )
            })}
          </div>
        ) : null}

        <div className="flex flex-wrap gap-3 text-sm">
          <div className="flex items-center gap-1.5">
            <X className="size-3.5 text-destructive" />
            <span className="text-muted-foreground">你的答案：</span>
            <span className="font-medium text-destructive">{item.user_answer}</span>
          </div>
          <div className="flex items-center gap-1.5">
            <Check className="size-3.5 text-secondary-green" />
            <span className="text-muted-foreground">正确答案：</span>
            <span className="font-medium text-secondary-green">{item.correct_answer}</span>
          </div>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <BookX className="size-3" />
            累计错误 {item.error_count} 次
          </div>
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Check className="size-3" />
            订正 {item.fix_count} 次
          </div>
        </div>

        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
          <span>首次答错：{item.first_error_at}</span>
          <span>最近答错：{item.last_error_at}</span>
          {item.knowledge_points.map((point) => (
            <Badge key={point.id} variant="outline" className="font-normal">
              {point.name}
            </Badge>
          ))}
        </div>

        <Separator />

        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 w-fit px-2 text-xs"
          onClick={() => setExpanded((prev) => !prev)}
        >
          {expanded ? <ChevronDown className="size-3.5" /> : <ChevronRight className="size-3.5" />}
          查看解析
        </Button>

        {expanded && (
          <div className="rounded-xl bg-[var(--surface)] p-3 text-sm leading-relaxed text-[var(--text-muted)]">
            {item.explanation}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function FlashcardDeck({
  items,
  index,
  onNext,
}: {
  items: WrongItem[]
  index: number
  onNext: () => void
}) {
  const [revealed, setRevealed] = useState(false)
  const current = items[index]

  useEffect(() => {
    setRevealed(false)
  }, [index, current?.id])

  if (!current) {
    return (
      <div className="rounded-2xl border border-dashed border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] px-6 py-20 text-center text-sm text-[var(--text-muted)]">
        闪卡复习已完成
      </div>
    )
  }

  const upcoming = items.slice(index + 1, index + 3)
  const userLabels = splitLabels(current.user_answer)
  const correctLabels = splitLabels(current.correct_answer)

  return (
    <div className="relative mx-auto min-h-[520px] max-w-3xl">
      {upcoming
        .map((item, offset) => ({ item, offset }))
        .reverse()
        .map(({ item, offset }) => (
          <Card
            key={item.id}
            className="pointer-events-none absolute inset-x-0 top-0 rounded-3xl border-0 bg-[var(--surface-container-lowest)] shadow-lg"
            style={{
              transform: `translate(${(offset + 1) * 10}px, ${(offset + 1) * 10}px) scale(${1 - (offset + 1) * 0.03})`,
              zIndex: 5 - offset,
              opacity: 0.55 - offset * 0.08,
            }}
          >
            <CardContent className="min-h-[440px] py-6" />
          </Card>
        ))}

      <Card className="relative z-10 rounded-3xl border-0 bg-[var(--surface-container-lowest)] shadow-xl">
        <CardContent className="flex min-h-[440px] flex-col gap-4 p-6 md:p-7">
          <div className="flex items-start justify-between gap-3">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="outline" className="bg-[var(--secondary)]/8 text-[var(--secondary)]">
                <Shuffle className="mr-1 size-3" />
                闪卡模式
              </Badge>
              <Badge variant="outline">{QUESTION_TYPE_LABELS[current.question_type] ?? current.question_type}</Badge>
              <Badge variant={STATUS_CONFIG[current.status].variant}>
                <span className={cn("inline-block size-1.5 rounded-full", STATUS_CONFIG[current.status].dotColor)} />
                {STATUS_CONFIG[current.status].label}
              </Badge>
            </div>
            <span className="text-xs text-[var(--text-muted)]">
              {index + 1} / {items.length}
            </span>
          </div>

          <p className="text-sm leading-6 text-[var(--text-muted)]">
            {[current.subject_name, current.chapter_name].filter(Boolean).join(" · ")}
          </p>
          <p className="text-base leading-relaxed text-[var(--text-main)]">{current.stem}</p>

          {current.options.length > 0 ? (
            <div className="grid gap-2">
              {current.options.map((option) => {
                const selected = userLabels.includes(option.label)
                const correct = correctLabels.includes(option.label)
                return (
                  <div
                    key={option.label}
                    className={cn(
                      "flex items-start gap-3 rounded-xl border px-4 py-3 text-sm transition-colors",
                      answerChipClass(selected, correct),
                    )}
                  >
                    <span className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full border text-xs font-medium">
                      {option.label}
                    </span>
                    <span className="flex-1 leading-relaxed">{option.text}</span>
                    {correct ? <CheckCircle2 className="mt-0.5 size-4" /> : null}
                  </div>
                )
              })}
            </div>
          ) : null}

          {revealed ? (
            <div className="mt-auto space-y-3 rounded-2xl bg-[var(--surface)] p-4">
              <div className="flex flex-wrap gap-3 text-sm">
                <div className="flex items-center gap-1.5">
                  <X className="size-3.5 text-destructive" />
                  <span className="text-muted-foreground">你的答案：</span>
                  <span className="font-medium text-destructive">{current.user_answer}</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <Check className="size-3.5 text-secondary-green" />
                  <span className="text-muted-foreground">正确答案：</span>
                  <span className="font-medium text-secondary-green">{current.correct_answer}</span>
                </div>
              </div>
              <div className="rounded-xl bg-white p-3 text-sm leading-relaxed text-[var(--text-muted)] shadow-sm">
                {current.explanation}
              </div>
            </div>
          ) : (
            <div className="mt-auto rounded-2xl border border-dashed border-[var(--outline-variant)] bg-[var(--surface)] px-4 py-5 text-sm text-[var(--text-muted)]">
              先回忆答案，再翻面看解析。
            </div>
          )}

          <div className="mt-auto flex flex-wrap items-center justify-between gap-3 pt-2">
            <div className="flex flex-wrap gap-2 text-xs text-[var(--text-muted)]">
              {current.knowledge_points.map((point) => (
                <Badge key={point.id} variant="outline" className="font-normal">
                  {point.name}
                </Badge>
              ))}
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" className="rounded-full gap-2" onClick={() => setRevealed((value) => !value)}>
                {revealed ? "收起解析" : "查看答案"}
              </Button>
              <Button className="rounded-full gap-2" onClick={onNext} disabled={!revealed}>
                下一张
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default function WrongBookPage() {
  const [items, setItems] = useState<WrongItem[]>([])
  const [statusFilter, setStatusFilter] = useState<string>("all")
  const [subjectFilter, setSubjectFilter] = useState<string>("all")
  const [chapterFilter, setChapterFilter] = useState<string>("all")
  const [knowledgePointFilter, setKnowledgePointFilter] = useState<string>("all")
  const [viewMode, setViewMode] = useState<"list" | "flashcard">("list")
  const [flashcardIndex, setFlashcardIndex] = useState(0)
  const [flashcardOrder, setFlashcardOrder] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      const token = readBrowserAccessToken()
      if (!token) {
        setLoading(false)
        return
      }

      try {
        const examID = readStoredSession()?.activeExam?.id
        const params = new URLSearchParams({ page_size: "200" })
        if (examID) params.set("exam_id", examID)
        const query = `?${params.toString()}`
        const res = await authFetch(`${API_BASE}/learner/wrong-book${query}`, {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        })
        if (!res.ok) {
          throw new Error("wrong-book request failed")
        }
        const payload = (await res.json()) as { data?: WrongItem[] }
        if (!cancelled) {
          setItems(payload.data ?? [])
          setError(null)
        }
      } catch {
        if (!cancelled) {
          setError("无法加载错题本，请稍后重试。")
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const subjects = Array.from(
    new Map(items.map((i) => [i.subject_id, i.subject_name])).entries(),
    ([id, name]) => ({ id, name })
  ).filter((item) => item.id && item.name)
  const chapters = Array.from(
    new Map(
      items
        .filter((i) => subjectFilter === "all" || i.subject_id === subjectFilter)
        .map((i) => [i.chapter_id, i.chapter_name])
    ).entries(),
    ([id, name]) => ({ id, name })
  ).filter((item) => item.id && item.name)
  const knowledgePoints = Array.from(
    new Map(
      items
        .filter((i) => subjectFilter === "all" || i.subject_id === subjectFilter)
        .filter((i) => chapterFilter === "all" || i.chapter_id === chapterFilter)
        .flatMap((i) => i.knowledge_points.map((point) => [point.id, point.name] as const))
    ).entries(),
    ([id, name]) => ({ id, name })
  ).filter((item) => item.id && item.name)
  const statusOptions = [
    { id: "open", name: "未订正" },
    { id: "reviewing", name: "订正中" },
    { id: "mastered", name: "已掌握" },
  ]

  const filtered = items.filter((item) => {
    if (statusFilter !== "all" && item.status !== statusFilter) return false
    if (subjectFilter !== "all" && item.subject_id !== subjectFilter) return false
    if (chapterFilter !== "all" && item.chapter_id !== chapterFilter) return false
    if (
      knowledgePointFilter !== "all" &&
      !item.knowledge_points.some((point) => point.id === knowledgePointFilter)
    ) {
      return false
    }
    return true
  })

  const filteredSignature = filtered.map((item) => item.id).join("|")

  useEffect(() => {
    if (viewMode !== "flashcard") return
    const ids = filtered.map((item) => item.id)
    const shuffled = [...ids].sort(() => Math.random() - 0.5)
    setFlashcardOrder(shuffled)
    setFlashcardIndex(0)
  }, [filteredSignature, viewMode])

  const flashcardItems = useMemo(
    () =>
      flashcardOrder
        .map((id) => filtered.find((item) => item.id === id))
        .filter((item): item is WrongItem => Boolean(item)),
    [filtered, flashcardOrder],
  )

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8">
      {loading ? (
        <div className="mb-6 flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" />
          正在加载错题记录...
        </div>
      ) : null}
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1
            className="font-heading text-2xl font-bold tracking-tight"
            style={{ fontFamily: "var(--font-heading)" }}
          >
            错题本
          </h1>
          <p className="text-sm text-muted-foreground">
            {loading ? "正在加载错题记录..." : `共 ${filtered.length} 道错题`}
          </p>
        </div>
        <Button
          type="button"
          variant="outline"
          className="rounded-full gap-2"
          onClick={() => setViewMode((mode) => (mode === "list" ? "flashcard" : "list"))}
        >
          <Layers3 className="size-4" />
          {viewMode === "list" ? "闪卡模式" : "列表模式"}
        </Button>
      </div>

      {error ? (
        <div className="mb-6 rounded-lg border border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      ) : null}

      <div className="mb-6 flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
          <Filter className="size-4" />
          <span>筛选</span>
        </div>
        <Select value={subjectFilter} onValueChange={(v) => {
          if (v === null) return
          setSubjectFilter(v)
          setChapterFilter("all")
          setKnowledgePointFilter("all")
        }}>
          <SelectTrigger className="w-36">
            <span className="line-clamp-1">{filterLabel(subjectFilter, "全部科目", subjects)}</span>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部科目</SelectItem>
            {subjects.map((s) => (
              <SelectItem key={s.id} value={s.id}>
                {s.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={chapterFilter} onValueChange={(v) => {
          if (v === null) return
          setChapterFilter(v)
          setKnowledgePointFilter("all")
        }}>
          <SelectTrigger className="w-36">
            <span className="line-clamp-1">{filterLabel(chapterFilter, "全部章节", chapters)}</span>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部章节</SelectItem>
            {chapters.map((chapter) => (
              <SelectItem key={chapter.id} value={chapter.id}>
                {chapter.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={knowledgePointFilter} onValueChange={(v) => {
          if (v !== null) setKnowledgePointFilter(v)
        }}>
          <SelectTrigger className="w-40">
            <span className="line-clamp-1">
              {filterLabel(knowledgePointFilter, "全部知识点", knowledgePoints)}
            </span>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部知识点</SelectItem>
            {knowledgePoints.map((point) => (
              <SelectItem key={point.id} value={point.id}>
                {point.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={statusFilter} onValueChange={(v) => {
          if (v !== null) setStatusFilter(v)
        }}>
          <SelectTrigger className="w-28">
            <span className="line-clamp-1">{filterLabel(statusFilter, "全部状态", statusOptions)}</span>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="open">未订正</SelectItem>
            <SelectItem value="reviewing">订正中</SelectItem>
            <SelectItem value="mastered">已掌握</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {viewMode === "flashcard" ? (
        <div className="py-2">
          {!error && !loading && filtered.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-16 text-muted-foreground">
              <BookX className="size-10 opacity-40" />
              <p className="text-sm">暂无错题记录</p>
            </div>
          ) : flashcardItems.length === 0 ? (
            <div className="flex flex-col items-center gap-2 py-16 text-muted-foreground">
              <Shuffle className="size-10 opacity-40" />
              <p className="text-sm">当前筛选条件下没有可复习的错题</p>
            </div>
          ) : (
            <FlashcardDeck
              items={flashcardItems}
              index={flashcardIndex}
              onNext={() => setFlashcardIndex((current) => Math.min(current + 1, flashcardItems.length))}
            />
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {!error && !loading && filtered.length === 0 && (
            <div className="flex flex-col items-center gap-2 py-16 text-muted-foreground">
              <BookX className="size-10 opacity-40" />
              <p className="text-sm">暂无错题记录</p>
            </div>
          )}
          {filtered.map((item) => (
            <WrongQuestionCard key={item.id} item={item} />
          ))}
        </div>
      )}
    </div>
  )
}
