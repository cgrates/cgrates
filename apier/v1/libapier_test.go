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

package v1

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestComposeArgsReload(t *testing.T) {
	apv1 := &APIerSv1{}
	expArgs := utils.AttrReloadCacheWithAPIOpts{
		APIOpts: make(map[string]interface{}),
		Tenant:  "cgrates.org",
		ArgsCache: map[string][]string{
			utils.AttributeProfileIDs: {"cgrates.org:ATTR1"},
		},
	}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheAttributeProfiles,
		"cgrates.org:ATTR1", nil, nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	expArgs.ArgsCache[utils.AttributeFilterIndexIDs] = []string{"cgrates.org:*cdrs:*none:*any:*any"}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheAttributeProfiles,
		"cgrates.org:ATTR1", &[]string{}, []string{utils.MetaCDRs}, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	expArgs.ArgsCache[utils.AttributeFilterIndexIDs] = []string{
		"cgrates.org:*cdrs:*string:*req.Account:1001",
		"cgrates.org:*cdrs:*prefix:*req.Destination:1001",
	}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheAttributeProfiles,
		"cgrates.org:ATTR1", &[]string{"*string:~*req.Account:1001|~req.Subject", "*prefix:1001:~*req.Destination", "*gt:~req.Usage:0"}, []string{utils.MetaCDRs}, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	expArgs = utils.AttrReloadCacheWithAPIOpts{
		APIOpts: make(map[string]interface{}),
		Tenant:  "cgrates.org",
		ArgsCache: map[string][]string{
			utils.StatsQueueProfileIDs: {"cgrates.org:Stat2"},
			utils.StatFilterIndexIDs: {
				"cgrates.org:*string:*req.Account:1001",
				"cgrates.org:*prefix:*req.Destination:1001",
			},
		},
	}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"*string:~*req.Account:1001|~req.Subject", "*prefix:1001:~*req.Destination"}, nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	expArgs.ArgsCache[utils.StatFilterIndexIDs] = []string{"cgrates.org:*none:*any:*any"}

	if rply, err := apv1.composeArgsReload("cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{}, []string{utils.MetaCDRs}, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expArgs, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(expArgs), utils.ToJSON(rply))
	}

	if _, err := apv1.composeArgsReload("cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"FLTR1"}, []string{utils.MetaCDRs}, make(map[string]interface{})); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
}

type rpcRequest struct {
	Method string
	Params interface{}
}
type rpcMock chan *rpcRequest

func (r rpcMock) Call(_ *context.Context, method string, args, _ interface{}) error {
	r <- &rpcRequest{
		Method: method,
		Params: args,
	}
	return nil
}

func TestCallCache(t *testing.T) {
	cache := make(rpcMock, 1)
	ch := make(chan birpc.ClientConnector, 1)
	ch <- cache
	cn := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): ch,
	})
	apv1 := &APIerSv1{
		ConnMgr: cn,
		Config:  config.CgrConfig(),
	}
	if err := apv1.CallCache(utils.MetaNone, "", "", "", nil, nil, nil); err != nil {
		t.Fatal(err)
	} else if len(cache) != 0 {
		t.Fatal("Expected call cache to not be called")
	}
	exp := &rpcRequest{
		Method: utils.CacheSv1Clear,
		Params: &utils.AttrCacheIDsWithAPIOpts{
			Tenant:   "cgrates.org",
			CacheIDs: []string{utils.CacheStatQueueProfiles, utils.CacheStatFilterIndexes},
			APIOpts:  make(map[string]interface{}),
		},
	}
	if err := apv1.CallCache(utils.MetaClear, "cgrates.org", utils.CacheStatQueueProfiles, "", nil, nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if len(cache) != 1 {
		t.Fatal("Expected call cache to be called")
	} else if rply := <-cache; !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	exp = &rpcRequest{
		Method: utils.CacheSv1ReloadCache,
		Params: utils.AttrReloadCacheWithAPIOpts{
			APIOpts: make(map[string]interface{}),
			Tenant:  "cgrates.org",
			ArgsCache: map[string][]string{
				utils.StatsQueueProfileIDs: {"cgrates.org:Stat2"},
				utils.StatFilterIndexIDs: {
					"cgrates.org:*string:*req.Account:1001",
					"cgrates.org:*prefix:*req.Destination:1001",
				},
			},
		},
	}

	if err := apv1.CallCache(utils.MetaReload, "cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"*string:~*req.Account:1001|~req.Subject", "*prefix:1001:~*req.Destination"},
		nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if len(cache) != 1 {
		t.Fatal("Expected call cache to be called")
	} else if rply := <-cache; !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}
	exp.Method = utils.CacheSv1LoadCache
	if err := apv1.CallCache(utils.MetaLoad, "cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"*string:~*req.Account:1001|~req.Subject", "*prefix:1001:~*req.Destination"},
		nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if len(cache) != 1 {
		t.Fatal("Expected call cache to be called")
	} else if rply := <-cache; !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}
	exp.Method = utils.CacheSv1RemoveItems
	if err := apv1.CallCache(utils.MetaRemove, "cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"*string:~*req.Account:1001|~req.Subject", "*prefix:1001:~*req.Destination"},
		nil, make(map[string]interface{})); err != nil {
		t.Fatal(err)
	} else if len(cache) != 1 {
		t.Fatal("Expected call cache to be called")
	} else if rply := <-cache; !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	if err := apv1.CallCache(utils.MetaLoad, "cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"FLTR1", "*prefix:1001:~*req.Destination"},
		nil, make(map[string]interface{})); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	} else if len(cache) != 0 {
		t.Fatal("Expected call cache to not be called")
	}
	if err := apv1.CallCache(utils.MetaRemove, "cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"FLTR1", "*prefix:1001:~*req.Destination"},
		nil, make(map[string]interface{})); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	} else if len(cache) != 0 {
		t.Fatal("Expected call cache to not be called")
	}
	if err := apv1.CallCache(utils.MetaReload, "cgrates.org", utils.CacheStatQueueProfiles,
		"cgrates.org:Stat2", &[]string{"FLTR1", "*prefix:1001:~*req.Destination"},
		nil, make(map[string]interface{})); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	} else if len(cache) != 0 {
		t.Fatal("Expected call cache to not be called")
	}
}
