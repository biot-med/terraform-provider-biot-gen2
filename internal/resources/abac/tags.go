package abac

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The access-control service stamps the conditions, actions and rules it ships with a
// <<BuiltIn>> tag, and re-adds it on every update: TagsServiceImpl.getEntityTags carries it
// over from the previous value whether or not the request asked for it.
//
// That would make `tags` unusable on an imported built-in object - sending ["team-a"] comes
// back as ["<<BuiltIn>>", "team-a"], and Terraform rejects an applied value that differs
// from a non-null configured one. So the tag is kept out of `tags` entirely and reported
// through the read-only `built_in` attribute instead, which leaves `tags` behaving like any
// other optional field.
//
// Nothing else is added or removed. Tags written here are the tags BioT stores.
const BuiltInTag = "<<BuiltIn>>"

func HasTag(tags []string, wanted string) bool {
	for _, tag := range tags {
		if tag == wanted {
			return true
		}
	}

	return false
}

// WithoutBuiltInTag strips the server-managed built-in marker from an API response, leaving
// only the tags the configuration is responsible for.
func WithoutBuiltInTag(tags []string) []string {
	out := []string{}
	for _, tag := range tags {
		if tag == BuiltInTag {
			continue
		}
		out = append(out, tag)
	}

	return out
}

// RejectBuiltInTag reports a diagnostic if the configured tags include the built-in marker.
// Call it from ValidateConfig: the marker is stripped from every response, so a configured
// one would never survive into state and the apply would fail with "inconsistent result
// after apply" - a confusing way to learn that the tag is managed by BioT.
func RejectBuiltInTag(tags types.Set, entity Entity, diagnostics *diag.Diagnostics) {
	if tags.IsNull() || tags.IsUnknown() {
		return
	}

	// Walked element by element rather than through ElementsAs: a tag taken from a variable
	// is still unknown at validate time, and that is not an error.
	for _, element := range tags.Elements() {
		tag, ok := element.(types.String)
		if !ok || tag.IsNull() || tag.IsUnknown() {
			continue
		}

		if tag.ValueString() == BuiltInTag {
			diagnostics.AddAttributeError(
				path.Root("tags"),
				fmt.Sprintf("The tag %q is managed by BioT", BuiltInTag),
				fmt.Sprintf(`%q marks the %ss that BioT ships, and the service adds and removes it on its own.

Remove it from tags. Whether BioT ships this %s is reported by the read-only
built_in attribute.`, BuiltInTag, entity.Noun, entity.Noun),
			)
		}
	}
}
