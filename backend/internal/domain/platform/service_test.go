package platform

import (
	"errors"
	"testing"
)

func TestIsMissingAdminSettingsRelationErrorMatchesPostgresAndSQLite(t *testing.T) {
	t.Helper()

	if !isMissingAdminSettingsRelationError(errors.New(`ERROR: relation "admin_settings" does not exist (SQLSTATE 42P01)`)) {
		t.Fatalf("expected postgres missing-table error to match")
	}
	if !isMissingAdminSettingsRelationError(errors.New("no such table: admin_settings")) {
		t.Fatalf("expected sqlite missing-table error to match")
	}
	if isMissingAdminSettingsRelationError(errors.New("permission denied")) {
		t.Fatalf("expected unrelated error not to match")
	}
}

func TestWithAdminSettingsSchemaRetryRetriesMissingTableOnce(t *testing.T) {
	t.Helper()

	attempts := 0
	ensures := 0

	value, err := withAdminSettingsSchemaRetry(func() error {
		ensures++
		return nil
	}, func() (string, error) {
		attempts++
		if attempts == 1 {
			return "", errors.New(`ERROR: relation "admin_settings" does not exist (SQLSTATE 42P01)`)
		}
		return "ok", nil
	})

	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if value != "ok" {
		t.Fatalf("expected retried value to be returned, got %q", value)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if ensures != 1 {
		t.Fatalf("expected ensure schema to run once, got %d", ensures)
	}
}

func TestWithAdminSettingsSchemaRetryDoesNotRetryOtherErrors(t *testing.T) {
	t.Helper()

	attempts := 0
	ensures := 0
	wantErr := errors.New("permission denied")

	_, err := withAdminSettingsSchemaRetry(func() error {
		ensures++
		return nil
	}, func() (string, error) {
		attempts++
		return "", wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
	if ensures != 0 {
		t.Fatalf("expected ensure schema not to run, got %d", ensures)
	}
}
