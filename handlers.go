package scim

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func errorHandler(w http.ResponseWriter, r *http.Request, scimErr scimError) {
	raw, err := json.Marshal(scimErr)
	if err != nil {
		log.Fatalf("failed marshaling scim error: %v", err)
	}
	w.WriteHeader(scimErr.Status)
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
		errorHandler(w, r, scimErrorResourceNotFound(id))
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
		errorHandler(w, r, scimErrorResourceNotFound(name))
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
	data, _ := ioutil.ReadAll(r.Body)

	attributes, scimErr := resourceType.validate(s.schemas, data, write)
	if scimErr != scimErrorNil {
		errorHandler(w, r, scimErr)
		return
	}

	resource, postErr := resourceType.handler.Create(attributes)
	if postErr != PostErrorNil {
		errorHandler(w, r, postErr.err)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType, r.Host+r.RequestURI+"/"+resource.ID))
	if err != nil {
		log.Fatalf("failed marshaling resource: %v", err)
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to retrieve a known resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.4
func (s Server) resourceGetHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	resource, getErr := resourceType.handler.Get(id)
	if getErr != GetErrorNil {
		errorHandler(w, r, getErr.err)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType, r.Host+r.RequestURI))
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcesGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users" or "/Groups", to retrieve
// all known resources.
func (s Server) resourcesGetHandler(w http.ResponseWriter, r *http.Request, resourceType resourceType) {
	resources, getErr := resourceType.handler.GetAll()
	if getErr != GetErrorNil {
		errorHandler(w, r, getErr.err)
		return
	}

	respResources := make([]CoreAttributes, 0)
	for _, resource := range resources {
		respResources = append(respResources, resource.response(resourceType, r.Host+r.RequestURI+"/"+resource.ID))
	}

	response := listResponse{
		TotalResults: len(resources),
		Resources:    respResources,
	}
	raw, err := json.Marshal(response)
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed list response %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcePutHandler receives an HTTP PUT to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.5.1
func (s Server) resourcePutHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	data, _ := ioutil.ReadAll(r.Body)

	attributes, scimErr := resourceType.validate(s.schemas, data, replace)
	if scimErr != scimErrorNil {
		errorHandler(w, r, scimErr)
		return
	}

	resource, putError := resourceType.handler.Replace(id, attributes)
	if putError != PutErrorNil {
		errorHandler(w, r, putError.err)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType, r.Host+r.RequestURI))
	if err != nil {
		log.Fatalf("failed marshaling resource: %v", err)
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceDeleteHandler receives an HTTP DELETE request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to delete a known resource.
//
// RFC: https://tools.ietf.org/html/rfc7644#section-3.6
func (s Server) resourceDeleteHandler(w http.ResponseWriter, r *http.Request, id string, resourceType resourceType) {
	deleteErr := resourceType.handler.Delete(id)
	if deleteErr != DeleteErrorNil {
		errorHandler(w, r, deleteErr.err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
