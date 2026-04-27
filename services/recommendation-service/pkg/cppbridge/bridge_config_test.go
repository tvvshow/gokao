package cppbridge

import (
	"path/filepath"
	"testing"
)

func TestResolveBridgeDataFilesUsesRepoFixtures(t *testing.T) {
	files, err := resolveBridgeDataFiles(BridgeConfig{})
	if err != nil {
		t.Fatalf("resolveBridgeDataFiles returned error: %v", err)
	}

	if filepath.Base(files.universities) != "universities.csv" {
		t.Fatalf("unexpected universities file: %s", files.universities)
	}
	if filepath.Base(files.majors) != "majors.csv" {
		t.Fatalf("unexpected majors file: %s", files.majors)
	}
	if filepath.Base(files.historical) != "historical_data.csv" {
		t.Fatalf("unexpected historical file: %s", files.historical)
	}
}
