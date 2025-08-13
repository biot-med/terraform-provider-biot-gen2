package biotplanmodifiers

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type JsonNormalizePlanModifier struct{}

func (m JsonNormalizePlanModifier) Description(ctx context.Context) string {
	return "Normalizes JSON string values to prevent unnecessary diffs caused by formatting"
}

func (m JsonNormalizePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m JsonNormalizePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		// Nothing to do
		return
	}

	original := req.PlanValue.ValueString()

	var parsed interface{}
	err := json.Unmarshal([]byte(original), &parsed)
	if err != nil {
		// Not valid JSON, leave as is
		return
	}

	normalizedBytes, err := json.Marshal(parsed)
	if err != nil {
		// Failed to marshal, leave as is
		return
	}

	normalized := string(normalizedBytes)

	// If normalized JSON differs, update plan value
	if normalized != original {
		resp.PlanValue = types.StringValue(normalized)
	}
}
