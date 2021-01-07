/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package engine

import (
	"context"
	"sync"

	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

// NewKafkaPoster creates a kafka poster
func NewKafkaPoster(dialURL string, attempts int, opts map[string]interface{}) *KafkaPoster {
	kfkPstr := &KafkaPoster{
		dialURL:  dialURL,
		attempts: attempts,
		topic:    utils.DefaultQueueID,
	}
	if vals, has := opts[utils.KafkaTopic]; has {
		kfkPstr.topic = utils.IfaceAsString(vals)
	}
	return kfkPstr
}

// KafkaPoster is a kafka poster
type KafkaPoster struct {
	dialURL    string
	topic      string // identifier of the CDR queue where we publish
	attempts   int
	sync.Mutex // protect writer
	writer     *kafka.Writer
}

// Post is the method being called when we need to post anything in the queue
// the optional chn will permits channel caching
func (pstr *KafkaPoster) Post(content []byte, key string) (err error) {
	pstr.newPostWriter()
	pstr.Lock()
	if err = pstr.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: content,
	}); err == nil {
		pstr.Unlock()
		return
	}
	pstr.Unlock()
	return
}

// Close closes the kafka writer
func (pstr *KafkaPoster) Close() {
	pstr.Lock()
	if pstr.writer != nil {
		pstr.writer.Close()
	}
	pstr.writer = nil
	pstr.Unlock()
}

func (pstr *KafkaPoster) newPostWriter() {
	pstr.Lock()
	if pstr.writer == nil {
		pstr.writer = kafka.NewWriter(kafka.WriterConfig{
			Brokers:     []string{pstr.dialURL},
			MaxAttempts: pstr.attempts,
			Topic:       pstr.topic,
		})
	}
	pstr.Unlock()
}
