// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

// NewKafkaPoster creates a kafka poster
func NewKafkaPoster(dialURL string, attempts int) (*KafkaPoster, error) {
	kfkPstr := &KafkaPoster{
		attempts: attempts,
	}
	if err := kfkPstr.parseURL(dialURL); err != nil {
		return nil, err
	}
	return kfkPstr, nil
}

// KafkaPoster is a kafka poster
type KafkaPoster struct {
	dialURL    string
	topic      string // identifier of the CDR queue where we publish
	attempts   int
	sync.Mutex // protect writer
	writer     *kafka.Writer
}

func (pstr *KafkaPoster) parseURL(dialURL string) error {
	pstr.topic = defaultQueueID
	i := strings.IndexByte(dialURL, '?')
	if i < 0 {
		pstr.dialURL = dialURL
		return nil
	}
	pstr.dialURL = dialURL[:i]
	rawQuery := dialURL[i+1:]
	qry, err := url.ParseQuery(rawQuery)
	if err != nil {
		return err
	}
	pstr.dialURL = strings.Split(dialURL, "?")[0]
	if vals, has := qry[utils.KafkaTopic]; has && len(vals) != 0 {
		pstr.topic = vals[0]
	}
	return nil
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
		pstr.writer = nil
	}
	pstr.Unlock()
}

func (pstr *KafkaPoster) newPostWriter() {
	pstr.Lock()
	if pstr.writer == nil {
		pstr.writer = &kafka.Writer{
			Addr:        kafka.TCP(pstr.dialURL),
			Topic:       pstr.topic,
			MaxAttempts: pstr.attempts,
		}
	}
	pstr.Unlock()
}
