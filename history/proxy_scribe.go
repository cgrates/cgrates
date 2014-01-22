/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package history

import "net/rpc"

const (
	JSON = "json"
	GOB  = "gob"
)

type ProxyScribe struct {
	Client *rpc.Client
}

func NewProxyScribe(addr string) (*ProxyScribe, error) {
	client, err := rpc.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}
	return &ProxyScribe{Client: client}, nil
}

func (ps *ProxyScribe) Record(rec *Record, out *int) error {
	return ps.Client.Call("Scribe.Record", rec, out)
}
