//go:build kafka
// +build kafka

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
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/segmentio/kafka-go"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	kafkaSSLConfigDir string
	kafkaSSLCfgPath   string
	kafkaSSLCfg       *config.CGRConfig
	kafkaSSLRpc       *birpc.Client

	sTestsKafkaSSL = []func(t *testing.T){
		testKafkaSSLLoadConfig,
		testKafkaSSLFlushDBs,

		testKafkaSSLStartEngine,
		testKafkaSSLRPCConn,
		testKafkaSSLExportEvent,           // exports event to ssl-topic, then the reader will consume said event and export it to processed-topic
		testKafkaSSLVerifyProcessedExport, // checks whether ERs managed to successfully read and export the events served by Kafka server
		testKafkaSSLStopEngine,
	}
)

// The test is exporting and reading from a kafka broker with the following configuration

/*
listeners=PLAINTEXT://:9092,SSL://localhost:9093
...
advertised.listeners=PLAINTEXT://localhost:9092,SSL://localhost:9093
...
ssl.truststore.location=/home/kafka/kafka/ssl/kafka.server.truststore.jks
ssl.truststore.password=123456
ssl.keystore.type=PKCS12
ssl.keystore.location=/home/kafka/kafka/ssl/kafka.server.keystore.p12
ssl.keystore.password=123456
ssl.key.password=123456
ssl.client.auth=none
ssl.protocol=TLSv1.2
security.inter.broker.protocol=SSL
*/

// How to create TLS keys and certificates:

/*
1. Generate CA if needed (openssl req -new -x509 -keyout ca-key.pem -out ca.pem -days 365);
2. Add the generated CA to the brokers’ truststore;
3. Generate key-certificate pair using the CA from step 1 to sign it and convert the pem files to p12 format;
4. Import both the certificate of the CA and the signed certificate into the broker keystore.
*/

func TestKafkaSSL(t *testing.T) {
	kafkaSSLConfigDir = "kafka_ssl"
	for _, stest := range sTestsKafkaSSL {
		t.Run(kafkaSSLConfigDir, stest)
	}
}

func testKafkaSSLLoadConfig(t *testing.T) {
	var err error
	kafkaSSLCfgPath = path.Join(*dataDir, "conf", "samples", kafkaSSLConfigDir)
	if kafkaSSLCfg, err = config.NewCGRConfigFromPath(context.Background(), kafkaSSLCfgPath); err != nil {
		t.Error(err)
	}
}

func testKafkaSSLFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(kafkaSSLCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(kafkaSSLCfg); err != nil {
		t.Fatal(err)
	}
}

func testKafkaSSLStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(kafkaSSLCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testKafkaSSLRPCConn(t *testing.T) {
	var err error
	kafkaSSLRpc, err = engine.NewRPCClient(kafkaSSLCfg.ListenCfg(), *encoding)
	if err != nil {
		t.Fatal(err)
	}
}

func testKafkaSSLExportEvent(t *testing.T) {
	event := &utils.CGREventWithEeIDs{
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
	if err := kafkaSSLRpc.Call(context.Background(), utils.EeSv1ProcessEvent, event, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testKafkaSSLVerifyProcessedExport(t *testing.T) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "processed-topic",
		// MinBytes: 10e3, // 10KB
		// MaxBytes: 10e6, // 10MB
	})

	ctx, cancel := context.WithCancel(context.Background())
	var rcv string
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			break
		}
		rcv = string(m.Value)
		cancel()
	}

	exp := `{"Account":"1001","AnswerTime":"2013-11-07T08:42:28Z","Category":"call","Cost":1.01,"Destination":"1002","OriginHost":"192.168.1.1","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","SetupTime":"2013-11-07T08:42:25Z","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice","Usage":10000000000}`

	if rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

	if err := r.Close(); err != nil {
		t.Fatal("failed to close reader:", err)
	}
}

func testKafkaSSLStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
