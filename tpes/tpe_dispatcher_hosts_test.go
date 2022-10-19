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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestTPEnewTPDispatchersHost(t *testing.T) {
	// dataDB := &engine.DataDBM
	// dm := &engine.NewDataManager()
	cfg := config.NewDefaultCGRConfig()
	connMng := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(&engine.DataDBMock{
		GetDispatcherHostDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.DispatcherHost, error) {
			dsph := &engine.DispatcherHost{
				Tenant: "cgrates.org",
				RemoteHost: &config.RemoteHost{
					ID:              "DSH1",
					Address:         "*internal",
					ConnectAttempts: 1,
					Reconnects:      3,
					ConnectTimeout:  time.Minute,
					ReplyTimeout:    2 * time.Minute,
				},
			}
			return dsph, nil
		},
	}, nil, connMng)
	exp := &TPDispatcherHosts{
		dm: dm,
	}
	rcv := newTPDispatcherHosts(dm)
	if rcv.dm != exp.dm {
		t.Errorf("Expected %v \nbut received %v", exp, rcv)
	}
}

func TestTPEExportItemsDispatchersHost(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	tpDsph := TPDispatcherHosts{
		dm: dm,
	}
	dsph := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:              "DSH1",
			Address:         "*internal",
			ConnectAttempts: 1,
			Reconnects:      3,
			ConnectTimeout:  time.Minute,
			ReplyTimeout:    2 * time.Minute,
		},
	}
	tpDsph.dm.SetDispatcherHost(context.Background(), dsph)
	err := tpDsph.exportItems(context.Background(), wrtr, "cgrates.org", []string{"DSH1"})
	if err != nil {
		t.Errorf("Expected nil\n but received %v", err)
	}
}

func TestTPEExportItemsDispatcherHostsNoDbConn(t *testing.T) {
	engine.Cache.Clear(nil)
	wrtr := new(bytes.Buffer)
	tpDsph := TPDispatcherHosts{
		dm: nil,
	}
	dsph := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:              "DSH1",
			Address:         "*internal",
			ConnectAttempts: 1,
			Reconnects:      3,
			ConnectTimeout:  time.Minute,
			ReplyTimeout:    2 * time.Minute,
		},
	}
	tpDsph.dm.SetDispatcherHost(context.Background(), dsph)
	err := tpDsph.exportItems(context.Background(), wrtr, "cgrates.org", []string{"DSH1"})
	if err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %v\n but received %v", utils.ErrNoDatabaseConn, err)
	}
}

func TestTPEExportItemsDispatchersIDNotFoundHost(t *testing.T) {
	wrtr := new(bytes.Buffer)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	tpDsph := TPDispatcherHosts{
		dm: dm,
	}
	dsph := &engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:              "DSH1",
			Address:         "*internal",
			ConnectAttempts: 1,
			Reconnects:      3,
			ConnectTimeout:  time.Minute,
			ReplyTimeout:    2 * time.Minute,
		},
	}
	tpDsph.dm.SetDispatcherHost(context.Background(), dsph)
	err := tpDsph.exportItems(context.Background(), wrtr, "cgrates.org", []string{"DSH2"})
	errExpect := "<DSP_HOST_NOT_FOUND> cannot find DispatcherHost with id: <DSH2>"
	if err.Error() != errExpect {
		t.Errorf("Expected %v\n but received %v", errExpect, err)
	}
}
