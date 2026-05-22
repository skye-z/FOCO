package main

import (
	"log"
	"net/http"
	"os"

	"foco/backend/api/internal/app"
	dbpkg "foco/backend/api/internal/db"
	apihttp "foco/backend/api/internal/http"
	"xorm.io/xorm"
)

func main() {
	addr := os.Getenv("GO_API_PORT")
	if addr == "" {
		addr = "8080"
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	publishableKey := os.Getenv("NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY")
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	dbURL := os.Getenv("SUPABASE_DB_URL")
	redisURL := os.Getenv("REDIS_URL")

	var engine *xorm.Engine
	var err error
	if dbURL != "" {
		engine, err = dbpkg.OpenEngine(dbURL)
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		defer engine.Close()
		dbpkg.RunMigrations(engine)
	}

	deps := app.BuildDependencies(engine, app.Config{
		SupabaseURL:    supabaseURL,
		PublishableKey: publishableKey,
		ServiceRoleKey: serviceRoleKey,
		DatabaseURL:    dbURL,
		RedisURL:       redisURL,
	})
	router := apihttp.NewRouter(deps)

	log.Printf("api listening on :%s", addr)
	if err := http.ListenAndServe(":"+addr, router); err != nil {
		log.Fatal(err)
	}
}
