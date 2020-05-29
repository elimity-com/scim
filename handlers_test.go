package scim

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func newTestServer(basePath string) Server {
	userSchema := getUserSchema()

	userSchemaExtension := getUserExtensionSchema()

	return Server{
		Config: ServiceProviderConfig{
			BasePathResolver: func(r *http.Request) string {
				return basePath
			},
		},
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
				"userName": fmt.Sprintf("test%d", i),
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

func TestInvalidEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		basePath       string
		target         string
		body           io.Reader
		expectedStatus int
	}{
		{
			name:           "invalid get request, no base path",
			method:         http.MethodGet,
			target:         "/v2/Invalid",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid get request, with base path",
			method:         http.MethodGet,
			basePath:       "/my/test/base/path",
			target:         "/my/test/base/path/v2/Invalid",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid get request, outside base path",
			method:         http.MethodGet,
			basePath:       "my/test/base/path/v2",
			target:         "/v2/Invalid",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid schema request, no base path",
			method:         http.MethodGet,
			target:         "/Schemas/urn:ietf:params:scim:schemas:core:2.0:Group",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid schema request, with base path",
			method:         http.MethodGet,
			basePath:       "/my/test/base/path",
			target:         "/my/test/base/path/Schemas/urn:ietf:params:scim:schemas:core:2.0:Group",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid resource types request, no base path",
			method:         http.MethodGet,
			target:         "/ResourceTypes/Group",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid resource types request, with base path",
			method:         http.MethodGet,
			basePath:       "/my/test/base/path",
			target:         "/my/test/base/path/ResourceTypes/Group",
			expectedStatus: http.StatusNotFound,
		}, {
			name:           "invalid post request, no base path",
			method:         http.MethodPost,
			target:         "/Users",
			body:           strings.NewReader(`{"id": "other"}`),
			expectedStatus: http.StatusBadRequest,
		}, {
			name:           "invalid post request, with base path",
			method:         http.MethodPost,
			basePath:       "/my/test/base/path",
			target:         "/my/test/base/path/Users",
			body:           strings.NewReader(`{"id": "other"}`),
			expectedStatus: http.StatusBadRequest,
		}, {
			name:           "invalid put request, no base path",
			method:         http.MethodPut,
			target:         "/Users/0001",
			body:           strings.NewReader(`{"more": "test"}`),
			expectedStatus: http.StatusBadRequest,
		}, {
			name:           "invalid put request, with base path",
			method:         http.MethodPut,
			basePath:       "/my/test/base/path",
			target:         "/my/test/base/path/Users/0001",
			body:           strings.NewReader(`{"more": "test"}`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}
}

func TestServerSchemasEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		target   string
	}{
		{
			name:   "schemas request without version",
			target: "/Schemas",
		}, {
			name:   "schemas request with version",
			target: "/v2/Schemas",
		}, {
			name:     "schemas request without version, with base path",
			basePath: "/my/test/base/path",
			target:   "/my/test/base/path/Schemas",
		}, {
			name:     "schemas request with version, with base path",
			basePath: "/my/test/base/path",
			target:   "/my/test/base/path/v2/Schemas",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			var response listResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Error(err)
			}

			if response.TotalResults != 2 {
				t.Errorf("handler returned unexpected body: got %v want 2 total result", rr.Body.String())
			}

			if len(response.Resources) != 2 {
				t.Fatal("resources contains more than one schema")
			}

			s, ok := response.Resources[0].(map[string]interface{})
			if !ok {
				t.Fatal("schema is not an object")
			}

			if s["id"].(string) != "urn:ietf:params:scim:schemas:core:2.0:User" &&
				s["id"].(string) != "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User" {
				t.Errorf("schema does not contain the correct id: %v", s["id"])
			}
		})
	}
}

func TestServerSchemaEndpointValid(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		schema   string
	}{
		{
			name:   "User schema",
			schema: "urn:ietf:params:scim:schemas:core:2.0:User",
		}, {
			name:   "Enterprice user schema",
			schema: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		}, {
			name:     "User schema, with base path",
			schema:   "urn:ietf:params:scim:schemas:core:2.0:User",
			basePath: "/my/test/base/path",
		}, {
			name:     "Enterprice user schema, with base path",
			schema:   "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
			basePath: "/my/test/base/path",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/Schemas/%s", tt.basePath, tt.schema), nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			var s map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &s); err != nil {
				t.Fatal(err)
			}

			if s["id"].(string) != tt.schema {
				t.Errorf("schema does not contain the correct id: %s", s["id"])
			}
		})
	}
}

func TestServerResourceTypesHandler(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		target   string
	}{
		{
			name:   "resource types request without version",
			target: "/ResourceTypes",
		}, {
			name:   "resource types request with version",
			target: "/v2/ResourceTypes",
		}, {
			name:     "resource types request without version, with base path",
			basePath: "/my/test/base/path",
			target:   "/my/test/base/path/ResourceTypes",
		}, {
			name:     "resource types request with version, with base path",
			basePath: "/my/test/base/path",
			target:   "/my/test/base/path/v2/ResourceTypes",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			var response listResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}

			if response.TotalResults != 2 {
				t.Errorf("handler returned unexpected body: got %v want 1 total result", rr.Body.String())
			}

			if len(response.Resources) != 2 {
				t.Fatal("resources contains more than one schema")
			}

			resourceType, ok := response.Resources[0].(map[string]interface{})
			if !ok {
				t.Errorf("resource type is not an object")
			}

			if resourceType["name"].(string) != "User" &&
				resourceType["name"].(string) != "EnterpriseUser" {
				t.Errorf("schema does not contain the correct id: %v", resourceType["name"])
			}
		})
	}
}

func TestServerResourceTypeHandlerValid(t *testing.T) {
	tests := []struct {
		name         string
		basePath     string
		resourceType string
	}{
		{
			name:         "User schema",
			resourceType: "User",
		}, {
			name:         "Enterprice user schema",
			resourceType: "EnterpriseUser",
		}, {
			name:         "User schema, with base path",
			resourceType: "User",
			basePath:     "/my/test/base/path",
		}, {
			name:         "Enterprice user schema, with base path",
			resourceType: "EnterpriseUser",
			basePath:     "/my/test/base/path",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/ResourceTypes/%s", tt.basePath, tt.resourceType), nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			var resourceType map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &resourceType); err != nil {
				t.Fatal(err)
			}
			if resourceType["id"] != tt.resourceType {
				t.Errorf("schema does not contain the correct name: %s", resourceType["name"])
			}
		})
	}
}

func TestServerServiceProviderConfigHandler(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		target   string
	}{
		{
			name:   "service provide config request without version",
			target: "/ServiceProviderConfig",
		}, {
			name:   "service provide config request with version",
			target: "/v2/ServiceProviderConfig",
		}, {
			name:     "service provide config request without version, with base path",
			basePath: "/my/test/base/path",
			target:   "/my/test/base/path/ServiceProviderConfig",
		}, {
			name:     "service provide config request with version, with base path",
			basePath: "/my/test/base/path",
			target:   "/my/test/base/path/v2/ServiceProviderConfig",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}
		})
	}
}

func TestServerResourcePostHandlerValid(t *testing.T) {
	tests := []struct {
		name             string
		basePath         string
		target           string
		body             io.Reader
		expectedUserName string
	}{
		{
			name:             "Users post request without version",
			target:           "/Users",
			body:             strings.NewReader(`{"id": "other", "userName": "test1"}`),
			expectedUserName: "test1",
		}, {
			name:             "Users post request with version",
			target:           "/v2/Users",
			body:             strings.NewReader(`{"id": "other", "userName": "test2"}`),
			expectedUserName: "test2",
		}, {
			name:             "Users post request without version, with base path",
			basePath:         "/my/test/base/path",
			target:           "/my/test/base/path/Users",
			body:             strings.NewReader(`{"id": "other", "userName": "test3"}`),
			expectedUserName: "test3",
		}, {
			name:             "Users post request with version, with base path",
			basePath:         "/my/test/base/path",
			target:           "/my/test/base/path/v2/Users",
			body:             strings.NewReader(`{"id": "other", "userName": "test4"}`),
			expectedUserName: "test4",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.target, tt.body)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusCreated {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
			}

			if rr.Header().Get("Content-Type") != "application/scim+json" {
				t.Error("handler did not return the header content type correctly")
			}

			var resource map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
				t.Fatal(err)
			}

			if resource["userName"] != tt.expectedUserName {
				t.Error("handler did not return the resource correctly")
			}

			meta, ok := resource["meta"].(map[string]interface{})
			if !ok {
				t.Error("handler did not return the resource meta correctly")
			}

			if meta["resourceType"] != "User" {
				t.Error("handler did not return the resource meta resource type correctly")
			}

			if len(fmt.Sprintf("%v", meta["created"])) == 0 {
				t.Error("handler did not return the resource meta created correctly")
			}

			if len(fmt.Sprintf("%v", meta["lastModified"])) == 0 {
				t.Error("handler did not return the resource meta last modified correctly")
			}

			if meta["location"] != strings.TrimPrefix(fmt.Sprintf("%s/Users/%s", tt.basePath, resource["id"]), "/") {
				t.Error("handler did not return the resource meta location correctly", meta["location"])
			}

			if meta["version"] != fmt.Sprintf("v%s", resource["id"]) {
				t.Error("handler did not return the resource meta version correctly")
			}

			if rr.Header().Get("Etag") != meta["version"] {
				t.Error("handler did not return the header entity tag correctly")
			}
		})
	}

}

func TestServerResourceGetHandler(t *testing.T) {

	tests := []struct {
		name                 string
		basePath             string
		target               string
		expectedUserName     string
		expectedVersion      string
		expectedCreated      string
		expectedLastModified string
	}{
		{
			name:                 "Users get request without version",
			target:               "/Users/0001",
			expectedUserName:     "test1",
			expectedVersion:      "v1",
			expectedCreated:      "2020-01-01T15:04:05+07:00",
			expectedLastModified: "2020-02-01T16:05:04+07:00",
		}, {
			name:                 "Users get request with version",
			target:               "/v2/Users/0002",
			expectedUserName:     "test2",
			expectedVersion:      "v2",
			expectedCreated:      "2020-01-02T15:04:05+07:00",
			expectedLastModified: "2020-02-02T16:05:04+07:00",
		}, {
			name:                 "Users get request without version, with base path",
			basePath:             "/my/test/base/path",
			target:               "/my/test/base/path/Users/0003",
			expectedUserName:     "test3",
			expectedVersion:      "v3",
			expectedCreated:      "2020-01-03T15:04:05+07:00",
			expectedLastModified: "2020-02-03T16:05:04+07:00",
		}, {
			name:                 "Users get request with version, with base path",
			basePath:             "/my/test/base/path",
			target:               "/my/test/base/path/v2/Users/0004",
			expectedUserName:     "test4",
			expectedVersion:      "v4",
			expectedCreated:      "2020-01-04T15:04:05+07:00",
			expectedLastModified: "2020-02-04T16:05:04+07:00",
		},
	}

	for _, tt := range tests {
		tt := tt // scopelint
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rr := httptest.NewRecorder()
			newTestServer(tt.basePath).ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			if rr.Header().Get("Content-Type") != "application/scim+json" {
				t.Error("handler did not return the header content type correctly")
			}

			if rr.Header().Get("Etag") != tt.expectedVersion {
				t.Error("handler did not return the header entity tag correctly")
			}

			var resource map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
				t.Fatal(err)
			}

			if resource["userName"] != tt.expectedUserName {
				t.Error("handler did not return the resource correctly")
			}

			meta, ok := resource["meta"].(map[string]interface{})
			if !ok {
				t.Error("handler did not return the resource meta correctly")
			}

			if meta["resourceType"] != "User" {
				t.Error("handler did not return the resource meta resource type correctly")
			}

			if meta["created"] != tt.expectedCreated {
				t.Error("handler did not return the resource meta created correctly")
			}

			if meta["lastModified"] != tt.expectedLastModified {
				t.Error("handler did not return the resource meta last modified correctly")
			}

			if meta["location"] != strings.TrimPrefix(fmt.Sprintf("%s/Users/%s", tt.basePath, resource["id"]), "/") {
				t.Error("handler did not return the resource meta location correctly", meta["location"])
			}

			if meta["version"] != tt.expectedVersion {
				t.Error("handler did not return the resource meta version correctly")
			}
		})
	}
}

func TestServerResourceGetHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var scimErr scimError
	if err := json.Unmarshal(rr.Body.Bytes(), &scimErr); err != nil {
		t.Error(err)
	}
	if scimErr != scimErrorResourceNotFound("9999") {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourcesGetHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users", nil)
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Error(err)
	}

	if response.TotalResults != 20 {
		t.Errorf("handler returned unexpected body: got %v want 20 total result", response.TotalResults)
	}
}

func TestServerResourcesGetHandlerPagination(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=2&startIndex=2", nil)
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Error(err)
	}

	if response.TotalResults != 20 {
		t.Errorf("handler returned unexpected body: got %v want 20 total result", response.TotalResults)
	}
}

func TestServerResourcesGetHandlerMaxCount(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users?count=20000", nil)
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Error(err)
	}

	if response.TotalResults != 20 {
		t.Errorf("handler returned unexpected body: got %v want 20 total result", response.TotalResults)
	}
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
		    "op":"remove",
		    "path":"displayName"
		  }
		]
	}`))
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Header().Get("Content-Type") != "application/scim+json" {
		t.Error("handler did not return the header content type correctly")
	}

	expectedVersion := "v1.patch"

	if rr.Header().Get("Etag") != expectedVersion {
		t.Error("handler did not return the header entity tag correctly")
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Error response: %v\n", resource)
	}

	if resource["displayName"] != nil {
		t.Errorf("handler did not remove the displayName attribute")
	}

	if resource["active"] != false {
		t.Errorf("handler did not deactivate user")
	}

	if resource["emails"] == nil || len(resource["emails"].([]interface{})) < 1 {
		t.Errorf("handler did not add user's email address")
	}

	meta, ok := resource["meta"].(map[string]interface{})
	if !ok {
		t.Error("handler did not return the resource meta correctly")
	}

	if meta["resourceType"] != "User" {
		t.Error("handler did not return the resource meta resource type correctly")
	}

	if meta["created"] != "2020-01-01T15:04:05+07:00" {
		t.Error("handler did not return the resource meta created correctly")
	}

	if meta["lastModified"] == "2020-02-01T16:05:04+07:00" {
		t.Error("handler did not return the resource meta last modified correctly")
	}

	if meta["location"] != "Users/0001" {
		t.Error("handler did not return the resource meta version correctly")
	}

	if meta["version"] != expectedVersion {
		t.Error("handler did not return the resource meta version correctly")
	}
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
	newTestServer("").ServeHTTP(rr, req)

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Error response: %v\n", resource)
	}
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
	newTestServer("").ServeHTTP(rr, req)

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Error response: %v\n", resource)
	}
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
	newTestServer("").ServeHTTP(rr, req)

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != expectedStatus {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Logf("Error response: %v\n", resource)
	}
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
	req := httptest.NewRequest(http.MethodPut, "/Users/0001", strings.NewReader(`{"id": "test", "userName": "other"}`))
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	if resource["userName"] != "other" {
		t.Errorf("handler did not replace previous resource")
	}
}

func TestServerResourcePutHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/9999", strings.NewReader(`{"userName": "other"}`))
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var scimErr scimError
	if err := json.Unmarshal(rr.Body.Bytes(), &scimErr); err != nil {
		t.Error(err)
	}

	if scimErr != scimErrorResourceNotFound("9999") {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourceDeleteHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/0001", nil)
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestServerResourceDeleteHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer("").ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var scimErr scimError
	if err := json.Unmarshal(rr.Body.Bytes(), &scimErr); err != nil {
		t.Error(err)
	}

	if scimErr != scimErrorResourceNotFound("9999") {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}
