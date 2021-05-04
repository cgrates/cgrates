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
	"fmt"
	"sync"
	"time"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/cgrates/utils"
)

// NewAMQPv1Poster creates a poster for amqpv1
func NewAMQPv1Poster(dialURL string, attempts int, opts map[string]interface{}) Poster {
	pstr := &AMQPv1Poster{
		dialURL:  dialURL,
		queueID:  "/" + utils.DefaultQueueID,
		attempts: attempts,
	}
	if vals, has := opts[utils.AMQPQueueID]; has {
		pstr.queueID = "/" + utils.IfaceAsString(vals)
	}
	return pstr
}

// AMQPv1Poster a poster for amqpv1
type AMQPv1Poster struct {
	sync.Mutex
	dialURL  string
	queueID  string // identifier of the CDR queue where we publish
	attempts int
	client   *amqpv1.Client
}

// Close closes the connections
func (pstr *AMQPv1Poster) Close() {
	pstr.Lock()
	if pstr.client != nil {
		pstr.client.Close()
	}
	pstr.client = nil
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
		if pstr.client != nil {
			pstr.client.Close() // Make shure the connection is closed before reseting it
		}
		pstr.client = nil
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
		sender, err := s.NewSender(
			amqpv1.LinkTargetAddress(pstr.queueID),
		)
		if err != nil {
			if i+1 < pstr.attempts {
				time.Sleep(time.Duration(fib()) * time.Second)
			}
			// if pstr.isRecoverableError(err) {
			// 	s.Close(ctx)
			// 	pstr.client.Close()
			// 	pstr.client = nil
			// 	stmp, err := pstr.newPosterSession()
			// 	if err == nil {
			// 		s = stmp
			// 	}
			// }
			continue
		}
		// Send message
		err = sender.Send(ctx, amqpv1.NewMessage(content))
		sender.Close(ctx)
		if err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
		// if pstr.isRecoverableError(err) {
		// 	s.Close(ctx)
		// 	pstr.client.Close()
		// 	pstr.client = nil
		// 	stmp, err := pstr.newPosterSession()
		// 	if err == nil {
		// 		s = stmp
		// 	}
		// }
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
	if pstr.client == nil {
		var client *amqpv1.Client
		client, err = amqpv1.Dial(pstr.dialURL)
		if err != nil {
			return nil, err
		}
		pstr.client = client
	}
	return pstr.client.NewSession()
}
