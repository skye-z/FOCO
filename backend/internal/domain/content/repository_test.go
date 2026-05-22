package content

import (
	"strings"
	"testing"
	"time"
)

func TestBuildQuestionPublishSyncUpdateSQLCastsUpdatedAt(t *testing.T) {
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	versionID := "version-1"
	sql, args := buildQuestionPublishSyncUpdateSQL([]ContentPackageQuestion{
		{
			Id:                        "question-1",
			CurrentPublishedVersionId: &versionID,
			Status:                    "published",
			UpdatedAt:                 now,
		},
	})

	if !strings.Contains(sql, "$4::timestamptz") {
		t.Fatalf("expected updated_at to be cast as timestamptz, got SQL:\n%s", sql)
	}
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d", len(args))
	}
}
