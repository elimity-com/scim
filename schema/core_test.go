package schema

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
)

func ExampleCoreAttribute_WithRequired() {
	// Customize the pre-built CoreUserSchema to make "emails" required.
	userSchema := CoreUserSchema()
	for i, attr := range userSchema.Attributes {
		if attr.Name() == "emails" {
			userSchema.Attributes[i] = attr.WithRequired(true)
		}
	}

	emails, _ := userSchema.Attributes.ContainsAttribute("emails")
	fmt.Println(emails.Required())
	// Output: true
}

func TestCoreAttribute_WithDescription(t *testing.T) {
	attr := SimpleCoreAttribute(SimpleStringParams(StringParams{
		Name: "test",
	}))
	desc := optional.NewString("new description")
	got := attr.WithDescription(desc)
	if got.Description() != "new description" {
		t.Errorf("WithDescription: got %q, want %q", got.Description(), "new description")
	}
	if attr.Description() != "" {
		t.Error("WithDescription modified the original attribute")
	}
}

func TestCoreAttribute_WithMutability(t *testing.T) {
	attr := SimpleCoreAttribute(SimpleStringParams(StringParams{
		Name: "test",
	}))
	got := attr.WithMutability(AttributeMutabilityImmutable())
	if got.Mutability() != "immutable" {
		t.Errorf("WithMutability: got %q, want %q", got.Mutability(), "immutable")
	}
	if attr.Mutability() != "readWrite" {
		t.Error("WithMutability modified the original attribute")
	}
}

func TestCoreAttribute_WithRequired(t *testing.T) {
	attr := SimpleCoreAttribute(SimpleStringParams(StringParams{
		Name: "test",
	}))
	got := attr.WithRequired(true)
	if !got.Required() {
		t.Error("WithRequired(true): got false, want true")
	}
	if attr.Required() {
		t.Error("WithRequired modified the original attribute")
	}
}

func TestCoreAttribute_WithReturned(t *testing.T) {
	attr := SimpleCoreAttribute(SimpleStringParams(StringParams{
		Name: "test",
	}))
	got := attr.WithReturned(AttributeReturnedNever())
	if got.Returned() != "never" {
		t.Errorf("WithReturned: got %q, want %q", got.Returned(), "never")
	}
	if attr.Returned() != "default" {
		t.Error("WithReturned modified the original attribute")
	}
}

func TestCoreAttribute_validate_allowsDistinctTypeValuePairs(t *testing.T) {
	emails := ComplexCoreAttribute(ComplexParams{
		Name:        "emails",
		MultiValued: true,
		SubAttributes: []SimpleParams{
			SimpleStringParams(StringParams{Name: "value"}),
			SimpleStringParams(StringParams{Name: "type"}),
			SimpleBooleanParams(BooleanParams{Name: "primary"}),
		},
	})

	_, scimErr := emails.validate([]interface{}{
		map[string]interface{}{"type": "work", "value": "john@work.com"},
		map[string]interface{}{"type": "home", "value": "john@home.com"},
	})
	if scimErr != nil {
		t.Errorf("unexpected error for distinct (type, value) pairs: %v", scimErr)
	}
}

func TestCoreAttribute_validate_allowsDuplicateTypeWithDifferentValue(t *testing.T) {
	emails := ComplexCoreAttribute(ComplexParams{
		Name:        "emails",
		MultiValued: true,
		SubAttributes: []SimpleParams{
			SimpleStringParams(StringParams{Name: "value"}),
			SimpleStringParams(StringParams{Name: "type"}),
		},
	})

	_, scimErr := emails.validate([]interface{}{
		map[string]interface{}{"type": "work", "value": "john@work.com"},
		map[string]interface{}{"type": "work", "value": "jane@work.com"},
	})
	if scimErr != nil {
		t.Errorf("unexpected error for same type with different value: %v", scimErr)
	}
}

func TestCoreAttribute_validate_rejectsDuplicatePrimary(t *testing.T) {
	emails := ComplexCoreAttribute(ComplexParams{
		Name:        "emails",
		MultiValued: true,
		SubAttributes: []SimpleParams{
			SimpleStringParams(StringParams{Name: "value"}),
			SimpleStringParams(StringParams{Name: "type"}),
			SimpleBooleanParams(BooleanParams{Name: "primary"}),
		},
	})

	_, scimErr := emails.validate([]interface{}{
		map[string]interface{}{"type": "work", "value": "john@work.com", "primary": true},
		map[string]interface{}{"type": "home", "value": "john@home.com", "primary": true},
	})
	if scimErr == nil {
		t.Fatal("expected error for duplicate primary")
	}
	if scimErr.Status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", scimErr.Status)
	}
	if scimErr.ScimType != errors.ScimTypeInvalidValue {
		t.Errorf("expected scimType %q, got %q", errors.ScimTypeInvalidValue, scimErr.ScimType)
	}
}

func TestCoreAttribute_validate_rejectsDuplicateTypeValuePairs(t *testing.T) {
	emails := ComplexCoreAttribute(ComplexParams{
		Name:        "emails",
		MultiValued: true,
		SubAttributes: []SimpleParams{
			SimpleStringParams(StringParams{Name: "value"}),
			SimpleStringParams(StringParams{Name: "type"}),
			SimpleBooleanParams(BooleanParams{Name: "primary"}),
		},
	})

	_, scimErr := emails.validate([]interface{}{
		map[string]interface{}{"type": "work", "value": "john@work.com"},
		map[string]interface{}{"type": "work", "value": "john@work.com"},
	})
	if scimErr == nil {
		t.Error("expected error for duplicate (type, value) pairs")
	}
}

func TestCoreAttribute_validate_rejectsDuplicateTypeValuePairsWithBadRequest(t *testing.T) {
	emails := ComplexCoreAttribute(ComplexParams{
		Name:        "emails",
		MultiValued: true,
		SubAttributes: []SimpleParams{
			SimpleStringParams(StringParams{Name: "value"}),
			SimpleStringParams(StringParams{Name: "type"}),
		},
	})

	_, scimErr := emails.validate([]interface{}{
		map[string]interface{}{"type": "work", "value": "john@work.com"},
		map[string]interface{}{"type": "work", "value": "john@work.com"},
	})
	if scimErr == nil {
		t.Fatal("expected error for duplicate (type, value) pairs")
	}
	if scimErr.Status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", scimErr.Status)
	}
	if scimErr.ScimType != errors.ScimTypeInvalidValue {
		t.Errorf("expected scimType %q, got %q", errors.ScimTypeInvalidValue, scimErr.ScimType)
	}
}

func TestCoreAttribute_validate_skipsDuplicateCheckWithoutTypeSubAttr(t *testing.T) {
	members := ComplexCoreAttribute(ComplexParams{
		Name:        "members",
		MultiValued: true,
		SubAttributes: []SimpleParams{
			SimpleStringParams(StringParams{Name: "value"}),
			SimpleStringParams(StringParams{Name: "displayName"}),
		},
	})

	_, scimErr := members.validate([]interface{}{
		map[string]interface{}{"value": "user1", "displayName": "User 1"},
		map[string]interface{}{"value": "user1", "displayName": "User 1"},
	})
	if scimErr != nil {
		t.Errorf("unexpected error for attribute without type sub-attribute: %v", scimErr)
	}
}
