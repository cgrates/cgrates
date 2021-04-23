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
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// updateInternalConns updates the connection list by specifying the subsystem for internal connections
func updateInternalConns(conns []string, subsystem string) (c []string) {
	subsystem = utils.MetaInternal + utils.ConcatenatedKeySep + subsystem
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		// if we have the connection internal we change the name so we can have internal rpc for each subsystem
		if conn == utils.MetaInternal {
			c[i] = subsystem
		}
	}
	return
}

// updateInternalConns updates the connection list by specifying the subsystem for internal connections
func updateBiRPCInternalConns(conns []string, subsystem string) (c []string) {
	subsystem = utils.ConcatenatedKeySep + subsystem
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		// if we have the connection internal we change the name so we can have internal rpc for each subsystem
		if conn == utils.MetaInternal ||
			conn == rpcclient.BiRPCInternal {
			c[i] += subsystem
		}
	}
	return
}

func getInternalJSONConns(conns []string) (c []string) {
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		if strings.HasPrefix(conn, utils.MetaInternal) {
			c[i] = utils.MetaInternal
		}
	}
	return
}

func getBiRPCInternalJSONConns(conns []string) (c []string) {
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		if strings.HasPrefix(conn, utils.MetaInternal) {
			c[i] = utils.MetaInternal
		} else if strings.HasPrefix(conn, rpcclient.BiRPCInternal) {
			c[i] = rpcclient.BiRPCInternal
		}
	}
	return
}
