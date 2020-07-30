package filter

import (
	"fmt"
	"strconv"
	"strings"

	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/schema"
)

var invalid = func(i interface{}) bool {
	return false
}

func eq(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "binary":
		return func(value interface{}) bool {
			str, ok := value.(string)
			if !ok {
				return false
			}
			return compValue == str
		}
	case "boolean":
		boolean, err := strconv.ParseBool(compValue)
		if err != nil {
			return invalid
		}
		return func(value interface{}) bool {
			t, ok := value.(bool)
			if !ok {
				return false
			}
			return t == boolean
		}
	case "decimal", "integer", "dateTime", "reference", "string":
		f, _ := c(compValue, attr, func(i int) bool {
			return i == 0
		})
		return f
	}
	return invalid
}

func co(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal", "integer":
		return func(value interface{}) bool {
			return strings.Contains(fmt.Sprint(value), compValue)
		}
	case "boolean":
		return func(value interface{}) bool {
			return strings.Contains(fmt.Sprintf("%t", value), compValue)
		}
	case "binary", "reference", "dateTime", "string":
		return func(value interface{}) bool {
			str, ok := value.(string)
			if !ok {
				return false
			}

			if !attr.CaseExact() {
				return strings.Contains(strings.ToLower(str), strings.ToLower(compValue))
			}
			return strings.Contains(str, compValue)
		}
	}
	return invalid
}

func sw(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal", "integer":
		return func(value interface{}) bool {
			return strings.HasPrefix(fmt.Sprint(value), compValue)
		}
	case "boolean":
		return func(value interface{}) bool {
			return strings.HasPrefix(fmt.Sprintf("%t", value), compValue)
		}
	case "binary", "reference", "dateTime", "string":
		return func(value interface{}) bool {
			str, ok := value.(string)
			if !ok {
				return false
			}

			if !attr.CaseExact() {
				return strings.HasPrefix(strings.ToLower(str), strings.ToLower(compValue))
			}
			return strings.HasPrefix(str, compValue)
		}
	}
	return invalid
}

func ew(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal", "integer":
		return func(value interface{}) bool {
			return strings.HasSuffix(fmt.Sprint(value), compValue)
		}
	case "boolean":
		return func(value interface{}) bool {
			return strings.HasSuffix(fmt.Sprintf("%t", value), compValue)
		}
	case "binary", "reference", "dateTime", "string":
		return func(value interface{}) bool {
			str, ok := value.(string)
			if !ok {
				return false
			}

			if !attr.CaseExact() {
				return strings.HasSuffix(strings.ToLower(str), strings.ToLower(compValue))
			}
			return strings.HasSuffix(str, compValue)
		}
	}
	return invalid
}

func c(compValue string, attr schema.CoreAttribute, comp func(int) bool) (func(interface{}) bool, error) {
	switch typ := attr.AttributeType(); typ {
	case "decimal":
		f64, err := strconv.ParseFloat(compValue, 64)
		if err != nil {
			return nil, err
		}
		return func(i interface{}) bool {
			var value float64
			switch i := i.(type) {
			case float32:
				value = float64(i)
			case float64:
				value = i
			default:
				return false
			}
			if value == f64 {
				return comp(0)
			}
			if value < f64 {
				return comp(-1)
			}
			return comp(1)
		}, nil
	case "integer":
		i64, err := strconv.ParseInt(compValue, 10, 64)
		if err != nil {
			return nil, err
		}
		return func(i interface{}) bool {
			var value int64
			switch i := i.(type) {
			case int:
				value = int64(i)
			case int32:
				value = int64(i)
			case int64:
				value = i
			default:
				return false
			}
			if value == i64 {
				return comp(0)
			}
			if value < i64 {
				return comp(-1)
			}
			return comp(1)
		}, nil
	case "binary", "boolean":
		return nil, invalidFilterError(fmt.Sprintf("invalid attribute type %q", typ))
	case "dateTime":
		return func(value interface{}) bool {
			date, ok := value.(string)
			if !ok {
				return false
			}
			dateValue, err := datetime.Parse(date)
			if err != nil {
				return false
			}
			dateCompValue, err := datetime.Parse(compValue)
			if err != nil {
				return false
			}

			if dateValue.Equal(dateCompValue) {
				return comp(0)
			}
			if dateValue.Before(dateCompValue) {
				return comp(-1)
			}
			return comp(1)
		}, nil
	case "string", "reference":
		return func(value interface{}) bool {
			str, ok := value.(string)
			if !ok {
				return false
			}

			if !attr.CaseExact() {
				return comp(strings.Compare(strings.ToLower(str), strings.ToLower(compValue)))
			}
			return comp(strings.Compare(str, compValue))
		}, nil
	}
	// should never happen
	return invalid, nil
}
