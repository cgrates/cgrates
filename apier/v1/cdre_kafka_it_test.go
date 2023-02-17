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

package v1

import (
	"context"
	"log"
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
)

var (
	kafkaCfgPath   string
	kafkaCfg       *config.CGRConfig
	kafkaRPC       *rpc.Client
	kafkaConfigDIR string

	sTestsCDReKafka = []func(t *testing.T){
		testKafkaInitCfg,
		testKafkaInitDataDb,
		testKafkaResetStorDb,
		testKafkaStartEngine,
		testKafkaRPCConn,
		testKafkaAddCDRs,
		testKafkaExportCDRs,
		testKafkaVerifyExport,
		testKafkaDeleteTopic,
		testKafkaKillEngine,
	}
)

func TestKafkaExport(t *testing.T) {
	kafkaConfigDIR = "cdre"
	for _, stest := range sTestsCDReKafka {
		t.Run(kafkaConfigDIR, stest)
	}
}

func testKafkaInitCfg(t *testing.T) {
	var err error
	kafkaCfgPath = path.Join("/usr/share/cgrates", "conf", "samples", kafkaConfigDIR)
	kafkaCfg, err = config.NewCGRConfigFromPath(kafkaCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	kafkaCfg.DataFolderPath = "/usr/share/cgrates" // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(kafkaCfg)
}

func testKafkaInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(kafkaCfg); err != nil {
		t.Fatal(err)
	}
}

func testKafkaResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(kafkaCfg); err != nil {
		t.Fatal(err)
	}
}

func testKafkaStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(kafkaCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testKafkaRPCConn(t *testing.T) {
	var err error
	kafkaRPC, err = newRPCClient(kafkaCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testKafkaAddCDRs(t *testing.T) {
	storedCdrs := []*engine.CDR{
		{
			CGRID:       "Cdr1",
			OrderID:     101,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr2",
			OrderID:     102,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR2",
			OriginHost:  "192.168.1.1",
			Source:      "test2",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(5) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr3",
			OrderID:     103,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR3",
			OriginHost:  "192.168.1.1",
			Source:      "test2",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now(),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(30) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
		{
			CGRID:       "Cdr4",
			OrderID:     104,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR4",
			OriginHost:  "192.168.1.1",
			Source:      "test3",
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   time.Now(),
			AnswerTime:  time.Time{},
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(0) * time.Second,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	for _, cdr := range storedCdrs {
		var reply string
		if err := kafkaRPC.Call(utils.CDRsV1ProcessCDR, &engine.CDRWithArgDispatcher{CDR: cdr}, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	time.Sleep(100 * time.Millisecond)
}

func testKafkaExportCDRs(t *testing.T) {
	attr := ArgExportCDRs{
		ExportArgs: map[string]interface{}{
			utils.ExportTemplate: "kafka_exporter",
		},
		Verbose: true,
	}
	var rply RplExportedCDRs
	if err := kafkaRPC.Call(utils.APIerSv1ExportCDRs, attr, &rply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rply.ExportedCGRIDs) != 2 {
		t.Errorf("Unexpected number of CDR exported: %s ", utils.ToJSON(rply))
	}
}

func testKafkaVerifyExport(t *testing.T) {
	topic := "cgrates_cdrs"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	batch := conn.ReadBatch(10e3, 1e6) // fetch 10KB min, 1MB max

	b := make([]byte, 10e3) // 10KB max per message
	var cdrs []string
	msgChan := make(chan []string)
	go func() {
		var i int
		for i < 2 {
			n, err := batch.Read(b)
			if err != nil {
				break
			}
			cdrs = append(cdrs, string(b[:n]))
			i++
		}
		msgChan <- cdrs
	}()

	expCDRs := []string{
		`{"Account":"1001","CGRID":"Cdr2","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR2","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"5s"}`,
		`{"Account":"1001","CGRID":"Cdr3","Category":"call","Cost":"-1.0000","Destination":"+4986517174963","OriginID":"OriginCDR3","RunID":"*default","Source":"test2","Tenant":"cgrates.org","Usage":"30s"}`,
	}
	select {
	case rcvCDRs := <-msgChan:
		sort.Strings(rcvCDRs)
		if !reflect.DeepEqual(rcvCDRs, expCDRs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expCDRs, rcvCDRs)
		}
	case <-time.After(30 * time.Second):
		t.Error("Timeout: Failed to consume the messages in due time")
	}

	if err := batch.Close(); err != nil {
		log.Fatal("failed to close batch:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close connection:", err)
	}

}

func testKafkaDeleteTopic(t *testing.T) {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions("cgrates_cdrs")
	if err != nil {
		t.Fatal(err)
	}

	if len(partitions) != 1 || partitions[0].Topic != "cgrates_cdrs" {
		t.Fatal("expected topic named cgrates_cdrs to exist")
	}

	if err := conn.DeleteTopics("cgrates_cdrs"); err != nil {
		t.Fatal(err)
	}

	experr := `[5] Leader Not Available: the cluster is in the middle of a leadership election and there is currently no leader for this partition and hence it is unavailable for writes`
	_, err = conn.ReadPartitions("cgrates_cdrs")
	if err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

}

func testKafkaKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
