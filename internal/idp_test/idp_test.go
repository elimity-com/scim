package idp_test

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

//go:embed testdata
var testdata embed.FS

func TestIdP(t *testing.T) {
	idps, _ := testdata.ReadDir("testdata")
	for _, idp := range idps {
		t.Run(idp.Name(), func(t *testing.T) {
			idpPath := fmt.Sprintf("testdata/%s", idp.Name())
			de, _ := fs.ReadDir(testdata, idpPath)
			for _, f := range de {
				path := fmt.Sprintf("%s/%s", idpPath, f.Name())
				raw, _ := fs.ReadFile(testdata, path)
				var test testCase
				_ = unmarshal(raw, &test)
				t.Run(strings.TrimSuffix(f.Name(), ".json"), func(t *testing.T) {
					if err := testRequest(test, idp.Name()); err != nil {
						t.Error(err)
					}
				})
			}
		})
	}
}

func testRequest(t testCase, idpName string) error {
	rr := httptest.NewRecorder()
	br := bytes.NewReader(t.Request)
	server, err := getNewServer(idpName)
	if err != nil {
		return err
	}
	server.ServeHTTP(
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
