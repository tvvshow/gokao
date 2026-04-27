package services

import (
	"strings"
	"testing"
)

func TestDefaultMigrationDefinitionsUseAdmissionDataTable(t *testing.T) {
	migrations := defaultMigrationDefinitions()

	var admissionMigration *migrationDefinition
	var indexMigration *migrationDefinition
	for i := range migrations {
		switch migrations[i].Version {
		case "003":
			admissionMigration = &migrations[i]
		case "004":
			indexMigration = &migrations[i]
		}
	}

	if admissionMigration == nil {
		t.Fatal("missing migration version 003")
	}
	if !strings.Contains(admissionMigration.SQL, "CREATE TABLE IF NOT EXISTS admission_data") {
		t.Fatalf("expected admission_data table migration, got %s", admissionMigration.SQL)
	}
	if strings.Contains(admissionMigration.SQL, "CREATE TABLE IF NOT EXISTS admissions") {
		t.Fatalf("unexpected legacy admissions table migration: %s", admissionMigration.SQL)
	}
	if admissionMigration.RollbackSQL != "DROP TABLE IF EXISTS admission_data;" {
		t.Fatalf("unexpected rollback SQL: %s", admissionMigration.RollbackSQL)
	}

	if indexMigration == nil {
		t.Fatal("missing migration version 004")
	}
	if !strings.Contains(indexMigration.SQL, "idx_admission_data_year ON admission_data(year)") {
		t.Fatalf("expected admission_data index migration, got %s", indexMigration.SQL)
	}
	if strings.Contains(indexMigration.SQL, "idx_admissions_year ON admissions(year)") {
		t.Fatalf("unexpected legacy admissions index migration: %s", indexMigration.SQL)
	}
	if strings.Contains(indexMigration.RollbackSQL, "idx_admissions_year") {
		t.Fatalf("unexpected legacy admissions rollback SQL: %s", indexMigration.RollbackSQL)
	}
}

func TestRollbackStatementsStayInSyncWithMigrationDefinitions(t *testing.T) {
	rollbacks := rollbackStatements()
	definitions := defaultMigrationDefinitions()

	if len(rollbacks) != len(definitions) {
		t.Fatalf("rollback count mismatch: got %d want %d", len(rollbacks), len(definitions))
	}

	for _, migration := range definitions {
		if got := rollbacks[migration.Version]; got != migration.RollbackSQL {
			t.Fatalf("rollback mismatch for %s: got %q want %q", migration.Version, got, migration.RollbackSQL)
		}
	}
}

