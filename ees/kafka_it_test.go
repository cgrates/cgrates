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
along with this program. If not, see <https://www.gnu.org/licenses/>
*/

package ees

import (
	"context"
	"path"
	"testing"
	"time"

	birpcctx "github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	kafkaConfigDir string
	kafkaCfgPath   string
	kafkaCfg       *config.CGRConfig
	kafkaRpc       *birpc.Client

	sTestsKafka = []func(t *testing.T){
		testCreateDirectory,
		testKafkaLoadConfig,
		testKafkaResetDataDB,
		testKafkaResetStorDB,
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
	if kafkaCfg, err = config.NewCGRConfigFromPath(kafkaCfgPath); err != nil {
		t.Error(err)
	}
}

func testKafkaResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(kafkaCfg); err != nil {
		t.Fatal(err)
	}
}

func testKafkaResetStorDB(t *testing.T) {
	if err := engine.InitStorDb(kafkaCfg); err != nil {
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
	kafkaRpc, err = newRPCClient(kafkaCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testKafkaCreateTopic(t *testing.T) {
	cl, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	adm := kadm.NewClient(cl)
	_, err = adm.CreateTopics(context.Background(), 1, 1, nil, utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}
}

func testKafkaExportEvent(t *testing.T) {
	event := &engine.CGREventWithEeIDs{
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
	if err := kafkaRpc.Call(birpcctx.Background(), utils.EeSv1ProcessEvent, event, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testKafkaVerifyExport(t *testing.T) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers("localhost:9092"),
		kgo.ConsumeTopics(utils.KafkaDefaultTopic),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fetches := cl.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		t.Fatal(errs[0].Err)
	}

	var rcv string
	fetches.EachRecord(func(r *kgo.Record) {
		rcv = string(r.Value)
	})

	exp := `{"Account":"1001","AnswerTime":"2013-11-07T08:42:28Z","Category":"call","Cost":1.01,"Destination":"1002","OriginHost":"192.168.1.1","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","SetupTime":"2013-11-07T08:42:25Z","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice","Usage":10000000000}`

	if rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func testKafkaDeleteTopic(t *testing.T) {
	cl, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"))
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	adm := kadm.NewClient(cl)

	topics, err := adm.ListTopics(context.Background(), utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}
	if !topics.Has(utils.KafkaDefaultTopic) {
		t.Fatal("expected topic named cgrates to exist")
	}

	if _, err := adm.DeleteTopics(context.Background(), utils.KafkaDefaultTopic); err != nil {
		t.Fatal(err)
	}

	topics, err = adm.ListTopics(context.Background(), utils.KafkaDefaultTopic)
	if err != nil {
		t.Fatal(err)
	}
	if topics.Has(utils.KafkaDefaultTopic) {
		t.Error("expected topic to be deleted")
	}
}
