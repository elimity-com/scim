package scim

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func TestInvalidRequests(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		target         string
		body           io.Reader
		expectedStatus int
	}{
		{
			name:           "invalid get request",
			method:         http.MethodGet,
			target:         "/Invalid",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid get request, with version",
			method:         http.MethodGet,
			target:         "/v2/Invalid",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid schema request",
			method:         http.MethodGet,
			target:         "/Schemas/urn:ietf:params:scim:schemas:core:2.0:Invalid",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid post request",
			method:         http.MethodPost,
			target:         "/Users",
			body:           strings.NewReader(`{"id": "other"}`),
			expectedStatus: http.StatusBadRequest,
		}, {
			name:           "invalid put request",
			method:         http.MethodPut,
			target:         "/Users/0001",
			body:           strings.NewReader(`{"more": "test"}`),
			expectedStatus: http.StatusBadRequest,
		}, {
			name:           "users post request with invalid externalId",
			method:         http.MethodPost,
			target:         "/Users",
			body:           strings.NewReader(`{"id": "other", "userName": "test1", "externalId": {"not": "this"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "users post request with invalid userName",
			method:         http.MethodPost,
			target:         "/Users",
			body:           strings.NewReader(`{"id": "other", "userName":  {"not": "this""}}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "users put request with invalid externalId",
			method:         http.MethodPut,
			target:         "/v2/Users/0002",
			body:           strings.NewReader(`{"id": "other", "userName": "test2", "externalId": {"test":"test"}}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "users put request with invalid userName",
			method:         http.MethodPut,
			target:         "/Users/0003",
			body:           strings.NewReader(`{"id": "other", "userName": {"test": "test"}}`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, test.expectedStatus, rr.Code)
		})
	}
}

func TestServerMeEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Me", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusNotImplemented, rr.Code)
}

func TestServerResourceDeleteHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/0001", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusNoContent, rr.Code)
}

func TestServerResourceDeleteHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusNotFound, rr.Code)

	var scimErr *errors.ScimError
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &scimErr))
	expectedError := &errors.ScimError{
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("Resource %d not found.", 9999),
	}
	assertEqualSCIMErrors(t, expectedError, scimErr)
}

func TestServerResourceGetHandler(t *testing.T) {
	tests := []struct {
		name                 string
		target               string
		expectedUserName     string
		expectedExternalID   string
		expectedVersion      string
		expectedCreated      string
		expectedLastModified string
	}{
		{
			name:                 "Users get request without version",
			target:               "/Users/0001",
			expectedUserName:     "test01",
			expectedExternalID:   "external1",
			expectedVersion:      "v1",
			expectedCreated:      "2020-01-01T15:04:05+07:00",
			expectedLastModified: "2020-02-01T16:05:04+07:00",
		}, {
			name:                 "Users get request with version",
			target:               "/v2/Users/0002",
			expectedUserName:     "test02",
			expectedExternalID:   "external2",
			expectedVersion:      "v2",
			expectedCreated:      "2020-01-02T15:04:05+07:00",
			expectedLastModified: "2020-02-02T16:05:04+07:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)

			assertEqual(t, "application/scim+json", rr.Header().Get("Content-Type"))

			assertEqual(t, tt.expectedVersion, rr.Header().Get("Etag"))

			var resource map[string]interface{}
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &resource))

			assertEqual(t, tt.expectedUserName, resource["userName"])
			assertEqual(t, tt.expectedExternalID, resource["externalId"])

			meta, ok := resource["meta"].(map[string]interface{})
			assertTypeOk(t, ok, "object")

			assertEqual(t, "User", meta["resourceType"])
			assertEqual(t, tt.expectedCreated, meta["created"])
			assertEqual(t, tt.expectedLastModified, meta["lastModified"])
			assertEqual(t, fmt.Sprintf("Users/%s", resource["id"]), meta["location"])
			assertEqual(t, tt.expectedVersion, meta["version"])
		})
	}
}

func TestServerResourceGetHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusNotFound, rr.Code)

	var scimErr *errors.ScimError
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &scimErr))
	expectedError := &errors.ScimError{
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("Resource %d not found.", 9999),
	}
	assertEqualSCIMErrors(t, expectedError, scimErr)
}

func TestServerResourcePatchHandlerFailOnBadType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"replace",
		    "path":"active",
		    "value":"test"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	var resource map[string]interface{}
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &resource))

	assertEqualStatusCode(t, http.StatusBadRequest, rr.Code)

	assertEqual(t, errors.ScimErrorInvalidValue.Detail, resource["detail"])
}

func TestServerResourcePatchHandlerInvalidPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"replace",
		    "path":"name.invalid",
		    "value":"test"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusBadRequest, rr.Code)

	var scimErr *errors.ScimError
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &scimErr))
	assertEqualSCIMErrors(t, &errors.ScimErrorInvalidPath, scimErr)
}

func TestServerResourcePatchHandlerInvalidRemoveOp(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Groups/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"remove",
		    "path":"members[invalid eq \"empty\"]"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusBadRequest, rr.Code)
}

func TestServerResourcePatchHandlerMapTypeSubAttribute(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestServer().ServeHTTP(recorder, httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
			"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
			"Operations":[
			  {
				"op": "replace",
				"path": "emails[type eq \"work\"].value",
				"value": "hoge@example.com"
			  }
			]
		}`)))
	assertEqualStatusCode(t, http.StatusOK, recorder.Code)

	recorder2 := httptest.NewRecorder()
	newTestServer().ServeHTTP(recorder2, httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
			"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
			"Operations":[
			  {
				"op": "replace",
				"path": "emails[type eq \"work\"].value",
				"value": 10000
			  }
			]
		}`)))
	assertEqualStatusCode(t, http.StatusBadRequest, recorder2.Code)
}

func TestServerResourcePatchHandlerReturnsNoContent(t *testing.T) {
	reqs := []*http.Request{
		httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
			"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
			"Operations":[
			  {
				"op": "add",
				"path": "userName",
				"value": "test01"
			  }
			]
		}`)),
		httptest.NewRequest(http.MethodPatch, "/Users/0002", strings.NewReader(`{
			"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
			"Operations":[
			  {
				"op": "replace",
				"path": "userName",
				"value": "test02"
			  }
			]
		}`)),
		httptest.NewRequest(http.MethodPatch, "/Users/0003", strings.NewReader(`{
			"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
			"Operations":[
			  {
				"op": "remove",
				"path": "name.givenName"
			  }
			]
		}`)),
	}
	for _, req := range reqs {
		rr := httptest.NewRecorder()
		newTestServer().ServeHTTP(rr, req)

		assertEqualStatusCode(t, http.StatusNoContent, rr.Code)
	}
}

// Tests valid add, replace, and remove operations.
func TestServerResourcePatchHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"add",
		    "value":{
		      "emails":[
		        {
			  "value":"babs@jensen.org",
			  "type":"home"
		        }
		      ]
		    }
		  },
		  {
		    "op":"replace",
		    "path":"active",
		    "value":false
		  },
		  {
		    "op":"replace",
		    "path":"externalId",
		    "value": "external_test_replace"
		  },
		  {
		    "op":"remove",
		    "path":"displayName"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	assertEqual(t, "application/scim+json", rr.Header().Get("Content-Type"))

	expectedVersion := "v1.patch"

	assertEqual(t, expectedVersion, rr.Header().Get("Etag"))

	var resource map[string]interface{}
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &resource))

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	assertNil(t, resource["displayName"], "displayName")
	assertFalse(t, resource["active"].(bool))
	assertEqual(t, "external_test_replace", resource["externalId"])

	if resource["emails"] == nil || len(resource["emails"].([]interface{})) < 1 {
		t.Errorf("handler did not add user's email address")
	}

	meta, ok := resource["meta"].(map[string]interface{})
	assertTrue(t, ok)

	assertEqual(t, "User", meta["resourceType"])
	assertEqual(t, "2020-01-01T15:04:05+07:00", meta["created"])
	assertNotEqual(t, "2020-02-01T16:05:04+07:00", meta["lastModified"])
	assertEqual(t, "Users/0001", meta["location"])
	assertEqual(t, expectedVersion, meta["version"])
}

func TestServerResourcePatchHandlerValidPathHasSubAttributes(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"replace",
		    "path":"name.givenName",
		    "value":"test"
		  },
		  {
		    "op":"replace",
		    "path":"name.familyName",
		    "value":"test"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)
}

func TestServerResourcePatchHandlerValidRemoveOp(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Groups/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"remove",
		    "path":"members[value eq \"2819c223-7f76-...413861904646\"]"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusNoContent, rr.Code)
}

func TestServerResourcePostHandlerValid(t *testing.T) {
	tests := []struct {
		name               string
		target             string
		body               io.Reader
		expectedUserName   string
		expectedExternalID interface{}
	}{
		{
			name:               "Users post request without version",
			target:             "/Users",
			body:               strings.NewReader(`{"id": "other", "userName": "test1", "externalId": "external_test1"}`),
			expectedUserName:   "test1",
			expectedExternalID: "external_test1",
		}, {
			name:               "Users post request with version",
			target:             "/v2/Users",
			body:               strings.NewReader(`{"id": "other", "userName": "test2", "externalId": "external_test2"}`),
			expectedUserName:   "test2",
			expectedExternalID: "external_test2",
		}, {
			name:               "Users post request without externalId",
			target:             "/v2/Users",
			body:               strings.NewReader(`{"id": "other", "userName": "test3"}`),
			expectedUserName:   "test3",
			expectedExternalID: nil,
		}, {
			name:               "Users post request with immutable attribute",
			target:             "/v2/Users",
			body:               strings.NewReader(`{"id": "other", "userName": "test3", "immutableThing": "test"}`),
			expectedUserName:   "test3",
			expectedExternalID: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.target, test.body)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusCreated, rr.Code)

			assertEqual(t, "application/scim+json", rr.Header().Get("Content-Type"))

			var resource map[string]interface{}
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &resource))

			assertEqual(t, test.expectedUserName, resource["userName"])

			assertEqual(t, test.expectedExternalID, resource["externalId"])

			meta, ok := resource["meta"].(map[string]interface{})
			assertTypeOk(t, ok, "object")

			assertEqual(t, "User", meta["resourceType"])
			assertNotNil(t, meta["created"], "created")
			assertNotNil(t, meta["lastModified"], "last modified")
			assertEqual(t, fmt.Sprintf("Users/%s", resource["id"]), meta["location"])
			assertEqual(t, fmt.Sprintf("v%s", resource["id"]), meta["version"])
			// ETag and version needs to be the same.
			assertEqual(t, rr.Header().Get("Etag"), meta["version"])
		})
	}
}

func TestServerResourcePutHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/9999", strings.NewReader(`{"userName": "other"}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusNotFound, rr.Code)

	var scimErr *errors.ScimError
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &scimErr))
	expectedError := &errors.ScimError{
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("Resource %d not found.", 9999),
	}
	assertEqualSCIMErrors(t, expectedError, scimErr)

	if scimErr == nil || scimErr.Status != http.StatusNotFound ||
		scimErr.Detail != fmt.Sprintf("Resource %d not found.", 9999) {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourcePutHandlerValid(t *testing.T) {
	tests := []struct {
		name               string
		target             string
		body               io.Reader
		expectedUserName   string
		expectedExternalID interface{}
	}{
		{
			name:               "Users put request",
			target:             "/v2/Users/0002",
			body:               strings.NewReader(`{"id": "other", "userName": "test2", "externalId": "external_test2"}`),
			expectedUserName:   "test2",
			expectedExternalID: "external_test2",
		}, {
			name:               "Users put request without externalId",
			target:             "/Users/0003",
			body:               strings.NewReader(`{"id": "other", "userName": "test3"}`),
			expectedUserName:   "test3",
			expectedExternalID: nil,
		}, {
			name:               "Users put request with immutable attribute",
			target:             "/Users/0003",
			body:               strings.NewReader(`{"id": "other", "userName": "test3", "immutableThing": "test"}`),
			expectedUserName:   "test3",
			expectedExternalID: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, test.target, test.body)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)

			assertEqual(t, "application/scim+json", rr.Header().Get("Content-Type"))

			var resource map[string]interface{}
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &resource))

			assertEqual(t, test.expectedUserName, resource["userName"])
			assertEqual(t, test.expectedExternalID, resource["externalId"])

			meta, ok := resource["meta"].(map[string]interface{})
			assertTypeOk(t, ok, "meta")
			assertEqual(t, "User", meta["resourceType"])
		})
	}
}

func TestServerResourceTypeHandlerValid(t *testing.T) {
	tests := []struct {
		name          string
		resourceType  string
		versionPrefix string
	}{
		{
			name:         "User schema",
			resourceType: "User",
		}, {
			name:         "Enterprise user schema",
			resourceType: "EnterpriseUser",
		}, {
			name:          "User schema, with base path",
			resourceType:  "User",
			versionPrefix: "/v2",
		}, {
			name:          "Enterprise user schema, with base path",
			resourceType:  "EnterpriseUser",
			versionPrefix: "/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/ResourceTypes/%s", tt.versionPrefix, tt.resourceType), nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)

			var resourceType map[string]interface{}
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &resourceType))

			assertEqual(t, tt.resourceType, resourceType["id"])
		})
	}
}

func TestServerResourceTypesHandler(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "resource types request without version",
			target: "/ResourceTypes",
		}, {
			name:   "resource types request with version",
			target: "/v2/ResourceTypes",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)

			var response listResponse
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

			assertEqual(t, 3, response.TotalResults)
			assertLen(t, response.Resources, 3)

			resourceTypes := make([]string, 3)
			for i, resource := range response.Resources {
				resourceType, ok := resource.(map[string]interface{})
				assertTypeOk(t, ok, "object")
				resourceTypes[i] = resourceType["name"].(string)
			}

			assertEqualStrings(t, []string{"User", "EnterpriseUser", "Group"}, resourceTypes)
		})
	}
}

func TestServerResourcesGetAllHandlerNegativeCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=-1", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	var response listResponse
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertEqual(t, 20, response.TotalResults)
	assertEqual(t, 0, len(response.Resources))
}

func TestServerResourcesGetAllHandlerNonIntCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=BadBanana", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusBadRequest, rr.Code)

	var response errors.ScimError
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertEqual(t, http.StatusBadRequest, response.Status)
	assertEqual(t, "Bad Request. Invalid parameter provided in request: count.", response.Detail)
}

func TestServerResourcesGetAllHandlerNonIntStartIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?startIndex=BadBanana", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusBadRequest, rr.Code)

	var response errors.ScimError
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertEqual(t, http.StatusBadRequest, response.Status)
	assertEqual(t, "Bad Request. Invalid parameter provided in request: startIndex.", response.Detail)
}

func TestServerResourcesGetHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	var response listResponse
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertEqual(t, 20, response.TotalResults)
	assertEqual(t, 20, len(response.Resources))
}

func TestServerResourcesGetHandlerMaxCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=20000", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	var response listResponse
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertEqual(t, 20, response.TotalResults)
}

func TestServerResourcesGetHandlerPagination(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=2&startIndex=2", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	var response listResponse
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertEqual(t, 20, response.TotalResults)
}

func TestServerSchemaEndpointValid(t *testing.T) {
	tests := []struct {
		name          string
		schema        string
		versionPrefix string
	}{
		{
			name:   "User schema",
			schema: "urn:ietf:params:scim:schemas:core:2.0:User",
		}, {
			name:   "Enterprise user schema",
			schema: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		}, {
			name:          "User schema, with base path",
			schema:        "urn:ietf:params:scim:schemas:core:2.0:User",
			versionPrefix: "/v2",
		}, {
			name:          "Enterprise user schema, with base path",
			schema:        "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
			versionPrefix: "/v2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf(
				"%s/Schemas/%s", test.versionPrefix, test.schema,
			), nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)

			var s map[string]interface{}
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &s))
			assertEqual(t, test.schema, s["id"].(string))
		})
	}
}

func TestServerSchemasEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "schemas request without version",
			target: "/Schemas",
		}, {
			name:   "schemas request with version",
			target: "/v2/Schemas",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)

			var response listResponse
			assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))

			expectedLen := 3
			assertEqual(t, expectedLen, response.TotalResults)
			assertLen(t, response.Resources, expectedLen)

			resourceIDs := make([]string, 3)
			for i, resource := range response.Resources {
				resourceType, ok := resource.(map[string]interface{})
				assertTypeOk(t, ok, "object")
				resourceIDs[i] = resourceType["id"].(string)
			}

			assertEqualStrings(t, []string{
				"urn:ietf:params:scim:schemas:core:2.0:User",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
				"urn:ietf:params:scim:schemas:core:2.0:Group",
			}, resourceIDs)
		})
	}
}

func TestServerSchemasEndpointFilter(t *testing.T) {
	params := url.Values{
		"filter": []string{"id co \"extension\""},
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf(
		"/Schemas?%s", params.Encode(),
	), nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assertEqualStatusCode(t, http.StatusOK, rr.Code)

	var response listResponse
	assertUnmarshalNoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assertLen(t, response.Resources, 1)
	assertEqual(t, 3, response.TotalResults)
}

func TestServerServiceProviderConfigHandler(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "service provide config request without version",
			target: "/ServiceProviderConfig",
		}, {
			name:   "service provide config request with version",
			target: "/v2/ServiceProviderConfig",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assertEqualStatusCode(t, http.StatusOK, rr.Code)
		})
	}
}

func getUserExtensionSchema() schema.Schema {
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        optional.NewString("EnterpriseUser"),
		Description: optional.NewString("Enterprise User"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "employeeNumber",
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "organization",
			})),
		},
	}
}

func getUserSchema() schema.Schema {
	return schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        optional.NewString("User"),
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "userName",
				Required:   true,
				Uniqueness: schema.AttributeUniquenessServer(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Name:     "active",
				Required: false,
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "readonlyThing",
				Required:   false,
				Mutability: schema.AttributeMutabilityReadOnly(),
			})),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "immutableThing",
				Required:   false,
				Mutability: schema.AttributeMutabilityImmutable(),
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:     "Name",
				Required: false,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "familyName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "givenName",
					}),
				},
			}),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "displayName",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Name:        "emails",
				MultiValued: true,
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Name: "value",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "display",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Name: "type",
						CanonicalValues: []string{
							"work", "home", "other",
						},
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Name: "primary",
					}),
				},
			}),
		},
	}
}

func newTestResourceHandler() ResourceHandler {
	data := make(map[string]testData)

	// Generate enough test data to test pagination
	for i := 1; i < 21; i++ {
		data[fmt.Sprintf("000%d", i)] = testData{
			resourceAttributes: ResourceAttributes{
				"userName":   fmt.Sprintf("test%02d", i),
				"externalId": fmt.Sprintf("external%d", i),
			},
			meta: map[string]string{
				"created":      fmt.Sprintf("2020-01-%02dT15:04:05+07:00", i),
				"lastModified": fmt.Sprintf("2020-02-%02dT16:05:04+07:00", i),
				"version":      fmt.Sprintf("v%d", i),
			},
		}
	}

	return testResourceHandler{
		data: data,
	}
}

func newTestServer() Server {
	userSchema := getUserSchema()
	userSchemaExtension := getUserExtensionSchema()
	return Server{
		Config: ServiceProviderConfig{},
		ResourceTypes: []ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      userSchema,
				Handler:     newTestResourceHandler(),
			},
			{
				ID:          optional.NewString("EnterpriseUser"),
				Name:        "EnterpriseUser",
				Endpoint:    "/EnterpriseUsers",
				Description: optional.NewString("Enterprise User Account"),
				Schema:      userSchema,
				SchemaExtensions: []SchemaExtension{
					{Schema: userSchemaExtension},
				},
				Handler: newTestResourceHandler(),
			},
			{
				ID:          optional.NewString("Group"),
				Name:        "Group",
				Endpoint:    "/Groups",
				Description: optional.NewString("Group"),
				Schema:      schema.CoreGroupSchema(),
				Handler:     newTestResourceHandler(),
			},
		},
	}
}
