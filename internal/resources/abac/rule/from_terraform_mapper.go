package rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/utils"
)

// ruleFields are the parts of a rule that both create and update send.
type ruleFields struct {
	actions            []abacapi.RuleActionRef
	conditions         []abacapi.RuleConditionRef
	apiExecutionPoints []abacapi.RuleAPIExecutionPoint
	tags               []string
}

func MapTerraformRuleToCreateRequest(ctx context.Context, model TerraformAbacRule) (abacapi.CreateRuleRequest, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	fields, ok := mapRuleFields(ctx, model, &diagnostics)
	if !ok {
		return abacapi.CreateRuleRequest{}, diagnostics
	}

	return abacapi.CreateRuleRequest{
		ID:                 model.ID.ValueString(),
		Description:        utils.StringOrEmpty(model.Description),
		Actions:            fields.actions,
		Conditions:         fields.conditions,
		APIExecutionPoints: fields.apiExecutionPoints,
		Tags:               fields.tags,
	}, diagnostics
}

// MapTerraformRuleToUpdateRequest builds the PATCH body. Every field is always sent: the
// service treats an absent field as "leave unchanged", and replaces actions, conditions and
// execution points wholesale when they are present, which is exactly Terraform's model.
//
// tags in particular must always be present - the service reads them unguarded while
// rewriting execution points, and an absent value fails the request.
func MapTerraformRuleToUpdateRequest(ctx context.Context, model TerraformAbacRule) (abacapi.UpdateRuleRequest, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	fields, ok := mapRuleFields(ctx, model, &diagnostics)
	if !ok {
		return abacapi.UpdateRuleRequest{}, diagnostics
	}

	return abacapi.UpdateRuleRequest{
		Description:        utils.StringOrEmpty(model.Description),
		Actions:            fields.actions,
		Conditions:         fields.conditions,
		APIExecutionPoints: fields.apiExecutionPoints,
		Tags:               fields.tags,
	}, diagnostics
}

// mapRuleFields converts the collections. Every slice it returns is non-nil, because a nil
// slice marshals to JSON null, which the service ignores rather than treating as empty.
func mapRuleFields(ctx context.Context, model TerraformAbacRule, diagnostics *diag.Diagnostics) (ruleFields, bool) {
	actionIDs, diags := utils.SetToStringSlice(ctx, model.ActionIDs)
	diagnostics.Append(diags...)

	var conditions []TerraformRuleCondition
	if !model.Conditions.IsNull() && !model.Conditions.IsUnknown() {
		diagnostics.Append(model.Conditions.ElementsAs(ctx, &conditions, false)...)
	}

	var points []TerraformRuleAPIExecutionPoint
	if !model.APIExecutionPoints.IsNull() && !model.APIExecutionPoints.IsUnknown() {
		diagnostics.Append(model.APIExecutionPoints.ElementsAs(ctx, &points, false)...)
	}

	tags, diags := utils.SetToStringSlice(ctx, model.Tags)
	diagnostics.Append(diags...)

	if diagnostics.HasError() {
		return ruleFields{}, false
	}

	fields := ruleFields{
		actions:            make([]abacapi.RuleActionRef, 0, len(actionIDs)),
		conditions:         make([]abacapi.RuleConditionRef, 0, len(conditions)),
		apiExecutionPoints: make([]abacapi.RuleAPIExecutionPoint, 0, len(points)),
		tags:               tags,
	}

	for _, id := range actionIDs {
		fields.actions = append(fields.actions, abacapi.RuleActionRef{ID: id})
	}

	for _, condition := range conditions {
		fields.conditions = append(fields.conditions, abacapi.RuleConditionRef{
			ID:       condition.ID.ValueString(),
			Inverted: condition.Inverted.ValueBool(),
		})
	}

	for _, point := range points {
		fields.apiExecutionPoints = append(fields.apiExecutionPoints, abacapi.RuleAPIExecutionPoint{
			APIID:          point.APIID.ValueString(),
			ExecutionPoint: point.ExecutionPoint.ValueString(),
			Order:          point.Order.ValueInt64(),
			Enabled:        point.Enabled.ValueBool(),
		})
	}

	return fields, true
}
