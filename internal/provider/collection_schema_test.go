package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/zilliztech/terraform-provider-zillizcloud/client"
)

func TestConvertSchemaFields(t *testing.T) {
	tests := []struct {
		name     string
		input    []CollectionSchemaFieldModel
		expected []client.CollectionSchemaField
	}{
		{
			name: "Array field with element_data_type",
			input: []CollectionSchemaFieldModel{
				{
					FieldName:       types.StringValue("tags"),
					DataType:        types.StringValue("Array"),
					ElementDataType: types.StringValue("VarChar"),
					IsPrimary:       types.BoolValue(false),
					ElementTypeParams: map[string]types.String{
						"max_length":   types.StringValue("128"),
						"max_capacity": types.StringValue("100"),
					},
				},
			},
			expected: []client.CollectionSchemaField{
				{
					FieldName:       "tags",
					DataType:        "Array",
					ElementDataType: "VarChar",
					IsPrimary:       false,
					ElementTypeParams: map[string]any{
						"max_length":   "128",
						"max_capacity": "100",
					},
				},
			},
		},
		{
			name: "Regular field without element_data_type",
			input: []CollectionSchemaFieldModel{
				{
					FieldName:         types.StringValue("id"),
					DataType:          types.StringValue("Int64"),
					ElementDataType:   types.StringValue(""),
					IsPrimary:         types.BoolValue(true),
					ElementTypeParams: map[string]types.String{},
				},
			},
			expected: []client.CollectionSchemaField{
				{
					FieldName:         "id",
					DataType:          "Int64",
					ElementDataType:   "",
					IsPrimary:         true,
					ElementTypeParams: map[string]any{},
				},
			},
		},
		{
			name: "Vector field",
			input: []CollectionSchemaFieldModel{
				{
					FieldName:       types.StringValue("embedding"),
					DataType:        types.StringValue("FloatVector"),
					ElementDataType: types.StringValue(""),
					IsPrimary:       types.BoolValue(false),
					ElementTypeParams: map[string]types.String{
						"dim": types.StringValue("768"),
					},
				},
			},
			expected: []client.CollectionSchemaField{
				{
					FieldName:       "embedding",
					DataType:        "FloatVector",
					ElementDataType: "",
					IsPrimary:       false,
					ElementTypeParams: map[string]any{
						"dim": "768",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSchemaFields(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d fields, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				actual := result[i]

				if actual.FieldName != expected.FieldName {
					t.Errorf("field %d: expected FieldName %q, got %q", i, expected.FieldName, actual.FieldName)
				}
				if actual.DataType != expected.DataType {
					t.Errorf("field %d: expected DataType %q, got %q", i, expected.DataType, actual.DataType)
				}
				if actual.ElementDataType != expected.ElementDataType {
					t.Errorf("field %d: expected ElementDataType %q, got %q", i, expected.ElementDataType, actual.ElementDataType)
				}
				if actual.IsPrimary != expected.IsPrimary {
					t.Errorf("field %d: expected IsPrimary %v, got %v", i, expected.IsPrimary, actual.IsPrimary)
				}

				if len(actual.ElementTypeParams) != len(expected.ElementTypeParams) {
					t.Errorf("field %d: expected %d params, got %d", i, len(expected.ElementTypeParams), len(actual.ElementTypeParams))
				}

				for k, expectedVal := range expected.ElementTypeParams {
					actualVal, ok := actual.ElementTypeParams[k]
					if !ok {
						t.Errorf("field %d: missing param %q", i, k)
						continue
					}
					if actualVal != expectedVal {
						t.Errorf("field %d: param %q: expected %v, got %v", i, k, expectedVal, actualVal)
					}
				}
			}
		})
	}
}

func TestConvertSchemaFieldModel(t *testing.T) {
	tests := []struct {
		name     string
		input    client.CollectionField
		expected CollectionSchemaFieldModel
	}{
		{
			name: "Array field from API response",
			input: client.CollectionField{
				Name:            "tags",
				Type:            "Array",
				ElementDataType: "VarChar",
				PrimaryKey:      false,
				Params: []client.FieldParam{
					{Key: "max_length", Value: "128"},
					{Key: "max_capacity", Value: "100"},
				},
			},
			expected: CollectionSchemaFieldModel{
				FieldName:       types.StringValue("tags"),
				DataType:        types.StringValue("Array"),
				ElementDataType: types.StringValue("VarChar"),
				IsPrimary:       types.BoolValue(false),
				ElementTypeParams: map[string]types.String{
					"max_length":   types.StringValue("128"),
					"max_capacity": types.StringValue("100"),
				},
			},
		},
		{
			name: "Regular field without element_data_type",
			input: client.CollectionField{
				Name:            "id",
				Type:            "Int64",
				ElementDataType: "",
				PrimaryKey:      true,
				Params:          []client.FieldParam{},
			},
			expected: CollectionSchemaFieldModel{
				FieldName:         types.StringValue("id"),
				DataType:          types.StringValue("Int64"),
				ElementDataType:   types.StringValue(""),
				IsPrimary:         types.BoolValue(true),
				ElementTypeParams: map[string]types.String{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSchemaFieldModel(tt.input)

			if result.FieldName.ValueString() != tt.expected.FieldName.ValueString() {
				t.Errorf("expected FieldName %q, got %q", tt.expected.FieldName.ValueString(), result.FieldName.ValueString())
			}
			if result.DataType.ValueString() != tt.expected.DataType.ValueString() {
				t.Errorf("expected DataType %q, got %q", tt.expected.DataType.ValueString(), result.DataType.ValueString())
			}
			if result.ElementDataType.ValueString() != tt.expected.ElementDataType.ValueString() {
				t.Errorf("expected ElementDataType %q, got %q", tt.expected.ElementDataType.ValueString(), result.ElementDataType.ValueString())
			}
			if result.IsPrimary.ValueBool() != tt.expected.IsPrimary.ValueBool() {
				t.Errorf("expected IsPrimary %v, got %v", tt.expected.IsPrimary.ValueBool(), result.IsPrimary.ValueBool())
			}

			if len(result.ElementTypeParams) != len(tt.expected.ElementTypeParams) {
				t.Errorf("expected %d params, got %d", len(tt.expected.ElementTypeParams), len(result.ElementTypeParams))
			}

			for k, expectedVal := range tt.expected.ElementTypeParams {
				actualVal, ok := result.ElementTypeParams[k]
				if !ok {
					t.Errorf("missing param %q", k)
					continue
				}
				if actualVal.ValueString() != expectedVal.ValueString() {
					t.Errorf("param %q: expected %v, got %v", k, expectedVal.ValueString(), actualVal.ValueString())
				}
			}
		})
	}
}
