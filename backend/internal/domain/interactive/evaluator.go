package interactive

import (
	"fmt"
	"strconv"
	"strings"
)

func EvaluateStep(step StepSchema, submission map[string]any) EvaluationResult {
	switch step.WidgetType {
	case "ordering_matching":
		mode, _ := step.EvaluationConfig["mode"].(string)
		if mode == "matching" {
			return EvaluationResult{IsCorrect: compareMap(step.EvaluationConfig["correct_pairs"], submission["pairs"])}
		}
		return EvaluationResult{IsCorrect: compareStringSlice(step.EvaluationConfig["correct_order"], submission["ordered_ids"])}
	case "highlight_marking":
		if highlights, ok := step.EvaluationConfig["expected_highlights"]; ok {
			return EvaluationResult{IsCorrect: compareStringSlice(highlights, submission["marked_ids"])}
		}
		return EvaluationResult{IsCorrect: compareStringSlice(step.EvaluationConfig["correct_marked_ids"], submission["marked_ids"])}
	case "formula_builder":
		if correctFormula, ok := step.EvaluationConfig["correct_formula"].(string); ok && correctFormula != "" {
			return EvaluationResult{IsCorrect: compareFormula(step, submission["slot_values"], correctFormula)}
		}
		return EvaluationResult{IsCorrect: compareMap(step.EvaluationConfig["required_slots"], submission["slot_values"])}
	case "parameter_lab":
		if quizAnswer, ok := step.EvaluationConfig["quiz_answer"].(string); ok && strings.TrimSpace(quizAnswer) != "" {
			return EvaluationResult{IsCorrect: compareFlexibleText(quizAnswer, submission["answer"])}
		}
		if expectedState, ok := step.EvaluationConfig["expected_state"]; ok {
			return EvaluationResult{IsCorrect: compareMap(expectedState, submission["state"])}
		}
		if targetRange, ok := step.EvaluationConfig["target_range"]; ok {
			return EvaluationResult{IsCorrect: compareTargetRange(targetRange, submission["state"])}
		}
		return EvaluationResult{IsCorrect: false}
	case "choice_cloze":
		mode, _ := step.EvaluationConfig["mode"].(string)
		correctSelections, hasCorrectSelections := step.EvaluationConfig["correct_selections"]
		switch mode {
		case "multi_choice":
			if hasCorrectSelections {
				return EvaluationResult{IsCorrect: compareStringSlice(correctSelections, submission["selected_option_ids"])}
			}
			return EvaluationResult{IsCorrect: compareStringSlice(step.EvaluationConfig["correct_option_ids"], submission["selected_option_ids"])}
		case "fill_blank":
			if hasCorrectSelections {
				return EvaluationResult{IsCorrect: compareStringSlice(correctSelections, submission["blank_values"])}
			}
			return EvaluationResult{IsCorrect: compareMap(step.EvaluationConfig["correct_blanks"], submission["blank_values"])}
		default:
			if hasCorrectSelections {
				if correctSelectionList, ok := anyToStringSlice(correctSelections); ok {
					if len(correctSelectionList) == 1 {
						return EvaluationResult{IsCorrect: compareStringValue(correctSelectionList[0], submission["selected_option_id"])}
					}
					return EvaluationResult{IsCorrect: compareStringSlice(correctSelectionList, submission["selected_option_ids"])}
				}
			}
			return EvaluationResult{IsCorrect: compareStringValue(step.EvaluationConfig["correct_option_id"], submission["selected_option_id"])}
		}
	default:
		return EvaluationResult{IsCorrect: false}
	}
}

func compareTargetRange(targetRange any, state any) bool {
	ranges, ok := targetRange.(map[string]any)
	if !ok {
		return false
	}
	submitted, ok := anyToMap(state)
	if !ok {
		return false
	}
	for key, rangeValue := range ranges {
		bounds, ok := anyToFloatSlice(rangeValue)
		if !ok || len(bounds) != 2 {
			return false
		}
		value, ok := submitted[key]
		if !ok {
			return false
		}
		f, ok := anyToFloat(value)
		if !ok {
			return false
		}
		if f < bounds[0] || f > bounds[1] {
			return false
		}
	}
	return true
}

func compareStringSlice(left any, right any) bool {
	ls, ok := anyToStringSlice(left)
	if !ok {
		return false
	}
	rs, ok := anyToStringSlice(right)
	if !ok {
		return false
	}
	if len(ls) != len(rs) {
		return false
	}
	for i := range ls {
		if ls[i] != rs[i] {
			return false
		}
	}
	return true
}

func compareMap(left any, right any) bool {
	lm, ok := anyToMap(left)
	if !ok {
		return false
	}
	rm, ok := anyToMap(right)
	if !ok {
		return false
	}
	if len(lm) != len(rm) {
		return false
	}
	for k, lv := range lm {
		rv, ok := rm[k]
		if !ok {
			return false
		}
		switch lvTyped := lv.(type) {
		case []string, []any:
			if !compareStringSlice(lvTyped, rv) {
				return false
			}
		default:
			if lv != rv {
				return false
			}
		}
	}
	return true
}

func compareStringValue(left any, right any) bool {
	ls, ok := left.(string)
	if !ok {
		return false
	}
	rs, ok := right.(string)
	if !ok {
		return false
	}
	return ls == rs
}

func compareFlexibleText(left string, right any) bool {
	expected := normalizeAnswerText(left)
	actual := normalizeAnswerText(stringValue(right))
	if expected == "" || actual == "" {
		return false
	}
	if expected == actual {
		return true
	}
	expectedNumber, expectedOK := parseAnswerNumber(expected)
	actualNumber, actualOK := parseAnswerNumber(actual)
	if expectedOK && actualOK {
		const tolerance = 0.000001
		diff := expectedNumber - actualNumber
		if diff < 0 {
			diff = -diff
		}
		return diff <= tolerance
	}
	return false
}

func normalizeAnswerText(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "，", ",")
	return value
}

func parseAnswerNumber(value string) (float64, bool) {
	value = strings.TrimSuffix(value, "万元")
	value = strings.TrimSuffix(value, "万")
	value = strings.TrimSuffix(value, "%")
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func anyToStringSlice(v any) ([]string, bool) {
	switch t := v.(type) {
	case []string:
		return t, true
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			switch value := item.(type) {
			case string:
				out = append(out, value)
			case map[string]any:
				out = append(out, firstStringValue(value["id"], value["text"], value["label"], value["name"], value["value"]))
			default:
				out = append(out, stringValue(value))
			}
		}
		return out, true
	case map[string]any:
		out := make([]string, 0, len(t))
		for _, value := range t {
			out = append(out, stringValue(value))
		}
		return out, true
	default:
		return nil, false
	}
}

func anyToMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func anyToFloatSlice(v any) ([]float64, bool) {
	switch t := v.(type) {
	case []any:
		out := make([]float64, 0, len(t))
		for _, item := range t {
			f, ok := anyToFloat(item)
			if !ok {
				return nil, false
			}
			out = append(out, f)
		}
		return out, true
	case []float64:
		return t, true
	default:
		return nil, false
	}
}

func anyToFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case int32:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func compareFormula(step StepSchema, slotValues any, correctFormula string) bool {
	submitted, ok := anyToMap(slotValues)
	if !ok {
		return false
	}
	rawSlots, ok := step.Content["slots"].([]any)
	if !ok {
		if slots, ok := step.Content["slots"].([]map[string]any); ok {
			rawSlots = make([]any, 0, len(slots))
			for _, slot := range slots {
				rawSlots = append(rawSlots, slot)
			}
		} else {
			return false
		}
	}
	parts := make([]string, 0, len(rawSlots))
	for index, rawSlot := range rawSlots {
		key := slotKey(rawSlot, index)
		value, ok := submitted[key]
		if !ok {
			value, _ = submitted[strconv.Itoa(index)]
		}
		parts = append(parts, strings.TrimSpace(stringValue(value)))
	}
	return normalizeFormula(strings.Join(parts, "")) == normalizeFormula(correctFormula)
}

func slotKey(rawSlot any, index int) string {
	switch t := rawSlot.(type) {
	case map[string]any:
		return firstStringValue(t["key"], t["id"], t["label"], strconv.Itoa(index))
	default:
		return strconv.Itoa(index)
	}
}

func stringValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(v)
	}
}

func firstStringValue(values ...any) string {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func normalizeFormula(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), " ", "")
}
