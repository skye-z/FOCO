package interactive

import "context"

type AdminUnitSummary struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	ExamID              string `json:"exam_id"`
	SubjectID           string `json:"subject_id,omitempty"`
	SubjectName         string `json:"subject_name,omitempty"`
	Status              string `json:"status"`
	StepCount           int    `json:"step_count"`
	VersionNo           int    `json:"version_no"`
	VersionID           string `json:"version_id"`
	UpdatedAt           string `json:"updated_at"`
	PublishedVersionNo  *int   `json:"published_version_no,omitempty"`
	HasUnpublishedDraft bool   `json:"has_unpublished_draft"`
}

type AdminVersionSummary struct {
	VersionID   string  `json:"version_id"`
	UnitID      string  `json:"unit_id"`
	VersionNo   int     `json:"version_no"`
	Status      string  `json:"status"`
	PublishedAt *string `json:"published_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type AdminVersionDetail struct {
	VersionID string       `json:"version_id"`
	UnitID    string       `json:"unit_id"`
	VersionNo int          `json:"version_no"`
	Status    string       `json:"status"`
	Title     string       `json:"title"`
	Steps     []StepSchema `json:"steps"`
}

type AttemptScope struct {
	UserID string
	ExamID string
}

type Repository interface {
	ListUnits(ctx context.Context) ([]UnitSummary, error)
	GetUnit(ctx context.Context, unitVersionID string) (*UnitView, error)
	CreateAttempt(ctx context.Context, unitVersionID, userID string) (*UnitAttempt, error)
	GetAttemptScope(ctx context.Context, attemptID string) (*AttemptScope, error)
	GetStep(ctx context.Context, stepID string) (*StepSchema, error)
	SaveStepAction(ctx context.Context, attemptID, stepID string, payload map[string]any) error
	SaveStepFeedback(ctx context.Context, attemptID, stepID string, feedback *StepFeedback) error
	CompleteAttempt(ctx context.Context, attemptID string) error
	CreateConceptCard(ctx context.Context, attemptID string) (*ConceptCard, error)

	AdminListUnits(ctx context.Context, examID, subjectID string) ([]AdminUnitSummary, error)
	AdminCreateUnit(ctx context.Context, examID, subjectID, title string) (*AdminUnitSummary, error)
	AdminListVersions(ctx context.Context, unitID string) ([]AdminVersionSummary, error)
	AdminCreateVersion(ctx context.Context, unitID string) (*AdminVersionDetail, error)
	AdminGetVersionDetail(ctx context.Context, versionID string) (*AdminVersionDetail, error)
	AdminUpdateVersion(ctx context.Context, versionID, title string, steps []StepSchema) (*AdminVersionDetail, error)
	AdminPublishVersion(ctx context.Context, versionID string) (*AdminVersionDetail, error)
	AdminDeleteUnit(ctx context.Context, unitID string) error
}
