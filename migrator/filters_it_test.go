//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrator

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fltrCfgIn       *config.CGRConfig
	fltrCfgOut      *config.CGRConfig
	fltrMigrator    *Migrator
	inPath, outPath string
)

var sTestsFltrIT = []func(t *testing.T){
	testFltrITConnect,
	testFltrITFlush,
	testFltrITMove,
}

func TestFiltersITMove(t *testing.T) {
	inPath = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	outPath = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	testFltrStart("TestFiltersITMove", t)
}

func TestFiltersITMoveEncoding(t *testing.T) {
	inPath = path.Join(*utils.DataDir, "conf", "samples", "tutmongo")
	outPath = path.Join(*utils.DataDir, "conf", "samples", "tutmongojson")
	testFltrStart("TestFiltersITMoveEncoding", t)
}

func TestFiltersITMoveEncoding2(t *testing.T) {
	inPath = path.Join(*utils.DataDir, "conf", "samples", "tutredis")
	outPath = path.Join(*utils.DataDir, "conf", "samples", "tutmysqljson")
	testFltrStart("TestFiltersITMoveEncoding2", t)
}

func testFltrStart(testName string, t *testing.T) {
	var err error
	if fltrCfgIn, err = config.NewCGRConfigFromPath(context.Background(), inPath); err != nil {
		t.Fatal(err)
	}
	if fltrCfgOut, err = config.NewCGRConfigFromPath(context.Background(), outPath); err != nil {
		t.Fatal(err)
	}
	for _, stest := range sTestsFltrIT {
		t.Run(testName, stest)
	}
	fltrMigrator.Close()
}

func testFltrITConnect(t *testing.T) {
	locker := engine.NewGuardianLocker(config.NewDefaultCGRConfig())
	cacheIn := engine.NewCacheS(fltrCfgIn, nil, nil, nil, locker)
	dataDBIn, err := NewMigratorDataDBs([]string{utils.MetaDefault}, fltrCfgIn.GeneralCfg().DBDataEncoding, fltrCfgIn, cacheIn, locker)
	if err != nil {
		t.Fatal(err)
	}
	cacheOut := engine.NewCacheS(fltrCfgOut, nil, nil, nil, locker)
	dataDBOut, err := NewMigratorDataDBs([]string{utils.MetaDefault}, fltrCfgOut.GeneralCfg().DBDataEncoding, fltrCfgOut, cacheOut, locker)
	if err != nil {
		t.Fatal(err)
	}
	fltrMigrator = NewMigrator(dataDBIn, dataDBOut, false, inPath == outPath)
}

func testFltrITFlush(t *testing.T) {
	fltrMigrator.dmTo.DB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(fltrMigrator.dmTo.DB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
	fltrMigrator.dmFrom.DB()[utils.MetaDefault].Flush("")
	if err := engine.SetDBVersions(fltrMigrator.dmFrom.DB()[utils.MetaDefault]); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testFltrITMove(t *testing.T) {
	expFilters := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*engine.FilterRule{{
			Type:    utils.MetaPrefix,
			Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
			Values:  []string{"1001"},
		}},
	}
	expFilters.Compile()
	if err := fltrMigrator.dmFrom.SetFilter(context.TODO(), expFilters, true); err != nil {
		t.Error(err)
	}
	currentVersion := engine.CurrentDataDBVersions()
	if err := fltrMigrator.dmFrom.DB()[utils.MetaDefault].SetVersions(currentVersion, false); err != nil {
		t.Error("Error when setting version for Filters ", err.Error())
	}
	if err, _ := fltrMigrator.Migrate([]string{utils.MetaFilters}); err != nil {
		t.Error("Error when migrating Filters ", err.Error())
	}
	result, err := fltrMigrator.dmTo.GetFilter(context.TODO(), expFilters.Tenant, expFilters.ID, false, false, utils.NonTransactional)
	if err != nil {
		t.Error(err)
	}
	result.Compile()
	if !reflect.DeepEqual(expFilters, result) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expFilters), utils.ToJSON(result))
	}
	// the migrated filter must be removed from the source db
	if _, err = fltrMigrator.dmFrom.GetFilter(context.TODO(), expFilters.Tenant, expFilters.ID, false, false, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}
}
