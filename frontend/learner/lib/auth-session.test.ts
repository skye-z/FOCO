import { describe, expect, it } from "vitest"

import {
  mergeSessionWithActiveEnrollment,
  type StoredAuthSession,
} from "./auth-session"

describe("mergeSessionWithActiveEnrollment", () => {
  it("persists the active exam from the learner identity payload", () => {
    const session: StoredAuthSession = {
      accessToken: "token",
      refreshToken: "refresh",
      user: { id: "user-1", email: "learner@example.com" },
    }

    expect(
      mergeSessionWithActiveEnrollment(session, {
        id: "enrollment-1",
        exam_id: "exam-1",
        exam_name: "CFA 一级",
        exam_code: "cfa-level-1",
        status: "in_progress",
      })
    ).toEqual({
      accessToken: "token",
      refreshToken: "refresh",
      user: { id: "user-1", email: "learner@example.com" },
      activeExam: {
        id: "exam-1",
        name: "CFA 一级",
      },
    })
  })

  it("preserves the rest of the session when there is no active enrollment", () => {
    const session: StoredAuthSession = {
      accessToken: "token",
      refreshToken: "refresh",
      activeExam: { id: "exam-old", name: "旧考试" },
    }

    expect(mergeSessionWithActiveEnrollment(session, null)).toEqual({
      accessToken: "token",
      refreshToken: "refresh",
    })
  })
})
