package account

import (
	"strings"
	"testing"
	"time"
)

func TestListAdminUsersQueryUsesAuthUsersAsPrimarySource(t *testing.T) {
	t.Helper()

	if !strings.Contains(listAdminUsersSQL, "FROM auth.users AS auth_users") {
		t.Fatalf("expected users query to start from auth.users, query was:\n%s", listAdminUsersSQL)
	}
	if !strings.Contains(listAdminUsersSQL, "FULL OUTER JOIN profiles") {
		t.Fatalf("expected users query to retain profile-only rows too, query was:\n%s", listAdminUsersSQL)
	}
	if strings.Contains(listAdminUsersSQL, "COALESCE(auth_users.id::text, profiles.id)::uuid") {
		t.Fatalf("expected query to avoid mixing text and uuid in COALESCE, query was:\n%s", listAdminUsersSQL)
	}
}

func TestBuildAdminUserSummariesAggregatesRolesPerUser(t *testing.T) {
	t.Helper()

	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	rows := []adminUserRow{
		{
			Id:          "user-1",
			Email:       "admin@example.com",
			DisplayName: "Admin",
			Status:      "active",
			CreatedAt:   now,
			Role:        "admin",
		},
		{
			Id:          "user-1",
			Email:       "admin@example.com",
			DisplayName: "Admin",
			Status:      "active",
			CreatedAt:   now,
			Role:        "editor",
		},
		{
			Id:          "user-2",
			Email:       "new-user@example.com",
			DisplayName: "new-user",
			Status:      "active",
			CreatedAt:   now.Add(-time.Hour),
			Role:        "",
		},
	}

	got := buildAdminUserSummaries(rows)
	if len(got) != 2 {
		t.Fatalf("expected 2 users, got %d", len(got))
	}
	if got[0].Id != "user-1" || len(got[0].Roles) != 2 {
		t.Fatalf("expected first user to aggregate 2 roles, got %+v", got[0])
	}
	if got[1].Id != "user-2" {
		t.Fatalf("expected second user to be retained without profile-backed role rows, got %+v", got[1])
	}
	if got[1].Email != "new-user@example.com" {
		t.Fatalf("expected auth user email to be preserved, got %+v", got[1])
	}
}
