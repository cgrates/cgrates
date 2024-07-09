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
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type clMock func(_ string, _ any, _ any) error

func (c clMock) Call(ctx *context.Context, m string, a any, r any) error {
	return c(m, a, r)
}

func TestCDRSV1ProcessCDRNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clMock := clMock(func(_ string, args any, reply any) error {
		rply, cancast := reply.(*AttrSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		newArgs, cancast := args.(*utils.CGREvent)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		if newArgs.Tenant == utils.EmptyString {
			return fmt.Errorf("Tenant is missing")
		}
		*rply = AttrSProcessEventReply{
			AlteredFields: []string{utils.AccountField},
			CGREvent: &utils.CGREvent{
				ID:   "TestBiRPCv1AuthorizeEventNoTenant",
				Time: utils.TimePointer(time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC)),
				Event: map[string]any{
					"Account":     "1002",
					"Category":    "call",
					"Destination": "1003",
					"OriginHost":  "local",
					"OriginID":    "123456",
					"ToR":         "*voice",
					"Usage":       "10s",
				},
			},
		}
		return nil
	})
	chanClnt := make(chan birpc.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanClnt,
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMngr,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	cdr := &CDRWithAPIOpts{ // no tenant, take the default
		CDR: &CDR{
			CGRID:       "Cdr1",
			OrderID:     123,
			ToR:         utils.MetaVoice,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.MetaRated,
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(0),
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
	}
	var reply string
	if err := cdrs.V1ProcessCDR(context.Background(), cdr, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRSV1ProcessEventNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ string, args any, reply any) error {
		rply, cancast := reply.(*[]*ChrgSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		newArgs, cancast := args.(*utils.CGREvent)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		if newArgs.Tenant == utils.EmptyString {
			return fmt.Errorf("Tenant is missing")
		}
		*rply = []*ChrgSProcessEventReply{}
		return nil
	})
	chanClnt := make(chan birpc.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanClnt,
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMngr,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	args := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute,
			},
		},
	}
	var reply string

	if err := cdrs.V1ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
	}
}

func TestCDRSV1V1ProcessExternalCDRNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ string, args any, reply any) error {
		rply, cancast := reply.(*[]*ChrgSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		newArgs, cancast := args.(*utils.CGREvent)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		if newArgs.Tenant == utils.EmptyString {
			return fmt.Errorf("Tenant is missing")
		}
		*rply = []*ChrgSProcessEventReply{}
		return nil
	})
	chanClnt := make(chan birpc.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanClnt,
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMngr,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}

	args := &ExternalCDRWithAPIOpts{
		ExternalCDR: &ExternalCDR{
			ToR:         utils.MetaVoice,
			OriginID:    "testDspCDRsProcessExternalCDR",
			OriginHost:  "127.0.0.1",
			Source:      utils.UnitTest,
			RequestType: utils.MetaRated,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1003",
			Subject:     "1003",
			Destination: "1001",
			SetupTime:   "2014-08-04T13:00:00Z",
			AnswerTime:  "2014-08-04T13:00:07Z",
			Usage:       "1s",
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	var reply string

	if err := cdrs.V1ProcessExternalCDR(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}

func TestArgV1ProcessClone(t *testing.T) {
	attr := &ArgV1ProcessEvent{
		Flags: []string{"flg,flg2,flg3"},
		CGREvent: utils.CGREvent{
			ID:   "TestBiRPCv1AuthorizeEventNoTenant",
			Time: utils.TimePointer(time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC)),
			Event: map[string]any{
				"Account":     "1002",
				"Category":    "call",
				"Destination": "1003",
				"OriginHost":  "local",
				"OriginID":    "123456",
				"ToR":         "*voice",
				"Usage":       "10s",
			},
		},
		clnb: true,
	}
	if val := attr.Clone(); reflect.DeepEqual(attr, val) {
		t.Errorf("expected %v,received %v", utils.ToJSON(val), utils.ToJSON(attr))
	}

}

func TestCDRV1CountCDRs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultTimezone = "UTC"
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	args := &utils.RPCCDRsFilterWithAPIOpts{

		RPCCDRsFilter: &utils.RPCCDRsFilter{},
		Tenant:        "cgrates.org",
		APIOpts:       map[string]any{},
	}

	i := int64(3)
	if err := cdrS.V1GetCDRsCount(context.Background(), args, &i); err != nil {
		t.Error(err)
	}
}
func TestV1CountCDRsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultTimezone = "UTC"
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts:       []string{"1001"},
			RunIDs:         []string{utils.MetaDefault},
			SetupTimeStart: "fdd",
		},
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAPIKey: "cdrs12345",
		},
	}
	i := utils.Int64Pointer(23)
	if err := cdrS.V1GetCDRsCount(context.Background(), args, i); err == nil {
		t.Error(err)
	}
}
func TestV1RateCDRs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultTimezone = "UTC"
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	arg := &ArgRateCDRs{
		Flags:         []string{utils.MetaAttributes, utils.MetaStats, utils.MetaExport, utils.MetaStore, utils.OptsThresholdS, utils.MetaThresholds, utils.MetaStats, utils.OptsChargerS, utils.MetaChargers, utils.OptsRALs, utils.MetaRALs, utils.OptsRerate, utils.MetaRerate, utils.OptsRefund, utils.MetaRefund},
		Tenant:        "cgrates.rg",
		RPCCDRsFilter: utils.RPCCDRsFilter{},
		APIOpts:       map[string]any{},
	}

	var reply string
	if err := cdrS.V1RateCDRs(context.Background(), arg, &reply); err == nil {
		t.Error(err)
	}

}

func TestCDRServerThdsProcessEvent(t *testing.T) {
	clMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {

				rpl := &[]string{"event"}

				*reply.(*[]string) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- clMock
	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ThresholdSConnsCfg): clientconn,
	})
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ThreshSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMgr,
		cdrDb:   db,
		dm:      dm,
	}
	crgEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "id",
		Time:   utils.TimePointer(time.Date(2019, 12, 1, 15, 0, 0, 0, time.UTC)),
	}

	if err := cdrS.thdSProcessEvent(crgEv); err != nil {
		t.Error(err)
	}

}
func TestCDRServerStatSProcessEvent(t *testing.T) {
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {

				rpl := &[]string{"status"}

				*reply.(*[]string) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg): clientconn,
	})
	cfg.CdrsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.StatSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltrs := NewFilterS(cfg, connMgr, dm)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		dm:      dm,
		filterS: fltrs,
		cdrDb:   db,
		connMgr: connMgr,
	}
	crgEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "id",
		Time:   utils.TimePointer(time.Date(2019, 12, 1, 15, 0, 0, 0, time.UTC)),
	}

	if err := cdrS.statSProcessEvent(crgEv); err != nil {
		t.Error(err)
	}
}

func TestCDRServerEesProcessEvent(t *testing.T) {
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpls := &map[string]map[string]any{
					"eeS": {
						"process": "event",
					},
				}
				*reply.(*map[string]map[string]any) = *rpls

				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock

	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.EEsConnsCfg): clientconn,
	})
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.EEsConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), connMgr)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}

	cgrEv := &CGREventWithEeIDs{
		EeIDs: []string{"ees"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "id",
			Time:   utils.TimePointer(time.Date(2019, 12, 1, 15, 0, 0, 0, time.UTC)),
		},
	}
	if err := cdrS.eeSProcessEvent(cgrEv); err != nil {
		t.Error(err)
	}

}

func TestCDRefundEventCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderRefundIncrements: func(ctx *context.Context, args, reply any) error {
				return nil
			},
		},
	}
	ec := &EventCost{
		CGRID: "event",
		RunID: "runid",
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ResponderRefundIncrements): clientconn,
	})
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ResponderRefundIncrements)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	if _, err := cdrS.refundEventCost(ec, "*postpaid", "tor"); err != nil {
		t.Error(err)
	}
}

func TestGetCostFromRater(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{

			utils.ResponderDebit: func(ctx *context.Context, args, reply any) error {
				rpl := &CallCost{
					Category: "category",
					Tenant:   "cgrates",
				}
				*reply.(*CallCost) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg): clientconn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMgr,
		cdrDb:   db,
		dm:      dm,
	}
	cd := &CDR{
		Category:    "category",
		Tenant:      "cgrates.org",
		RequestType: utils.PseudoPrepaid,
	}
	cdr := &CDRWithAPIOpts{
		CDR:     cd,
		APIOpts: map[string]any{},
	}
	exp := &CallCost{
		Category: "category",
		Tenant:   "cgrates",
	}
	if val, err := cdrS.getCostFromRater(cdr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, val) {
		t.Errorf("expected %+v ,received %+v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestRefundEventCost(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderRefundIncrements: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg): clientconn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	ec := &EventCost{
		CGRID:     "cgrid",
		RunID:     "rnID",
		StartTime: time.Date(2022, 12, 1, 11, 0, 0, 0, time.UTC),
		Charges: []*ChargingInterval{
			{
				CompressFactor: 2,
				Increments: []*ChargingIncrement{
					{
						Usage: 10 * time.Minute,
						Cost:  20,
					}, {
						Usage: 5 * time.Minute,
						Cost:  15,
					},
				},
			}, {},
		},
	}
	if val, err := cdrS.refundEventCost(ec, utils.MetaPrepaid, "tor"); err != nil {
		t.Error(err)
	} else if !val {
		t.Error("expected true")
	}
}

func TestCDRSV2ProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1

	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ string, args any, reply any) error {

		return nil
	})
	chanClnt := make(chan birpc.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanClnt,
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMngr,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}

	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	args := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute,
			},
		},
	}
	evs := &[]*utils.EventWithFlags{}

	if err := cdrs.V2ProcessEvent(context.Background(), args, evs); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: evs,
		Error:  nil}
	cacheKey := utils.ConcatenatedKey(utils.CDRsV2ProcessEvent, args.CGREvent.ID)
	rcv, has := Cache.Get(utils.CacheRPCResponses, cacheKey)
	if !has {
		t.Error("expected value")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))

	}

}
func TestCDRSV2ProcessEventCacheSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	args := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute,
			},
		},
	}
	evs := &[]*utils.EventWithFlags{}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV2ProcessEvent, args.CGREvent.ID),
		&utils.CachedRPCResponse{Result: evs, Error: nil},
		nil, true, utils.NonTransactional)

	if err := cdrs.V2ProcessEvent(context.Background(), args, evs); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{Result: evs, Error: nil}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV2ProcessEvent, args.CGREvent.ID))
	if !has {
		t.Error("expected to have a values")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestCDRSV1ProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1

	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ string, args any, reply any) error {

		return nil
	})
	chanClnt := make(chan birpc.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): chanClnt,
	})
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), connMngr)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMngr,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}

	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	args := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute,
			},
		},
	}
	reply := utils.StringPointer("result")

	if err := cdrs.V1ProcessEvent(context.Background(), args, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil}
	cacheKey := utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, args.CGREvent.ID)
	rcv, has := Cache.Get(utils.CacheRPCResponses, cacheKey)
	if !has {
		t.Error("expected value")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))

	}

}
func TestCDRSV1ProcessEventCacheSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	args := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaChargers},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute,
			},
		},
	}
	reply := utils.StringPointer("result")
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, args.CGREvent.ID),
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)

	if err := cdrs.V1ProcessEvent(context.Background(), args, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{Result: reply, Error: nil}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, args.CGREvent.ID))
	if !has {
		t.Error("expected to have a values")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestV1ProcessEvent(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(cfg)
	}()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.AttributeSConnsCfg)}
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)}
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.EEsConnsCfg)}
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	cfg.CdrsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.CdrsCfg().StoreCdrs = true
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpl := &AttrSProcessEventReply{
					AlteredFields: []string{"*req.OfficeGroup"},
					CGREvent: &utils.CGREvent{
						Event: map[string]any{
							utils.CGRID: "cgrid",
						},
					},
				}
				*reply.(*AttrSProcessEventReply) = *rpl

				return nil
			},
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpl := []*ChrgSProcessEventReply{
					{
						ChargerSProfile:    "chrgs1",
						AttributeSProfiles: []string{"attr1", "attr2"},
						CGREvent: &utils.CGREvent{
							Event: map[string]any{
								utils.CGRID: "cgrid2",
							},
						},
					},
				}
				*reply.(*[]*ChrgSProcessEventReply) = rpl
				return nil
			},
			utils.ResponderRefundIncrements: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{
					ID: "cgrates.org:1001",
					BalanceMap: map[string]Balances{
						utils.MetaMonetary: {
							&Balance{Value: 20},
						}}}
				*reply.(*Account) = *rpl
				return nil
			},
			utils.ResponderDebit: func(ctx *context.Context, args, reply any) error {
				rpl := &CallCost{}
				*reply.(*CallCost) = *rpl
				return nil
			},
			utils.ResponderGetCost: func(ctx *context.Context, args, reply any) error {
				rpl := &CallCost{}
				*reply.(*CallCost) = *rpl
				return nil
			},
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpl := &map[string]map[string]any{}
				*reply.(*map[string]map[string]any) = *rpl
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpl := &[]string{}
				*reply.(*[]string) = *rpl
				return nil
			},
			utils.StatSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				rpl := &[]string{}
				*reply.(*[]string) = *rpl
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.AttributeSConnsCfg): clientconn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):       clientconn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder):      clientconn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.EEsConnsCfg):        clientconn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     clientconn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):          clientconn,
	})

	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	arg := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaStats, utils.MetaExport, utils.MetaStore, utils.OptsThresholdS, utils.MetaThresholds, utils.MetaStats, utils.OptsChargerS, utils.MetaChargers, utils.OptsRALs, utils.MetaRALs, utils.OptsRerate, utils.MetaRerate, utils.OptsRefund, utils.MetaRefund},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute,
			},
		},
	}
	var reply string
	if err := cdrS.V1ProcessEvent(context.Background(), arg, &reply); err == nil {
		t.Error(err)
	}
}
func TestCdrprocessEventsErrLog(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	buf := new(bytes.Buffer)
	setlog := func(b *bytes.Buffer) {
		utils.Logger.SetLogLevel(4)
		utils.Logger.SetSyslog(nil)
		log.SetOutput(b)
	}
	setlog(buf)
	removelog := func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}
	tmp := Cache
	defer func() {
		removelog()
		Cache = tmp
	}()
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheCDRsTBL: {
			Limit:     3,
			StaticTTL: true,
		},
	}
	Cache.Clear(nil)
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.EeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {

				return utils.ErrPartiallyExecuted
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs): clientConn,
	})
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}

	evs := []*utils.CGREvent{
		{ID: "TestV1ProcessEventNoTenant",
			Event: map[string]any{
				utils.CGRID:        "test1",
				utils.RunID:        utils.MetaDefault,
				utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:        123 * time.Minute},
		}}
	expLog := `with AttributeS`
	if _, err := cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  true,
			chrgS:  true,
			refund: true,
			ralS:   true,
			store:  true,
			reRate: true,
			export: true,
			thdS:   true,
			stS:    true,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	removelog()
	buf2 := new(bytes.Buffer)
	setlog(buf2)
	expLog = `with ChargerS`
	if _, err := cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  true,
			refund: true,
			ralS:   true,
			store:  true,
			reRate: true,
			export: true,
			thdS:   true,
			stS:    true,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf2.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	removelog()
	buf3 := new(bytes.Buffer)
	setlog(buf3)
	Cache.Set(utils.CacheCDRIDs, utils.ConcatenatedKey("test1", utils.MetaDefault), "val", []string{}, true, utils.NonTransactional)
	expLog = `with CacheS`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: false,
			ralS:   true,
			store:  true,
			reRate: false,
			export: true,
			thdS:   true,
			stS:    true,
		}); err == nil || err != utils.ErrExists {
		t.Error(err)
	} else if rcvLog := buf3.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf4 := new(bytes.Buffer)
	removelog()
	setlog(buf4)
	evs[0].Event[utils.AnswerTime] = "time"
	expLog = `could not retrieve previously`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: true,
			ralS:   true,
			store:  true,
			reRate: false,
			export: true,
			thdS:   true,
			stS:    true,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf4.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf5 := new(bytes.Buffer)
	removelog()
	setlog(buf5)
	evs[0].Event[utils.AnswerTime] = time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC)
	expLog = `refunding CDR`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: true,
			ralS:   false,
			store:  false,
			reRate: false,
			export: false,
			thdS:   false,
			stS:    false,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf5.String(); strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	db.db.Set(utils.CacheCDRsTBL, utils.ConcatenatedKey("test1", utils.MetaDefault, "testV1CDRsRefundOutOfSessionCost"), "val", []string{}, true, utils.NonTransactional)

	buf6 := new(bytes.Buffer)
	removelog()
	setlog(buf6)
	expLog = `refunding CDR`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: true,
			ralS:   false,
			store:  true,
			reRate: false,
			export: false,
			thdS:   false,
			stS:    false,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf6.String(); strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf7 := new(bytes.Buffer)
	removelog()
	setlog(buf7)
	expLog = `exporting cdr`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: true,
			ralS:   false,
			store:  false,
			reRate: true,
			export: true,
			thdS:   false,
			stS:    false,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf7.String(); strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf8 := new(bytes.Buffer)
	removelog()
	setlog(buf8)
	expLog = `processing event`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: true,
			ralS:   false,
			store:  false,
			reRate: true,
			export: false,
			thdS:   true,
			stS:    false,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf8.String(); strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf9 := new(bytes.Buffer)
	removelog()
	setlog(buf9)
	expLog = `processing event`
	if _, err = cdrs.processEvents(evs,
		cdrProcessingArgs{
			attrS:  false,
			chrgS:  false,
			refund: true,
			ralS:   false,
			store:  false,
			reRate: true,
			export: false,
			thdS:   false,
			stS:    true,
		}); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf9.String(); strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
}

func TestV1ProcessCDR(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	cd := &CDR{
		CGRID:       "cgrid1",
		RunID:       "run1",
		Category:    "category",
		Tenant:      "cgrates.org",
		RequestType: utils.PseudoPrepaid,
	}
	cdr := &CDRWithAPIOpts{
		CDR:     cd,
		APIOpts: map[string]any{},
	}
	reply := utils.StringPointer("reply")

	if err = cdrS.V1ProcessCDR(context.Background(), cdr, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID))

	if !has {
		t.Errorf("has no value")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestV1ProcessCDRSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
	}
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	cd := &CDR{
		CGRID:       "cgrid1",
		RunID:       "run1",
		Category:    "category",
		Tenant:      "cgrates.org",
		RequestType: utils.PseudoPrepaid,
	}
	cdr := &CDRWithAPIOpts{
		CDR:     cd,
		APIOpts: map[string]any{},
	}
	reply := utils.StringPointer("reply")
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID),
		&utils.CachedRPCResponse{Result: reply, Error: err},
		nil, true, utils.NonTransactional)
	if err = cdrS.V1ProcessCDR(context.Background(), cdr, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID))
	if !has {
		t.Errorf("has no value")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestV1StoreSessionCost(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	clMock := clMock(func(_ string, _, _ any) error {
		return nil
	})
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- clMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{})
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	attr := &AttrCDRSStoreSMCost{
		Cost: &SMCost{
			CGRID:    "cgrid1",
			RunID:    "run1",
			OriginID: "originid",
			CostDetails: &EventCost{
				Usage: utils.DurationPointer(1 * time.Minute),
				Cost:  utils.Float64Pointer(32.3),
			},
		},
		CheckDuplicate: false,
	}
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	reply := utils.StringPointer("reply")
	if err = cdrS.V1StoreSessionCost(context.Background(), attr, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, attr.Cost.CGRID, attr.Cost.RunID))
	if !has {
		t.Errorf("has no value")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
	attr.Cost.CGRID = utils.EmptyString
	if err := cdrS.V1StoreSessionCost(context.Background(), attr, reply); err == nil || err.Error() != fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing) {
		t.Error(err)
	}
}

func TestV1StoreSessionCostSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	clMock := clMock(func(_ string, _, _ any) error {

		return nil
	})
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- clMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{})
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	attr := &AttrCDRSStoreSMCost{
		Cost: &SMCost{
			CGRID:    "cgrid1",
			RunID:    "run1",
			OriginID: "originid",
			CostDetails: &EventCost{
				Usage: utils.DurationPointer(1 * time.Minute),
				Cost:  utils.Float64Pointer(32.3),
			},
		},
		CheckDuplicate: false,
	}
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	reply := utils.StringPointer("reply")
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, attr.Cost.CGRID, attr.Cost.RunID),
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)

	if err = cdrS.V1StoreSessionCost(context.Background(), attr, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, attr.Cost.CGRID, attr.Cost.RunID))

	if !has {
		t.Errorf("has no value")
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
	cdrS.guard = guardian.Guardian
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	attr.CheckDuplicate = true
	if err = cdrS.V1StoreSessionCost(context.Background(), attr, reply); err != nil {
		t.Error(err)
	}
}

func TestV2StoreSessionCost(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderRefundRounding: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg): clientconn,
	})
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	args := &ArgsV2CDRSStoreSMCost{
		CheckDuplicate: false,
		Cost: &V2SMCost{
			CGRID:      "cgrid",
			RunID:      "runid",
			OriginHost: "host",
			OriginID:   "originid",
			CostSource: "cgrates",
			CostDetails: &EventCost{
				CGRID:          "evcgrid",
				RunID:          "evrunid",
				StartTime:      time.Date(2021, 11, 1, 2, 0, 0, 0, time.UTC),
				Usage:          utils.DurationPointer(122),
				Cost:           utils.Float64Pointer(134),
				Charges:        []*ChargingInterval{},
				AccountSummary: &AccountSummary{},
				Rating:         Rating{},
				Accounting:     Accounting{},
				RatingFilters:  RatingFilters{},
				Rates:          ChargedRates{},
				Timings:        ChargedTimings{},
			},
		},
	}
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	var reply string

	if err := cdrS.V2StoreSessionCost(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{Result: utils.StringPointer("OK"), Error: nil}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID))

	if !has {
		t.Error("has no value")
	}
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
	args = &ArgsV2CDRSStoreSMCost{
		Cost: &V2SMCost{},
	}
	if err = cdrS.V2StoreSessionCost(context.Background(), args, &reply); err == nil {
		t.Error(err)
	}
}

func TestV2StoreSessionCostSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderRefundRounding: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg): clientconn,
	})
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	config.SetCgrConfig(cfg)
	Cache = NewCacheS(cfg, dm, nil)
	args := &ArgsV2CDRSStoreSMCost{
		CheckDuplicate: false,
		Cost: &V2SMCost{
			CGRID:      "cgrid",
			RunID:      "runid",
			OriginHost: "host",
			OriginID:   "originid",
			CostSource: "cgrates",
			CostDetails: &EventCost{
				CGRID:     "evcgrid",
				RunID:     "evrunid",
				StartTime: time.Date(2021, 11, 1, 2, 0, 0, 0, time.UTC),
				Usage:     utils.DurationPointer(122),
				Cost:      utils.Float64Pointer(134),
				Charges: []*ChargingInterval{
					{
						RatingID: "rating1",
						Increments: []*ChargingIncrement{
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id",
								CompressFactor: 5,
							},
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id",
								CompressFactor: 5,
							},
						},
						CompressFactor: 3,
						usage:          utils.DurationPointer(10 * time.Minute),
						ecUsageIdx:     utils.DurationPointer(4 * time.Minute),
						cost:           utils.Float64Pointer(38),
					},
					{
						RatingID: "rating2",
						Increments: []*ChargingIncrement{
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id",
								CompressFactor: 5,
							},
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id",
								CompressFactor: 5,
							},
						},
						CompressFactor: 3,
						usage:          utils.DurationPointer(10 * time.Minute),
						ecUsageIdx:     utils.DurationPointer(4 * time.Minute),
						cost:           utils.Float64Pointer(38),
					},
				},
				AccountSummary: &AccountSummary{
					Tenant:           "Tenant",
					ID:               "acc_id",
					BalanceSummaries: BalanceSummaries{},
					AllowNegative:    false,
					Disabled:         true,
				},
				Rating: Rating{
					"rating1": &RatingUnit{},
					"rating2": &RatingUnit{},
				},
				Accounting:    Accounting{},
				RatingFilters: RatingFilters{},
				Rates:         ChargedRates{},
				Timings:       ChargedTimings{},
			},
		},
	}
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	reply := utils.StringPointer("reply")
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID),
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)

	if err := cdrS.V2StoreSessionCost(context.Background(), args, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{Result: reply, Error: nil}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestV1RateCDRSErr(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ChargerSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	arg := &ArgRateCDRs{
		Flags: []string{utils.MetaStore, utils.MetaExport, utils.MetaThresholds, utils.MetaStats, utils.MetaChargers, utils.MetaAttributes},

		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	var reply string

	if err := cdrS.V1RateCDRs(context.Background(), arg, &reply); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestV1GetCDRsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items),
		dm:      dm,
	}
	args := utils.RPCCDRsFilterWithAPIOpts{

		RPCCDRsFilter: &utils.RPCCDRsFilter{
			CGRIDs:         []string{"CGRIDs"},
			NotCGRIDs:      []string{"NotCGRIDs"},
			OriginHosts:    []string{"OriginHosts"},
			SetupTimeStart: "time",
			SetupTimeEnd:   "2020-04-18T11:46:26.371Z",
		},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	var cdrs *[]*CDR
	if err := cdrS.V1GetCDRs(context.Background(), args, cdrs); err == nil {
		t.Error(utils.NewErrServerError(utils.ErrNotFound))
	}
	args.RPCCDRsFilter.SetupTimeStart = ""
	if err := cdrS.V1GetCDRs(context.Background(), args, cdrs); err == nil || err.Error() != fmt.Sprintf("SERVER_ERROR: %s", utils.ErrNotFound) {
		t.Error(utils.NewErrServerError(utils.ErrNotFound))
	}
}
func TestGetCostFromRater2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	cfg.CdrsCfg().SchedulerConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.SchedulerConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderDebit: func(ctx *context.Context, args, reply any) error {

				return utils.ErrAccountNotFound
			},
			utils.SchedulerSv1ExecuteActionPlans: func(ctx *context.Context, args, reply any) error {
				rpl := "reply"
				*reply.(*string) = rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg):     clientconn,
		utils.ConcatenatedKey(utils.MetaInternal, utils.SchedulerConnsCfg): clientconn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	cdr := &CDRWithAPIOpts{

		CDR: &CDR{
			ToR:         "tor",
			Tenant:      "tenant",
			Category:    "cdr",
			Subject:     "cdrsubj",
			Account:     "acc_cdr",
			Destination: "acc_dest",
			RequestType: utils.MetaDynaprepaid,
			Usage:       1 * time.Minute,
		},
		APIOpts: map[string]any{},
	}

	if _, err := cdrS.getCostFromRater(cdr); err == nil || err != utils.ErrAccountNotFound {
		t.Error(err)
	}
}

func TestGetCostFromRater3(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ccMock := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderGetCost: func(ctx *context.Context, args, reply any) error {

				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg): clientconn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	cdr := &CDRWithAPIOpts{

		CDR: &CDR{
			ToR:         "tor",
			Tenant:      "tenant",
			Category:    "cdr",
			Subject:     "cdrsubj",
			Account:     "acc_cdr",
			Destination: "acc_dest",
			RequestType: "default",
			Usage:       1 * time.Minute,
		},
		APIOpts: map[string]any{},
	}

	if _, err := cdrS.getCostFromRater(cdr); err == nil || err != rpcclient.ErrUnsupporteServiceMethod {
		t.Errorf("expected %+v,received %v", rpcclient.ErrUnsupporteServiceMethod, err)
	}
}

func TestV2StoreSessionCost2(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ccMOck := &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderRefundRounding: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan birpc.ClientConnector, 1)
	clientconn <- ccMOck
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg): clientconn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}

	args := &ArgsV2CDRSStoreSMCost{

		CheckDuplicate: false,
		Cost: &V2SMCost{
			CGRID:      "cgrid",
			RunID:      "runid",
			OriginHost: "host",
			OriginID:   "originid",
			CostSource: "cgrates",
			CostDetails: &EventCost{
				CGRID:     "evcgrid",
				RunID:     "evrunid",
				StartTime: time.Date(2021, 11, 1, 2, 0, 0, 0, time.UTC),
				Usage:     utils.DurationPointer(122),
				Cost:      utils.Float64Pointer(134),
				Charges: []*ChargingInterval{
					{

						RatingID: utils.MetaRounding,
						Increments: []*ChargingIncrement{
							{

								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id",
								CompressFactor: 5,
							},
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id2",
								CompressFactor: 5,
							},
						},
						CompressFactor: 3,
						usage:          utils.DurationPointer(10 * time.Minute),
						ecUsageIdx:     utils.DurationPointer(4 * time.Minute),
						cost:           utils.Float64Pointer(38),
					},
					{
						RatingID: utils.MetaRounding,

						Increments: []*ChargingIncrement{
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id",
								CompressFactor: 5,
							},
							{
								Usage:          5 * time.Minute,
								Cost:           23,
								AccountingID:   "acc_id2",
								CompressFactor: 5,
							},
						},
						CompressFactor: 3,
						usage:          utils.DurationPointer(10 * time.Minute),
						ecUsageIdx:     utils.DurationPointer(4 * time.Minute),
						cost:           utils.Float64Pointer(38),
					},
				},
				AccountSummary: &AccountSummary{
					Tenant:           "Tenant",
					ID:               "acc_id",
					BalanceSummaries: BalanceSummaries{},
					AllowNegative:    false,
					Disabled:         true,
				},

				Accounting: Accounting{
					"acc_id": &BalanceCharge{
						RatingID: utils.MetaRounding,
					},
					"acc_id2": &BalanceCharge{
						RatingID: utils.MetaRounding,
					},
				},
				RatingFilters: RatingFilters{
					"filtersid": RatingMatchedFilters{
						utils.Subject:               "string",
						utils.DestinationPrefixName: "string",
						utils.DestinationID:         "dest",
						utils.RatingPlanID:          "rating",
					},
					"ratefilter": RatingMatchedFilters{
						utils.Subject:               "string",
						utils.DestinationPrefixName: "string",
						utils.DestinationID:         "dest",
						utils.RatingPlanID:          "rating",
					},
				},
				Rates:   ChargedRates{},
				Timings: ChargedTimings{},
				Rating: Rating{
					utils.MetaRounding: &RatingUnit{
						ConnectFee:       21,
						RoundingMethod:   "method",
						RoundingDecimals: 3,
						MaxCost:          22,
						MaxCostStrategy:  "sr",
						RatesID:          "rates",
						RatingFiltersID:  "filtersid",
					},
				},
			},
		},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	var reply string
	if err := cdrS.V2StoreSessionCost(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %+v,received %+v", utils.OK, reply)
	}
	clientconn2 := make(chan birpc.ClientConnector, 1)
	clientconn2 <- &ccMock{

		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderRefundRounding: func(ctx *context.Context, args, reply any) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return utils.ErrNotFound
			},
		},
	}
	cdrS.connMgr.rpcInternal[utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)] = clientconn2

	if err := cdrS.V2StoreSessionCost(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}

}
func TestV1RateCDRSSuccesful(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ChargerSConnsCfg)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	cdr := &CDR{
		CGRID:       "Cdr1",
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "OriginCDR1",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.MetaRated,
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(0),
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}
	if err := cdrS.cdrDb.SetCDR(cdr, true); err != nil {
		t.Error(err)
	}
	arg := &ArgRateCDRs{
		Flags: []string{utils.MetaStore, utils.MetaExport, utils.MetaThresholds, utils.MetaStats, utils.MetaChargers, utils.MetaAttributes},
		RPCCDRsFilter: utils.RPCCDRsFilter{
			CGRIDs: []string{"Cdr1"},
		},
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}
	var reply *string

	if err := cdrS.V1RateCDRs(context.Background(), arg, reply); err == nil {
		t.Error(err)
	}
}

func TestCdrServerStoreSMCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ChargerSConnsCfg)}
	db := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheSessionCostsTBL: {
			Limit:     2,
			TTL:       2 * time.Minute,
			StaticTTL: true,
			Remote:    false,
			Replicate: true,
		},
	})
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	smCost := &SMCost{
		CGRID:      "cgrid",
		RunID:      "runid",
		OriginHost: "originhost",
		OriginID:   "origin",
		CostSource: "cost",
		Usage:      1 * time.Minute,
		CostDetails: &EventCost{
			CGRID:          "ecId",
			RunID:          "ecRunId",
			StartTime:      time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			Usage:          utils.DurationPointer(1 * time.Hour),
			Cost:           utils.Float64Pointer(12.1),
			Charges:        []*ChargingInterval{},
			AccountSummary: &AccountSummary{},
			Accounting:     Accounting{},
			RatingFilters:  RatingFilters{},
			Rates:          ChargedRates{},
		},
	}
	guardian := guardian.GuardianLocker{}
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
		guard:   &guardian,
	}
	if err := cdrS.cdrDb.SetSMCost(smCost); err != nil {
		t.Error(err)
	}

}

func TestCdrSRateCDR(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmp := Cache
	defer func() {
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmp
	}()
	Cache.Clear(nil)
	cfg.CdrsCfg().SMCostRetries = 1
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ChargerSConnsCfg)}
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	db := NewInternalDB(nil, nil, true, map[string]*config.ItemOpt{
		utils.CacheSessionCostsTBL: {
			Limit:     2,
			TTL:       2 * time.Minute,
			StaticTTL: true,
			Remote:    false,
			Replicate: true,
		},
	})
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ResponderDebit: func(ctx *context.Context, args, reply any) error {
				cc := &CallCost{
					Category:    "generic",
					Tenant:      "cgrates.org",
					Subject:     "1001",
					Account:     "1001",
					Destination: "data",
					ToR:         "*data",
					Cost:        0,
				}
				*reply.(*CallCost) = *cc
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): clientConn,
	})
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	smc := &SMCost{
		CGRID:       "cgrid",
		RunID:       "runid",
		OriginHost:  "originHost",
		OriginID:    "originID",
		CostSource:  "cost_source",
		Usage:       2 * time.Minute,
		CostDetails: &EventCost{},
	}
	if err := cdrS.cdrDb.SetSMCost(smc); err != nil {
		t.Error(err)
	}
	db.db.Set(utils.CacheSessionCostsTBL, "cgrates:item1", &SMCost{
		CGRID:      "CGRID",
		RunID:      utils.MetaDefault,
		OriginHost: utils.FreeSWITCHAgent,
		OriginID:   "Origin1",
		Usage:      time.Second,
		CostSource: utils.MetaSessionS,
	}, []string{utils.ConcatenatedKey(utils.CGRID, "CGRID_22"),
		utils.ConcatenatedKey(utils.RunID, "RUN_ID1"),
		utils.ConcatenatedKey(utils.OriginHost, "ORG_Host1")}, true, utils.NonTransactional)
	cdrOpts := &CDRWithAPIOpts{
		CDR: &CDR{
			CGRID:       "CGRID_22",
			RunID:       "RUN_ID1",
			OriginHost:  "ORG_Host1",
			OrderID:     222,
			Usage:       4 * time.Second,
			RequestType: utils.Prepaid,
			ExtraFields: map[string]string{
				utils.LastUsed: "extra",
			},
		},
	}

	if _, err := cdrS.rateCDR(cdrOpts); err != nil {
		t.Error(err)
	}

	cdrOpts.ExtraFields = map[string]string{}
	cdrOpts.Usage = 0
	if _, err := cdrS.rateCDR(cdrOpts); err != nil {
		t.Error(err)
	}

	cdrOpts.CostDetails = &EventCost{
		CGRID:     "7636f3f1a06dffa038ba7900fb57f52d28830a24",
		RunID:     utils.MetaDefault,
		StartTime: time.Date(2018, 7, 27, 0, 59, 21, 0, time.UTC),
		Usage:     utils.DurationPointer(2 * time.Second),
		Charges: []*ChargingInterval{
			{
				RatingID: "cc68da4",
				Increments: []*ChargingIncrement{
					{
						Usage:          102400,
						AccountingID:   "0d87a64",
						CompressFactor: 103,
					},
				},
				CompressFactor: 1,
			},
		},
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				{
					UUID:  "9a767726-fe69-4940-b7bd-f43de9f0f8a5",
					ID:    "addon_data",
					Type:  utils.MetaData,
					Value: 10726871040},
			},
		},
	}
	if _, err := cdrS.rateCDR(cdrOpts); err != nil {
		t.Error(err)
	}
}

func TestChrgrSProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clienConn := make(chan birpc.ClientConnector, 1)
	clienConn <- &ccMock{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ChargerSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				*reply.(*[]*ChrgSProcessEventReply) = []*ChrgSProcessEventReply{
					{
						ChargerSProfile:    "Charger1",
						AttributeSProfiles: []string{"cgrates.org:ATTR_1001_SIMPLEAUTH"},
						AlteredFields:      []string{utils.MetaReqRunID, "*req.Password"},
						CGREvent: &utils.CGREvent{ // matching Charger1
							Tenant: "cgrates.org",
							ID:     "event1",
							Event: map[string]any{
								utils.AccountField: "1001",
								"Password":         "CGRateS.org",
								"RunID":            utils.MetaDefault,
							},
							APIOpts: map[string]any{
								utils.OptsContext:              "simpleauth",
								utils.MetaSubsys:               utils.MetaChargers,
								utils.OptsAttributesProfileIDs: []string{"ATTR_1001_SIMPLEAUTH"},
							},
						},
					},
				}
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): clienConn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	cgrEv := utils.CGREvent{
		ID: "TestV1ProcessEventNoTenant",
		Event: map[string]any{
			utils.CGRID:        "test1",
			utils.RunID:        utils.MetaDefault,
			utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
			utils.Destination:  "+4986517174963",
			utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
			utils.Usage:        123 * time.Minute,
		},
	}
	expcgrEv := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.AccountField: "1001",
				"Password":         "CGRateS.org",
				"RunID":            utils.MetaDefault,
			},
			APIOpts: map[string]any{
				utils.OptsContext:              "simpleauth",
				utils.MetaSubsys:               utils.MetaChargers,
				utils.OptsAttributesProfileIDs: []string{"ATTR_1001_SIMPLEAUTH"},
			},
		},
	}
	if val, err := cdrS.chrgrSProcessEvent(&cgrEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(val, expcgrEv) {
		t.Errorf("expected %v,received %v", utils.ToJSON(expcgrEv), utils.ToJSON(val))
	}
}

func TestCDRServerListenAndServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpConnMgr := connMgr
	defer func() {
		connMgr = tmpConnMgr
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMgr,
		dm:      dm,
		cdrDb:   db,
	}
	stopChan := make(chan struct{}, 1)
	go func() {
		time.Sleep(10 * time.Millisecond)
		stopChan <- struct{}{}
	}()
	cdrS.ListenAndServe(stopChan)
}

func TestCDRServerListenAndServe2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpConnMgr := connMgr
	defer func() {
		connMgr = tmpConnMgr
		config.SetCgrConfig(config.NewDefaultCGRConfig())
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMgr,
		dm:      dm,
		cdrDb:   db,
	}
	stopChan := make(chan struct{}, 1)
	go func() {
		time.Sleep(40 * time.Millisecond)
		stopChan <- struct{}{}
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cdrS.storDBChan <- db
	}()
	cdrS.ListenAndServe(stopChan)
}

func TestStoreSMCostErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	tmpConnMgr := connMgr
	tmpCache := Cache
	defer func() {
		connMgr = tmpConnMgr
		config.SetCgrConfig(config.NewDefaultCGRConfig())
		Cache = tmpCache
	}()
	Cache.Clear(nil)
	cfg.DataDbCfg().Items = map[string]*config.ItemOpt{
		utils.CacheSessionCostsTBL: {
			Replicate: true,
		},
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	db.db.Set(utils.CacheSessionCostsTBL, "CGRID:CGRID", nil, []string{"GRP_1"}, true, utils.NonTransactional)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		connMgr: connMgr,
		dm:      dm,
		cdrDb:   db,
		guard:   guardian.Guardian,
	}
	smCost := &SMCost{
		CGRID:      "CGRID",
		RunID:      utils.MetaDefault,
		OriginHost: utils.FreeSWITCHAgent,
		OriginID:   "Origin1",
		Usage:      time.Second,
		CostSource: utils.MetaSessionS,
		CostDetails: &EventCost{
			CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
			RunID:     utils.MetaDefault,
			StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
			Charges: []*ChargingInterval{
				{
					RatingID: "c1a5ab9",
					Increments: []*ChargingIncrement{
						{
							Usage:          0,
							Cost:           0.1,
							AccountingID:   "9bdad10",
							CompressFactor: 1,
						},
					},
					CompressFactor: 1,
				},
			},
			AccountSummary: &AccountSummary{
				Tenant: "cgrates.org",
				ID:     "dan",
				BalanceSummaries: []*BalanceSummary{
					{
						UUID:  "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
						Type:  utils.MetaMonetary,
						Value: 50,
					},
				},
				AllowNegative: false,
				Disabled:      false,
			},
			Rating: Rating{
				"3cd6425": &RatingUnit{
					RoundingMethod:   "*up",
					RoundingDecimals: 5,
					TimingID:         "7f324ab",
					RatesID:          "4910ecf",
					RatingFiltersID:  "43e77dc",
				},
			},
			Accounting: Accounting{
				"a012888": &BalanceCharge{
					AccountID:   "cgrates.org:dan",
					BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
					Units:       0.01,
				},
			},
			RatingFilters: RatingFilters{
				"43e77dc": RatingMatchedFilters{
					"DestinationID":     "GERMANY",
					"DestinationPrefix": "+49",
					"RatingPlanID":      "RPL_RETAIL1",
					"Subject":           "*out:cgrates.org:call:*any",
				},
			},
			Rates: ChargedRates{
				"ec1a177": RateGroups{
					&RGRate{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Minute,
						RateUnit:           time.Second},
				},
			},
			Timings: ChargedTimings{
				"7f324ab": &ChargedTiming{
					StartTime: "00:00:00",
				}}}}

	if err := cdrS.storeSMCost(smCost, true); err != nil {
		t.Error(err)
	}
}

func TestCDRSGetCDRs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)}
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	cfg.CdrsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.CdrsCfg().SMCostRetries = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(serviceMethod string, _, _ any) error {
		if serviceMethod == utils.EeSv1ProcessEvent {

			return nil
		}
		if serviceMethod == utils.ThresholdSv1ProcessEvent {
			return nil
		}
		if serviceMethod == utils.StatSv1ProcessEvent {
			return nil
		}

		return utils.ErrNotImplemented
	})
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
		connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs):        clientConn,
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn,
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):      clientConn,
		}),
	}
	arg := &ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event1",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginHost:   "host",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.OriginID:     "OriginCDR2",
				utils.CGRID:        "TestCGRID",
				utils.Subject:      "1001",
				utils.RunID:        utils.MetaDefault,
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
			APIOpts: map[string]any{
				utils.OptsStatS: true,
			},
		},
		Flags: []string{
			utils.MetaStore, utils.MetaRALs, utils.MetaExport, utils.MetaThresholds,
		},
	}
	var reply string
	if err := cdrS.V1ProcessEvent(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Expected OK")
	}

	var cdrs []*CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{RequestTypes: []string{utils.MetaPrepaid}},
	}
	if err := cdrS.V1GetCDRs(context.Background(), args, &cdrs); err != nil {
		t.Error(err)
	}
}

func TestV1RateCDRsSuccesful(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: nil,
	}
	cdr := &CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2023, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1002",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2023, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2023, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		Cost:        1.01,
	}
	if err := db.SetCDR(cdr, true); err != nil {
		t.Error(err)
	}
	var reply string
	arg := &ArgRateCDRs{
		Flags:  []string{utils.MetaRerate},
		Tenant: "cgrates.org",
		RPCCDRsFilter: utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
		},
		APIOpts: map[string]any{},
	}
	if err := cdrS.V1RateCDRs(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Expected reply to be ok")
	}
}

func TestCdrsSetCloneable(t *testing.T) {
	tests := []struct {
		input    bool
		expected bool
	}{
		{input: true, expected: true},
		{input: false, expected: false},
	}
	for _, tt := range tests {
		attr := &ArgV1ProcessEvents{
			Flags:     []string{},
			CGREvents: []*utils.CGREvent{},
			APIOpts:   make(map[string]any),
			clnb:      !tt.input,
		}
		attr.SetCloneable(tt.input)
		if attr.clnb != tt.expected {
			t.Errorf("SetCloneable(%v) = %v; expected %v", tt.input, attr.clnb, tt.expected)
		}
	}
}

func TestCdrsSetCloneableEvent(t *testing.T) {
	arg := &ArgV1ProcessEvent{}
	arg.SetCloneable(true)
	if !arg.clnb {
		t.Errorf("expected clnb to be true, got false")
	}
	arg.SetCloneable(false)

	if arg.clnb {
		t.Errorf("expected clnb to be false, got true")
	}
}

func TestCdrsRPCClone(t *testing.T) {

	arg := &ArgV1ProcessEvent{}

	cloned1, err1 := arg.RPCClone()

	if err1 != nil {
		t.Errorf("unexpected error: %v", err1)
	}
	if cloned1 != arg {
		t.Errorf("expected cloned object to be identical, got different objects")
	}
	arg.SetCloneable(true)
	cloned2, err2 := arg.RPCClone()
	if err2 != nil {
		t.Errorf("unexpected error: %v", err2)
	}
	if cloned2 == arg {
		t.Errorf("expected cloned object to be different, got identical objects")
	}
}

func TestCdrsRPCCloneArgs(t *testing.T) {
	arg := &ArgV1ProcessEvents{}

	cloned1, err1 := arg.RPCClone()
	if err1 != nil {
		t.Errorf("unexpected error: %v", err1)
	}
	if cloned1 != arg {
		t.Errorf("expected cloned object to be same, got different")
	}
	arg.SetCloneable(true)
	cloned2, err2 := arg.RPCClone()
	if err2 != nil {
		t.Errorf("unexpected error: %v", err2)
	}
	if cloned2 == arg {
		t.Errorf("expected cloned object to be different, got same")
	}
}

func TestNewMapEventFromReqForm(t *testing.T) {
	form := url.Values{}
	form.Add("key1", "value1")
	form.Add("key2", "value2")

	req, err := http.NewRequest("POST", "http://cgrates.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Form = form
	mp, err := newMapEventFromReqForm(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedRemoteAddr := req.RemoteAddr
	if mp[utils.Source] != expectedRemoteAddr {
		t.Errorf("expected mp[%q] to be %q, got %q", utils.Source, expectedRemoteAddr, mp[utils.Source])
	}
	if mp["key1"] != "value1" {
		t.Errorf("expected mp[%q] to be %q, got %q", "key1", "value1", mp["key1"])
	}
	if mp["key2"] != "value2" {
		t.Errorf("expected mp[%q] to be %q, got %q", "key2", "value2", mp["key2"])
	}
}

func TestCDRSV1ProcessEvents(t *testing.T) {

	ctx := context.Background()
	cfg := config.NewDefaultCGRConfig()
	cdrS := &CDRServer{
		cgrCfg: cfg,
	}
	arg := &ArgV1ProcessEvents{
		Flags:     []string{},
		CGREvents: []*utils.CGREvent{},
		APIOpts:   make(map[string]any),
	}
	expectedReply := utils.OK
	var reply string
	err := cdrS.V1ProcessEvents(ctx, arg, &reply)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if reply != expectedReply {
		t.Errorf("Expected reply %q, got %q", expectedReply, reply)
	}
}

func TestCDRSCallValidServiceMethod(t *testing.T) {
	cdrS := &CDRServer{}

	args := struct{}{}
	reply := new(struct{})

	err := cdrS.Call(nil, "CDRServer.testMethod", args, reply)

	if err == nil {
		t.Errorf("UNSUPPORTED_SERVICE_METHOD, got %v", err)
	}
}

func TestCDRSCallInvalidServiceMethod(t *testing.T) {
	cdrS := &CDRServer{}

	args := struct{}{}
	reply := new(struct{})

	err := cdrS.Call(nil, "CDRServer.InvalidMethod", args, reply)

	if err != rpcclient.ErrUnsupporteServiceMethod {
		t.Errorf("Expected error %v, got %v", rpcclient.ErrUnsupporteServiceMethod, err)
	}
}

func TestNewMapEventFromReqForm_ParseForm(t *testing.T) {
	formData := url.Values{}
	formData.Add("key", "value")
	req := httptest.NewRequest("POST", "/", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-cgrates-urlencoded")
	_, err := newMapEventFromReqForm(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
