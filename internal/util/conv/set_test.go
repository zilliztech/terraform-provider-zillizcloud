package conv

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSliceToSet_String(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected types.Set
	}{
		{
			name:     "nil slice returns null set",
			input:    nil,
			expected: types.SetNull(types.StringType),
		},
		{
			name:     "empty slice returns empty set",
			input:    []string{},
			expected: types.SetValueMust(types.StringType, []attr.Value{}),
		},
		{
			name:  "single element",
			input: []string{"hello"},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("hello"),
			}),
		},
		{
			name:  "multiple elements",
			input: []string{"a", "b", "c"},
			expected: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceToSet(tt.input)
			assert.True(t, result.Equal(tt.expected), "Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestSliceToSet_Int64(t *testing.T) {
	tests := []struct {
		name     string
		input    []int64
		expected types.Set
	}{
		{
			name:     "nil slice returns null set",
			input:    nil,
			expected: types.SetNull(types.Int64Type),
		},
		{
			name:     "empty slice returns empty set",
			input:    []int64{},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{}),
		},
		{
			name:  "single element",
			input: []int64{42},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
		},
		{
			name:  "multiple elements",
			input: []int64{1, 2, 3},
			expected: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(2),
				types.Int64Value(3),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceToSet(tt.input)
			assert.True(t, result.Equal(tt.expected), "Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestSliceToSet_Bool(t *testing.T) {
	tests := []struct {
		name     string
		input    []bool
		expected types.Set
	}{
		{
			name:     "nil slice returns null set",
			input:    nil,
			expected: types.SetNull(types.BoolType),
		},
		{
			name:     "empty slice returns empty set",
			input:    []bool{},
			expected: types.SetValueMust(types.BoolType, []attr.Value{}),
		},
		{
			name:  "single element",
			input: []bool{true},
			expected: types.SetValueMust(types.BoolType, []attr.Value{
				types.BoolValue(true),
			}),
		},
		{
			name:  "multiple elements",
			input: []bool{true, false, true},
			expected: types.SetValueMust(types.BoolType, []attr.Value{
				types.BoolValue(true),
				types.BoolValue(false),
				types.BoolValue(true),
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceToSet(tt.input)
			assert.True(t, result.Equal(tt.expected), "Expected %v, got %v", tt.expected, result)
		})
	}
}

func TestSetToSlice_String(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       types.Set
		expected    []string
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "null set returns error",
			input:       types.SetNull(types.StringType),
			expected:    nil,
			expectError: true,
			errorCheck:  IsNullError,
		},
		{
			name:        "unknown set returns empty slice",
			input:       types.SetUnknown(types.StringType),
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "empty set returns empty slice",
			input:       types.SetValueMust(types.StringType, []attr.Value{}),
			expected:    []string{},
			expectError: false,
		},
		{
			name: "single element",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("hello"),
			}),
			expected:    []string{"hello"},
			expectError: false,
		},
		{
			name: "multiple elements",
			input: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
			expected:    []string{"a", "b", "c"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetToSlice[string](ctx, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err), "Error check failed for: %v", err)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestSetToSlice_Int64(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       types.Set
		expected    []int64
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "null set returns error",
			input:       types.SetNull(types.Int64Type),
			expected:    nil,
			expectError: true,
			errorCheck:  IsNullError,
		},
		{
			name:        "unknown set returns empty slice",
			input:       types.SetUnknown(types.Int64Type),
			expected:    []int64{},
			expectError: false,
		},
		{
			name:        "empty set returns empty slice",
			input:       types.SetValueMust(types.Int64Type, []attr.Value{}),
			expected:    []int64{},
			expectError: false,
		},
		{
			name: "single element",
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(42),
			}),
			expected:    []int64{42},
			expectError: false,
		},
		{
			name: "multiple elements",
			input: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(2),
				types.Int64Value(3),
			}),
			expected:    []int64{1, 2, 3},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetToSlice[int64](ctx, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err), "Error check failed for: %v", err)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestSetToSlice_Bool(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		input       types.Set
		expected    []bool
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:        "null set returns error",
			input:       types.SetNull(types.BoolType),
			expected:    nil,
			expectError: true,
			errorCheck:  IsNullError,
		},
		{
			name:        "unknown set returns empty slice",
			input:       types.SetUnknown(types.BoolType),
			expected:    []bool{},
			expectError: false,
		},
		{
			name:        "empty set returns empty slice",
			input:       types.SetValueMust(types.BoolType, []attr.Value{}),
			expected:    []bool{},
			expectError: false,
		},
		{
			name: "single element",
			input: types.SetValueMust(types.BoolType, []attr.Value{
				types.BoolValue(true),
			}),
			expected:    []bool{true},
			expectError: false,
		},
		{
			name: "multiple elements",
			input: types.SetValueMust(types.BoolType, []attr.Value{
				types.BoolValue(true),
				types.BoolValue(false),
				types.BoolValue(true),
			}),
			expected:    []bool{true, false, true},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetToSlice[bool](ctx, tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					assert.True(t, tt.errorCheck(err), "Error check failed for: %v", err)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	ctx := context.Background()

	// Test string round trip
	t.Run("string round trip", func(t *testing.T) {
		original := []string{"hello", "world", "test"}
		set := SliceToSet(original)
		result, err := SetToSlice[string](ctx, set)
		require.NoError(t, err)
		assert.ElementsMatch(t, original, result)
	})

	// Test int64 round trip
	t.Run("int64 round trip", func(t *testing.T) {
		original := []int64{1, 2, 3, 42}
		set := SliceToSet(original)
		result, err := SetToSlice[int64](ctx, set)
		require.NoError(t, err)
		assert.ElementsMatch(t, original, result)
	})

	// Test bool round trip
	t.Run("bool round trip", func(t *testing.T) {
		original := []bool{true, false, true, false}
		set := SliceToSet(original)
		result, err := SetToSlice[bool](ctx, set)
		require.NoError(t, err)
		assert.ElementsMatch(t, original, result)
	})

	// Test nil handling
	t.Run("nil slice round trip should fail", func(t *testing.T) {
		var original []string
		set := SliceToSet(original) // nil slice -> null set
		result, err := SetToSlice[string](ctx, set)
		require.Error(t, err)
		assert.True(t, IsNullError(err))
		assert.Nil(t, result)
	})
}

func TestConvenienceFunctions(t *testing.T) {
	ctx := context.Background()

	// Test StringSliceToSet and SetToStringSlice
	t.Run("string convenience functions", func(t *testing.T) {
		original := []string{"a", "b", "c"}
		set := SliceToSet(original)
		result, err := SetToSlice[string](ctx, set)
		require.NoError(t, err)
		assert.ElementsMatch(t, original, result)
	})

	// Test Int64SliceToSet and SetToInt64Slice
	t.Run("int64 convenience functions", func(t *testing.T) {
		original := []int64{1, 2, 3}
		set := SliceToSet(original)
		result, err := SetToSlice[int64](ctx, set)
		require.NoError(t, err)
		assert.ElementsMatch(t, original, result)
	})

	// Test BoolSliceToSet and SetToBoolSlice
	t.Run("bool convenience functions", func(t *testing.T) {
		original := []bool{true, false}
		set := SliceToSet(original)
		result, err := SetToSlice[bool](ctx, set)
		require.NoError(t, err)
		assert.ElementsMatch(t, original, result)
	})
}
