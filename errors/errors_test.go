package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestScimErrorMarshalling(t *testing.T) {
	scimErr := ScimError{
		ScimType: ScimTypeTooMany,
		Detail:   "Just too many.",
		Status:   http.StatusTooManyRequests,
	}

	raw, err := json.Marshal(scimErr)
	if err != nil {
		t.Error(err)
	}

	var s struct {
		Schemas []string `json:"schemas"`
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Error(err)
	}

	if len(s.Schemas) != 1 || s.Schemas[0] != "urn:ietf:params:scim:api:messages:2.0:Error" {
		t.Errorf("did not get the correct schemas")
	}

	var e ScimError
	if err := json.Unmarshal(raw, &e); err != nil {
		t.Error(err)
	}

	if e.ScimType != scimErr.ScimType {
		t.Errorf("got invalid scim type: %s", e.ScimType)
	}

	if e.Detail != scimErr.Detail {
		t.Errorf("got invalid detail: %s", e.Detail)
	}

	if e.Status != scimErr.Status {
		t.Errorf("got invalid status: %d", e.Status)
	}
}

func TestCheckScimError(t *testing.T) {
	for _, test := range []struct {
		statusCode int
		method     string
		applicable bool
	}{
		// valid status + method
		{307, http.MethodGet, true},
		{308, http.MethodDelete, true},
		{400, http.MethodPut, true},
		{401, http.MethodPatch, true},
		{403, http.MethodPost, true},
		{404, http.MethodGet, true},
		{500, http.MethodDelete, false},
		{501, http.MethodPut, true},

		// invalid method
		{400, http.MethodConnect, false},
		{400, http.MethodHead, false},
		{400, http.MethodOptions, false},
		{400, http.MethodTrace, false},

		// invalid combination
		{409, http.MethodGet, false},
		{412, http.MethodGet, false},
		{412, http.MethodPost, false},
		{413, http.MethodGet, false},
		{413, http.MethodDelete, false},
		{413, http.MethodPut, false},
		{413, http.MethodPatch, false},
	} {
		err := CheckScimError(ScimError{
			Status: test.statusCode,
		}, test.method)
		if test.applicable {
			if err.Status == 500 {
				t.Error("no status code 500 expected")
			}
		} else {
			if err.Status != 500 {
				t.Errorf("status code 500 expected, got %d", err.Status)
			}
		}
	}
}

func TestCheckError(t *testing.T) {
	err := fmt.Errorf("error message")
	scimErr := CheckScimError(err, http.MethodGet)
	if scimErr.Detail != err.Error() {
		t.Error("invalid detail message")
	}
	if scimErr.Status != 500 {
		t.Errorf("status code 500 expected, got %d", scimErr.Status)
	}
}
