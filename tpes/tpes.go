// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package tpes

import (
	"archive/zip"
	"bytes"
	"io"
	"slices"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewTPeS(cfg *config.CGRConfig, dm *engine.DataManager, cm *engine.ConnManager) *TPeS {
	return &TPeS{
		cfg:     cfg,
		connMgr: cm,
		dm:      dm,
	}
}

// TPeS is managing the TariffPlanExporter
type TPeS struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
	dm      *engine.DataManager
}

type ArgsExportTP struct {
	Tenant      string
	APIOpts     map[string]any
	ExportItems map[string][]string // map[exportType][]string{"itemID1", "itemID2"}
}

func getTariffPlansKeys(ctx *context.Context, dm *engine.DataManager, tenant, exportType string) (profileIDs []string, err error) {
	exporter, ok := exporters[exportType]
	if !ok {
		return nil, utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, exportType)
	}
	prefix := exporter.keyPrefix + tenant + utils.ConcatenatedKeySep
	// dbKeys are full keys (for example "alp_cgrates.org:ATTR_1"); return only profile IDs.
	db, _, err := dm.DBConns().GetConn(exporter.dbItemID)
	if err != nil {
		return nil, err
	}
	var dbKeys []string
	if dbKeys, err = db.GetKeysForPrefix(ctx, prefix, utils.EmptyString); err != nil {
		return nil, err
	}
	profileIDs = make([]string, 0, len(dbKeys))
	for _, key := range dbKeys {
		profileIDs = append(profileIDs, key[len(prefix):])
	}
	slices.Sort(profileIDs)
	return
}

// V1ExportTariffPlan is the API executed to export tariff plan items
func (tpE *TPeS) V1ExportTariffPlan(ctx *context.Context, args *ArgsExportTP, reply *[]byte) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = tpE.cfg.GeneralCfg().DefaultTenant
	}
	// Empty ExportItems means export all known tariff-plan items.
	if len(args.ExportItems) == 0 {
		args.ExportItems = make(map[string][]string)
		for exportType := range exporters {
			var itemIDs []string
			if itemIDs, err = getTariffPlansKeys(ctx, tpE.dm, args.Tenant, exportType); err != nil {
				return
			} else if len(itemIDs) != 0 {
				args.ExportItems[exportType] = itemIDs
			}
		}
	} else {
		for exportType := range args.ExportItems {
			if _, has := exporters[exportType]; !has {
				return utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, exportType)
			}
		}
	}
	buff := new(bytes.Buffer)
	zBuff := zip.NewWriter(buff)
	for exportType, itemIDs := range args.ExportItems {
		if len(itemIDs) == 0 {
			continue
		}
		exporter := exporters[exportType]
		var writer io.Writer
		if writer, err = zBuff.CreateHeader(&zip.FileHeader{
			Method:   zip.Deflate,
			Name:     exporter.fileName,
			Modified: time.Now(),
		}); err != nil {
			return
		}
		if err = exporter.exportItems(ctx, tpE.dm, exportType, writer, args.Tenant, itemIDs); err != nil {
			return utils.NewErrServerError(err)
		}
	}
	if err = zBuff.Close(); err != nil {
		return err
	}
	*reply = buff.Bytes()
	return
}
