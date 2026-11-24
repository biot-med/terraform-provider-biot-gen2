package portalbuilder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"biot.com/terraform-provider-biot-gen2/internal/api"
	"biot.com/terraform-provider-biot-gen2/internal/api/codesnippet"
	portalbuilderapi "biot.com/terraform-provider-biot-gen2/internal/api/portalbuilder"
)

// A view's configuration refers to templates and code snippets by id, and ids differ between
// environments. So config_json holds a portable form instead:
//
//   - a template tab names its template:      "template": {"name": "patient"}
//   - a conditional value holds its code:     "codeSnippet": "user._template.name === 'nurse'"
//
// resolver translates between that form and BioT's, caching lookups for one operation.
type resolver struct {
	client *api.APIClient

	templateNames map[string]string // template id -> name
	templateIDs   map[string]string // template name -> id
	snippetCode   map[string]string // snippet id -> code
	snippetIDs    map[string]string // code -> id of a snippet the view already uses
}

func newResolver(client *api.APIClient) *resolver {
	return &resolver{
		client:        client,
		templateNames: map[string]string{},
		templateIDs:   map[string]string{},
		snippetCode:   map[string]string{},
		snippetIDs:    map[string]string{},
	}
}

const (
	snippetIDKey   = "codeSnippetId"
	snippetCodeKey = "codeSnippet"
)

// errTemplateNotFound is returned for a template name that does not exist in this environment.
var errTemplateNotFound = errors.New("template not found")

// remember records the templates of listed views, which come with their names, saving lookups.
func (r *resolver) remember(views []portalbuilderapi.View) {
	for _, view := range views {
		if view.Template != nil {
			r.templateNames[view.Template.ID] = view.Template.Name
			r.templateIDs[view.Template.Name] = view.Template.ID
		}
	}
}

/* Templates */

func (r *resolver) templateName(ctx context.Context, id string) (string, error) {
	if name, ok := r.templateNames[id]; ok {
		return name, nil
	}

	template, err := r.client.Template.Get(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to look up template %s: %w", id, err)
	}

	r.templateNames[id], r.templateIDs[template.Name] = template.Name, id
	return template.Name, nil
}

func (r *resolver) templateID(ctx context.Context, name string) (string, error) {
	if id, ok := r.templateIDs[name]; ok {
		return id, nil
	}

	response, err := r.client.Template.Search(ctx, map[string]interface{}{
		"filter": map[string]interface{}{
			"name": map[string]interface{}{"in": []string{name}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to look up template %q: %w", name, err)
	}

	switch len(response.Data) {
	case 0:
		return "", fmt.Errorf("%w: %q", errTemplateNotFound, name)
	case 1:
		id := response.Data[0].ID
		r.templateNames[id], r.templateIDs[name] = name, id
		return id, nil
	default:
		var entities []string
		for _, template := range response.Data {
			entities = append(entities, template.EntityTypeName)
		}
		return "", fmt.Errorf("more than one template is named %q (entities: %s)", name, strings.Join(entities, ", "))
	}
}

// templateNotFoundError explains a missing template in terms of what to do about it.
func templateNotFoundError(name string, where string) string {
	return fmt.Sprintf("%s refers to template %q, which does not exist in this environment. BioT creates a "+
		"template's views only once the template exists, so apply the template first: manage it with "+
		"biot_template and add it to this resource's depends_on, or apply the templates in an earlier run.",
		where, name)
}

// tabTemplates returns the template objects of the template tabs in a parsed config, in order,
// with the index of the tab each belongs to.
func tabTemplates(config map[string]any) (indexes []int, templates []map[string]any) {
	tabs, _ := config["tabs"].([]any)
	for i, tab := range tabs {
		tabObject, _ := tab.(map[string]any)
		if template, ok := tabObject["template"].(map[string]any); ok {
			indexes = append(indexes, i)
			templates = append(templates, template)
		}
	}
	return indexes, templates
}

/* Code snippets */

func (r *resolver) snippetCodeOf(ctx context.Context, id string) (string, error) {
	if code, ok := r.snippetCode[id]; ok {
		return code, nil
	}

	snippet, err := r.client.CodeSnippet.Get(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to look up code snippet %s: %w", id, err)
	}

	r.snippetCode[id] = snippet.Code
	return snippet.Code, nil
}

// rememberSnippets records the snippets a view uses now, so that toBioT reuses them for code that
// has not changed instead of creating a new snippet on every apply.
func (r *resolver) rememberSnippets(ctx context.Context, view portalbuilderapi.ViewResponse) error {
	encoded, err := json.Marshal(view.SelectedConfig())
	if err != nil {
		return err
	}

	var config any
	if err := json.Unmarshal(encoded, &config); err != nil {
		return err
	}

	return walkObjects(config, "", func(object map[string]any, at string) error {
		id, ok := object[snippetIDKey].(string)
		if !ok {
			return nil
		}
		code, err := r.snippetCodeOf(ctx, id)
		if err != nil {
			return err
		}
		r.snippetIDs[code] = id
		return nil
	})
}

// snippetFor returns the id of a snippet holding code: one the view already uses, or a new one.
func (r *resolver) snippetFor(ctx context.Context, code string) (string, error) {
	if id, ok := r.snippetIDs[code]; ok {
		return id, nil
	}

	id, err := r.createSnippet(ctx, code)
	if err != nil {
		return "", err
	}

	// The same code in several places of config_json shares one snippet.
	r.snippetIDs[code] = id
	return id, nil
}

// createSnippet creates a snippet for code and returns its id. It is unnamed: BioT requires names
// to be unique, and the snippet only exists to be referred to by this view.
func (r *resolver) createSnippet(ctx context.Context, code string) (string, error) {
	snippet, err := r.client.CodeSnippet.Create(ctx, codesnippet.CreateRequest{
		Language: codesnippet.LanguageJavaScript,
		Code:     code,
	})
	if err != nil {
		return "", err
	}

	r.snippetCode[snippet.ID] = code
	return snippet.ID, nil
}

// walkObjects calls fn for every JSON object in value, depth first, with a readable path to it
// such as tabs[0].display. Map keys are visited in sorted order so errors are deterministic.
func walkObjects(value any, at string, fn func(object map[string]any, at string) error) error {
	switch typed := value.(type) {
	case map[string]any:
		if err := fn(typed, at); err != nil {
			return err
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			inner := key
			if at != "" {
				inner = at + "." + key
			}
			if err := walkObjects(typed[key], inner, fn); err != nil {
				return err
			}
		}
	case []any:
		for i, inner := range typed {
			if err := walkObjects(inner, fmt.Sprintf("%s[%d]", at, i), fn); err != nil {
				return err
			}
		}
	}
	return nil
}

/* Translation */

// toPortable turns a config read from BioT into the form config_json holds.
func (r *resolver) toPortable(ctx context.Context, raw string) (string, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return "", err
	}

	_, templates := tabTemplates(config)
	for _, template := range templates {
		id, ok := template["id"].(string)
		if !ok {
			continue
		}
		name, err := r.templateName(ctx, id)
		if err != nil {
			return "", err
		}
		delete(template, "id")
		template["name"] = name
	}

	err := walkObjects(config, "", func(object map[string]any, at string) error {
		id, ok := object[snippetIDKey].(string)
		if !ok {
			return nil
		}
		code, err := r.snippetCodeOf(ctx, id)
		if err != nil {
			return err
		}
		delete(object, snippetIDKey)
		object[snippetCodeKey] = code
		return nil
	})
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return normalizeJSON(string(encoded))
}

// toBioT turns config_json into the body the update API takes: template names become this
// environment's ids, and snippet code becomes a snippet (see snippetFor). Errors name the place in
// config_json they come from.
func (r *resolver) toBioT(ctx context.Context, raw string) (string, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return "", err
	}

	indexes, templates := tabTemplates(config)
	for i, template := range templates {
		name, ok := template["name"].(string)
		if !ok {
			continue
		}
		id, err := r.templateID(ctx, name)
		if errors.Is(err, errTemplateNotFound) {
			return "", errors.New(templateNotFoundError(name, fmt.Sprintf("config_json tabs[%d]", indexes[i])))
		}
		if err != nil {
			return "", err
		}
		delete(template, "name")
		template["id"] = id
	}

	// Code the view already uses keeps its snippet; new or changed code gets a new one. The
	// snippets the view stops using are deleted by BioT once they are a day old.
	err := walkObjects(config, "", func(object map[string]any, at string) error {
		code, ok := object[snippetCodeKey].(string)
		if !ok {
			return nil
		}
		id, err := r.snippetFor(ctx, code)
		if err != nil {
			return fmt.Errorf("failed to create the code snippet at config_json %s: %w", at, err)
		}
		delete(object, snippetCodeKey)
		object[snippetIDKey] = id
		return nil
	})
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
