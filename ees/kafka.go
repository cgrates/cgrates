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
package ees

import (
	"context"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

// NewKafkaEE creates a kafka poster
func NewKafkaEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *KafkaEE {
	kfkPstr := &KafkaEE{
		cfg:   cfg,
		dc:    dc,
		topic: utils.DefaultQueueID,
		reqs:  newConcReq(cfg.ConcurrentRequests),
	}
	if vals, has := cfg.Opts[utils.KafkaTopic]; has {
		kfkPstr.topic = utils.IfaceAsString(vals)
	}
	return kfkPstr
}

// KafkaEE is a kafka poster
type KafkaEE struct {
	topic  string // identifier of the CDR queue where we publish
	writer *kafka.Writer

	cfg          *config.EventExporterCfg
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *KafkaEE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *KafkaEE) Connect() (_ error) {
	pstr.Lock()
	if pstr.writer == nil {
		pstr.writer = &kafka.Writer{
			Addr:        kafka.TCP(pstr.Cfg().ExportPath),
			Topic:       pstr.topic,
			MaxAttempts: pstr.Cfg().Attempts,
		}
		// pstr.writer = kafka.NewWriter(kafka.WriterConfig{
		// Brokers:     []string{pstr.Cfg().ExportPath},
		// MaxAttempts: pstr.Cfg().Attempts,
		// Topic:       pstr.topic,
		// })
	}
	pstr.Unlock()
	return
}

func (pstr *KafkaEE) ExportEvent(content interface{}, key string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	if pstr.writer == nil {
		pstr.RUnlock()
		pstr.reqs.done()
		return utils.ErrDisconnected
	}
	err = pstr.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: content.([]byte),
	})
	pstr.RUnlock()
	pstr.reqs.done()
	return
}

func (pstr *KafkaEE) Close() (err error) {
	pstr.Lock()
	if pstr.writer != nil {
		err = pstr.writer.Close()
		pstr.writer = nil
	}
	pstr.Unlock()
	return
}

func (pstr *KafkaEE) GetMetrics() *utils.SafeMapStorage { return pstr.dc }
