/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package actions

import (
	"fmt"
	"net"

	"github.com/cgrates/cgrates/utils"
)

// ActData is the data source for the actioners
type ActData struct {
	Req  map[string]interface{}
	Opts map[string]interface{}
}

// String implements utils.DataProvider
func (aD *ActData) String() string {
	return utils.ToIJSON(aD)
}

// RemoteHost implements utils.DataProvider
func (aD *ActData) RemoteHost() net.Addr {
	return nil
}

// FieldAsString implements utils.DataProvider
func (aD *ActData) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	if iface, err = aD.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(iface), nil
}

// FieldAsInterface implements utils.DataProvider
func (aD *ActData) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if len(fldPath) < 1 {
		return nil, fmt.Errorf("invalid fieldPath: <%+v>", fldPath)
	}
	switch fldPath[0] {
	case utils.MetaReq:
	case utils.MetaOpts:
	default:
		return nil, fmt.Errorf("invalid prefix for fieldPath: <%+v>", fldPath)
	}

	return
}

// Set implements utils.RWDataProvider
func (aD *ActData) Set(fldPath []string, val interface{}) (err error) {
	return

}

// Set implements utils.RWDataProvider
func (aD *ActData) Remove(fldPath []string) (err error) {
	return
}
