//go:build kafka

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

package general_tests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

// TestKafkaSSL tests exporting to and reading from a kafka broker through SSL.
//
//	go test -tags=kafka ./general_tests/ -run TestKafkaSSL -dbtype "*internal" -v
func TestKafkaSSL(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	brokerPlainURL := "localhost:19092"
	brokerSSLURL := "localhost:19094"
	mainTopic := "cgrates-cdrs"
	processedTopic := "processed"

	caPath := startKafkaSSLBroker(t, brokerPlainURL, brokerSSLURL)

	createKafkaTopics(t, brokerPlainURL, true, mainTopic, processedTopic)

	content := fmt.Sprintf(`{
"ees": {
    "enabled": true,
    "exporters": [
        {
            "id": "kafka_ssl",
            "type": "*kafka_json_map",
            "export_path": "%s",
            "synchronous": true,
            "opts": {
                "kafkaTopic": "%s",
                "kafkaTLS": true,
                "kafkaCAPath": "%s",
                "kafkaSkipTLSVerify": false
            },
            "failed_posts_dir": "*none"
        },
        {
            "id": "kafka_processed",
            "type": "*kafka_json_map",
            "export_path": "%s",
            "synchronous": true,
            "opts": {
                "kafkaTopic": "%s"
            },
            "failed_posts_dir": "*none"
        }
    ]
},
"ers": {
    "enabled": true,
    "sessions_conns":[],
    "ees_conns": ["*localhost"],
    "readers": [
        {
            "id": "kafka_ssl",
            "type": "*kafka_json_map",
            "run_delay": "-1",
            "flags": ["*dryrun"],
            "source_path": "%s",
            "ees_success_ids": ["kafka_processed"],
            "opts": {
                "kafkaTopic": "%s",
                "kafkaGroupID": "",
                "kafkaTLS": true,
                "kafkaCAPath": "%s",
                "kafkaSkipTLSVerify": false
            },
            "fields": [
                {"tag": "ToR", "path": "*cgreq.ToR", "type": "*variable", "value": "~*req.ToR", "mandatory": true},
                {"tag": "OriginID", "path": "*cgreq.OriginID", "type": "*variable", "value": "~*req.OriginID", "mandatory": true},
                {"tag": "Account", "path": "*cgreq.Account", "type": "*variable", "value": "~*req.Account", "mandatory": true},
                {"tag": "Destination", "path": "*cgreq.Destination", "type": "*variable", "value": "~*req.Destination", "mandatory": true},
                {"tag": "Usage", "path": "*cgreq.Usage", "type": "*variable", "value": "~*req.Usage", "mandatory": true}
            ]
        }
    ]
}
}`, brokerSSLURL, mainTopic, caPath,
		brokerPlainURL, processedTopic,
		brokerSSLURL, mainTopic, caPath)

	ng := engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      engine.InternalDBCfg,
	}
	client, _ := ng.Run(t)

	t.Run("export kafka event", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var reply map[string]map[string]any
		if err := client.Call(ctx, utils.EeSv1ProcessEvent,
			&engine.CGREventWithEeIDs{
				EeIDs: []string{"kafka_ssl"},
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
			}, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("verify kafka export", func(t *testing.T) {
		cl, err := kgo.NewClient(
			kgo.SeedBrokers(brokerPlainURL),
			kgo.ConsumeTopics("processed"),
			kgo.FetchMaxWait(10*time.Millisecond),
		)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { cl.Close() })

		want := `{"Account":"1001","Destination":"1002","OriginID":"abcdef","ToR":"*voice","Usage":"10000000000"}`

		readErr := make(chan error, 1)
		msg := make(chan string, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			fetches := cl.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				readErr <- errs[0].Err
				return
			}
			fetches.EachRecord(func(r *kgo.Record) {
				msg <- string(r.Value)
			})
		}()

		select {
		case err := <-readErr:
			t.Errorf("PollFetches failed unexpectedly: %v", err)
		case got := <-msg:
			if got != want {
				t.Errorf("PollFetches = %v, want %v", got, want)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("PollFetches took too long (>%s)", 2*time.Second)
		}
	})
}

func createKafkaTopics(tb testing.TB, brokerURL string, cleanup bool, topics ...string) {
	tb.Helper()
	cl, err := kgo.NewClient(kgo.SeedBrokers(brokerURL))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { cl.Close() })

	adm := kadm.NewClient(cl)
	_, err = adm.CreateTopics(context.Background(), 1, 1, nil, topics...)
	if err != nil {
		tb.Fatal(err)
	}

	if cleanup {
		tb.Cleanup(func() {
			if _, err := adm.DeleteTopics(context.Background(), topics...); err != nil {
				tb.Log(err)
			}
		})
	}
}

func startKafkaSSLBroker(t *testing.T, plainURL, sslURL string) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err) // rootless podman needs to traverse this
	}

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTpl, caTpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, _ := x509.ParseCertificate(caDER)

	srvKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	srvTpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	srvDER, err := x509.CreateCertificate(rand.Reader, srvTpl, caCert, &srvKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}

	writePEM := func(name string, blocks ...*pem.Block) {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range blocks {
			if err := pem.Encode(f, b); err != nil {
				f.Close()
				t.Fatal(err)
			}
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	srvKeyDER, _ := x509.MarshalPKCS8PrivateKey(srvKey)
	writePEM("keystore.pem",
		&pem.Block{Type: "PRIVATE KEY", Bytes: srvKeyDER},
		&pem.Block{Type: "CERTIFICATE", Bytes: srvDER},
		&pem.Block{Type: "CERTIFICATE", Bytes: caDER},
	)
	writePEM("truststore.pem",
		&pem.Block{Type: "CERTIFICATE", Bytes: caDER},
	)
	caPath := filepath.Join(dir, "ca.crt")
	writePEM("ca.crt",
		&pem.Block{Type: "CERTIFICATE", Bytes: caDER},
	)

	_, plainPort, _ := net.SplitHostPort(plainURL)
	_, sslPort, _ := net.SplitHostPort(sslURL)
	ctrlPort := "19093"

	props := fmt.Sprintf(`node.id=1
process.roles=broker,controller
listeners=PLAINTEXT://0.0.0.0:%s,SSL://0.0.0.0:%s,CONTROLLER://0.0.0.0:%s
advertised.listeners=PLAINTEXT://localhost:%s,SSL://localhost:%s
controller.quorum.voters=1@localhost:%s
controller.listener.names=CONTROLLER
listener.security.protocol.map=PLAINTEXT:PLAINTEXT,SSL:SSL,CONTROLLER:PLAINTEXT
offsets.topic.replication.factor=1
log.dirs=/tmp/kraft-combined-logs
ssl.keystore.type=PEM
ssl.keystore.location=/etc/kafka/secrets/keystore.pem
ssl.truststore.type=PEM
ssl.truststore.location=/etc/kafka/secrets/truststore.pem
ssl.client.auth=none
ssl.endpoint.identification.algorithm=
`, plainPort, sslPort, ctrlPort, plainPort, sslPort, ctrlPort)
	if err := os.WriteFile(filepath.Join(dir, "server.properties"), []byte(props), 0644); err != nil {
		t.Fatal(err)
	}

	containerName := fmt.Sprintf("kafka-ssl-test-%d", time.Now().UnixNano()%100000)
	volumeName := containerName + "-data"
	_ = exec.Command("podman", "rm", "-f", containerName).Run()

	out, err := exec.Command("podman", "run", "-d", "--name", containerName,
		"-p", plainPort+":"+plainPort,
		"-p", sslPort+":"+sslPort,
		"-v", dir+":/etc/kafka/secrets:ro,z",
		"-v", volumeName+":/tmp/kraft-combined-logs",
		"docker.io/apache/kafka:4.2.0",
		"bash", "-c",
		"/opt/kafka/bin/kafka-storage.sh format -t ssl-test-cluster-01 -c /etc/kafka/secrets/server.properties && exec /opt/kafka/bin/kafka-server-start.sh /etc/kafka/secrets/server.properties",
	).CombinedOutput()
	if err != nil {
		t.Fatalf("podman run: %s: %v", out, err)
	}
	t.Cleanup(func() {
		_ = exec.Command("podman", "rm", "-f", containerName).Run()
		_ = exec.Command("podman", "volume", "rm", volumeName).Run()
	})

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := tls.DialWithDialer(
			&net.Dialer{Timeout: time.Second},
			"tcp", sslURL,
			&tls.Config{RootCAs: certPoolFrom(caCert)},
		)
		if err == nil {
			conn.Close()
			return caPath
		}
		time.Sleep(500 * time.Millisecond)
	}

	logs, _ := exec.Command("podman", "logs", "--tail", "20", containerName).CombinedOutput()
	t.Fatalf("kafka SSL did not become ready within 30s. Logs:\n%s", logs)
	return "" // unreachable
}

func certPoolFrom(cert *x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return pool
}
