package interactive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"xorm.io/xorm"
)

type XormRepository struct {
	engine *xorm.Engine
}

type adminUnitSummaryRow struct {
	ID                  string `xorm:"id"`
	Title               string `xorm:"title"`
	ExamID              string `xorm:"exam_id"`
	SubjectID           string `xorm:"subject_id"`
	SubjectName         string `xorm:"subject_name"`
	Status              string `xorm:"status"`
	StepCount           int    `xorm:"step_count"`
	VersionNo           int    `xorm:"version_no"`
	VersionID           string `xorm:"version_id"`
	UpdatedAt           string `xorm:"updated_at"`
	PublishedVersionNo  *int   `xorm:"published_version_no"`
	HasUnpublishedDraft bool   `xorm:"has_unpublished_draft"`
}

type adminVersionSummaryRow struct {
	VersionID   string  `xorm:"version_id"`
	UnitID      string  `xorm:"unit_id"`
	VersionNo   int     `xorm:"version_no"`
	Status      string  `xorm:"status"`
	PublishedAt *string `xorm:"published_at"`
	UpdatedAt   string  `xorm:"updated_at"`
}

func NewXormRepository(engine *xorm.Engine) *XormRepository {
	return &XormRepository{engine: engine}
}

func (r *XormRepository) ListUnits(ctx context.Context) ([]UnitSummary, error) {
	var rows []struct {
		ID        string `xorm:"id"`
		Title     string `xorm:"title"`
		StepCount int    `xorm:"step_count"`
	}
	err := r.engine.Context(ctx).SQL(`
		select iuv.id::text as id,
		       coalesce(iuv.metadata->>'title', iu.title) as title,
		       count(iuvs.id)::int as step_count
		from interactive_unit_versions iuv
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		join exams e on e.id = iu.exam_id
		left join interactive_unit_version_steps iuvs on iuvs.unit_version_id = iuv.id
		where iuv.status = 'published'
		  and e.status = 'active'
		group by iuv.id, coalesce(iuv.metadata->>'title', iu.title), iuv.published_at, iu.updated_at
		order by coalesce(iuv.published_at, iu.updated_at) desc, iuv.id desc
	`).Find(&rows)
	if err != nil {
		return nil, err
	}
	units := make([]UnitSummary, 0, len(rows))
	for _, row := range rows {
		units = append(units, UnitSummary{
			ID:        row.ID,
			Title:     row.Title,
			StepCount: row.StepCount,
		})
	}
	return units, nil
}

func (r *XormRepository) GetUnit(ctx context.Context, unitVersionID string) (*UnitView, error) {
	var view UnitView
	has, err := r.engine.Context(ctx).SQL(`
		select iuv.id::text, coalesce((iuv.metadata->>'title'), iu.title)
		from interactive_unit_versions iuv
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		where iuv.id = ?::uuid
	`, unitVersionID).Get(&view)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	var stepRows []struct {
		ID                 string `xorm:"id"`
		WidgetType         string `xorm:"widget_type"`
		Content            string `xorm:"content"`
		InitialState       string `xorm:"initial_state"`
		AllowedActions     string `xorm:"allowed_actions"`
		EvaluationConfig   string `xorm:"evaluation_config"`
		FeedbackMap        string `xorm:"feedback_map"`
		HintPolicy         string `xorm:"hint_policy"`
		KnowledgePointIDs  string `xorm:"knowledge_point_ids"`
		KnowledgePointTags string `xorm:"knowledge_point_tags"`
	}
	err = r.engine.Context(ctx).SQL(`
		select id::text, widget_type, content::text, initial_state::text, allowed_actions::text, evaluation_config::text, feedback_map::text, hint_policy::text, knowledge_point_ids::text, knowledge_point_tags::text
		from interactive_unit_version_steps
		where unit_version_id = ?::uuid
		order by step_no asc
	`, unitVersionID).Find(&stepRows)
	if err != nil {
		return nil, err
	}

	view.Steps = make([]StepSchema, 0, len(stepRows))
	for _, sr := range stepRows {
		view.Steps = append(view.Steps, StepSchema{
			ID:                 sr.ID,
			WidgetType:         sr.WidgetType,
			Content:            decodeMap([]byte(sr.Content)),
			InitialState:       decodeMap([]byte(sr.InitialState)),
			AllowedActions:     decodeMap([]byte(sr.AllowedActions)),
			EvaluationConfig:   decodeMap([]byte(sr.EvaluationConfig)),
			FeedbackMap:        decodeMap([]byte(sr.FeedbackMap)),
			HintPolicy:         decodeMap([]byte(sr.HintPolicy)),
			KnowledgePointIDs:  decodeStringSlice([]byte(sr.KnowledgePointIDs)),
			KnowledgePointTags: decodeStringSlice([]byte(sr.KnowledgePointTags)),
		})
	}
	return &view, nil
}

func (r *XormRepository) CreateAttempt(ctx context.Context, unitVersionID, userID string) (*UnitAttempt, error) {
	attempt := &UnitAttempt{
		ID:            newUUIDLikeString(),
		UnitVersionID: unitVersionID,
		UserID:        userID,
		Status:        "in_progress",
	}
	_, err := r.engine.Context(ctx).Exec(`
		insert into unit_attempts(id, user_id, unit_version_id, status, created_at)
		values (?::uuid, ?, ?::uuid, ?, now())
	`, attempt.ID, attempt.UserID, attempt.UnitVersionID, attempt.Status)
	return attempt, err
}

func (r *XormRepository) GetAttemptScope(ctx context.Context, attemptID string) (*AttemptScope, error) {
	var row struct {
		UserID string `xorm:"user_id"`
		ExamID string `xorm:"exam_id"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select ua.user_id,
		       iu.exam_id::text as exam_id
		from unit_attempts ua
		join interactive_unit_versions iuv on iuv.id = ua.unit_version_id
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		where ua.id = ?::uuid
	`, attemptID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &AttemptScope{UserID: row.UserID, ExamID: row.ExamID}, nil
}

func (r *XormRepository) GetStep(ctx context.Context, stepID string) (*StepSchema, error) {
	var sr struct {
		ID                 string `xorm:"id"`
		WidgetType         string `xorm:"widget_type"`
		Content            string `xorm:"content"`
		InitialState       string `xorm:"initial_state"`
		AllowedActions     string `xorm:"allowed_actions"`
		EvaluationConfig   string `xorm:"evaluation_config"`
		FeedbackMap        string `xorm:"feedback_map"`
		HintPolicy         string `xorm:"hint_policy"`
		KnowledgePointIDs  string `xorm:"knowledge_point_ids"`
		KnowledgePointTags string `xorm:"knowledge_point_tags"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select id::text, widget_type, content::text, initial_state::text, allowed_actions::text, evaluation_config::text, feedback_map::text, hint_policy::text, knowledge_point_ids::text, knowledge_point_tags::text
		from interactive_unit_version_steps
		where id = ?::uuid
	`, stepID).Get(&sr)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &StepSchema{
		ID:                 sr.ID,
		WidgetType:         sr.WidgetType,
		Content:            decodeMap([]byte(sr.Content)),
		InitialState:       decodeMap([]byte(sr.InitialState)),
		AllowedActions:     decodeMap([]byte(sr.AllowedActions)),
		EvaluationConfig:   decodeMap([]byte(sr.EvaluationConfig)),
		FeedbackMap:        decodeMap([]byte(sr.FeedbackMap)),
		HintPolicy:         decodeMap([]byte(sr.HintPolicy)),
		KnowledgePointIDs:  decodeStringSlice([]byte(sr.KnowledgePointIDs)),
		KnowledgePointTags: decodeStringSlice([]byte(sr.KnowledgePointTags)),
	}, nil
}

func (r *XormRepository) SaveStepAction(ctx context.Context, attemptID, stepID string, payload map[string]any) error {
	body, _ := json.Marshal(payload)
	_, err := r.engine.Context(ctx).Exec(`
		insert into step_actions(id, attempt_id, step_id, action_payload, created_at)
		values (?::uuid, ?::uuid, ?::uuid, ?::jsonb, now())
	`, newUUIDLikeString(), attemptID, stepID, string(body))
	return err
}

func (r *XormRepository) SaveStepFeedback(ctx context.Context, attemptID, stepID string, feedback *StepFeedback) error {
	_, err := r.engine.Context(ctx).Exec(`
		insert into step_feedback(id, attempt_id, step_id, is_correct, allow_continue, hint, created_at)
		values (?::uuid, ?::uuid, ?::uuid, ?, ?, ?, now())
	`, newUUIDLikeString(), attemptID, stepID, feedback.IsCorrect, feedback.AllowContinue, feedback.Hint)
	return err
}

func (r *XormRepository) CompleteAttempt(ctx context.Context, attemptID string) error {
	_, err := r.engine.Context(ctx).Exec(`
		update unit_attempts
		set status = 'completed', completed_at = now()
		where id = ?::uuid
	`, attemptID)
	return err
}

func (r *XormRepository) CreateConceptCard(ctx context.Context, attemptID string) (*ConceptCard, error) {
	existing, err := r.getConceptCardByAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	var info struct {
		UserID        string `xorm:"user_id"`
		UnitVersionID string `xorm:"unit_version_id"`
		UnitTitle     string `xorm:"unit_title"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select ua.user_id, ua.unit_version_id::text, coalesce(iuv.metadata->>'title', iu.title) as unit_title
		from unit_attempts ua
		join interactive_unit_versions iuv on iuv.id = ua.unit_version_id
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		where ua.id = ?::uuid
	`, attemptID).Get(&info)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("attempt not found: %s", attemptID)
	}

	totalSteps, err := r.countUnitSteps(ctx, info.UnitVersionID)
	if err != nil {
		return nil, err
	}
	correctSteps, err := r.countCorrectFeedback(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	content := map[string]any{
		"title":         info.UnitTitle,
		"summary":       fmt.Sprintf("Completed %s with %d/%d correct steps.", info.UnitTitle, correctSteps, totalSteps),
		"correct_steps": correctSteps,
		"total_steps":   totalSteps,
	}
	if totalSteps > 0 && correctSteps < totalSteps {
		content["next_focus"] = "Review the blocked step hints and replay the lab once."
	} else {
		content["next_focus"] = "You can now explain this concept chain without hints."
	}
	body, _ := json.Marshal(content)
	card := &ConceptCard{
		ID:            newUUIDLikeString(),
		AttemptID:     attemptID,
		UnitVersionID: info.UnitVersionID,
		Content:       content,
	}
	_, err = r.engine.Context(ctx).Exec(`
		insert into concept_cards(id, user_id, attempt_id, unit_version_id, content, created_at)
		values (?::uuid, ?, ?::uuid, ?::uuid, ?::jsonb, now())
	`, card.ID, info.UserID, attemptID, info.UnitVersionID, string(body))
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (r *XormRepository) getConceptCardByAttempt(ctx context.Context, attemptID string) (*ConceptCard, error) {
	var row struct {
		ID            string `xorm:"id"`
		AttemptID     string `xorm:"attempt_id"`
		UnitVersionID string `xorm:"unit_version_id"`
		Content       string `xorm:"content"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select id::text, attempt_id::text, unit_version_id::text, content::text
		from concept_cards
		where attempt_id = ?::uuid
		order by created_at desc
		limit 1
	`, attemptID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &ConceptCard{
		ID:            row.ID,
		AttemptID:     row.AttemptID,
		UnitVersionID: row.UnitVersionID,
		Content:       decodeMap([]byte(row.Content)),
	}, nil
}

func (r *XormRepository) countUnitSteps(ctx context.Context, unitVersionID string) (int, error) {
	var result struct {
		Count int `xorm:"count"`
	}
	_, err := r.engine.Context(ctx).SQL(`
		select count(1)::int as count
		from interactive_unit_version_steps
		where unit_version_id = ?::uuid
	`, unitVersionID).Get(&result)
	return result.Count, err
}

func (r *XormRepository) countCorrectFeedback(ctx context.Context, attemptID string) (int, error) {
	var result struct {
		Count int `xorm:"count"`
	}
	_, err := r.engine.Context(ctx).SQL(`
		select count(1)::int as count
		from step_feedback
		where attempt_id = ?::uuid and is_correct = true
	`, attemptID).Get(&result)
	return result.Count, err
}

func (r *XormRepository) AdminListUnits(ctx context.Context, examID, subjectID string) ([]AdminUnitSummary, error) {
	query := `
		select iu.id::text as id,
		       iu.title as title,
		       iu.exam_id::text as exam_id,
		       coalesce(iu.subject_id::text, '') as subject_id,
		       coalesce(s.name, '') as subject_name,
		       iu.status as status,
		       coalesce(step_counts.step_count, 0) as step_count,
		       coalesce(latest_v.version_no, 0) as version_no,
		       coalesce(latest_v.id::text, '') as version_id,
		       to_char(iu.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as updated_at,
		       pub_v.version_no as published_version_no,
		       coalesce(draft_exists.has_draft, false) as has_unpublished_draft
		from interactive_units iu
		join exams e on e.id = iu.exam_id
		left join subjects s on s.id = iu.subject_id
		left join (
		    select iuv.interactive_unit_id, max(iuv.version_no) as version_no, iuv.id
		    from interactive_unit_versions iuv
		    group by iuv.interactive_unit_id, iuv.id
		    order by max(iuv.version_no) desc
		) latest_v on latest_v.interactive_unit_id = iu.id and latest_v.version_no = (
		    select max(iuv2.version_no) from interactive_unit_versions iuv2 where iuv2.interactive_unit_id = iu.id
		)
		left join (
		    select unit_version_id, count(1)::int as step_count
		    from interactive_unit_version_steps
		    group by unit_version_id
		) step_counts on step_counts.unit_version_id = latest_v.id
		left join interactive_unit_versions pub_v on pub_v.id = iu.current_published_version_id
		left join (
		    select distinct interactive_unit_id, true as has_draft
		    from interactive_unit_versions
		    where status = 'draft'
		) draft_exists on draft_exists.interactive_unit_id = iu.id
	`
	args := []any{}
	conditions := []string{"e.status = 'active'"}
	if examID != "" {
		conditions = append(conditions, "iu.exam_id = ?::uuid")
		args = append(args, examID)
	}
	if subjectID != "" {
		conditions = append(conditions, "iu.subject_id = ?::uuid")
		args = append(args, subjectID)
	}
	if len(conditions) > 0 {
		for i, c := range conditions {
			if i == 0 {
				query += " where " + c
			} else {
				query += " and " + c
			}
		}
	}
	query += " order by iu.updated_at desc"

	var rows []adminUnitSummaryRow
	err := r.engine.Context(ctx).SQL(query, args...).Find(&rows)
	if err != nil {
		return nil, err
	}
	return mapAdminUnitSummaryRows(rows), nil
}

func mapAdminUnitSummaryRows(rows []adminUnitSummaryRow) []AdminUnitSummary {
	units := make([]AdminUnitSummary, 0, len(rows))
	for _, row := range rows {
		units = append(units, AdminUnitSummary{
			ID:                  row.ID,
			Title:               row.Title,
			ExamID:              row.ExamID,
			SubjectID:           row.SubjectID,
			SubjectName:         row.SubjectName,
			Status:              row.Status,
			StepCount:           row.StepCount,
			VersionNo:           row.VersionNo,
			VersionID:           row.VersionID,
			UpdatedAt:           row.UpdatedAt,
			PublishedVersionNo:  row.PublishedVersionNo,
			HasUnpublishedDraft: row.HasUnpublishedDraft,
		})
	}
	return units
}

func (r *XormRepository) AdminCreateUnit(ctx context.Context, examID, subjectID, title string) (*AdminUnitSummary, error) {
	unitID := newUUIDLikeString()
	versionID := newUUIDLikeString()

	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()

	_, err := sess.Exec(`
		insert into interactive_units(id, exam_id, subject_id, title, status, created_at, updated_at)
		values (?::uuid, ?::uuid, nullif(?, '')::uuid, ?, 'draft', now(), now())
	`, unitID, examID, subjectID, title)
	if err != nil {
		return nil, err
	}

	_, err = sess.Exec(`
		insert into interactive_unit_versions(id, interactive_unit_id, version_no, status, metadata, created_at, updated_at)
		values (?::uuid, ?::uuid, 1, 'draft', '{}'::jsonb, now(), now())
	`, versionID, unitID)
	if err != nil {
		return nil, err
	}

	return &AdminUnitSummary{
		ID:        unitID,
		Title:     title,
		ExamID:    examID,
		SubjectID: subjectID,
		Status:    "draft",
		VersionNo: 1,
		VersionID: versionID,
	}, nil
}

func (r *XormRepository) AdminListVersions(ctx context.Context, unitID string) ([]AdminVersionSummary, error) {
	var rows []adminVersionSummaryRow
	err := r.engine.Context(ctx).SQL(`
		select iuv.id::text as version_id,
		       iuv.interactive_unit_id::text as unit_id,
		       iuv.version_no,
		       iuv.status,
		       to_char(iuv.published_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as published_at,
		       to_char(iuv.updated_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as updated_at
		from interactive_unit_versions iuv
		where iuv.interactive_unit_id = ?::uuid
		order by iuv.version_no desc
	`, unitID).Find(&rows)
	if err != nil {
		return nil, err
	}
	return mapAdminVersionSummaryRows(rows), nil
}

func mapAdminVersionSummaryRows(rows []adminVersionSummaryRow) []AdminVersionSummary {
	versions := make([]AdminVersionSummary, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, AdminVersionSummary{
			VersionID:   row.VersionID,
			UnitID:      row.UnitID,
			VersionNo:   row.VersionNo,
			Status:      row.Status,
			PublishedAt: row.PublishedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}
	return versions
}

func (r *XormRepository) AdminCreateVersion(ctx context.Context, unitID string) (*AdminVersionDetail, error) {
	var max struct {
		MaxNo int `xorm:"max_no"`
	}
	_, err := r.engine.Context(ctx).SQL(`
		select coalesce(max(version_no), 0) as max_no
		from interactive_unit_versions
		where interactive_unit_id = ?::uuid
	`, unitID).Get(&max)
	if err != nil {
		return nil, err
	}

	versionID := newUUIDLikeString()
	newVersionNo := max.MaxNo + 1

	var source struct {
		VersionID string `xorm:"version_id"`
		Title     string `xorm:"title"`
	}
	_, _ = r.engine.Context(ctx).SQL(`
		select iuv.id::text as version_id,
		       coalesce(iuv.metadata->>'title', iu.title) as title
		from interactive_unit_versions iuv
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		where iuv.interactive_unit_id = ?::uuid
		order by iuv.version_no desc
		limit 1
	`, unitID).Get(&source)

	_, err = r.engine.Context(ctx).Exec(`
		insert into interactive_unit_versions(id, interactive_unit_id, version_no, status, metadata, created_at, updated_at)
		values (?::uuid, ?::uuid, ?, 'draft', ?::jsonb, now(), now())
	`, versionID, unitID, newVersionNo, buildInteractiveUnitMetadataJSON(source.Title))
	if err != nil {
		return nil, err
	}

	if source.VersionID != "" {
		sourceSteps, _ := r.getVersionSteps(ctx, source.VersionID)
		for i, step := range sourceSteps {
			stepID := newUUIDLikeString()
			contentJSON, _ := json.Marshal(step.Content)
			initialStateJSON, _ := json.Marshal(step.InitialState)
			allowedActionsJSON, _ := json.Marshal(step.AllowedActions)
			evalConfigJSON, _ := json.Marshal(step.EvaluationConfig)
			feedbackMapJSON, _ := json.Marshal(step.FeedbackMap)
			hintPolicyJSON, _ := json.Marshal(step.HintPolicy)
			kpIDsJSON, _ := json.Marshal(step.KnowledgePointIDs)
			kpTagsJSON, _ := json.Marshal(step.KnowledgePointTags)
			_, _ = r.engine.Context(ctx).Exec(`
				insert into interactive_unit_version_steps(id, unit_version_id, step_no, widget_type, content, initial_state, allowed_actions, evaluation_config, feedback_map, hint_policy, knowledge_point_ids, knowledge_point_tags, created_at)
				values (?::uuid, ?::uuid, ?, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, now())
			`, stepID, versionID, i+1, step.WidgetType,
				string(contentJSON), string(initialStateJSON), string(allowedActionsJSON),
				string(evalConfigJSON), string(feedbackMapJSON), string(hintPolicyJSON), string(kpIDsJSON), string(kpTagsJSON))
		}
	}

	return r.AdminGetVersionDetail(ctx, versionID)
}

func (r *XormRepository) AdminGetVersionDetail(ctx context.Context, versionID string) (*AdminVersionDetail, error) {
	var meta struct {
		VersionID string `xorm:"version_id"`
		UnitID    string `xorm:"unit_id"`
		VersionNo int    `xorm:"version_no"`
		Status    string `xorm:"status"`
		Title     string `xorm:"title"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select iuv.id::text as version_id,
		       iuv.interactive_unit_id::text as unit_id,
		       iuv.version_no,
		       iuv.status,
		       coalesce(iuv.metadata->>'title', iu.title) as title
		from interactive_unit_versions iuv
		join interactive_units iu on iu.id = iuv.interactive_unit_id
		where iuv.id = ?::uuid
	`, versionID).Get(&meta)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	steps, err := r.getVersionSteps(ctx, versionID)
	if err != nil {
		return nil, err
	}

	return &AdminVersionDetail{
		VersionID: meta.VersionID,
		UnitID:    meta.UnitID,
		VersionNo: meta.VersionNo,
		Status:    meta.Status,
		Title:     meta.Title,
		Steps:     steps,
	}, nil
}

func (r *XormRepository) AdminUpdateVersion(ctx context.Context, versionID, title string, steps []StepSchema) (*AdminVersionDetail, error) {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()

	var currentStatus string
	has, err := sess.SQL(`select status from interactive_unit_versions where id = ?::uuid`, versionID).Get(&currentStatus)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrVersionNotFound
	}
	if currentStatus != "draft" {
		return nil, ErrVersionReadOnly
	}

	_, err = sess.Exec(`
		update interactive_unit_versions
		set metadata = ?::jsonb,
		    updated_at = now()
		where id = ?::uuid
	`, buildInteractiveUnitMetadataJSON(title), versionID)
	if err != nil {
		return nil, err
	}

	_, err = sess.Exec(`
		delete from interactive_unit_version_steps
		where unit_version_id = ?::uuid
	`, versionID)
	if err != nil {
		return nil, err
	}

	for i, step := range steps {
		stepID := step.ID
		if stepID == "" {
			stepID = newUUIDLikeString()
		}
		contentJSON, _ := json.Marshal(step.Content)
		initialStateJSON, _ := json.Marshal(step.InitialState)
		allowedActionsJSON, _ := json.Marshal(step.AllowedActions)
		evalConfigJSON, _ := json.Marshal(step.EvaluationConfig)
		feedbackMapJSON, _ := json.Marshal(step.FeedbackMap)
		hintPolicyJSON, _ := json.Marshal(step.HintPolicy)
		knowledgePointIDsJSON, _ := json.Marshal(step.KnowledgePointIDs)
		knowledgePointTagsJSON, _ := json.Marshal(step.KnowledgePointTags)

		_, err = sess.Exec(`
			insert into interactive_unit_version_steps(id, unit_version_id, step_no, widget_type, content, initial_state, allowed_actions, evaluation_config, feedback_map, hint_policy, knowledge_point_ids, knowledge_point_tags, created_at)
			values (?::uuid, ?::uuid, ?, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, now())
		`, stepID, versionID, i+1, step.WidgetType,
			string(contentJSON), string(initialStateJSON), string(allowedActionsJSON),
			string(evalConfigJSON), string(feedbackMapJSON), string(hintPolicyJSON), string(knowledgePointIDsJSON), string(knowledgePointTagsJSON))
		if err != nil {
			return nil, err
		}
	}

	return r.AdminGetVersionDetail(ctx, versionID)
}

func (r *XormRepository) AdminPublishVersion(ctx context.Context, versionID string) (*AdminVersionDetail, error) {
	detail, err := r.AdminGetVersionDetail(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, ErrVersionNotFound
	}
	if detail.Status != "draft" {
		return nil, ErrVersionNotPublishable
	}
	if strings.TrimSpace(detail.Title) == "" {
		return nil, ErrVersionNotPublishable
	}
	if len(detail.Steps) == 0 {
		return nil, ErrVersionNotPublishable
	}

	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()

	res, err := sess.Exec(`
		update interactive_unit_versions
		set status = 'published', published_at = now(), updated_at = now()
		where id = ?::uuid and status = 'draft'
	`, versionID)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, ErrVersionNotPublishable
	}

	var unitID string
	_, err = sess.SQL(`select interactive_unit_id::text from interactive_unit_versions where id = ?::uuid`, versionID).Get(&unitID)
	if err != nil {
		return nil, err
	}

	var versionTitle string
	_, _ = sess.SQL(`select coalesce(metadata->>'title', '') from interactive_unit_versions where id = ?::uuid`, versionID).Get(&versionTitle)

	_, err = sess.Exec(`
		update interactive_units
		set current_published_version_id = ?::uuid, status = 'published', title = nullif(?, ''), updated_at = now()
		where id = ?::uuid
	`, versionID, versionTitle, unitID)
	if err != nil {
		return nil, err
	}

	return r.AdminGetVersionDetail(ctx, versionID)
}

func (r *XormRepository) AdminDeleteUnit(ctx context.Context, unitID string) error {
	_, err := r.engine.Context(ctx).Exec(`delete from interactive_units where id = ?::uuid`, unitID)
	return err
}

func (r *XormRepository) getVersionSteps(ctx context.Context, versionID string) ([]StepSchema, error) {
	var stepRows []struct {
		ID                 string `xorm:"id"`
		WidgetType         string `xorm:"widget_type"`
		Content            string `xorm:"content"`
		InitialState       string `xorm:"initial_state"`
		AllowedActions     string `xorm:"allowed_actions"`
		EvaluationConfig   string `xorm:"evaluation_config"`
		FeedbackMap        string `xorm:"feedback_map"`
		HintPolicy         string `xorm:"hint_policy"`
		KnowledgePointIDs  string `xorm:"knowledge_point_ids"`
		KnowledgePointTags string `xorm:"knowledge_point_tags"`
	}
	err := r.engine.Context(ctx).SQL(`
		select id::text, widget_type, content::text, initial_state::text, allowed_actions::text, evaluation_config::text, feedback_map::text, hint_policy::text, knowledge_point_ids::text, knowledge_point_tags::text
		from interactive_unit_version_steps
		where unit_version_id = ?::uuid
		order by step_no asc
	`, versionID).Find(&stepRows)
	if err != nil {
		return nil, err
	}

	steps := make([]StepSchema, 0, len(stepRows))
	for _, sr := range stepRows {
		steps = append(steps, StepSchema{
			ID:                 sr.ID,
			WidgetType:         sr.WidgetType,
			Content:            decodeMap([]byte(sr.Content)),
			InitialState:       decodeMap([]byte(sr.InitialState)),
			AllowedActions:     decodeMap([]byte(sr.AllowedActions)),
			EvaluationConfig:   decodeMap([]byte(sr.EvaluationConfig)),
			FeedbackMap:        decodeMap([]byte(sr.FeedbackMap)),
			HintPolicy:         decodeMap([]byte(sr.HintPolicy)),
			KnowledgePointIDs:  decodeStringSlice([]byte(sr.KnowledgePointIDs)),
			KnowledgePointTags: decodeStringSlice([]byte(sr.KnowledgePointTags)),
		})
	}
	return steps, nil
}

func decodeMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		return map[string]any{}
	}
	return out
}

func decodeStringSlice(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		return []string{}
	}
	return out
}

func buildInteractiveUnitMetadataJSON(title string) string {
	metadata := map[string]any{}
	if trimmed := strings.TrimSpace(title); trimmed != "" {
		metadata["title"] = trimmed
	}
	body, _ := json.Marshal(metadata)
	return string(body)
}
