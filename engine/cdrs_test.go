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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
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
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
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
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}
	i := utils.Int64Pointer(23)
	if err := cdrS.V1CountCDRs(args, i); err == nil {
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
		APIOpts:       map[string]interface{}{},
	}

	var reply string
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
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := &AttrSProcessEventReply{
					AlteredFields: []string{"*req.OfficeGroup"},
					CGREvent: &utils.CGREvent{
						Event: map[string]interface{}{
							utils.CGRID: "cgrid",
						},
					},
				}
				*reply.(*AttrSProcessEventReply) = *rpl

				return nil
			},
			utils.ChargerSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := []*ChrgSProcessEventReply{
					{
						ChargerSProfile:    "chrgs1",
						AttributeSProfiles: []string{"attr1", "attr2"},
						CGREvent: &utils.CGREvent{
							Event: map[string]interface{}{
								utils.CGRID: "cgrid2",
							},
						},
					},
				}
				*reply.(*[]*ChrgSProcessEventReply) = rpl
				return nil
			},
			utils.ResponderRefundIncrements: func(args, reply interface{}) error {
				rpl := &Account{
					ID: "cgrates.org:1001",
					BalanceMap: map[string]Balances{
						utils.MetaMonetary: {
							&Balance{Value: 20},
						}}}
				*reply.(*Account) = *rpl
				return nil
			},
			utils.ResponderDebit: func(args, reply interface{}) error {
				rpl := &CallCost{}
				*reply.(*CallCost) = *rpl
				return nil
			},
			utils.ResponderGetCost: func(args, reply interface{}) error {
				rpl := &CallCost{}
				*reply.(*CallCost) = *rpl
				return nil
			},
			utils.EeSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := &map[string]map[string]interface{}{}
				*reply.(*map[string]map[string]interface{}) = *rpl
				return nil
			},
			utils.ThresholdSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := &[]string{}
				*reply.(*[]string) = *rpl
				return nil
			},
			utils.StatSv1ProcessEvent: func(args, reply interface{}) error {
				rpl := &[]string{}
				*reply.(*[]string) = *rpl
				return nil
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	if err := cdrS.V1ProcessEvent(arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %v,received %v", utils.OK, reply)
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
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.EeSv1ProcessEvent: func(args, reply interface{}) error {

				return utils.ErrPartiallyExecuted
			},
		},
	}
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
			Event: map[string]interface{}{
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
	if _, err := cdrs.processEvents(evs, true, true, true, true, true, true, true, true, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	removelog()
	buf2 := new(bytes.Buffer)
	setlog(buf2)
	expLog = `with ChargerS`
	if _, err := cdrs.processEvents(evs, true, false, true, true, true, true, true, true, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf2.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	removelog()
	buf3 := new(bytes.Buffer)
	setlog(buf3)
	Cache.Set(utils.CacheCDRIDs, utils.ConcatenatedKey("test1", utils.MetaDefault), "val", []string{}, true, utils.NonTransactional)
	expLog = `with CacheS`
	if _, err = cdrs.processEvents(evs, false, false, false, true, true, false, true, true, true); err == nil || err != utils.ErrExists {
		t.Error(err)
	} else if rcvLog := buf3.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf4 := new(bytes.Buffer)
	removelog()
	setlog(buf4)
	evs[0].Event[utils.AnswerTime] = "time"
	expLog = `converting event`
	if _, err = cdrs.processEvents(evs, false, false, true, true, true, false, true, true, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf4.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf5 := new(bytes.Buffer)
	removelog()
	setlog(buf5)
	evs[0].Event[utils.AnswerTime] = time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC)
	expLog = `refunding CDR`
	if _, err = cdrs.processEvents(evs, false, false, true, false, false, false, false, false, false); err != nil {
		t.Error(err)
	} else if rcvLog := buf5.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	db.db.Set(utils.CacheCDRsTBL, utils.ConcatenatedKey("test1", utils.MetaDefault, "testV1CDRsRefundOutOfSessionCost"), "val", []string{}, true, utils.NonTransactional)

	buf6 := new(bytes.Buffer)
	removelog()
	setlog(buf6)
	expLog = `refunding CDR`
	if _, err = cdrs.processEvents(evs, false, false, true, false, true, false, false, false, false); err == nil || err != utils.ErrExists {
		t.Error(err)
	} else if rcvLog := buf6.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf7 := new(bytes.Buffer)
	removelog()
	setlog(buf7)
	expLog = `exporting cdr`
	if _, err = cdrs.processEvents(evs, false, false, true, false, false, true, true, false, false); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf7.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf8 := new(bytes.Buffer)
	removelog()
	setlog(buf8)
	expLog = `processing event`
	if _, err = cdrs.processEvents(evs, false, false, true, false, false, true, false, true, false); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf8.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("expected %v,received %v", expLog, rcvLog)
	}
	buf9 := new(bytes.Buffer)
	removelog()
	setlog(buf9)
	expLog = `processing event`
	if _, err = cdrs.processEvents(evs, false, false, true, false, false, true, false, false, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	} else if rcvLog := buf9.String(); !strings.Contains(rcvLog, expLog) {
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
	attr.Cost.CGRID = utils.EmptyString
	if err := cdrS.V1StoreSessionCost(attr, reply); err == nil || err.Error() != fmt.Sprintf("%s: CGRID", utils.MandatoryInfoMissing) {
		t.Error(err)
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
	cdrS.guard = guardian.Guardian
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 0
	attr.CheckDuplicate = true
	if err = cdrS.V1StoreSessionCost(attr, reply); err != nil {
		t.Error(err)
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
	var reply string

	if err := cdrS.V2StoreSessionCost(args, &reply); err != nil {
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
	if err = cdrS.V2StoreSessionCost(args, &reply); err == nil {
		t.Error(err)
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
		APIOpts: map[string]interface{}{},
	}
	var reply string

	if err := cdrS.V1RateCDRs(arg, &reply); err == nil || err != utils.ErrNotFound {
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
		APIOpts: map[string]interface{}{},
	}
	var cdrs *[]*CDR
	if err := cdrS.V1GetCDRs(args, cdrs); err == nil {
		t.Error(utils.NewErrServerError(utils.ErrNotFound))
	}
	args.RPCCDRsFilter.SetupTimeStart = ""
	if err := cdrS.V1GetCDRs(args, cdrs); err == nil || err.Error() != fmt.Sprintf("SERVER_ERROR: %s", utils.ErrNotFound) {
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderDebit: func(args, reply interface{}) error {

				return utils.ErrAccountNotFound
			},
			utils.SchedulerSv1ExecuteActionPlans: func(args, reply interface{}) error {
				rpl := "reply"
				*reply.(*string) = rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMock
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		APIOpts: map[string]interface{}{},
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderGetCost: func(args, reply interface{}) error {

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
		APIOpts: map[string]interface{}{},
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
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundRounding: func(args, reply interface{}) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return nil
			},
		},
	}
	clientconn := make(chan rpcclient.ClientConnector, 1)
	clientconn <- ccMOck
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		APIOpts: map[string]interface{}{},
	}
	var reply string
	if err := cdrS.V2StoreSessionCost(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected %+v,received %+v", utils.OK, reply)
	}
	clientconn2 := make(chan rpcclient.ClientConnector, 1)
	clientconn2 <- &ccMock{

		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderRefundRounding: func(args, reply interface{}) error {
				rpl := &Account{}
				*reply.(*Account) = *rpl
				return utils.ErrNotFound
			},
		},
	}
	cdrS.connMgr.rpcInternal[utils.ConcatenatedKey(utils.MetaInternal, utils.RateSConnsCfg)] = clientconn2

	if err := cdrS.V2StoreSessionCost(args, &reply); err != nil {
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
		APIOpts: map[string]interface{}{},
	}
	var reply *string

	if err := cdrS.V1RateCDRs(arg, reply); err == nil {
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
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ResponderDebit: func(args, reply interface{}) error {
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
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	clienConn := make(chan rpcclient.ClientConnector, 1)
	clienConn <- &ccMock{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.ChargerSv1ProcessEvent: func(args, reply interface{}) error {
				*reply.(*[]*ChrgSProcessEventReply) = []*ChrgSProcessEventReply{
					{
						ChargerSProfile:    "Charger1",
						AttributeSProfiles: []string{"cgrates.org:ATTR_1001_SIMPLEAUTH"},
						AlteredFields:      []string{utils.MetaReqRunID, "*req.Password"},
						CGREvent: &utils.CGREvent{ // matching Charger1
							Tenant: "cgrates.org",
							ID:     "event1",
							Event: map[string]interface{}{
								utils.AccountField: "1001",
								"Password":         "CGRateS.org",
								"RunID":            utils.MetaDefault,
							},
							APIOpts: map[string]interface{}{
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
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
	}
	expcgrEv := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				"Password":         "CGRateS.org",
				"RunID":            utils.MetaDefault,
			},
			APIOpts: map[string]interface{}{
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

func TestCdrSCall123(t *testing.T) {
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
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- cdrS
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{})
	config.SetCgrConfig(cfg)
	SetConnManager(connMgr)
	var reply string
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
	if err := cdrS.Call(utils.CDRsV1StoreSessionCost, attr, &reply); err != nil {
		t.Error(err)
	}
}
