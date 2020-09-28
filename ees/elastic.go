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
	"sync"

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
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	sync.RWMutex
	dc utils.MapStorage
}

// init will create all the necessary dependencies, including opening the file
func (eEe *ElasticEe) init() (err error) {
	// compose the config out of opts
	// create the client
	if eEe.eClnt, err = elasticsearch.NewDefaultClient(); err != nil {
		return
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
	eeReq := NewEventExporterRequest(req, eEe.dc,
		eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Tenant,
		eEe.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Timezone,
			eEe.cgrCfg.GeneralCfg().DefaultTimezone),
		eEe.filterS)
	if err = eeReq.SetFields(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].ContentFields()); err != nil {
		return
	}
	updateEEMetrics(eEe.dc, cgrEv.Event, utils.FirstNonEmpty(eEe.cgrCfg.EEsCfg().Exporters[eEe.cfgIdx].Timezone,
		eEe.cgrCfg.GeneralCfg().DefaultTimezone))
	return
}

func (eEe *ElasticEe) GetMetrics() utils.MapStorage {
	return eEe.dc.Clone()
}
