//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	locker := engine.NewGuardianLocker(config.NewDefaultCGRConfig())
	cacheIn := engine.NewCacheS(stsCfgIn, nil, nil, nil, locker)
	dataDBIn, err := NewMigratorDataDBs([]string{utils.MetaDefault}, stsCfgIn.GeneralCfg().DBDataEncoding, stsCfgIn, cacheIn, locker)
	if err != nil {
		t.Fatal(err)
	}
	cacheOut := engine.NewCacheS(stsCfgOut, nil, nil, nil, locker)
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, stsCfgOut.GeneralCfg().DBDataEncoding, stsCfgOut, cacheOut, locker)
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
