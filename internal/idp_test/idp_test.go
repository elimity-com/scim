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
					if err := testRequest(t, test, idp.Name()); err != nil {
						t.Error(err)
					}
				})
			}
		})
	}
}

func testRequest(t *testing.T, tc testCase, idpName string) error {
	rr := httptest.NewRecorder()
	br := bytes.NewReader(tc.Request)
	getNewServer(t, idpName).ServeHTTP(
		rr,
		httptest.NewRequest(tc.Method, tc.Path, br),
	)
	if code := rr.Code; code != tc.StatusCode {
		return fmt.Errorf("expected %d, got %d", tc.StatusCode, code)
	}
	if len(tc.Response) != 0 {
		var response map[string]interface{}
		if err := unmarshal(rr.Body.Bytes(), &response); err != nil {
			return err
		}
		if !reflect.DeepEqual(tc.Response, response) {
			return fmt.Errorf("expected, got:\n%v\n%v", tc.Response, response)
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
