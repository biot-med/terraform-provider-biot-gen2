package abac

import (
	"context"
	"fmt"
	"iter"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

// ListSpec describes how one access-control list resource turns search results into
// `terraform query` results. See ListResults.
type ListSpec[T any] struct {
	// Items is the search to run, usually one of the abacapi.Client Search methods.
	Items iter.Seq2[T, error]

	// Keep filters results the service cannot filter itself. Nil keeps everything.
	Keep func(T) bool

	ID          func(T) string
	DisplayName func(T) string

	// Model maps a result to the resource's Terraform model. It only runs when Terraform asks
	// for full objects, i.e. for -generate-config-out.
	Model func(context.Context, T) (any, diag.Diagnostics)
}

// ListResults streams search results as list results. It is shared by the condition, action
// and rule list resources, which differ only in their search and mapping.
//
// It stops at the block's limit (Terraform's default is 100) counting only results that were
// kept, so client-side filters such as built_in still return up to the limit.
func ListResults[T any](ctx context.Context, req list.ListRequest, entity Entity, spec ListSpec[T]) iter.Seq[list.ListResult] {
	return func(push func(list.ListResult) bool) {
		var pushed int64

		for item, err := range spec.Items {
			if err != nil {
				var diags diag.Diagnostics
				diags.AddError("API Error", fmt.Sprintf("Failed to list %ss: %s", entity.Noun, err))
				push(list.ListResult{Diagnostics: diags})
				return
			}

			if spec.Keep != nil && !spec.Keep(item) {
				continue
			}

			result := req.NewListResult(ctx)
			result.DisplayName = spec.DisplayName(item)
			result.Diagnostics.Append(result.Identity.SetAttribute(ctx, path.Root("id"), spec.ID(item))...)

			if req.IncludeResource {
				model, diags := spec.Model(ctx, item)
				result.Diagnostics.Append(diags...)
				if !diags.HasError() {
					result.Diagnostics.Append(result.Resource.Set(ctx, model)...)
				}
			}

			if !push(result) {
				return
			}

			pushed++
			if req.Limit > 0 && pushed >= req.Limit {
				return
			}
		}
	}
}

// MatchesBuiltIn applies a list block's optional built_in filter to an object's raw tags. The
// search endpoints for conditions and actions cannot filter on tags, so this runs locally.
func MatchesBuiltIn(filter types.Bool, tags []string) bool {
	if filter.IsNull() || filter.IsUnknown() {
		return true
	}

	return HasTag(tags, BuiltInTag) == filter.ValueBool()
}

// ListNotConfigured is the result stream for a list resource that never received its API
// client - a provider bug, reported instead of panicking.
func ListNotConfigured(entity Entity) iter.Seq[list.ListResult] {
	var diags diag.Diagnostics
	diags.AddError("Provider not configured",
		fmt.Sprintf("The biot provider has no API client for listing %ss. This is a bug in the provider.", entity.Noun))

	return list.ListResultsStreamDiagnostics(diags)
}

// ExcludeStaticTypes is the search filter that leaves out CLASS conditions and actions: the
// static implementations BioT ships, which cannot be created and have no params to manage.
// It is applied by the service - responses do not carry the type - and expressed as "not
// CLASS" rather than a list of allowed types, so that types added later are still listed.
func ExcludeStaticTypes() transport.SearchFilter {
	return transport.SearchFilter{NotIn: []string{"CLASS"}}
}

// DescribeObject is the display name for a condition or action in `terraform query` output.
// Terraform already prints the id next to it, so it describes the rest.
func DescribeObject(value string, description string) string {
	if description == "" {
		return value
	}

	return value + ": " + description
}
