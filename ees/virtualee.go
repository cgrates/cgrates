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

func NewVirtualEE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) *VirtualEE {
	return &VirtualEE{
		cfg: cfg,
		em:  em,
	}
}

// VirtualEE implements EventExporter interface for .csv files
type VirtualEE struct {
	cfg *config.EventExporterCfg
	em  *utils.ExporterMetrics
}

func (vEe *VirtualEE) Cfg() *config.EventExporterCfg { return vEe.cfg }
func (vEe *VirtualEE) Connect() error                { return nil }

func (vEe *VirtualEE) ExportEvent(payload any, _ string) error {
	utils.Logger.Info(
		fmt.Sprintf("<%s> <%s> exported: <%s>",
			utils.EEs, vEe.Cfg().ID, utils.ToJSON(payload)))
	return nil
}

func (vEe *VirtualEE) Close() error                       { return nil }
func (vEe *VirtualEE) GetMetrics() *utils.ExporterMetrics { return vEe.em }

func (vEe *VirtualEE) PrepareMap(cgrEv *utils.CGREvent) (any, error) {
	return cgrEv.Event, nil
}

func (vEe *VirtualEE) PrepareOrderMap(onm *utils.OrderedNavigableMap) (any, error) {
	preparedMap := make(map[string]any)
	for el := onm.GetFirstElement(); el != nil; el = el.Next() {
		path := el.Value
		item, _ := onm.Field(path)
		path = path[:len(path)-1] // remove the last index
		preparedMap[strings.Join(path, utils.NestingSep)] = item.String()
	}
	return preparedMap, nil
}
