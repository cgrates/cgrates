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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
	"reflect"
	"testing"
)

func TestNewCustomJSONServerCodec(t *testing.T) {

	reader, writer := io.Pipe()

	codec := NewCustomJSONServerCodec(struct {
		io.Reader
		io.Writer
		io.Closer
	}{Reader: reader, Writer: writer, Closer: writer})

	c := codec.(*jsonServerCodec)

	if c.dec == nil {
		t.Error("Decoder is nil")
	}
	if c.enc == nil {
		t.Error("Encoder is nil")
	}
	if c.c == nil {
		t.Error("Connection is nil")
	}
	if c.pending == nil {
		t.Error("Pending map is nil")
	}

	writer.Close()
}

func TestJSONCodecReset(t *testing.T) {

	r := serverRequest{
		Method:  "test",
		Params:  &json.RawMessage{},
		Id:      &json.RawMessage{},
		isApier: false,
	}

	r.reset()

	if r.Method != "" {
		t.Error("Method didn't clear")
	}
	if r.Params != nil {
		t.Error("Params didn't clear")
	}
	if r.Id != nil {
		t.Error("Id didn't clear")
	}
}

type bufferClose struct {
	*bytes.Buffer
}

func (b *bufferClose) Close() error {
	return nil
}

func TestJSONCodecReadRequestHeader2(t *testing.T) {
	tests := []struct {
		name  string
		write []byte
		exp   error
	}{
		{
			name: "ValidRequest",
			write: []byte(`{
				"method": "APIerSv1.RemTP",
				"params": [{
					"TPid": "test"
				}],
				"id": 0
			}`),
			exp: nil,
		},
		{
			name:  "InvalidRequest",
			write: []byte("test"),
			exp:   fmt.Errorf("invalid character 'e' in literal true (expecting 'r')"),
		},
		{
			name: "ValidApierRequest",
			write: []byte(`{
				"method": "ApierV.Method",
				"params": [{
					"TPid": "test"
				}],
				"id": 0
			}`),
			exp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &bufferClose{
				Buffer: &bytes.Buffer{},
			}
			s := NewCustomJSONServerCodec(conn)

			rr := rpc.Request{
				ServiceMethod: "Service.Method",
				Seq:           1,
			}

			conn.Write(tt.write)

			err := s.ReadRequestHeader(&rr)
			if err != nil {
				if err.Error() != tt.exp.Error() {
					t.Errorf("Received an error: %v", err)
				}
			}
		})
	}
}

func TestJSONReadRequestBody(t *testing.T) {

	tests := []struct {
		name    string
		arg     any
		exp     string
		params  bool
		isApier bool
	}{
		{
			name:    "nil argument",
			arg:     nil,
			params:  false,
			isApier: false,
		},
		{
			name:    "check error missing param",
			arg:     "test",
			exp:     "jsonrpc: request body missing params",
			params:  true,
			isApier: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &bufferClose{
				Buffer: &bytes.Buffer{},
			}
			s := NewCustomJSONServerCodec(conn)
			sj := s.(*jsonServerCodec)

			if !tt.params {
				jm, _ := json.RawMessage.MarshalJSON([]byte(`{
					"method": "ApierV.Method",
					"params": [{
						"TPid": "test"
					}],
					"id": 0
				}`))
				js := json.RawMessage(jm)
				sj.req.Params = &js
			}
			if tt.isApier {
				sj.req.isApier = true
			}

			rcv := s.ReadRequestBody(tt.arg)

			if rcv != nil {
				if rcv.Error() != tt.exp {
					t.Errorf("recived %s, expected %s", rcv, tt.exp)
				}
			}
		})
	}
}

func TestJSONReadRequestBody2(t *testing.T) {
	conn := &bufferClose{
		Buffer: &bytes.Buffer{},
	}
	s := NewCustomJSONServerCodec(conn)
	sj := s.(*jsonServerCodec)

	b, _ := json.Marshal([1]any{&CGREvent{Tenant: "cgrates.org"}})
	js := json.RawMessage(b)
	sj.req.Params = &js

	sj.req.isApier = true

	args := &MethodParameters{
		Method:     "test",
		Parameters: [1]any{&CGREvent{Tenant: "cgrates.net"}},
	}
	rcv := s.ReadRequestBody(args)

	if rcv != nil {
		t.Errorf("recived %v, expected %v", rcv, nil)
	}

	a := args.Parameters
	if a.(map[string]any)[Tenant] != "cgrates.org" {
		t.Errorf("expected %s, recived %s", args.Parameters.(*CGREvent).Tenant, "cgrates.org")
	}
}

func TestJSONReadRequestBody3(t *testing.T) {
	conn := &bufferClose{
		Buffer: &bytes.Buffer{},
	}
	s := NewCustomJSONServerCodec(conn)
	sj := s.(*jsonServerCodec)

	b, _ := json.Marshal([1]any{&CGREvent{Tenant: "cgrates.org"}})
	js := json.RawMessage(b)
	sj.req.Params = &js

	sj.req.isApier = false

	args := &MethodParameters{
		Method:     "test",
		Parameters: [1]any{&CGREvent{Tenant: "cgrates.net"}},
	}
	rcv := s.ReadRequestBody(args)

	if rcv != nil {
		t.Errorf("recived %v, expected %v", rcv, nil)
	}

	a := args.Parameters.([1]any)
	if a[0].(*CGREvent).Tenant != "cgrates.net" {
		t.Errorf("expected %s, recived %s", a[0].(*CGREvent).Tenant, "cgrate.net")
	}
}

/*
- NewCustomJSONServerCodec
- defer Close
- ReadRequestHeader
- check whether c.req was populated correctly
- check whether c.seq was incremented
- check whether c.req.Id was added to c.pending array
- ReadRequestBody
- check if x.Parameters was properly modified
- WriteResponse
- check if the rpc.Response is populated correctly
*/
func TestJSONCodecOK(t *testing.T) {
	//Create server
	conn := &bufferClose{
		Buffer: &bytes.Buffer{},
	}
	s := NewCustomJSONServerCodec(conn)
	c := s.(*jsonServerCodec)

	defer s.Close()

	//ReadRequesHeader
	rr := rpc.Request{
		ServiceMethod: "Service.Method",
		Seq:           1,
	}

	conn.Write([]byte(`{
		"method": "ApierV.Method",
		"params": [{
			"TPid": "test"
		}],
		"id": 0
	}`))

	err := s.ReadRequestHeader(&rr)
	if err != nil {
		t.Error(err)
	}

	if c.req.Method != "ApierV.Method" {
		t.Errorf("received %v, expected %v", c.req.Method, "ApierV.Method")
	}

	exp := json.RawMessage([]byte{48})
	rcv := *c.pending[1]
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("received %v, expected %v", rcv, exp)
	}

	if c.seq != 1 {
		t.Errorf("received %v, expected %v", c.seq, 1)
	}

	//ReadRequestBody
	b, _ := json.Marshal([1]any{&CGREvent{Tenant: "cgrates.org"}})
	js := json.RawMessage(b)
	c.req.Params = &js

	c.req.isApier = true

	args := &MethodParameters{
		Method:     "test",
		Parameters: [1]any{&CGREvent{Tenant: "cgrates.net"}},
	}
	err = s.ReadRequestBody(args)

	if err != nil {
		t.Errorf("received %v, expected %v", err, nil)
	}

	a := args.Parameters
	if a.(map[string]any)[Tenant] != "cgrates.org" {
		t.Errorf("expected %s, received %s", args.Parameters.(*CGREvent).Tenant, "cgrates.org")
	}

	//WriteResponse
	res := rpc.Response{
		Seq: 1,
	}
	err = c.WriteResponse(&res, "test")
	if err != nil {
		t.Error(err)
	}

	mp := make(map[string]any)
	err = c.dec.Decode(&mp)
	if err != nil {
		t.Error(err)
	}

	if mp["error"] != nil ||
		mp["result"] != "test" ||
		mp["id"] != 0. {
		t.Error("unexpected reply", mp)
	}
}

func TestJSONCodecWriteResponse(t *testing.T) {
	conn := &bufferClose{
		Buffer: &bytes.Buffer{},
	}
	s := NewCustomJSONServerCodec(conn)
	c := s.(*jsonServerCodec)

	defer s.Close()

	res := rpc.Response{
		Seq: 1,
	}
	err := c.WriteResponse(&res, "test")
	if err.Error() != "invalid sequence number in response" {
		t.Error(err)
	}
}

func TestJSONCodecWriteResponse2(t *testing.T) {
	conn := &bufferClose{
		Buffer: &bytes.Buffer{},
	}
	s := NewCustomJSONServerCodec(conn)
	c := s.(*jsonServerCodec)

	conn.Write([]byte(`{
		"method": "ApierV.Method",
		"params": [{
			"TPid": "test"
		}]
	}`))

	defer s.Close()

	rr := rpc.Request{
		ServiceMethod: "Service.Method",
		Seq:           1,
	}
	err := s.ReadRequestHeader(&rr)
	if err != nil {
		t.Error(err)
	}

	res := rpc.Response{
		Seq:   1,
		Error: "error",
	}
	err = c.WriteResponse(&res, "test")
	if err != nil {
		t.Error(err)
	}

	mp := make(map[string]any)
	err = c.dec.Decode(&mp)
	if err != nil {
		t.Error(err)
	}

	if mp["error"] != "error" ||
		mp["result"] != nil ||
		mp["id"] != nil {
		t.Error("unexpected reply", mp)
	}
}

func TestJSONCodecClose(t *testing.T) {
	conn := &bufferClose{
		Buffer: &bytes.Buffer{},
	}
	s := NewCustomJSONServerCodec(conn)

	err := s.Close()

	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
