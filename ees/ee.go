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
	GetMetrics() utils.MapStorage                  // called to get metrics
}

// NewEventExporter produces exporters
func NewEventExporter(cgrCfg *config.CGRConfig, cfgIdx int, filterS *engine.FilterS, dc utils.MapStorage) (ee EventExporter, err error) {
	switch cgrCfg.EEsCfg().Exporters[cfgIdx].Type {
	case utils.MetaFileCSV:
		return NewFileCSVee(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaFileFWV:
		return NewFileFWVee(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaHTTPPost:
		return NewHTTPPostEe(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaHTTPjsonMap, utils.MetaAMQPjsonMap, utils.MetaAMQPV1jsonMap, utils.MetaSQSjsonMap, utils.MetaKafkajsonMap, utils.MetaS3jsonMap:
		return NewHTTPJsonMapEe(cgrCfg, cfgIdx, filterS, dc)
	case utils.MetaVirt:
		return NewVirtualExporter(cgrCfg, cfgIdx, filterS, dc)
	default:
		return nil, fmt.Errorf("unsupported exporter type: <%s>", cgrCfg.EEsCfg().Exporters[cfgIdx].Type)
	}
}
