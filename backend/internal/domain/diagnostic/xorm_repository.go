package diagnostic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"xorm.io/xorm"
)

type XormRepository struct {
	engine *xorm.Engine
}

func NewXormRepository(engine *xorm.Engine) *XormRepository {
	return &XormRepository{engine: engine}
}

func (r *XormRepository) FindLatestAttempt(ctx context.Context, userID, examID string) (*Attempt, error) {
	query := `
		select id::text as id,
		       user_id,
		       exam_id::text as exam_id,
		       trigger_type,
		       status,
		       started_at,
		       completed_at,
		       coalesce(summary::text, '{}') as summary
		from diagnostic_attempts
		where user_id = ?::uuid
	`
	args := []any{userID}
	if examID != "" {
		query += ` and exam_id = ?::uuid`
		args = append(args, examID)
	}
	query += ` order by created_at desc limit 1`

	var row struct {
		ID          string     `xorm:"id"`
		UserID      string     `xorm:"user_id"`
		ExamID      string     `xorm:"exam_id"`
		TriggerType string     `xorm:"trigger_type"`
		Status      string     `xorm:"status"`
		StartedAt   *time.Time `xorm:"started_at"`
		CompletedAt *time.Time `xorm:"completed_at"`
		Summary     string     `xorm:"summary"`
	}
	has, err := r.engine.Context(ctx).SQL(query, args...).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	attempt := &Attempt{
		ID:          row.ID,
		UserID:      row.UserID,
		ExamID:      row.ExamID,
		TriggerType: row.TriggerType,
		Status:      row.Status,
		StartedAt:   row.StartedAt,
		CompletedAt: row.CompletedAt,
	}

	var summary struct {
		Items  json.RawMessage `json:"items"`
		Result *ProfileSummary `json:"result"`
	}
	_ = json.Unmarshal([]byte(row.Summary), &summary)
	attempt.Items = unmarshalItems(summary.Items)
	attempt.Result = summary.Result
	return attempt, nil
}

func (r *XormRepository) CreateAttempt(ctx context.Context, attempt *Attempt) error {
	summaryJSON, _ := json.Marshal(map[string]any{
		"items":  marshalItems(attempt.Items),
		"result": nil,
	})
	_, err := r.engine.Context(ctx).Exec(`
		insert into diagnostic_attempts(
			id, user_id, exam_id, trigger_type, status, started_at, completed_at, summary, created_at
		) values (
			?::uuid, ?::uuid, ?::uuid, ?, ?, ?, null, ?::jsonb, now()
		)
	`, attempt.ID, attempt.UserID, attempt.ExamID, attempt.TriggerType, attempt.Status, attempt.StartedAt, string(summaryJSON))
	return err
}

func (r *XormRepository) UpdateAttempt(ctx context.Context, attempt *Attempt) error {
	summaryJSON, _ := json.Marshal(map[string]any{
		"items":  marshalItems(attempt.Items),
		"result": attempt.Result,
	})
	_, err := r.engine.Context(ctx).Exec(`
		update diagnostic_attempts
		set status = ?,
		    completed_at = ?,
		    summary = ?::jsonb
		where id = ?::uuid
	`, attempt.Status, attempt.CompletedAt, string(summaryJSON), attempt.ID)
	return err
}

func (r *XormRepository) ListDiagnosticQuestions(ctx context.Context, examID string, limit int) ([]Question, error) {
	if limit <= 0 {
		limit = 10
	}

	var rows []struct {
		QuestionVersionID string `xorm:"question_version_id"`
		SubjectID         string `xorm:"subject_id"`
		SubjectName       string `xorm:"subject_name"`
		ChapterID         string `xorm:"chapter_id"`
		ChapterName       string `xorm:"chapter_name"`
		QuestionType      string `xorm:"question_type"`
		Difficulty        int    `xorm:"difficulty"`
		Stem              string `xorm:"stem"`
		Options           string `xorm:"options"`
		CorrectAnswer     string `xorm:"correct_answer"`
		KnowledgePoints   string `xorm:"knowledge_points"`
		SortOrder         int    `xorm:"sort_order"`
	}
	err := r.engine.Context(ctx).SQL(`
		select qv.id::text as question_version_id,
		       s.id::text as subject_id,
		       s.name as subject_name,
		       coalesce(c.id::text, '') as chapter_id,
		       coalesce(c.name, '') as chapter_name,
		       qv.question_type as question_type,
		       qv.difficulty as difficulty,
		       qv.stem::text as stem,
		       coalesce(qv.options::text, '{"choices":[]}') as options,
		       qv.correct_answer::text as correct_answer,
		       coalesce(
		           jsonb_agg(jsonb_build_object('id', kp.id::text, 'name', kp.name)) filter (where kp.id is not null),
		           '[]'::jsonb
		       )::text as knowledge_points,
		       s.sort_order as sort_order
		from questions q
		join question_versions qv on qv.id = q.current_published_version_id
		join subjects s on s.id = q.subject_id
		left join chapters c on c.id = q.chapter_id
		left join question_version_knowledge_points qvkp on qvkp.question_version_id = qv.id
		left join knowledge_points kp on kp.id = qvkp.knowledge_point_id
		where q.exam_id = ?::uuid
		  and q.status = 'published'
		group by qv.id, s.id, s.name, c.id, c.name, qv.question_type, qv.difficulty, qv.stem, qv.options, qv.correct_answer, s.sort_order
		order by s.sort_order asc, c.sort_order asc, qv.difficulty asc, random()
	`, examID).Find(&rows)
	if err != nil {
		return nil, err
	}

	const perChapterTarget = 3

	type chapterKey struct {
		subjectID string
		chapterID string
	}

	picks := make([]int, 0)
	chapterCount := map[chapterKey]int{}

	for i, row := range rows {
		if row.ChapterID == "" {
			continue
		}
		key := chapterKey{subjectID: row.SubjectID, chapterID: row.ChapterID}
		if chapterCount[key] >= perChapterTarget {
			continue
		}
		chapterCount[key]++
		picks = append(picks, i)
	}

	questions := make([]Question, 0, len(picks))
	for _, idx := range picks {
		row := rows[idx]
		options, labelByID := parseOptions(row.Options)
		questions = append(questions, Question{
			ID:                "dq-" + row.QuestionVersionID,
			QuestionVersionID: row.QuestionVersionID,
			SubjectID:         row.SubjectID,
			SubjectName:       row.SubjectName,
			ChapterID:         row.ChapterID,
			ChapterName:       row.ChapterName,
			QuestionType:      row.QuestionType,
			Stem:              extractText(row.Stem),
			Options:           options,
			CorrectLabels:     parseCorrectLabels(row.CorrectAnswer, labelByID),
			KnowledgePoints:   decodeKnowledgePoints(row.KnowledgePoints),
		})
	}
	return questions, nil
}

func (r *XormRepository) SaveProfile(ctx context.Context, profile *Profile) error {
	summaryJSON, _ := json.Marshal(profile.Summary)
	sourceJSON, _ := json.Marshal(map[string]any{
		"recommended_subject_ids":         profile.Summary.RecommendedSubjectIDs,
		"recommended_chapter_ids":         profile.Summary.RecommendedChapterIDs,
		"recommended_knowledge_point_ids": profile.Summary.RecommendedKnowledgePointIDs,
	})
	_, err := r.engine.Context(ctx).Exec(`
		insert into learner_profiles(
			id, user_id, exam_id, profile_version, profile_summary, confidence_score, source_snapshot, computed_at
		) values (
			?::uuid, ?::uuid, ?::uuid, ?, ?::jsonb, ?, ?::jsonb, ?
		)
	`, profile.ID, profile.UserID, profile.ExamID, profile.ProfileVersion, string(summaryJSON), profile.ConfidenceScore, string(sourceJSON), profile.ComputedAt)
	return err
}

func (r *XormRepository) FindLatestProfile(ctx context.Context, userID, examID string) (*Profile, error) {
	var row struct {
		ID              string    `xorm:"id"`
		UserID          string    `xorm:"user_id"`
		ExamID          string    `xorm:"exam_id"`
		ProfileVersion  int       `xorm:"profile_version"`
		ProfileSummary  string    `xorm:"profile_summary"`
		ConfidenceScore int       `xorm:"confidence_score"`
		ComputedAt      time.Time `xorm:"computed_at"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select id::text as id,
		       user_id,
		       exam_id::text as exam_id,
		       profile_version,
		       profile_summary::text as profile_summary,
		       confidence_score,
		       computed_at
		from learner_profiles
		where user_id = ?::uuid and exam_id = ?::uuid
		order by profile_version desc
		limit 1
	`, userID, examID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	var summary ProfileSummary
	_ = json.Unmarshal([]byte(row.ProfileSummary), &summary)

	return &Profile{
		ID:              row.ID,
		UserID:          row.UserID,
		ExamID:          row.ExamID,
		ProfileVersion:  row.ProfileVersion,
		Summary:         summary,
		ConfidenceScore: row.ConfidenceScore,
		ComputedAt:      row.ComputedAt,
	}, nil
}

func parseOptions(raw string) ([]Option, map[string]string) {
	var payload struct {
		Choices []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"choices"`
	}
	_ = json.Unmarshal([]byte(raw), &payload)
	options := make([]Option, 0, len(payload.Choices))
	labels := make(map[string]string, len(payload.Choices))
	for index, choice := range payload.Choices {
		label := string(rune('A' + index))
		options = append(options, Option{Label: label, Text: choice.Text})
		labels[choice.ID] = label
	}
	return options, labels
}

func parseCorrectLabels(raw string, labelByID map[string]string) []string {
	var payload struct {
		SelectedOptionIDs []string `json:"selected_option_ids"`
	}
	_ = json.Unmarshal([]byte(raw), &payload)
	labels := make([]string, 0, len(payload.SelectedOptionIDs))
	for _, optionID := range payload.SelectedOptionIDs {
		if label, ok := labelByID[optionID]; ok {
			labels = append(labels, label)
		}
	}
	return normalizeLabels(labels)
}

func extractText(raw string) string {
	var payload struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && strings.TrimSpace(payload.Text) != "" {
		return payload.Text
	}
	return strings.TrimSpace(raw)
}

func decodeKnowledgePoints(raw string) []KnowledgePoint {
	var result []KnowledgePoint
	_ = json.Unmarshal([]byte(raw), &result)
	if result == nil {
		return []KnowledgePoint{}
	}
	return result
}

type questionJSON struct {
	ID                string           `json:"id"`
	QuestionVersionID string           `json:"question_version_id"`
	SubjectID         string           `json:"subject_id"`
	SubjectName       string           `json:"subject_name"`
	ChapterID         string           `json:"chapter_id"`
	ChapterName       string           `json:"chapter_name"`
	QuestionType      string           `json:"question_type"`
	Stem              string           `json:"stem"`
	Options           []Option         `json:"options"`
	CorrectLabels     []string         `json:"correct_labels"`
	KnowledgePoints   []KnowledgePoint `json:"knowledge_points"`
}

func marshalItems(items []Question) []questionJSON {
	out := make([]questionJSON, len(items))
	for i, q := range items {
		out[i] = questionJSON{
			ID: q.ID, QuestionVersionID: q.QuestionVersionID,
			SubjectID: q.SubjectID, SubjectName: q.SubjectName,
			ChapterID: q.ChapterID, ChapterName: q.ChapterName,
			QuestionType: q.QuestionType, Stem: q.Stem,
			Options: q.Options, CorrectLabels: q.CorrectLabels,
			KnowledgePoints: q.KnowledgePoints,
		}
	}
	return out
}

func unmarshalItems(raw json.RawMessage) []Question {
	var trimmed json.RawMessage
	if err := json.Unmarshal(raw, &trimmed); err == nil {
		raw = trimmed
	}
	var items []questionJSON
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	out := make([]Question, len(items))
	for i, q := range items {
		out[i] = Question{
			ID: q.ID, QuestionVersionID: q.QuestionVersionID,
			SubjectID: q.SubjectID, SubjectName: q.SubjectName,
			ChapterID: q.ChapterID, ChapterName: q.ChapterName,
			QuestionType: q.QuestionType, Stem: q.Stem,
			Options: q.Options, CorrectLabels: q.CorrectLabels,
			KnowledgePoints: q.KnowledgePoints,
		}
	}
	return out
}
