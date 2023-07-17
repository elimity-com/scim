package schema

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/elimity-com/scim/optional"
)

var testSchema = Schema{
	ID:          "empty",
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
	expectedJSON, err := ioutil.ReadFile("./testdata/schema_test.json")
	if err != nil {
		t.Errorf("failed to acquire test data")
		return
	}

	actualJSON, err := testSchema.MarshalJSON()
	if err != nil {
		t.Errorf("failed to marshal schema into JSON")
		return
	}

	normalizedActual, err := normalizeJSON(actualJSON)
	normalizedExpected, expectedErr := normalizeJSON(expectedJSON)
	if err != nil || expectedErr != nil {
		t.Errorf("failed to normalize test JSON")
		return
	}

	if normalizedActual != normalizedExpected {
		t.Errorf("schema output by MarshalJSON did not match the expected output. want %s, got %s", normalizedExpected, normalizedActual)
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
	for _, test := range []map[string]interface{}{
		{ // missing required field
			"field": "present",
			"booleans": []interface{}{
				true,
			},
		},
		{ // missing required multivalued field
			"required": "present",
			"booleans": []interface{}{},
		},
		{ // wrong type element of slice
			"required": "present",
			"booleans": []interface{}{
				"present",
			},
		},
		{ // duplicate names
			"required": "present",
			"Required": "present",
			"booleans": []interface{}{
				true,
			},
		},
		{ // wrong string type
			"required": true,
			"booleans": []interface{}{
				true,
			},
		},
		{ // wrong complex type
			"required": "present",
			"complex":  "present",
			"booleans": []interface{}{
				true,
			},
		},
		{ // wrong complex element type
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"complex": []interface{}{
				"present",
			},
		},
		{ // duplicate complex element names
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"complex": []interface{}{
				map[string]interface{}{
					"sub": "present",
					"Sub": "present",
				},
			},
		},
		{ // wrong type complex element
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"complex": []interface{}{
				map[string]interface{}{
					"sub": true,
				},
			},
		},
		{ // invalid type binary
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"binary": true,
		},
		{ // invalid type dateTime
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"dateTime": "04:56:22Z2008-01-23T",
		},
		{ // invalid type integer
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"integer": 1.1,
		},
		{ // invalid type decimal
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"decimal": "1.1",
		},
		{ // invalid type integer (json.Number)
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"integerNumber": json.Number("1.1"),
		},
		{ // invalid type decimal (json.Number)
			"required": "present",
			"booleans": []interface{}{
				true,
			},
			"decimalNumber": json.Number("fail"),
		},
	} {
		if _, scimErr := testSchema.Validate(test); scimErr == nil {
			t.Errorf("invalid resource expected")
		}
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
