package scim

import (
	"encoding/json"
	"testing"
)

func TestScimErrorUnmarshalJSONInvalid(t *testing.T) {
	data1 := []byte(`{"status":400}`)
	var scimErr1 scimError
	err := json.Unmarshal(data1, &scimErr1)
	if err == nil {
		t.Errorf("error expected")
	}

	data2 := []byte(`{"status":"test"}`)
	var scimErr2 scimError
	err = json.Unmarshal(data2, &scimErr2)
	if err == nil {
		t.Errorf("error expected")
	}
}
