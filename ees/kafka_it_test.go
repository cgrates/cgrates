//go:build integration
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

package ees

import (
	"net"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	kafka "github.com/segmentio/kafka-go"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	kafkaConfigDir string
	kafkaCfgPath   string
	kafkaCfg       *config.CGRConfig
	kafkaRpc       *birpc.Client

	sTestsKafka = []func(t *testing.T){
		testCreateDirectory,
		testKafkaLoadConfig,
		testKafkaResetDBs,

		testKafkaStartEngine,
		testKafkaRPCConn,
		testKafkaCreateTopic,
		testKafkaExportEvent,
		testKafkaVerifyExport,
		testKafkaDeleteTopic,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestKafkaExport(t *testing.T) {
	kafkaConfigDir = "ees"
	for _, stest := range sTestsKafka {
		t.Run(kafkaConfigDir, stest)
	}
}

func testKafkaLoadConfig(t *testing.T) {
	var err error
	kafkaCfgPath = path.Join(*utils.DataDir, "conf", "samples", kafkaConfigDir)
	if kafkaCfg, err = config.NewCGRConfigFromPath(context.Background(), kafkaCfgPath); err != nil {
		t.Error(err)
	}
}

func testKafkaResetDBs(t *testing.T) {
	if err := engine.InitDataDB(kafkaCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(kafkaCfg); err != nil {
		t.Fatal(err)
	}
}

func testKafkaStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(kafkaCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testKafkaRPCConn(t *testing.T) {
	var err error
	kafkaRpc, err = engine.NewRPCClient(kafkaCfg.ListenCfg(), *utils.Encoding)
	if err != nil {
		t.Fatal(err)
	}
}

func testKafkaCreateTopic(t *testing.T) {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		t.Fatal(err)
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatal(err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             utils.KafkaDefaultTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Fatal(err)
	}
}

func testKafkaExportEvent(t *testing.T) {
	event := &utils.CGREventWithEeIDs{
		EeIDs: []string{"KafkaExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "KafkaEvent",
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
	if err := kafkaRpc.Call(context.Background(), utils.EeSv1ProcessEvent, event, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testKafkaVerifyExport(t *testing.T) {
	// make a new reader that consumes from the cgrates topic
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     utils.KafkaDefaultTopic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	ctx, cancel := context.WithCancel(context.Background())
	m, err := r.ReadMessage(ctx)
	if err != nil {
		t.Error(err)
	}
	rcv := string(m.Value)
	cancel()

	exp := `{"Account":"1001","AnswerTime":"2013-11-07T08:42:28Z","Category":"call","Cost":1.01,"Destination":"1002","OriginHost":"192.168.1.1","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","SetupTime":"2013-11-07T08:42:25Z","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice","Usage":10000000000}`

	if rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	if err := r.Close(); err != nil {
		t.Fatal("failed to close reader:", err)
	}
}

func testKafkaDeleteTopic(t *testing.T) {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions("cgrates")
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 1 || partitions[0].Topic != "cgrates" {
		t.Fatal("expected topic named cgrates to exist")
	}

	if err := conn.DeleteTopics(utils.KafkaDefaultTopic); err != nil {
		t.Fatal(err)
	}

	experr := `[3] Unknown Topic Or Partition: the request is for a topic or partition that does not exist on this broker`
	_, err = conn.ReadPartitions("cgrates")
	if err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}
