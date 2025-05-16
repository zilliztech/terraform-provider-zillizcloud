package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
