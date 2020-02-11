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

package config

import (
	"net"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// DataProvider is a data source from multiple formats
type DataProvider interface {
	String() string // printable version of data
	FieldAsInterface(fldPath []string) (interface{}, error)
	FieldAsString(fldPath []string) (string, error)
	RemoteHost() net.Addr
}

func DPDynamicInterface(dnVal string, dP DataProvider) (interface{}, error) {
	if strings.HasPrefix(dnVal, utils.DynamicDataPrefix) {
		dnVal = strings.TrimPrefix(dnVal, utils.DynamicDataPrefix)
		return dP.FieldAsInterface(strings.Split(dnVal, utils.NestingSep))
	}
	return utils.StringToInterface(dnVal), nil
}

func DPDynamicString(dnVal string, dP DataProvider) (string, error) {
	if strings.HasPrefix(dnVal, utils.DynamicDataPrefix) {
		dnVal = strings.TrimPrefix(dnVal, utils.DynamicDataPrefix)
		return dP.FieldAsString(strings.Split(dnVal, utils.NestingSep))
	}
	return dnVal, nil
}
