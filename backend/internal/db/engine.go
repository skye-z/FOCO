package db

import (
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"xorm.io/xorm"
)

func OpenEngine(dsn string) (*xorm.Engine, error) {
	connString := dsn
	if strings.Contains(connString, "pooler.supabase.com") {
		if strings.Contains(connString, "?") {
			connString += "&default_query_exec_mode=simple_protocol"
		} else {
			connString += "?default_query_exec_mode=simple_protocol"
		}
	}

	engine, err := xorm.NewEngine("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("create engine: %w", err)
	}

	if err := engine.Ping(); err != nil {
		_ = engine.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	engine.SetMaxOpenConns(10)
	engine.SetMaxIdleConns(5)
	engine.SetConnMaxIdleTime(0)
	return engine, nil
}

func RunMigrations(engine *xorm.Engine) {
	for _, sql := range migrationStatements() {
		if _, err := engine.Exec(sql); err != nil {
			fmt.Printf("migration warning: %v\n", err)
		}
	}
}

func migrationStatements() []string {
	return []string{
		`ALTER TABLE interactive_unit_version_steps ADD COLUMN IF NOT EXISTS knowledge_point_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE interactive_unit_version_steps ADD COLUMN IF NOT EXISTS knowledge_point_tags jsonb NOT NULL DEFAULT '[]'::jsonb`,
		`CREATE TABLE IF NOT EXISTS practice_sessions (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
			exam_enrollment_id uuid NULL REFERENCES exam_enrollments(id) ON DELETE SET NULL,
			status text NOT NULL DEFAULT 'in_progress',
			total_count int NOT NULL,
			answered_count int NOT NULL DEFAULT 0,
			correct_count int NOT NULL DEFAULT 0,
			total_duration_seconds int NOT NULL DEFAULT 0,
			xp_earned int NOT NULL DEFAULT 0,
			coins_earned int NOT NULL DEFAULT 0,
			created_at timestamptz NOT NULL DEFAULT now(),
			updated_at timestamptz NOT NULL DEFAULT now(),
			completed_at timestamptz NULL
		)`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS user_id uuid NULL`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS exam_id uuid NULL`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS exam_enrollment_id uuid NULL REFERENCES exam_enrollments(id) ON DELETE SET NULL`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS scope jsonb NOT NULL DEFAULT '{}'::jsonb`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'in_progress'`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS started_at timestamptz NOT NULL DEFAULT now()`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS total_count int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS answered_count int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS correct_count int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS total_duration_seconds int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS xp_earned int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS coins_earned int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now()`,
		`ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS completed_at timestamptz NULL`,
		`CREATE TABLE IF NOT EXISTS practice_session_items (
			id uuid PRIMARY KEY,
			session_id uuid NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
			question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
			question_version_id uuid NOT NULL REFERENCES question_versions(id) ON DELETE CASCADE,
			subject_id uuid NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
			chapter_id uuid NULL REFERENCES chapters(id) ON DELETE SET NULL,
			question_type text NOT NULL,
			score int NOT NULL DEFAULT 1,
			position int NOT NULL,
			stem text NOT NULL,
			options jsonb NOT NULL DEFAULT '[]'::jsonb,
			correct_labels jsonb NOT NULL DEFAULT '[]'::jsonb,
			explanation text NOT NULL DEFAULT '',
			knowledge_point_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
			user_answer jsonb NULL,
			is_correct boolean NULL,
			duration_seconds int NOT NULL DEFAULT 0,
			submitted_at timestamptz NULL,
			created_at timestamptz NOT NULL DEFAULT now(),
			UNIQUE(session_id, position)
		)`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS question_id uuid NULL REFERENCES questions(id) ON DELETE CASCADE`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS practice_session_id uuid NULL REFERENCES practice_sessions(id) ON DELETE CASCADE`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS session_id uuid NULL REFERENCES practice_sessions(id) ON DELETE CASCADE`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS question_version_id uuid NULL REFERENCES question_versions(id) ON DELETE CASCADE`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS subject_id uuid NULL REFERENCES subjects(id) ON DELETE CASCADE`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS chapter_id uuid NULL REFERENCES chapters(id) ON DELETE SET NULL`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS question_type text NOT NULL DEFAULT 'single_choice'`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS score int NOT NULL DEFAULT 1`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS position int NOT NULL DEFAULT 1`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS seq_no int NOT NULL DEFAULT 1`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'pending'`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS last_submission_id uuid NULL`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS stem text NOT NULL DEFAULT ''`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS explanation text NOT NULL DEFAULT ''`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS options jsonb NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS correct_labels jsonb NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS knowledge_point_ids jsonb NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS user_answer jsonb NULL`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS is_correct boolean NULL`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS duration_seconds int NOT NULL DEFAULT 0`,
		`ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS submitted_at timestamptz NULL`,
		`ALTER TABLE exams ADD COLUMN IF NOT EXISTS next_exam_date timestamptz NULL`,
		`ALTER TABLE exams ADD COLUMN IF NOT EXISTS next_next_exam_date timestamptz NULL`,
		`UPDATE exams
		 SET next_exam_date = COALESCE(next_exam_date, '2026-08-25T00:00:00Z'::timestamptz),
		     next_next_exam_date = COALESCE(next_next_exam_date, '2026-11-17T00:00:00Z'::timestamptz),
		     updated_at = now()
		 WHERE id = '73000000-0000-0000-0000-000000000101'::uuid`,
		`CREATE TABLE IF NOT EXISTS diagnostic_attempts (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
			trigger_type text NOT NULL,
			status text NOT NULL,
			started_at timestamptz NULL,
			completed_at timestamptz NULL,
			summary jsonb NULL,
			created_at timestamptz NOT NULL DEFAULT now()
		)`,
		`ALTER TABLE diagnostic_attempts ADD COLUMN IF NOT EXISTS started_at timestamptz NULL`,
		`ALTER TABLE diagnostic_attempts ADD COLUMN IF NOT EXISTS completed_at timestamptz NULL`,
		`ALTER TABLE diagnostic_attempts ADD COLUMN IF NOT EXISTS summary jsonb NULL`,
		`CREATE TABLE IF NOT EXISTS learner_profiles (
			id uuid PRIMARY KEY,
			user_id uuid NOT NULL,
			exam_id uuid NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
			profile_version int NOT NULL,
			profile_summary jsonb NOT NULL,
			confidence_score numeric NOT NULL DEFAULT 0,
			source_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb,
			computed_at timestamptz NOT NULL DEFAULT now(),
			UNIQUE(user_id, exam_id, profile_version)
		)`,
		`CREATE TABLE IF NOT EXISTS unit_attempts (
			id uuid PRIMARY KEY,
			user_id text NOT NULL,
			unit_version_id uuid NOT NULL,
			status text NOT NULL DEFAULT 'in_progress',
			created_at timestamptz NOT NULL DEFAULT now(),
			completed_at timestamptz NULL
		)`,
		`CREATE TABLE IF NOT EXISTS step_actions (
			id uuid PRIMARY KEY,
			attempt_id uuid NOT NULL REFERENCES unit_attempts(id) ON DELETE CASCADE,
			step_id uuid NOT NULL,
			action_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
			created_at timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS step_feedback (
			id uuid PRIMARY KEY,
			attempt_id uuid NOT NULL REFERENCES unit_attempts(id) ON DELETE CASCADE,
			step_id uuid NOT NULL,
			is_correct boolean NOT NULL DEFAULT false,
			allow_continue boolean NOT NULL DEFAULT false,
			hint text NOT NULL DEFAULT '',
			created_at timestamptz NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS concept_cards (
			id uuid PRIMARY KEY,
			user_id text NOT NULL,
			attempt_id uuid NOT NULL REFERENCES unit_attempts(id) ON DELETE CASCADE,
			unit_version_id uuid NOT NULL,
			content jsonb NOT NULL DEFAULT '{}'::jsonb,
			created_at timestamptz NOT NULL DEFAULT now()
		)`,
	}
}
