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
	"bytes"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPTrends(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetTrendProfileDrvF: func(ctx *context.Context, tnt string, id string) (*engine.TrendProfile, error) {
			trd := &engine.TrendProfile{
				Tenant: "cgrates.org",
				ID:     "TRD_2",
			}
			return trd, nil
		},
	}, nil, connMng)
	exp := &TPThresholds{
		dm: dm,
	}
	rcv := newTPTrends(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportTrends(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	tpTrd := TPTrends{
		dm: dm,
	}
	trd := &engine.TrendProfile{
		Tenant: "cgrates.org",
		ID:     "TRD_2",
	}
	tpTrd.dm.SetTrendProfile(context.Background(), trd)
	err := tpTrd.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TRD_2"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsTrendsNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpTrd := TPTrends{
		dm: nil,
	}
	trd := &engine.TrendProfile{
		Tenant: "cgrates.org",
		ID:     "TRD_2",
	}
	tpTrd.dm.SetTrendProfile(context.Background(), trd)
	err := tpTrd.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TRD_2"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsTrendsIDNotFound(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	tpTrd := TPTrends{
		dm: dm,
	}
	trd := &engine.TrendProfile{
		Tenant: "cgrates.org",
		ID:     "TRD_2",
	}
	tpTrd.dm.SetTrendProfile(context.Background(), trd)
	err := tpTrd.exportItems(context.Background(), wrtr, "cgrates.org", []string{"TRD_3"})
	errExpect := "<NOT_FOUND> cannot find TrendProfile with id: <TRD_3>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
