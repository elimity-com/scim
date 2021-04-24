package patch

import "testing"

func TestNewPathValidator(t *testing.T) {
	t.Run("Invalid JSON", func(t *testing.T) {
		// The quotes in the value filter are not escaped.
		op := `{"op":"add","path":"complexMultiValued[attr1 eq "value"].attr1","value":"value"}`
		if _, err := NewPathValidator(op, patchSchema); err == nil {
			t.Error("expected JSON error, got none")
		}
	})
	t.Run("Invalid Op", func(t *testing.T) {
		// The quotes in the value filter are not escaped.
		op := `{"op":"invalid","path":"attr1","value":"value"}`
		validator, _ := NewPathValidator(op, patchSchema)
		if err := validator.Validate(); err == nil {
			t.Errorf("expected error, got none")
		}
	})
}
