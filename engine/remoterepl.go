/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

// SetReplicateHost will set the connID in cache
func SetReplicateHost(objType, objID, connID string) {
	if connID == utils.EmptyString {
		return
	}
	Cache.SetWithoutReplicate(utils.CacheReplicationHosts, objType+objID+utils.ConcatenatedKeySep+connID, connID, []string{objType + objID},
		true, utils.NonTransactional)
}

// replicate will call Set/Remove APIs on ReplicatorSv1
func replicate(connMgr *ConnManager, connIDs []string, filtered bool, objType, objID, method string, args interface{}) (err error) {
	// the reply is string for Set/Remove APIs
	// ignored in favor of the error
	var reply string
	if !filtered {
		// is not partial so send to all defined connections
		return utils.CastRPCErr(connMgr.Call(connIDs, nil, method, args, &reply))
	}
	// is partial so get all the replicationHosts from cache based on object Type and ID
	// alp_cgrates.org:ATTR1
	rplcHostIDsIfaces := Cache.tCache.GetGroupItems(utils.CacheReplicationHosts, objType+objID)
	rplcHostIDs := make(utils.StringSet)
	for _, hostID := range rplcHostIDsIfaces {
		rplcHostIDs.Add(hostID.(string))
	}
	// using the replication hosts call the method
	return utils.CastRPCErr(connMgr.CallWithConnIDs(connIDs, rplcHostIDs,
		method, args, &reply))
}

// replicateMultipleIDs will do the same thing as replicate but uses multiple objectIDs
// used when setting the LoadIDs
func replicateMultipleIDs(connMgr *ConnManager, connIDs []string, filtered bool, objType string, objIDs []string, method string, args interface{}) (err error) {
	// the reply is string for Set/Remove APIs
	// ignored in favor of the error
	var reply string
	if !filtered {
		// is not partial so send to all defined connections
		return utils.CastRPCErr(connMgr.Call(connIDs, nil, method, args, &reply))
	}
	// is partial so get all the replicationHosts from cache based on object Type and ID
	// combine all hosts in a single set so if we receive a get with one ID in list
	// send all list to that hos
	rplcHostIDs := make(utils.StringSet)
	for _, objID := range objIDs {
		rplcHostIDsIfaces := Cache.tCache.GetGroupItems(utils.CacheReplicationHosts, objType+objID)
		for _, hostID := range rplcHostIDsIfaces {
			rplcHostIDs.Add(hostID.(string))
		}
	}
	// using the replication hosts call the method
	return utils.CastRPCErr(connMgr.CallWithConnIDs(connIDs, rplcHostIDs,
		method, args, &reply))
}
