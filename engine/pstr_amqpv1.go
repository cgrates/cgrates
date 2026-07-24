// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/cgrates/utils"
)

// NewAMQPv1Poster creates a poster for amqpv1
func NewAMQPv1Poster(dialURL string, attempts int) (Poster, error) {
	URL, qID, err := parseURL(dialURL)
	if err != nil {
		return nil, err
	}
	return &AMQPv1Poster{
		dialURL:  URL,
		queueID:  "/" + qID,
		attempts: attempts,
	}, nil
}

// AMQPv1Poster a poster for amqpv1
type AMQPv1Poster struct {
	sync.Mutex
	dialURL  string
	queueID  string // identifier of the CDR queue where we publish
	attempts int
	conn     *amqpv1.Conn
}

// Close closes the connections
func (pstr *AMQPv1Poster) Close() {
	pstr.Lock()
	if pstr.conn != nil {
		pstr.conn.Close()
	}
	pstr.conn = nil
	pstr.Unlock()
}

// Post is the method being called when we need to post anything in the queue
func (pstr *AMQPv1Poster) Post(content []byte, _ string) (err error) {
	var s *amqpv1.Session
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if s, err = pstr.newPosterSession(); err == nil {
			break
		}
		// reset client and try again
		// used in case of closed connection because of idle time
		if pstr.conn != nil {
			pstr.conn.Close() // Make shure the connection is closed before reseting it
		}
		pstr.conn = nil
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<AMQPv1Poster> creating new post channel, err: %s", err.Error()))
		return err
	}

	ctx := context.Background()
	for i := 0; i < pstr.attempts; i++ {
		sender, err := s.NewSender(ctx, pstr.queueID, nil)
		if err != nil {
			if i+1 < pstr.attempts {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
			continue
		}
		// Send message
		err = sender.Send(ctx, amqpv1.NewMessage(content), nil)
		sender.Close(ctx)
		if err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		return
	}
	if s != nil {
		s.Close(ctx)
	}
	return
}

func (pstr *AMQPv1Poster) newPosterSession() (s *amqpv1.Session, err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.conn == nil {
		var client *amqpv1.Conn
		client, err = amqpv1.Dial(context.TODO(), pstr.dialURL, nil)
		if err != nil {
			return nil, err
		}
		pstr.conn = client
	}
	return pstr.conn.NewSession(context.TODO(), nil)
}
