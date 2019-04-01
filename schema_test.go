package scim

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidationSchema(t *testing.T) {
	// validate raw meta schema with meta schema
	if err := metaSchema.validate(strings.NewReader(rawMetaSchema)); err != nil {
		t.Errorf(err.Error())
	}

	s := `{
  		"id": "urn:ietf:params:scim:schemas:core:2.0:User",
  		"name": "User",
  		"description": "User Account",
  		"attributes": [
    		{
      			"name": "userName",
      			"type": "string",
      			"multiValued": false,
      			"required": true,
      			"caseExact": false,
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "server"
    		},
    		{
      			"name": "name",
      			"type": "complex",
      			"multiValued": false,
      			"required": false,
      			"subAttributes": [
        			{
          				"name": "familyName",
          				"type": "string",
          				"multiValued": false,
          				"required": false,
          				"caseExact": false,
          				"mutability": "readWrite",
          				"returned": "default",
          				"uniqueness": "none"
        			},
        			{
          				"name": "givenName",
          				"type": "string",
          				"multiValued": false,
          				"required": false,
          				"caseExact": false,
          				"mutability": "readWrite",
          				"returned": "default",
          				"uniqueness": "none"
        			}
      			],
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "none"
    		},
    		{
				"name": "displayName",
      			"type": "string",
      			"multiValued": false,
      			"required": false,
      			"caseExact": false,
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "none"
    		},
    		{
      			"name": "emails",
      			"type": "complex",
      			"multiValued": true,
      			"required": false,
      			"subAttributes": [
      				{
      					"name": "value",
      					"type": "string",
      					"multiValued": false,
      					"required": false,
      					"caseExact": false,
      					"mutability": "readWrite",
      					"returned": "default",
      					"uniqueness": "none"
      				},
      				{
      					"name": "display",
      					"type": "string",
      					"multiValued": false,
      					"required": false,
      					"caseExact": false,
      					"mutability": "readWrite",
      					"returned": "default",
      					"uniqueness": "none"
      				},
      				{
      					"name": "type",
      					"type": "string",
      					"multiValued": false,
      					"required": false,
      					"caseExact": false,
      					"canonicalValues": [
      						"work",
      						"home",
      						"other"
      					],
      					"mutability": "readWrite",
      					"returned": "default",
      					"uniqueness": "none"
      				},
      				{
      					"name": "primary",
      					"type": "boolean",
      					"multiValued": false,
      					"required": false,
      					"mutability": "readWrite",
      					"returned": "default"
      				}
      			],
      			"mutability": "readWrite",
      			"returned": "default",
      			"uniqueness": "none"
    		}
  		],
  		"meta": {
    		"resourceType": "Schema",
    		"location": "/v2/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
  		}
	}`

	// validate simple user schema with meta schema
	if err := metaSchema.validate(strings.NewReader(s)); err != nil {
		t.Errorf(err.Error())
	}

	var schema schema
	if err := json.Unmarshal([]byte(s), &schema); err != nil {
		t.Errorf(err.Error())
	}

	// validate user with simple user schema
	if err := schema.validate(strings.NewReader(`{
		"schemas": [
			"schemas"
		],
		"id": "id",
     	"userName": "username",
     	"externalId": "eid",
		"name": {
       		"familyName": "family name",
       		"givenName": "given name"
		}
	}`)); err != nil {
		t.Errorf(err.Error())
	}

	// invalid user
	if err := schema.validate(strings.NewReader(`{
		"schemas": [
			"schemas"
		],
		"id": "id",
     	"userName": "username",
     	"externalId": "eid",
		"name": "name"
	}`)); err == nil {
		t.Errorf("expected: could not convert name to type complex")
	}
	if err := schema.validate(strings.NewReader(`{
		"schemas": [
			"test"
		],
		"id": "test",
     	"userName": "test",
     	"externalId": "test",
		"name": {
       		"familyName": {}
		}
	}`)); err == nil {
		t.Errorf("expected: could not convert map[] to type string")
	}
}

func TestSchema_Validate(t *testing.T) {
	// invalid json
	if err := metaSchema.validate(strings.NewReader(``)); err == nil {
		t.Errorf("expected: unexpected end of json input")
	}
	if err := metaSchema.validate(strings.NewReader(`[]`)); err == nil {
		t.Errorf("expected: cannot unmarshal array into Go value of type map[string]interface{}")
	}

	// duplicate key
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"ID": "test"
	}`)); err == nil {
		t.Errorf("expected: duplicate key: id")
	}
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": [
			{
				"name": "test2",
				"NAME": "test2",
				"type": "string",
				"multiValued": false
			}
		]
	}`)); err == nil {
		t.Errorf("expected: duplicate key: name")
	}

	// required error
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test"
	}`)); err == nil {
		t.Errorf("expected: could not find required value attributes")
	}
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": [
			{
				"name": "test2"
			}
		]
	}`)); err == nil {
		t.Errorf("expected: could not find required value type")
	}

	// converting error
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": "test"
	}`)); err == nil {
		t.Errorf("expected: could not convert test to type complex")
	}
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": [
			"test"
		]
	}`)); err == nil {
		t.Errorf("expected: element of slice was not a complex value: test")
	}
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": [
			{
				"name": "test2",
				"type": "string",
				"multiValued": true,
				"canonicalValues": [
					{}
				]
			}
		]
	}`)); err == nil {
		t.Errorf("expected: could not convert map[] to type string")
	}
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": [
			{
				"name": "test2",
				"type": "string",
				"multiValued": "true"
			}
		]
	}`)); err == nil {
		t.Errorf("expected: could not convert true to type boolean")
	}
	if err := metaSchema.validate(strings.NewReader(`{
		"id": {},
		"name": "test",
		"attributes": [
		]
	}`)); err == nil {
		t.Errorf("expected: could not convert map[] to type string")
	}

	// canonical values error
	if err := metaSchema.validate(strings.NewReader(`{
		"id": "test",
		"name": "test",
		"attributes": [
			{
				"name": "test2",
				"type": "string",
				"multiValued": true,
				"returned": "test"
			}
		]
	}`)); err == nil {
		t.Errorf("expected: test not in canonical values")
	}
}
