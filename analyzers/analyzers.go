/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package analyzers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService initializes a AnalyzerService
func NewAnalyzerService(cfg *config.CGRConfig, filterS chan *engine.FilterS) (aS *AnalyzerService, err error) {
	aS = &AnalyzerService{
		cfg:         cfg,
		filterSChan: filterS,
	}
	err = aS.initDB()
	return
}

// AnalyzerService is the service handling analyzer
type AnalyzerService struct {
	db  bleve.Index
	cfg *config.CGRConfig

	// because we do not use the filters only for API
	// start the service without them
	// and populate them on the first API call
	filterSChan chan *engine.FilterS
	filterS     *engine.FilterS
}

func (aS *AnalyzerService) initDB() (err error) {
	dbPath := path.Join(aS.cfg.AnalyzerSCfg().DBPath, "db")
	if _, err = os.Stat(dbPath); err == nil {
		aS.db, err = bleve.Open(dbPath)
	} else if os.IsNotExist(err) {
		indxType, storeType := getIndex(aS.cfg.AnalyzerSCfg().IndexType)
		aS.db, err = bleve.NewUsing(dbPath,
			bleve.NewIndexMapping(), indxType, storeType, nil)
	}
	return
}

func (aS *AnalyzerService) clenaUp() (err error) {
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
func (aS *AnalyzerService) deleteHits(hits search.DocumentMatchCollection) (err error) {
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
func (aS *AnalyzerService) ListenAndServe(exitChan <-chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AnalyzerS))
	if err = aS.clenaUp(); err != nil { // clean up the data at the system start
		return
	}
	for {
		select {
		case <-exitChan:
			return
		case <-time.After(aS.cfg.AnalyzerSCfg().CleanupInterval):
			if err = aS.clenaUp(); err != nil {
				return
			}
		}
	}
}

// Shutdown is called to shutdown the service
func (aS *AnalyzerService) Shutdown() error {
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown initialized", utils.AnalyzerS))
	aS.db.Close()
	utils.Logger.Info(fmt.Sprintf("<%s> service shutdown complete", utils.AnalyzerS))
	return nil
}

func (aS *AnalyzerService) logTrafic(id uint64, method string,
	params, result, err interface{},
	enc, from, to string, sTime, eTime time.Time) error {
	if strings.HasPrefix(method, utils.AnalyzerSv1) {
		return nil
	}
	return aS.db.Index(utils.ConcatenatedKey(method, strconv.FormatInt(sTime.Unix(), 10)),
		NewInfoRPC(id, method, params, result, err, enc, from, to, sTime, eTime))
}

// QueryArgs the structure that we use to filter the API calls
type QueryArgs struct {
	// a string based on the query language(https://blevesearch.com/docs/Query-String-Query/) that we send to bleve
	HeaderFilters string
	// a list of filters that we use to filter the call similar to how we filter the events
	ContentFilters []string
}

// V1StringQuery returns a list of API that match the query
func (aS *AnalyzerService) V1StringQuery(args *QueryArgs, reply *[]map[string]interface{}) error {
	s := bleve.NewSearchRequest(bleve.NewQueryStringQuery(args.HeaderFilters))
	s.Fields = []string{utils.Meta} // return all fields
	searchResults, err := aS.db.Search(s)
	if err != nil {
		return err
	}
	rply := make([]map[string]interface{}, 0, searchResults.Hits.Len())
	lCntFltrs := len(args.ContentFilters)
	if lCntFltrs != 0 &&
		aS.filterS == nil { // populate the filter on the first API that requeres them
		aS.filterS = <-aS.filterSChan
		aS.filterSChan <- aS.filterS
	}
	for _, obj := range searchResults.Hits {
		// make sure that the result is corectly marshaled
		rep := json.RawMessage(utils.IfaceAsString(obj.Fields[utils.Reply]))
		req := json.RawMessage(utils.IfaceAsString(obj.Fields[utils.RequestParams]))
		obj.Fields[utils.Reply] = rep
		obj.Fields[utils.RequestParams] = req
		// try to pretty print the duration
		if dur, err := utils.IfaceAsDuration(obj.Fields[utils.RequestDuration]); err == nil {
			obj.Fields[utils.RequestDuration] = dur.String()
		}
		if lCntFltrs != 0 {
			repDP, err := unmarshalJSON(rep)
			if err != nil {
				return err
			}
			reqDP, err := unmarshalJSON(req)
			if err != nil {
				return err
			}
			if pass, err := aS.filterS.Pass(aS.cfg.GeneralCfg().DefaultTenant,
				args.ContentFilters, utils.MapStorage{
					utils.MetaReq: reqDP,
					utils.MetaRep: repDP,
				}); err != nil {
				return err
			} else if !pass {
				continue
			}
		}
		rply = append(rply, obj.Fields)
	}
	*reply = rply
	return nil
}
