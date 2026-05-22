package content

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestValidateContentPackageAllowsPublishedQuestionSnapshotsWithoutVersionArrays(t *testing.T) {
	t.Helper()

	raw := map[string]any{
		"exams":                 []map[string]any{{"id": "exam-1"}},
		"subjects":              []map[string]any{},
		"chapters":              []map[string]any{},
		"knowledge_points":      []map[string]any{},
		"knowledge_point_edges": []map[string]any{},
		"questions":             []map[string]any{{"id": "q-1"}},
	}
	payload := &ContentPackagePayload{
		Exams:     []Exam{{Id: "exam-1"}},
		Questions: []ContentPackageQuestion{{Id: "q-1"}},
	}

	report := validateContentPackage(raw, payload)
	if len(report.ValidationErrors) != 0 {
		t.Fatalf("expected simplified content package to pass validation, got %v", report.ValidationErrors)
	}
}

func TestContentPackagePayloadParsesBundledCFAFixture(t *testing.T) {
	t.Helper()

	data, err := os.ReadFile("../../../../cfa.json")
	if err != nil {
		t.Fatalf("read cfa fixture: %v", err)
	}
	var payload ContentPackagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal bundled CFA content package: %v", err)
	}
	if len(payload.Questions) == 0 {
		t.Fatalf("expected bundled CFA content package to include questions")
	}
}

func TestValidateContentPackageRequiresQuestions(t *testing.T) {
	t.Helper()

	raw := map[string]any{
		"exams":                 []map[string]any{{"id": "exam-1"}},
		"subjects":              []map[string]any{},
		"chapters":              []map[string]any{},
		"knowledge_points":      []map[string]any{},
		"knowledge_point_edges": []map[string]any{},
		"questions":             []map[string]any{},
	}
	payload := &ContentPackagePayload{
		Exams:     []Exam{{Id: "exam-1"}},
		Questions: []ContentPackageQuestion{},
	}

	report := validateContentPackage(raw, payload)
	if len(report.ValidationErrors) == 0 {
		t.Fatalf("expected validation errors for missing questions")
	}
}

func TestServiceRenameNodeUpdatesSupportedNodeTypes(t *testing.T) {
	t.Helper()

	repo := &renameNodeRepoStub{}
	svc := NewService(repo)
	ctx := context.Background()

	if err := svc.RenameNode(ctx, "exam", "exam-1", "New Exam"); err != nil {
		t.Fatalf("expected exam rename to succeed, got %v", err)
	}
	if repo.renameType != "exam" || repo.renameID != "exam-1" || repo.renameName != "New Exam" {
		t.Fatalf("expected exam rename call to be recorded, got %+v", repo)
	}

	if err := svc.RenameNode(ctx, "subject", "subject-1", "New Subject"); err != nil {
		t.Fatalf("expected subject rename to succeed, got %v", err)
	}
	if repo.renameType != "subject" || repo.renameID != "subject-1" || repo.renameName != "New Subject" {
		t.Fatalf("expected subject rename call to be recorded, got %+v", repo)
	}

	if err := svc.RenameNode(ctx, "chapter", "chapter-1", "New Chapter"); err != nil {
		t.Fatalf("expected chapter rename to succeed, got %v", err)
	}
	if repo.renameType != "chapter" || repo.renameID != "chapter-1" || repo.renameName != "New Chapter" {
		t.Fatalf("expected chapter rename call to be recorded, got %+v", repo)
	}
}

func TestServiceRenameNodeRejectsUnsupportedType(t *testing.T) {
	t.Helper()

	repo := &renameNodeRepoStub{}
	svc := NewService(repo)
	err := svc.RenameNode(context.Background(), "knowledge_point", "kp-1", "New KP")
	if err == nil {
		t.Fatalf("expected unsupported node type to fail")
	}
	if repo.renameType != "" {
		t.Fatalf("expected repo not to be called for unsupported type")
	}
}

type renameNodeRepoStub struct {
	renameType string
	renameID   string
	renameName string
}

func (s *renameNodeRepoStub) GetExamTree(ctx context.Context) ([]ExamTreeNode, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) ListKnowledgePoints(ctx context.Context, examId string) ([]KnowledgePoint, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) FilterQuestions(ctx context.Context, filter ExamTreeFilter) ([]QuestionCard, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) GetVersionDetail(ctx context.Context, versionId string) (*QuestionVersionDetail, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) ListQuestionVersions(ctx context.Context, questionId string) ([]QuestionVersionSummary, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) UpdateVersion(ctx context.Context, versionId string, stem, options, correctAnswer, explanation string, difficulty int) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) UpdateQuestion(ctx context.Context, questionId, subjectId string, chapterId *string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) SetVersionKnowledgePoints(ctx context.Context, versionId string, kpIds []string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateExam(ctx context.Context, exam *Exam) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateSubject(ctx context.Context, subject *Subject) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateChapter(ctx context.Context, chapter *Chapter) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateKnowledgePoint(ctx context.Context, kp *KnowledgePoint) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) ListKnowledgePointEdges(ctx context.Context, examId string) ([]KnowledgePointEdge, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateKnowledgePointEdge(ctx context.Context, edge *KnowledgePointEdge) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateQuestion(ctx context.Context, q *Question) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) GetQuestion(ctx context.Context, questionId string) (*Question, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CreateQuestionVersion(ctx context.Context, v *QuestionVersion) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) GetMaxVersionNo(ctx context.Context, questionId string) (int, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CloneQuestionVersion(ctx context.Context, sourceVersionId string, v *QuestionVersion) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) PublishVersion(ctx context.Context, versionId, publishedBy string, publishNote *string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) BuildKnowledgeGraph(ctx context.Context, examId string) (*KnowledgeGraphPayload, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) ExportContentPackage(ctx context.Context) (*ContentPackagePayload, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) ImportContentPackage(ctx context.Context, payload *ContentPackagePayload) (*ContentPackageImportReport, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CountQuestionsByExam(ctx context.Context, examId string) (int64, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CountQuestionsBySubject(ctx context.Context, subjectId string) (int64, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) CountQuestionsByChapter(ctx context.Context, chapterId string) (int64, error) {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) DeleteExam(ctx context.Context, examId string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) DeleteSubject(ctx context.Context, subjectId string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) DeleteChapter(ctx context.Context, chapterId string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) DeleteQuestion(ctx context.Context, questionId string) error {
	panic("unexpected call")
}

func (s *renameNodeRepoStub) RenameExam(ctx context.Context, examId, name string) error {
	s.renameType = "exam"
	s.renameID = examId
	s.renameName = name
	return nil
}

func (s *renameNodeRepoStub) RenameSubject(ctx context.Context, subjectId, name string) error {
	s.renameType = "subject"
	s.renameID = subjectId
	s.renameName = name
	return nil
}

func (s *renameNodeRepoStub) RenameChapter(ctx context.Context, chapterId, name string) error {
	s.renameType = "chapter"
	s.renameID = chapterId
	s.renameName = name
	return nil
}
