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
)

const (
	GENERAL_JSN        = "general"
	CACHE_JSN          = "caches"
	LISTEN_JSN         = "listen"
	HTTP_JSN           = "http"
	DATADB_JSN         = "data_db"
	STORDB_JSN         = "stor_db"
	FilterSjsn         = "filters"
	RALS_JSN           = "rals"
	SCHEDULER_JSN      = "schedulers"
	CDRS_JSN           = "cdrs"
	SessionSJson       = "sessions"
	FreeSWITCHAgentJSN = "freeswitch_agent"
	KamailioAgentJSN   = "kamailio_agent"
	AsteriskAgentJSN   = "asterisk_agent"
	DA_JSN             = "diameter_agent"
	RA_JSN             = "radius_agent"
	HttpAgentJson      = "http_agent"
	ATTRIBUTE_JSN      = "attributes"
	RESOURCES_JSON     = "resources"
	STATS_JSON         = "stats"
	THRESHOLDS_JSON    = "thresholds"
	RouteSJson         = "routes"
	LoaderJson         = "loaders"
	MAILER_JSN         = "mailer"
	SURETAX_JSON       = "suretax"
	DispatcherSJson    = "dispatchers"
	RegistrarCJson     = "registrarc"
	CgrLoaderCfgJson   = "loader"
	CgrMigratorCfgJson = "migrator"
	ChargerSCfgJson    = "chargers"
	TlsCfgJson         = "tls"
	AnalyzerCfgJson    = "analyzers"
	ApierS             = "apiers"
	DNSAgentJson       = "dns_agent"
	ERsJson            = "ers"
	EEsJson            = "ees"
	RateSJson          = "rates"
	ActionSJson        = "actions"
	RPCConnsJsonName   = "rpc_conns"
	SIPAgentJson       = "sip_agent"
	TemplatesJson      = "templates"
	ConfigSJson        = "configs"
	APIBanCfgJson      = "apiban"
	CoreSCfgJson       = "cores"
	AccountSCfgJson    = "accounts"
)

var (
	sortedCfgSections = []string{GENERAL_JSN, RPCConnsJsonName, DATADB_JSN, STORDB_JSN, LISTEN_JSN, TlsCfgJson, HTTP_JSN, SCHEDULER_JSN,
		CACHE_JSN, FilterSjsn, RALS_JSN, CDRS_JSN, ERsJson, SessionSJson, AsteriskAgentJSN, FreeSWITCHAgentJSN,
		KamailioAgentJSN, DA_JSN, RA_JSN, HttpAgentJson, DNSAgentJson, ATTRIBUTE_JSN, ChargerSCfgJson, RESOURCES_JSON, STATS_JSON,
		THRESHOLDS_JSON, RouteSJson, LoaderJson, MAILER_JSN, SURETAX_JSON, CgrLoaderCfgJson, CgrMigratorCfgJson, DispatcherSJson,
		AnalyzerCfgJson, ApierS, EEsJson, RateSJson, SIPAgentJson, RegistrarCJson, TemplatesJson, ConfigSJson, APIBanCfgJson, CoreSCfgJson,
		ActionSJson, AccountSCfgJson}
)

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromBytes(buf []byte) (cgrJsonCfg *CgrJsonCfg, err error) {
	cgrJsonCfg = new(CgrJsonCfg)
	err = NewRjReaderFromBytes(buf).Decode(cgrJsonCfg)
	return
}

// Main object holding the loaded config as section raw messages
type CgrJsonCfg map[string]*json.RawMessage

func (self CgrJsonCfg) GeneralJsonCfg() (*GeneralJsonCfg, error) {
	rawCfg, hasKey := self[GENERAL_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(GeneralJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RPCConnJsonCfg() (map[string]*RPCConnsJson, error) {
	rawCfg, hasKey := self[RPCConnsJsonName]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string]*RPCConnsJson)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CacheJsonCfg() (*CacheJsonCfg, error) {
	rawCfg, hasKey := self[CACHE_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CacheJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ListenJsonCfg() (*ListenJsonCfg, error) {
	rawCfg, hasKey := self[LISTEN_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ListenJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) HttpJsonCfg() (*HTTPJsonCfg, error) {
	rawCfg, hasKey := self[HTTP_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(HTTPJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) DbJsonCfg(section string) (*DbJsonCfg, error) {
	rawCfg, hasKey := self[section]
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
	rawCfg, hasKey := jsnCfg[FilterSjsn]
	if !hasKey {
		return nil, nil
	}
	cfg := new(FilterSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RalsJsonCfg() (*RalsJsonCfg, error) {
	rawCfg, hasKey := self[RALS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RalsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SchedulerJsonCfg() (*SchedulerJsonCfg, error) {
	rawCfg, hasKey := self[SCHEDULER_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SchedulerJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CdrsJsonCfg() (*CdrsJsonCfg, error) {
	rawCfg, hasKey := self[CDRS_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CdrsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ERsJsonCfg() (erSCfg *ERsJsonCfg, err error) {
	rawCfg, hasKey := self[ERsJson]
	if !hasKey {
		return
	}
	erSCfg = new(ERsJsonCfg)
	err = json.Unmarshal(*rawCfg, &erSCfg)
	return
}

func (self CgrJsonCfg) EEsJsonCfg() (erSCfg *EEsJsonCfg, err error) {
	rawCfg, hasKey := self[EEsJson]
	if !hasKey {
		return
	}
	erSCfg = new(EEsJsonCfg)
	err = json.Unmarshal(*rawCfg, &erSCfg)
	return
}

func (self CgrJsonCfg) SessionSJsonCfg() (*SessionSJsonCfg, error) {
	rawCfg, hasKey := self[SessionSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SessionSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) FreeswitchAgentJsonCfg() (*FreeswitchAgentJsonCfg, error) {
	rawCfg, hasKey := self[FreeSWITCHAgentJSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(FreeswitchAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) KamAgentJsonCfg() (*KamAgentJsonCfg, error) {
	rawCfg, hasKey := self[KamailioAgentJSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(KamAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) AsteriskAgentJsonCfg() (*AsteriskAgentJsonCfg, error) {
	rawCfg, hasKey := self[AsteriskAgentJSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AsteriskAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) DiameterAgentJsonCfg() (*DiameterAgentJsonCfg, error) {
	rawCfg, hasKey := self[DA_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DiameterAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RadiusAgentJsonCfg() (*RadiusAgentJsonCfg, error) {
	rawCfg, hasKey := self[RA_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RadiusAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) HttpAgentJsonCfg() (*[]*HttpAgentJsonCfg, error) {
	rawCfg, hasKey := self[HttpAgentJson]
	if !hasKey {
		return nil, nil
	}
	httpAgnt := make([]*HttpAgentJsonCfg, 0)
	if err := json.Unmarshal(*rawCfg, &httpAgnt); err != nil {
		return nil, err
	}
	return &httpAgnt, nil
}

func (self CgrJsonCfg) DNSAgentJsonCfg() (da *DNSAgentJsonCfg, err error) {
	rawCfg, hasKey := self[DNSAgentJson]
	if !hasKey {
		return
	}
	da = new(DNSAgentJsonCfg)
	err = json.Unmarshal(*rawCfg, da)
	return
}

func (cgrJsn CgrJsonCfg) AttributeServJsonCfg() (*AttributeSJsonCfg, error) {
	rawCfg, hasKey := cgrJsn[ATTRIBUTE_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AttributeSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cgrJsn CgrJsonCfg) ChargerServJsonCfg() (*ChargerSJsonCfg, error) {
	rawCfg, hasKey := cgrJsn[ChargerSCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ChargerSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ResourceSJsonCfg() (*ResourceSJsonCfg, error) {
	rawCfg, hasKey := self[RESOURCES_JSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ResourceSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) StatSJsonCfg() (*StatServJsonCfg, error) {
	rawCfg, hasKey := self[STATS_JSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(StatServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ThresholdSJsonCfg() (*ThresholdSJsonCfg, error) {
	rawCfg, hasKey := self[THRESHOLDS_JSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ThresholdSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RouteSJsonCfg() (*RouteSJsonCfg, error) {
	rawCfg, hasKey := self[RouteSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RouteSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) LoaderJsonCfg() ([]*LoaderJsonCfg, error) {
	rawCfg, hasKey := self[LoaderJson]
	if !hasKey {
		return nil, nil
	}
	cfg := make([]*LoaderJsonCfg, 0)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) MailerJsonCfg() (*MailerJsonCfg, error) {
	rawCfg, hasKey := self[MAILER_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(MailerJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SureTaxJsonCfg() (*SureTaxJsonCfg, error) {
	rawCfg, hasKey := self[SURETAX_JSON]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SureTaxJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) DispatcherSJsonCfg() (*DispatcherSJsonCfg, error) {
	rawCfg, hasKey := self[DispatcherSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(DispatcherSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RegistrarCJsonCfgs() (*RegistrarCJsonCfgs, error) {
	rawCfg, hasKey := self[RegistrarCJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RegistrarCJsonCfgs)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) LoaderCfgJson() (*LoaderCfgJson, error) {
	rawCfg, hasKey := self[CgrLoaderCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(LoaderCfgJson)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) MigratorCfgJson() (*MigratorCfgJson, error) {
	rawCfg, hasKey := self[CgrMigratorCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(MigratorCfgJson)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) TlsCfgJson() (*TlsJsonCfg, error) {
	rawCfg, hasKey := self[TlsCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(TlsJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) AnalyzerCfgJson() (*AnalyzerSJsonCfg, error) {
	rawCfg, hasKey := self[AnalyzerCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AnalyzerSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ApierCfgJson() (*ApierJsonCfg, error) {
	rawCfg, hasKey := self[ApierS]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ApierJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) RateCfgJson() (*RateSJsonCfg, error) {
	rawCfg, hasKey := self[RateSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(RateSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) SIPAgentJsonCfg() (*SIPAgentJsonCfg, error) {
	rawCfg, hasKey := self[SIPAgentJson]
	if !hasKey {
		return nil, nil
	}
	sipAgnt := new(SIPAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, sipAgnt); err != nil {
		return nil, err
	}
	return sipAgnt, nil
}

func (self CgrJsonCfg) TemplateSJsonCfg() (map[string][]*FcTemplateJsonCfg, error) {
	rawCfg, hasKey := self[TemplatesJson]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string][]*FcTemplateJsonCfg)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ConfigSJsonCfg() (*ConfigSCfgJson, error) {
	rawCfg, hasKey := self[ConfigSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ConfigSCfgJson)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ApiBanCfgJson() (*APIBanJsonCfg, error) {
	rawCfg, hasKey := self[APIBanCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(APIBanJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CoreSCfgJson() (*CoreSJsonCfg, error) {
	rawCfg, hasKey := self[CoreSCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(CoreSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) ActionSCfgJson() (*ActionSJsonCfg, error) {
	rawCfg, hasKey := self[ActionSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ActionSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) AccountSCfgJson() (*AccountSJsonCfg, error) {
	rawCfg, hasKey := self[AccountSCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AccountSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
