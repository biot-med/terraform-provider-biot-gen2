package utils

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// from native to types

func InterfaceToJsonString(ctx context.Context, key string, val interface{}) types.String {
	if val == nil {
		// Return Null string if val is nil
		return types.StringNull()
	}

	// Wrap val inside a map with the given key
	wrapped := map[string]interface{}{
		key: val,
	}

	bytes, err := json.Marshal(wrapped)
	if err != nil {
		// Handle error - return null string or consider returning error instead
		return types.StringNull()
	}

	return types.StringValue(string(bytes))
}

func StringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func StringOrNullPtr(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func Int64OrNullPtr(n *int64) types.Int64 {
	if n == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*n)
}

func BoolOrNullPtr(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}

func ConvertStringList(in []string) []types.String {
	if in == nil {
		return []types.String{}
	}

	out := []types.String{}
	for _, v := range in {
		out = append(out, types.StringValue(v))
	}
	return out
}

// from types to native

func StringOrEmpty(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func StringOrNilPtr(s types.String) *string {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	val := s.ValueString()
	return &val
}

func BoolOrNilPtr(b types.Bool) *bool {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	val := b.ValueBool()
	return &val
}

func Int64OrNilPtr(n types.Int64) *int64 {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	val := n.ValueInt64()
	return &val
}

func ConvertTerraformStringList(in []types.String) []string {
	if in == nil {
		return nil
	}
	out := []string{}
	for _, v := range in {
		if !v.IsNull() && !v.IsUnknown() {
			out = append(out, v.ValueString())
		} else {
			out = append(out, "") // or skip if you prefer
		}
	}
	return out
}
