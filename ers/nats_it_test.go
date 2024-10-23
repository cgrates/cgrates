//go:build integration

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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
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

	cmd := exec.Command("nats-server", "-js")
	if err := cmd.Start(); err != nil {
		t.Fatal(err) // most probably not installed
	}
	t.Cleanup(func() { cmd.Process.Kill() })

	var js jetstream.JetStream // to reuse jetstream instance

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf/samples/ers_nats"),
		PreStartHook: func(t testing.TB, c *config.CGRConfig) {
			nc := connectToNATSServer(t, "nats://127.0.0.1:4222")

			// Initialize a stream manager and create a stream.
			var err error
			js, err = jetstream.New(nc)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := js.CreateStream(context.Background(), jetstream.StreamConfig{
				Name:     "stream",
				Subjects: []string{"cgrates_cdrs", "cgrates_cdrs_processed"},
			}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := js.DeleteStream(context.Background(), "stream"); err != nil {
					t.Errorf("failed to clean up stream: %v", err)
				}
			})
		},
	}
	ng.Run(t)

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
	cons, err := js.CreateOrUpdateConsumer(context.Background(), "stream", jetstream.ConsumerConfig{
		FilterSubject: "cgrates_cdrs_processed",
		Durable:       "cgrates_processed",
		AckPolicy:     jetstream.AckAllPolicy,
	})
	if err != nil {
		t.Error(err)
	}

	// Wait for the messages to be consumed and processed.
	time.Sleep(20 * time.Millisecond)

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
				{"tag": "Key", "type": "*variable", "value": "~*req.Key", "path": "*uch.Key"}
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
			"id": "nats_reader",
			"type": "*nats_json_map",
			"source_path": "%s",
			"ees_success_ids": ["nats_processed"],
			"flags": ["*dryrun"],
			"opts": {
				%s
			},
			"fields":[
				{"tag": "Key", "type": "*variable", "value": "~*req.Key", "path": "*cgreq.Key"},
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
	jwtFilePath := path.Join(baseJWTPath, "u.creds")
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
	jwtCfgPath := path.Join(baseJWTPath, "resolver.conf")
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
			cmd := exec.Command("nats-server", tc.serverFlags...)
			if err := cmd.Start(); err != nil {
				t.Fatal(err) // most probably not installed
			}
			t.Cleanup(func() { cmd.Process.Kill() })
			var nc *nats.Conn // to reuse nats conn

			ng := engine.TestEngine{
				ConfigJSON: fmt.Sprintf(natsCfg, tc.sourcePath, tc.readerOpts),
				PreStartHook: func(t testing.TB, c *config.CGRConfig) {
					rdrOpts := c.ERsCfg().ReaderCfg("nats_reader").Opts.NATS
					nop, err := GetNatsOpts(rdrOpts, c.GeneralCfg().NodeID, time.Second)
					if err != nil {
						t.Fatal(err)
					}
					nc = connectToNATSServer(t, tc.sourcePath, nop...)
				},
			}
			client, _ := ng.Run(t)

			// For non-jetstream connections, we need to make sure the
			// engine is ready to read published messages right away.
			time.Sleep(2 * time.Millisecond)

			for i := 0; i < 3; i++ {
				key := fmt.Sprintf("key%d", i+1)
				expData := fmt.Sprintf(`{"Key": "%s"}`, key)
				if err := nc.Publish("cgrates_cdrs", []byte(expData)); err != nil {
					t.Error(err)
				}
				checkNATSExports(t, client, key)
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
	jwtFilePath := path.Join(baseJWTPath, "u.creds")
	if err := os.WriteFile(jwtFilePath, []byte(`-----BEGIN NATS USER JWT-----
eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJIMkwzWEtSTVoyMklDSFBGSDRXQzM1U0hLRVY3RVZUTEJERlpESVhTN0xOWEhCNkhUV0ZBIiwiaWF0IjoxNzI2NzUyOTAzLCJpc3MiOiJBQktCWlJVVFY0M1NZN1E2VlA3TVBYTldBQzdOTVZCUzJGUEpRMzZHQ1dWNVZIRzVCVVlFNTRGSSIsInN1YiI6IlVCSUtHTlRPRFJPTU8yWEdZTk5TQ1hQNUNLSlBUSVRWUTY1TjdZVEZQWVFKNFdFWVY3T0xPTkE1IiwibmF0cyI6eyJwdWIiOnt9LCJzdWIiOnt9LCJzdWJzIjotMSwiZGF0YSI6LTEsInBheWxvYWQiOi0xLCJpc3N1ZXJfYWNjb3VudCI6IkFCVFZITzJNUkVOQlcyTk81N0tXSzZMWENaU0g0M09IUEtYRFlHWEJENlFZWlFDSkhGRURNNVFYIiwidHlwZSI6InVzZXIiLCJ2ZXJzaW9uIjoyfX0.rIFeciJthv_V4OfRc1wQXxk7-E3Wa6-87suJI_sn808Az7psEdvFagNosCqgdGd_d7AUDhY2eCipcIEZxnPeBA
------END NATS USER JWT------

************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN USER NKEY SEED-----
SUAD3BD2VMG6RJZVODS5BKYHOY6HGM7U5VDFYEEDHXIQXBX5R7RMKVQS4Y
------END USER NKEY SEED------

*************************************************************`), 0664); err != nil {
		t.Fatal(err)
	}
	jwtCfgPath := path.Join(baseJWTPath, "resolver.conf")
	if err := os.WriteFile(jwtCfgPath, []byte(`operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJQRVBMTFM3WFQyUzQ2NUZIS0JDSjdTNDdWS1VPRjJEVFhWRE9QVzZTWjY3R0JTU0I3SllBIiwiaWF0IjoxNzI2NzUyOTAzLCJpc3MiOiJPQzZJRFFQVTZJNFFOTVZYUEpSR1RSVFpHRFNTSEVSMzVMSkQ3V0k3RlJEUUpPT1ZUUjVKRjczWiIsIm5hbWUiOiJPIiwic3ViIjoiT0M2SURRUFU2STRRTk1WWFBKUkdUUlRaR0RTU0hFUjM1TEpEN1dJN0ZSRFFKT09WVFI1SkY3M1oiLCJuYXRzIjp7InNpZ25pbmdfa2V5cyI6WyJPQUdPWkNFRjNWR1FFV1RRTkFZTllJRFpFREVZSU03WDI3S0xKQzdaU1c3WDVQSjZJSllEWEgyUyJdLCJ0eXBlIjoib3BlcmF0b3IiLCJ2ZXJzaW9uIjoyfX0.KS5aRNmSKIcIDrEIZvpLPgmrwKGMExSTOnsp579ihwqRLFt5ZrDXEl8I81F-lgvRGZ0_yX4fUIQ9-J4trZDkDw

system_account: ADXMVBEVF2FHZTELG3CMW4VT4WRN6JK3SSXSS34G77BUX7YD6KWU2GMI

resolver: MEMORY
resolver_preload: {
    ADXMVBEVF2FHZTELG3CMW4VT4WRN6JK3SSXSS34G77BUX7YD6KWU2GMI: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiI0UzJRVU1JWDdGTjZJR0lUM0xTWERMN1ZaQTVJQlUyNjJFTkY3WTZGR1JMNDJWNE5YUkdBIiwiaWF0IjoxNzI2NzUyOTAzLCJpc3MiOiJPQUdPWkNFRjNWR1FFV1RRTkFZTllJRFpFREVZSU03WDI3S0xKQzdaU1c3WDVQSjZJSllEWEgyUyIsIm5hbWUiOiJTWVMiLCJzdWIiOiJBRFhNVkJFVkYyRkhaVEVMRzNDTVc0VlQ0V1JONkpLM1NTWFNTMzRHNzdCVVg3WUQ2S1dVMkdNSSIsIm5hdHMiOnsibGltaXRzIjp7InN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6LTEsImltcG9ydHMiOi0xLCJleHBvcnRzIjotMSwid2lsZGNhcmRzIjp0cnVlLCJjb25uIjotMSwibGVhZiI6LTF9LCJkZWZhdWx0X3Blcm1pc3Npb25zIjp7InB1YiI6eyJhbGxvdyI6WyIkU1lTLlx1MDAzZSJdfSwic3ViIjp7ImFsbG93IjpbIiRTWVMuXHUwMDNlIl19fSwiYXV0aG9yaXphdGlvbiI6e30sInR5cGUiOiJhY2NvdW50IiwidmVyc2lvbiI6Mn19.nZFNLl_sfCsaX2jTPFkCsHRbDNt0WGlpR0tx3K8J9KP8Ds8VmiQl7OvEmYjZflVKDVJgXvIwICIT-aY56klsAg
    ABTVHO2MRENBW2NO57KWK6LXCZSH43OHPKXDYGXBD6QYZQCJHFEDM5QX: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJKWFZETkU2NVhITFg0Qk03M0EzUDNIWkFZT0QyTFlZUUc3SElIMjQ1QUlHN1Y0NVBVSTRBIiwiaWF0IjoxNzI2NzUyOTAzLCJpc3MiOiJPQUdPWkNFRjNWR1FFV1RRTkFZTllJRFpFREVZSU03WDI3S0xKQzdaU1c3WDVQSjZJSllEWEgyUyIsIm5hbWUiOiJBIiwic3ViIjoiQUJUVkhPMk1SRU5CVzJOTzU3S1dLNkxYQ1pTSDQzT0hQS1hEWUdYQkQ2UVlaUUNKSEZFRE01UVgiLCJuYXRzIjp7ImxpbWl0cyI6eyJzdWJzIjotMSwiZGF0YSI6LTEsInBheWxvYWQiOi0xLCJpbXBvcnRzIjotMSwiZXhwb3J0cyI6LTEsIndpbGRjYXJkcyI6dHJ1ZSwiY29ubiI6LTEsImxlYWYiOi0xLCJtZW1fc3RvcmFnZSI6LTEsImRpc2tfc3RvcmFnZSI6LTEsInN0cmVhbXMiOi0xLCJjb25zdW1lciI6LTF9LCJzaWduaW5nX2tleXMiOlsiQUJLQlpSVVRWNDNTWTdRNlZQN01QWE5XQUM3Tk1WQlMyRlBKUTM2R0NXVjVWSEc1QlVZRTU0RkkiXSwiZGVmYXVsdF9wZXJtaXNzaW9ucyI6eyJwdWIiOnt9LCJzdWIiOnt9fSwiYXV0aG9yaXphdGlvbiI6e30sInR5cGUiOiJhY2NvdW50IiwidmVyc2lvbiI6Mn19.GWvAWjlECPNT5afsc96NVIEOJLFQi5fL9gBEzY4-z7mGyJN41qjJpx7l_LvLb8icToW2nca81J9NFwT5yf-NAA
}`), 0664); err != nil {
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
			cmd := exec.Command("nats-server", tc.serverFlags...)
			if err := cmd.Start(); err != nil {
				t.Fatal(err) // most probably not installed
			}
			t.Cleanup(func() { cmd.Process.Kill() })
			var js jetstream.JetStream // to reuse jetstream instance

			ng := engine.TestEngine{
				ConfigJSON: fmt.Sprintf(natsCfg, tc.sourcePath, tc.readerOpts),
				PreStartHook: func(t testing.TB, c *config.CGRConfig) {
					rdrOpts := c.ERsCfg().ReaderCfg("nats_reader").Opts.NATS
					nop, err := GetNatsOpts(rdrOpts, c.GeneralCfg().NodeID, time.Second)
					if err != nil {
						t.Fatal(err)
					}
					nc := connectToNATSServer(t, tc.sourcePath, nop...)

					// Initialize a stream manager and create a stream.
					js, err = jetstream.New(nc)
					if err != nil {
						t.Fatal(err)
					}
					if _, err = js.CreateStream(context.Background(), jetstream.StreamConfig{
						Name:     "stream",
						Subjects: []string{"cgrates_cdrs", "cgrates_cdrs_processed"},
					}); err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() {
						if err := js.DeleteStream(context.Background(), "stream"); err != nil {
							t.Error(err)
						}
					})
				},
			}
			client, _ := ng.Run(t)

			for i := 0; i < 3; i++ {
				key := fmt.Sprintf("key%d", i+1)
				expData := fmt.Sprintf(`{"Key": "%s"}`, key)
				if _, err := js.Publish(context.Background(), "cgrates_cdrs", []byte(expData)); err != nil {
					t.Error(err)
				}
				checkNATSExports(t, client, key)
			}
		})
	}
}

func connectToNATSServer(t testing.TB, url string, opts ...nats.Option) *nats.Conn {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	time.Sleep(5 * time.Millisecond) // takes around 5ms for the server to be available
	fib := utils.FibDuration(time.Millisecond, 0)
	for time.Now().Before(deadline) {
		nc, err := nats.Connect(url, opts...)
		if err == nil { // successfully connected
			t.Cleanup(func() {
				nc.Close()
			})
			return nc
		}
		time.Sleep(fib())
	}

	t.Fatalf("NATS server did not become available within %s", time.Second)
	return nil
}

func checkNATSExports(t *testing.T, client *birpc.Client, wantKey any) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	time.Sleep(2 * time.Millisecond) // takes around 1-2ms for the export to happen
	fib := utils.FibDuration(time.Millisecond, 0)

	itemID := "Key"
	var err error
	var key any
	for time.Now().Before(deadline) {
		err = client.Call(context.Background(), utils.CacheSv1GetItem,
			&utils.ArgsGetCacheItemWithAPIOpts{
				Tenant: "cgrates.org",
				ArgsGetCacheItem: utils.ArgsGetCacheItem{
					CacheID: utils.CacheUCH,
					ItemID:  itemID,
				},
			}, &key)

		if err == nil && key == wantKey {
			return
		}
		time.Sleep(fib())
	}

	if err != nil {
		t.Errorf("CacheSv1.GetItem(%q) unexpected err: %q", itemID, err)
		return
	}
	if key != wantKey {
		t.Errorf("CacheSv1.GetItem(%q)=%q, want %q", itemID, key, wantKey)
	}
}

/*

In order to generate the resolver.conf and u.creds for the jetstream test, run the following:

package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

func main() {
	// create an operator key pair (private key)
	okp, err := nkeys.CreateOperator()
	if err != nil {
		log.Fatal(err)
	}
	// extract the public key
	opk, err := okp.PublicKey()
	if err != nil {
		log.Fatal(err)
	}

	// create an operator claim using the public key for the identifier
	oc := jwt.NewOperatorClaims(opk)
	oc.Name = "O"
	// add an operator signing key to sign accounts
	oskp, err := nkeys.CreateOperator()
	if err != nil {
		log.Fatal(err)
	}
	// get the public key for the signing key
	ospk, err := oskp.PublicKey()
	if err != nil {
		log.Fatal(err)
	}
	// add the signing key to the operator - this makes any account
	// issued by the signing key to be valid for the operator
	oc.SigningKeys.Add(ospk)

	// self-sign the operator JWT - the operator trusts itself
	operatorJWT, err := oc.Encode(okp)
	if err != nil {
		log.Fatal(err)
	}

	// Create system account
	sysAkp, err := nkeys.CreateAccount()
	if err != nil {
		log.Fatal(err)
	}
	sysApk, err := sysAkp.PublicKey()
	if err != nil {
		log.Fatal(err)
	}
	sysAc := jwt.NewAccountClaims(sysApk)
	sysAc.Name = "SYS"
	// Add necessary system permissions
	sysAc.DefaultPermissions.Pub.Allow.Add("$SYS.>")
	sysAc.DefaultPermissions.Sub.Allow.Add("$SYS.>")
	sysAccountJWT, err := sysAc.Encode(oskp)
	if err != nil {
		log.Fatal(err)
	}

	// create an account keypair
	akp, err := nkeys.CreateAccount()
	if err != nil {
		log.Fatal(err)
	}
	// extract the public key for the account
	apk, err := akp.PublicKey()
	if err != nil {
		log.Fatal(err)
	}
	// create the claim for the account using the public key of the account
	ac := jwt.NewAccountClaims(apk)
	ac.Name = "A"

	// enable jetstream for account
	ac.Limits.JetStreamLimits = jwt.JetStreamLimits{
		MemoryStorage: -1,
		DiskStorage:   -1,
		Streams:       -1,
		Consumer:      -1,
	}

	// create a signing key that we can use for issuing users
	askp, err := nkeys.CreateAccount()
	if err != nil {
		log.Fatal(err)
	}
	// extract the public key
	aspk, err := askp.PublicKey()
	if err != nil {
		log.Fatal(err)
	}
	// add the signing key (public) to the account
	ac.SigningKeys.Add(aspk)

	// now we could encode an issue the account using the operator
	// key that we generated above, but this will illustrate that
	// the account could be self-signed, and given to the operator
	// who can then re-sign it
	accountJWT, err := ac.Encode(akp)
	if err != nil {
		log.Fatal(err)
	}

	// the operator would decode the provided token, if the token
	// is not self-signed or signed by an operator or tampered with
	// the decoding would fail
	ac, err = jwt.DecodeAccountClaims(accountJWT)
	if err != nil {
		log.Fatal(err)
	}
	// here the operator is going to use its private signing key to
	// re-issue the account
	accountJWT, err = ac.Encode(oskp)
	if err != nil {
		log.Fatal(err)
	}

	// now back to the account, the account can issue users
	// need not be known to the operator - the users are trusted
	// because they will be signed by the account. The server will
	// look up the account get a list of keys the account has and
	// verify that the user was issued by one of those keys
	ukp, err := nkeys.CreateUser()
	if err != nil {
		log.Fatal(err)
	}
	upk, err := ukp.PublicKey()
	if err != nil {
		log.Fatal(err)
	}
	uc := jwt.NewUserClaims(upk)
	// since the jwt will be issued by a signing key, the issuer account
	// must be set to the public ID of the account
	uc.IssuerAccount = apk
	userJwt, err := uc.Encode(askp)
	if err != nil {
		log.Fatal(err)
	}
	// the seed is a version of the keypair that is stored as text
	useed, err := ukp.Seed()
	if err != nil {
		log.Fatal(err)
	}
	// generate a creds formatted file that can be used by a NATS client
	creds, err := jwt.FormatUserConfig(userJwt, useed)
	if err != nil {
		log.Fatal(err)
	}

	// now we are going to put it together into something that can be run
	// we create a directory to store the server configuration, the creds
	// file and a small go program that uses the creds file
	dir, err := os.MkdirTemp(os.TempDir(), "jwt_example")
	if err != nil {
		log.Fatal(err)
	}
	// print where we generated the file
	fmt.Printf("cfg path: %s\n", dir)

	// we are generating a memory resolver server configuration
	// it lists the operator and all account jwts the server should
	// know about
	resolver := fmt.Sprintf(`operator: %s

system_account: %s

resolver: MEMORY
resolver_preload: {
    %s: %s
    %s: %s
}

jetstream: enabled
`, operatorJWT, sysApk, sysApk, sysAccountJWT, apk, accountJWT)
	if err := os.WriteFile(path.Join(dir, "resolver.conf"),
		[]byte(resolver), 0644); err != nil {
		log.Fatal(err)
	}

	// store the creds
	credsPath := path.Join(dir, "u.creds")
	if err := os.WriteFile(credsPath, creds, 0644); err != nil {
		log.Fatal(err)
	}
}

*/
