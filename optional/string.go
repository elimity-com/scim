package optional

func NewString(value string) String {
	return String{
		value:   value,
		present: true,
	}
}

type String struct {
	value   string
	present bool
}

func (s String) Value() string {
	return s.value
}

func (s String) Present() bool {
	return s.present
}
