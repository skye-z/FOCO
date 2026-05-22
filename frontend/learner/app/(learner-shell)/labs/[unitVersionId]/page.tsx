"use client"

import * as React from "react"
import { useRouter, useParams } from "next/navigation"
import { authFetch, readBrowserAccessToken } from "@/lib/auth-session"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"
import { Separator } from "@/components/ui/separator"
import {
  ArrowLeft,
  Check,
  X,
  ChevronRight,
  Loader2,
  FlaskConical,
  Trophy,
  RotateCcw,
  GripVertical,
  Lightbulb,
} from "lucide-react"

const API_BASE = "/api/v1"
const LAB_SESSION_STORAGE_PREFIX = "foco.lab.session"

type StepSchema = {
  id: string
  widget_type: string
  content: Record<string, any>
  initial_state: Record<string, any>
  allowed_actions: Record<string, any>
  evaluation_config: Record<string, any>
  feedback_map: Record<string, any>
  hint_policy: Record<string, any>
  knowledge_point_ids: string[]
  knowledge_point_tags: string[]
}

type UnitView = {
  unit_version_id: string
  title: string
  steps: StepSchema[]
}

type StepFeedback = {
  is_correct: boolean
  allow_continue: boolean
  hint: string
}

function itemKey(item: any, index: number) {
  if (typeof item === "string") return item
  return item?.id || item?.key || item?.text || item?.label || String(index)
}

function itemLabel(item: any, index: number) {
  if (typeof item === "string") return item
  return item?.text || item?.label || item?.name || item?.id || String(index)
}

type HighlightSegment = {
  id: string
  text: string
  selectable: boolean
}

type LabSessionState = {
  attemptId: string | null
  currentStep: number
  feedbacks: Record<number, StepFeedback>
}

function highlightTargets(step: StepSchema) {
  const content = step.content
  if (Array.isArray(content?.items) && content.items.length > 0) {
    return content.items.map((item: any, index: number) => ({
      id: itemKey(item, index),
      text: itemLabel(item, index),
    }))
  }
  const expectedHighlights = step.evaluation_config?.expected_highlights
  if (Array.isArray(expectedHighlights) && expectedHighlights.length > 0) {
    return expectedHighlights.map((text: any, index: number) => ({
      id: String(text ?? index),
      text: String(text ?? ""),
    }))
  }
  return []
}

function charHighlightSegments(source: string): HighlightSegment[] {
  return Array.from(source).map((char, index) => ({
    id: `char-${index}`,
    text: char,
    selectable: !/\s/.test(char),
  }))
}

function selectedTextsFromSegments(segments: HighlightSegment[], markedIds: string[]) {
  const selected = new Set(markedIds)
  const groups: string[] = []
  let buffer = ""

  for (const segment of segments) {
    if (segment.selectable && selected.has(segment.id)) {
      buffer += segment.text
      continue
    }
    if (!segment.selectable && /\s/.test(segment.text) && buffer) {
      buffer += segment.text
      continue
    }
    if (buffer.trim()) groups.push(buffer.trim())
    buffer = ""
  }
  if (buffer.trim()) groups.push(buffer.trim())

  return groups
}

function correctHighlightIds(segments: HighlightSegment[], targets: Array<{ text: string }>) {
  const expected = targets.map((target) => target.text).filter(Boolean)
  const source = segments.map((segment) => segment.text).join("")
  const ids = new Set<string>()

  for (const text of expected) {
    let start = source.indexOf(text)
    while (start !== -1) {
      for (let i = start; i < start + Array.from(text).length; i++) {
        if (segments[i]?.selectable) ids.add(segments[i].id)
      }
      start = source.indexOf(text, start + text.length)
    }
  }

  return ids
}

type CompletionSummary = {
  attempt_id: string
  status: string
  concept_card: {
    id: string
    content: Record<string, any>
  } | null
}

function authHeaders(): Record<string, string> {
  const token = readBrowserAccessToken()
  const headers: Record<string, string> = { "Content-Type": "application/json" }
  if (token) headers["Authorization"] = `Bearer ${token}`
  return headers
}

const WIDGET_LABELS: Record<string, string> = {
  ordering_matching: "排序匹配",
  highlight_marking: "高亮标注",
  formula_builder: "公式构建",
  parameter_lab: "参数实验",
  choice_cloze: "选择填空",
}

function firstText(...values: any[]) {
  for (const value of values) {
    if (typeof value === "string" && value.trim()) return value.trim()
  }
  return ""
}

function formatValue(value: any): string {
  if (Array.isArray(value)) return value.map(formatValue).filter(Boolean).join("、")
  if (value && typeof value === "object") {
    return Object.entries(value)
      .map(([key, val]) => `${key}: ${formatValue(val)}`)
      .join("；")
  }
  if (value === undefined || value === null || value === "") return ""
  return String(value)
}

function optionLabelById(step: StepSchema, id: string) {
  const options = Array.isArray(step.content?.options) ? step.content.options : []
  const option = options.find((item: any, index: number) => itemKey(item, index) === id)
  return option ? itemLabel(option, options.indexOf(option)) : id
}

function correctAnswerText(step: StepSchema) {
  const config = step.evaluation_config ?? {}
  switch (step.widget_type) {
    case "ordering_matching": {
      const order = Array.isArray(config.correct_order) ? config.correct_order : []
      if (order.length > 0) {
        return order.map((id: any) => optionLabelById({ ...step, content: { ...step.content, options: step.content?.items ?? [] } }, String(id))).join(" → ")
      }
      return formatValue(config.correct_pairs || config.correct_mapping)
    }
    case "highlight_marking":
      return formatValue(config.expected_highlights || config.correct_marked_ids || config.correct_marks)
    case "choice_cloze": {
      const values =
        config.correct_selections ||
        config.correct_answers ||
        config.correct_option_ids ||
        (config.correct_option_id ? [config.correct_option_id] : [])
      return Array.isArray(values)
        ? values.map((value: any) => optionLabelById(step, String(value))).join("、")
        : formatValue(values)
    }
    case "parameter_lab":
      return firstText(config.quiz_answer, config.answer, step.content?.quiz_answer) ||
        formatValue(config.expected_state || config.target_state || config.target_range)
    case "formula_builder":
      return formatValue(config.correct_formula || config.required_slots || config.correct_mapping)
    default:
      return ""
  }
}

function explanationText(step: StepSchema, feedback: StepFeedback) {
  const map = step.feedback_map ?? {}
  return firstText(
    feedback.is_correct ? map.correct : map.incorrect,
    feedback.is_correct ? map.success_message : map.failure_message,
    map.explanation,
    map.misconception_explanation,
    map.remediation,
    step.hint_policy?.default_hint,
  )
}

function AnswerReview({ step, feedback }: { step: StepSchema; feedback: StepFeedback }) {
  const answer = correctAnswerText(step)
  const explanation = explanationText(step, feedback)
  if (!answer && !explanation) return null

  return (
    <div className="space-y-2 rounded-lg border bg-muted/20 p-3 text-sm">
      {answer ? (
        <div>
          <p className="mb-1 text-xs font-medium text-muted-foreground">正确答案</p>
          <p className="leading-relaxed text-foreground">{answer}</p>
        </div>
      ) : null}
      {explanation ? (
        <div>
          <p className="mb-1 text-xs font-medium text-muted-foreground">解析</p>
          <p className="leading-relaxed text-muted-foreground">{explanation}</p>
        </div>
      ) : null}
    </div>
  )
}

function OrderingWidget({
  step,
  feedback,
  onSubmit,
  submitting,
}: {
  step: StepSchema
  feedback: StepFeedback | null
  onSubmit: (payload: Record<string, any>) => void
  submitting: boolean
}) {
  const [items, setItems] = React.useState<any[]>([])
  const initialized = React.useRef(false)
  const dragIndex = React.useRef<number | null>(null)

  React.useEffect(() => {
    if (initialized.current) return
    initialized.current = true
    const content = step.content
    if (content && content.items) {
      const list = Array.isArray(content.items) ? [...content.items] : []
      for (let i = list.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1))
        ;[list[i], list[j]] = [list[j], list[i]]
      }
      setItems(list)
    }
  }, [step])

  const moveItem = (index: number, direction: "up" | "down") => {
    const newItems = [...items]
    const targetIndex = direction === "up" ? index - 1 : index + 1
    if (targetIndex < 0 || targetIndex >= newItems.length) return
    ;[newItems[index], newItems[targetIndex]] = [newItems[targetIndex], newItems[index]]
    setItems(newItems)
  }

  const moveItemTo = (fromIndex: number, toIndex: number) => {
    if (feedback || fromIndex === toIndex || fromIndex < 0 || toIndex < 0) return
    setItems((prev) => {
      const next = [...prev]
      const [moved] = next.splice(fromIndex, 1)
      next.splice(toIndex, 0, moved)
      return next
    })
  }

  return (
    <div className="space-y-4">
      {(step.content?.instruction || step.content?.title) && (
        <p className="text-sm text-muted-foreground">{step.content.instruction || step.content.title}</p>
      )}
      <div className="space-y-2">
        {items.map((item, idx) => (
          <div
            key={itemKey(item, idx)}
            draggable={!feedback}
            onDragStart={(event) => {
              dragIndex.current = idx
              event.dataTransfer.effectAllowed = "move"
            }}
            onDragOver={(event) => {
              if (!feedback) event.preventDefault()
            }}
            onDrop={(event) => {
              event.preventDefault()
              if (dragIndex.current === null) return
              moveItemTo(dragIndex.current, idx)
              dragIndex.current = null
            }}
            onDragEnd={() => {
              dragIndex.current = null
            }}
            className={cn(
              "flex items-center gap-2 rounded-lg border bg-card p-3 text-sm transition-all",
              !feedback && "cursor-grab active:cursor-grabbing",
              feedback && "opacity-80",
              feedback?.is_correct && "border-green-300 bg-green-50",
              feedback && !feedback.is_correct && "border-red-300 bg-red-50"
            )}
          >
            <GripVertical className="size-4 text-muted-foreground" />
            <span className="flex-1">{itemLabel(item, idx)}</span>
            {!feedback && (
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-7"
                  onClick={() => moveItem(idx, "up")}
                  disabled={idx === 0}
                >
                  ▲
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-7"
                  onClick={() => moveItem(idx, "down")}
                  disabled={idx === items.length - 1}
                >
                  ▼
                </Button>
              </div>
            )}
          </div>
        ))}
      </div>
      {!feedback && (
        <div className="space-y-2">
          <p className="text-xs text-muted-foreground">拖动卡片调整顺序，也可以用右侧按钮微调。</p>
          <Button
            onClick={() => onSubmit({ ordered_ids: items.map((item, idx) => itemKey(item, idx)) })}
            disabled={submitting}
            className="w-full"
          >
            {submitting ? <Loader2 className="mr-2 size-4 animate-spin" /> : <Check className="mr-2 size-4" />}
            提交答案
          </Button>
        </div>
      )}
    </div>
  )
}

function HighlightWidget({
  step,
  feedback,
  onSubmit,
  submitting,
}: {
  step: StepSchema
  feedback: StepFeedback | null
  onSubmit: (payload: Record<string, any>) => void
  submitting: boolean
}) {
  const [markedIds, setMarkedIds] = React.useState<string[]>([])
  const segments = React.useMemo(() => {
    const content = step.content
    if (!content) return []
    const source = String(content.passage || content.text || "")
    if (source) {
      return charHighlightSegments(source)
    }
    if (content.segments) {
      return Array.isArray(content.segments)
        ? content.segments.map((seg: any, index: number) => ({
            id: itemKey(seg, index),
            text: itemLabel(seg, index),
            selectable: true,
          }))
        : []
    }
    const targets = highlightTargets(step)
    if (targets.length > 0) {
      return targets.map((target) => ({ ...target, selectable: true }))
    }
    return []
  }, [step])

  const toggleMark = (id: string) => {
    if (feedback) return
    setMarkedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  const correctIds = React.useMemo(() => {
    const ids = step.evaluation_config?.correct_marked_ids
    if (Array.isArray(ids) && ids.length > 0) return ids.map((item: any) => String(item))
    return Array.from(correctHighlightIds(segments, highlightTargets(step)))
  }, [step, segments])

  return (
    <div className="space-y-4">
      {(step.content?.instruction || step.content?.title) && (
        <p className="text-sm text-muted-foreground">{step.content.instruction || step.content.title}</p>
      )}
      <div className="rounded-lg border bg-card p-4 text-sm leading-loose">
        {segments.map((seg: HighlightSegment, idx: number) => {
          const isMarked = markedIds.includes(seg.id)
          const isCorrect = feedback && correctIds.includes(seg.id)
          const isWrong = feedback && isMarked && !correctIds.includes(seg.id)
          const isMissed = feedback && !isMarked && correctIds.includes(seg.id)
          if (!seg.selectable) {
            return <span key={seg.id || idx}>{seg.text}</span>
          }
          return (
            <span
              key={seg.id || idx}
              onClick={() => toggleMark(seg.id)}
              className={cn(
                "inline cursor-pointer rounded px-1 transition-colors",
                !feedback && isMarked && "bg-yellow-200 font-medium",
                isCorrect && "bg-green-200",
                isWrong && "bg-red-200 line-through",
                isMissed && "bg-green-100 underline"
              )}
            >
              {seg.text}
            </span>
          )
        })}
      </div>
      {!feedback && (
        <div className="flex items-center justify-between">
          <p className="text-xs text-muted-foreground">点击原文中的文字进行标注</p>
          <Button
            onClick={() => onSubmit({ marked_ids: selectedTextsFromSegments(segments, markedIds) })}
            disabled={submitting || markedIds.length === 0}
          >
            {submitting ? <Loader2 className="mr-2 size-4 animate-spin" /> : <Check className="mr-2 size-4" />}
            提交标注
          </Button>
        </div>
      )}
    </div>
  )
}

function ChoiceClozeWidget({
  step,
  feedback,
  onSubmit,
  submitting,
}: {
  step: StepSchema
  feedback: StepFeedback | null
  onSubmit: (payload: Record<string, any>) => void
  submitting: boolean
}) {
  const [selectedId, setSelectedId] = React.useState<string | null>(null)
  const [selectedIds, setSelectedIds] = React.useState<string[]>([])
  const [blankValues, setBlankValues] = React.useState<string[]>([])
  const hasOptions = Array.isArray(step.content?.options) && step.content.options.length > 0
  const mode =
    step.evaluation_config?.mode ||
    (Array.isArray(step.content?.blanks) && step.content.blanks.length > 0 && !hasOptions
      ? "fill_blank"
      : undefined) ||
    (Array.isArray(step.evaluation_config?.correct_selections) &&
    step.evaluation_config.correct_selections.length > 1
      ? "multi_choice"
      : "single")

  const options = React.useMemo(() => {
    const content = step.content
    if (content?.options && Array.isArray(content.options)) {
      return content.options.map((opt: any, idx: number) => {
        const fallbackLabel = String.fromCharCode(65 + idx)
        if (typeof opt === "string") {
          return { id: opt, label: fallbackLabel, text: opt }
        }
        const text = String(opt?.text || opt?.name || opt?.label || opt?.value || opt?.id || fallbackLabel)
        const rawLabel = String(opt?.label || fallbackLabel)
        return {
          id: String(opt?.id || opt?.key || opt?.value || text),
          label: rawLabel.length <= 3 ? rawLabel : fallbackLabel,
          text,
        }
      })
    }
    return []
  }, [step])
  const blanks = React.useMemo(() => {
    const raw = Array.isArray(step.content?.blanks)
      ? step.content.blanks
      : Array.isArray(step.evaluation_config?.correct_answers)
        ? step.evaluation_config.correct_answers
        : []
    return raw.map((blank: any, index: number) => ({
      id: blank?.id || String(index),
      hint: blank?.hint || blank?.label || `第 ${index + 1} 空`,
    }))
  }, [step])

  React.useEffect(() => {
    setBlankValues((prev) => blanks.map((_, index) => prev[index] ?? ""))
  }, [blanks])

  const toggleMulti = (id: string) => {
    if (feedback) return
    setSelectedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  const correctId = step.evaluation_config?.correct_option_id
  const correctIds = step.evaluation_config?.correct_option_ids
  const correctSelections = Array.isArray(step.evaluation_config?.correct_selections)
    ? step.evaluation_config.correct_selections.map((item: any) => String(item))
    : []
  const instructionText = firstText(step.content?.instruction, step.content?.title)
  const questionText = firstText(step.content?.prompt, step.content?.text)
  const passageText = firstText(step.content?.passage)

  return (
    <div className="space-y-4">
      {(instructionText || questionText || passageText) && (
        <div className="space-y-2">
          {instructionText ? <p className="text-sm text-muted-foreground">{instructionText}</p> : null}
          {questionText ? (
            <div className="rounded-lg border bg-muted/20 p-4 text-sm leading-relaxed text-foreground">
              {questionText}
            </div>
          ) : null}
          {passageText ? (
            <div className="rounded-lg border bg-muted/30 p-4 text-sm leading-relaxed">{passageText}</div>
          ) : null}
        </div>
      )}
      {mode === "fill_blank" && blanks.length > 0 ? (
        <div className="space-y-2">
          {blanks.map((blank, index) => (
            <input
              key={blank.id}
              type="text"
              value={blankValues[index] ?? ""}
              onChange={(event) => {
                if (feedback) return
                const next = [...blankValues]
                next[index] = event.target.value
                setBlankValues(next)
              }}
              disabled={!!feedback}
              placeholder={blank.hint}
              className="w-full rounded-lg border bg-background px-3 py-2 text-sm"
            />
          ))}
        </div>
      ) : (
      <div className="space-y-2">
        {options.map((opt: { id?: string; label?: string; text?: string }, idx: number) => {
          const id = opt.id || String(idx)
          const isSingleSelected = mode !== "multi_choice" && selectedId === id
          const isMultiSelected = mode === "multi_choice" && selectedIds.includes(id)
          const isSelected = isSingleSelected || isMultiSelected
          const optionText = opt.text ? String(opt.text) : ""

          const isCorrectOption =
            feedback &&
            (correctId === id ||
              (Array.isArray(correctIds) && correctIds.includes(id)) ||
              (optionText ? correctSelections.includes(optionText) : false))
          const isWrongSelection = feedback && isSelected && !isCorrectOption

          return (
            <button
              key={id}
              onClick={() => {
                if (feedback) return
                if (mode === "multi_choice") toggleMulti(id)
                else setSelectedId(id)
              }}
              className={cn(
                "flex w-full items-start gap-3 rounded-lg border p-3 text-left text-sm transition-all",
                !feedback && isSelected && "border-primary bg-primary/5",
                isCorrectOption && "border-green-400 bg-green-50",
                isWrongSelection && "border-red-400 bg-red-50"
              )}
            >
              <span
                className={cn(
                  "flex size-6 shrink-0 items-center justify-center rounded-full border text-xs font-medium",
                  !feedback && isSelected && "border-primary bg-primary text-primary-foreground",
                  isCorrectOption && "border-green-500 bg-green-500 text-white",
                  isWrongSelection && "border-red-500 bg-red-500 text-white"
                )}
              >
                {opt.label || String.fromCharCode(65 + idx)}
              </span>
              <span className="min-w-0 flex-1 leading-relaxed">{opt.text}</span>
            </button>
          )
        })}
      </div>
      )}
      {!feedback && (
        <Button
          onClick={() => {
            if (mode === "multi_choice") onSubmit({ selected_option_ids: selectedIds })
            else if (mode === "fill_blank") onSubmit({ blank_values: blankValues.map((value) => value.trim()) })
            else onSubmit({ selected_option_id: selectedId })
          }}
          disabled={
            submitting ||
            (mode === "multi_choice"
              ? selectedIds.length === 0
              : mode === "fill_blank"
                ? blankValues.every((value) => !value.trim())
                : !selectedId)
          }
          className="w-full"
        >
          {submitting ? <Loader2 className="mr-2 size-4 animate-spin" /> : <Check className="mr-2 size-4" />}
          提交答案
        </Button>
      )}
    </div>
  )
}

function ParameterLabWidget({
  step,
  feedback,
  onSubmit,
  submitting,
}: {
  step: StepSchema
  feedback: StepFeedback | null
  onSubmit: (payload: Record<string, any>) => void
  submitting: boolean
}) {
  const params = React.useMemo(() => {
    const content = step.content
    if (content?.parameters && Array.isArray(content.parameters)) return content.parameters
    if (content?.params && Array.isArray(content.params)) return content.params.map((param: any, idx: number) => ({
      id: param.id || param.key || String(idx),
      key: param.key || param.name || param.id || String(idx),
      label: param.label || param.name || param.id || String(idx),
      min: param.min,
      max: param.max,
      step: param.step,
      default_value: param.default_value ?? param.default,
      type: param.type,
      options: param.options,
    }))
    return []
  }, [step])

  const [state, setState] = React.useState<Record<string, any>>({})
  const [answer, setAnswer] = React.useState("")

  React.useEffect(() => {
    const init: Record<string, any> = {}
    const initialValues =
      step.initial_state?.values && typeof step.initial_state.values === "object"
        ? step.initial_state.values
        : step.initial_state
    if (initialValues) Object.assign(init, initialValues)
    params.forEach((p: { id?: string; key?: string; name?: string; default_value?: any; default?: any }) => {
      const k = p.key || p.name || p.id || ""
      if (k && init[k] === undefined) init[k] = p.default_value ?? p.default ?? 0
    })
    setState(init)
    setAnswer("")
  }, [step, params])

  const updateParam = (key: string, value: any) => {
    if (feedback) return
    setState((prev) => ({ ...prev, [key]: value }))
  }
  const targetMetric = firstText(step.content?.target_metric, step.content?.formula, step.content?.description)
  const quizQuestion = firstText(step.content?.quiz_question, step.evaluation_config?.quiz_question)
  const targetAnswer = correctAnswerText(step)
  const requiresAnswer = Boolean(firstText(step.evaluation_config?.quiz_answer, step.content?.quiz_answer))

  return (
    <div className="space-y-4">
      {(step.content?.instruction || step.content?.title) && (
        <p className="text-sm text-muted-foreground">{step.content.instruction || step.content.title}</p>
      )}
      {(targetMetric || quizQuestion || targetAnswer) && (
        <div className="space-y-2 rounded-lg border bg-muted/20 p-3 text-sm">
          {targetMetric ? (
            <p>
              <span className="font-medium">实验目标：</span>
              {targetMetric}
            </p>
          ) : null}
          {quizQuestion ? (
            <p>
              <span className="font-medium">观察问题：</span>
              {quizQuestion}
            </p>
          ) : null}
          {targetAnswer ? (
            <p className="text-muted-foreground">
              {requiresAnswer ? "调整参数并填写观察答案后提交。" : "调整参数后提交，系统会按目标状态判断影响方向与范围。"}
            </p>
          ) : null}
        </div>
      )}
      <div className="space-y-4">
        {params.map((p: { id?: string; key?: string; label?: string; type?: string; min?: number; max?: number; step?: number; options?: any[]; default_value?: any }, idx: number) => {
          const k = p.key || p.id || String(idx)
          const val = state[k]
          return (
            <div key={k} className="space-y-1.5">
              <label className="text-sm font-medium">{p.label || k}</label>
              {p.type === "slider" || (!p.type && p.min !== undefined) ? (
                <div className="flex items-center gap-3">
                  <input
                    type="range"
                    min={p.min ?? 0}
                    max={p.max ?? 100}
                    step={p.step ?? 1}
                    value={Number(val) || 0}
                    onChange={(e) => updateParam(k, Number(e.target.value))}
                    disabled={!!feedback}
                    className="flex-1"
                  />
                  <span className="min-w-[3rem] text-right text-sm font-mono">{val ?? p.default_value ?? 0}</span>
                </div>
              ) : p.type === "select" && Array.isArray(p.options) ? (
                <select
                  value={String(val ?? "")}
                  onChange={(e) => updateParam(k, e.target.value)}
                  disabled={!!feedback}
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                >
                  {p.options.map((o: any, oi: number) => (
                    <option key={oi} value={String(typeof o === "object" ? o.value : o)}>
                      {typeof o === "object" ? o.label : o}
                    </option>
                  ))}
                </select>
              ) : (
                <input
                  type="text"
                  value={String(val ?? "")}
                  onChange={(e) => updateParam(k, e.target.value)}
                  disabled={!!feedback}
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                />
              )}
            </div>
          )
        })}
      </div>
      {Object.keys(state).length > 0 && (
        <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
          {Object.entries(state).map(([key, value]) => (
            <Badge key={key} variant="outline" className="font-normal">
              {key}: {formatValue(value)}
            </Badge>
          ))}
        </div>
      )}
      {requiresAnswer ? (
        <div className="space-y-1.5">
          <label className="text-sm font-medium">你的观察答案</label>
          <input
            type="text"
            value={answer}
            onChange={(event) => {
              if (feedback) return
              setAnswer(event.target.value)
            }}
            disabled={!!feedback}
            placeholder="填写计算结果或影响方向"
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
          />
        </div>
      ) : null}
      {!feedback && (
        <Button
          onClick={() => onSubmit({ state, answer: answer.trim() })}
          disabled={submitting || (requiresAnswer && !answer.trim())}
          className="w-full"
        >
          {submitting ? <Loader2 className="mr-2 size-4 animate-spin" /> : <Check className="mr-2 size-4" />}
          提交实验结果
        </Button>
      )}
    </div>
  )
}

function FormulaBuilderWidget({
  step,
  feedback,
  onSubmit,
  submitting,
}: {
  step: StepSchema
  feedback: StepFeedback | null
  onSubmit: (payload: Record<string, any>) => void
  submitting: boolean
}) {
  const slots = React.useMemo(() => {
    const content = step.content
    if (content?.slots && Array.isArray(content.slots)) return content.slots
    return []
  }, [step])

  const [slotValues, setSlotValues] = React.useState<Record<string, string>>({})

  const updateSlot = (key: string, value: string) => {
    if (feedback) return
    setSlotValues((prev) => ({ ...prev, [key]: value }))
  }

  return (
    <div className="space-y-4">
      {(step.content?.instruction || step.content?.title) && (
        <p className="text-sm text-muted-foreground">{step.content.instruction || step.content.title}</p>
      )}
      {step.content?.formula_template || step.content?.answer || step.content?.formula ? (
        <div className="rounded-lg border bg-muted/30 p-4 text-center font-mono text-lg">
          {step.content.formula_template || step.content.answer || step.content.formula}
        </div>
      ) : null}
      <div className="space-y-3">
        {slots.map((slot: { id?: string; key?: string; label?: string; options?: any[] }, idx: number) => {
          const k = slot.key || slot.id || String(idx)
          const opts = Array.isArray(slot.options) ? slot.options : []
          return (
            <div key={k} className="space-y-1.5">
              <label className="text-sm font-medium">{slot.label || k}</label>
              {opts.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {opts.map((o: any, oi: number) => {
                    const oVal = typeof o === "object" ? o.value : o
                    const oLabel = typeof o === "object" ? o.label : o
                    const isSelected = slotValues[k] === String(oVal)
                    const isCorrect = feedback && step.evaluation_config?.required_slots?.[k] === String(oVal)
                    const isWrong = feedback && isSelected && !isCorrect
                    return (
                      <button
                        key={oi}
                        onClick={() => updateSlot(k, String(oVal))}
                        className={cn(
                          "rounded-md border px-3 py-1.5 text-sm transition-all",
                          !feedback && isSelected && "border-primary bg-primary/10",
                          isCorrect && "border-green-500 bg-green-50 text-green-700",
                          isWrong && "border-red-500 bg-red-50 text-red-700"
                        )}
                      >
                        {oLabel}
                      </button>
                    )
                  })}
                </div>
              ) : (
                <input
                  type="text"
                  value={slotValues[k] || ""}
                  onChange={(e) => updateSlot(k, e.target.value)}
                  disabled={!!feedback}
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm"
                />
              )}
            </div>
          )
        })}
      </div>
      {!feedback && (
        <Button
          onClick={() => onSubmit({ slot_values: slotValues })}
          disabled={submitting}
          className="w-full"
        >
          {submitting ? <Loader2 className="mr-2 size-4 animate-spin" /> : <Check className="mr-2 size-4" />}
          提交公式
        </Button>
      )}
    </div>
  )
}

function StepWidget({
  step,
  feedback,
  onSubmit,
  submitting,
}: {
  step: StepSchema
  feedback: StepFeedback | null
  onSubmit: (payload: Record<string, any>) => void
  submitting: boolean
}) {
  switch (step.widget_type) {
    case "ordering_matching":
      return <OrderingWidget step={step} feedback={feedback} onSubmit={onSubmit} submitting={submitting} />
    case "highlight_marking":
      return <HighlightWidget step={step} feedback={feedback} onSubmit={onSubmit} submitting={submitting} />
    case "choice_cloze":
      return <ChoiceClozeWidget step={step} feedback={feedback} onSubmit={onSubmit} submitting={submitting} />
    case "parameter_lab":
      return <ParameterLabWidget step={step} feedback={feedback} onSubmit={onSubmit} submitting={submitting} />
    case "formula_builder":
      return <FormulaBuilderWidget step={step} feedback={feedback} onSubmit={onSubmit} submitting={submitting} />
    default:
      return (
        <div className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
          暂不支持的交互类型: {step.widget_type}
        </div>
      )
  }
}

function CompletionScreen({
  summary,
  onBack,
}: {
  summary: CompletionSummary
  onBack: () => void
}) {
  const card = summary.concept_card
  return (
    <div className="mx-auto max-w-lg space-y-6 py-12 text-center">
      <div className="mx-auto flex size-20 items-center justify-center rounded-full bg-green-100">
        <Trophy className="size-10 text-green-600" />
      </div>
      <h2 className="text-2xl font-bold">单元完成!</h2>
      <p className="text-muted-foreground">你已完成所有交互步骤</p>

      {card?.content && (
        <Card className="text-left">
          <CardContent className="space-y-3 p-5">
            <h3 className="flex items-center gap-2 font-semibold">
              <FlaskConical className="size-4" />
              概念卡片
            </h3>
            {card.content.title && <p className="font-medium">{card.content.title}</p>}
            {card.content.summary && (
              <p className="text-sm text-muted-foreground leading-relaxed">{card.content.summary}</p>
            )}
            {card.content.key_takeaways && Array.isArray(card.content.key_takeaways) && (
              <ul className="space-y-1">
                {card.content.key_takeaways.map((t: string, i: number) => (
                  <li key={i} className="flex items-start gap-2 text-sm">
                    <Check className="mt-0.5 size-4 shrink-0 text-green-500" />
                    {t}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      )}

      <Button onClick={onBack} className="w-full">
        <RotateCcw className="mr-2 size-4" />
        返回实验室列表
      </Button>
    </div>
  )
}

export default function InteractiveLabPage() {
  const router = useRouter()
  const params = useParams<{ unitVersionId: string }>()
  const unitVersionId = params.unitVersionId

  const [unit, setUnit] = React.useState<UnitView | null>(null)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  const [attemptId, setAttemptId] = React.useState<string | null>(null)
  const [currentStep, setCurrentStep] = React.useState(0)
  const [feedbacks, setFeedbacks] = React.useState<Record<number, StepFeedback>>({})
  const [submitting, setSubmitting] = React.useState(false)
  const [completed, setCompleted] = React.useState<CompletionSummary | null>(null)
  const restoredSession = React.useRef(false)
  const sessionStorageKey = `${LAB_SESSION_STORAGE_PREFIX}.${unitVersionId}`

  React.useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const res = await authFetch(`${API_BASE}/learner/interactive-units/${unitVersionId}`, {
          headers: authHeaders(),
        })
        if (!res.ok) throw new Error("Failed to load unit")
        const json = await res.json()
        if (!cancelled) setUnit(json.data ?? json)
      } catch {
        if (!cancelled) setError("无法加载交互单元")
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [unitVersionId])

  React.useEffect(() => {
    if (typeof window === "undefined") return
    const raw = window.localStorage.getItem(sessionStorageKey)
    if (raw) {
      try {
        const saved = JSON.parse(raw) as LabSessionState
        if (saved.attemptId) setAttemptId(saved.attemptId)
        if (Number.isFinite(saved.currentStep)) setCurrentStep(Math.max(0, saved.currentStep))
        if (saved.feedbacks) setFeedbacks(saved.feedbacks)
      } catch {
        window.localStorage.removeItem(sessionStorageKey)
      }
    }
    restoredSession.current = true
  }, [sessionStorageKey])

  React.useEffect(() => {
    if (!unit) return
    setCurrentStep((step) => Math.min(step, Math.max(unit.steps.length - 1, 0)))
  }, [unit])

  React.useEffect(() => {
    if (typeof window === "undefined" || !restoredSession.current) return
    if (completed) {
      window.localStorage.removeItem(sessionStorageKey)
      return
    }
    if (!attemptId) return
    const state: LabSessionState = { attemptId, currentStep, feedbacks }
    window.localStorage.setItem(sessionStorageKey, JSON.stringify(state))
  }, [attemptId, currentStep, feedbacks, completed, sessionStorageKey])

  const startAttempt = React.useCallback(async () => {
    try {
      const res = await authFetch(`${API_BASE}/learner/interactive-units/${unitVersionId}/attempts`, {
        method: "POST",
        headers: authHeaders(),
      })
      if (!res.ok) throw new Error("Failed to start attempt")
      const json = await res.json()
      setAttemptId(json.data?.id ?? json.data?.attempt_id)
      setCurrentStep(0)
      setFeedbacks({})
      setCompleted(null)
    } catch {
      setError("无法开始尝试，请稍后重试")
    }
  }, [unitVersionId])

  const submitAction = React.useCallback(
    async (stepId: string, payload: Record<string, any>) => {
      if (!attemptId) return
      setSubmitting(true)
      try {
        const res = await authFetch(
          `${API_BASE}/learner/interactive-unit-attempts/${attemptId}/steps/${stepId}/actions`,
          {
            method: "POST",
            headers: authHeaders(),
            body: JSON.stringify(payload),
          }
        )
        if (!res.ok) throw new Error("Submit failed")
        const json = await res.json()
        const fb: StepFeedback = json.data ?? json
        setFeedbacks((prev) => ({ ...prev, [currentStep]: fb }))
      } catch {
        setFeedbacks((prev) => ({
          ...prev,
          [currentStep]: { is_correct: false, allow_continue: true, hint: "提交失败，请继续" },
        }))
      } finally {
        setSubmitting(false)
      }
    },
    [attemptId, currentStep]
  )

  const goNext = React.useCallback(() => {
    if (!unit) return
    if (currentStep < unit.steps.length - 1) {
      setCurrentStep((s) => s + 1)
    }
  }, [unit, currentStep])

  const finishAttempt = React.useCallback(async () => {
    if (!attemptId) return
    setSubmitting(true)
    try {
      const res = await authFetch(`${API_BASE}/learner/interactive-unit-attempts/${attemptId}/complete`, {
        method: "POST",
        headers: authHeaders(),
      })
      if (!res.ok) throw new Error("Complete failed")
      const json = await res.json()
      setCompleted(json.data ?? json)
    } catch {
      setCompleted({
        attempt_id: attemptId,
        status: "completed",
        concept_card: null,
      })
    } finally {
      setSubmitting(false)
    }
  }, [attemptId])

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-3">
        <p className="text-sm text-destructive">{error}</p>
        <Button variant="outline" onClick={() => router.push("/labs")}>
          <ArrowLeft className="mr-2 size-4" />
          返回列表
        </Button>
      </div>
    )
  }

  if (!unit) return null

  if (completed) {
    return (
      <div className="min-h-screen p-6">
        <CompletionScreen summary={completed} onBack={() => router.push("/labs")} />
      </div>
    )
  }

  if (!attemptId) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 px-6">
        <div className="flex size-16 items-center justify-center rounded-full bg-teal-100">
          <FlaskConical className="size-8 text-teal-600" />
        </div>
        <h1 className="text-xl font-bold">{unit.title}</h1>
        <p className="text-center text-sm text-muted-foreground">
          共 {unit.steps.length} 个交互步骤
        </p>
        <div className="flex flex-wrap justify-center gap-2">
          {unit.steps.map((s, i) => (
            <Badge key={s.id} variant="outline" className="text-xs">
              {i + 1}. {WIDGET_LABELS[s.widget_type] || s.widget_type}
            </Badge>
          ))}
        </div>
        <Button size="lg" onClick={startAttempt}>
          开始学习
        </Button>
        <Button variant="ghost" onClick={() => router.push("/labs")}>
          <ArrowLeft className="mr-2 size-4" />
          返回列表
        </Button>
      </div>
    )
  }

  const step = unit.steps[currentStep]
  const fb = feedbacks[currentStep] ?? null
  const progress = ((currentStep + (fb ? 1 : 0)) / unit.steps.length) * 100
  const isLastStep = currentStep === unit.steps.length - 1
  const canContinue = Boolean(fb)

  return (
    <div className="min-h-screen">
      <div className="sticky top-0 z-10 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex max-w-2xl items-center gap-3 px-4 py-3">
          <Button variant="ghost" size="icon" className="size-8" onClick={() => router.push("/labs")}>
            <ArrowLeft className="size-4" />
          </Button>
          <div className="flex-1">
            <p className="text-sm font-medium truncate">{unit.title}</p>
            <p className="text-xs text-muted-foreground">
              步骤 {currentStep + 1} / {unit.steps.length}
            </p>
          </div>
          <Badge variant="outline" className="text-xs shrink-0">
            {WIDGET_LABELS[step.widget_type]}
          </Badge>
        </div>
        <Progress value={progress} className="h-1 rounded-none" />
      </div>

      <div className="mx-auto max-w-2xl space-y-6 px-4 py-6">
        <Card>
          <CardContent className="space-y-5 p-5">
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span>步骤 {currentStep + 1}</span>
              <Separator orientation="vertical" className="h-3" />
              <span>{WIDGET_LABELS[step.widget_type]}</span>
            </div>

            <StepWidget
              step={step}
              feedback={fb}
              onSubmit={(payload) => submitAction(step.id, payload)}
              submitting={submitting}
            />

            {fb && (
              <div className="space-y-3">
                <div
                  className={cn(
                    "flex items-center gap-2 rounded-lg p-3 text-sm",
                    fb.is_correct ? "bg-green-50 text-green-700" : "bg-red-50 text-red-700"
                  )}
                >
                  {fb.is_correct ? <Check className="size-4" /> : <X className="size-4" />}
                  {fb.is_correct ? "回答正确!" : "回答不正确"}
                </div>
                {fb.hint && (
                  <div className="flex items-start gap-2 rounded-lg bg-amber-50 p-3 text-sm text-amber-700">
                    <Lightbulb className="mt-0.5 size-4 shrink-0" />
                    {fb.hint}
                  </div>
                )}
                <AnswerReview step={step} feedback={fb} />
                {canContinue && (
                  <div className="flex justify-end">
                    <Button
                      onClick={isLastStep ? finishAttempt : goNext}
                      disabled={submitting}
                    >
                      {isLastStep ? (
                        <>
                          <Trophy className="mr-2 size-4" />
                          完成单元
                        </>
                      ) : (
                        <>
                          下一题
                          <ChevronRight className="ml-1 size-4" />
                        </>
                      )}
                    </Button>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        <div className="flex justify-center gap-1">
          {unit.steps.map((_, i) => (
            <button
              key={i}
              onClick={() => {
                if (i <= currentStep || (i === currentStep + 1 && feedbacks[currentStep]?.allow_continue)) {
                  setCurrentStep(i)
                }
              }}
              className={cn(
                "size-2.5 rounded-full transition-colors",
                i === currentStep && "bg-primary",
                i < currentStep && "bg-primary/50",
                i > currentStep && "bg-muted-foreground/20"
              )}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
