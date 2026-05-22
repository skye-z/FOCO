package practice

import "time"

type CreateSessionInput struct {
	UserID            string
	ExamID            string
	Mode              string
	QuestionTypes     []string
	Difficulty        string
	Count             int
	SubjectIDs        []string
	ChapterIDs        []string
	KnowledgePointIDs []string
}

type DiagnosticRecommendation struct {
	RecommendedDifficulty        string
	RecommendedSubjectIDs        []string
	RecommendedChapterIDs        []string
	RecommendedKnowledgePointIDs []string
}

type ResolvedScope struct {
	Difficulty        string
	SubjectIDs        []string
	ChapterIDs        []string
	KnowledgePointIDs []string
}

type QuestionOption struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

type PracticeItemContent struct {
	Stem    string           `json:"stem"`
	Options []QuestionOption `json:"options"`
}

type PracticeSessionItemView struct {
	ItemID            string              `json:"item_id"`
	QuestionVersionID string              `json:"question_version_id"`
	QuestionType      string              `json:"question_type"`
	Score             int                 `json:"score"`
	Content           PracticeItemContent `json:"content"`
}

type PracticeSessionView struct {
	SessionID  string                    `json:"session_id"`
	ExamName   string                    `json:"exam_name"`
	TotalCount int                       `json:"total_count"`
	Items      []PracticeSessionItemView `json:"items"`
}

type KnowledgePointRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SubmitAnswerInput struct {
	UserID          string
	SessionID       string
	ItemID          string
	Answer          any
	DurationSeconds int
}

type SubmitResult struct {
	IsCorrect       bool                `json:"is_correct"`
	CorrectAnswer   any                 `json:"correct_answer"`
	Explanation     string              `json:"explanation"`
	KnowledgePoints []KnowledgePointRef `json:"knowledge_points"`
	XpEarned        int                 `json:"xp_earned"`
}

type SessionSummary struct {
	Total           int `json:"total"`
	Correct         int `json:"correct"`
	Wrong           int `json:"wrong"`
	Accuracy        int `json:"accuracy"`
	XpEarned        int `json:"xp_earned"`
	CoinsEarned     int `json:"coins_earned"`
	DurationMinutes int `json:"duration_minutes"`
}

type WrongBookFilter struct {
	UserID           string
	ExamID           string
	SubjectID        string
	ChapterID        string
	KnowledgePointID string
	Status           string
	Page             int
	PageSize         int
}

type WrongBookItem struct {
	ID              string              `json:"id"`
	QuestionID      string              `json:"question_id"`
	QuestionType    string              `json:"question_type"`
	Stem            string              `json:"stem"`
	Options         []QuestionOption    `json:"options"`
	UserAnswer      string              `json:"user_answer"`
	CorrectAnswer   string              `json:"correct_answer"`
	Explanation     string              `json:"explanation"`
	ErrorCount      int                 `json:"error_count"`
	FixCount        int                 `json:"fix_count"`
	FirstErrorAt    string              `json:"first_error_at"`
	LastErrorAt     string              `json:"last_error_at"`
	Status          string              `json:"status"`
	SubjectID       string              `json:"subject_id"`
	SubjectName     string              `json:"subject_name"`
	ChapterID       string              `json:"chapter_id"`
	ChapterName     string              `json:"chapter_name"`
	KnowledgePoints []KnowledgePointRef `json:"knowledge_points"`
}

type CandidateQuestion struct {
	QuestionID        string
	QuestionVersionID string
	SubjectID         string
	ChapterID         string
	QuestionType      string
	Difficulty        int
	Stem              string
	Options           string
	CorrectAnswer     string
	Explanation       string
	KnowledgePointIDs []string
}

type EnrollmentRef struct {
	ID       string
	ExamID   string
	ExamName string
}

type PracticeSessionRecord struct {
	ID                   string
	UserID               string
	ExamID               string
	ExamEnrollmentID     string
	ScopeJSON            string
	Status               string
	TotalCount           int
	AnsweredCount        int
	CorrectCount         int
	TotalDurationSeconds int
	XpEarned             int
	CoinsEarned          int
	StartedAt            time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type PracticeSessionItemRecord struct {
	ID                string
	SessionID         string
	QuestionID        string
	QuestionVersionID string
	SubjectID         string
	ChapterID         string
	QuestionType      string
	Score             int
	Position          int
	Stem              string
	Options           string
	CorrectLabels     string
	Explanation       string
	KnowledgePointIDs string
	CreatedAt         time.Time
}

type SubmissionRecord struct {
	SessionID         string
	ItemID            string
	ExamID            string
	QuestionType      string
	CorrectLabels     []string
	Explanation       string
	KnowledgePointIDs []string
	SubmittedAt       *time.Time
	IsCorrect         *bool
}
