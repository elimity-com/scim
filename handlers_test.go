package scim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func newTestServer() Server {
	userSchema := schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
		Name:        "User",
		Description: optional.NewString("User Account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:       "userName",
				Required:   true,
				Uniqueness: schema.AttributeUniquenessServer(),
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

	userSchemaExtension := schema.Schema{
		ID:          "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
		Name:        "EnterpriseUser",
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

func newTestResourceHandler() ResourceHandler {
	data := make(map[string]ResourceAttributes)
	data["0001"] = ResourceAttributes{
		"userName": "test",
	}

	return testResourceHandler{
		data: data,
	}
}

func TestInvalidEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v2/Invalid", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerSchemasEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Schemas", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

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
}

func TestServerSchemaEndpointInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Schemas/urn:ietf:params:scim:schemas:core:2.0:Group", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

}

func TestServerSchemaEndpointValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var s map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &s); err != nil {
		t.Fatal(err)
	}

	if s["id"].(string) != "urn:ietf:params:scim:schemas:core:2.0:User" {
		t.Errorf("schema does not contain the correct id: %s", s["id"])
	}
}

func TestServerResourceTypesHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ResourceTypes", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

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
}

func TestServerResourceTypeHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ResourceTypes/Group", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerResourceTypeHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ResourceTypes/User", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resourceType map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resourceType); err != nil {
		t.Fatal(err)
	}
	if resourceType["id"] != "User" {
		t.Errorf("schema does not contain the correct name: %s", resourceType["name"])
	}
}

func TestServerServiceProviderConfigHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ServiceProviderConfig", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServerResourcePostHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/Users", strings.NewReader(`{"id": "other"}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServerResourcePostHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/Users", strings.NewReader(`{"id": "other", "userName": "test"}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	if resource["userName"] != "test" {
		t.Error("handler did not return the resource correctly")
	}
}

func TestServerResourceGetHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/0001", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	if resource["userName"] != "test" {
		t.Error("handler did not return the resource correctly")
	}
}

func TestServerResourceGetHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

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
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Error(err)
	}

	if response.TotalResults != 1 {
		t.Errorf("handler returned unexpected body: got %v want 1 total result", rr.Body.String())
	}
}

func TestServerResourcePutHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/0001", strings.NewReader(`{"more": "test"}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestServerResourcePutHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/0001", strings.NewReader(`{"id": "test", "userName": "other"}`))
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

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
	newTestServer().ServeHTTP(rr, req)

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
	newTestServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestServerResourceDeleteHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/9999", nil)
	rr := httptest.NewRecorder()
	newTestServer().ServeHTTP(rr, req)

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
