//go:build flaky

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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	kafka "github.com/segmentio/kafka-go"
)

var (
	cdrsExpCfgPath string
	cdrsExpCfgDir  string
	cdrsExpCfg     *config.CGRConfig
	cdrsExpRPC     *birpc.Client

	cdrsExpHTTPEv     = make(chan map[string]any, 1)
	cdrsExpHTTPServer *http.Server

	cdrsExpAMQPCon *amqp.Connection

	cdrsExpEv = &utils.CGREvent{

		ID:     "Export",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.CGRID:        "TestCGRID",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestCDRsExp",
			utils.OriginHost:   "192.168.1.0",
			utils.RequestType:  utils.MetaRated,
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC),
			utils.AnswerTime:   time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:        10 * time.Second,
			utils.Cost:         1.201,
		},
		APIOpts: map[string]any{},
	}

	cdrsExpEvExp = map[string]any{
		utils.CGRID:        "TestCGRID",
		utils.ToR:          utils.MetaVoice,
		utils.OriginID:     "TestCDRsExp",
		utils.RequestType:  utils.MetaRated,
		utils.Tenant:       "cgrates.org",
		utils.Category:     "call",
		utils.AccountField: "1001",
		utils.Subject:      "1001",
		utils.Destination:  "1002",
		utils.RunID:        utils.MetaRaw,
		utils.OrderID:      "0",
	}

	cdrsExpTests = []func(t *testing.T){
		testCDRsExpInitConfig,
		testCDRsExpInitDB,
		testCDRsExpPrepareHTTP,
		testCDRsExpPrepareAMQP,
		testCDRsExpStartEngine,
		testCDRsExpInitRPC,
		testCDRsExpLoadAddCharger,
		testCDRsExpExportEvent,
		testCDRsExpHTTP,
		testCDRsExpAMQP,
		testCDRsExpKafka,
		testCDRsExpFileFailover,
		testCDRsExpStopEngine,
		testCDRsExpStopHTTPServer,
		testCDRsExpCloseAMQP,
	}
)

func TestCDRsExp(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		cdrsExpCfgDir = "cdrsexport_internal"
	case utils.MetaMySQL:
		cdrsExpCfgDir = "cdrsexport_mysql"
	case utils.MetaMongo:
		cdrsExpCfgDir = "cdrsexport_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range cdrsExpTests {
		t.Run(cdrsExpCfgDir, stest)
	}
}

func testCDRsExpInitConfig(t *testing.T) {
	var err error
	cdrsExpCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsExpCfgDir)
	if cdrsExpCfg, err = config.NewCGRConfigFromPath(cdrsExpCfgPath); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExpInitDB(t *testing.T) {
	if err := engine.InitDataDb(cdrsExpCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(cdrsExpCfg); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir); err != nil {
		t.Fatal("Error removing folder: ", cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir, err)
	}
	if err := os.MkdirAll(cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir, 0755); err != nil {
		t.Error(err)
	}

}

func testCDRsExpPrepareHTTP(t *testing.T) {
	srvMux := http.NewServeMux()
	srvMux.HandleFunc("/cdr_http", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(r.Body); err != nil {
			t.Logf("Error reading the body:%s", err)
			return
		}
		vals, err := url.ParseQuery(buf.String())
		if err != nil {
			t.Logf("Error parsing the values: %s", err)
			return
		}

		ev := make(map[string]any)
		for k, v := range vals {
			ev[k] = v[0]
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, utils.OK)
		cdrsExpHTTPEv <- ev
	})
	cdrsExpHTTPServer = &http.Server{Addr: ":12081", Handler: srvMux}

	go func(t2 *testing.T) {
		if err := cdrsExpHTTPServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			t2.Log(err)
		}
	}(t)
}

func testCDRsExpPrepareAMQP(t *testing.T) {
	var err error
	if cdrsExpAMQPCon, err = amqp.Dial("amqp://guest:guest@localhost:5672/"); err != nil {
		t.Fatal(err)
	}
	defer cdrsExpAMQPCon.Close()

	var ch *amqp.Channel
	if ch, err = cdrsExpAMQPCon.Channel(); err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare("exchangename", "fanout", true, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExpStartEngine(t *testing.T) {
	runtime.Gosched()
	if _, err := engine.StopStartEngine(cdrsExpCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExpInitRPC(t *testing.T) {
	cdrsExpRPC = engine.NewRPCClient(t, cdrsExpCfg.ListenCfg())
}

func testCDRsExpLoadAddCharger(t *testing.T) {
	// //add a default charger
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "*raw",
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{"*constant:*opts.AddOrderID:true"},
			Weight:       20,
		},
	}
	var result string
	if err := cdrsExpRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testCDRsExpExportEvent(t *testing.T) {
	// stop RabbitMQ server so we can test reconnects
	if err := exec.Command("service", "rabbitmq-server", "stop").Run(); err != nil {
		t.Error(err)
	}
	var reply string
	if err := cdrsExpRPC.Call(context.Background(), utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags:    []string{"*export:true", utils.MetaRALs},
			CGREvent: *cdrsExpEv,
		}, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() { // some exporters will fail
		t.Error("Unexpected error: ", err)
	}
	// time.Sleep(50 * time.Millisecond)
	// filesInDir, _ := os.ReadDir(cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir)
	// if len(filesInDir) != 0 {
	// 	t.Errorf("Should be no files in directory: %s", cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir)
	// }
	// start RabbitMQ server so we can test reconnects
	if err := exec.Command("service", "rabbitmq-server", "start").Run(); err != nil {
		t.Error(err)
	}
	time.Sleep(2 * time.Second)
	var err error
	if cdrsExpAMQPCon, err = amqp.Dial("amqp://guest:guest@localhost:5672/"); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExpHTTP(t *testing.T) {
	select {
	case rcvCDR := <-cdrsExpHTTPEv:
		if !reflect.DeepEqual(cdrsExpEvExp, rcvCDR) {
			t.Errorf("Expected %s received %s", utils.ToJSON(cdrsExpEvExp), utils.ToJSON(rcvCDR))
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
}

func testCDRsExpAMQP(t *testing.T) {
	ch, err := cdrsExpAMQPCon.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume("cgrates_cdrs", "", true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case d := <-msgs:
		var rcvCDR map[string]any
		if err := json.Unmarshal(d.Body, &rcvCDR); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(cdrsExpEvExp, rcvCDR) {
			t.Errorf("Expected %s received %s", utils.ToJSON(cdrsExpEvExp), utils.ToJSON(rcvCDR))
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
}

func testCDRsExpKafka(t *testing.T) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "cgrates_cdrs",
		GroupID: "tmp",
		MaxWait: time.Millisecond,
	})

	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var m kafka.Message
	var err error
	if m, err = reader.ReadMessage(ctx); err != nil {
		t.Fatal(err)
	}
	var rcvCDR map[string]any
	if err := json.Unmarshal(m.Value, &rcvCDR); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(cdrsExpEvExp, rcvCDR) {
		t.Errorf("Expected %s received %s", utils.ToJSON(cdrsExpEvExp), utils.ToJSON(rcvCDR))
	}
	cancel()
}

func checkContent(ev *ees.ExportEvents, content []any) error {
	match := false
	for _, bev := range ev.Events {
		for _, con := range content {
			if reflect.DeepEqual(bev, con) {
				match = true
				break
			}
		}
		if match {
			break
		}
	}
	if !match {
		exp := make([]string, len(content))
		for i, con := range content {
			exp[i] = utils.IfaceAsString(con)
		}
		recv := make([]string, len(ev.Events))
		for i, con := range ev.Events {
			recv[i] = utils.IfaceAsString(con)
		}
		return fmt.Errorf("Expecting: one of %s, received: %s", utils.ToJSON(exp), utils.ToJSON(recv))
	}
	return nil
}
func testCDRsExpFileFailover(t *testing.T) {
	time.Sleep(time.Second)
	filesInDir, _ := os.ReadDir(cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir)
	}
	expectedFormats := utils.NewStringSet([]string{utils.MetaAMQPV1jsonMap, utils.MetaSQSjsonMap, utils.MetaS3jsonMap})
	rcvFormats := utils.StringSet{}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(cdrsExpCfg.EEsCfg().Exporters[1].FailedPostsDir, fileName)

		ev, err := ees.NewExportEventsFromFile(filePath)
		if err != nil {
			t.Errorf("<%s> for file <%s>", err, fileName)
			continue
		} else if len(ev.Events) == 0 {
			t.Error("Expected at least one event")
			continue
		}
		rcvFormats.Add(ev.Type)
		if err := checkContent(ev, []any{[]byte(utils.ToJSON(cdrsExpEvExp))}); err != nil {
			t.Errorf("For file <%s> and event <%s> received %s", filePath, utils.ToJSON(ev), err)
		}
	}
	if !reflect.DeepEqual(expectedFormats, rcvFormats) {
		t.Errorf("Missing format expecting: %s received: %s", utils.ToJSON(expectedFormats), utils.ToJSON(rcvFormats))
	}
}

func testCDRsExpStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testCDRsExpStopHTTPServer(t *testing.T) {
	var err error
	if err = cdrsExpHTTPServer.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func testCDRsExpCloseAMQP(t *testing.T) {
	ch, err := cdrsExpAMQPCon.Channel()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = ch.QueueDelete("cgrates_cdrs", false, false, true); err != nil {
		ch.Close()
		t.Fatal(err)
	}
	ch.Close()
	if err := cdrsExpAMQPCon.Close(); err != nil {
		t.Fatal(err)
	}
}
