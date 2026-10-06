package rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/resources/abac"
	"biot.com/terraform-provider-biot-gen2/internal/utils"
)

func MapRuleResponseToTerraformModel(ctx context.Context, response abacapi.RuleResponse) (TerraformAbacRule, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	actionIDs := make([]string, 0, len(response.Actions))
	for _, action := range response.Actions {
		actionIDs = append(actionIDs, action.ID)
	}

	actionSet, diags := utils.StringSliceToSet(ctx, actionIDs)
	diagnostics.Append(diags...)

	// A rule with no conditions may come back with the key missing entirely, which decodes to
	// nil. It is mapped to an empty set, never null, to match the schema default of [].
	conditions := make([]TerraformRuleCondition, 0, len(response.Conditions))
	for _, condition := range response.Conditions {
		conditions = append(conditions, TerraformRuleCondition{
			ID:       types.StringValue(condition.ID),
			Inverted: types.BoolValue(condition.Inverted),
		})
	}

	conditionSet, diags := types.SetValueFrom(ctx, conditionObjectType, conditions)
	diagnostics.Append(diags...)

	points := make([]TerraformRuleAPIExecutionPoint, 0, len(response.APIExecutionPoints))
	for _, point := range response.APIExecutionPoints {
		points = append(points, TerraformRuleAPIExecutionPoint{
			APIID:          types.StringValue(point.APIID),
			ExecutionPoint: types.StringValue(point.ExecutionPoint),
			Order:          types.Int64Value(point.Order),
			Enabled:        types.BoolValue(point.Enabled),
		})
	}

	pointSet, diags := types.SetValueFrom(ctx, apiExecutionPointObjectType, points)
	diagnostics.Append(diags...)

	// The server-managed built-in marker is reported through built_in instead of appearing
	// in tags - see abac/tags.go.
	tagSet, diags := utils.StringSliceToSet(ctx, abac.WithoutBuiltInTag(response.Tags))
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return TerraformAbacRule{}, diagnostics
	}

	return TerraformAbacRule{
		ID: types.StringValue(response.ID),
		// Mapped directly rather than through utils.StringOrNull: the service
		// field-initialises description to "", and turning that into null would not match a
		// configured empty description.
		Description:        types.StringValue(response.Description),
		ActionIDs:          actionSet,
		Conditions:         conditionSet,
		APIExecutionPoints: pointSet,
		Tags:               tagSet,
		BuiltIn:            types.BoolValue(abac.HasTag(response.Tags, abac.BuiltInTag)),
	}, diagnostics
}
