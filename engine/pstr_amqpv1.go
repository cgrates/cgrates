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
	"net"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	amqpv1 "pack.ag/amqp"
)

func NewAMQPv1Poster(dialURL string, attempts int, fallbackFileDir string) (Poster, error) {
	URL, qID, err := parseURL(dialURL)
	if err != nil {
		return nil, err
	}
	return &AMQPv1Poster{
		dialURL:         URL,
		queueID:         "/" + qID,
		attempts:        attempts,
		fallbackFileDir: fallbackFileDir,
	}, nil
}

type AMQPv1Poster struct {
	sync.Mutex
	dialURL         string
	queueID         string // identifier of the CDR queue where we publish
	attempts        int
	fallbackFileDir string
	client          *amqpv1.Client
}

func (pstr *AMQPv1Poster) Close() {
	pstr.Lock()
	if pstr.client != nil {
		pstr.client.Close()
	}
	pstr.client = nil
	pstr.Unlock()
}

func (pstr *AMQPv1Poster) Post(content []byte, fallbackFileName, _ string) (err error) {
	var s *amqpv1.Session
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if s, err = pstr.newPosterSession(); err == nil {
			break
		}
		// reset client and try again
		// used in case of closed conection because of idle time
		if pstr.client != nil {
			pstr.client.Close() // Make shure the connection is closed before reseting it
		}
		pstr.client = nil
		time.Sleep(time.Duration(fib()) * time.Second)
	}
	if err != nil {
		if fallbackFileName != utils.META_NONE {
			utils.Logger.Warning(fmt.Sprintf("<AMQPv1Poster> creating new post channel, err: %s", err.Error()))
			err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		}
		return err
	}

	ctx := context.Background()
	for i := 0; i < pstr.attempts; i++ {
		sender, err := s.NewSender(
			amqpv1.LinkTargetAddress(pstr.queueID),
		)
		if err != nil {
			time.Sleep(time.Duration(fib()) * time.Second)
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
		time.Sleep(time.Duration(fib()) * time.Second)
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
	if err != nil && fallbackFileName != utils.META_NONE {
		err = writeToFile(pstr.fallbackFileDir, fallbackFileName, content)
		return err
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

func isRecoverableCloseError(err error) bool {
	return err == amqpv1.ErrConnClosed ||
		err == amqpv1.ErrLinkClosed ||
		err == amqpv1.ErrSessionClosed
}

func (pstr *AMQPv1Poster) isRecoverableError(err error) bool {
	switch err.(type) {
	case *amqpv1.Error, *amqpv1.DetachError, net.Error:
		if netErr, ok := err.(net.Error); ok {
			if !netErr.Temporary() {
				return false
			}
		}
	default:
		if !isRecoverableCloseError(err) {
			return false
		}
	}
	return true
}
