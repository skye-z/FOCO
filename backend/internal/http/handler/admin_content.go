package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"foco/backend/api/internal/domain/content"
)

type ContentService interface {
	GetExamTree(ctx context.Context) ([]content.ExamTreeNode, error)
	ListKnowledgePoints(ctx context.Context, examId string) ([]content.KnowledgePoint, error)
	FilterQuestions(ctx context.Context, filter content.ExamTreeFilter) ([]content.QuestionCard, error)
	GetVersionDetail(ctx context.Context, versionId string) (*content.QuestionVersionDetail, error)
	ListQuestionVersions(ctx context.Context, questionId string) ([]content.QuestionVersionSummary, error)
	SaveVersionOrCreateDraft(ctx context.Context, versionId, stem, options, correctAnswer, explanation string, difficulty int, subjectId string, chapterId *string, kpIds []string) (*content.QuestionVersionDetail, error)
	RestoreVersionAsDraft(ctx context.Context, versionId string) (*content.QuestionVersionDetail, error)

	CreateExam(ctx context.Context, code, name string, description *string) (*content.Exam, error)
	CreateSubject(ctx context.Context, examId, code, name string, sortOrder int) (*content.Subject, error)
	CreateChapter(ctx context.Context, subjectId, code, name string, sortOrder int) (*content.Chapter, error)
	RenameNode(ctx context.Context, nodeType, nodeId, name string) error
	CreateKnowledgePoint(ctx context.Context, examId, code, name string, description *string) (*content.KnowledgePoint, error)
	ListKnowledgePointEdges(ctx context.Context, examId string) ([]content.KnowledgePointEdge, error)
	CreateKnowledgePointEdge(ctx context.Context, examId, fromId, toId, edgeType string, weight *float64) (*content.KnowledgePointEdge, error)

	CreateQuestion(ctx context.Context, examId, subjectId string, chapterId *string) (*content.Question, error)
	CreateQuestionVersion(ctx context.Context, questionId, questionType string, difficulty int, stem, options, correctAnswer, explanation string) (*content.QuestionVersion, error)
	PublishVersion(ctx context.Context, versionId, publishedBy string, publishNote *string) error
	BuildKnowledgeGraph(ctx context.Context, examId string) (*content.KnowledgeGraphPayload, error)
	ExportContentPackage(ctx context.Context) (*content.ContentPackagePayload, error)
	ImportContentPackage(ctx context.Context, raw map[string]any, payload *content.ContentPackagePayload) (*content.ContentPackageImportReport, error)

	DeleteExam(ctx context.Context, examId string) error
	DeleteSubject(ctx context.Context, subjectId string) error
	DeleteChapter(ctx context.Context, chapterId string) error
	DeleteQuestion(ctx context.Context, questionId string) error
	CountQuestionsByExam(ctx context.Context, examId string) (int64, error)
	CountQuestionsBySubject(ctx context.Context, subjectId string) (int64, error)
	CountQuestionsByChapter(ctx context.Context, chapterId string) (int64, error)
}

type AdminContentHandler struct {
	contentService ContentService
}

func NewAdminContentHandler(contentService ContentService) *AdminContentHandler {
	return &AdminContentHandler{contentService: contentService}
}

func (h *AdminContentHandler) ExamTree(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	tree, err := h.contentService.GetExamTree(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  tree,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *AdminContentHandler) KnowledgePoints(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	examId := r.URL.Query().Get("exam_id")
	kps, err := h.contentService.ListKnowledgePoints(r.Context(), examId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  kps,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *AdminContentHandler) ListQuestions(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	q := r.URL.Query()
	filter := content.ExamTreeFilter{
		ExamId:         q.Get("exam_id"),
		SubjectId:      q.Get("subject_id"),
		ChapterId:      q.Get("chapter_id"),
		KnowledgePoint: q.Get("knowledge_point_id"),
		Difficulty:     q.Get("difficulty"),
		Status:         q.Get("status"),
	}
	cards, err := h.contentService.FilterQuestions(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  cards,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *AdminContentHandler) GetVersionDetail(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	versionId := r.PathValue("versionId")
	detail, err := h.contentService.GetVersionDetail(r.Context(), versionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  detail,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *AdminContentHandler) ListQuestionVersions(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	items, err := h.contentService.ListQuestionVersions(r.Context(), r.PathValue("questionId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items, "meta": map[string]any{}, "error": nil})
}

func (h *AdminContentHandler) UpdateVersion(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	versionId := r.PathValue("versionId")

	var body struct {
		Stem              any      `json:"stem"`
		Options           any      `json:"options"`
		CorrectAnswer     any      `json:"correct_answer"`
		Explanation       any      `json:"explanation"`
		Difficulty        int      `json:"difficulty"`
		SubjectId         string   `json:"subject_id"`
		ChapterId         *string  `json:"chapter_id"`
		KnowledgePointIds []string `json:"knowledge_point_ids"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}

	stemStr := jsonString(body.Stem)
	optsStr := jsonString(body.Options)
	ansStr := jsonString(body.CorrectAnswer)
	expStr := jsonString(body.Explanation)

	detail, err := h.contentService.SaveVersionOrCreateDraft(
		r.Context(), versionId,
		stemStr, optsStr, ansStr, expStr,
		body.Difficulty, body.SubjectId, body.ChapterId, body.KnowledgePointIds,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  detail,
		"meta":  map[string]any{},
		"error": nil,
	})
}

func (h *AdminContentHandler) RestoreVersion(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	detail, err := h.contentService.RestoreVersionAsDraft(r.Context(), r.PathValue("versionId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": detail, "meta": map[string]any{}, "error": nil})
}

func (h *AdminContentHandler) KnowledgeGraph(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	graph, err := h.contentService.BuildKnowledgeGraph(r.Context(), r.URL.Query().Get("exam_id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": graph, "meta": map[string]any{}, "error": nil})
}

func (h *AdminContentHandler) ExportContentPackage(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}
	payload, err := h.contentService.ExportContentPackage(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="foco-content-package.json"`)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *AdminContentHandler) ImportContentPackage(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"data":  nil,
			"meta":  map[string]any{},
			"error": "content service unavailable",
		})
		return
	}
	var raw map[string]any
	var payload content.ContentPackagePayload
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"data":  nil,
				"meta":  map[string]any{},
				"error": "invalid content package json",
			})
			return
		}
	}
	b, _ := json.Marshal(raw)
	if err := json.Unmarshal(b, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"data":  nil,
			"meta":  map[string]any{},
			"error": "invalid content package structure: " + err.Error(),
		})
		return
	}
	report, err := h.contentService.ImportContentPackage(r.Context(), raw, &payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"data":  nil,
			"meta":  map[string]any{},
			"error": err.Error(),
		})
		return
	}
	status := http.StatusOK
	if len(report.ValidationErrors) > 0 {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]any{"data": report, "meta": map[string]any{}, "error": nil})
}

func jsonString(v any) string {
	if v == nil {
		return "{}"
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}

func (h *AdminContentHandler) CreateExam(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	exam, err := h.contentService.CreateExam(r.Context(), body.Code, body.Name, body.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": exam})
}

func (h *AdminContentHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExamId    string `json:"exam_id"`
		Code      string `json:"code"`
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	sub, err := h.contentService.CreateSubject(r.Context(), body.ExamId, body.Code, body.Name, body.SortOrder)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": sub})
}

func (h *AdminContentHandler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SubjectId string `json:"subject_id"`
		Code      string `json:"code"`
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	ch, err := h.contentService.CreateChapter(r.Context(), body.SubjectId, body.Code, body.Name, body.SortOrder)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": ch})
}

func (h *AdminContentHandler) RenameExam(w http.ResponseWriter, r *http.Request) {
	h.renameNode(w, r, "exam", r.PathValue("examId"))
}

func (h *AdminContentHandler) RenameSubject(w http.ResponseWriter, r *http.Request) {
	h.renameNode(w, r, "subject", r.PathValue("subjectId"))
}

func (h *AdminContentHandler) RenameChapter(w http.ResponseWriter, r *http.Request) {
	h.renameNode(w, r, "chapter", r.PathValue("chapterId"))
}

func (h *AdminContentHandler) renameNode(w http.ResponseWriter, r *http.Request, nodeType, nodeID string) {
	var body struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name is required"})
		return
	}
	if err := h.contentService.RenameNode(r.Context(), nodeType, nodeID, body.Name); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": nodeID, "name": body.Name, "type": nodeType}})
}

func (h *AdminContentHandler) CreateKnowledgePoint(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExamId      string  `json:"exam_id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	kp, err := h.contentService.CreateKnowledgePoint(r.Context(), body.ExamId, body.Code, body.Name, body.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": kp})
}

func (h *AdminContentHandler) ListKnowledgePointEdges(w http.ResponseWriter, r *http.Request) {
	examId := r.URL.Query().Get("exam_id")
	edges, err := h.contentService.ListKnowledgePointEdges(r.Context(), examId)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": edges})
}

func (h *AdminContentHandler) CreateKnowledgePointEdge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExamId string   `json:"exam_id"`
		FromId string   `json:"from_knowledge_point_id"`
		ToId   string   `json:"to_knowledge_point_id"`
		Type   string   `json:"edge_type"`
		Weight *float64 `json:"weight"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	edge, err := h.contentService.CreateKnowledgePointEdge(r.Context(), body.ExamId, body.FromId, body.ToId, body.Type, body.Weight)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": edge})
}

func (h *AdminContentHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ExamId    string  `json:"exam_id"`
		SubjectId string  `json:"subject_id"`
		ChapterId *string `json:"chapter_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	q, err := h.contentService.CreateQuestion(r.Context(), body.ExamId, body.SubjectId, body.ChapterId)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": q})
}

func (h *AdminContentHandler) CreateQuestionVersion(w http.ResponseWriter, r *http.Request) {
	questionId := r.PathValue("questionId")
	var body struct {
		QuestionType  string `json:"question_type"`
		Difficulty    int    `json:"difficulty"`
		Stem          any    `json:"stem"`
		Options       any    `json:"options"`
		CorrectAnswer any    `json:"correct_answer"`
		Explanation   any    `json:"explanation"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	v, err := h.contentService.CreateQuestionVersion(
		r.Context(), questionId, body.QuestionType, body.Difficulty,
		jsonString(body.Stem), jsonString(body.Options),
		jsonString(body.CorrectAnswer), jsonString(body.Explanation),
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": v})
}

func (h *AdminContentHandler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	versionId := r.PathValue("versionId")
	var body struct {
		PublishNote *string `json:"publish_note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	err := h.contentService.PublishVersion(r.Context(), versionId, "", body.PublishNote)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	detail, _ := h.contentService.GetVersionDetail(r.Context(), versionId)
	writeJSON(w, http.StatusOK, map[string]any{"data": detail})
}

func (h *AdminContentHandler) DeleteExam(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")
	count, _ := h.contentService.CountQuestionsByExam(r.Context(), examId)
	if count > 0 {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "该考试下存在题目，请先删除相关题目"})
		return
	}
	if err := h.contentService.DeleteExam(r.Context(), examId); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": nil})
}

func (h *AdminContentHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	subjectId := r.PathValue("subjectId")
	count, _ := h.contentService.CountQuestionsBySubject(r.Context(), subjectId)
	if count > 0 {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "该科目下存在题目，请先删除相关题目"})
		return
	}
	if err := h.contentService.DeleteSubject(r.Context(), subjectId); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": nil})
}

func (h *AdminContentHandler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	chapterId := r.PathValue("chapterId")
	count, _ := h.contentService.CountQuestionsByChapter(r.Context(), chapterId)
	if count > 0 {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "该章节下存在题目，请先删除相关题目"})
		return
	}
	if err := h.contentService.DeleteChapter(r.Context(), chapterId); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": nil})
}

func (h *AdminContentHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionId := r.PathValue("questionId")
	if err := h.contentService.DeleteQuestion(r.Context(), questionId); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": nil})
}
