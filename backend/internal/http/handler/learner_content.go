package handler

import (
	"net/http"

	"foco/backend/api/internal/domain/content"
)

type LearnerContentHandler struct {
	contentService ContentService
}

func NewLearnerContentHandler(contentService ContentService) *LearnerContentHandler {
	return &LearnerContentHandler{contentService: contentService}
}

func (h *LearnerContentHandler) ExamContentTree(w http.ResponseWriter, r *http.Request) {
	if h.contentService == nil {
		http.Error(w, "content service unavailable", http.StatusInternalServerError)
		return
	}

	examId := r.PathValue("examId")
	if examId == "" {
		http.Error(w, "exam_id required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	tree, err := h.contentService.GetExamTree(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var examNode *content.ExamTreeNode
	for i := range tree {
		if tree[i].Id == examId {
			examNode = &tree[i]
			break
		}
	}
	if examNode == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "exam not found",
		})
		return
	}

	kps, err := h.contentService.ListKnowledgePoints(ctx, examId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kpByExam := make(map[string][]content.KnowledgePoint, len(kps))
	for _, kp := range kps {
		kpByExam[kp.ExamId] = append(kpByExam[kp.ExamId], kp)
	}

	subjects := make([]map[string]any, 0, len(examNode.Children))
	for _, subj := range examNode.Children {
		chapters := make([]map[string]any, 0, len(subj.Children))
		for _, ch := range subj.Children {
			kpItems := make([]map[string]any, 0)
			chapters = append(chapters, map[string]any{
				"id":   ch.Id,
				"name": ch.Name,
				"knowledge_points": kpItems,
			})
		}
		subjects = append(subjects, map[string]any{
			"id":       subj.Id,
			"name":     subj.Name,
			"chapters": chapters,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"exam": map[string]any{
				"id":   examNode.Id,
				"name": examNode.Name,
			},
			"subjects": subjects,
		},
		"meta":  map[string]any{},
		"error": nil,
	})
}
