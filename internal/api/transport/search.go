package transport

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// SearchFilter is one field's condition in BioT's SearchRequestV2 filter (FilterV2).
type SearchFilter struct {
	Eq    any      `json:"eq,omitempty"`
	In    []string `json:"in,omitempty"`
	NotIn []string `json:"notIn,omitempty"`
}

type searchOrder struct {
	Prop  string `json:"prop"`
	Order string `json:"order"`
}

type searchRequest struct {
	Filter map[string]SearchFilter `json:"filter,omitempty"`
	Sort   []searchOrder           `json:"sort"`
	Page   int                     `json:"page"`
	Limit  int                     `json:"limit"`
}

type searchResponse[T any] struct {
	Data     []T `json:"data"`
	Metadata struct {
		Page struct {
			TotalResults int64 `json:"totalResults"`
		} `json:"page"`
	} `json:"metadata"`
}

// searchPageSize matches SearchRequestV2's default limit. The service sets no maximum.
const searchPageSize = 100

// SearchAll pages through a BioT search endpoint and yields every match. endpoint is the
// collection URL, e.g. .../access-control/v1/actions; the searchRequest query parameter is
// added here.
//
// Results are sorted by id, so that pages stay stable while they are being read. Paging stops
// at an empty page or once the reported total has been read - not on a short page, which would
// end early if the service ever capped the page size below what was asked for. Only when the
// response carries no total does a short page end it.
//
// An error is yielded once, after which iteration stops.
func SearchAll[T any](ctx context.Context, c *Client, endpoint string, filter map[string]SearchFilter) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		var read int64

		for page := 0; ; page++ {
			encoded, err := EncodeSearchRequest(searchRequest{
				Filter: filter,
				Sort:   []searchOrder{{Prop: "id", Order: "ASC"}},
				Page:   page,
				Limit:  searchPageSize,
			})
			if err != nil {
				yield(zero, err)
				return
			}

			url := fmt.Sprintf("%s?searchRequest=%s", endpoint, encoded)
			response, err := Do[searchResponse[T]](ctx, c, http.MethodGet, url, nil)
			if err != nil {
				yield(zero, err)
				return
			}

			for _, item := range response.Data {
				if !yield(item, nil) {
					return
				}
			}

			read += int64(len(response.Data))
			if len(response.Data) == 0 {
				return
			}

			// Normally the reported total ends paging. If it is missing (decodes as 0), fall
			// back to a short page as the end, rather than stopping after the first page.
			if total := response.Metadata.Page.TotalResults; total > 0 {
				if read >= total {
					return
				}
			} else if len(response.Data) < searchPageSize {
				return
			}
		}
	}
}
