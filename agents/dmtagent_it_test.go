/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package agents

import (
	"flag"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var testIntegration = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args
var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

var daCfgPath string
var daCfg *config.CGRConfig

func TestDmtAgentInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	daCfgPath = path.Join(*dataDir, "conf", "samples", "dmtagent")
	// Init config first
	var err error
	daCfg, err = config.NewCGRConfigFromFolder(daCfgPath)
	if err != nil {
		t.Error(err)
	}
	daCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(daCfg)
}

// Remove data in both rating and accounting db
func TestDmtAgentResetDataDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestDmtAgentResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestDmtAgentStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func TestDmtAgentSendCCR(t *testing.T) {
	if !*testIntegration {
		return
	}
	dmtClient, err := NewDiameterClient(daCfg.DiameterAgentCfg().Listen, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm,
		daCfg.DiameterAgentCfg().VendorId, daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, daCfg.DiameterAgentCfg().DictionariesDir)
	if err != nil {
		t.Fatal(err)
	}
	cdr := &engine.StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE,
		AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002", Supplier: "SUPPL1",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), MediationRunId: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, Pdd: time.Duration(7) * time.Second, ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"}, Cost: 1.01, RatedAccount: "dan", RatedSubject: "dans", Rated: true,
	}
	ccr := storedCdrToCCR(cdr, "UNIT_TEST", daCfg.DiameterAgentCfg().OriginRealm, daCfg.DiameterAgentCfg().VendorId,
		daCfg.DiameterAgentCfg().ProductName, utils.DIAMETER_FIRMWARE_REVISION, time.Duration(300)*time.Second, true)
	m, err := ccr.AsDiameterMessage()
	if err != nil {
		t.Error(err)
	}
	if err := dmtClient.SendMessage(m); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(10) * time.Second)
}

func TestDmtAgentStopEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
