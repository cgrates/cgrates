/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/esapi"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch"
)

func NewElasticExporter(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc *utils.SafeMapStorage) (eEe *ElasticEe, err error) {
	eEe = &ElasticEe{
		id:      cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg:  cgrCfg,
		cfgIdx:  cfgIdx,
		filterS: filterS,
		dc:      dc,
		reqs:    newConcReq(cgrCfg.EEsCfg().Exporters[cfgIdx].ConcurrentRequests),
	}
	err = eEe.init()
	return
}

// ElasticEe implements EventExporter interface for ElasticSearch export
type ElasticEe struct {
	id      string
	eClnt   *elasticsearch.Client
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	dc      *utils.SafeMapStorage
	opts    esapi.IndexRequest // this variable is used only for storing the options from OptsMap
	reqs    *concReq
}

// init will create all the necessary dependencies, including opening the file
func (eEe *ElasticEe) init() (err error) {
	// create the client
	if eEe.eClnt, err = elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: strings.Split(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].ExportPath, utils.InfieldSep),
		}); err != nil {
		return
	}
	//parse opts
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsIndex]; !has {
		eEe.opts.Index = utils.CDRsTBL
	} else {
		eEe.opts.Index = utils.IfaceAsString(val)
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsIfPrimaryTerm]; has {
		var intVal int64
		if intVal, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		eEe.opts.IfPrimaryTerm = utils.IntPointer(int(intVal))
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsIfSeqNo]; has {
		var intVal int64
		if intVal, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		eEe.opts.IfSeqNo = utils.IntPointer(int(intVal))
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsOpType]; has {
		eEe.opts.OpType = utils.IfaceAsString(val)
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsPipeline]; has {
		eEe.opts.Pipeline = utils.IfaceAsString(val)
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsRouting]; has {
		eEe.opts.Routing = utils.IfaceAsString(val)
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsTimeout]; has {
		if eEe.opts.Timeout, err = utils.IfaceAsDuration(val); err != nil {
			return
		}
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsVersionLow]; has {
		var intVal int64
		if intVal, err = utils.IfaceAsTInt64(val); err != nil {
			return
		}
		eEe.opts.Version = utils.IntPointer(int(intVal))
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsVersionType]; has {
		eEe.opts.VersionType = utils.IfaceAsString(val)
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.ElsWaitForActiveShards]; has {
		eEe.opts.WaitForActiveShards = utils.IfaceAsString(val)
	}
	return
}

// ID returns the identificator of this exporter
func (eEe *ElasticEe) ID() string {
	return eEe.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (eEe *ElasticEe) OnEvicted(_ string, _ interface{}) {

}

// ExportEvent implements EventExporter
func (eEe *ElasticEe) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	eEe.reqs.get()
	defer func() {
		updateEEMetrics(eEe.dc, cgrEv.ID, cgrEv.Event, err != nil, utils.FirstNonEmpty(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Timezone,
			eEe.cgrCfg.GeneralCfg().DefaultTimezone))
		eEe.reqs.done()
	}()
	eEe.dc.Lock()
	eEe.dc.MapStorage[utils.NumberOfEvents] = eEe.dc.MapStorage[utils.NumberOfEvents].(int64) + 1
	eEe.dc.Unlock()
	valMp := make(map[string]interface{})
	if len(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].ContentFields()) == 0 {
		valMp = cgrEv.Event
	} else {
		oNm := map[string]*utils.OrderedNavigableMap{
			utils.MetaExp: utils.NewOrderedNavigableMap(),
		}
		eeReq := engine.NewExportRequest(map[string]utils.DataStorage{
			utils.MetaReq:  utils.MapStorage(cgrEv.Event),
			utils.MetaDC:   eEe.dc,
			utils.MetaOpts: utils.MapStorage(cgrEv.APIOpts),
			utils.MetaCfg:  eEe.cgrCfg.GetDataProvider(),
		}, utils.FirstNonEmpty(cgrEv.Tenant, eEe.cgrCfg.GeneralCfg().DefaultTenant),
			eEe.filterS, oNm)
		if err = eeReq.SetFields(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].ContentFields()); err != nil {
			return
		}
		for el := eeReq.ExpData[utils.MetaExp].GetFirstElement(); el != nil; el = el.Next() {
			path := el.Value
			nmIt, _ := eeReq.ExpData[utils.MetaExp].Field(path)
			path = path[:len(path)-1] // remove the last index
			valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
		}
	}

	// Set up the request object
	cgrID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.CGRID), utils.GenUUID())
	runID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RunID), utils.MetaDefault)
	eReq := esapi.IndexRequest{
		Index:               eEe.opts.Index,
		DocumentID:          utils.ConcatenatedKey(cgrID, runID),
		Body:                strings.NewReader(utils.ToJSON(valMp)),
		Refresh:             "true",
		IfPrimaryTerm:       eEe.opts.IfPrimaryTerm,
		IfSeqNo:             eEe.opts.IfSeqNo,
		OpType:              eEe.opts.OpType,
		Parent:              eEe.opts.Parent,
		Pipeline:            eEe.opts.Pipeline,
		Routing:             eEe.opts.Routing,
		Timeout:             eEe.opts.Timeout,
		Version:             eEe.opts.Version,
		VersionType:         eEe.opts.VersionType,
		WaitForActiveShards: eEe.opts.WaitForActiveShards,
	}

	var resp *esapi.Response
	if resp, err = eReq.Do(context.Background(), eEe.eClnt); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.IsError() {
		var e map[string]interface{}
		if err = json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return
		} else {
			utils.Logger.Warning(fmt.Sprintf("<%s> Exporter with id: <%s> received error: <%+v> when indexing document",
				utils.EventExporterS, eEe.id, e))
		}
	}
	return
}

func (eEe *ElasticEe) GetMetrics() *utils.SafeMapStorage {
	return eEe.dc.Clone()
}
