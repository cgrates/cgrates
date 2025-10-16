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
package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// func TestConfigV1SetConfigWithDB(t *testing.T) {
// 	cfg := NewDefaultCGRConfig()
// 	cfg.rldCh = make(chan string, 100)
// 	db := make(CgrJsonCfg)
// 	cfg.db = db

// 	v2 := NewDefaultCGRConfig()
// 	v2.GeneralCfg().NodeID = "Test"
// 	v2.GeneralCfg().DefaultCaching = utils.MetaClear
// 	var reply string
// 	if err := cfg.V1SetConfig(context.Background(), &SetConfigArgs{
// 		Config: v2.AsMapInterface(utils.InfieldSep),
// 	}, &reply); err != nil {
// 		t.Fatal(err)
// 	}

// 	exp := &GeneralJsonCfg{
// 		Node_id:         utils.StringPointer("Test"),
// 		Default_caching: utils.StringPointer(utils.MetaClear),
// 		Opts:            &GeneralOptsJson{},
// 	}
// 	rpl := new(GeneralJsonCfg)
// 	if err := db.GetSection(context.Background(), GeneralJSON, rpl); err != nil {
// 		t.Fatal(err)
// 	} else if !reflect.DeepEqual(exp, rpl) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
// 	}
// 	exp2 := &AccountSJsonCfg{Opts: &AccountsOptsJson{}}
// 	rpl2 := new(AccountSJsonCfg)
// 	if err := db.GetSection(context.Background(), AccountSJSON, rpl2); err != nil {
// 		t.Fatal(err)
// 	} else if !reflect.DeepEqual(exp2, rpl2) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl2))
// 	}
// }

func TestConfigV1StoreCfgInDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	db := make(CgrJsonCfg)
	cfg.db = db

	cfg.GeneralCfg().NodeID = "Test"
	cfg.GeneralCfg().DefaultCaching = utils.MetaClear
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), &SectionWithAPIOpts{Sections: []string{utils.MetaAll}}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
		Opts:            &GeneralOptsJson{},
	}
	rpl := new(GeneralJsonCfg)
	if err := db.GetSection(context.Background(), GeneralJSON, rpl); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := &AccountSJsonCfg{Opts: &AccountsOptsJson{}}
	rpl2 := new(AccountSJsonCfg)
	if err := db.GetSection(context.Background(), AccountSJSON, rpl2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl2) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl2))
	}
}

func TestConfigV1StoreCfgInDBErr1(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	args := &SectionWithAPIOpts{
		Sections: []string{},
	}

	var reply string
	expected := "no DB connection for config"
	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestConfigV1StoreCfgInDBErr2(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)

	cfg.db = &mockDb{}

	args := &SectionWithAPIOpts{
		Sections: []string{CDRsJSON},
	}

	var reply string
	expected := utils.ErrNotImplemented
	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestConfigV1StoreCfgInDBErr3(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)

	args := &SectionWithAPIOpts{
		Sections: []string{"cdrs"},
	}

	cfg.db = new(mockDb)
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

// func TestConfigV1SetConfigFromJSONWithDB(t *testing.T) {
// 	cfg := NewDefaultCGRConfig()
// 	cfg.rldCh = make(chan string, 100)
// 	db := make(CgrJsonCfg)
// 	cfg.db = db

// 	v2 := NewDefaultCGRConfig()
// 	v2.GeneralCfg().NodeID = "Test"
// 	v2.GeneralCfg().DefaultCaching = utils.MetaClear
// 	var reply string
// 	if err := cfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{
// 		Config: utils.ToJSON(v2.AsMapInterface(utils.InfieldSep)),
// 	}, &reply); err != nil {
// 		t.Fatal(err)
// 	}

// 	exp := &GeneralJsonCfg{
// 		Node_id:         utils.StringPointer("Test"),
// 		Default_caching: utils.StringPointer(utils.MetaClear),
// 		Opts:            &GeneralOptsJson{},
// 	}
// 	rpl := new(GeneralJsonCfg)
// 	if err := db.GetSection(context.Background(), GeneralJSON, rpl); err != nil {
// 		t.Fatal(err)
// 	} else if !reflect.DeepEqual(exp, rpl) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
// 	}
// 	exp2 := &AccountSJsonCfg{Opts: &AccountsOptsJson{}}
// 	rpl2 := new(AccountSJsonCfg)
// 	if err := db.GetSection(context.Background(), AccountSJSON, rpl2); err != nil {
// 		t.Fatal(err)
// 	} else if !reflect.DeepEqual(exp2, rpl2) {
// 		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl2))
// 	}
// }

func TestConfigV1SetConfigFromJSONWithDBErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	args := &SetConfigFromJSONArgs{
		Config: `{
			"cdrs":{
				"enabled": false,
			}
		}
		`,
	}

	cfg.rldCh = make(chan string, 100)

	cfg.db = &mockDb{}
	var reply string
	expected := utils.ErrNotImplemented
	if err := cfg.V1SetConfigFromJSON(context.Background(), args, &reply); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, reply)
	}
}

func TestConfigLoadFromDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	db := make(CgrJsonCfg)
	g1 := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}

	if err := db.SetSection(context.Background(), GeneralJSON, g1); err != nil {
		t.Fatal(err)
	}

	if err := cfg.LoadFromDB(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	expGeneral := &GeneralCfg{
		NodeID:           "Test",
		DefaultCaching:   utils.MetaClear,
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
		DefaultReqType:   utils.MetaRated,
		DefaultCategory:  utils.Call,
		DefaultTenant:    "cgrates.org",
		DefaultTimezone:  "Local",
		ConnectAttempts:  5,
		Reconnects:       -1,
		ConnectTimeout:   time.Second,
		ReplyTimeout:     2 * time.Second,
		DigestSeparator:  ",",
		DigestEqual:      ":",
		MaxParallelConns: 100,
		Opts: &GeneralOpts{
			ExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	if !reflect.DeepEqual(expGeneral, cfg.GeneralCfg()) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expGeneral), utils.ToJSON(cfg.GeneralCfg()))
	}

	cfg.GeneralCfg().NodeID = "Test2"
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), &SectionWithAPIOpts{Sections: []string{utils.MetaAll}}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test2"),
		Default_caching: utils.StringPointer(utils.MetaClear),
		Opts:            &GeneralOptsJson{},
	}
	rpl := new(GeneralJsonCfg)
	if err := db.GetSection(context.Background(), GeneralJSON, rpl); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := &AccountSJsonCfg{Opts: &AccountsOptsJson{}}
	rpl2 := new(AccountSJsonCfg)
	if err := db.GetSection(context.Background(), AccountSJSON, rpl2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl2) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl2))
	}
}

func TestGetSectionAsMap(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	expected := map[string]any{
		"db_type":     utils.MetaInternal,
		"db_host":     "",
		"db_port":     0,
		"db_name":     "",
		"db_user":     "",
		"db_password": "",
		"opts": map[string]any{
			utils.InternalDBBackupPathCfg:      "/var/lib/cgrates/internal_db/backup/configdb",
			utils.InternalDBDumpIntervalCfg:    "0s",
			utils.InternalDBDumpPathCfg:        "/var/lib/cgrates/internal_db/configdb",
			utils.InternalDBFileSizeLimitCfg:   int64(1073741824),
			utils.InternalDBRewriteIntervalCfg: "0s",
			utils.InternalDBStartTimeoutCfg:    "5m0s",
			"redisMaxConns":                    10,
			"redisConnectAttempts":             20,
			"redisSentinel":                    "",
			"redisCluster":                     false,
			"redisClusterSync":                 "5s",
			"redisClusterOndownDelay":          "0s",
			"redisConnectTimeout":              "0s",
			"redisReadTimeout":                 "0s",
			"redisWriteTimeout":                "0s",
			utils.MongoQueryTimeoutCfg:         "10s",
			utils.MongoConnSchemeCfg:           "mongodb",
			"redisTLS":                         false,
			"redisClientCertificate":           "",
			"redisClientKey":                   "",
			"redisCACertificate":               "",
		},
	}

	rcv, err := cfg.getSectionAsMap(ConfigDBJSON)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

type mockDb struct {
	GetSectionF      func(*context.Context, string, any) error
	SetSectionF      func(*context.Context, string, any) error
	DumpConfigDBF    func() error
	RewriteConfigDBF func() error
	BackupConfigDBF  func(string, bool) error
}

func (m *mockDb) GetSection(ctx *context.Context, sec string, val any) error {
	if m.GetSectionF != nil {
		return m.GetSectionF(ctx, sec, val)
	}
	return utils.ErrNotImplemented
}

func (m *mockDb) SetSection(ctx *context.Context, sec string, val any) error {
	if m.SetSectionF != nil {
		return m.SetSectionF(ctx, sec, val)
	}
	return utils.ErrNotImplemented
}

func (m *mockDb) DumpConfigDB() error {
	if m.DumpConfigDBF != nil {
		return m.DumpConfigDBF()
	}
	return utils.ErrNotImplemented
}

func (m *mockDb) RewriteConfigDB() error {
	if m.RewriteConfigDBF != nil {
		return m.RewriteConfigDBF()
	}
	return utils.ErrNotImplemented
}

func (m *mockDb) BackupConfigDB(path string, zip bool) error {
	if m.BackupConfigDBF != nil {
		return m.BackupConfigDBF(path, zip)
	}
	return utils.ErrNotImplemented
}

func TestStoreDiffSectionGeneral(t *testing.T) {
	section := GeneralJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.generalCfg = &GeneralCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.generalCfg = &GeneralCfg{}

	if err := storeDiffSection(context.Background(), section, &mockDb{}, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLogger(t *testing.T) {
	section := LoggerJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loggerCfg = &LoggerCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loggerCfg = &LoggerCfg{}

	if err := storeDiffSection(context.Background(), section, &mockDb{}, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}
func TestStoreDiffSectionEFs(t *testing.T) {
	section := EFsJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.efsCfg = &EFsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.efsCfg = &EFsCfg{}

	if err := storeDiffSection(context.Background(), section, &mockDb{}, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}
func TestStoreDiffSectionRPCConns(t *testing.T) {
	section := RPCConnsJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.rpcConns = RPCConns{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.rpcConns = RPCConns{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCache(t *testing.T) {
	section := CacheJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.cacheCfg = &CacheCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.cacheCfg = &CacheCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionListen(t *testing.T) {
	section := ListenJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.listenCfg = &ListenCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.listenCfg = &ListenCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionHTTP(t *testing.T) {
	section := HTTPJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.httpCfg = &HTTPCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.httpCfg = &HTTPCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDataDB(t *testing.T) {
	section := DBJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dbCfg = &DbCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dbCfg = &DbCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionFilterS(t *testing.T) {
	section := FilterSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.filterSCfg = &FilterSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.filterSCfg = &FilterSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCDRs(t *testing.T) {
	section := CDRsJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.cdrsCfg = &CdrsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.cdrsCfg = &CdrsCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionERs(t *testing.T) {
	section := ERsJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.ersCfg = &ERsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.ersCfg = &ERsCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionEEs(t *testing.T) {
	section := EEsJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.eesCfg = &EEsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.eesCfg = &EEsCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSessionS(t *testing.T) {
	section := SessionSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sessionSCfg = &SessionSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.sessionSCfg = &SessionSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionFreeSWITCH(t *testing.T) {
	section := FreeSWITCHAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.fsAgentCfg = &FsAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.fsAgentCfg = &FsAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionKamailio(t *testing.T) {
	section := KamailioAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.kamAgentCfg = &KamAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.kamAgentCfg = &KamAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAsterisk(t *testing.T) {
	section := AsteriskAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.asteriskAgentCfg = &AsteriskAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.asteriskAgentCfg = &AsteriskAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDiameter(t *testing.T) {
	section := DiameterAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.diameterAgentCfg = &DiameterAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.diameterAgentCfg = &DiameterAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRadius(t *testing.T) {
	section := RadiusAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.diameterAgentCfg = &DiameterAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.diameterAgentCfg = &DiameterAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionHTTPAgent(t *testing.T) {
	section := HTTPAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.httpAgentCfg = HTTPAgentCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.httpAgentCfg = HTTPAgentCfgs{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDNS(t *testing.T) {
	section := DNSAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dnsAgentCfg = &DNSAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dnsAgentCfg = &DNSAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAttributeS(t *testing.T) {
	section := AttributeSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.attributeSCfg = &AttributeSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.attributeSCfg = &AttributeSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionChargerS(t *testing.T) {
	section := ChargerSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.chargerSCfg = &ChargerSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.chargerSCfg = &ChargerSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionResourceS(t *testing.T) {
	section := ResourceSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.resourceSCfg = &ResourceSConfig{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.resourceSCfg = &ResourceSConfig{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionStatS(t *testing.T) {
	section := StatSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.statsCfg = &StatSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.statsCfg = &StatSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionThresholdS(t *testing.T) {
	section := ThresholdSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.thresholdSCfg = &ThresholdSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.thresholdSCfg = &ThresholdSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRouteS(t *testing.T) {
	section := RouteSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.routeSCfg = &RouteSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.routeSCfg = &RouteSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLoaderS(t *testing.T) {
	section := LoaderSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loaderCfg = LoaderSCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loaderCfg = LoaderSCfgs{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSureTax(t *testing.T) {
	section := SureTaxJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sureTaxCfg = &SureTaxCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.sureTaxCfg = &SureTaxCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRegistrarC(t *testing.T) {
	section := RegistrarCJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.registrarCCfg = &RegistrarCCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.registrarCCfg = &RegistrarCCfgs{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLoader(t *testing.T) {
	section := LoaderJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loaderCgrCfg = &LoaderCgrCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loaderCgrCfg = &LoaderCgrCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionMigrator(t *testing.T) {
	section := MigratorJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.migratorCgrCfg = &MigratorCgrCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.migratorCgrCfg = &MigratorCgrCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionTls(t *testing.T) {
	section := TlsJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.tlsCfg = &TLSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.tlsCfg = &TLSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAnalyzerS(t *testing.T) {
	section := AnalyzerSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.analyzerSCfg = &AnalyzerSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.analyzerSCfg = &AnalyzerSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAdminS(t *testing.T) {
	section := AdminSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.admS = &AdminSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.admS = &AdminSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRateS(t *testing.T) {
	section := RateSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.rateSCfg = &RateSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.rateSCfg = &RateSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSIP(t *testing.T) {
	section := SIPAgentJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sipAgentCfg = &SIPAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV1.sipAgentCfg = &SIPAgentCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionTemplates(t *testing.T) {
	section := TemplatesJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.templates = FCTemplates{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.templates = FCTemplates{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionConfigS(t *testing.T) {
	section := ConfigSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.configSCfg = &ConfigSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.configSCfg = &ConfigSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAPIBan(t *testing.T) {
	section := APIBanJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.apiBanCfg = &APIBanCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.apiBanCfg = &APIBanCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCoreS(t *testing.T) {
	section := CoreSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.coreSCfg = &CoreSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.coreSCfg = &CoreSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionActionS(t *testing.T) {
	section := ActionSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.actionSCfg = &ActionSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.actionSCfg = &ActionSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAccountS(t *testing.T) {
	section := AccountSJSON

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.accountSCfg = &AccountSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.accountSCfg = &AccountSCfg{}

	if err := storeDiffSection(context.Background(), section, new(mockDb), cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestV1SetConfigErr1(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	args := &SetConfigArgs{
		Config: map[string]any{
			"cores": map[string]any{
				"caps":                "0",
				"caps_strategy":       "*busy",
				"caps_stats_interval": "0",
				"shutdown_timeout":    "1s",
			},
		},
	}

	var reply string
	expected := "json: cannot unmarshal string into Go struct field CoreSJsonCfg.Caps of type int"
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestV1SetConfigErr2(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	args := &SetConfigArgs{
		Config: map[string]any{
			"cdrs": map[string]any{
				"enabled":      false,
				"extra_fields": []string{},

				"session_cost_retries": 5,
				"chargers_conns":       []string{},
				"attributes_conns":     []string{},
				"thresholds_conns":     []string{},
				"stats_conns":          []string{},
				"online_cdr_exports":   []string{},
				"actions_conns":        []string{},
				"ees_conns":            []string{},
			},
		},
		DryRun: true,
	}
	var reply string
	cfg.sessionSCfg.Enabled = true
	cfg.sessionSCfg.TerminateAttempts = 0
	expected := "<SessionS> 'terminate_attempts' should be at least 1"
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestV1SetConfigErr3(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for key := range utils.StatelessDataDBPartitions {
		cfg.cacheCfg.Partitions[key].Limit = 0
	}
	args := &SetConfigArgs{
		Config: map[string]any{
			"cdrs": map[string]any{
				"enabled":              false,
				"extra_fields":         []string{},
				"session_cost_retries": 5,
				"chargers_conns":       []string{},
				"attributes_conns":     []string{},
				"thresholds_conns":     []string{},
				"stats_conns":          []string{},
				"online_cdr_exports":   []string{},
				"actions_conns":        []string{},
				"ees_conns":            []string{},
			},
		},
	}

	cfg.rldCh = make(chan string, 100)

	var reply string
	cfg.db = new(mockDb)
	expected := utils.ErrNotImplemented
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestLoadFromDBErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	expected := utils.ErrNotImplemented
	if err := cfg.LoadFromDB(context.Background(), new(mockDb)); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestLoadCfgFromDBErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	expected := utils.ErrNotImplemented
	sections := []string{"general"}
	if err := cfg.loadCfgFromDB(context.Background(), new(mockDb), sections, false); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}
func TestLoadCfgFromDBErr2(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	sections := []string{"test"}
	expected := "Invalid section: <test>"
	if err := cfg.loadCfgFromDB(context.Background(), new(mockDb), sections, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %v \n but received \n %v", new(mockDb), err)
	}
}

func TestV1GetConfig(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.cacheDP[GeneralJSON] = &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Default_request_type: utils.StringPointer(utils.MetaRated),
		Default_category:     utils.StringPointer(utils.Call),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Default_caching:      utils.StringPointer(utils.MetaReload),
		Connect_attempts:     utils.IntPointer(3),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("2s"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Max_parallel_conns:   utils.IntPointer(100),
	}
	args := &SectionWithAPIOpts{
		Sections: []string{GeneralJSON},
	}

	var reply map[string]any
	section := &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Default_request_type: utils.StringPointer(utils.MetaRated),
		Default_category:     utils.StringPointer(utils.Call),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Default_caching:      utils.StringPointer(utils.MetaReload),
		Connect_attempts:     utils.IntPointer(3),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("2s"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Max_parallel_conns:   utils.IntPointer(100),
	}
	expected := map[string]any{
		GeneralJSON: section,
	}
	if err := cfg.V1GetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n but received %+v \n", utils.ToJSON(expected), utils.ToJSON(reply))
	}

}

func TestApisLoadFromPathErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	path := "/tmp/test.json"
	errExpect := `path:"/tmp/test.json" is not reachable`
	if err := cfg.LoadFromPath(context.Background(), path); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}
}

func TestLoadCfgFromDBErr3(t *testing.T) {

	cfg := NewDefaultCGRConfig()
	sections := []string{"config_db", "test"}
	expected := "Invalid section: <test>"

	if err := cfg.loadCfgFromDB(context.Background(), new(mockDb), sections, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}

}

func TestLoadCfgFromDBErr4(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	sections := []string{"config_db"}
	expected := "Invalid section: <config_db>"
	if err := cfg.loadCfgFromDB(context.Background(), new(mockDb), sections, false); err == nil || err.Error() != expected {
		t.Errorf("Expected <%v> \n but received \n <%v><%T>", expected, err, err)
	}
}

func TestLoadCfgFromDBErr5(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	sections := []string{"config_db"}

	if err := cfg.loadCfgFromDB(context.Background(), new(mockDb), sections, true); err != nil {
		t.Errorf("Expected <nil> \n but received \n <%v>", err)
	}
}

func TestCGRConfigTpeSCfg(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	rcv := cfg.TpeSCfg()
	if rcv != cfg.tpeSCfg {
		t.Errorf("Expected <%v>, Received <%v>", cfg.tpeSCfg, rcv)
	}
}

func TestCGRConfigV1StoreCfgInDB(t *testing.T) {

	cfg := NewDefaultCGRConfig()
	cfg.rldCh = make(chan string, 100)
	cfg.db = &mockDb{
		GetSectionF: func(ctx *context.Context, s string, i any) error { return nil },
		SetSectionF: func(ctx *context.Context, s string, i any) error { return utils.ErrNotImplemented },
	}

	args := &SectionWithAPIOpts{
		Sections: []string{GeneralJSON},
	}

	var reply string

	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %v \n but received \n %v", utils.ErrNotImplemented, err)
	}
}
