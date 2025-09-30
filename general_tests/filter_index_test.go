//go:build performance

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

package general_tests

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var perf = flag.Bool("perf", false, "check performance")

func TestFilterIndexUpdates(t *testing.T) {
	var cfgJSON string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgJSON = `{
"data_db": {
	"db_type": "*internal"
}
}`
	case utils.MetaRedis:
		cfgJSON = `{
"data_db": {
	"db_type": "*redis"
	// "opts": {
	// 	"redisPoolPipelineWindow": "0"
	// }
}
}`
	case utils.MetaMongo:
		cfgJSON = `{
"data_db": {
	"db_type": "*mongo", 
	"db_port": 27017
}
}`
	}
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(cfgJSON)
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	defer dataDB.Close()
	if err := dataDB.Flush(""); err != nil {
		t.Fatal(err)
	}
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)

	checkIndexes := func(want map[string]utils.StringSet) {
		t.Helper()
		if *perf {
			before := time.Now()
			defer func() {
				t.Helper()
				dur := time.Since(before)
				t.Logf("checkIndexes() took %s", dur)
			}()
		}

		indexes, err := dm.GetIndexes(utils.CacheAttributeFilterIndexes, "cgrates.org:simpleauth",
			true, false)
		if err != nil {
			if len(want) == 0 && err == utils.ErrNotFound {
				return
			}
			t.Fatalf("dm.GetIndexes unexpected error: %v", err)
		}
		if !reflect.DeepEqual(want, indexes) {
			t.Errorf("dm.GetIndexes = %s, want %s",
				utils.ToJSON(indexes), utils.ToJSON(want))
		}
	}

	setFilter := func(id, typ string, values ...string) {
		t.Helper()
		if *perf {
			before := time.Now()
			defer func() {
				t.Helper()
				dur := time.Since(before)
				t.Logf("setFilter(%q) took %s", id, dur)
			}()
		}
		fltr := &engine.Filter{
			Tenant: "cgrates.org",
			ID:     id,
			Rules: []*engine.FilterRule{
				{
					Type:    typ,
					Element: "~*req.Destination",
					Values:  values,
				},
			},
		}
		if err := dm.SetFilter(fltr, true); err != nil {
			t.Fatal(err)
		}
	}

	setProfiles := func(startIdx, prflCount int, fltrIDs ...string) {
		t.Helper()
		if *perf {
			before := time.Now()
			defer func() {
				t.Helper()
				dur := time.Since(before)
				t.Logf("setProfiles(%d, %d, %q) took %s",
					startIdx, prflCount, fltrIDs, dur)
			}()
		}
		for i := range prflCount {
			ap := &engine.AttributeProfile{
				Tenant:    "cgrates.org",
				ID:        fmt.Sprintf("test%d", startIdx+i),
				FilterIDs: fltrIDs,
				Contexts:  []string{"simpleauth"},
				Attributes: []*engine.Attribute{
					{
						Path:  "*req.Password",
						Type:  utils.MetaConstant,
						Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
					},
				},
			}
			if err := dm.SetAttributeProfile(ap, true); err != nil {
				t.Fatal(err)
			}
		}
	}

	removeProfiles := func(startIdx, prflCount int) {
		t.Helper()
		if *perf {
			before := time.Now()
			defer func() {
				t.Helper()
				dur := time.Since(before)
				t.Logf("removeProfiles(%d, %d) took %s", startIdx, prflCount, dur)
			}()
		}
		for i := range prflCount {
			if err := dm.RemoveAttributeProfile("cgrates.org", fmt.Sprintf("test%d", startIdx+i), true); err != nil {
				t.Fatal(err)
			}
		}
	}

	setFilter("FLTR1", utils.MetaString, "1001", "1002")
	setProfiles(0, 1, "FLTR1")
	checkIndexes(map[string]utils.StringSet{
		"*string:*req.Destination:1001": {
			"test0": {},
		},
		"*string:*req.Destination:1002": {
			"test0": {},
		},
	})

	setProfiles(0, 2, "*string:~*req.Account:2001|2002", "FLTR1")
	checkIndexes(map[string]utils.StringSet{
		"*string:*req.Account:2001": {
			"test0": {},
			"test1": {},
		},
		"*string:*req.Account:2002": {
			"test0": {},
			"test1": {},
		},
		"*string:*req.Destination:1001": {
			"test0": {},
			"test1": {},
		},
		"*string:*req.Destination:1002": {
			"test0": {},
			"test1": {},
		},
	})

	setProfiles(0, 2, "FLTR1")
	checkIndexes(map[string]utils.StringSet{
		"*string:*req.Destination:1001": {
			"test0": {},
			"test1": {},
		},
		"*string:*req.Destination:1002": {
			"test0": {},
			"test1": {},
		},
	})

	setFilter("FLTR2", utils.MetaPrefix, "1003", "1004")
	setProfiles(0, 2, "FLTR2")
	checkIndexes(map[string]utils.StringSet{
		"*prefix:*req.Destination:1003": {
			"test0": {},
			"test1": {},
		},
		"*prefix:*req.Destination:1004": {
			"test0": {},
			"test1": {},
		},
	})

	setFilter("FLTR2", utils.MetaPrefix, "1004", "1005")
	checkIndexes(map[string]utils.StringSet{
		"*prefix:*req.Destination:1004": {
			"test0": {},
			"test1": {},
		},
		"*prefix:*req.Destination:1005": {
			"test0": {},
			"test1": {},
		},
	})

	setFilter("FLTR2", utils.MetaString, "1005", "1006")
	checkIndexes(map[string]utils.StringSet{
		"*string:*req.Destination:1005": {
			"test0": {},
			"test1": {},
		},
		"*string:*req.Destination:1006": {
			"test0": {},
			"test1": {},
		},
	})

	removeProfiles(0, 1) // Remove test0
	checkIndexes(map[string]utils.StringSet{
		"*string:*req.Destination:1005": {
			"test1": {},
		},
		"*string:*req.Destination:1006": {
			"test1": {},
		},
	})

	removeProfiles(1, 1) // Remove test1
	checkIndexes(map[string]utils.StringSet{})
}

func BenchmarkFilterIndexUpdates(b *testing.B) {
	cases := []struct {
		name          string
		profiles      int
		initialValues int
		finalValues   []string
	}{
		{"Remove1_1P_10V", 1, 10, generateValues(0, 9)}, // Keep values 0-8 (remove value 9)
		{"Remove5_1P_10V", 1, 10, generateValues(0, 5)}, // Keep values 0-4 (remove values 5-9)
		{"Remove9_1P_10V", 1, 10, generateValues(0, 1)}, // Keep value 0 (remove values 1-9)
		{"Add1_1P_10V", 1, 10, generateValues(0, 11)},   // Keep values 0-9, add value 10
		{"Add5_1P_10V", 1, 10, generateValues(0, 15)},   // Keep values 0-9, add values 10-14

		{"Remove1_10P_100V", 10, 100, generateValues(0, 99)},  // Keep values 0-98 (remove value 99)
		{"Remove10_10P_100V", 10, 100, generateValues(0, 90)}, // Keep values 0-89 (remove values 90-99)
		{"Remove50_10P_100V", 10, 100, generateValues(0, 50)}, // Keep values 0-49 (remove values 50-99)
		{"Remove90_10P_100V", 10, 100, generateValues(0, 10)}, // Keep values 0-9 (remove values 10-99)
		{"Add1_10P_100V", 10, 100, generateValues(0, 101)},    // Keep values 0-99, add value 100
		{"Add10_10P_100V", 10, 100, generateValues(0, 110)},   // Keep values 0-99, add values 100-109
		{"Add50_10P_100V", 10, 100, generateValues(0, 150)},   // Keep values 0-99, add values 100-149

		{"Remove1_100P_500V", 100, 500, generateValues(0, 499)},  // Keep values 0-498 (remove value 499)
		{"Remove50_100P_500V", 100, 500, generateValues(0, 450)}, // Keep values 0-449 (remove values 450-499)
		{"Remove450_100P_500V", 100, 500, generateValues(0, 50)}, // Keep values 0-49 (remove values 50-499)
		{"Add1_100P_500V", 100, 500, generateValues(0, 501)},     // Keep values 0-499, add value 500
		{"Add50_100P_500V", 100, 500, generateValues(0, 550)},    // Keep values 0-499, add values 500-549
		{"Add100_100P_500V", 100, 500, generateValues(0, 600)},   // Keep values 0-499, add values 500-599

		{"Remove990_10P_1000V", 10, 1000, generateValues(0, 10)}, // Keep values 0-9 (remove values 10-999)
		{"Remove495_50P_500V", 50, 500, generateValues(0, 5)},    // Keep values 0-4 (remove values 5-499)
		{"Remove198_100P_200V", 100, 200, generateValues(0, 2)},  // Keep values 0-1 (remove values 2-199)
		{"Add990_10P_10V", 10, 10, generateValues(0, 1000)},      // Keep values 0-9, add values 10-999
		{"Add495_50P_5V", 50, 5, generateValues(0, 500)},         // Keep values 0-4, add values 5-499

		{"Replace1_1P_10V", 1, 10,
			append(generateValues(0, 9), generateValues(1000, 1001)...)}, // Keep 0-8, replace 9 with 1000
		{"Replace5_10P_100V", 10, 100,
			append(generateValues(0, 95), generateValues(1000, 1005)...)}, // Keep 0-94, replace 95-99 with 1000-1004
		{"Replace50_100P_500V", 100, 500,
			append(generateValues(0, 450), generateValues(1000, 1050)...)}, // Keep 0-449, replace 450-499 with 1000-1049
		{"Replace400_100P_500V", 100, 500,
			append(generateValues(0, 100), generateValues(1000, 1400)...)}, // Keep 0-99, replace 100-499 with 1000-1399
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			benchmarkFilterUpdate(b, tc.profiles, tc.initialValues, tc.finalValues)
		})
	}
}

func benchmarkFilterUpdate(b *testing.B, profileCount, initialValueCount int, finalValues []string) {
	var cfgJSON string
	switch *utils.DBType {
	case utils.MetaInternal:
		cfgJSON = `{
"data_db": {
	"db_type": "*internal"
}
}`
	case utils.MetaRedis:
		cfgJSON = `{
"data_db": {
	"db_type": "*redis",
	// "opts": {
	// 	"redisPoolPipelineWindow": "0"
	// }
}
}`
	case utils.MetaMongo:
		cfgJSON = `{
"data_db": {
	"db_type": "*mongo", 
	"db_port": 27017
}
}`
	}

	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(cfgJSON)
	if err != nil {
		b.Fatal(err)
	}

	dataDB, err := engine.NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		b.Fatal(err)
	}
	defer dataDB.Close()

	if err := dataDB.Flush(""); err != nil {
		b.Fatal(err)
	}

	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)

	initialValues := generateValues(0, initialValueCount)
	initialFilter := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "BENCH_FLTR",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  initialValues,
			},
		},
	}

	if err := dm.SetFilter(initialFilter, true); err != nil {
		b.Fatal(err)
	}

	for i := range profileCount {
		profile := &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        fmt.Sprintf("bench_test%d", i),
			FilterIDs: []string{"BENCH_FLTR"},
			Contexts:  []string{"simpleauth"},
			Attributes: []*engine.Attribute{
				{
					Path:  "*req.Password",
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
				},
			},
		}
		if err := dm.SetAttributeProfile(profile, true); err != nil {
			b.Fatal(err)
		}
	}

	updateFilter := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "BENCH_FLTR",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  finalValues,
			},
		},
	}

	for b.Loop() {
		b.StopTimer()
		if err := dm.SetFilter(initialFilter, true); err != nil {
			b.Fatal(err)
		}

		b.StartTimer()
		if err := dm.SetFilter(updateFilter, true); err != nil {
			b.Fatal(err)
		}
	}
}

func generateValues(start, end int) []string {
	values := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		values = append(values, strconv.Itoa(i))
	}
	return values
}
