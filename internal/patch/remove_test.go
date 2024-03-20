package patch

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elimity-com/scim/schema"
)

// The following example shows how remove all members of a group.
func Example_removeAllMembers() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op":   "remove",
		"path": "members",
	})
	validator, _ := NewValidator(operation, schema.CoreGroupSchema())
	fmt.Println(validator.Validate())
	// Output:
	// <nil> <nil>
}

// The following example shows how remove a value from a complex multi-valued attribute.
func Example_removeComplexMultiValuedAttributeValue() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op":   "remove",
		"path": `emails[type eq "work" and value eq "elimity.com"]`,
	})
	validator, _ := NewValidator(operation, schema.CoreUserSchema())
	fmt.Println(validator.Validate())
	// Output:
	// <nil> <nil>
}

// The following example shows how remove a single group from a user.
func Example_removeSingleGroup() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op":   "remove",
		"path": "groups",
		"value": []interface{}{
			map[string]interface{}{
				"$ref":  nil,
				"value": "f648f8d5ea4e4cd38e9c",
			},
		},
	})
	validator, _ := NewValidator(operation, schema.CoreUserSchema())
	fmt.Println(validator.Validate())
	// Output:
	// [map[]] <nil>
}

// The following example shows how remove a single member from a group.
func Example_removeSingleMember() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op":   "remove",
		"path": `members[value eq "0001"]`,
	})
	validator, _ := NewValidator(operation, schema.CoreGroupSchema())
	fmt.Println(validator.Validate())
	// Output:
	// <nil> <nil>
}

// The following example shows how to replace all of the members of a group with a different members list.
func Example_replaceAllMembers() {
	operations := []map[string]interface{}{
		{
			"op":   "remove",
			"path": "members",
		},
		{
			"op":   "remove",
			"path": "members",
			"value": []interface{}{
				map[string]interface{}{
					"value": "f648f8d5ea4e4cd38e9c",
				},
			},
		},
		{
			"op":   "add",
			"path": "members",
			"value": []interface{}{
				map[string]interface{}{
					"display": "di-wu",
					"$ref":    "https://example.com/v2/Users/0001",
					"value":   "0001",
				},
				map[string]interface{}{
					"display": "example",
					"$ref":    "https://example.com/v2/Users/0002",
					"value":   "0002",
				},
			},
		},
	}
	for _, op := range operations {
		operation, _ := json.Marshal(op)
		validator, _ := NewValidator(operation, schema.CoreGroupSchema())
		fmt.Println(validator.Validate())
	}
	// Output:
	// <nil> <nil>
	// [map[value:f648f8d5ea4e4cd38e9c]] <nil>
	// [map[$ref:https://example.com/v2/Users/0001 display:di-wu value:0001] map[$ref:https://example.com/v2/Users/0002 display:example value:0002]] <nil>
}

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
		valid   map[string]interface{}
		invalid map[string]interface{}
	}{
		// If "path" is unspecified, the operation fails.
		{invalid: map[string]interface{}{"op": "remove"}},

		// If the target location is a single-value attribute.
		{valid: map[string]interface{}{"op": "remove", "path": "attr1"}},
		// If the target location is a multi-valued attribute and no filter is specified.
		{valid: map[string]interface{}{"op": "remove", "path": "multiValued"}},
		// If the target location is a multi-valued attribute and a complex filter is specified comparing a "value".
		{valid: map[string]interface{}{"op": "remove", "path": `multivalued[value eq "value"]`}},
		// If the target location is a complex multi-valued attribute and a complex filter is specified based on the
		// attribute's sub-attributes
		{valid: map[string]interface{}{"op": "remove", "path": `complexMultiValued[attr1 eq "value"]`}},
		{valid: map[string]interface{}{"op": "remove", "path": `complexMultiValued[attr1 eq "value"].attr1`}},
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
					t.Fatal(err)
				}
				if _, err := validator.Validate(); err == nil {
					t.Errorf("The following operatation should be an INVALID add operation:\n(case %d): %s", i, op)
				}
			}
		})
	}
}
