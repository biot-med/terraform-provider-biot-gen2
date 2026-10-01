package utils

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// from native to types

func Float64OrNullPtr(n *float64) types.Number {
	if n == nil {
		return types.NumberNull()
	}
	return types.NumberValue(big.NewFloat(*n))
}

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

func Float64OrNilPtr(n types.Number) *float64 {
	if n.IsNull() || n.IsUnknown() {
		return nil
	}
	val, _ := n.ValueBigFloat().Float64()
	return &val
}

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

// set and JSON helpers

// StringSliceToSet converts a native string slice into a Terraform set.
//
// A nil slice becomes an empty set rather than a null one. The access-control service omits
// empty collections from its responses entirely (@JsonInclude(NON_NULL)), so a missing
// "tags" key means "no tags", not "unknown" - and mapping it to null would not match a
// configured value of [].
func StringSliceToSet(ctx context.Context, in []string) (types.Set, diag.Diagnostics) {
	if in == nil {
		in = []string{}
	}

	return types.SetValueFrom(ctx, types.StringType, in)
}

// SetToStringSlice converts a Terraform set into a native string slice.
//
// It never returns nil. A nil slice marshals to JSON null, and the access-control service
// treats a null field as "leave unchanged" rather than "clear", which would make an empty
// tag set impossible to express.
func SetToStringSlice(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	out := []string{}
	if set.IsNull() || set.IsUnknown() {
		return out, nil
	}

	diags := set.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return []string{}, diags
	}

	return out, diags
}

// JsonStringToMap parses a JSON object string into a map. A null, unknown or empty string
// yields an empty map, never nil, for the same reason as SetToStringSlice.
func JsonStringToMap(s types.String) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	if s.IsNull() || s.IsUnknown() || s.ValueString() == "" {
		return out, nil
	}

	if err := json.Unmarshal([]byte(s.ValueString()), &out); err != nil {
		return nil, err
	}

	return out, nil
}

// MapToJsonString renders a map as a compact JSON object string. A nil map becomes "{}" so
// that it matches a schema default of "{}" instead of drifting to null.
func MapToJsonString(m map[string]interface{}) (types.String, error) {
	if m == nil {
		m = map[string]interface{}{}
	}

	bytes, err := json.Marshal(m)
	if err != nil {
		return types.StringNull(), err
	}

	return types.StringValue(string(bytes)), nil
}
