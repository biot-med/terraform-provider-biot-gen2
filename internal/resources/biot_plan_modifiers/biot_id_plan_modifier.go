package biotplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CopyIDFromStateByNameSetModifier struct{}

func (m CopyIDFromStateByNameSetModifier) Description(_ context.Context) string {
	return "Copies ID from state to plan based on attribute name."
}

func (m CopyIDFromStateByNameSetModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m CopyIDFromStateByNameSetModifier) PlanModifySet(
	ctx context.Context,
	req planmodifier.SetRequest,
	resp *planmodifier.SetResponse,
) {
	newSet, diags := copyIDFromStateByNameForSet(ctx, req.PlanValue, req.StateValue)
	resp.Diagnostics.Append(diags...)
	if !diags.HasError() {
		resp.PlanValue = newSet
	}
}

func copyIDFromStateByNameForSet(
	ctx context.Context,
	planSet types.Set,
	stateSet types.Set,
) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Extract planned objects
	var planObjs []types.Object
	diags.Append(planSet.ElementsAs(ctx, &planObjs, false)...)
	if diags.HasError() {
		return planSet, diags
	}

	// Build name-to-ID map from state
	nameToID := make(map[string]string)
	if !stateSet.IsNull() && !stateSet.IsUnknown() {
		var stateObjs []types.Object
		diags.Append(stateSet.ElementsAs(ctx, &stateObjs, false)...)
		if diags.HasError() {
			return planSet, diags
		}

		for _, obj := range stateObjs {
			attrMap := obj.Attributes()

			name, okName := attrMap["name"].(types.String)
			id, okID := attrMap["id"].(types.String)

			if okName && okID &&
				!name.IsNull() && !name.IsUnknown() &&
				!id.IsNull() && !id.IsUnknown() {
				nameToID[name.ValueString()] = id.ValueString()
			}
		}
	}

	// Update plan with IDs from state
	updatedObjs := make([]attr.Value, 0, len(planObjs))
	for _, obj := range planObjs {
		// Get attribute map from the object
		attrMap := obj.Attributes()

		nameVal, okName := attrMap["name"].(types.String)
		idVal, okID := attrMap["id"].(types.String)

		// Copy ID from state if name matches and ID is missing
		if okName && !nameVal.IsNull() && !nameVal.IsUnknown() && 
			(!okID || idVal.IsNull() || idVal.IsUnknown()) {

			if idFromState, exists := nameToID[nameVal.ValueString()]; exists {
				attrMap["id"] = types.StringValue(idFromState)
			}
		}

		objType := obj.Type(ctx).(types.ObjectType)
		newObj, objDiags := types.ObjectValue(objType.AttributeTypes(), attrMap)
		diags.Append(objDiags...)
		if objDiags.HasError() {
			return planSet, diags
		}

		updatedObjs = append(updatedObjs, newObj)
	}

	// Rebuild Set
	newSet, setDiags := types.SetValue(planSet.ElementType(ctx), updatedObjs)
	diags.Append(setDiags...)
	return newSet, diags
}