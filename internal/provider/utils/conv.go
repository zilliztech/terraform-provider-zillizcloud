package utils

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
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

// parseValue attempts to parse a string into bool, int64, float64, or fallback to string.
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

// ConvertPropertiesToMap converts a types.Map to a map[string]any for API calls.
func ConvertPropertiesToMap(props types.Map) (map[string]any, error) {
	if props.IsNull() || props.IsUnknown() {
		return nil, nil
	}

	result := make(map[string]any)
	elements := make(map[string]types.String)
	diags := props.ElementsAs(context.Background(), &elements, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to convert properties: %v", diags)
	}

	for k, v := range elements {
		val := v.ValueString()
		// Parse string values to their appropriate types
		if val == "true" {
			result[k] = true
		} else if val == "false" {
			result[k] = false
		} else if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			result[k] = intVal
		} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			// Convert float to int if it's a whole number
			if floatVal == float64(int64(floatVal)) {
				result[k] = int64(floatVal)
			} else {
				result[k] = val
			}
		} else {
			result[k] = val
		}
	}
	return result, nil
}

// ConvertMapToProperties converts a map[string]any to types.Map for state management.
func ConvertMapToProperties(props map[string]any) (types.Map, error) {
	if len(props) == 0 {
		return types.MapNull(types.StringType), nil
	}

	elements := make(map[string]types.String)
	for k, v := range props {
		// Convert all values to strings for storage
		elements[k] = types.StringValue(fmt.Sprintf("%v", v))
	}

	mapVal, diags := types.MapValueFrom(context.Background(), types.StringType, elements)
	if diags.HasError() {
		return types.MapNull(types.StringType), fmt.Errorf("failed to create map: %v", diags)
	}

	return mapVal, nil
}
