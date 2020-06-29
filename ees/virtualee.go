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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewVirtualExporter(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS,
	dc utils.MapStorage) (vEe *VirtualEe, err error) {
	dc[utils.ExportID] = cgrCfg.EEsCfg().Exporters[cfgIdx].ID
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
	defer vEe.Unlock()

	vEe.dc[utils.NumberOfEvents] = vEe.dc[utils.NumberOfEvents].(int) + 1

	req := utils.MapStorage{}
	for k, v := range cgrEv.Event {
		req[k] = v
	}
	eeReq := NewEventExporterRequest(req, vEe.dc, cgrEv.Tenant, vEe.cgrCfg.GeneralCfg().DefaultTimezone,
		vEe.filterS)
	if err = eeReq.SetFields(vEe.cgrCfg.EEsCfg().Exporters[vEe.cfgIdx].ContentFields()); err != nil {
		vEe.dc[utils.NegativeExports].(utils.StringSet).Add(cgrEv.ID)
		return
	}
	if aTime, err := cgrEv.FieldAsTime(utils.AnswerTime, vEe.cgrCfg.GeneralCfg().DefaultTimezone); err == nil {
		if vEe.dc[utils.FirstEventATime].(time.Time).IsZero() || vEe.dc[utils.FirstEventATime].(time.Time).Before(aTime) {
			vEe.dc[utils.FirstEventATime] = aTime
		}
		if aTime.After(vEe.dc[utils.LastEventATime].(time.Time)) {
			vEe.dc[utils.LastEventATime] = aTime
		}
	}
	if oID, err := cgrEv.FieldAsInt64(utils.OrderID); err == nil {
		if vEe.dc[utils.FirstExpOrderID].(int64) > oID || vEe.dc[utils.FirstExpOrderID].(int64) == 0 {
			vEe.dc[utils.FirstExpOrderID] = oID
		}
		if vEe.dc[utils.LastExpOrderID].(int64) < oID {
			vEe.dc[utils.LastExpOrderID] = oID
		}
	}
	if cost, err := cgrEv.FieldAsFloat64(utils.Cost); err == nil {
		vEe.dc[utils.TotalCost] = vEe.dc[utils.TotalCost].(float64) + cost
	}
	if tor, err := cgrEv.FieldAsString(utils.ToR); err == nil {
		if usage, err := cgrEv.FieldAsDuration(utils.Usage); err == nil {
			switch tor {
			case utils.VOICE:
				vEe.dc[utils.TotalDuration] = vEe.dc[utils.TotalDuration].(time.Duration) + usage
			case utils.SMS:
				vEe.dc[utils.TotalSMSUsage] = vEe.dc[utils.TotalSMSUsage].(time.Duration) + usage
			case utils.MMS:
				vEe.dc[utils.TotalMMSUsage] = vEe.dc[utils.TotalMMSUsage].(time.Duration) + usage
			case utils.GENERIC:
				vEe.dc[utils.TotalGenericUsage] = vEe.dc[utils.TotalGenericUsage].(time.Duration) + usage
			case utils.DATA:
				vEe.dc[utils.TotalDataUsage] = vEe.dc[utils.TotalDataUsage].(time.Duration) + usage
			}
		}
	}
	vEe.dc[utils.PositiveExports].(utils.StringSet).Add(cgrEv.ID)
	return
}
