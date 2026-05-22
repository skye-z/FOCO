package content

import (
	"context"
	"errors"
	"time"

	cachepkg "foco/backend/api/internal/cache"
	"github.com/google/uuid"
)

type Service struct {
	repo  Repository
	cache *cachepkg.Manager
}

var ErrUnsupportedRenameNodeType = errors.New("unsupported rename node type")

func NewService(repo Repository, caches ...*cachepkg.Manager) *Service {
	var cacheManager *cachepkg.Manager
	if len(caches) > 0 {
		cacheManager = caches[0]
	}
	return &Service{repo: repo, cache: cacheManager}
}

func (s *Service) GetExamTree(ctx context.Context) ([]ExamTreeNode, error) {
	if s.cache == nil {
		return s.repo.GetExamTree(ctx)
	}
	var result []ExamTreeNode
	err := s.cache.GetJSON(ctx, contentNamespace(), "exam-tree", 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.GetExamTree(ctx)
	})
	return result, err
}

func (s *Service) ListKnowledgePoints(ctx context.Context, examId string) ([]KnowledgePoint, error) {
	if s.cache == nil {
		return s.repo.ListKnowledgePoints(ctx, examId)
	}
	var result []KnowledgePoint
	err := s.cache.GetJSON(ctx, contentNamespace(), "knowledge-points:"+examId, 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListKnowledgePoints(ctx, examId)
	})
	return result, err
}

func (s *Service) FilterQuestions(ctx context.Context, filter ExamTreeFilter) ([]QuestionCard, error) {
	if s.cache == nil {
		return s.repo.FilterQuestions(ctx, filter)
	}
	var result []QuestionCard
	err := s.cache.GetJSON(ctx, contentNamespace(), "questions:"+filterCacheKey(filter), 2*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.FilterQuestions(ctx, filter)
	})
	return result, err
}

func (s *Service) GetVersionDetail(ctx context.Context, versionId string) (*QuestionVersionDetail, error) {
	if s.cache == nil {
		return s.repo.GetVersionDetail(ctx, versionId)
	}
	var result *QuestionVersionDetail
	err := s.cache.GetJSON(ctx, contentNamespace(), "version-detail:"+versionId, 2*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.GetVersionDetail(ctx, versionId)
	})
	return result, err
}

func (s *Service) ListQuestionVersions(ctx context.Context, questionId string) ([]QuestionVersionSummary, error) {
	if s.cache == nil {
		return s.repo.ListQuestionVersions(ctx, questionId)
	}
	var result []QuestionVersionSummary
	err := s.cache.GetJSON(ctx, contentNamespace(), "question-versions:"+questionId, 2*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListQuestionVersions(ctx, questionId)
	})
	return result, err
}

func (s *Service) UpdateVersionFull(ctx context.Context, versionId, stem, options, correctAnswer, explanation string, difficulty int, subjectId string, chapterId *string, kpIds []string) error {
	if err := s.repo.UpdateVersion(ctx, versionId, stem, options, correctAnswer, explanation, difficulty); err != nil {
		return err
	}
	detail, err := s.repo.GetVersionDetail(ctx, versionId)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateQuestion(ctx, detail.QuestionId, subjectId, chapterId); err != nil {
		return err
	}
	if err := s.repo.SetVersionKnowledgePoints(ctx, versionId, kpIds); err != nil {
		return err
	}
	s.invalidateContent(ctx)
	return nil
}

func (s *Service) CreateExam(ctx context.Context, code, name string, description *string) (*Exam, error) {
	exam := &Exam{
		Id:          uuid.New().String(),
		Code:        code,
		Name:        name,
		Status:      "active",
		Description: description,
	}
	if err := s.repo.CreateExam(ctx, exam); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return exam, nil
}

func (s *Service) CreateSubject(ctx context.Context, examId, code, name string, sortOrder int) (*Subject, error) {
	subject := &Subject{
		Id:        uuid.New().String(),
		ExamId:    examId,
		Code:      code,
		Name:      name,
		SortOrder: sortOrder,
	}
	if err := s.repo.CreateSubject(ctx, subject); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return subject, nil
}

func (s *Service) CreateChapter(ctx context.Context, subjectId, code, name string, sortOrder int) (*Chapter, error) {
	chapter := &Chapter{
		Id:        uuid.New().String(),
		SubjectId: subjectId,
		Code:      code,
		Name:      name,
		SortOrder: sortOrder,
	}
	if err := s.repo.CreateChapter(ctx, chapter); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return chapter, nil
}

func (s *Service) RenameNode(ctx context.Context, nodeType, nodeId, name string) error {
	switch nodeType {
	case "exam":
		err := s.repo.RenameExam(ctx, nodeId, name)
		s.invalidateContentIfOK(ctx, err)
		return err
	case "subject":
		err := s.repo.RenameSubject(ctx, nodeId, name)
		s.invalidateContentIfOK(ctx, err)
		return err
	case "chapter":
		err := s.repo.RenameChapter(ctx, nodeId, name)
		s.invalidateContentIfOK(ctx, err)
		return err
	default:
		return ErrUnsupportedRenameNodeType
	}
}

func (s *Service) CreateKnowledgePoint(ctx context.Context, examId, code, name string, description *string) (*KnowledgePoint, error) {
	kp := &KnowledgePoint{
		Id:          uuid.New().String(),
		ExamId:      examId,
		Code:        code,
		Name:        name,
		Description: description,
		Status:      "active",
	}
	if err := s.repo.CreateKnowledgePoint(ctx, kp); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return kp, nil
}

func (s *Service) ListKnowledgePointEdges(ctx context.Context, examId string) ([]KnowledgePointEdge, error) {
	if s.cache == nil {
		return s.repo.ListKnowledgePointEdges(ctx, examId)
	}
	var result []KnowledgePointEdge
	err := s.cache.GetJSON(ctx, contentNamespace(), "knowledge-point-edges:"+examId, 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ListKnowledgePointEdges(ctx, examId)
	})
	return result, err
}

func (s *Service) CreateKnowledgePointEdge(ctx context.Context, examId, fromId, toId, edgeType string, weight *float64) (*KnowledgePointEdge, error) {
	edge := &KnowledgePointEdge{
		Id:                   uuid.New().String(),
		ExamId:               examId,
		FromKnowledgePointId: fromId,
		ToKnowledgePointId:   toId,
		EdgeType:             edgeType,
		Weight:               weight,
	}
	if err := s.repo.CreateKnowledgePointEdge(ctx, edge); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return edge, nil
}

func (s *Service) CreateQuestion(ctx context.Context, examId, subjectId string, chapterId *string) (*Question, error) {
	q := &Question{
		Id:        uuid.New().String(),
		ExamId:    examId,
		SubjectId: subjectId,
		ChapterId: chapterId,
		Status:    "draft",
	}
	if err := s.repo.CreateQuestion(ctx, q); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return q, nil
}

type CreateVersionResult struct {
	Question        *Question
	QuestionVersion *QuestionVersion
}

func (s *Service) CreateQuestionVersion(ctx context.Context, questionId, questionType string, difficulty int, stem, options, correctAnswer, explanation string) (*QuestionVersion, error) {
	maxNo, err := s.repo.GetMaxVersionNo(ctx, questionId)
	if err != nil {
		return nil, err
	}
	v := &QuestionVersion{
		Id:            uuid.New().String(),
		QuestionId:    questionId,
		VersionNo:     maxNo + 1,
		Status:        "draft",
		QuestionType:  questionType,
		Difficulty:    difficulty,
		Stem:          stem,
		Options:       &options,
		CorrectAnswer: correctAnswer,
		Explanation:   explanation,
	}
	if err := s.repo.CreateQuestionVersion(ctx, v); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return v, nil
}

func (s *Service) SaveVersionOrCreateDraft(ctx context.Context, versionId, stem, options, correctAnswer, explanation string, difficulty int, subjectId string, chapterId *string, kpIds []string) (*QuestionVersionDetail, error) {
	detail, err := s.repo.GetVersionDetail(ctx, versionId)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}

	targetVersionId := versionId
	if detail.Status == "published" {
		maxNo, err := s.repo.GetMaxVersionNo(ctx, detail.QuestionId)
		if err != nil {
			return nil, err
		}
		clone := &QuestionVersion{
			Id:         uuid.New().String(),
			QuestionId: detail.QuestionId,
			VersionNo:  maxNo + 1,
		}
		if err := s.repo.CloneQuestionVersion(ctx, versionId, clone); err != nil {
			return nil, err
		}
		targetVersionId = clone.Id
	}

	if err := s.repo.UpdateVersion(ctx, targetVersionId, stem, options, correctAnswer, explanation, difficulty); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateQuestion(ctx, detail.QuestionId, subjectId, chapterId); err != nil {
		return nil, err
	}
	if err := s.repo.SetVersionKnowledgePoints(ctx, targetVersionId, kpIds); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return s.repo.GetVersionDetail(ctx, targetVersionId)
}

func (s *Service) RestoreVersionAsDraft(ctx context.Context, versionId string) (*QuestionVersionDetail, error) {
	detail, err := s.repo.GetVersionDetail(ctx, versionId)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}
	maxNo, err := s.repo.GetMaxVersionNo(ctx, detail.QuestionId)
	if err != nil {
		return nil, err
	}
	clone := &QuestionVersion{
		Id:         uuid.New().String(),
		QuestionId: detail.QuestionId,
		VersionNo:  maxNo + 1,
	}
	if err := s.repo.CloneQuestionVersion(ctx, versionId, clone); err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return s.repo.GetVersionDetail(ctx, clone.Id)
}

func (s *Service) PublishVersion(ctx context.Context, versionId, publishedBy string, publishNote *string) error {
	err := s.repo.PublishVersion(ctx, versionId, publishedBy, publishNote)
	s.invalidateContentIfOK(ctx, err)
	return err
}

func (s *Service) BuildKnowledgeGraph(ctx context.Context, examId string) (*KnowledgeGraphPayload, error) {
	if s.cache == nil {
		return s.repo.BuildKnowledgeGraph(ctx, examId)
	}
	var result *KnowledgeGraphPayload
	err := s.cache.GetJSON(ctx, contentNamespace(), "knowledge-graph:"+examId, 10*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.BuildKnowledgeGraph(ctx, examId)
	})
	return result, err
}

func (s *Service) ExportContentPackage(ctx context.Context) (*ContentPackagePayload, error) {
	if s.cache == nil {
		return s.repo.ExportContentPackage(ctx)
	}
	var result *ContentPackagePayload
	err := s.cache.GetJSON(ctx, contentNamespace(), "content-package-export", 2*time.Minute, &result, func(ctx context.Context) (any, error) {
		return s.repo.ExportContentPackage(ctx)
	})
	return result, err
}

func (s *Service) ImportContentPackage(ctx context.Context, raw map[string]any, payload *ContentPackagePayload) (*ContentPackageImportReport, error) {
	report := validateContentPackage(raw, payload)
	if len(report.ValidationErrors) > 0 {
		return report, nil
	}
	imported, err := s.repo.ImportContentPackage(ctx, payload)
	if err != nil {
		return nil, err
	}
	s.invalidateContent(ctx)
	return imported, nil
}

func validateContentPackage(raw map[string]any, payload *ContentPackagePayload) *ContentPackageImportReport {
	report := &ContentPackageImportReport{ValidationErrors: []string{}}
	requiredKeys := []string{
		"exams",
		"subjects",
		"chapters",
		"knowledge_points",
		"knowledge_point_edges",
		"questions",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			report.ValidationErrors = append(report.ValidationErrors, "缺少字段: "+key)
		}
	}
	if len(payload.Exams) == 0 {
		report.ValidationErrors = append(report.ValidationErrors, "exams 不能为空")
	}
	if len(payload.Questions) == 0 {
		report.ValidationErrors = append(report.ValidationErrors, "questions 不能为空")
	}
	return report
}

func (s *Service) DeleteExam(ctx context.Context, examId string) error {
	err := s.repo.DeleteExam(ctx, examId)
	s.invalidateContentIfOK(ctx, err)
	return err
}

func (s *Service) DeleteSubject(ctx context.Context, subjectId string) error {
	err := s.repo.DeleteSubject(ctx, subjectId)
	s.invalidateContentIfOK(ctx, err)
	return err
}

func (s *Service) DeleteChapter(ctx context.Context, chapterId string) error {
	err := s.repo.DeleteChapter(ctx, chapterId)
	s.invalidateContentIfOK(ctx, err)
	return err
}

func (s *Service) DeleteQuestion(ctx context.Context, questionId string) error {
	err := s.repo.DeleteQuestion(ctx, questionId)
	s.invalidateContentIfOK(ctx, err)
	return err
}

func (s *Service) CountQuestionsByExam(ctx context.Context, examId string) (int64, error) {
	return s.repo.CountQuestionsByExam(ctx, examId)
}

func (s *Service) CountQuestionsBySubject(ctx context.Context, subjectId string) (int64, error) {
	return s.repo.CountQuestionsBySubject(ctx, subjectId)
}

func (s *Service) CountQuestionsByChapter(ctx context.Context, chapterId string) (int64, error) {
	return s.repo.CountQuestionsByChapter(ctx, chapterId)
}

func (s *Service) invalidateContent(ctx context.Context) {
	if s.cache != nil {
		s.cache.Invalidate(ctx, contentNamespace(), interactiveNamespace(), accountStatsNamespace())
	}
}

func (s *Service) invalidateContentIfOK(ctx context.Context, err error) {
	if err == nil {
		s.invalidateContent(ctx)
	}
}

func contentNamespace() string      { return "content:all" }
func interactiveNamespace() string  { return "interactive:all" }
func accountStatsNamespace() string { return "account:stats" }

func filterCacheKey(filter ExamTreeFilter) string {
	return filter.ExamId + "|" + filter.SubjectId + "|" + filter.ChapterId + "|" + filter.KnowledgePoint + "|" + filter.Difficulty + "|" + filter.Status
}
