package scim

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/elimity-com/scim/errors"
)

func errorHandler(w http.ResponseWriter, _ *http.Request, scimErr scimError) {
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
func (s Server) schemasHandler(w http.ResponseWriter, r *http.Request) {
	params, paramsErr := s.parseRequestParams(r)
	if paramsErr != nil {
		errorHandler(w, r, *paramsErr)
		return
	}

	schemas := s.getSchemas()
	start, end := clamp(params.StartIndex-1, params.Count, len(schemas))
	var resources []interface{}
	for _, v := range schemas[start:end] {
		resources = append(resources, v)
	}

	raw, err := json.Marshal(listResponse{
		TotalResults: len(schemas),
		ItemsPerPage: params.Count,
		StartIndex:   params.StartIndex,
		Resources:    resources,
	})
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling list response: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// schemaHandler receives an HTTP GET to retrieve individual schema definitions which can be returned by appending the
// schema URI to the /Schemas endpoint. For example: "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User"
func (s Server) schemaHandler(w http.ResponseWriter, r *http.Request, id string) {
	schema := s.getSchema(id)
	if schema.ID != id {
		errorHandler(w, r, scimErrorResourceNotFound(id))
		return
	}

	raw, err := json.Marshal(schema)
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling schema: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceTypesHandler receives an HTTP GET to this endpoint, "/ResourceTypes", which is used to discover the types of
// resources available on a SCIM service provider (e.g., Users and Groups).  Each resource type defines the endpoints,
// the core schema URI that defines the resource, and any supported schema extensions.
func (s Server) resourceTypesHandler(w http.ResponseWriter, r *http.Request) {
	params, paramsErr := s.parseRequestParams(r)
	if paramsErr != nil {
		errorHandler(w, r, *paramsErr)
		return
	}

	start, end := clamp(params.StartIndex-1, params.Count, len(s.ResourceTypes))
	var resources []interface{}
	for _, v := range s.ResourceTypes[start:end] {
		resources = append(resources, v.getRaw())
	}

	raw, err := json.Marshal(listResponse{
		TotalResults: len(s.ResourceTypes),
		ItemsPerPage: params.Count,
		StartIndex:   params.StartIndex,
		Resources:    resources,
	})
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling list response: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User"
func (s Server) resourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	var resourceType ResourceType
	for _, r := range s.ResourceTypes {
		if r.Name == name {
			resourceType = r
			break
		}
	}
	if resourceType.Name != name {
		errorHandler(w, r, scimErrorResourceNotFound(name))
		return
	}

	raw, err := json.Marshal(resourceType.getRaw())
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling resource type: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// serviceProviderConfigHandler receives an HTTP GET to this endpoint will return a JSON structure that describes the
// SCIM specification features available on a service provider.
func (s Server) serviceProviderConfigHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(s.Config.getRaw())
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling service provider config: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcePatchHandler receives an HTTP PATCH to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
func (s Server) resourcePatchHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	patch, scimErr := resourceType.validatePatch(r)
	if scimErr != errors.ValidationErrorNil {
		errorHandler(w, r, scimValidationError(scimErr))
		return
	}

	resource, patchErr := resourceType.Handler.Patch(r, id, patch)
	if patchErr != errors.PatchErrorNil {
		errorHandler(w, r, scimPatchError(patchErr, id))
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcePostHandler receives an HTTP POST request to the resource endpoint, such as "/Users" or "/Groups", as
// defined by the associated resource type endpoint discovery to create new resources.
func (s Server) resourcePostHandler(w http.ResponseWriter, r *http.Request, resourceType ResourceType) {
	data, _ := ioutil.ReadAll(r.Body)

	attributes, scimErr := resourceType.validate(data)
	if scimErr != errors.ValidationErrorNil {
		errorHandler(w, r, scimValidationError(scimErr))
		return
	}

	resource, postErr := resourceType.Handler.Create(r, attributes)
	if postErr != errors.PostErrorNil {
		errorHandler(w, r, scimPostError(postErr))
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourceGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to retrieve a known resource.
func (s Server) resourceGetHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	resource, getErr := resourceType.Handler.Get(r, id)
	if getErr != errors.GetErrorNil {
		errorHandler(w, r, scimGetError(getErr, id))
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
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
func (s Server) resourcesGetHandler(w http.ResponseWriter, r *http.Request, resourceType ResourceType) {
	params, paramsErr := s.parseRequestParams(r)
	if paramsErr != nil {
		errorHandler(w, r, *paramsErr)
		return
	}

	page, getError := resourceType.Handler.GetAll(r, params)
	if getError != errors.GetErrorNil {
		errorHandler(w, r, scimGetAllError(getError))
		return
	}

	resources := []interface{}{}
	for _, v := range page.Resources {
		resources = append(resources, v.response(resourceType))
	}

	raw, err := json.Marshal(listResponse{
		TotalResults: page.TotalResults,
		Resources:    resources,
		StartIndex:   params.StartIndex,
		ItemsPerPage: params.Count,
	})
	if err != nil {
		errorHandler(w, r, scimErrorInternalServer)
		log.Fatalf("failed marshalling list response: %v", err)
		return
	}
	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// resourcePutHandler receives an HTTP PUT to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
func (s Server) resourcePutHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	data, _ := ioutil.ReadAll(r.Body)

	attributes, scimErr := resourceType.validate(data)
	if scimErr != errors.ValidationErrorNil {
		errorHandler(w, r, scimValidationError(scimErr))
		return
	}

	resource, putError := resourceType.Handler.Replace(r, id, attributes)
	if putError != errors.PutErrorNil {
		errorHandler(w, r, scimPutError(putError, id))
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
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

// resourceDeleteHandler receives an HTTP DELETE request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to delete a known resource.
func (s Server) resourceDeleteHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	deleteErr := resourceType.Handler.Delete(r, id)
	if deleteErr != errors.DeleteErrorNil {
		errorHandler(w, r, scimDeleteError(deleteErr, id))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
