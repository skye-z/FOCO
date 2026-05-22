package apihttp

import (
	"net/http"

	"foco/backend/api/internal/http/handler"
	"foco/backend/api/internal/http/middleware"
)

type Dependencies struct {
	TokenVerifier        middleware.TokenVerifier
	AccountService       handler.AccountService
	AdminUserService     handler.AdminUserService
	AdminSettingsService handler.AdminSettingsService
	AuditWriter          handler.AuditWriter
	SeedService          handler.SeedService
	StatsService         handler.StatsService
	ContentService       handler.ContentService
	DiagnosticService    handler.DiagnosticService
	HomeService          handler.HomeService
	ProfileService       handler.ProfileService
	PracticeService      handler.PracticeService
	InteractiveService   handler.InteractiveService
	SeedChinese          http.HandlerFunc
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handler.HealthHandler)

	seedHandler := handler.NewSeedHandler(deps.SeedService)
	adminSettingsHandler := handler.NewAdminSettingsHandler(deps.AdminSettingsService)
	mux.HandleFunc("POST /api/v1/seed/admin", seedHandler.SeedDefaultAdmin)
	mux.HandleFunc("GET /api/v1/public/settings", adminSettingsHandler.GetPublic)

	learnerIdentityHandler := handler.NewLearnerIdentityHandler(deps.AccountService)
	learnerContentHandler := handler.NewLearnerContentHandler(deps.ContentService)
	learnerDiagnosticHandler := handler.NewLearnerDiagnosticHandler(deps.DiagnosticService)
	learnerHomeHandler := handler.NewLearnerHomeHandler(deps.HomeService)
	learnerProfileHandler := handler.NewLearnerProfileHandler(deps.ProfileService)
	learnerPracticeHandler := handler.NewLearnerPracticeHandler(deps.PracticeService)
	learnerInteractiveHandler := handler.NewLearnerInteractiveHandler(deps.InteractiveService)
	adminStatsHandler := handler.NewAdminStatsHandler(deps.StatsService)
	adminUsersHandler := handler.NewAdminUsersHandler(deps.AdminUserService)
	adminContentHandler := handler.NewAdminContentHandler(deps.ContentService)
	adminInteractiveHandler := handler.NewAdminInteractiveHandler(deps.InteractiveService)
	withAuth := middleware.AuthMiddleware(deps.TokenVerifier)

	mux.Handle("GET /api/v1/me", withAuth(http.HandlerFunc(learnerIdentityHandler.Me)))
	mux.Handle("POST /api/v1/learner/bootstrap", withAuth(http.HandlerFunc(learnerIdentityHandler.Bootstrap)))
	mux.Handle("GET /api/v1/learner/exams", withAuth(http.HandlerFunc(learnerIdentityHandler.ListExams)))
	mux.Handle("POST /api/v1/learner/exam-enrollments", withAuth(http.HandlerFunc(learnerIdentityHandler.CreateEnrollment)))
	mux.Handle("GET /api/v1/learner/diagnostic/current", withAuth(http.HandlerFunc(learnerDiagnosticHandler.GetCurrent)))
	mux.Handle("POST /api/v1/learner/diagnostic/restart", withAuth(http.HandlerFunc(learnerDiagnosticHandler.Restart)))
	mux.Handle("POST /api/v1/learner/diagnostic/{attemptId}/submit", withAuth(http.HandlerFunc(learnerDiagnosticHandler.Submit)))
	mux.Handle("GET /api/v1/learner/home", withAuth(http.HandlerFunc(learnerHomeHandler.GetHome)))
	mux.Handle("GET /api/v1/learner/recommendations", withAuth(http.HandlerFunc(learnerHomeHandler.GetRecommendations)))
	mux.Handle("GET /api/v1/learner/exams/{examId}/content-tree", withAuth(http.HandlerFunc(learnerContentHandler.ExamContentTree)))
	mux.Handle("GET /api/v1/learner/profile", withAuth(http.HandlerFunc(learnerProfileHandler.GetProfile)))
	mux.Handle("GET /api/v1/learner/wrong-book", withAuth(http.HandlerFunc(learnerPracticeHandler.WrongBook)))
	mux.Handle("POST /api/v1/learner/practice-sessions", withAuth(http.HandlerFunc(learnerPracticeHandler.CreateSession)))
	mux.Handle("GET /api/v1/learner/practice-sessions/{sessionId}", withAuth(http.HandlerFunc(learnerPracticeHandler.GetSession)))
	mux.Handle("POST /api/v1/learner/practice-sessions/{sessionId}/items/{itemId}/submit", withAuth(http.HandlerFunc(learnerPracticeHandler.SubmitAnswer)))
	mux.Handle("GET /api/v1/learner/practice-sessions/{sessionId}/summary", withAuth(http.HandlerFunc(learnerPracticeHandler.Summary)))
	mux.Handle("GET /api/v1/learner/interactive-units", withAuth(http.HandlerFunc(learnerInteractiveHandler.ListUnits)))
	mux.Handle("GET /api/v1/learner/interactive-units/{unitVersionId}", withAuth(http.HandlerFunc(learnerInteractiveHandler.GetUnit)))
	mux.Handle("POST /api/v1/learner/interactive-units/{unitVersionId}/attempts", withAuth(http.HandlerFunc(learnerInteractiveHandler.StartAttempt)))
	mux.Handle("POST /api/v1/learner/interactive-unit-attempts/{attemptId}/steps/{stepId}/actions", withAuth(http.HandlerFunc(learnerInteractiveHandler.SubmitStepAction)))
	mux.Handle("POST /api/v1/learner/interactive-unit-attempts/{attemptId}/complete", withAuth(http.HandlerFunc(learnerInteractiveHandler.CompleteAttempt)))
	mux.Handle("GET /api/v1/admin/stats", withAuth(http.HandlerFunc(adminStatsHandler.Overview)))
	mux.Handle("GET /api/v1/admin/settings", withAuth(http.HandlerFunc(adminSettingsHandler.Get)))
	mux.Handle("PATCH /api/v1/admin/settings", withAuth(http.HandlerFunc(adminSettingsHandler.Update)))
	mux.Handle("GET /api/v1/admin/users", withAuth(http.HandlerFunc(adminUsersHandler.List)))
	mux.Handle("POST /api/v1/admin/users/{userId}/roles", withAuth(http.HandlerFunc(adminUsersHandler.GrantRole)))
	mux.Handle("POST /api/v1/admin/users/{userId}/disable", withAuth(http.HandlerFunc(adminUsersHandler.DisableUser)))
	mux.Handle("POST /api/v1/admin/users/{userId}/reset-password", withAuth(http.HandlerFunc(adminUsersHandler.ResetPassword)))
	mux.Handle("GET /api/v1/admin/exam-tree", withAuth(http.HandlerFunc(adminContentHandler.ExamTree)))
	mux.Handle("GET /api/v1/admin/content-package/export", withAuth(http.HandlerFunc(adminContentHandler.ExportContentPackage)))
	mux.Handle("POST /api/v1/admin/content-package/import", withAuth(http.HandlerFunc(adminContentHandler.ImportContentPackage)))
	mux.Handle("GET /api/v1/admin/knowledge-points", withAuth(http.HandlerFunc(adminContentHandler.KnowledgePoints)))
	mux.Handle("GET /api/v1/admin/knowledge-graph", withAuth(http.HandlerFunc(adminContentHandler.KnowledgeGraph)))
	mux.Handle("GET /api/v1/admin/questions", withAuth(http.HandlerFunc(adminContentHandler.ListQuestions)))
	mux.Handle("GET /api/v1/admin/questions/{questionId}/versions", withAuth(http.HandlerFunc(adminContentHandler.ListQuestionVersions)))
	mux.Handle("POST /api/v1/admin/exams", withAuth(http.HandlerFunc(adminContentHandler.CreateExam)))
	mux.Handle("PATCH /api/v1/admin/exams/{examId}", withAuth(http.HandlerFunc(adminContentHandler.RenameExam)))
	mux.Handle("DELETE /api/v1/admin/exams/{examId}", withAuth(http.HandlerFunc(adminContentHandler.DeleteExam)))
	mux.Handle("POST /api/v1/admin/subjects", withAuth(http.HandlerFunc(adminContentHandler.CreateSubject)))
	mux.Handle("PATCH /api/v1/admin/subjects/{subjectId}", withAuth(http.HandlerFunc(adminContentHandler.RenameSubject)))
	mux.Handle("DELETE /api/v1/admin/subjects/{subjectId}", withAuth(http.HandlerFunc(adminContentHandler.DeleteSubject)))
	mux.Handle("POST /api/v1/admin/chapters", withAuth(http.HandlerFunc(adminContentHandler.CreateChapter)))
	mux.Handle("PATCH /api/v1/admin/chapters/{chapterId}", withAuth(http.HandlerFunc(adminContentHandler.RenameChapter)))
	mux.Handle("DELETE /api/v1/admin/chapters/{chapterId}", withAuth(http.HandlerFunc(adminContentHandler.DeleteChapter)))
	mux.Handle("POST /api/v1/admin/knowledge-points", withAuth(http.HandlerFunc(adminContentHandler.CreateKnowledgePoint)))
	mux.Handle("GET /api/v1/admin/knowledge-point-edges", withAuth(http.HandlerFunc(adminContentHandler.ListKnowledgePointEdges)))
	mux.Handle("POST /api/v1/admin/knowledge-point-edges", withAuth(http.HandlerFunc(adminContentHandler.CreateKnowledgePointEdge)))
	mux.Handle("POST /api/v1/admin/questions", withAuth(http.HandlerFunc(adminContentHandler.CreateQuestion)))
	mux.Handle("DELETE /api/v1/admin/questions/{questionId}", withAuth(http.HandlerFunc(adminContentHandler.DeleteQuestion)))
	mux.Handle("POST /api/v1/admin/questions/{questionId}/versions", withAuth(http.HandlerFunc(adminContentHandler.CreateQuestionVersion)))
	mux.Handle("GET /api/v1/admin/question-versions/{versionId}", withAuth(http.HandlerFunc(adminContentHandler.GetVersionDetail)))
	mux.Handle("PATCH /api/v1/admin/question-versions/{versionId}", withAuth(http.HandlerFunc(adminContentHandler.UpdateVersion)))
	mux.Handle("POST /api/v1/admin/question-versions/{versionId}/restore", withAuth(http.HandlerFunc(adminContentHandler.RestoreVersion)))
	mux.Handle("POST /api/v1/admin/question-versions/{versionId}/publish", withAuth(http.HandlerFunc(adminContentHandler.PublishVersion)))

	mux.Handle("GET /api/v1/admin/interactive-units", withAuth(http.HandlerFunc(adminInteractiveHandler.ListUnits)))
	mux.Handle("POST /api/v1/admin/interactive-units", withAuth(http.HandlerFunc(adminInteractiveHandler.CreateUnit)))
	mux.Handle("GET /api/v1/admin/interactive-units/{unitId}/versions", withAuth(http.HandlerFunc(adminInteractiveHandler.ListVersions)))
	mux.Handle("POST /api/v1/admin/interactive-units/{unitId}/versions", withAuth(http.HandlerFunc(adminInteractiveHandler.CreateVersion)))
	mux.Handle("GET /api/v1/admin/interactive-unit-versions/{versionId}", withAuth(http.HandlerFunc(adminInteractiveHandler.GetVersionDetail)))
	mux.Handle("PATCH /api/v1/admin/interactive-unit-versions/{versionId}", withAuth(http.HandlerFunc(adminInteractiveHandler.UpdateVersion)))
	mux.Handle("POST /api/v1/admin/interactive-unit-versions/{versionId}/publish", withAuth(http.HandlerFunc(adminInteractiveHandler.PublishVersion)))
	mux.Handle("DELETE /api/v1/admin/interactive-units/{unitId}", withAuth(http.HandlerFunc(adminInteractiveHandler.DeleteUnit)))

	mux.HandleFunc("POST /api/v1/seed/chinese", deps.SeedChinese)

	return middleware.SecurityHeadersMiddleware(middleware.CORSMiddleware(middleware.RateLimitMiddleware(mux)))
}
