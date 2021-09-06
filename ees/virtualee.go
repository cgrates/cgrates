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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewVirtualEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *VirtualEE {
	return &VirtualEE{
		cfg: cfg,
		dc:  dc,
	}
}

// VirtualEE implements EventExporter interface for .csv files
type VirtualEE struct {
	cfg *config.EventExporterCfg
	dc  *utils.SafeMapStorage
}

func (vEe *VirtualEE) Cfg() *config.EventExporterCfg                          { return vEe.cfg }
func (vEe *VirtualEE) Connect() error                                         { return nil }
func (vEe *VirtualEE) ExportEvent(interface{}, string) error                  { return nil }
func (vEe *VirtualEE) Close() error                                           { return nil }
func (vEe *VirtualEE) GetMetrics() *utils.SafeMapStorage                      { return vEe.dc }
func (vEe *VirtualEE) PrepareMap(map[string]interface{}) (interface{}, error) { return nil, nil }
func (vEe *VirtualEE) PrepareOrderMap(*utils.OrderedNavigableMap) (interface{}, error) {
	return nil, nil
}
