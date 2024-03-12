package schema

import (
	"fmt"
	"os"
	"testing"
)

func TestDefaultSchemas(t *testing.T) {
	for _, test := range []struct {
		file   string
		schema Schema
	}{
		{
			file:   "user_schema.json",
			schema: CoreUserSchema(),
		},
		{
			file:   "group_schema.json",
			schema: CoreGroupSchema(),
		},
		{
			file:   "enterprise_user_schema.json",
			schema: ExtensionEnterpriseUser(),
		},
	} {
		expectedJSON, err := os.ReadFile(fmt.Sprintf("./testdata/%s", test.file))
		if err != nil {
			t.Errorf("Failed to acquire test data")
			return
		}

		actualJSON, err := test.schema.MarshalJSON()
		if err != nil {
			t.Errorf("Failed to marshal schema into JSON")
			return
		}

		normalizedActual, err := normalizeJSON(actualJSON)
		normalizedExpected, expectedErr := normalizeJSON(expectedJSON)
		if err != nil || expectedErr != nil {
			t.Errorf("Failed to normalize test JSON")
			return
		}

		if normalizedActual != normalizedExpected {
			t.Errorf("Schema output by MarshalJSON did not match the expected output. Want %s, Got %s", normalizedExpected, normalizedActual)
		}
	}
}
