package profile

import (
	"context"
	"encoding/json"

	"xorm.io/xorm"
)

type XormRepository struct {
	engine *xorm.Engine
}

func NewXormRepository(engine *xorm.Engine) *XormRepository {
	return &XormRepository{engine: engine}
}

func (r *XormRepository) GetWalletBalance(ctx context.Context, userID string) (int, error) {
	var row struct {
		CoinsBalance int `xorm:"coins_balance"`
	}
	has, err := r.engine.Context(ctx).SQL(`
		select coins_balance
		from wallets
		where user_id = ?::uuid
		limit 1
	`, userID).Get(&row)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil
	}
	return row.CoinsBalance, nil
}

func (r *XormRepository) GetStreakStats(ctx context.Context, userID, examID string) (*StreakStats, error) {
	var row struct {
		CurrentStreak int    `xorm:"current_streak"`
		BestStreak    int    `xorm:"best_streak"`
		Status        string `xorm:"status"`
	}

	query := `
		select s.current_streak, s.best_streak, s.status
		from streaks s
		join exam_enrollments ee on ee.id = s.exam_enrollment_id
		where ee.user_id = ?::uuid
	`
	args := []any{userID}
	if examID != "" {
		query += ` and ee.exam_id = ?::uuid`
		args = append(args, examID)
	}
	query += `
		order by ee.created_at desc
		limit 1
	`

	has, err := r.engine.Context(ctx).SQL(query, args...).Get(&row)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &StreakStats{
		CurrentStreak: row.CurrentStreak,
		BestStreak:    row.BestStreak,
		Status:        row.Status,
	}, nil
}

func (r *XormRepository) ListPracticeSessions(ctx context.Context, userID, examID string, limit int) ([]PracticeSessionSummary, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		select id::text as id,
		       status,
		       total_count,
		       answered_count,
		       correct_count,
		       xp_earned,
		       coins_earned,
		       total_duration_seconds,
		       created_at,
		       completed_at
		from practice_sessions
		where user_id = ?::uuid
	`
	args := []any{userID}
	if examID != "" {
		query += ` and exam_id = ?::uuid`
		args = append(args, examID)
	}
	query += ` order by created_at desc limit ?`
	args = append(args, limit)

	rows := make([]PracticeSessionSummary, 0)
	if err := r.engine.Context(ctx).SQL(query, args...).Find(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *XormRepository) ListKnowledgePointResults(ctx context.Context, userID, examID string) ([]KnowledgePointResult, error) {
	query := `
		select kp.id::text as knowledge_point_id,
		       kp.name as knowledge_point_name,
		       coalesce(psi.is_correct, false) as correct,
		       psi.submitted_at as submitted_at
		from practice_session_items psi
		join practice_sessions ps on ps.id = coalesce(psi.session_id, psi.practice_session_id)
		join lateral jsonb_array_elements_text(coalesce(psi.knowledge_point_ids, '[]'::jsonb)) as kp_ref(id) on true
		join knowledge_points kp on kp.id::text = kp_ref.id
		where ps.user_id = ?::uuid
		  and psi.submitted_at is not null
	`
	args := []any{userID}
	if examID != "" {
		query += ` and ps.exam_id = ?::uuid`
		args = append(args, examID)
	}
	query += ` order by kp.name asc`

	rows := make([]KnowledgePointResult, 0)
	if err := r.engine.Context(ctx).SQL(query, args...).Find(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *XormRepository) BuildKnowledgeGraph(ctx context.Context, examID string) (*KnowledgeGraph, error) {
	nodes := []KnowledgeGraphNode{}
	edges := []KnowledgeGraphEdge{}

	examQuery := `
		select ('exam:' || id::text) as id,
		       'exam' as type,
		       id::text as ref_id,
		       name as label,
		       code as description,
		       code as "group"
		from exams
		where status = 'active'
	`
	examArgs := []any{}
	if examID != "" {
		examQuery += ` and id = ?::uuid`
		examArgs = append(examArgs, examID)
	}
	examQuery += ` order by created_at asc`
	if err := r.engine.Context(ctx).SQL(examQuery, examArgs...).Find(&nodes); err != nil {
		return nil, err
	}

	var subjectRows []KnowledgeGraphNode
	subjectQuery := `
		select ('subject:' || id::text) as id,
		       'subject' as type,
		       id::text as ref_id,
		       name as label,
		       code as description,
		       code as "group"
		from subjects
	`
	subjectArgs := []any{}
	if examID != "" {
		subjectQuery += ` where exam_id = ?::uuid`
		subjectArgs = append(subjectArgs, examID)
	}
	subjectQuery += ` order by exam_id asc, sort_order asc, created_at asc`
	if err := r.engine.Context(ctx).SQL(subjectQuery, subjectArgs...).Find(&subjectRows); err != nil {
		return nil, err
	}
	nodes = append(nodes, subjectRows...)

	var examSubjectEdges []KnowledgeGraphEdge
	examSubjectQuery := `
		select ('examsubject:' || exam_id::text || ':' || id::text) as id,
		       ('exam:' || exam_id::text) as source,
		       ('subject:' || id::text) as target,
		       'exam_subject' as type,
		       '包含' as label
		from subjects
	`
	if examID != "" {
		examSubjectQuery += ` where exam_id = ?::uuid`
	}
	examSubjectQuery += ` order by exam_id asc, sort_order asc, created_at asc`
	if err := r.engine.Context(ctx).SQL(examSubjectQuery, subjectArgs...).Find(&examSubjectEdges); err != nil {
		return nil, err
	}
	edges = append(edges, examSubjectEdges...)

	var chapterRows []KnowledgeGraphNode
	chapterQuery := `
		select ('chapter:' || c.id::text) as id,
		       'chapter' as type,
		       c.id::text as ref_id,
		       c.name as label,
		       c.code as description,
		       c.code as "group"
		from chapters c
		join subjects s on s.id = c.subject_id
	`
	chapterArgs := []any{}
	if examID != "" {
		chapterQuery += ` where s.exam_id = ?::uuid`
		chapterArgs = append(chapterArgs, examID)
	}
	chapterQuery += ` order by c.subject_id asc, c.sort_order asc, c.created_at asc`
	if err := r.engine.Context(ctx).SQL(chapterQuery, chapterArgs...).Find(&chapterRows); err != nil {
		return nil, err
	}
	nodes = append(nodes, chapterRows...)

	var subjectChapterEdges []KnowledgeGraphEdge
	chapterEdgeQuery := `
		select ('subjectchapter:' || c.subject_id::text || ':' || c.id::text) as id,
		       ('subject:' || c.subject_id::text) as source,
		       ('chapter:' || c.id::text) as target,
		       'subject_chapter' as type,
		       '包含' as label
		from chapters c
		join subjects s on s.id = c.subject_id
	`
	if examID != "" {
		chapterEdgeQuery += ` where s.exam_id = ?::uuid`
	}
	chapterEdgeQuery += ` order by c.subject_id asc, c.sort_order asc, c.created_at asc`
	if err := r.engine.Context(ctx).SQL(chapterEdgeQuery, chapterArgs...).Find(&subjectChapterEdges); err != nil {
		return nil, err
	}
	edges = append(edges, subjectChapterEdges...)

	var kpRows []KnowledgeGraphNode
	kpQuery := `
		select ('kp:' || kp.id::text) as id,
		       'knowledge_point' as type,
		       kp.id::text as ref_id,
		       kp.name as label,
		       coalesce(kp.description, '') as description,
		       kp.code as "group"
		from knowledge_points kp
		where kp.status = 'active'
	`
	kpArgs := []any{}
	if examID != "" {
		kpQuery += ` and kp.exam_id = ?::uuid`
		kpArgs = append(kpArgs, examID)
	}
	kpQuery += ` order by kp.exam_id asc, kp.name asc`
	if err := r.engine.Context(ctx).SQL(kpQuery, kpArgs...).Find(&kpRows); err != nil {
		return nil, err
	}
	nodes = append(nodes, kpRows...)

	var locationEdges []KnowledgeGraphEdge
	locationEdgeQuery := `
		select distinct
		       case
		           when q.chapter_id is not null then ('chapterkp:' || q.chapter_id::text || ':' || qvkp.knowledge_point_id::text)
		           else ('subjectkp:' || q.subject_id::text || ':' || qvkp.knowledge_point_id::text)
		       end as id,
		       case
		           when q.chapter_id is not null then ('chapter:' || q.chapter_id::text)
		           else ('subject:' || q.subject_id::text)
		       end as source,
		       ('kp:' || qvkp.knowledge_point_id::text) as target,
		       case
		           when q.chapter_id is not null then 'chapter_knowledge_point'
		           else 'subject_knowledge_point'
		       end as type,
		       '关联' as label
		from questions q
		join question_versions qv on qv.id = q.current_published_version_id
		join question_version_knowledge_points qvkp on qvkp.question_version_id = qv.id
		join knowledge_points kp on kp.id = qvkp.knowledge_point_id
		where kp.status = 'active'
	`
	if examID != "" {
		locationEdgeQuery += ` and q.exam_id = ?::uuid`
	}
	if err := r.engine.Context(ctx).SQL(locationEdgeQuery, kpArgs...).Find(&locationEdges); err != nil {
		return nil, err
	}
	edges = append(edges, locationEdges...)

	var kpEdges []KnowledgeGraphEdge
	kpEdgeQuery := `
		select kpe.id::text as id,
		       ('kp:' || kpe.from_knowledge_point_id::text) as source,
		       ('kp:' || kpe.to_knowledge_point_id::text) as target,
		       kpe.edge_type as type,
		       kpe.edge_type as label
		from knowledge_point_edges kpe
	`
	kpEdgeArgs := []any{}
	if examID != "" {
		kpEdgeQuery += ` where kpe.exam_id = ?::uuid`
		kpEdgeArgs = append(kpEdgeArgs, examID)
	}
	kpEdgeQuery += ` order by kpe.created_at asc`
	if err := r.engine.Context(ctx).SQL(kpEdgeQuery, kpEdgeArgs...).Find(&kpEdges); err != nil {
		return nil, err
	}
	edges = append(edges, kpEdges...)

	return &KnowledgeGraph{Nodes: nodes, Edges: edges}, nil
}

func decodeStringArray(raw string) []string {
	if raw == "" {
		return []string{}
	}
	var values []string
	_ = json.Unmarshal([]byte(raw), &values)
	if values == nil {
		return []string{}
	}
	return values
}
