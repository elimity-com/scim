package patch_test

import (
	"fmt"
	"github.com/elimity-com/scim/internal/patch"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"testing"
)

var (
	patchSchema = schema.Schema{
		ID:          "test:PatchEntity",
		Name:        optional.NewString("Patch"),
		Description: optional.NewString("Patch Entity"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "attr1",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:        "multiValued",
				MultiValued: true,
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name: "complex",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "attr1",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "attr2",
					}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "complexMultiValued",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "attr1",
					}),
				},
			}),
		},
	}
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
		{
			// The operation must contain a "value" member whose content specifies the value to be added.
			valid:   `{"op":"add","path":"attr1","value":"value"}`,
			invalid: `{"op":"add","path":"attr1"}`,
		},
		{
			// The value MAY be a quoted value, or it may be a JSON object containing the sub-attributes of the complex
			// attribute specified in the operation's "path".
			//
			// This is interpreted as:
			// > The value MUST contain a value with the data type of the attribute specified in the operation's "path".
			// The idea is that path can be either fine-grained or point to a whole object.
			// Thus value of "value" depends on what path points to.
			valid:   `{"op":"add","path":"complex.attr1","value":"value"}`,
			invalid: `{"op":"add","path":"complex.attr1","value":{"attr1":"value"}}`,
		},
		{
			// see previous case...
			valid:   `{"op":"add","path":"complex","value":{"attr1":"value"}}`,
			invalid: `{"op":"add","path":"complex","value":"value"}`,
		},
		{
			// If omitted, the target location is assumed to be the resource itself. The "value" parameter contains a
			// set of attributes to be added to the resource.
			valid:   `{"op":"add","value":{"attr1":"value"}}`,
			invalid: `{"op":"add","value":"value"}`,
		},
		{
			// If the target location specifies a multi-valued attribute, a new value is added to the attribute.
			valid: `{"op":"add","value":{"multiValued":"value"}}`,
		},
		{
			// Example on page 36 (RFC7644, Section 3.5.2.1).
			valid: `{"op":"add","path":"complexMultiValued","value":[{"attr1":"value"}]}`,
		},
		{
			// Example on page 37 (RFC7644, Section 3.5.2.1).
			valid: `{"op":"add","value":{"attr1":"value","complexMultiValued":[{"attr1":"value"}]}}`,
		},

		// Invalid types.
		{invalid: `{"op":"add","path":"attr1","value":1}`},
		{invalid: `{"op":"add","path":"multiValued","value":1}`},
		{invalid: `{"op":"add","path":"complex.attr1","value":1}`},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// valid
			if op := test.valid; op != "" {
				validator, err := patch.NewPathValidator(op, patchSchema)
				if err != nil {
					t.Fatal(err)
				}
				if err := validator.Validate(); err != nil {
					t.Errorf("The following operatation should be an VALID add operation:\n(case %d): %s\n%v", i, op, err)
				}
			}
			// invalid
			if op := test.invalid; op != "" {
				validator, err := patch.NewPathValidator(op, patchSchema)
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
