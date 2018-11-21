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
	"io"
	"os"
)

const (
	GENERAL_JSN        = "general"
	CACHE_JSN          = "cache"
	LISTEN_JSN         = "listen"
	HTTP_JSN           = "http"
	DATADB_JSN         = "data_db"
	STORDB_JSN         = "stor_db"
	FilterSjsn         = "filters"
	RALS_JSN           = "rals"
	SCHEDULER_JSN      = "scheduler"
	CDRS_JSN           = "cdrs"
	MEDIATOR_JSN       = "mediator"
	CDRE_JSN           = "cdre"
	CDRC_JSN           = "cdrc"
	SessionSJson       = "sessions"
	FreeSWITCHAgentJSN = "freeswitch_agent"
	KamailioAgentJSN   = "kamailio_agent"
	AsteriskAgentJSN   = "asterisk_agent"
	SM_JSN             = "session_manager"
	FS_JSN             = "freeswitch"
	OSIPS_JSN          = "opensips"
	DA_JSN             = "diameter_agent"
	RA_JSN             = "radius_agent"
	HttpAgentJson      = "http_agent"
	HISTSERV_JSN       = "historys"
	PUBSUBSERV_JSN     = "pubsubs"
	ALIASESSERV_JSN    = "aliases"
	USERSERV_JSN       = "users"
	ATTRIBUTE_JSN      = "attributes"
	RESOURCES_JSON     = "resources"
	STATS_JSON         = "stats"
	THRESHOLDS_JSON    = "thresholds"
	SupplierSJson      = "suppliers"
	FILTERS_JSON       = "filters"
	LoaderJson         = "loaders"
	MAILER_JSN         = "mailer"
	SURETAX_JSON       = "suretax"
	DispatcherSJson    = "dispatcher"
	CgrLoaderCfgJson   = "loader"
	CgrMigratorCfgJson = "migrator"
	ChargerSCfgJson    = "chargers"
	TlsCfgJson         = "tls"
	AnalyzerCfgJson    = "analyzers"
)

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromReader(r io.Reader) (*CgrJsonCfg, error) {
	var cgrJsonCfg CgrJsonCfg
	jr := NewRawJSONReader(r)
	if err := json.NewDecoder(jr).Decode(&cgrJsonCfg); err != nil {
		return nil, err
	}
	return &cgrJsonCfg, nil
}

// Loads the config out of file
func NewCgrJsonCfgFromFile(fpath string) (*CgrJsonCfg, error) {
	cfgFile, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()
	return NewCgrJsonCfgFromReader(cfgFile)
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

func (self CgrJsonCfg) CdreJsonCfgs() (map[string]*CdreJsonCfg, error) {
	rawCfg, hasKey := self[CDRE_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := make(map[string]*CdreJsonCfg)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) CdrcJsonCfg() ([]*CdrcJsonCfg, error) {
	rawCfg, hasKey := self[CDRC_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := make([]*CdrcJsonCfg, 0)
	if err := json.Unmarshal(*rawCfg, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
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

func (self CgrJsonCfg) PubSubServJsonCfg() (*PubSubServJsonCfg, error) {
	rawCfg, hasKey := self[PUBSUBSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(PubSubServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) AliasesServJsonCfg() (*AliasesServJsonCfg, error) {
	rawCfg, hasKey := self[ALIASESSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(AliasesServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (self CgrJsonCfg) UserServJsonCfg() (*UserServJsonCfg, error) {
	rawCfg, hasKey := self[USERSERV_JSN]
	if !hasKey {
		return nil, nil
	}
	cfg := new(UserServJsonCfg)
	if err := json.Unmarshal(*rawCfg, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
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

func (self CgrJsonCfg) SupplierSJsonCfg() (*SupplierSJsonCfg, error) {
	rawCfg, hasKey := self[SupplierSJson]
	if !hasKey {
		return nil, nil
	}
	cfg := new(SupplierSJsonCfg)
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
