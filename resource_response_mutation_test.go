package scim_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/schema"
)

// TestResponseDoesNotMutateHandlerAttributes verifies that serving a GET request
// does not modify the attributes map stored inside the resource handler.
//
// Resource.response() previously did:
//
//	response := r.Attributes  // reference copy, not a value copy
//	response["id"] = r.ID     // writes back into the handler's stored map
//
// The result: a second GET on the same resource would find framework-injected
// keys ("id", "schemas", "meta") already present in the attributes, leading to
// duplicate or stale data in the response.
func TestResponseDoesNotMutateHandlerAttributes(t *testing.T) {
	handler := &testResourceHandler{
		data: map[string]testData{
			"0001": {
				attributes: scim.ResourceAttributes{
					"userName": "alice",
				},
			},
		},
		schema: schema.CoreUserSchema(),
	}

	s, err := scim.NewServer(&scim.ServerArgs{
		ServiceProviderConfig: &scim.ServiceProviderConfig{},
		ResourceTypes: []scim.ResourceType{
			{
				Name:     "User",
				Endpoint: "/Users",
				Schema:   schema.CoreUserSchema(),
				Handler:  handler,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/Users/0001", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// After the GET, the handler's stored map must not contain any of the
	// framework-injected keys.  If response() mutated the map, these will be
	// present and a subsequent GET would return them as user-supplied data.
	stored := handler.data["0001"].attributes
	for _, key := range []string{"id", "schemas", "meta"} {
		if _, ok := stored[key]; ok {
			t.Errorf("Resource.response() mutated handler attributes: key %q was injected into the stored map", key)
		}
	}
}
