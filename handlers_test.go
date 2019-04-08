package scim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErr(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	NewServer(nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServer_SchemasHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/Schemas", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	user, _ := NewSchemaFromFile("testdata/simple_user_schema.json")
	NewServer([]Schema{user}, nil).ServeHTTP(rr, req)

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

func TestServer_SchemaHandlerInvalid(t *testing.T) {
	req, err := http.NewRequest("GET", "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	NewServer(nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServer_SchemaHandlerValid(t *testing.T) {
	req, err := http.NewRequest("GET", "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	user, _ := NewSchemaFromFile("testdata/simple_user_schema.json")
	NewServer([]Schema{user}, nil).ServeHTTP(rr, req)

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

func TestServer_ResourceTypesHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/ResourceTypes", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	user, _ := NewResourceTypeFromFile("testdata/simple_user_resource_type.json")
	NewServer(nil, []ResourceType{user}).ServeHTTP(rr, req)

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

func TestServer_ResourceTypeHandlerInvalid(t *testing.T) {
	req, err := http.NewRequest("GET", "/ResourceTypes/User", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	NewServer(nil, nil).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestServer_ResourceTypeHandlerValid(t *testing.T) {
	req, err := http.NewRequest("GET", "/ResourceTypes/User", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	user, _ := NewResourceTypeFromFile("testdata/simple_user_resource_type.json")
	NewServer(nil, []ResourceType{user}).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resourceType resourceType
	json.Unmarshal(rr.Body.Bytes(), &resourceType)
	if resourceType.ID != "User" {
		t.Errorf("schema does not contain the correct name: %s", resourceType.Name)
	}
}