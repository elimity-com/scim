package patch

import (
	"encoding/json"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"testing"
)

func TestNewPathValidator(t *testing.T) {
	t.Run("Valid Integer", func(t *testing.T) {
		for _, op := range []map[string]interface{}{
			{"op": "add", "path": "attr2", "value": 1234},
			{"op": "add", "path": "attr2", "value": "1234"},
		} {
			operation, _ := json.Marshal(op)
			validator, err := NewValidator(operation, patchSchema)
			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
			schema.SetAllowStringValues(true)
			defer schema.SetAllowStringValues(false)
			v, err := validator.Validate()
			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
			n, ok := v.(int64)
			if !ok {
				t.Fatalf("unexpected type, got %T", v)
			}
			if n != 1234 {
				t.Fatalf("unexpected integer, got %d", n)
			}
		}
	})

	t.Run("Valid Float", func(t *testing.T) {
		for _, op := range []map[string]interface{}{
			{"op": "add", "path": "attr3", "value": 12.34},
			{"op": "add", "path": "attr3", "value": "12.34"},
		} {
			operation, _ := json.Marshal(op)
			validator, err := NewValidator(operation, patchSchema)
			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
			schema.SetAllowStringValues(true)
			defer schema.SetAllowStringValues(false)
			v, err := validator.Validate()
			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
			n, ok := v.(float64)
			if !ok {
				t.Fatalf("unexpected type, got %T", v)
			}
			if n != 12.34 {
				t.Fatalf("unexpected integer, got %f", n)
			}
		}
	})

	t.Run("Valid Booleans", func(t *testing.T) {
		tests := []struct {
			op       map[string]interface{}
			expected bool
		}{
			{map[string]interface{}{"op": "add", "path": "attr4", "value": true}, true},
			{map[string]interface{}{"op": "add", "path": "attr4", "value": "True"}, true},
			{map[string]interface{}{"op": "add", "path": "attr4", "value": false}, false},
			{map[string]interface{}{"op": "add", "path": "attr4", "value": "False"}, false},
		}
		for _, tc := range tests {
			operation, _ := json.Marshal(tc.op)
			validator, err := NewValidator(operation, patchSchema)
			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
			schema.SetAllowStringValues(true)
			defer schema.SetAllowStringValues(false)
			v, err := validator.Validate()
			if err != nil {
				t.Fatalf("unexpected error, got %v", err)
			}
			b, ok := v.(bool)
			if !ok {
				t.Fatalf("unexpected type, got %T", v)
			}
			if b != tc.expected {
				t.Fatalf("unexpected integer, got %v", b)
			}
		}
	})
	t.Run("Invalid Op", func(t *testing.T) {
		// "op" must be one of "add", "remove", or "replace".
		op, _ := json.Marshal(map[string]interface{}{
			"op":    "invalid",
			"path":  "attr1",
			"value": "value",
		})
		validator, _ := NewValidator(op, patchSchema)
		if _, err := validator.Validate(); err == nil {
			t.Errorf("expected error, got none")
		}
	})
	t.Run("Invalid Attribute", func(t *testing.T) {
		// "invalid pr" is not a valid path filter.
		// This error will be caught by the path filter validator.
		op, _ := json.Marshal(map[string]interface{}{
			"op":    "add",
			"path":  "invalid pr",
			"value": "value",
		})
		if _, err := NewValidator(op, patchSchema); err == nil {
			t.Error("expected JSON error, got none")
		}
	})
}

func TestOperationValidator_getRefAttribute(t *testing.T) {
	for _, test := range []struct {
		pathFilter       string
		expectedAttrName string
	}{
		{`userName`, `userName`},
		{`name.givenName`, `givenName`},
		{`urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber`, `employeeNumber`},
	} {
		op, _ := json.Marshal(map[string]interface{}{
			"op":    "add",
			"path":  test.pathFilter,
			"value": "value",
		})
		validator, err := NewValidator(
			op,
			schema.CoreUserSchema(),
			schema.ExtensionEnterpriseUser(),
		)
		if err != nil {
			t.Fatal(err)
		}
		attr, err := validator.getRefAttribute(validator.Path.AttributePath)
		if err != nil {
			t.Fatal(err)
		}
		if name := attr.Name(); name != test.expectedAttrName {
			t.Errorf("expected %s, got %s", test.expectedAttrName, name)
		}
	}

	op, _ := json.Marshal(map[string]interface{}{
		"op":    "invalid",
		"path":  "complex",
		"value": "value",
	})
	validator, _ := NewValidator(
		op,
		schema.CoreUserSchema(),
		schema.ExtensionEnterpriseUser(),
	)
	if _, err := validator.getRefAttribute(filter.AttributePath{
		AttributeName: "invalid",
	}); err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestOperationValidator_getRefSubAttribute(t *testing.T) {
	for _, test := range []struct {
		attributeName    string
		subAttributeName string
	}{
		{`name`, `givenName`},
		{`groups`, `display`},
	} {
		op, _ := json.Marshal(map[string]interface{}{
			"op":    "invalid",
			"path":  test.attributeName,
			"value": "value",
		})
		validator, err := NewValidator(
			op,
			schema.CoreUserSchema(),
			schema.ExtensionEnterpriseUser(),
		)
		if err != nil {
			t.Fatal(err)
		}
		refAttr, ok := schema.CoreUserSchema().Attributes.ContainsAttribute(test.attributeName)
		if !ok {
			t.Fatal()
		}
		attr, err := validator.getRefSubAttribute(&refAttr, test.subAttributeName)
		if err != nil {
			t.Fatal(err)
		}
		if name := attr.Name(); name != test.subAttributeName {
			t.Errorf("expected %s, got %s", test.subAttributeName, name)
		}
	}
}
