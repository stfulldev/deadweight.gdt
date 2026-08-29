package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestCanonicalSchemaMatchesVersionOneModel(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "schema", "deadweight.gdt.schema.json"))
	if err != nil {
		t.Fatalf("ReadFile(schema) error = %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" || schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("schema root = %#v", schema)
	}
	if !reflect.DeepEqual(stringSlice(t, schema["required"]), []string{"version"}) {
		t.Fatalf("required = %#v", schema["required"])
	}

	properties := objectMap(t, schema["properties"])
	wantProperties := []string{"budgets", "fail_on_partial", "preset", "profile", "profiles", "version"}
	if got := sortedKeys(properties); !reflect.DeepEqual(got, wantProperties) {
		t.Fatalf("properties = %#v, want %#v", got, wantProperties)
	}
	version := objectMap(t, properties["version"])
	if version["type"] != "integer" || version["const"] != float64(CurrentVersion) {
		t.Fatalf("version schema = %#v", version)
	}
	not := objectMap(t, schema["not"])
	if !reflect.DeepEqual(stringSlice(t, not["required"]), []string{"preset", "profile"}) {
		t.Fatalf("selector exclusion = %#v", not)
	}

	definitions := objectMap(t, schema["$defs"])
	id := objectMap(t, definitions["id"])
	if id["type"] != "string" || id["pattern"] != IDPattern {
		t.Fatalf("id schema = %#v", id)
	}
	budgets := objectMap(t, definitions["budgets"])
	assertStrictObject(t, "budgets", budgets)
	budgetProperties := objectMap(t, budgets["properties"])
	wantMetrics := make([]string, 0, len(metrics.OrderedNames()))
	for _, name := range metrics.OrderedNames() {
		wantMetrics = append(wantMetrics, string(name))
	}
	sort.Strings(wantMetrics)
	if got := sortedKeys(budgetProperties); !reflect.DeepEqual(got, wantMetrics) {
		t.Fatalf("budget properties = %#v, want %#v", got, wantMetrics)
	}
	for _, name := range wantMetrics {
		property := objectMap(t, budgetProperties[name])
		if property["type"] != "integer" || property["minimum"] != float64(0) {
			t.Errorf("budget %q schema = %#v", name, property)
		}
	}

	profiles := objectMap(t, properties["profiles"])
	assertStrictObject(t, "profiles", profiles)
	patterns := objectMap(t, profiles["patternProperties"])
	if got := sortedKeys(patterns); !reflect.DeepEqual(got, []string{IDPattern}) {
		t.Fatalf("profile patterns = %#v", got)
	}
	profile := objectMap(t, definitions["profile"])
	assertStrictObject(t, "profile", profile)
	profileProperties := objectMap(t, profile["properties"])
	wantProfileFields := []string{"budgets", "description", "extends", "name", "platform", "quality", "renderer", "target_fps"}
	if got := sortedKeys(profileProperties); !reflect.DeepEqual(got, wantProfileFields) {
		t.Fatalf("profile properties = %#v, want %#v", got, wantProfileFields)
	}
	if got := stringSlice(t, objectMap(t, profileProperties["renderer"])["enum"]); !reflect.DeepEqual(got, RendererIDs()) {
		t.Fatalf("renderer enum = %#v", got)
	}
	if got := stringSlice(t, objectMap(t, profileProperties["quality"])["enum"]); !reflect.DeepEqual(got, QualityIDs()) {
		t.Fatalf("quality enum = %#v", got)
	}
	if objectMap(t, profileProperties["target_fps"])["minimum"] != float64(0) || objectMap(t, profileProperties["platform"])["minLength"] != float64(1) {
		t.Fatalf("profile numeric/string constraints = %#v", profileProperties)
	}
}

func assertStrictObject(t *testing.T, name string, object map[string]any) {
	t.Helper()
	if object["type"] != "object" || object["additionalProperties"] != false {
		t.Fatalf("%s is not a strict object: %#v", name, object)
	}
}

func objectMap(t *testing.T, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value = %T %#v, want object", value, value)
	}

	return object
}

func stringSlice(t *testing.T, value any) []string {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("value = %T %#v, want array", value, value)
	}
	result := make([]string, len(items))
	for index, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("item %d = %T %#v, want string", index, item, item)
		}
		result[index] = text
	}

	return result
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
