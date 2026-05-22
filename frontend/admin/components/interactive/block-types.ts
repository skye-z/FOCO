import type { StepSchema } from "./step-list";

export type VisualBlockType =
  | "formula_drag"
  | "parameter_chart"
  | "reasoning_sort"
  | "concept_match"
  | "condition_mark"
  | "choice"
  | "fill_blank";

export type VisualBlockDefinition = {
  type: VisualBlockType;
  label: string;
  description: string;
  widgetType: StepSchema["widget_type"];
};

export type BlockCompleteness = {
  complete: boolean;
  missing: string[];
};

function getContent(step: StepSchema) {
  return step.content ?? {};
}

function asArray(value: any) {
  return Array.isArray(value) ? value : [];
}

function firstString(...values: any[]) {
  for (const value of values) {
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return "";
}

export const VISUAL_BLOCKS: VisualBlockDefinition[] = [
  {
    type: "formula_drag",
    label: "拖拽公式组件",
    description: "用变量槽位和候选变量可视化搭建公式。",
    widgetType: "formula_builder",
  },
  {
    type: "parameter_chart",
    label: "调整图表 计算参数",
    description: "通过参数滑动和图表观察结果变化。",
    widgetType: "parameter_lab",
  },
  {
    type: "reasoning_sort",
    label: "排序推理链",
    description: "按正确逻辑顺序重建推理步骤。",
    widgetType: "ordering_matching",
  },
  {
    type: "concept_match",
    label: "匹配概念卡片",
    description: "把概念与定义、案例或结论正确配对。",
    widgetType: "ordering_matching",
  },
  {
    type: "condition_mark",
    label: "标记关键条件",
    description: "从题干中标出影响判断的关键条件。",
    widgetType: "highlight_marking",
  },
  {
    type: "choice",
    label: "选择",
    description: "通过可视化选项判断关键概念或结论。",
    widgetType: "choice_cloze",
  },
  {
    type: "fill_blank",
    label: "填空",
    description: "填写变量、中间值或最终结论。",
    widgetType: "choice_cloze",
  },
];

export function getBlockDefinition(type: VisualBlockType) {
  return VISUAL_BLOCKS.find((block) => block.type === type);
}

export function createDefaultVisualStep(type: VisualBlockType): StepSchema {
  const definition = getBlockDefinition(type);
  const widgetType = definition?.widgetType ?? "ordering_matching";

  return {
    widget_type: widgetType,
    content: {
      ...createDefaultContent(type),
      visual_block_type: type,
      title: "",
    },
    initial_state: {},
    allowed_actions: {},
    evaluation_config: createDefaultEvaluationConfig(type),
    feedback_map: {},
    hint_policy: {},
    knowledge_point_ids: [],
    knowledge_point_tags: [],
  };
}

export function getStepVisualBlockType(step: StepSchema): VisualBlockType {
  const content = getContent(step);
  const explicit = content.visual_block_type as VisualBlockType | undefined;
  if (explicit) return explicit;

  switch (step.widget_type) {
    case "formula_builder":
      return "formula_drag";
    case "parameter_lab":
      return "parameter_chart";
    case "highlight_marking":
      return "condition_mark";
    case "choice_cloze":
      if (asArray(content.blanks).length > 0 && asArray(content.options).length === 0) {
        return "fill_blank";
      }
      return "choice";
    case "ordering_matching":
      if (asArray(content.pairs).length > 0 && asArray(content.items).length === 0) {
        return "concept_match";
      }
      return "reasoning_sort";
    default:
      return "reasoning_sort";
  }
}

export function getStepTitle(step: StepSchema) {
  const content = getContent(step);
  return (
    firstString(
      content.title,
      content.instruction,
      content.prompt,
      content.text,
      content.formula,
      content.answer,
      content.description,
    ) || "未命名交互块"
  );
}

export function getStepSummary(step: StepSchema) {
  const content = getContent(step);
  const type = getStepVisualBlockType(step);

  if (type === "formula_drag") {
    return firstString(content.formula_template, content.answer, content.formula)
      ? `公式: ${firstString(content.formula_template, content.answer, content.formula)}`
      : "尚未配置块内容";
  }
  if (type === "parameter_chart") {
    const parameters = asArray(content.parameters).length > 0 ? asArray(content.parameters) : asArray(content.params);
    if (parameters.length > 0) {
      return `参数 ${parameters.length} 个`;
    }
    return firstString(content.formula, content.quiz_question, content.description) || "尚未配置块内容";
  }
  if (type === "reasoning_sort") {
    const items = asArray(content.items);
    if (items.length > 0) {
      return `步骤项 ${items.length} 个`;
    }
  }
  if (type === "concept_match") {
    const pairs = asArray(content.pairs);
    if (pairs.length > 0) {
      return `配对 ${pairs.length} 组`;
    }
  }
  if (type === "condition_mark") {
    const passage = firstString(content.passage, content.text);
    if (passage) {
      return `题干 ${passage.length} 字`;
    }
  }
  if (type === "choice") {
    const options = asArray(content.options);
    if (options.length > 0) {
      return `选项 ${options.length} 个`;
    }
    const blanks = asArray(content.blanks);
    if (blanks.length > 0) {
      return `填空 ${blanks.length} 个`;
    }
  }
  if (type === "fill_blank") {
    const blanks = asArray(content.blanks);
    if (blanks.length > 0) {
      return `填空 ${blanks.length} 个`;
    }
  }

  return "尚未配置块内容";
}

export function evaluateBlockCompleteness(step: StepSchema): BlockCompleteness {
  const missing: string[] = [];
  const content = getContent(step);
  const type = getStepVisualBlockType(step);

  if (!getStepTitle(step)) {
    missing.push("块标题");
  }
  switch (type) {
    case "formula_drag":
      if (!firstString(content.formula_template, content.answer, content.formula)) missing.push("公式模板");
      if (!asArray(content.slots).length) missing.push("公式槽位");
      if (!asArray(content.variable_bank).length && !asArray(content.slots).some((slot) => asArray(slot?.options).length > 0)) {
        missing.push("候选变量");
      }
      break;
    case "parameter_chart":
      if (!firstString(content.target_metric, content.formula, content.quiz_question)) missing.push("目标指标");
      if (!asArray(content.parameters).length && !asArray(content.params).length) missing.push("参数列表");
      break;
    case "reasoning_sort":
      if (!asArray(content.items).length) missing.push("推理步骤");
      break;
    case "concept_match":
      if (!asArray(content.pairs).length) missing.push("配对项");
      break;
    case "condition_mark":
      if (!firstString(content.passage, content.text)) missing.push("题干文本");
      if (!asArray(content.items).length && !asArray(content.expected_highlights).length) missing.push("关键条件");
      break;
    case "choice":
      if (!firstString(content.prompt, content.text, content.instruction)) missing.push("题干");
      if (!asArray(content.options).length && !asArray(content.correct_selections).length) missing.push("选项");
      break;
    case "fill_blank":
      if (!firstString(content.prompt, content.text, content.instruction)) missing.push("题干");
      if (!asArray(content.blanks).length && !(typeof content.blanks === "number" && content.blanks > 0) && !asArray(content.correct_selections).length) {
        missing.push("填空项");
      }
      break;
  }

  return {
    complete: missing.length === 0,
    missing,
  };
}

function createDefaultContent(type: VisualBlockType): Record<string, any> {
  switch (type) {
    case "formula_drag":
      return {
        formula_template: "PV = FV / (1 + r)^n",
        slots: [],
        variable_bank: [],
      };
    case "parameter_chart":
      return {
        chart_type: "line",
        parameters: [],
        target_metric: "",
      };
    case "reasoning_sort":
      return {
        items: [],
      };
    case "concept_match":
      return {
        pairs: [],
      };
    case "condition_mark":
      return {
        passage: "",
        items: [],
      };
    case "choice":
      return {
        prompt: "",
        blanks: [],
        mode: "single_choice",
      };
    case "fill_blank":
      return {
        prompt: "",
        blanks: [],
        mode: "fill_blank",
      };
  }
}

function createDefaultEvaluationConfig(type: VisualBlockType): Record<string, any> {
  switch (type) {
    case "formula_drag":
      return { correct_mapping: {} };
    case "parameter_chart":
      return { target_state: {} };
    case "reasoning_sort":
      return { correct_order: [] };
    case "concept_match":
      return { correct_pairs: [] };
    case "condition_mark":
      return { correct_marks: [] };
    case "choice":
      return { correct_answers: [] };
    case "fill_blank":
      return { correct_answers: [] };
  }
}
