package database

import (
	"reflect"
	"testing"

	"github.com/oktetopython/gaokao/services/data-service/internal/models"
)

func TestMigratableModelsIncludeStatisticsTables(t *testing.T) {
	found := map[reflect.Type]bool{}
	for _, model := range migratableModels() {
		typ := reflect.TypeOf(model)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		found[typ] = true
	}

	required := []reflect.Type{
		reflect.TypeOf(models.University{}),
		reflect.TypeOf(models.Major{}),
		reflect.TypeOf(models.AdmissionData{}),
		reflect.TypeOf(models.UniversityStatistics{}),
		reflect.TypeOf(models.MajorStatistics{}),
	}

	for _, typ := range required {
		if !found[typ] {
			t.Fatalf("migratable models missing %s", typ.Name())
		}
	}
}

