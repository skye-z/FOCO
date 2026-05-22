package interactive

import "testing"

func TestMapAdminUnitSummaryRowsPreservesIDs(t *testing.T) {
	rows := []adminUnitSummaryRow{
		{ID: "unit-1", Title: "Unit 1", ExamID: "exam-1", VersionID: "v1"},
		{ID: "unit-2", Title: "Unit 2", ExamID: "exam-1", VersionID: "v2"},
	}

	units := mapAdminUnitSummaryRows(rows)
	if len(units) != 2 {
		t.Fatalf("expected 2 mapped units, got %d", len(units))
	}
	if units[0].ID != "unit-1" || units[1].ID != "unit-2" {
		t.Fatalf("expected IDs to be preserved, got %+v", units)
	}
}

func TestMapAdminVersionSummaryRowsPreservesVersionIDs(t *testing.T) {
	rows := []adminVersionSummaryRow{
		{VersionID: "version-1", UnitID: "unit-1", VersionNo: 2, Status: "draft", UpdatedAt: "2026-05-21T12:00:00Z"},
		{VersionID: "version-2", UnitID: "unit-1", VersionNo: 1, Status: "published", UpdatedAt: "2026-05-20T12:00:00Z"},
	}

	versions := mapAdminVersionSummaryRows(rows)
	if len(versions) != 2 {
		t.Fatalf("expected 2 mapped versions, got %d", len(versions))
	}
	if versions[0].VersionID != "version-1" || versions[1].VersionID != "version-2" {
		t.Fatalf("expected version IDs to be preserved, got %+v", versions)
	}
}

func TestBuildInteractiveUnitMetadataJSON(t *testing.T) {
	if got := buildInteractiveUnitMetadataJSON(" "); got != "{}" {
		t.Fatalf("expected empty metadata json, got %s", got)
	}

	if got := buildInteractiveUnitMetadataJSON("  TVM Lab  "); got != `{"title":"TVM Lab"}` {
		t.Fatalf("expected trimmed title metadata json, got %s", got)
	}
}
