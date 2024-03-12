package patch

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestOperationValidator_ValidateUpdate(t *testing.T) {
	// The goal this test is to cover Section 3.5.2.1/3 of RFC7644.
	// More info: https://tools.ietf.org/html/rfc7644#section-3.5.2.1
	// More info: https://tools.ietf.org/html/rfc7644#section-3.5.2.3
	// (!) Both "add" and "replace" behave the same in regard to valid paths.

	// Some more indirect things are covered by these tests:
	// - If the target location does not exist, the attribute and value are added.
	// - If the target location specifies a complex attribute, a set of sub-attributes shall be specified in the "value"
	//   parameter.
	// - If the target location specifies an attribute that does not exist (has no value), the attribute is added with
	//   the new value.
	// - If the target location exists, the value is replaced.
	// - If the target location already contains the value specified, no changes SHOULD be made to the resource, and a
	//   success response should be returned.
	//
	// Some things that are NOT covered:
	// - Unless other operations change the resource, this operation shall not change the modify timestamp of the
	//   resource.
	for i, test := range []struct {
		valid   map[string]interface{}
		invalid map[string]interface{}
	}{
		// The operation must contain a "value" member whose content specifies the value to be added.
		{
			valid:   map[string]interface{}{"op": "add", "path": "attr1", "value": "value"},
			invalid: map[string]interface{}{"op": "add", "path": "attr1"},
		},

		// A URI prefix in the path.
		{
			valid:   map[string]interface{}{"op": "add", "path": "test:PatchEntity:attr1", "value": "value"},
			invalid: map[string]interface{}{"op": "add", "path": "invalid:attr1", "value": "value"},
		},
		{valid: map[string]interface{}{"op": "add", "path": "test:PatchExtension:attr1", "value": "value"}},

		// The value MAY be a quoted value, or it may be a JSON object containing the sub-attributes of the complex
		// attribute specified in the operation's "path".
		//
		// This is interpreted as:
		// > The value MUST contain a value with the data type of the attribute specified in the operation's "path".
		// The idea is that path can be either fine-grained or point to a whole object.
		// Thus value of "value" depends on what path points to.
		{
			valid:   map[string]interface{}{"op": "add", "path": "complex.attr1", "value": "value"},
			invalid: map[string]interface{}{"op": "add", "path": "complex.attr1", "value": map[string]interface{}{"attr1": "value"}},
		},
		{
			valid:   map[string]interface{}{"op": "add", "path": "complex", "value": map[string]interface{}{"attr1": "value"}},
			invalid: map[string]interface{}{"op": "add", "path": "complex", "value": "value"},
		},

		// If omitted, the target location is assumed to be the resource itself. The "value" parameter contains a
		// set of attributes to be added to the resource.
		{
			valid:   map[string]interface{}{"op": "add", "value": map[string]interface{}{"attr1": "value"}},
			invalid: map[string]interface{}{"op": "add", "value": "value"},
		},
		{invalid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"invalid": "value"}}},
		{invalid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"invalid:attr1": "value"}}},

		// If the target location specifies a multi-valued attribute, a new value is added to the attribute.
		{valid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"multiValued": "value"}}},

		// Example on page 36 (RFC7644, Section 3.5.2.1).
		{valid: map[string]interface{}{"op": "add", "path": "complexMultiValued", "value": []interface{}{map[string]interface{}{"attr1": "value"}}}},
		{valid: map[string]interface{}{"op": "add", "path": "complexMultiValued", "value": map[string]interface{}{"attr1": "value"}}},

		// Example on page 37 (RFC7644, Section 3.5.2.1).
		{valid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"attr1": "value", "complexMultiValued": []interface{}{map[string]interface{}{"attr1": "value"}}}}},

		{
			valid:   map[string]interface{}{"op": "add", "path": `complexMultiValued[attr1 eq "value"].attr1`, "value": "value"},
			invalid: map[string]interface{}{"op": "add", "path": `complexMultiValued[attr1 eq "value"].attr2`, "value": "value"},
		},
		{
			valid:   map[string]interface{}{"op": "add", "path": `test:PatchEntity:complexMultiValued[attr1 eq "value"].attr1`, "value": "value"},
			invalid: map[string]interface{}{"op": "add", "path": `test:PatchEntity:complexMultiValued[attr2 eq "value"].attr1`, "value": "value"},
		},

		// Valid path, attribute not found.
		{invalid: map[string]interface{}{"op": "add", "path": "invalid", "value": "value"}},
		{invalid: map[string]interface{}{"op": "add", "path": "complex.invalid", "value": "value"}},

		// Sub-attributes in complex assignments.
		{valid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"complex.attr1": "value"}}},

		// Has no sub-attributes.
		{invalid: map[string]interface{}{"op": "add", "path": "attr1.invalid", "value": "value"}},

		// Invalid types.
		{invalid: map[string]interface{}{"op": "add", "path": "attr1", "value": 1}},
		{invalid: map[string]interface{}{"op": "add", "path": "multiValued", "value": 1}},
		{invalid: map[string]interface{}{"op": "add", "path": "multiValued", "value": []interface{}{1}}},
		{invalid: map[string]interface{}{"op": "add", "path": "complex.attr1", "value": 1}},
		{invalid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"attr1": 1}}},
		{invalid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"multiValued": 1}}},
		{invalid: map[string]interface{}{"op": "add", "value": map[string]interface{}{"multiValued": []interface{}{1}}}},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// valid
			if op := test.valid; op != nil {
				operation, _ := json.Marshal(op)
				validator, err := NewValidator(operation, patchSchema, patchSchemaExtension)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := validator.Validate(); err != nil {
					t.Errorf("The following operatation should be an VALID add operation:\n(case %d): %s\n%v", i, op, err)
				}
			}
			// invalid
			if op := test.invalid; op != nil {
				operation, _ := json.Marshal(op)
				validator, err := NewValidator(operation, patchSchema, patchSchemaExtension)
				if err != nil {
					return
				}
				if _, err := validator.Validate(); err == nil {
					t.Errorf("The following operatation should be an INVALID add operation:\n(case %d): %s", i, op)
				}
			}
		})
	}
}
