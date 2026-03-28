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
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			Name:       "requiredReadOnly",
			Required:   true,
			Mutability: AttributeMutabilityReadOnly(),
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

func TestJSONUnmarshalling(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		originalJSON, err := testSchema.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		var got Schema
		if err := json.Unmarshal(originalJSON, &got); err != nil {
			t.Fatal(err)
		}

		gotJSON, err := got.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		normalizedOriginal, err := normalizeJSON(originalJSON)
		if err != nil {
			t.Fatal(err)
		}
		normalizedGot, err := normalizeJSON(gotJSON)
		if err != nil {
			t.Fatal(err)
		}

		if normalizedOriginal != normalizedGot {
			t.Errorf("round trip mismatch.\nWant: %s\nGot:  %s", normalizedOriginal, normalizedGot)
		}
	})

	t.Run("user schema round trip", func(t *testing.T) {
		originalJSON, err := CoreUserSchema().MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		var got Schema
		if err := json.Unmarshal(originalJSON, &got); err != nil {
			t.Fatal(err)
		}

		gotJSON, err := got.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		normalizedOriginal, err := normalizeJSON(originalJSON)
		if err != nil {
			t.Fatal(err)
		}
		normalizedGot, err := normalizeJSON(gotJSON)
		if err != nil {
			t.Fatal(err)
		}

		if normalizedOriginal != normalizedGot {
			t.Errorf("round trip mismatch.\nWant: %s\nGot:  %s", normalizedOriginal, normalizedGot)
		}
	})

	t.Run("group schema round trip", func(t *testing.T) {
		originalJSON, err := CoreGroupSchema().MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		var got Schema
		if err := json.Unmarshal(originalJSON, &got); err != nil {
			t.Fatal(err)
		}

		gotJSON, err := got.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		normalizedOriginal, err := normalizeJSON(originalJSON)
		if err != nil {
			t.Fatal(err)
		}
		normalizedGot, err := normalizeJSON(gotJSON)
		if err != nil {
			t.Fatal(err)
		}

		if normalizedOriginal != normalizedGot {
			t.Errorf("round trip mismatch.\nWant: %s\nGot:  %s", normalizedOriginal, normalizedGot)
		}
	})

	t.Run("from file", func(t *testing.T) {
		data, err := os.ReadFile("./testdata/schema_test.json")
		if err != nil {
			t.Fatal(err)
		}

		var got Schema
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}

		if got.ID != "test-schema-id" {
			t.Errorf("ID: want %q, got %q", "test-schema-id", got.ID)
		}
		if got.Name.Value() != "test" {
			t.Errorf("Name: want %q, got %q", "test", got.Name.Value())
		}
		if len(got.Attributes) != 11 {
			t.Errorf("Attributes: want 11, got %d", len(got.Attributes))
		}
	})

	t.Run("enterprise user extension round trip", func(t *testing.T) {
		originalJSON, err := ExtensionEnterpriseUser().MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		var got Schema
		if err := json.Unmarshal(originalJSON, &got); err != nil {
			t.Fatal(err)
		}

		gotJSON, err := got.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		normalizedOriginal, err := normalizeJSON(originalJSON)
		if err != nil {
			t.Fatal(err)
		}
		normalizedGot, err := normalizeJSON(gotJSON)
		if err != nil {
			t.Fatal(err)
		}

		if normalizedOriginal != normalizedGot {
			t.Errorf("round trip mismatch.\nWant: %s\nGot:  %s", normalizedOriginal, normalizedGot)
		}
	})

	t.Run("custom schema round trip", func(t *testing.T) {
		custom := Schema{
			ID:          "urn:example:custom:1.0:Device",
			Name:        optional.NewString("Device"),
			Description: optional.NewString("A custom device resource"),
			Attributes: []CoreAttribute{
				SimpleCoreAttribute(SimpleStringParams(StringParams{
					Name:       "serialNumber",
					Required:   true,
					Uniqueness: AttributeUniquenessServer(),
					CaseExact:  true,
				})),
				SimpleCoreAttribute(SimpleBooleanParams(BooleanParams{
					Name:       "active",
					Mutability: AttributeMutabilityReadWrite(),
				})),
				SimpleCoreAttribute(SimpleNumberParams(NumberParams{
					Name: "firmwareVersion",
					Type: AttributeTypeDecimal(),
				})),
				SimpleCoreAttribute(SimpleDateTimeParams(DateTimeParams{
					Name:       "lastSeen",
					Mutability: AttributeMutabilityReadOnly(),
					Returned:   AttributeReturnedAlways(),
				})),
				SimpleCoreAttribute(SimpleReferenceParams(ReferenceParams{
					Name:           "owner",
					ReferenceTypes: []AttributeReferenceType{AttributeReferenceTypeExternal, AttributeReferenceTypeURI},
				})),
				SimpleCoreAttribute(SimpleStringParams(StringParams{
					Name:            "status",
					CanonicalValues: []string{"online", "offline", "maintenance"},
				})),
				ComplexCoreAttribute(ComplexParams{
					Name:        "location",
					MultiValued: false,
					SubAttributes: []SimpleParams{
						SimpleStringParams(StringParams{Name: "building"}),
						SimpleNumberParams(NumberParams{
							Name: "floor",
							Type: AttributeTypeInteger(),
						}),
					},
				}),
			},
		}

		originalJSON, err := custom.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		var got Schema
		if err := json.Unmarshal(originalJSON, &got); err != nil {
			t.Fatal(err)
		}

		if got.ID != custom.ID {
			t.Errorf("ID: want %q, got %q", custom.ID, got.ID)
		}
		if got.Name.Value() != custom.Name.Value() {
			t.Errorf("Name: want %q, got %q", custom.Name.Value(), got.Name.Value())
		}
		if got.Description.Value() != custom.Description.Value() {
			t.Errorf("Description: want %q, got %q", custom.Description.Value(), got.Description.Value())
		}

		gotJSON, err := got.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		normalizedOriginal, err := normalizeJSON(originalJSON)
		if err != nil {
			t.Fatal(err)
		}
		normalizedGot, err := normalizeJSON(gotJSON)
		if err != nil {
			t.Fatal(err)
		}

		if normalizedOriginal != normalizedGot {
			t.Errorf("round trip mismatch.\nWant: %s\nGot:  %s", normalizedOriginal, normalizedGot)
		}
	})

	t.Run("unknown type", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"a","type":"unknown"}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for unknown attribute type")
		}
	})

	t.Run("unknown mutability", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"a","type":"string","mutability":"unknown"}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for unknown mutability")
		}
	})

	t.Run("unknown returned", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"a","type":"string","returned":"unknown"}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for unknown returned")
		}
	})

	t.Run("unknown uniqueness", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"a","type":"string","uniqueness":"unknown"}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for unknown uniqueness")
		}
	})

	t.Run("invalid attribute name", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"_invalid","type":"string"}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for invalid attribute name")
		}
	})

	t.Run("invalid sub-attribute name", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"a","type":"complex","subAttributes":[{"name":"1bad","type":"string"}]}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for invalid sub-attribute name")
		}
	})

	t.Run("duplicate sub-attribute names", func(t *testing.T) {
		data := []byte(`{"id":"x","attributes":[{"name":"a","type":"complex","subAttributes":[{"name":"b","type":"string"},{"name":"b","type":"string"}]}]}`)
		var got Schema
		if err := json.Unmarshal(data, &got); err == nil {
			t.Error("expected error for duplicate sub-attribute names")
		}
	})
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
			"schemas":          []interface{}{"test-schema-id"},
			"required":         "present",
			"requiredReadOnly": "ignoreme",
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
			name: "missing schemas attribute",
			resource: map[string]interface{}{
				"required": "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "schemas attribute is not an array",
			resource: map[string]interface{}{
				"schemas":  "test-schema-id",
				"required": "present",
				"booleans": []interface{}{true},
			},
		},
		{
			name: "wrong schema ID",
			resource: map[string]interface{}{
				"schemas":  []interface{}{"wrong-schema-id"},
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
