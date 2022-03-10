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
	APIOpts     map[string]interface{}
	ExportItems map[string][]string // map[expType][]string{"itemID1", "itemID2"}
}

// V1ExportTariffPlan is the API executed to export tariff plan items
func (tpE *TPeS) V1ExportTariffPlan(ctx *context.Context, args *ArgsExportTP, reply *[]byte) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = tpE.cfg.GeneralCfg().DefaultTenant
	}
	for eType := range args.ExportItems {
		if _, has := tpE.exps[eType]; !has {
			return utils.ErrPrefix(utils.ErrUnsupportedTPExporterType, eType)
		}
	}

	buff := new(bytes.Buffer)
	zBuff := zip.NewWriter(buff)
	for expType, expItms := range args.ExportItems {
		var wrtr io.Writer
		//here we will create all the header for each subsystem type for the csv
		if wrtr, err = zBuff.CreateHeader(&zip.FileHeader{
			Method:   zip.Deflate, // to be compressed
			Name:     getFileName(expType),
			Modified: time.Now(),
		}); err != nil {
			return
		}
		var expBts []byte
		// expBts will containt the bytes with all profiles in CSV format
		if expBts, err = tpE.exps[expType].exportItems(ctx, args.Tenant, expItms); err != nil {
			return utils.NewErrServerError(err)
		}
		// write all the bytes into our zip
		if _, err = wrtr.Write(expBts); err != nil {
			return
		}
	}
	if err = zBuff.Close(); err != nil {
		return err
	}
	*reply = buff.Bytes()
	return
}
