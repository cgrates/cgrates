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
	"sync"

	"github.com/elastic/go-elasticsearch/esapi"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	elasticsearch "github.com/elastic/go-elasticsearch"
)

func NewElasticExporter(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (eEe *ElasticEe, err error) {
	eEe = &ElasticEe{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
	err = eEe.init()
	return
}

// ElasticEe implements EventExporter interface for ElasticSearch export
type ElasticEe struct {
	id      string
	eClnt   *elasticsearch.Client
	index   string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	sync.RWMutex
	dc utils.MapStorage
}

// init will create all the necessary dependencies, including opening the file
func (eEe *ElasticEe) init() (err error) {
	// create the client
	if eEe.eClnt, err = elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: strings.Split(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].ExportPath, utils.INFIELD_SEP),
		}); err != nil {
		return
	}
	if val, has := eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Opts[utils.Index]; !has {
		eEe.index = utils.CDRsTBL
	} else {
		eEe.index = utils.IfaceAsString(val)
	}
	return
}

// ID returns the identificator of this exporter
func (eEe *ElasticEe) ID() string {
	return eEe.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (eEe *ElasticEe) OnEvicted(_ string, _ interface{}) {
	return
}

// ExportEvent implements EventExporter
func (eEe *ElasticEe) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	eEe.Lock()
	defer func() {
		if err != nil {
			eEe.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			eEe.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		eEe.Unlock()
	}()
	eEe.dc[utils.NumberOfEvents] = eEe.dc[utils.NumberOfEvents].(int64) + 1

	req := utils.MapStorage{}
	for k, v := range cgrEv.Event {
		req[k] = v
	}
	valMp := make(map[string]string)
	eeReq := NewEventExporterRequest(req, eEe.dc,
		eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Tenant,
		eEe.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Timezone,
			eEe.cgrCfg.GeneralCfg().DefaultTimezone),
		eEe.filterS)
	if err = eeReq.SetFields(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].ContentFields()); err != nil {
		return
	}
	for el := eeReq.cnt.GetFirstElement(); el != nil; el = el.Next() {
		var nmIt utils.NMInterface
		if nmIt, err = eeReq.cnt.Field(el.Value); err != nil {
			return
		}
		itm, isNMItem := nmIt.(*config.NMItem)
		if !isNMItem {
			err = fmt.Errorf("cannot encode reply value: %s, err: not NMItems", utils.ToJSON(el.Value))
			return
		}
		if itm == nil {
			continue // all attributes, not writable to diameter packet
		}
		valMp[strings.Join(itm.Path, utils.NestingSep)] = utils.IfaceAsString(itm.Data)
	}
	updateEEMetrics(eEe.dc, cgrEv.Event, utils.FirstNonEmpty(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Timezone,
		eEe.cgrCfg.GeneralCfg().DefaultTimezone))
	// Set up the request object
	cgrID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.CGRID), utils.GenUUID())
	runID := utils.FirstNonEmpty(engine.MapEvent(cgrEv.Event).GetStringIgnoreErrors(utils.RunID), utils.MetaDefault)
	eReq := esapi.IndexRequest{
		Index:      eEe.index,
		DocumentID: utils.ConcatenatedKey(cgrID, runID),
		Body:       strings.NewReader(utils.ToJSON(valMp)),
		Refresh:    "true",
	}
	var resp *esapi.Response
	if resp, err = eReq.Do(context.Background(), eEe.eClnt); err != nil {
		resp.Body.Close()
		return
	} else if resp.IsError() {
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

func (eEe *ElasticEe) GetMetrics() utils.MapStorage {
	return eEe.dc.Clone()
}
