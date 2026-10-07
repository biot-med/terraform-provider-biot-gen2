package rule

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
)

func point(apiID, executionPoint string) attr.Value {
	return types.ObjectValueMust(apiExecutionPointObjectType.AttrTypes, map[string]attr.Value{
		"api_id":              types.StringValue(apiID),
		"api_execution_point": types.StringValue(executionPoint),
		"order":               types.Int64Value(1),
		"enabled":             types.BoolValue(true),
	})
}

func points(values ...attr.Value) types.Set {
	if values == nil {
		values = []attr.Value{}
	}
	return types.SetValueMust(apiExecutionPointObjectType, values)
}

func summaries(diags diag.Diagnostics) string {
	var out []string
	for _, d := range diags.Errors() {
		out = append(out, d.Summary())
	}
	return strings.Join(out, "; ")
}

func TestValidateAPIExecutionPoints(t *testing.T) {
	cases := []struct {
		name    string
		points  types.Set
		wantErr string // empty: no error expected
	}{
		{"empty list is allowed - BioT stores rules with none", points(), ""},
		{"null", types.SetNull(apiExecutionPointObjectType), ""},
		{"unknown", types.SetUnknown(apiExecutionPointObjectType), ""},
		{"one valid point", points(point("GET/x", "PRE_REQUEST")), ""},
		{"same API at both execution points", points(point("GET/x", "PRE_REQUEST"), point("GET/x", "POST_REQUEST")), ""},
		{"invalid execution point", points(point("GET/x", "DURING")), `Invalid api_execution_point "DURING"`},
		{
			"duplicate api + execution point",
			points(point("GET/x", "PRE_REQUEST"), types.ObjectValueMust(apiExecutionPointObjectType.AttrTypes, map[string]attr.Value{
				"api_id":              types.StringValue("GET/x"),
				"api_execution_point": types.StringValue("PRE_REQUEST"),
				"order":               types.Int64Value(2), // differs only in order
				"enabled":             types.BoolValue(true),
			})),
			"PRE_REQUEST on GET/x is listed more than once",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var diags diag.Diagnostics
			validateAPIExecutionPoints(c.points, &diags)

			got := summaries(diags)
			if c.wantErr == "" && got != "" {
				t.Fatalf("unexpected error: %s", got)
			}
			if c.wantErr != "" && !strings.Contains(got, c.wantErr) {
				t.Fatalf("got %q, want an error containing %q", got, c.wantErr)
			}
		})
	}
}

func TestValidateNewRuleExecutionPoints(t *testing.T) {
	cases := []struct {
		name    string
		points  types.Set
		wantErr bool
	}{
		{"empty list is rejected on create (@NotEmpty on CreateRuleRequest)", points(), true},
		{"one point", points(point("GET/x", "PRE_REQUEST")), false},
		{"null", types.SetNull(apiExecutionPointObjectType), false},
		{"unknown - decided at apply time", types.SetUnknown(apiExecutionPointObjectType), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var diags diag.Diagnostics
			validateNewRuleExecutionPoints(c.points, &diags)

			if diags.HasError() != c.wantErr {
				t.Fatalf("HasError = %v, want %v (%s)", diags.HasError(), c.wantErr, summaries(diags))
			}
			if c.wantErr && !strings.Contains(summaries(diags), "A new rule needs at least one API execution point") {
				t.Fatalf("unexpected message: %s", summaries(diags))
			}
		})
	}
}

// BioT leaves apiExecutionPoints out of the response when a rule has none. State must hold an
// empty set - not null - so that a configured `api_execution_points = []` plans with no diff.
func TestRuleWithoutExecutionPointsMapsToEmptySet(t *testing.T) {
	model, diags := MapRuleResponseToTerraformModel(context.Background(), abacapi.RuleResponse{
		ID:      "BLOCK_MANUFACTURER_ORGANIZATION_USER_RULE",
		Actions: []abacapi.RuleActionRef{{ID: "ACCESS_DENIED_ACTION"}},
		// APIExecutionPoints deliberately absent, as in the service's response.
	})
	if diags.HasError() {
		t.Fatalf("mapping failed: %s", summaries(diags))
	}

	if model.APIExecutionPoints.IsNull() || model.APIExecutionPoints.IsUnknown() {
		t.Fatalf("api_execution_points is null/unknown, want an empty set")
	}
	if n := len(model.APIExecutionPoints.Elements()); n != 0 {
		t.Fatalf("api_execution_points has %d elements, want 0", n)
	}
	if !model.APIExecutionPoints.Equal(points()) {
		t.Fatalf("api_execution_points = %s, want []", model.APIExecutionPoints)
	}
}
