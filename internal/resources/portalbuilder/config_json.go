package portalbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	portalbuilderapi "biot.com/terraform-provider-biot-gen2/internal/api/portalbuilder"
)

// normalizeJSON renders JSON with sorted keys and no whitespace - the same form
// JsonNormalizePlanModifier gives the planned value, so the two compare as strings.
func normalizeJSON(raw string) (string, error) {
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return "", err
	}

	normalized, err := json.Marshal(parsed)
	if err != nil {
		return "", err
	}

	return string(normalized), nil
}

// equivalentJSON reports whether two configurations mean the same to BioT. It ignores nulls and
// empty lists, which BioT adds and drops freely: a hand-written config that leaves out
// "codeSnippetId": null or "cookies": [] is not a change.
func equivalentJSON(a string, b string) bool {
	var parsedA, parsedB any
	if json.Unmarshal([]byte(a), &parsedA) != nil || json.Unmarshal([]byte(b), &parsedB) != nil {
		return false
	}

	return reflect.DeepEqual(stripEmpty(parsedA), stripEmpty(parsedB))
}

func stripEmpty(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		stripped := make(map[string]any, len(typed))
		for key, inner := range typed {
			if inner = stripEmpty(inner); inner != nil {
				stripped[key] = inner
			}
		}
		return stripped
	case []any:
		if len(typed) == 0 {
			return nil
		}
		stripped := make([]any, len(typed))
		for i, inner := range typed {
			stripped[i] = stripEmpty(inner)
		}
		return stripped
	default:
		return value
	}
}

// configJSONFromView renders a view's selected configuration in normalized form.
func configJSONFromView(view portalbuilderapi.ViewResponse) (string, error) {
	encoded, err := json.Marshal(view.SelectedConfig())
	if err != nil {
		return "", err
	}

	return normalizeJSON(string(encoded))
}

// validateConfigJSON checks that config_json is an object made of the sections an update takes,
// so a typo in a section name fails the plan instead of silently clearing that section.
func validateConfigJSON(raw string) error {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return fmt.Errorf("config_json must be a JSON object: %w", err)
	}

	for key := range parsed {
		if !slices.Contains(portalbuilderapi.ConfigKeys, key) {
			return fmt.Errorf("config_json has an unknown section %q; the allowed sections are %s",
				key, strings.Join(portalbuilderapi.ConfigKeys, ", "))
		}
	}

	// Template tabs refer to templates by name, which works in every environment; an id would
	// only work in one, and would show as a difference on every plan.
	var config map[string]any
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return fmt.Errorf("config_json must be a JSON object: %w", err)
	}
	indexes, templates := tabTemplates(config)
	for i, template := range templates {
		if _, hasID := template["id"]; hasID {
			return fmt.Errorf("config_json tabs[%d] refers to its template by id; use the template's name "+
				"instead, for example template = { name = \"patient\" }", indexes[i])
		}
		if name, _ := template["name"].(string); name == "" {
			return fmt.Errorf("config_json tabs[%d] needs the name of its template", indexes[i])
		}
	}

	// Likewise conditional values hold their snippet's code rather than its id.
	return walkObjects(config, "", func(object map[string]any, at string) error {
		if id, hasID := object[snippetIDKey]; hasID && id != nil {
			return fmt.Errorf("config_json %s refers to a code snippet by id; put the snippet's code in %q "+
				"instead, for example %q: \"user._template.name === 'nurse'\"", at, snippetCodeKey, snippetCodeKey)
		}
		if code, hasCode := object[snippetCodeKey]; hasCode {
			if text, _ := code.(string); strings.TrimSpace(text) == "" {
				return fmt.Errorf("config_json %s.%s must be the snippet's code", at, snippetCodeKey)
			}
		}
		return nil
	})
}

// keepEquivalentStateModifier plans the prior state's value when the configured JSON is
// equivalent to it (see equivalentJSON), so formatting, nulls and empty lists never show as a diff.
type keepEquivalentStateModifier struct{}

func (m keepEquivalentStateModifier) Description(ctx context.Context) string {
	return "Keeps the current value when the configured JSON is equivalent to it."
}

func (m keepEquivalentStateModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m keepEquivalentStateModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if equivalentJSON(req.PlanValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
