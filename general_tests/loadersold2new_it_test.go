//go:build flaky

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
package general_tests

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	testLdro2nRtCfgPath string
	testLdro2nRtCfg     *config.CGRConfig
	testLdro2nRtRPC     *birpc.Client
	testLdro2nDirs      = []string{"/tmp/In", "/tmp/Out", "/tmp/parsed"}

	testLdro2nRtTests = []func(t *testing.T){
		testCreateDirs,
		testLdro2nRtLoadConfig,
		testLdro2nRtFlushDBs,

		testLdro2nRtStartEngine,
		testLdro2nRtRPCConn,
		testLdro2nRtLoadTP,
		testLdro2nRtCheckData,
		testLdro2nRtStopCgrEngine,
		testRemoveDirs,
	}
)

func TestLdro2nRtChange(t *testing.T) {
	var testLdro2nRtCfgDir string
	switch *utils.DBType {
	case utils.MetaInternal:
		testLdro2nRtCfgDir = "loaders_old2new_internal"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		testLdro2nRtCfgDir = "loaders_old2new_mysql"
	case utils.MetaMongo:
		testLdro2nRtCfgDir = "loaders_old2new_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	testLdro2nRtCfgPath = path.Join(*utils.DataDir, "conf", "samples", testLdro2nRtCfgDir)
	for _, testLdro2nRtTest := range testLdro2nRtTests {
		t.Run(testLdro2nRtCfgDir, testLdro2nRtTest)
	}
}

func testCreateDirs(t *testing.T) {
	for _, dir := range testLdro2nDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
}

func testRemoveDirs(t *testing.T) {
	for _, dir := range testLdro2nDirs {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}
}
func testLdro2nRtLoadConfig(t *testing.T) {
	var err error
	if testLdro2nRtCfg, err = config.NewCGRConfigFromPath(context.Background(), testLdro2nRtCfgPath); err != nil {
		t.Error(err)
	}
}

func testLdro2nRtFlushDBs(t *testing.T) {
	if err := engine.InitDB(testLdro2nRtCfg); err != nil {
		t.Fatal(err)
	}
}

func testLdro2nRtStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testLdro2nRtCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testLdro2nRtRPCConn(t *testing.T) {
	testLdro2nRtRPC = engine.NewRPCClient(t, testLdro2nRtCfg.ListenCfg(), *utils.Encoding)
}

func testLdro2nRtLoadTP(t *testing.T) {
	for _, file := range []string{
		"Timings.csv",
		"Destinations.csv",
		"Rates.csv",
		"DestinationRates.csv",
		"RatingPlans.csv",
	} {
		if err := copyFile(filepath.Join(*utils.DataDir, "tariffplans", "oldtutorial2", file), filepath.Join("/tmp/In", file)); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(time.Second)
	if err := copyFile(filepath.Join(*utils.DataDir, "tariffplans", "oldtutorial2", "RatingProfiles.csv"), filepath.Join("/tmp/In", "RatingProfiles.csv")); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}

func testLdro2nRtStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func copyFile(src, dst string) (err error) {
	var in *os.File
	if in, err = os.Open(src); err != nil {
		return
	}
	defer in.Close()

	var si os.FileInfo
	if si, err = in.Stat(); err != nil {
		return
	}
	var out *os.File
	if out, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, si.Mode()); err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	return out.Sync()
}

func testLdro2nRtCheckData(t *testing.T) {
	expIDs := []string{"call*any", "data*any", "callSPECIAL_1002", "generic*any", "call1001"}
	sort.Strings(expIDs)
	var rateIDs []string
	if err := testLdro2nRtRPC.Call(context.Background(), utils.AdminSv1GetRateProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rateIDs); err != nil {
		t.Error(err)
	} else if sort.Strings(rateIDs); !reflect.DeepEqual(rateIDs, expIDs) {
		t.Errorf("expected: %s, \nreceived: %s", utils.ToJSON(expIDs), utils.ToJSON(rateIDs))
	}

	expRatePrf := utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "call*any",
		FilterIDs: []string{"*ai:~*opts.*startTime:2014-01-14T00:00:00Z", "*string:~*req.Category:call"},
		MinCost:   utils.NewDecimal(0, 0),
		MaxCost:   utils.NewDecimal(0, 0),
		Rates: map[string]*utils.Rate{
			"RP_RETAIL1_DR_1007_MAXCOST_DISC_*any": {
				ID:              "RP_RETAIL1_DR_1007_MAXCOST_DISC_*any",
				FilterIDs:       []string{"*prefix:~*req.Destination:1007"},
				ActivationTimes: "",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimalFromFloat64(0),
					FixedFee:      utils.NewDecimalFromFloat64(0),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.01),
					Unit:          utils.NewDecimalFromFloat64(1000000000),
					Increment:     utils.NewDecimalFromFloat64(1000000000),
				}},
			},
			"RP_RETAIL1_DR_FS_10CNT_OFFPEAK_EVENING": {
				ID:              "RP_RETAIL1_DR_FS_10CNT_OFFPEAK_EVENING",
				FilterIDs:       []string{"*prefix:~*req.Destination:10"},
				ActivationTimes: "* 00 * * 1,2,3,4,5",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimalFromFloat64(0),
					FixedFee:      utils.NewDecimalFromFloat64(0.2),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.1),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(60000000000),
				}, {
					IntervalStart: utils.NewDecimalFromFloat64(60000000000),
					FixedFee:      utils.NewDecimalFromFloat64(0),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.05),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(1000000000),
				}},
			},
			"RP_RETAIL1_DR_FS_10CNT_OFFPEAK_MORNING": {
				ID:              "RP_RETAIL1_DR_FS_10CNT_OFFPEAK_MORNING",
				FilterIDs:       []string{"*prefix:~*req.Destination:10"},
				ActivationTimes: "* 00 * * 1,2,3,4,5",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimalFromFloat64(0),
					FixedFee:      utils.NewDecimalFromFloat64(0.2),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.1),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(60000000000),
				}, {
					IntervalStart: utils.NewDecimalFromFloat64(60000000000),
					FixedFee:      utils.NewDecimalFromFloat64(0),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.05),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(1000000000),
				}},
			},
			"RP_RETAIL1_DR_FS_10CNT_OFFPEAK_WEEKEND": {
				ID:              "RP_RETAIL1_DR_FS_10CNT_OFFPEAK_WEEKEND",
				FilterIDs:       []string{"*prefix:~*req.Destination:10"},
				ActivationTimes: "* 00 * * 6,0",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimalFromFloat64(0),
					FixedFee:      utils.NewDecimalFromFloat64(0.2),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.1),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(60000000000),
				}, {
					IntervalStart: utils.NewDecimalFromFloat64(60000000000),
					FixedFee:      utils.NewDecimalFromFloat64(0),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.05),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(1000000000),
				}},
			},
			"RP_RETAIL1_DR_FS_40CNT_PEAK": {
				ID:              "RP_RETAIL1_DR_FS_40CNT_PEAK",
				FilterIDs:       []string{"*prefix:~*req.Destination:10"},
				ActivationTimes: "* 00 * * 1,2,3,4,5",
				Weights:         utils.DynamicWeights{{Weight: 10}},
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimalFromFloat64(0),
					FixedFee:      utils.NewDecimalFromFloat64(0.8),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.4),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(30000000000),
				}, {
					IntervalStart: utils.NewDecimalFromFloat64(60000000000),
					FixedFee:      utils.NewDecimalFromFloat64(0),
					RecurrentFee:  utils.NewDecimalFromFloat64(0.2),
					Unit:          utils.NewDecimalFromFloat64(60000000000),
					Increment:     utils.NewDecimalFromFloat64(10000000000),
				}},
			},
		},
	}

	var rplyRatePrf utils.RateProfile
	if err := testLdro2nRtRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyRatePrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRatePrf, expRatePrf) {
		t.Errorf("expected: %s, \nreceived: %s",
			utils.ToJSON(expRatePrf), utils.ToJSON(rplyRatePrf))
	}
}
