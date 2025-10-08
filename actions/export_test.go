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

package actions

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestACHTTPPostExecute(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.EFsCfg().PosterAttempts = 1
	apAction := &utils.APAction{
		ID:   "TEST_ACTION_HTTPPOST",
		Type: utils.CDRLog,
		Diktats: []*utils.APDiktat{
			{
				Opts: map[string]any{
					"*url": "~*balance.TestBalance.Value",
				},
			},
		},
	}
	http, err := newActHTTPPost(context.Background(), "cgrates.org", new(utils.CGREvent),
		new(engine.FilterS), cfg, apAction)
	if err != nil {
		t.Error(err)
	}

	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]any{
			utils.Usage: 10 * time.Minute,
		},
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	expected := `<EEs> Exporter <TEST_ACTION_HTTPPOST> could not export because err: <Post "~*balance.TestBalance.Value": unsupported protocol scheme "">`
	if err := http.execute(context.Background(), dataStorage, utils.EmptyString); err != nil {
		t.Error(err)
	} else if rcv := buf.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	// channels cannot be marshaled
	dataStorage[utils.MetaReq] = make(chan struct{}, 1)
	expected = "json: unsupported type: chan struct {}"
	if err := http.execute(context.Background(), dataStorage, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	dataStorage = utils.MapStorage{
		utils.MetaOpts: map[string]any{
			utils.MetaAsync: 10 * time.Minute,
		},
	}
	http.aCfg.Opts = make(map[string]any)
	http.aCfg.Opts[utils.MetaAsync] = true
	http.config.EFsCfg().FailedPostsDir = utils.MetaNone
	if err := http.execute(context.Background(), dataStorage, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestACHTTPPostValues(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.EEsCfg().ExporterCfg(utils.MetaDefault).Attempts = 1
	cfg.EEsCfg().ExporterCfg(utils.MetaDefault).FailedPostsDir = utils.MetaNone
	apAction := &utils.APAction{
		ID:   "TEST_ACTION_HTTPPostValues",
		Type: utils.MetaHTTPPost,
		Diktats: []*utils.APDiktat{
			{
				Opts: map[string]any{
					"*balancePath": "~*balance.TestBalance.Value",
				},
			},
		},
	}
	http, err := newActHTTPPost(context.Background(), "cgrates.org", new(utils.CGREvent),
		new(engine.FilterS), cfg, apAction)
	if err != nil {
		t.Error(err)
	}
	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]any{
			utils.AccountField: 1003,
		},
	}

	if err := http.execute(context.Background(), dataStorage,
		utils.EmptyString); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestACHTTPPostID(t *testing.T) {
	apAction := &utils.APAction{
		ID:   "TestACHTTPPostID",
		Type: utils.MetaHTTPPost,
	}
	http := &actHTTPPost{
		aCfg: apAction,
	}
	if rcv := http.id(); rcv != apAction.ID {
		t.Errorf("Expected %+v, received %+v", apAction.ID, rcv)
	}

	actESP := &actExport{
		aCfg: apAction,
	}
	if rcv := actESP.id(); rcv != apAction.ID {
		t.Errorf("Expected %+v, received %+v", apAction.ID, rcv)
	}
}
