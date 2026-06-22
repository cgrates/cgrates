//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package migrator

import (
	"path"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	stsPathIn   string
	stsPathOut  string
	stsCfgIn    *config.CGRConfig
	stsCfgOut   *config.CGRConfig
	stsMigrator *Migrator
)

var sTestsStsIT = []func(t *testing.T){
	testStsITConnect,
	testStsITFlush,
}

func TestStatsQueueITRedis(t *testing.T) {
	var err error
	stsPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	stsCfgIn, err = config.NewCGRConfigFromPath(context.Background(), stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	stsCfgOut, err = config.NewCGRConfigFromPath(context.Background(), stsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsStsIT {
		t.Run("TestStatsQueueITMigrateRedis", stest)
	}
	stsMigrator.Close()
}

func TestStatsQueueITMongo(t *testing.T) {
	var err error
	stsPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	stsCfgIn, err = config.NewCGRConfigFromPath(context.Background(), stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	stsCfgOut, err = config.NewCGRConfigFromPath(context.Background(), stsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsStsIT {
		t.Run("TestStatsQueueITMigrateMongo", stest)
	}
	stsMigrator.Close()
}

func TestStatsQueueITMove(t *testing.T) {
	var err error
	stsPathIn = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	stsCfgIn, err = config.NewCGRConfigFromPath(context.Background(), stsPathIn)
	if err != nil {
		t.Fatal(err)
	}
	stsPathOut = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	stsCfgOut, err = config.NewCGRConfigFromPath(context.Background(), stsPathOut)
	if err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsStsIT {
		t.Run("TestStatsQueueITMove", stest)
	}
	stsMigrator.Close()
}

func testStsITConnect(t *testing.T) {
	cacheIn := engine.NewCacheS(stsCfgIn, nil, nil, nil)
	dataDBIn, err := NewMigratorDataDBs([]string{utils.MetaDefault}, stsCfgIn.GeneralCfg().DBDataEncoding, stsCfgIn, cacheIn)
	if err != nil {
		t.Fatal(err)
	}
	cacheOut := engine.NewCacheS(stsCfgOut, nil, nil, nil)
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, stsCfgOut.GeneralCfg().DBDataEncoding, stsCfgOut, cacheOut)
	if err != nil {
		t.Fatal(err)
	}
	stsMigrator = NewMigrator(dataDBIn, dataDBOut, false, stsPathIn == stsPathOut)
}

func testStsITFlush(t *testing.T) {
	stsMigrator.dmTo.DB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(stsMigrator.dmTo.DB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
}
