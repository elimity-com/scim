package filter

import (
	"fmt"
	"strings"

	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/schema"
)

func eq(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal":
	case "integer":
	case "binary":
	case "boolean":
	case "complex":
	case "dateTime":
	case "reference":
	default: // "string"
		return func(value interface{}) bool {
			str, ok := value.(string)
			if !ok {
				return false
			}

			if !attr.CaseExact() {
				return strings.EqualFold(str, compValue)
			}
			return str == compValue
		}
	}
	return nil
}

func co(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal":
	case "integer":
	case "binary":
	case "boolean":
	case "complex":
	case "dateTime":
	case "reference":
	default: // "string"
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
	return nil
}

func sw(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal":
	case "integer":
	case "binary":
	case "boolean":
	case "complex":
	case "dateTime":
	case "reference":
	default: // "string"
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
	return nil
}

func ew(compValue string, attr schema.CoreAttribute) func(interface{}) bool {
	switch attr.AttributeType() {
	case "decimal":
	case "integer":
	case "binary":
	case "boolean":
	case "complex":
	case "dateTime":
	case "reference":
	default: // "string"
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
	return nil
}

func c(compValue string, attr schema.CoreAttribute, comp func(int) bool) (func(interface{}) bool, error) {
	switch typ := attr.AttributeType(); typ {
	case "decimal":
	case "integer":
	case "binary", "boolean":
		return nil, invalidFilterError(fmt.Sprintf("invalid attribute type %q", typ))
	case "complex":
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
	case "reference":
	default: // "string"
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
	return nil, nil
}
