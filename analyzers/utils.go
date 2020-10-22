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

package analyzers

import (
	"log"

	"github.com/cgrates/cgrates/utils"
)

var h = new(Handler)

type Handler struct {
	enc  string
	from string
	to   string
}

func (h *Handler) handleTrafic(x interface{}) {
	log.Println(utils.ToJSON(x))
}

func (h *Handler) handleRequest(id uint64, method string, args interface{}) {
	h.handleTrafic(&RPCServerRequest{
		ID:     id,
		Method: method,
		Params: []interface{}{args},
	})
}
func (h *Handler) handleResponse(id uint64, x, err interface{}) {
	var e interface{}
	switch val := x.(type) {
	default:
	case nil:
	case string:
		e = val
	case error:
		e = val.Error()
	}
	h.handleTrafic(&RPCServerResponse{
		ID:     id,
		Result: x,
		Error:  e,
	})
}

type RPCServerRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     uint64        `json:"id"`
}

func (r *RPCServerRequest) reset() {
	r.Method = ""
	r.Params = nil
	r.ID = 0
}

type RPCServerResponse struct {
	ID     uint64      `json:"id"`
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}
