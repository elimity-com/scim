package filter

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser"
)

func TestFilterBoolean(t *testing.T) {
	for i, test := range []struct {
		comparator string
		invalid    bool
	}{
		{
			comparator: "eq",
			invalid:    false,
		},
		{
			comparator: "ne",
			invalid:    true,
		},
		{
			comparator: "co",
			invalid:    false,
		},
		{
			comparator: "sw",
			invalid:    false,
		},
		{
			comparator: "ew",
			invalid:    false,
		},
		{
			comparator: "gt",
			invalid:    true,
		},
		{
			comparator: "ge",
			invalid:    true,
		},
		{
			comparator: "lt",
			invalid:    true,
		},
		{
			comparator: "le",
			invalid:    true,
		},
	} {
		t.Run(test.comparator, func(t *testing.T) {
			f := newTypeFilter(fmt.Sprintf("bool %s \"true\"", test.comparator))
			valid, err := f.IsValid(map[string]interface{}{
				"bool": true,
			})
			if i <= 4 {
				if err != nil {
					t.Error(err)
				}
			} else {
				// can not compare gt, ge, lt, ge
				if err == nil {
					t.Errorf("error expected, got none")
				}
			}

			if test.invalid {
				if valid {
					t.Errorf("resource should be invalid")
				}
				return
			}

			if !valid {
				t.Errorf("resource should be valid")
			}
		})
	}
}

func TestFilterDateTime(t *testing.T) {
	testCases := []struct {
		filter   string
		resource map[string]interface{}
	}{
		{
			filter: "time %s \"2020-07-30T12:32:11Z\"",
			resource: map[string]interface{}{
				"time": "2020-07-30T12:32:11Z",
			},
		},
		{
			filter: "time %s \"2020-07-30T12:32:11Z\"",
			resource: map[string]interface{}{
				"time": "2020-07-01T12:32:11Z",
			},
		},
		{
			filter: "time %s \"2020-07-30T12:32:11Z\"",
			resource: map[string]interface{}{
				"time": "2020-08-01T11:45:56Z",
			},
		},
	}

	for _, test := range []struct {
		comparator string
		invalid    []bool
	}{
		{
			comparator: "eq",
			invalid: []bool{
				false, true, true,
			},
		},
		{
			comparator: "ne",
			invalid: []bool{
				true, false, false,
			},
		},
		{
			comparator: "co",
			invalid: []bool{
				false, true, true,
			},
		},
		{
			comparator: "sw",
			invalid: []bool{
				false, true, true,
			},
		},
		{
			comparator: "ew",
			invalid: []bool{
				false, true, true,
			},
		},
		{
			comparator: "gt",
			invalid: []bool{
				true, true, false,
			},
		},
		{
			comparator: "ge",
			invalid: []bool{
				false, true, false,
			},
		},
		{
			comparator: "lt",
			invalid: []bool{
				true, false, true,
			},
		},
		{
			comparator: "le",
			invalid: []bool{
				false, false, true,
			},
		},
	} {
		t.Run(test.comparator, func(t *testing.T) {
			for i, c := range testCases {
				t.Run(c.filter, func(t *testing.T) {
					f := newTypeFilter(fmt.Sprintf(c.filter, test.comparator))
					valid, err := f.IsValid(c.resource)
					if err != nil {
						t.Error(err)
					}

					if test.invalid[i] {
						if valid {
							t.Errorf("resource should be invalid")
						}
						return
					}

					if !valid {
						t.Errorf("resource should be valid")
					}
				})
			}
		})
	}
}

func TestFilterNumber(t *testing.T) {
	testCases := []struct {
		filter   string
		resource map[string]interface{}
	}{
		{
			filter: "dec %s \"0\"",
			resource: map[string]interface{}{
				"dec": 0.0,
			},
		},
		{
			filter: "dec %s \"0\"",
			resource: map[string]interface{}{
				"dec": 0.1,
			},
		},
		{
			filter: "dec %s \"0\"",
			resource: map[string]interface{}{
				"dec": float64(-1),
			},
		},
		{
			filter: "dec %s \"0\"",
			resource: map[string]interface{}{
				"dec": float32(1),
			},
		},

		{
			filter: "int %s \"0\"",
			resource: map[string]interface{}{
				"int": 0,
			},
		},
		{
			filter: "int %s \"0\"",
			resource: map[string]interface{}{
				"int": int32(-1),
			},
		},
		{
			filter: "int %s \"0\"",
			resource: map[string]interface{}{
				"int": int64(1),
			},
		},
	}

	for _, test := range []struct {
		comparator string
		invalid    []bool
	}{
		{
			comparator: "eq",
			invalid: []bool{
				false, true, true, true,
				false, true, true,
			},
		},
		{
			comparator: "ne",
			invalid: []bool{
				true, false, false, false,
				true, false, false,
			},
		},
		{
			comparator: "co",
			invalid: []bool{
				false, false, true, true,
				false, true, true,
			},
		},
		{
			comparator: "sw",
			invalid: []bool{
				false, false, true, true,
				false, true, true,
			},
		},
		{
			comparator: "ew",
			invalid: []bool{
				false, true, true, true,
				false, true, true,
			},
		},
		{
			comparator: "gt",
			invalid: []bool{
				true, false, true, false,
				true, true, false,
			},
		},
		{
			comparator: "ge",
			invalid: []bool{
				false, false, true, false,
				false, true, false,
			},
		},
		{
			comparator: "lt",
			invalid: []bool{
				true, true, false, true,
				true, false, true,
			},
		},
		{
			comparator: "le",
			invalid: []bool{
				false, true, false, true,
				false, false, true,
			},
		},
	} {
		t.Run(test.comparator, func(t *testing.T) {
			for i, c := range testCases {
				t.Run(c.filter, func(t *testing.T) {
					f := newTypeFilter(fmt.Sprintf(c.filter, test.comparator))
					valid, err := f.IsValid(c.resource)
					if err != nil {
						t.Error(err)
					}

					if test.invalid[i] {
						if valid {
							t.Errorf("resource should be invalid")
						}
						return
					}

					if !valid {
						t.Errorf("resource should be valid")
					}
				})
			}
		})
	}
}

func TestFilterString(t *testing.T) {
	testCases := []struct {
		filter   string
		resource map[string]interface{}
	}{
		{
			filter: "str %s \"x\"",
			resource: map[string]interface{}{
				"str": "x",
			},
		},
		{
			filter: "str %s \"x\"",
			resource: map[string]interface{}{
				"str": "!x",
			},
		},
		{
			filter: "str %s \"X\"",
			resource: map[string]interface{}{
				"str": "x",
			},
		},

		{
			filter: "strCE %s \"x\"",
			resource: map[string]interface{}{
				"strCE": "x",
			},
		},
		{
			filter: "strCE %s \"x\"",
			resource: map[string]interface{}{
				"strCE": "!x",
			},
		},
		{
			filter: "strCE %s \"X\"",
			resource: map[string]interface{}{
				"strCE": "x",
			},
		},
	}

	for _, test := range []struct {
		comparator string
		invalid    []bool
	}{
		{
			comparator: "eq",
			invalid: []bool{
				false, true, false,
				false, true, true,
			},
		},
		{
			comparator: "ne",
			invalid: []bool{
				true, false, true,
				true, false, false,
			},
		},
		{
			comparator: "co",
			invalid: []bool{
				false, false, false,
				false, false, true,
			},
		},
		{
			comparator: "sw",
			invalid: []bool{
				false, true, false,
				false, true, true,
			},
		},
		{
			comparator: "ew",
			invalid: []bool{
				false, false, false,
				false, false, true,
			},
		},
		{
			comparator: "gt",
			invalid: []bool{
				true, true, true,
				true, true, false,
			},
		},
		{
			comparator: "ge",
			invalid: []bool{
				false, true, false,
				false, true, false,
			},
		},
		{
			comparator: "lt",
			invalid: []bool{
				true, false, true,
				true, false, true,
			},
		},
		{
			comparator: "le",
			invalid: []bool{
				false, false, false,
				false, false, true,
			},
		},
	} {
		t.Run(test.comparator, func(t *testing.T) {
			for i, c := range testCases {
				t.Run(c.filter, func(t *testing.T) {
					f := newTypeFilter(fmt.Sprintf(c.filter, test.comparator))
					valid, err := f.IsValid(c.resource)
					if err != nil {
						t.Error(err)
					}

					if test.invalid[i] {
						if valid {
							t.Errorf("resource should be invalid")
						}
						return
					}

					if !valid {
						t.Errorf("resource should be valid")
					}
				})
			}
		})
	}
}

func newTypeFilter(f string) Filter {
	parser := filter.NewParser(strings.NewReader(f))
	exp, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	return Filter{
		Expression: exp,
		schema: schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
					Name: "str",
				})),
				schema.SimpleCoreAttribute(schema.SimpleReferenceParams(schema.ReferenceParams{
					Name: "strCE",
				})),
				schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
					Name: "dec",
					Type: schema.AttributeTypeDecimal(),
				})),
				schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
					Name: "int",
					Type: schema.AttributeTypeInteger(),
				})),
				schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
					Name: "bool",
				})),
				schema.SimpleCoreAttribute(schema.SimpleDateTimeParams(schema.DateTimeParams{
					Name: "time",
				})),
			},
		},
	}
}
