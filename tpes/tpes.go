/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package tpes

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewTPeS(cfg *config.CGRConfig, dm *engine.DataManager, cm *engine.ConnManager) (tpE *TPeS) {
	tpE = &TPeS{
		cfg:     cfg,
		connMgr: cm,
		dm:      dm,
		exps:    make(map[string]tpExporter),
	}

	var err error
	for expType := range tpExporterTypes {
		if tpE.exps[expType], err = newTPExporter(expType, dm); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> cannot create exporter of type <%s>", utils.TPeS, expType))
		}
	}
	return
}

// TPeS is managing the TariffPlanExporter
type TPeS struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
	dm      *engine.DataManager
	fltr    *engine.FilterS
	exps    map[string]tpExporter
}

type ArgsExportTP struct {
	Tenant      string
	APIOpts     map[string]any
	ExportItems map[string][]string // map[expType][]string{"itemID1", "itemID2"}
}

func getTariffPlansKeys(ctx *context.Context, dm *engine.DataManager, tnt, expType string) (profileIDs []string, err error) {
	var itemID string
	var prfx string
	switch expType {
	case utils.MetaAttributes:
		prfx = utils.AttributeProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaAttributeProfiles
	case utils.MetaActions:
		prfx = utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaActionProfiles
	case utils.MetaAccounts:
		prfx = utils.AccountPrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaAccounts
	case utils.MetaChargers:
		prfx = utils.ChargerProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaChargerProfiles
	case utils.MetaFilters:
		prfx = utils.FilterPrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaFilters
	case utils.MetaRates:
		prfx = utils.RateProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaRateProfiles
	case utils.MetaResources:
		prfx = utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaResourceProfiles
	case utils.MetaRoutes:
		prfx = utils.RouteProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaRouteProfiles
	case utils.MetaStats:
		prfx = utils.StatQueueProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaStatQueueProfiles
	case utils.MetaThresholds:
		prfx = utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaThresholdProfiles
	case utils.MetaRankings:
		prfx = utils.RankingProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaRankingProfiles
	case utils.MetaTrends:
		prfx = utils.TrendProfilePrefix + tnt + utils.ConcatenatedKeySep
		itemID = utils.MetaTrendProfiles
	default:
		return nil, fmt.Errorf("Unsuported exporter type")
	}
	// dbKeys will contain the full name of the key, but we will need just the IDs e.g. "alp_cgrates.org:ATTR_1" -- just ATTR_1
	dataDB, _, err := dm.DBConns().GetConn(itemID)
	if err != nil {
		return nil, err
	}
	var dbKeys []string
	if dbKeys, err = dataDB.GetKeysForPrefix(ctx, prfx); err != nil {
		return nil, err
	}
	profileIDs = make([]string, 0, len(dbKeys))
	for _, key := range dbKeys {
		profileIDs = append(profileIDs, key[len(prfx):])
	}
	return
}

// V1ExportTariffPlan is the API executed to export tariff plan items
func (tpE *TPeS) V1ExportTariffPlan(ctx *context.Context, args *ArgsExportTP, reply *[]byte) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = tpE.cfg.GeneralCfg().DefaultTenant
	}
	/*
	  IMPORTANT!!
	*/
	// in case the export items are empty, export all tariffplans for every subsystem from database in zip format and containing CSV files
	if len(args.ExportItems) == 0 {
		args.ExportItems = make(map[string][]string)
		for subsystem := range tpExporterTypes {
			var itemIDs []string
			if itemIDs, err = getTariffPlansKeys(ctx, tpE.dm, args.Tenant, subsystem); err != nil {
				return
			} else if len(itemIDs) != 0 {
				// the map e.g. : *filters: {"ATTR_1", "ATTR_1"}
				args.ExportItems[subsystem] = itemIDs
			}
		}
	} else {
		// else export just the wanted IDs
		for eType := range args.ExportItems {
			if _, has := tpE.exps[eType]; !has {
				return utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, eType)
			}
		}
	}
	buff := new(bytes.Buffer)
	zBuff := zip.NewWriter(buff)
	for expType, expItms := range args.ExportItems {
		// if there are not items to be exported, continue with the next subsystem
		if len(expItms) == 0 {
			continue
		}
		var wrtr io.Writer
		// here we will create all the header for each subsystem type for the csv
		if wrtr, err = zBuff.CreateHeader(&zip.FileHeader{
			Method:   zip.Deflate, // to be compressed
			Name:     exportFileName[expType],
			Modified: time.Now(),
		}); err != nil {
			return
		}
		// our buffer will contain the bytes with all profiles in CSV format
		if err = tpE.exps[expType].exportItems(ctx, wrtr, args.Tenant, expItms); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	if err = zBuff.Close(); err != nil {
		return err
	}
	*reply = buff.Bytes()
	return
}
