//go:build integration
// +build integration

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
		150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
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
