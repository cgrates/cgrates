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
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerService initializes a AnalyzerService
func NewAnalyzerService(cfg *config.CGRConfig) (aS *AnalyzerService, err error) {
	aS = &AnalyzerService{cfg: cfg}
	err = aS.initDB()
	return
}

// AnalyzerService is the service handling analyzer
type AnalyzerService struct {
	db  bleve.Index
	cfg *config.CGRConfig
}

func (aS *AnalyzerService) initDB() (err error) {
	if _, err = os.Stat(aS.cfg.AnalyzerSCfg().DBPath); err == nil {
		aS.db, err = bleve.Open(aS.cfg.AnalyzerSCfg().DBPath)
	} else if os.IsNotExist(err) {
		indxType, storeType := getIndex(aS.cfg.AnalyzerSCfg().IndexType)
		aS.db, err = bleve.NewUsing(aS.cfg.AnalyzerSCfg().DBPath,
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
func (aS *AnalyzerService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AnalyzerS))
	if err = aS.clenaUp(); err != nil { // clean up the data at the system start
		return
	}
	for {
		select {
		case e := <-exitChan:
			exitChan <- e // put back for the others listening for shutdown request
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

// V1StringQuery returns a list of API that match the query
func (aS *AnalyzerService) V1StringQuery(searchstr string, reply *[]map[string]interface{}) error {

	s := bleve.NewSearchRequest(bleve.NewQueryStringQuery(searchstr))
	s.Fields = []string{utils.Meta} // return all fields
	searchResults, err := aS.db.Search(s)
	if err != nil {
		return err
	}
	rply := make([]map[string]interface{}, searchResults.Hits.Len())
	for i, obj := range searchResults.Hits {
		rply[i] = obj.Fields
		// make sure that the result is corectly marshaled
		rply[i][utils.Reply] = json.RawMessage(utils.IfaceAsString(obj.Fields[utils.Reply]))
		rply[i][utils.RequestParams] = json.RawMessage(utils.IfaceAsString(obj.Fields[utils.RequestParams]))
		// try to pretty print the duration
		if dur, err := utils.IfaceAsDuration(rply[i][utils.RequestDuration]); err == nil {
			rply[i][utils.RequestDuration] = dur.String()
		}
	}
	*reply = rply
	return nil
}
