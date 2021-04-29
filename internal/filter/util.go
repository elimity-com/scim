package filter

func toInt(i interface{}) (int, bool) {
	switch i := i.(type) {
	case int:
		return i, true
	case int64:
		return int(i), true
	case int32:
		return int(i), true
	case int16:
		return int(i), true
	case int8:
		return int(i), true
	default:
		return 0, false
	}
}

func toFloat(i interface{}) (float64, bool) {
	switch i := i.(type) {
	case float64:
		return i, true
	case float32:
		return float64(i), true
	default:
		if i, ok := toInt(i); ok {
			return float64(i), true
		}
		return 0, false
	}
}
