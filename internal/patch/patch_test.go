package patch

import (
	"fmt"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"testing"
)

func TestNewPathValidator(t *testing.T) {
	t.Run("Invalid JSON", func(t *testing.T) {
		// The quotes in the value filter are not escaped.
		op := `{"op":"add","path":"complexMultiValued[attr1 eq "value"].attr1","value":"value"}`
		if _, err := NewValidator(op, patchSchema); err == nil {
			t.Error("expected JSON error, got none")
		}
	})
	t.Run("Invalid Op", func(t *testing.T) {
		// "op" must be one of "add", "remove", or "replace".
		op := `{"op":"invalid","path":"attr1","value":"value"}`
		validator, _ := NewValidator(op, patchSchema)
		if _, err := validator.Validate(); err == nil {
			t.Errorf("expected error, got none")
		}
	})
	t.Run("Invalid Attribute", func(t *testing.T) {
		// "invalid pr" is not a valid Path filter.
		// This error will be caught by the Path filter validator.
		op := `{"op":"add","path":"invalid pr","value":"value"}`
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
		validator, err := NewValidator(
			fmt.Sprintf(`{"op":"invalid","path":%q,"value":"value"}`, test.pathFilter),
			schema.CoreUserSchema(), schema.ExtensionEnterpriseUser(),
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

	validator, _ := NewValidator(
		`{"op":"invalid","path":"complex","value":"value"}`,
		schema.CoreUserSchema(), schema.ExtensionEnterpriseUser(),
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
		validator, err := NewValidator(
			fmt.Sprintf(`{"op":"invalid","path":%q,"value":"value"}`, test.attributeName),
			schema.CoreUserSchema(), schema.ExtensionEnterpriseUser(),
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
