//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"

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
	if err := engine.InitDB(kafkaCfg); err != nil {
		t.Fatal(err)
	}
}

func testKafkaStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(kafkaCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testKafkaRPCConn(t *testing.T) {
	kafkaRpc = engine.NewRPCClient(t, kafkaCfg.ListenCfg(), *utils.Encoding)
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
