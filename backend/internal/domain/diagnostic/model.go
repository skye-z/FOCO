package diagnostic

import "time"

const (
	TriggerTypeInitialAuto   = "initial_auto"
	TriggerTypeManualRestart = "manual_restart"
	StatusPending            = "pending"
	StatusCompleted          = "completed"
)

type KnowledgePoint struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Option struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

type Question struct {
	ID                string           `json:"id"`
	QuestionVersionID string           `json:"question_version_id"`
	SubjectID         string           `json:"subject_id"`
	SubjectName       string           `json:"subject_name"`
	ChapterID         string           `json:"chapter_id"`
	ChapterName       string           `json:"chapter_name"`
	QuestionType      string           `json:"question_type"`
	Stem              string           `json:"stem"`
	Options           []Option         `json:"options"`
	CorrectLabels     []string         `json:"-"`
	KnowledgePoints   []KnowledgePoint `json:"knowledge_points"`
}

type Attempt struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	ExamID      string          `json:"exam_id"`
	TriggerType string          `json:"trigger_type"`
	Status      string          `json:"status"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Items       []Question      `json:"items,omitempty"`
	Result      *ProfileSummary `json:"result,omitempty"`
}

type KnowledgeMastery struct {
	KnowledgePointID   string `json:"knowledge_point_id"`
	KnowledgePointName string `json:"knowledge_point_name"`
	MasteryScore       int    `json:"mastery_score"`
	ConfidenceScore    int    `json:"confidence_score"`
	ForgettingDueAt    string `json:"forgetting_due_at"`
}

type AreaSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Accuracy    int    `json:"accuracy"`
	Attempts    int    `json:"attempts"`
	Recommended bool   `json:"recommended"`
}

type ProfileSummary struct {
	HasCompleted                   bool               `json:"has_completed"`
	CompletedAt                    string             `json:"completed_at,omitempty"`
	OverallAccuracy                int                `json:"overall_accuracy"`
	SummaryText                    string             `json:"summary_text"`
	RecommendedDifficulty          string             `json:"recommended_difficulty"`
	RecommendedSubjectIDs          []string           `json:"recommended_subject_ids"`
	RecommendedSubjectNames        []string           `json:"recommended_subject_names"`
	RecommendedChapterIDs          []string           `json:"recommended_chapter_ids"`
	RecommendedChapterNames        []string           `json:"recommended_chapter_names"`
	RecommendedKnowledgePointIDs   []string           `json:"recommended_knowledge_point_ids"`
	RecommendedKnowledgePointNames []string           `json:"recommended_knowledge_point_names"`
	Subjects                       []AreaSummary      `json:"subjects"`
	Chapters                       []AreaSummary      `json:"chapters"`
	KnowledgePoints                []KnowledgeMastery `json:"knowledge_points"`
}

type Profile struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	ExamID          string         `json:"exam_id"`
	ProfileVersion  int            `json:"profile_version"`
	Summary         ProfileSummary `json:"summary"`
	ConfidenceScore int            `json:"confidence_score"`
	ComputedAt      time.Time      `json:"computed_at"`
}

type CurrentPayload struct {
	Status    string          `json:"status"`
	AttemptID string          `json:"attempt_id,omitempty"`
	Items     []Question      `json:"items,omitempty"`
	Summary   *ProfileSummary `json:"summary,omitempty"`
}

type SubmitInput struct {
	UserID    string
	AttemptID string
	Answers   map[string][]string
}
