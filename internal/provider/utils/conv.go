package utils

import "strconv"

func ToInt(v any) (int, bool) {
	switch v := v.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		i, err := strconv.Atoi(v)
		return i, err == nil
	default:
		return 0, false
	}
}

func ToBool(v any) (bool, bool) {
	switch v := v.(type) {
	case bool:
		return v, true
	case string:
		b, err := strconv.ParseBool(v)
		return b, err == nil
	default:
		return false, false
	}
}
