package interactive

type StepSchema struct {
	ID                 string         `json:"id"`
	WidgetType         string         `json:"widget_type"`
	Content            map[string]any `json:"content"`
	InitialState       map[string]any `json:"initial_state"`
	AllowedActions     map[string]any `json:"allowed_actions"`
	EvaluationConfig   map[string]any `json:"evaluation_config"`
	FeedbackMap        map[string]any `json:"feedback_map"`
	HintPolicy         map[string]any `json:"hint_policy"`
	KnowledgePointIDs  []string       `json:"knowledge_point_ids"`
	KnowledgePointTags []string       `json:"knowledge_point_tags"`
}

type EvaluationResult struct {
	IsCorrect bool
	Hint      string
}
