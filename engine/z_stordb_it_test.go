// +build integration

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
	"path"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cfg    *config.CGRConfig
	storDB StorDB
)

// subtests to be executed for each confDIR
var sTestsStorDBit = []func(t *testing.T){
	testStorDBitFlush,
	testStorDBitIsDBEmpty,
	testStorDBitCRUDVersions,
	testStorDBitCRUDTPAccountProfiles,
	testStorDBitCRUDTPActionProfiles,
	testStorDBitCRUDTPDispatcherProfiles,
	testStorDBitCRUDTPDispatcherHosts,
	testStorDBitCRUDTPFilters,
	testStorDBitCRUDTPRateProfiles,
	testStorDBitCRUDTPRoutes,
	testStorDBitCRUDTPThresholds,
	testStorDBitCRUDTPAttributes,
	testStorDBitCRUDTPChargers,
	testStorDBitCRUDTpTimings,
	testStorDBitCRUDTpDestinations,
	testStorDBitCRUDTpRates,
	testStorDBitCRUDTpDestinationRates,
	testStorDBitCRUDTpRatingPlans,
	testStorDBitCRUDTpRatingProfiles,
	testStorDBitCRUDTpSharedGroups,
	testStorDBitCRUDTpActions,
	testStorDBitCRUDTpActionPlans,
	testStorDBitCRUDTpActionTriggers,
	testStorDBitCRUDTpAccountActions,
	testStorDBitCRUDTpResources,
	testStorDBitCRUDTpStats,
	testStorDBitCRUDCDRs,
	testStorDBitCRUDSMCosts,
	testStorDBitCRUDSMCosts2,
}

func TestStorDBit(t *testing.T) {
	//var stestName string
	switch *dbType {
	case utils.MetaInternal:
		cfg = config.NewDefaultCGRConfig()
		config.SetCgrConfig(cfg)
		storDB = NewInternalDB(nil, nil, false)
	case utils.MetaMySQL:
		if cfg, err = config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "storage", "mysql")); err != nil {
			t.Fatal(err)
		}
		if storDB, err = NewMySQLStorage(cfg.StorDbCfg().Host,
			cfg.StorDbCfg().Port, cfg.StorDbCfg().Name,
			cfg.StorDbCfg().User, cfg.StorDbCfg().Password,
			100, 10, 0); err != nil {
			t.Fatal(err)
		}
	case utils.MetaMongo:
		if cfg, err = config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "storage", "mongo")); err != nil {
			t.Fatal(err)
		}
		if storDB, err = NewMongoStorage(cfg.StorDbCfg().Host,
			cfg.StorDbCfg().Port, cfg.StorDbCfg().Name,
			cfg.StorDbCfg().User, cfg.StorDbCfg().Password,
			cfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, cfg.StorDbCfg().StringIndexedFields, 10*time.Second); err != nil {
			t.Fatal(err)
		}
	case utils.MetaPostgres:
		if cfg, err = config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "storage", "postgres")); err != nil {
			t.Fatal(err)
		}
		if storDB, err = NewPostgresStorage(cfg.StorDbCfg().Host,
			cfg.StorDbCfg().Port, cfg.StorDbCfg().Name,
			cfg.StorDbCfg().User, cfg.StorDbCfg().Password,
			utils.IfaceAsString(cfg.StorDbCfg().Opts[utils.SSLModeCfg]),
			100, 10, 0); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsStorDBit {
		stestFullName := runtime.FuncForPC(reflect.ValueOf(stest).Pointer()).Name()
		split := strings.Split(stestFullName, ".")
		stestName := split[len(split)-1]
		// Fixme: Implement mongo needed versions methods
		if stestName != "testStorDBitCRUDVersions" {
			stestName := split[len(split)-1]
			t.Run(stestName, stest)
		}
	}
}

func testStorDBitIsDBEmpty(t *testing.T) {
	x := storDB.GetStorageType()
	switch x {
	case utils.MONGO:
		test, err := storDB.IsDBEmpty()
		if err != nil {
			t.Error(err)
		} else if test != true {
			t.Errorf("Expecting: true got :%+v", test)
		}
	case utils.POSTGRES, utils.MYSQL:
		test, err := storDB.IsDBEmpty()
		if err != nil {
			t.Error(err)
		} else if test != false {
			t.Errorf("Expecting: false got :%+v", test)
		}
	}
}

func testStorDBitCRUDTPAccountProfiles(t *testing.T) {
	//READ
	if _, err := storDB.GetTPAccountProfiles("sub_ID1", utils.EmptyString, "TEST_ID1"); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	var actPrf = []*utils.TPAccountProfile{
		&utils.TPAccountProfile{
			TPid:   testTPID,
			Tenant: "cgrates.org",
			ID:     "1001",
			Weight: 20,
			Balances: []*utils.TPAccountBalance{
				&utils.TPAccountBalance{
					ID:        "MonetaryBalance",
					FilterIDs: []string{},
					Weight:    10,
					Type:      utils.MONETARY,
					CostIncrement: []*utils.TPBalanceCostIncrement{
						&utils.TPBalanceCostIncrement{
							FilterIDs:    []string{"fltr1", "fltr2"},
							Increment:    utils.Float64Pointer(1.3),
							FixedFee:     utils.Float64Pointer(2.3),
							RecurrentFee: utils.Float64Pointer(3.3),
						},
					},
					CostAttributes: []string{"attr1", "attr2"},
					UnitFactors: []*utils.TPBalanceUnitFactor{
						&utils.TPBalanceUnitFactor{
							FilterIDs: []string{"fltr1", "fltr2"},
							Factor:    100,
						},
						&utils.TPBalanceUnitFactor{
							FilterIDs: []string{"fltr3"},
							Factor:    200,
						},
					},
					Units: 14,
				},
			},
			ThresholdIDs: []string{utils.META_NONE},
		},
	}
	if err := storDB.SetTPAccountProfiles(actPrf); err != nil {
		t.Error(err)
	}

	//READ
	if rcv, err := storDB.GetTPAccountProfiles(actPrf[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], actPrf[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n", utils.ToJSON(actPrf[0]), utils.ToJSON(rcv[0]))
	}

	//UPDATE AND READ
	actPrf[0].FilterIDs = []string{"*string:~*req.Account:1007"}
	if err := storDB.SetTPAccountProfiles(actPrf); err != nil {
		t.Error(err)
	} else if rcv, err := storDB.GetTPAccountProfiles(actPrf[0].TPid,
		utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], actPrf[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n", utils.ToJSON(actPrf[0]), utils.ToJSON(rcv[0]))
	}

	//REMOVE AND READ
	if err := storDB.RemTpData(utils.EmptyString, actPrf[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPActionProfiles(actPrf[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPActionProfiles(t *testing.T) {
	//READ
	if _, err := storDB.GetTPActionProfiles("sub_ID1", utils.EmptyString, "TEST_ID1"); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	var actPrf = []*utils.TPActionProfile{
		{
			Tenant:    "cgrates.org",
			TPid:      "TEST_ID1",
			ID:        "sub_id1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weight:    20,
			Schedule:  utils.ASAP,
			Actions: []*utils.TPAPAction{
				{
					ID:        "TOPUP",
					FilterIDs: []string{},
					Type:      "*topup",
					Path:      "~*balance.TestBalance.Value",
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			TPid:      "TEST_ID1",
			ID:        "sub_id2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			Weight:    10,
			Schedule:  utils.ASAP,
			Actions: []*utils.TPAPAction{
				{
					ID:        "TOPUP",
					FilterIDs: []string{},
					Type:      "*topup",
					Path:      "~*balance.TestBalance.Value",
				},
			},
		},
	}
	if err := storDB.SetTPActionProfiles(actPrf); err != nil {
		t.Error("Unable to set TPActionProfile")
	}

	//READ
	if rcv, err := storDB.GetTPActionProfiles(actPrf[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], actPrf[0]) || reflect.DeepEqual(rcv[1], actPrf[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToJSON(actPrf[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//UPDATE AND READ
	actPrf[0].FilterIDs = []string{"*string:~*req.Account:1007"}
	actPrf[1].Weight = 20
	if err := storDB.SetTPActionProfiles(actPrf); err != nil {
		t.Error(err)
	} else if rcv, err := storDB.GetTPActionProfiles(actPrf[0].TPid,
		utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], actPrf[0]) || reflect.DeepEqual(rcv[1], actPrf[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToJSON(actPrf[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//REMOVE AND READ
	if err := storDB.RemTpData(utils.EmptyString, actPrf[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPActionProfiles(actPrf[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPDispatcherProfiles(t *testing.T) {
	//READ
	if _, err := storDB.GetTPDispatcherProfiles("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	var dsp = []*utils.TPDispatcherProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Strategy: utils.MetaFirst,
			Weight:   10,
		},
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Dsp2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-08-15T14:00:00Z",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	if err := storDB.SetTPDispatcherProfiles(dsp); err != nil {
		t.Errorf("Unable to set TPDispatcherProfile")
	}

	//READ
	if rcv, err := storDB.GetTPDispatcherProfiles(dsp[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(dsp[0], rcv[0]) ||
		reflect.DeepEqual(dsp[0], rcv[1])) {
		t.Errorf("Expected: \n%+v\n, received: \n%+v\n || \n%+v",
			utils.ToJSON(dsp[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//UPDATE and WRITE
	dsp[0].FilterIDs = []string{"*prefix:~*Account:1005"}
	dsp[1].FilterIDs = []string{"*prefix:~*Account:1005"}
	if err := storDB.SetTPDispatcherProfiles(dsp); err != nil {
		t.Errorf("Unable to set TPDispatcherProfile")
	}

	//READ
	if rcv, err := storDB.GetTPDispatcherProfiles(dsp[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(dsp[0], rcv[0]) ||
		reflect.DeepEqual(dsp[0], rcv[1])) {
		t.Errorf("Expected: \n%+v\n, received: \n%+v\n || \n%+v",
			utils.ToJSON(dsp[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, dsp[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPDispatcherProfiles(dsp[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPDispatcherHosts(t *testing.T) {
	//READ
	if _, err := storDB.GetTPDispatcherHosts("TP_ID", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpDispatcherHosts := []*utils.TPDispatcherHost{
		{
			TPid:   "TP_ID",
			Tenant: "cgrates.org",
			ID:     "ALL1",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "127.0.0.1:2012",
				Transport: utils.MetaJSON,
				TLS:       true,
			},
		},
		{
			TPid:   "TP_ID",
			Tenant: "cgrates.org",
			ID:     "ALL2",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "127.0.0.1:2014",
				Transport: utils.MetaJSON,
				TLS:       false,
			},
		},
	}
	if err := storDB.SetTPDispatcherHosts(tpDispatcherHosts); err != nil {
		t.Error("Unable to set TPDispatcherHosts")
	}

	//READ
	if rcv, err := storDB.GetTPDispatcherHosts(tpDispatcherHosts[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(tpDispatcherHosts[0], rcv[0]) ||
		reflect.DeepEqual(tpDispatcherHosts[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpDispatcherHosts[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//UPDATE
	tpDispatcherHosts[0].Conn.Address = "127.0.0.1:2855"
	tpDispatcherHosts[1].Conn.Address = "127.0.0.1:2855"
	if err := storDB.SetTPDispatcherHosts(tpDispatcherHosts); err != nil {
		t.Error("Unable to set TPDispatcherHosts")
	}

	//READ AFTER UPDATE
	if rcv, err := storDB.GetTPDispatcherHosts(tpDispatcherHosts[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(tpDispatcherHosts[0], rcv[0]) ||
		reflect.DeepEqual(tpDispatcherHosts[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpDispatcherHosts[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpDispatcherHosts[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPDispatcherHosts(tpDispatcherHosts[0].TPid,
		utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPFilters(t *testing.T) {
	//READ
	if _, err := storDB.GetTPFilters("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpFilters := []*utils.TPFilterProfile{
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Filter1",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaString,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
		},
		{
			TPid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "Filter2",
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Destination",
					Values:  []string{"10"},
				},
			},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2015-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
		},
	}
	if err := storDB.SetTPFilters(tpFilters); err != nil {
		t.Errorf("Unable to set TPFilters")
	}

	//READ
	if rcv, err := storDB.GetTPFilters(tpFilters[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(tpFilters[0], rcv[0]) ||
		reflect.DeepEqual(tpFilters[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpFilters[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//UPDATE and WRITE
	tpFilters[0].Filters[0].Element = "Account"
	tpFilters[1].Filters[0].Element = "Account"
	if err := storDB.SetTPFilters(tpFilters); err != nil {
		t.Errorf("Unable to set TPFilters")
	}

	//READ
	if rcv, err := storDB.GetTPFilters(tpFilters[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(tpFilters[0], rcv[0]) ||
		reflect.DeepEqual(tpFilters[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpFilters[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpFilters[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPFilters(tpFilters[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPRateProfiles(t *testing.T) {
	//READ
	if _, err := storDB.GetTPRateProfiles("ID_RP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpr := []*utils.TPRateProfile{
		{
			TPid:             "id_RP1",
			Tenant:           "cgrates.org",
			ID:               "RP1",
			FilterIDs:        []string{"*string:~*req.Subject:1001"},
			Weight:           0,
			RoundingMethod:   "*up",
			RoundingDecimals: 4,
			MinCost:          0.1,
			MaxCost:          0.6,
			MaxCostStrategy:  "*free",
			Rates: map[string]*utils.TPRate{
				"FIRST_GI": {
					ID:        "FIRST_GI",
					FilterIDs: []string{"*gi:~*req.Usage:0"},
					Weight:    0,
					IntervalRates: []*utils.TPIntervalRate{
						{
							RecurrentFee: 0.12,
							Unit:         "1m",
							Increment:    "1m",
						},
					},
					Blocker: false,
				},
			},
		},
	}
	if err := storDB.SetTPRateProfiles(tpr); err != nil {
		t.Error(err)
	}

	//READ
	if rcv, err := storDB.GetTPRateProfiles(tpr[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv[0], tpr[0]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(tpr[0]), utils.ToJSON(rcv[0]))
	}

	//UPDATE and WRITE
	tpr[0].MaxCost = 2.0
	tpr[0].MinCost = 0.2
	if err := storDB.SetTPRateProfiles(tpr); err != nil {
		t.Error(err)
	}

	//READ
	if rcv, err := storDB.GetTPRateProfiles(tpr[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv[0], tpr[0]) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(tpr[0]), utils.ToJSON(rcv[0]))
	}

	//REMOVE AND READ
	if err := storDB.RemTpData(utils.EmptyString, tpr[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPRateProfiles(tpr[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPRoutes(t *testing.T) {
	//READ
	if _, err := storDB.GetTPRoutes("TP1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpRoutes := []*utils.TPRouteProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_1",
			FilterIDs: []string{"*string:~*req.Accout:1007"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{},
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RatingPlanIDs:   []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_2",
			FilterIDs: []string{"*string:~*req.Destination:100"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2015-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			Sorting:           "*lowest_cost",
			SortingParameters: []string{},
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RatingPlanIDs:   []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam2",
				},
			},
			Weight: 10,
		},
	}
	if err := storDB.SetTPRoutes(tpRoutes); err != nil {
		t.Errorf("Unable to set TPRoutes")
	}

	//READ
	if rcv, err := storDB.GetTPRoutes(tpRoutes[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], tpRoutes[0]) ||
		reflect.DeepEqual(tpRoutes[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpRoutes[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//UPDATE
	tpRoutes[0].Sorting = "*higher_cost"
	tpRoutes[1].Sorting = "*higher_cost"
	if err := storDB.SetTPRoutes(tpRoutes); err != nil {
		t.Errorf("Unable to set TPRoutes")
	}

	//READ
	if rcv, err := storDB.GetTPRoutes(tpRoutes[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], tpRoutes[0]) ||
		reflect.DeepEqual(tpRoutes[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpRoutes[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpRoutes[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPRoutes(tpRoutes[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPThresholds(t *testing.T) {
	//READ
	if _, err := storDB.GetTPThresholds("TH1", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpThresholds := []*utils.TPThresholdProfile{
		{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Threshold1",
			FilterIDs: []string{"*string:~*req.Account:1002", "*string:~*req.DryRun:*default"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			MaxHits:   -1,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"Thresh1"},
			Async:     true,
		},
		{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Threshold2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2015-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			MaxHits:   -1,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    20,
			ActionIDs: []string{"Thresh1"},
			Async:     true,
		},
	}
	if err := storDB.SetTPThresholds(tpThresholds); err != nil {
		t.Errorf("Unable to set TPThresholds")
	}

	//READ
	if rcv, err := storDB.GetTPThresholds(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(tpThresholds[0], rcv[0]) ||
		reflect.DeepEqual(tpThresholds[0], rcv[1])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpThresholds[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))

	}

	//UPDATE
	tpThresholds[0].FilterIDs = []string{"*string:~*req.Destination:101"}
	tpThresholds[1].FilterIDs = []string{"*string:~*req.Destination:101"}
	if err := storDB.SetTPThresholds(tpThresholds); err != nil {
		t.Errorf("Unable to set TPThresholds")
	}

	//READ
	if rcv, err := storDB.GetTPThresholds(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpThresholds[0], rcv[0]) &&
		!reflect.DeepEqual(tpThresholds[0], rcv[1]) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpThresholds[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))

	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpThresholds[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPRoutes(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPAttributes(t *testing.T) {
	//READ
	if _, err := storDB.GetTPAttributes("TP_ID", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpAProfile := []*utils.TPAttributeProfile{
		{
			TPid:   "TP_ID",
			Tenant: "cgrates.org",
			ID:     "APROFILE_ID1",
			Attributes: []*utils.TPAttribute{
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField + utils.InInFieldSep,
					Value:     "102",
					FilterIDs: []string{"*string:~*req.Account:102"},
				},
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField + utils.InInFieldSep,
					Value:     "101",
					FilterIDs: []string{"*string:~*req.Account:101"},
				},
			},
		},
		{
			TPid:   "TP_ID",
			Tenant: "cgrates.org",
			ID:     "APROFILE_ID2",
			Attributes: []*utils.TPAttribute{
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.Destination + utils.InInFieldSep,
					Value:     "11",
					FilterIDs: []string{"*string:~*req.Destination:11"},
				},
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.Destination + utils.InInFieldSep,
					Value:     "11",
					FilterIDs: []string{"*string:~*req.Destination:10"},
				},
			},
		},
	}
	if err := storDB.SetTPAttributes(tpAProfile); err != nil {
		t.Errorf("Unable to set TPActionProfile")
	}

	//READ
	if rcv, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], tpAProfile[0]) ||
		reflect.DeepEqual(rcv[1], tpAProfile[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpAProfile[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//UPDATE
	tpAProfile[0].Attributes[0].Value = "107"
	tpAProfile[1].Attributes[0].Value = "107"
	if err := storDB.SetTPAttributes(tpAProfile); err != nil {
		t.Error(err)
	}

	//READ
	if rcv, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], tpAProfile[0]) ||
		reflect.DeepEqual(rcv[1], tpAProfile[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpAProfile[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpAProfile[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTPChargers(t *testing.T) {
	//READ
	if _, err := storDB.GetTPChargers("TP_ID", utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	tpChargers := []*utils.TPChargerProfile{
		{
			TPid:      "TP_ID",
			Tenant:    "cgrates.org",
			ID:        "Chrgs1",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			RunID:  utils.MetaDefault,
			Weight: 20,
		},
		{
			TPid:      "TP_id",
			Tenant:    "cgrates.org",
			ID:        "Chrgs2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "",
			},
			RunID:  utils.MetaDefault,
			Weight: 10,
		},
	}
	if err := storDB.SetTPChargers(tpChargers); err != nil {
		t.Error("Unable to set tpChargerProfile")
	}

	//READ
	if rcv, err := storDB.GetTPChargers(tpChargers[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], tpChargers[0]) ||
		reflect.DeepEqual(rcv[1], tpChargers[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpChargers[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//UPDATE
	tpChargers[0].FilterIDs = []string{"*string:~*req.Account:1001"}
	tpChargers[1].FilterIDs = []string{"*string:~*req.Account:1001"}
	if err := storDB.SetTPChargers(tpChargers); err != nil {
		t.Error("Unable to set tpChargerProfile")
	}

	//READ
	if rcv, err := storDB.GetTPChargers(tpChargers[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !(reflect.DeepEqual(rcv[0], tpChargers[0]) ||
		reflect.DeepEqual(rcv[1], tpChargers[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
			utils.ToIJSON(tpChargers[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpChargers[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPChargers(tpChargers[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpTimings(t *testing.T) {
	// READ
	if _, err := storDB.GetTPTimings("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.ApierTPTiming{
		{
			TPid:      "testTPid",
			ID:        "testTag1",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "01:00:00",
		},
		{
			TPid:      "testTPid",
			ID:        "testTag2",
			Years:     "*any",
			Months:    "*any",
			MonthDays: "*any",
			WeekDays:  "1;2;3;4;5",
			Time:      "01:00:00",
		},
	}
	if err := storDB.SetTPTimings(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPTimings("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Time = "02:00:00"
	snd[1].Time = "02:00:00"
	if err := storDB.SetTPTimings(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPTimings("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPTimings("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpDestinations(t *testing.T) {
	// READ
	if _, err := storDB.GetTPDestinations("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	snd := []*utils.TPDestination{
		{
			TPid:     "testTPid",
			ID:       "testTag1",
			Prefixes: []string{`0256`, `0257`, `0723`, `+49`},
		},
		{
			TPid:     "testTPid",
			ID:       "testTag2",
			Prefixes: []string{`0256`, `0257`, `0723`, `+49`},
		},
	}
	if err := storDB.SetTPDestinations(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinations("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		prfs := make(map[string]bool)
		for _, prf := range snd[0].Prefixes {
			prfs[prf] = true
		}
		pfrOk := true
		for i := range rcv[0].Prefixes {
			found1, _ := prfs[rcv[0].Prefixes[i]]
			found2, _ := prfs[rcv[1].Prefixes[i]]
			if !found1 && !found2 {
				pfrOk = false
			}
		}
		if pfrOk {
			rcv[0].Prefixes = snd[0].Prefixes
			rcv[1].Prefixes = snd[0].Prefixes
		}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Prefixes = []string{`9999`, `0257`, `0723`, `+49`}
	snd[1].Prefixes = []string{`9999`, `0257`, `0723`, `+49`}
	if err := storDB.SetTPDestinations(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinations("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		prfs := make(map[string]bool)
		for _, prf := range snd[0].Prefixes {
			prfs[prf] = true
		}
		pfrOk := true
		for i := range rcv[0].Prefixes {
			found1, _ := prfs[rcv[0].Prefixes[i]]
			found2, _ := prfs[rcv[1].Prefixes[i]]
			if !found1 && !found2 {
				pfrOk = false
			}
		}
		if pfrOk {
			rcv[0].Prefixes = snd[0].Prefixes
			rcv[1].Prefixes = snd[0].Prefixes
		}
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPDestinations("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRates(t *testing.T) {
	// READ
	if _, err := storDB.GetTPRates("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPRateRALs{
		{
			TPid: "testTPid",
			ID:   "1",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.0,
					Rate:               0.0,
					RateUnit:           "60s",
					RateIncrement:      "60s",
					GroupIntervalStart: "0s",
				},
				{
					ConnectFee:         0.0,
					Rate:               0.0,
					RateUnit:           "60s",
					RateIncrement:      "60s",
					GroupIntervalStart: "1s",
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.0,
					Rate:               0.0,
					RateUnit:           "60s",
					RateIncrement:      "60s",
					GroupIntervalStart: "0s",
				},
			},
		},
	}
	snd[0].RateSlots[0].SetDurations()
	snd[0].RateSlots[1].SetDurations()
	snd[1].RateSlots[0].SetDurations()
	if err := storDB.SetTPRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRates("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].RateSlots[0].GroupIntervalStart = "3s"
	snd[1].RateSlots[0].GroupIntervalStart = "3s"
	snd[0].RateSlots[0].SetDurations()
	snd[1].RateSlots[0].SetDurations()
	if err := storDB.SetTPRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRates("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPRates("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpDestinationRates(t *testing.T) {
	// READ
	if _, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPDestinationRate{
		{
			TPid: "testTPid",
			ID:   "1",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "GERMANY",
					RateId:           "RT_1CENT",
					RoundingMethod:   "*up",
					RoundingDecimals: 0,
					MaxCost:          0.0,
					MaxCostStrategy:  "",
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "GERMANY",
					RateId:           "RT_1CENT",
					RoundingMethod:   "*up",
					RoundingDecimals: 0,
					MaxCost:          0.0,
					MaxCostStrategy:  "",
				},
			},
		},
	}

	if err := storDB.SetTPDestinationRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].DestinationRates[0].MaxCostStrategy = "test"
	snd[1].DestinationRates[0].MaxCostStrategy = "test"
	if err := storDB.SetTPDestinationRates(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPDestinationRates("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRatingPlans(t *testing.T) {
	// READ
	if _, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPRatingPlan{
		{
			TPid: "testTPid",
			ID:   "1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "1",
					TimingId:           "ALWAYS",
					Weight:             0.0,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "2",
					TimingId:           "ALWAYS",
					Weight:             2,
				},
			},
		},
	}
	if err := storDB.SetTPRatingPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].RatingPlanBindings[0].TimingId = "test"
	snd[1].RatingPlanBindings[0].TimingId = "test"
	if err := storDB.SetTPRatingPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPRatingPlans("testTPid", "", nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpRatingProfiles(t *testing.T) {
	// READ
	var filter = utils.TPRatingProfile{
		TPid: "testTPid",
	}
	if _, err := storDB.GetTPRatingProfiles(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPRatingProfile{
		{
			TPid:     "testTPid",
			LoadId:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "test",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "test",
					FallbackSubjects: "",
				},
			},
		},
		{
			TPid:     "testTPid",
			LoadId:   "TEST_LOADID2",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "test",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2014-07-29T15:00:00Z",
					RatingPlanId:     "test",
					FallbackSubjects: "",
				},
			},
		},
	}
	if err := storDB.SetTPRatingProfiles(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingProfiles(&filter); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) ||
			reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].RatingPlanActivations = append(snd[0].RatingPlanActivations,
		&utils.TPRatingActivation{
			ActivationTime:   "2019-02-11T15:00:00Z",
			RatingPlanId:     "test",
			FallbackSubjects: "",
		})
	snd[1].RatingPlanActivations = append(snd[1].RatingPlanActivations,
		&utils.TPRatingActivation{
			ActivationTime:   "2019-02-11T15:00:00Z",
			RatingPlanId:     "test",
			FallbackSubjects: "",
		})
	if err := storDB.SetTPRatingProfiles(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPRatingProfiles(&filter); err != nil {
		t.Error(err)
	} else {
		if len(snd) != len(rcv) ||
			len(snd[0].RatingPlanActivations) != len(rcv[0].RatingPlanActivations) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPRatingProfiles(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpSharedGroups(t *testing.T) {
	// READ
	if _, err := storDB.GetTPSharedGroups("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPSharedGroups{
		{
			TPid: "testTPid",
			ID:   "1",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "test",
					Strategy:      "*lowest_cost",
					RatingSubject: "test",
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "test",
					Strategy:      "*lowest_cost",
					RatingSubject: "test",
				},
			},
		},
	}
	if err := storDB.SetTPSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPSharedGroups("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].SharedGroups[0].Strategy = "test"
	snd[1].SharedGroups[0].Strategy = "test"
	if err := storDB.SetTPSharedGroups(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPSharedGroups("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPSharedGroups("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActions(t *testing.T) {
	// READ
	if _, err := storDB.GetTPActions("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPActions{
		{
			TPid: "testTPid",
			ID:   "1",
			Actions: []*utils.TPAction{
				{
					Identifier:      "",
					BalanceId:       "",
					BalanceUuid:     "",
					BalanceType:     "*monetary",
					Units:           "10",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "",
					DestinationIds:  "DST_ON_NET",
					RatingSubject:   "",
					Categories:      "",
					SharedGroups:    "",
					BalanceWeight:   "",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          11.0,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			Actions: []*utils.TPAction{
				{
					Identifier:      "",
					BalanceId:       "",
					BalanceUuid:     "",
					BalanceType:     "*monetary",
					Units:           "10",
					ExpiryTime:      "*unlimited",
					Filter:          "",
					TimingTags:      "",
					DestinationIds:  "DST_ON_NET",
					RatingSubject:   "",
					Categories:      "",
					SharedGroups:    "",
					BalanceWeight:   "",
					ExtraParameters: "",
					BalanceBlocker:  "false",
					BalanceDisabled: "false",
					Weight:          11.0,
				},
			},
		},
	}
	if err := storDB.SetTPActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActions("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Actions[0].Weight = 12.1
	snd[1].Actions[0].Weight = 12.1
	if err := storDB.SetTPActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActions("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPActions("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActionPlans(t *testing.T) {
	// READ
	if _, err := storDB.GetTPActionPlans("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPActionPlan{
		{
			TPid: "testTPid",
			ID:   "1",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "1",
					TimingId:  "1",
					Weight:    1,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "1",
					TimingId:  "1",
					Weight:    1,
				},
			},
		},
	}
	if err := storDB.SetTPActionPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionPlans("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].ActionPlan[0].TimingId = "test"
	snd[1].ActionPlan[0].TimingId = "test"
	if err := storDB.SetTPActionPlans(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionPlans("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPActionPlans("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpActionTriggers(t *testing.T) {
	// READ
	if _, err := storDB.GetTPActionTriggers("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPActionTriggers{
		{
			TPid: "testTPid",
			ID:   "1",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:                    "1",
					UniqueID:              "",
					ThresholdType:         "1",
					ThresholdValue:        0,
					Recurrent:             true,
					MinSleep:              "",
					ExpirationDate:        "2014-07-29T15:00:00Z",
					ActivationDate:        "2014-07-29T15:00:00Z",
					BalanceId:             "test",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "call",
					BalanceWeight:         "0.0",
					BalanceExpirationDate: "2014-07-29T15:00:00Z",
					BalanceTimingTags:     "T1",
					BalanceRatingSubject:  "test",
					BalanceCategories:     "",
					BalanceSharedGroups:   "SHARED_1",
					BalanceBlocker:        "false",
					BalanceDisabled:       "false",
					ActionsId:             "test",
					Weight:                1.0,
				},
			},
		},
		{
			TPid: "testTPid",
			ID:   "2",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:                    "2",
					UniqueID:              "",
					ThresholdType:         "1",
					ThresholdValue:        0,
					Recurrent:             true,
					MinSleep:              "",
					ExpirationDate:        "2014-07-29T15:00:00Z",
					ActivationDate:        "2014-07-29T15:00:00Z",
					BalanceId:             "test",
					BalanceType:           "*monetary",
					BalanceDestinationIds: "call",
					BalanceWeight:         "0.0",
					BalanceExpirationDate: "2014-07-29T15:00:00Z",
					BalanceTimingTags:     "T1",
					BalanceRatingSubject:  "test",
					BalanceCategories:     "",
					BalanceSharedGroups:   "SHARED_1",
					BalanceBlocker:        "false",
					BalanceDisabled:       "false",
					ActionsId:             "test",
					Weight:                1.0,
				},
			},
		},
	}
	if err := storDB.SetTPActionTriggers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionTriggers("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].ActionTriggers[0].ActionsId = "test2"
	snd[1].ActionTriggers[0].ActionsId = "test2"
	if err := storDB.SetTPActionTriggers(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPActionTriggers("testTPid", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPActionTriggers("testTPid", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpAccountActions(t *testing.T) {
	// READ
	var filter = utils.TPAccountActions{
		TPid: "testTPid",
	}
	if _, err := storDB.GetTPAccountActions(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*utils.TPAccountActions{
		{
			TPid:             "testTPid",
			LoadId:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Account:          "1001",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         true,
		},
		{
			TPid:             "testTPid",
			LoadId:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Account:          "1002",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
			AllowNegative:    true,
			Disabled:         true,
		},
	}
	if err := storDB.SetTPAccountActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPAccountActions(&filter); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// UPDATE
	snd[0].Disabled = false
	snd[1].Disabled = false
	if err := storDB.SetTPAccountActions(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPAccountActions(&filter); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0], rcv[0]) || reflect.DeepEqual(snd[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v", utils.ToIJSON(snd[0]), utils.ToIJSON(rcv[0]), utils.ToIJSON(rcv[1]))
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPAccountActions(&filter); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpResources(t *testing.T) {
	// READ
	if _, err := storDB.GetTPResources("testTPid", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	var snd = []*utils.TPResourceProfile{
		{
			TPid:         "testTPid",
			ID:           "testTag1",
			Weight:       0.0,
			Limit:        "test",
			ThresholdIDs: []string{"1x", "2x"},
			FilterIDs:    []string{"FILTR_RES_1"},
			Blocker:      true,
			Stored:       true,
		},
		{
			TPid:               "testTPid",
			ID:                 "testTag2",
			ActivationInterval: &utils.TPActivationInterval{ActivationTime: "test"},
			Weight:             0.0,
			Limit:              "test",
			ThresholdIDs:       []string{"1x", "2x"},
			FilterIDs:          []string{"FLTR_RES_2"},
			Blocker:            true,
			Stored:             false,
		},
	}
	if err := storDB.SetTPResources(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPResources("testTPid", "", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].TPid, rcv[0].TPid) || reflect.DeepEqual(snd[0].TPid, rcv[1].TPid)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].ID, rcv[0].ID) || reflect.DeepEqual(snd[0].ID, rcv[1].ID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ID, rcv[0].ID, rcv[1].ID)
		}
		if !(reflect.DeepEqual(snd[0].ActivationInterval, rcv[0].ActivationInterval) || reflect.DeepEqual(snd[0].ActivationInterval, rcv[1].ActivationInterval)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].Weight, rcv[0].Weight) || reflect.DeepEqual(snd[0].Weight, rcv[1].Weight)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Weight, rcv[0].Weight, rcv[1].Weight)
		}
		if !(reflect.DeepEqual(snd[0].Limit, rcv[0].Limit) || reflect.DeepEqual(snd[0].Limit, rcv[1].Limit)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Limit, rcv[0].Limit, rcv[1].Limit)
		}
		sort.Strings(rcv[0].ThresholdIDs)
		sort.Strings(rcv[1].ThresholdIDs)
		if !(reflect.DeepEqual(snd[0].ThresholdIDs, rcv[0].ThresholdIDs) || reflect.DeepEqual(snd[0].ThresholdIDs, rcv[1].ThresholdIDs)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ThresholdIDs, rcv[0].ThresholdIDs, rcv[1].ThresholdIDs)
		}
	}
	// UPDATE
	snd[0].Weight = 2.1
	snd[1].Weight = 2.1
	if err := storDB.SetTPResources(snd); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPResources("testTPid", "", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].TPid, rcv[0].TPid) || reflect.DeepEqual(snd[0].TPid, rcv[1].TPid)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].ID, rcv[0].ID) || reflect.DeepEqual(snd[0].ID, rcv[1].ID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ID, rcv[0].ID, rcv[1].ID)
		}
		if !(reflect.DeepEqual(snd[0].ActivationInterval, rcv[0].ActivationInterval) ||
			reflect.DeepEqual(snd[0].ActivationInterval, rcv[1].ActivationInterval)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].TPid, rcv[0].TPid, rcv[1].TPid)
		}
		if !(reflect.DeepEqual(snd[0].Weight, rcv[0].Weight) || reflect.DeepEqual(snd[0].Weight, rcv[1].Weight)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Weight, rcv[0].Weight, rcv[1].Weight)
		}
		if !(reflect.DeepEqual(snd[0].Limit, rcv[0].Limit) || reflect.DeepEqual(snd[0].Limit, rcv[1].Limit)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Limit, rcv[0].Limit, rcv[1].Limit)
		}
		sort.Strings(rcv[0].ThresholdIDs)
		sort.Strings(rcv[1].ThresholdIDs)
		if !(reflect.DeepEqual(snd[0].ThresholdIDs, rcv[0].ThresholdIDs) || reflect.DeepEqual(snd[0].ThresholdIDs, rcv[1].ThresholdIDs)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ThresholdIDs, rcv[0].ThresholdIDs, rcv[1].ThresholdIDs)
		}
	}
	// REMOVE
	if err := storDB.RemTpData("", "testTPid", nil); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetTPResources("testTPid", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDTpStats(t *testing.T) {
	// READ
	if _, err := storDB.GetTPStats("TEST_TPID", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	//WRITE
	eTPs := []*utils.TPStatProfile{
		{
			TPid:      "TEST_TPID",
			Tenant:    "Test",
			ID:        "Stats1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*asr",
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20.0,
			Stored:       true,
			MinItems:     1,
		},
	}

	if err := storDB.SetTPStats(eTPs); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetTPStats("TEST_TPID", "", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs[0], rcv[0]) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(eTPs[0]), utils.ToJSON(rcv[0]))
	}

	// UPDATE
	eTPs[0].Metrics = []*utils.MetricWithFilters{
		{
			MetricID: "*asr",
		},
		{
			MetricID: utils.MetaACD,
		},
	}
	if err := storDB.SetTPStats(eTPs); err != nil {
		t.Error(err)
	}
	eTPsReverse := []*utils.TPStatProfile{
		{
			TPid:      "TEST_TPID",
			Tenant:    "Test",
			ID:        "Stats1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: "*asr",
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20.0,
			Stored:       true,
			MinItems:     1,
		},
	}
	// READ
	if rcv, err := storDB.GetTPStats("TEST_TPID", "", ""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs[0], rcv[0]) && !reflect.DeepEqual(eTPsReverse[0], rcv[0]) {
		t.Errorf("Expecting: %+v,\n received: %+v || reveived : %+v", utils.ToJSON(eTPs[0]), utils.ToJSON(rcv[0]), utils.ToJSON(eTPsReverse[0]))
	}

	// REMOVE
	if err := storDB.RemTpData(utils.TBLTPStats, "TEST_TPID", nil); err != nil {
		t.Error(err)
	}
	// READ
	if ids, err := storDB.GetTPStats("TEST_TPID", "", ""); err != utils.ErrNotFound {
		t.Error(err)
		t.Error(utils.ToJSON(ids))
	}
}

func testStorDBitCRUDCDRs(t *testing.T) {
	// READ
	var filter = utils.CDRsFilter{}
	if _, _, err := storDB.GetCDRs(&filter, false); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*CDR{
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OrderID:     0,
			OriginHost:  "host1",
			OriginID:    "1",
			Usage:       1000000000,
			CostDetails: NewBareEventCost(),
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OrderID:     0,
			OriginHost:  "host2",
			OriginID:    "2",
			Usage:       1000000000,
			CostDetails: NewBareEventCost(),
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
	}
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, false); err != nil {
			t.Error(err)
		}
	}
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, false); err == nil || err != utils.ErrExists {
			t.Error(err) // for mongo will fail because of indexes
		}
	}
	// READ
	if rcv, _, err := storDB.GetCDRs(&filter, false); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].CGRID, rcv[0].CGRID) || reflect.DeepEqual(snd[0].CGRID, rcv[1].CGRID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CGRID, rcv[0].CGRID, rcv[1].CGRID)
		}
		if !(reflect.DeepEqual(snd[0].RunID, rcv[0].RunID) || reflect.DeepEqual(snd[0].RunID, rcv[1].RunID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].RunID, rcv[0].RunID, rcv[1].RunID)
		}
		// if !(reflect.DeepEqual(snd[0].OrderID, rcv[0].OrderID) || reflect.DeepEqual(snd[0].OrderID, rcv[1].OrderID)) {
		// 	t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OrderID, rcv[0].OrderID, rcv[1].OrderID)
		// }
		if !(reflect.DeepEqual(snd[0].OriginHost, rcv[0].OriginHost) || reflect.DeepEqual(snd[0].OriginHost, rcv[1].OriginHost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginHost, rcv[0].OriginHost, rcv[1].OriginHost)
		}
		if !(reflect.DeepEqual(snd[0].Source, rcv[0].Source) || reflect.DeepEqual(snd[0].Source, rcv[1].Source)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Source, rcv[0].Source, rcv[1].Source)
		}
		if !(reflect.DeepEqual(snd[0].OriginID, rcv[0].OriginID) || reflect.DeepEqual(snd[0].OriginID, rcv[1].OriginID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginID, rcv[0].OriginID, rcv[1].OriginID)
		}
		if !(reflect.DeepEqual(snd[0].ToR, rcv[0].ToR) || reflect.DeepEqual(snd[0].ToR, rcv[1].ToR)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ToR, rcv[0].ToR, rcv[1].ToR)
		}
		if !(reflect.DeepEqual(snd[0].RequestType, rcv[0].RequestType) || reflect.DeepEqual(snd[0].RequestType, rcv[1].RequestType)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].RequestType, rcv[0].RequestType, rcv[1].RequestType)
		}
		if !(reflect.DeepEqual(snd[0].Tenant, rcv[0].Tenant) || reflect.DeepEqual(snd[0].Tenant, rcv[1].Tenant)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Tenant, rcv[0].Tenant, rcv[1].Tenant)
		}
		if !(reflect.DeepEqual(snd[0].Category, rcv[0].Category) || reflect.DeepEqual(snd[0].Category, rcv[1].Category)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Category, rcv[0].Category, rcv[1].Category)
		}
		if !(reflect.DeepEqual(snd[0].Account, rcv[0].Account) || reflect.DeepEqual(snd[0].Account, rcv[1].Account)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Account, rcv[0].Account, rcv[1].Account)
		}
		if !(reflect.DeepEqual(snd[0].Subject, rcv[0].Subject) || reflect.DeepEqual(snd[0].Subject, rcv[1].Subject)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Subject, rcv[0].Subject, rcv[1].Subject)
		}
		if !(reflect.DeepEqual(snd[0].Destination, rcv[0].Destination) || reflect.DeepEqual(snd[0].Destination, rcv[1].Destination)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Destination, rcv[0].Destination, rcv[1].Destination)
		}
		if !(snd[0].SetupTime.Equal(rcv[0].SetupTime) || snd[0].SetupTime.Equal(rcv[1].SetupTime)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].SetupTime, rcv[0].SetupTime, rcv[1].SetupTime)
		}
		if !(snd[0].AnswerTime.Equal(rcv[0].AnswerTime) || snd[0].AnswerTime.Equal(rcv[1].AnswerTime)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].AnswerTime, rcv[0].AnswerTime, rcv[1].AnswerTime)
		}
		if !(reflect.DeepEqual(snd[0].Usage, rcv[0].Usage) || reflect.DeepEqual(snd[0].Usage, rcv[1].Usage)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Usage, rcv[0].Usage, rcv[1].Usage)
		}
		if !(reflect.DeepEqual(snd[0].ExtraFields, rcv[0].ExtraFields) || reflect.DeepEqual(snd[0].ExtraFields, rcv[1].ExtraFields)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ExtraFields, rcv[0].ExtraFields, rcv[1].ExtraFields)
		}
		if !(reflect.DeepEqual(snd[0].CostSource, rcv[0].CostSource) || reflect.DeepEqual(snd[0].CostSource, rcv[1].CostSource)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CostSource, rcv[0].CostSource, rcv[1].CostSource)
		}
		if !(reflect.DeepEqual(snd[0].Cost, rcv[0].Cost) || reflect.DeepEqual(snd[0].Cost, rcv[1].Cost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Cost, rcv[0].Cost, rcv[1].Cost)
		}
		if !(reflect.DeepEqual(snd[0].ExtraInfo, rcv[0].ExtraInfo) || reflect.DeepEqual(snd[0].ExtraInfo, rcv[1].ExtraInfo)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].ExtraInfo, rcv[0].ExtraInfo, rcv[1].ExtraInfo)
		}
		if !(reflect.DeepEqual(snd[0].PreRated, rcv[0].PreRated) || reflect.DeepEqual(snd[0].PreRated, rcv[1].PreRated)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].PreRated, rcv[0].PreRated, rcv[1].PreRated)
		}
		if !(reflect.DeepEqual(snd[0].Partial, rcv[0].Partial) || reflect.DeepEqual(snd[0].Partial, rcv[1].Partial)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].Partial, rcv[0].Partial, rcv[1].Partial)
		}
		if !reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) {
			t.Errorf("Expecting: %+v, received: %+v", snd[0].CostDetails, rcv[0].CostDetails)
		}
	}
	// UPDATE
	snd[0].OriginHost = "host3"
	snd[1].OriginHost = "host3"
	for _, cdr := range snd {
		if err := storDB.SetCDR(cdr, true); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, _, err := storDB.GetCDRs(&filter, false); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].OriginHost, rcv[0].OriginHost) || reflect.DeepEqual(snd[0].OriginHost, rcv[1].OriginHost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginHost, rcv[0].OriginHost, rcv[1].OriginHost)
		}
	}
	// REMOVE
	if _, _, err := storDB.GetCDRs(&filter, true); err != nil {
		t.Error(err)
	}
	// READ
	if _, _, err := storDB.GetCDRs(&filter, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDSMCosts(t *testing.T) {
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:       "1",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:       "2",
			OriginHost:  "host2",
			OriginID:    "2",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host2", ""); err != nil {
		t.Error(err)
	} else {
		if !(reflect.DeepEqual(snd[0].CGRID, rcv[0].CGRID) || reflect.DeepEqual(snd[0].CGRID, rcv[1].CGRID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].CGRID, rcv[0].CGRID, rcv[1].CGRID)
		}
		if !(reflect.DeepEqual(snd[0].RunID, rcv[0].RunID) || reflect.DeepEqual(snd[0].RunID, rcv[1].RunID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].RunID, rcv[0].RunID, rcv[1].RunID)
		}
		if !(reflect.DeepEqual(snd[0].OriginHost, rcv[0].OriginHost) || reflect.DeepEqual(snd[0].OriginHost, rcv[1].OriginHost)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginHost, rcv[0].OriginHost, rcv[1].OriginHost)
		}
		if !(reflect.DeepEqual(snd[0].OriginID, rcv[0].OriginID) || reflect.DeepEqual(snd[0].OriginID, rcv[1].OriginID)) {
			t.Errorf("Expecting: %+v, received: %+v || %+v", snd[0].OriginID, rcv[0].OriginID, rcv[1].OriginID)
		}
		if !reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) {
			t.Errorf("Expecting: %+v, received: %+v ", utils.ToJSON(snd[0].CostDetails), utils.ToJSON(rcv[0].CostDetails))
		}
	}
	// REMOVE
	for _, smc := range snd {
		if err := storDB.RemoveSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitCRUDSMCosts2(t *testing.T) {
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		{
			CGRID:       "CGRID1",
			RunID:       "11",
			OriginHost:  "host22",
			OriginID:    "O1",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID2",
			RunID:       "12",
			OriginHost:  "host22",
			OriginID:    "O2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID3",
			RunID:       "13",
			OriginHost:  "host23",
			OriginID:    "O3",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host22", ""); err != nil {
		t.Fatal(err)
	} else if len(rcv) != 2 {
		t.Errorf("Expected 2 results received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{
		RunIDs:         []string{"12", "13"},
		NotRunIDs:      []string{"11"},
		OriginHosts:    []string{"host22", "host23"},
		NotOriginHosts: []string{"host21"},
	}); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "", ""); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 {
		t.Errorf("Expected 1 result received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{}); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testStorDBitFlush(t *testing.T) {
	if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage", cfg.StorDbCfg().Type)); err != nil {
		t.Error(err)
	}
}

func testStorDBitCRUDVersions(t *testing.T) {
	// CREATE
	vrs := Versions{utils.CostDetails: 1}
	if err := storDB.SetVersions(vrs, true); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", vrs, rcv)
	}

	// UPDATE
	vrs = Versions{utils.CostDetails: 2, "OTHER_KEY": 1}
	if err := storDB.SetVersions(vrs, false); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", vrs, rcv)
	}

	// REMOVE
	vrs = Versions{"OTHER_KEY": 1}
	if err := storDB.RemoveVersions(vrs); err != nil {
		t.Error(err)
	}
	if rcv, err := storDB.GetVersions(utils.CostDetails); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 || rcv[utils.CostDetails] != 2 {
		t.Errorf("Received: %+v", rcv)
	}

	if _, err := storDB.GetVersions("UNKNOWN"); err != utils.ErrNotFound {
		t.Error(err)
	}

	vrs = Versions{"UNKNOWN": 1}
	if err := storDB.RemoveVersions(vrs); err != nil {
		t.Error(err)
	}

	if err := storDB.RemoveVersions(nil); err != nil {
		t.Error(err)
	}

	if rcv, err := storDB.GetVersions(""); err != utils.ErrNotFound {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Received: %+v", rcv)
	}

}
