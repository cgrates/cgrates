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

	"github.com/cgrates/birpc/context"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService initializes a AnalyzerService
func NewAnalyzerService(cfg *config.CGRConfig) (aS *AnalyzerService, err error) {
	aS = &AnalyzerService{
		cfg: cfg,
	}
	err = aS.initDB()
	return
}

// AnalyzerService is the service handling analyzer
type AnalyzerService struct {
	db  bleve.Index
	cfg *config.CGRConfig

	filterS *engine.FilterS
}

// SetFilterS will set the filterS used in APIs
// this function is called before the API is registerd
func (aS *AnalyzerService) SetFilterS(fS *engine.FilterS) {
	aS.filterS = fS
}

func (aS *AnalyzerService) initDB() (err error) {
	dbPath := path.Join(aS.cfg.AnalyzerSCfg().DBPath, utils.AnzDBDir)
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
func (aS *AnalyzerService) ListenAndServe(stopChan <-chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AnalyzerS))
	if err = aS.clenaUp(); err != nil { // clean up the data at the system start
		return
	}
	for {
		select {
		case <-stopChan:
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
	return aS.db.Index(utils.ConcatenatedKey(enc, from, to, method, strconv.FormatInt(sTime.Unix(), 10)),
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
	lenContentFltrs := len(args.ContentFilters)
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
		if val, has := obj.Fields[utils.ReplyError]; !has || len(utils.IfaceAsString(val)) == 0 {
			obj.Fields[utils.ReplyError] = nil
		}
		if lenContentFltrs != 0 {
			dp, err := getDPFromSearchresult(req, rep, obj.Fields)
			if err != nil {
				return err
			}
			if pass, err := aS.filterS.Pass(context.TODO(), aS.cfg.GeneralCfg().DefaultTenant,
				args.ContentFilters, dp); err != nil {
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
