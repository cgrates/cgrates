/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package analyzers

import (
	"strings"
	"testing"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestQueryFromFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters []string
		check   func(*testing.T, query.Query)
		wantErr string
	}{
		{
			name:    "exact string",
			filters: []string{"*string:~*hdr.RequestMethod:CoreSv1.Ping"},
			check: func(t *testing.T, q query.Query) {
				term, ok := q.(*query.TermQuery)
				if !ok || term.FieldVal != "RequestMethod" || term.Term != "CoreSv1.Ping" {
					t.Fatalf("got %#v", q)
				}
			},
		},
		{
			name:    "prefix",
			filters: []string{"*prefix:~*hdr.RequestEncoding:*go"},
			check: func(t *testing.T, q query.Query) {
				prefix, ok := q.(*query.PrefixQuery)
				if !ok || prefix.FieldVal != "RequestEncoding" || prefix.Prefix != "*go" {
					t.Fatalf("got %#v", q)
				}
			},
		},
		{
			name:    "string alternatives",
			filters: []string{"*string:~*hdr.RequestMethod:CoreSv1.Ping|CoreSv1.Status"},
			check: func(t *testing.T, q query.Query) {
				disjunction, ok := q.(*query.DisjunctionQuery)
				if !ok || len(disjunction.Disjuncts) != 2 {
					t.Fatalf("got %#v", q)
				}
				for i, value := range []string{"CoreSv1.Ping", "CoreSv1.Status"} {
					term, ok := disjunction.Disjuncts[i].(*query.TermQuery)
					if !ok || term.FieldVal != "RequestMethod" || term.Term != value {
						t.Fatalf("alternative %d: got %#v", i, disjunction.Disjuncts[i])
					}
				}
			},
		},
		{
			name:    "date range",
			filters: []string{"*lt:~*hdr.RequestStartTime:2025-01-15T10:30:00+02:00"},
			check: func(t *testing.T, q query.Query) {
				dateRange, ok := q.(*query.DateRangeQuery)
				if !ok || dateRange.FieldVal != utils.RequestStartTime {
					t.Fatalf("got %#v", q)
				}
			},
		},
		{
			name:    "duration range",
			filters: []string{"*gt:~*hdr.RequestDuration:1m"},
			check: func(t *testing.T, q query.Query) {
				durationRange, ok := q.(*query.NumericRangeQuery)
				if !ok || durationRange.FieldVal != utils.RequestDuration || durationRange.Min == nil ||
					*durationRange.Min != float64(time.Minute) {
					t.Fatalf("got %#v", q)
				}
			},
		},
		{
			name: "stop after named filter",
			filters: []string{
				"*string:~*hdr.RequestMethod:CoreSv1.Ping",
				"METHOD_FILTER",
				"*string:~*hdr.RequestEncoding:*json",
			},
			check: func(t *testing.T, q query.Query) {
				term, ok := q.(*query.TermQuery)
				if !ok || term.FieldVal != "RequestMethod" || term.Term != "CoreSv1.Ping" {
					t.Fatalf("got %#v", q)
				}
			},
		},
		{
			name:    "malformed inline after named filter",
			filters: []string{"METHOD_FILTER", "*regex:~*hdr.RequestMethod:["},
			wantErr: "error parsing regexp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := queryFromFilters(tt.filters)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("got %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tt.check(t, q)
		})
	}
}

func TestQueryIncludesFilterMatches(t *testing.T) {
	filterTime := time.Date(2025, time.January, 15, 10, 30, 0, 0,
		time.FixedZone("test", 2*60*60))
	tests := []struct {
		name   string
		filter string
		doc    map[string]any
	}{
		{
			name:   "stop word keyword",
			filter: "*string:~*hdr.RequestSource:the",
			doc:    map[string]any{"RequestSource": "the"},
		},
		{
			name:   "OR alternatives",
			filter: "*string:~*hdr.RequestEncoding:*|*json",
			doc:    map[string]any{"RequestEncoding": "*json"},
		},
		{
			name:   "converted value stays in FilterS",
			filter: "*string:~*hdr.RequestEncoding:gob{*string2hex}",
			doc:    map[string]any{"RequestEncoding": "0x676f62"},
		},
		{
			name:   "strict duration boundary",
			filter: "*gt:~*hdr.RequestDuration:1m",
			doc:    map[string]any{utils.RequestDuration: time.Minute + time.Nanosecond},
		},
		{
			name:   "same moment in another timezone",
			filter: "*lt:~*hdr.RequestStartTime:" + filterTime.Format(time.RFC3339Nano),
			doc: map[string]any{
				utils.RequestStartTime: filterTime.UTC().Format(time.RFC3339Nano),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, err := bleve.NewMemOnly(newAnalyzerIndexMapping())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := idx.Close(); err != nil {
					t.Error(err)
				}
			})
			if err := idx.Index("doc", tt.doc); err != nil {
				t.Fatal(err)
			}

			searchRequest := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
			searchRequest.Fields = []string{utils.Meta}
			result, err := idx.Search(searchRequest)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Hits) != 1 {
				t.Fatalf("got %d indexed documents", len(result.Hits))
			}
			fields := result.Hits[0].Fields
			if duration, err := utils.IfaceAsDuration(fields[utils.RequestDuration]); err == nil {
				fields[utils.RequestDuration] = duration.String()
			}
			filter, err := engine.NewFilterFromInline("", tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			pass, err := filter.Rules[0].Pass(context.Background(), utils.MapStorage{
				utils.MetaHdr: utils.MapStorage(fields),
			})
			if err != nil {
				t.Fatal(err)
			}
			if !pass {
				t.Fatal("indexed document does not pass the final filter")
			}

			q, err := queryFromFilters([]string{tt.filter})
			if err != nil {
				t.Fatal(err)
			}
			searchResult, err := idx.Search(bleve.NewSearchRequest(q))
			if err != nil {
				t.Fatal(err)
			}
			if searchResult.Total != 1 {
				t.Fatal("final match was excluded by the Bleve query")
			}
		})
	}
}
