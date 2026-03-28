package scim

import (
	"testing"

	scimErrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
	filter "github.com/scim2/filter-parser/v2"
)

func TestApplyPatch_AddDistinctTypeValuePair(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
		"emails": []interface{}{
			map[string]interface{}{"type": "work", "value": "john@work.com"},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("emails"), Value: []interface{}{
			map[string]interface{}{"type": "home", "value": "john@home.com"},
		}},
	}, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	emails, ok := result["emails"].([]interface{})
	if !ok {
		t.Fatal("expected emails to be a list")
	}
	if len(emails) != 2 {
		t.Errorf("expected 2 emails, got %d", len(emails))
	}
}

// RFC 7644 Section 3.5.2: the server SHALL set primary to false on the
// existing value when a new value with primary: true is added.
func TestApplyPatch_AddDuplicatePrimary_AutoClears(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{"type": "work", "value": "john@work.com", "primary": true},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("emails"), Value: []interface{}{
			map[string]interface{}{"type": "home", "value": "john@home.com", "primary": true},
		}},
	}, s)
	if err != nil {
		t.Fatalf("expected auto-clear of old primary, got error: %v", err)
	}
	emails := result["emails"].([]interface{})
	if len(emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(emails))
	}
	work := emails[0].(map[string]interface{})
	if p, _ := work["primary"].(bool); p {
		t.Error("expected old primary (work) to be cleared")
	}
	home := emails[1].(map[string]interface{})
	if p, _ := home["primary"].(bool); !p {
		t.Error("expected new primary (home) to remain true")
	}
}

func TestApplyPatch_AddDuplicateTypeValuePair(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
		"emails": []interface{}{
			map[string]interface{}{"type": "work", "value": "john@work.com"},
		},
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("emails"), Value: []interface{}{
			map[string]interface{}{"type": "work", "value": "john@work.com"},
		}},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeInvalidValue)
}

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

func TestApplyPatch_AddSubAttribute_ParentNotExist(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
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
		t.Errorf("expected familyName 'Doe', got %v", name["familyName"])
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

func TestApplyPatch_AddWithValueExpression(t *testing.T) {
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
			Op:    PatchOperationAdd,
			Path:  mustParsePath(`emails[type eq "work"].primary`),
			Value: true,
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	emails := result["emails"].([]interface{})
	workEmail := emails[0].(map[string]interface{})
	if workEmail["primary"] != true {
		t.Errorf("expected primary to be true on work email, got %v", workEmail["primary"])
	}
	homeEmail := emails[1].(map[string]interface{})
	if _, ok := homeEmail["primary"]; ok {
		t.Error("expected home email to not have primary set")
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

// RFC 7644 Section 3.5.2: "a client MAY 'add' a value to an 'immutable'
// attribute if the attribute had no previous value".
func TestApplyPatch_ImmutableAttribute_AddNew_Succeeds(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("externalId"), Value: "ext-1"},
	}, s)
	if err != nil {
		t.Fatal(err)
	}
	if result["externalId"] != "ext-1" {
		t.Errorf("expected externalId 'ext-1', got %v", result["externalId"])
	}
}

func TestApplyPatch_ImmutableAttribute_Remove_ReturnsMutability(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"externalId": "ext-1",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationRemove, Path: mustParsePath("externalId")},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeMutability)
}

func TestApplyPatch_ImmutableAttribute_Replace_ReturnsMutability(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"externalId": "ext-1",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("externalId"), Value: "ext-2"},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeMutability)
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

// RFC 7644 Section 3.5.2: "a client MUST NOT modify an attribute that has
// mutability 'readOnly'".
func TestApplyPatch_ReadOnlyAttribute_Add_ReturnsMutability(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationAdd, Path: mustParsePath("id"), Value: "123"},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeMutability)
}

func TestApplyPatch_ReadOnlyAttribute_Remove_ReturnsMutability(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"id": "123",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationRemove, Path: mustParsePath("id")},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeMutability)
}

func TestApplyPatch_ReadOnlyAttribute_Replace_ReturnsMutability(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"id": "123",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("id"), Value: "456"},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeMutability)
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

func TestApplyPatch_RemoveRequiredAttribute(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName":    "john",
		"displayName": "John Doe",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationRemove, Path: mustParsePath("userName")},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeInvalidValue)
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

// RFC 7644 Section 3.5.2.2: "If 'path' is unspecified, the operation fails
// with HTTP status code 400 and a 'scimType' error code of 'noTarget'".
func TestApplyPatch_RemoveWithNoPath_ReturnsNoTarget(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationRemove},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeNoTarget)
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

// RFC 7644 Section 3.5.2.3: "If the target location specifies a complex
// attribute, a set of sub-attributes SHALL be specified in the 'value'
// parameter, which replaces any existing values or adds where an attribute did
// not previously exist. Sub-attributes that are not specified in the 'value'
// parameter are left unchanged".
func TestApplyPatch_ReplaceComplexAttribute_MergesSubAttributes(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"name": map[string]interface{}{
			"givenName":  "John",
			"familyName": "Doe",
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:   PatchOperationReplace,
			Path: mustParsePath("name"),
			Value: map[string]interface{}{
				"familyName": "Smith",
			},
		},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	name, ok := result["name"].(map[string]interface{})
	if !ok {
		t.Fatal("expected name to be a map")
	}
	if name["familyName"] != "Smith" {
		t.Errorf("expected familyName to be 'Smith', got %v", name["familyName"])
	}
	if name["givenName"] != "John" {
		t.Errorf("expected givenName to remain 'John', got %v", name["givenName"])
	}
}

func TestApplyPatch_ReplaceDuplicateTypeValuePair(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
		"emails": []interface{}{
			map[string]interface{}{"type": "work", "value": "john@work.com"},
		},
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("emails"), Value: []interface{}{
			map[string]interface{}{"type": "work", "value": "john@work.com"},
			map[string]interface{}{"type": "work", "value": "john@work.com"},
		}},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeInvalidValue)
}

func TestApplyPatch_ReplaceMultiValuedWithoutFilter(t *testing.T) {
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

	newEmails := []interface{}{
		map[string]interface{}{
			"value": "jane@new.com",
			"type":  "new",
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("emails"), Value: newEmails},
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
	email := emails[0].(map[string]interface{})
	if email["value"] != "jane@new.com" {
		t.Errorf("expected email 'jane@new.com', got %v", email["value"])
	}
}

// RFC 7644 Section 3.5.2.3: "If the target location path specifies an attribute
// that does not exist, the service provider SHALL treat the operation as an 'add'".
func TestApplyPatch_ReplaceNonExistentAttribute_TreatedAsAdd(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{Op: PatchOperationReplace, Path: mustParsePath("displayName"), Value: "John Doe"},
	}, s)
	if err != nil {
		t.Fatal(err)
	}

	if result["displayName"] != "John Doe" {
		t.Errorf("expected displayName to be 'John Doe', got %v", result["displayName"])
	}
}

// RFC 7644 Section 3.5.2: when primary is set to true via a value expression,
// the server SHALL clear primary on all other values, even when the modified
// element appears before the existing primary in the list.
func TestApplyPatch_ReplacePrimaryViaValueExpr_AutoClears(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{"type": "home", "value": "john@home.com"},
			map[string]interface{}{"type": "work", "value": "john@work.com", "primary": true},
		},
	}

	result, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:    PatchOperationReplace,
			Path:  mustParsePath(`emails[type eq "home"].primary`),
			Value: true,
		},
	}, s)
	if err != nil {
		t.Fatalf("expected auto-clear, got error: %v", err)
	}
	emails := result["emails"].([]interface{})
	home := emails[0].(map[string]interface{})
	if p, _ := home["primary"].(bool); !p {
		t.Error("expected home to become primary")
	}
	work := emails[1].(map[string]interface{})
	if p, _ := work["primary"].(bool); p {
		t.Error("expected work primary to be cleared")
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

// RFC 7644 Section 3.5.2.3: "If the target location is a multi-valued attribute
// for which a value selection filter ('valuePath') has been supplied and no
// record match was made, the service provider SHALL indicate failure by
// returning HTTP status code 400 and a 'scimType' error code of 'noTarget'".
func TestApplyPatch_ReplaceValueExprNoTarget_AttributeMissing(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"userName": "john",
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:    PatchOperationReplace,
			Path:  mustParsePath(`emails[type eq "work"].value`),
			Value: "new@work.com",
		},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeNoTarget)
}

func TestApplyPatch_ReplaceValueExprNoTarget_NoMatch(t *testing.T) {
	s := testUserSchema()
	attrs := ResourceAttributes{
		"emails": []interface{}{
			map[string]interface{}{
				"value": "john@home.com",
				"type":  "home",
			},
		},
	}

	_, err := ApplyPatch(attrs, []PatchOperation{
		{
			Op:    PatchOperationReplace,
			Path:  mustParsePath(`emails[type eq "work"].value`),
			Value: "new@work.com",
		},
	}, s)
	assertScimError(t, err, scimErrors.ScimTypeNoTarget)
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

func mustParsePath(s string) *filter.Path {
	p, err := filter.ParsePath([]byte(s))
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
				Name:     "userName",
				Required: true,
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "displayName",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "externalId",
				Mutability: schema.AttributeMutabilityImmutable(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "id",
				Mutability: schema.AttributeMutabilityReadOnly(),
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
