package condition

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/utils"
)

func MapTerraformConditionToCreateRequest(ctx context.Context, model TerraformAbacCondition) (abacapi.CreateConditionRequest, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	params, tags, ok := mapConditionMutableFields(ctx, model, &diagnostics)
	if !ok {
		return abacapi.CreateConditionRequest{}, diagnostics
	}

	return abacapi.CreateConditionRequest{
		ID:          model.ID.ValueString(),
		Description: utils.StringOrEmpty(model.Description),
		Value:       model.Value.ValueString(),
		Params:      params,
		Tags:        tags,
	}, diagnostics
}

// MapTerraformConditionToUpdateRequest builds the PATCH body. Every mutable field is always
// sent: the service treats an absent field as "leave unchanged", and Terraform always holds
// the complete desired state, so a sparse patch could never clear a value.
//
// id and value are absent because the service refuses to change either on PATCH. The schema
// marks both RequiresReplace, so a change to them never reaches this function.
func MapTerraformConditionToUpdateRequest(ctx context.Context, model TerraformAbacCondition) (abacapi.UpdateConditionRequest, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	params, tags, ok := mapConditionMutableFields(ctx, model, &diagnostics)
	if !ok {
		return abacapi.UpdateConditionRequest{}, diagnostics
	}

	return abacapi.UpdateConditionRequest{
		Description: utils.StringOrEmpty(model.Description),
		Params:      params,
		Tags:        tags,
	}, diagnostics
}

func mapConditionMutableFields(ctx context.Context, model TerraformAbacCondition, diagnostics *diag.Diagnostics) (map[string]interface{}, []string, bool) {
	params, err := utils.JsonStringToMap(model.Params)
	if err != nil {
		diagnostics.AddAttributeError(
			path.Root("params"),
			"Invalid JSON in params",
			"params must be a JSON object. Build it with jsonencode({...}).\n\nParse error: "+err.Error(),
		)
		return nil, nil, false
	}

	tags, tagDiagnostics := utils.SetToStringSlice(ctx, model.Tags)
	diagnostics.Append(tagDiagnostics...)
	if diagnostics.HasError() {
		return nil, nil, false
	}

	return params, tags, true
}
