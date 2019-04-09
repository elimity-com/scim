package scim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErr(t *testing.T) {
	req := httptest.NewRequest("GET", "/Invalid", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	NewServer(config, nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerSchemasHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/Schemas", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	user, _ := NewSchemaFromFile("testdata/simple_user_schema.json")
	NewServer(config, []Schema{user}, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
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

	schema, ok := schemas[0].(map[string]interface{})
	if !ok {
		t.Errorf("schema is not an object")
	}

	if schema["ID"].(string) != "urn:ietf:params:scim:schemas:core:2.0:User" {
		t.Errorf("schema does not contain the correct id: %v", schema["ID"])
	}
}

func TestServerSchemaHandlerInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	NewServer(config, nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerSchemaHandlerValid(t *testing.T) {
	req := httptest.NewRequest("GET", "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	user, _ := NewSchemaFromFile("testdata/simple_user_schema.json")
	NewServer(config, []Schema{user}, nil).ServeHTTP(rr, req)

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
	req := httptest.NewRequest("GET", "/ResourceTypes", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	user, _ := NewResourceTypeFromFile("testdata/simple_user_resource_type.json")
	NewServer(config, nil, []ResourceType{user}).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response listResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
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
	req := httptest.NewRequest("GET", "/ResourceTypes/User", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	NewServer(config, nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServerResourceTypeHandlerValid(t *testing.T) {
	req := httptest.NewRequest("GET", "/ResourceTypes/User", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	user, _ := NewResourceTypeFromFile("testdata/simple_user_resource_type.json")
	NewServer(config, nil, []ResourceType{user}).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resourceType resourceType
	json.Unmarshal(rr.Body.Bytes(), &resourceType)
	if resourceType.ID != "User" {
		t.Errorf("schema does not contain the correct name: %s", resourceType.Name)
	}
}

func TestServerServiceProviderConfigHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/ServiceProviderConfig", nil)
	rr := httptest.NewRecorder()
	config, err := NewServiceProviderConfigFromFile("testdata/simple_service_provider_config.json")
	if err != nil {
		t.Error(err)
	}
	NewServer(config, nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
