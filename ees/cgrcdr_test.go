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
package ees

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCgrCdrExportEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	newDM := engine.NewDataManager(dbCM, cfg, nil)
	filterS := engine.NewFilterS(cfg, nil, newDM)
	eeCfg := config.NewEventExporterCfg("CDR1", utils.MetaCgrcdr, "", utils.MetaNone, 1, nil)
	cgrcdrEe, err := NewEventExporter(eeCfg, cfg, filterS, nil, newDM)
	if err != nil {
		t.Error(err)
	}
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event",
		Event: map[string]any{
			utils.OrderID:      1,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "dsafdsaf",
			utils.OriginHost:   "192.168.1.1",
			utils.RequestType:  utils.MetaRated,
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:        10 * time.Second,
			utils.Cost:         2.34567,
			"ExtraFields":      map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
		APIOpts: map[string]any{
			utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
			utils.RunID:        utils.MetaDefault,
		},
	}
	val := cgrcdrEe.ExtraData(cgrEvent)
	if err := cgrcdrEe.ExportEvent(context.Background(), nil, val); err != nil {
		t.Error(err)
	}
	fltr, err := engine.NewFilterFromInline(cfg.GeneralCfg().DefaultTenant, "*string:~*opts.*cdrID:918692d9a7ec66e5ee66d07b57fbebc40ab40e61")
	if err != nil {
		t.Error(err)
	}
	cdr, err := newDM.GetCDRs(context.Background(), []*engine.Filter{fltr}, nil)
	if err != nil {
		t.Error(err)
	}
	if len(cdr) == 0 {
		t.Errorf("expected to generate a CDR")
	}

}
