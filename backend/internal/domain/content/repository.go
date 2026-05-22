package content

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"xorm.io/xorm"
)

type Repository interface {
	GetExamTree(ctx context.Context) ([]ExamTreeNode, error)
	ListKnowledgePoints(ctx context.Context, examId string) ([]KnowledgePoint, error)
	FilterQuestions(ctx context.Context, filter ExamTreeFilter) ([]QuestionCard, error)
	GetVersionDetail(ctx context.Context, versionId string) (*QuestionVersionDetail, error)
	ListQuestionVersions(ctx context.Context, questionId string) ([]QuestionVersionSummary, error)
	UpdateVersion(ctx context.Context, versionId string, stem, options, correctAnswer, explanation string, difficulty int) error
	UpdateQuestion(ctx context.Context, questionId, subjectId string, chapterId *string) error
	SetVersionKnowledgePoints(ctx context.Context, versionId string, kpIds []string) error

	CreateExam(ctx context.Context, exam *Exam) error
	CreateSubject(ctx context.Context, subject *Subject) error
	CreateChapter(ctx context.Context, chapter *Chapter) error
	RenameExam(ctx context.Context, examId, name string) error
	RenameSubject(ctx context.Context, subjectId, name string) error
	RenameChapter(ctx context.Context, chapterId, name string) error
	CreateKnowledgePoint(ctx context.Context, kp *KnowledgePoint) error
	ListKnowledgePointEdges(ctx context.Context, examId string) ([]KnowledgePointEdge, error)
	CreateKnowledgePointEdge(ctx context.Context, edge *KnowledgePointEdge) error

	CreateQuestion(ctx context.Context, q *Question) error
	GetQuestion(ctx context.Context, questionId string) (*Question, error)
	CreateQuestionVersion(ctx context.Context, v *QuestionVersion) error
	GetMaxVersionNo(ctx context.Context, questionId string) (int, error)
	CloneQuestionVersion(ctx context.Context, sourceVersionId string, v *QuestionVersion) error
	PublishVersion(ctx context.Context, versionId, publishedBy string, publishNote *string) error
	BuildKnowledgeGraph(ctx context.Context, examId string) (*KnowledgeGraphPayload, error)
	ExportContentPackage(ctx context.Context) (*ContentPackagePayload, error)
	ImportContentPackage(ctx context.Context, payload *ContentPackagePayload) (*ContentPackageImportReport, error)

	CountQuestionsByExam(ctx context.Context, examId string) (int64, error)
	CountQuestionsBySubject(ctx context.Context, subjectId string) (int64, error)
	CountQuestionsByChapter(ctx context.Context, chapterId string) (int64, error)
	DeleteExam(ctx context.Context, examId string) error
	DeleteSubject(ctx context.Context, subjectId string) error
	DeleteChapter(ctx context.Context, chapterId string) error
	DeleteQuestion(ctx context.Context, questionId string) error
}

type XormRepository struct {
	engine *xorm.Engine
}

type knowledgeGraphQuestionRow struct {
	QuestionId  string  `xorm:"question_id"`
	VersionId   string  `xorm:"version_id"`
	StemPreview string  `xorm:"stem_preview"`
	ExamId      string  `xorm:"exam_id"`
	SubjectId   string  `xorm:"subject_id"`
	ChapterId   *string `xorm:"chapter_id"`
}

type knowledgeGraphQuestionEdgeRow struct {
	QuestionId       string `xorm:"question_id"`
	KnowledgePointId string `xorm:"knowledge_point_id"`
}

func NewRepository(engine *xorm.Engine) *XormRepository {
	return &XormRepository{engine: engine}
}

func execBulk(sess *xorm.Session, query string, args []any) (sql.Result, error) {
	all := make([]interface{}, 0, len(args)+1)
	all = append(all, query)
	all = append(all, args...)
	return sess.Exec(all...)
}

func (r *XormRepository) GetExamTree(ctx context.Context) ([]ExamTreeNode, error) {
	var exams []Exam
	err := r.engine.Context(ctx).
		Where("status = 'active'").
		OrderBy("name asc").
		Find(&exams)
	if err != nil {
		return nil, err
	}

	var allSubjects []Subject
	err = r.engine.Context(ctx).
		OrderBy("exam_id asc, sort_order asc").
		Find(&allSubjects)
	if err != nil {
		return nil, err
	}

	subjectByExam := map[string][]Subject{}
	for _, s := range allSubjects {
		subjectByExam[s.ExamId] = append(subjectByExam[s.ExamId], s)
	}

	var allChapters []Chapter
	err = r.engine.Context(ctx).
		OrderBy("subject_id asc, sort_order asc").
		Find(&allChapters)
	if err != nil {
		return nil, err
	}

	chapterBySubject := map[string][]Chapter{}
	for _, c := range allChapters {
		chapterBySubject[c.SubjectId] = append(chapterBySubject[c.SubjectId], c)
	}

	nodes := make([]ExamTreeNode, 0, len(exams))
	for _, e := range exams {
		subjects := subjectByExam[e.Id]
		subjectNodes := make([]ExamTreeNode, 0, len(subjects))
		for _, s := range subjects {
			chapters := chapterBySubject[s.Id]
			chapterNodes := make([]ExamTreeNode, 0, len(chapters))
			for _, c := range chapters {
				chapterNodes = append(chapterNodes, ExamTreeNode{
					Id:   c.Id,
					Code: c.Code,
					Name: c.Name,
					Type: "chapter",
				})
			}
			subjectNodes = append(subjectNodes, ExamTreeNode{
				Id:       s.Id,
				Code:     s.Code,
				Name:     s.Name,
				Type:     "subject",
				Children: chapterNodes,
			})
		}
		node := ExamTreeNode{
			Id:       e.Id,
			Code:     e.Code,
			Name:     e.Name,
			Type:     "exam",
			Children: subjectNodes,
		}
		if e.NextExamDate != nil {
			node.NextExamDate = e.NextExamDate.Format("2006-01-02")
			days := int(time.Until(*e.NextExamDate).Hours() / 24)
			if days < 0 {
				days = 0
			}
			node.CountdownDays = &days
		}
		if e.NextNextExamDate != nil {
			node.NextNextExamDate = e.NextNextExamDate.Format("2006-01-02")
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (r *XormRepository) ListKnowledgePoints(ctx context.Context, examId string) ([]KnowledgePoint, error) {
	var kps []KnowledgePoint
	sess := r.engine.Context(ctx).Where("status = 'active'")
	if examId != "" {
		sess = sess.Where("exam_id = ?::uuid", examId)
	}
	err := sess.OrderBy("name asc").Find(&kps)
	return kps, err
}

func (r *XormRepository) FilterQuestions(ctx context.Context, filter ExamTreeFilter) ([]QuestionCard, error) {
	sess := r.engine.Context(ctx).
		Table("questions").
		Select(`questions.id, questions.exam_id, questions.subject_id, subjects.name as subject_name,
			questions.chapter_id, chapters.name as chapter_name,
			questions.status,
			coalesce(qv.question_type, '') as question_type,
			coalesce(qv.difficulty, 0) as difficulty,
			coalesce(qv.version_no, 0) as version_no,
			coalesce(qv.id::text, '') as version_id,
			coalesce(qv.stem::text, '') as stem_preview,
			pub_qv.version_no as published_version_no,
			draft_qv.version_no as draft_version_no,
			(draft_qv.id is not null) as has_unpublished_draft`).
		Join("LEFT", "subjects", "subjects.id = questions.subject_id").
		Join("LEFT", "chapters", "chapters.id = questions.chapter_id").
		Join("LEFT", "(SELECT DISTINCT ON (question_id) id, question_id, question_type, difficulty, version_no, stem FROM question_versions ORDER BY question_id, version_no DESC) qv", "qv.question_id = questions.id").
		Join("LEFT", "question_versions pub_qv", "pub_qv.id = questions.current_published_version_id").
		Join("LEFT", "(SELECT DISTINCT ON (question_id) id, question_id, version_no FROM question_versions WHERE status = 'draft' ORDER BY question_id, version_no DESC) draft_qv", "draft_qv.question_id = questions.id")

	if filter.ExamId != "" {
		sess = sess.Where("questions.exam_id = ?::uuid", filter.ExamId)
	}
	if filter.SubjectId != "" {
		sess = sess.Where("questions.subject_id = ?::uuid", filter.SubjectId)
	}
	if filter.ChapterId != "" {
		sess = sess.Where("questions.chapter_id = ?::uuid", filter.ChapterId)
	}
	if filter.Status != "" {
		sess = sess.Where("questions.status = ?", filter.Status)
	}
	if filter.Difficulty != "" {
		sess = sess.Where("qv.difficulty = ?", filter.Difficulty)
	}
	if filter.KnowledgePoint != "" {
		sess = sess.Join("INNER", "question_version_knowledge_points qvkp", "qvkp.question_version_id = qv.id").
			Where("qvkp.knowledge_point_id = ?::uuid", filter.KnowledgePoint)
	}

	sess = sess.OrderBy("questions.created_at desc").Limit(200)

	var results []QuestionCard
	err := sess.Find(&results)
	if err != nil {
		return nil, err
	}

	for i := range results {
		results[i].StemPreview = truncateStem(results[i].StemPreview)
	}

	return results, nil
}

func truncateStem(raw string) string {
	if raw == "" {
		return ""
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return raw
	}
	if text, ok := parsed["text"].(string); ok {
		if len(text) > 120 {
			return text[:120] + "..."
		}
		return text
	}
	b, _ := json.Marshal(parsed)
	s := string(b)
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

func (r *XormRepository) GetVersionDetail(ctx context.Context, versionId string) (*QuestionVersionDetail, error) {
	var detail QuestionVersionDetail
	_, err := r.engine.Context(ctx).SQL(`
		SELECT
			qv.id as version_id, qv.question_id, q.exam_id, q.subject_id, q.chapter_id,
			qv.question_type, qv.difficulty, qv.version_no, qv.status,
			qv.stem::text as stem, coalesce(qv.options::text, '{}') as options,
			qv.correct_answer::text as correct_answer, qv.explanation::text as explanation
		FROM question_versions qv
		INNER JOIN questions q ON q.id = qv.question_id
		WHERE qv.id = ?
	`, versionId).Get(&detail)
	if err != nil {
		return nil, err
	}

	var kpRows []struct {
		KnowledgePointId string `xorm:"knowledge_point_id"`
	}
	err = r.engine.Context(ctx).
		Table("question_version_knowledge_points").
		Where("question_version_id = ?::uuid", versionId).
		Cols("knowledge_point_id").
		Find(&kpRows)
	if err != nil {
		return nil, err
	}
	kpIds := make([]string, 0, len(kpRows))
	for _, kp := range kpRows {
		kpIds = append(kpIds, kp.KnowledgePointId)
	}
	detail.KnowledgePointIds = kpIds
	return &detail, nil
}

func (r *XormRepository) ListQuestionVersions(ctx context.Context, questionId string) ([]QuestionVersionSummary, error) {
	var rows []QuestionVersionSummary
	err := r.engine.Context(ctx).SQL(`
		SELECT
			qv.id as version_id,
			qv.question_id,
			qv.version_no,
			qv.status,
			qv.published_at,
			qv.updated_at,
			qv.publish_note,
			(q.current_published_version_id = qv.id) as is_current,
			(qv.status = 'published') as is_published
		FROM question_versions qv
		INNER JOIN questions q ON q.id = qv.question_id
		WHERE qv.question_id = ?::uuid
		ORDER BY qv.version_no DESC
	`, questionId).Find(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *XormRepository) UpdateVersion(ctx context.Context, versionId string, stem, options, correctAnswer, explanation string, difficulty int) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()
	_, err := sess.Exec(
		"UPDATE question_versions SET stem = $1::jsonb, options = $2::jsonb, correct_answer = $3::jsonb, explanation = $4::jsonb, difficulty = $5, updated_at = now() WHERE id = $6::uuid",
		stem, options, correctAnswer, explanation, difficulty, versionId,
	)
	return err
}

func (r *XormRepository) UpdateQuestion(ctx context.Context, questionId, subjectId string, chapterId *string) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()
	chap := ""
	if chapterId != nil {
		chap = *chapterId
	}
	_, err := sess.Exec(
		"UPDATE questions SET subject_id = $1::uuid, chapter_id = nullif($2, '')::uuid, updated_at = now() WHERE id = $3::uuid",
		subjectId, chap, questionId,
	)
	return err
}

func (r *XormRepository) SetVersionKnowledgePoints(ctx context.Context, versionId string, kpIds []string) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()
	_, err := sess.Exec("DELETE FROM question_version_knowledge_points WHERE question_version_id = $1::uuid", versionId)
	if err != nil {
		return err
	}
	for _, kpId := range kpIds {
		_, err = sess.Exec(
			"INSERT INTO question_version_knowledge_points(question_version_id, knowledge_point_id, created_at) VALUES ($1::uuid, $2::uuid, now())",
			versionId, kpId,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *XormRepository) CreateExam(ctx context.Context, exam *Exam) error {
	_, err := r.engine.Context(ctx).Insert(exam)
	return err
}

func (r *XormRepository) CreateSubject(ctx context.Context, subject *Subject) error {
	_, err := r.engine.Context(ctx).Insert(subject)
	return err
}

func (r *XormRepository) CreateChapter(ctx context.Context, chapter *Chapter) error {
	_, err := r.engine.Context(ctx).Insert(chapter)
	return err
}

func (r *XormRepository) RenameExam(ctx context.Context, examId, name string) error {
	_, err := r.engine.Context(ctx).Exec(
		"UPDATE exams SET name = $1, updated_at = now() WHERE id = $2::uuid",
		name, examId,
	)
	return err
}

func (r *XormRepository) RenameSubject(ctx context.Context, subjectId, name string) error {
	_, err := r.engine.Context(ctx).Exec(
		"UPDATE subjects SET name = $1, updated_at = now() WHERE id = $2::uuid",
		name, subjectId,
	)
	return err
}

func (r *XormRepository) RenameChapter(ctx context.Context, chapterId, name string) error {
	_, err := r.engine.Context(ctx).Exec(
		"UPDATE chapters SET name = $1, updated_at = now() WHERE id = $2::uuid",
		name, chapterId,
	)
	return err
}

func (r *XormRepository) CreateKnowledgePoint(ctx context.Context, kp *KnowledgePoint) error {
	_, err := r.engine.Context(ctx).Insert(kp)
	return err
}

func (r *XormRepository) ListKnowledgePointEdges(ctx context.Context, examId string) ([]KnowledgePointEdge, error) {
	var edges []KnowledgePointEdge
	sess := r.engine.Context(ctx)
	if examId != "" {
		sess = sess.Where("exam_id = ?::uuid", examId)
	}
	err := sess.OrderBy("created_at asc").Find(&edges)
	return edges, err
}

func (r *XormRepository) CreateKnowledgePointEdge(ctx context.Context, edge *KnowledgePointEdge) error {
	_, err := r.engine.Context(ctx).Insert(edge)
	return err
}

func (r *XormRepository) CreateQuestion(ctx context.Context, q *Question) error {
	_, err := r.engine.Context(ctx).Insert(q)
	return err
}

func (r *XormRepository) GetQuestion(ctx context.Context, questionId string) (*Question, error) {
	var q Question
	_, err := r.engine.Context(ctx).ID(questionId).Get(&q)
	if err != nil {
		return nil, err
	}
	if q.Id == "" {
		return nil, nil
	}
	return &q, nil
}

func (r *XormRepository) CreateQuestionVersion(ctx context.Context, v *QuestionVersion) error {
	_, err := r.engine.Context(ctx).Insert(v)
	return err
}

func (r *XormRepository) CloneQuestionVersion(ctx context.Context, sourceVersionId string, v *QuestionVersion) error {
	_, err := r.engine.Context(ctx).Exec(`
		INSERT INTO question_versions(
			id, question_id, version_no, status, question_type, difficulty, stem, options,
			correct_answer, explanation, publish_note, content_hash, created_at, updated_at
		)
		SELECT
			$1::uuid, question_id, $2, 'draft', question_type, difficulty, stem, options,
			correct_answer, explanation, NULL, content_hash, now(), now()
		FROM question_versions
		WHERE id = $3::uuid
	`, v.Id, v.VersionNo, sourceVersionId)
	return err
}

func (r *XormRepository) GetMaxVersionNo(ctx context.Context, questionId string) (int, error) {
	var max struct {
		MaxNo int `xorm:"max_no"`
	}
	_, err := r.engine.Context(ctx).SQL(
		"SELECT COALESCE(MAX(version_no), 0) as max_no FROM question_versions WHERE question_id = ?::uuid",
		questionId,
	).Get(&max)
	return max.MaxNo, err
}

func (r *XormRepository) PublishVersion(ctx context.Context, versionId, publishedBy string, publishNote *string) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()

	_, err := sess.Exec(
		"UPDATE question_versions SET status = 'published', published_at = now(), published_by = nullif($1, '')::uuid, publish_note = $2, updated_at = now() WHERE id = $3::uuid AND status = 'draft'",
		publishedBy, publishNote, versionId,
	)
	if err != nil {
		return err
	}

	var questionId string
	_, err = sess.SQL("SELECT question_id FROM question_versions WHERE id = ?::uuid", versionId).Get(&questionId)
	if err != nil {
		return err
	}

	_, err = sess.Exec(
		"UPDATE questions SET current_published_version_id = $1::uuid, status = 'published', updated_at = now() WHERE id = $2::uuid",
		versionId, questionId,
	)
	return err
}

func (r *XormRepository) BuildKnowledgeGraph(ctx context.Context, examId string) (*KnowledgeGraphPayload, error) {
	exams := []Exam{}
	examSession := r.engine.Context(ctx).Where("status = 'active'")
	if examId != "" {
		examSession = examSession.Where("id = ?::uuid", examId)
	}
	if err := examSession.OrderBy("created_at asc").Find(&exams); err != nil {
		return nil, err
	}

	var subjects []Subject
	subjectSession := r.engine.Context(ctx)
	if examId != "" {
		subjectSession = subjectSession.Where("exam_id = ?::uuid", examId)
	}
	if err := subjectSession.OrderBy("exam_id asc, sort_order asc, created_at asc").Find(&subjects); err != nil {
		return nil, err
	}

	var chapters []Chapter
	chapterQuery := `
		SELECT c.*
		FROM chapters c
		INNER JOIN subjects s ON s.id = c.subject_id
	`
	chapterArgs := []any{}
	if examId != "" {
		chapterQuery += " WHERE s.exam_id = ?::uuid"
		chapterArgs = append(chapterArgs, examId)
	}
	chapterQuery += " ORDER BY c.subject_id asc, c.sort_order asc, c.created_at asc"
	if err := r.engine.Context(ctx).SQL(chapterQuery, chapterArgs...).Find(&chapters); err != nil {
		return nil, err
	}

	kps, err := r.ListKnowledgePoints(ctx, examId)
	if err != nil {
		return nil, err
	}
	edges, err := r.ListKnowledgePointEdges(ctx, examId)
	if err != nil {
		return nil, err
	}

	var questionRows []knowledgeGraphQuestionRow
	query := `
		SELECT
			q.id as question_id,
			qv.id as version_id,
			q.exam_id as exam_id,
			q.subject_id as subject_id,
			q.chapter_id as chapter_id,
			coalesce(qv.stem::text, '') as stem_preview
		FROM questions q
		LEFT JOIN question_versions qv ON qv.id = q.current_published_version_id
	`
	args := []any{}
	if examId != "" {
		query += " WHERE q.exam_id = ?::uuid"
		args = append(args, examId)
	}
	err = r.engine.Context(ctx).SQL(query, args...).Find(&questionRows)
	if err != nil {
		return nil, err
	}

	var questionEdgeRows []knowledgeGraphQuestionEdgeRow
	query = `
		SELECT q.id as question_id, qvkp.knowledge_point_id
		FROM questions q
		INNER JOIN question_versions qv ON qv.id = q.current_published_version_id
		INNER JOIN question_version_knowledge_points qvkp ON qvkp.question_version_id = qv.id
	`
	args = []any{}
	if examId != "" {
		query += " WHERE q.exam_id = ?::uuid"
		args = append(args, examId)
	}
	err = r.engine.Context(ctx).SQL(query, args...).Find(&questionEdgeRows)
	if err != nil {
		return nil, err
	}

	return assembleKnowledgeGraph(exams, subjects, chapters, kps, edges, questionRows, questionEdgeRows), nil
}

func assembleKnowledgeGraph(
	exams []Exam,
	subjects []Subject,
	chapters []Chapter,
	kps []KnowledgePoint,
	kpEdges []KnowledgePointEdge,
	questionRows []knowledgeGraphQuestionRow,
	questionEdgeRows []knowledgeGraphQuestionEdgeRow,
) *KnowledgeGraphPayload {
	payload := &KnowledgeGraphPayload{
		Nodes: []KnowledgeGraphNode{},
		Edges: []KnowledgeGraphEdge{},
	}

	for _, exam := range exams {
		payload.Nodes = append(payload.Nodes, KnowledgeGraphNode{
			Id:          "exam:" + exam.Id,
			Type:        "exam",
			RefId:       exam.Id,
			Label:       exam.Name,
			Description: exam.Code,
			Group:       exam.Code,
		})
	}

	for _, subject := range subjects {
		payload.Nodes = append(payload.Nodes, KnowledgeGraphNode{
			Id:          "subject:" + subject.Id,
			Type:        "subject",
			RefId:       subject.Id,
			Label:       subject.Name,
			Description: subject.Code,
			Group:       subject.Code,
		})
		payload.Edges = append(payload.Edges, KnowledgeGraphEdge{
			Id:     "examsubject:" + subject.ExamId + ":" + subject.Id,
			Source: "exam:" + subject.ExamId,
			Target: "subject:" + subject.Id,
			Type:   "exam_subject",
			Label:  "包含",
		})
	}

	for _, chapter := range chapters {
		payload.Nodes = append(payload.Nodes, KnowledgeGraphNode{
			Id:          "chapter:" + chapter.Id,
			Type:        "chapter",
			RefId:       chapter.Id,
			Label:       chapter.Name,
			Description: chapter.Code,
			Group:       chapter.Code,
		})
		payload.Edges = append(payload.Edges, KnowledgeGraphEdge{
			Id:     "subjectchapter:" + chapter.SubjectId + ":" + chapter.Id,
			Source: "subject:" + chapter.SubjectId,
			Target: "chapter:" + chapter.Id,
			Type:   "subject_chapter",
			Label:  "包含",
		})
	}

	for _, kp := range kps {
		payload.Nodes = append(payload.Nodes, KnowledgeGraphNode{
			Id:          "kp:" + kp.Id,
			Type:        "knowledge_point",
			RefId:       kp.Id,
			Label:       kp.Name,
			Description: ptrToString(kp.Description),
			Group:       kp.Code,
		})
	}

	for _, edge := range kpEdges {
		payload.Edges = append(payload.Edges, KnowledgeGraphEdge{
			Id:     edge.Id,
			Source: "kp:" + edge.FromKnowledgePointId,
			Target: "kp:" + edge.ToKnowledgePointId,
			Type:   edge.EdgeType,
			Label:  edge.EdgeType,
		})
	}

	for _, row := range questionRows {
		payload.Nodes = append(payload.Nodes, KnowledgeGraphNode{
			Id:          "question:" + row.QuestionId,
			Type:        "question",
			RefId:       row.QuestionId,
			Label:       truncateStem(row.StemPreview),
			Description: row.VersionId,
		})
		if row.ChapterId != nil && *row.ChapterId != "" {
			payload.Edges = append(payload.Edges, KnowledgeGraphEdge{
				Id:     "chapterquestion:" + *row.ChapterId + ":" + row.QuestionId,
				Source: "chapter:" + *row.ChapterId,
				Target: "question:" + row.QuestionId,
				Type:   "chapter_question",
				Label:  "包含",
			})
		} else {
			payload.Edges = append(payload.Edges, KnowledgeGraphEdge{
				Id:     "subjectquestion:" + row.SubjectId + ":" + row.QuestionId,
				Source: "subject:" + row.SubjectId,
				Target: "question:" + row.QuestionId,
				Type:   "subject_question",
				Label:  "包含",
			})
		}
	}

	for _, row := range questionEdgeRows {
		payload.Edges = append(payload.Edges, KnowledgeGraphEdge{
			Id:     "qkp:" + row.QuestionId + ":" + row.KnowledgePointId,
			Source: "question:" + row.QuestionId,
			Target: "kp:" + row.KnowledgePointId,
			Type:   "question_tag",
			Label:  "关联",
		})
	}

	return payload
}

func (r *XormRepository) ExportContentPackage(ctx context.Context) (*ContentPackagePayload, error) {
	payload := &ContentPackagePayload{}
	if err := r.engine.Context(ctx).OrderBy("created_at asc").Find(&payload.Exams); err != nil {
		return nil, err
	}
	if err := r.engine.Context(ctx).OrderBy("created_at asc").Find(&payload.Subjects); err != nil {
		return nil, err
	}
	if err := r.engine.Context(ctx).OrderBy("created_at asc").Find(&payload.Chapters); err != nil {
		return nil, err
	}
	if err := r.engine.Context(ctx).OrderBy("created_at asc").Find(&payload.KnowledgePoints); err != nil {
		return nil, err
	}
	if err := r.engine.Context(ctx).OrderBy("created_at asc").Find(&payload.KnowledgePointEdges); err != nil {
		return nil, err
	}
	questions, err := r.exportPublishedContentPackageQuestions(ctx)
	if err != nil {
		return nil, err
	}
	payload.Questions = questions
	iUnits, err := r.exportInteractiveUnits(ctx)
	if err != nil {
		return nil, err
	}
	payload.InteractiveUnits = iUnits
	return payload, nil
}

const batchSize = 200

func (r *XormRepository) ImportContentPackage(ctx context.Context, payload *ContentPackagePayload) (*ContentPackageImportReport, error) {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()
	report := &ContentPackageImportReport{}
	if err := sess.Begin(); err != nil {
		return nil, err
	}

	if len(payload.Exams) > 0 {
		if err := batchUpsertExams(sess, payload.Exams); err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.ExamsImported = len(payload.Exams)
	}
	if len(payload.Subjects) > 0 {
		if err := batchUpsertSubjects(sess, payload.Subjects); err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.SubjectsImported = len(payload.Subjects)
	}
	if len(payload.Chapters) > 0 {
		if err := batchUpsertChapters(sess, payload.Chapters); err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.ChaptersImported = len(payload.Chapters)
	}
	if len(payload.KnowledgePoints) > 0 {
		if err := batchUpsertKnowledgePoints(sess, payload.KnowledgePoints); err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.KnowledgePointsImported = len(payload.KnowledgePoints)
	}
	if len(payload.KnowledgePointEdges) > 0 {
		if err := batchUpsertKnowledgePointEdges(sess, payload.KnowledgePointEdges); err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.KnowledgePointEdgesImported = len(payload.KnowledgePointEdges)
	}
	if len(payload.Questions) > 0 {
		n, err := batchUpsertQuestions(sess, payload.Questions)
		if err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.QuestionsImported = n
		report.QuestionVersionsImported = n
		report.QuestionVersionKnowledgePointsImported = countQuestionVersionKP(payload.Questions)
	}
	if len(payload.InteractiveUnits) > 0 {
		units, steps, err := batchUpsertInteractiveUnits(sess, payload.InteractiveUnits)
		if err != nil {
			_ = sess.Rollback()
			return nil, err
		}
		report.InteractiveUnitsImported = units
		report.InteractiveUnitStepsImported = steps
	}

	if err := sess.Commit(); err != nil {
		return nil, err
	}
	return report, nil
}

func countQuestionVersionKP(questions []ContentPackageQuestion) int {
	n := 0
	for _, q := range questions {
		n += len(q.KnowledgePointIds)
	}
	return n
}

func batchUpsertExams(sess *xorm.Session, items []Exam) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		var sb strings.Builder
		sb.WriteString(`INSERT INTO exams (id, code, name, status, description, next_exam_date, next_next_exam_date, created_at, updated_at) VALUES `)
		args := make([]any, 0, len(batch)*9)
		for j, item := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("($%d::uuid, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				j*9+1, j*9+2, j*9+3, j*9+4, j*9+5, j*9+6, j*9+7, j*9+8, j*9+9))
			args = append(args, item.Id, item.Code, item.Name, item.Status, item.Description, item.NextExamDate, item.NextNextExamDate, item.CreatedAt, item.UpdatedAt)
		}
		sb.WriteString(` ON CONFLICT (id) DO UPDATE SET code=EXCLUDED.code, name=EXCLUDED.name, status=EXCLUDED.status, description=EXCLUDED.description, next_exam_date=EXCLUDED.next_exam_date, next_next_exam_date=EXCLUDED.next_next_exam_date, updated_at=EXCLUDED.updated_at`)
		if _, err := execBulk(sess, sb.String(), args); err != nil {
			return err
		}
	}
	return nil
}

func batchUpsertSubjects(sess *xorm.Session, items []Subject) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		var sb strings.Builder
		sb.WriteString(`INSERT INTO subjects (id, exam_id, code, name, sort_order, created_at, updated_at) VALUES `)
		args := make([]any, 0, len(batch)*7)
		for j, item := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d, $%d, $%d, $%d, $%d)",
				j*7+1, j*7+2, j*7+3, j*7+4, j*7+5, j*7+6, j*7+7))
			args = append(args, item.Id, item.ExamId, item.Code, item.Name, item.SortOrder, item.CreatedAt, item.UpdatedAt)
		}
		sb.WriteString(` ON CONFLICT (id) DO UPDATE SET exam_id=EXCLUDED.exam_id, code=EXCLUDED.code, name=EXCLUDED.name, sort_order=EXCLUDED.sort_order, updated_at=EXCLUDED.updated_at`)
		if _, err := execBulk(sess, sb.String(), args); err != nil {
			return err
		}
	}
	return nil
}

func batchUpsertChapters(sess *xorm.Session, items []Chapter) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		var sb strings.Builder
		sb.WriteString(`INSERT INTO chapters (id, subject_id, code, name, sort_order, created_at, updated_at) VALUES `)
		args := make([]any, 0, len(batch)*7)
		for j, item := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d, $%d, $%d, $%d, $%d)",
				j*7+1, j*7+2, j*7+3, j*7+4, j*7+5, j*7+6, j*7+7))
			args = append(args, item.Id, item.SubjectId, item.Code, item.Name, item.SortOrder, item.CreatedAt, item.UpdatedAt)
		}
		sb.WriteString(` ON CONFLICT (id) DO UPDATE SET subject_id=EXCLUDED.subject_id, code=EXCLUDED.code, name=EXCLUDED.name, sort_order=EXCLUDED.sort_order, updated_at=EXCLUDED.updated_at`)
		if _, err := execBulk(sess, sb.String(), args); err != nil {
			return err
		}
	}
	return nil
}

func batchUpsertKnowledgePoints(sess *xorm.Session, items []KnowledgePoint) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		var sb strings.Builder
		sb.WriteString(`INSERT INTO knowledge_points (id, exam_id, code, name, description, status, created_at, updated_at) VALUES `)
		args := make([]any, 0, len(batch)*8)
		for j, item := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d, $%d, $%d, $%d, $%d, $%d)",
				j*8+1, j*8+2, j*8+3, j*8+4, j*8+5, j*8+6, j*8+7, j*8+8))
			args = append(args, item.Id, item.ExamId, item.Code, item.Name, item.Description, item.Status, item.CreatedAt, item.UpdatedAt)
		}
		sb.WriteString(` ON CONFLICT (id) DO UPDATE SET exam_id=EXCLUDED.exam_id, code=EXCLUDED.code, name=EXCLUDED.name, description=EXCLUDED.description, status=EXCLUDED.status, updated_at=EXCLUDED.updated_at`)
		if _, err := execBulk(sess, sb.String(), args); err != nil {
			return err
		}
	}
	return nil
}

func batchUpsertKnowledgePointEdges(sess *xorm.Session, items []KnowledgePointEdge) error {
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		var sb strings.Builder
		sb.WriteString(`INSERT INTO knowledge_point_edges (id, exam_id, from_knowledge_point_id, to_knowledge_point_id, edge_type, weight, created_at) VALUES `)
		args := make([]any, 0, len(batch)*7)
		for j, item := range batch {
			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d::uuid, $%d::uuid, $%d, $%d, $%d)",
				j*7+1, j*7+2, j*7+3, j*7+4, j*7+5, j*7+6, j*7+7))
			args = append(args, item.Id, item.ExamId, item.FromKnowledgePointId, item.ToKnowledgePointId, item.EdgeType, item.Weight, item.CreatedAt)
		}
		sb.WriteString(` ON CONFLICT (id) DO UPDATE SET exam_id=EXCLUDED.exam_id, from_knowledge_point_id=EXCLUDED.from_knowledge_point_id, to_knowledge_point_id=EXCLUDED.to_knowledge_point_id, edge_type=EXCLUDED.edge_type, weight=EXCLUDED.weight`)
		if _, err := execBulk(sess, sb.String(), args); err != nil {
			return err
		}
	}
	return nil
}

func batchUpsertQuestions(sess *xorm.Session, items []ContentPackageQuestion) (int, error) {
	total := len(items)
	for _, item := range items {
		versionID := item.CurrentPublishedVersionId
		if versionID == nil || *versionID == "" {
			return 0, xorm.ErrParamsType
		}
	}

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := items[i:end]

		var qsb strings.Builder
		qsb.WriteString(`INSERT INTO questions (id, exam_id, subject_id, chapter_id, status, created_at, updated_at) VALUES `)
		qargs := make([]any, 0, len(batch)*7)
		for j, item := range batch {
			if j > 0 {
				qsb.WriteString(", ")
			}
			qsb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d::uuid, $%d::uuid, $%d, $%d, $%d)",
				j*7+1, j*7+2, j*7+3, j*7+4, j*7+5, j*7+6, j*7+7))
			qargs = append(qargs, item.Id, item.ExamId, item.SubjectId, item.ChapterId, item.Status, item.CreatedAt, item.UpdatedAt)
		}
		qsb.WriteString(` ON CONFLICT (id) DO UPDATE SET exam_id=EXCLUDED.exam_id, subject_id=EXCLUDED.subject_id, chapter_id=EXCLUDED.chapter_id, status=EXCLUDED.status, updated_at=EXCLUDED.updated_at`)
		if _, err := execBulk(sess, qsb.String(), qargs); err != nil {
			return 0, err
		}

		var vsb strings.Builder
		vsb.WriteString(`INSERT INTO question_versions (id, question_id, version_no, status, question_type, difficulty, stem, options, correct_answer, explanation, published_at, published_by, publish_note, created_at, updated_at) VALUES `)
		vargs := make([]any, 0, len(batch)*14)
		for j, item := range batch {
			if j > 0 {
				vsb.WriteString(", ")
			}
			vid := *item.CurrentPublishedVersionId
			base := j * 14
			vsb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d, 'published', $%d, $%d, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d, nullif($%d,'')::uuid, $%d, $%d, $%d)",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12, base+13, base+14))
			vargs = append(vargs, vid, item.Id, item.VersionNo, item.QuestionType, item.Difficulty,
				item.Stem, nullableJSONText(item.Options), item.CorrectAnswer, item.Explanation,
				item.PublishedAt, ptrToString(item.PublishedBy), item.PublishNote, item.CreatedAt, item.UpdatedAt)
		}
		vsb.WriteString(` ON CONFLICT (id) DO UPDATE SET question_id=EXCLUDED.question_id, version_no=EXCLUDED.version_no, status='published', question_type=EXCLUDED.question_type, difficulty=EXCLUDED.difficulty, stem=EXCLUDED.stem, options=EXCLUDED.options, correct_answer=EXCLUDED.correct_answer, explanation=EXCLUDED.explanation, published_at=EXCLUDED.published_at, published_by=EXCLUDED.published_by, publish_note=EXCLUDED.publish_note, updated_at=EXCLUDED.updated_at`)
		if _, err := execBulk(sess, vsb.String(), vargs); err != nil {
			return 0, err
		}

		var kpIDs []string
		var kpVersionIDs []string
		var kpCreatedAt []time.Time
		for _, item := range batch {
			vid := *item.CurrentPublishedVersionId
			_, _ = sess.Exec("DELETE FROM question_version_knowledge_points WHERE question_version_id = $1::uuid", vid)
			createdAt := item.UpdatedAt
			if createdAt.IsZero() {
				createdAt = item.CreatedAt
			}
			if createdAt.IsZero() {
				createdAt = time.Now()
			}
			for _, kpid := range item.KnowledgePointIds {
				kpVersionIDs = append(kpVersionIDs, vid)
				kpIDs = append(kpIDs, kpid)
				kpCreatedAt = append(kpCreatedAt, createdAt)
			}
		}
		if len(kpIDs) > 0 {
			for k := 0; k < len(kpIDs); k += batchSize {
				kend := k + batchSize
				if kend > len(kpIDs) {
					kend = len(kpIDs)
				}
				var ksb strings.Builder
				ksb.WriteString(`INSERT INTO question_version_knowledge_points (question_version_id, knowledge_point_id, created_at) VALUES `)
				kargs := make([]any, 0, (kend-k)*3)
				for m := k; m < kend; m++ {
					if m > k {
						ksb.WriteString(", ")
					}
					idx := m - k
					ksb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d)", idx*3+1, idx*3+2, idx*3+3))
					kargs = append(kargs, kpVersionIDs[m], kpIDs[m], kpCreatedAt[m])
				}
				ksb.WriteString(` ON CONFLICT DO NOTHING`)
				if _, err := execBulk(sess, ksb.String(), kargs); err != nil {
					return 0, err
				}
			}
		}

		sql, uargs := buildQuestionPublishSyncUpdateSQL(batch)
		if _, err := execBulk(sess, sql, uargs); err != nil {
			return 0, err
		}
	}
	return total, nil
}

func buildQuestionPublishSyncUpdateSQL(batch []ContentPackageQuestion) (string, []any) {
	var usb strings.Builder
	usb.WriteString(`UPDATE questions SET current_published_version_id = v.id, status = v.status, updated_at = v.updated_at FROM (VALUES `)
	uargs := make([]any, 0, len(batch)*4)
	for j, item := range batch {
		if j > 0 {
			usb.WriteString(", ")
		}
		usb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d, $%d::timestamptz)", j*4+1, j*4+2, j*4+3, j*4+4))
		uargs = append(uargs, *item.CurrentPublishedVersionId, item.Id, item.Status, item.UpdatedAt)
	}
	usb.WriteString(`) AS v(id, qid, status, updated_at) WHERE questions.id = v.qid::uuid`)
	return usb.String(), uargs
}

func batchUpsertInteractiveUnits(sess *xorm.Session, items []ContentPackageInteractiveUnit) (int, int, error) {
	totalUnits := 0
	totalSteps := 0

	for _, iu := range items {
		targetUnitID := iu.ID
		var existingUnit struct {
			ID string `xorm:"id"`
		}
		hasExistingUnit, err := sess.SQL(`
			SELECT id::text AS id
			FROM interactive_units
			WHERE exam_id = $1::uuid
			  AND coalesce(subject_id::text, '') = $2
			  AND title = $3
			ORDER BY updated_at DESC
			LIMIT 1
		`, iu.ExamID, iu.SubjectID, iu.Title).Get(&existingUnit)
		if err != nil {
			return 0, 0, err
		}
		if hasExistingUnit && existingUnit.ID != "" {
			targetUnitID = existingUnit.ID
		}

		_, err = sess.Exec(`INSERT INTO interactive_units (id, exam_id, subject_id, title, status, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, nullif($3,'')::uuid, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET exam_id=$2, subject_id=nullif($3,'')::uuid, title=$4, status=$5, updated_at=$7`,
			targetUnitID, iu.ExamID, iu.SubjectID, iu.Title, iu.Status, iu.CreatedAt, iu.UpdatedAt)
		if err != nil {
			return 0, 0, err
		}
		totalUnits++

		versionMetadata := "{}"
		if iu.Title != "" {
			b, _ := json.Marshal(iu.Title)
			versionMetadata = `{"title":` + string(b) + `}`
		}
		_, _ = sess.Exec(`INSERT INTO interactive_unit_versions (id, interactive_unit_id, version_no, status, metadata, published_at, created_at, updated_at)
			VALUES ($1::uuid, $2::uuid, $3, $4, $5::jsonb, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET interactive_unit_id=$2::uuid, status=$4, metadata=$5::jsonb, published_at=$6, updated_at=$8`,
			iu.VersionID, targetUnitID, iu.VersionNo, iu.Status, versionMetadata, iu.PublishedAt, iu.CreatedAt, iu.UpdatedAt)

		if iu.Status == "published" {
			_, _ = sess.Exec(`UPDATE interactive_units SET current_published_version_id=$1::uuid WHERE id=$2::uuid`, iu.VersionID, targetUnitID)
		}

		_, _ = sess.Exec(`DELETE FROM interactive_unit_version_steps WHERE unit_version_id=$1::uuid`, iu.VersionID)

		if len(iu.Steps) > 0 {
			var ssb strings.Builder
			ssb.WriteString(`INSERT INTO interactive_unit_version_steps (id, unit_version_id, step_no, widget_type, content, initial_state, allowed_actions, evaluation_config, feedback_map, hint_policy, knowledge_point_ids, knowledge_point_tags, created_at) VALUES `)
			sargs := make([]any, 0, len(iu.Steps)*12)
			for k, step := range iu.Steps {
				if k > 0 {
					ssb.WriteString(", ")
				}
				base := k * 12
				ssb.WriteString(fmt.Sprintf("($%d::uuid, $%d::uuid, $%d, $%d, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d::jsonb, $%d::jsonb, now())",
					base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10, base+11, base+12))
				sargs = append(sargs, step.ID, iu.VersionID, step.StepNo, step.WidgetType, step.Content, step.InitialState, step.AllowedActions, step.EvaluationConfig, step.FeedbackMap, step.HintPolicy, step.KnowledgePointIDs, step.KnowledgePointTags)
			}
			ssb.WriteString(` ON CONFLICT (id) DO UPDATE SET step_no=EXCLUDED.step_no, widget_type=EXCLUDED.widget_type, content=EXCLUDED.content, initial_state=EXCLUDED.initial_state, allowed_actions=EXCLUDED.allowed_actions, evaluation_config=EXCLUDED.evaluation_config, feedback_map=EXCLUDED.feedback_map, hint_policy=EXCLUDED.hint_policy, knowledge_point_ids=EXCLUDED.knowledge_point_ids, knowledge_point_tags=EXCLUDED.knowledge_point_tags`)
			if _, err := execBulk(sess, ssb.String(), sargs); err != nil {
				return 0, 0, err
			}
			totalSteps += len(iu.Steps)
		}
	}
	return totalUnits, totalSteps, nil
}

func (r *XormRepository) exportPublishedContentPackageQuestions(ctx context.Context) ([]ContentPackageQuestion, error) {
	type row struct {
		QuestionID                string     `xorm:"question_id"`
		ExamID                    string     `xorm:"exam_id"`
		SubjectID                 string     `xorm:"subject_id"`
		ChapterID                 *string    `xorm:"chapter_id"`
		QuestionStatus            string     `xorm:"question_status"`
		CurrentPublishedVersionID *string    `xorm:"current_published_version_id"`
		VersionNo                 int        `xorm:"version_no"`
		QuestionType              string     `xorm:"question_type"`
		Difficulty                int        `xorm:"difficulty"`
		Stem                      string     `xorm:"stem"`
		Options                   *string    `xorm:"options"`
		CorrectAnswer             string     `xorm:"correct_answer"`
		Explanation               string     `xorm:"explanation"`
		PublishedAt               *time.Time `xorm:"published_at"`
		PublishedBy               *string    `xorm:"published_by"`
		PublishNote               *string    `xorm:"publish_note"`
		QuestionCreatedAt         time.Time  `xorm:"question_created_at"`
		QuestionUpdatedAt         time.Time  `xorm:"question_updated_at"`
	}

	var rows []row
	err := r.engine.Context(ctx).SQL(`
		SELECT
			q.id AS question_id,
			q.exam_id AS exam_id,
			q.subject_id AS subject_id,
			q.chapter_id AS chapter_id,
			q.status AS question_status,
			q.current_published_version_id AS current_published_version_id,
			qv.version_no AS version_no,
			qv.question_type AS question_type,
			qv.difficulty AS difficulty,
			qv.stem AS stem,
			qv.options AS options,
			qv.correct_answer AS correct_answer,
			qv.explanation AS explanation,
			qv.published_at AS published_at,
			qv.published_by AS published_by,
			qv.publish_note AS publish_note,
			q.created_at AS question_created_at,
			q.updated_at AS question_updated_at
		FROM questions q
		INNER JOIN question_versions qv ON qv.id = q.current_published_version_id
		ORDER BY q.created_at ASC
	`).Find(&rows)
	if err != nil {
		return nil, err
	}

	questionIDs := make([]string, 0, len(rows))
	for _, item := range rows {
		questionIDs = append(questionIDs, item.QuestionID)
	}

	knowledgePointIDsByQuestion, err := r.loadPublishedKnowledgePointIDsByQuestion(ctx, questionIDs)
	if err != nil {
		return nil, err
	}

	result := make([]ContentPackageQuestion, 0, len(rows))
	for _, item := range rows {
		result = append(result, ContentPackageQuestion{
			Id:                        item.QuestionID,
			ExamId:                    item.ExamID,
			SubjectId:                 item.SubjectID,
			ChapterId:                 item.ChapterID,
			Status:                    item.QuestionStatus,
			CurrentPublishedVersionId: item.CurrentPublishedVersionID,
			VersionNo:                 item.VersionNo,
			QuestionType:              item.QuestionType,
			Difficulty:                item.Difficulty,
			Stem:                      item.Stem,
			Options:                   item.Options,
			CorrectAnswer:             item.CorrectAnswer,
			Explanation:               item.Explanation,
			PublishedAt:               item.PublishedAt,
			PublishedBy:               item.PublishedBy,
			PublishNote:               item.PublishNote,
			KnowledgePointIds:         knowledgePointIDsByQuestion[item.QuestionID],
			CreatedAt:                 item.QuestionCreatedAt,
			UpdatedAt:                 item.QuestionUpdatedAt,
		})
	}
	return result, nil
}

func (r *XormRepository) loadPublishedKnowledgePointIDsByQuestion(ctx context.Context, questionIDs []string) (map[string][]string, error) {
	result := map[string][]string{}
	if len(questionIDs) == 0 {
		return result, nil
	}

	type row struct {
		QuestionID       string `xorm:"question_id"`
		KnowledgePointID string `xorm:"knowledge_point_id"`
	}
	var rows []row
	err := r.engine.Context(ctx).SQL(`
		SELECT q.id AS question_id, qvkp.knowledge_point_id AS knowledge_point_id
		FROM questions q
		INNER JOIN question_version_knowledge_points qvkp ON qvkp.question_version_id = q.current_published_version_id
		WHERE q.id IN (`+placeholders(len(questionIDs))+`)
		ORDER BY q.id ASC
	`, toInterfaceSlice(questionIDs)...).Find(&rows)
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.QuestionID] = append(result[item.QuestionID], item.KnowledgePointID)
	}
	return result, nil
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	parts := make([]string, 0, count)
	for i := 0; i < count; i++ {
		parts = append(parts, "?")
	}
	return strings.Join(parts, ",")
}

func toInterfaceSlice(items []string) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func (r *XormRepository) CountQuestionsByExam(ctx context.Context, examId string) (int64, error) {
	return r.engine.Context(ctx).Table("questions").Where("exam_id = ?::uuid", examId).Count()
}

func (r *XormRepository) CountQuestionsBySubject(ctx context.Context, subjectId string) (int64, error) {
	return r.engine.Context(ctx).Table("questions").Where("subject_id = ?::uuid", subjectId).Count()
}

func (r *XormRepository) CountQuestionsByChapter(ctx context.Context, chapterId string) (int64, error) {
	return r.engine.Context(ctx).Table("questions").Where("chapter_id = ?::uuid", chapterId).Count()
}

func (r *XormRepository) DeleteExam(ctx context.Context, examId string) error {
	_, err := r.engine.Context(ctx).Exec("UPDATE exams SET status = 'archived', updated_at = now() WHERE id = ?::uuid", examId)
	return err
}

func (r *XormRepository) DeleteSubject(ctx context.Context, subjectId string) error {
	_, err := r.engine.Context(ctx).Exec("DELETE FROM subjects WHERE id = ?::uuid", subjectId)
	return err
}

func (r *XormRepository) DeleteChapter(ctx context.Context, chapterId string) error {
	_, err := r.engine.Context(ctx).Exec("DELETE FROM chapters WHERE id = ?::uuid", chapterId)
	return err
}

func (r *XormRepository) DeleteQuestion(ctx context.Context, questionId string) error {
	sess := r.engine.NewSession().Context(ctx)
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	_, err := sess.Exec("DELETE FROM question_version_knowledge_points WHERE question_version_id IN (SELECT id FROM question_versions WHERE question_id = ?::uuid)", questionId)
	if err != nil {
		_ = sess.Rollback()
		return err
	}
	_, err = sess.Exec("DELETE FROM question_versions WHERE question_id = ?::uuid", questionId)
	if err != nil {
		_ = sess.Rollback()
		return err
	}
	_, err = sess.Exec("DELETE FROM questions WHERE id = ?::uuid", questionId)
	if err != nil {
		_ = sess.Rollback()
		return err
	}
	return sess.Commit()
}

func ptrToEmpty(s *string) string {
	if s == nil {
		return "{}"
	}
	return *s
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullableJSONText(s *string) any {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	return *s
}

func upsertByID[T any](sess *xorm.Session, probe any, id string, value T) error {
	has, err := sess.ID(id).Get(probe)
	if err != nil {
		return err
	}
	if has {
		_, err = sess.ID(id).AllCols().Update(&value)
		return err
	}
	_, err = sess.Insert(&value)
	return err
}

func (r *XormRepository) exportInteractiveUnits(ctx context.Context) ([]ContentPackageInteractiveUnit, error) {
	type unitRow struct {
		ID            string     `xorm:"id"`
		ExamID        string     `xorm:"exam_id"`
		SubjectID     string     `xorm:"subject_id"`
		Title         string     `xorm:"title"`
		Status        string     `xorm:"status"`
		VersionID     string     `xorm:"version_id"`
		VersionNo     int        `xorm:"version_no"`
		VersionStatus string     `xorm:"version_status"`
		PublishedAt   *time.Time `xorm:"published_at"`
		CreatedAt     time.Time  `xorm:"created_at"`
		UpdatedAt     time.Time  `xorm:"updated_at"`
	}

	var unitRows []unitRow
	err := r.engine.Context(ctx).SQL(`
		SELECT iu.id::text AS id,
		       iu.exam_id::text AS exam_id,
		       coalesce(iu.subject_id::text, '') AS subject_id,
		       iu.title,
		       iu.status,
		       pub_v.id::text AS version_id,
		       pub_v.version_no,
		       pub_v.status AS version_status,
		       pub_v.published_at,
		       iu.created_at,
		       iu.updated_at
		FROM interactive_units iu
		LEFT JOIN interactive_unit_versions pub_v ON pub_v.id = iu.current_published_version_id
		ORDER BY iu.created_at ASC
	`).Find(&unitRows)
	if err != nil {
		return nil, err
	}

	result := make([]ContentPackageInteractiveUnit, 0, len(unitRows))
	for _, ur := range unitRows {
		if ur.VersionID == "" {
			continue
		}

		type stepRow struct {
			ID                 string `xorm:"id"`
			StepNo             int    `xorm:"step_no"`
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
		var stepRows []stepRow
		err := r.engine.Context(ctx).SQL(`
			SELECT id::text, step_no, widget_type, content::text, initial_state::text, allowed_actions::text,
			       evaluation_config::text, feedback_map::text, hint_policy::text,
			       knowledge_point_ids::text, knowledge_point_tags::text
			FROM interactive_unit_version_steps
			WHERE unit_version_id = ?::uuid
			ORDER BY step_no ASC
		`, ur.VersionID).Find(&stepRows)
		if err != nil {
			return nil, err
		}

		steps := make([]ContentPackageInteractiveStep, 0, len(stepRows))
		for _, sr := range stepRows {
			steps = append(steps, ContentPackageInteractiveStep{
				ID: sr.ID, StepNo: sr.StepNo, WidgetType: sr.WidgetType,
				Content: sr.Content, InitialState: sr.InitialState, AllowedActions: sr.AllowedActions,
				EvaluationConfig: sr.EvaluationConfig, FeedbackMap: sr.FeedbackMap, HintPolicy: sr.HintPolicy,
				KnowledgePointIDs: sr.KnowledgePointIDs, KnowledgePointTags: sr.KnowledgePointTags,
			})
		}

		result = append(result, ContentPackageInteractiveUnit{
			ID: ur.ID, ExamID: ur.ExamID, SubjectID: ur.SubjectID, Title: ur.Title,
			Status: ur.Status, VersionNo: ur.VersionNo, VersionID: ur.VersionID,
			PublishedAt: ur.PublishedAt, Steps: steps,
			CreatedAt: ur.CreatedAt, UpdatedAt: ur.UpdatedAt,
		})
	}
	return result, nil
}
