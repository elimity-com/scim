package idp

// These tests are based on: https://developer.okta.com/docs/reference/scim/scim-20/#scim-group-operations
// Date: 31 May 2021

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateGroup(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/okta/create_group.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPost, "/Groups", bytes.NewReader(rawReq))
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
	rawResp, err := ioutil.ReadFile("testdata/okta/create_group_resp.json")
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

func TestRetrieveSpecificGroup(t *testing.T) {
	var (
		req    = httptest.NewRequest(http.MethodGet, "/Groups/abf4dd94-a4c0-4f67-89c9-76b03340cb9b", nil)
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
	rawResp, err := ioutil.ReadFile("testdata/okta/retrieve_group_resp.json")
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

func TestUpdateSpecificGroupMembership(t *testing.T) {
	rawReqs, err := ioutil.ReadFile("testdata/okta/update_group_membership.json")
	if err != nil {
		t.Fatal(err)
	}
	var requests []json.RawMessage
	if err := unmarshal(rawReqs, &requests); err != nil {
		t.Fatal(err)
	}
	for _, rawReq := range requests {
		var (
			req    = httptest.NewRequest(http.MethodPatch, "/Groups/abf4dd94-a4c0-4f67-89c9-76b03340cb9b", bytes.NewReader(rawReq))
			rr     = httptest.NewRecorder()
			server = newOktaTestServer()
		)
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatal(rr.Code, rr.Body.String())
		}
	}
}

func TestUpdateSpecificGroupName(t *testing.T) {
	rawReq, err := ioutil.ReadFile("testdata/okta/update_group_name.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req    = httptest.NewRequest(http.MethodPatch, "/Groups/abf4dd94-a4c0-4f67-89c9-76b03340cb9b", bytes.NewReader(rawReq))
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
	rawResp, err := ioutil.ReadFile("testdata/okta/update_group_name_resp.json")
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
