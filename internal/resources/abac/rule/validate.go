package rule

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The checks below turn server-side rejections into plan-time errors that point at the right
// attribute. Every one of them skips values that are still unknown at validate time - a
// condition id taken from another resource, for example - since those are not errors.

var validExecutionPoints = []string{"PRE_REQUEST", "POST_REQUEST"}

// validateActionIDs: the service requires at least one action (@NotEmpty on create).
func validateActionIDs(actionIDs types.Set, diagnostics *diag.Diagnostics) {
	if actionIDs.IsNull() || actionIDs.IsUnknown() {
		return
	}

	if len(actionIDs.Elements()) == 0 {
		diagnostics.AddAttributeError(
			path.Root("action_ids"),
			"A rule needs at least one action",
			"action_ids is empty. Add the id of at least one biot_abac_action.",
		)
	}
}

// validateConditions: each condition may be referenced once. A set would happily hold
// {id = "A", inverted = true} and {id = "A", inverted = false} as two different elements, and
// the service rejects that with DUPLICATE_CONDITION_IDS.
func validateConditions(conditions types.Set, diagnostics *diag.Diagnostics) {
	if conditions.IsNull() || conditions.IsUnknown() {
		return
	}

	seen := map[string]bool{}
	for _, element := range conditions.Elements() {
		id, ok := knownString(element, "id")
		if !ok {
			continue
		}

		if seen[id] {
			diagnostics.AddAttributeError(
				path.Root("conditions"),
				fmt.Sprintf("Condition %q is listed more than once", id),
				"A rule can reference each condition once. Keep a single entry for it, with the inverted value you want.",
			)
		}
		seen[id] = true
	}
}

// validateAPIExecutionPoints: api_execution_point must be one the service knows, and each
// (api_id, api_execution_point) pair may appear once - the service keys execution points by
// that pair and fails on duplicates.
//
// An empty list is allowed here. BioT stores rules with no execution points - an update may set
// them to [] - and those must import and plan cleanly. Only creating a rule requires one; see
// validateNewRuleExecutionPoints.
func validateAPIExecutionPoints(points types.Set, diagnostics *diag.Diagnostics) {
	if points.IsNull() || points.IsUnknown() {
		return
	}

	seen := map[string]bool{}
	for _, element := range points.Elements() {
		executionPoint, executionPointKnown := knownString(element, "api_execution_point")
		if executionPointKnown && !contains(validExecutionPoints, executionPoint) {
			diagnostics.AddAttributeError(
				path.Root("api_execution_points").AtSetValue(element).AtName("api_execution_point"),
				fmt.Sprintf("Invalid api_execution_point %q", executionPoint),
				fmt.Sprintf("api_execution_point must be one of: %s.", strings.Join(validExecutionPoints, ", ")),
			)
		}

		apiID, apiIDKnown := knownString(element, "api_id")
		if !apiIDKnown || !executionPointKnown {
			continue
		}

		key := apiID + " " + executionPoint
		if seen[key] {
			diagnostics.AddAttributeError(
				path.Root("api_execution_points"),
				fmt.Sprintf("%s on %s is listed more than once", executionPoint, apiID),
				"A rule can run once per API and execution point. Keep a single entry for it.",
			)
		}
		seen[key] = true
	}
}

// knownString reads a string attribute from a set element, reporting false if the element or
// the attribute is null or not yet known.
func knownString(element any, name string) (string, bool) {
	object, ok := element.(types.Object)
	if !ok || object.IsNull() || object.IsUnknown() {
		return "", false
	}

	value, ok := object.Attributes()[name].(types.String)
	if !ok || value.IsNull() || value.IsUnknown() {
		return "", false
	}

	return value.ValueString(), true
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}

	return false
}

// validateNewRuleExecutionPoints: creating a rule requires at least one execution point
// (@NotEmpty on CreateRuleRequest), while updating one may leave it with none. So this runs only
// for a plan that creates the rule - see ModifyPlan - and never in ValidateConfig, which also
// runs for rules that already exist.
func validateNewRuleExecutionPoints(points types.Set, diagnostics *diag.Diagnostics) {
	if points.IsNull() || points.IsUnknown() {
		return
	}

	if len(points.Elements()) == 0 {
		diagnostics.AddAttributeError(
			path.Root("api_execution_points"),
			"A new rule needs at least one API execution point",
			"BioT requires at least one API execution point when a rule is created, and this plan creates the rule - either it is new, or it is being replaced (a change of id, -replace, or replace_triggered_by). Add the API the rule should run on.\n\nRules that already exist in BioT with no execution points are fine with api_execution_points = [] as long as they are updated rather than created.",
		)
	}
}
