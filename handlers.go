package scim

import (
	"encoding/json"
	"net/http"
)

func errorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"status": "404"}`))
}

// schemasHandler receives an HTTP GET to retrieve information about resource schemas supported by a SCIM service
// provider. An HTTP GET to the endpoint "/Schemas" returns all supported schemas in ListResponse format.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) schemasHandler(w http.ResponseWriter, r *http.Request) {
	var schemas []schema
	for _, v := range s.schemas {
		schemas = append(schemas, v)
	}

	response := listResponse{
		TotalResults: len(schemas),
		ItemsPerPage: len(schemas),
		StartIndex:   1,
		Resources:    schemas,
	}
	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// schemaHandler receives an HTTP GET to retrieve individual schema definitions which can be returned by appending the
// schema URI to the /Schemas endpoint. For example: "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) schemaHandler(w http.ResponseWriter, r *http.Request, id string) {
	schema, ok := s.schemas[id]
	if !ok {
		errorHandler(w, r)
		return
	}

	raw, _ := json.Marshal(schema)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourceTypesHandler receives an HTTP GET to this endpoint, "/ResourceTypes", which is used to discover the types of
// resources available on a SCIM service provider (e.g., Users and Groups).  Each resource type defines the endpoints,
// the core schema URI that defines the resource, and any supported schema extensions.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) resourceTypesHandler(w http.ResponseWriter, r *http.Request) {
	var resourceTypes []resourceType
	for _, v := range s.resourceTypes {
		resourceTypes = append(resourceTypes, v)
	}

	response := listResponse{
		TotalResults: len(resourceTypes),
		ItemsPerPage: len(resourceTypes),
		StartIndex:   0,
		Resources:    resourceTypes,
	}
	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User"
func (s Server) resourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	resourceType, ok := s.resourceTypes[name]
	if !ok {
		errorHandler(w, r)
		return
	}

	raw, _ := json.Marshal(resourceType)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
