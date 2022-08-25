package scim_test

import (
	"encoding/json"
	"github.com/elimity-com/scim/logging"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

func Test_Group_Filter(t *testing.T) {
	s := newTestServerForFilter()

	tests := []struct {
		name                string
		filter              string
		expectedDisplayName string
	}{
		{name: "Happy path", filter: "displayName eq \"testGroup\"", expectedDisplayName: "testGroup"},
		{name: "Happy path with plus sign", filter: "displayName eq \"testGroup+test\"", expectedDisplayName: "testGroup+test"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/Groups?filter="+url.QueryEscape(tt.filter), nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)

			bytes, err := ioutil.ReadAll(w.Result().Body)
			if err != nil {
				t.Fatal(err)
			}

			if w.Result().StatusCode != http.StatusOK {
				t.Fatal(w.Result().StatusCode, string(bytes))
			}

			var result map[string]interface{}
			if err := json.Unmarshal(bytes, &result); err != nil {
				t.Fatal(err)
			}

			resources, ok := result["Resources"].([]interface{})
			if !ok {
				t.Fatal("Resources is not the right type or missing")
			}

			if len(resources) != 1 {
				t.Fatal("one Resource expected")
			}

			firstResource, ok := resources[0].(map[string]interface{})
			if !ok {
				t.Fatal("first Resource is not the right type or missing")
			}

			displayName, ok := firstResource["displayName"].(string)
			if !ok {
				t.Fatal("displayName is not the right type or missing")
			}

			if displayName != tt.expectedDisplayName {
				t.Fatal("displayName not eq " + displayName)
			}
		})
	}
}

func Test_User_Filter(t *testing.T) {
	s := newTestServerForFilter()

	tests := []struct {
		name             string
		filter           string
		expectedUserName string
	}{
		{name: "Happy path", filter: "userName eq \"testUser\"", expectedUserName: "testUser"},
		{name: "Happy path with plus sign", filter: "userName eq \"testUser+test\"", expectedUserName: "testUser+test"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/Users?filter="+url.QueryEscape(tt.filter), nil)
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)

			bytes, err := ioutil.ReadAll(w.Result().Body)
			if err != nil {
				t.Fatal(err)
			}

			if w.Result().StatusCode != http.StatusOK {
				t.Fatal(w.Result().StatusCode, string(bytes))
			}

			var result map[string]interface{}
			if err := json.Unmarshal(bytes, &result); err != nil {
				t.Fatal(err)
			}

			resources, ok := result["Resources"].([]interface{})
			if !ok {
				t.Fatal("Resources is not the right type or missing")
			}

			if len(resources) != 1 {
				t.Fatal("one Resource expected")
			}

			firstResource, ok := resources[0].(map[string]interface{})
			if !ok {
				t.Fatal("first Resource is not the right type or missing")
			}

			userName, ok := firstResource["userName"].(string)
			if !ok {
				t.Fatal("userName is not the right type or missing")
			}

			if userName != tt.expectedUserName {
				t.Fatal("userName not eq " + userName)
			}
		})
	}
}

func newTestServerForFilter() scim.Server {
	return scim.NewServer(
		scim.ServiceProviderConfig{},
		[]scim.ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      schema.CoreUserSchema(),
				Handler: &testResourceHandler{
					data: map[string]testData{
						"0001": {attributes: map[string]interface{}{"userName": "testUser"}},
						"0002": {attributes: map[string]interface{}{"userName": "testUser+test"}},
					},
					schema: schema.CoreUserSchema(),
				},
			},
			{
				ID:          optional.NewString("Group"),
				Name:        "Group",
				Endpoint:    "/Groups",
				Description: optional.NewString("Group"),
				Schema:      schema.CoreGroupSchema(),
				Handler: &testResourceHandler{
					data: map[string]testData{
						"0001": {attributes: map[string]interface{}{"displayName": "testGroup"}},
						"0002": {attributes: map[string]interface{}{"displayName": "testGroup+test"}},
					},
					schema: schema.CoreGroupSchema(),
				},
			},
		},
		logging.NullLogger{},
	)
}
