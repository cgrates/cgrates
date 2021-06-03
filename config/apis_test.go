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

type mockDb struct{}

func (mockDb) GeneralJsonCfg() (*GeneralJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}

func (mockDb) RPCConnJsonCfg() (RPCConnsJson, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) CacheJsonCfg() (*CacheJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ListenJsonCfg() (*ListenJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) HttpJsonCfg() (*HTTPJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) DbJsonCfg(section string) (*DbJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) FilterSJsonCfg() (*FilterSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) CdrsJsonCfg() (*CdrsJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ERsJsonCfg() (*ERsJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) EEsJsonCfg() (*EEsJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) SessionSJsonCfg() (*SessionSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) FreeswitchAgentJsonCfg() (*FreeswitchAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) KamAgentJsonCfg() (*KamAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) AsteriskAgentJsonCfg() (*AsteriskAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) DiameterAgentJsonCfg() (*DiameterAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) RadiusAgentJsonCfg() (*RadiusAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) HttpAgentJsonCfg() (*[]*HttpAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) DNSAgentJsonCfg() (*DNSAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) AttributeServJsonCfg() (*AttributeSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ChargerServJsonCfg() (*ChargerSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ResourceSJsonCfg() (*ResourceSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) StatSJsonCfg() (*StatServJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ThresholdSJsonCfg() (*ThresholdSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) RouteSJsonCfg() (*RouteSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) LoaderJsonCfg() ([]*LoaderJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) SureTaxJsonCfg() (*SureTaxJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) DispatcherSJsonCfg() (*DispatcherSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) RegistrarCJsonCfgs() (*RegistrarCJsonCfgs, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) LoaderCfgJson() (*LoaderCfgJson, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) MigratorCfgJson() (*MigratorCfgJson, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) TlsCfgJson() (*TlsJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) AnalyzerCfgJson() (*AnalyzerSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) AdminSCfgJson() (*AdminSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) RateCfgJson() (*RateSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) SIPAgentJsonCfg() (*SIPAgentJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) TemplateSJsonCfg() (FcTemplatesJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ConfigSJsonCfg() (*ConfigSCfgJson, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ApiBanCfgJson() (*APIBanJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) CoreSJSON() (*CoreSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) ActionSCfgJson() (*ActionSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) AccountSCfgJson() (*AccountSJsonCfg, error) {
	return nil, utils.ErrAccountNotFound
}
func (mockDb) SetSection(*context.Context, string, interface{}) error {
	return utils.ErrAccountNotFound
}

func TestStoreDiffSectionGeneral(t *testing.T) {
	section := GeneralJSON

	db := &mockDb{}
	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.generalCfg = &GeneralCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.generalCfg = &GeneralCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRPCConns(t *testing.T) {
	section := RPCConnsJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.rpcConns = RPCConns{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.rpcConns = RPCConns{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCache(t *testing.T) {
	section := CacheJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.cacheCfg = &CacheCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.cacheCfg = &CacheCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionListen(t *testing.T) {
	section := ListenJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.listenCfg = &ListenCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.listenCfg = &ListenCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionHTTP(t *testing.T) {
	section := HTTPJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.httpCfg = &HTTPCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.httpCfg = &HTTPCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionStorDB(t *testing.T) {
	section := StorDBJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.storDbCfg = &StorDbCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.storDbCfg = &StorDbCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDataDB(t *testing.T) {
	section := DataDBJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dataDbCfg = &DataDbCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dataDbCfg = &DataDbCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionFilterS(t *testing.T) {
	section := FilterSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.filterSCfg = &FilterSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.filterSCfg = &FilterSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCDRs(t *testing.T) {
	section := CDRsJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.cdrsCfg = &CdrsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.cdrsCfg = &CdrsCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionERs(t *testing.T) {
	section := ERsJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.ersCfg = &ERsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.ersCfg = &ERsCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionEEs(t *testing.T) {
	section := EEsJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.eesCfg = &EEsCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.eesCfg = &EEsCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSessionS(t *testing.T) {
	section := SessionSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sessionSCfg = &SessionSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.sessionSCfg = &SessionSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionFreeSWITCH(t *testing.T) {
	section := FreeSWITCHAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.fsAgentCfg = &FsAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.fsAgentCfg = &FsAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionKamailio(t *testing.T) {
	section := KamailioAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.kamAgentCfg = &KamAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.kamAgentCfg = &KamAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAsterisk(t *testing.T) {
	section := AsteriskAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.asteriskAgentCfg = &AsteriskAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.asteriskAgentCfg = &AsteriskAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDiameter(t *testing.T) {
	section := DiameterAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.diameterAgentCfg = &DiameterAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.diameterAgentCfg = &DiameterAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRadius(t *testing.T) {
	section := RadiusAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.diameterAgentCfg = &DiameterAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.diameterAgentCfg = &DiameterAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionHTTPAgent(t *testing.T) {
	section := HTTPAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.httpAgentCfg = HTTPAgentCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.httpAgentCfg = HTTPAgentCfgs{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDNS(t *testing.T) {
	section := DNSAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dnsAgentCfg = &DNSAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dnsAgentCfg = &DNSAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAttributeS(t *testing.T) {
	section := AttributeSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.attributeSCfg = &AttributeSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.attributeSCfg = &AttributeSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionChargerS(t *testing.T) {
	section := ChargerSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.chargerSCfg = &ChargerSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.chargerSCfg = &ChargerSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionResourceS(t *testing.T) {
	section := ResourceSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.resourceSCfg = &ResourceSConfig{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.resourceSCfg = &ResourceSConfig{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionStatS(t *testing.T) {
	section := StatSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.statsCfg = &StatSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.statsCfg = &StatSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionThresholdS(t *testing.T) {
	section := ThresholdSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.thresholdSCfg = &ThresholdSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.thresholdSCfg = &ThresholdSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRouteS(t *testing.T) {
	section := RouteSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.routeSCfg = &RouteSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.routeSCfg = &RouteSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLoaderS(t *testing.T) {
	section := LoaderSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loaderCfg = LoaderSCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loaderCfg = LoaderSCfgs{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSureTax(t *testing.T) {
	section := SureTaxJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sureTaxCfg = &SureTaxCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.sureTaxCfg = &SureTaxCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionDispatcherS(t *testing.T) {
	section := DispatcherSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.dispatcherSCfg = &DispatcherSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.dispatcherSCfg = &DispatcherSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRegistrarC(t *testing.T) {
	section := RegistrarCJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.registrarCCfg = &RegistrarCCfgs{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.registrarCCfg = &RegistrarCCfgs{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionLoader(t *testing.T) {
	section := LoaderJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.loaderCgrCfg = &LoaderCgrCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.loaderCgrCfg = &LoaderCgrCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionMigrator(t *testing.T) {
	section := MigratorJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.migratorCgrCfg = &MigratorCgrCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.migratorCgrCfg = &MigratorCgrCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionTls(t *testing.T) {
	section := TlsJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.tlsCfg = &TLSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.tlsCfg = &TLSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAnalyzerS(t *testing.T) {
	section := AnalyzerSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.analyzerSCfg = &AnalyzerSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.analyzerSCfg = &AnalyzerSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAdminS(t *testing.T) {
	section := AdminSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.admS = &AdminSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.admS = &AdminSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionRateS(t *testing.T) {
	section := RateSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.rateSCfg = &RateSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.rateSCfg = &RateSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionSIP(t *testing.T) {
	section := SIPAgentJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.sipAgentCfg = &SIPAgentCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV1.sipAgentCfg = &SIPAgentCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionTemplates(t *testing.T) {
	section := TemplatesJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.templates = FCTemplates{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.templates = FCTemplates{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionConfigS(t *testing.T) {
	section := ConfigSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.configSCfg = &ConfigSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.configSCfg = &ConfigSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAPIBan(t *testing.T) {
	section := APIBanJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.apiBanCfg = &APIBanCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.apiBanCfg = &APIBanCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionCoreS(t *testing.T) {
	section := CoreSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.coreSCfg = &CoreSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.coreSCfg = &CoreSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionActionS(t *testing.T) {
	section := ActionSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.actionSCfg = &ActionSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.actionSCfg = &ActionSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
		t.Error(err)
	}
}

func TestStoreDiffSectionAccountS(t *testing.T) {
	section := AccountSJSON
	db := &mockDb{}

	cgrCfgV1 := NewDefaultCGRConfig()
	cgrCfgV1.accountSCfg = &AccountSCfg{}

	cgrCfgV2 := NewDefaultCGRConfig()
	cgrCfgV2.accountSCfg = &AccountSCfg{}

	if err := storeDiffSection(context.Background(), section, db, cgrCfgV1, cgrCfgV2); err != utils.ErrAccountNotFound || err == nil {
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

	cfg.rldChans[DataDBJSON] = make(chan struct{}, 1)
	cfg.rldChans[StorDBJSON] = make(chan struct{}, 1)
	cfg.rldChans[RPCConnsJSON] = make(chan struct{}, 1)
	cfg.rldChans[HTTPJSON] = make(chan struct{}, 1)
	cfg.rldChans[CDRsJSON] = make(chan struct{}, 1)
	cfg.rldChans[ERsJSON] = make(chan struct{}, 1)
	cfg.rldChans[SessionSJSON] = make(chan struct{}, 1)
	cfg.rldChans[AsteriskAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[FreeSWITCHAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[KamailioAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[DiameterAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[RadiusAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[HTTPAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[DNSAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[AttributeSJSON] = make(chan struct{}, 1)
	cfg.rldChans[ChargerSJSON] = make(chan struct{}, 1)
	cfg.rldChans[ResourceSJSON] = make(chan struct{}, 1)
	cfg.rldChans[StatSJSON] = make(chan struct{}, 1)
	cfg.rldChans[ThresholdSJSON] = make(chan struct{}, 1)
	cfg.rldChans[RouteSJSON] = make(chan struct{}, 1)
	cfg.rldChans[LoaderSJSON] = make(chan struct{}, 1)
	cfg.rldChans[DispatcherSJSON] = make(chan struct{}, 1)
	cfg.rldChans[AnalyzerSJSON] = make(chan struct{}, 1)
	cfg.rldChans[AdminSJSON] = make(chan struct{}, 1)
	cfg.rldChans[EEsJSON] = make(chan struct{}, 1)
	cfg.rldChans[SIPAgentJSON] = make(chan struct{}, 1)
	cfg.rldChans[RateSJSON] = make(chan struct{}, 1)
	cfg.rldChans[RegistrarCJSON] = make(chan struct{}, 1)
	cfg.rldChans[AccountSJSON] = make(chan struct{}, 1)
	cfg.rldChans[ActionSJSON] = make(chan struct{}, 1)

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
