package app

import (
	"net/http"

	auditpkg "foco/backend/api/internal/audit"
	authpkg "foco/backend/api/internal/auth"
	cachepkg "foco/backend/api/internal/cache"
	accountpkg "foco/backend/api/internal/domain/account"
	contentpkg "foco/backend/api/internal/domain/content"
	diagnosticpkg "foco/backend/api/internal/domain/diagnostic"
	homepkg "foco/backend/api/internal/domain/home"
	interactivepkg "foco/backend/api/internal/domain/interactive"
	platformpkg "foco/backend/api/internal/domain/platform"
	practicepkg "foco/backend/api/internal/domain/practice"
	profilepkg "foco/backend/api/internal/domain/profile"
	apihttp "foco/backend/api/internal/http"
	handlerpkg "foco/backend/api/internal/http/handler"
	"xorm.io/xorm"
)

type Config struct {
	SupabaseURL    string
	PublishableKey string
	ServiceRoleKey string
	DatabaseURL    string
	RedisURL       string
}

func BuildDependencies(engine *xorm.Engine, cfg Config) apihttp.Dependencies {
	deps := apihttp.Dependencies{}

	if engine != nil {
		cacheManager := cachepkg.NewManager(cfg.RedisURL)
		platformService := platformpkg.NewService(engine, cacheManager)
		_ = platformService.EnsureSchema()
		roleRepo := authpkg.NewXormRoleRepository(engine)
		deps.TokenVerifier = authpkg.NewSupabaseTokenVerifier(cfg.SupabaseURL, cfg.PublishableKey, http.DefaultClient, roleRepo, cacheManager)
		deps.AuditWriter = auditpkg.NewXormWriter(engine)
		deps.AccountService = accountpkg.NewService(accountpkg.NewRepository(engine), cacheManager)
		deps.AdminUserService = accountpkg.NewAdminUserService(accountpkg.NewRepository(engine), cfg.SupabaseURL, cfg.ServiceRoleKey, http.DefaultClient, cacheManager)
		deps.AdminSettingsService = platformService
		deps.StatsService = accountpkg.NewService(accountpkg.NewRepository(engine), cacheManager)
		deps.ContentService = contentpkg.NewService(contentpkg.NewRepository(engine), cacheManager)
		deps.DiagnosticService = diagnosticpkg.NewService(diagnosticpkg.NewXormRepository(engine), cacheManager)
		deps.HomeService = homepkg.NewService(homepkg.NewXormRepository(engine), cacheManager)
		deps.ProfileService = profilepkg.NewService(profilepkg.NewXormRepository(engine), cacheManager)
		deps.PracticeService = practicepkg.NewService(practicepkg.NewXormRepository(engine), cacheManager)
		deps.InteractiveService = interactivepkg.NewService(interactivepkg.NewXormRepository(engine), cacheManager)
		deps.SeedChinese = handlerpkg.NewSeedChineseHandler(engine).Run
		deps.SeedService = &seedRunner{cfg: cfg, engine: engine}
	}

	return deps
}

type seedRunner struct {
	cfg    Config
	engine *xorm.Engine
}

func (s *seedRunner) SeedDefaultAdmin() (string, bool, error) {
	return accountpkg.SeedDefaultAdmin(accountpkg.SeedConfig{
		SupabaseURL:    s.cfg.SupabaseURL,
		ServiceRoleKey: s.cfg.ServiceRoleKey,
		Engine:         s.engine,
	})
}
