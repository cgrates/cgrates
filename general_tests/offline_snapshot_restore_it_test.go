//go:build integration
// +build integration

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
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

func TestOfflineInternalSnapshotAndRestore(t *testing.T) { // run with sudo
	switch *utils.DBType {
	case utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres, utils.MetaRedis:
		t.SkipNow()
	}

	paths := []string{ // run 2 times, once for empty path values, once for populated
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_limit"),       // dump -1 and rewrite -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_ms_limit"), // dump ms and rewrite ms and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_limit"),       // dump -1 and rewrite -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_ms_limit"), // dump ms and rewrite ms and limit passed
	}
	for i, pth := range paths {
		dirNames := []string{"db", "*default", "*accounts", "*ip_profiles", "*ip_allocations",
			"*action_profiles", "*versions", "*charger_profiles", "*attribute_profiles",
			"*resource_profiles", "*resources", "*statqueue_profiles", "*statqueues",
			"*threshold_profiles", "*thresholds", "*filters", "*route_profiles",
			"*rate_profiles", "*ranking_profiles", "*rankings", "*trend_profiles",
			"*trends", "*load_ids", "*resource_filter_indexes", "*ip_filter_indexes",
			"*stat_filter_indexes", "*threshold_filter_indexes",
			"*route_filter_indexes", "*attribute_filter_indexes",
			"*charger_filter_indexes", "*rate_profile_filter_indexes",
			"*rate_filter_indexes", "*action_profile_filter_indexes",
			"*account_filter_indexes", "*reverse_filter_indexes", "*cdrs",
		}
		slices.Sort(dirNames)
		dfltCfg := config.NewDefaultCGRConfig()
		if err := os.MkdirAll(dfltCfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBDumpPath, 0755); err != nil {
			t.Fatal(err)
		}
		backupPath := dfltCfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBBackupPath
		if i == 3 {
			backupPath = "/tmp/customBackupPath"
		}
		if err := os.MkdirAll(backupPath, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(dfltCfg.ConfigDBCfg().Opts.InternalDBDumpPath, 0755); err != nil {
			t.Fatal(err)
		}
		t.Run("OfflineInternal"+strconv.Itoa(i), func(t *testing.T) {
			ng := engine.TestEngine{
				ConfigPath:       pth,
				GracefulShutdown: true,
				Encoding:         *utils.Encoding,
				LogBuffer:        &bytes.Buffer{},
			}
			t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
			client, cfg := ng.Run(t)
			time.Sleep(100 * time.Millisecond)

			t.Run("LoadTariffs", func(t *testing.T) {
				var reply string
				if err := client.Call(context.Background(), utils.LoaderSv1Run,
					&loaders.ArgsProcessFolder{
						APIOpts: map[string]any{
							utils.MetaCache: utils.MetaNone,
						},
					}, &reply); err != nil {
					t.Error(err)
				} else if reply != utils.OK {
					t.Error("Unexpected reply returned:", reply)
				}
				time.Sleep(100 * time.Millisecond)
			})

			var attrs []*utils.APIAttributeProfile

			t.Run("GetAttributes", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &attrs); err != nil {
					t.Errorf("AdminSv1GetAttributeProfiles failed unexpectedly: %v", err)
				}
				if len(attrs) != 8 {
					t.Fatalf("AdminSv1GetAttributeProfiles len(attrs)=%v, want 8", len(attrs))
				}
				sort.Slice(attrs, func(i, j int) bool {
					return attrs[i].ID > attrs[j].ID
				})
			})

			var chrgrs []*utils.ChargerProfile

			t.Run("GetChargers", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &chrgrs); err != nil {
					t.Errorf("AdminSv1GetChargerProfiles failed unexpectedly: %v", err)
				}
				if len(chrgrs) != 3 {
					t.Fatalf("AdminSv1GetChargerProfiles len(chrgrs)=%v, want 3", len(chrgrs))
				}
				sort.Slice(chrgrs, func(i, j int) bool {
					return chrgrs[i].ID > chrgrs[j].ID
				})
			})

			var fltrs []*engine.Filter

			t.Run("GetFilters", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetFilters,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &fltrs); err != nil {
					t.Errorf("AdminSv1GetFilters failed unexpectedly: %v", err)
				}
				if len(fltrs) != 22 {
					t.Fatalf("AdminSv1GetFilters len(fltrs)=%v, want 22", len(fltrs))
				}
				sort.Slice(fltrs, func(i, j int) bool {
					return fltrs[i].ID > fltrs[j].ID
				})
			})

			var rsrcs []*utils.ResourceProfile

			t.Run("GetResources", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetResourceProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rsrcs); err != nil {
					t.Errorf("AdminSv1GetResourceProfiles failed unexpectedly: %v", err)
				}
				if len(rsrcs) != 1 {
					t.Fatalf("AdminSv1GetResourceProfiles len(rsrcs)=%v, want 1", len(rsrcs))
				}
				sort.Slice(rsrcs, func(i, j int) bool {
					return rsrcs[i].ID > rsrcs[j].ID
				})
			})

			var stats []*engine.StatQueueProfile

			t.Run("GetStatQueueProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &stats); err != nil {
					t.Errorf("AdminSv1GetStatQueueProfiles failed unexpectedly: %v", err)
				}
				if len(stats) != 7 {
					t.Fatalf("AdminSv1GetStatQueueProfiles len(stats)=%v, want 7", len(stats))
				}
				sort.Slice(stats, func(i, j int) bool {
					return stats[i].ID > stats[j].ID
				})
			})

			var routes []*utils.RouteProfile

			t.Run("GetRouteProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &routes); err != nil {
					t.Errorf("AdminSv1GetRouteProfiles failed unexpectedly: %v", err)
				}
				if len(routes) != 12 {
					t.Fatalf("AdminSv1GetRouteProfiles len(routes)=%v, want 12", len(routes))
				}
				sort.Slice(routes, func(i, j int) bool {
					return routes[i].ID > routes[j].ID
				})
			})

			var thrsholds []*engine.ThresholdProfile

			t.Run("GetThresholdProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &thrsholds); err != nil {
					t.Errorf("AdminSv1GetThresholdProfiles failed unexpectedly: %v", err)
				}
				if len(thrsholds) != 1 {
					t.Fatalf("AdminSv1GetThresholdProfiles len(thrsholds)=%v, want 1", len(thrsholds))
				}
				sort.Slice(thrsholds, func(i, j int) bool {
					return thrsholds[i].ID > thrsholds[j].ID
				})
			})

			var rankings []*utils.RankingProfile

			t.Run("GetRankingProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rankings); err != nil {
					t.Errorf("AdminSv1GetRankingProfiles failed unexpectedly: %v", err)
				}
				if len(rankings) != 2 {
					t.Fatalf("AdminSv1GetRankingProfiles len(rankings)=%v, want 2", len(rankings))
				}
				sort.Slice(rankings, func(i, j int) bool {
					return rankings[i].ID > rankings[j].ID
				})
			})

			var trends []*utils.TrendProfile

			t.Run("GetTrendProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &trends); err != nil {
					t.Errorf("AdminSv1GetTrendProfiles failed unexpectedly: %v", err)
				}
				if len(trends) != 2 {
					t.Fatalf("AdminSv1GetTrendProfiles len(trends)=%v, want 2", len(trends))
				}
				sort.Slice(trends, func(i, j int) bool {
					return trends[i].ID > trends[j].ID
				})
			})

			var rates []*utils.RateProfile

			t.Run("GetRateProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetRateProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rates); err != nil {
					t.Errorf("AdminSv1GetRateProfiles failed unexpectedly: %v", err)
				}
				if len(rates) != 2 {
					t.Fatalf("AdminSv1GetRateProfiles len(rates)=%v, want 2", len(rates))
				}
				sort.Slice(rates, func(i, j int) bool {
					return rates[i].ID > rates[j].ID
				})
			})

			var acts []*utils.ActionProfile

			t.Run("GetActionProfiles", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetActionProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &acts); err != nil {
					t.Errorf("AdminSv1GetActionProfiles failed unexpectedly: %v", err)
				}
				if len(acts) != 1 {
					t.Fatalf("AdminSv1GetActionProfiles len(acts)=%v, want 1", len(acts))
				}
				sort.Slice(acts, func(i, j int) bool {
					return acts[i].ID > acts[j].ID
				})
			})

			var acnts []*utils.Account

			t.Run("GetAccounts", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &acnts); err != nil {
					t.Errorf("AdminSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(acnts) != 2 {
					t.Fatalf("AdminSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
				}
				sort.Slice(acnts, func(i, j int) bool {
					return acnts[i].ID > acnts[j].ID
				})
			})

			t.Run("CheckBackupFolder1", func(t *testing.T) { // make sure its empty
				var dirs int
				if err := filepath.Walk(backupPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if info.Name() != "customBackupPath" && info.Name() != "db" {
							t.Errorf("unexpected directory name <%v>", info.Name())
						}
						dirs++
					} else {
						t.Errorf("expected no files inside, received <%v>", info)
					}
					return nil
				}); err != nil {
					t.Error(err)
				} else if dirs != 1 { // expected only the folder itself
					t.Errorf("expected <%d> directories, received <%d>", 1, dirs)
				}
			})

			dumpDirs1 := []string{}
			dumpFileNames := []string{}
			t.Run("CheckDumpFolder1", func(t *testing.T) {
				if i == 1 || i == 3 {
					time.Sleep(500 * time.Millisecond) // make sure to wait for intervals to kick in
				}
				// t.Log("tree here")
				// time.Sleep(5 * time.Second)
				var dirs, files int
				if err := filepath.Walk(cfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						dumpDirs1 = append(dumpDirs1, info.Name())
						dirs++
					} else {
						dumpFileNames = append(dumpFileNames, info.Name())
						files++
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
				slices.Sort(dumpDirs1)
				if !reflect.DeepEqual(dirNames, dumpDirs1) {
					t.Errorf("expected <%v>, received <%v>", dirNames, dumpDirs1)
				} else if files != 37 {
					t.Errorf("expected <%d> files, received <%d>", 37, files)
				}
			})

			t.Run("RunSnapshotDB", func(t *testing.T) {
				bpath := ""
				zip := false
				if i == 3 {
					bpath = backupPath
					zip = true
				}
				var rply string
				if err := client.Call(context.Background(), utils.AdminSv1SnapshotDB, apis.BackupParams{BackupFolderPath: bpath, Zip: zip}, &rply); err != nil {
					t.Fatal(err)
				}
			})

			t.Run("CheckBackupFolder2", func(t *testing.T) { // make sure its populated correctly and same as dump folder
				backupDirs := []string{}
				backupFileNames := []string{}
				dumpDirsCopy := dumpDirs1
				foundBackupFolder := false
				var dirs, files int
				if err := filepath.Walk(backupPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						backupDirs = append(backupDirs, info.Name())
						if strings.HasPrefix(info.Name(), "backup_") {
							if !foundBackupFolder {
								dumpDirsCopy = append(dumpDirsCopy, info.Name())
								foundBackupFolder = true
							} else {
								t.Errorf("found more than 1 backup folder <%v>", backupDirs)
							}
						}
						dirs++
					} else {
						backupFileNames = append(backupFileNames, info.Name())
						files++
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
				slices.Sort(backupDirs)
				slices.Sort(dumpDirsCopy)
				if i == 3 {
					dumpDirsCopy = []string{"customBackupPath"} // no folders should be in the same root as backuppath for zipped backup
					if !reflect.DeepEqual(backupDirs, dumpDirsCopy) {
						t.Errorf("expected <%v>, received <%v>", dumpDirsCopy, backupDirs)
					}
					if len(backupFileNames) != 1 {
						t.Errorf("expected 1 backup zip file, received <%v>", backupFileNames)
					} else if strings.HasPrefix(backupFileNames[0], "backup_") && strings.HasPrefix(backupFileNames[0], ".zip") {
						t.Errorf("expected backup zip file, received <%v>", backupFileNames)
					}
				} else {
					if !reflect.DeepEqual(backupDirs, dumpDirsCopy) {
						t.Errorf("expected <%v>, received <%v>", dumpDirsCopy, backupDirs)
					} else if !reflect.DeepEqual(dumpFileNames, backupFileNames) { // make sure it was a perfect file copy
						t.Errorf("expected <%v>, received <%v>", dumpFileNames, backupFileNames)
					}
				}
			})

			t.Run("CheckDumpFolder2", func(t *testing.T) { // make sure new files are added
				snapshotDirs := []string{}
				snapshotFileNames := []string{}
				// t.Log("tree here2")
				time.Sleep(50 * time.Millisecond)
				var dirs, files int
				if err := filepath.Walk(cfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						snapshotDirs = append(snapshotDirs, info.Name())
						dirs++
					} else {
						snapshotFileNames = append(snapshotFileNames, info.Name())
						files++
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
				slices.Sort(snapshotDirs)
				if !reflect.DeepEqual(dirNames, snapshotDirs) {
					t.Errorf("expected <%v>, received <%v>", dirNames, snapshotDirs)
				} else if files != 36 { // expected 1 less file then dump since rewriting happens on dumps *versions
					t.Errorf("expected <%d> files, received <%d>", 36, files)
				}
				for _, sfn := range snapshotFileNames {
					// make sure file names are not the same as discarded(backed up) dump files
					if sfn != "db" && slices.Contains(dumpFileNames, sfn) {
						t.Fatalf("expected snapshot file names to be different from backed up dump file names, snapshot files <%v>, \ndump files <%v>", snapshotFileNames, dumpFileNames)
					}
				}
			})

			t.Run("GetAttributes2", func(t *testing.T) {
				var rcv []*utils.APIAttributeProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetAttributeProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 8 {
					t.Fatalf("AdminSv1GetAttributeProfiles len(rcv)=%v, want 8", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, attrs) {
					t.Errorf("Expected attributes to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(attrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetChargers2", func(t *testing.T) {
				var rcv []*utils.ChargerProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetChargerProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 3 {
					t.Fatalf("AdminSv1GetChargerProfiles len(rcv)=%v, want 3", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, chrgrs) {
					t.Errorf("Expected Chargers to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(chrgrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetFilters2", func(t *testing.T) {
				var rcv []*engine.Filter
				if err := client.Call(context.Background(), utils.AdminSv1GetFilters,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetFilters failed unexpectedly: %v", err)
				}
				if len(rcv) != 22 {
					t.Fatalf("AdminSv1GetFilters len(rcv)=%v, want 22", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, fltrs) {
					t.Errorf("Expected Filters to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(fltrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetResources2", func(t *testing.T) {
				var rcv []*utils.ResourceProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetResourceProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetResourceProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetResourceProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rsrcs) {
					t.Errorf("Expected Resources to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rsrcs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetStatQueueProfiles2", func(t *testing.T) {
				var rcv []*engine.StatQueueProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetStatQueueProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 7 {
					t.Fatalf("AdminSv1GetStatQueueProfiles len(rcv)=%v, want 7", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, stats) {
					t.Errorf("Expected StatQueueProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(stats), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRouteProfiles2", func(t *testing.T) {
				var rcv []*utils.RouteProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRouteProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 12 {
					t.Fatalf("AdminSv1GetRouteProfiles len(rcv)=%v, want 12", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, routes) {
					t.Errorf("Expected RouteProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(routes), utils.ToJSON(rcv))
				}
			})

			t.Run("GetThresholdProfiles2", func(t *testing.T) {
				var rcv []*engine.ThresholdProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetThresholdProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetThresholdProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, thrsholds) {
					t.Errorf("Expected ThresholdProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(thrsholds), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRankingProfiles2", func(t *testing.T) {
				var rcv []*utils.RankingProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRankingProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetRankingProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rankings) {
					t.Errorf("Expected RankingProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rankings), utils.ToJSON(rcv))
				}
			})

			t.Run("GetTrendProfiles2", func(t *testing.T) {
				var rcv []*utils.TrendProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetTrendProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetTrendProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, trends) {
					t.Errorf("Expected TrendProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(trends), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRateProfiles2", func(t *testing.T) {
				var rcv []*utils.RateProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRateProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRateProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetRateProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rates) {
					t.Errorf("Expected RateProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rates), utils.ToJSON(rcv))
				}
			})

			t.Run("GetActionProfiles2", func(t *testing.T) {
				var rcv []*utils.ActionProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetActionProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetActionProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetActionProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, acts) {
					t.Errorf("Expected ActionProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acts), utils.ToJSON(rcv))
				}
			})

			t.Run("GetAccounts2", func(t *testing.T) {
				var rcv []*utils.Account
				if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv2GetAccounts len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, acnts) {
					t.Errorf("Expected Accounts to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acnts), utils.ToJSON(rcv))
				}
			})

			t.Run("EngineShutdown", func(t *testing.T) {
				if err := engine.KillEngine(100); err != nil {
					t.Error(err)
				}
			})

			t.Run("CountDBFiles", func(t *testing.T) {
				var dirs, files int
				if err := filepath.Walk(cfg.DbCfg().DBConns[utils.MetaDefault].Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						dirs++
					} else {
						files++
					}
					return nil
				}); err != nil {
					t.Error(err)
				} else if dirs != 36 {
					t.Errorf("expected <%d> directories, received <%d>", 36, dirs)
				} else if files != 33 {
					t.Errorf("expected <%d> files, received <%d>", 33, files)
				}
			})

			ng.PreserveDataDB = true
			client, cfg = ng.Run(t)
			time.Sleep(100 * time.Millisecond)

			t.Run("GetAttributes3", func(t *testing.T) {
				var rcv []*utils.APIAttributeProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetAttributeProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 8 {
					t.Fatalf("AdminSv1GetAttributeProfiles len(rcv)=%v, want 8", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, attrs) {
					t.Errorf("Expected attributes to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(attrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetChargers3", func(t *testing.T) {
				var rcv []*utils.ChargerProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetChargerProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 3 {
					t.Fatalf("AdminSv1GetChargerProfiles len(rcv)=%v, want 3", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, chrgrs) {
					t.Errorf("Expected Chargers to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(chrgrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetFilters3", func(t *testing.T) {
				var rcv []*engine.Filter
				if err := client.Call(context.Background(), utils.AdminSv1GetFilters,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetFilters failed unexpectedly: %v", err)
				}
				if len(rcv) != 22 {
					t.Fatalf("AdminSv1GetFilters len(rcv)=%v, want 22", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, fltrs) {
					t.Errorf("Expected Filters to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(fltrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetResources3", func(t *testing.T) {
				var rcv []*utils.ResourceProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetResourceProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetResourceProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetResourceProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rsrcs) {
					t.Errorf("Expected Resources to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rsrcs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetStatQueueProfiles3", func(t *testing.T) {
				var rcv []*engine.StatQueueProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetStatQueueProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 7 {
					t.Fatalf("AdminSv1GetStatQueueProfiles len(rcv)=%v, want 7", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, stats) {
					t.Errorf("Expected StatQueueProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(stats), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRouteProfiles3", func(t *testing.T) {
				var rcv []*utils.RouteProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRouteProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 12 {
					t.Fatalf("AdminSv1GetRouteProfiles len(rcv)=%v, want 12", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, routes) {
					t.Errorf("Expected RouteProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(routes), utils.ToJSON(rcv))
				}
			})

			t.Run("GetThresholdProfiles3", func(t *testing.T) {
				var rcv []*engine.ThresholdProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetThresholdProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetThresholdProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, thrsholds) {
					t.Errorf("Expected ThresholdProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(thrsholds), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRankingProfiles3", func(t *testing.T) {
				var rcv []*utils.RankingProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRankingProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetRankingProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rankings) {
					t.Errorf("Expected RankingProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rankings), utils.ToJSON(rcv))
				}
			})

			t.Run("GetTrendProfiles3", func(t *testing.T) {
				var rcv []*utils.TrendProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetTrendProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetTrendProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, trends) {
					t.Errorf("Expected TrendProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(trends), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRateProfiles3", func(t *testing.T) {
				var rcv []*utils.RateProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRateProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRateProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetRateProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rates) {
					t.Errorf("Expected RateProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rates), utils.ToJSON(rcv))
				}
			})

			t.Run("GetActionProfiles3", func(t *testing.T) {
				var rcv []*utils.ActionProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetActionProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetActionProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetActionProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, acts) {
					t.Errorf("Expected ActionProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acts), utils.ToJSON(rcv))
				}
			})

			t.Run("GetAccountsAndRemove", func(t *testing.T) {
				var rcv []*utils.Account
				if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv2GetAccounts len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, acnts) {
					t.Errorf("Expected Accounts to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acnts), utils.ToJSON(rcv))
				}

				for _, acnt := range rcv {
					args := &utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{
							Tenant: acnt.Tenant,
							ID:     acnt.ID,
						},
					}
					var reply string
					if err := client.Call(context.Background(), utils.AdminSv1RemoveAccount,
						args,
						&reply); err != nil {
						t.Error(err)
					} else if reply != utils.OK {
						t.Error("Unexpected reply returned", reply)
					}
				}

			})

			var newAcc *utils.AccountWithAPIOpts

			t.Run("SetNewAccount", func(t *testing.T) {
				newAcc = &utils.AccountWithAPIOpts{
					Account: &utils.Account{
						Tenant: "cgrates.org",
						ID:     "AccAfterSnapshot",
					},
				}
				var replySet string
				if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
					newAcc, &replySet); err != nil {
					t.Error(err)
				}
			})

			t.Run("RestoreFirstBackup", func(t *testing.T) {
				bpath := ""
				if i == 3 {
					bpath = backupPath
				}
				if i == 4 {
					entries, err := os.ReadDir(backupPath)
					if err != nil {
						t.Fatal(err)
					}
					if len(entries) != 1 {
						t.Errorf("expected 1 entry, received <%v>", entries)
					}
					bpath = filepath.Join(backupPath, entries[0].Name())
				}
				var rply string
				if err := client.Call(context.Background(), utils.AdminSv1RestoreDB, bpath, &rply); err != nil {
					t.Fatal(err)
				}
			})

			t.Run("CheckNewAccount", func(t *testing.T) {
				var acntGot utils.Account
				if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
					&utils.TenantIDWithAPIOpts{
						TenantID: &utils.TenantID{
							Tenant: "cgrates.org",
							ID:     "AccAfterSnapshot",
						},
					}, &acntGot); err == nil || err.Error() != "NOT_FOUND" {
					t.Errorf("expected <%v>, received <%v>", "NOT_FOUND", err)
				}
			})

			t.Run("GetAttributes4", func(t *testing.T) {
				var rcv []*utils.APIAttributeProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetAttributeProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 8 {
					t.Fatalf("AdminSv1GetAttributeProfiles len(rcv)=%v, want 8", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, attrs) {
					t.Errorf("Expected attributes to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(attrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetChargers4", func(t *testing.T) {
				var rcv []*utils.ChargerProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetChargerProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 3 {
					t.Fatalf("AdminSv1GetChargerProfiles len(rcv)=%v, want 3", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, chrgrs) {
					t.Errorf("Expected Chargers to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(chrgrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetFilters4", func(t *testing.T) {
				var rcv []*engine.Filter
				if err := client.Call(context.Background(), utils.AdminSv1GetFilters,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetFilters failed unexpectedly: %v", err)
				}
				if len(rcv) != 22 {
					t.Fatalf("AdminSv1GetFilters len(rcv)=%v, want 22", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, fltrs) {
					t.Errorf("Expected Filters to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(fltrs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetResources4", func(t *testing.T) {
				var rcv []*utils.ResourceProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetResourceProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetResourceProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetResourceProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rsrcs) {
					t.Errorf("Expected Resources to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rsrcs), utils.ToJSON(rcv))
				}
			})

			t.Run("GetStatQueueProfiles4", func(t *testing.T) {
				var rcv []*engine.StatQueueProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetStatQueueProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 7 {
					t.Fatalf("AdminSv1GetStatQueueProfiles len(rcv)=%v, want 7", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, stats) {
					t.Errorf("Expected StatQueueProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(stats), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRouteProfiles4", func(t *testing.T) {
				var rcv []*utils.RouteProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRouteProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 12 {
					t.Fatalf("AdminSv1GetRouteProfiles len(rcv)=%v, want 12", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, routes) {
					t.Errorf("Expected RouteProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(routes), utils.ToJSON(rcv))
				}
			})

			t.Run("GetThresholdProfiles4", func(t *testing.T) {
				var rcv []*engine.ThresholdProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetThresholdProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetThresholdProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, thrsholds) {
					t.Errorf("Expected ThresholdProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(thrsholds), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRankingProfiles4", func(t *testing.T) {
				var rcv []*utils.RankingProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRankingProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRankingProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetRankingProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rankings) {
					t.Errorf("Expected RankingProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rankings), utils.ToJSON(rcv))
				}
			})

			t.Run("GetTrendProfiles4", func(t *testing.T) {
				var rcv []*utils.TrendProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetTrendProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetTrendProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetTrendProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, trends) {
					t.Errorf("Expected TrendProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(trends), utils.ToJSON(rcv))
				}
			})

			t.Run("GetRateProfiles4", func(t *testing.T) {
				var rcv []*utils.RateProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetRateProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetRateProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv1GetRateProfiles len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, rates) {
					t.Errorf("Expected RateProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(rates), utils.ToJSON(rcv))
				}
			})

			t.Run("GetActionProfiles4", func(t *testing.T) {
				var rcv []*utils.ActionProfile
				if err := client.Call(context.Background(), utils.AdminSv1GetActionProfiles,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv1GetActionProfiles failed unexpectedly: %v", err)
				}
				if len(rcv) != 1 {
					t.Fatalf("AdminSv1GetActionProfiles len(rcv)=%v, want 1", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, acts) {
					t.Errorf("Expected ActionProfiles to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acts), utils.ToJSON(rcv))
				}
			})

			t.Run("GetAccounts4", func(t *testing.T) { // accounts deleted before restoring should be recovered
				var rcv []*utils.Account
				if err := client.Call(context.Background(), utils.AdminSv1GetAccounts,
					&utils.ArgsItemIDs{
						Tenant: "cgrates.org",
					}, &rcv); err != nil {
					t.Errorf("AdminSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(rcv) != 2 {
					t.Fatalf("AdminSv2GetAccounts len(rcv)=%v, want 2", len(rcv))
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].ID > rcv[j].ID
				})

				if !reflect.DeepEqual(rcv, acnts) {
					t.Errorf("Expected Accounts to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acnts), utils.ToJSON(rcv))
				}
			})

			if err := engine.KillEngine(100); err != nil {
				t.Error(err)
			}
			if i > 2 {
				if err := os.RemoveAll(backupPath); err != nil {
					t.Error(err)
				}
			}

			if err := os.RemoveAll("/var/lib/cgrates/internal_db"); err != nil {
				t.Error(err)
			}

		})
	}
}
