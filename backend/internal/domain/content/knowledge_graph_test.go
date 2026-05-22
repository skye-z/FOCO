package content

import "testing"

func TestAssembleKnowledgeGraphIncludesFullExamSubjectChapterKnowledgePointQuestionNetwork(t *testing.T) {
	t.Helper()

	exam := Exam{Id: "exam-1", Code: "cfa-l1", Name: "CFA Level I"}
	subject := Subject{Id: "subject-1", ExamId: exam.Id, Code: "ethics", Name: "Ethics", SortOrder: 1}
	chapter := Chapter{Id: "chapter-1", SubjectId: subject.Id, Code: "std", Name: "Standards", SortOrder: 1}
	kp := KnowledgePoint{Id: "kp-1", ExamId: exam.Id, Code: "kp-code", Name: "Professionalism"}
	question := knowledgeGraphQuestionRow{
		QuestionId:  "question-1",
		VersionId:   "version-1",
		StemPreview: `{"content":"Which standard applies?"}`,
		ExamId:      exam.Id,
		SubjectId:   subject.Id,
		ChapterId:   &chapter.Id,
	}
	qkp := knowledgeGraphQuestionEdgeRow{QuestionId: question.QuestionId, KnowledgePointId: kp.Id}

	payload := assembleKnowledgeGraph(
		[]Exam{exam},
		[]Subject{subject},
		[]Chapter{chapter},
		[]KnowledgePoint{kp},
		nil,
		[]knowledgeGraphQuestionRow{question},
		[]knowledgeGraphQuestionEdgeRow{qkp},
	)

	assertNodeType(t, payload.Nodes, "exam:"+exam.Id, "exam")
	assertNodeType(t, payload.Nodes, "subject:"+subject.Id, "subject")
	assertNodeType(t, payload.Nodes, "chapter:"+chapter.Id, "chapter")
	assertNodeType(t, payload.Nodes, "kp:"+kp.Id, "knowledge_point")
	assertNodeType(t, payload.Nodes, "question:"+question.QuestionId, "question")

	assertEdgeType(t, payload.Edges, "examsubject:"+exam.Id+":"+subject.Id, "exam_subject")
	assertEdgeType(t, payload.Edges, "subjectchapter:"+subject.Id+":"+chapter.Id, "subject_chapter")
	assertEdgeType(t, payload.Edges, "chapterquestion:"+chapter.Id+":"+question.QuestionId, "chapter_question")
	assertEdgeType(t, payload.Edges, "qkp:"+question.QuestionId+":"+kp.Id, "question_tag")
	assertEdgeMissing(t, payload.Edges, "subjectquestion:"+subject.Id+":"+question.QuestionId)
	assertEdgeMissing(t, payload.Edges, "examkp:"+exam.Id+":"+kp.Id)
}

func TestAssembleKnowledgeGraphFallsBackToSubjectWhenQuestionHasNoChapter(t *testing.T) {
	t.Helper()

	exam := Exam{Id: "exam-1", Code: "cfa-l1", Name: "CFA Level I"}
	subject := Subject{Id: "subject-1", ExamId: exam.Id, Code: "ethics", Name: "Ethics", SortOrder: 1}
	kp := KnowledgePoint{Id: "kp-1", ExamId: exam.Id, Code: "kp-code", Name: "Professionalism"}
	question := knowledgeGraphQuestionRow{
		QuestionId:  "question-1",
		VersionId:   "version-1",
		StemPreview: `{"content":"Which standard applies?"}`,
		ExamId:      exam.Id,
		SubjectId:   subject.Id,
		ChapterId:   nil,
	}
	qkp := knowledgeGraphQuestionEdgeRow{QuestionId: question.QuestionId, KnowledgePointId: kp.Id}

	payload := assembleKnowledgeGraph(
		[]Exam{exam},
		[]Subject{subject},
		nil,
		[]KnowledgePoint{kp},
		nil,
		[]knowledgeGraphQuestionRow{question},
		[]knowledgeGraphQuestionEdgeRow{qkp},
	)

	assertEdgeType(t, payload.Edges, "subjectquestion:"+subject.Id+":"+question.QuestionId, "subject_question")
	assertEdgeMissing(t, payload.Edges, "chapterquestion:chapter-1:"+question.QuestionId)
}

func assertNodeType(t *testing.T, nodes []KnowledgeGraphNode, id, nodeType string) {
	t.Helper()
	for _, node := range nodes {
		if node.Id == id {
			if node.Type != nodeType {
				t.Fatalf("expected node %s to have type %s, got %s", id, nodeType, node.Type)
			}
			return
		}
	}
	t.Fatalf("expected node %s to exist", id)
}

func assertEdgeType(t *testing.T, edges []KnowledgeGraphEdge, id, edgeType string) {
	t.Helper()
	for _, edge := range edges {
		if edge.Id == id {
			if edge.Type != edgeType {
				t.Fatalf("expected edge %s to have type %s, got %s", id, edgeType, edge.Type)
			}
			return
		}
	}
	t.Fatalf("expected edge %s to exist", id)
}

func assertEdgeMissing(t *testing.T, edges []KnowledgeGraphEdge, id string) {
	t.Helper()
	for _, edge := range edges {
		if edge.Id == id {
			t.Fatalf("expected edge %s to be absent", id)
		}
	}
}
