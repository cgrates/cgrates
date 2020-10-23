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
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/blevesearch/bleve"
	// import the bleve packages in order to register the indextype and storagetype
	"github.com/blevesearch/bleve/document"
	_ "github.com/blevesearch/bleve/index/scorch"
	_ "github.com/blevesearch/bleve/index/store/boltdb"
	_ "github.com/blevesearch/bleve/index/store/goleveldb"
	_ "github.com/blevesearch/bleve/index/store/moss"
	_ "github.com/blevesearch/bleve/index/upsidedown"
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
	fmt.Println(aS.cfg.AnalyzerSCfg().DBPath)
	if _, err = os.Stat(aS.cfg.AnalyzerSCfg().DBPath); err == nil {
		fmt.Println("exista")
		aS.db, err = bleve.Open(aS.cfg.AnalyzerSCfg().DBPath)
	} else if os.IsNotExist(err) {
		fmt.Println("nu exista")
		aS.db, err = bleve.NewUsing(aS.cfg.AnalyzerSCfg().DBPath, bleve.NewIndexMapping(),
			aS.cfg.AnalyzerSCfg().IndexType, aS.cfg.AnalyzerSCfg().StoreType, nil)
	}
	return
}

// ListenAndServe will initialize the service
func (aS *AnalyzerService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AnalyzerS))
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
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
	info *extraInfo, sTime, eTime time.Time) error {
	var e interface{}
	switch val := err.(type) {
	default:
	case nil:
	case string:
		e = val
	case error:
		e = val.Error()
	}
	return aS.db.Index(utils.ConcatenatedKey(method, strconv.FormatInt(sTime.Unix(), 10)),
		InfoRPC{
			Duration:  eTime.Sub(sTime),
			StartTime: sTime,
			EndTime:   eTime,

			Encoding: info.enc,
			From:     info.from,
			To:       info.to,

			ID:     id,
			Method: method,
			Params: params,
			Result: result,
			Error:  e,
		})
}

func (aS *AnalyzerService) V1Search(searchstr string, reply *[]*document.Document) error {
	s := bleve.NewSearchRequest(bleve.NewQueryStringQuery(searchstr))
	searchResults, err := aS.db.Search(s)
	if err != nil {
		return err
	}
	rply := make([]*document.Document, searchResults.Hits.Len())
	for i, obj := range searchResults.Hits {
		fmt.Println(obj.ID)
		fmt.Println(obj.Index)
		d, _ := aS.db.Document(obj.ID)
		fmt.Println(d.Fields[0].Name())
		rply[i] = d
	}
	*reply = rply
	return nil
}
