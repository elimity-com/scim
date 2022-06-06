package scim

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

func ErrorHandler(w http.ResponseWriter, _ *http.Request, scimErr *errors.ScimError) {
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

// ResourceDeleteHandler receives an HTTP DELETE request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to delete a known resource.
func (s Server) ResourceDeleteHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	deleteErr := resourceType.Handler.Delete(r, id)
	if deleteErr != nil {
		scimErr := errors.CheckScimError(deleteErr, http.MethodDelete)
		ErrorHandler(w, r, &scimErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ResourceGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to retrieve a known resource.
func (s Server) ResourceGetHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	resource, getErr := resourceType.Handler.Get(r, id)
	if getErr != nil {
		scimErr := errors.CheckScimError(getErr, http.MethodGet)
		ErrorHandler(w, r, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ResourcePatchHandler receives an HTTP PATCH to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
func (s Server) ResourcePatchHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	patch, scimErr := resourceType.validatePatch(r)
	if scimErr != nil {
		ErrorHandler(w, r, scimErr)
		return
	}

	resource, patchErr := resourceType.Handler.Patch(r, id, patch)
	if patchErr != nil {
		scimErr := errors.CheckScimError(patchErr, http.MethodPatch)
		ErrorHandler(w, r, &scimErr)
		return
	}

	if len(resource.Attributes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ResourcePostHandler receives an HTTP POST request to the resource endpoint, such as "/Users" or "/Groups", as
// defined by the associated resource type endpoint discovery to create new resources.
func (s Server) ResourcePostHandler(w http.ResponseWriter, r *http.Request, resourceType ResourceType) {
	data, _ := readBody(r)

	attributes, scimErr := resourceType.validate(data)
	if scimErr != nil {
		ErrorHandler(w, r, scimErr)
		return
	}

	resource, postErr := resourceType.Handler.Create(r, attributes)
	if postErr != nil {
		scimErr := errors.CheckScimError(postErr, http.MethodPost)
		ErrorHandler(w, r, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ResourcePutHandler receives an HTTP PUT to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
func (s Server) ResourcePutHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	data, _ := readBody(r)

	attributes, scimErr := resourceType.validate(data)
	if scimErr != nil {
		ErrorHandler(w, r, scimErr)
		return
	}

	resource, putError := resourceType.Handler.Replace(r, id, attributes)
	if putError != nil {
		scimErr := errors.CheckScimError(putError, http.MethodPut)
		ErrorHandler(w, r, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling resource: %v", err)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ResourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User".
func (s Server) ResourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	var resourceType ResourceType
	for _, r := range s.ResourceTypes {
		if r.Name == name {
			resourceType = r
			break
		}
	}

	if resourceType.Name != name {
		scimErr := errors.ScimErrorResourceNotFound(name)
		ErrorHandler(w, r, &scimErr)
		return
	}

	raw, err := json.Marshal(resourceType.getRaw())
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling resource type: %v", err)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ResourceTypesHandler receives an HTTP GET to this endpoint, "/ResourceTypes", which is used to discover the types of
// resources available on a SCIM service provider (e.g., Users and Groups).  Each resource type defines the endpoints,
// the core schema URI that defines the resource, and any supported schema extensions.
func (s Server) ResourceTypesHandler(w http.ResponseWriter, r *http.Request) {
	params, paramsErr := s.parseRequestParams(r, schema.ResourceTypeSchema())
	if paramsErr != nil {
		ErrorHandler(w, r, paramsErr)
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
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling list response: %v", err)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ResourcesGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users" or "/Groups", to retrieve
// all known resources.
func (s Server) ResourcesGetHandler(w http.ResponseWriter, r *http.Request, resourceType ResourceType) {
	params, paramsErr := s.parseRequestParams(r, resourceType.Schema, resourceType.getSchemaExtensions()...)
	if paramsErr != nil {
		ErrorHandler(w, r, paramsErr)
		return
	}

	page, getError := resourceType.Handler.GetAll(r, params)
	if getError != nil {
		scimErr := errors.CheckScimError(getError, http.MethodGet)
		ErrorHandler(w, r, &scimErr)
		return
	}

	raw, err := json.Marshal(listResponse{
		TotalResults: page.TotalResults,
		Resources:    page.resources(resourceType),
		StartIndex:   params.StartIndex,
		ItemsPerPage: params.Count,
	})
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshalling list response: %v", err)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// SchemaHandler receives an HTTP GET to retrieve individual schema definitions which can be returned by appending the
// schema URI to the /Schemas endpoint. For example: "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User".
func (s Server) SchemaHandler(w http.ResponseWriter, r *http.Request, id string) {
	getSchema := s.getSchema(id)
	if getSchema.ID != id {
		scimErr := errors.ScimErrorResourceNotFound(id)
		ErrorHandler(w, r, &scimErr)
		return
	}

	raw, err := json.Marshal(getSchema)
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling schema: %v", err)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// SchemasHandler receives an HTTP GET to retrieve information about resource schemas supported by a SCIM service
// provider. An HTTP GET to the endpoint "/Schemas" returns all supported schemas in ListResponse format.
func (s Server) SchemasHandler(w http.ResponseWriter, r *http.Request) {
	params, paramsErr := s.parseRequestParams(r, schema.Definition())
	if paramsErr != nil {
		ErrorHandler(w, r, paramsErr)
		return
	}

	var (
		start, end = clamp(params.StartIndex-1, params.Count, len(s.getSchemas()))
		resources  []interface{}
	)
	for _, v := range s.getSchemas()[start:end] {
		resource := v.ToMap()
		if params.FilterValidator != nil {
			if err := params.FilterValidator.PassesFilter(resource); err != nil {
				continue
			}
		}
		resources = append(resources, resource)
	}

	raw, err := json.Marshal(listResponse{
		TotalResults: len(s.getSchemas()),
		ItemsPerPage: params.Count,
		StartIndex:   params.StartIndex,
		Resources:    resources,
	})
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling list response: %v", err)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}

// ServiceProviderConfigHandler receives an HTTP GET to this endpoint will return a JSON structure that describes the
// SCIM specification features available on a service provider.
func (s Server) ServiceProviderConfigHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(s.Config.getRaw())
	if err != nil {
		ErrorHandler(w, r, &errors.ScimErrorInternal)
		log.Fatalf("failed marshaling service provider config: %v", err)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Printf("failed writing response: %v", err)
	}
}
