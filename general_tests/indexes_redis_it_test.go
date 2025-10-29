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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestIndexesRedis(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := engine.NewRedisStorage(cfg.DataDbCfg().Host+":"+cfg.DataDbCfg().Port, 10, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding, 10, 20,
		utils.EmptyString, false, 0, 0, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
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
