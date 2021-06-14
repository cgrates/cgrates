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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func TestChargersmatchingChargerProfilesForEventErrPass(t *testing.T) {
	Cache.Clear(nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false

	dbm := &DataDBMock{
		GetChargerProfileDrvF: func(ctx *context.Context, s1, s2 string) (*ChargerProfile, error) {
			return &ChargerProfile{
				Tenant:    s1,
				ID:        s2,
				RunID:     utils.MetaDefault,
				FilterIDs: []string{"fltr1"},
			}, nil
		},
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) {
			return []string{s + "cgrates.org:chr1"}, nil
		},
		GetFilterDrvF: func(ctx *context.Context, s1, s2 string) (*Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dmFilter := NewDataManager(dbm, defaultCfg.CacheCfg(), nil)
	cS := &ChargerService{
		dm: dmFilter,
		filterS: &FilterS{
			dm:  dmFilter,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
	}
	cgrEv := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			"Charger":        "ChargerProfile1",
			utils.AnswerTime: time.Date(2021, 4, 19, 12, 0, 0, 0, time.UTC),
			"UsageInterval":  "10s",
			utils.Weight:     "10.0",
		},
		APIOpts: map[string]interface{}{
			utils.Subsys: utils.MetaChargers,
		},
	}

	experr := utils.ErrNotImplemented
	rcv, err := cS.matchingChargerProfilesForEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestChargersprocessEventCallNilErr(t *testing.T) {
	Cache.Clear(nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false
	defaultCfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, defaultCfg.CacheCfg(), nil)
	cP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				rply := AttrSProcessEventReply{
					AlteredFields: []string{utils.AccountField},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]interface{}{
							utils.AccountField: "1001",
						},
					},
				}
				*reply.(*AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
		connMgr: NewConnManager(defaultCfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): rpcInternal,
		}),
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}

	exp := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "1001",
			AlteredFields:   []string{utils.MetaReqRunID, utils.AccountField},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cgrEvID",
				Event: map[string]interface{}{
					utils.AccountField: "1001",
				},
			},
		},
	}
	rcv, err := cS.processEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}

}

func TestChargersprocessEventCallErr(t *testing.T) {
	Cache.Clear(nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false
	defaultCfg.ChargerSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	data := NewInternalDB(nil, nil, true)
	dm := NewDataManager(data, defaultCfg.CacheCfg(), nil)
	cP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				return utils.ErrNotFound
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
		connMgr: NewConnManager(defaultCfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): rpcInternal,
		}),
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}

	exp := []*ChrgSProcessEventReply{
		{
			ChargerSProfile: "1001",
			AlteredFields:   []string{utils.MetaReqRunID},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cgrEvID",
				Event: map[string]interface{}{
					utils.AccountField: "1001",
					"RunID":            utils.MetaDefault,
				},
				APIOpts: map[string]interface{}{
					utils.Subsys:      utils.MetaChargers,
					utils.OptsContext: utils.MetaChargers,
				},
			},
		},
	}
	rcv, err := cS.processEvent(context.Background(), cgrEv.Tenant, cgrEv)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1ProcessEventErrNotFound(t *testing.T) {
	Cache.Clear(nil)
	dataDB := NewInternalDB(nil, nil, true)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false
	defaultCfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := NewDataManager(dataDB, defaultCfg.CacheCfg(), nil)

	cP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				rply := AttrSProcessEventReply{
					AlteredFields: []string{utils.AccountField},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]interface{}{
							utils.AccountField: "1001",
						},
					},
				}
				*reply.(*AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
		connMgr: NewConnManager(defaultCfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): rpcInternal,
		}),
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1002",
		},
	}
	reply := &[]*ChrgSProcessEventReply{}

	experr := utils.ErrNotFound
	err := cS.V1ProcessEvent(context.Background(), args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1ProcessEventErrOther(t *testing.T) {
	Cache.Clear(nil)
	dataDB := NewInternalDB(nil, nil, true)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false
	defaultCfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := NewDataManager(dataDB, defaultCfg.CacheCfg(), nil)

	cP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			"invalidMethod": func(ctx *context.Context, args, reply interface{}) error {
				rply := AttrSProcessEventReply{
					AlteredFields: []string{utils.AccountField},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]interface{}{
							utils.AccountField: "1001",
						},
					},
				}
				*reply.(*AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
		connMgr: NewConnManager(defaultCfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): rpcInternal,
		}),
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}
	reply := &[]*ChrgSProcessEventReply{}

	exp := &[]*ChrgSProcessEventReply{}
	experr := fmt.Sprintf("SERVER_ERROR: %s", rpcclient.ErrUnsupporteServiceMethod)
	err := cS.V1ProcessEvent(context.Background(), args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1ProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	dataDB := NewInternalDB(nil, nil, true)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false
	defaultCfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := NewDataManager(dataDB, defaultCfg.CacheCfg(), nil)

	cP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	ccM := &ccMock{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				rply := AttrSProcessEventReply{
					AlteredFields: []string{utils.AccountField},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "cgrEvID",
						Event: map[string]interface{}{
							utils.AccountField: "1001",
						},
					},
				}
				*reply.(*AttrSProcessEventReply) = rply
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
		connMgr: NewConnManager(defaultCfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): rpcInternal,
		}),
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}
	reply := &[]*ChrgSProcessEventReply{}

	exp := &[]*ChrgSProcessEventReply{
		{
			ChargerSProfile: "1001",
			AlteredFields:   []string{utils.MetaReqRunID, utils.AccountField},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "cgrEvID",
				Event: map[string]interface{}{
					utils.AccountField: "1001",
				},
			},
		},
	}
	err := cS.V1ProcessEvent(context.Background(), args, reply)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1GetChargersForEventNilErr(t *testing.T) {
	Cache.Clear(nil)
	dataDB := NewInternalDB(nil, nil, true)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false
	defaultCfg.ChargerSCfg().AttributeSConns = []string{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	dm := NewDataManager(dataDB, defaultCfg.CacheCfg(), nil)

	cP := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		RunID:     utils.MetaDefault,
		FilterIDs: []string{"*string:~*req.Account:1001"},
	}
	if err := dm.SetChargerProfile(context.Background(), cP, true); err != nil {
		t.Fatal(err)
	}

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}
	reply := &ChargerProfiles{}

	exp := &ChargerProfiles{
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			RunID:     "*default",
		},
	}
	err := cS.V1GetChargersForEvent(context.Background(), args, reply)

	if err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}

	if err := dm.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func TestChargersV1GetChargersForEventErr(t *testing.T) {
	Cache.Clear(nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ChargerSCfg().IndexedSelects = false

	dbm := &DataDBMock{
		GetKeysForPrefixF: func(ctx *context.Context, s string) ([]string, error) {
			return []string{":"}, nil
		},
	}
	dm := NewDataManager(dbm, defaultCfg.CacheCfg(), nil)

	cS := &ChargerService{
		dm: dm,
		filterS: &FilterS{
			dm:  dm,
			cfg: defaultCfg,
		},
		cfg: defaultCfg,
	}
	args := &utils.CGREvent{
		ID: "cgrEvID",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
	}
	reply := &ChargerProfiles{}

	exp := &ChargerProfiles{}
	experr := fmt.Sprintf("SERVER_ERROR: %s", utils.ErrNotImplemented)
	err := cS.V1GetChargersForEvent(context.Background(), args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(exp), utils.ToJSON(reply))
	}
}
