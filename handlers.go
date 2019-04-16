package scim

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func errorHandler(w http.ResponseWriter, r *http.Request, scimErr scimError) {
	raw, err := json.Marshal(scimErr)
	if err != nil {
		log.Fatalf("failed marshaling scim error: %v", err)
	}
	w.WriteHeader(scimErr.status)
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
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
	raw, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("failed marshaling list response: %v", err)
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// schemaHandler receives an HTTP GET to retrieve individual schema definitions which can be returned by appending the
// schema URI to the /Schemas endpoint. For example: "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) schemaHandler(w http.ResponseWriter, r *http.Request, id string) {
	schema, ok := s.schemas[id]
	if !ok {
		errorHandler(w, r, resourceNotFound(id))
		return
	}

	raw, err := json.Marshal(schema)
	if err != nil {
		log.Fatalf("failed marshaling schema: %v", err)
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
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
	raw, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("failed marshaling list response: %v", err)
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User"
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) resourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	resourceType, ok := s.resourceTypes[name]
	if !ok {
		errorHandler(w, r, resourceNotFound(name))
		return
	}

	raw, err := json.Marshal(resourceType)
	if err != nil {
		log.Fatalf("failed marshaling resource type: %v", err)
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// serviceProviderConfigHandler receives an HTTP GET to this endpoint will return a JSON structure that describes the
// SCIM specification features available on a service provider.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-4
func (s Server) serviceProviderConfigHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(s.config)
	if err != nil {
		log.Fatalf("failed marshaling service provider config: %v", err)
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcePostHandler receives an HTTP POST request to the resource endpoint, such as "/Users" or "/Groups", as
// defined by the associated resource type endpoint discovery to create new resources.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.3
func (s Server) resourcePostHandler(w http.ResponseWriter, r *http.Request, resourceType resourceType) {
	w.WriteHeader(http.StatusOK)
	_, err := io.WriteString(w, fmt.Sprintf(`{"desc": "create %s (%s)"}`, resourceType.Name, s.schemas[resourceType.Schema].ID))
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to retrieve a known resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.4
func (s Server) resourceGetHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	_, err := io.WriteString(w, fmt.Sprintf(`{"desc": "get %s (%s) with id: %s"}`, resourceType.Name, s.schemas[resourceType.Schema].ID, id))
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcesGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users" or "/Groups", to retrieve
// all known resources.
func (s Server) resourcesGetHandler(w http.ResponseWriter, r *http.Request, resourceType resourceType) {
	_, err := io.WriteString(w, fmt.Sprintf(`{"desc": "get all %ss (%s)"}`, resourceType.Name, s.schemas[resourceType.Schema].ID))
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcePutHandler receives an HTTP PUT to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.5.1
func (s Server) resourcePutHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	_, err := io.WriteString(w, fmt.Sprintf(`{"desc": "replace %s (%s) with id: %s"}`, resourceType.Name, s.schemas[resourceType.Schema].ID, id))
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceDeleteHandler receives an HTTP DELETE to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to delete a resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.6
func (s Server) resourceDeleteHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	_, err := io.WriteString(w, fmt.Sprintf(`{"desc": "delete %s (%s) with id: %s"}`, resourceType.Name, s.schemas[resourceType.Schema].ID, id))
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}
