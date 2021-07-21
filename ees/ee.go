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
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type EventExporter interface {
	ID() string                                    // return the exporter identificator
	ExportEvent(cgrEv *utils.CGREvent) (err error) // called on each event to be exported
	OnEvicted(itmID string, value interface{})     // called when the exporter needs to terminate
	GetMetrics() *utils.SafeMapStorage             // called to get metrics
}

// NewEventExporter produces exporters
func NewEventExporter(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS) (ee EventExporter, err error) {
	var dc *utils.SafeMapStorage
	if dc, err = newEEMetrics(utils.FirstNonEmpty(
		cgrCfg.EEsCfg().Exporters[cfgIdx].Timezone,
		cgrCfg.GeneralCfg().DefaultTimezone)); err != nil {
		return
	}
	switch cgrCfg.EEsCfg().Exporters[cfgIdx].Type {
	case utils.MetaFileCSV:
		return NewFileCSVee(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaFileFWV:
		return NewFileFWVee(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaHTTPPost:
		return NewHTTPPostEe(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaHTTPjsonMap:
		return NewHTTPjsonMapEE(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaAMQPjsonMap, utils.MetaAMQPV1jsonMap,
		utils.MetaSQSjsonMap, utils.MetaKafkajsonMap,
		utils.MetaS3jsonMap, utils.MetaNatsjsonMap:
		return NewPosterJSONMapEE(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaVirt:
		return NewVirtualExporter(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaElastic:
		return NewElasticExporter(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaSQL:
		return NewSQLEe(cgrCfg, cfgIdx, filterS, dc)
	default:
		return nil, fmt.Errorf("unsupported exporter type: <%s>", cgrCfg.EEsCfg().Exporters[cfgIdx].Type)
	}
}

func newConcReq(limit int) (c *concReq) {
	c = &concReq{limit: limit}
	if limit > 0 {
		c.reqs = make(chan struct{}, limit)
		for i := 0; i < limit; i++ {
			c.reqs <- struct{}{}
		}
	}
	return
}

type concReq struct {
	reqs  chan struct{}
	limit int
}

func (c *concReq) get() {
	if c.limit > 0 {
		<-c.reqs
	}
}
func (c *concReq) done() {
	if c.limit > 0 {
		c.reqs <- struct{}{}
	}
}
