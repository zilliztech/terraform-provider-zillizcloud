package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
