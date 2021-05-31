package idp_test

// These tests are based on: https://docs.microsoft.com/en-us/azure/active-directory/app-provisioning/use-scim-to-provision-users-and-groups#group-operations
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

func TestAzureADGroup(t *testing.T) {
	t.Run("Create Group", azureADCreateGroup)
	t.Run("Get Group", azureADGetGroup)
	t.Run("Get Group By Display Name", azureADGetGroupByDisplayName)
	t.Run("Update Group", azureADUpdateGroup)
}

func azureADCreateGroup(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/azuread/create_group.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPost, "/Groups", bytes.NewReader(rawReq))
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
	rawResp, err := ioutil.ReadFile("testdata/azuread/create_group_resp.json")
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

func azureADGetGroup(t *testing.T) {
	var (
		req    = httptest.NewRequest(http.MethodGet, "/Groups/40734ae655284ad3abcc", nil)
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
	rawResp, err := ioutil.ReadFile("testdata/azuread/get_group_resp.json")
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

func azureADGetGroupByDisplayName(t *testing.T) {
	var (
		filter = url.QueryEscape(`displayName eq "displayName"`)
		req    = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/Groups?filter=%s", filter), nil)
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
	rawResp, err := ioutil.ReadFile("testdata/azuread/get_group_dn_resp.json")
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

func azureADUpdateGroup(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/azuread/update_group.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPatch, "/Groups/fa2ce26709934589afc5", bytes.NewReader(rawReq))
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
	rawResp, err := ioutil.ReadFile("testdata/azuread/update_group_resp.json")
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
