package filter

import (
	"fmt"
	datetime "github.com/di-wu/xsd-datetime"
	"github.com/elimity-com/scim/schema"
	"github.com/scim2/filter-parser/v2"
	"strings"
	"time"
)

func cmpBool(ref bool, cmp func(v, ref bool) error) func(interface{}) error {
	return func(i interface{}) error {
		value, ok := i.(bool)
		if !ok {
			return fmt.Errorf("given value is not a boolean: %v", i)
		}
		return cmp(value, ref)
	}
}

func cmpDataTime(ref time.Time, cmp func(v, ref time.Time) error) func(interface{}) error {
	return func(i interface{}) error {
		date, ok := i.(string)
		if !ok {
			return fmt.Errorf("given value is not a string: %v", i)
		}
		value, err := datetime.Parse(date)
		if err != nil {
			return err
		}
		return cmp(value, ref)
	}
}

func cmpDecimal(ref float64, cmp func(v, ref float64) error) func(interface{}) error {
	return func(i interface{}) error {
		value, ok := toFloat(i)
		if !ok {
			return fmt.Errorf("given value is not a float: %v", i)
		}
		return cmp(value, ref)
	}
}

func cmpInteger(ref int, cmp func(v, ref int) error) func(interface{}) error {
	return func(i interface{}) error {
		value, ok := toInt(i)
		if !ok {
			return fmt.Errorf("given value is not an integer: %v", i)
		}
		return cmp(value, ref)
	}
}

// createCompareFunction returns a compare function based on the attribute expression and attribute.
// e.g. `userName eq "john"` will return a string comparator that checks whether the passed value is equal to "john".
func createCompareFunction(e *filter.AttributeExpression, attr schema.CoreAttribute) (func(interface{}) error, error) {
	switch typ := attr.AttributeType(); typ {
	case "binary":
		ref, ok := e.CompareValue.(string)
		if !ok {
			return nil, fmt.Errorf("a binary attribute needs to be compared to a string")
		}
		switch e.Operator {
		case filter.EQ:
			return strComp(ref, attr, func(v, ref string) error {
				if v != ref {
					return fmt.Errorf("%s is not equal to %s", v, ref)
				}
				return nil
			})
		case filter.NE:
			return strComp(ref, attr, func(v, ref string) error {
				if v == ref {
					return fmt.Errorf("%s is equal to %s", v, ref)
				}
				return nil
			})
		case filter.CO:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.Contains(v, ref) {
					return fmt.Errorf("%s does not contain %s", v, ref)
				}
				return nil
			})
		case filter.SW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasPrefix(v, ref) {
					return fmt.Errorf("%s does not start with %s", v, ref)
				}
				return nil
			})
		case filter.EW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasSuffix(v, ref) {
					return fmt.Errorf("%s does not end with %s", v, ref)
				}
				return nil
			})
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
			return cmpDataTime(ref, func(v, ref time.Time) error {
				if !v.Equal(ref) {
					return fmt.Errorf("%s is not equal to %s", v.Format(time.RFC3339), ref.Format(time.RFC3339))
				}
				return nil
			}), nil
		case filter.NE:
			return cmpDataTime(ref, func(v, ref time.Time) error {
				if v.Equal(ref) {
					return fmt.Errorf("%s is equal to %s", v.Format(time.RFC3339), ref.Format(time.RFC3339))
				}
				return nil
			}), nil
		case filter.CO:
			return strComp(date, attr, func(v, ref string) error {
				if !strings.Contains(v, ref) {
					return fmt.Errorf("%s does not contain %s", v, ref)
				}
				return nil
			})
		case filter.SW:
			return strComp(date, attr, func(v, ref string) error {
				if !strings.HasPrefix(v, ref) {
					return fmt.Errorf("%s does not start with %s", v, ref)
				}
				return nil
			})
		case filter.EW:
			return strComp(date, attr, func(v, ref string) error {
				if !strings.HasSuffix(v, ref) {
					return fmt.Errorf("%s does not end with %s", v, ref)
				}
				return nil
			})
		case filter.GT:
			return cmpDataTime(ref, func(v, ref time.Time) error {
				if !v.After(ref) {
					return fmt.Errorf("%s is not greater than %s", v.Format(time.RFC3339), ref.Format(time.RFC3339))
				}
				return nil
			}), nil
		case filter.LT:
			return cmpDataTime(ref, func(v, ref time.Time) error {
				if !v.Before(ref) {
					return fmt.Errorf("%s is not less than %s", v.Format(time.RFC3339), ref.Format(time.RFC3339))
				}
				return nil
			}), nil
		case filter.GE:
			return cmpDataTime(ref, func(v, ref time.Time) error {
				if !v.After(ref) && !v.Equal(ref) {
					return fmt.Errorf("%s is not greater or equal to %s", v.Format(time.RFC3339), ref.Format(time.RFC3339))
				}
				return nil
			}), nil
		case filter.LE:
			return cmpDataTime(ref, func(v, ref time.Time) error {
				if !v.Before(ref) && !v.Equal(ref) {
					return fmt.Errorf("%s is not less or equal to %s", v.Format(time.RFC3339), ref.Format(time.RFC3339))
				}
				return nil
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
			return strComp(ref, attr, func(v, ref string) error {
				if v != ref {
					return fmt.Errorf("%s is not equal to %s", v, ref)
				}
				return nil
			})
		case filter.NE:
			return strComp(ref, attr, func(v, ref string) error {
				if v == ref {
					return fmt.Errorf("%s is equal to %s", v, ref)
				}
				return nil
			})
		case filter.CO:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.Contains(v, ref) {
					return fmt.Errorf("%s does not contain %s", v, ref)
				}
				return nil
			})
		case filter.SW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasPrefix(v, ref) {
					return fmt.Errorf("%s does not start with %s", v, ref)
				}
				return nil
			})
		case filter.EW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasSuffix(v, ref) {
					return fmt.Errorf("%s does not end with %s", v, ref)
				}
				return nil
			})
		case filter.GT:
			return strComp(ref, attr, func(v, ref string) error {
				if strings.Compare(v, ref) <= 0 {
					return fmt.Errorf("%s is not lexicographically greater than %s", v, ref)
				}
				return nil
			})
		case filter.LT:
			return strComp(ref, attr, func(v, ref string) error {
				if strings.Compare(v, ref) >= 0 {
					return fmt.Errorf("%s is not lexicographically less than %s", v, ref)
				}
				return nil
			})
		case filter.GE:
			return strComp(ref, attr, func(v, ref string) error {
				if strings.Compare(v, ref) < 0 {
					return fmt.Errorf("%s is not lexicographically greater or equal to %s", v, ref)
				}
				return nil
			})
		case filter.LE:
			return strComp(ref, attr, func(v, ref string) error {
				if strings.Compare(v, ref) > 0 {
					return fmt.Errorf("%s is not lexicographically less or equal to %s", v, ref)
				}
				return nil
			})
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
			return cmpBool(ref, func(v, ref bool) error {
				if v != ref {
					return fmt.Errorf("%t is not equal to %t", v, ref)
				}
				return nil
			}), nil
		case filter.NE:
			return cmpBool(ref, func(v, ref bool) error {
				if v == ref {
					return fmt.Errorf("%t is equal to %t", v, ref)
				}
				return nil
			}), nil
		case filter.CO:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.Contains(v, ref) {
					return fmt.Errorf("%s does not contain %s", v, ref)
				}
				return nil
			})
		case filter.SW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasPrefix(v, ref) {
					return fmt.Errorf("%s does not start with %s", v, ref)
				}
				return nil
			})
		case filter.EW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasSuffix(v, ref) {
					return fmt.Errorf("%s does not end with %s", v, ref)
				}
				return nil
			})
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
			return cmpDecimal(ref, func(v, ref float64) error {
				if v != ref {
					return fmt.Errorf("%f is not equal to %f", v, ref)
				}
				return nil
			}), nil
		case filter.NE:
			return cmpDecimal(ref, func(v, ref float64) error {
				if v == ref {
					return fmt.Errorf("%f is equal to %f", v, ref)
				}
				return nil
			}), nil
		case filter.CO:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.Contains(v, ref) {
					return fmt.Errorf("%s does not contain %s", v, ref)
				}
				return nil
			})
		case filter.SW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasPrefix(v, ref) {
					return fmt.Errorf("%s does not start with %s", v, ref)
				}
				return nil
			})
		case filter.EW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasSuffix(v, ref) {
					return fmt.Errorf("%s does not end with %s", v, ref)
				}
				return nil
			})
		case filter.GT:
			return cmpDecimal(ref, func(v, ref float64) error {
				if v <= ref {
					return fmt.Errorf("%f is not greater than %f", v, ref)
				}
				return nil
			}), nil
		case filter.LT:
			return cmpDecimal(ref, func(v, ref float64) error {
				if v >= ref {
					return fmt.Errorf("%f is not less than %f", v, ref)
				}
				return nil
			}), nil
		case filter.GE:
			return cmpDecimal(ref, func(v, ref float64) error {
				if v < ref {
					return fmt.Errorf("%f is not greater or equal to %f", v, ref)
				}
				return nil
			}), nil
		case filter.LE:
			return cmpDecimal(ref, func(v, ref float64) error {
				if v > ref {
					return fmt.Errorf("%f is not less or equal to %f", v, ref)
				}
				return nil
			}), nil
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
			return cmpInteger(ref, func(v, ref int) error {
				if v != ref {
					return fmt.Errorf("%d is not equal to %d", v, ref)
				}
				return nil
			}), nil
		case filter.NE:
			return cmpInteger(ref, func(v, ref int) error {
				if v == ref {
					return fmt.Errorf("%d is equal to %d", v, ref)
				}
				return nil
			}), nil
		case filter.CO:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.Contains(v, ref) {
					return fmt.Errorf("%s does not contain %s", v, ref)
				}
				return nil
			})
		case filter.SW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasPrefix(v, ref) {
					return fmt.Errorf("%s does not start with %s", v, ref)
				}
				return nil
			})
		case filter.EW:
			return strComp(ref, attr, func(v, ref string) error {
				if !strings.HasSuffix(v, ref) {
					return fmt.Errorf("%s does not end with %s", v, ref)
				}
				return nil
			})
		case filter.GT:
			return cmpInteger(ref, func(v, ref int) error {
				if v <= ref {
					return fmt.Errorf("%d is not greater than %d", v, ref)
				}
				return nil
			}), nil
		case filter.LT:
			return cmpInteger(ref, func(v, ref int) error {
				if v >= ref {
					return fmt.Errorf("%d is not less than %d", v, ref)
				}
				return nil
			}), nil
		case filter.GE:
			return cmpInteger(ref, func(v, ref int) error {
				if v < ref {
					return fmt.Errorf("%d is not greater or equal to %d", v, ref)
				}
				return nil
			}), nil
		case filter.LE:
			return cmpInteger(ref, func(v, ref int) error {
				if v > ref {
					return fmt.Errorf("%d is not less or equal to %d", v, ref)
				}
				return nil
			}), nil
		default:
			return nil, fmt.Errorf("unknown operator in expression: %s", e)
		}
	default:
		panic(fmt.Sprintf("unknown attribute type: %s", typ))
	}
}

// The entire operator value must be a substring of the attribute value for a match.
func strComp(ref interface{}, attr schema.CoreAttribute, cmp func(v, ref string) error) (func(interface{}) error, error) {
	switch typ := attr.AttributeType(); typ {
	case "boolean":
		ref, ok := ref.(bool)
		if !ok {
			return nil, fmt.Errorf("a boolean attribute needs to be compared to a boolean")
		}
		return func(i interface{}) error {
			value, ok := i.(bool)
			if !ok {
				return fmt.Errorf("given value is not a boolean: %v", i)
			}
			return cmp(fmt.Sprintf("%t", value), fmt.Sprintf("%t", ref))
		}, nil
	case "decimal":
		ref, ok := toFloat(ref)
		if !ok {
			return nil, fmt.Errorf("a decimal attribute needs to be compared to a float/int")
		}
		return func(i interface{}) error {
			value, ok := toFloat(i)
			if !ok {
				return fmt.Errorf("given value is not a float: %v", i)
			}
			// fmt.Sprintf("%f") would give them both the same precision.
			return cmp(fmt.Sprint(value), fmt.Sprint(ref))
		}, nil
	case "integer":
		ref, ok := toInt(ref)
		if !ok {
			return nil, fmt.Errorf("a integer attribute needs to be compared to a int")
		}
		return func(i interface{}) error {
			value, ok := toInt(i)
			if !ok {
				return fmt.Errorf("given value is not an integer: %v", i)
			}
			return cmp(fmt.Sprintf("%d", value), fmt.Sprintf("%d", ref))
		}, nil
	case "binary", "reference", "dateTime", "string":
		ref, ok := ref.(string)
		if !ok {
			return nil, fmt.Errorf("a %s attribute needs to be compared to a string", typ)
		}
		if attr.CaseExact() {
			return func(i interface{}) error {
				value, ok := i.(string)
				if !ok {
					return fmt.Errorf("given value is not a string: %v", i)
				}
				return cmp(value, ref)
			}, nil
		}
		return func(i interface{}) error {
			value, ok := i.(string)
			if !ok {
				return fmt.Errorf("given value is not a string: %v", i)
			}
			return cmp(strings.ToLower(value), strings.ToLower(ref))
		}, nil
	default:
		panic(fmt.Sprintf("unknown attribute type: %s", typ))
	}
}
