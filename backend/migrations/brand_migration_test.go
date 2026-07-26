package migrations

import (
	"strings"
	"testing"
)

func TestLegacySiteBrandMigrationIsNarrowAndIdempotent(t *testing.T) {
	content, err := FS.ReadFile("193_normalize_legacy_site_brand.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(content))
	for _, required := range []string{
		"where key = 'site_name'",
		"lower(btrim(value)) = 'sub2api'",
		"where key = 'site_subtitle'",
		"lower(btrim(value)) = 'subscription to api conversion platform'",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing narrow guard %q", required)
		}
	}
	if strings.Contains(sql, "delete from") || strings.Contains(sql, "drop table") {
		t.Fatal("brand migration must not delete tables or rows")
	}
}
