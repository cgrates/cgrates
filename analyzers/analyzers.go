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
	"encoding/json"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerS initializes a AnalyzerService
func NewAnalyzerS(cfg *config.CGRConfig) (aS *AnalyzerS, err error) {
	aS = &AnalyzerS{
		cfg: cfg,
	}
	err = aS.initDB()
	return
}

// AnalyzerS is the service handling analyzer
type AnalyzerS struct {
	db  bleve.Index
	cfg *config.CGRConfig

	fltrS *engine.FilterS
}

// SetFilterS will set the filterS used in APIs
// this function is called before the API is registerd
func (aS *AnalyzerS) SetFilterS(fS *engine.FilterS) {
	aS.fltrS = fS
}

func (aS *AnalyzerS) initDB() (err error) {
	if aS.cfg.AnalyzerSCfg().IndexType == utils.MetaInternal {
		aS.db, err = bleve.NewMemOnly(newAnalyzerIndexMapping())
		return
	}
	dbPath := path.Join(aS.cfg.AnalyzerSCfg().DBPath, utils.AnzDBDir)
	if _, err = os.Stat(dbPath); err == nil {
		aS.db, err = bleve.Open(dbPath)
	} else if os.IsNotExist(err) {
		indxType, storeType := getIndex(aS.cfg.AnalyzerSCfg().IndexType)
		aS.db, err = bleve.NewUsing(dbPath,
			newAnalyzerIndexMapping(), indxType, storeType, nil)
	}
	return
}

func (aS *AnalyzerS) clenaUp() (err error) {
	t2 := bleve.NewDateRangeQuery(time.Time{}, time.Now().Add(-aS.cfg.AnalyzerSCfg().TTL))
	t2.SetField(utils.RequestStartTime)
	searchReq := bleve.NewSearchRequest(t2)
	var res *bleve.SearchResult
	if res, err = aS.db.Search(searchReq); err != nil {
		return
	}
	return aS.deleteHits(res.Hits)
}

// extracted as function in order to test this
func (aS *AnalyzerS) deleteHits(hits search.DocumentMatchCollection) (err error) {
	hasErr := false
	for _, hit := range hits {
		if err = aS.db.Delete(hit.ID); err != nil {
			hasErr = true
		}
	}
	if hasErr {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// ListenAndServe will initialize the service
func (aS *AnalyzerS) ListenAndServe(ctx *context.Context) (err error) {
	if err = aS.clenaUp(); err != nil { // clean up the data at the system start
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(aS.cfg.AnalyzerSCfg().CleanupInterval):
			if err = aS.clenaUp(); err != nil {
				return
			}
		}
	}
}

// Shutdown is called to shutdown the service
func (aS *AnalyzerS) Shutdown() error {
	return aS.db.Close()
}

func (aS *AnalyzerS) logTrafic(id uint64, method string,
	params, result, err any,
	enc, from, to string, sTime, eTime time.Time) error {
	if strings.HasPrefix(method, utils.AnalyzerSv1) {
		return nil
	}
	return aS.db.Index(utils.ConcatenatedKey(enc, from, to, method, strconv.FormatInt(rand.Int63n(100000000000000), 16)),
		NewInfoRPC(id, method, params, result, err, enc, from, to, sTime, eTime))
}

// QueryArgs contains filters and pagination options for Analyzer queries.
type QueryArgs struct {
	// Filters contains inline filters and named filter IDs.
	Filters []string
	// Limit is the maximum number of results to return.
	Limit int
	// Offset is the starting position for results, used for pagination.
	Offset int
}

// V1StringQuery returns RPC calls matching the filters.
func (aS *AnalyzerS) V1StringQuery(ctx *context.Context, args *QueryArgs, reply *[]map[string]any) error {
	if len(args.Filters) == 0 {
		return aS.queryAll(ctx, args, reply)
	}
	q, err := queryFromFilters(args.Filters)
	if err != nil {
		return err
	}
	return aS.queryFiltered(ctx, q, args, reply)
}

func (aS *AnalyzerS) queryAll(ctx *context.Context,
	args *QueryArgs, reply *[]map[string]any) error {
	s := bleve.NewSearchRequest(bleve.NewMatchAllQuery())
	s.Fields = []string{utils.Meta}
	s.SortBy([]string{"-" + utils.RequestStartTime, "_id"})
	if args.Limit > 0 {
		s.Size = args.Limit
	}
	if args.Offset > 0 {
		s.From = args.Offset
	}
	searchResults, err := aS.db.SearchInContext(ctx, s)
	if err != nil {
		return err
	}
	rply := make([]map[string]any, 0, searchResults.Hits.Len())
	for _, obj := range searchResults.Hits {
		prepareQueryFields(obj.Fields)
		rply = append(rply, obj.Fields)
	}
	*reply = rply
	return nil
}

func (aS *AnalyzerS) queryFiltered(ctx *context.Context, q query.Query,
	args *QueryArgs, reply *[]map[string]any) error {
	limit := args.Limit
	if limit <= 0 {
		limit = defaultQueryLimit
	}
	offset := args.Offset
	if offset < 0 {
		offset = 0
	}
	rply := make([]map[string]any, 0, limit)
	skipped := 0
	var searchAfter []string
	for len(rply) < limit {
		s := bleve.NewSearchRequest(q)
		s.Fields = []string{utils.Meta}
		s.SortBy([]string{"-" + utils.RequestStartTime, "_id"})
		s.Size = queryBatchSize
		s.SearchAfter = searchAfter
		searchResults, err := aS.db.SearchInContext(ctx, s)
		if err != nil {
			return err
		}
		if searchResults.Hits.Len() == 0 {
			break
		}
		for _, obj := range searchResults.Hits {
			req, rep := prepareQueryFields(obj.Fields)
			dp, err := getDPFromSearchresult(req, rep, obj.Fields)
			if err != nil {
				return err
			}
			pass, err := aS.fltrS.Pass(ctx, aS.cfg.GeneralCfg().DefaultTenant, args.Filters, dp)
			if err != nil {
				return err
			}
			if !pass {
				continue
			}
			if skipped < offset {
				skipped++
				continue
			}
			rply = append(rply, obj.Fields)
			if len(rply) == limit {
				break
			}
		}
		// Resume after the last hit, since filtering can reject the whole batch.
		searchAfter = searchResults.Hits[searchResults.Hits.Len()-1].Sort
	}
	*reply = rply
	return nil
}

func prepareQueryFields(fields map[string]any) (req, rep json.RawMessage) {
	req = json.RawMessage(utils.IfaceAsString(fields[utils.RequestParams]))
	rep = json.RawMessage(utils.IfaceAsString(fields[utils.Reply]))
	fields[utils.RequestParams] = req
	fields[utils.Reply] = rep
	if dur, err := utils.IfaceAsDuration(fields[utils.RequestDuration]); err == nil {
		fields[utils.RequestDuration] = dur.String()
	}
	if val, has := fields[utils.ReplyError]; !has || len(utils.IfaceAsString(val)) == 0 {
		fields[utils.ReplyError] = nil
	}
	return req, rep
}
