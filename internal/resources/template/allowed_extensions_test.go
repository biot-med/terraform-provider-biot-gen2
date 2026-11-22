package template

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	templateapi "biot.com/terraform-provider-biot-gen2/internal/api/template"
)

func extensions(values ...string) types.List {
	elements := []attr.Value{}
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elements)
}

func fileAttribute(allowedExtensions types.List) BaseTerraformAttribute {
	return BaseTerraformAttribute{
		Name:       types.StringValue("_sampleFile"),
		Type:       types.StringValue("FILE"),
		Validation: &TerraformValidation{AllowedExtensions: allowedExtensions},
	}
}

func requestJSON(t *testing.T, request any) string {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(body)
}

func TestCustomAttributeSendsAllowedExtensionsInOrder(t *testing.T) {
	attrs := []TerraformCustomAttribute{{BaseTerraformAttribute: fileAttribute(extensions("png", "<none>", "pdf"))}}

	body := requestJSON(t, mapCustomAttributes(context.Background(), attrs))

	expected := `"allowedExtensions":` + requestJSON(t, []string{"png", "<none>", "pdf"})
	if !strings.Contains(body, expected) {
		t.Errorf("expected the list in configured order, got %s", body)
	}
}

func TestCustomAttributeLeavesOutUnsetAllowedExtensions(t *testing.T) {
	for name, list := range map[string]types.List{
		"null":    types.ListNull(types.StringType),
		"unknown": types.ListUnknown(types.StringType),
	} {
		t.Run(name, func(t *testing.T) {
			attrs := []TerraformCustomAttribute{{BaseTerraformAttribute: fileAttribute(list)}}

			body := requestJSON(t, mapCustomAttributes(context.Background(), attrs))

			if strings.Contains(body, "allowedExtensions") {
				t.Errorf("expected no allowedExtensions, got %s", body)
			}
		})
	}
}

func TestBuiltinAttributeNeverSendsAllowedExtensions(t *testing.T) {
	attrs := []TerraformBuiltinAttribute{{BaseTerraformAttribute: fileAttribute(extensions("svg"))}}

	body := requestJSON(t, mapBuiltinAttributes(context.Background(), attrs))

	if strings.Contains(body, "allowedExtensions") {
		t.Errorf("expected no allowedExtensions on a built-in attribute, got %s", body)
	}
}

func TestResponseAllowedExtensions(t *testing.T) {
	cases := map[string]struct {
		json     string
		expected types.List
	}{
		"list keeps order": {`{"allowedExtensions":["pdf","*"]}`, extensions("pdf", "*")},
		"missing is null":  {`{}`, types.ListNull(types.StringType)},
		"null is null":     {`{"allowedExtensions":null}`, types.ListNull(types.StringType)},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var validation templateapi.Validation
			if err := json.Unmarshal([]byte(c.json), &validation); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got := mapToTerraformValidation(context.Background(), &validation).AllowedExtensions

			if !got.Equal(c.expected) {
				t.Errorf("expected %s, got %s", c.expected, got)
			}
		})
	}
}

func allowedExtensionsSchema(t *testing.T, attributes map[string]schema.Attribute) schema.ListAttribute {
	t.Helper()
	validation := attributes["validation"].(schema.SingleNestedAttribute)
	return validation.Attributes["allowed_extensions"].(schema.ListAttribute)
}

func TestAllowedExtensionsIsReadOnlyOnBuiltinAttributes(t *testing.T) {
	builtin := allowedExtensionsSchema(t, builtinAttributeSchema())
	if builtin.Optional || !builtin.Computed {
		t.Errorf("built-in allowed_extensions must be Computed only, got Optional=%v Computed=%v", builtin.Optional, builtin.Computed)
	}

	custom := allowedExtensionsSchema(t, customAttributeSchema())
	if !custom.Optional || !custom.Computed {
		t.Errorf("custom allowed_extensions must be Optional and Computed, got Optional=%v Computed=%v", custom.Optional, custom.Computed)
	}
}
