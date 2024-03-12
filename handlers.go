package scim

import (
	"encoding/json"
	"net/http"

	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/schema"
)

func errorHandler(w http.ResponseWriter, scimErr *errors.ScimError) {
	raw, err := json.Marshal(scimErr)
	if err != nil {
		log.Error(
			"failed marshaling scim error",
			"scimError", scimErr,
			"error", err,
		)
		return
	}

	w.WriteHeader(scimErr.Status)
	_, err = w.Write(raw)
	if err != nil {
		log.Error(
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
		errorHandler(w, &scimErr)
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
		errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
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
		log.Error(
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
		errorHandler(w, scimErr)
		return
	}

	resource, patchErr := resourceType.Handler.Patch(r, id, patch)
	if patchErr != nil {
		scimErr := errors.CheckScimError(patchErr, http.MethodPatch)
		errorHandler(w, &scimErr)
		return
	}

	if len(resource.Attributes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
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
		log.Error(
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
		errorHandler(w, scimErr)
		return
	}

	resource, postErr := resourceType.Handler.Create(r, attributes)
	if postErr != nil {
		scimErr := errors.CheckScimError(postErr, http.MethodPost)
		errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
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
		log.Error(
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
		errorHandler(w, scimErr)
		return
	}

	resource, putError := resourceType.Handler.Replace(r, id, attributes)
	if putError != nil {
		scimErr := errors.CheckScimError(putError, http.MethodPut)
		errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resource.response(resourceType))
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
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
		log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// resourceTypeHandler receives an HTTP GET to retrieve individual resource types which can be returned by appending the
// resource types name to the /ResourceTypes endpoint. For example: "/ResourceTypes/User".
func (s Server) resourceTypeHandler(w http.ResponseWriter, r *http.Request, name string) {
	var resourceType ResourceType
	for _, r := range s.ResourceTypes {
		if r.Name == name {
			resourceType = r
			break
		}
	}

	if resourceType.Name != name {
		scimErr := errors.ScimErrorResourceNotFound(name)
		errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(resourceType.getRaw())
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
			"failed marshaling resource type",
			"resourceType", resourceType,
			"error", err,
		)
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Error(
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
		errorHandler(w, paramsErr)
		return
	}

	start, end := clamp(params.StartIndex-1, params.Count, len(s.ResourceTypes))
	var resources []interface{}
	for _, v := range s.ResourceTypes[start:end] {
		resources = append(resources, v.getRaw())
	}

	lr := listResponse{
		TotalResults: len(s.ResourceTypes),
		ItemsPerPage: params.Count,
		StartIndex:   params.StartIndex,
		Resources:    resources,
	}
	raw, err := json.Marshal(lr)
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
			"failed marshaling list response",
			"listResponse", lr,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Error(
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
		errorHandler(w, paramsErr)
		return
	}

	page, getError := resourceType.Handler.GetAll(r, params)
	if getError != nil {
		scimErr := errors.CheckScimError(getError, http.MethodGet)
		errorHandler(w, &scimErr)
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
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
			"failed marshaling list response",
			"listResponse", lr,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Error(
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
		errorHandler(w, &scimErr)
		return
	}

	raw, err := json.Marshal(getSchema)
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
			"failed marshaling schema",
			"schema", getSchema,
			"error", err,
		)
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Error(
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
		errorHandler(w, paramsErr)
		return
	}

	var (
		start, end = clamp(params.StartIndex-1, params.Count, len(s.getSchemas()))
		resources  []interface{}
	)
	if validator := params.FilterValidator; validator != nil {
		if err := validator.Validate(); err != nil {
			errorHandler(w, &errors.ScimErrorInvalidFilter)
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
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
			"failed marshaling list response",
			"listResponse", lr,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Error(
			"failed writing response",
			"error", err,
		)
	}
}

// serviceProviderConfigHandler receives an HTTP GET to this endpoint will return a JSON structure that describes the
// SCIM specification features available on a service provider.
func (s Server) serviceProviderConfigHandler(w http.ResponseWriter, r *http.Request) {
	raw, err := json.Marshal(s.Config.getRaw())
	if err != nil {
		errorHandler(w, &errors.ScimErrorInternal)
		log.Error(
			"failed marshaling service provider config",
			"serviceProviderConfig", s.Config,
			"error", err,
		)
		return
	}

	_, err = w.Write(raw)
	if err != nil {
		log.Error(
			"failed writing response",
			"error", err,
		)
	}
}
