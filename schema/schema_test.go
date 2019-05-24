package schema

import (
	"testing"

	"github.com/elimity-com/scim/optional"
)

var minimalUserSchema = NewSchema(
	"urn:ietf:params:scim:schemas:core:2.0:User",
	"User",
	optional.NewString("User Account"),
	[]CoreAttribute{
		SimpleCoreAttribute(SimpleStringParams(StringParams{
			Name:       "userName",
			Required:   true,
			Uniqueness: AttributeUniquenessServer,
		})),
	},
)

func TestMinimalUserSchema(t *testing.T) {
	if minimalUserSchema.validate(map[string]interface{}{
		"field": "string",
	}) {
		t.Errorf("invalid resource expected")
	}
}
