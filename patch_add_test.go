package scim_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPatch_addAttributes(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/patch/add/attributes.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req = httptest.NewRequest(http.MethodPatch, "/Users/0001", bytes.NewReader(raw))
		rr  = httptest.NewRecorder()
	)
	newTestServer().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	rm, ok := resource["emails"]
	if !ok {
		t.Fatal(resource["emails"])
	}
	rl, ok := rm.([]interface{})
	if !ok {
		t.Fatal(rm)
	}
	if len(rl) != 1 {
		t.Fatal(rl)
	}
	m, ok := rl[0].(map[string]interface{})
	if !ok {
		t.Fatal(rl[0])
	}
	if m["value"] != "babs@jensen.org" {
		t.Error(m["value"])
	}
	if m["type"] != "home" {
		t.Error(m["type"])
	}
	nn, ok := resource["nickname"]
	if !ok {
		t.Fatal(nn)
	}
	if nn != "Babs" {
		t.Error(nn)
	}
}

func TestPatch_addMember(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/patch/add/member.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req = httptest.NewRequest(http.MethodPatch, "/Groups/0001", bytes.NewReader(raw))
		rr  = httptest.NewRecorder()
	)
	newTestServer().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	rm, ok := resource["members"]
	if !ok {
		t.Fatal(resource["members"])
	}
	rl, ok := rm.([]interface{})
	if !ok {
		t.Fatal(rm)
	}
	if len(rl) != 1 {
		t.Fatal(rl)
	}
	m, ok := rl[0].(map[string]interface{})
	if !ok {
		t.Fatal(rl[0])
	}
	if m["display"] != "Babs Jensen" {
		t.Error(m["display"])
	}
	if m["$ref"] != "https://example.com/v2/Users/2819c223-7f76-453a-919d-413861904646" {
		t.Error(m["$ref"])
	}
	if m["value"] != "2819c223-7f76-453a-919d-413861904646" {
		t.Error(m["value"])
	}
}

func TestPatch_alreadyExists(t *testing.T) {
	for _, test := range []struct {
		jsonFilePath string
		targetPath   string
		changed      bool
	}{
		{
			jsonFilePath: "testdata/patch/add/member.json",
			targetPath:   "/Groups/0001",
			changed:      true,
		},
		{
			jsonFilePath: "testdata/patch/add/attributes.json",
			targetPath:   "/Users/0001",
			changed:      true,
		},
		{
			jsonFilePath: "testdata/patch/add/complex.json",
			targetPath:   "/Users/0001",
			changed:      false,
		},
	} {
		server := newTestServer()
		raw, err := ioutil.ReadFile(test.jsonFilePath)
		if err != nil {
			t.Fatal(err)
		}
		var (
			req = httptest.NewRequest(http.MethodPatch, test.targetPath, bytes.NewReader(raw))
			rr  = httptest.NewRecorder()
		)
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatal(rr.Code, rr.Body.String())
		}
		req = httptest.NewRequest(http.MethodPatch, test.targetPath, bytes.NewReader(raw))
		rr = httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if test.changed {
			if rr.Code != http.StatusOK {
				t.Error(rr.Code, rr.Body.String())
			}
		} else {
			if rr.Code != http.StatusNoContent {
				t.Error(rr.Code, rr.Body.String())
			}
		}
	}
}

func TestPatch_complex(t *testing.T) {
	raw, err := ioutil.ReadFile("testdata/patch/add/complex.json")
	if err != nil {
		t.Fatal(err)
	}
	var (
		req = httptest.NewRequest(http.MethodPatch, "/Users/0001", bytes.NewReader(raw))
		rr  = httptest.NewRecorder()
	)
	newTestServer().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatal(rr.Code, rr.Body.String())
	}
	var resource map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	rm, ok := resource["name"]
	if !ok {
		t.Fatal(resource["members"])
	}
	m, ok := rm.(map[string]interface{})
	if !ok {
		t.Fatal(rm)
	}
	if m["givenName"] != "Barbara" {
		t.Error(m["givenName"])
	}
	if m["familyName"] != "Jensen" {
		t.Error(m["familyName"])
	}
	if m["formatted"] != "Barbara Jensen" {
		t.Error(m["formatted"])
	}
}
