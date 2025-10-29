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

package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type clMock func(_ *context.Context, _ string, _ any, _ any) error

func (c clMock) Call(ctx *context.Context, m string, a any, r any) error {
	return c(ctx, m, a, r)
}

func TestCDRSV1ProcessCDRNoTenant(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clMock := clMock(func(_ *context.Context, _ string, args any, reply any) error {
		rply, cancast := reply.(*AttrSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		newArgs, cancast := args.(*AttrArgsProcessEvent)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		if newArgs.Tenant == utils.EmptyString {
			return fmt.Errorf("Tenant is missing")
		}
		*rply = AttrSProcessEventReply{
			AlteredFields: []string{utils.Account},
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
	cdr := &CDRWithArgDispatcher{ // no tenant, take the default
		CDR: &CDR{
			CGRID:       "Cdr1",
			OrderID:     123,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.META_RATED,
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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ *context.Context, _ string, args any, reply any) error {
		rply, cancast := reply.(*[]*ChrgSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		newArgs, cancast := args.(*utils.CGREventWithArgDispatcher)
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
				utils.CGRID:       "test1",
				utils.RunID:       utils.MetaDefault,
				utils.OriginID:    "testV1CDRsRefundOutOfSessionCost",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "testV1CDRsRefundOutOfSessionCost",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       123 * time.Minute,
			},
		},
	}
	var reply string

	if err := cdrs.V1ProcessEvent(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRSV1V1ProcessExternalCDRNoTenant(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)}
	clMock := clMock(func(_ *context.Context, _ string, args any, reply any) error {
		rply, cancast := reply.(*[]*ChrgSProcessEventReply)
		if !cancast {
			return fmt.Errorf("can't cast")
		}
		newArgs, cancast := args.(*utils.CGREventWithArgDispatcher)
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

	args := &ExternalCDRWithArgDispatcher{
		ExternalCDR: &ExternalCDR{
			ToR:         utils.VOICE,
			OriginID:    "testDspCDRsProcessExternalCDR",
			OriginHost:  "127.0.0.1",
			Source:      utils.UNIT_TEST,
			RequestType: utils.META_RATED,
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

func TestCDRSRateExportCDRS(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CdrsCfg().OnlineCDRExports = []string{"stringy"}
	cfg.CdreProfiles = map[string]*config.CdreCfg{"stringy": {}}
	cdrs := &CDRServer{
		cgrCfg:  cfg,
		connMgr: nil,
		cdrDb:   db,
		dm:      dm,
	}
	cdr := &CDR{
		ToR:         utils.VOICE,
		OriginID:    "testDspCDRsProcessExternalCDR",
		OriginHost:  "127.0.0.1",
		Source:      utils.UNIT_TEST,
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1003",
		Subject:     "1003",
		Destination: "1001",

		Usage: time.Second,
	}
	if err := cdrs.cdrDb.SetCDR(cdr, true); err != nil {
		t.Error(err)
	}
	arg := &ArgRateCDRs{
		Flags:         []string{"*export:true"},
		RPCCDRsFilter: utils.RPCCDRsFilter{},
	}
	var reply string
	if err := cdrs.V1RateCDRs(arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected ok ,received %v", reply)
	}
}

func TestCDRSStoreSessionCost22(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
	}
	cdr := &AttrCDRSStoreSMCost{
		Cost: &SMCost{
			CGRID:      "test1",
			RunID:      utils.MetaDefault,
			OriginID:   "testV1CDRsRefundOutOfSessionCost",
			CostSource: utils.MetaSessionS,
			OriginHost: "",
			Usage:      time.Duration(3 * time.Minute),
			CostDetails: &EventCost{
				CGRID:     "test1",
				RunID:     utils.MetaDefault,
				StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				Usage:     utils.DurationPointer(time.Duration(3 * time.Minute)),
				Cost:      utils.Float64Pointer(2.3),
				Charges: []*ChargingInterval{
					{
						RatingID: "c1a5ab9",
						Increments: []*ChargingIncrement{
							{
								Usage:          time.Duration(2 * time.Minute),
								Cost:           2.0,
								AccountingID:   "a012888",
								CompressFactor: 1,
							},
							{
								Usage:          time.Duration(1 * time.Second),
								Cost:           0.005,
								AccountingID:   "44d6c02",
								CompressFactor: 60,
							},
						},
						CompressFactor: 1,
					},
				},
				AccountSummary: &AccountSummary{
					Tenant: "cgrates.org",
					ID:     "testV1CDRsRefundOutOfSessionCost",
					BalanceSummaries: []*BalanceSummary{
						{
							UUID:  "random",
							Type:  utils.MONETARY,
							Value: 50,
						},
					},
					AllowNegative: false,
					Disabled:      false,
				},
				Rating: Rating{
					"c1a5ab9": &RatingUnit{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						RatesID:          "ec1a177",
						RatingFiltersID:  "43e77dc",
					},
				},
				Accounting: Accounting{
					"a012888": &BalanceCharge{
						AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
						BalanceUUID: "random",
						Units:       120.7,
					},
					"44d6c02": &BalanceCharge{
						AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
						BalanceUUID: "random",
						Units:       120.7,
					},
				},
				Rates: ChargedRates{
					"ec1a177": RateGroups{
						&Rate{
							GroupIntervalStart: time.Duration(0),
							Value:              0.01,
							RateIncrement:      time.Duration(1 * time.Minute),
							RateUnit:           time.Duration(1 * time.Second)},
					},
				},
			},
		},
	}
	var reply string
	if err := cdrS.V1StoreSessionCost(cdr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected ok,received %v", reply)
	}
	if _, err := cdrS.cdrDb.GetSMCosts(cdr.Cost.CGRID, cdr.Cost.RunID, "", cdr.Cost.OriginID); err != nil {
		t.Error(err)
	}
}

func TestCDRSV2StoreSessionCost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cc := &CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		ToR:         "*voice",
		Cost:        0,
	}
	args := &ArgsV2CDRSStoreSMCost{
		CheckDuplicate: true,
		Cost: &V2SMCost{
			CGRID:       "testRPCMethodsCdrsStoreSessionCost",
			RunID:       utils.MetaDefault,
			OriginHost:  "",
			OriginID:    "testdatagrp_grp1",
			CostSource:  "SMR",
			Usage:       1536,
			CostDetails: NewEventCostFromCallCost(cc, "testRPCMethodsCdrsStoreSessionCost", utils.MetaDefault),
		},
	}
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
		guard:  guardian.Guardian,
	}
	var reply string
	if err := cdrS.V2StoreSessionCost(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected ok,received %s", reply)
	}
}

func TestCDRSRateCDRs(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	arg := utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		TenantArg: &utils.TenantArg{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("cdrs12345"),
		},
	}

	var cdrs []*CDR
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
	}
	exp := []*CDR{
		{Account: "1001", RunID: "*default"},
	}
	cdrS.cdrDb.SetCDR(exp[0], true)
	if err := cdrS.V1GetCDRs(arg, &cdrs); err != nil {
		t.Error(err)
	}
	var cnt int64
	if err := cdrS.V1CountCDRs(&arg, &cnt); err != nil {
		t.Error(err)
	} else if cnt != 1 {
		t.Errorf("Expected 1,Received %d", cnt)
	}
}

func TestCDRSRateCDRSucces(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CdrsCfg().SMCostRetries = 0
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, reply any) error {
		if serviceMethod == utils.ResponderDebit {
			rpl := CallCost{
				Destination: "1002",
				Timespans: []*TimeSpan{
					{
						TimeStart:     time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
						TimeEnd:       time.Date(2018, 8, 24, 16, 00, 36, 0, time.UTC),
						ratingInfo:    &RatingInfo{},
						DurationIndex: 0,
						RateInterval: &RateInterval{
							Rating: &RIRate{
								Rates: RateGroups{
									&Rate{GroupIntervalStart: 0,
										Value:         100,
										RateIncrement: 1,
										RateUnit:      time.Nanosecond}}}},
					},
				},
				ToR: utils.SMS,
			}
			*reply.(*CallCost) = rpl
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): clientConn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	cdr := &CDRWithArgDispatcher{
		CDR: &CDR{CGRID: "Cdr1",
			OrderID:     101,
			ToR:         utils.VOICE,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			SetupTime:   time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			RunID:       utils.MetaDefault,
			Usage:       time.Duration(10) * time.Second,
			ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "ExtraHeader2": "ExtraVal2"},
			Cost:        1.01},
	}
	if _, err := cdrS.rateCDR(cdr); err != nil {
		t.Error(err)
	}
}

func TestV2StoreSessionCost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CacheCfg()[utils.CacheRPCResponses].Limit = 1
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		dm:     dm,
		cdrDb:  db,
		guard:  guardian.Guardian,
	}
	args := &ArgsV2CDRSStoreSMCost{
		CheckDuplicate: true,
		Cost: &V2SMCost{
			CGRID:      "testRPCMethodsCdrsStoreSessionCost",
			RunID:      utils.MetaDefault,
			OriginHost: "",
			OriginID:   "testdatagrp_grp1",
			CostSource: "SMR",
			Usage:      1536,
			CostDetails: &EventCost{
				AccountSummary: &AccountSummary{},
			},
		},
	}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID),
		&utils.CachedRPCResponse{Result: utils.OK, Error: nil},
		nil, true, utils.NonTransactional)
	var reply string
	if err := cdrS.V2StoreSessionCost(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRSThDSProcessEvent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {
		if serviceMethod == utils.ThresholdSv1ProcessEvent {
			return nil
		}
		return utils.ErrNotImplemented
	})
	connMgr := NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn,
	})
	cdrS := &CDRServer{
		cgrCfg:  cfg,
		cdrDb:   db,
		dm:      dm,
		connMgr: connMgr,
	}
	cgrEv := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CGREvent1",
			Event: map[string]any{
				utils.Account:     "1002",
				utils.Subject:     "1002",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("sup12345"),
		},
	}
	if err := cdrS.thdSProcessEvent(cgrEv); err != nil {
		t.Error(err)
	}
}

// func TestCdrSStoreSessionCost(t *testing.T) {
// 	cfg, _ := config.NewDefaultCGRConfig()
// 	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
// 	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
// 	dm := NewDataManager(db, cfg.CacheCfg(), nil)
// 	cdrS := &CDRServer{
// 		cgrCfg:  cfg,
// 		cdrDb:   db,
// 		dm:      dm,
// 		connMgr: connMgr,
// 	}

// }

func TestCRDSRefundEventCost(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMethod string, _, _ any) error {

		if serviceMethod == utils.ResponderRefundIncrements {

			return nil
		}
		return utils.ErrNotImplemented
	})
	cdrS := &CDRServer{cgrCfg: cfg, dm: dm, cdrDb: db, connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs): clientConn,
	})}
	ec := &EventCost{
		Cost: utils.Float64Pointer(10),
		Accounting: map[string]*BalanceCharge{
			"3463957": {
				Units:         0.002623,
				RatingID:      "",
				AccountID:     "cgrates.org:1001",
				ExtraChargeID: "",
			},
			"fee8a3a": {
				Units:         0.0787,
				RatingID:      "",
				AccountID:     "cgrates.org:1001",
				ExtraChargeID: "",
			},
		},
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",

			BalanceSummaries: []*BalanceSummary{
				{ID: "voice2", Type: utils.VOICE, Value: 10, Disabled: false},
			},
			AllowNegative: true,
			Disabled:      false,
		},
		Charges: []*ChargingInterval{
			{
				RatingID: "c1a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Duration(0),
						Cost:           0.1,
						AccountingID:   "3463957",
						CompressFactor: 1,
					},
					{
						Usage:          time.Duration(1 * time.Second),
						Cost:           0,
						AccountingID:   "fee8a3a",
						CompressFactor: 10,
					},
				},
				CompressFactor: 1,
			},
		}}

	if _, err := cdrS.refundEventCost(ec, "*prepaid", "*monetary"); err != nil {
		t.Error(err)
	}
}

func TestCDRSV1StoreSessionCostCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CacheCfg()[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		cdrDb:  db,
		dm:     dm,
	}
	cc := &CallCost{
		Category:    utils.CALL,
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		ToR:         "ToR",
		Cost:        10,
	}
	attr := &AttrCDRSStoreSMCost{
		Cost: &SMCost{
			CGRID:       "b783a8b",
			RunID:       utils.MetaDefault,
			OriginHost:  "",
			OriginID:    "testdatagrp_grp1",
			CostSource:  "SMR",
			Usage:       1536,
			CostDetails: NewEventCostFromCallCost(cc, "b783a8b", utils.MetaDefault),
		},
		CheckDuplicate: false,
	}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, attr.Cost.CGRID, attr.Cost.RunID), &utils.CachedRPCResponse{Result: utils.OK, Error: nil},
		nil, true, utils.NonTransactional)
	var reply string
	config.SetCgrConfig(cfg)
	if err := cdrS.V1StoreSessionCost(attr, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRsV1ProcessEventAll(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	cfg.CacheCfg()[utils.CacheRPCResponses].Limit = 1
	cfg.CdrsCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)}
	cfg.CdrsCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, serviceMeth string, _, _ any) error {
		if serviceMeth == utils.ThresholdSv1ProcessEvent {
			return nil
		}
		if serviceMeth == utils.StatSv1ProcessEvent {
			return nil
		}
		return utils.ErrNotImplemented
	})
	cdrS := &CDRServer{
		cgrCfg: cfg,
		dm:     dm,
		cdrDb:  db,
		connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS):      clientConn,
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): clientConn,
		}),
	}
	var reply string
	arg := &ArgV1ProcessEvent{
		Flags: []string{utils.MetaThresholds, utils.MetaStatS},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.RequestType: utils.META_POSTPAID,
				utils.Category:    "call",
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Minute,
			},
		},
	}
	if err := cdrS.V1ProcessEvent(arg, &reply); err != nil {
		t.Error(err)
	}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessEvent, arg.CGREvent.ID), &utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)
	config.SetCgrConfig(cfg)
	Cache.Clear(nil)
	if err := cdrS.V1ProcessEvent(arg, &reply); err == nil { //unfinished
		t.Error(err)
	}
}

func TestCDRSV2StoreCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	Cache.Clear(nil)
	cfg.CacheCfg()[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		dm:     dm,
		cdrDb:  db,
		guard:  guardian.Guardian,
	}
	cc := &CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		ToR:         "*voice",
		Cost:        0,
	}
	args := &ArgsV2CDRSStoreSMCost{
		CheckDuplicate: true,
		Cost: &V2SMCost{
			CGRID:       "CGR_ID",
			RunID:       utils.MetaDefault,
			OriginHost:  "",
			OriginID:    "testdatagrp_grp1",
			CostSource:  "SMS",
			Usage:       1536,
			CostDetails: NewEventCostFromCallCost(cc, "CGR_ID", utils.MetaDefault),
		},
	}
	config.SetCgrConfig(cfg)
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1StoreSessionCost, args.Cost.CGRID, args.Cost.RunID),
		&utils.CachedRPCResponse{Result: utils.OK, Error: nil},
		nil, true, utils.NonTransactional)
	var reply string
	if err := cdrS.V2StoreSessionCost(args, &reply); err != nil {
		t.Error(err)
	}
}

func TestRemoveThresholdProfileRpl(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	cfg.DataDbCfg().Items[utils.MetaThresholdProfiles].Replicate = true
	cfg.DataDbCfg().RplConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1)}
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, servicemethod string, _, _ any) error {
		if servicemethod == utils.ReplicatorSv1RemoveThresholdProfile {

			return nil
		}
		return utils.ErrNotImplemented
	})

	dm := NewDataManager(db, cfg.CacheCfg(), NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.ReplicatorSv1): clientConn,
	}))
	thp := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_Test",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2024, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2024, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:  -1,
		MinSleep: time.Duration(5 * time.Minute),
		Blocker:  false,
		Weight:   20.0,
	}
	config.SetCgrConfig(cfg)
	dm.SetThresholdProfile(thp, true)
	if err := dm.RemoveThresholdProfile("cgrates.org", "THD_Test", utils.NonTransactional, true); err != nil {
		t.Error(err)
	}

}

func TestCDRSV1ProcessCDRCache(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CacheCfg()[utils.CacheRPCResponses].Limit = 1
	Cache.Clear(nil)
	defer func() {
		cfg2, _ := config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg2)
	}()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	cdr := &CDRWithArgDispatcher{
		CDR: &CDR{
			Tenant:      "cgrates.org",
			OriginID:    "testDspCDRsProcessCDR",
			OriginHost:  "192.168.1.1",
			Source:      "testDspCDRsProcessCDR",
			RequestType: utils.META_RATED,
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			Usage:       time.Duration(2) * time.Minute,
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("cdrs12345"),
		},
	}
	var reply string
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.CDRsV1ProcessCDR, cdr.CGRID, cdr.RunID),
		&utils.CachedRPCResponse{Result: reply, Error: err},
		nil, true, utils.NonTransactional)
	cdrS := &CDRServer{
		cgrCfg: cfg,
		dm:     dm,
		cdrDb:  db,
		guard:  guardian.Guardian,
	}
	config.SetCgrConfig(cfg)
	if err := cdrS.V1ProcessCDR(cdr, &reply); err != nil {
		t.Error(err)
	}
}

func TestCDRSGetCostFromRater(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().RaterConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, servicemethod string, _, reply any) error {
		if servicemethod == utils.ResponderGetCost {
			return nil
		}
		return utils.ErrNotImplemented
	})
	cdrS := &CDRServer{
		cgrCfg: cfg,
		dm:     dm,
		cdrDb:  db,
		connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder): clientConn,
		}),
	}
	cdr := &CDRWithArgDispatcher{
		CDR: &CDR{
			Tenant:      "cgrates.org",
			RequestType: "default",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			ToR:         "voice",
			SetupTime:   time.Now(),
			AnswerTime:  time.Now().Add(time.Duration(2) * time.Second),
			Usage:       time.Duration(2) * time.Minute,
		},
	}
	if _, err := cdrS.getCostFromRater(cdr); err != nil {
		t.Error(err)
	}
}

func TestCDRSChrgsProcessEvent(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	cfg.CdrsCfg().ChargerSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.ChargerSv1)}
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	clientConn := make(chan birpc.ClientConnector, 1)
	clientConn <- clMock(func(ctx *context.Context, servicemethod string, _, reply any) error {

		if servicemethod == utils.ChargerSv1ProcessEvent {
			chargers := []*ChrgSProcessEventReply{
				{
					ChargerSProfile:    "Charger1",
					AttributeSProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
					AlteredFields:      []string{utils.MetaReqRunID},
					CGREvent: &utils.CGREvent{ // matching Charger1
						Tenant: "cgrates.org",
						ID:     "event1",
						Event: map[string]any{
							"Charger":        "ChargerProfile1",
							utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
							"UsageInterval":  "1s",
							utils.Weight:     "200.0",
							"RunID":          utils.MetaDefault,
						},
					},
				},
			}
			*reply.(*[]*ChrgSProcessEventReply) = chargers
			return nil
		}
		return utils.ErrNotImplemented
	})

	cdrs := &CDRServer{
		cgrCfg: cfg,
		dm:     dm,
		cdrDb:  db,
		connMgr: NewConnManager(cfg, map[string]chan birpc.ClientConnector{
			utils.ConcatenatedKey(utils.MetaInternal, utils.ChargerSv1): clientConn,
		}),
	}
	cgrEv := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID: "event1",
			Event: map[string]any{
				"Charger":        "ChargerProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "200.0",
			},
		}}
	if _, err := cdrs.chrgrSProcessEvent(cgrEv); err != nil {
		t.Error(err)
	}

}
