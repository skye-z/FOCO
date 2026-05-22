import { describe, expect, it } from "vitest"

import {
  buildPracticeSessionRequest,
  filterVisibleChapters,
  isSmartPracticeLocked,
  resolveInitialPracticeExamId,
  type ContentTreeSubject,
} from "./practice-setup"

describe("practice setup helpers", () => {
  const subjects: ContentTreeSubject[] = [
    {
      id: "subject-1",
      name: "数量方法",
      chapters: [
        { id: "chapter-1", name: "现值折现" },
        { id: "chapter-2", name: "年金终值" },
      ],
    },
    {
      id: "subject-2",
      name: "职业道德",
      chapters: [
        { id: "chapter-3", name: "独立性与客观性" },
      ],
    },
  ]

  it("prefers the active exam id when present", () => {
    expect(
      resolveInitialPracticeExamId("exam-active", [
        { id: "exam-fallback" },
        { id: "exam-active" },
      ]),
    ).toBe("exam-active")
  })

  it("falls back to the first exam when no active exam is stored", () => {
    expect(
      resolveInitialPracticeExamId(null, [
        { id: "exam-1" },
        { id: "exam-2" },
      ]),
    ).toBe("exam-1")
  })

  it("builds a manual practice session request", () => {
    expect(
      buildPracticeSessionRequest({
        examId: "exam-1",
        mode: "manual",
        questionTypes: ["single_choice", "multiple_choice", "judgment"],
        difficulty: "medium",
        count: 12,
        selectedSubjectIds: ["subject-1"],
        selectedChapterIds: ["chapter-2"],
      }),
    ).toEqual({
      exam_id: "exam-1",
      mode: "manual",
      question_types: ["single_choice", "multiple_choice", "judgment"],
      difficulty: "medium",
      count: 12,
      subject_ids: ["subject-1"],
      chapter_ids: ["chapter-2"],
    })
  })

  it("filters visible chapters by selected subjects and keyword", () => {
    expect(filterVisibleChapters(subjects, new Set(["subject-1"]), "年金")).toEqual([
      {
        chapter: { id: "chapter-2", name: "年金终值" },
        subjectId: "subject-1",
        subjectName: "数量方法",
      },
    ])
  })

  it("locks smart practice until the diagnostic is completed", () => {
    expect(isSmartPracticeLocked(null)).toBe(true)
    expect(isSmartPracticeLocked({ has_completed: false })).toBe(true)
    expect(isSmartPracticeLocked({ has_completed: true })).toBe(false)
  })
})
