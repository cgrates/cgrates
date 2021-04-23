/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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

package apis

import (
	"testing"

	"github.com/cgrates/birpc"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCacheHasItem(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	client := make(chan birpc.ClientConnector, 1)
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	connMngr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): client,
	})
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMngr)

	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: utils.CGRateSorg,
			ID:     "TestGetAttributeProfile",
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.RequestType",
					Type:  utils.MetaConstant,
					Value: utils.MetaPrepaid,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaLoad,
		},
	}
	var reply string

	adms := NewAdminSv1(cfg, dm, connMngr)
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	expected := "SERVER_ERROR: context deadline exceeded"
	if err := adms.SetAttributeProfile(ctx, attrPrf, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	cancel()

	ch := engine.NewCacheS(cfg, dm, nil)
	cache := NewCacheSv1(ch)
	var replyBool bool
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TestGetAttributeProfile",
		},
	}
	if err := cache.HasItem(nil, args, &replyBool); err != nil {
		t.Error(err)
	} else if replyBool {
		t.Errorf("Unexpected replyBool returned")
	}
}
