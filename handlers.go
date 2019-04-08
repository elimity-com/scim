package scim

import (
	"encoding/json"
	"net/http"
)

func errorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("error!"))
}

// schemasHandler receives an HTTP GET to retrieve information about resource schemas supported by a SCIM service
// provider. An HTTP GET to the endpoint "/Schemas" returns all supported schemas in ListResponse format.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) schemasHandler(w http.ResponseWriter, r *http.Request) {
	response := listResponse{
		TotalResults: len(s.schemas),
		ItemsPerPage: len(s.schemas),
		StartIndex:   0,
		Resources:    s.schemas,
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
	var schema *Schema
	for _, s := range s.schemas {
		if s.ID == id {
			schema = &s
			break
		}
	}

	response := listResponse{
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

// resourceTypesHandler receives an HTTP GET to this endpoint, "/ResourceTypes", which is used to discover the types of
// resources available on a SCIM service provider (e.g., Users and Groups).  Each resource type defines the endpoints,
// the core schema URI that defines the resource, and any supported schema extensions.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) resourceTypesHandler(w http.ResponseWriter, r *http.Request) {
	response := listResponse{
		TotalResults: len(s.resourceTypes),
		ItemsPerPage: len(s.resourceTypes),
		StartIndex:   0,
		Resources:    s.resourceTypes,
	}
	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User"
func (s Server) resourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	var resourceType *ResourceType
	for _, t := range s.resourceTypes {
		if t.Name == name {
			resourceType = &t
			break
		}
	}

	response := listResponse{
		StartIndex: 0,
	}

	if resourceType != nil {
		response.TotalResults = 1
		response.ItemsPerPage = 1
		response.Resources = resourceType
	}

	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
