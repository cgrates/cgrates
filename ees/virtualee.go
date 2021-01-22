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
)

func NewVirtualExporter(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (vEe *VirtualEe, err error) {
	vEe = &VirtualEe{id: cgrCfg.EEsCfg().Exporters[cfgIdx].ID,
		cgrCfg: cgrCfg, cfgIdx: cfgIdx, filterS: filterS, dc: dc}
	err = vEe.init()
	return
}

// VirtualEe implements EventExporter interface for .csv files
type VirtualEe struct {
	id      string
	cgrCfg  *config.CGRConfig
	cfgIdx  int // index of config instance within ERsCfg.Readers
	filterS *engine.FilterS
	sync.RWMutex
	dc utils.MapStorage
}

// init will create all the necessary dependencies, including opening the file
func (vEe *VirtualEe) init() (err error) {
	return
}

// ID returns the identificator of this exporter
func (vEe *VirtualEe) ID() string {
	return vEe.id
}

// OnEvicted implements EventExporter, doing the cleanup before exit
func (vEe *VirtualEe) OnEvicted(_ string, _ interface{}) {
	return
}

// ExportEvent implements EventExporter
func (vEe *VirtualEe) ExportEvent(cgrEv *utils.CGREvent) (err error) {
	vEe.Lock()
	defer func() {
		if err != nil {
			vEe.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		} else {
			vEe.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
		}
		vEe.Unlock()
	}()
	vEe.dc[utils.NumberOfEvents] = vEe.dc[utils.NumberOfEvents].(int64) + 1

	req := utils.MapStorage(cgrEv.Event)
	oNm := map[string]*utils.OrderedNavigableMap{
		utils.MetaExp: utils.NewOrderedNavigableMap(),
	}
	eeReq := engine.NewEventRequest(req, vEe.dc, cgrEv.Opts,
		vEe.cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Tenant,
		vEe.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(vEe.cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Timezone,
			vEe.cgrCfg.GeneralCfg().DefaultTimezone),
		vEe.filterS, oNm)
	if err = eeReq.SetFields(vEe.cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].ContentFields()); err != nil {
		return
	}
	updateEEMetrics(vEe.dc, cgrEv.Event, utils.FirstNonEmpty(vEe.cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].Timezone,
		vEe.cgrCfg.GeneralCfg().DefaultTimezone))
	return
}

func (vEe *VirtualEe) GetMetrics() utils.MapStorage {
	return vEe.dc.Clone()
}
