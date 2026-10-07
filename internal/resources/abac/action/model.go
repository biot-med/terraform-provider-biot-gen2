package action

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TerraformAbacAction mirrors api.AbacActionResponse.
//
// Params is a JSON object string rather than a typed map: the service models it as
// Map<String, Object> and real actions nest arrays and objects inside it, which a
// types.Map of strings cannot represent.
type TerraformAbacAction struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Value       types.String `tfsdk:"value"`
	Params      types.String `tfsdk:"params"`
	Tags        types.Set    `tfsdk:"tags"`

	// BuiltIn is derived from the provenance markers on the API response rather than being
	// a field of its own. See abac/tags.go.
	BuiltIn types.Bool `tfsdk:"built_in"`
}
