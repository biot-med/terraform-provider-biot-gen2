package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

type staticToken struct{}

func (staticToken) AccessToken(context.Context) (string, error) { return "t", nil }

type item struct {
	ID string `json:"id"`
}

// fakeSearch serves `total` items. capLimit > 0 makes it return fewer per page than asked for;
// omitTotal leaves metadata out; failOnPage returns a 500 for that page.
func fakeSearch(t *testing.T, total, capLimit int, omitTotal bool, failOnPage int, requests *int32) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(requests, 1)
		var req searchRequest
		raw, _ := url.QueryUnescape(r.URL.Query().Get("searchRequest"))
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Errorf("bad searchRequest %q: %v", raw, err)
		}
		if req.Page == failOnPage {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"BOOM","message":"boom"}`))
			return
		}

		size := req.Limit
		if capLimit > 0 && capLimit < size {
			size = capLimit
		}
		// Offsets follow the page size the server actually uses.
		data := []item{}
		for i := req.Page * size; i < total && i < (req.Page+1)*size; i++ {
			data = append(data, item{ID: fmt.Sprintf("i%03d", i)})
		}

		body := map[string]any{"data": data}
		if !omitTotal {
			body["metadata"] = map[string]any{"page": map[string]any{"totalResults": total}}
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)

	return New(srv.URL).WithTokens(staticToken{})
}

func collect(t *testing.T, c *Client) ([]string, error) {
	t.Helper()
	var ids []string
	for it, err := range SearchAll[item](context.Background(), c, c.BaseURL+"/things", nil) {
		if err != nil {
			return ids, err
		}
		ids = append(ids, it.ID)
	}
	return ids, nil
}

func TestSearchAllPaging(t *testing.T) {
	cases := []struct {
		name         string
		total        int
		capLimit     int
		omitTotal    bool
		wantRequests int32
	}{
		{"total reported", 250, 0, false, 3},
		{"total reported, exact multiple of the page size", 200, 0, false, 2},
		{"no total: a short page ends it", 250, 0, true, 3},
		{"no total, exact multiple: an empty page ends it", 200, 0, true, 3},
		{"server caps the page size: the total keeps it going", 120, 50, false, 3},
		{"nothing found", 0, 0, false, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var requests int32
			ids, err := collect(t, fakeSearch(t, c.total, c.capLimit, c.omitTotal, -1, &requests))
			if err != nil {
				t.Fatal(err)
			}
			if len(ids) != c.total {
				t.Fatalf("got %d items, want %d", len(ids), c.total)
			}
			seen := map[string]bool{}
			for _, id := range ids {
				if seen[id] {
					t.Fatalf("duplicate item %s", id)
				}
				seen[id] = true
			}
			if requests != c.wantRequests {
				t.Fatalf("made %d requests, want %d", requests, c.wantRequests)
			}
		})
	}
}

func TestSearchAllStopsOnError(t *testing.T) {
	var requests int32
	ids, err := collect(t, fakeSearch(t, 250, 0, false, 1, &requests))
	if err == nil {
		t.Fatal("expected an error from page 1")
	}
	if len(ids) != 100 || requests != 2 {
		t.Fatalf("got %d items over %d requests, want 100 over 2", len(ids), requests)
	}
}

func TestSearchAllStopsWhenTheCallerDoes(t *testing.T) {
	var requests int32
	c := fakeSearch(t, 250, 0, false, -1, &requests)
	n := 0
	for range SearchAll[item](context.Background(), c, c.BaseURL+"/things", nil) {
		n++
		if n == 5 {
			break
		}
	}
	if requests != 1 {
		t.Fatalf("made %d requests after the caller stopped, want 1", requests)
	}
}
