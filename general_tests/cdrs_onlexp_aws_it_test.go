// +build integration

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

package general_tests

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
	amqpv1 "github.com/vcabbage/amqp"
)

var cdrsAWSCfgPath string
var cdrsAWSCfg *config.CGRConfig
var cdrsAWSRpc *rpcclient.RpcClient

func TestCDRsOnExpAWSInitConfig(t *testing.T) {
	var err error
	cdrsAWSCfgPath = path.Join(*dataDir, "conf", "samples", "cdrsonexpamqp")
	if cdrsAWSCfg, err = config.NewCGRConfigFromFolder(cdrsAWSCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

// InitDb so we can rely on count
func TestCDRsOnExpAWSInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsAWSCfg); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(cdrsAWSCfg.GeneralCfg().FailedPostsDir); err != nil {
		t.Fatal("Error removing folder: ", cdrsAWSCfg.GeneralCfg().FailedPostsDir, err)
	}

	if err := os.Mkdir(cdrsAWSCfg.GeneralCfg().FailedPostsDir, 0700); err != nil {
		t.Error(err)
	}

}

func TestCDRsOnExpAWSStartMasterEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsAWSCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestCDRsOnExpAWSProccesCDR(t *testing.T) {
	cdrsAWSRpc, err = rpcclient.NewRpcClient("tcp", cdrsAWSCfg.ListenCfg().RPCJSONListen, false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), "json", nil, false)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	testCdr1 := &engine.CDR{
		CGRID:       utils.Sha1("httpjsonrpc1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		ToR:         utils.VOICE,
		OriginID:    "httpjsonrpc1",
		OriginHost:  "192.168.1.1",
		Source:      "UNKNOWN",
		RequestType: utils.META_PSEUDOPREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
		AnswerTime:  time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage:       time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID,
		Cost:        1.201,
		PreRated:    true,
	}
	var reply string
	if err := cdrsAWSRpc.Call("CdrsV2.ProcessCdr", testCdr1, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func TestCDRsOnExpAWSAMQPData(t *testing.T) {
	// client, err :=
	amqpv1.Dial("amqp://guest:guest@localhost:5672/?queue_id=cgrates_cdrs")
	// if err != nil {
	// 	t.Fatal("Dialing AMQP server:", err)
	// }
	// defer client.Close()
	// // Open a session
	// session, err := client.NewSession()
	// if err != nil {
	// 	t.Fatal("Creating AMQP session:", err)
	// }
	// ctx := context.Background()
	// defer session.Close(ctx)

	// receiver, err := session.NewReceiver(amqpv1.LinkSourceAddress("/cgrates_cdrs"))
	// if err != nil {
	// 	t.Fatal("Creating receiver link:", err)
	// }
	// defer func() {
	// 	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	// 	receiver.Close(ctx)
	// 	cancel()
	// }()

	// ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	// // Receive next message
	// msg, err := receiver.Receive(ctx)
	// if err != nil {
	// 	t.Fatal("Reading message from AMQP1.0:", err)
	// }

	// // Accept message
	// msg.Accept()

	// msgData := msg.GetData()
	// cancel()

	// var rcvCDR map[string]string
	// if err := json.Unmarshal(msgData, &rcvCDR); err != nil {
	// 	t.Error(err)
	// }
	// if rcvCDR[utils.CGRID] != utils.Sha1("httpjsonrpc1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()) {
	// 	t.Errorf("Unexpected CDR received: %+v", rcvCDR)
	// }
}

func TestCDRsOnExpAWSAMQPPosterFileFailover(t *testing.T) {
	time.Sleep(time.Duration(10 * time.Second))
	failoverContent := [][]byte{[]byte(`{"CGRID":"57548d485d61ebcba55afbe5d939c82a8e9ff670"}`), []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)}
	filesInDir, _ := ioutil.ReadDir(cdrsAWSCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsAWSCfg.GeneralCfg().FailedPostsDir)
	}
	var foundFile bool
	var fileName string
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName = file.Name()
		if strings.HasPrefix(fileName, "cdr|*aws_json_map") {
			foundFile = true
			filePath := path.Join(cdrsAWSCfg.GeneralCfg().FailedPostsDir, fileName)
			if readBytes, err := ioutil.ReadFile(filePath); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(failoverContent[0], readBytes) && !reflect.DeepEqual(failoverContent[1], readBytes) { // Checking just the prefix should do since some content is dynamic
				t.Errorf("Expecting: %v or %v, received: %v", string(failoverContent[0]), string(failoverContent[1]), string(readBytes))
			}
			if err := os.Remove(filePath); err != nil {
				t.Error("Failed removing file: ", filePath)
			}
		}
	}
	if !foundFile {
		t.Fatal("Could not find the file in folder")
	}
}

func TestCDRsOnExpAWSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
