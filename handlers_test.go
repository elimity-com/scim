package scim

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"

	"github.com/stretchr/testify/assert"
)

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
				Endpoint:    "/EnterpriseUser",
				Description: optional.NewString("Enterprise User Account"),
				Schema:      userSchema,
				SchemaExtensions: []SchemaExtension{
					{Schema: userSchemaExtension},
				},
				Handler: newTestResourceHandler(),
			},
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

func newTestResourceHandler() ResourceHandler {
	data := make(map[string]testData)

	// Generate enough test data to test pagination
	for i := 1; i < 21; i++ {
		data[fmt.Sprintf("000%d", i)] = testData{
			resourceAttributes: ResourceAttributes{
				"userName":   fmt.Sprintf("test%d", i),
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
			target:         "/Schemas/urn:ietf:params:scim:schemas:core:2.0:Group",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid resource types request",
			method:         http.MethodGet,
			target:         "/ResourceTypes/Group",
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

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "status code mismatch")
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

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

			var response listResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, 2, response.TotalResults)

			assert.Len(t, response.Resources, 2)

			resourceIDs := make([]string, 2)
			for i, resource := range response.Resources {
				resourceType, ok := resource.(map[string]interface{})
				assert.True(t, ok, "schema is not an object")
				resourceIDs[i] = resourceType["id"].(string)
			}

			assert.Equal(
				t,
				[]string{
					"urn:ietf:params:scim:schemas:core:2.0:User",
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
				}, resourceIDs)
		})
	}
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
			name:   "Enterprice user schema",
			schema: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		}, {
			name:          "User schema, with base path",
			schema:        "urn:ietf:params:scim:schemas:core:2.0:User",
			versionPrefix: "/v2",
		}, {
			name:          "Enterprice user schema, with base path",
			schema:        "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
			versionPrefix: "/v2",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/Schemas/%s", tt.versionPrefix, tt.schema), nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

			var s map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &s)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, tt.schema, s["id"].(string))
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

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

			var response listResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, 2, response.TotalResults)
			assert.Len(t, response.Resources, 2, "unexpected or missing resources")

			resourceTypes := make([]string, 2)
			for i, resource := range response.Resources {
				resourceType, ok := resource.(map[string]interface{})
				assert.True(t, ok, "resource type is not an object")
				resourceTypes[i] = resourceType["name"].(string)
			}

			assert.Equal(t, []string{"User", "EnterpriseUser"}, resourceTypes)
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
			name:         "Enterprice user schema",
			resourceType: "EnterpriseUser",
		}, {
			name:          "User schema, with base path",
			resourceType:  "User",
			versionPrefix: "/v2",
		}, {
			name:          "Enterprice user schema, with base path",
			resourceType:  "EnterpriseUser",
			versionPrefix: "/v2",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/ResourceTypes/%s", tt.versionPrefix, tt.resourceType), nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

			var resourceType map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &resourceType)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, tt.resourceType, resourceType["id"])
		})
	}
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
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")
		})
	}
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
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.target, tt.body)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusCreated, rr.Code, "status code mismatch")

			assert.Equal(t, "application/scim+json", rr.Header().Get("Content-Type"))

			var resource map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &resource)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, tt.expectedUserName, resource["userName"])

			assert.Equal(t, tt.expectedExternalID, resource["externalId"])

			meta, ok := resource["meta"].(map[string]interface{})
			assert.True(t, ok, "handler did not return the resource meta correctly")

			assert.Equal(t, "User", meta["resourceType"])
			assert.NotEmpty(t, meta["created"], "missing meta created")
			assert.NotEmpty(t, meta["lastModified"], "missing meta last modified")
			assert.Equal(t, fmt.Sprintf("Users/%s", resource["id"]), meta["location"])
			assert.Equal(t, fmt.Sprintf("v%s", resource["id"]), meta["version"])
			assert.Equal(t, rr.Header().Get("Etag"), meta["version"], "ETag and version needs to be the same")
		})
	}
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
			expectedUserName:     "test1",
			expectedExternalID:   "external1",
			expectedVersion:      "v1",
			expectedCreated:      "2020-01-01T15:04:05+07:00",
			expectedLastModified: "2020-02-01T16:05:04+07:00",
		}, {
			name:                 "Users get request with version",
			target:               "/v2/Users/0002",
			expectedUserName:     "test2",
			expectedExternalID:   "external2",
			expectedVersion:      "v2",
			expectedCreated:      "2020-01-02T15:04:05+07:00",
			expectedLastModified: "2020-02-02T16:05:04+07:00",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

			assert.Equal(t, "application/scim+json", rr.Header().Get("Content-Type"))

			assert.Equal(t, tt.expectedVersion, rr.Header().Get("Etag"))

			var resource map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &resource)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, tt.expectedUserName, resource["userName"])

			assert.Equal(t, tt.expectedExternalID, resource["externalId"])

			meta, ok := resource["meta"].(map[string]interface{})
			assert.True(t, ok, "handler did not return the resource meta correctly")

			assert.Equal(t, "User", meta["resourceType"])
			assert.Equal(t, tt.expectedCreated, meta["created"])
			assert.Equal(t, tt.expectedLastModified, meta["lastModified"])
			assert.Equal(t, fmt.Sprintf("Users/%s", resource["id"]), meta["location"])
			assert.Equal(t, tt.expectedVersion, meta["version"])
		})
	}
}

func TestServerResourceGetHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "status code mismatch")

	var scimErr *errors.ScimError
	err := json.Unmarshal(rr.Body.Bytes(), &scimErr)
	assert.NoError(t, err, "json unmarshalling failed")

	expectedError := &errors.ScimError{
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("Resource %d not found.", 9999),
	}
	assert.Equal(t, expectedError, scimErr)
}

func TestServerResourcesGetHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

	var response listResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, 20, response.TotalResults)
}

func TestServerResourcesGetHandlerPagination(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=2&startIndex=2", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

	var response listResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, 20, response.TotalResults)
}

func TestServerResourcesGetHandlerMaxCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=20000", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

	var response listResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, 20, response.TotalResults)
}

// Tests valid add, replace, and remove operations
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

	assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

	assert.Equal(t, "application/scim+json", rr.Header().Get("Content-Type"))

	expectedVersion := "v1.patch"

	assert.Equal(t, expectedVersion, rr.Header().Get("Etag"))

	var resource map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resource)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

	assert.Nil(t, resource["displayName"], "handler did not remove the displayName attribute")
	assert.False(t, resource["active"].(bool), "handler did not deactivate user")
	assert.Equal(t, "external_test_replace", resource["externalId"])

	if resource["emails"] == nil || len(resource["emails"].([]interface{})) < 1 {
		t.Errorf("handler did not add user's email address")
	}

	meta, ok := resource["meta"].(map[string]interface{})
	assert.True(t, ok, "handler did not return the resource meta correctly")

	assert.Equal(t, "User", meta["resourceType"])
	assert.Equal(t, "2020-01-01T15:04:05+07:00", meta["created"])
	assert.NotEqual(t, "2020-02-01T16:05:04+07:00", meta["lastModified"])
	assert.Equal(t, "Users/0001", meta["location"])
	assert.Equal(t, expectedVersion, meta["version"])
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
	err := json.Unmarshal(rr.Body.Bytes(), &resource)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, http.StatusBadRequest, rr.Code, "status code mismatch")
}

func TestServerResourcePatchHandlerFailOnUndefinedAttribute(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"add",
		    "value":{
		      "notActuallyAThing": "adfad"
		    }
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	var resource map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resource)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, http.StatusBadRequest, rr.Code, "status code mismatch")
}

func runPatchImmutableTest(t *testing.T, op, path string, expectedStatus int) {
	req := httptest.NewRequest(http.MethodPatch, "/Users/0001", strings.NewReader(fmt.Sprintf(`{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations":[
		  {
		    "op":"%s",
		    "path":"%s",
		    "value":"test"
		  }
		]
	}`, op, path)))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	var resource map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resource)
	assert.NoError(t, err, "json unmarshalling failed")

	assert.Equal(t, expectedStatus, rr.Code, "status code mismatch")
}

// Ensure we error when changing an immutable or readonly property while allowing adding of immutable properties.
func TestServerResourcePatchHandlerFailOnImmutable(t *testing.T) {
	runPatchImmutableTest(t, PatchOperationAdd, "immutableThing", http.StatusOK)
	runPatchImmutableTest(t, PatchOperationRemove, "immutableThing", http.StatusBadRequest)
	runPatchImmutableTest(t, PatchOperationReplace, "immutableThing", http.StatusBadRequest)
	runPatchImmutableTest(t, PatchOperationReplace, "readonlyThing", http.StatusBadRequest)
	runPatchImmutableTest(t, PatchOperationRemove, "readonlyThing", http.StatusBadRequest)
	runPatchImmutableTest(t, PatchOperationReplace, "readonlyThing", http.StatusBadRequest)
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
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, tt.target, tt.body)
			rr := httptest.NewRecorder()
			newTestServer().ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "status code mismatch")

			assert.Equal(t, "application/scim+json", rr.Header().Get("Content-Type"))

			var resource map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &resource)
			assert.NoError(t, err, "json unmarshalling failed")

			assert.Equal(t, tt.expectedUserName, resource["userName"])

			assert.Equal(t, tt.expectedExternalID, resource["externalId"])

			meta, ok := resource["meta"].(map[string]interface{})
			assert.True(t, ok, "handler did not return the resource meta correctly")

			assert.Equal(t, "User", meta["resourceType"])
		})
	}
}

func TestServerResourcePutHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/9999", strings.NewReader(`{"userName": "other"}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "status code mismatch")

	var scimErr *errors.ScimError
	err := json.Unmarshal(rr.Body.Bytes(), &scimErr)
	assert.NoError(t, err, "json unmarshalling failed")

	expectedError := &errors.ScimError{
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("Resource %d not found.", 9999),
	}
	assert.Equal(t, expectedError, scimErr)

	if scimErr == nil || scimErr.Status != http.StatusNotFound ||
		scimErr.Detail != fmt.Sprintf("Resource %d not found.", 9999) {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourceDeleteHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/0001", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code, "status code mismatch")
}

func TestServerResourceDeleteHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "status code mismatch")

	var scimErr *errors.ScimError
	err := json.Unmarshal(rr.Body.Bytes(), &scimErr)
	assert.NoError(t, err, "json unmarshalling failed")

	expectedError := &errors.ScimError{
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("Resource %d not found.", 9999),
	}
	assert.Equal(t, expectedError, scimErr, "wrong scim error")
}
