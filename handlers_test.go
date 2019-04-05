package scim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestError(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	NewServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	expected := `error!`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestServer_SchemasHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/Schemas", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	user, _ := NewSchemaFromFile("testdata/simple_user_schema.json")
	NewServer(user).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.TotalResults != 1 {
		t.Errorf("handler returned unexpected body: got %v want 1 total result", rr.Body.String())
	}
}

func TestServer_SchemaHandlerInvalid(t *testing.T) {
	req, err := http.NewRequest("GET", "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	NewServer().ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.TotalResults != 0 {
		t.Errorf("handler returned unexpected body: got %v want no total result", rr.Body.String())
	}
}

func TestServer_SchemaHandlerValid(t *testing.T) {
	req, err := http.NewRequest("GET", "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	user, _ := NewSchemaFromFile("testdata/simple_user_schema.json")
	NewServer(user).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ListResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	if response.TotalResults != 1 {
		t.Errorf("handler returned unexpected body: got %v want 1 total result", rr.Body.String())
	}
}
