package patch

import (
	"fmt"
	"github.com/elimity-com/scim/schema"
	"testing"
)

func TestOperationValidator_ValidateAdd(t *testing.T) {
	// The goal this test is to cover Section 3.5.2.1 of RFC7644.
	// More info: https://tools.ietf.org/html/rfc7644#section-3.5.2.1

	// Some more indirect things are covered by these tests:
	// - If the target location does not exist, the attribute and value are added.
	// - If the target location specifies a complex attribute, a set of sub-attributes shall be specified in the "value"
	//   parameter.
	// - If the target location specifies an attribute that does not exist (has no value), the attribute is added with
	//   the new value.
	// - If the target location exists, the value is replaced.
	// - If the target location already contains the value specified, no changes SHOULD be made to the resource, and a
	//   success response should be returned.
	// Some things that are NOT covered:
	// - Unless other operations change the resource, this operation shall not change the modify timestamp of the
	//   resource.
	for i, test := range []struct {
		valid   string
		invalid string
	}{
		// The operation must contain a "value" member whose content specifies the value to be added.
		{
			valid:   `{"op":"add","path":"attr1","value":"value"}`,
			invalid: `{"op":"add","path":"attr1"}`,
		},

		// A URI prefix in the path.
		{
			valid:   `{"op":"add","path":"test:PatchEntity:attr1","value":"value"}`,
			invalid: `{"op":"add","path":"invalid:attr1","value":"value"}`,
		},
		{valid: `{"op":"add","path":"test:PatchExtension:attr1","value":"value"}`},

		// The value MAY be a quoted value, or it may be a JSON object containing the sub-attributes of the complex
		// attribute specified in the operation's "path".
		//
		// This is interpreted as:
		// > The value MUST contain a value with the data type of the attribute specified in the operation's "path".
		// The idea is that path can be either fine-grained or point to a whole object.
		// Thus value of "value" depends on what path points to.
		{
			valid:   `{"op":"add","path":"complex.attr1","value":"value"}`,
			invalid: `{"op":"add","path":"complex.attr1","value":{"attr1":"value"}}`,
		},
		{
			valid:   `{"op":"add","path":"complex","value":{"attr1":"value"}}`,
			invalid: `{"op":"add","path":"complex","value":"value"}`,
		},

		// If omitted, the target location is assumed to be the resource itself. The "value" parameter contains a
		// set of attributes to be added to the resource.
		{
			valid:   `{"op":"add","value":{"attr1":"value"}}`,
			invalid: `{"op":"add","value":"value"}`,
		},
		{invalid: `{"op":"add","value":{"invalid":"value"}}`},
		{invalid: `{"op":"add","value":{"invalid:attr1":"value"}}`},

		// If the target location specifies a multi-valued attribute, a new value is added to the attribute.
		{valid: `{"op":"add","value":{"multiValued":"value"}}`},

		// Example on page 36 (RFC7644, Section 3.5.2.1).
		{valid: `{"op":"add","path":"complexMultiValued","value":[{"attr1":"value"}]}`},
		{valid: `{"op":"add","path":"complexMultiValued","value":{"attr1":"value"}}`},

		// Example on page 37 (RFC7644, Section 3.5.2.1).
		{valid: `{"op":"add","value":{"attr1":"value","complexMultiValued":[{"attr1":"value"}]}}`},
		{valid: `{"op":"add","value":{"attr1":"value","complexMultiValued":[{"attr1":"value"}]}}`},

		// TODO: support value filters.
		// {valid: `{"op":"add","path":"complexMultiValued[attr1 eq \"value\"].attr1","value":"value"}`},
		// {valid: `{"op":"add","path":"test:PatchEntity:complexMultiValued[attr1 eq \"value\"].attr1","value":"value"}`},

		// Valid path, attribute not found.
		{invalid: `{"op":"add","path":"invalid","value":"value"}`},
		{invalid: `{"op":"add","path":"complex.invalid","value":"value"}`},

		// Sub-attributes in complex assignments.
		{valid: `{"op":"add","value":{"complex.attr1":"value"}}`},

		// Has no sub-attributes.
		{invalid: `{"op":"add","path":"attr1.invalid","value":"value"}`},

		// Invalid types.
		{invalid: `{"op":"add","path":"attr1","value":1}`},
		{invalid: `{"op":"add","path":"multiValued","value":1}`},
		{invalid: `{"op":"add","path":"multiValued","value":[1]}`},
		{invalid: `{"op":"add","path":"complex.attr1","value":1}`},
		{invalid: `{"op":"add","value":{"attr1":1}}`},
		{invalid: `{"op":"add","value":{"multiValued":1}}`},
		{invalid: `{"op":"add","value":{"multiValued":[1]}}`},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// valid
			if op := test.valid; op != "" {
				validator, err := NewValidator(op, patchSchema, patchSchemaExtension)
				if err != nil {
					t.Fatal(err)
				}
				if err := validator.Validate(); err != nil {
					t.Errorf("The following operatation should be an VALID add operation:\n(case %d): %s\n%v", i, op, err)
				}
			}
			// invalid
			if op := test.invalid; op != "" {
				validator, err := NewValidator(op, patchSchema, patchSchemaExtension)
				if err != nil {
					t.Fatal(err)
				}
				if err := validator.Validate(); err == nil {
					t.Errorf("The following operatation should be an INVALID add operation:\n(case %d): %s", i, op)
				}
			}
		})
	}
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
			fmt.Sprintf(`{"op":"invalid","path":"%s","value":"value"}`, test.pathFilter),
			schema.CoreUserSchema(), schema.ExtensionEnterpriseUser(),
		)
		if err != nil {
			t.Fatal(err)
		}
		attr, err := validator.getRefAttribute(validator.path.AttributePath)
		if err != nil {
			t.Fatal(err)
		}
		if name := attr.Name(); name != test.expectedAttrName {
			t.Errorf("expected %s, got %s", test.expectedAttrName, name)
		}
	}
}
