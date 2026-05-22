package interactive

import "testing"

func TestEvaluateOrderingStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "ordering_matching",
		EvaluationConfig: map[string]any{
			"mode":          "ordering",
			"correct_order": []string{"a", "b", "c"},
		},
	}
	result := EvaluateStep(step, map[string]any{
		"ordered_ids": []string{"a", "b", "c"},
	})
	if !result.IsCorrect {
		t.Fatalf("expected ordering step to be correct")
	}
}

func TestEvaluateMatchingStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "ordering_matching",
		EvaluationConfig: map[string]any{
			"mode": "matching",
			"correct_pairs": map[string]any{
				"concept_1": "answer_1",
				"concept_2": "answer_2",
			},
		},
	}
	result := EvaluateStep(step, map[string]any{
		"pairs": map[string]any{
			"concept_1": "answer_1",
			"concept_2": "answer_2",
		},
	})
	if !result.IsCorrect {
		t.Fatalf("expected matching step to be correct")
	}
}

func TestEvaluateHighlightMarkingStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "highlight_marking",
		EvaluationConfig: map[string]any{
			"correct_marked_ids": []string{"s1", "s3"},
		},
	}
	result := EvaluateStep(step, map[string]any{
		"marked_ids": []string{"s1", "s3"},
	})
	if !result.IsCorrect {
		t.Fatalf("expected highlight marking step to be correct")
	}
}

func TestEvaluateFormulaBuilderStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "formula_builder",
		EvaluationConfig: map[string]any{
			"required_slots": map[string]any{
				"numerator":   "cash_flow",
				"denominator": "discount_factor",
			},
		},
	}
	result := EvaluateStep(step, map[string]any{
		"slot_values": map[string]any{
			"numerator":   "cash_flow",
			"denominator": "discount_factor",
		},
	})
	if !result.IsCorrect {
		t.Fatalf("expected formula builder step to be correct")
	}
}

func TestEvaluateParameterLabStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "parameter_lab",
		EvaluationConfig: map[string]any{
			"expected_state": map[string]any{
				"discount_rate": 0.08,
				"periods":       5.0,
			},
		},
	}
	result := EvaluateStep(step, map[string]any{
		"state": map[string]any{
			"discount_rate": 0.08,
			"periods":       5.0,
		},
	})
	if !result.IsCorrect {
		t.Fatalf("expected parameter lab step to be correct")
	}
}

func TestEvaluateParameterLabQuizAnswerStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "parameter_lab",
		EvaluationConfig: map[string]any{
			"quiz_answer": "-75万",
		},
	}
	result := EvaluateStep(step, map[string]any{
		"answer": "-75",
	})
	if !result.IsCorrect {
		t.Fatalf("expected parameter lab quiz answer to accept equivalent numeric answer")
	}
}

func TestEvaluateChoiceClozeSingleChoiceStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "choice_cloze",
		EvaluationConfig: map[string]any{
			"mode":              "single_choice",
			"correct_option_id": "opt_b",
		},
	}
	result := EvaluateStep(step, map[string]any{
		"selected_option_id": "opt_b",
	})
	if !result.IsCorrect {
		t.Fatalf("expected single choice/cloze step to be correct")
	}
}

func TestEvaluateChoiceClozeFillBlankStep(t *testing.T) {
	step := StepSchema{
		WidgetType: "choice_cloze",
		EvaluationConfig: map[string]any{
			"mode": "fill_blank",
			"correct_blanks": map[string]any{
				"principle": "independence",
				"action":    "disclose",
			},
		},
	}
	result := EvaluateStep(step, map[string]any{
		"blank_values": map[string]any{
			"principle": "independence",
			"action":    "disclose",
		},
	})
	if !result.IsCorrect {
		t.Fatalf("expected fill blank choice/cloze step to be correct")
	}
}
