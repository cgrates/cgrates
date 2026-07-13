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

package main

import (
	"net"
	"path"
	"slices"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCGRLoaderStoreRemove(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	case utils.MetaRedis:
		dbCfg = engine.RedisDBCfg
	case utils.MetaMySQL:
		dbCfg = engine.MySQLDBCfg
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	default:
		t.Fatal("unsupported dbtype value")
	}

	testEngine := engine.TestEngine{
		ConfigPath: path.Join(*utils.DataDir, "conf", "samples", "tutinternal"),
		DBCfg:      dbCfg,
		Encoding:   utils.MetaJSON,
	}
	client, cfg := testEngine.Run(t)
	tpPath := path.Join(*utils.DataDir, "tariffplans", "testit")

	engine.LoadCSVsWithCGRLoader(t, cfg.ConfigPath, tpPath, nil, nil)
	_ = client.Close()
	testEngine.Stop(t)
	client = testEngine.Start(t)

	args := func(id string) *utils.TenantIDWithAPIOpts {
		return &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: utils.CGRateSorg, ID: id},
		}
	}

	var attribute *utils.APIAttributeProfile
	if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		args("ATTR_ACNT_1001"), &attribute); err != nil {
		t.Fatal(err)
	}
	if attribute == nil || attribute.Tenant != utils.CGRateSorg || attribute.ID != "ATTR_ACNT_1001" ||
		!slices.Contains(attribute.FilterIDs, "FLTR_ACCOUNT_1001") ||
		len(attribute.Attributes) != 1 || attribute.Attributes[0] == nil ||
		attribute.Attributes[0].Path != "*req.OfficeGroup" ||
		attribute.Attributes[0].Value != "Marketing" {
		t.Errorf("unexpected attribute profile: %s", utils.ToJSON(attribute))
	}

	var resourceProfile *utils.ResourceProfile
	if err := client.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		args("RES_ACNT_1001"), &resourceProfile); err != nil {
		t.Fatal(err)
	}
	if resourceProfile == nil || resourceProfile.Limit != 1 || resourceProfile.UsageTTL != time.Hour ||
		!slices.Contains(resourceProfile.FilterIDs, "FLTR_ACCOUNT_1001") {
		t.Errorf("unexpected resource profile: %s", utils.ToJSON(resourceProfile))
	}

	var resource *utils.Resource
	if err := client.Call(context.Background(), utils.ResourceSv1GetResource,
		args("RES_ACNT_1001"), &resource); err != nil {
		t.Fatal(err)
	}
	if resource == nil || resource.Tenant != utils.CGRateSorg || resource.ID != "RES_ACNT_1001" || len(resource.Usages) != 0 {
		t.Errorf("unexpected resource: %s", utils.ToJSON(resource))
	}

	var rateProfile *utils.RateProfile
	if err := client.Call(context.Background(), utils.AdminSv1GetRateProfile,
		args("RT_SPECIAL_1002"), &rateProfile); err != nil {
		t.Fatal(err)
	}
	if rateProfile == nil {
		t.Fatal("rate profile is nil")
	}
	rate, has := rateProfile.Rates["RT_ALWAYS"]
	if !has || rate == nil || len(rate.IntervalRates) != 1 || rate.IntervalRates[0] == nil ||
		rate.IntervalRates[0].RecurrentFee == nil ||
		rate.IntervalRates[0].RecurrentFee.Cmp(utils.NewDecimal(1, 2).Big) != 0 {
		t.Errorf("unexpected rate profile: %s", utils.ToJSON(rateProfile))
	}

	var account *utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		args("ACC_PRF_1"), &account); err != nil {
		t.Fatal(err)
	}
	if account == nil {
		t.Fatal("account is nil")
	}
	balance, has := account.Balances["MonetaryBalance"]
	if !has || balance == nil || balance.Units == nil ||
		balance.Units.Cmp(utils.NewDecimal(14, 0).Big) != 0 || len(balance.UnitFactors) != 2 {
		t.Errorf("unexpected account: %s", utils.ToJSON(account))
	}

	var filter *engine.Filter
	if err := client.Call(context.Background(), utils.AdminSv1GetFilter,
		args("FLTR_1"), &filter); err != nil {
		t.Fatal(err)
	}
	if filter == nil {
		t.Fatal("filter is nil")
	}
	var hasPrefixRule bool
	for _, rule := range filter.Rules {
		if rule != nil && rule.Type == utils.MetaPrefix && rule.Element == "~*req.Destination" &&
			slices.Contains(rule.Values, "10") && slices.Contains(rule.Values, "20") {
			hasPrefixRule = true
			break
		}
	}
	if len(filter.Rules) != 3 || !hasPrefixRule {
		t.Errorf("unexpected filter: %s", utils.ToJSON(filter))
	}

	engine.LoadCSVsWithCGRLoader(t, cfg.ConfigPath, tpPath, nil, nil, "-remove")

	for _, check := range []struct {
		name   string
		method string
		args   any
		reply  any
	}{
		{name: "attribute profile", method: utils.AdminSv1GetAttributeProfile, args: args("ATTR_ACNT_1001"), reply: new(*utils.APIAttributeProfile)},
		{name: "resource profile", method: utils.AdminSv1GetResourceProfile, args: args("RES_ACNT_1001"), reply: new(*utils.ResourceProfile)},
		{name: "resource", method: utils.ResourceSv1GetResource, args: args("RES_ACNT_1001"), reply: new(*utils.Resource)},
		{name: "rate profile", method: utils.AdminSv1GetRateProfile, args: args("RT_SPECIAL_1002"), reply: new(*utils.RateProfile)},
		{name: "account", method: utils.AdminSv1GetAccount, args: args("ACC_PRF_1"), reply: new(*utils.Account)},
		{name: "filter", method: utils.AdminSv1GetFilter, args: args("FLTR_1"), reply: new(*engine.Filter)},
	} {
		t.Run("removed "+check.name, func(t *testing.T) {
			checkCGRLoaderNotFound(t, client, check.method, check.args, check.reply)
		})
	}
}

func TestCGRLoaderTenant(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaRedis, utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	checkCGRLoaderTenant(t, path.Join(*utils.DataDir, "conf", "samples", "tutinternal"))
}

type testCache struct {
	tenant chan string
}

func (recorder *testCache) ReloadCache(_ *context.Context,
	args *utils.AttrReloadCacheWithAPIOpts, reply *string) error {
	select {
	case recorder.tenant <- args.Tenant:
	default:
	}
	*reply = utils.OK
	return nil
}

func (*testCache) Clear(_ *context.Context, _ *utils.AttrCacheIDsWithAPIOpts,
	reply *string) error {
	*reply = utils.OK
	return nil
}

func checkCGRLoaderTenant(t *testing.T, cfgPath string) {
	t.Helper()
	recorder := &testCache{tenant: make(chan string, 1)}
	server := birpc.NewServer()
	if err := server.RegisterName(utils.CacheSv1, recorder); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()

	engine.LoadCSVsWithCGRLoader(t, cfgPath, "", nil, map[string]string{
		utils.AttributesCsv: "cgrates.org,ATTR_TENANT_TEST,,,,,,,,\n",
	}, "-caches_address="+listener.Addr().String(), "-tenant=tenant.com")
	select {
	case tenant := <-recorder.tenant:
		if tenant != "tenant.com" {
			t.Errorf("cache tenant is %q, want %q", tenant, "tenant.com")
		}
	default:
		t.Error("cgr-loader did not reload the cache")
	}
}

func checkCGRLoaderNotFound(t *testing.T, client *birpc.Client, method string, args, reply any) {
	t.Helper()
	err := client.Call(context.Background(), method, args, reply)
	if err == nil {
		t.Errorf("%s returned no error, want %v", method, utils.ErrNotFound)
		return
	}
	if err = utils.CastRPCErr(err); err != utils.ErrNotFound {
		t.Errorf("%s returned %v, want %v", method, err, utils.ErrNotFound)
	}
}
