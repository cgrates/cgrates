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
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/segmentio/kafka-go"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// TestKafkaSSL tests exporting to and reading from a kafka broker through an SSL connection.
// Steps to set up a local kafka server with SSL setup can be found at the bottom of the file.
func TestKafkaSSL(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	brokerPlainURL := "localhost:9092"
	brokerSSLURL := "localhost:9094"
	mainTopic := "cgrates-cdrs"
	processedTopic := "processed"

	content := fmt.Sprintf(`{

"general": {
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
	// "cache": {
	// 	"*kafka_json_map": {"limit": -1, "ttl": "5s", "precache": false},
	// },
    "exporters": [
        {
            "id": "kafka_ssl",								
            "type": "*kafka_json_map",									
            "export_path": "%s",
			"synchronous": true,
            "opts": {
                "kafkaTopic": "%s",
				"kafkaBatchSize": 1,
                "kafkaTLS": true,
                "kafkaCAPath": "/tmp/ssl/kafka/ca.crt",
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
                "kafkaTopic": "%s",
				"kafkaBatchSize": 1
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
                "kafkaCAPath": "/tmp/ssl/kafka/ca.crt",
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

}`, brokerSSLURL, mainTopic, brokerPlainURL, processedTopic, brokerSSLURL, mainTopic)

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	createKafkaTopics(t, brokerPlainURL, true, mainTopic, processedTopic)

	// export event to cgrates-cdrs topic, then the reader will consume it and
	// export it to the 'processed' topic
	t.Run("export kafka event", func(t *testing.T) {
		n := 1
		var wg sync.WaitGroup
		wg.Add(n)

		var reply map[string]map[string]interface{}
		for range n {
			go func() {
				defer wg.Done()
				if err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
					&engine.CGREventWithEeIDs{
						EeIDs: []string{"kafka_ssl"},
						CGREvent: &utils.CGREvent{
							Tenant: "cgrates.org",
							ID:     "KafkaEvent",
							Event: map[string]interface{}{
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
			}()
		}
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Errorf("timed out waiting for %s replies", utils.EeSv1ProcessEvent)
		}
	})

	// Check whether ERs managed to successfully consume the event from the
	// 'cgrates-cdrs' topic and exported it to the 'processed' topic.
	t.Run("verify kafka export", func(t *testing.T) {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{brokerPlainURL},
			Topic:   "processed",
			MaxWait: time.Millisecond,
		})
		t.Cleanup(func() {
			if err := r.Close(); err != nil {
				t.Error("failed to close reader:", err)
			}
		})

		want := `{"Account":"1001","Destination":"1002","OriginID":"abcdef","ToR":"*voice","Usage":"10000000000"}`

		readErr := make(chan error)
		msg := make(chan string)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			m, err := r.FetchMessage(ctx)
			if err != nil {
				readErr <- err
				return
			}
			msg <- string(m.Value)
		}()

		select {
		case err := <-readErr:
			t.Errorf("kafka.Reader.ReadMessage() failed unexpectedly: %v", err)
		case got := <-msg:
			if got != want {
				t.Errorf("kafka.Reader.ReadMessage() = %v, want %v", got, want)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("kafka.Reader.ReadMessage() took too long (>%s)", 2*time.Second)
		}
	})
}

func createKafkaTopics(t *testing.T, brokerURL string, cleanup bool, topics ...string) {
	t.Helper()
	conn, err := kafka.Dial("tcp", brokerURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	controller, err := conn.Controller()
	if err != nil {
		t.Fatal(err)
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		t.Fatal(err)
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, 0, len(topics))
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	if err := controllerConn.CreateTopics(topicConfigs...); err != nil {
		t.Fatal(err)
	}

	if cleanup {
		t.Cleanup(func() {
			if err := conn.DeleteTopics(topics...); err != nil {
				t.Log(err)
			}
		})
	}
}

// Kafka broker has the following configuration:

/*
/opt/kafka/config/kraft/server.properties
----------------------------------------

listeners=PLAINTEXT://:9092,CONTROLLER://:9093,SSL://:9094
...
advertised.listeners=PLAINTEXT://localhost:9092,SSL://localhost:9094
...
ssl.truststore.location=/tmp/ssl/kafka/kafka.truststore.jks
ssl.truststore.password=123456
ssl.keystore.location=/tmp/ssl/kafka/kafka.keystore.jks
ssl.keystore.password=123456
ssl.key.password=123456
ssl.client.auth=none
*/

// Script to generate TLS keys and certificates:

/*
#!/bin/bash

mkdir -p /tmp/ssl/kafka

# Generate CA key
openssl genpkey -algorithm RSA -out /tmp/ssl/kafka/ca.key

# Generate CA certificate
openssl req -x509 -new -key /tmp/ssl/kafka/ca.key -days 3650 -out /tmp/ssl/kafka/ca.crt \
-subj "/C=US/ST=California/L=San Francisco/O=MyOrg/CN=localhost/emailAddress=example@email.com"

# Generate server key and CSR
openssl req -new -newkey rsa:4096 -nodes -keyout /tmp/ssl/kafka/server.key \
-out /tmp/ssl/kafka/server.csr \
-subj "/C=US/ST=California/L=San Francisco/O=MyOrg/CN=localhost/emailAddress=example@email.com"

# Create SAN configuration file
echo "authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1=localhost
IP.1=127.0.0.1
" > /tmp/ssl/kafka/san.cnf

# Sign server certificate with CA
openssl x509 -req -in /tmp/ssl/kafka/server.csr -CA /tmp/ssl/kafka/ca.crt \
-CAkey /tmp/ssl/kafka/ca.key -CAcreateserial -out /tmp/ssl/kafka/server.crt \
-days 3650 -extfile /tmp/ssl/kafka/san.cnf

# Convert server certificate and key to PKCS12 format
openssl pkcs12 -export \
    -in /tmp/ssl/kafka/server.crt \
    -inkey /tmp/ssl/kafka/server.key \
    -name kafka-broker \
    -out /tmp/ssl/kafka/kafka.p12 \
    -password pass:123456

# Import PKCS12 file into Java keystore
keytool -importkeystore \
    -srckeystore /tmp/ssl/kafka/kafka.p12 \
    -destkeystore /tmp/ssl/kafka/kafka.keystore.jks \
    -srcstoretype pkcs12 \
    -srcstorepass 123456 \
    -deststorepass 123456 \
    -noprompt

# Create truststore and import CA certificate
keytool -keystore /tmp/ssl/kafka/kafka.truststore.jks -alias CARoot -import -file /tmp/ssl/kafka/ca.crt \
    -storepass 123456 \
    -noprompt
*/
