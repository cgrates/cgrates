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
	"bufio"
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
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestOfflineInternalSnapshotAndRestore(t *testing.T) { // run with sudo
	switch *utils.DBType {
	case utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres, utils.MetaRedis:
		t.SkipNow()
	}
	paths := []string{
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_limit"),       // dump -1 and rewrite -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_ms_limit"), // dump ms and rewrite ms and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_limit"),       // dump -1 and rewrite -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_ms_limit"), // dump ms and rewrite ms and limit passed
	}
	for i, pth := range paths {
		dataDBDirNames := []string{"*account_action_plans", "*accounts", "*action_plans", "*action_triggers", "*actions", "*attribute_filter_indexes", "*attribute_profiles", "*charger_filter_indexes", "*charger_profiles", "*default", "*destinations", "*dispatcher_filter_indexes", "*dispatcher_hosts", "*dispatcher_profiles", "*filters", "*ip_allocations", "*ip_filter_indexes", "*ip_profiles", "*load_ids", "*ranking_profiles", "*rankings", "*rating_plans", "*rating_profiles", "*resource_filter_indexes", "*resource_profiles", "*resources", "*reverse_destinations", "*reverse_filter_indexes", "*route_filter_indexes", "*route_profiles", "*sessions_backup", "*shared_groups", "*stat_filter_indexes", "*statqueue_profiles", "*statqueues", "*threshold_filter_indexes", "*threshold_profiles", "*thresholds", "*timings", "*trend_profiles", "*trends", "*versions", "datadb"}
		slices.Sort(dataDBDirNames)
		storDBDirNames := []string{"*cdrs", "*default", "*session_costs", "*tp_account_actions", "*tp_action_plans", "*tp_action_triggers", "*tp_actions", "*tp_attributes", "*tp_chargers", "*tp_destination_rates", "*tp_destinations", "*tp_dispatcher_hosts", "*tp_dispatcher_profiles", "*tp_filters", "*tp_ips", "*tp_rankings", "*tp_rates", "*tp_rating_plans", "*tp_rating_profiles", "*tp_resources", "*tp_routes", "*tp_shared_groups", "*tp_stats", "*tp_thresholds", "*tp_timings", "*tp_trends", "*versions", "stordb"}
		slices.Sort(storDBDirNames)
		dfltCfg := config.NewDefaultCGRConfig()
		if err := os.MkdirAll(dfltCfg.DataDbCfg().Opts.InternalDBDumpPath, 0755); err != nil {
			t.Fatal(err)
		}
		dataDBBackupPath := dfltCfg.DataDbCfg().Opts.InternalDBBackupPath
		if err := os.MkdirAll(dfltCfg.StorDbCfg().Opts.InternalDBDumpPath, 0755); err != nil {
			t.Fatal(err)
		}
		storDBBackupPath := dfltCfg.StorDbCfg().Opts.InternalDBBackupPath
		if i == 3 {
			dataDBBackupPath = "/tmp/datadb/customBackupPath"
			storDBBackupPath = "/tmp/stordb/customBackupPath"
		}
		if err := os.MkdirAll(dataDBBackupPath, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(storDBBackupPath, 0755); err != nil {
			t.Fatal(err)
		}
		t.Run("OfflineInternal"+strconv.Itoa(i), func(t *testing.T) {
			ng := engine.TestEngine{
				ConfigPath:       pth,
				GracefulShutdown: true,
				PreserveDataDB:   true,
				// LogBuffer:        &bytes.Buffer{},
			}
			// t.Cleanup(func() { fmt.Println(ng.LogBuffer) })
			client, cfg := ng.Run(t)
			time.Sleep(100 * time.Millisecond)

			t.Run("LoadTariffs", func(t *testing.T) {
				attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
				var loadInst utils.LoadInstance
				if err := client.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
					t.Error(err)
				}
				time.Sleep(100 * time.Millisecond)
			})

			t.Run("ExportDataDB", func(t *testing.T) {
				// exports Attributes, Chargers, Dispatchers, DispatcherHosts, Filters, Resources, Stats, Routes, Thresholds, Rankings, Trends
				var reply string
				if err := client.Call(context.Background(), utils.APIerSv1ExportToFolder, utils.ArgExportToFolder{Path: "/tmp/ExportPath1"}, &reply); err != nil {
					t.Error(err)
				} else if reply != utils.OK {
					t.Errorf("Expected: <%v>, received: <%v>", utils.OK, reply)
				}
			})

			var acnts []*engine.Account

			t.Run("GetAccounts", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
					&utils.AttrGetAccounts{
						Tenant: "cgrates.org",
					}, &acnts); err != nil {
					t.Errorf("APIerSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(acnts) != 2 {
					t.Fatalf("APIerSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
				}
				sort.Slice(acnts, func(i, j int) bool {
					return acnts[i].ID > acnts[j].ID
				})
			})

			ratingPl := new(engine.RatingPlan)
			ratingPl2 := new(engine.RatingPlan)
			ratingPl3 := new(engine.RatingPlan)
			ratingPl4 := new(engine.RatingPlan)
			ratingPl5 := new(engine.RatingPlan)
			ratingPl6 := new(engine.RatingPlan)
			ratingPl7 := new(engine.RatingPlan)
			ratingPl8 := new(engine.RatingPlan)
			ratingPl9 := new(engine.RatingPlan)
			ratingPl10 := new(engine.RatingPlan)
			ratingPl11 := new(engine.RatingPlan)

			t.Run("GetRatingPlans", func(t *testing.T) {
				rplnId := "RP_TESTIT1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_SPECIAL_1002"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl2); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_RETAIL1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl3); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_ANY2CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl4); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_ANY1CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl5); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_TEST"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl6); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_MOBILE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl7); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_LOCAL"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl8); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_FREE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl9); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_ANY2CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl10); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				rplnId = "RP_ANY1CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, ratingPl11); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
			})

			var rProf engine.RatingProfile

			t.Run("GetRatingProfile", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingProfile,
					&utils.AttrGetRatingProfile{
						Tenant:   "cgrates.org",
						Category: "free",
						Subject:  "RP_FREE",
					}, &rProf); err != nil {
					t.Error(err)
				}

			})

			var dests []*engine.Destination
			t.Run("GetDestinations", func(t *testing.T) {
				attrs := &v2.AttrGetDestinations{DestinationIDs: []string{}}
				if err := client.Call(context.Background(), utils.APIerSv2GetDestinations, attrs, &dests); err != nil {
					t.Error("Got error on APIerSv2.GetDestinations: ", err.Error())
				}
			})

			var rdID1, rdID2, rdID3, rdID4, rdID5, rdID6, rdID7 []string

			t.Run("GetReverseDestination", func(t *testing.T) {
				checkRD := func(t *testing.T, dst string) (rpl []string) {
					if err := client.Call(context.Background(), utils.APIerSv1GetReverseDestination, utils.StringPointer(dst), &rpl); err != nil {
						t.Errorf("Error dst <%s>, <%v>", dst, err)
					}
					return
				}
				rdID1 = checkRD(t, "1001")
				rdID2 = checkRD(t, "1002")
				rdID3 = checkRD(t, "+49151")
				rdID4 = checkRD(t, "077")
				rdID5 = checkRD(t, "10")
				rdID6 = checkRD(t, "+246")
				rdID7 = checkRD(t, "+135")
			})

			var actsMp map[string]engine.Actions
			t.Run("GetActions", func(t *testing.T) {
				if err := client.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{}, &actsMp); err != nil {
					t.Error(err)
				}
			})

			var atr engine.ActionTriggers

			t.Run("GetActionTriggers", func(t *testing.T) {
				var reply string
				if err := client.Call(context.Background(), utils.APIerSv1SetActionTrigger, engine.AttrSetActionTrigger{
					GroupID:  "GroupID",
					UniqueID: "ID",
					ActionTrigger: map[string]any{
						utils.ThresholdType:         "THR",
						utils.ThresholdValue:        10,
						utils.Recurrent:             false,
						utils.Executed:              false,
						utils.MinSleep:              time.Second,
						utils.ExpirationDate:        time.Now(),
						utils.ActivationDate:        time.Now(),
						utils.BalanceID:             "*default",
						utils.BalanceType:           "*call",
						utils.BalanceDestinationIds: []any{"DST1", "DST2"},
						utils.BalanceWeight:         10,
						utils.BalanceExpirationDate: time.Now(),
						utils.BalanceTimingTags:     []string{"*asap"},
						utils.BalanceCategories:     []string{utils.Call},
						utils.BalanceSharedGroups:   []string{"SHRGroup"},
						utils.BalanceBlocker:        true,
						utils.ActionsID:             "ACT1",
						utils.MinQueuedItems:        5,
					},
				}, &reply); err != nil {
					t.Error(err)
				} else if reply != utils.OK {
					t.Errorf("Calling v1.SetActionTrigger got: %v", reply)
				}
				if err := client.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atr); err != nil {
					t.Error(err)
				}
			})

			var aps []*engine.ActionPlan
			t.Run("GetActionPlan", func(t *testing.T) {
				var reply string
				atms1 := &engine.AttrSetActionPlan{
					Id: "ATMS_1",
					ActionPlan: []*engine.AttrActionPlan{
						{
							ActionsId: "ACTION_TOPUP_RESET_SMS",
							MonthDays: "1",
							Time:      "00:00:00",
							Weight:    20.0},
					},
				}
				if err := client.Call(context.Background(), utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
					t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
				} else if reply != utils.OK {
					t.Errorf("Unexpected reply returned: %s", reply)
				}
				if err := client.Call(context.Background(), utils.APIerSv1GetActionPlan,
					&v1.AttrGetActionPlan{ID: "ATMS_1"}, &aps); err != nil {
					t.Error(err)
				}
			})

			t.Run("CheckDataDBBackupFolder1", func(t *testing.T) { // make sure its empty
				var dirs int
				if err := filepath.Walk(dataDBBackupPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if info.Name() != "customBackupPath" && info.Name() != "datadb" {
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

			t.Run("CheckStorDBBackupFolder1", func(t *testing.T) { // make sure its empty
				var dirs int
				if err := filepath.Walk(storDBBackupPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						if info.Name() != "customBackupPath" && info.Name() != "stordb" {
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

			datadbDumpDirs1 := []string{}
			stordbDumpDirs1 := []string{}
			storDBDumpFileNames := []string{}

			t.Run("CheckDataDBDumpFolder1", func(t *testing.T) {
				if i == 1 || i == 3 {
					time.Sleep(500 * time.Millisecond) // make sure to wait for intervals to kick in
				}
				// time.Sleep(5 * time.Second)
				var dirs, files int
				if err := filepath.Walk(cfg.DataDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						datadbDumpDirs1 = append(datadbDumpDirs1, info.Name())
						dirs++
					} else {
						datadbDumpDirs1 = append(datadbDumpDirs1, info.Name())
						files++
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
				// slices.Sort(datadbDumpDirs1)
				for _, ddbn := range dataDBDirNames {
					if !slices.Contains(datadbDumpDirs1, ddbn) {
						t.Errorf("expected <%v>, received <%v>", dataDBDirNames, ddbn)
					}
				}
				if dirs != 43 {
					t.Errorf("expected <%d> dirs, received <%d>", 43, dirs)
				} else if files != 43 && files != 44 && files != 45 {
					t.Errorf("expected <43> or <44> or <45> files, received <%d>", files)
				}
			})

			t.Run("CheckStorDBDumpFolder1", func(t *testing.T) {
				var dirs, files int
				if err := filepath.Walk(cfg.StorDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						stordbDumpDirs1 = append(stordbDumpDirs1, info.Name())
						dirs++
					} else {
						storDBDumpFileNames = append(storDBDumpFileNames, info.Name())
						files++
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
				slices.Sort(stordbDumpDirs1)
				if !reflect.DeepEqual(storDBDirNames, stordbDumpDirs1) {
					t.Errorf("expected <%v>, received <%v>", storDBDirNames, stordbDumpDirs1)
				} else if files != 27 {
					t.Errorf("expected <%d> files, received <%d>", 27, files)
				}
			})

			t.Run("RunSnapshotDataDB", func(t *testing.T) {
				var rply string
				bpath := ""
				zip := false
				if i == 3 {
					bpath = dataDBBackupPath
					zip = true
				}
				if err := client.Call(context.Background(), utils.APIerSv1SnapshotDataDB, &v1.BackupParams{BackupFolderPath: bpath, Zip: zip}, &rply); err != nil {
					t.Fatal(err)
				}

			})

			t.Run("RunSnapshotStorDB", func(t *testing.T) {
				bpath := ""
				zip := false
				if i == 3 {
					bpath = storDBBackupPath
					zip = true
				}
				var rply string
				if err := client.Call(context.Background(), utils.APIerSv1SnapshotStorDB, &v1.BackupParams{BackupFolderPath: bpath, Zip: zip}, &rply); err != nil {
					t.Fatal(err)
				}
			})

			t.Run("CheckDataDBBackupFolder2", func(t *testing.T) { // make sure its populated correctly and same as dump folder
				backupDirs := []string{}
				dumpDirsCopy := datadbDumpDirs1
				foundBackupFolder := false
				if err := filepath.Walk(dataDBBackupPath, func(_ string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					backupDirs = append(backupDirs, info.Name())
					if strings.HasPrefix(info.Name(), "backup_") {
						if !foundBackupFolder {
							if i == 3 {
								dumpDirsCopy = append(dumpDirsCopy, info.Name())
							} else { // add backup string to the slice for simplicity in comparing
								dumpDirsCopy = append(dumpDirsCopy, "") // add it in the expected position pushing the rest of the elements to the right
								copy(dumpDirsCopy[2:], dumpDirsCopy[1:])
								dumpDirsCopy[1] = info.Name()
							}
							foundBackupFolder = true
						} else {
							t.Errorf("found more than 1 backup folder <%v>", backupDirs)
						}
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
				if i == 3 {
					dumpDirsCopy = []string{"customBackupPath"} // no folders should be in the same root as backuppath for zipped backup
					if !reflect.DeepEqual(backupDirs[0], dumpDirsCopy[0]) {
						t.Errorf("expected <%v>, received <%v>", dumpDirsCopy[0], backupDirs[0])
					}
					if len(backupDirs) != 2 {
						t.Errorf("expected 2 directories 1 backup zip file and one the forlder itself, received <%v>", backupDirs)
					} else if !strings.HasPrefix(backupDirs[1], "backup_") || !strings.HasSuffix(backupDirs[1], ".zip") {
						t.Errorf("expected backup zip file, received <%v>", backupDirs)
					}
				} else {
					if !reflect.DeepEqual(backupDirs, dumpDirsCopy) {
						dumpDirsCopy[31] = backupDirs[31]
						if !reflect.DeepEqual(backupDirs, dumpDirsCopy) {
							t.Errorf("expected <%v>, received <%v>", dumpDirsCopy, backupDirs)
						}
					} else if !reflect.DeepEqual(dumpDirsCopy, backupDirs) {
						t.Errorf("expected <%v>, received <%v>", dumpDirsCopy, backupDirs)
					}
				}
			})

			t.Run("CheckStorDBBackupFolder2", func(t *testing.T) { // make sure its populated correctly and same as dump folder
				backupDirs := []string{}
				backupFileNames := []string{}
				dumpDirsCopy := stordbDumpDirs1
				foundBackupFolder := false
				var dirs, files int
				if err := filepath.Walk(storDBBackupPath, func(_ string, info os.FileInfo, err error) error {
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
					} else if !reflect.DeepEqual(storDBDumpFileNames, backupFileNames) { // make sure it was a perfect file copy
						t.Errorf("expected <%v>, received <%v>", storDBDumpFileNames, backupFileNames)
					}
				}
			})

			t.Run("CheckDataDBDumpFolder2", func(t *testing.T) { // make sure new files are added
				snapshotDirs := []string{}
				snapshotFileNames := []string{}
				time.Sleep(50 * time.Millisecond)
				var dirs, files int
				if err := filepath.Walk(cfg.DataDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
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
				if !reflect.DeepEqual(dataDBDirNames, snapshotDirs) {
					t.Errorf("expected <%v>, received <%v>", dataDBDirNames, snapshotDirs)
				} else if files != 42 && files != 43 && files != 44 { // depending on when rewriting is done
					t.Errorf("expected 42, 43 or 44 files, received <%d>", files)
				}
				for _, sfn := range snapshotFileNames {
					// make sure file names are not the same as discarded(backed up) dump files
					if sfn != "0Rewrite0" && sfn != "datadb" && slices.Contains(datadbDumpDirs1, sfn) {
						t.Fatalf("expected snapshot file names to be different from backed up dump file names, snapshot files <%v>, \ndump files <%v>", snapshotFileNames, datadbDumpDirs1)
					}
				}
			})

			t.Run("CheckStorDBDumpFolder2", func(t *testing.T) { // make sure new files are added
				snapshotDirs := []string{}
				snapshotFileNames := []string{}
				time.Sleep(50 * time.Millisecond)
				var dirs, files int
				if err := filepath.Walk(cfg.StorDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
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
				if !reflect.DeepEqual(storDBDirNames, snapshotDirs) {
					t.Errorf("expected <%v>, received <%v>", storDBDirNames, snapshotDirs)
				} else if files != 27 { // expected 1 less file then dump since rewriting happens on dumps *versions
					t.Errorf("expected <%d> files, received <%d>", 27, files)
				}
				for _, sfn := range snapshotFileNames {
					// make sure file names are not the same as discarded(backed up) dump files
					if sfn != "stordb" && slices.Contains(storDBDumpFileNames, sfn) {
						t.Fatalf("expected snapshot file names to be different from backed up dump file names, snapshot files <%v>, \ndump files <%v>", snapshotFileNames, storDBDumpFileNames)
					}
				}
			})

			t.Run("GetAccounts2", func(t *testing.T) {
				var acnts2 []*engine.Account
				if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
					&utils.AttrGetAccounts{
						Tenant: "cgrates.org",
					}, &acnts2); err != nil {
					t.Errorf("APIerSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(acnts2) != 2 {
					t.Fatalf("APIerSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
				}
				sort.Slice(acnts2, func(i, j int) bool {
					return acnts2[i].ID > acnts2[j].ID
				})
				if !reflect.DeepEqual(acnts2, acnts) {
					t.Errorf("Expected accounts to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acnts), utils.ToJSON(acnts2))
				}
			})

			t.Run("GetRatingPlans2", func(t *testing.T) {
				restoreRply := new(engine.RatingPlan)
				rplnId := "RP_TESTIT1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply.Ratings {
					if !reflect.DeepEqual(restoreRply.Ratings[rateId], ratingPl.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply.Ratings[rateId], ratingPl.Ratings[rateId])
					}
				}
				for rateId := range restoreRply.DestinationRates {
					if !reflect.DeepEqual(restoreRply.DestinationRates[rateId], ratingPl.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply.DestinationRates[rateId], ratingPl.DestinationRates[rateId])
					}
				}
				restoreRply2 := new(engine.RatingPlan)
				rplnId = "RP_SPECIAL_1002"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply2); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply2.Ratings {
					if !reflect.DeepEqual(restoreRply2.Ratings[rateId], ratingPl2.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply2.Ratings[rateId], ratingPl2.Ratings[rateId])
					}
				}
				for rateId := range restoreRply2.DestinationRates {
					if !reflect.DeepEqual(restoreRply2.DestinationRates[rateId], ratingPl2.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply2.DestinationRates[rateId], ratingPl2.DestinationRates[rateId])
					}
				}
				restoreRply3 := new(engine.RatingPlan)
				rplnId = "RP_RETAIL1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply3); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply3.Ratings {
					if !reflect.DeepEqual(restoreRply3.Ratings[rateId], ratingPl3.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply3.Ratings[rateId], ratingPl3.Ratings[rateId])
					}
				}
				for rateId := range restoreRply3.DestinationRates {
					if !reflect.DeepEqual(restoreRply3.DestinationRates[rateId], ratingPl3.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply3.DestinationRates[rateId], ratingPl3.DestinationRates[rateId])
					}
				}
				restoreRply4 := new(engine.RatingPlan)
				rplnId = "RP_ANY2CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply4); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply4.Ratings {
					if !reflect.DeepEqual(restoreRply4.Ratings[rateId], ratingPl4.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply4.Ratings[rateId], ratingPl4.Ratings[rateId])
					}
				}
				for rateId := range restoreRply4.DestinationRates {
					if !reflect.DeepEqual(restoreRply4.DestinationRates[rateId], ratingPl4.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply4.DestinationRates[rateId], ratingPl4.DestinationRates[rateId])
					}
				}
				restoreRply5 := new(engine.RatingPlan)
				rplnId = "RP_ANY1CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply5); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply5.Ratings {
					if !reflect.DeepEqual(restoreRply5.Ratings[rateId], ratingPl5.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply5.Ratings[rateId], ratingPl5.Ratings[rateId])
					}
				}
				for rateId := range restoreRply5.DestinationRates {
					if !reflect.DeepEqual(restoreRply5.DestinationRates[rateId], ratingPl5.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply5.DestinationRates[rateId], ratingPl5.DestinationRates[rateId])
					}
				}
				restoreRply6 := new(engine.RatingPlan)
				rplnId = "RP_TEST"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply6); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply6.Ratings {
					if !reflect.DeepEqual(restoreRply6.Ratings[rateId], ratingPl6.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply6.Ratings[rateId], ratingPl6.Ratings[rateId])
					}
				}
				for rateId := range restoreRply6.DestinationRates {
					if !reflect.DeepEqual(restoreRply6.DestinationRates[rateId], ratingPl6.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply6.DestinationRates[rateId], ratingPl6.DestinationRates[rateId])
					}
				}
				restoreRply7 := new(engine.RatingPlan)
				rplnId = "RP_MOBILE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply7); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply7.Ratings {
					if !reflect.DeepEqual(restoreRply7.Ratings[rateId], ratingPl7.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply7.Ratings[rateId], ratingPl7.Ratings[rateId])
					}
				}
				for rateId := range restoreRply7.DestinationRates {
					if !reflect.DeepEqual(restoreRply7.DestinationRates[rateId], ratingPl7.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply7.DestinationRates[rateId], ratingPl7.DestinationRates[rateId])
					}
				}
				restoreRply8 := new(engine.RatingPlan)
				rplnId = "RP_LOCAL"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply8); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply8.Ratings {
					if !reflect.DeepEqual(restoreRply8.Ratings[rateId], ratingPl8.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply8.Ratings[rateId], ratingPl8.Ratings[rateId])
					}
				}
				for rateId := range restoreRply8.DestinationRates {
					if !reflect.DeepEqual(restoreRply8.DestinationRates[rateId], ratingPl8.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply8.DestinationRates[rateId], ratingPl8.DestinationRates[rateId])
					}
				}
				restoreRply9 := new(engine.RatingPlan)
				rplnId = "RP_FREE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply9); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply9.Ratings {
					if !reflect.DeepEqual(restoreRply9.Ratings[rateId], ratingPl9.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply9.Ratings[rateId], ratingPl9.Ratings[rateId])
					}
				}
				for rateId := range restoreRply9.DestinationRates {
					if !reflect.DeepEqual(restoreRply9.DestinationRates[rateId], ratingPl9.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply9.DestinationRates[rateId], ratingPl9.DestinationRates[rateId])
					}
				}
				restoreRply10 := new(engine.RatingPlan)
				rplnId = "RP_ANY2CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply10); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply10.Ratings {
					if !reflect.DeepEqual(restoreRply10.Ratings[rateId], ratingPl10.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply10.Ratings[rateId], ratingPl10.Ratings[rateId])
					}
				}
				for rateId := range restoreRply10.DestinationRates {
					if !reflect.DeepEqual(restoreRply10.DestinationRates[rateId], ratingPl10.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply10.DestinationRates[rateId], ratingPl10.DestinationRates[rateId])
					}
				}
				restoreRply11 := new(engine.RatingPlan)
				rplnId = "RP_ANY1CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply11); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply11.Ratings {
					if !reflect.DeepEqual(restoreRply11.Ratings[rateId], ratingPl11.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply11.Ratings[rateId], ratingPl11.Ratings[rateId])
					}
				}
				for rateId := range restoreRply11.DestinationRates {
					if !reflect.DeepEqual(restoreRply11.DestinationRates[rateId], ratingPl11.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply11.DestinationRates[rateId], ratingPl11.DestinationRates[rateId])
					}
				}
			})
			t.Run("GetRatingProfiles2", func(t *testing.T) {
				var rcvRprof engine.RatingProfile
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingProfile,
					&utils.AttrGetRatingProfile{
						Tenant:   "cgrates.org",
						Category: "free",
						Subject:  "RP_FREE",
					}, &rcvRprof); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(rcvRprof, rProf) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rProf, rcvRprof)
				}
			})

			t.Run("GetDestinations2", func(t *testing.T) {
				sort.Slice(dests, func(i, j int) bool {
					return dests[i].Id < dests[j].Id
				})
				var rcv []*engine.Destination
				attrs := &v2.AttrGetDestinations{DestinationIDs: []string{}}
				if err := client.Call(context.Background(), utils.APIerSv2GetDestinations, attrs, &rcv); err != nil {
					t.Error("Got error on APIerSv2.GetDestinations: ", err.Error())
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].Id < rcv[j].Id
				})
				if !reflect.DeepEqual(dests, rcv) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", dests, rcv)
				}
			})

			t.Run("GetReverseDestination2", func(t *testing.T) {
				checkRD := func(t *testing.T, dst string) (rpl []string) {
					if err := client.Call(context.Background(), utils.APIerSv1GetReverseDestination, utils.StringPointer(dst), &rpl); err != nil {
						t.Errorf("Error dst <%s>, <%v>", dst, err)
					}
					return
				}
				rcvRdID1 := checkRD(t, "1001")
				rcvRdID2 := checkRD(t, "1002")
				rcvRdID3 := checkRD(t, "+49151")
				rcvRdID4 := checkRD(t, "077")
				rcvRdID5 := checkRD(t, "10")
				rcvRdID6 := checkRD(t, "+246")
				rcvRdID7 := checkRD(t, "+135")
				if !reflect.DeepEqual(rdID1, rcvRdID1) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID1, rcvRdID1)
				}
				if !reflect.DeepEqual(rdID2, rcvRdID2) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID2, rcvRdID2)
				}
				if !reflect.DeepEqual(rdID3, rcvRdID3) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID3, rcvRdID3)
				}
				if !reflect.DeepEqual(rdID4, rcvRdID4) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID4, rcvRdID4)
				}
				if !reflect.DeepEqual(rdID5, rcvRdID5) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID5, rcvRdID5)
				}
				if !reflect.DeepEqual(rdID6, rcvRdID6) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID6, rcvRdID6)
				}
				if !reflect.DeepEqual(rdID7, rcvRdID7) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID7, rcvRdID7)
				}
			})

			t.Run("GetActions2", func(t *testing.T) {
				var rcv map[string]engine.Actions
				if err := client.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{}, &rcv); err != nil {
					t.Error(err)
				}
				if len(actsMp) != len(rcv) {
					t.Errorf("Length of expected <%v>, doesnt match <%v>", len(actsMp), len(rcv))
				}

				for id, acts := range actsMp {
					if len(acts) != len(rcv[id]) {
						t.Errorf("Length of expected <%v>, doesnt match <%v>", len(acts), len(rcv[id]))
					}
					for i, act := range acts {
						if rcv[id][i].Balance.Blocker == nil {
							rcv[id][i].Balance.Blocker = utils.BoolPointer(false)
						}
						if rcv[id][i].Balance.Disabled == nil {
							rcv[id][i].Balance.Disabled = utils.BoolPointer(false)
						}
						if rcv[id][i].Balance.ExpirationDate == nil {
							rcv[id][i].Balance.ExpirationDate = act.Balance.ExpirationDate
						}
						if !reflect.DeepEqual(utils.ToJSON(act), utils.ToJSON(rcv[id][i])) {
							t.Errorf("expected <%+v>, \nreceived <%+v>", utils.ToJSON(act), utils.ToJSON(rcv[id][i]))
						}
					}
				}

			})

			t.Run("GetActionTriggers2", func(t *testing.T) {
				var rcv engine.ActionTriggers
				if err := client.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{}}, &rcv); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(atr, rcv) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", atr, rcv)
				}
			})

			t.Run("GetActionPlan2", func(t *testing.T) {
				var rcv []*engine.ActionPlan
				if err := client.Call(context.Background(), utils.APIerSv1GetActionPlan,
					&v1.AttrGetActionPlan{ID: "ATMS_1"}, &rcv); err != nil {
					t.Error(err)
				}
				if len(aps) != 1 || len(rcv) != 1 {
					t.Errorf("expected aps len 1, got <%v>, expected rcv len 1, got <%v>", len(aps), len(rcv))
				}
				if !reflect.DeepEqual(aps[0].Id, rcv[0].Id) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", aps[0].Id, rcv[0].Id)
				}
				for id, actts := range aps[0].ActionTimings {
					if !reflect.DeepEqual(actts.ActionsID, rcv[0].ActionTimings[id].ActionsID) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.ActionsID, rcv[0].ActionTimings[id].ActionsID)
					}
					if !reflect.DeepEqual(actts.Uuid, rcv[0].ActionTimings[id].Uuid) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.Uuid, rcv[0].ActionTimings[id].Uuid)
					}
					if !reflect.DeepEqual(actts.ExtraData, rcv[0].ActionTimings[id].ExtraData) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.ExtraData, rcv[0].ActionTimings[id].ExtraData)
					}
					if !reflect.DeepEqual(actts.Weight, rcv[0].ActionTimings[id].Weight) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.Weight, rcv[0].ActionTimings[id].Weight)
					}
				}
			})
			t.Run("EngineShutdown", func(t *testing.T) {
				if err := engine.KillEngine(100); err != nil {
					t.Error(err)
				}
			})
			t.Run("CountDataDBFiles", func(t *testing.T) {
				time.Sleep(100 * time.Millisecond) // wait for engine to shutdown
				var dirs, files int
				if err := filepath.Walk(cfg.DataDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
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
				} else if dirs != 43 {
					t.Errorf("expected <%d> directories, received <%d>", 43, dirs)
				} else if (i == 0 && files != 28 && files != 29) ||
					(i == 2 && files != 28 && files != 29) {
					t.Errorf("expected 28 or 29 files, received <%d>", files)
				} else if (i == 1 && files != 28 && files != 30 && files != 29) ||
					(i == 3 && files != 28 && files != 30 && files != 29) {
					t.Errorf("expected 28 or 29 or 30 files, received <%d>", files)
				}
			})

			t.Run("CountStorDBFiles", func(t *testing.T) {
				var dirs, files int
				if err := filepath.Walk(cfg.StorDbCfg().Opts.InternalDBDumpPath, func(_ string, info os.FileInfo, err error) error {
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
				} else if dirs != 28 {
					t.Errorf("expected <%d> directories, received <%d>", 28, dirs)
				} else if files != 1 {
					t.Errorf("expected <%d> files, received <%d>", 1, files)
				}
			})

			ng.PreserveDataDB = true
			ng.PreserveStorDB = true
			client, cfg = ng.Run(t)
			time.Sleep(100 * time.Millisecond)
			t.Run("ExportDataDB2", func(t *testing.T) {
				var reply string
				if err := client.Call(context.Background(), utils.APIerSv1ExportToFolder, utils.ArgExportToFolder{Path: "/tmp/ExportPath2"}, &reply); err != nil {
					t.Error(err)
				} else if reply != utils.OK {
					t.Errorf("Expected: <%v>, received: <%v>", utils.OK, reply)
				}
			})

			t.Run("CompareExports", func(t *testing.T) {
				readLines := func(filePath string) ([]string, error) {
					file, err := os.Open(filePath)
					if err != nil {
						return nil, fmt.Errorf("error opening file %s: %v", filePath, err)
					}
					defer file.Close()

					var lines []string
					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						lines = append(lines, scanner.Text())
					}
					if err := scanner.Err(); err != nil {
						return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
					}
					sort.Strings(lines)
					return lines, nil
				}
				if err := filepath.Walk("/tmp/ExportPath1", func(path1 string, info1 os.FileInfo, err1 error) error {
					if err1 != nil {
						return err1
					}

					if (i == 1 && path1 == "/tmp/ExportPath1/Chargers.csv") || (i == 3 && path1 == "/tmp/ExportPath1/Chargers.csv") {
						return filepath.SkipDir // chargers have expired before 2nd export
					}
					relPath, err := filepath.Rel("/tmp/ExportPath1", path1) // save path that comes after /tmp/ExportPath1
					if err != nil {
						return fmt.Errorf("error calculating relative path: %v", err)
					}
					path2 := filepath.Join("/tmp/ExportPath2", relPath)
					if _, err := os.Stat(path2); err != nil {
						return err
					}
					if info1.Mode().IsRegular() {
						lines1, err := readLines(path1)
						if err != nil {
							return err
						}
						lines2, err := readLines(path2)
						if err != nil {
							return err
						}
						if len(lines1) != len(lines2) {
							return fmt.Errorf("Line count doesnt match: <%v> \n\n<%v>", lines1, lines2)
						}

						for i := range lines1 {
							if lines1[i] != lines2[i] {
								return fmt.Errorf("Files differ: %v <%v> \nand \n%v <%v>", path1, lines1[i], path2, lines2[i])
							}
						}
					}
					return nil
				}); err != nil {
					t.Error(err)
				}
			})

			t.Run("GetRatingPlans4", func(t *testing.T) {
				restoreRply := new(engine.RatingPlan)
				rplnId := "RP_TESTIT1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply.Ratings {
					if !reflect.DeepEqual(restoreRply.Ratings[rateId], ratingPl.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply.Ratings[rateId], ratingPl.Ratings[rateId])
					}
				}
				for rateId := range restoreRply.DestinationRates {
					if !reflect.DeepEqual(restoreRply.DestinationRates[rateId], ratingPl.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply.DestinationRates[rateId], ratingPl.DestinationRates[rateId])
					}
				}
				restoreRply2 := new(engine.RatingPlan)
				rplnId = "RP_SPECIAL_1002"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply2); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply2.Ratings {
					if !reflect.DeepEqual(restoreRply2.Ratings[rateId], ratingPl2.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply2.Ratings[rateId], ratingPl2.Ratings[rateId])
					}
				}
				for rateId := range restoreRply2.DestinationRates {
					if !reflect.DeepEqual(restoreRply2.DestinationRates[rateId], ratingPl2.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply2.DestinationRates[rateId], ratingPl2.DestinationRates[rateId])
					}
				}
				restoreRply3 := new(engine.RatingPlan)
				rplnId = "RP_RETAIL1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply3); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply3.Ratings {
					if !reflect.DeepEqual(restoreRply3.Ratings[rateId], ratingPl3.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply3.Ratings[rateId], ratingPl3.Ratings[rateId])
					}
				}
				for rateId := range restoreRply3.DestinationRates {
					if !reflect.DeepEqual(restoreRply3.DestinationRates[rateId], ratingPl3.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply3.DestinationRates[rateId], ratingPl3.DestinationRates[rateId])
					}
				}
				restoreRply4 := new(engine.RatingPlan)
				rplnId = "RP_ANY2CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply4); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply4.Ratings {
					if !reflect.DeepEqual(restoreRply4.Ratings[rateId], ratingPl4.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply4.Ratings[rateId], ratingPl4.Ratings[rateId])
					}
				}
				for rateId := range restoreRply4.DestinationRates {
					if !reflect.DeepEqual(restoreRply4.DestinationRates[rateId], ratingPl4.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply4.DestinationRates[rateId], ratingPl4.DestinationRates[rateId])
					}
				}
				restoreRply5 := new(engine.RatingPlan)
				rplnId = "RP_ANY1CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply5); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply5.Ratings {
					if !reflect.DeepEqual(restoreRply5.Ratings[rateId], ratingPl5.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply5.Ratings[rateId], ratingPl5.Ratings[rateId])
					}
				}
				for rateId := range restoreRply5.DestinationRates {
					if !reflect.DeepEqual(restoreRply5.DestinationRates[rateId], ratingPl5.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply5.DestinationRates[rateId], ratingPl5.DestinationRates[rateId])
					}
				}
				restoreRply6 := new(engine.RatingPlan)
				rplnId = "RP_TEST"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply6); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply6.Ratings {
					if !reflect.DeepEqual(restoreRply6.Ratings[rateId], ratingPl6.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply6.Ratings[rateId], ratingPl6.Ratings[rateId])
					}
				}
				for rateId := range restoreRply6.DestinationRates {
					if !reflect.DeepEqual(restoreRply6.DestinationRates[rateId], ratingPl6.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply6.DestinationRates[rateId], ratingPl6.DestinationRates[rateId])
					}
				}
				restoreRply7 := new(engine.RatingPlan)
				rplnId = "RP_MOBILE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply7); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply7.Ratings {
					if !reflect.DeepEqual(restoreRply7.Ratings[rateId], ratingPl7.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply7.Ratings[rateId], ratingPl7.Ratings[rateId])
					}
				}
				for rateId := range restoreRply7.DestinationRates {
					if !reflect.DeepEqual(restoreRply7.DestinationRates[rateId], ratingPl7.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply7.DestinationRates[rateId], ratingPl7.DestinationRates[rateId])
					}
				}
				restoreRply8 := new(engine.RatingPlan)
				rplnId = "RP_LOCAL"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply8); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply8.Ratings {
					if !reflect.DeepEqual(restoreRply8.Ratings[rateId], ratingPl8.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply8.Ratings[rateId], ratingPl8.Ratings[rateId])
					}
				}
				for rateId := range restoreRply8.DestinationRates {
					if !reflect.DeepEqual(restoreRply8.DestinationRates[rateId], ratingPl8.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply8.DestinationRates[rateId], ratingPl8.DestinationRates[rateId])
					}
				}
				restoreRply9 := new(engine.RatingPlan)
				rplnId = "RP_FREE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply9); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply9.Ratings {
					if !reflect.DeepEqual(restoreRply9.Ratings[rateId], ratingPl9.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply9.Ratings[rateId], ratingPl9.Ratings[rateId])
					}
				}
				for rateId := range restoreRply9.DestinationRates {
					if !reflect.DeepEqual(restoreRply9.DestinationRates[rateId], ratingPl9.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply9.DestinationRates[rateId], ratingPl9.DestinationRates[rateId])
					}
				}
				restoreRply10 := new(engine.RatingPlan)
				rplnId = "RP_ANY2CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply10); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply10.Ratings {
					if !reflect.DeepEqual(restoreRply10.Ratings[rateId], ratingPl10.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply10.Ratings[rateId], ratingPl10.Ratings[rateId])
					}
				}
				for rateId := range restoreRply10.DestinationRates {
					if !reflect.DeepEqual(restoreRply10.DestinationRates[rateId], ratingPl10.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply10.DestinationRates[rateId], ratingPl10.DestinationRates[rateId])
					}
				}
				restoreRply11 := new(engine.RatingPlan)
				rplnId = "RP_ANY1CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply11); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply11.Ratings {
					if !reflect.DeepEqual(restoreRply11.Ratings[rateId], ratingPl11.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply11.Ratings[rateId], ratingPl11.Ratings[rateId])
					}
				}
				for rateId := range restoreRply11.DestinationRates {
					if !reflect.DeepEqual(restoreRply11.DestinationRates[rateId], ratingPl11.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply11.DestinationRates[rateId], ratingPl11.DestinationRates[rateId])
					}
				}
			})
			t.Run("GetRatingProfiles4", func(t *testing.T) {
				var rcvRprof engine.RatingProfile
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingProfile,
					&utils.AttrGetRatingProfile{
						Tenant:   "cgrates.org",
						Category: "free",
						Subject:  "RP_FREE",
					}, &rcvRprof); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(rcvRprof, rProf) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rProf, rcvRprof)
				}
			})

			t.Run("GetDestinations4", func(t *testing.T) {
				sort.Slice(dests, func(i, j int) bool {
					return dests[i].Id < dests[j].Id
				})
				var rcv []*engine.Destination
				attrs := &v2.AttrGetDestinations{DestinationIDs: []string{}}
				if err := client.Call(context.Background(), utils.APIerSv2GetDestinations, attrs, &rcv); err != nil {
					t.Error("Got error on APIerSv2.GetDestinations: ", err.Error())
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].Id < rcv[j].Id
				})
				if !reflect.DeepEqual(dests, rcv) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", dests, rcv)
				}
			})

			t.Run("GetReverseDestination4", func(t *testing.T) {
				checkRD := func(t *testing.T, dst string) (rpl []string) {
					if err := client.Call(context.Background(), utils.APIerSv1GetReverseDestination, utils.StringPointer(dst), &rpl); err != nil {
						t.Errorf("Error dst <%s>, <%v>", dst, err)
					}
					return
				}
				rcvRdID1 := checkRD(t, "1001")
				rcvRdID2 := checkRD(t, "1002")
				rcvRdID3 := checkRD(t, "+49151")
				rcvRdID4 := checkRD(t, "077")
				rcvRdID5 := checkRD(t, "10")
				rcvRdID6 := checkRD(t, "+246")
				rcvRdID7 := checkRD(t, "+135")
				if !reflect.DeepEqual(rdID1, rcvRdID1) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID1, rcvRdID1)
				}
				if !reflect.DeepEqual(rdID2, rcvRdID2) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID2, rcvRdID2)
				}
				if !reflect.DeepEqual(rdID3, rcvRdID3) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID3, rcvRdID3)
				}
				if !reflect.DeepEqual(rdID4, rcvRdID4) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID4, rcvRdID4)
				}
				if !reflect.DeepEqual(rdID5, rcvRdID5) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID5, rcvRdID5)
				}
				if !reflect.DeepEqual(rdID6, rcvRdID6) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID6, rcvRdID6)
				}
				if !reflect.DeepEqual(rdID7, rcvRdID7) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID7, rcvRdID7)
				}
			})

			t.Run("GetActions4", func(t *testing.T) {
				var rcv map[string]engine.Actions
				if err := client.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{}, &rcv); err != nil {
					t.Error(err)
				}
				if len(actsMp) != len(rcv) {
					t.Errorf("Length of expected <%v>, doesnt match <%v>", len(actsMp), len(rcv))
				}

				for id, acts := range actsMp {
					if len(acts) != len(rcv[id]) {
						t.Errorf("Length of expected <%v>, doesnt match <%v>", len(acts), len(rcv[id]))
					}
					for i, act := range acts {
						if rcv[id][i].Balance.Blocker == nil {
							rcv[id][i].Balance.Blocker = utils.BoolPointer(false)
						}
						if rcv[id][i].Balance.Disabled == nil {
							rcv[id][i].Balance.Disabled = utils.BoolPointer(false)
						}
						if rcv[id][i].Balance.ExpirationDate == nil {
							rcv[id][i].Balance.ExpirationDate = act.Balance.ExpirationDate
						}
						if !reflect.DeepEqual(utils.ToJSON(act), utils.ToJSON(rcv[id][i])) {
							t.Errorf("expected <%+v>, \nreceived <%+v>", utils.ToJSON(act), utils.ToJSON(rcv[id][i]))
						}
					}
				}

			})

			t.Run("GetActionTriggers4", func(t *testing.T) {
				var rcv engine.ActionTriggers
				if err := client.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{}}, &rcv); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(atr, rcv) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", atr, rcv)
				}
			})

			t.Run("GetActionPlan4", func(t *testing.T) {
				var rcv []*engine.ActionPlan
				if err := client.Call(context.Background(), utils.APIerSv1GetActionPlan,
					&v1.AttrGetActionPlan{ID: "ATMS_1"}, &rcv); err != nil {
					t.Error(err)
				}
				if len(aps) != 1 || len(rcv) != 1 {
					t.Errorf("expected aps len 1, got <%v>, expected rcv len 1, got <%v>", len(aps), len(rcv))
				}
				if !reflect.DeepEqual(aps[0].Id, rcv[0].Id) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", aps[0].Id, rcv[0].Id)
				}
				for id, actts := range aps[0].ActionTimings {
					if !reflect.DeepEqual(actts.ActionsID, rcv[0].ActionTimings[id].ActionsID) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.ActionsID, rcv[0].ActionTimings[id].ActionsID)
					}
					if !reflect.DeepEqual(actts.Uuid, rcv[0].ActionTimings[id].Uuid) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.Uuid, rcv[0].ActionTimings[id].Uuid)
					}
					if !reflect.DeepEqual(actts.ExtraData, rcv[0].ActionTimings[id].ExtraData) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.ExtraData, rcv[0].ActionTimings[id].ExtraData)
					}
					if !reflect.DeepEqual(actts.Weight, rcv[0].ActionTimings[id].Weight) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.Weight, rcv[0].ActionTimings[id].Weight)
					}
				}
			})

			t.Run("GetAccountsAndRemove", func(t *testing.T) {
				var acnts2 []*engine.Account
				if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
					&utils.AttrGetAccounts{
						Tenant: "cgrates.org",
					}, &acnts2); err != nil {
					t.Errorf("APIerSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(acnts2) != 2 {
					t.Fatalf("APIerSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
				}
				sort.Slice(acnts2, func(i, j int) bool {
					return acnts2[i].ID > acnts2[j].ID
				})
				if !reflect.DeepEqual(acnts2, acnts) {
					t.Errorf("Expected accounts to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acnts), utils.ToJSON(acnts2))
				}
				for _, acnt := range acnts2 {
					args := &utils.AttrRemoveAccount{
						Account: acnt.ID,
					}
					var reply string
					if err := client.Call(context.Background(), utils.APIerSv1RemoveAccount,
						args,
						&reply); err != nil {
						t.Error(err)
					} else if reply != utils.OK {
						t.Error("Unexpected reply returned", reply)
					}
				}
			})

			var newAcc *utils.AttrSetAccount

			t.Run("SetNewAccount", func(t *testing.T) {
				newAcc = &utils.AttrSetAccount{
					Tenant:  "cgrates.org",
					Account: "AccAfterSnapshot",
				}
				var replySet string
				if err := client.Call(context.Background(), utils.APIerSv1SetAccount,
					newAcc, &replySet); err != nil {
					t.Error(err)
				}
			})

			t.Run("RestoreFirstDataDBBackup", func(t *testing.T) {
				bpath := ""
				if i == 3 {
					bpath = dataDBBackupPath
				}
				if i == 4 {
					entries, err := os.ReadDir(dataDBBackupPath)
					if err != nil {
						t.Fatal(err)
					}
					if len(entries) != 1 {
						t.Errorf("expected 1 entry, received <%v>", entries)
					}
					bpath = filepath.Join(dataDBBackupPath, entries[0].Name())
				}
				var rply string
				if err := client.Call(context.Background(), utils.APIerSv1RestoreDataDB, bpath, &rply); err != nil {
					t.Fatal(err)
				}
			})

			t.Run("RestoreFirstStorDBBackup", func(t *testing.T) {
				bpath := ""
				if i == 3 {
					bpath = storDBBackupPath
				}
				if i == 4 {
					entries, err := os.ReadDir(storDBBackupPath)
					if err != nil {
						t.Fatal(err)
					}
					if len(entries) != 1 {
						t.Errorf("expected 1 entry, received <%v>", entries)
					}
					bpath = filepath.Join(storDBBackupPath, entries[0].Name())
				}
				var rply string
				if err := client.Call(context.Background(), utils.APIerSv1RestoreStorDB, bpath, &rply); err != nil {
					t.Fatal(err)
				}
			})

			t.Run("CheckNewAccount", func(t *testing.T) {
				var acntGot engine.Account
				if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
					&utils.AttrGetAccount{
						Tenant:  "cgrates.org",
						Account: "AccAfterSnapshot",
					}, &acntGot); err == nil || err.Error() != "NOT_FOUND" {
					t.Errorf("expected <%v>, received <%v>", "NOT_FOUND", err)
				}
			})

			t.Run("GetRatingPlans4", func(t *testing.T) {
				restoreRply := new(engine.RatingPlan)
				rplnId := "RP_TESTIT1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply.Ratings {
					if !reflect.DeepEqual(restoreRply.Ratings[rateId], ratingPl.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply.Ratings[rateId], ratingPl.Ratings[rateId])
					}
				}
				for rateId := range restoreRply.DestinationRates {
					if !reflect.DeepEqual(restoreRply.DestinationRates[rateId], ratingPl.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply.DestinationRates[rateId], ratingPl.DestinationRates[rateId])
					}
				}
				restoreRply2 := new(engine.RatingPlan)
				rplnId = "RP_SPECIAL_1002"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply2); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply2.Ratings {
					if !reflect.DeepEqual(restoreRply2.Ratings[rateId], ratingPl2.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply2.Ratings[rateId], ratingPl2.Ratings[rateId])
					}
				}
				for rateId := range restoreRply2.DestinationRates {
					if !reflect.DeepEqual(restoreRply2.DestinationRates[rateId], ratingPl2.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply2.DestinationRates[rateId], ratingPl2.DestinationRates[rateId])
					}
				}
				restoreRply3 := new(engine.RatingPlan)
				rplnId = "RP_RETAIL1"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply3); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply3.Ratings {
					if !reflect.DeepEqual(restoreRply3.Ratings[rateId], ratingPl3.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply3.Ratings[rateId], ratingPl3.Ratings[rateId])
					}
				}
				for rateId := range restoreRply3.DestinationRates {
					if !reflect.DeepEqual(restoreRply3.DestinationRates[rateId], ratingPl3.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply3.DestinationRates[rateId], ratingPl3.DestinationRates[rateId])
					}
				}
				restoreRply4 := new(engine.RatingPlan)
				rplnId = "RP_ANY2CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply4); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply4.Ratings {
					if !reflect.DeepEqual(restoreRply4.Ratings[rateId], ratingPl4.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply4.Ratings[rateId], ratingPl4.Ratings[rateId])
					}
				}
				for rateId := range restoreRply4.DestinationRates {
					if !reflect.DeepEqual(restoreRply4.DestinationRates[rateId], ratingPl4.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply4.DestinationRates[rateId], ratingPl4.DestinationRates[rateId])
					}
				}
				restoreRply5 := new(engine.RatingPlan)
				rplnId = "RP_ANY1CNT"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply5); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply5.Ratings {
					if !reflect.DeepEqual(restoreRply5.Ratings[rateId], ratingPl5.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply5.Ratings[rateId], ratingPl5.Ratings[rateId])
					}
				}
				for rateId := range restoreRply5.DestinationRates {
					if !reflect.DeepEqual(restoreRply5.DestinationRates[rateId], ratingPl5.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply5.DestinationRates[rateId], ratingPl5.DestinationRates[rateId])
					}
				}
				restoreRply6 := new(engine.RatingPlan)
				rplnId = "RP_TEST"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply6); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply6.Ratings {
					if !reflect.DeepEqual(restoreRply6.Ratings[rateId], ratingPl6.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply6.Ratings[rateId], ratingPl6.Ratings[rateId])
					}
				}
				for rateId := range restoreRply6.DestinationRates {
					if !reflect.DeepEqual(restoreRply6.DestinationRates[rateId], ratingPl6.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply6.DestinationRates[rateId], ratingPl6.DestinationRates[rateId])
					}
				}
				restoreRply7 := new(engine.RatingPlan)
				rplnId = "RP_MOBILE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply7); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply7.Ratings {
					if !reflect.DeepEqual(restoreRply7.Ratings[rateId], ratingPl7.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply7.Ratings[rateId], ratingPl7.Ratings[rateId])
					}
				}
				for rateId := range restoreRply7.DestinationRates {
					if !reflect.DeepEqual(restoreRply7.DestinationRates[rateId], ratingPl7.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply7.DestinationRates[rateId], ratingPl7.DestinationRates[rateId])
					}
				}
				restoreRply8 := new(engine.RatingPlan)
				rplnId = "RP_LOCAL"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply8); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply8.Ratings {
					if !reflect.DeepEqual(restoreRply8.Ratings[rateId], ratingPl8.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply8.Ratings[rateId], ratingPl8.Ratings[rateId])
					}
				}
				for rateId := range restoreRply8.DestinationRates {
					if !reflect.DeepEqual(restoreRply8.DestinationRates[rateId], ratingPl8.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply8.DestinationRates[rateId], ratingPl8.DestinationRates[rateId])
					}
				}
				restoreRply9 := new(engine.RatingPlan)
				rplnId = "RP_FREE"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply9); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply9.Ratings {
					if !reflect.DeepEqual(restoreRply9.Ratings[rateId], ratingPl9.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply9.Ratings[rateId], ratingPl9.Ratings[rateId])
					}
				}
				for rateId := range restoreRply9.DestinationRates {
					if !reflect.DeepEqual(restoreRply9.DestinationRates[rateId], ratingPl9.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply9.DestinationRates[rateId], ratingPl9.DestinationRates[rateId])
					}
				}
				restoreRply10 := new(engine.RatingPlan)
				rplnId = "RP_ANY2CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply10); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply10.Ratings {
					if !reflect.DeepEqual(restoreRply10.Ratings[rateId], ratingPl10.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply10.Ratings[rateId], ratingPl10.Ratings[rateId])
					}
				}
				for rateId := range restoreRply10.DestinationRates {
					if !reflect.DeepEqual(restoreRply10.DestinationRates[rateId], ratingPl10.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply10.DestinationRates[rateId], ratingPl10.DestinationRates[rateId])
					}
				}
				restoreRply11 := new(engine.RatingPlan)
				rplnId = "RP_ANY1CNT_SEC"
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingPlan, &rplnId, restoreRply11); err != nil {
					t.Error("Got error on APIerSv1.GetRatingPlan: ", err.Error())
				}
				for rateId := range restoreRply11.Ratings {
					if !reflect.DeepEqual(restoreRply11.Ratings[rateId], ratingPl11.Ratings[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply11.Ratings[rateId], ratingPl11.Ratings[rateId])
					}
				}
				for rateId := range restoreRply11.DestinationRates {
					if !reflect.DeepEqual(restoreRply11.DestinationRates[rateId], ratingPl11.DestinationRates[rateId]) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", restoreRply11.DestinationRates[rateId], ratingPl11.DestinationRates[rateId])
					}
				}
			})
			t.Run("GetRatingProfiles4", func(t *testing.T) {
				var rcvRprof engine.RatingProfile
				if err := client.Call(context.Background(), utils.APIerSv1GetRatingProfile,
					&utils.AttrGetRatingProfile{
						Tenant:   "cgrates.org",
						Category: "free",
						Subject:  "RP_FREE",
					}, &rcvRprof); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(rcvRprof, rProf) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rProf, rcvRprof)
				}
			})

			t.Run("GetDestinations4", func(t *testing.T) {
				sort.Slice(dests, func(i, j int) bool {
					return dests[i].Id < dests[j].Id
				})
				var rcv []*engine.Destination
				attrs := &v2.AttrGetDestinations{DestinationIDs: []string{}}
				if err := client.Call(context.Background(), utils.APIerSv2GetDestinations, attrs, &rcv); err != nil {
					t.Error("Got error on APIerSv2.GetDestinations: ", err.Error())
				}
				sort.Slice(rcv, func(i, j int) bool {
					return rcv[i].Id < rcv[j].Id
				})
				if !reflect.DeepEqual(dests, rcv) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", dests, rcv)
				}
			})

			t.Run("GetReverseDestination4", func(t *testing.T) {
				checkRD := func(t *testing.T, dst string) (rpl []string) {
					if err := client.Call(context.Background(), utils.APIerSv1GetReverseDestination, utils.StringPointer(dst), &rpl); err != nil {
						t.Errorf("Error dst <%s>, <%v>", dst, err)
					}
					return
				}
				rcvRdID1 := checkRD(t, "1001")
				rcvRdID2 := checkRD(t, "1002")
				rcvRdID3 := checkRD(t, "+49151")
				rcvRdID4 := checkRD(t, "077")
				rcvRdID5 := checkRD(t, "10")
				rcvRdID6 := checkRD(t, "+246")
				rcvRdID7 := checkRD(t, "+135")
				if !reflect.DeepEqual(rdID1, rcvRdID1) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID1, rcvRdID1)
				}
				if !reflect.DeepEqual(rdID2, rcvRdID2) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID2, rcvRdID2)
				}
				if !reflect.DeepEqual(rdID3, rcvRdID3) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID3, rcvRdID3)
				}
				if !reflect.DeepEqual(rdID4, rcvRdID4) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID4, rcvRdID4)
				}
				if !reflect.DeepEqual(rdID5, rcvRdID5) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID5, rcvRdID5)
				}
				if !reflect.DeepEqual(rdID6, rcvRdID6) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID6, rcvRdID6)
				}
				if !reflect.DeepEqual(rdID7, rcvRdID7) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", rdID7, rcvRdID7)
				}
			})

			t.Run("GetActions4", func(t *testing.T) {
				var rcv map[string]engine.Actions
				if err := client.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{}, &rcv); err != nil {
					t.Error(err)
				}
				if len(actsMp) != len(rcv) {
					t.Errorf("Length of expected <%v>, doesnt match <%v>", len(actsMp), len(rcv))
				}

				for id, acts := range actsMp {
					if len(acts) != len(rcv[id]) {
						t.Errorf("Length of expected <%v>, doesnt match <%v>", len(acts), len(rcv[id]))
					}
					for i, act := range acts {
						if rcv[id][i].Balance.Blocker == nil {
							rcv[id][i].Balance.Blocker = utils.BoolPointer(false)
						}
						if rcv[id][i].Balance.Disabled == nil {
							rcv[id][i].Balance.Disabled = utils.BoolPointer(false)
						}
						if rcv[id][i].Balance.ExpirationDate == nil {
							rcv[id][i].Balance.ExpirationDate = act.Balance.ExpirationDate
						}
						if !reflect.DeepEqual(utils.ToJSON(act), utils.ToJSON(rcv[id][i])) {
							t.Errorf("expected <%+v>, \nreceived <%+v>", utils.ToJSON(act), utils.ToJSON(rcv[id][i]))
						}
					}
				}

			})

			t.Run("GetActionTriggers4", func(t *testing.T) {
				var rcv engine.ActionTriggers
				if err := client.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{}}, &rcv); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(atr, rcv) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", atr, rcv)
				}
			})

			t.Run("GetActionPlan4", func(t *testing.T) {
				var rcv []*engine.ActionPlan
				if err := client.Call(context.Background(), utils.APIerSv1GetActionPlan,
					&v1.AttrGetActionPlan{ID: "ATMS_1"}, &rcv); err != nil {
					t.Error(err)
				}
				if len(aps) != 1 || len(rcv) != 1 {
					t.Errorf("expected aps len 1, got <%v>, expected rcv len 1, got <%v>", len(aps), len(rcv))
				}
				if !reflect.DeepEqual(aps[0].Id, rcv[0].Id) {
					t.Errorf("expected <%+v>, \nreceived <%+v>", aps[0].Id, rcv[0].Id)
				}
				for id, actts := range aps[0].ActionTimings {
					if !reflect.DeepEqual(actts.ActionsID, rcv[0].ActionTimings[id].ActionsID) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.ActionsID, rcv[0].ActionTimings[id].ActionsID)
					}
					if !reflect.DeepEqual(actts.Uuid, rcv[0].ActionTimings[id].Uuid) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.Uuid, rcv[0].ActionTimings[id].Uuid)
					}
					if !reflect.DeepEqual(actts.ExtraData, rcv[0].ActionTimings[id].ExtraData) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.ExtraData, rcv[0].ActionTimings[id].ExtraData)
					}
					if !reflect.DeepEqual(actts.Weight, rcv[0].ActionTimings[id].Weight) {
						t.Errorf("expected <%+v>, \nreceived <%+v>", actts.Weight, rcv[0].ActionTimings[id].Weight)
					}
				}
			})

			t.Run("GetAccounts4", func(t *testing.T) { // accounts deleted before restoring should be recovered
				var acnts2 []*engine.Account
				if err := client.Call(context.Background(), utils.APIerSv2GetAccounts,
					&utils.AttrGetAccounts{
						Tenant: "cgrates.org",
					}, &acnts2); err != nil {
					t.Errorf("APIerSv2GetAccounts failed unexpectedly: %v", err)
				}
				if len(acnts2) != 2 {
					t.Fatalf("APIerSv2GetAccounts len(acnts)=%v, want 2", len(acnts))
				}
				sort.Slice(acnts2, func(i, j int) bool {
					return acnts2[i].ID > acnts2[j].ID
				})
				if !reflect.DeepEqual(acnts2, acnts) {
					t.Errorf("Expected accounts to be the same. Before shutdown \n<%v>\nAfter rebooting <%v>", utils.ToJSON(acnts), utils.ToJSON(acnts2))
				}
			})

			t.Run("EngineShutdown", func(t *testing.T) { // make sure to kill engine to display porper logs and delete files after engine is shutdown
				if err := engine.KillEngine(100); err != nil {
					t.Error(err)
				}
			})

			if i > 2 {
				if err := os.RemoveAll(dataDBBackupPath); err != nil {
					t.Error(err)
				}
				if err := os.RemoveAll(storDBBackupPath); err != nil {
					t.Error(err)
				}
			}

			if err := os.RemoveAll("/var/lib/cgrates/internal_db"); err != nil {
				t.Error(err)
			}
			if err := os.RemoveAll("/tmp/ExportPath1"); err != nil {
				t.Error(err)
			}
			if err := os.RemoveAll("/tmp/ExportPath2"); err != nil {
				t.Error(err)
			}
		})
	}
}
