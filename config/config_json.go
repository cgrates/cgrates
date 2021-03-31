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

	"github.com/cgrates/cgrates/utils"
)

const (
	GENERAL_JSN        = "general"
	CACHE_JSN          = "caches"
	LISTEN_JSN         = "listen"
	HTTP_JSN           = "http"
	DATADB_JSN         = "data_db"
	STORDB_JSN         = "stor_db"
	FilterSjsn         = "filters"
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
	sortedCfgSections = []string{GENERAL_JSN, RPCConnsJsonName, DATADB_JSN, STORDB_JSN, LISTEN_JSN, TlsCfgJson, HTTP_JSN,
		CACHE_JSN, FilterSjsn, CDRS_JSN, ERsJson, SessionSJson, AsteriskAgentJSN, FreeSWITCHAgentJSN,
		KamailioAgentJSN, DA_JSN, RA_JSN, HttpAgentJson, DNSAgentJson, ATTRIBUTE_JSN, ChargerSCfgJson, RESOURCES_JSON, STATS_JSON,
		THRESHOLDS_JSON, RouteSJson, LoaderJson, MAILER_JSN, SURETAX_JSON, CgrLoaderCfgJson, CgrMigratorCfgJson, DispatcherSJson,
		AnalyzerCfgJson, ApierS, EEsJson, RateSJson, SIPAgentJson, RegistrarCJson, TemplatesJson, ConfigSJson, APIBanCfgJson, CoreSCfgJson,
		ActionSJson, AccountSCfgJson}
	sortedSectionsSet = utils.NewStringSet(sortedCfgSections)
)

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromBytes(buf []byte) (cgrJsonCfg *CgrJsonCfg, err error) {
	cgrJsonCfg = new(CgrJsonCfg)
	err = NewRjReaderFromBytes(buf).Decode(cgrJsonCfg)
	return
}

// Main object holding the loaded config as section raw messages
type CgrJsonCfg map[string]*json.RawMessage

func (jsnCfg CgrJsonCfg) GeneralJsonCfg() (*GeneralJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[GENERAL_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(GeneralJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RPCConnJsonCfg() (map[string]*RPCConnsJson, error) {
	rawCfg, hasKey := jsnCfg[RPCConnsJsonName]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string]*RPCConnsJson)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) CacheJsonCfg() (*CacheJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[CACHE_JSN]
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
	rawCfg, hasKey := jsnCfg[LISTEN_JSN]
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
	rawCfg, hasKey := jsnCfg[HTTP_JSN]
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

func (jsnCfg CgrJsonCfg) CdrsJsonCfg() (*CdrsJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[CDRS_JSN]
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
	rawCfg, hasKey := jsnCfg[ERsJson]
	if !hasKey {
		return
	}
	erSCfg = new(ERsJsonCfg)
	err = json.Unmarshal(*rawCfg, &erSCfg)
	return
}

func (jsnCfg CgrJsonCfg) EEsJsonCfg() (erSCfg *EEsJsonCfg, err error) {
	rawCfg, hasKey := jsnCfg[EEsJson]
	if !hasKey {
		return
	}
	erSCfg = new(EEsJsonCfg)
	err = json.Unmarshal(*rawCfg, &erSCfg)
	return
}

func (jsnCfg CgrJsonCfg) SessionSJsonCfg() (*SessionSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[SessionSJson]
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
	rawCfg, hasKey := jsnCfg[FreeSWITCHAgentJSN]
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
	rawCfg, hasKey := jsnCfg[KamailioAgentJSN]
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
	rawCfg, hasKey := jsnCfg[AsteriskAgentJSN]
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
	rawCfg, hasKey := jsnCfg[DA_JSN]
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
	rawCfg, hasKey := jsnCfg[RA_JSN]
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
	rawCfg, hasKey := jsnCfg[HttpAgentJson]
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
	rawCfg, hasKey := jsnCfg[DNSAgentJson]
	if !hasKey {
		return
	}
	da = new(DNSAgentJsonCfg)
	err = json.Unmarshal(*rawCfg, da)
	return
}

func (jsnCfg CgrJsonCfg) AttributeServJsonCfg() (*AttributeSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ATTRIBUTE_JSN]
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
	rawCfg, hasKey := jsnCfg[ChargerSCfgJson]
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
	rawCfg, hasKey := jsnCfg[RESOURCES_JSON]
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
	rawCfg, hasKey := jsnCfg[STATS_JSON]
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
	rawCfg, hasKey := jsnCfg[THRESHOLDS_JSON]
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
	rawCfg, hasKey := jsnCfg[RouteSJson]
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
	rawCfg, hasKey := jsnCfg[LoaderJson]
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
	rawCfg, hasKey := jsnCfg[MAILER_JSN]
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
	rawCfg, hasKey := jsnCfg[SURETAX_JSON]
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
	rawCfg, hasKey := jsnCfg[DispatcherSJson]
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
	rawCfg, hasKey := jsnCfg[RegistrarCJson]
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
	rawCfg, hasKey := jsnCfg[CgrLoaderCfgJson]
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
	rawCfg, hasKey := jsnCfg[CgrMigratorCfgJson]
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
	rawCfg, hasKey := jsnCfg[TlsCfgJson]
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
	rawCfg, hasKey := jsnCfg[AnalyzerCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AnalyzerSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ApierCfgJson() (*ApierJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[ApierS]
	if !hasKey {
		return nil, nil
	}
	cfg := new(ApierJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) RateCfgJson() (*RateSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[RateSJson]
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
	rawCfg, hasKey := jsnCfg[SIPAgentJson]
	if !hasKey {
		return nil, nil
	}
	sipAgnt := new(SIPAgentJsonCfg)
	if err := json.Unmarshal(*rawCfg, sipAgnt); err != nil {
		return nil, err
	}
	return sipAgnt, nil
}

func (jsnCfg CgrJsonCfg) TemplateSJsonCfg() (map[string][]*FcTemplateJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[TemplatesJson]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string][]*FcTemplateJsonCfg)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) ConfigSJsonCfg() (*ConfigSCfgJson, error) {
	rawCfg, hasKey := jsnCfg[ConfigSJson]
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
	rawCfg, hasKey := jsnCfg[APIBanCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(APIBanJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (jsnCfg CgrJsonCfg) CoreSCfgJson() (*CoreSJsonCfg, error) {
	rawCfg, hasKey := jsnCfg[CoreSCfgJson]
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
	rawCfg, hasKey := jsnCfg[ActionSJson]
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
	rawCfg, hasKey := jsnCfg[AccountSCfgJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AccountSJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
