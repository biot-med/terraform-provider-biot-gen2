package biotplanmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	validationType = types.ObjectType{AttrTypes: map[string]attr.Type{
		"allowed_extensions": types.ListType{ElemType: types.StringType},
	}}
	attributeType = types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":       types.StringType,
		"id":         types.StringType,
		"type":       types.StringType,
		"validation": validationType,
	}}
)

func list(values ...string) types.List {
	elements := []attr.Value{}
	for _, v := range values {
		elements = append(elements, types.StringValue(v))
	}
	return types.ListValueMust(types.StringType, elements)
}

func attribute(attributeTypeName string, allowedExtensions types.List) attr.Value {
	return types.ObjectValueMust(attributeType.AttrTypes, map[string]attr.Value{
		"name":       types.StringValue("_sampleFile"),
		"id":         types.StringValue("id-1"),
		"type":       types.StringValue(attributeTypeName),
		"validation": types.ObjectValueMust(validationType.AttrTypes, map[string]attr.Value{"allowed_extensions": allowedExtensions}),
	})
}

func plannedAllowedExtensions(t *testing.T, plan, state attr.Value) types.List {
	t.Helper()
	resp := planModifySetResponse(t, plan, state)
	var objs []types.Object
	resp.ElementsAs(context.Background(), &objs, false)
	return objs[0].Attributes()["validation"].(types.Object).Attributes()["allowed_extensions"].(types.List)
}

func planModifySetResponse(t *testing.T, plan, state attr.Value) types.Set {
	t.Helper()
	newSet, diags := copyIDFromStateByNameForSet(context.Background(),
		types.SetValueMust(attributeType, []attr.Value{plan}),
		types.SetValueMust(attributeType, []attr.Value{state}))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	return newSet
}

func TestUnknownAllowedExtensionsAreCopiedFromState(t *testing.T) {
	got := plannedAllowedExtensions(t, attribute("FILE", types.ListUnknown(types.StringType)), attribute("FILE", list("svg", "png")))
	if !got.Equal(list("svg", "png")) {
		t.Errorf("expected the state list, got %s", got)
	}
}

func TestUnknownAllowedExtensionsOfANonFileAttributeStayNull(t *testing.T) {
	got := plannedAllowedExtensions(t, attribute("LABEL", types.ListUnknown(types.StringType)), attribute("LABEL", types.ListNull(types.StringType)))
	if !got.IsNull() {
		t.Errorf("expected null, got %s", got)
	}
}

func TestConfiguredAllowedExtensionsAreKept(t *testing.T) {
	got := plannedAllowedExtensions(t, attribute("FILE", list("pdf")), attribute("FILE", list("*")))
	if !got.Equal(list("pdf")) {
		t.Errorf("expected the configured list, got %s", got)
	}
}

func TestAllowedExtensionsAreNotCopiedWhenTheTypeChanges(t *testing.T) {
	got := plannedAllowedExtensions(t, attribute("LABEL", types.ListUnknown(types.StringType)), attribute("FILE", list("*")))
	if !got.IsUnknown() {
		t.Errorf("expected unknown, got %s", got)
	}
}
