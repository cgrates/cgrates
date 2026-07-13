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
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const (
	defaultQueryLimit = 10
	queryBatchSize    = 1000
)

var keywordHeaderFields = map[string]struct{}{
	"RequestEncoding":    {},
	"RequestSource":      {},
	"RequestDestination": {},
	"RequestMethod":      {},
	utils.ReplyError:     {},
}

func newAnalyzerIndexMapping() *mapping.IndexMappingImpl {
	indexMapping := bleve.NewIndexMapping()
	documentMapping := bleve.NewDocumentStaticMapping()
	documentMapping.AddSubDocumentMapping("_all", bleve.NewDocumentDisabledMapping())
	for field := range keywordHeaderFields {
		fieldMapping := bleve.NewKeywordFieldMapping()
		fieldMapping.IncludeTermVectors = false
		fieldMapping.DocValues = false
		fieldMapping.SkipFreqNorm = true
		documentMapping.AddFieldMappingsAt(field, fieldMapping)
	}
	idMapping := bleve.NewNumericFieldMapping()
	idMapping.Index = false
	idMapping.DocValues = false
	documentMapping.AddFieldMappingsAt("RequestID", idMapping)
	durationMapping := bleve.NewNumericFieldMapping()
	durationMapping.DocValues = false
	durationMapping.SkipFreqNorm = true
	documentMapping.AddFieldMappingsAt(utils.RequestDuration, durationMapping)
	dateMapping := bleve.NewDateTimeFieldMapping()
	dateMapping.SkipFreqNorm = true
	documentMapping.AddFieldMappingsAt(utils.RequestStartTime, dateMapping)
	for _, field := range []string{utils.RequestParams, utils.Reply} {
		fieldMapping := bleve.NewTextFieldMapping()
		fieldMapping.Index = false
		fieldMapping.IncludeTermVectors = false
		fieldMapping.DocValues = false
		documentMapping.AddFieldMappingsAt(field, fieldMapping)
	}
	indexMapping.DefaultMapping = documentMapping
	return indexMapping
}

func queryFromFilters(filters []string) (query.Query, error) {
	queries := make([]query.Query, 0, len(filters))
	// Later Bleve conditions could hide an earlier FilterS error or side effect.
	addToQuery := true
	for _, filterID := range filters {
		if !strings.HasPrefix(filterID, utils.Meta) {
			addToQuery = false
			continue
		}
		filter, err := engine.NewFilterFromInline("", filterID)
		if err != nil {
			return nil, fmt.Errorf("invalid analyzer query filter <%s>: %w", filterID, err)
		}
		rule := filter.Rules[0]
		if !addToQuery {
			continue
		}
		field, ok := strings.CutPrefix(rule.Element, utils.DynamicDataPrefix+utils.MetaHdr+utils.NestingSep)
		if !ok {
			addToQuery = false
			continue
		}
		q := headerQuery(field, rule)
		if q == nil {
			addToQuery = false
			continue
		}
		queries = append(queries, q)
	}
	switch len(queries) {
	case 0:
		return bleve.NewMatchAllQuery(), nil
	case 1:
		return queries[0], nil
	default:
		return bleve.NewConjunctionQuery(queries...), nil
	}
}

func headerQuery(field string, rule *engine.FilterRule) query.Query {
	switch field {
	case utils.RequestStartTime:
		return dateRangeQuery(rule)
	case utils.RequestDuration:
		return durationRangeQuery(rule)
	}
	if rule.Type != utils.MetaString && rule.Type != utils.MetaPrefix {
		return nil
	}
	if _, ok := keywordHeaderFields[field]; !ok {
		return nil
	}
	queries := make([]query.Query, 0, len(rule.Values))
	for _, rawValue := range rule.Values {
		value, ok := literalFilterValue(rawValue)
		if !ok {
			return nil
		}
		var q query.FieldableQuery
		if rule.Type == utils.MetaString {
			q = bleve.NewTermQuery(value)
		} else {
			q = bleve.NewPrefixQuery(value)
		}
		q.SetField(field)
		queries = append(queries, q)
	}
	switch len(queries) {
	case 0:
		return nil
	case 1:
		return queries[0]
	default:
		return bleve.NewDisjunctionQuery(queries...)
	}
}

func dateRangeQuery(rule *engine.FilterRule) query.Query {
	queries := make([]query.Query, 0, len(rule.Values))
	for _, rawValue := range rule.Values {
		value, ok := literalFilterValue(rawValue)
		if !ok {
			return nil
		}
		bound, ok := dateBound(value)
		if !ok {
			return nil
		}
		// *lt may match the same moment in another timezone.
		inclusive := rule.Type != utils.MetaGreaterThan
		var q *query.DateRangeQuery
		switch rule.Type {
		case utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
			q = query.NewDateRangeInclusiveQuery(bound, time.Time{}, &inclusive, nil)
		case utils.MetaLessThan, utils.MetaLessOrEqual:
			q = query.NewDateRangeInclusiveQuery(time.Time{}, bound, nil, &inclusive)
		default:
			return nil
		}
		q.SetField(utils.RequestStartTime)
		if err := q.Validate(); err != nil {
			return nil
		}
		queries = append(queries, q)
	}
	switch len(queries) {
	case 0:
		return nil
	case 1:
		return queries[0]
	default:
		return bleve.NewDisjunctionQuery(queries...)
	}
}

func dateBound(value string) (time.Time, bool) {
	// FilterRule formats literal times as RFC3339 before comparing them.
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return time.Time{}, false
	}
	filterTime, ok := utils.StringToInterface(value).(time.Time)
	if !ok {
		return time.Time{}, false
	}
	bound, err := time.Parse(time.RFC3339Nano, utils.IfaceAsString(filterTime))
	if err != nil || bound.IsZero() {
		return time.Time{}, false
	}
	return bound, true
}

func durationRangeQuery(rule *engine.FilterRule) query.Query {
	queries := make([]query.Query, 0, len(rule.Values))
	for _, rawValue := range rule.Values {
		value, ok := literalFilterValue(rawValue)
		if !ok {
			return nil
		}
		bound, ok := durationBound(value)
		if !ok {
			return nil
		}
		inclusive := rule.Type == utils.MetaGreaterOrEqual || rule.Type == utils.MetaLessOrEqual
		var q *query.NumericRangeQuery
		switch rule.Type {
		case utils.MetaGreaterThan, utils.MetaGreaterOrEqual:
			q = query.NewNumericRangeInclusiveQuery(&bound, nil, &inclusive, nil)
		case utils.MetaLessThan, utils.MetaLessOrEqual:
			q = query.NewNumericRangeInclusiveQuery(nil, &bound, nil, &inclusive)
		default:
			return nil
		}
		q.SetField(utils.RequestDuration)
		queries = append(queries, q)
	}
	switch len(queries) {
	case 0:
		return nil
	case 1:
		return queries[0]
	default:
		return bleve.NewDisjunctionQuery(queries...)
	}
}

func durationBound(value string) (float64, bool) {
	switch bound := utils.StringToInterface(value).(type) {
	case int64:
		return float64(bound), true
	case float64:
		if math.IsNaN(bound) || math.IsInf(bound, 0) {
			return 0, false
		}
		return bound, true
	case time.Duration:
		return float64(bound), true
	default:
		return 0, false
	}
}

func literalFilterValue(value string) (string, bool) {
	parser, err := utils.NewRSRParser(value)
	if err != nil || parser == nil {
		return "", false
	}
	// FilterS handles values that depend on other fields or change before comparison.
	if parser.Rules != parser.Path || strings.HasPrefix(parser.Path, utils.DynamicDataPrefix) {
		return "", false
	}
	return parser.Path, true
}
