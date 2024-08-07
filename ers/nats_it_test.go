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

package ers

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func TestNatsConcurrentReaders(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	cfgPath := path.Join(*utils.DataDir, "conf", "samples", "ers_nats")
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatal("could not init cfg", err.Error())
	}

	exec.Command("pkill", "nats-server")
	cmd := exec.Command("nats-server", "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	time.Sleep(50 * time.Millisecond)
	defer cmd.Process.Kill()

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
	if _, err = js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:     "stream",
		Subjects: []string{"cgrates_cdrs", "cgrates_cdrs_processed"},
	}); err != nil {
		t.Fatal(err)
	}
	defer js.DeleteStream(context.Background(), "stream")

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

}

var natsCfg string = `{

"general": {
	"node_id": "nats_test",
	"log_level": 7
},

"data_db": {
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "nats_processed",
			"type": "*virt",
			"fields": [
				{"tag": "CGRID", "type": "*variable", "value": "~*req.CGRID", "path": "*uch.CGRID"}
			]
		}
	]
},

"ers": {														
	"enabled": true,											
	"sessions_conns":[],
	"ees_conns": ["*internal"],								
	"readers": [
		{
			"id": "nats_reader1",									
			"type": "*nats_json_map",									
			"source_path": "%s",			
			"ees_success_ids": ["nats_processed"],	
			"flags": ["*dryrun"],										
			"opts": {
				%s	
			},
			"fields":[											
				{"tag": "CGRID", "type": "*variable", "value": "~*req.CGRID", "path": "*cgreq.CGRID"},
				{"tag": "readerId", "type": "*variable", "value": "~*vars.*readerID", "path": "*cgreq.ReaderID"},
			]
		}
	]
}
	
}`

func TestNatsNormalTT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	// SeedFile setup
	baseNKeyPath := t.TempDir()
	seedFilePath := path.Join(baseNKeyPath, "seed.txt")
	if err := os.WriteFile(seedFilePath, []byte("SUAOUIE5CU47NCO22GHFEZXGCRCJDVTHDLMIP4L7UQNCR5SW4FZICI7O3Q"), 0664); err != nil {
		t.Fatal(err)
	}
	nkeyCfgPath := path.Join(baseNKeyPath, "nats.cfg")
	if err := os.WriteFile(nkeyCfgPath, []byte(`authorization: {
		users: [
		  { nkey: UBSNABLSM4Y2KY4ZFWPDOB4NVNYCGVD5YB7ROC4EGSDR7Z7V57PXAIQY }
		]
	  }`), 0664); err != nil {
		t.Fatal(err)
	}

	// JWTFile setup
	baseJWTPath := t.TempDir()
	jwtFilePath := path.Join(baseJWTPath, "jwt.txt")
	if err := os.WriteFile(jwtFilePath, []byte(`-----BEGIN NATS USER JWT-----
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
	jwtCfgPath := path.Join(baseJWTPath, "nats.cfg")
	if err := os.WriteFile(jwtCfgPath, []byte(`// Operator "memory"
operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJFRk5ERUdSNU1aUEw1VElQTFVKMlNMTFdZV0VDU0NJSEhVU1lISE5IR1BZVUpaWE5XUlNRIiwiaWF0IjoxNjI0ODc1NzYwLCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJtZW1vcnkiLCJzdWIiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hdHMiOnsidHlwZSI6Im9wZXJhdG9yIiwidmVyc2lvbiI6Mn19.MZfwcw5j6zY8SfFQppGIa3VjYYZK2_n1kV16Nk5jVCgwS8dKWzRQK_XjFYWwQ15Cq9YY73jcTA6LO0DmQGsdBA

resolver: MEMORY

resolver_preload: {
	// Account "account"
	AAYJDVLZGWN7Y3FBPCVXCRVQZDFMYGHU4YXLGSXX7U254KH5SC3R5QKM: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJUNUFINlFUUEJIQlYyV1ZGWEkzMlFOT0RCVkI2Vkg3WTNJNzVQTjJBRzNPV0xESVc3TFFRIiwiaWF0IjoxNjI0ODc1ODEwLCJpc3MiOiJPQ0VSUlQ2WFNEQ1dBWTNFWVNTTjQ2UUxGQko3RFJHNTIzU1hIMkg0UjQ3WFZVWFYyUlJCSVNMSyIsIm5hbWUiOiJhY2NvdW50Iiwic3ViIjoiQUFZSkRWTFpHV043WTNGQlBDVlhDUlZRWkRGTVlHSFU0WVhMR1NYWDdVMjU0S0g1U0MzUjVRS00iLCJuYXRzIjp7ImxpbWl0cyI6eyJzdWJzIjotMSwiZGF0YSI6LTEsInBheWxvYWQiOi0xLCJpbXBvcnRzIjotMSwiZXhwb3J0cyI6LTEsIndpbGRjYXJkcyI6dHJ1ZSwiY29ubiI6LTEsImxlYWYiOi0xfSwiZGVmYXVsdF9wZXJtaXNzaW9ucyI6eyJwdWIiOnt9LCJzdWIiOnt9fSwidHlwZSI6ImFjY291bnQiLCJ2ZXJzaW9uIjoyfX0.unslgXhO_ui9NpYkq5CuEmaU0rz5B1dbxr0bM98kXi2E-TB7RnTXPRGJpqTX16DKCdYhklfIVnI0zPMWHkaJCg

}
`), 0664); err != nil {
		t.Fatal(err)
	}

	testcases := []struct {
		name        string
		serverFlags []string
		sourcePath  string
		readerOpts  string
	}{
		{
			name:        "NoAuth",
			serverFlags: []string{},
			sourcePath:  "nats://127.0.0.1:4222",
			readerOpts:  "",
		},
		{
			name:        "UsernameAndPassword",
			serverFlags: []string{"--user", "user", "--pass", "password"},
			sourcePath:  "nats://user:password@127.0.0.1:4222",
			readerOpts:  "",
		},
		{
			name:        "TokenAuth",
			serverFlags: []string{"--auth", "token"},
			sourcePath:  "nats://token@127.0.0.1:4222",
			readerOpts:  "",
		},
		{
			name:        "NkeyAuth",
			serverFlags: []string{"-c", nkeyCfgPath},
			sourcePath:  "nats://127.0.0.1:4222",
			readerOpts: fmt.Sprintf(
				`"natsSeedFile": "%s"`,
				seedFilePath,
			),
		},
		{
			name:        "JWTAuth",
			serverFlags: []string{"-c", jwtCfgPath},
			sourcePath:  "nats://127.0.0.1:4222",
			readerOpts: fmt.Sprintf(
				`"natsJWTFile": "%s"`,
				jwtFilePath,
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			exec.Command("pkill", "nats-server")
			cmd := exec.Command("nats-server", tc.serverFlags...)
			if err := cmd.Start(); err != nil {
				t.Fatal(err) // most probably not installed
			}
			time.Sleep(50 * time.Millisecond)
			defer cmd.Process.Kill()

			configJSON := fmt.Sprintf(natsCfg, tc.sourcePath, tc.readerOpts)
			cfgPath := t.TempDir()
			filePath := filepath.Join(cfgPath, "cgrates.json")
			err := os.WriteFile(filePath, []byte(configJSON), 0644)
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := config.NewCGRConfigFromPath(cfgPath)
			if err != nil {
				t.Fatal(err)
			}

			rdrCfgOpts := cfg.ERsCfg().Readers[1].Opts.NATS
			nop, err := GetNatsOpts(rdrCfgOpts, cfg.GeneralCfg().NodeID, time.Second)
			if err != nil {
				t.Fatal(err)
			}

			// Establish a connection to nats.
			nc, err := nats.Connect(tc.sourcePath, nop...)
			if err != nil {
				t.Fatal(err)
			}

			if _, err = engine.StartEngine(cfgPath, *utils.WaitRater); err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() {
				engine.KillEngine(*utils.WaitRater)
				nc.Close()
			})

			client, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 3; i++ {
				randomCGRID := utils.UUIDSha1Prefix()
				expData := fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)
				if err = nc.Publish("cgrates_cdrs", []byte(expData)); err != nil {
					t.Error(err)
				}

				time.Sleep(20 * time.Millisecond) // wait for exports

				var cgrID any
				if err = client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
					Tenant: "cgrates.org",
					ArgsGetCacheItem: utils.ArgsGetCacheItem{
						CacheID: utils.CacheUCH,
						ItemID:  "CGRID",
					},
				}, &cgrID); err != nil {
					t.Error(err)
				} else if cgrID != randomCGRID {
					t.Errorf("expected %v, received %v", randomCGRID, cgrID)
				}
			}
		})
	}
}

func TestNatsJetStreamTT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	// SeedFile setup
	baseNKeyPath := t.TempDir()
	seedFilePath := path.Join(baseNKeyPath, "seed.txt")
	if err := os.WriteFile(seedFilePath, []byte("SUAOUIE5CU47NCO22GHFEZXGCRCJDVTHDLMIP4L7UQNCR5SW4FZICI7O3Q"), 0664); err != nil {
		t.Fatal(err)
	}
	nkeyCfgPath := path.Join(baseNKeyPath, "nats.cfg")
	if err := os.WriteFile(nkeyCfgPath, []byte(`authorization: {
		users: [
		  { nkey: UBSNABLSM4Y2KY4ZFWPDOB4NVNYCGVD5YB7ROC4EGSDR7Z7V57PXAIQY }
		]
	  }`), 0664); err != nil {
		t.Fatal(err)
	}

	// JWTFile setup
	baseJWTPath := t.TempDir()
	// baseJWTPath := "/tmp/natsCfg3"
	// os.Mkdir("/tmp/natsCfg3", 0644)
	jwtFilePath := path.Join(baseJWTPath, "jwt.txt")
	if err := os.WriteFile(jwtFilePath, []byte(`-----BEGIN NATS USER JWT-----
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
	jwtCfgPath := path.Join(baseJWTPath, "nats.cfg")
	if err := os.WriteFile(jwtCfgPath, []byte(`// Operator "memory"
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

	testcases := []struct {
		name        string
		serverFlags []string
		sourcePath  string
		readerOpts  string
	}{
		{
			name:        "NoAuth",
			serverFlags: []string{"-js"},
			sourcePath:  "nats://127.0.0.1:4222",
			readerOpts: `
			"natsJetStream": true,
			"natsStreamName": "stream"`,
		},
		{
			name:        "UsernameAndPassword",
			serverFlags: []string{"-js", "--user", "user", "--pass", "password"},
			sourcePath:  "nats://user:password@127.0.0.1:4222",
			readerOpts: `
			"natsJetStream": true,
			"natsStreamName": "stream"`,
		},
		{
			name:        "TokenAuth",
			serverFlags: []string{"-js", "--auth", "token"},
			sourcePath:  "nats://token@127.0.0.1:4222",
			readerOpts: `
			"natsJetStream": true,
			"natsStreamName": "stream"`,
		},
		{
			name:        "NkeyAuth",
			serverFlags: []string{"-js", "-c", nkeyCfgPath},
			sourcePath:  "nats://127.0.0.1:4222",
			readerOpts: fmt.Sprintf(`
			"natsJetStream": true,
			"natsStreamName": "stream",
			"natsSeedFile": "%s"`, seedFilePath),
		},
		{
			name:        "JWTAuth",
			serverFlags: []string{"-js", "-c", jwtCfgPath},
			sourcePath:  "nats://127.0.0.1:4222",
			readerOpts: fmt.Sprintf(`
			"natsJetStream": true,
			"natsStreamName": "stream",
			"natsJWTFile": "%s"`, jwtFilePath),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			exec.Command("pkill", "nats-server")
			cmd := exec.Command("nats-server", tc.serverFlags...)
			if err := cmd.Start(); err != nil {
				t.Fatal(err) // most probably not installed
			}
			time.Sleep(50 * time.Millisecond)
			defer cmd.Process.Kill()

			configJSON := fmt.Sprintf(natsCfg, tc.sourcePath, tc.readerOpts)
			cfgPath := t.TempDir()
			filePath := filepath.Join(cfgPath, "cgrates.json")
			err := os.WriteFile(filePath, []byte(configJSON), 0644)
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := config.NewCGRConfigFromPath(cfgPath)
			if err != nil {
				t.Fatal(err)
			}

			rdrCfgOpts := cfg.ERsCfg().Readers[1].Opts.NATS
			nop, err := GetNatsOpts(rdrCfgOpts, cfg.GeneralCfg().NodeID, 2*time.Second)
			if err != nil {
				t.Fatal(err)
			}

			// Establish a connection to nats.
			nc, err := nats.Connect(tc.sourcePath, nop...)
			if err != nil {
				t.Fatal(err)
			}
			defer nc.Close()

			// Initialize a stream manager and create a stream.
			js, err := jetstream.New(nc)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = js.CreateStream(context.Background(), jetstream.StreamConfig{
				Name:     "stream",
				Subjects: []string{"cgrates_cdrs", "cgrates_cdrs_processed"},
			}); err != nil {
				t.Fatal(err)
			}
			defer js.DeleteStream(context.Background(), "stream")

			if _, err = engine.StartEngine(cfgPath, *utils.WaitRater); err != nil {
				t.Fatal(err)
			}
			defer engine.KillEngine(*utils.WaitRater)

			client, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen)
			if err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 3; i++ {
				randomCGRID := utils.UUIDSha1Prefix()
				expData := fmt.Sprintf(`{"CGRID": "%s"}`, randomCGRID)
				if _, err := js.Publish(context.Background(), "cgrates_cdrs", []byte(expData)); err != nil {
					t.Error(err)
				}

				time.Sleep(20 * time.Millisecond) // wait for exports

				var cgrID any
				if err = client.Call(context.Background(), utils.CacheSv1GetItem, &utils.ArgsGetCacheItemWithAPIOpts{
					Tenant: "cgrates.org",
					ArgsGetCacheItem: utils.ArgsGetCacheItem{
						CacheID: utils.CacheUCH,
						ItemID:  "CGRID",
					},
				}, &cgrID); err != nil {
					t.Error(err)
				} else if cgrID != randomCGRID {
					t.Errorf("expected %v, received %v", randomCGRID, cgrID)
				}
			}
		})
	}
}
