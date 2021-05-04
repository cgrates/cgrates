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

package engine

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestCallCacheNoCaching(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cM := NewConnManager(defaultCfg, nil)
	cacheConns := []string{}
	caching := utils.MetaNone
	args := map[string][]string{
		utils.FilterIDs:   {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
		utils.ResourceIDs: {},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	expArgs := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(args, expArgs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expArgs, args)
	}
}

func TestCallCacheReloadCacheFirstCallFail(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				expArgs := utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					ArgsCache: map[string][]string{
						utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return utils.ErrUnsupporteServiceMethod
			},
		},
	}
	client <- ccM

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog := "Reloading cache\n"
	experr := utils.ErrUnsupporteServiceMethod
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog := buf.String()[20:]
	if rcvlog != explog {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog, rcvlog)
	}
}

func TestCallCacheReloadCacheSecondCallFailed(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(args, reply interface{}) error {
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return utils.ErrUnsupporteServiceMethod
			},
		},
	}
	client <- ccM

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaReload
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	explog1 := "Reloading cache"
	explog2 := "Clearing indexes"
	experr := utils.ErrUnsupporteServiceMethod
	explog3 := fmt.Sprintf("WARNING: Got error on cache clear: %s\n", experr)
	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, true)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	rcvlog1 := buf.String()[20 : 20+len(explog1)]
	if rcvlog1 != explog1 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog1, rcvlog1)
	}

	rcvlog2 := buf.String()[41+len(rcvlog1) : 41+len(rcvlog1)+len(explog2)]
	if rcvlog2 != explog2 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog2, rcvlog2)
	}

	rcvlog3 := buf.String()[62+len(rcvlog1)+len(explog2):]
	if rcvlog3 != explog3 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", explog3, rcvlog3)
	}
}

func TestCallCacheLoadCache(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1LoadCache: func(args, reply interface{}) error {
				expArgs := utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					ArgsCache: map[string][]string{
						utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
		},
	}
	client <- ccM

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaLoad
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCallCacheRemoveItems(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1RemoveItems: func(args, reply interface{}) error {
				expArgs := utils.AttrReloadCacheWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					ArgsCache: map[string][]string{
						utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
					},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
					CacheIDs: []string{"cacheID"},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
		},
	}
	client <- ccM

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaRemove
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{"cacheID"}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}

func TestCallCacheClear(t *testing.T) {
	tmp1, tmp2 := connMgr, Cache
	defer func() {
		connMgr = tmp1
		Cache = tmp2
	}()

	defaultCfg := config.NewDefaultCGRConfig()
	Cache = NewCacheS(defaultCfg, nil, nil)
	cacheConns := []string{"cacheConn1"}
	client := make(chan rpcclient.ClientConnector, 1)
	ccM := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.CacheSv1Clear: func(args, reply interface{}) error {
				expArgs := &utils.AttrCacheIDsWithAPIOpts{
					APIOpts: map[string]interface{}{
						utils.Subsys: utils.MetaChargers,
					},
				}

				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf(
						"\nWrong value of args: \nexpected: <%+v>, \nreceived: <%+v>",
						expArgs, args,
					)
				}
				return nil
			},
		},
	}
	client <- ccM

	cM := NewConnManager(defaultCfg, map[string]chan rpcclient.ClientConnector{
		"cacheConn1": client,
	})
	caching := utils.MetaClear
	args := map[string][]string{
		utils.FilterIDs: {"cgrates.org:FLTR_ID1", "cgrates.org:FLTR_ID2"},
	}
	cacheIDs := []string{}
	opts := map[string]interface{}{
		utils.Subsys: utils.MetaChargers,
	}

	err := CallCache(cM, cacheConns, caching, args, cacheIDs, opts, false)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}
}
