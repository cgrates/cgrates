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
package config

import (
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestConfigV1SetConfigWithDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	cfg.db = db

	v2 := NewDefaultCGRConfig()
	v2.GeneralCfg().NodeID = "Test"
	v2.GeneralCfg().DefaultCaching = utils.MetaClear
	var reply string
	if err := cfg.V1SetConfig(context.Background(), &SetConfigArgs{
		Config: v2.AsMapInterface(utils.InfieldSep),
	}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestConfigV1StoreCfgInDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
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
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestConfigV1StoreCfgInDBErr1(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
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
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}

	generalJsonCfg := func() (*GeneralJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	cfg.db = &mockDb{
		GeneralJsonCfgF: generalJsonCfg,
	}

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
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}

	args := &SectionWithAPIOpts{
		Sections: []string{"cdrs"},
	}

	cfg.db = new(mockDb)
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), args, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Error(err)
	}
}

func TestConfigV1SetConfigFromJSONWithDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	cfg.db = db

	v2 := NewDefaultCGRConfig()
	v2.GeneralCfg().NodeID = "Test"
	v2.GeneralCfg().DefaultCaching = utils.MetaClear
	var reply string
	if err := cfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{
		Config: utils.ToJSON(v2.AsMapInterface(utils.InfieldSep)),
	}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestConfigV1SetConfigFromJSONWithDBErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	args := &SetConfigFromJSONArgs{
		Config: `{
			"cdrs":{
				"enabled": false,
				"store_cdrs": true,
			}
		}
		`,
	}

	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}

	cfg.db = &mockDb{
		GeneralJsonCfgF: nil,
	}
	var reply string
	expected := utils.ErrNotImplemented
	if err := cfg.V1SetConfigFromJSON(context.Background(), args, &reply); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, reply)
	}
}

func TestConfigLoadFromDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	g1 := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}

	if err := db.SetSection(context.Background(), GeneralJSON, g1); err != nil {
		t.Fatal(err)
	}

	if err := cfg.LoadFromDB(db); err != nil {
		t.Fatal(err)
	}
	expGeneral := &GeneralCfg{
		NodeID:           "Test",
		DefaultCaching:   utils.MetaClear,
		Logger:           utils.MetaSysLog,
		LogLevel:         6,
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
		PosterAttempts:   3,
		FailedPostsDir:   "/var/spool/cgrates/failed_posts",
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
		RSRSep:           ";",
		FailedPostsTTL:   5 * time.Second,
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
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestGetSectionAsMap(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	var cons []string
	expected := map[string]interface{}{
		"db_type":     utils.MetaInternal,
		"db_host":     "",
		"db_port":     0,
		"db_name":     "",
		"db_user":     "",
		"db_password": "",
		"items":       map[string]interface{}{},
		"opts": map[string]interface{}{
			"redis_sentinel":             "",
			"redis_cluster":              false,
			"redis_cluster_sync":         "5s",
			"redis_cluster_ondown_delay": "0",
			"query_timeout":              "10s",
			"redis_tls":                  false,
			"redis_client_certificate":   "",
			"redis_client_key":           "",
			"redis_ca_certificate":       "",
		},
		"remote_conn_id":       "",
		"remote_conns":         cons,
		"replication_cache":    "",
		"replication_conns":    cons,
		"replication_filtered": false,
	}

	rcv, err := cfg.getSectionAsMap(ConfigDBJSON)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		fmt.Printf("%T", rcv)
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

type mockDb struct {
	GeneralJsonCfgF         func() (*GeneralJsonCfg, error)
	RPCConnJsonCfgF         func() (RPCConnsJson, error)
	CacheJsonCfgF           func() (*CacheJsonCfg, error)
	ListenJsonCfgF          func() (*ListenJsonCfg, error)
	HttpJsonCfgF            func() (*HTTPJsonCfg, error)
	DbJsonCfgF              func(section string) (*DbJsonCfg, error)
	FilterSJsonCfgF         func() (*FilterSJsonCfg, error)
	CdrsJsonCfgF            func() (*CdrsJsonCfg, error)
	ERsJsonCfgF             func() (*ERsJsonCfg, error)
	EEsJsonCfgF             func() (*EEsJsonCfg, error)
	SessionSJsonCfgF        func() (*SessionSJsonCfg, error)
	FreeswitchAgentJsonCfgF func() (*FreeswitchAgentJsonCfg, error)
	KamAgentJsonCfgF        func() (*KamAgentJsonCfg, error)
	AsteriskAgentJsonCfgF   func() (*AsteriskAgentJsonCfg, error)
	DiameterAgentJsonCfgF   func() (*DiameterAgentJsonCfg, error)
	RadiusAgentJsonCfgF     func() (*RadiusAgentJsonCfg, error)
	HttpAgentJsonCfgF       func() (*[]*HttpAgentJsonCfg, error)
	DNSAgentJsonCfgF        func() (*DNSAgentJsonCfg, error)
	AttributeServJsonCfgF   func() (*AttributeSJsonCfg, error)
	ChargerServJsonCfgF     func() (*ChargerSJsonCfg, error)
	ResourceSJsonCfgF       func() (*ResourceSJsonCfg, error)
	StatSJsonCfgF           func() (*StatServJsonCfg, error)
	ThresholdSJsonCfgF      func() (*ThresholdSJsonCfg, error)
	RouteSJsonCfgF          func() (*RouteSJsonCfg, error)
	LoaderJsonCfgF          func() ([]*LoaderJsonCfg, error)
	SureTaxJsonCfgF         func() (*SureTaxJsonCfg, error)
	DispatcherSJsonCfgF     func() (*DispatcherSJsonCfg, error)
	RegistrarCJsonCfgsF     func() (*RegistrarCJsonCfgs, error)
	LoaderCfgJsonF          func() (*LoaderCfgJson, error)
	MigratorCfgJsonF        func() (*MigratorCfgJson, error)
	TlsCfgJsonF             func() (*TlsJsonCfg, error)
	AnalyzerCfgJsonF        func() (*AnalyzerSJsonCfg, error)
	AdminSCfgJsonF          func() (*AdminSJsonCfg, error)
	RateCfgJsonF            func() (*RateSJsonCfg, error)
	SIPAgentJsonCfgF        func() (*SIPAgentJsonCfg, error)
	TemplateSJsonCfgF       func() (FcTemplatesJsonCfg, error)
	ConfigSJsonCfgF         func() (*ConfigSCfgJson, error)
	ApiBanCfgJsonF          func() (*APIBanJsonCfg, error)
	CoreSJSONF              func() (*CoreSJsonCfg, error)
	ActionSCfgJsonF         func() (*ActionSJsonCfg, error)
	AccountSCfgJsonF        func() (*AccountSJsonCfg, error)
	SetSectionF             func(*context.Context, string, interface{}) error
}

func (m *mockDb) GeneralJsonCfg() (*GeneralJsonCfg, error) {
	if m.GeneralJsonCfgF != nil {
		return m.GeneralJsonCfgF()
	}
	return &GeneralJsonCfg{}, nil
}

func (m *mockDb) RPCConnJsonCfg() (RPCConnsJson, error) {
	if m.RPCConnJsonCfgF != nil {
		return m.RPCConnJsonCfgF()
	}
	return RPCConnsJson{}, nil
}
func (m *mockDb) CacheJsonCfg() (*CacheJsonCfg, error) {
	if m.CacheJsonCfgF != nil {
		return m.CacheJsonCfgF()
	}
	return &CacheJsonCfg{}, nil
}
func (m *mockDb) ListenJsonCfg() (*ListenJsonCfg, error) {
	if m.ListenJsonCfgF != nil {
		return m.ListenJsonCfgF()
	}
	return &ListenJsonCfg{}, nil
}
func (m *mockDb) HttpJsonCfg() (*HTTPJsonCfg, error) {
	if m.HttpJsonCfgF != nil {
		return m.HttpJsonCfgF()
	}
	return &HTTPJsonCfg{}, nil
}
func (m *mockDb) DbJsonCfg(section string) (*DbJsonCfg, error) {
	if m.DbJsonCfgF != nil {
		return m.DbJsonCfgF(section)
	}
	return &DbJsonCfg{}, nil
}
func (m *mockDb) FilterSJsonCfg() (*FilterSJsonCfg, error) {
	if m.FilterSJsonCfgF != nil {
		return m.FilterSJsonCfgF()
	}
	return &FilterSJsonCfg{}, nil
}
func (m *mockDb) CdrsJsonCfg() (*CdrsJsonCfg, error) {
	if m.CdrsJsonCfgF != nil {
		return m.CdrsJsonCfgF()
	}
	return &CdrsJsonCfg{}, nil
}
func (m *mockDb) ERsJsonCfg() (*ERsJsonCfg, error) {
	if m.ERsJsonCfgF != nil {
		return m.ERsJsonCfgF()
	}
	return &ERsJsonCfg{}, nil
}
func (m *mockDb) EEsJsonCfg() (*EEsJsonCfg, error) {
	if m.EEsJsonCfgF != nil {
		return m.EEsJsonCfgF()
	}
	return &EEsJsonCfg{}, nil
}
func (m *mockDb) SessionSJsonCfg() (*SessionSJsonCfg, error) {
	if m.SessionSJsonCfgF != nil {
		return m.SessionSJsonCfgF()
	}
	return &SessionSJsonCfg{}, nil
}
func (m *mockDb) FreeswitchAgentJsonCfg() (*FreeswitchAgentJsonCfg, error) {
	if m.FreeswitchAgentJsonCfgF != nil {
		return m.FreeswitchAgentJsonCfgF()
	}
	return &FreeswitchAgentJsonCfg{}, nil
}
func (m *mockDb) KamAgentJsonCfg() (*KamAgentJsonCfg, error) {
	if m.KamAgentJsonCfgF != nil {
		return m.KamAgentJsonCfgF()
	}
	return &KamAgentJsonCfg{}, nil
}
func (m *mockDb) AsteriskAgentJsonCfg() (*AsteriskAgentJsonCfg, error) {
	if m.AsteriskAgentJsonCfgF != nil {
		return m.AsteriskAgentJsonCfgF()
	}
	return &AsteriskAgentJsonCfg{}, nil
}
func (m *mockDb) DiameterAgentJsonCfg() (*DiameterAgentJsonCfg, error) {
	if m.DiameterAgentJsonCfgF != nil {
		return m.DiameterAgentJsonCfgF()
	}
	return &DiameterAgentJsonCfg{}, nil
}
func (m *mockDb) RadiusAgentJsonCfg() (*RadiusAgentJsonCfg, error) {
	if m.RadiusAgentJsonCfgF != nil {
		return m.RadiusAgentJsonCfgF()
	}
	return &RadiusAgentJsonCfg{}, nil
}
func (m *mockDb) HttpAgentJsonCfg() (*[]*HttpAgentJsonCfg, error) {
	if m.HttpAgentJsonCfgF != nil {
		return m.HttpAgentJsonCfgF()
	}
	return &[]*HttpAgentJsonCfg{}, nil
}
func (m *mockDb) DNSAgentJsonCfg() (*DNSAgentJsonCfg, error) {
	if m.DNSAgentJsonCfgF != nil {
		return m.DNSAgentJsonCfgF()
	}
	return &DNSAgentJsonCfg{}, nil
}
func (m *mockDb) AttributeServJsonCfg() (*AttributeSJsonCfg, error) {
	if m.AttributeServJsonCfgF != nil {
		return m.AttributeServJsonCfgF()
	}
	return &AttributeSJsonCfg{}, nil
}
func (m *mockDb) ChargerServJsonCfg() (*ChargerSJsonCfg, error) {
	if m.ChargerServJsonCfgF != nil {
		return m.ChargerServJsonCfgF()
	}
	return &ChargerSJsonCfg{}, nil
}
func (m *mockDb) ResourceSJsonCfg() (*ResourceSJsonCfg, error) {
	if m.ResourceSJsonCfgF != nil {
		return m.ResourceSJsonCfgF()
	}
	return &ResourceSJsonCfg{}, nil
}
func (m *mockDb) StatSJsonCfg() (*StatServJsonCfg, error) {
	if m.StatSJsonCfgF != nil {
		return m.StatSJsonCfgF()
	}
	return &StatServJsonCfg{}, nil
}
func (m *mockDb) ThresholdSJsonCfg() (*ThresholdSJsonCfg, error) {
	if m.ThresholdSJsonCfgF != nil {
		return m.ThresholdSJsonCfgF()
	}
	return &ThresholdSJsonCfg{}, nil
}
func (m *mockDb) RouteSJsonCfg() (*RouteSJsonCfg, error) {
	if m.RouteSJsonCfgF != nil {
		return m.RouteSJsonCfgF()
	}
	return &RouteSJsonCfg{}, nil
}
func (m *mockDb) LoaderJsonCfg() ([]*LoaderJsonCfg, error) {
	if m.LoaderJsonCfgF != nil {
		return m.LoaderJsonCfgF()
	}
	return []*LoaderJsonCfg{}, nil
}
func (m *mockDb) SureTaxJsonCfg() (*SureTaxJsonCfg, error) {
	if m.SureTaxJsonCfgF != nil {
		return m.SureTaxJsonCfgF()
	}
	return &SureTaxJsonCfg{}, nil
}
func (m *mockDb) DispatcherSJsonCfg() (*DispatcherSJsonCfg, error) {
	if m.DispatcherSJsonCfgF != nil {
		return m.DispatcherSJsonCfgF()
	}
	return &DispatcherSJsonCfg{}, nil
}
func (m *mockDb) RegistrarCJsonCfgs() (*RegistrarCJsonCfgs, error) {
	if m.RegistrarCJsonCfgsF != nil {
		return m.RegistrarCJsonCfgsF()
	}
	return &RegistrarCJsonCfgs{}, nil
}
func (m *mockDb) LoaderCfgJson() (*LoaderCfgJson, error) {
	if m.LoaderCfgJsonF != nil {
		return m.LoaderCfgJsonF()
	}
	return &LoaderCfgJson{}, nil
}
func (m *mockDb) MigratorCfgJson() (*MigratorCfgJson, error) {
	if m.MigratorCfgJsonF != nil {
		return m.MigratorCfgJsonF()
	}
	return &MigratorCfgJson{}, nil
}
func (m *mockDb) TlsCfgJson() (*TlsJsonCfg, error) {
	if m.TlsCfgJsonF != nil {
		return m.TlsCfgJsonF()
	}
	return &TlsJsonCfg{}, nil
}
func (m *mockDb) AnalyzerCfgJson() (*AnalyzerSJsonCfg, error) {
	if m.AnalyzerCfgJsonF != nil {
		return m.AnalyzerCfgJsonF()
	}
	return &AnalyzerSJsonCfg{}, nil
}
func (m *mockDb) AdminSCfgJson() (*AdminSJsonCfg, error) {
	if m.AdminSCfgJsonF != nil {
		return m.AdminSCfgJsonF()
	}
	return &AdminSJsonCfg{}, nil
}
func (m *mockDb) RateCfgJson() (*RateSJsonCfg, error) {
	if m.RateCfgJsonF != nil {
		return m.RateCfgJsonF()
	}
	return &RateSJsonCfg{}, nil
}
func (m *mockDb) SIPAgentJsonCfg() (*SIPAgentJsonCfg, error) {
	if m.SIPAgentJsonCfgF != nil {
		return m.SIPAgentJsonCfgF()
	}
	return &SIPAgentJsonCfg{}, nil
}
func (m *mockDb) TemplateSJsonCfg() (FcTemplatesJsonCfg, error) {
	if m.TemplateSJsonCfgF != nil {
		return m.TemplateSJsonCfgF()
	}
	return FcTemplatesJsonCfg{}, nil
}
func (m *mockDb) ConfigSJsonCfg() (*ConfigSCfgJson, error) {
	if m.ConfigSJsonCfgF != nil {
		return m.ConfigSJsonCfgF()
	}
	return &ConfigSCfgJson{}, nil
}
func (m *mockDb) ApiBanCfgJson() (*APIBanJsonCfg, error) {
	if m.ApiBanCfgJsonF != nil {
		return m.ApiBanCfgJsonF()
	}
	return &APIBanJsonCfg{}, nil
}
func (m *mockDb) CoreSJSON() (*CoreSJsonCfg, error) {
	if m.CoreSJSONF != nil {
		return m.CoreSJSONF()
	}
	return &CoreSJsonCfg{}, nil
}
func (m *mockDb) ActionSCfgJson() (*ActionSJsonCfg, error) {
	if m.ActionSCfgJsonF != nil {
		return m.ActionSCfgJsonF()
	}
	return &ActionSJsonCfg{}, nil
}
func (m *mockDb) AccountSCfgJson() (*AccountSJsonCfg, error) {
	if m.AccountSCfgJsonF != nil {
		return m.AccountSCfgJsonF()
	}
	return &AccountSJsonCfg{}, nil
}
func (m *mockDb) SetSection(*context.Context, string, interface{}) error {
	return utils.ErrNotImplemented
}

func TestStoreDiffSectionGeneral(t *testing.T) {
	section := GeneralJSON

	generalJsonCfg := func() (*GeneralJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		GeneralJsonCfgF: generalJsonCfg,
	}
	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.generalCfg = &GeneralCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.generalCfg = &GeneralCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRPCConns(t *testing.T) {
	section := RPCConnsJSON

	rpcConns := func() (RPCConnsJson, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		RPCConnJsonCfgF: rpcConns,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.rpcConns = RPCConns{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.rpcConns = RPCConns{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCache(t *testing.T) {
	section := CacheJSON

	cacheJsonCfg := func() (*CacheJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		CacheJsonCfgF: cacheJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.cacheCfg = &CacheCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.cacheCfg = &CacheCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionListen(t *testing.T) {
	section := ListenJSON

	listenJsonCfg := func() (*ListenJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ListenJsonCfgF: listenJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.listenCfg = &ListenCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.listenCfg = &ListenCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionHTTP(t *testing.T) {
	section := HTTPJSON

	httpJsonCfg := func() (*HTTPJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		HttpJsonCfgF: httpJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.httpCfg = &HTTPCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.httpCfg = &HTTPCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionStorDB(t *testing.T) {
	section := StorDBJSON

	storDbJsonCfg := func(section string) (*DbJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		DbJsonCfgF: storDbJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.storDbCfg = &StorDbCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.storDbCfg = &StorDbCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDataDB(t *testing.T) {
	section := DataDBJSON

	dataDbJsonCfg := func(section string) (*DbJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		DbJsonCfgF: dataDbJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dataDbCfg = &DataDbCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dataDbCfg = &DataDbCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionFilterS(t *testing.T) {
	section := FilterSJSON

	filterSJsonCfg := func() (*FilterSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		FilterSJsonCfgF: filterSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.filterSCfg = &FilterSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.filterSCfg = &FilterSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCDRs(t *testing.T) {
	section := CDRsJSON

	cdrsJsonCfg := func() (*CdrsJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		CdrsJsonCfgF: cdrsJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.cdrsCfg = &CdrsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.cdrsCfg = &CdrsCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionERs(t *testing.T) {
	section := ERsJSON

	erSJsonCfg := func() (*ERsJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ERsJsonCfgF: erSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.ersCfg = &ERsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.ersCfg = &ERsCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionEEs(t *testing.T) {
	section := EEsJSON

	eeSJsonCfg := func() (*EEsJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		EEsJsonCfgF: eeSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.eesCfg = &EEsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.eesCfg = &EEsCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSessionS(t *testing.T) {
	section := SessionSJSON

	sessionSJsonCfg := func() (*SessionSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		SessionSJsonCfgF: sessionSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sessionSCfg = &SessionSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.sessionSCfg = &SessionSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionFreeSWITCH(t *testing.T) {
	section := FreeSWITCHAgentJSON

	freeswitchAgentJsonCfg := func() (*FreeswitchAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		FreeswitchAgentJsonCfgF: freeswitchAgentJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.fsAgentCfg = &FsAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.fsAgentCfg = &FsAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionKamailio(t *testing.T) {
	section := KamailioAgentJSON

	kamailioJsonCfg := func() (*KamAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		KamAgentJsonCfgF: kamailioJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.kamAgentCfg = &KamAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.kamAgentCfg = &KamAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAsterisk(t *testing.T) {
	section := AsteriskAgentJSON

	asteriskJsonCfg := func() (*AsteriskAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		AsteriskAgentJsonCfgF: asteriskJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.asteriskAgentCfg = &AsteriskAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.asteriskAgentCfg = &AsteriskAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDiameter(t *testing.T) {
	section := DiameterAgentJSON

	diameterJsonCfg := func() (*DiameterAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		DiameterAgentJsonCfgF: diameterJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.diameterAgentCfg = &DiameterAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.diameterAgentCfg = &DiameterAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRadius(t *testing.T) {
	section := RadiusAgentJSON

	radiusJsonCfg := func() (*RadiusAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		RadiusAgentJsonCfgF: radiusJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.diameterAgentCfg = &DiameterAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.diameterAgentCfg = &DiameterAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionHTTPAgent(t *testing.T) {
	section := HTTPAgentJSON

	httpAgentJsonCfg := func() (*[]*HttpAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		HttpAgentJsonCfgF: httpAgentJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.httpAgentCfg = HTTPAgentCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.httpAgentCfg = HTTPAgentCfgs{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDNS(t *testing.T) {
	section := DNSAgentJSON

	dnsJsonCfg := func() (*DNSAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		DNSAgentJsonCfgF: dnsJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dnsAgentCfg = &DNSAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dnsAgentCfg = &DNSAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAttributeS(t *testing.T) {
	section := AttributeSJSON

	attributeSJsonCfg := func() (*AttributeSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		AttributeServJsonCfgF: attributeSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.attributeSCfg = &AttributeSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.attributeSCfg = &AttributeSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionChargerS(t *testing.T) {
	section := ChargerSJSON

	chargerSJsonCfg := func() (*ChargerSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ChargerServJsonCfgF: chargerSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.chargerSCfg = &ChargerSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.chargerSCfg = &ChargerSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionResourceS(t *testing.T) {
	section := ResourceSJSON

	resourceSJsonCfg := func() (*ResourceSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ResourceSJsonCfgF: resourceSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.resourceSCfg = &ResourceSConfig{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.resourceSCfg = &ResourceSConfig{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionStatS(t *testing.T) {
	section := StatSJSON

	statSJsonCfg := func() (*StatServJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		StatSJsonCfgF: statSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.statsCfg = &StatSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.statsCfg = &StatSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionThresholdS(t *testing.T) {
	section := ThresholdSJSON

	thresholdSJsonCfg := func() (*ThresholdSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ThresholdSJsonCfgF: thresholdSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.thresholdSCfg = &ThresholdSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.thresholdSCfg = &ThresholdSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRouteS(t *testing.T) {
	section := RouteSJSON

	routeSJsonCfg := func() (*RouteSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		RouteSJsonCfgF: routeSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.routeSCfg = &RouteSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.routeSCfg = &RouteSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLoaderS(t *testing.T) {
	section := LoaderSJSON

	loaderSJsonCfg := func() ([]*LoaderJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		LoaderJsonCfgF: loaderSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loaderCfg = LoaderSCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loaderCfg = LoaderSCfgs{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSureTax(t *testing.T) {
	section := SureTaxJSON

	sureTaxJsonCfg := func() (*SureTaxJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		SureTaxJsonCfgF: sureTaxJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sureTaxCfg = &SureTaxCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.sureTaxCfg = &SureTaxCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDispatcherS(t *testing.T) {
	section := DispatcherSJSON

	dispatcherSJsonCfg := func() (*DispatcherSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		DispatcherSJsonCfgF: dispatcherSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dispatcherSCfg = &DispatcherSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dispatcherSCfg = &DispatcherSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRegistrarC(t *testing.T) {
	section := RegistrarCJSON

	registrarCJsonCfgs := func() (*RegistrarCJsonCfgs, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		RegistrarCJsonCfgsF: registrarCJsonCfgs,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.registrarCCfg = &RegistrarCCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.registrarCCfg = &RegistrarCCfgs{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLoader(t *testing.T) {
	section := LoaderJSON

	loaderJsonCfg := func() (*LoaderCfgJson, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		LoaderCfgJsonF: loaderJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loaderCgrCfg = &LoaderCgrCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loaderCgrCfg = &LoaderCgrCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionMigrator(t *testing.T) {
	section := MigratorJSON

	migratorJsonCfg := func() (*MigratorCfgJson, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		MigratorCfgJsonF: migratorJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.migratorCgrCfg = &MigratorCgrCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.migratorCgrCfg = &MigratorCgrCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionTls(t *testing.T) {
	section := TlsJSON

	tlsJsonCfg := func() (*TlsJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		TlsCfgJsonF: tlsJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.tlsCfg = &TLSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.tlsCfg = &TLSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAnalyzerS(t *testing.T) {
	section := AnalyzerSJSON

	analyzerJsonCfg := func() (*AnalyzerSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		AnalyzerCfgJsonF: analyzerJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.analyzerSCfg = &AnalyzerSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.analyzerSCfg = &AnalyzerSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAdminS(t *testing.T) {
	section := AdminSJSON

	adminSJsonCfg := func() (*AdminSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		AdminSCfgJsonF: adminSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.admS = &AdminSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.admS = &AdminSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRateS(t *testing.T) {
	section := RateSJSON

	rateSJsonCfg := func() (*RateSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		RateCfgJsonF: rateSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.rateSCfg = &RateSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.rateSCfg = &RateSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSIP(t *testing.T) {
	section := SIPAgentJSON

	sipAgentJsonCfg := func() (*SIPAgentJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		SIPAgentJsonCfgF: sipAgentJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sipAgentCfg = &SIPAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV1.sipAgentCfg = &SIPAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionTemplates(t *testing.T) {
	section := TemplatesJSON

	templatesJsonCfg := func() (FcTemplatesJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		TemplateSJsonCfgF: templatesJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.templates = FCTemplates{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.templates = FCTemplates{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionConfigS(t *testing.T) {
	section := ConfigSJSON

	configSJsonCfg := func() (*ConfigSCfgJson, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ConfigSJsonCfgF: configSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.configSCfg = &ConfigSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.configSCfg = &ConfigSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAPIBan(t *testing.T) {
	section := APIBanJSON

	apiBanJsonCfg := func() (*APIBanJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ApiBanCfgJsonF: apiBanJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.apiBanCfg = &APIBanCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.apiBanCfg = &APIBanCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCoreS(t *testing.T) {
	section := CoreSJSON

	coreSJsonCfg := func() (*CoreSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		CoreSJSONF: coreSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.coreSCfg = &CoreSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.coreSCfg = &CoreSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionActionS(t *testing.T) {
	section := ActionSJSON

	actionSJsonCfg := func() (*ActionSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		ActionSCfgJsonF: actionSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.actionSCfg = &ActionSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.actionSCfg = &ActionSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAccountS(t *testing.T) {
	section := AccountSJSON

	accountSJsonCfg := func() (*AccountSJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		AccountSCfgJsonF: accountSJsonCfg,
	}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.accountSCfg = &AccountSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.accountSCfg = &AccountSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrNotImplemented || err == nil {
		t.Error(err)
	}
}

func TestV1ReloadConfig(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.db = &CgrJsonCfg{}
	cfg.ConfigPath = path.Join("/usr", "share", "cgrates", "conf", "samples", "tutmongo2")
	args := &ReloadArgs{
		Section: utils.MetaAll,
	}

	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}

	var reply string
	if err := cfg.V1ReloadConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Errorf("Expected %v \n but received \n %v", "OK", reply)
	}

	args = &ReloadArgs{
		Section: ConfigDBJSON,
	}

	expected := "Invalid section: <config_db> "
	if err := cfg.V1ReloadConfig(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("%T and %T", expected, err.Error())
		t.Errorf("Expected %q \n but received \n %q", expected, err)
	}
}

func TestV1SetConfigErr1(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	args := &SetConfigArgs{
		Config: map[string]interface{}{
			"cores": map[string]interface{}{
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
		Config: map[string]interface{}{
			"cdrs": map[string]interface{}{
				"enabled":              false,
				"extra_fields":         []string{},
				"store_cdrs":           true,
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
	args := &SetConfigArgs{
		Config: map[string]interface{}{
			"cdrs": map[string]interface{}{
				"enabled":              false,
				"extra_fields":         []string{},
				"store_cdrs":           true,
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

	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}

	var reply string
	cfg.db = &mockDb{
		GeneralJsonCfgF: nil,
	}
	expected := utils.ErrNotImplemented
	if err := cfg.V1SetConfig(context.Background(), args, &reply); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestLoadFromDBErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	generalJsonCfg := func() (*GeneralJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	jsnCfg := &mockDb{
		GeneralJsonCfgF: generalJsonCfg,
	}
	expected := utils.ErrNotImplemented
	if err := cfg.LoadFromDB(jsnCfg); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestLoadCfgFromDBErr(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	generalJsonCfg := func() (*GeneralJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	jsnCfg := &mockDb{
		GeneralJsonCfgF: generalJsonCfg,
	}
	expected := utils.ErrNotImplemented
	sections := []string{"general"}
	if err := cfg.loadCfgFromDB(jsnCfg, sections, false); err == nil || err != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestLoadCfgFromDBErr2(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	generalJsonCfg := func() (*GeneralJsonCfg, error) {

		return nil, utils.ErrNotImplemented
	}
	db := &mockDb{
		GeneralJsonCfgF: generalJsonCfg,
	}
	sections := []string{"test"}
	expected := "Invalid section: <test> "
	if err := cfg.loadCfgFromDB(db, sections, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %v \n but received \n %v", expected, err)
	}
}

func TestV1GetConfig(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	cfg.cacheDP[GeneralJSON] = &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
		Logger:               utils.StringPointer(utils.MetaSysLog),
		Log_level:            utils.IntPointer(6),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Poster_attempts:      utils.IntPointer(3),
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
		Rsr_separator:        utils.StringPointer(";"),
		Digest_equal:         utils.StringPointer(":"),
		Failed_posts_ttl:     utils.StringPointer("2ns"),
		Max_parallel_conns:   utils.IntPointer(100),
	}
	args := &SectionWithAPIOpts{
		Sections: []string{GeneralJSON},
	}

	var reply map[string]interface{}
	section := &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
		Logger:               utils.StringPointer(utils.MetaSysLog),
		Log_level:            utils.IntPointer(6),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Poster_attempts:      utils.IntPointer(3),
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
		Rsr_separator:        utils.StringPointer(";"),
		Digest_equal:         utils.StringPointer(":"),
		Failed_posts_ttl:     utils.StringPointer("2ns"),
		Max_parallel_conns:   utils.IntPointer(100),
	}
	expected := map[string]interface{}{
		GeneralJSON: section,
	}
	if err := cfg.V1GetConfig(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n but received %+v \n", utils.ToJSON(expected), utils.ToJSON(reply))
	}

}
