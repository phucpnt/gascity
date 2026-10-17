package api

import (
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/gastownhall/gascity/internal/beads"
)

func TestBeadListSelectors(t *testing.T) {
	state := newFakeState(t)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := []beads.Bead{
		{ID: "gc-request", Metadata: map[string]string{"source": "chat", "thread": "42", "note": " a,b=c "}},
		{ID: "gc-next", Metadata: map[string]string{"source": "chat", "thread": "43"}},
		{ID: "gc-step", ParentID: "gc-request", Metadata: map[string]string{"thread": "42"}},
		{ID: "gc-empty", Metadata: map[string]string{"thread": ""}},
		{ID: "gc-backlog"},
	}
	for i := range rows {
		rows[i].Title = rows[i].ID
		rows[i].Type = "task"
		rows[i].Status = "closed"
		rows[i].Assignee = "worker"
		rows[i].CreatedAt = base.Add(time.Duration(i) * time.Minute)
	}
	// A Counter-capable backend exercises the bounded all=true path as well
	// as the usual hydrated path. Parent/key-existence selectors must narrow
	// candidates before either Total or a page limit is calculated.
	state.stores["myrig"] = &countingListStore{Store: beads.NewMemStoreFrom(len(rows), rows, nil)}
	h := newTestCityHandler(t, state)
	for _, tc := range []struct {
		name  string
		query url.Values
		want  []string
	}{
		{"metadata equality", url.Values{"metadata": {"source=chat"}}, []string{"gc-next", "gc-request"}},
		{"metadata conjunction", url.Values{"metadata": {"source=chat", "thread=42"}}, []string{"gc-request"}},
		{"value delimiters and whitespace", url.Values{"metadata": {"note= a,b=c "}}, []string{"gc-request"}},
		{"parent absent", url.Values{"parent": {"none"}}, []string{"gc-backlog", "gc-empty", "gc-next", "gc-request"}},
		{"parent present", url.Values{"parent": {"any"}}, []string{"gc-step"}},
		{"parent identity", url.Values{"parent": {"gc-request"}}, []string{"gc-step"}},
		{"key existence includes empty", url.Values{"has_metadata": {"thread"}}, []string{"gc-empty", "gc-step", "gc-next", "gc-request"}},
		{"key existence conjunction", url.Values{"has_metadata": {"thread", "source"}}, []string{"gc-next", "gc-request"}},
		{"absent key", url.Values{"has_metadata": {"unknown"}}, nil},
		{"empty value requires key", url.Values{"metadata": {"thread="}}, []string{"gc-empty"}},
		{"parent and metadata", url.Values{"parent": {"none"}, "metadata": {"source=chat"}, "has_metadata": {"thread"}}, []string{"gc-next", "gc-request"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := tc.query
			q.Set("rig", "myrig")
			q.Set("all", "true")
			q.Set("assignee", "worker")
			q.Set("limit", "1")
			var got []string
			for page := 0; ; page++ {
				if page > len(rows) {
					t.Fatal("pagination did not terminate")
				}
				rec := getBeads(t, h, cityURL(state, "/beads?")+q.Encode())
				if rec.Code != http.StatusOK {
					t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
				}
				items, total, next := decodeListBody(t, rec)
				if total != len(tc.want) {
					t.Fatalf("total = %d, want %d", total, len(tc.want))
				}
				for _, b := range items {
					got = append(got, b.ID)
				}
				if next == "" {
					break
				}
				q.Set("cursor", next)
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("ids = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBeadListSelectorsRejectMalformedFilters(t *testing.T) {
	state := newFakeState(t)
	h := newTestCityHandler(t, state)
	for _, query := range []url.Values{
		{"metadata": {"source"}},
		{"metadata": {"=chat"}},
		{"metadata": {"bad key=value"}},
		{"metadata": {"source=chat", "source=other"}},
		{"has_metadata": {"bad key"}},
	} {
		rec := getBeads(t, h, cityURL(state, "/beads?")+query.Encode())
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400; %s", query.Encode(), rec.Code, rec.Body.String())
		}
	}
}
