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

// Most of the logic follows standard library implementation in this file
package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cenkalti/rpc2"
)

type concReqsBiJSONCoded struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer

	// temporary work space
	msg            message
	serverRequest  serverRequest
	clientResponse clientResponse

	// JSON-RPC clients can use arbitrary json values as request IDs.
	// Package rpc expects uint64 request IDs.
	// We assign uint64 sequence numbers to incoming requests
	// but save the original request ID in the pending map.
	// When rpc responds, we use the sequence number in
	// the response to find the original request ID.
	mutex   sync.Mutex // protects seq, pending
	pending map[uint64]*json.RawMessage
	seq     uint64
}

// NewConcReqsBiJSONCoded returns a new rpc2.Codec using JSON-RPC on conn.
func NewConcReqsBiJSONCoded(conn io.ReadWriteCloser) rpc2.Codec {
	return &concReqsBiJSONCoded{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]*json.RawMessage),
	}
}

// serverRequest and clientResponse combined
type message struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

type clientResponse struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

type clientRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Id     *uint64       `json:"id"`
}

func (c *concReqsBiJSONCoded) ReadHeader(req *rpc2.Request, resp *rpc2.Response) error {
	if err := ConReqs.Allocate(); err != nil {
		return err
	}
	c.msg = message{}
	if err := c.dec.Decode(&c.msg); err != nil {
		return err
	}

	if c.msg.Method != "" {
		// request comes to server
		c.serverRequest.Id = c.msg.Id
		c.serverRequest.Method = c.msg.Method
		c.serverRequest.Params = c.msg.Params

		req.Method = c.serverRequest.Method

		// JSON request id can be any JSON value;
		// RPC package expects uint64.  Translate to
		// internal uint64 and save JSON on the side.
		if c.serverRequest.Id == nil {
			// Notification
		} else {
			c.mutex.Lock()
			c.seq++
			c.pending[c.seq] = c.serverRequest.Id
			c.serverRequest.Id = nil
			req.Seq = c.seq
			c.mutex.Unlock()
		}
	} else {
		// response comes to client
		err := json.Unmarshal(*c.msg.Id, &c.clientResponse.Id)
		if err != nil {
			return err
		}
		c.clientResponse.Result = c.msg.Result
		c.clientResponse.Error = c.msg.Error

		resp.Error = ""
		resp.Seq = c.clientResponse.Id
		if c.clientResponse.Error != nil || c.clientResponse.Result == nil {
			x, ok := c.clientResponse.Error.(string)
			if !ok {
				return fmt.Errorf("invalid error %v", c.clientResponse.Error)
			}
			if x == "" {
				x = "unspecified error"
			}
			resp.Error = x
		}
	}
	return nil
}

func (c *concReqsBiJSONCoded) ReadRequestBody(x interface{}) error {
	if x == nil {
		return nil
	}
	if c.serverRequest.Params == nil {
		return errMissingParams
	}
	var params *[]interface{}
	switch x := x.(type) {
	case *[]interface{}:
		params = x
	default:
		params = &[]interface{}{x}
	}
	return json.Unmarshal(*c.serverRequest.Params, params)
}

func (c *concReqsBiJSONCoded) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}
	return json.Unmarshal(*c.clientResponse.Result, x)
}

func (c *concReqsBiJSONCoded) WriteRequest(r *rpc2.Request, param interface{}) error {
	req := &clientRequest{Method: r.Method}
	switch param := param.(type) {
	case []interface{}:
		req.Params = param
	default:
		req.Params = []interface{}{param}
	}
	if r.Seq == 0 {
		// Notification
		req.Id = nil
	} else {
		seq := r.Seq
		req.Id = &seq
	}
	return c.enc.Encode(req)
}

func (c *concReqsBiJSONCoded) WriteResponse(r *rpc2.Response, x interface{}) error {
	defer ConReqs.Deallocate()
	c.mutex.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		// Invalid request so no id.  Use JSON null.
		b = &null
	}
	resp := serverResponse{Id: b}
	if r.Error == "" {
		resp.Result = x
	} else {
		resp.Error = r.Error
	}
	return c.enc.Encode(resp)
}

func (c *concReqsBiJSONCoded) Close() error {
	return c.c.Close()
}
