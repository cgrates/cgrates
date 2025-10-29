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

package engine

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAMQPPosterParseURL(t *testing.T) {
	amqp := &AMQPPoster{}
	expected := &AMQPPoster{
		dialURL:      "amqp://guest:guest@localhost:5672/?heartbeat=5",
		queueID:      "q1",
		exchange:     "E1",
		exchangeType: "fanout",
		routingKey:   "CGRCDR",
	}
	dialURL := "amqp://guest:guest@localhost:5672/?queue_id=q1&exchange=E1&routing_key=CGRCDR&heartbeat=5&exchange_type=fanout"
	if err := amqp.parseURL(dialURL); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, amqp) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(amqp))
	}
}

func TestKafkaParseURL(t *testing.T) {
	u := "127.0.0.1:9092?topic=cdr_billing"
	exp := &KafkaPoster{
		dialURL:  "127.0.0.1:9092",
		topic:    "cdr_billing",
		attempts: 10,
	}
	if kfk, err := NewKafkaPoster(u, 10); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, kfk) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(kfk))
	}
	u = "localhost:9092?topic=cdr_billing"
	exp = &KafkaPoster{
		dialURL:  "localhost:9092",
		topic:    "cdr_billing",
		attempts: 10,
	}
	if kfk, err := NewKafkaPoster(u, 10); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, kfk) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(kfk))
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		input       string
		expectedURL string
		expectedQID string
		expectError bool
	}{
		{
			input:       "http://cgrates.com/path?queue_id=myQueue",
			expectedURL: "http://cgrates.com/path",
			expectedQID: "myQueue",
			expectError: false,
		},
		{
			input:       "http://cgrates.com/path",
			expectedURL: "http://cgrates.com/path",
			expectedQID: defaultQueueID,
			expectError: false,
		},
		{
			input:       ":/invalid-url",
			expectedURL: "",
			expectedQID: "",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			url, qid, err := parseURL(test.input)
			if test.expectError {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if url != test.expectedURL {
					t.Fatalf("Expected URL to be %v, got %v", test.expectedURL, url)
				}
				if qid != test.expectedQID {
					t.Fatalf("Expected qID to be %v, got %v", test.expectedQID, qid)
				}
			}
		})
	}
}
