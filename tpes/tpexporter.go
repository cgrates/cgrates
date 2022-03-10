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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tpExporterTypes = utils.NewStringSet([]string{
	utils.MetaAttributes,
	utils.MetaResources,
	/*
		utils.MetaFilters,
		utils.MetaStats,
		utils.MetaThresholds,
		utils.MetaRoutes,
		utils.MetaChargers,
		utils.MetaDispatchers,
		utils.MetaDispatcherHosts,
		utils.MetaRateProfiles,
		utils.MetaActions,
		utils.MetaAccounts */})

// tpExporter is the interface implementing exports of tariff plan items
type tpExporter interface {
	exportItems(ctx *context.Context, tnt string, itmIDs []string) (expContent []byte, err error)
}

// newTPExporter constructs tpExporters
func newTPExporter(expType string, dm *engine.DataManager) (tpE tpExporter, err error) {
	switch expType {
	case utils.MetaAttributes:
		return newTPAttributes(dm), nil
	case utils.MetaResources:
		return newTPResources(dm), nil
	default:
		return nil, utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, expType)
	}
}

func getFileName(subsystem string) string {
	switch subsystem {
	case utils.MetaAttributes:
		return utils.AttributesCsv
	case utils.MetaResources:
		return utils.ResourcesCsv
	default:
		return utils.EmptyString
	}
}

/*
func writeOut(tpData []interface{}) (csvBts []byte, err error) {
	if len(tpData) == 0 {
		return
	}
	buff := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buff)
	csvWriter.Comma = utils.CSVSep
	for _, tpItem := range tpData {
		record, err := engine.CsvDump(tpItem)
		if err != nil {
			return nil, err
		}
		if err := csvWriter.Write(record); err != nil {
			return nil, err
		}
	}

	csvWriter.Flush()
	return buff.Bytes(), nil
}
*/
