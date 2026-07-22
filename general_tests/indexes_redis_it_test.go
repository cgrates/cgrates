//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestIndexesRedis(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := engine.NewRedisStorage(cfg.DataDbCfg().Host+":"+cfg.DataDbCfg().Port, 10, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding, 10, 20,
		utils.EmptyString, false, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	limit := engine.RedisLimit + 1
	indx := make(map[string]utils.StringSet)
	for i := 0; i < limit; i++ {
		indx["*string:*req.Destination:"+strconv.Itoa(i)] = utils.StringSet{"ATTR_New": {}}
	}
	if err = db.SetIndexesDrv(utils.CacheAttributeFilterIndexes, "cgrates.org:*any", indx,
		false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
}
