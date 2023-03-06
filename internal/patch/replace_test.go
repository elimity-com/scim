package patch

import (
	"fmt"
	"github.com/elimity-com/scim/schema"
)

// The following example shows how to replace all values of one or more specific attributes.
func Example_replaceAnyAttribute() {
	operation := `{
	"op": "replace",
	"value": {
		"emails": [
			{
				"value": "quint",
				"type": "work",
				"primary": true
			},
			{
				"value": "me@di-wu.be",
				"type": "home"
			}
		],
		"nickname": "di-wu"
	}
}`
	validator, _ := NewValidator(operation, schema.CoreUserSchema())
	fmt.Println(validator.Validate())
	// Output:
	// map[emails:[map[primary:true type:work value:quint] map[type:home value:me@di-wu.be]] nickname:di-wu] <nil>
}

// The following example shows how to replace all of the members of a group with a different members list in a single
// replace operation.
func Example_replaceMembers() {
	operations := []string{`{
	"op": "replace",
	"path": "members",
	"value": [
		{
			"display": "di-wu",
			"$ref": "https://example.com/v2/Users/0001",
			"value": "0001"
		},
		{
			"display": "example",
			"$ref": "https://example.com/v2/Users/0002",
			"value": "0002"
		}
	]
}`,
	}
	for _, op := range operations {
		validator, _ := NewValidator(op, schema.CoreGroupSchema())
		fmt.Println(validator.Validate())
	}
	// Output:
	// [map[$ref:https://example.com/v2/Users/0001 display:di-wu value:0001] map[$ref:https://example.com/v2/Users/0002 display:example value:0002]] <nil>
}

// The following example shows how to change a specific sub-attribute "streetAddress" of complex attribute "emails"
// selected by a "valuePath" filter.
func Example_replaceSpecificSubAttribute() {
	operation := `{
	"op": "replace",
	"path": "addresses[type eq \"work\"].streetAddress",
	"value": "ExampleStreet 100"
}`
	validator, _ := NewValidator(operation, schema.CoreUserSchema())
	fmt.Println(validator.Validate())
	// Output:
	// ExampleStreet 100 <nil>
}

// The following example shows how to change a User's entire "work" address, using a "valuePath" filter.
func Example_replaceWorkAddress() {
	operation := `{
	"op": "replace",
	"path": "addresses[type eq \"work\"]",
	"value": {
		"type": "work",
		"streetAddress": "ExampleStreet 1",
		"locality": "ExampleCity",
		"postalCode": "0001",
		"country": "BE",
		"primary": true
	}
}`
	validator, _ := NewValidator(operation, schema.CoreUserSchema())
	fmt.Println(validator.Validate())
	// Output:
	// [map[country:BE locality:ExampleCity postalCode:0001 primary:true streetAddress:ExampleStreet 1 type:work]] <nil>
}
