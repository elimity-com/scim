package idp_test

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/elimity-com/scim"
)

//go:embed testdata
var testdata embed.FS

func TestIdP(t *testing.T) {
	idps, _ := testdata.ReadDir("testdata")
	for _, idp := range idps {
		var newServer func() scim.Server
		switch idp.Name() {
		case "okta":
			newServer = newOktaTestServer
		case "azuread":
			newServer = newAzureADTestServer
		}
		t.Run(idp.Name(), func(t *testing.T) {
			idpPath := fmt.Sprintf("testdata/%s", idp.Name())
			de, _ := fs.ReadDir(testdata, idpPath)
			for _, f := range de {
				path := fmt.Sprintf("%s/%s", idpPath, f.Name())
				raw, _ := fs.ReadFile(testdata, path)
				var test testCase
				_ = unmarshal(raw, &test)
				t.Run(strings.TrimSuffix(f.Name(), ".json"), func(t *testing.T) {
					if err := testRequest(test, newServer); err != nil {
						t.Error(err)
					}
				})
			}
		})
	}
}

func testRequest(t testCase, newServer func() scim.Server) error {
	rr := httptest.NewRecorder()
	var br io.Reader
	if len(t.Request) != 0 {
		br = bytes.NewReader(t.Request)
	}
	newServer().ServeHTTP(
		rr,
		httptest.NewRequest(t.Method, t.Path, br),
	)
	if code := rr.Code; code != t.StatusCode {
		return fmt.Errorf("expected %d, got %d", t.StatusCode, code)
	}
	if len(t.Response) != 0 {
		var response map[string]interface{}
		if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
			return err
		}
		if !reflect.DeepEqual(t.Response, response) {
			return fmt.Errorf("expected, got:\n%v\n%v", t.Response, response)
		}
	}
	return nil
}

type testCase struct {
	Request    json.RawMessage
	Response   map[string]interface{}
	Method     string
	Path       string
	StatusCode int
}
