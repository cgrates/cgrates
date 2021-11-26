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

package loaders

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/cgrates/rpcclient"
)

func TestRemoveFromDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	for _, lType := range []string{utils.MetaAttributes, utils.MetaResources, utils.MetaFilters, utils.MetaStats,
		utils.MetaThresholds, utils.MetaRoutes, utils.MetaChargers, utils.MetaDispatchers, utils.MetaDispatcherHosts,
		utils.MetaRateProfiles, utils.MetaActionProfiles, utils.MetaAccounts} {
		if err := removeFromDB(context.Background(), dm, lType, "cgrates.org", "ID", true, false, utils.NewOrderedNavigableMap()); err != utils.ErrNotFound {
			t.Error(err)
		}
	}
	expErrMsg := "cannot find RateIDs in map"
	if err := removeFromDB(context.Background(), dm, utils.MetaRateProfiles, "cgrates.org", "ID", true, true, utils.NewOrderedNavigableMap()); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := removeFromDB(context.Background(), dm, utils.MetaRateProfiles, "cgrates.org", "ID", true, true, newOrderNavMap(utils.MapStorage{utils.RateIDs: "RT1"})); err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := removeFromDB(context.Background(), dm, utils.EmptyString, "cgrates.org", "ID", true, false, utils.NewOrderedNavigableMap()); err != nil {
		t.Error(err)
	}
}

func testDryRunWithData(lType string, data []utils.MapStorage) (string, error) {
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	err := dryRun(context.Background(), lType, utils.InfieldSep, "test", TenantIDFromOrderedNavigableMap(data[0]), data)
	return buf.String(), err
}

func testDryRun(t *testing.T, lType string) string {
	data := utils.NewOrderedNavigableMap()
	data.SetAsSlice(utils.NewFullPath(utils.Tenant), []*utils.DataNode{utils.NewLeafNode("cgrates.org")})
	data.SetAsSlice(utils.NewFullPath(utils.ID), []*utils.DataNode{utils.NewLeafNode("ID")})
	buf, err := testDryRunWithData(lType, []*utils.OrderedNavigableMap{data})
	if err != nil {
		t.Fatal(lType, err)
	}
	return buf
}

func newOrderNavMap(mp utils.MapStorage) (o *utils.OrderedNavigableMap) {
	o = utils.NewOrderedNavigableMap()
	for k, v := range mp {
		o.SetAsSlice(utils.NewFullPath(k), []*utils.DataNode{utils.NewLeafNode(v)})
	}
	return
}
func TestDryRun(t *testing.T) {
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: AttributeProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Attributes\":null,\"Blocker\":false,\"Weight\":0}",
		testDryRun(t, utils.MetaAttributes); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ResourceProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"UsageTTL\":0,\"Limit\":0,\"AllocationMessage\":\"\",\"Blocker\":false,\"Stored\":false,\"Weight\":0,\"ThresholdIDs\":null}",
		testDryRun(t, utils.MetaResources); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: Filter: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"Rules\":[]}",
		testDryRun(t, utils.MetaFilters); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: StatsQueueProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"QueueLength\":0,\"TTL\":0,\"MinItems\":0,\"Metrics\":null,\"Stored\":false,\"Blocker\":false,\"Weight\":0,\"ThresholdIDs\":null}",
		testDryRun(t, utils.MetaStats); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ThresholdProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"MaxHits\":0,\"MinHits\":0,\"MinSleep\":0,\"Blocker\":false,\"Weight\":0,\"ActionProfileIDs\":null,\"Async\":false}",
		testDryRun(t, utils.MetaThresholds); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: RouteProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Sorting\":\"\",\"SortingParameters\":null,\"Routes\":null,\"Weights\":null}",
		testDryRun(t, utils.MetaRoutes); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ChargerProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"RunID\":\"\",\"AttributeIDs\":null,\"Weight\":0}",
		testDryRun(t, utils.MetaChargers); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: DispatcherProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Strategy\":\"\",\"StrategyParams\":{},\"Weight\":0,\"Hosts\":null}",
		testDryRun(t, utils.MetaDispatchers); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: RateProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Weights\":null,\"MinCost\":0,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"Rates\":{}}",
		testDryRun(t, utils.MetaRateProfiles); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: ActionProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Weight\":0,\"Schedule\":\"\",\"Targets\":{},\"Actions\":null}",
		testDryRun(t, utils.MetaActionProfiles); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
	if expLog, rplyLog := "[INFO] <LoaderS-test> DRY_RUN: Accounts: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Weights\":null,\"Opts\":{},\"Balances\":{},\"ThresholdIDs\":null}",
		testDryRun(t, utils.MetaAccounts); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	rplyLog, err := testDryRunWithData(utils.MetaDispatcherHosts, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", utils.Address: "127.0.0.1"})})
	if err != nil {
		t.Fatal(err)
	}
	if expLog := "[INFO] <LoaderS-test> DRY_RUN: DispatcherHost: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"Address\":\"127.0.0.1\",\"Transport\":\"*json\",\"ConnectAttempts\":0,\"Reconnects\":0,\"ConnectTimeout\":0,\"ReplyTimeout\":0,\"TLS\":false,\"ClientKey\":\"\",\"ClientCertificate\":\"\",\"CaCertificate\":\"\"}"; !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestDryRunWithUpdateStructErrors(t *testing.T) {
	expErrMsg := `strconv.ParseFloat: parsing "notWeight": invalid syntax`
	if _, err := testDryRunWithData(utils.MetaAttributes, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := testDryRunWithData(utils.MetaResources, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := testDryRunWithData(utils.MetaStats, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := testDryRunWithData(utils.MetaThresholds, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := testDryRunWithData(utils.MetaChargers, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := testDryRunWithData(utils.MetaDispatchers, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := testDryRunWithData(utils.MetaActionProfiles, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func TestDryRunWithModelsErrors(t *testing.T) {
	expErrMsg := `strconv.ParseFloat: parsing "float": invalid syntax`
	if _, err := testDryRunWithData(utils.MetaResources, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Limit": "float"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `time: invalid duration "float"`
	if _, err := testDryRunWithData(utils.MetaStats, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "TTL": "float"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := testDryRunWithData(utils.MetaThresholds, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "MinSleep": "float"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	expErrMsg = `invalid Weight <float> in string: <;float>`
	if _, err := testDryRunWithData(utils.MetaRoutes, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Weights": ";float"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := testDryRunWithData(utils.MetaRateProfiles, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Weights": ";float"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := testDryRunWithData(utils.MetaAccounts, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Weights": ";float"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `time: invalid duration "float"`
	if _, err := testDryRunWithData(utils.MetaDispatcherHosts, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "ReplyTimeout": "float", "Address": "127.0.0.1"})}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := testDryRunWithData(utils.MetaFilters, []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "ReplyTimeout": "float"})}); err != utils.ErrWrongPath {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `emtpy RSRParser in rule: <>`
	data := newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})
	data.SetAsSlice(utils.NewFullPath("Rules.Type"), []*utils.DataNode{utils.NewLeafNode("*no")})
	if _, err := testDryRunWithData(utils.MetaFilters, []*utils.OrderedNavigableMap{data}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func TestSetToDBWithUpdateStructErrors(t *testing.T) {
	expErrMsg := `strconv.ParseFloat: parsing "notWeight": invalid syntax`
	if err := setToDB(context.Background(), nil, utils.MetaAttributes, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaResources, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaStats, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaThresholds, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaChargers, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaDispatchers, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaActionProfiles, utils.InfieldSep, utils.NewTenantID(""), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Weight: "notWeight"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

}

func TestSetToDBWithModelsErrors(t *testing.T) {
	expErrMsg := `strconv.ParseFloat: parsing "float": invalid syntax`
	if err := setToDB(context.Background(), nil, utils.MetaResources, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Limit": "float"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	expErrMsg = `time: invalid duration "float"`
	if err := setToDB(context.Background(), nil, utils.MetaStats, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "TTL": "float"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaThresholds, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "MinSleep": "float"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	expErrMsg = `invalid Weight <float> in string: <;float>`
	if err := setToDB(context.Background(), nil, utils.MetaRoutes, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Weights": ";float"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaRateProfiles, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Weights": ";float"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaAccounts, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Weights": ";float"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `time: invalid duration "float"`
	if err := setToDB(context.Background(), nil, utils.MetaDispatcherHosts, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "ReplyTimeout": "float", "Address": "127.0.0.1"})}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaFilters, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "ReplyTimeout": "float"})}, true, false); err != utils.ErrWrongPath {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	expErrMsg = `emtpy RSRParser in rule: <>`
	data := newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})
	data.SetAsSlice(utils.NewFullPath("Rules.Type"), []*utils.DataNode{utils.NewLeafNode("*no")})
	if err := setToDB(context.Background(), nil, utils.MetaFilters, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{data}, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if err := setToDB(context.Background(), nil, utils.EmptyString, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Error(err)
	}
}

func TestSetToDBWithDBError(t *testing.T) {
	if err := setToDB(context.Background(), nil, utils.MetaAttributes, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaResources, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaStats, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaThresholds, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaChargers, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaDispatchers, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaActionProfiles, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}

	if err := setToDB(context.Background(), nil, utils.MetaFilters, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaRoutes, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaDispatcherHosts, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Address": "127.0.0.1"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaRateProfiles, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
	if err := setToDB(context.Background(), nil, utils.MetaAccounts, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != utils.ErrNoDatabaseConn {
		t.Fatal(err)
	}
}

func TestSetToDB(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	if err := setToDB(context.Background(), dm, utils.MetaAttributes, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v1 := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v1, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v1), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaResources, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v2 := &engine.ResourceProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetResourceProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v2, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v2), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaStats, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v3 := &engine.StatQueueProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetStatQueueProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v3, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v3), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaThresholds, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v4 := &engine.ThresholdProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetThresholdProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v4, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v4), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaChargers, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v5 := &engine.ChargerProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetChargerProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v5, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v5), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaDispatchers, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v6 := &engine.DispatcherProfile{Tenant: "cgrates.org", ID: "ID", StrategyParams: make(map[string]interface{})}
	if prf, err := dm.GetDispatcherProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v6, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v6), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaActionProfiles, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v7 := &engine.ActionProfile{Tenant: "cgrates.org", ID: "ID", Targets: map[string]utils.StringSet{}}
	if prf, err := dm.GetActionProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v7, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v7), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaFilters, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v8 := &engine.Filter{Tenant: "cgrates.org", ID: "ID", Rules: make([]*engine.FilterRule, 0)}
	v8.Compile()
	if prf, err := dm.GetFilter(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v8, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v8), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaRoutes, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v9 := &engine.RouteProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetRouteProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v9, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v9), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaDispatcherHosts, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Address": "127.0.0.1"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v10 := &engine.DispatcherHost{Tenant: "cgrates.org", RemoteHost: &config.RemoteHost{ID: "ID", Address: "127.0.0.1", Transport: utils.MetaJSON}}
	if prf, err := dm.GetDispatcherHost(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v10, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v10), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaRateProfiles, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v11 := &utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{}, MinCost: utils.NewDecimal(0, 0), MaxCost: utils.NewDecimal(0, 0)}
	if prf, err := dm.GetRateProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v11, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v11), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaAccounts, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, false); err != nil {
		t.Fatal(err)
	}
	v12 := &utils.Account{Tenant: "cgrates.org", ID: "ID", Balances: map[string]*utils.Balance{}, Opts: make(map[string]interface{})}
	if prf, err := dm.GetAccount(context.Background(), "cgrates.org", "ID"); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v12, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v12), utils.ToJSON(prf))
	}

	if err := setToDB(context.Background(), dm, utils.MetaRateProfiles, utils.InfieldSep, utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, true, true); err != nil {
		t.Fatal(err)
	}
	v13 := &utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{}, MinCost: utils.NewDecimal(0, 0), MaxCost: utils.NewDecimal(0, 0)}
	if prf, err := dm.GetRateProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v13, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v13), utils.ToJSON(prf))
	}
}

func TestLoaderProcess(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)
	if expLd := (&loader{
		cfg:       cfg,
		ldrCfg:    cfg.LoaderCfg()[0],
		dm:        dm,
		filterS:   fS,
		connMgr:   cM,
		dataCache: cache,
		Locker:    newLocker(cfg.LoaderCfg()[0].GetLockFilePath()),
	}); !reflect.DeepEqual(expLd, ld) {
		t.Errorf("Expeceted: %+v, received: %+v", expLd, ld)
	}

	expErrMsg := `unsupported loader action: <"notSupported">`
	if err := ld.process(context.Background(), utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{}, utils.MetaAttributes, "notSupported", utils.MetaNone, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if err := ld.process(context.Background(), utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{}, utils.MetaAttributes, utils.MetaParse, utils.MetaNone, true, false); err != nil {
		t.Error(err)
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	if err := ld.process(context.Background(), utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaAttributes, utils.MetaDryRun, utils.MetaNone, true, false); err != nil {
		t.Error(err)
	}

	if expLog, rplyLog := "[INFO] <LoaderS-*default> DRY_RUN: AttributeProfile: {\"Tenant\":\"cgrates.org\",\"ID\":\"ID\",\"FilterIDs\":null,\"Attributes\":null,\"Blocker\":false,\"Weight\":0}",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

	if err := ld.process(context.Background(), utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaAttributes, utils.MetaStore, utils.MetaNone, true, false); err != nil {
		t.Error(err)
	}
	v1 := &engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", true, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(v1, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v1), utils.ToJSON(prf))
	}
	if err := ld.process(context.Background(), utils.NewTenantID("cgrates.org:ID"), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaAttributes, utils.MetaRemove, utils.MetaNone, true, false); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}

type ccMock map[string]func(ctx *context.Context, args interface{}, reply interface{}) error

func (ccM ccMock) Call(ctx *context.Context, serviceMethod string, args interface{}, reply interface{}) (err error) {
	if call, has := ccM[serviceMethod]; has {
		return call(ctx, args, reply)
	}
	return rpcclient.ErrUnsupporteServiceMethod
}

func TestLoaderProcessCallCahe(t *testing.T) {
	var reloadCache, clearCache interface{}
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	connID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCache)
	tntID := "cgrates.org:ID"
	iCh := make(chan birpc.ClientConnector, 1)
	iCh <- ccMock{
		utils.CacheSv1ReloadCache: func(_ *context.Context, args, _ interface{}) error { reloadCache = args; return nil },
		utils.CacheSv1Clear:       func(_ *context.Context, args, _ interface{}) error { clearCache = args; return nil },
	}
	cM.AddInternalConn(connID, utils.CacheSv1, iCh)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, []string{connID})

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaAttributes, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{AttributeProfileIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheAttributeFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaResources, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetResourceProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.ResourceProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{ResourceProfileIDs: []string{tntID}, ResourceIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheResourceFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaStats, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetStatQueueProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.StatQueueProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{StatsQueueProfileIDs: []string{tntID}, StatsQueueIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheStatFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaThresholds, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetThresholdProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.ThresholdProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{ThresholdProfileIDs: []string{tntID}, ThresholdIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheThresholdFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaRoutes, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetRouteProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.RouteProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{RouteProfileIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheRouteFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaChargers, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetChargerProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.ChargerProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{ChargerProfileIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheChargerFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaDispatchers, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetDispatcherProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.DispatcherProfile{Tenant: "cgrates.org", ID: "ID", StrategyParams: make(map[string]interface{})}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{DispatcherProfileIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheDispatcherFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaRateProfiles, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetRateProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&utils.RateProfile{Tenant: "cgrates.org", ID: "ID", Rates: map[string]*utils.Rate{}, MinCost: utils.NewDecimal(0, 0), MaxCost: utils.NewDecimal(0, 0)}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{RateProfileIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheRateProfilesFilterIndexes, utils.CacheRateFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaActionProfiles, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetActionProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.ActionProfile{Tenant: "cgrates.org", ID: "ID", Targets: map[string]utils.StringSet{}}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{ActionProfileIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheActionProfiles, utils.CacheActionProfilesFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}

	reloadCache, clearCache = nil, nil

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaFilters, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetFilter(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.Filter{Tenant: "cgrates.org", ID: "ID", Rules: make([]*engine.FilterRule, 0)}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{FilterIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if !reflect.DeepEqual(nil, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(nil), utils.ToJSON(clearCache))
	}

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID", "Address": "127.0.0.1"})}, utils.MetaDispatcherHosts, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetDispatcherHost(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.DispatcherHost{Tenant: "cgrates.org", RemoteHost: &config.RemoteHost{ID: "ID", Address: "127.0.0.1", Transport: utils.MetaJSON}}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{DispatcherHostIDs: []string{tntID}}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if !reflect.DeepEqual(nil, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(nil), utils.ToJSON(clearCache))
	}

	reloadCache, clearCache = nil, nil

	if err := ld.process(context.Background(), utils.NewTenantID(tntID), []*utils.OrderedNavigableMap{newOrderNavMap(utils.MapStorage{utils.Tenant: "cgrates.org", utils.ID: "ID"})}, utils.MetaAccounts, utils.MetaStore, utils.MetaReload, true, false); err != nil {
		t.Error(err)
	}
	if prf, err := dm.GetAccount(context.Background(), "cgrates.org", "ID"); err != nil {
		t.Fatal(err)
	} else if v := (&utils.Account{Tenant: "cgrates.org", ID: "ID", Balances: map[string]*utils.Balance{}, Opts: make(map[string]interface{})}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if exp := (&utils.AttrReloadCacheWithAPIOpts{}); !reflect.DeepEqual(exp, reloadCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(reloadCache))
	}
	if exp := (&utils.AttrCacheIDsWithAPIOpts{CacheIDs: []string{utils.CacheAccounts, utils.CacheAccountsFilterIndexes}}); !reflect.DeepEqual(exp, clearCache) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(exp), utils.ToJSON(clearCache))
	}
}

func TestLoaderProcessData(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)

	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	if err := ld.processData(context.Background(), NewStringCSV(`cgrates.org,ID
cgrates.org,ID2`, utils.CSVSep, -1), fc, utils.MetaAttributes, utils.MetaStore, utils.MetaNone, true, false); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID2", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID2"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}
}

type mockCSV struct{}

func (mockCSV) Path() (_ string)        { return }
func (mockCSV) Read() ([]string, error) { return nil, utils.ErrNotFound }
func (mockCSV) Close() (_ error)        { return }

func TestLoaderProcessDataErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)

	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	expErrMsg := "inline parse error for string: <*string>"
	if err := ld.processData(context.Background(), NewStringCSV(`cgrates.org,ID
cgrates.org,ID2`, utils.CSVSep, -1), fc, utils.MetaAttributes, utils.MetaStore, utils.MetaNone, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}

	fc = []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	expErrMsg = `unsupported loader action: <"notSupported">`
	if err := ld.processData(context.Background(), NewStringCSV(`cgrates.org,ID
cgrates.org,ID2`, utils.CSVSep, -1), fc, utils.MetaAttributes, "notSupported", utils.MetaNone, true, false); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}

	if err := ld.processData(context.Background(), mockCSV{}, fc, utils.MetaAttributes, "notSupported", utils.MetaNone, true, false); err != utils.ErrNotFound {
		t.Errorf("Expeceted: %q, received: %v", utils.ErrNotFound, err)
	}
}

func TestLoaderProcessFileURL(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	ld := newLoader(cfg, cfg.LoaderCfg()[0], dm, cache, fS, cM, nil)

	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}

	mux := http.NewServeMux()
	mux.Handle("/ok/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) { rw.Write([]byte(`cgrates.org,ID`)) }))
	s := httptest.NewServer(mux)
	defer s.Close()
	runtime.Gosched()

	if err := ld.processFile(context.Background(), &config.LoaderDataType{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}, s.URL+"/ok", utils.EmptyString, utils.MetaStore, utils.MetaNone, true); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}

	if err := ld.processFile(context.Background(), &config.LoaderDataType{
		Type:     utils.MetaAttributes,
		Filename: utils.AttributesCsv,
		Fields:   fc,
	}, s.URL+"/notFound", utils.EmptyString, utils.MetaStore, utils.MetaNone, true); err != utils.ErrNotFound {
		t.Errorf("Expeceted: %v, received: %v", utils.ErrNotFound, err)
	}

}

type mockLock struct{}

// lockFolder will attempt to lock the folder by creating the lock file
func (mockLock) Lock() error                { return utils.ErrExists }
func (mockLock) Unlock() (_ error)          { return }
func (mockLock) Locked() (_ bool, _ error)  { return true, utils.ErrExists }
func (mockLock) IsLockFile(string) (_ bool) { return }

func TestLoaderProcessIFile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessIFileIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	tmpOut, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessIFileOut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpOut)
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  tmpIn,
		TpOutDir: tmpOut,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)
	expErrMsg := fmt.Sprintf(`rename %s/Chargers.csv %s/Chargers.csv: no such file or directory`, tmpIn, tmpOut)
	if err := ld.processIFile(utils.EmptyString, utils.ChargersCsv); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ld.processIFile(utils.EmptyString, utils.AttributesCsv); err != nil {
		t.Fatal(err)
	}
	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.AttributesCsv)); err == nil {
		t.Errorf("Expected file to be moved")
	} else if !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(tmpOut, utils.AttributesCsv)); err != nil {
		t.Errorf("Expected file to be moved")
	}

	ld.Locker = mockLock{}
	if err := ld.processIFile(utils.EmptyString, utils.AttributesCsv); err != utils.ErrExists {
		t.Fatal(err)
	}
}

func TestLoaderProcessFolder(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Path: utils.Tenant, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.0", utils.RSRConstSep)},
		{Path: utils.ID, Type: utils.MetaVariable, Value: config.NewRSRParsersMustCompile("~*req.1", utils.RSRConstSep)},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	tmpOut, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderOut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpOut)
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  tmpIn,
		TpOutDir: tmpOut,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	f, err = os.Create(path.Join(tmpIn, utils.ChargersCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := ld.processFolder(context.Background(), utils.MetaNone, true, true); err != nil {
		t.Fatal(err)
	}

	if prf, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != nil {
		t.Fatal(err)
	} else if v := (&engine.AttributeProfile{Tenant: "cgrates.org", ID: "ID"}); !reflect.DeepEqual(v, prf) {
		t.Errorf("Expeceted: %v, received: %v", utils.ToJSON(v), utils.ToJSON(prf))
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.AttributesCsv)); err == nil {
		t.Errorf("Expected file to be moved")
	} else if !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(tmpOut, utils.AttributesCsv)); err != nil {
		t.Errorf("Expected file to be moved")
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.ChargersCsv)); err == nil {
		t.Errorf("Expected file to be moved")
	} else if !os.IsNotExist(err) {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(tmpOut, utils.ChargersCsv)); err != nil {
		t.Errorf("Expected file to be moved")
	}

	ld.Locker = mockLock{}
	if err := ld.processFolder(context.Background(), utils.MetaNone, true, true); err != utils.ErrExists {
		t.Fatal(err)
	}
}

func TestLoaderProcessFolderErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cM := engine.NewConnManager(cfg)
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), cM)
	fS := engine.NewFilterS(cfg, cM, dm)
	cache := map[string]*ltcache.Cache{}
	for k, cfg := range cfg.LoaderCfg()[0].Cache {
		cache[k] = ltcache.NewCache(cfg.Limit, cfg.TTL, cfg.StaticTTL, nil)
	}
	fc := []*config.FCTemplate{
		{Filters: []string{"*string"}},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderErrorsIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	tmpOut, err := os.MkdirTemp(utils.EmptyString, "TestLoaderProcessFolderErrorsOut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpOut)
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  tmpIn,
		TpOutDir: tmpOut,
		Data: []*config.LoaderDataType{
			{
				Type:     utils.MetaAttributes,
				Filename: utils.AttributesCsv,
				Fields:   fc,
			},
		},
		FieldSeparator: utils.FieldsSep,
		Action:         utils.MetaStore,
		Opts: &config.LoaderSOptsCfg{
			WithIndex: true,
			Cache:     utils.MetaNone,
		},
	}, dm, cache, fS, cM, nil)

	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(`cgrates.org,ID`); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	expErrMsg := "inline parse error for string: <*string>"
	if err := ld.processFolder(context.Background(), utils.MetaNone, true, true); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if _, err := os.Stat(path.Join(tmpIn, utils.AttributesCsv)); err != nil {
		t.Errorf("Expected file to not be moved because of template error: %v", err)
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	if err := ld.processFolder(context.Background(), utils.MetaNone, true, false); err != nil {
		t.Fatal(err)
	}

	if _, err := dm.GetAttributeProfile(context.Background(), "cgrates.org", "ID", false, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Fatal(err)
	}

	if expLog, rplyLog := "<LoaderS-test> loaderType: <*attributes> cannot open files, err: inline parse error for string: <*string>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}

}

func TestLoaderMoveUnprocessedFilesErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:      "test",
		Enabled: true,
		TpInDir: "notAFolder",
	}, nil, nil, nil, nil, nil)

	expErrMsg := "open notAFolder: no such file or directory"
	if err := ld.moveUnprocessedFiles(); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	tmpIn, err := os.MkdirTemp(utils.EmptyString, "TestLoaderMoveUnprocessedFilesErrorsIn")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpIn)
	ld.ldrCfg.TpInDir = tmpIn
	ld.ldrCfg.TpOutDir = "notAFolder"
	f, err := os.Create(path.Join(tmpIn, utils.AttributesCsv))
	if err != nil {
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	expErrMsg = fmt.Sprintf("rename %s/Attributes.csv notAFolder/Attributes.csv: no such file or directory", tmpIn)
	if err := ld.moveUnprocessedFiles(); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func TestLoaderHandleFolder(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		RunDelay: time.Nanosecond,
		TpInDir:  "/tmp/TestLoaderHandleFolder",
		Opts:     &config.LoaderSOptsCfg{},
	}, nil, nil, nil, nil, nil)
	ld.Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.handleFolder(stop)

	if expLog, rplyLog := "[INFO] <LoaderS-test> stop monitoring path </tmp/TestLoaderHandleFolder>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestLoaderListenAndServe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		RunDelay: time.Nanosecond,
		TpInDir:  "/tmp/TestLoaderListenAndServe",
		Opts:     &config.LoaderSOptsCfg{},
	}, nil, nil, nil, nil, nil)
	ld.Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.ListenAndServe(stop)
	runtime.Gosched()
	time.Sleep(time.Nanosecond)
	if expLog, rplyLog := "[INFO] Starting <LoaderS-test>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}

func TestLoaderListenAndServeI(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	ld := newLoader(cfg, &config.LoaderSCfg{
		ID:       "test",
		Enabled:  true,
		TpInDir:  "/tmp/TestLoaderListenAndServeI",
		RunDelay: -1,
		Opts:     &config.LoaderSOptsCfg{},
	}, nil, nil, nil, nil, nil)
	ld.Locker = mockLock{}
	stop := make(chan struct{})
	close(stop)

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	ld.ListenAndServe(stop)
	runtime.Gosched()
	time.Sleep(time.Nanosecond)
	if expLog, rplyLog := "[INFO] Starting <LoaderS-test>",
		buf.String(); !strings.Contains(rplyLog, expLog) {
		t.Errorf("Expected %+q, received %+q", expLog, rplyLog)
	}
}
