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

/*
import (
	"os/exec"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var cgrRater1Cmd, cgrSmg1Cmd *exec.Cmd

func TestHaPoolInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	daCfgPath = path.Join(*dataDir, "conf", "samples", "hapool", "cgrrater1")
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
func TestHaPoolResetDataDb(t *testing.T) {
	TestDmtAgentResetDataDb(t)
}

// Wipe out the cdr database
func TestHaPoolResetStorDb(t *testing.T) {
	TestDmtAgentResetStorDb(t)
}

// Start CGR Engine
func TestHaPoolStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	engine.KillEngine(*waitRater) // just to make sure
	var err error
	cgrRater1 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrrater1")
	if cgrRater1Cmd, err = engine.StartEngine(cgrRater1, *waitRater); err != nil {
		t.Fatal("cgrRater1: ", err)
	}
	cgrRater2 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrrater2")
	if _, err = engine.StartEngine(cgrRater2, *waitRater); err != nil {
		t.Fatal("cgrRater2: ", err)
	}
	cgrSmg1 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrsmg1")
	if cgrSmg1Cmd, err = engine.StartEngine(cgrSmg1, *waitRater); err != nil {
		t.Fatal("cgrSmg1: ", err)
	}
	cgrSmg2 := path.Join(*dataDir, "conf", "samples", "hapool", "cgrsmg2")
	if _, err = engine.StartEngine(cgrSmg2, *waitRater); err != nil {
		t.Fatal("cgrSmg2: ", err)
	}
	cgrDa := path.Join(*dataDir, "conf", "samples", "hapool", "dagent")
	if _, err = engine.StartEngine(cgrDa, *waitRater); err != nil {
		t.Fatal("cgrDa: ", err)
	}

}

// Connect rpc client to rater
func TestHaPoolApierRpcConn(t *testing.T) {
	TestDmtAgentApierRpcConn(t)
}

// Load the tariff plan, creating accounts and their balances
func TestHaPoolTPFromFolder(t *testing.T) {
	TestDmtAgentTPFromFolder(t)
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:47:26Z"'
func TestHaPoolSendCCRInit(t *testing.T) {
	TestDmtAgentSendCCRInit(t)
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:52:26Z"'
func TestHaPoolSendCCRUpdate(t *testing.T) {
	TestDmtAgentSendCCRUpdate(t)
}

// cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2015-11-07T08:42:26Z" TimeEnd="2015-11-07T08:57:26Z"'
func TestHaPoolSendCCRUpdate2(t *testing.T) {
	TestDmtAgentSendCCRUpdate2(t)
}

func TestHaPoolSendCCRTerminate(t *testing.T) {
	TestDmtAgentSendCCRTerminate(t)
}

func TestHaPoolCdrs(t *testing.T) {
	TestDmtAgentCdrs(t)
}

func TestHaPoolStopEngine(t *testing.T) {
	TestDmtAgentStopEngine(t)
}
*/
