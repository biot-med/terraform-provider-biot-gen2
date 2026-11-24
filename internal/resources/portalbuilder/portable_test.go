package portalbuilder

import (
	"context"
	"strings"
	"testing"
)

// A resolver whose caches already hold every template, so the tests make no API calls.
func seededResolver() *resolver {
	resolver := newResolver(nil)
	resolver.templateNames["t-patient"], resolver.templateIDs["patient"] = "patient", "t-patient"
	resolver.snippetCode["s-nurse"] = "user._template.name === 'nurse'"
	return resolver
}

func TestTemplateTabsUseNamesInConfigAndIdsInBioT(t *testing.T) {
	ctx := context.Background()
	bioT := `{"tabs":[{"type":"TEMPLATE","template":{"id":"t-patient","title":"Patients"}},{"type":"EXTERNAL_PAGE","externalPage":{"url":"u"}}]}`
	portable := `{"tabs":[{"type":"TEMPLATE","template":{"name":"patient","title":"Patients"}},{"type":"EXTERNAL_PAGE","externalPage":{"url":"u"}}]}`

	got, err := seededResolver().toPortable(ctx, bioT)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := normalizeJSON(portable); got != want {
		t.Errorf("toPortable =\n%s\nwant\n%s", got, want)
	}

	back, err := seededResolver().toBioT(ctx, portable)
	if err != nil {
		t.Fatal(err)
	}
	if !equivalentJSON(back, bioT) {
		t.Errorf("toBioT =\n%s\nwant\n%s", back, bioT)
	}
}

func TestValidateConfigJSONRejectsTemplateIDs(t *testing.T) {
	if err := validateConfigJSON(`{"tabs":[{"type":"TEMPLATE","template":{"id":"t-patient"}}]}`); err == nil {
		t.Error("a template tab with an id was accepted")
	}
	if err := validateConfigJSON(`{"tabs":[{"type":"TEMPLATE","template":{"title":"x"}}]}`); err == nil {
		t.Error("a template tab without a name was accepted")
	}
	if err := validateConfigJSON(`{"tabs":[{"type":"TEMPLATE","template":{"name":"patient"}}]}`); err != nil {
		t.Errorf("a template tab with a name was rejected: %s", err)
	}
}

func TestSnippetIDsBecomeCode(t *testing.T) {
	bioT := `{"attributes":[{"selected":[{"name":"_name","display":{"type":"CODE_SNIPPET","value":null,"codeSnippetId":"s-nurse"}}]}],` +
		`"tabs":[{"type":"EXTERNAL_PAGE","display":{"type":"CONST","value":true,"codeSnippetId":null}}]}`

	got, err := seededResolver().toPortable(context.Background(), bioT)
	if err != nil {
		t.Fatal(err)
	}

	want, _ := normalizeJSON(`{"attributes":[{"selected":[{"name":"_name","display":{"type":"CODE_SNIPPET","value":null,"codeSnippet":"user._template.name === 'nurse'"}}]}],` +
		`"tabs":[{"type":"EXTERNAL_PAGE","display":{"type":"CONST","value":true,"codeSnippetId":null}}]}`)
	if got != want {
		t.Errorf("toPortable =\n%s\nwant\n%s", got, want)
	}
	if err := validateConfigJSON(got); err != nil {
		t.Errorf("the portable form fails validation: %s", err)
	}
}

func TestValidateConfigJSONRejectsSnippetIDs(t *testing.T) {
	err := validateConfigJSON(`{"tabs":[{"type":"EXTERNAL_PAGE","display":{"type":"CODE_SNIPPET","codeSnippetId":"s-nurse"}}]}`)
	if err == nil || !strings.Contains(err.Error(), "tabs[0].display") {
		t.Errorf("error = %v, want one naming tabs[0].display", err)
	}
	if err := validateConfigJSON(`{"tabs":[{"display":{"type":"CODE_SNIPPET","codeSnippet":" "}}]}`); err == nil {
		t.Error("an empty codeSnippet was accepted")
	}
}

// Unchanged code keeps the snippet the view already uses. The seeded resolver has no API client,
// so a test that tried to create a snippet would panic.
func TestUnchangedSnippetCodeIsReused(t *testing.T) {
	resolver := seededResolver()
	resolver.snippetIDs["user._template.name === 'nurse'"] = "s-nurse"

	got, err := resolver.toBioT(context.Background(),
		`{"tabs":[{"type":"EXTERNAL_PAGE","display":{"type":"CODE_SNIPPET","codeSnippet":"user._template.name === 'nurse'"}}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"codeSnippetId":"s-nurse"`) {
		t.Errorf("toBioT = %s, want it to reuse s-nurse", got)
	}
}
