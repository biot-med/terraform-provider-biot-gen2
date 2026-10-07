package rule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TerraformAbacRule mirrors abacapi.RuleResponse.
//
// The nested collections are held as types.Set rather than Go slices so that ValidateConfig
// can inspect them while some elements are still unknown. The mappers convert them to and from
// the element structs below.
type TerraformAbacRule struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`

	// ActionIDs flattens the API's [{id: ...}] into a plain set of ids. An action reference
	// carries nothing but its id, so the object wrapper would only make the HCL noisier.
	ActionIDs types.Set `tfsdk:"action_ids"`

	Conditions         types.Set `tfsdk:"conditions"`
	APIExecutionPoints types.Set `tfsdk:"api_execution_points"`
	Tags               types.Set `tfsdk:"tags"`

	// BuiltIn is derived from the provenance markers on the API response rather than being
	// a field of its own. See abac/tags.go.
	BuiltIn types.Bool `tfsdk:"built_in"`
}

type TerraformRuleCondition struct {
	ID       types.String `tfsdk:"id"`
	Inverted types.Bool   `tfsdk:"inverted"`
}

type TerraformRuleAPIExecutionPoint struct {
	APIID          types.String `tfsdk:"api_id"`
	ExecutionPoint types.String `tfsdk:"api_execution_point"`
	Order          types.Int64  `tfsdk:"order"`
	Enabled        types.Bool   `tfsdk:"enabled"`
}

var conditionObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":       types.StringType,
		"inverted": types.BoolType,
	},
}

var apiExecutionPointObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"api_id":              types.StringType,
		"api_execution_point": types.StringType,
		"order":               types.Int64Type,
		"enabled":             types.BoolType,
	},
}
