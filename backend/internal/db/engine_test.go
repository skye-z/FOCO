package db

import (
	"strings"
	"testing"
)

func TestMigrationStatementsPatchExistingPracticeSessionsSchema(t *testing.T) {
	statements := strings.Join(migrationStatements(), "\n")

	if !strings.Contains(statements, "ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS exam_enrollment_id") {
		t.Fatalf("expected practice_sessions migration to patch exam_enrollment_id for existing tables, got:\n%s", statements)
	}
	if !strings.Contains(statements, "ALTER TABLE practice_sessions ADD COLUMN IF NOT EXISTS total_count") {
		t.Fatalf("expected practice_sessions migration to patch total_count for existing tables, got:\n%s", statements)
	}
	if !strings.Contains(statements, "ALTER TABLE practice_session_items ADD COLUMN IF NOT EXISTS submitted_at") {
		t.Fatalf("expected practice_session_items migration to patch submitted_at for existing tables, got:\n%s", statements)
	}
	if !strings.Contains(statements, "UPDATE exams") || !strings.Contains(statements, "SET next_exam_date = COALESCE") {
		t.Fatalf("expected exam date backfill migration for confirmed exam data, got:\n%s", statements)
	}
}
