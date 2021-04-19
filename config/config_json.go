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
	"encoding/json"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	GeneralJSON         = "general"
	CacheJSON           = "caches"
	ListenJSON          = "listen"
	HTTPJSON            = "http"
	DataDBJSON          = "data_db"
	StorDBJSON          = "stor_db"
	FilterSJSON         = "filters"
	CDRsJSON            = "cdrs"
	SessionSJSON        = "sessions"
	FreeSWITCHAgentJSON = "freeswitch_agent"
	KamailioAgentJSON   = "kamailio_agent"
	AsteriskAgentJSON   = "asterisk_agent"
	DiameterAgentJSON   = "diameter_agent"
	RadiusAgentJSON     = "radius_agent"
	HTTPAgentJSON       = "http_agent"
	AttributeSJSON      = "attributes"
	ResourceSJSON       = "resources"
	StatSJSON           = "stats"
	ThresholdSJSON      = "thresholds"
	RouteSJSON          = "routes"
	LoaderSJSON         = "loaders"
	MailerJSON          = "mailer"
	SureTaxJSON         = "suretax"
	DispatcherSJSON     = "dispatchers"
	RegistrarCJSON      = "registrarc"
	LoaderJSON          = "loader"
	MigratorJSON        = "migrator"
	ChargerSJSON        = "chargers"
	TlsJSON             = "tls"
	AnalyzerSJSON       = "analyzers"
	AdminSJSON          = "admins"
	DNSAgentJSON        = "dns_agent"
	ERsJSON             = "ers"
	EEsJSON             = "ees"
	RateSJSON           = "rates"
	ActionSJSON         = "actions"
	RPCConnsJSON        = "rpc_conns"
	SIPAgentJSON        = "sip_agent"
	TemplatesJSON       = "templates"
	ConfigSJSON         = "configs"
	APIBanJSON          = "apiban"
	CoreSJSON           = "cores"
	AccountSJSON        = "accounts"
	ConfigDBJSON        = "config_db"
)

var (
	sortedCfgSections = []string{GeneralJSON, RPCConnsJSON, DataDBJSON, StorDBJSON, ListenJSON, TlsJSON, HTTPJSON,
		CacheJSON, FilterSJSON, CDRsJSON, ERsJSON, SessionSJSON, AsteriskAgentJSON, FreeSWITCHAgentJSON,
		KamailioAgentJSON, DiameterAgentJSON, RadiusAgentJSON, HTTPAgentJSON, DNSAgentJSON, AttributeSJSON,
		ChargerSJSON, ResourceSJSON, StatSJSON, ThresholdSJSON, RouteSJSON, LoaderSJSON, MailerJSON, SureTaxJSON,
		LoaderJSON, MigratorJSON, DispatcherSJSON, AnalyzerSJSON, AdminSJSON, EEsJSON, RateSJSON, SIPAgentJSON,
		RegistrarCJSON, TemplatesJSON, ConfigSJSON, APIBanJSON, CoreSJSON, ActionSJSON, AccountSJSON, ConfigDBJSON}
	sortedSectionsSet = utils.NewStringSet(sortedCfgSections)
)

type ConfigDB interface {
	GeneralJsonCfg() (*GeneralJsonCfg, error)
	RPCConnJsonCfg() (RPCConnsJson, error)
	CacheJsonCfg() (*CacheJsonCfg, error)
	ListenJsonCfg() (*ListenJsonCfg, error)
	HttpJsonCfg() (*HTTPJsonCfg, error)
	DbJsonCfg(section string) (*DbJsonCfg, error)
	FilterSJsonCfg() (*FilterSJsonCfg, error)
	CdrsJsonCfg() (*CdrsJsonCfg, error)
	ERsJsonCfg() (*ERsJsonCfg, error)
	EEsJsonCfg() (*EEsJsonCfg, error)
	SessionSJsonCfg() (*SessionSJsonCfg, error)
	FreeswitchAgentJsonCfg() (*FreeswitchAgentJsonCfg, error)
	KamAgentJsonCfg() (*KamAgentJsonCfg, error)
	AsteriskAgentJsonCfg() (*AsteriskAgentJsonCfg, error)
	DiameterAgentJsonCfg() (*DiameterAgentJsonCfg, error)
	RadiusAgentJsonCfg() (*RadiusAgentJsonCfg, error)
	HttpAgentJsonCfg() (*[]*HttpAgentJsonCfg, error)
	DNSAgentJsonCfg() (*DNSAgentJsonCfg, error)
	AttributeServJsonCfg() (*AttributeSJsonCfg, error)
	ChargerServJsonCfg() (*ChargerSJsonCfg, error)
	ResourceSJsonCfg() (*ResourceSJsonCfg, error)
	StatSJsonCfg() (*StatServJsonCfg, error)
	ThresholdSJsonCfg() (*ThresholdSJsonCfg, error)
	RouteSJsonCfg() (*RouteSJsonCfg, error)
	LoaderJsonCfg() ([]*LoaderJsonCfg, error)
	MailerJsonCfg() (*MailerJsonCfg, error)
	SureTaxJsonCfg() (*SureTaxJsonCfg, error)
	DispatcherSJsonCfg() (*DispatcherSJsonCfg, error)
	RegistrarCJsonCfgs() (*RegistrarCJsonCfgs, error)
	LoaderCfgJson() (*LoaderCfgJson, error)
	MigratorCfgJson() (*MigratorCfgJson, error)
	TlsCfgJson() (*TlsJsonCfg, error)
	AnalyzerCfgJson() (*AnalyzerSJsonCfg, error)
	AdminSCfgJson() (*AdminSJsonCfg, error)
	RateCfgJson() (*RateSJsonCfg, error)
	SIPAgentJsonCfg() (*SIPAgentJsonCfg, error)
	TemplateSJsonCfg() (FcTemplatesJsonCfg, error)
	ConfigSJsonCfg() (*ConfigSCfgJson, error)
	ApiBanCfgJson() (*APIBanJsonCfg, error)
	CoreSJSON() (*CoreSJsonCfg, error)
	ActionSCfgJson() (*ActionSJsonCfg, error)
	AccountSCfgJson() (*AccountSJsonCfg, error)
	SetSection(*context.Context, string, interface{}) error
	ConfigDBJsonCfg() (*ConfigDBJsonCfg, error) // only used when loading from file
}

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromBytes(buf []byte) (cgrJsonCfg *CgrJsonCfg, err error) {
	cgrJsonCfg = new(CgrJsonCfg)
	err = NewRjReaderFromBytes(buf).Decode(cgrJsonCfg)
	return
}

// Main object holding the loaded config as section raw messages
type CgrJsonCfg map[string]*json.RawMessage

func (jsnCfg CgrJsonCfg) GeneralJsonCfg() (*GeneralJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[GeneralJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(GeneralJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RPCConnJsonCfg() (RPCConnsJson, error) {
	rawCfg, hasKey := jsnCfg[RPCConnsJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := make(RPCConnsJson)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) CacheJsonCfg() (*CacheJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[CacheJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CacheJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ListenJsonCfg() (*ListenJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ListenJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ListenJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) HttpJsonCfg() (*HTTPJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[HTTPJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(HTTPJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) DbJsonCfg(section string) (*DbJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[section]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DbJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) FilterSJsonCfg() (*FilterSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[FilterSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(FilterSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) CdrsJsonCfg() (*CdrsJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[CDRsJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CdrsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ERsJsonCfg() (erSCfg *ERsJsonCfg, err error) {
	rawCfg, hasKey := jsnCfg[ERsJSON]
	if !hasKey {
		return
	}
	erSCfg = new(ERsJsonCfg)
	err = json.Unmarshal(*rawCfg, &erSCfg)
	return
}

func (jsnCfg CgrJsonCfg) EEsJsonCfg() (erSCfg *EEsJsonCfg, err error) {
	rawCfg, hasKey := jsnCfg[EEsJSON]
	if !hasKey {
		return
	}
	erSCfg = new(EEsJsonCfg)
	err = json.Unmarshal(*rawCfg, &erSCfg)
	return
}

func (jsnCfg CgrJsonCfg) SessionSJsonCfg() (*SessionSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[SessionSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SessionSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) FreeswitchAgentJsonCfg() (*FreeswitchAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[FreeSWITCHAgentJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(FreeswitchAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) KamAgentJsonCfg() (*KamAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[KamailioAgentJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(KamAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) AsteriskAgentJsonCfg() (*AsteriskAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[AsteriskAgentJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AsteriskAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) DiameterAgentJsonCfg() (*DiameterAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[DiameterAgentJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DiameterAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RadiusAgentJsonCfg() (*RadiusAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[RadiusAgentJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RadiusAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) HttpAgentJsonCfg() (*[]*HttpAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[HTTPAgentJSON]
	if !hasKey {
		return nil, nil
	}
	httpAgnt := make([]*HttpAgentJsonCfg, 0)
	if err := json.Unmarshal(*rawCfg, &httpAgnt); err != nil {
		return nil, err
	}
	return &httpAgnt, nil
}

func (jsnCfg CgrJsonCfg) DNSAgentJsonCfg() (da *DNSAgentJsonCfg, err error) {
	rawCfg, hasKey := jsnCfg[DNSAgentJSON]
	if !hasKey {
		return
	}
	da = new(DNSAgentJsonCfg)
	err = json.Unmarshal(*rawCfg, da)
	return
}

func (jsnCfg CgrJsonCfg) AttributeServJsonCfg() (*AttributeSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[AttributeSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AttributeSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ChargerServJsonCfg() (*ChargerSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ChargerSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ChargerSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ResourceSJsonCfg() (*ResourceSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ResourceSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ResourceSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) StatSJsonCfg() (*StatServJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[StatSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(StatServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ThresholdSJsonCfg() (*ThresholdSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ThresholdSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ThresholdSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RouteSJsonCfg() (*RouteSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[RouteSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RouteSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) LoaderJsonCfg() ([]*LoaderJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[LoaderSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := make([]*LoaderJsonCfg, 0)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) MailerJsonCfg() (*MailerJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[MailerJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(MailerJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) SureTaxJsonCfg() (*SureTaxJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[SureTaxJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SureTaxJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) DispatcherSJsonCfg() (*DispatcherSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[DispatcherSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DispatcherSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RegistrarCJsonCfgs() (*RegistrarCJsonCfgs, error) {
	rawCfg, hasKey := jsnCfg[RegistrarCJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RegistrarCJsonCfgs)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) LoaderCfgJson() (*LoaderCfgJson, error) {
	rawCfg, hasKey := jsnCfg[LoaderJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(LoaderCfgJson)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) MigratorCfgJson() (*MigratorCfgJson, error) {
	rawCfg, hasKey := jsnCfg[MigratorJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(MigratorCfgJson)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) TlsCfgJson() (*TlsJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[TlsJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(TlsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) AnalyzerCfgJson() (*AnalyzerSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[AnalyzerSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AnalyzerSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) AdminSCfgJson() (*AdminSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[AdminSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AdminSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RateCfgJson() (*RateSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[RateSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RateSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) SIPAgentJsonCfg() (*SIPAgentJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[SIPAgentJSON]
	if !hasKey {
		return nil, nil
	}
	sipAgnt := new(SIPAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, sipAgnt); err != nil {
		return nil, err
	}
	return sipAgnt, nil
}

func (jsnCfg CgrJsonCfg) TemplateSJsonCfg() (FcTemplatesJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[TemplatesJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := make(FcTemplatesJsonCfg)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ConfigSJsonCfg() (*ConfigSCfgJson, error) {
	rawCfg, hasKey := jsnCfg[ConfigSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ConfigSCfgJson)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ApiBanCfgJson() (*APIBanJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[APIBanJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(APIBanJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) CoreSJSON() (*CoreSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[CoreSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CoreSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ActionSCfgJson() (*ActionSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ActionSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ActionSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) AccountSCfgJson() (*AccountSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[AccountSJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AccountSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) SetSection(_ *context.Context, section string, jsn interface{}) error {
	data, err := json.Marshal(jsn)
	if err != nil {
		return err
	}
	d := json.RawMessage(data)
	jsnCfg[section] = &d
	return nil
}

func (jsnCfg CgrJsonCfg) ConfigDBJsonCfg() (*ConfigDBJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ConfigDBJSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ConfigDBJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
