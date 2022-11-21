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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type clMock func(_ string, _ interface{}, _ interface{}) error

func (c clMock) Call(m string, a interface{}, r interface{}) error {
	return c(m, a, r)
}

func TestCDRSV1ProcessCDRNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {
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
				Event: map[string]interface{}{
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
	chanClnt := make(chan rpcclient.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	if err := cdrs.V1ProcessCDR(cdr, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRSV1ProcessEventNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {
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
	chanClnt := make(chan rpcclient.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
			Event: map[string]interface{}{
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

	if err := cdrs.V1ProcessEvent(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRSV1V1ProcessExternalCDRNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {
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
	chanClnt := make(chan rpcclient.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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

	if err := cdrs.V1ProcessExternalCDR(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestArgV1ProcessClone(t *testing.T) {
	attr := &ArgV1ProcessEvent{
		Flags: []string{"flg,flg2,flg3"},
		CGREvent: utils.CGREvent{
			ID:   "TestBiRPCv1AuthorizeEventNoTenant",
			Time: utils.TimePointer(time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC)),
			Event: map[string]interface{}{
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

func TestCDRV1GetCDRs(t *testing.T) {
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
		APIOpts:       map[string]interface{}{},
	}
	cdrs := &[]*CDR{
		{
			CGRID: "cgrid"},
		{
			CGRID: "cgr1d",
		},
	}
	if err := cdrS.V1GetCDRs(*args, cdrs); err == nil {
		t.Error(err)
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
		APIOpts:       map[string]interface{}{},
	}

	i := int64(3)
	if err := cdrS.V1CountCDRs(args, &i); err != nil {
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
		Flags:         []string{"flag1", "flag2", "flag3"},
		Tenant:        "cgrates",
		RPCCDRsFilter: utils.RPCCDRsFilter{},
		APIOpts:       map[string]interface{}{},
	}

	reply := "reply"
	if err := cdrS.V1RateCDRs(arg, &reply); err == nil {
		t.Error(err)
	}

}

func TestCDRServerThdsProcessEvent(t *testing.T) {
	clMock := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(args, reply interface{}) error {

				rpl := &[]string{"event"}

				*reply.(*[]string) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- clMock
	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.StatSv1ProcessEvent: func(args, reply interface{}) error {

				rpl := &[]string{"status"}

				*reply.(*[]string) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(args, reply interface{}) error {
				rpls := &map[string]map[string]interface{}{
					"eeS": {
						"process": "event",
					},
				}
				*reply.(*map[string]map[string]interface{}) = *rpls

				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock

	cfg := config.NewDefaultCGRConfig()
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundIncrements: func(args, reply interface{}) error {
				return nil
			},
		},
	}
	ec := &EventCost{
		CGRID: "event",
		RunID: "runid",
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		calls: map[string]func(args interface{}, reply interface{}) error{

			utils.ResponderDebit: func(args, reply interface{}) error {
				rpl := &CallCost{
					Category: "category",
					Tenant:   "cgrates",
				}
				*reply.(*CallCost) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		APIOpts: map[string]interface{}{},
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundIncrements: func(args, reply interface{}) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {

		return nil
	})
	chanClnt := make(chan rpcclient.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
			Event: map[string]interface{}{
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

	if err := cdrs.V2ProcessEvent(args, evs); err != nil {
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
			Event: map[string]interface{}{
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

	if err := cdrs.V2ProcessEvent(args, evs); err != nil {
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
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {

		return nil
	})
	chanClnt := make(chan rpcclient.ClientConnector, 1)
	chanClnt <- clMock
	connMngr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
			Event: map[string]interface{}{
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

	if err := cdrs.V1ProcessEvent(args, reply); err != nil {
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
			Event: map[string]interface{}{
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

	if err := cdrs.V1ProcessEvent(args, reply); err != nil {
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
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.AttributeSConnsCfg)}
	cfg.CdrsCfg().StoreCdrs = true

	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	ccMock := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := &AttrSProcessEventReply{}
				*reply.(*AttrSProcessEventReply) = *rpl

				return nil
			},
			utils.ChargerSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := []*ChrgSProcessEventReply{
					{
						ChargerSProfile:    "chrgs1",
						AttributeSProfiles: []string{"attr1", "attr2"},
					}, {
						ChargerSProfile:    "chrgs1",
						AttributeSProfiles: []string{"attr1", "attr2"},
					},
				}
				*reply.(*[]*ChrgSProcessEventReply) = rpl
				return nil
			},
		},
	}

	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.AttributeSConnsCfg): clientconn,
	})

	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	arg := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaExport, utils.MetaStore, utils.OptsThresholdS, utils.MetaThresholds, utils.MetaStats, utils.OptsChargerS, utils.MetaChargers, utils.OptsRALs, utils.MetaRALs, utils.OptsRerate, utils.MetaRerate, utils.OptsRefund, utils.MetaRefund},
		CGREvent: utils.CGREvent{
			ID: "TestV1ProcessEventNoTenant",
			Event: map[string]interface{}{
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
	if err := cdrS.V1ProcessEvent(arg, utils.StringPointer("val")); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
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
		APIOpts: map[string]interface{}{},
	}
	reply := utils.StringPointer("reply")

	if err = cdrS.V1ProcessCDR(cdr, reply); err != nil {
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
		APIOpts: map[string]interface{}{},
	}
	reply := utils.StringPointer("reply")
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID),
		&utils.CachedRPCResponse{Result: reply, Error: err},
		nil, true, utils.NonTransactional)
	if err = cdrS.V1ProcessCDR(cdr, reply); err != nil {
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
	clMock := clMock(func(_ string, _, _ interface{}) error {

		return nil
	})
	clientconn := make(chan rpcclient.ClientConnector, 1)

	clientconn <- clMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{})
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

	if err = cdrS.V1StoreSessionCost(attr, reply); err != nil {
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
}

func TestV1StoreSessionCostSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	clMock := clMock(func(_ string, _, _ interface{}) error {

		return nil
	})
	clientconn := make(chan rpcclient.ClientConnector, 1)

	clientconn <- clMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{})
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

	if err = cdrS.V1StoreSessionCost(attr, reply); err != nil {
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
}

func TestV2StoreSessionCost(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	ccMock := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundRounding: func(args, reply interface{}) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	reply := utils.StringPointer("reply")

	if err := cdrS.V2StoreSessionCost(args, reply); err != nil {
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
}

func TestV2StoreSessionCostSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	ccMock := &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundRounding: func(args, reply interface{}) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	reply := utils.StringPointer("reply")
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID),
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)

	if err := cdrS.V2StoreSessionCost(args, reply); err != nil {
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

func TestV1RateCDRS(t *testing.T) {
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
		RPCCDRsFilter: utils.RPCCDRsFilter{
			CGRIDs:          []string{"id", "cgr"},
			NotRequestTypes: []string{"noreq"},
			NotCGRIDs:       []string{"cgrid"},
			RunIDs:          []string{"runid"},
			OriginIDs:       []string{"o_id"},
			NotOriginIDs:    []string{"noid"},
		},
		Tenant:  "cgrates.org",
		APIOpts: map[string]interface{}{},
	}
	reply := utils.StringPointer("reply")

	if err := cdrS.V1RateCDRs(arg, reply); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

}
