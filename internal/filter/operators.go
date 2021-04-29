package filter

import (
	"fmt"
	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"strings"
	"time"
)

// createCompareFunction returns a compare function based on the attribute expression and attribute.
// e.g. `userName eq "john"` will return a string comparator that checks whether the passed value is equal to "john".
func createCompareFunction(e *filter.AttributeExpression, attr schema.CoreAttribute) (func(interface{}) bool, error) {
	switch typ := attr.AttributeType(); typ {
	case "binary":
		ref, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a binary attribute needs to be compared to a string")
		}
		switch e.Operator {
		case filter.EQ:
			return strComp(ref, attr, func(v, ref string) bool { return v == ref })
		case filter.NE:
			return strComp(ref, attr, func(v, ref string) bool { return v != ref })
		case filter.CO:
			return strComp(ref, attr, strings.Contains)
		case filter.SW:
			return strComp(ref, attr, strings.HasPrefix)
		case filter.EW:
			return strComp(ref, attr, strings.HasSuffix)
		case filter.GT, filter.LT, filter.GE, filter.LE:
			return nil, fmt.Errorf("invalid attribute type %q", typ)
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	case "dateTime":
		date, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a dateTime attribute needs to be compared to a string")
		}
		ref, err := datetime.Parse(date)
		if err != nil {
			return nil, fmt.Errorf("a dateTime attribute needs to be compared to a dateTime")
		}
		switch e.Operator {
		case filter.EQ:
			return cmpDataTime(ref, func(v, ref time.Time) bool {
				return v.Equal(ref)
			}), nil
		case filter.NE:
			return cmpDataTime(ref, func(v, ref time.Time) bool {
				return !v.Equal(ref)
			}), nil
		case filter.CO:
			return strComp(date, attr, strings.Contains)
		case filter.SW:
			return strComp(date, attr, strings.HasPrefix)
		case filter.EW:
			return strComp(date, attr, strings.HasSuffix)
		case filter.GT:
			return cmpDataTime(ref, func(v, ref time.Time) bool {
				return v.After(ref)
			}), nil
		case filter.LT:
			return cmpDataTime(ref, func(v, ref time.Time) bool {
				return v.Before(ref)
			}), nil
		case filter.GE:
			return cmpDataTime(ref, func(v, ref time.Time) bool {
				return v.After(ref) || v.Equal(ref)
			}), nil
		case filter.LE:
			return cmpDataTime(ref, func(v, ref time.Time) bool {
				return v.Before(ref) || v.Equal(ref)
			}), nil
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	case "reference", "string":
		ref, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a %s attribute needs to be compared to a string", typ)
		}
		switch e.Operator {
		case filter.EQ:
			return strComp(ref, attr, func(v, ref string) bool { return v == ref })
		case filter.NE:
			return strComp(ref, attr, func(v, ref string) bool { return v != ref })
		case filter.CO:
			return strComp(ref, attr, strings.Contains)
		case filter.SW:
			return strComp(ref, attr, strings.HasPrefix)
		case filter.EW:
			return strComp(ref, attr, strings.HasSuffix)
		case filter.GT:
			return strComp(ref, attr, func(v, ref string) bool { return strings.Compare(v, ref) > 0 })
		case filter.LT:
			return strComp(ref, attr, func(v, ref string) bool { return strings.Compare(v, ref) < 0 })
		case filter.GE:
			return strComp(ref, attr, func(v, ref string) bool { return strings.Compare(v, ref) >= 0 })
		case filter.LE:
			return strComp(ref, attr, func(v, ref string) bool { return strings.Compare(v, ref) <= 0 })
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	case "boolean":
		ref, ok := e.CompareValue.(bool)
		if !ok {
			return nil, fmt.Errorf("a boolean attribute needs to be compared to a boolean")
		}
		switch e.Operator {
		case filter.EQ:
			return cmpBool(ref, func(v, ref bool) bool { return v == ref }), nil
		case filter.NE:
			return cmpBool(ref, func(v, ref bool) bool { return v != ref }), nil
		case filter.CO:
			return strComp(ref, attr, strings.Contains)
		case filter.SW:
			return strComp(ref, attr, strings.HasPrefix)
		case filter.EW:
			return strComp(ref, attr, strings.HasSuffix)
		case filter.GT, filter.LT, filter.GE, filter.LE:
			return nil, fmt.Errorf("invalid attribute type %q", typ)
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	case "decimal":
		ref, ok := toFloat(e.CompareValue)
		if !ok {
			return nil, fmt.Errorf("a decimal attribute needs to be compared to a float/int")
		}
		switch e.Operator {
		case filter.EQ:
			return cmpDecimal(ref, func(v, ref float64) bool { return v == ref }), nil
		case filter.NE:
			return cmpDecimal(ref, func(v, ref float64) bool { return v != ref }), nil
		case filter.CO:
			return strComp(ref, attr, strings.Contains)
		case filter.SW:
			return strComp(ref, attr, strings.HasPrefix)
		case filter.EW:
			return strComp(ref, attr, strings.HasSuffix)
		case filter.GT:
			return cmpDecimal(ref, func(v, ref float64) bool { return v > ref }), nil
		case filter.LT:
			return cmpDecimal(ref, func(v, ref float64) bool { return v < ref }), nil
		case filter.GE:
			return cmpDecimal(ref, func(v, ref float64) bool { return v >= ref }), nil
		case filter.LE:
			return cmpDecimal(ref, func(v, ref float64) bool { return v <= ref }), nil
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	case "integer":
		ref, ok := toInt(e.CompareValue)
		if !ok {
			return nil, fmt.Errorf("a integer attribute needs to be compared to a int")
		}
		switch e.Operator {
		case filter.EQ:
			return cmpInteger(ref, func(v, ref int) bool { return v == ref }), nil
		case filter.NE:
			return cmpInteger(ref, func(v, ref int) bool { return v != ref }), nil
		case filter.CO:
			return strComp(ref, attr, strings.Contains)
		case filter.SW:
			return strComp(ref, attr, strings.HasPrefix)
		case filter.EW:
			return strComp(ref, attr, strings.HasSuffix)
		case filter.GT:
			return cmpInteger(ref, func(v, ref int) bool { return v > ref }), nil
		case filter.LT:
			return cmpInteger(ref, func(v, ref int) bool { return v < ref }), nil
		case filter.GE:
			return cmpInteger(ref, func(v, ref int) bool { return v >= ref }), nil
		case filter.LE:
			return cmpInteger(ref, func(v, ref int) bool { return v <= ref }), nil
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	default:
		panic(fmt.Sprintf("unknown attribute type: %s", typ))
	}
}

func cmpDataTime(ref time.Time, cmp func(v, ref time.Time) bool) func(interface{}) bool {
	return func(i interface{}) bool {
		date, ok := i.(string)
		if !ok {
			return false
		}
		value, err := datetime.Parse(date)
		if err != nil {
			return false
		}
		return cmp(value, ref)
	}
}

func cmpInteger(ref int, cmp func(v, ref int) bool) func(interface{}) bool {
	return func(i interface{}) bool {
		value, ok := toInt(i)
		if !ok {
			return false
		}
		return cmp(value, ref)
	}
}

func cmpDecimal(ref float64, cmp func(v, ref float64) bool) func(interface{}) bool {
	return func(i interface{}) bool {
		value, ok := toFloat(i)
		if !ok {
			return false
		}
		return cmp(value, ref)
	}
}

func cmpBool(ref bool, cmp func(v, ref bool) bool) func(interface{}) bool {
	return func(i interface{}) bool {
		value, ok := i.(bool)
		if !ok {
			return false
		}
		return cmp(value, ref)
	}
}

// The entire operator value must be a substring of the attribute value for a match.
func strComp(ref interface{}, attr schema.CoreAttribute, cmp func(v, ref string) bool) (func(interface{}) bool, error) {
	switch typ := attr.AttributeType(); typ {
	case "boolean":
		ref, ok := ref.(bool)
		if !ok {
			return nil, fmt.Errorf("a boolean attribute needs to be compared to a boolean")
		}
		return func(i interface{}) bool {
			value, ok := i.(bool)
			if !ok {
				return false
			}
			return cmp(fmt.Sprintf("%t", value), fmt.Sprintf("%t", ref))
		}, nil
	case "decimal":
		ref, ok := toFloat(ref)
		if !ok {
			return nil, fmt.Errorf("a decimal attribute needs to be compared to a float/int")
		}
		return func(i interface{}) bool {
			value, ok := toFloat(i)
			if !ok {
				return false
			}
			// fmt.Sprintf("%f") would give them both the same precision.
			return cmp(fmt.Sprint(value), fmt.Sprint(ref))
		}, nil
	case "integer":
		ref, ok := toInt(ref)
		if !ok {
			return nil, fmt.Errorf("a integer attribute needs to be compared to a int")
		}
		return func(i interface{}) bool {
			value, ok := toInt(i)
			if !ok {
				return false
			}
			return cmp(fmt.Sprintf("%d", value), fmt.Sprintf("%d", ref))
		}, nil
	case "binary", "reference", "dateTime", "string":
		ref, ok := ref.(string)
		if !ok {
			return nil, fmt.Errorf("a %s attribute needs to be compared to a string", typ)
		}
		if attr.CaseExact() {
			return func(i interface{}) bool {
				value, ok := i.(string)
				if !ok {
					return false
				}
				return cmp(value, ref)
			}, nil
		}
		return func(i interface{}) bool {
			value, ok := i.(string)
			if !ok {
				return false
			}
			return cmp(strings.ToLower(value), strings.ToLower(ref))
		}, nil
	default:
		panic(fmt.Sprintf("unknown attribute type: %s", typ))
	}
}
