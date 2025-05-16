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
package general_tests

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestOfflineInternal(t *testing.T) { // run with sudo
	paths := []string{
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal"),                     // dump -1
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms"),                  // dump ms
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite"),             // dump -1 and rewrite -1
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_ms"),          // dump -1 and rewrite ms
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite"),          // dump ms and rewrite -1
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_ms"),       // dump ms and rewrite ms
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_limit"),               // dump -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_limit"),            // dump ms and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_limit"),       // dump -1 and rewrite -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_rewrite_ms_limit"),    // dump -1 and rewrite ms and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_limit"),    // dump ms and rewrite -1 and limit passed
		path.Join(*utils.DataDir, "conf", "samples", "offline_internal_ms_rewrite_ms_limit"), // dump ms and rewrite ms and limit passed
	}
	for i, pth := range paths {
		if err := os.MkdirAll(config.NewDefaultCGRConfig().DataDbCfg().Opts.InternalDBDumpPath, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(config.NewDefaultCGRConfig().StorDbCfg().Opts.InternalDBDumpPath, 0755); err != nil {
			t.Fatal(err)
		}
		t.Run("OfflineInternal"+strconv.Itoa(i), func(t *testing.T) {
			ng := engine.TestEngine{
				ConfigPath:       pth,
				PreInitDB:        true,
				GracefulShutdown: true,
			}
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
				if err := client.Call(context.Background(), utils.APIerSv1SetActionTrigger, v1.AttrSetActionTrigger{
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
				atms1 := &v1.AttrSetActionPlan{
					Id: "ATMS_1",
					ActionPlan: []*v1.AttrActionPlan{
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

			t.Run("EngineShutdown", func(t *testing.T) {
				if err := engine.KillEngine(100); err != nil {
					t.Error(err)
				}
			})
			t.Run("CountDataDBFiles", func(t *testing.T) {
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
				} else if dirs != 40 {
					t.Errorf("expected <%d> directories, received <%d>", 40, dirs)
				} else if i > 6 && (files != 29 && files != 30) { // depends if rewriting is scheduled or not by the time we shutdown
					t.Errorf("expected 29 or 30 files, received <%d>", files)
				} else if i < 6 && files != 28 {
					t.Errorf("expected <%d> files, received <%d>", 28, files)
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
				} else if dirs != 27 {
					t.Errorf("expected <%d> directories, received <%d>", 27, dirs)
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

						for i := 0; i < len(lines1); i++ {
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
