package scim

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func errorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	io.WriteString(w, `{"status": "404"}`)
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
		Resources:    resourceTypes,
	}
	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User"
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
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

// serviceProviderConfigHandler receives an HTTP GET to this endpoint will return a JSON structure that describes the
// SCIM specification features available on a service provider.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) serviceProviderConfigHandler(w http.ResponseWriter, r *http.Request) {
	raw, _ := json.Marshal(s.config)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourcePostHandler receives an HTTP POST request to the resource endpoint, such as "/Users" or "/Groups", as
// defined by the associated resource type endpoint discovery to create new resources.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.3
func (s Server) resourcePostHandler(w http.ResponseWriter, r *http.Request, resourceType resourceType) {
	data, _ := ioutil.ReadAll(r.Body)

	attributes, err := s.schemas[resourceType.Schema].validate(data, write)
	if err != nil {
		errorHandler(w, r)
		return
	}

	resource, err := resourceType.handler.Create(attributes)
	if err != nil {
		errorHandler(w, r)
		return
	}

	raw, _ := json.Marshal(resource)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourceGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to retrieve a known resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.4
func (s Server) resourceGetHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	resource, err := resourceType.handler.Get(id)
	if err != nil {
		errorHandler(w, r)
		return
	}

	raw, _ := json.Marshal(resource)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourcesGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users" or "/Groups", to retrieve
// all known resources.
func (s Server) resourcesGetHandler(w http.ResponseWriter, r *http.Request, resourceType resourceType) {
	resources, err := resourceType.handler.GetAll()
	if err != nil {
		errorHandler(w, r)
		return
	}

	response := listResponse{
		TotalResults: len(resources),
		Resources:    resources,
	}
	raw, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourcePutHandler receives an HTTP PUT  to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.5.1
func (s Server) resourcePutHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	data, _ := ioutil.ReadAll(r.Body)

	attributes, err := s.schemas[resourceType.Schema].validate(data, replace)
	if err != nil {
		errorHandler(w, r)
		return
	}

	resource, err := resourceType.handler.Replace(id, attributes)
	if err != nil {
		errorHandler(w, r)
		return
	}

	raw, _ := json.Marshal(resource)
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// resourceDeleteHandler receives an HTTP DELETE request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to delete a known resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.6
func (s Server) resourceDeleteHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	err := resourceType.handler.Delete(id)
	if err != nil {
		errorHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
