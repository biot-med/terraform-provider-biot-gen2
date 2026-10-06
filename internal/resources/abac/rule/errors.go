package rule

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	abacapi "biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
	"biot.com/terraform-provider-biot-gen2/internal/resources/abac"
)

// Errors a rule can hit about the objects it references, rather than about itself. They need
// their own handling: ACTIONS_NOT_FOUND here means "you referenced an action that does not
// exist", not "this object is gone", so the shared Entity mapping would describe it wrongly.
const (
	codeActionsNotFound      = "ACTIONS_NOT_FOUND"
	codeConditionsNotFound   = "CONDITIONS_NOT_FOUND"
	codeDuplicateActionIDs   = "DUPLICATE_ACTION_IDS"
	codeDuplicateConditionID = "DUPLICATE_CONDITION_IDS"
)

// addRuleError handles reference errors, then defers to the shared mapping for everything
// about the rule itself.
func addRuleError(diagnostics *diag.Diagnostics, operation string, id string, err error) {
	apiError, ok := transport.AsAPIError(err)
	if !ok {
		abac.AddError(diagnostics, Entity, operation, id, err)
		return
	}

	var details abacapi.ErrorDetails
	_ = apiError.DecodeDetails(&details)

	switch apiError.Code {
	case codeActionsNotFound:
		diagnostics.AddAttributeError(
			path.Root("action_ids"),
			fmt.Sprintf("Rule %q references actions that do not exist", id),
			missingReferenceDetail("action", details.ActionIDs, "action_ids = [biot_abac_action.<name>.id]"),
		)

	case codeConditionsNotFound:
		diagnostics.AddAttributeError(
			path.Root("conditions"),
			fmt.Sprintf("Rule %q references conditions that do not exist", id),
			missingReferenceDetail("condition", details.ConditionIDs, "id = biot_abac_condition.<name>.id"),
		)

	case codeDuplicateActionIDs, codeDuplicateConditionID:
		diagnostics.AddError(fmt.Sprintf("Duplicate references in rule %q", id), apiError.Message)

	default:
		abac.AddError(diagnostics, Entity, operation, id, err)
	}
}

func missingReferenceDetail(noun string, ids []string, example string) string {
	missing := strings.Join(ids, ", ")
	if missing == "" {
		missing = "(not reported by the server)"
	}

	return fmt.Sprintf(`Missing %s ids: %s

If the %ss are managed in this configuration, refer to the resource instead of typing its id, for example:

    %s

That gives Terraform the dependency it needs to create them before the rule. A typed id string gives it none.`, noun, missing, noun, example)
}
