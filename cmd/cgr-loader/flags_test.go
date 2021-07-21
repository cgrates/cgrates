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

package main

import "testing"

func TestCGRLoaderFlags(t *testing.T) {
	if err := cgrLoaderFlags.Parse([]string{"-config_path", "/etc/cgrates"}); err != nil {
		t.Error(err)
	} else if *cfgPath != "/etc/cgrates" {
		t.Errorf("Expected /etc/cgrates, received %+v", *cfgPath)
	}

	if err := cgrLoaderFlags.Parse([]string{"-datadb_type", "*redis"}); err != nil {
		t.Error(err)
	} else if *dataDBType != "*redis" {
		t.Errorf("Expected *redis, received %+v", *dataDBType)
	}

	if err := cgrLoaderFlags.Parse([]string{"-datadb_host", "CGRATES"}); err != nil {
		t.Error(err)
	} else if *dataDBHost != "CGRATES" {
		t.Errorf("Expected CGRATES, received %+v", *dataDBHost)
	}

	if err := cgrLoaderFlags.Parse([]string{"-datadb_port", "6379"}); err != nil {
		t.Error(err)
	} else if *dataDBPort != "6379" {
		t.Errorf("Expected 6379, received %+v", *dataDBPort)
	}

	if err := cgrLoaderFlags.Parse([]string{"-datadb_name", "cgrates1"}); err != nil {
		t.Error(err)
	} else if *dataDBName != "cgrates1" {
		t.Errorf("Expected cgrates1, received %+v", *dataDBName)
	}

	if err := cgrLoaderFlags.Parse([]string{"-datadb_user", "USER1"}); err != nil {
		t.Error(err)
	} else if *dataDBUser != "USER1" {
		t.Errorf("Expected USER1, received %+v", *dataDBUser)
	}

	if err := cgrLoaderFlags.Parse([]string{"-datadb_passwd", "cgrates.org"}); err != nil {
		t.Error(err)
	} else if *dataDBPasswd != "cgrates.org" {
		t.Errorf("Expected cgrates.org, received %+v", *dataDBPasswd)
	}

	if err := cgrLoaderFlags.Parse([]string{"-dbdata_encoding", "json"}); err != nil {
		t.Error(err)
	} else if *dbDataEncoding != "json" {
		t.Errorf("Expected json, received %+v", *dbDataEncoding)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisSentinel", "sentinel_name"}); err != nil {
		t.Error(err)
	} else if *dbRedisSentinel != "sentinel_name" {
		t.Errorf("Expected jsentinel_name, received %+v", *dbRedisSentinel)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisCluster", "true"}); err != nil {
		t.Error(err)
	} else if *dbRedisCluster != true {
		t.Errorf("Expected true, received %+v", *dbRedisCluster)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisClusterSync", "3s"}); err != nil {
		t.Error(err)
	} else if *dbRedisClusterSync != "3s" {
		t.Errorf("Expected 3s, received %+v", *dbRedisClusterSync)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisClusterOndownDelay", "0"}); err != nil {
		t.Error(err)
	} else if *dbRedisClusterDownDelay != "0" {
		t.Errorf("Expected 0, received %+v", *dbRedisClusterDownDelay)
	}

	if err := cgrLoaderFlags.Parse([]string{"-mongoQueryTimeout", "5s"}); err != nil {
		t.Error(err)
	} else if *dbQueryTimeout != "5s" {
		t.Errorf("Expected 5s, received %+v", *dbQueryTimeout)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisTLS", "true"}); err != nil {
		t.Error(err)
	} else if *dbRedisTls != true {
		t.Errorf("Expected true, received %+v", *dbRedisTls)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisClientCertificate", "/path/to/certificate"}); err != nil {
		t.Error(err)
	} else if *dbRedisClientCertificate != "/path/to/certificate" {
		t.Errorf("Expected path/to/certificate, received %+v", *dbRedisClientCertificate)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisClientKey", "123"}); err != nil {
		t.Error(err)
	} else if *dbRedisClientKey != "123" {
		t.Errorf("Expected 123, received %+v", *dbRedisClientKey)
	}

	if err := cgrLoaderFlags.Parse([]string{"-redisCACertificate", "/path/to/CACertificate"}); err != nil {
		t.Error(err)
	} else if *dbRedisCACertificate != "/path/to/CACertificate" {
		t.Errorf("Expected /path/to/CACertificate, received %+v", *dbRedisCACertificate)
	}

	if err := cgrLoaderFlags.Parse([]string{"-stordb_type", "*mongo"}); err != nil {
		t.Error(err)
	} else if *storDBType != "*mongo" {
		t.Errorf("Expected *mongo, received %+v", *storDBType)
	}

	if err := cgrLoaderFlags.Parse([]string{"-stordb_host", "CGRATES"}); err != nil {
		t.Error(err)
	} else if *storDBHost != "CGRATES" {
		t.Errorf("Expected CGRATES, received %+v", *storDBHost)
	}

	if err := cgrLoaderFlags.Parse([]string{"-stordb_port", "6533"}); err != nil {
		t.Error(err)
	} else if *storDBPort != "6533" {
		t.Errorf("Expected 6533, received %+v", *storDBPort)
	}

	if err := cgrLoaderFlags.Parse([]string{"-stordb_name", "stordb"}); err != nil {
		t.Error(err)
	} else if *storDBName != "stordb" {
		t.Errorf("Expected stordb, received %+v", *storDBName)
	}

	if err := cgrLoaderFlags.Parse([]string{"-stordb_user", "cgrates_user"}); err != nil {
		t.Error(err)
	} else if *storDBUser != "cgrates_user" {
		t.Errorf("Expected cgrates_user, received %+v", *storDBUser)
	}

	if err := cgrLoaderFlags.Parse([]string{"-stordb_passwd", "cgrates.org"}); err != nil {
		t.Error(err)
	} else if *storDBPasswd != "cgrates.org" {
		t.Errorf("Expected cgrates.org, received %+v", *storDBPasswd)
	}

	if err := cgrLoaderFlags.Parse([]string{"-caching", "*none"}); err != nil {
		t.Error(err)
	} else if *cachingArg != "*none" {
		t.Errorf("Expected *none, received %+v", *cachingArg)
	}

	if err := cgrLoaderFlags.Parse([]string{"-tpid", "Default_tp"}); err != nil {
		t.Error(err)
	} else if *tpid != "Default_tp" {
		t.Errorf("Expected Default_tp, received %+v", *tpid)
	}

	if err := cgrLoaderFlags.Parse([]string{"-path", "/etc/tariffplans"}); err != nil {
		t.Error(err)
	} else if *dataPath != "/etc/tariffplans" {
		t.Errorf("Expected /etc/tariffplans, received %+v", *dataPath)
	}

	if err := cgrLoaderFlags.Parse([]string{"-version", "true"}); err != nil {
		t.Error(err)
	} else if *version != true {
		t.Errorf("Expected /etc/tariffplans, received %+v", *version)
	}

	if err := cgrLoaderFlags.Parse([]string{"-verbose", "true"}); err != nil {
		t.Error(err)
	} else if *verbose != true {
		t.Errorf("Expected /etc/tariffplans, received %+v", *verbose)
	}

	if err := cgrLoaderFlags.Parse([]string{"-dry_run", "true"}); err != nil {
		t.Error(err)
	} else if *dryRun != true {
		t.Errorf("Expected /etc/tariffplans, received %+v", *dryRun)
	}

	if err := cgrLoaderFlags.Parse([]string{"-field_sep", ","}); err != nil {
		t.Error(err)
	} else if *fieldSep != "," {
		t.Errorf("Expected , , received %+v", *fieldSep)
	}

	if err := cgrLoaderFlags.Parse([]string{"-import_id", "unique_id"}); err != nil {
		t.Error(err)
	} else if *importID != "unique_id" {
		t.Errorf("Expected unique_id, received %+v", *importID)
	}

	if err := cgrLoaderFlags.Parse([]string{"-timezone", "UTC"}); err != nil {
		t.Error(err)
	} else if *timezone != "UTC" {
		t.Errorf("Expected UTC, received %+v", *timezone)
	}

	if err := cgrLoaderFlags.Parse([]string{"-disable_reverse_mappings", "true"}); err != nil {
		t.Error(err)
	} else if *disableReverse != true {
		t.Errorf("Expected true, received %+v", *disableReverse)
	}

	if err := cgrLoaderFlags.Parse([]string{"-flush_stordb", "true"}); err != nil {
		t.Error(err)
	} else if *flushStorDB != true {
		t.Errorf("Expected true, received %+v", *flushStorDB)
	}

	if err := cgrLoaderFlags.Parse([]string{"-remove", "true"}); err != nil {
		t.Error(err)
	} else if *remove != true {
		t.Errorf("Expected true, received %+v", *remove)
	}

	if err := cgrLoaderFlags.Parse([]string{"-api_key", "14422"}); err != nil {
		t.Error(err)
	} else if *apiKey != "14422" {
		t.Errorf("Expected 14422, received %+v", *apiKey)
	}

	if err := cgrLoaderFlags.Parse([]string{"-route_id", "route_idss"}); err != nil {
		t.Error(err)
	} else if *routeID != "route_idss" {
		t.Errorf("Expected route_idss, received %+v", *routeID)
	}

	if err := cgrLoaderFlags.Parse([]string{"-tenant", "tenant.com"}); err != nil {
		t.Error(err)
	} else if *tenant != "tenant.com" {
		t.Errorf("Expected tenant.com, received %+v", *tenant)
	}

	if err := cgrLoaderFlags.Parse([]string{"-from_stordb", "true"}); err != nil {
		t.Error(err)
	} else if *fromStorDB != true {
		t.Errorf("Expected true, received %+v", *fromStorDB)
	}

	if err := cgrLoaderFlags.Parse([]string{"-to_stordb", "true"}); err != nil {
		t.Error(err)
	} else if *toStorDB != true {
		t.Errorf("Expected true, received %+v", *toStorDB)
	}

	if err := cgrLoaderFlags.Parse([]string{"-caches_address", "*internal"}); err != nil {
		t.Error(err)
	} else if *cacheSAddress != "*internal" {
		t.Errorf("Expected *internal, received %+v", *cacheSAddress)
	}

	if err := cgrLoaderFlags.Parse([]string{"-scheduler_address", "*internal"}); err != nil {
		t.Error(err)
	} else if *schedulerAddress != "*internal" {
		t.Errorf("Expected *internal, received %+v", *schedulerAddress)
	}

	if err := cgrLoaderFlags.Parse([]string{"-rpc_encoding", "*gob"}); err != nil {
		t.Error(err)
	} else if *rpcEncoding != "*gob" {
		t.Errorf("Expected **gob, received %+v", *rpcEncoding)
	}
}
