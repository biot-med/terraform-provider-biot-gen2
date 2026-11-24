package portalbuilder

import (
	"encoding/json"
	"strings"
	"testing"

	portalbuilderapi "biot.com/terraform-provider-biot-gen2/internal/api/portalbuilder"
)

func TestEquivalentJSON(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"formatting and key order", `{"tabs":[{"type":"TEMPLATE"}], "attributes":[]}`, `{"attributes":[],"tabs":[{"type":"TEMPLATE"}]}`, true},
		{"nulls BioT adds", `{"tabs":[{"display":{"type":"CONST","value":true}}]}`, `{"tabs":[{"display":{"codeSnippetId":null,"type":"CONST","value":true}}]}`, true},
		{"empty lists BioT adds", `{"tabs":[{"externalPage":{"url":"u"}}]}`, `{"tabs":[{"externalPage":{"url":"u","cookies":[]}}]}`, true},
		{"missing section vs empty", `{}`, `{"tabs":[]}`, true},
		{"order of items matters", `{"tabs":[{"type":"A"},{"type":"B"}]}`, `{"tabs":[{"type":"B"},{"type":"A"}]}`, false},
		{"different value", `{"measurements":{"defaultView":"LIVE"}}`, `{"measurements":{"defaultView":"HISTORY"}}`, false},
		{"invalid JSON", `{`, `{}`, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := equivalentJSON(c.a, c.b); got != c.want {
				t.Errorf("equivalentJSON(%s, %s) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestValidateConfigJSON(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr string
	}{
		{"all sections", `{"attributes":[],"measurements":{},"tabs":[]}`, ""},
		{"empty object", `{}`, ""},
		{"unknown section", `{"tab":[]}`, `unknown section "tab"`},
		{"not an object", `[]`, "must be a JSON object"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateConfigJSON(c.raw)
			if c.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, c.wantErr)
			}
		})
	}
}

// A GET response mixes selected and available items; config_json keeps only the selected ones,
// in the shape the update API takes.
func TestConfigJSONFromView(t *testing.T) {
	response := `{
		"id": "v1", "type": "TEMPLATE_EXPAND", "portalType": "ORGANIZATION_PORTAL",
		"entityTypeName": "patient", "template": {"id": "t1", "name": "patient", "displayName": "Patient"},
		"attributes": [
			{"classification": "REGULAR", "selected": [{"name": "_name"}], "attributes": [{"name": "_name"}, {"name": "_email"}]},
			{"classification": "STATUS", "selected": [], "attributes": [{"name": "_status"}]}
		],
		"measurements": {"attributes": [{"name": "hr"}], "selected": {"views": ["LIVE"], "defaultView": "LIVE", "chart": [], "headline": []}},
		"tabs": {"available": [{"type": "EXTERNAL_PAGE"}], "selected": [{"type": "EXTERNAL_PAGE", "externalPage": {"url": "https://x"}}], "supportedTypes": ["EXTERNAL_PAGE"]}
	}`

	var view portalbuilderapi.ViewResponse
	if err := json.Unmarshal([]byte(response), &view); err != nil {
		t.Fatal(err)
	}

	got, err := configJSONFromView(view)
	if err != nil {
		t.Fatal(err)
	}

	want, _ := normalizeJSON(`{
		"attributes": [{"classification": "REGULAR", "selected": [{"name": "_name"}]}],
		"measurements": {"views": ["LIVE"], "defaultView": "LIVE", "chart": [], "headline": []},
		"tabs": [{"type": "EXTERNAL_PAGE", "externalPage": {"url": "https://x"}}]
	}`)
	if got != want {
		t.Errorf("config_json =\n%s\nwant\n%s", got, want)
	}
}
