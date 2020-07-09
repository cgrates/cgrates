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

package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/rpc"
	"strings"
	"sync"
)

var errMissingParams = errors.New("jsonrpc: request body missing params")

type MethodParameters struct {
	Method     string
	Parameters interface{}
}

type jsonServerCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer

	// temporary work space
	req serverRequest

	// JSON-RPC clients can use arbitrary json values as request IDs.
	// Package rpc expects uint64 request IDs.
	// We assign uint64 sequence numbers to incoming requests
	// but save the original request ID in the pending map.
	// When rpc responds, we use the sequence number in
	// the response to find the original request ID.
	mutex   sync.Mutex // protects seq, pending
	seq     uint64
	pending map[uint64]*json.RawMessage

	allocated bool // populated if we have allocated a channel for concurrent requests
}

// NewCustomJSONServerCodec is used only when DispatcherS is active to handle APIer methods generically
func NewCustomJSONServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &jsonServerCodec{
		dec:     json.NewDecoder(conn),
		enc:     json.NewEncoder(conn),
		c:       conn,
		pending: make(map[uint64]*json.RawMessage),
	}
}

type serverRequest struct {
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
	Id      *json.RawMessage `json:"id"`
	isApier bool
}

func (r *serverRequest) reset() {
	r.Method = ""
	r.Params = nil
	r.Id = nil
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result"`
	Error  interface{}      `json:"error"`
}

func (c *jsonServerCodec) ReadRequestHeader(r *rpc.Request) error {
	c.req.reset()
	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}
	// in case we get a request with APIerSv1 or APIerSv2 we redirect
	// to Dispatcher to send it according to ArgDispatcher
	if c.req.isApier = strings.HasPrefix(c.req.Method, ApierV); c.req.isApier {
		r.ServiceMethod = DispatcherSv1Apier
	} else {
		r.ServiceMethod = c.req.Method
	}

	// JSON request id can be any JSON value;
	// RPC package expects uint64.  Translate to
	// internal uint64 and save JSON on the side.
	c.mutex.Lock()
	c.seq++
	c.pending[c.seq] = c.req.Id
	c.req.Id = nil
	r.Seq = c.seq
	c.mutex.Unlock()

	return nil
}

func (c *jsonServerCodec) ReadRequestBody(x interface{}) error {
	if err := ConReqs.Allocate(); err != nil {
		return err
	}
	c.allocated = true
	if x == nil {
		return nil
	}
	if c.req.Params == nil {
		return errMissingParams
	}
	// following example from ReadRequestHeader in case we get APIerSv1
	// or APIerSv2 we compose the parameters
	if c.req.isApier {
		cx := x.(*MethodParameters)
		cx.Method = c.req.Method
		var params [1]interface{}
		params[0] = &cx.Parameters
		return json.Unmarshal(*c.req.Params, &params)
	}
	// JSON params is array value.
	// RPC params is struct.
	// Unmarshal into array containing struct for now.
	// Should think about making RPC more general.
	var params [1]interface{}
	params[0] = x
	return json.Unmarshal(*c.req.Params, &params)

}

var null = json.RawMessage([]byte("null"))

func (c *jsonServerCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if c.allocated {
		defer func() {
			ConReqs.Deallocate()
			c.allocated = false
		}()
	}
	c.mutex.Lock()
	b, ok := c.pending[r.Seq]
	if !ok {
		c.mutex.Unlock()
		return errors.New("invalid sequence number in response")
	}
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	if b == nil {
		// Invalid request so no id. Use JSON null.
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

func (c *jsonServerCodec) Close() error {
	return c.c.Close()
}
