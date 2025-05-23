package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

// SliceToMap
//
// Converts a slice of struct with fields `key` and `value` (tagged as `json:"key"` and `json:"value"`)
// into a map[string]any. It infers the type of the value as bool, int64, float64, or string.
//
// Example:
// Input:
//
//	[]CollectionProperty{
//	  {Key: "collection.ttl.seconds", Value: "86400"},
//	  {Key: "mmap.enabled", Value: "true"},
//	  {Key: "ratio", Value: "3.14"},
//	  {Key: "description", Value: "demo"},
//	}
//
// Output:
//
//	map[string]any{
//	  "collection.ttl.seconds": int64(86400),
//	  "mmap.enabled": true,
//	  "ratio": float64(3.14),
//	  "description": "demo",
//	}
func SliceToMap(input any) (map[string]any, error) {
	v := reflect.ValueOf(input)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input must be a slice")
	}

	result := make(map[string]any)

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)

		// Dereference pointer if needed
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		itemType := item.Type()
		var keyStr, valStr string

		for j := 0; j < item.NumField(); j++ {
			field := itemType.Field(j)
			value := item.Field(j)

			jsonTag := field.Tag.Get("json")
			tagParts := strings.Split(jsonTag, ",")
			tagName := tagParts[0]

			switch tagName {
			case "key":
				keyStr = value.String()
			case "value":
				valStr = value.String()
			}
		}

		if keyStr == "" {
			continue
		}

		result[keyStr] = parseValue(valStr)
	}

	return result, nil
}

// parseValue attempts to parse a string into bool, int64, float64, or fallback to string
func parseValue(s string) any {
	s = strings.TrimSpace(s)
	if s == "true" || s == "false" {
		val, _ := strconv.ParseBool(s)
		return val
	}
	if iVal, err := strconv.ParseInt(s, 10, 64); err == nil {
		return iVal
	}
	if fVal, err := strconv.ParseFloat(s, 64); err == nil {
		return fVal
	}
	return s
}
