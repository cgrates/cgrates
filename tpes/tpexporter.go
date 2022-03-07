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

package tpes

import (
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tpExporterTypes = utils.NewStringSet([]string{utils.MetaAttributes, utils.MetaResources, utils.MetaFilters, utils.MetaStats,
	utils.MetaThresholds, utils.MetaRoutes, utils.MetaChargers, utils.MetaDispatchers, utils.MetaDispatcherHosts,
	utils.MetaRateProfiles, utils.MetaActions, utils.MetaAccounts})

// tpExporter is the interface implementing exports of tariff plan items
type tpExporter interface {
	exportItems(itmIDs []string) (expContent []byte, err error)
}

// newTPExporter constructs tpExporters
func newTPExporter(expType string, dm *engine.DataManager) (tpE tpExporter, err error) {
	switch expType {
	default:
		return nil, utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, expType)
	}
	return
}
