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

package actions

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestACHTTPPostExecute(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().PosterAttempts = 1
	apAction := &engine.APAction{
		ID:   "TEST_ACTION_HTTPPOST",
		Type: utils.CDRLog,
		Diktats: []*engine.APDiktat{
			{
				Path:  "~*balance.TestBalance.Value",
				Value: "10",
			},
		},
	}
	http := newActHTTPPost(cfg, apAction)

	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]interface{}{
			utils.Usage: 10 * time.Minute,
		},
	}

	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)

	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	expected := `<EEs> Exporter <TEST_ACTION_HTTPPOST> could not export because err: <Post "~*balance.TestBalance.Value": unsupported protocol scheme "">`
	if err := http.execute(nil, dataStorage, utils.EmptyString); err != nil {
		t.Error(err)
	} else if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	buff.Reset()

	// channels cannot be marshaled
	dataStorage[utils.MetaReq] = make(chan struct{}, 1)
	expected = "json: unsupported type: chan struct {}"
	if err := http.execute(nil, dataStorage, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	dataStorage = utils.MapStorage{
		utils.MetaOpts: map[string]interface{}{
			utils.MetaAsync: 10 * time.Minute,
		},
	}
	http.aCfg.Opts = make(map[string]interface{})
	http.aCfg.Opts[utils.MetaAsync] = true
	http.config.GeneralCfg().FailedPostsDir = utils.MetaNone
	if err := http.execute(nil, dataStorage, utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestACHTTPPostValues(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().PosterAttempts = 1
	cfg.GeneralCfg().FailedPostsDir = utils.MetaNone
	apAction := &engine.APAction{
		ID:   "TEST_ACTION_HTTPPostValues",
		Type: utils.MetaHTTPPost,
		Diktats: []*engine.APDiktat{
			{
				Path:  "~*balance.TestBalance.Value",
				Value: "80",
			},
		},
	}
	http := newActHTTPPost(cfg, apAction)
	dataStorage := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: 1003,
		},
	}

	if err := http.execute(nil, dataStorage,
		utils.EmptyString); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}
}

func TestACHTTPPostID(t *testing.T) {
	apAction := &engine.APAction{
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
