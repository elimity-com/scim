package idp_test

import (
	"github.com/elimity-com/scim/logging"
	"net/http"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func newOktaTestServer() scim.Server {
	return scim.NewServer(
		scim.ServiceProviderConfig{},
		[]scim.ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      schema.CoreUserSchema(),
				Handler:     oktaUserResourceHandler{},
			},

			{
				ID:          optional.NewString("Group"),
				Name:        "Group",
				Endpoint:    "/Groups",
				Description: optional.NewString("Group"),
				Schema:      schema.CoreGroupSchema(),
				Handler:     oktaGroupResourceHandler{},
			},
		},
		logging.NullLogger{},
	)
}

type oktaGroupResourceHandler struct{}

func (o oktaGroupResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{
		ID:         "abf4dd94-a4c0-4f67-89c9-76b03340cb9b",
		Attributes: attributes,
	}, nil
}

func (o oktaGroupResourceHandler) Delete(r *http.Request, id string) error {
	return errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

func (o oktaGroupResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	return scim.Resource{
		ID: id,
		Attributes: scim.ResourceAttributes{
			"displayName": "Test SCIMv2",
			"members": []interface{}{
				map[string]interface{}{
					"value":   "b1c794f24f4c49f4b5d503a4cb2686ea",
					"display": "SCIM 2 Group A",
				},
			},
		},
	}, nil
}

func (o oktaGroupResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	return scim.Page{}, errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

func (o oktaGroupResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	return scim.Resource{
		ID: id,
		Attributes: scim.ResourceAttributes{
			"displayName": "Test SCIMv20",
		},
	}, nil
}

func (o oktaGroupResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{}, errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

type oktaUserResourceHandler struct{}

func (t oktaUserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	delete(attributes, "password")
	return scim.Resource{
		ID:         "23a35c27-23d3-4c03-b4c5-6443c09e7173",
		ExternalID: optional.NewString("00ujl29u0le5T6Aj10h7"),
		Attributes: attributes,
	}, nil
}

func (t oktaUserResourceHandler) Delete(r *http.Request, id string) error {
	return errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

func (t oktaUserResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	return scim.Resource{
		ID: id,
		Attributes: scim.ResourceAttributes{
			"userName": "test.user@okta.local",
			"name": map[string]interface{}{
				"givenName":  "Test",
				"familyName": "User",
			},
			"active": true,
			"emails": []interface{}{
				map[string]interface{}{
					"primary": true,
					"value":   "test.user@okta.local",
					"type":    "work",
					"display": "test.user@okta.local",
				},
			},
		},
	}, nil
}

func (t oktaUserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	return scim.Page{}, nil
}

func (t oktaUserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	return scim.Resource{
		ID: id,
		Attributes: scim.ResourceAttributes{
			"userName": "test.user@okta.local",
			"name": map[string]interface{}{
				"givenName":  "Another",
				"familyName": "User",
			},
			"active": false,
			"emails": []interface{}{
				map[string]interface{}{
					"primary": true,
					"value":   "test.user@okta.local",
					"type":    "work",
					"display": "test.user@okta.local",
				},
			},
		},
	}, nil
}

func (t oktaUserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{
		ID:         id,
		Attributes: attributes,
	}, nil
}
