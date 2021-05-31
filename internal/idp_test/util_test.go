package idp_test

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// deepEqual is the same as reflect.DeepEqual, but taking "null" attributes into account.
func deepEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return equalsNil(a) && equalsNil(b)
	}
	v1 := reflect.ValueOf(a)
	v2 := reflect.ValueOf(a)
	if v1.Type() != v2.Type() {
		return false
	}
	switch a := a.(type) {
	case []interface{}:
		if len(a) == 0 {
			return equalsNil(b)
		}
		b := b.([]interface{})
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if !deepEqual(v, b[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		am := nonNilAttributes(a)
		bm := nonNilAttributes(b.(map[string]interface{}))
		for k, v := range am {
			if !deepEqual(v, bm[k]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}

// equalsNil checks whether the given value is nil in terms of SCIM.
// e.g. empty array "[]" == null.
func equalsNil(a interface{}) bool {
	switch a := a.(type) {
	case []interface{}:
		return len(a) == 0
	case map[string]interface{}:
		return len(a) == 0
	default:
		return a == nil
	}
}

// nonNilAttributes removes all nested nil attributes from the given resource.
func nonNilAttributes(a map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range a {
		if !equalsNil(v) {
			switch v := v.(type) {
			case []interface{}:
				switch v[0].(type) {
				case map[string]interface{}:
					var a []interface{}
					for _, v := range v {
						v := nonNilAttributes(v.(map[string]interface{}))
						a = append(a, v)
					}
					m[k] = a
				default:
					m[k] = v
				}
			case map[string]interface{}:
				m[k] = nonNilAttributes(v)
			case string:
				if v != "" {
					m[k] = v
				}
			default:
				m[k] = v
			}
		}
	}
	return m
}

func unmarshal(data []byte, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	return d.Decode(v)
}
