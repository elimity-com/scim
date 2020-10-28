package scim

import (
	"github.com/elimity-com/scim/errors"
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	if expected != actual {
		t.Errorf("not equal: expected %v, actual %v", expected, actual)
	}
}

func assertEqualSCIMErrors(t *testing.T, expected, actual *errors.ScimError) {
	if expected.ScimType != actual.ScimType ||
		expected.Detail != actual.Detail ||
		expected.Status != actual.Status {
		t.Errorf("wrong scim error: expected %v, actual %v", expected, actual)
	}
}

func assertEqualStatusCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("status code mismatch: expected %d, actual %d", expected, actual)
	}
}

func assertEqualStrings(t *testing.T, expected, actual []string) {
	assertLen(t, actual, len(expected))
	for i, id := range expected {
		if rID := actual[i]; rID != id {
			t.Errorf("%s is not equal to %sd", rID, id)
		}
	}
}

func assertFalse(t *testing.T, ok bool) {
	if ok {
		t.Error("value should be false")
	}
}

func assertLen(t *testing.T, object interface{}, length int) {
	ok, l := getLen(object)
	if !ok {
		t.Errorf("given object is not a slice/array")
	}
	if l != length {
		t.Errorf("expected %d entities, got %d", length, l)
	}
}

func assertNil(t *testing.T, object interface{}, name string) {
	if object != nil {
		t.Errorf("object should be nil: %s", name)
	}
}

func assertNotEqual(t *testing.T, expected, actual interface{}) {
	if expected == actual {
		t.Errorf("%v and %v should not be equal", expected, actual)
	}
}

func assertNotNil(t *testing.T, object interface{}, name string) {
	if object == nil {
		t.Errorf("missing object: %s", name)
	}
}

func assertTrue(t *testing.T, ok bool) {
	if !ok {
		t.Error("value should be true")
	}
}

func assertTypeOk(t *testing.T, ok bool, expectedType string) {
	if !ok {
		t.Errorf("type is not a(n) %s", expectedType)
	}
}

func assertUnmarshalNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("json unmarshalling failed: %s", err)
	}
}

func getLen(x interface{}) (ok bool, length int) {
	v := reflect.ValueOf(x)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()
	return true, v.Len()
}
