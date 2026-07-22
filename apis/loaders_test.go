// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func TestLoadersNewLoaderSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewGuardianLocker(cfg)
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	ldrS := loaders.NewLoaderS(cfg, dm, fltrs, nil)

	exp := &LoaderSv1{
		ldrS: ldrS,
	}
	rcv := NewLoaderSv1(ldrS)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}
