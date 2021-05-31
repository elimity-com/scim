package idp_test

// These tests are based on: https://docs.microsoft.com/en-us/azure/active-directory/app-provisioning/use-scim-to-provision-users-and-groups#user-operations
// Date: 31 May 2021

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAzureADUser(t *testing.T) {
	t.Run("Create User", azureADCreateUser)
	t.Run("Get User", azureADGetUser)
	t.Run("Get User Not Found", azureADGetUser404)
	t.Run("Get User By Query", azureGetUserByQuery)
	t.Run("Get User By Query Zero Results", azureGetUserByQuery0)
	t.Run("Update User", azureUpdateUser)
}

func azureADCreateUser(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/azuread/create_user.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPost, "/Users", bytes.NewReader(rawReq))
		rr     = httptest.NewRecorder()
		server = newAzureADTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/azuread/create_user_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}
	if !deepEqual(reference, response) {
		t.Error(nonNilAttributes(reference))
		t.Error(nonNilAttributes(response))
	}
}

func azureADGetUser(t *testing.T) {
	var (
		req    = httptest.NewRequest(http.MethodGet, "/Users/5d48a0a8e9f04aa38008", nil)
		rr     = httptest.NewRecorder()
		server = newAzureADTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/azuread/get_user_resp.json")
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

func azureADGetUser404(t *testing.T) {
	var (
		req    = httptest.NewRequest(http.MethodGet, "/Users/5171a35d82074e068ce2", nil)
		rr     = httptest.NewRecorder()
		server = newAzureADTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatal(rr.Code, rr.Body.String())
	}
}

func azureGetUserByQuery(t *testing.T) {
	var (
		filter = url.QueryEscape(`userName eq "Test_User_dfeef4c5-5681-4387-b016-bdf221e82081"`)
		req    = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/Users?filter=%s", filter), nil)
		rr     = httptest.NewRecorder()
		server = newAzureADTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/azuread/get_user_query_resp.json")
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

func azureGetUserByQuery0(t *testing.T) {
	var (
		filter = url.QueryEscape(`userName eq "non-existent user"`)
		req    = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/Users?filter=%s", filter), nil)
		rr     = httptest.NewRecorder()
		server = newAzureADTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/azuread/get_user_query_0_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}

	if !deepEqual(reference, response) {
		t.Error(reference)
		t.Error(response)
	}
}

func azureUpdateUser(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/azuread/update_user.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPatch, "/Users/6764549bef60420686bc", bytes.NewReader(rawReq))
		rr     = httptest.NewRecorder()
		server = newAzureADTestServer()
	)
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var response map[string]interface{}
	if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	rawResp, err := ioutil.ReadFile("testdata/azuread/update_user_resp.json")
	if err != nil {
		t.Fatal(err)
	}
	var reference map[string]interface{}
	if err := unmarshal(rawResp, &reference); err != nil {
		t.Fatal(err)
	}
	if !deepEqual(reference, response) {
		t.Error(reference)
		t.Error(response)
	}
}
