package patch

import (
	"fmt"
	"testing"
)

func TestOperationValidator_ValidateRemove(t *testing.T) {
	// The goal this test is to cover Section 3.5.2.2 of RFC7644.
	// More info: https://tools.ietf.org/html/rfc7644#section-3.5.2.2

	// Some more indirect things are covered by these tests:
	// - If the target location is a single-value attribute, the attribute and its associated value is removed.
	// - If the target location is a multi-valued attribute and no filter is specified, the attribute and all values
	//   are removed.
	// - If the target location is a multi-valued attribute and a complex filter is specified comparing a "value", the
	//   values matched by the filter are removed.
	// - If the target location is a complex multi-valued attribute and a complex filter is specified based on the
	//   attribute's sub-attributes, the matching records are removed.

	for i, test := range []struct {
		valid   string
		invalid string
	}{
		// If "path" is unspecified, the operation fails.
		{invalid: `{"op":"remove"}`},

		// If the target location is a single-value attribute.
		{valid: `{"op":"remove","path":"attr1"}`},
		// If the target location is a multi-valued attribute and no filter is specified.
		{valid: `{"op":"remove","path":"multiValued"}`},
		// If the target location is a complex multi-valued attribute and a complex filter is specified comparing a
		// "value".
		{valid: `{"op":"remove","path":"complexMultiValued[attr1 eq \"value\"]"}`},
		// If the target location is a complex multi-valued attribute and a complex filter is specified based on the
		// attribute's sub-attributes
		{valid: `{"op":"remove","path":"complexMultiValued[attr1 eq \"value\"].attr1"}`},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// valid
			if op := test.valid; op != "" {
				validator, err := NewValidator(op, patchSchema, patchSchemaExtension)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := validator.Validate(); err != nil {
					t.Errorf("The following operatation should be an VALID add operation:\n(case %d): %s\n%v", i, op, err)
				}
			}
			// invalid
			if op := test.invalid; op != "" {
				validator, err := NewValidator(op, patchSchema, patchSchemaExtension)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := validator.Validate(); err == nil {
					t.Errorf("The following operatation should be an INVALID add operation:\n(case %d): %s", i, op)
				}
			}
		})
	}
}
