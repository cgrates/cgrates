//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestIndexesRedis(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := engine.NewRedisStorage("127.0.0.1:6379", 10, utils.CGRateSLwr,
		cfg.DbCfg().DBConns[utils.MetaDefault].Password, cfg.GeneralCfg().DBDataEncoding, cfg.DbCfg().DBConns[utils.MetaDefault].Opts.RedisMaxConns,
		cfg.DbCfg().DBConns[utils.MetaDefault].Opts.RedisConnectAttempts, utils.EmptyString, false, 0, 0, 0, 0, 0,
		150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString, 1000, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	limit := engine.RedisLimit + 1
	indx := make(map[string]utils.StringSet)
	for i := range limit {
		indx["*string:*req.Destination:"+strconv.Itoa(i)] = utils.StringSet{"ATTR_New": {}}
	}
	if err = db.SetIndexesDrv(context.Background(), utils.CacheAttributeFilterIndexes, "cgrates.org:*any", indx,
		false, utils.NonTransactional); err != nil {
		t.Fatal(err)
	}
}
