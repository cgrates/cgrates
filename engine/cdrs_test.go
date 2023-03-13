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
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	cfg.CdrsCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {
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
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {
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
	clMock := clMock(func(_ string, args interface{}, reply interface{}) error {
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
