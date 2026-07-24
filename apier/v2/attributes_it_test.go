//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v2

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	attrSRPC        *rpc.Client
	alsPrfConfigDIR string //run tests for specific configuration

	sTestsAlsPrf = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testAttributeSSetAlsPrf,
		testAttributeSUpdateAlsPrf,
		testAttributeSKillEngine,
	}
)

// Test start here
func TestAttributeSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		alsPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		alsPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		alsPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	alsPrfCfgPath = path.Join(*utils.DataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromPath(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	alsPrfCfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(alsPrfCfg)
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(alsPrfCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrSRPC, err = newRPCClient(alsPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeSSetAlsPrf(t *testing.T) {
	extAlsPrf := &AttributeWithCache{
		ExternalAttributeProfile: &engine.ExternalAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: "1001",
				},
			},
			Weight: 20,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv2SetAttributeProfile, extAlsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ExternalAttribute"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
}

func testAttributeSUpdateAlsPrf(t *testing.T) {
	extAlsPrf := &AttributeWithCache{
		ExternalAttributeProfile: &engine.ExternalAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: "1001",
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Value: "~*req.Account",
				},
			},
			Weight: 20,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv2SetAttributeProfile, extAlsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Value: config.NewRSRParsersMustCompile("~*req.Account", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	sort.Strings(alsPrf.AttributeProfile.Contexts)
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ExternalAttribute"}}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.Contexts)
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
}

func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
