package scim

import (
	"testing"

	"github.com/elimity-com/scim/schema"
	filterlib "github.com/scim2/filter-parser/v2"
)

func TestApplyPatch_AddSimpleAttribute(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("displayName"), Value: "John Doe"},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if result["displayName"] != "John Doe" {
		t.Errorf("expected displayName to be 'John Doe', got %v", result["displayName"])
	}
	// Original should not be modified.
	if _, ok := attrs["displayName"]; ok {
		t.Error("original attrs should not be modified")
	}
}

func TestApplyPatch_AddSubAttribute(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"name": map[string]interface{}{
			"givenName": "John",
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("name.familyName"), Value: "Doe"},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	name, ok := result["name"].(map[string]interface{})
	if !ok {
		t.Fatal("expected name to be a map")
	}
	if name["familyName"] != "Doe" {
		t.Errorf("expected familyName to be 'Doe', got %v", name["familyName"])
	}
	if name["givenName"] != "John" {
		t.Errorf("expected givenName to remain 'John', got %v", name["givenName"])
	}
}

func TestApplyPatch_AddToMultiValued(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{
				"value": "john@example.com",
				"type":  "work",
			},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("emails"), Value: []interface{}{
			map[string]interface{}{
				"value": "john@home.com",
				"type":  "home",
			},
		}},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	emails, ok := result["emails"].([]interface{})
	if !ok {
		t.Fatal("expected emails to be a list")
	}
	if len(emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(emails))
	}
}

func TestApplyPatch_AddWithNoPath(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op: PatchOperationAdd,
			Value: map[string]interface{}{
				"displayName": "John Doe",
				"userName":    "johnny",
			},
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if result["displayName"] != "John Doe" {
		t.Errorf("expected displayName to be 'John Doe', got %v", result["displayName"])
	}
	if result["userName"] != "johnny" {
		t.Errorf("expected userName to be 'johnny', got %v", result["userName"])
	}
}

func TestApplyPatch_DoesNotMutateOriginal(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
		"name": map[string]interface{}{
			"givenName": "John",
		},
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("userName"), Value: "jane"},
		{Op: PatchOperationAdd, Path: mustParsePath("name.familyName"), Value: "Doe"},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if attrs["userName"] != "john" {
		t.Error("original userName should not be modified")
	}
	name := attrs["name"].(map[string]interface{})
	if _, ok := name["familyName"]; ok {
		t.Error("original name map should not be modified")
	}
}

func TestApplyPatch_MultipleOperations(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("displayName"), Value: "John Doe"},
		{Op: PatchOperationReplace, Path: mustParsePath("userName"), Value: "jane"},
		{Op: PatchOperationRemove, Path: mustParsePath("displayName")},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if result["userName"] != "jane" {
		t.Errorf("expected userName to be 'jane', got %v", result["userName"])
	}
	if _, ok := result["displayName"]; ok {
		t.Error("expected displayName to be removed after sequence of operations")
	}
}

func TestApplyPatch_RemoveAttribute(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName":    "john",
		"displayName": "John Doe",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationRemove, Path: mustParsePath("displayName")},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := result["displayName"]; ok {
		t.Error("expected displayName to be removed")
	}
	if result["userName"] != "john" {
		t.Error("expected userName to remain")
	}
}

func TestApplyPatch_RemoveSubAttribute(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"name": map[string]interface{}{
			"givenName":  "John",
			"familyName": "Doe",
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationRemove, Path: mustParsePath("name.familyName")},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	name, ok := result["name"].(map[string]interface{})
	if !ok {
		t.Fatal("expected name to be a map")
	}
	if _, ok := name["familyName"]; ok {
		t.Error("expected familyName to be removed")
	}
	if name["givenName"] != "John" {
		t.Errorf("expected givenName to remain 'John'")
	}
}

func TestApplyPatch_RemoveSubAttributeFromValueExpression(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{
				"value":   "john@work.com",
				"type":    "work",
				"primary": true,
			},
			map[string]interface{}{
				"value": "john@home.com",
				"type":  "home",
			},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:   PatchOperationRemove,
			Path: mustParsePath(`emails[type eq "work"].primary`),
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	emails := result["emails"].([]interface{})
	workEmail := emails[0].(map[string]interface{})
	if _, ok := workEmail["primary"]; ok {
		t.Error("expected primary to be removed from work email")
	}
	if workEmail["value"] != "john@work.com" {
		t.Error("expected value to remain on work email")
	}
}

func TestApplyPatch_RemoveWithValueExpression(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{
				"value": "john@work.com",
				"type":  "work",
			},
			map[string]interface{}{
				"value": "john@home.com",
				"type":  "home",
			},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:   PatchOperationRemove,
			Path: mustParsePath(`emails[type eq "work"]`),
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	emails, ok := result["emails"].([]interface{})
	if !ok {
		t.Fatal("expected emails to be a list")
	}
	if len(emails) != 1 {
		t.Fatalf("expected 1 email, got %d", len(emails))
	}
	remaining := emails[0].(map[string]interface{})
	if remaining["type"] != "home" {
		t.Errorf("expected home email to remain, got %v", remaining["type"])
	}
}

func TestApplyPatch_ReplaceSimpleAttribute(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("userName"), Value: "jane"},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if result["userName"] != "jane" {
		t.Errorf("expected userName to be 'jane', got %v", result["userName"])
	}
}

func TestApplyPatch_ReplaceWithNoPath(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName":    "john",
		"displayName": "John Doe",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op: PatchOperationReplace,
			Value: map[string]interface{}{
				"displayName": nil,
				"userName":    "jane",
			},
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if result["userName"] != "jane" {
		t.Errorf("expected userName to be 'jane', got %v", result["userName"])
	}
	if _, ok := result["displayName"]; ok {
		t.Error("expected displayName to be removed")
	}
}

func TestApplyPatch_ReplaceWithValueExpression(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{
				"value": "john@work.com",
				"type":  "work",
			},
			map[string]interface{}{
				"value": "john@home.com",
				"type":  "home",
			},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:    PatchOperationReplace,
			Path:  mustParsePath(`emails[type eq "work"].value`),
			Value: "john@newwork.com",
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	emails, ok := result["emails"].([]interface{})
	if !ok {
		t.Fatal("expected emails to be a list")
	}
	workEmail := emails[0].(map[string]interface{})
	if workEmail["value"] != "john@newwork.com" {
		t.Errorf("expected work email to be updated, got %v", workEmail["value"])
	}
	homeEmail := emails[1].(map[string]interface{})
	if homeEmail["value"] != "john@home.com" {
		t.Errorf("expected home email to remain, got %v", homeEmail["value"])
	}
}

func mustParsePath(s string) *filterlib.Path {
	p, err := filterlib.ParsePath([]byte(s))
	if err != nil {
		panic(err)
	}
	return &p
}

func testUserSchema() schema.Schema {
	return schema.Schema{
		ID: schema.UserSchema,
		Attributes: schema.Attributes{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "userName",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "displayName",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name: "name",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{Name: "givenName"}),
					schema.SimpleStringParams(schema.StringParams{Name: "familyName"}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "emails",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{Name: "value"}),
					schema.SimpleStringParams(schema.StringParams{Name: "type"}),
					schema.SimpleBooleanParams(schema.BooleanParams{Name: "primary"}),
				},
			}),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "members",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{Name: "value"}),
					schema.SimpleStringParams(schema.StringParams{Name: "displayName"}),
				},
			}),
		},
	}
}
