package condition

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/resources/abac"
	"biot.com/terraform-provider-biot-gen2/internal/utils"
)

func MapConditionResponseToTerraformModel(ctx context.Context, response abacapi.ConditionResponse) (TerraformAbacCondition, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	// A condition with no params comes back without the key at all, which decodes to a nil
	// map. MapToJsonString renders that as "{}" so it matches the schema default.
	params, err := utils.MapToJsonString(response.Params)
	if err != nil {
		diagnostics.AddError(
			"Unreadable params in API response",
			"Failed to render the params of condition "+response.ID+" as JSON: "+err.Error(),
		)
		return TerraformAbacCondition{}, diagnostics
	}

	// The server-managed built-in marker is reported through built_in instead of appearing
	// in tags - see abac/tags.go.
	tags, tagDiagnostics := utils.StringSliceToSet(ctx, abac.WithoutBuiltInTag(response.Tags))
	diagnostics.Append(tagDiagnostics...)
	if diagnostics.HasError() {
		return TerraformAbacCondition{}, diagnostics
	}

	return TerraformAbacCondition{
		ID: types.StringValue(response.ID),
		// Mapped directly rather than through utils.StringOrNull: the service
		// field-initialises description to "", and turning that into null would not match a
		// configured empty description.
		Description: types.StringValue(response.Description),
		Value:       types.StringValue(response.Value),
		Params:      params,
		Tags:        tags,
		BuiltIn:     types.BoolValue(abac.HasTag(response.Tags, abac.BuiltInTag)),
	}, diagnostics
}
