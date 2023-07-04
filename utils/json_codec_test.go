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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
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
	tests := []struct{
		name string
		arg any 
		exp error
		params bool
	}{
		{
			name: "nil argument",
			arg: nil,
			exp: nil,
			params: false,
		},
		{
			name: "check error missing param",
			arg: "test",
			exp: errMissingParams,
			params: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &bufferClose{
				Buffer: &bytes.Buffer{},
			}
			s := NewCustomJSONServerCodec(conn)
			sj := s.(*jsonServerCodec)

			if tt.params {
				sj.req.Params = nil
			}

			err := s.ReadRequestBody(tt.arg)

			if err != nil {
				if err.Error() != tt.exp.Error() {
					t.Errorf("recived %s, expected %s", err, tt.exp)
				}
			}
		})
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
