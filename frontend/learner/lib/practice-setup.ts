export type PracticeExamOption = {
  id: string
}

export type ContentTreeChapter = {
  id: string
  name: string
}

export type ContentTreeSubject = {
  id: string
  name: string
  chapters: ContentTreeChapter[]
}

export type PracticeDifficulty = "easy" | "medium" | "hard"
export type PracticeQuestionType = "single_choice" | "multiple_choice" | "judgment"

export type PracticeSessionRequest = {
  exam_id: string
  mode: "manual" | "intelligent"
  question_types: PracticeQuestionType[]
  difficulty: PracticeDifficulty
  count: number
  subject_ids?: string[]
  chapter_ids?: string[]
}

export function resolveInitialPracticeExamId(
  activeExamId: string | null,
  exams: PracticeExamOption[],
): string | null {
  if (activeExamId) {
    return activeExamId
  }

  return exams[0]?.id ?? null
}

export function buildPracticeSessionRequest(input: {
  examId: string
  mode: "manual" | "intelligent"
  questionTypes: PracticeQuestionType[]
  difficulty: PracticeDifficulty
  count: number
  selectedSubjectIds: string[]
  selectedChapterIds: string[]
}): PracticeSessionRequest {
  const body: PracticeSessionRequest = {
    exam_id: input.examId,
    mode: input.mode,
    question_types: input.questionTypes,
    difficulty: input.difficulty,
    count: input.count,
  }

  if (input.selectedSubjectIds.length > 0) {
    body.subject_ids = input.selectedSubjectIds
  }

  if (input.selectedChapterIds.length > 0) {
    body.chapter_ids = input.selectedChapterIds
  }

  return body
}

export function filterVisibleChapters(
  subjects: ContentTreeSubject[],
  selectedSubjectIds: Set<string>,
  keyword: string,
) {
  const normalizedKeyword = keyword.trim().toLowerCase()

  return subjects
    .filter((subject) =>
      selectedSubjectIds.size === 0 || selectedSubjectIds.has(subject.id)
    )
    .flatMap((subject) =>
      subject.chapters
        .filter((chapter) =>
          normalizedKeyword === ""
            ? true
            : chapter.name.toLowerCase().includes(normalizedKeyword)
        )
        .map((chapter) => ({
          chapter,
          subjectId: subject.id,
          subjectName: subject.name,
        }))
    )
}

export function isSmartPracticeLocked(summary: { has_completed: boolean } | null) {
  return !summary?.has_completed
}
