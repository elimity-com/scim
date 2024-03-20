package scim

import (
	"encoding/json"
	"net/http"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

func (s Server) errorHandler(w http.ResponseWriter, scimErr *errors.ScimError) {
	raw, err := json.Marshal(scimErr)
	if err != nil {
		s.log.Error(
			"failed marshaling scim error",
			"scimError", scimErr,
			"error", err,
		)
		return
	}

	w.WriteHeader(scimErr.Status)
	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourceDeleteHandler receives an HTTP DELETE request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to delete a known resource.
func (s Server) resourceDeleteHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	deleteErr := resourceType.Handler.Delete(r, id)
	if deleteErr != nil {
		scimErr := errors.CheckScimError(deleteErr, http.MethodDelete)
		s.errorHandler(w, &scimErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// resourceGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}",
// where "{id}" is a resource identifier to retrieve a known resource.
func (s Server) resourceGetHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	resource, getErr := resourceType.Handler.Get(r, id)
	if getErr != nil {
		scimErr := errors.CheckScimError(getErr, http.MethodGet)
		s.errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling resource",
			"resource", resource,
			"error", err,
		)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourcePatchHandler receives an HTTP PATCH to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
func (s Server) resourcePatchHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	patch, scimErr := resourceType.validatePatch(r)
	if scimErr != nil {
		s.errorHandler(w, scimErr)
		return
	}

	resource, patchErr := resourceType.Handler.Patch(r, id, patch)
	if patchErr != nil {
		scimErr := errors.CheckScimError(patchErr, http.MethodPatch)
		s.errorHandler(w, &scimErr)
		return
	}

	if len(resource.Attributes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling resource",
			"resource", resource,
			"error", err,
		)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	w.WriteHeader(http.StatusOK)

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourcePostHandler receives an HTTP POST request to the resource endpoint, such as "/Users" or "/Groups", as
// defined by the associated resource type endpoint discovery to create new resources.
func (s Server) resourcePostHandler(w http.ResponseWriter, r *http.Request, resourceType ResourceType) {
	data, _ := readBody(r)

	attributes, scimErr := resourceType.validate(data)
	if scimErr != nil {
		s.errorHandler(w, scimErr)
		return
	}

	resource, postErr := resourceType.Handler.Create(r, attributes)
	if postErr != nil {
		scimErr := errors.CheckScimError(postErr, http.MethodPost)
		s.errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling resource",
			"resource", resource,
			"error", err,
		)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourcePutHandler receives an HTTP PUT to the resource endpoint, e.g., "/Users/{id}" or "/Groups/{id}", where
// "{id}" is a resource identifier to replace a resource's attributes.
func (s Server) resourcePutHandler(w http.ResponseWriter, r *http.Request, id string, resourceType ResourceType) {
	data, _ := readBody(r)

	attributes, scimErr := resourceType.validate(data)
	if scimErr != nil {
		s.errorHandler(w, scimErr)
		return
	}

	resource, putError := resourceType.Handler.Replace(r, id, attributes)
	if putError != nil {
		scimErr := errors.CheckScimError(putError, http.MethodPut)
		s.errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling resource",
			"resource", resource,
			"error", err,
		)
		return
	}

	if resource.Meta.Version != "" {
		w.Header().Set("Etag", resource.Meta.Version)
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User".
func (s Server) resourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	var resourceType ResourceType
	for _, r := range s.resourceTypes {
		if r.Name == name {
			resourceType = r
			break
		}
	}

	if resourceType.Name != name {
		scimErr := errors.ScimErrorResourceNotFound(name)
		s.errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resourceType.getRaw())
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling resource type",
			"resourceType", resourceType,
			"error", err,
		)
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourceTypesHandler receives an HTTP GET to this endpoint, "/ResourceTypes", which is used to discover the types of
// resources available on a SCIM service provider (e.g., Users and Groups).  Each resource type defines the endpoints,
// the core schema URI that defines the resource, and any supported schema extensions.
func (s Server) resourceTypesHandler(w http.ResponseWriter, r *http.Request) {
	params, paramsErr := s.parseRequestParams(r, schema.ResourceTypeSchema())
	if paramsErr != nil {
		s.errorHandler(w, paramsErr)
		return
	}

	start, end := clamp(params.StartIndex-1, params.Count, len(s.resourceTypes))
	var resources []interface{}
	for _, v := range s.resourceTypes[start:end] {
		resources = append(resources, v.getRaw())
	}

	lr := listResponse{
		TotalResults: len(s.resourceTypes),
		ItemsPerPage: params.Count,
		StartIndex:   params.StartIndex,
		Resources:    resources,
	}
	raw, err := json.Marshal(lr)
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling list response",
			"listResponse", lr,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourcesGetHandler receives an HTTP GET request to the resource endpoint, e.g., "/Users" or "/Groups", to retrieve
// all known resources.
func (s Server) resourcesGetHandler(w http.ResponseWriter, r *http.Request, resourceType ResourceType) {
	params, paramsErr := s.parseRequestParams(r, resourceType.Schema, resourceType.getSchemaExtensions()...)
	if paramsErr != nil {
		s.errorHandler(w, paramsErr)
		return
	}

	page, getError := resourceType.Handler.GetAll(r, params)
	if getError != nil {
		scimErr := errors.CheckScimError(getError, http.MethodGet)
		s.errorHandler(w, &scimErr)
		return
	}

	lr := listResponse{
		TotalResults: page.TotalResults,
		Resources:    page.resources(resourceType),
		StartIndex:   params.StartIndex,
		ItemsPerPage: params.Count,
	}
	raw, err := json.Marshal(lr)
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling list response",
			"listResponse", lr,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// schemaHandler receives an HTTP GET to retrieve individual schema definitions which can be returned by appending the
// schema URI to the /Schemas endpoint. For example: "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User".
func (s Server) schemaHandler(w http.ResponseWriter, r *http.Request, id string) {
	getSchema := s.getSchema(id)
	if getSchema.ID != id {
		scimErr := errors.ScimErrorResourceNotFound(id)
		s.errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(getSchema)
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling schema",
			"schema", getSchema,
			"error", err,
		)
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// schemasHandler receives an HTTP GET to retrieve information about resource schemas supported by a SCIM service
// provider. An HTTP GET to the endpoint "/Schemas" returns all supported schemas in ListResponse format.
func (s Server) schemasHandler(w http.ResponseWriter, r *http.Request) {
	params, paramsErr := s.parseRequestParams(r, schema.Definition())
	if paramsErr != nil {
		s.errorHandler(w, paramsErr)
		return
	}

	var (
		start, end = clamp(params.StartIndex-1, params.Count, len(s.getSchemas()))
		resources  []interface{}
	)
	if validator := params.FilterValidator; validator != nil {
		if err := validator.Validate(); err != nil {
			s.errorHandler(w, &errors.ScimErrorInvalidFilter)
			return
		}
	}
	for _, v := range s.getSchemas()[start:end] {
		resource := v.ToMap()
		if validator := params.FilterValidator; validator != nil {
			if err := validator.PassesFilter(resource); err != nil {
				continue
			}
		}
		resources = append(resources, resource)
	}

	lr := listResponse{
		TotalResults: len(s.getSchemas()),
		ItemsPerPage: params.Count,
		StartIndex:   params.StartIndex,
		Resources:    resources,
	}
	raw, err := json.Marshal(lr)
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling list response",
			"listResponse", lr,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// serviceProviderConfigHandler receives an HTTP GET to this endpoint will return a JSON structure that describes the
// SCIM specification features available on a service provider.
func (s Server) serviceProviderConfigHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(s.config.getRaw())
	if err != nil {
		s.errorHandler(w, &errors.ScimErrorInternal)
		s.log.Error(
			"failed marshaling service provider config",
			"serviceProviderConfig", s.config,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		s.log.Error(
			"failed writing response",
			"error", err,
		)
	}
}
