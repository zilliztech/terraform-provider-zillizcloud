package validator

import (
	"context"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// K8sLabelMapValidator validates that map keys/values conform to Kubernetes label rules.
// Rules implemented:
//   - Key may be <prefix>/<name> where prefix is an optional DNS subdomain (<=253 chars total, labels separated by dots),
//     name is required: <=63 chars, alphanumeric at both ends, and can contain dashes, underscores, dots in between.
//   - Value: may be empty; if non-empty, <=63 chars, alphanumeric at both ends, and can contain dashes, underscores, dots in between.
type K8sLabelMapValidator struct{}

func (k K8sLabelMapValidator) Description(_ context.Context) string {
	return "Keys and values must satisfy Kubernetes label syntax."
}

func (k K8sLabelMapValidator) MarkdownDescription(_ context.Context) string {
	return "Keys must be optional DNS subdomain prefix + '/' + name; name <=63 chars, alphanumeric at ends, dashes/underscores/dots allowed in between. Values may be empty; if set, same 63-char and character rules."
}

func (k K8sLabelMapValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Compile regexes once per validate call (file-level var would be nicer, but this is simple and safe)
	// name: must be 1-63, start/end alnum, middle [a-zA-Z0-9_.-]
	nameRe := regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9_.-]{0,61}[A-Za-z0-9])?$`)
	// DNS label: 1-63, start/end alnum, middle [a-z0-9-]
	dnsLabelRe := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

	// value: empty OR same as name rule
	valueValid := func(s string) bool {
		if s == "" {
			return true
		}
		if len(s) > 63 {
			return false
		}
		return nameRe.MatchString(s)
	}

	// prefix: DNS subdomain (series of DNS labels separated by dots), total length <=253
	prefixValid := func(s string) bool {
		if s == "" {
			return true
		}
		if len(s) > 253 {
			return false
		}
		parts := strings.Split(s, ".")
		for _, p := range parts {
			if !dnsLabelRe.MatchString(p) {
				return false
			}
		}
		return true
	}

	elements := req.ConfigValue.Elements()
	for key, val := range elements {
		// Validate key: optional prefix + "/" + name OR just name
		var prefix, name string
		if i := strings.Index(key, "/"); i >= 0 {
			prefix = key[:i]
			name = key[i+1:]
		} else {
			name = key
		}

		if name == "" || len(name) > 63 || !nameRe.MatchString(name) || (strings.Contains(key, "/") && (prefix == "" || !prefixValid(prefix))) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtMapKey(key),
				"Invalid label key",
				"Keys must be of the form '<prefix>/<name>' where prefix is an optional DNS subdomain (<=253 chars) and name is <=63 chars, start/end alphanumeric; middle may contain '-', '_', or '.'.",
			)
		}

		// Validate value is a string and matches rules
		if val.IsNull() || val.IsUnknown() {
			continue
		}
		if val.Type(ctx) != types.StringType {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtMapKey(key),
				"Invalid label value type",
				"Label values must be strings.",
			)
			continue
		}
		// Extract string value directly
		valueStr := ""
		if s, ok := val.(types.String); ok {
			valueStr = s.ValueString()
		} else {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtMapKey(key),
				"Invalid label value type",
				"Label values must be strings.",
			)
			continue
		}
		if !valueValid(valueStr) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtMapKey(key),
				"Invalid label value",
				"Values must be empty or <=63 chars, start/end alphanumeric; middle may contain '-', '_', or '.'.",
			)
		}
	}
}
