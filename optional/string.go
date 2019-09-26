package optional

// New returns an optional string with given value.
func New(value string) String {
	return String{
		value:   value,
		present: true,
	}
}

// String represents an optional string value.
type String struct {
	value   string
	present bool
}

// Value returns the value of the optional string.
func (s String) Value() string {
	return s.value
}

// Present returns whether it contains a value or not.
func (s String) Present() bool {
	return s.present
}
