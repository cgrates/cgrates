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

package cores

import (
	"bufio"
	"encoding/gob"
	"io/ioutil"
	"log"
	"net/rpc"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

type mockRWC struct{}

func (*mockRWC) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (mk *mockRWC) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (mk *mockRWC) Close() error {
	return nil
}

//Mocking For getting a nil error when the interface argument is nil in encoding
type mockReadWriteCloserErrorNilInterface struct {
	mockRWC
}

func (mk *mockReadWriteCloserErrorNilInterface) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestWriteResponseInterface(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	resp := &rpc.Response{
		ServiceMethod: utils.APIerSv1Ping,
		Seq:           123,
	}
	conn := new(mockReadWriteCloserErrorNilInterface)
	exp := newGobServerCodec(conn)
	expected := "gob: cannot encode nil value"
	if err := exp.WriteResponse(resp, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

type mockReadWriteCloserErrorNilResponse struct {
	mockRWC
}

func (mk *mockReadWriteCloserErrorNilResponse) Write(p []byte) (n int, err error) {
	return 4, utils.ErrNotImplemented
}

func TestWriteResponseResponse(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	resp := &rpc.Response{
		ServiceMethod: utils.APIerSv1Ping,
		Seq:           123,
		Error:         "err",
	}
	conn := new(mockReadWriteCloserErrorNilResponse)
	buf := bufio.NewWriter(conn)
	gsrv := gobServerCodec{
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
	}
	if err := gsrv.WriteResponse(resp, "string"); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}

	buf = bufio.NewWriter(conn)
	gsrv.encBuf = buf

	if err := gsrv.WriteResponse(resp, "string"); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestReadRequestHeader(t *testing.T) {
	conn := new(mockReadWriteCloserErrorNilResponse)
	buf := bufio.NewWriter(conn)
	gsrv := gobServerCodec{
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
	}
	expected := "gob: DecodeValue of unassignable value"
	if err := gsrv.ReadRequestHeader(nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestReadRequestBody(t *testing.T) {
	conn := new(mockReadWriteCloserErrorNilResponse)
	buf := bufio.NewWriter(conn)
	gsrv := gobServerCodec{
		enc:    gob.NewEncoder(buf),
		encBuf: buf,
		rwc:    conn,
		dec:    gob.NewDecoder(conn),
	}
	expected := "gob: attempt to decode into a non-pointer"
	if err := gsrv.ReadRequestBody(2); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestClose(t *testing.T) {
	conn := new(mockRWC)
	exp := newGobServerCodec(conn)
	//now after calling, it will be closed
	if err := exp.Close(); err != nil {
		t.Error(err)
	}

	//calling again the function won t close
	if err := exp.Close(); err != nil {
		t.Error(err)
	}
}
