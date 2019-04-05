package scim

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func errorHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("error!"))
}

// schemasHandler receives an HTTP GET to retrieve information about resource schemas supported by a SCIM service
// provider. An HTTP GET to the endpoint "/Schemas" returns all supported schemas in ListResponse format.
func (s Server) schemasHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	response := ListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: len(s.schemas),
		ItemsPerPage: len(s.schemas),
		StartIndex:   0,
		Resources:    s.schemas,
	}
	raw, _ := json.Marshal(response)
	w.Write(raw)
}

// schemaHandler receives an HTTP GET to retrieve individual schema definitions can be returned by appending the schema
// URI to the /Schemas endpoint. For example: `/Schemas/urn:ietf:params:scim:schemas:core:2.0:User`
func (s Server) schemaHandler(w http.ResponseWriter, r *http.Request, params url.Values) {
	var schema *Schema
	for _, s := range s.schemas {
		id, ok := params["id"]
		if ok && len(id) == 1 && s.ID == id[0] {
			schema = &s
			break
		}
	}

	response := ListResponse{
		Schemas:    []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		StartIndex: 0,
	}

	if schema != nil {
		response.TotalResults = 1
		response.ItemsPerPage = 1
		response.Resources = schema
	}

	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
