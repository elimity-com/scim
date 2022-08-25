package idp_test

import (
	"github.com/elimity-com/scim/logging"
	"net/http"
	"time"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
)

var azureCreatedTime = time.Date(
	2018, time.Month(3), 27,
	19, 59, 26, 0, time.UTC,
)

func newAzureADTestServer() scim.Server {
	return scim.NewServer(
		scim.ServiceProviderConfig{
			MaxResults: 20,
		},
		[]scim.ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      schema.CoreUserSchema(),
				SchemaExtensions: []scim.SchemaExtension{
					{Schema: schema.ExtensionEnterpriseUser()},
				},
				Handler: azureADUserResourceHandler{},
			},

			{
				ID:          optional.NewString("Group"),
				Name:        "Group",
				Endpoint:    "/Groups",
				Description: optional.NewString("Group"),
				Schema:      schema.CoreGroupSchema(),
				Handler:     azureADGroupResourceHandler{},
			},
		},
		logging.NullLogger{},
	)
}

type azureADGroupResourceHandler struct{}

func (a azureADGroupResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{
		ID:         "927fa2c08dcb4a7fae9e",
		ExternalID: optional.NewString(attributes["externalId"].(string)),
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &azureCreatedTime,
			LastModified: &azureCreatedTime,
		},
	}, nil
}

func (a azureADGroupResourceHandler) Delete(r *http.Request, id string) error {
	return errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

func (a azureADGroupResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	return scim.Resource{
		ID:         id,
		ExternalID: optional.NewString("60f1bb27-2e1e-402d-bcc4-ec999564a194"),
		Attributes: scim.ResourceAttributes{
			"displayName": "displayName",
		},
		Meta: scim.Meta{
			Created:      &azureCreatedTime,
			LastModified: &azureCreatedTime,
		},
	}, nil
}

func (a azureADGroupResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	return scim.Page{
		TotalResults: 1,
		Resources: []scim.Resource{
			{
				ID:         "8c601452cc934a9ebef9",
				ExternalID: optional.NewString("0db508eb-91e2-46e4-809c-30dcbda0c685"),
				Attributes: scim.ResourceAttributes{
					"displayName": "displayName",
				},
				Meta: scim.Meta{
					Created:      &azureCreatedTime,
					LastModified: &azureCreatedTime,
				},
			},
		},
	}, nil
}

func (a azureADGroupResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	return scim.Resource{
		ID:         id,
		ExternalID: optional.NewString("60f1bb27-2e1e-402d-bcc4-ec999564a194"),
		Attributes: scim.ResourceAttributes{
			"displayName": "1879db59-3bdf-4490-ad68-ab880a269474updatedDisplayName",
		},
		Meta: scim.Meta{
			Created:      &azureCreatedTime,
			LastModified: &azureCreatedTime,
		},
	}, nil
}

func (a azureADGroupResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{}, errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

type azureADUserResourceHandler struct{}

func (a azureADUserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{
		ID:         "48af03ac28ad4fb88478",
		ExternalID: optional.NewString(attributes["externalId"].(string)),
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &azureCreatedTime,
			LastModified: &azureCreatedTime,
		},
	}, nil
}

func (a azureADUserResourceHandler) Delete(r *http.Request, id string) error {
	return errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}

func (a azureADUserResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	if id == "5171a35d82074e068ce2" {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	return scim.Resource{
		ID:         id,
		ExternalID: optional.NewString("58342554-38d6-4ec8-948c-50044d0a33fd"),
		Attributes: scim.ResourceAttributes{
			"userName": "Test_User_feed3ace-693c-4e5a-82e2-694be1b39934",
			"name": map[string]interface{}{
				"formatted":  "givenName familyName",
				"familyName": "familyName",
				"givenName":  "givenName",
			},
			"active": true,
			"emails": []interface{}{
				map[string]interface{}{
					"value":   "Test_User_22370c1a-9012-42b2-bf64-86099c2a1c22@testuser.com",
					"type":    "work",
					"primary": true,
				},
			},
		},
		Meta: scim.Meta{
			Created:      &azureCreatedTime,
			LastModified: &azureCreatedTime,
		},
	}, nil
}

func (a azureADUserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	f := params.FilterValidator.GetFilter().(*filter.AttributeExpression)
	if f.CompareValue.(string) == "non-existent user" {
		return scim.Page{}, nil
	}

	return scim.Page{
		TotalResults: 1,
		Resources: []scim.Resource{
			{
				ID:         "2441309d85324e7793ae",
				ExternalID: optional.NewString("7fce0092-d52e-4f76-b727-3955bd72c939"),
				Attributes: scim.ResourceAttributes{
					"userName": "Test_User_dfeef4c5-5681-4387-b016-bdf221e82081",
					"name": map[string]interface{}{
						"familyName": "familyName",
						"givenName":  "givenName",
					},
					"active": true,
					"emails": []interface{}{
						map[string]interface{}{
							"value":   "Test_User_91b67701-697b-46de-b864-bd0bbe4f99c1@testuser.com",
							"type":    "work",
							"primary": true,
						},
					},
				},
				Meta: scim.Meta{
					Created:      &azureCreatedTime,
					LastModified: &azureCreatedTime,
				},
			},
		},
	}, nil
}

func (a azureADUserResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	return scim.Resource{
		ID:         id,
		ExternalID: optional.NewString("6c75de36-30fa-4d2d-a196-6bdcdb6b6539"),
		Attributes: scim.ResourceAttributes{
			"userName": "5b50642d-79fc-4410-9e90-4c077cdd1a59@testuser.com",
			"name": map[string]interface{}{
				"formatted":  "givenName updatedFamilyName",
				"familyName": "updatedFamilyName",
				"givenName":  "givenName",
			},
			"active": false,
			"emails": []interface{}{
				map[string]interface{}{
					"value":   "updatedEmail@microsoft.com",
					"type":    "work",
					"primary": true,
				},
			},
		},
		Meta: scim.Meta{
			Created:      &azureCreatedTime,
			LastModified: &azureCreatedTime,
		},
	}, nil
}

func (a azureADUserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	return scim.Resource{}, errors.ScimError{
		Status: http.StatusNotImplemented,
	}
}
