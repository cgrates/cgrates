//go:build integration
// +build integration

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
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	amqpConfigDir  string
	amqpCfgPath    string
	amqpCfg        *config.CGRConfig
	amqpRPC        *birpc.Client
	amqpExportPath string

	sTestsAMQP = []func(t *testing.T){
		testCreateDirectory,
		testAMQPLoadConfig,
		testAMQPResetDBs,

		testAMQPStartEngine,
		testAMQPRPCConn,

		testAMQPExportEvent,
		testAMQPVerifyExport,

		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestAMQPExport(t *testing.T) {
	amqpConfigDir = "ees"
	for _, stest := range sTestsAMQP {
		t.Run(amqpConfigDir, stest)
	}
}

func testAMQPLoadConfig(t *testing.T) {
	var err error
	amqpCfgPath = path.Join(*utils.DataDir, "conf", "samples", amqpConfigDir)
	if amqpCfg, err = config.NewCGRConfigFromPath(context.Background(), amqpCfgPath); err != nil {
		t.Error(err)
	}
	for _, value := range amqpCfg.EEsCfg().Exporters {
		if value.ID == "AMQPExporter" {
			amqpExportPath = value.ExportPath
		}
	}
}

func testAMQPResetDBs(t *testing.T) {
	if err := engine.InitDB(amqpCfg); err != nil {
		t.Fatal(err)
	}
}

func testAMQPStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(amqpCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAMQPRPCConn(t *testing.T) {
	amqpRPC = engine.NewRPCClient(t, amqpCfg.ListenCfg(), *utils.Encoding)
}

func testAMQPExportEvent(t *testing.T) {
	event := &utils.CGREventWithEeIDs{
		EeIDs: []string{"AMQPExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AMQPEvent",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "abcdef",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813748, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         1.01,
			},
		},
	}

	var reply map[string]map[string]any
	if err := amqpRPC.Call(context.Background(), utils.EeSv1ProcessEvent, event, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testAMQPVerifyExport(t *testing.T) {
	conn, err := amqp.Dial(amqpExportPath)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()
	q, err := ch.QueueDeclare("cgrates_cdrs", true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	msg, err := ch.Consume(q.Name, utils.EmptyString, true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	exp := `{"Account":"1001","AnswerTime":"2013-11-07T08:42:28Z","Category":"call","Cost":1.01,"Destination":"1002","OriginHost":"192.168.1.1","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","SetupTime":"2013-11-07T08:42:25Z","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice","Usage":10000000000}`
	var rcv string
	select {
	case d := <-msg:
		rcv = string(d.Body)
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
	if rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	// Delete the queue after verifying if the export was successful
	_, err = ch.QueueDelete("cgrates_cdrs", false, false, true)
	if err != nil {
		t.Error(err)
	}
}
