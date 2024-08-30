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

package ers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func TestNatsERIT(t *testing.T) {
	cfgPath := path.Join(*utils.DataDir, "conf", "samples", "ers_nats")
	cfg, err := config.NewCGRConfigFromPath(context.Background(), cfgPath)
	if err != nil {
		t.Fatal("could not init cfg", err.Error())
	}

	natsServer, err := server.NewServer(&server.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		JetStream: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	// Establish a connection to nats.
	nc, err := nats.Connect(cfg.ERsCfg().Readers[1].SourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	// Initialize a stream manager and create a stream.
	js, err := jetstream.New(nc)
	if err != nil {
		t.Fatal(err)
	}
	js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "stream",
		Subjects: []string{"cgrates_cdrs", "cgrates_cdrs_processed"},
	})

	// Start the engine.
	if _, err := engine.StopStartEngine(cfgPath, 100); err != nil {
		t.Fatal(err)
	}
	defer engine.KillEngine(100)

	// Publish CDRs asynchronously to the nats subject.
	cdr := make(map[string]any)
	for i := 0; i < 10; i++ {
		cdr[utils.AccountField] = 1001 + i
		cdr[utils.Subject] = 1001 + i
		cdr[utils.Destination] = 2001 + i
		b, _ := json.Marshal(cdr)
		js.PublishAsync("cgrates_cdrs", b)
	}
	select {
	case <-js.PublishAsyncComplete():
	case <-time.After(5 * time.Second):
		t.Fatal("Did not resolve in time")
	}

	// Define a consumer for the subject where all the processed cdrs were published.
	var cons jetstream.Consumer
	cons, err = js.CreateOrUpdateConsumer(context.Background(), "stream", jetstream.ConsumerConfig{
		FilterSubject: "cgrates_cdrs_processed",
		Durable:       "cgrates_processed",
		AckPolicy:     jetstream.AckAllPolicy,
	})
	if err != nil {
		t.Error(err)
	}

	// Wait for the messages to be consumed and processed.
	time.Sleep(100 * time.Millisecond)

	// Retrieve info about the consumer.
	info, err := cons.Info(context.Background())
	if err != nil {
		t.Error(err)
	}

	if info.NumPending != 10 {
		t.Errorf("expected %d pending messages, received %d", 10, info.NumPending)
	}

	js.DeleteStream(context.Background(), "stream")

}

func testCheckNatsData(t *testing.T, randomOriginID, expData string, ch chan string) {
	select {
	case err := <-rdrErr:
		t.Fatal(err)
	case ev := <-rdrEvents:
		if ev.rdrCfg.ID != "nats" {
			t.Fatalf("Expected 'nats' received `%s`", ev.rdrCfg.ID)
		}
		expected := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     ev.cgrEvent.ID,
			Event: map[string]any{
				"OriginID": randomOriginID,
				"ReaderID": "nats",
			},
			APIOpts: map[string]any{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Fatalf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
		select {
		case msg := <-ch:
			if expData != msg {
				t.Errorf("Expected %q ,received %q", expData, msg)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout2")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout")
	}
}

func testCheckNatsJetStream(t *testing.T, cfg *config.CGRConfig) {
	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)
	var err error
	if rdr, err = NewNatsER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}

	nop, err := GetNatsOpts(rdr.Config().Opts, "testExp", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect(rdr.Config().SourcePath, nop...)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		t.Fatal(err)
	}

	_, err = js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "test",
		Subjects: []string{utils.DefaultQueueID},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer js.DeleteStream(context.Background(), "test")

	_, err = js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "test2",
		Subjects: []string{"processed_cdrs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer js.DeleteStream(context.Background(), "test2")

	ch := make(chan string, 3)
	var cons jetstream.Consumer
	cons, err = js.CreateOrUpdateConsumer(context.Background(), "test2", jetstream.ConsumerConfig{
		FilterSubject: "processed_cdrs",
		Durable:       "test4",
		AckPolicy:     jetstream.AckAllPolicy,
	})
	if err != nil {
		nc.Drain()
		t.Fatal(err)
	}

	_, err = cons.Consume(func(msg jetstream.Msg) {
		ch <- string(msg.Data())
	})
	if err != nil {
		t.Fatal(err)
	}

	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(10 * time.Nanosecond)

	for i := 0; i < 3; i++ {
		randomOriginID := utils.UUIDSha1Prefix()
		expData := fmt.Sprintf(`{"OriginID": "%s"}`, randomOriginID)
		if _, err = js.Publish(context.Background(), utils.DefaultQueueID, []byte(expData)); err != nil {
			t.Fatal(err)
		}

		nc.FlushTimeout(time.Second)
		nc.Flush()

		testCheckNatsData(t, randomOriginID, expData, ch)
	}
	close(rdrExit)
}

func testCheckNatsNormal(t *testing.T, cfg *config.CGRConfig) {
	rdrEvents = make(chan *erEvent, 1)
	rdrErr = make(chan error, 1)
	rdrExit = make(chan struct{}, 1)

	var err error
	if rdr, err = NewNatsER(cfg, 1, rdrEvents, make(chan *erEvent, 1),
		rdrErr, new(engine.FilterS), rdrExit); err != nil {
		t.Fatal(err)
	}

	nop, err := GetNatsOpts(rdr.Config().Opts, "testExp", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect(rdr.Config().SourcePath, nop...)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan string, 3)
	_, err = nc.QueueSubscribe("processed_cdrs", "test3", func(msg *nats.Msg) {
		ch <- string(msg.Data)
	})
	if err != nil {
		t.Fatal(err)
	}

	defer nc.Drain()
	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 3; i++ {
		randomOriginID := utils.UUIDSha1Prefix()
		expData := fmt.Sprintf(`{"OriginID": "%s"}`, randomOriginID)
		if err = nc.Publish(utils.DefaultQueueID, []byte(expData)); err != nil {
			t.Fatal(err)
		}

		nc.FlushTimeout(time.Second)
		nc.Flush()

		testCheckNatsData(t, randomOriginID, expData, ch)
	}
	close(rdrExit)
}

func TestNatsERJetStream(t *testing.T) {
	natsServer, err := server.NewServer(&server.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		JetStream: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsJetStream": true,
				"natsStreamName": "test",
				"natsJetStreamProcessed": true,
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsJetStream(t, cfg)
}

func TestNatsER(t *testing.T) {
	natsServer, err := server.NewServer(&server.Options{
		Host: "127.0.0.1",
		Port: 4222,
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsNormal(t, cfg)
}

func TestNatsERJetStreamUser(t *testing.T) {
	natsServer, err := server.NewServer(&server.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		JetStream: true,
		Username:  "user",
		Password:  "password",
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://user:password@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsJetStream": true,
				"natsStreamName": "test",
				"natsJetStreamProcessed": true,
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsJetStream(t, cfg)
}

func TestNatsERUser(t *testing.T) {
	natsServer, err := server.NewServer(&server.Options{
		Host:     "127.0.0.1",
		Port:     4222,
		Username: "user",
		Password: "password",
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://user:password@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsNormal(t, cfg)
}

func TestNatsERJetStreamToken(t *testing.T) {
	natsServer, err := server.NewServer(&server.Options{
		Host:          "127.0.0.1",
		Port:          4222,
		JetStream:     true,
		Authorization: "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://token@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsJetStream": true,
				"natsStreamName": "test",
				"natsJetStreamProcessed": true,
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsJetStream(t, cfg)
}

func TestNatsERToken(t *testing.T) {
	natsServer, err := server.NewServer(&server.Options{
		Host:          "127.0.0.1",
		Port:          4222,
		JetStream:     true,
		Authorization: "token",
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://token@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsSubjectProcessed": "processed_cdrs",
			}
		},
	],
},
}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsNormal(t, cfg)
}

func TestNatsERNkey(t *testing.T) {
	// prepare
	basePath := "/tmp/natsCfg"
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(basePath)
	seedFilePath := path.Join(basePath, "seed.txt")
	if err := os.WriteFile(seedFilePath, []byte("SUAOUIE5CU47NCO22GHFEZXGCRCJDVTHDLMIP4L7UQNCR5SW4FZICI7O3Q"), 0664); err != nil {
		t.Fatal(err)
	}

	natsServer, err := server.NewServer(&server.Options{
		Host: "127.0.0.1",
		Port: 4222,
		Nkeys: []*server.NkeyUser{
			{
				Nkey: "UBSNABLSM4Y2KY4ZFWPDOB4NVNYCGVD5YB7ROC4EGSDR7Z7V57PXAIQY",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsSubjectProcessed": "processed_cdrs",
				"natsSeedFile": %q,
				"natsSeedFileProcessed": %q,
			}
		},
	],
},
}`, seedFilePath, seedFilePath))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsNormal(t, cfg)
}

func TestNatsERJetStreamNKey(t *testing.T) {
	// prepare
	basePath := "/tmp/natsCfg"
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(basePath)
	seedFilePath := path.Join(basePath, "seed.txt")
	if err := os.WriteFile(seedFilePath, []byte("SUAOUIE5CU47NCO22GHFEZXGCRCJDVTHDLMIP4L7UQNCR5SW4FZICI7O3Q"), 0664); err != nil {
		t.Fatal(err)
	}

	natsServer, err := server.NewServer(&server.Options{
		Host:      "127.0.0.1",
		Port:      4222,
		JetStream: true,
		Nkeys: []*server.NkeyUser{
			{
				Nkey: "UBSNABLSM4Y2KY4ZFWPDOB4NVNYCGVD5YB7ROC4EGSDR7Z7V57PXAIQY",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsJetStream": true,
				"natsStreamName": "test",
				"natsSeedFile": %q,
				"natsJetStreamProcessed": true,
				"natsSubjectProcessed": "processed_cdrs",
				"natsSeedFileProcessed": %q,
			}
		},
	],
},
}`, seedFilePath, seedFilePath))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsJetStream(t, cfg)
}

func TestNatsERJWT(t *testing.T) {
	// prepare
	basePath := "/tmp/natsCfg"
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(basePath)
	seedFilePath := path.Join(basePath, "jwt.txt")
	if err := os.WriteFile(seedFilePath, []byte(`-----BEGIN NATS USER JWT-----
eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJETFRGNFpLVVdNNFRPRkVNQko0UEUzTVFFVlhIUkJJN0xZNDdZNEMzNTNMWlJKSU5CUkJRIiwiaWF0IjoxNjI0ODg1NzMwLCJpc3MiOiJBQVlKRFZMWkdXTjdZM0ZCUENWWENSVlFaREZNWUdIVTRZWExHU1hYN1UyNTRLSDVTQzNSNVFLTSIsIm5hbWUiOiJ1c2VyIiwic3ViIjoiVUQzQkdXSUJQVTNDV0w0SE9VVTRWSkVRV1RVQVNBRUc2T0ozWEhNM0VFSk5BMlBBM1VWTllUWk4iLCJuYXRzIjp7InB1YiI6e30sInN1YiI6e30sInN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6LTEsInR5cGUiOiJ1c2VyIiwidmVyc2lvbiI6Mn19.YmFL5nRMkEOXe77sQJPPRv_vwi89tzhVVl0AVjE4sXWyoWIHiCepNw28DbpJ0p_MlT8Qf0SY2cjAhIm-Qi7lDw
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
SUADIH32XQYWC2MI2YGM4AUQ3NMKZSZ5V2BZXQ237XXMLO7FFHDF5CTUDE
------END USER NKEY SEED------

*************************************************************`), 0664); err != nil {
		t.Fatal(err)
	}
	natsCfgPath := path.Join(basePath, "nats.cfg")
	if err := os.WriteFile(natsCfgPath, []byte(`// Operator "memory"
operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJFRk5ERUdSNU1aUEw1VElQTFVKMlNMTFdZV0VDU0NJSEhVU1lISE5IR1BZVUpaWE5XUlNRIiwiaWF0IjoxNjI0ODc1NzYwLCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJtZW1vcnkiLCJzdWIiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hdHMiOnsidHlwZSI6Im9wZXJhdG9yIiwidmVyc2lvbiI6Mn19.MZfwcw5j6zY8SfFQppGIa3VjYYZK2_n1kV16Nk5jVCgwS8dKWzRQK_XjFYWwQ15Cq9YY73jcTA6LO0DmQGsdBA

resolver: MEMORY

resolver_preload: {
	// Account "account"
	AAYJDVLZGWN7Y3FBPCVXCRVQZDFMYGHU4YXLGSXX7U254KH5SC3R5QKM: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJUNUFINlFUUEJIQlYyV1ZGWEkzMlFOT0RCVkI2Vkg3WTNJNzVQTjJBRzNPV0xESVc3TFFRIiwiaWF0IjoxNjI0ODc1ODEwLCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJhY2NvdW50Iiwic3ViIjoiQUFZSkRWTFpHV043WTNGQlBDVlhDUlZRWkRGTVlHSFU0WVhMR1NYWDdVMjU0S0g1U0MzUjVRS00iLCJuYXRzIjp7ImxpbWl0cyI6eyJzdWJzIjotMSwiZGF0YSI6LTEsInBheWxvYWQiOi0xLCJpbXBvcnRzIjotMSwiZXhwb3J0cyI6LTEsIndpbGRjYXJkcyI6dHJ1ZSwiY29ubiI6LTEsImxlYWYiOi0xfSwiZGVmYXVsdF9wZXJtaXNzaW9ucyI6eyJwdWIiOnt9LCJzdWIiOnt9fSwidHlwZSI6ImFjY291bnQiLCJ2ZXJzaW9uIjoyfX0.unslgXhO_ui9NpYkq5CuEmaU0rz5B1dbxr0bM98kXi2E-TB7RnTXPRGJpqTX16DKCdYhklfIVnI0zPMWHkaJCg

}
`), 0664); err != nil {
		t.Fatal(err)
	}
	natsServer, err := server.NewServer(&server.Options{
		Host:       "127.0.0.1",
		Port:       4222,
		ConfigFile: natsCfgPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsSubjectProcessed": "processed_cdrs",
				"natsJWTFile": %q,
				"natsJWTFileProcessed": %q,
			}
		},
	],
},
}`, seedFilePath, seedFilePath))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsNormal(t, cfg)
}

func TestNatsERJetStreamJWT(t *testing.T) {
	// prepare
	basePath := "/tmp/natsCfg"
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(basePath)
	seedFilePath := path.Join(basePath, "jwt.txt")
	if err := os.WriteFile(seedFilePath, []byte(`-----BEGIN NATS USER JWT-----
eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJXTUUyUkhMWEU1R0FZS0hITk5aNkhLSDQ2Q0VSUFNEUExPN1BMT0ZEWTZaNUdZM09aVkFRIiwiaWF0IjoxNjI0OTUzNTE5LCJpc3MiOiJBQVlKRFZMWkdXTjdZM0ZCUENWWENSVlFaREZNWUdIVTRZWExHU1hYN1UyNTRLSDVTQzNSNVFLTSIsIm5hbWUiOiJ1c2VyMiIsInN1YiI6IlVERVJVRElMMlZORVNKUzVGNUpDQVJHWUNSUVM1M0NWSUVBSllHU0hFR0dZSEI2R01BQ1AzM1VDIiwibmF0cyI6eyJwdWIiOnt9LCJzdWIiOnt9LCJzdWJzIjotMSwiZGF0YSI6LTEsInBheWxvYWQiOi0xLCJ0eXBlIjoidXNlciIsInZlcnNpb24iOjJ9fQ.YitrKhIlU45Q1m6A_HFsaxDgUUiIyjLJuHNKnG1cjzj5H6n697Iv3pDsIZVTh6pBYROg1aRV42bD3PpEZna2AA
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
SUAGM22ETPOJNZGYSGJL3HRLZ6R35FSVOYNINYN5Z5UPLP5K4SQJ753WU4
------END USER NKEY SEED------

*************************************************************`), 0664); err != nil {
		t.Fatal(err)
	}
	natsCfgPath := path.Join(basePath, "nats.cfg")
	if err := os.WriteFile(natsCfgPath, []byte(`// Operator "memory"
operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJFRk5ERUdSNU1aUEw1VElQTFVKMlNMTFdZV0VDU0NJSEhVU1lISE5IR1BZVUpaWE5XUlNRIiwiaWF0IjoxNjI0ODc1NzYwLCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJtZW1vcnkiLCJzdWIiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hdHMiOnsidHlwZSI6Im9wZXJhdG9yIiwidmVyc2lvbiI6Mn19.MZfwcw5j6zY8SfFQppGIa3VjYYZK2_n1kV16Nk5jVCgwS8dKWzRQK_XjFYWwQ15Cq9YY73jcTA6LO0DmQGsdBA

resolver: MEMORY

resolver_preload: {
	// Account "js"
	AAFIBB6C56ROU5XRVJLJYR3BTGGYK3HJGHEHQV7L7QZMTT3ZRBLHBS7F: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJQNEZOWllRQkNKWERZT09ITzRNVU5BQTRHR0w2UTVIRkxKQUJXVEc3WVFIRFVNUVlHUldRIiwiaWF0IjoxNjI0OTU0MjQ0LCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJqcyIsInN1YiI6IkFBRklCQjZDNTZST1U1WFJWSkxKWVIzQlRHR1lLM0hKR0hFSFFWN0w3UVpNVFQzWlJCTEhCUzdGIiwibmF0cyI6eyJsaW1pdHMiOnsic3VicyI6LTEsImRhdGEiOi0xLCJwYXlsb2FkIjotMSwiaW1wb3J0cyI6LTEsImV4cG9ydHMiOi0xLCJ3aWxkY2FyZHMiOnRydWUsImNvbm4iOi0xLCJsZWFmIjotMX0sImRlZmF1bHRfcGVybWlzc2lvbnMiOnsicHViIjp7fSwic3ViIjp7fX0sInR5cGUiOiJhY2NvdW50IiwidmVyc2lvbiI6Mn19.tGaVbpNXuSFxk3RDxicbi62nupiTv_-vTgps0t-LmvxKoNuzjvrnhyARwdh3qknMP54pDqzlUfldqubmEYLFBg

	// Account "account"
	AAYJDVLZGWN7Y3FBPCVXCRVQZDFMYGHU4YXLGSXX7U254KH5SC3R5QKM: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJTUkVHMkdLUVg1RlJKQ0lTUlFITVlNUU9CU09DSkRYMjVaUUpGTllDMkxMQkZBRlNOQU9BIiwiaWF0IjoxNjI0OTUzNzIzLCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJhY2NvdW50Iiwic3ViIjoiQUFZSkRWTFpHV043WTNGQlBDVlhDUlZRWkRGTVlHSFU0WVhMR1NYWDdVMjU0S0g1U0MzUjVRS00iLCJuYXRzIjp7ImxpbWl0cyI6eyJzdWJzIjotMSwiZGF0YSI6LTEsInBheWxvYWQiOi0xLCJpbXBvcnRzIjotMSwiZXhwb3J0cyI6LTEsIndpbGRjYXJkcyI6dHJ1ZSwiY29ubiI6LTEsImxlYWYiOi0xLCJtZW1fc3RvcmFnZSI6LTEsInN0cmVhbXMiOi0xLCJjb25zdW1lciI6LTF9LCJkZWZhdWx0X3Blcm1pc3Npb25zIjp7InB1YiI6e30sInN1YiI6e319LCJ0eXBlIjoiYWNjb3VudCIsInZlcnNpb24iOjJ9fQ.rcOqLmWL77kgoDS4GPK5qs-rpG1mQCkQ5FoCzT3VGqsIXNdpn72d38jbCeV40_6l8dI49IRtRHySv8k7VwaaAA

}
system_account:AAFIBB6C56ROU5XRVJLJYR3BTGGYK3HJGHEHQV7L7QZMTT3ZRBLHBS7F
`), 0664); err != nil {
		t.Fatal(err)
	}
	natsServer, err := server.NewServer(&server.Options{
		Host:       "127.0.0.1",
		Port:       4222,
		ConfigFile: natsCfgPath,
		JetStream:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	natsServer.Start()
	defer natsServer.Shutdown()

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*natsJSONMap",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "OriginID", "type": "*composed", "value": "~*req.OriginID", "path": "*cgreq.OriginID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			],
			"opts": {
				"natsJetStream": true,
				"natsStreamName": "test",
				"natsJWTFile": %q,
				"natsJetStreamProcessed": true,
				"natsSubjectProcessed": "processed_cdrs",
				"natsJWTFileProcessed": %q,
			}
		},
	],
},
}`, seedFilePath, seedFilePath))
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.CheckConfigSanity(); err != nil {
		t.Fatal(err)
	}
	testCheckNatsJetStream(t, cfg)
}
