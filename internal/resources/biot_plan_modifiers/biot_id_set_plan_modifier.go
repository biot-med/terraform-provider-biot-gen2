package biotplanmodifiers

import (
	"context"

	"maps"

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

	var planObjs []types.Object
	diags.Append(planSet.ElementsAs(ctx, &planObjs, false)...)
	if diags.HasError() {
		return planSet, diags
	}

	// Build name -> attribute (with ID and nested selectable_values)
	stateAttrMap := map[string]types.Object{}
	if !stateSet.IsNull() && !stateSet.IsUnknown() {
		var stateObjs []types.Object
		diags.Append(stateSet.ElementsAs(ctx, &stateObjs, false)...)
		if diags.HasError() {
			return planSet, diags
		}

		for _, obj := range stateObjs {
			attrMap := obj.Attributes()
			name, ok := attrMap["name"].(types.String)
			if ok && !name.IsNull() && !name.IsUnknown() {
				stateAttrMap[name.ValueString()] = obj
			}
		}
	}

	updatedObjs := make([]attr.Value, 0, len(planObjs))

	for _, planObj := range planObjs {
		attrMap := planObj.Attributes()
		nameVal, _ := attrMap["name"].(types.String)
		idVal, _ := attrMap["id"].(types.String)

		// Copy to new map
		newAttrMap := make(map[string]attr.Value, len(attrMap))
		maps.Copy(newAttrMap, attrMap)

		attrName := nameVal.ValueString()
		stateObj, found := stateAttrMap[attrName]
		if found {
			stateAttrMap := stateObj.Attributes()

			// 1. Copy attribute ID
			if idVal.IsNull() || idVal.IsUnknown() {
				if stateID, ok := stateAttrMap["id"].(types.String); ok && !stateID.IsNull() && !stateID.IsUnknown() {
					newAttrMap["id"] = stateID
				}
			}

			// 2. Copy selectable_values IDs
			planSelectable, planHas := attrMap["selectable_values"].(types.Set)
			stateSelectable, stateHas := stateAttrMap["selectable_values"].(types.Set)

			if planHas && stateHas && !planSelectable.IsNull() && !planSelectable.IsUnknown() {
				var planVals []types.Object
				diags.Append(planSelectable.ElementsAs(ctx, &planVals, false)...)

				var stateVals []types.Object
				diags.Append(stateSelectable.ElementsAs(ctx, &stateVals, false)...)

				// Build name -> id map from state selectable values
				stateValueIDMap := map[string]string{}
				for _, sv := range stateVals {
					svAttrs := sv.Attributes()
					name, _ := svAttrs["name"].(types.String)
					id, _ := svAttrs["id"].(types.String)
					if !name.IsNull() && !name.IsUnknown() && !id.IsNull() && !id.IsUnknown() {
						stateValueIDMap[name.ValueString()] = id.ValueString()
					}
				}

				// Rebuild plan selectable values with copied IDs
				var updatedValues []attr.Value
				for _, sv := range planVals {
					svAttrs := sv.Attributes()
					nameVal, _ := svAttrs["name"].(types.String)
					idVal, _ := svAttrs["id"].(types.String)

					newSvAttrs := make(map[string]attr.Value, len(svAttrs))
					maps.Copy(newSvAttrs, svAttrs)

					if (idVal.IsNull() || idVal.IsUnknown()) && !nameVal.IsNull() && !nameVal.IsUnknown() {
						if id, ok := stateValueIDMap[nameVal.ValueString()]; ok {
							newSvAttrs["id"] = types.StringValue(id)
						}
					}

					svType := sv.Type(ctx).(types.ObjectType)
					newSvObj, svDiags := types.ObjectValue(svType.AttributeTypes(), newSvAttrs)
					diags.Append(svDiags...)
					if svDiags.HasError() {
						return planSet, diags
					}
					updatedValues = append(updatedValues, newSvObj)
				}

				// Set updated selectable_values
				newSelectableSet, setDiags := types.SetValue(planSelectable.ElementType(ctx), updatedValues)
				diags.Append(setDiags...)
				if !setDiags.HasError() {
					newAttrMap["selectable_values"] = newSelectableSet
				}
			}
		}

		objType := planObj.Type(ctx).(types.ObjectType)
		newObj, objDiags := types.ObjectValue(objType.AttributeTypes(), newAttrMap)
		diags.Append(objDiags...)
		if objDiags.HasError() {
			return planSet, diags
		}

		updatedObjs = append(updatedObjs, newObj)
	}

	newSet, setDiags := types.SetValue(planSet.ElementType(ctx), updatedObjs)
	diags.Append(setDiags...)
	return newSet, diags
}
