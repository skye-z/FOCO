export type TreeNode = {
  id: string;
  code: string;
  name: string;
  type: "exam" | "subject" | "chapter";
  next_exam_date?: string;
  next_next_exam_date?: string;
  countdown_days?: number | null;
  children?: TreeNode[];
};

export type KnowledgePoint = {
  id: string;
  code: string;
  name: string;
};

export type QuestionCard = {
  id: string;
  exam_id: string;
  subject_name: string;
  chapter_name: string | null;
  subject_id: string;
  chapter_id: string | null;
  status: string;
  question_type: string;
  difficulty: number;
  version_no: number;
  version_id: string;
  stem_preview: string;
  published_version_no?: number | null;
  draft_version_no?: number | null;
  has_unpublished_draft?: boolean;
};

export type QuestionVersionSummary = {
  version_id: string;
  question_id: string;
  version_no: number;
  status: string;
  published_at?: string | null;
  updated_at: string;
  publish_note?: string | null;
  is_current: boolean;
  is_published: boolean;
};

export type VersionDetail = {
  version_id: string;
  question_id: string;
  exam_id: string;
  subject_id: string;
  chapter_id: string | null;
  question_type: string;
  difficulty: number;
  version_no: number;
  status: string;
  stem: string;
  options: string;
  correct_answer: string;
  explanation: string;
  knowledge_point_ids: string[];
};

export const DIFFICULTY_OPTIONS = [
  { value: "1", label: "简单" },
  { value: "2", label: "较易" },
  { value: "3", label: "中等" },
  { value: "4", label: "较难" },
  { value: "5", label: "困难" },
];

export function difficultyLabel(d: number) {
  const opt = DIFFICULTY_OPTIONS.find((o) => parseInt(o.value) === d);
  return opt?.label ?? "中等";
}

export function difficultyColor(d: number) {
  if (d <= 2) return "bg-emerald-50 text-emerald-700";
  if (d <= 3) return "bg-amber-50 text-amber-700";
  return "bg-red-50 text-red-700";
}

export function statusLabel(s: string) {
  switch (s) {
    case "draft":
      return "草稿";
    case "published":
    case "active":
      return "已发布";
    case "archived":
      return "已归档";
    default:
      return s;
  }
}

export function statusColor(s: string) {
  switch (s) {
    case "draft":
      return "bg-gray-100 text-gray-600";
    case "active":
    case "published":
      return "bg-blue-50 text-blue-700";
    case "archived":
      return "bg-gray-50 text-gray-400";
    default:
      return "bg-gray-100 text-gray-600";
  }
}

export function typeLabel(t: string) {
  switch (t) {
    case "single_choice":
      return "单选";
    case "multiple_choice":
      return "多选";
    case "judgment":
      return "判断";
    case "fill_blank":
      return "填空";
    case "essay":
      return "简答";
    default:
      return t || "未设置";
  }
}

export function typeColor(t: string) {
  switch (t) {
    case "single_choice":
      return "bg-violet-50 text-violet-700";
    case "multiple_choice":
      return "bg-purple-50 text-purple-700";
    case "judgment":
      return "bg-cyan-50 text-cyan-700";
    default:
      return "bg-gray-100 text-gray-600";
  }
}

export function questionStatusLabel(question: QuestionCard) {
  if (question.published_version_no || question.status === "published")
    return "已发布";
  if (question.has_unpublished_draft) return "草稿";
  return statusLabel(question.status);
}

export function questionStatusColor(question: QuestionCard) {
  if (question.published_version_no || question.status === "published")
    return "bg-blue-50 text-blue-700";
  if (question.has_unpublished_draft) return "bg-gray-100 text-gray-700";
  return statusColor(question.status);
}

import { toast } from "sonner";

export function showActionError(title: string, description?: string) {
  toast.error(title, {
    description: description ?? "请稍后重试，或刷新页面后再次操作。",
  });
}

export function jsonField(raw: string, key: string): string {
  try {
    const obj = JSON.parse(raw);
    if (typeof obj === "string") return obj;
    if (key === "answer") {
      if (Array.isArray(obj.selected_option_ids))
        return obj.selected_option_ids[0] ?? raw;
      if (obj.answer !== undefined) return obj.answer;
    }
    return obj[key] ?? raw;
  } catch {
    return raw;
  }
}

export type InteractiveUnitSummary = {
  id: string;
  title: string;
  exam_id: string;
  subject_id: string;
  subject_name?: string;
  status: string;
  step_count: number;
  version_no: number;
  version_id: string;
  published_version_no?: number | null;
  has_unpublished_draft?: boolean;
  updated_at: string;
};

export function widgetTypeLabel(w: string): string {
  switch (w) {
    case "ordering_matching": return "排序匹配";
    case "formula_builder": return "公式构建";
    case "parameter_lab": return "参数调节";
    case "highlight_marking": return "高亮标记";
    case "choice_cloze": return "选择填空";
    default: return w;
  }
}

export function jsonToOptions(raw: string): string[] {
  try {
    const obj = JSON.parse(raw);
    if (typeof obj !== "object" || obj === null) return ["", "", "", ""];
    if (Array.isArray(obj.choices)) {
      const result: string[] = [];
      for (let i = 0; i < 4; i++) {
        result.push(obj.choices[i]?.text ?? "");
      }
      return result;
    }
    const keys = Object.keys(obj).sort();
    const result: string[] = [];
    for (let i = 0; i < 4; i++) {
      const k = keys[i] ?? String.fromCharCode(65 + i);
      result.push(obj[k] ?? "");
    }
    return result;
  } catch {
    return ["", "", "", ""];
  }
}
