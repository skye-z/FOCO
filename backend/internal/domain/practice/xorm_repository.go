package practice

import (
	"context"
	"encoding/json"
	"fmt"
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

func (r *XormRepository) FindEnrollment(ctx context.Context, userID, examID string) (*EnrollmentRef, error) {
	var row struct {
		ID       string `xorm:"id"`
		ExamID   string `xorm:"exam_id"`
		ExamName string `xorm:"exam_name"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select ee.id::text as id,
		       ee.exam_id::text as exam_id,
		       e.name as exam_name
		from exam_enrollments ee
		join exams e on e.id = ee.exam_id
		where ee.user_id = ?::uuid
		  and ee.exam_id = ?::uuid
		order by case
			when ee.status = 'in_progress' then 0
			when ee.status = 'passed_manual' then 1
			else 2
		end, ee.created_at desc
		limit 1
	`, userID, examID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &EnrollmentRef{
		ID:       row.ID,
		ExamID:   row.ExamID,
		ExamName: row.ExamName,
	}, nil
}

func (r *XormRepository) FindLatestDiagnosticRecommendation(ctx context.Context, userID, examID string) (*DiagnosticRecommendation, error) {
	var row struct {
		ProfileSummary string `xorm:"profile_summary"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select profile_summary::text as profile_summary
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

	var payload struct {
		RecommendedDifficulty        string   `json:"recommended_difficulty"`
		RecommendedSubjectIDs        []string `json:"recommended_subject_ids"`
		RecommendedChapterIDs        []string `json:"recommended_chapter_ids"`
		RecommendedKnowledgePointIDs []string `json:"recommended_knowledge_point_ids"`
	}
	_ = json.Unmarshal([]byte(row.ProfileSummary), &payload)

	return &DiagnosticRecommendation{
		RecommendedDifficulty:        payload.RecommendedDifficulty,
		RecommendedSubjectIDs:        payload.RecommendedSubjectIDs,
		RecommendedChapterIDs:        payload.RecommendedChapterIDs,
		RecommendedKnowledgePointIDs: payload.RecommendedKnowledgePointIDs,
	}, nil
}

func (r *XormRepository) ListCandidateQuestions(ctx context.Context, input CreateSessionInput) ([]CandidateQuestion, error) {
	query := `
		select q.id::text as question_id,
		       q.subject_id::text as subject_id,
		       coalesce(q.chapter_id::text, '') as chapter_id,
		       qv.id::text as question_version_id,
		       qv.question_type as question_type,
		       qv.difficulty as difficulty,
		       qv.stem::text as stem,
		       coalesce(qv.options::text, '{"choices":[]}') as options,
		       qv.correct_answer::text as correct_answer,
		       qv.explanation::text as explanation,
		       coalesce(
		           jsonb_agg(qvkp.knowledge_point_id::text) filter (where qvkp.knowledge_point_id is not null),
		           '[]'::jsonb
		       )::text as knowledge_point_ids
		from questions q
		join question_versions qv on qv.id = q.current_published_version_id
		left join question_version_knowledge_points qvkp on qvkp.question_version_id = qv.id
		where q.exam_id = ?::uuid
		  and q.status = 'published'
	`
	args := []any{input.ExamID}

	if len(input.QuestionTypes) > 0 {
		query += ` and qv.question_type in (` + placeholders(len(input.QuestionTypes)) + `)`
		args = append(args, toInterface(input.QuestionTypes)...)
	}
	if len(input.SubjectIDs) > 0 {
		query += ` and q.subject_id::text in (` + placeholders(len(input.SubjectIDs)) + `)`
		args = append(args, toInterface(input.SubjectIDs)...)
	}
	if len(input.ChapterIDs) > 0 {
		query += ` and q.chapter_id::text in (` + placeholders(len(input.ChapterIDs)) + `)`
		args = append(args, toInterface(input.ChapterIDs)...)
	}
	if len(input.KnowledgePointIDs) > 0 {
		query += ` and exists (
			select 1
			from question_version_knowledge_points qvkp_filter
			where qvkp_filter.question_version_id = qv.id
			  and qvkp_filter.knowledge_point_id::text in (` + placeholders(len(input.KnowledgePointIDs)) + `)
		)`
		args = append(args, toInterface(input.KnowledgePointIDs)...)
	}

	switch input.Difficulty {
	case "easy":
		query += ` and qv.difficulty <= 2`
	case "medium":
		query += ` and qv.difficulty = 3`
	case "hard":
		query += ` and qv.difficulty >= 4`
	}

	query += `
		group by q.id, q.subject_id, q.chapter_id, qv.id, qv.question_type, qv.difficulty, qv.stem, qv.options, qv.correct_answer, qv.explanation
		order by random()
		limit ?
	`
	args = append(args, input.Count)

	var rows []struct {
		QuestionID        string `xorm:"question_id"`
		SubjectID         string `xorm:"subject_id"`
		ChapterID         string `xorm:"chapter_id"`
		QuestionVersionID string `xorm:"question_version_id"`
		QuestionType      string `xorm:"question_type"`
		Difficulty        int    `xorm:"difficulty"`
		Stem              string `xorm:"stem"`
		Options           string `xorm:"options"`
		CorrectAnswer     string `xorm:"correct_answer"`
		Explanation       string `xorm:"explanation"`
		KnowledgePointIDs string `xorm:"knowledge_point_ids"`
	}
	if err := r.engine.Context(ctx).SQL(query, args...).Find(&rows); err != nil {
		return nil, err
	}

	result := make([]CandidateQuestion, 0, len(rows))
	for _, row := range rows {
		result = append(result, CandidateQuestion{
			QuestionID:        row.QuestionID,
			QuestionVersionID: row.QuestionVersionID,
			SubjectID:         row.SubjectID,
			ChapterID:         row.ChapterID,
			QuestionType:      row.QuestionType,
			Difficulty:        row.Difficulty,
			Stem:              row.Stem,
			Options:           row.Options,
			CorrectAnswer:     row.CorrectAnswer,
			Explanation:       row.Explanation,
			KnowledgePointIDs: decodeStringSlice(row.KnowledgePointIDs),
		})
	}
	return result, nil
}

func (r *XormRepository) CreateSession(ctx context.Context, session *PracticeSessionRecord, items []PracticeSessionItemRecord) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	_, err := sess.Exec(`
		insert into practice_sessions(
			id, user_id, exam_id, exam_enrollment_id, scope, status, started_at, total_count, answered_count, correct_count,
			total_duration_seconds, xp_earned, coins_earned, created_at, updated_at
		) values (
			?::uuid, ?::uuid, ?::uuid, ?::uuid, ?::jsonb, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`, session.ID, session.UserID, session.ExamID, session.ExamEnrollmentID, session.ScopeJSON, session.Status, session.StartedAt, session.TotalCount, session.AnsweredCount, session.CorrectCount, session.TotalDurationSeconds, session.XpEarned, session.CoinsEarned, session.CreatedAt, session.UpdatedAt)
	if err != nil {
		_ = sess.Rollback()
		return err
	}

	for _, item := range items {
		_, err = sess.Exec(`
			insert into practice_session_items(
				id, practice_session_id, session_id, question_id, question_version_id, subject_id, chapter_id, question_type, score,
				position, seq_no, status, stem, options, correct_labels, explanation, knowledge_point_ids, created_at
			) values (
				?::uuid, ?::uuid, ?::uuid, ?::uuid, ?::uuid, ?::uuid, nullif(?, '')::uuid, ?, ?, ?, ?, 'pending', ?, ?::jsonb, ?::jsonb, ?, ?::jsonb, ?
			)
		`, item.ID, item.SessionID, item.SessionID, item.QuestionID, item.QuestionVersionID, item.SubjectID, item.ChapterID, item.QuestionType, item.Score, item.Position, item.Position, item.Stem, item.Options, item.CorrectLabels, item.Explanation, item.KnowledgePointIDs, item.CreatedAt)
		if err != nil {
			_ = sess.Rollback()
			return err
		}
	}

	return sess.Commit()
}

func (r *XormRepository) GetSession(ctx context.Context, userID, sessionID string) (*PracticeSessionView, error) {
	var meta struct {
		SessionID  string `xorm:"session_id"`
		ExamName   string `xorm:"exam_name"`
		TotalCount int    `xorm:"total_count"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select ps.id::text as session_id,
		       e.name as exam_name,
		       ps.total_count as total_count
		from practice_sessions ps
		join exams e on e.id = ps.exam_id
		where ps.id = ?::uuid and ps.user_id = ?::uuid
	`, sessionID, userID).Get(&meta)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	var rows []struct {
		ItemID            string  `xorm:"item_id"`
		QuestionVersionID string  `xorm:"question_version_id"`
		QuestionType      string  `xorm:"question_type"`
		Score             int     `xorm:"score"`
		Stem              string  `xorm:"stem"`
		Options           string  `xorm:"options"`
		SubmittedAt       *string `xorm:"submitted_at"`
		IsCorrect         *bool   `xorm:"is_correct"`
		UserAnswer        string  `xorm:"user_answer"`
	}
	if err := r.engine.Context(ctx).SQL(`
		select psi.id::text as item_id,
		       psi.question_version_id::text as question_version_id,
		       psi.question_type as question_type,
		       psi.score as score,
		       psi.stem as stem,
		       psi.options::text as options,
		       psi.submitted_at::text as submitted_at,
		       psi.is_correct as is_correct,
		       coalesce(psi.user_answer::text, '') as user_answer
		from practice_session_items psi
		join practice_sessions ps on ps.id = psi.session_id
		where ps.id = ?::uuid and ps.user_id = ?::uuid
		order by coalesce(psi.position, psi.seq_no) asc
	`, sessionID, userID).Find(&rows); err != nil {
		return nil, err
	}

	items := make([]PracticeSessionItemView, 0, len(rows))
	for _, row := range rows {
		item := PracticeSessionItemView{
			ItemID:            row.ItemID,
			QuestionVersionID: row.QuestionVersionID,
			QuestionType:      row.QuestionType,
			Score:             row.Score,
			Content: PracticeItemContent{
				Stem:    row.Stem,
				Options: decodeOptions(row.Options),
			},
			Submitted: row.SubmittedAt != nil,
			IsCorrect: row.IsCorrect,
		}
		if row.UserAnswer != "" {
			_ = json.Unmarshal([]byte(row.UserAnswer), &item.UserAnswer)
		}
		items = append(items, item)
	}

	return &PracticeSessionView{
		SessionID:  meta.SessionID,
		ExamName:   meta.ExamName,
		TotalCount: meta.TotalCount,
		Items:      items,
	}, nil
}

func (r *XormRepository) GetSubmission(ctx context.Context, userID, sessionID, itemID string) (*SubmissionRecord, error) {
	var row struct {
		SessionID         string     `xorm:"session_id"`
		ItemID            string     `xorm:"item_id"`
		ExamID            string     `xorm:"exam_id"`
		QuestionType      string     `xorm:"question_type"`
		CorrectLabels     string     `xorm:"correct_labels"`
		Explanation       string     `xorm:"explanation"`
		KnowledgePointIDs string     `xorm:"knowledge_point_ids"`
		SubmittedAt       *time.Time `xorm:"submitted_at"`
		IsCorrect         *bool      `xorm:"is_correct"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select ps.id::text as session_id,
		       psi.id::text as item_id,
		       ps.exam_id::text as exam_id,
		       psi.question_type as question_type,
		       psi.correct_labels::text as correct_labels,
		       psi.explanation as explanation,
		       psi.knowledge_point_ids::text as knowledge_point_ids,
		       psi.submitted_at as submitted_at,
		       psi.is_correct as is_correct
		from practice_session_items psi
		join practice_sessions ps on ps.id = coalesce(psi.session_id, psi.practice_session_id)
		where ps.id = ?::uuid
		  and ps.user_id = ?::uuid
		  and psi.id = ?::uuid
	`, sessionID, userID, itemID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &SubmissionRecord{
		SessionID:         row.SessionID,
		ItemID:            row.ItemID,
		ExamID:            row.ExamID,
		QuestionType:      row.QuestionType,
		CorrectLabels:     decodeStringSlice(row.CorrectLabels),
		Explanation:       row.Explanation,
		KnowledgePointIDs: decodeStringSlice(row.KnowledgePointIDs),
		SubmittedAt:       row.SubmittedAt,
		IsCorrect:         row.IsCorrect,
	}, nil
}

func (r *XormRepository) SaveSubmission(ctx context.Context, sessionID, itemID string, correct bool, userAnswer []string, durationSeconds, xpEarned, coinsEarned int) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	userAnswerJSON, _ := json.Marshal(userAnswer)
	result, err := sess.Exec(`
		update practice_session_items
		set user_answer = ?::jsonb,
		    is_correct = ?,
		    duration_seconds = ?,
		    submitted_at = now(),
		    status = 'answered'
		where id = ?::uuid
		  and coalesce(session_id, practice_session_id) = ?::uuid
		  and submitted_at is null
	`, string(userAnswerJSON), correct, durationSeconds, itemID, sessionID)
	if err != nil {
		_ = sess.Rollback()
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		_ = sess.Rollback()
		return fmt.Errorf("practice item already submitted")
	}

	correctDelta := 0
	if correct {
		correctDelta = 1
	}
	_, err = sess.Exec(`
		update practice_sessions
		set answered_count = least(total_count, answered_count + 1),
		    correct_count = correct_count + ?,
		    total_duration_seconds = total_duration_seconds + ?,
		    xp_earned = xp_earned + ?,
		    coins_earned = coins_earned + ?,
		    status = case when answered_count + 1 >= total_count then 'completed' else status end,
		    completed_at = case when answered_count + 1 >= total_count then now() else completed_at end,
		    updated_at = now()
		where id = ?::uuid
	`, correctDelta, durationSeconds, xpEarned, coinsEarned, sessionID)
	if err != nil {
		_ = sess.Rollback()
		return err
	}

	if coinsEarned > 0 {
		_, err = sess.Exec(`
			insert into wallets (id, user_id, coins_balance, created_at, updated_at)
			values (gen_random_uuid(), (select user_id from practice_sessions where id = ?::uuid), ?, now(), now())
			on conflict (user_id) do update set coins_balance = wallets.coins_balance + ?, updated_at = now()
		`, sessionID, coinsEarned, coinsEarned)
		if err != nil {
			_ = sess.Rollback()
			return err
		}
	}

	return sess.Commit()
}

func (r *XormRepository) ListKnowledgePoints(ctx context.Context, ids []string) ([]KnowledgePointRef, error) {
	if len(ids) == 0 {
		return []KnowledgePointRef{}, nil
	}

	var rows []KnowledgePointRef
	if err := r.engine.Context(ctx).SQL(`
		select id::text as id, name
		from knowledge_points
		where id::text in (`+placeholders(len(ids))+`)
		order by name asc
	`, toInterface(ids)...).Find(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *XormRepository) GetSummary(ctx context.Context, userID, sessionID string) (*SessionSummary, error) {
	var row struct {
		TotalCount           int `xorm:"total_count"`
		CorrectCount         int `xorm:"correct_count"`
		AnsweredCount        int `xorm:"answered_count"`
		XpEarned             int `xorm:"xp_earned"`
		CoinsEarned          int `xorm:"coins_earned"`
		TotalDurationSeconds int `xorm:"total_duration_seconds"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select total_count,
		       correct_count,
		       answered_count,
		       xp_earned,
		       coins_earned,
		       total_duration_seconds
		from practice_sessions
		where id = ?::uuid and user_id = ?::uuid
	`, sessionID, userID).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}

	wrong := row.AnsweredCount - row.CorrectCount
	if wrong < 0 {
		wrong = 0
	}
	accuracy := 0
	if row.AnsweredCount > 0 {
		accuracy = int(float64(row.CorrectCount) * 100.0 / float64(row.AnsweredCount))
	}

	return &SessionSummary{
		Total:           row.TotalCount,
		Correct:         row.CorrectCount,
		Wrong:           wrong,
		Accuracy:        accuracy,
		XpEarned:        row.XpEarned,
		CoinsEarned:     row.CoinsEarned,
		DurationMinutes: durationMinutes(row.TotalDurationSeconds),
	}, nil
}

func (r *XormRepository) ListWrongBook(ctx context.Context, filter WrongBookFilter) ([]WrongBookItem, error) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 200
	}
	if pageSize > 500 {
		pageSize = 500
	}

	query := `
		with submitted as (
			select psi.id::text as item_id,
			       psi.question_id::text as question_id,
			       psi.question_type,
			       psi.stem,
			       psi.options::text as options,
			       psi.user_answer::text as user_answer,
			       psi.correct_labels::text as correct_answer,
			       psi.explanation,
			       psi.is_correct,
			       psi.submitted_at,
			       psi.subject_id::text as subject_id,
			       coalesce(psi.chapter_id::text, '') as chapter_id,
			       s.name as subject_name,
			       coalesce(c.name, '') as chapter_name,
			       coalesce(kp_meta.knowledge_points, '[]') as knowledge_points
			from practice_session_items psi
			join practice_sessions ps on ps.id = coalesce(psi.session_id, psi.practice_session_id)
			join subjects s on s.id = psi.subject_id
			left join chapters c on c.id = psi.chapter_id
			left join lateral (
				select jsonb_agg(
					jsonb_build_object('id', kp.id::text, 'name', kp.name)
					order by kp.name
				)::text as knowledge_points
				from jsonb_array_elements_text(coalesce(psi.knowledge_point_ids, '[]'::jsonb)) kp_ref(id)
				join knowledge_points kp on kp.id::text = kp_ref.id
			) kp_meta on true
			where ps.user_id = ?::uuid
			  and psi.submitted_at is not null
	`
	args := []any{filter.UserID}
	if filter.ExamID != "" {
		query += ` and ps.exam_id = ?::uuid`
		args = append(args, filter.ExamID)
	}
	if filter.SubjectID != "" {
		query += ` and psi.subject_id = ?::uuid`
		args = append(args, filter.SubjectID)
	}
	if filter.ChapterID != "" {
		query += ` and psi.chapter_id = ?::uuid`
		args = append(args, filter.ChapterID)
	}
	if filter.KnowledgePointID != "" {
		query += ` and exists (
			select 1
			from jsonb_array_elements_text(coalesce(psi.knowledge_point_ids, '[]'::jsonb)) kp_filter(id)
			where kp_filter.id = ?
		)`
		args = append(args, filter.KnowledgePointID)
	}
	query += `
		),
		wrong_groups as (
			select question_id,
			       min(submitted_at) filter (where is_correct = false) as first_error_at,
			       max(submitted_at) filter (where is_correct = false) as last_error_at,
			       count(*) filter (where is_correct = false) as error_count
			from submitted
			group by question_id
			having count(*) filter (where is_correct = false) > 0
		),
		correction_groups as (
			select s.question_id,
			       count(*) filter (
			           where s.is_correct = true
			             and s.submitted_at > wg.first_error_at
			       ) as fix_count
			from submitted s
			join wrong_groups wg on wg.question_id = s.question_id
			group by s.question_id
		),
		latest_wrong as (
			select distinct on (question_id)
			       question_id,
			       item_id,
			       question_type,
			       stem,
			       options,
			       user_answer,
			       correct_answer,
			       explanation,
			       subject_id,
			       chapter_id,
			       subject_name,
			       chapter_name,
			       knowledge_points
			from submitted
			where is_correct = false
			order by question_id, submitted_at desc
		),
		latest_any as (
			select distinct on (question_id)
			       question_id,
			       is_correct as latest_correct
			from submitted
			order by question_id, submitted_at desc
		)
		select lw.item_id as id,
		       lw.question_id,
		       lw.question_type,
		       lw.stem,
		       lw.options,
		       lw.user_answer,
		       lw.correct_answer,
		       lw.explanation,
		       wg.error_count,
		       coalesce(cg.fix_count, 0) as fix_count,
		       wg.first_error_at,
		       wg.last_error_at,
		       case
		           when coalesce(la.latest_correct, false) = true then 'mastered'
		           when coalesce(cg.fix_count, 0) > 0 then 'reviewing'
		           else 'open'
		       end as status,
		       lw.subject_id,
		       lw.chapter_id,
		       lw.subject_name,
		       lw.chapter_name,
		       lw.knowledge_points
		from wrong_groups wg
		join latest_wrong lw on lw.question_id = wg.question_id
		left join latest_any la on la.question_id = wg.question_id
		left join correction_groups cg on cg.question_id = wg.question_id
		where 1=1
	`
	if filter.Status != "" {
		query += ` and case
		           when coalesce(la.latest_correct, false) = true then 'mastered'
		           when coalesce(cg.fix_count, 0) > 0 then 'reviewing'
		           else 'open'
		       end = ?`
		args = append(args, filter.Status)
	}
	query += ` order by wg.last_error_at desc limit ? offset ?`
	args = append(args, pageSize, (page-1)*pageSize)

	var rows []struct {
		ID              string    `xorm:"id"`
		QuestionID      string    `xorm:"question_id"`
		QuestionType    string    `xorm:"question_type"`
		Stem            string    `xorm:"stem"`
		Options         string    `xorm:"options"`
		UserAnswer      string    `xorm:"user_answer"`
		CorrectAnswer   string    `xorm:"correct_answer"`
		Explanation     string    `xorm:"explanation"`
		ErrorCount      int       `xorm:"error_count"`
		FixCount        int       `xorm:"fix_count"`
		FirstErrorAt    time.Time `xorm:"first_error_at"`
		LastErrorAt     time.Time `xorm:"last_error_at"`
		Status          string    `xorm:"status"`
		SubjectID       string    `xorm:"subject_id"`
		SubjectName     string    `xorm:"subject_name"`
		ChapterID       string    `xorm:"chapter_id"`
		ChapterName     string    `xorm:"chapter_name"`
		KnowledgePoints string    `xorm:"knowledge_points"`
	}
	if err := r.engine.Context(ctx).SQL(query, args...).Find(&rows); err != nil {
		return nil, err
	}

	items := make([]WrongBookItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, WrongBookItem{
			ID:              row.ID,
			QuestionID:      row.QuestionID,
			QuestionType:    row.QuestionType,
			Stem:            row.Stem,
			Options:         decodeOptions(row.Options),
			UserAnswer:      formatAnswerLabels(row.UserAnswer),
			CorrectAnswer:   formatAnswerLabels(row.CorrectAnswer),
			Explanation:     row.Explanation,
			ErrorCount:      row.ErrorCount,
			FixCount:        row.FixCount,
			FirstErrorAt:    row.FirstErrorAt.Format("2006-01-02"),
			LastErrorAt:     row.LastErrorAt.Format("2006-01-02"),
			Status:          row.Status,
			SubjectID:       row.SubjectID,
			SubjectName:     row.SubjectName,
			ChapterID:       row.ChapterID,
			ChapterName:     row.ChapterName,
			KnowledgePoints: decodeKnowledgePointRefs(row.KnowledgePoints),
		})
	}
	return items, nil
}

func decodeKnowledgePointRefs(raw string) []KnowledgePointRef {
	if strings.TrimSpace(raw) == "" {
		return []KnowledgePointRef{}
	}
	var refs []KnowledgePointRef
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return []KnowledgePointRef{}
	}
	return refs
}

func decodeStringSlice(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var values []string
	_ = json.Unmarshal([]byte(raw), &values)
	if values == nil {
		return []string{}
	}
	return values
}

func formatAnswerLabels(raw string) string {
	labels := decodeStringSlice(raw)
	if len(labels) == 0 {
		return "未作答"
	}
	return strings.Join(labels, ", ")
}

func decodeOptions(raw string) []QuestionOption {
	if strings.TrimSpace(raw) == "" {
		return []QuestionOption{}
	}
	var values []QuestionOption
	_ = json.Unmarshal([]byte(raw), &values)
	if values == nil {
		return []QuestionOption{}
	}
	return values
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	items := make([]string, 0, count)
	for range count {
		items = append(items, "?")
	}
	return strings.Join(items, ",")
}

func toInterface(items []string) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}
