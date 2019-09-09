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
package engine

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

// "amqp://guest:guest@localhost:5672/?topic=cgrates_cdrs"
func NewKafkaPoster(dialURL string, attempts int, fallbackFileDir string) (*KafkaPoster, error) {
	amqp := &KafkaPoster{
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}
	if err := amqp.parseURL(dialURL); err != nil {
		return nil, err
	}
	return amqp, nil
}

type KafkaPoster struct {
	dialURL         string
	topic           string // identifier of the CDR queue where we publish
	attempts        int
	fallbackFileDir string
	sync.Mutex      // protect writer
	writer          *kafka.Writer
}

func (pstr *KafkaPoster) parseURL(dialURL string) error {
	u, err := url.Parse(dialURL)
	if err != nil {
		return err
	}
	qry := u.Query()

	pstr.dialURL = strings.Split(dialURL, "?")[0]
	pstr.topic = defaultQueueID
	if vals, has := qry[utils.KafkaTopic]; has && len(vals) != 0 {
		pstr.topic = vals[0]
	}
	return nil
}

// Post is the method being called when we need to post anything in the queue
// the optional chn will permits channel caching
func (pstr *KafkaPoster) Post(content []byte, fallbackFileName, key string) (err error) {
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
	if err != nil && fallbackFileName != utils.META_NONE {
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		return err
	}
	return
}

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
