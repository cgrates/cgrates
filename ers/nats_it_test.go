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
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

func testCheckNatsData(t *testing.T, randomCGRID, expData string, ch chan *nats.Msg) {
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
			Time:   ev.cgrEvent.Time,
			Event: map[string]interface{}{
				"CGRID": randomCGRID,
			},
			APIOpts: map[string]interface{}{},
		}
		if !reflect.DeepEqual(ev.cgrEvent, expected) {
			t.Fatalf("Expected %s ,received %s", utils.ToJSON(expected), utils.ToJSON(ev.cgrEvent))
		}
		select {
		case msg := <-ch:
			if expData != string(msg.Data) {
				t.Errorf("Expected %q ,received %q", expData, string(msg.Data))
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

	nop, err := engine.GetNatsOpts(rdr.Config().Opts, "testExp", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect(rdr.Config().SourcePath, nop...)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	for name := range js.StreamNames() {
		if name == "test" {
			if err = js.DeleteStream("test"); err != nil {
				t.Fatal(err)
			}
			break
		}
		if name == "test2" {
			if err = js.DeleteStream("test2"); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	if _, err = js.AddStream(&nats.StreamConfig{
		Name:     "test",
		Subjects: []string{utils.DefaultQueueID},
	}); err != nil {
		t.Fatal(err)
	}

	if err = js.PurgeStream("test"); err != nil {
		t.Fatal(err)
	}

	if _, err = js.AddStream(&nats.StreamConfig{
		Name:     "test2",
		Subjects: []string{"processed_cdrs"},
	}); err != nil {
		t.Fatal(err)
	}

	if err = js.PurgeStream("test2"); err != nil {
		t.Fatal(err)
	}
	ch := make(chan *nats.Msg, 3)
	_, err = js.QueueSubscribe("processed_cdrs", "test3", func(msg *nats.Msg) {
		ch <- msg
	}, nats.Durable("test4"))
	if err != nil {
		t.Fatal(err)
	}

	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(10 * time.Nanosecond)

	for i := 0; i < 3; i++ {
		randomCGRID := utils.UUIDSha1Prefix()
		expData := fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)
		if _, err = js.Publish(utils.DefaultQueueID, []byte(expData)); err != nil {
			t.Fatal(err)
		}

		nc.FlushTimeout(time.Second)
		nc.Flush()

		testCheckNatsData(t, randomCGRID, expData, ch)
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

	nop, err := engine.GetNatsOpts(rdr.Config().Opts, "testExp", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	nc, err := nats.Connect(rdr.Config().SourcePath, nop...)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *nats.Msg, 3)
	_, err = nc.ChanQueueSubscribe("processed_cdrs", "test3", ch)
	if err != nil {
		t.Fatal(err)
	}

	defer nc.Drain()
	go rdr.Serve()
	runtime.Gosched()
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 3; i++ {
		randomCGRID := utils.UUIDSha1Prefix()
		expData := fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)
		if err = nc.Publish(utils.DefaultQueueID, []byte(expData)); err != nil {
			t.Fatal(err)
		}

		nc.FlushTimeout(time.Second)
		nc.Flush()

		testCheckNatsData(t, randomCGRID, expData, ch)
	}
	close(rdrExit)
}

func TestNatsERJetStream(t *testing.T) {
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsJetStream": true,
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-js", "--user", "user", "--pass", "password")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://user:password@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsJetStream": true,
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "--user", "user", "--pass", "password")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://user:password@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-js", "--auth", "token")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://token@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsJetStream": true,
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "--auth", "token")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://token@localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
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
	basePath := "/tmp/nkey"
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(basePath)
	seedFilePath := path.Join(basePath, "seed.txt")
	if err := os.WriteFile(seedFilePath, []byte("SUAOUIE5CU47NCO22GHFEZXGCRCJDVTHDLMIP4L7UQNCR5SW4FZICI7O3Q"), 0664); err != nil {
		t.Fatal(err)
	}
	natsCfgPath := path.Join(basePath, "nats.cfg")
	if err := os.WriteFile(natsCfgPath, []byte(`authorization: {
	users: [
	  { nkey: UBSNABLSM4Y2KY4ZFWPDOB4NVNYCGVD5YB7ROC4EGSDR7Z7V57PXAIQY }
	]
  }
`), 0664); err != nil {
		t.Fatal(err)
	}
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-c", natsCfgPath)
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
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
	basePath := "/tmp/nkey"
	if err := os.MkdirAll(basePath, 0755); err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(basePath)
	seedFilePath := path.Join(basePath, "seed.txt")
	if err := os.WriteFile(seedFilePath, []byte("SUAOUIE5CU47NCO22GHFEZXGCRCJDVTHDLMIP4L7UQNCR5SW4FZICI7O3Q"), 0664); err != nil {
		t.Fatal(err)
	}
	natsCfgPath := path.Join(basePath, "nats.cfg")
	if err := os.WriteFile(natsCfgPath, []byte(`authorization: {
users: [
  { nkey: UBSNABLSM4Y2KY4ZFWPDOB4NVNYCGVD5YB7ROC4EGSDR7Z7V57PXAIQY }
]
}
`), 0664); err != nil {
		t.Fatal(err)
	}
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-c", natsCfgPath, "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsJetStream": true,
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
	basePath := "/tmp/nkey"
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-c", natsCfgPath)
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
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
	basePath := "/tmp/nkey"
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
	// start the nats-server
	exec.Command("pkill", "nats-server")

	cmd := exec.Command("nats-server", "-c", natsCfgPath, "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(100 * time.Millisecond)
	defer cmd.Process.Kill()
	//

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(fmt.Sprintf(`{
"ers": {									// EventReaderService
	"enabled": true,						// starts the EventReader service: <true|false>
	"sessions_conns":["*localhost"],
	"readers": [
		{
			"id": "nats",										
			"type": "*nats_json_map",							
			"run_delay":  "-1",									
			"concurrent_requests": 1024,						
			"source_path": "nats://localhost:4222",				
			"processed_path": "",	
			"tenant": "cgrates.org",							
			"filters": [],										
			"flags": [],										
			"fields":[									
				{"tag": "CGRID", "type": "*composed", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
			],
			"opts": {
				"natsJetStream": true,
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
