package scim

import (
	"fmt"
	"testing"
)

func TestNewServer(t *testing.T) {
	cases := []struct {
		s   string
		err string
	}{
		{
			s: `{
  				"id": "urn:ietf:params:scim:schemas:core:2.0:User",
  				"name": "User",
  				"attributes": []
			}`,
			err: "required array is empty",
		},
	}

	for idx, test := range cases {
		t.Run(fmt.Sprintf("invalid schema %d", idx), func(t *testing.T) {
			if _, err := NewSchemaFromString(test.s); err == nil || err.Error() != test.err {
				t.Errorf("expected: %s / got: %v", test.err, err)
			}
		})
	}

	user, err := NewSchemaFromString(`{
  		"id": "urn:ietf:params:scim:schemas:core:2.0:User",
  		"name": "User",
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
    		}
  		]
	}`)
	if err != nil {
		t.Error(err)
	}

	_, err = NewSchemaFromFile("")
	if err == nil {
		t.Error("expected: no such file or directory")
	}

	NewServer(*user)
}
