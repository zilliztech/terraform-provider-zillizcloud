package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func JsonMapValidator(allowedPrefix string) validator.String {
	return &jsonMapValidatorImpl{
		prefix: allowedPrefix,
	}
}

type jsonMapValidatorImpl struct {
	prefix string
}

func (v *jsonMapValidatorImpl) Description(_ context.Context) string {
	return "Must be a valid JSON object with string keys and JSON-compatible values."
}

func (v *jsonMapValidatorImpl) MarkdownDescription(_ context.Context) string {
	return "Must be a **valid JSON object**, formatted as a `map[string]any`. Values may be strings, numbers, or booleans. " +
		"If provided, keys must start with `" + v.prefix + "`."
}

func (v *jsonMapValidatorImpl) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &parsed)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON format",
			fmt.Sprintf("Failed to parse JSON: %s", err.Error()),
		)
		return
	}

	for k := range parsed {
		if v.prefix != "" && !strings.HasPrefix(k, v.prefix) {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid property key",
				fmt.Sprintf("Key %q does not start with required prefix %q", k, v.prefix),
			)
		}
	}
}

func ParseJsonMap(raw types.String, path path.Path, diags *diag.Diagnostics) map[string]any {
	if raw.IsNull() || raw.IsUnknown() {
		return nil
	}

	var result map[string]any
	err := json.Unmarshal([]byte(raw.ValueString()), &result)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Invalid JSON",
			"Cannot decode map json: "+err.Error(),
		)
		return nil
	}
	return result
}

// ConvertMapToJson converts a map to a JSON string
func ConvertMapToJson(data interface{}) (string, error) {
	if data == nil {
		return "", nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertTFMapToGoMap(tfMap types.Map, attrPath path.Path, diags *diag.Diagnostics) map[string]any {
	result := make(map[string]any)

	if tfMap.IsNull() || tfMap.IsUnknown() {
		return result
	}

	for key, val := range tfMap.Elements() {
		if val.IsNull() || val.IsUnknown() {
			continue
		}

		switch val.Type(context.TODO()) {
		case types.StringType:
			goVal, err := val.ToTerraformValue(context.TODO())
			if err != nil {
				diags.AddError("Failed to convert string value", err.Error())
				continue
			}
			result[key] = goVal.String()

		case types.BoolType:
			boolVal, err := val.ToTerraformValue(context.TODO())
			if err != nil {
				diags.AddError("Failed to convert bool value", err.Error())
				continue
			}
			val, _ := strconv.ParseBool(boolVal.String())
			result[key] = val

		case types.NumberType:
			num, _ := val.ToTerraformValue(context.TODO())
			intVal, err := strconv.ParseInt(num.String(), 10, 64)
			if err != nil {
				floatVal, err := strconv.ParseFloat(num.String(), 64)
				if err != nil {
					diags.AddError("Failed to convert number value", err.Error())
					continue
				}
				result[key] = floatVal
			} else {
				result[key] = intVal
			}

		default:
			diags.AddAttributeError(
				attrPath.AtName(key),
				"Unsupported type in map",
				"Only string, bool, and number types are supported",
			)
		}
	}
	log.Printf("Converted map: %v", result)

	return result
}

func ConvertGoMapToTFMap(data map[string]any) types.Map {
	elements := make(map[string]attr.Value)

	for key, val := range data {
		switch v := val.(type) {
		case string:
			elements[key] = types.StringValue(v)
		case bool:
			elements[key] = types.BoolValue(v)
		case int:
			elements[key] = types.NumberValue(big.NewFloat(float64(v)))
		case int64:
			elements[key] = types.NumberValue(big.NewFloat(float64(v)))
		case float64:
			elements[key] = types.NumberValue(big.NewFloat(v))
		}
	}

	result, _ := types.MapValue(types.StringType, elements)
	return result
}
