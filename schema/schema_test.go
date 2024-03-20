package schema

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/elimity-com/scim/optional"
)

var testSchema = Schema{
	ID:          "test-schema-id",
	Name:        optional.NewString("test"),
	Description: optional.String{},
	Attributes: []CoreAttribute{
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			Name:     "required",
			Required: true,
		})),
		SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
			MultiValued: true,
			Name:        "booleans",
			Required:    true,
		})),
		ComplexCoreAttribute(ComplexParams{
			MultiValued: true,
			Name:        "complex",
			SubAttributes: []SimpleParams{
				SimpleStringParams(StringParams{Name: "sub"}),
			},
		}),

		SimpleCoreAttribute(SimpleBinaryParams(BinaryParams{
			Name: "binary",
		})),
		SimpleCoreAttribute(SimpleDateTimeParams(DateTimeParams{
			Name: "dateTime",
		})),
		SimpleCoreAttribute(SimpleReferenceParams(ReferenceParams{
			Name: "reference",
		})),
		SimpleCoreAttribute(SimpleNumberParams(NumberParams{
			Name: "integer",
			Type: AttributeTypeInteger(),
		})),
		SimpleCoreAttribute(SimpleNumberParams(NumberParams{
			Name: "integerNumber",
			Type: AttributeTypeInteger(),
		})),
		SimpleCoreAttribute(SimpleNumberParams(NumberParams{
			Name: "decimal",
			Type: AttributeTypeDecimal(),
		})),
		SimpleCoreAttribute(SimpleNumberParams(NumberParams{
			Name: "decimalNumber",
			Type: AttributeTypeDecimal(),
		})),
	},
}

func TestInvalidAttributeName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("did not panic")
		}
	}()

	_ = Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []CoreAttribute{
			SimpleCoreAttribute(SimpleStringParams(StringParams{Name: "_Invalid"})),
		},
	}
}

func TestJSONMarshalling(t *testing.T) {
	expectedJSON, err := os.ReadFile("./testdata/schema_test.json")
	if err != nil {
		t.Fatal("failed to acquire test data")
	}

	actualJSON, err := testSchema.MarshalJSON()
	if err != nil {
		t.Fatal("failed to marshal schema into JSON")
	}

	normalizedActual, err := normalizeJSON(actualJSON)
	normalizedExpected, expectedErr := normalizeJSON(expectedJSON)
	if err != nil || expectedErr != nil {
		t.Errorf("failed to normalize test JSON")
		return
	}
	if normalizedActual != normalizedExpected {
		t.Errorf("schema output by MarshalJSON did not match the expected output."+
			"\nWant: %s\nGot:  %s", normalizedExpected, normalizedActual)
	}
}

func TestResourceInvalid(t *testing.T) {
	var resource interface{}
	if _, scimErr := testSchema.Validate(resource); scimErr == nil {
		t.Error("invalid resource expected")
	}
}

func TestValidValidation(t *testing.T) {
	for _, test := range []map[string]interface{}{
		{
			"schemas":  []interface{}{"test-schema-id"},
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"complex": []interface{}{
				map[string]interface{}{
					"sub": "present",
				},
			},
			"binary":        "ZXhhbXBsZQ==",
			"dateTime":      "2008-01-23T04:56:22Z",
			"integer":       11,
			"decimal":       -2.1e5,
			"integerNumber": json.Number("11"),
			"decimalNumber": json.Number("11.12"),
		},
	} {
		if _, scimErr := testSchema.Validate(test); scimErr != nil {
			t.Errorf("valid resource expected")
		}
	}
}

func TestValidationInvalid(t *testing.T) {
	tests := []struct {
		name     string
		resource map[string]interface{}
	}{
		{
			name: "missing required field",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"field":    "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "missing required multivalued field",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{},
			},
		},
		{
			name: "wrong type element of slice",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{"present"},
			},
		},
		{
			name: "duplicate names",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"Required": "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "wrong string type",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": true,
				"booleans": []interface{}{true},
			},
		},
		{
			name: "wrong complex type",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"complex":  "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "wrong complex element type",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"complex": []interface{}{
					"present",
				},
			},
		},
		{
			name: "duplicate complex element names",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"complex": []interface{}{
					map[string]interface{}{
						"sub": "present",
						"Sub": "present",
					},
				},
			},
		},
		{
			name: "wrong type complex element",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"complex": []interface{}{
					map[string]interface{}{
						"sub": true,
					},
				},
			},
		},
		{
			name: "invalid type binary",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"binary":   true,
			},
		},
		{
			name: "invalid type dateTime",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"dateTime": "04:56:22Z2008-01-23T",
			},
		},
		{
			name: "invalid type integer",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"integer":  1.1,
			},
		},
		{
			name: "invalid type decimal",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"test-schema-id"},
				"required": "present",
				"booleans": []interface{}{true},
				"decimal":  "1.1",
			},
		},
		{
			name: "invalid type integer (json.Number)",
			resource: map[string]interface{}{
				"schemas":       []interface{}{"test-schema-id"},
				"required":      "present",
				"booleans":      []interface{}{true},
				"integerNumber": json.Number("1.1"),
			},
		},
		{
			name: "invalid type decimal (json.Number)",
			resource: map[string]interface{}{
				"schemas":       []interface{}{"test-schema-id"},
				"required":      "present",
				"booleans":      []interface{}{true},
				"decimalNumber": json.Number("fail"),
			},
		},
		{
			name: "missing 'schemas' attribute",
			resource: map[string]interface{}{
				"required": "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "'schemas' attribute is not an array of strings",
			resource: map[string]interface{}{
				"schemas":  "test-schema-id",
				"required": "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "wrong 'schemas' name",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"wrong_schema_name"},
				"required": "present",
				"booleans": []interface{}{true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, scimErr := testSchema.Validate(tt.resource); scimErr == nil {
				t.Errorf("invalid resource expected")
			}
		})
	}
}

func normalizeJSON(rawJSON []byte) (string, error) {
	dataMap := map[string]interface{}{}

	// Ignoring errors since we know it is valid
	err := json.Unmarshal(rawJSON, &dataMap)
	if err != nil {
		return "", err
	}

	ret, err := json.Marshal(dataMap)

	return string(ret), err
}
