package idp_test

// These tests are based on: https://developer.okta.com/docs/reference/scim/scim-20/#scim-user-operations
// Date: 31 May 2021

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestOktaUser(t *testing.T) {
	t.Run("Determine User Exists", oktaDetermineUserExists)
	t.Run("Create User", oktaCreateUser)
	t.Run("Retrieve Specific User", oktaRetrieveSpecificUser)
	t.Run("Update User", oktaUpdateUser)
	t.Run("Update Specific User", oktaUpdateSpecificUser)
}

func oktaCreateUser(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/okta/create_user.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPost, "/Users", bytes.NewReader(rawReq))
		rr     = httptest.NewRecorder()
		server = newOktaTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/okta/create_user_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}
	if !deepEqual(reference, response) {
		t.Error(reference, response)
	}
}

func oktaDetermineUserExists(t *testing.T) {
	var (
		filter = url.QueryEscape(`userName eq "test.user@okta.local"`)
		req    = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/Users?filter=%s&startIndex=1&count=100", filter), nil)
		rr     = httptest.NewRecorder()
		server = newOktaTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/okta/user_exists_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}

	// TODO: itemsPerPage, should this be equal to the requested amount (100 in this case) or 0, since there are no
	//       resources that match this? Okta expects: 0.
	reference["itemsPerPage"] = json.Number("100")

	if !deepEqual(reference, response) {
		t.Error(reference, response)
	}
}

func oktaRetrieveSpecificUser(t *testing.T) {
	var (
		req    = httptest.NewRequest(http.MethodGet, "/Users/23a35c27-23d3-4c03-b4c5-6443c09e7173", nil)
		rr     = httptest.NewRecorder()
		server = newOktaTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/okta/retrieve_user_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}
	if !deepEqual(reference, response) {
		t.Error(reference, response)
	}
}

func oktaUpdateSpecificUser(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/okta/update_user_patch.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPatch, "/Users/23a35c27-23d3-4c03-b4c5-6443c09e7173", bytes.NewReader(rawReq))
		rr     = httptest.NewRecorder()
		server = newOktaTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/okta/update_user_patch_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}
	if !deepEqual(reference, response) {
		t.Error(reference, response)
	}
}

func oktaUpdateUser(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/okta/update_user.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPut, "/Users/23a35c27-23d3-4c03-b4c5-6443c09e7173", bytes.NewReader(rawReq))
		rr     = httptest.NewRecorder()
		server = newOktaTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/okta/update_user_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}
	if !deepEqual(reference, response) {
		t.Error(reference, response)
	}
}
