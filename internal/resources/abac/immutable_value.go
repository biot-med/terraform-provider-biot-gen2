package abac

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RejectValueChange stops a plan that changes `value` on an existing condition or action.
// Call it from ModifyPlan.
//
// The service cannot change which implementation an object uses once it exists - `value` is
// absent from the update request and ignored by the mapper - and the console locks the field
// when editing. The alternative, replacing the object, deletes it first, and the service then
// removes it from every rule that uses it (ON DELETE CASCADE). Refusing the change is both
// what the product does elsewhere and the only option that never loses a reference.
//
// A change of `id` alongside `value` is allowed: that is a new object, and rules that refer to
// it by resource reference are updated to the new id in the same apply.
func RejectValueChange(ctx context.Context, entity Entity, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to compare on create (no state) or destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var stateID, planID, stateValue, planValue types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &stateID)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("id"), &planID)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("value"), &stateValue)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("value"), &planValue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// An unknown value is decided at apply time, when Terraform plans again with it known and
	// this check runs a second time.
	if planValue.IsUnknown() || planID.IsUnknown() {
		return
	}

	if planValue.Equal(stateValue) || !planID.Equal(stateID) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		path.Root("value"),
		fmt.Sprintf("The value of %s %q cannot be changed", entity.Noun, stateID.ValueString()),
		fmt.Sprintf(`BioT cannot change which implementation a %s uses once it exists, so value is fixed at creation - the console locks it when editing for the same reason. This is only about value: params, description and tags can all be changed in place.

To switch from %q to %q, give the %s a new id as well. Terraform will then create a new %s and update the rules that refer to it in the same apply.`,
			entity.Noun, stateValue.ValueString(), planValue.ValueString(), entity.Noun, entity.Noun),
	)
}
