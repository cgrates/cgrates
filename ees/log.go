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

package ees

import (
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewLogEE(cfg *config.EventExporterCfg, dc *utils.ExporterMetrics) *LogEE {
	return &LogEE{
		cfg: cfg,
		dc:  dc,
	}
}

// LogEE implements EventExporter interface for .csv files
type LogEE struct {
	cfg *config.EventExporterCfg
	dc  *utils.ExporterMetrics
}

func (vEe *LogEE) Cfg() *config.EventExporterCfg { return vEe.cfg }
func (vEe *LogEE) Connect() error                { return nil }
func (vEe *LogEE) ExportEvent(mp any, _ string) error {
	utils.Logger.Info(
		fmt.Sprintf("<%s> <%s> exported: <%s>",
			utils.EEs, vEe.Cfg().ID, utils.ToJSON(mp)))
	return nil
}
func (vEe *LogEE) Close() error                       { return nil }
func (vEe *LogEE) GetMetrics() *utils.ExporterMetrics { return vEe.dc }
func (vEe *LogEE) PrepareMap(mp *utils.CGREvent) (any, error) {
	return mp.Event, nil
}
func (vEe *LogEE) PrepareOrderMap(mp *utils.OrderedNavigableMap) (any, error) {
	valMp := make(map[string]any)
	for el := mp.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		nmIt, _ := mp.Field(path)
		path = path[:len(path)-1] // remove the last index
		valMp[strings.Join(path, utils.NestingSep)] = nmIt.String()
	}
	return valMp, nil
}
