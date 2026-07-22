package patch

import (
	"encoding/json"
	"fmt"
	"github.com/elimity-com/scim/schema"
)

// The following example shows how to add a member to a group.
func Example_addMemberToGroup() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op":   "add",
		"path": "members",
		"value": map[string]interface{}{
			"display": "di-wu",
			"$ref":    "https://example.com/v2/Users/0001",
			"value":   "0001",
		},
	})
	validator, _ := NewValidator(operation, schema.CoreGroupSchema())
	fmt.Println(validator.Validate())
	// Output:
	// [map[$ref:https://example.com/v2/Users/0001 display:di-wu value:0001]] <nil>
}

// The following example shows how to add one or more attributes to a User resource without using a "path" attribute.
func Example_addWithoutPath() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op": "add",
		"value": map[string]interface{}{
			"emails": []map[string]interface{}{
				{
					"value": "quint@elimity.com",
					"type":  "work",
				},
			},
			"nickname": "di-wu",
		},
	})
	validator, _ := NewValidator(operation, schema.CoreUserSchema())
	fmt.Println(validator.Validate())
	// Output:
	// map[emails:[map[type:work value:quint@elimity.com]] nickname:di-wu] <nil>
}

// The following example shows how to add extension attributes to a User resource without using a "path" attribute,
// where the extension is provided as a nested object keyed by its schema URI. See RFC 7643 Section 3.3
// and RFC 7644 Section 3.5.2.
func Example_addExtensionWithoutPath() {
	operation, _ := json.Marshal(map[string]interface{}{
		"op": "add",
		"value": map[string]interface{}{
			"userName": "test",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]interface{}{
				"organization": "SUSE",
			},
		},
	})
	validator, _ := NewValidator(operation, schema.CoreUserSchema(), schema.ExtensionEnterpriseUser())
	fmt.Println(validator.Validate())
	// Output:
	// map[urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:map[organization:SUSE] userName:test] <nil>
}
