package filter_test

import (
	"fmt"
	"github.com/elimity-com/scim/errors"
	internal "github.com/elimity-com/scim/internal/filter"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"testing"
)

func TestValidatorBoolean(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("bool %s true", op)
		}
		ref = schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
					Name: "bool",
				})),
			},
		}
		attr = map[string]interface{}{
			"bool": true,
		}
	)

	for _, test := range []struct {
		op    filter.CompareOperator
		valid bool // Whether the filter is valid.
	}{
		{filter.EQ, true},
		{filter.NE, false},
		{filter.CO, true},
		{filter.SW, true},
		{filter.EW, true},
		{filter.GT, false},
		{filter.LT, false},
		{filter.GE, false},
		{filter.LE, false},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			validator, err := internal.NewValidator(f, ref)
			if err != nil {
				t.Fatal(err)
			}
			if err := validator.PassesFilter(attr); (err == nil) != test.valid {
				t.Errorf("%s %v | actual %v, expected %v", f, attr, err, test.valid)
			}
		})
	}
}

func TestValidatorDateTime(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("time %s \"2021-01-01T12:00:00Z\"", op)
		}
		ref = schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleDateTimeParams(schema.DateTimeParams{
					Name: "time",
				})),
			},
		}
		attrs = [3]map[string]interface{}{
			{"time": "2021-01-01T08:00:00Z"}, // before
			{"time": "2021-01-01T12:00:00Z"}, // equal
			{"time": "2021-01-01T16:00:00Z"}, // after
		}
	)

	for _, test := range []struct {
		op    filter.CompareOperator
		valid [3]bool
	}{
		{filter.EQ, [3]bool{false, true, false}},
		{filter.NE, [3]bool{true, false, true}},
		{filter.CO, [3]bool{false, true, false}},
		{filter.SW, [3]bool{false, true, false}},
		{filter.EW, [3]bool{false, true, false}},
		{filter.GT, [3]bool{false, false, true}},
		{filter.LT, [3]bool{true, false, false}},
		{filter.GE, [3]bool{false, true, true}},
		{filter.LE, [3]bool{true, true, false}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			validator, err := internal.NewValidator(f, ref)
			if err != nil {
				t.Fatal(err)
			}
			for i, attr := range attrs {
				if err := validator.PassesFilter(attr); (err == nil) != test.valid[i] {
					t.Errorf("(%d) %s %v | actual %v, expected %v", i, f, attr, err, test.valid[i])
				}
			}
		})
	}
}

func TestValidatorDecimal(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("dec %s 1", op)
		}
		ref = schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
					Name: "dec",
					Type: schema.AttributeTypeDecimal(),
				})),
			},
		}
		attrs = [4]map[string]interface{}{
			{"dec": -0.1},       // less
			{"dec": float64(1)}, // equal
			{"dec": float32(1)}, // equal f32
			{"dec": 1.1},        // greater
		}
	)

	for _, test := range []struct {
		op    filter.CompareOperator
		valid [4]bool
	}{
		{filter.EQ, [4]bool{false, true, true, false}},
		{filter.NE, [4]bool{true, false, false, true}},
		{filter.CO, [4]bool{true, true, true, true}},
		{filter.SW, [4]bool{false, true, true, true}},
		{filter.EW, [4]bool{true, true, true, true}},
		{filter.GT, [4]bool{false, false, false, true}},
		{filter.LT, [4]bool{true, false, false, false}},
		{filter.GE, [4]bool{false, true, true, true}},
		{filter.LE, [4]bool{true, true, true, false}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			validator, err := internal.NewValidator(f, ref)
			if err != nil {
				t.Fatal(err)
			}
			for i, attr := range attrs {
				if err := validator.PassesFilter(attr); (err == nil) != test.valid[i] {
					t.Errorf("(%d) %s %v | actual %v, expected %v", i, f, attr, err, test.valid[i])
				}
			}
		})
	}
}

func TestValidatorInteger(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("int %s 0", op)
		}
		ref = schema.Schema{
			Attributes: []schema.CoreAttribute{
				schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
					Name: "int",
					Type: schema.AttributeTypeInteger(),
				})),
			},
		}
		attrs = [5]map[string]interface{}{
			{"int": -1},       // less
			{"int": int64(0)}, // equal i64
			{"int": int32(0)}, // equal i32
			{"int": 0},        // equal
			{"int": 10},       // greater
		}
	)

	for _, test := range []struct {
		op    filter.CompareOperator
		valid [5]bool
	}{
		{filter.EQ, [5]bool{false, true, true, true, false}},
		{filter.NE, [5]bool{true, false, false, false, true}},
		{filter.CO, [5]bool{false, true, true, true, true}},
		{filter.SW, [5]bool{false, true, true, true, false}},
		{filter.EW, [5]bool{false, true, true, true, true}},
		{filter.GT, [5]bool{false, false, false, false, true}},
		{filter.LT, [5]bool{true, false, false, false, false}},
		{filter.GE, [5]bool{false, true, true, true, true}},
		{filter.LE, [5]bool{true, true, true, true, false}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			validator, err := internal.NewValidator(f, ref)
			if err != nil {
				t.Fatal(err)
			}
			for i, attr := range attrs {
				if err := validator.PassesFilter(attr); (err == nil) != test.valid[i] {
					t.Errorf("(%d) %s %v | actual %v, expected %v", i, f, attr, err, test.valid[i])
				}
			}
		})
	}
}

// TestValidatorInvalidResourceTypes contains all the cases where an *errors.ScimError gets returned.
func TestValidatorInvalidResourceTypes(t *testing.T) {
	for _, test := range []struct {
		name     string
		filter   string
		attr     schema.CoreAttribute
		resource map[string]interface{}
	}{
		{
			"string", `attr eq "value"`,
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": 1, // expects a string
			},
		},
		{
			"stringMv", `attr eq "value"`,
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:        "attr",
				MultiValued: true,
			})),
			map[string]interface{}{
				"attr": []interface{}{1}, // expects a []interface{string}
			},
		},
		{
			"stringMv",
			`attr eq "value"`,
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Name:        "attr",
				MultiValued: true,
			})),
			map[string]interface{}{
				"attr": []string{"value"}, // expects a []interface{}
			},
		},
		{
			"dateTime", `attr eq "2006-01-02T15:04:05"`,
			schema.SimpleCoreAttribute(schema.SimpleDateTimeParams(schema.DateTimeParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": 1, // expects a string
			},
		},
		{
			"dateTime", `attr eq "2006-01-02T15:04:05"`,
			schema.SimpleCoreAttribute(schema.SimpleDateTimeParams(schema.DateTimeParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": "2006-01-02T", // expects a valid dateTime
			},
		},
		{
			"boolean", `attr eq true`,
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Name: "attr",
			})),
			map[string]interface{}{
				"attr": 1, // expects a boolean
			},
		},
		{
			"decimal", `attr eq 0`,
			schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
				Name: "attr",
				Type: schema.AttributeTypeDecimal(),
			})),
			map[string]interface{}{
				"attr": "0", // expects a boolean
			},
		},
		{
			"integer", `attr eq 0`,
			schema.SimpleCoreAttribute(schema.SimpleNumberParams(schema.NumberParams{
				Name: "attr",
				Type: schema.AttributeTypeInteger(),
			})),
			map[string]interface{}{
				"attr": 0.0, // expects an integer
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			validator, err := internal.NewValidator(test.filter, schema.Schema{
				Attributes: []schema.CoreAttribute{test.attr},
			})
			if err != nil {
				t.Fatal(err)
			}
			scimErr := validator.PassesFilter(test.resource)
			if scimErr == nil {
				t.Fatal()
			}
			if _, ok := scimErr.(*errors.ScimError); !ok {
				t.Error(scimErr)
			}
		})
	}
}

func TestValidatorString(t *testing.T) {
	var (
		exp = func(op filter.CompareOperator) string {
			return fmt.Sprintf("str %s \"x\"", op)
		}
		attrs = [3]map[string]interface{}{
			{"str": "x"},
			{"str": "X"},
			{"str": "y"},
		}
	)

	for _, test := range []struct {
		op      filter.CompareOperator
		valid   [3]bool
		validCE [3]bool
	}{
		{filter.EQ, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.NE, [3]bool{false, false, true}, [3]bool{false, true, true}},
		{filter.CO, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.SW, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.EW, [3]bool{true, true, false}, [3]bool{true, false, false}},
		{filter.GT, [3]bool{false, false, true}, [3]bool{false, false, true}},
		{filter.LT, [3]bool{false, false, false}, [3]bool{false, true, false}},
		{filter.GE, [3]bool{true, true, true}, [3]bool{true, false, true}},
		{filter.LE, [3]bool{true, true, false}, [3]bool{true, true, false}},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			f := exp(test.op)
			for i, attr := range attrs {
				validator, err := internal.NewValidator(f, schema.Schema{
					Attributes: []schema.CoreAttribute{
						schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
							Name: "str",
						})),
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := validator.PassesFilter(attr); (err == nil) != test.valid[i] {
					t.Errorf("(0.%d) %s %s | actual %v, expected %v", i, f, attr, err, test.valid[i])
				}
				validatorCE, err := internal.NewValidator(f, schema.Schema{
					Attributes: []schema.CoreAttribute{
						schema.SimpleCoreAttribute(schema.SimpleReferenceParams(schema.ReferenceParams{
							Name: "str",
						})),
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := validatorCE.PassesFilter(attr); (err == nil) != test.validCE[i] {
					t.Errorf("(1.%d) %s %s | actual %v, expected %v", i, f, attr, err, test.validCE[i])
				}
			}
		})
	}
}
