//go:build integration
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

	"gorm.io/gorm/logger"

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
	testStorDBitCRUDTPAccounts,
	testStorDBitCRUDTPActionProfiles,
	testStorDBitCRUDTPDispatcherProfiles,
	testStorDBitCRUDTPDispatcherHosts,
	testStorDBitCRUDTPFilters,
	testStorDBitCRUDTPRateProfiles,
	testStorDBitCRUDTPRoutes,
	testStorDBitCRUDTPThresholds,
	testStorDBitCRUDTPAttributes,
	testStorDBitCRUDTPChargers,
	testStorDBitCRUDTpResources,
	testStorDBitCRUDTpStats,
	testStorDBitCRUDCDRs,
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
			100, 10, 0, "UTC"); err != nil {
			t.Fatal(err)
		}
		storDB.(*SQLStorage).db.Config.Logger = logger.Default.LogMode(logger.Silent)
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
		storDB.(*SQLStorage).db.Config.Logger = logger.Default.LogMode(logger.Silent)
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
	case utils.Mongo:
		test, err := storDB.IsDBEmpty()
		if err != nil {
			t.Error(err)
		} else if test != true {
			t.Errorf("Expecting: true got :%+v", test)
		}
	case utils.Postgres, utils.MySQL:
		test, err := storDB.IsDBEmpty()
		if err != nil {
			t.Error(err)
		} else if test != false {
			t.Errorf("Expecting: false got :%+v", test)
		}
	}
}

func testStorDBitCRUDTPAccounts(t *testing.T) {
	//READ
	if _, err := storDB.GetTPAccounts("sub_ID1", utils.EmptyString, "TEST_ID1"); err != utils.ErrNotFound {
		t.Error(err)
	}

	//WRITE
	var actPrf = []*utils.TPAccount{
		{
			TPid:    testTPID,
			Tenant:  "cgrates.org",
			ID:      "1001",
			Weights: ";20",
			Balances: map[string]*utils.TPAccountBalance{
				"MonetaryBalance": {
					ID:      "MonetaryBalance",
					Weights: ";10",
					Type:    utils.MetaMonetary,
					CostIncrement: []*utils.TPBalanceCostIncrement{
						{
							FilterIDs:    []string{"fltr1", "fltr2"},
							Increment:    utils.Float64Pointer(1.3),
							FixedFee:     utils.Float64Pointer(2.3),
							RecurrentFee: utils.Float64Pointer(3.3),
						},
					},
					AttributeIDs: []string{"attr1", "attr2"},
					UnitFactors: []*utils.TPBalanceUnitFactor{
						{
							FilterIDs: []string{"fltr1", "fltr2"},
							Factor:    100,
						},
						{
							FilterIDs: []string{"fltr3"},
							Factor:    200,
						},
					},
					Units: 14,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := storDB.SetTPAccounts(actPrf); err != nil {
		t.Error(err)
	}

	//READ
	rcv, err := storDB.GetTPAccounts(actPrf[0].TPid, utils.EmptyString, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	sort.Strings(rcv[0].Balances["MonetaryBalance"].AttributeIDs)
	if !(reflect.DeepEqual(rcv[0], actPrf[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n", utils.ToJSON(actPrf[0]), utils.ToJSON(rcv[0]))
	}

	//UPDATE AND READ
	actPrf[0].FilterIDs = []string{"*string:~*req.Account:1007"}
	if err := storDB.SetTPAccounts(actPrf); err != nil {
		t.Error(err)
	} else if rcv, err := storDB.GetTPAccounts(actPrf[0].TPid,
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
			Schedule:  utils.MetaASAP,
			Actions: []*utils.TPAPAction{
				{
					ID:   "TOPUP",
					Type: "*topup",
					Diktats: []*utils.TPAPDiktat{{
						Path: "~*balance.TestBalance.Value",
					}},
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			TPid:      "TEST_ID1",
			ID:        "sub_id2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			Weight:    10,
			Schedule:  utils.MetaASAP,
			Actions: []*utils.TPAPAction{
				{
					ID:   "TOPUP",
					Type: "*topup",
					Diktats: []*utils.TPAPDiktat{{
						Path: "~*balance.TestBalance.Value",
					}},
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
			Strategy:  utils.MetaFirst,
			Weight:    10,
		},
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Dsp2",
			FilterIDs: []string{"*string:~*req.Destination:10"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
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
			utils.ToJSON(tpDispatcherHosts[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			utils.ToJSON(tpDispatcherHosts[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			utils.ToJSON(tpFilters[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			utils.ToJSON(tpFilters[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			TPid:            "id_RP1",
			Tenant:          "cgrates.org",
			ID:              "RP1",
			FilterIDs:       []string{"*string:~*req.Subject:1001"},
			Weights:         ";0",
			MinCost:         0.1,
			MaxCost:         0.6,
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.TPRate{
				"FIRST_GI": {
					ID:        "FIRST_GI",
					FilterIDs: []string{"*gi:~*req.Usage:0"},
					Weights:   ";0",
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
			Sorting:   "*lowest_cost",
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RateProfileIDs:  []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weights:         ";10",
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weights: ";20",
		},
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "SUPL_2",
			FilterIDs: []string{"*string:~*req.Destination:100"},
			Sorting:   "*lowest_cost",
			Routes: []*utils.TPRoute{
				{
					ID:              "supplier1",
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc1", "Acc2"},
					RateProfileIDs:  []string{"RPL_1"},
					ResourceIDs:     []string{"ResGroup1"},
					StatIDs:         []string{"Stat1"},
					Weights:         ";10",
					Blocker:         false,
					RouteParameters: "SortingParam2",
				},
			},
			Weights: ";10",
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
			utils.ToJSON(tpRoutes[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			utils.ToJSON(tpRoutes[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			TPid:             "TH1",
			Tenant:           "cgrates.org",
			ID:               "Threshold1",
			FilterIDs:        []string{"*string:~*req.Account:1002", "*string:~*req.DryRun:*default"},
			MaxHits:          -1,
			MinSleep:         "1s",
			Blocker:          true,
			Weight:           10,
			ActionProfileIDs: []string{"Thresh1"},
			Async:            true,
		},
		{
			TPid:             "TH1",
			Tenant:           "cgrates.org",
			ID:               "Threshold2",
			FilterIDs:        []string{"*string:~*req.Destination:10"},
			MaxHits:          -1,
			MinSleep:         "1s",
			Blocker:          true,
			Weight:           20,
			ActionProfileIDs: []string{"Thresh1"},
			Async:            true,
		},
	}
	if err := storDB.SetTPThresholds(tpThresholds); err != nil {
		t.Errorf("Unable to set TPThresholds")
	}

	//READ
	if rcv, err := storDB.GetTPThresholds(tpThresholds[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rcv[0].FilterIDs)
		sort.Strings(rcv[1].FilterIDs)
		if !(reflect.DeepEqual(tpThresholds[0], rcv[0]) ||
			reflect.DeepEqual(tpThresholds[0], rcv[1])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToJSON(tpThresholds[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
		}
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
			utils.ToJSON(tpThresholds[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))

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
					Value:     "101",
					FilterIDs: []string{"*string:~*req.Account:101"},
				},
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField + utils.InInFieldSep,
					Value:     "108",
					FilterIDs: []string{"*string:~*req.Account:102"},
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
					Value:     "12",
					FilterIDs: []string{"*string:~*req.Destination:11"},
				},
				{
					Type:      utils.MetaString,
					Path:      utils.MetaReq + utils.NestingSep + utils.Destination + utils.InInFieldSep,
					Value:     "13",
					FilterIDs: []string{"*string:~*req.Destination:10"},
				},
			},
		},
	}
	if err := storDB.SetTPAttributes(tpAProfile); err != nil {
		t.Errorf("Unable to set TPActionProfile:%s", err)
	}

	//READ
	if rcv, err := storDB.GetTPAttributes(tpAProfile[0].TPid, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rcv[0].Attributes, func(i, j int) bool { return rcv[0].Attributes[i].Value < rcv[0].Attributes[j].Value })
		sort.Slice(rcv[1].Attributes, func(i, j int) bool { return rcv[1].Attributes[i].Value < rcv[1].Attributes[j].Value })
		if !(reflect.DeepEqual(rcv[0], tpAProfile[0]) ||
			reflect.DeepEqual(rcv[1], tpAProfile[0])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToJSON(tpAProfile[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
		}
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
	} else {
		sort.Slice(rcv[0].Attributes, func(i, j int) bool { return rcv[0].Attributes[i].Value < rcv[0].Attributes[j].Value })
		sort.Slice(rcv[1].Attributes, func(i, j int) bool { return rcv[1].Attributes[i].Value < rcv[1].Attributes[j].Value })
		if !(reflect.DeepEqual(rcv[0], tpAProfile[0]) ||
			reflect.DeepEqual(rcv[1], tpAProfile[0])) {
			t.Errorf("Expecting:\n%+v\nReceived:\n%+v\n||\n%+v",
				utils.ToJSON(tpAProfile[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
		}
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
			TPid:         "TP_ID",
			Tenant:       "cgrates.org",
			ID:           "Chrgs1",
			FilterIDs:    []string{"*string:~*req.Account:1002"},
			AttributeIDs: make([]string, 0),
			RunID:        utils.MetaDefault,
			Weight:       20,
		},
		{
			TPid:         "TP_id",
			Tenant:       "cgrates.org",
			ID:           "Chrgs2",
			FilterIDs:    []string{"*string:~*req.Destination:10"},
			AttributeIDs: make([]string, 0),
			RunID:        utils.MetaDefault,
			Weight:       10,
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
			utils.ToJSON(tpChargers[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
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
			utils.ToJSON(tpChargers[0]), utils.ToJSON(rcv[0]), utils.ToJSON(rcv[1]))
	}

	//REMOVE and READ
	if err := storDB.RemTpData(utils.EmptyString, tpChargers[0].TPid, nil); err != nil {
		t.Error(err)
	} else if _, err := storDB.GetTPChargers(tpChargers[0].TPid, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
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
			TPid:         "testTPid",
			ID:           "testTag2",
			Weight:       0.0,
			Limit:        "test",
			ThresholdIDs: []string{"1x", "2x"},
			FilterIDs:    []string{"FLTR_RES_2"},
			Blocker:      true,
			Stored:       false,
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
			TPid:        "TEST_TPID",
			Tenant:      "Test",
			ID:          "Stats1",
			FilterIDs:   []string{"FLTR_1"},
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
			TPid:        "TEST_TPID",
			Tenant:      "Test",
			ID:          "Stats1",
			FilterIDs:   []string{"FLTR_1"},
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
			CGRID:      "88ed9c38005f07576a1e1af293063833b60edcc6",
			RunID:      "1",
			OrderID:    0,
			OriginHost: "host1",
			OriginID:   "1",
			Usage:      1000000000,
			// CostDetails: NewBareEventCost(),
			ExtraFields: map[string]string{"Service-Context-Id": "voice@huawei.com"},
		},
		{
			CGRID:      "88ed9c38005f07576a1e1af293063833b60edcc2",
			RunID:      "2",
			OrderID:    0,
			OriginHost: "host2",
			OriginID:   "2",
			Usage:      1000000000,
			// CostDetails: NewBareEventCost(),
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
		// if !reflect.DeepEqual(snd[0].CostDetails, rcv[0].CostDetails) {
		// t.Errorf("Expecting: %+v, received: %+v", snd[0].CostDetails, rcv[0].CostDetails)
		// }
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
