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

import (
	"testing"
)

// if the flag change this should fail
// do not use constants in this test
func TestFlags(t *testing.T) {
	if err := cgrMigratorFlags.Parse([]string{"-config_path", "true"}); err != nil {
		t.Fatal(err)
	} else if *cfgPath != "true" {
		t.Errorf("Expected true received:%v ", *cfgPath)
	}
	if err := cgrMigratorFlags.Parse([]string{"-exec", "true"}); err != nil {
		t.Fatal(err)
	} else if *exec != "true" {
		t.Errorf("Expected true received:%v ", *exec)
	}
	if err := cgrMigratorFlags.Parse([]string{"-version", "true"}); err != nil {
		t.Fatal(err)
	} else if !*version {
		t.Errorf("Expected true received:%v ", *version)
	}
	if err := cgrMigratorFlags.Parse([]string{"-datadb_type", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBType != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBType)
	}
	if err := cgrMigratorFlags.Parse([]string{"-datadb_host", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBHost != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBHost)
	}
	if err := cgrMigratorFlags.Parse([]string{"-datadb_port", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBPort != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBPort)
	}
	if err := cgrMigratorFlags.Parse([]string{"-datadb_name", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBName != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBName)
	}
	if err := cgrMigratorFlags.Parse([]string{"-datadb_user", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBUser != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBUser)
	}
	if err := cgrMigratorFlags.Parse([]string{"-datadb_passwd", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBPass != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBPass)
	}
	if err := cgrMigratorFlags.Parse([]string{"-dbdata_encoding", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDBDataEncoding != "true" {
		t.Errorf("Expected true received:%v ", *inDBDataEncoding)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisSentinel", "true"}); err != nil {
		t.Fatal(err)
	} else if *inDataDBRedisSentinel != "true" {
		t.Errorf("Expected true received:%v ", *inDataDBRedisSentinel)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisCluster", "true"}); err != nil {
		t.Fatal(err)
	} else if !*dbRedisCluster {
		t.Errorf("Expected true received:%v ", *dbRedisCluster)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisClusterSync", "true"}); err != nil {
		t.Fatal(err)
	} else if *dbRedisClusterSync != "true" {
		t.Errorf("Expected true received:%v ", *dbRedisClusterSync)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisClusterOndownDelay", "true"}); err != nil {
		t.Fatal(err)
	} else if *dbRedisClusterDownDelay != "true" {
		t.Errorf("Expected true received:%v ", *dbRedisClusterDownDelay)
	}
	if err := cgrMigratorFlags.Parse([]string{"-mongoQueryTimeout", "true"}); err != nil {
		t.Fatal(err)
	} else if *dbQueryTimeout != "true" {
		t.Errorf("Expected true received:%v ", *dbQueryTimeout)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisTLS", "true"}); err != nil {
		t.Fatal(err)
	} else if !*dbRedisTls {
		t.Errorf("Expected true received:%v ", *dbRedisTls)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisClientCertificate", "true"}); err != nil {
		t.Fatal(err)
	} else if *dbRedisClientCertificate != "true" {
		t.Errorf("Expected true received:%v ", *dbRedisClientCertificate)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisClientKey", "true"}); err != nil {
		t.Fatal(err)
	} else if *dbRedisClientKey != "true" {
		t.Errorf("Expected true received:%v ", *dbRedisClientKey)
	}
	if err := cgrMigratorFlags.Parse([]string{"-redisCACertificate", "true"}); err != nil {
		t.Fatal(err)
	} else if *dbRedisCACertificate != "true" {
		t.Errorf("Expected true received:%v ", *dbRedisCACertificate)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_type", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBType != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBType)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_host", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBHost != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBHost)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_port", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBPort != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBPort)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_name", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBName != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBName)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_user", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBUser != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBUser)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_password", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBPass != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBPass)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_datadb_encoding", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDBDataEncoding != "true" {
		t.Errorf("Expected true received:%v ", *outDBDataEncoding)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_redis_sentinel", "true"}); err != nil {
		t.Fatal(err)
	} else if *outDataDBRedisSentinel != "true" {
		t.Errorf("Expected true received:%v ", *outDataDBRedisSentinel)
	}
	if err := cgrMigratorFlags.Parse([]string{"-stordb_type", "true"}); err != nil {
		t.Fatal(err)
	} else if *inStorDBType != "true" {
		t.Errorf("Expected true received:%v ", *inStorDBType)
	}
	if err := cgrMigratorFlags.Parse([]string{"-stordb_host", "true"}); err != nil {
		t.Fatal(err)
	} else if *inStorDBHost != "true" {
		t.Errorf("Expected true received:%v ", *inStorDBHost)
	}
	if err := cgrMigratorFlags.Parse([]string{"-stordb_port", "true"}); err != nil {
		t.Fatal(err)
	} else if *inStorDBPort != "true" {
		t.Errorf("Expected true received:%v ", *inStorDBPort)
	}
	if err := cgrMigratorFlags.Parse([]string{"-stordb_name", "true"}); err != nil {
		t.Fatal(err)
	} else if *inStorDBName != "true" {
		t.Errorf("Expected true received:%v ", *inStorDBName)
	}
	if err := cgrMigratorFlags.Parse([]string{"-stordb_user", "true"}); err != nil {
		t.Fatal(err)
	} else if *inStorDBUser != "true" {
		t.Errorf("Expected true received:%v ", *inStorDBUser)
	}
	if err := cgrMigratorFlags.Parse([]string{"-stordb_passwd", "true"}); err != nil {
		t.Fatal(err)
	} else if *inStorDBPass != "true" {
		t.Errorf("Expected true received:%v ", *inStorDBPass)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_stordb_type", "true"}); err != nil {
		t.Fatal(err)
	} else if *outStorDBType != "true" {
		t.Errorf("Expected true received:%v ", *outStorDBType)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_stordb_host", "true"}); err != nil {
		t.Fatal(err)
	} else if *outStorDBHost != "true" {
		t.Errorf("Expected true received:%v ", *outStorDBHost)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_stordb_port", "true"}); err != nil {
		t.Fatal(err)
	} else if *outStorDBPort != "true" {
		t.Errorf("Expected true received:%v ", *outStorDBPort)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_stordb_name", "true"}); err != nil {
		t.Fatal(err)
	} else if *outStorDBName != "true" {
		t.Errorf("Expected true received:%v ", *outStorDBName)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_stordb_user", "true"}); err != nil {
		t.Fatal(err)
	} else if *outStorDBUser != "true" {
		t.Errorf("Expected true received:%v ", *outStorDBUser)
	}
	if err := cgrMigratorFlags.Parse([]string{"-out_stordb_password", "true"}); err != nil {
		t.Fatal(err)
	} else if *outStorDBPass != "true" {
		t.Errorf("Expected true received:%v ", *outStorDBPass)
	}
	if err := cgrMigratorFlags.Parse([]string{"-dry_run", "true"}); err != nil {
		t.Fatal(err)
	} else if !*dryRun {
		t.Errorf("Expected true received:%v ", *dryRun)
	}
	if err := cgrMigratorFlags.Parse([]string{"-verbose", "true"}); err != nil {
		t.Fatal(err)
	} else if !*verbose {
		t.Errorf("Expected true received:%v ", *verbose)
	}

}
