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
	"io"
)

func DecodeServerRequest(r io.Reader) (req *serverRequest, err error) {
	req = new(serverRequest)
	err = json.NewDecoder(r).Decode(req)
	return
}

// NewServerRequest used in dispatcherh tests
func NewServerRequest(method string, params, id json.RawMessage) *serverRequest {
	return &serverRequest{
		Method: method,
		Params: &params,
		Id:     &id,
	}
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
}

func WriteServerResponse(w io.Writer, id *json.RawMessage, result, err interface{}) error {
	return json.NewEncoder(w).Encode(
		serverResponse{
			Id:     id,
			Result: result,
			Error:  err,
		})
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result"`
	Error  interface{}      `json:"error"`
}
