package scim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() (*Server, error) {
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		return nil, err
	}
	userSchema, err := NewSchemaFromFile("testdata/simple_user_schema.json")
	if err != nil {
		return nil, err
	}
	userSchemaExtension, err := NewSchemaFromFile("testdata/simple_user_extension_schema.json")
	if err != nil {
		return nil, err
	}
	userResourceType, err := NewResourceTypeFromFile("testdata/simple_user_resource_type_with_extension.json", newTestResourceHandler())
	if err != nil {
		return nil, err
	}
	server, err := NewServer(config, []Schema{userSchema, userSchemaExtension}, []ResourceType{userResourceType})
	if err != nil {
		return nil, err
	}

	return &server, err
}

func TestErr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Invalid", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerSchemasHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Schemas", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}
	if response.TotalResults != 2 {
		t.Errorf("handler returned unexpected body: got %v want 2 total result", rr.Body.String())
	}

	schemas, ok := response.Resources.([]interface{})
	if !ok {
		t.Errorf("resources is not a list of objects")
	}

	if len(schemas) != 2 {
		t.Errorf("resources does not contain two schemas")
		return
	}

	schema, ok := schemas[1].(map[string]interface{})
	if !ok {
		t.Errorf("schema is not an object")
	}

	if schema["id"].(string) != "urn:ietf:params:scim:schemas:core:2.0:User" &&
		schema["id"].(string) != "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User" {
		t.Errorf("schema does not contain the correct id: %v", schema["id"])
	}
}

func TestServerSchemaHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Schemas/urn:ietf:params:scim:schemas:core:2.0:Group", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerSchemaHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var schema schema
	if err := json.Unmarshal(rr.Body.Bytes(), &schema); err != nil {
		t.Error(err)
	}

	if schema.ID != "urn:ietf:params:scim:schemas:core:2.0:User" {
		t.Errorf("schema does not contain the correct id: %s", schema.ID)
	}
}

func TestServerResourceTypesHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ResourceTypes", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}
	if response.TotalResults != 1 {
		t.Errorf("handler returned unexpected body: got %v want 1 total result", rr.Body.String())
	}

	schemas, ok := response.Resources.([]interface{})
	if !ok {
		t.Errorf("resources is not a list of objects")
	}

	if len(schemas) != 1 {
		t.Errorf("resources contains more than one schema")
		return
	}

	resourceType, ok := schemas[0].(map[string]interface{})
	if !ok {
		t.Errorf("schema is not an object")
	}

	if resourceType["name"].(string) != "User" {
		t.Errorf("schema does not contain the correct id: %v", resourceType["Name"])
	}
}

func TestServerResourceTypeHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ResourceTypes/Group", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerResourceTypeHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ResourceTypes/User", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resourceType resourceType
	err = json.Unmarshal(rr.Body.Bytes(), &resourceType)
	if err != nil {
		t.Error(err)
	}
	if resourceType.ID != "User" {
		t.Errorf("schema does not contain the correct name: %s", resourceType.Name)
	}
}

func TestServerServiceProviderConfigHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ServiceProviderConfig", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServerResourcePostHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/Users", strings.NewReader(`{"id": "other"}`))
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var scimErr scimError
	err = json.Unmarshal(rr.Body.Bytes(), &scimErr)
	if err != nil {
		t.Error(err)
	}
	if scimErr != scimErrorInvalidValue {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourcePostHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/Users", strings.NewReader(`{"id": "other", "userName": "test"}`))
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resource map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resource)
	if err != nil {
		t.Error(err)
	}
	if resource["userName"] != "test" {
		t.Errorf("handler did not return the resource correctly")
	}
}

func TestServerResourceGetHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/0001", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resource map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resource)
	if err != nil {
		t.Error(err)
	}
	if resource["userName"] != "test" {
		t.Errorf("handler did not return the resource correctly")
	}
}

func TestServerResourceGetHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users/9999", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var scimErr scimError
	err = json.Unmarshal(rr.Body.Bytes(), &scimErr)
	if err != nil {
		t.Error(err)
	}
	if scimErr != scimErrorResourceNotFound("9999") {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourcesGetHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/Users", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	var response listResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Error(err)
	}

	if response.TotalResults != 1 {
		t.Errorf("handler returned unexpected body: got %v want 1 total result", rr.Body.String())
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServerResourcePutHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/0001", strings.NewReader(`{"more": "test"}`))
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var scimErr scimError
	err = json.Unmarshal(rr.Body.Bytes(), &scimErr)
	if err != nil {
		t.Error(err)
	}
	if scimErr != scimErrorInvalidValue {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourcePutHandlerValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/0001", strings.NewReader(`{"id": "test", "userName": "other"}`))
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	var resource map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &resource)
	if err != nil {
		t.Error(err)
	}
	if resource["userName"] != "other" {
		t.Errorf("handler did not replace previous resource")
	}
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServerResourcePutHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/Users/9999", strings.NewReader(`{"userName": "other"}`))
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var scimErr scimError
	err = json.Unmarshal(rr.Body.Bytes(), &scimErr)
	if err != nil {
		t.Error(err)
	}
	if scimErr != scimErrorResourceNotFound("9999") {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}

func TestServerResourceDeleteHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/0001", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestServerResourceDeleteHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/Users/9999", nil)
	rr := httptest.NewRecorder()

	server, err := newTestServer()
	if err != nil {
		t.Error(err)
	}
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	var scimErr scimError
	err = json.Unmarshal(rr.Body.Bytes(), &scimErr)
	if err != nil {
		t.Error(err)
	}
	if scimErr != scimErrorResourceNotFound("9999") {
		t.Errorf("wrong scim error: %v", scimErr)
	}
}
