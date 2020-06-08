package schema

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestParseJSONSchema(t *testing.T) {
	for _, test := range []string{
		"user_schema.json",
		"group_schema.json",
		"enterprise_user_schema.json",
	} {
		expectedJSON, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", test))
		if err != nil {
			t.Errorf("Failed to acquire test data")
			return
		}

		schema, err := ParseJSONSchema(expectedJSON)
		if err != nil {
			fmt.Println(err)
			return
		}

		actualJSON, err := schema.MarshalJSON()
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
