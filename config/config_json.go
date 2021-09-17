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
	SectionToService = map[string]string{
		AttributeSJSON:      utils.AttributeS,
		ChargerSJSON:        utils.ChargerS,
		ThresholdSJSON:      utils.ThresholdS,
		StatSJSON:           utils.StatS,
		ResourceSJSON:       utils.ResourceS,
		RouteSJSON:          utils.RouteS,
		AdminSJSON:          utils.AdminS,
		CDRsJSON:            utils.CDRServer,
		SessionSJSON:        utils.SessionS,
		ERsJSON:             utils.ERs,
		DNSAgentJSON:        utils.DNSAgent,
		FreeSWITCHAgentJSON: utils.FreeSWITCHAgent,
		KamailioAgentJSON:   utils.KamailioAgent,
		AsteriskAgentJSON:   utils.AsteriskAgent,
		RadiusAgentJSON:     utils.RadiusAgent,
		DiameterAgentJSON:   utils.DiameterAgent,
		HTTPAgentJSON:       utils.HTTPAgent,
		LoaderSJSON:         utils.LoaderS,
		AnalyzerSJSON:       utils.AnalyzerS,
		DispatcherSJSON:     utils.DispatcherS,
		DataDBJSON:          utils.DataDB,
		StorDBJSON:          utils.StorDB,
		EEsJSON:             utils.EEs,
		RateSJSON:           utils.RateS,
		SIPAgentJSON:        utils.SIPAgent,
		RegistrarCJSON:      utils.RegistrarC,
		HTTPJSON:            utils.GlobalVarS,
		AccountSJSON:        utils.AccountS,
		ActionSJSON:         utils.ActionS,
		CoreSJSON:           utils.CoreS,
		RPCConnsJSON:        RPCConnsJSON,
	}
)

type ConfigDB interface {
	GetSection(ctx *context.Context, section string, val interface{}) error // in this case value must be a not nil pointer
	SetSection(ctx *context.Context, section string, val interface{}) error
}

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromBytes(buf []byte) (cgrJsonCfg *CgrJsonCfg, err error) {
	cgrJsonCfg = new(CgrJsonCfg)
	err = NewRjReaderFromBytes(buf).Decode(cgrJsonCfg)
	return
}

// Main object holding the loaded config as section raw messages
type CgrJsonCfg map[string]json.RawMessage

func (jsnCfg CgrJsonCfg) GetSection(ctx *context.Context, section string, val interface{}) (err error) {
	if rawCfg, hasKey := jsnCfg[section]; hasKey {
		err = json.Unmarshal(rawCfg, val)
	}
	return
}

func (jsnCfg CgrJsonCfg) SetSection(_ *context.Context, section string, jsn interface{}) (_ error) {
	data, err := json.Marshal(jsn)
	if err != nil {
		return err
	}
	jsnCfg[section] = json.RawMessage(data)
	return
}

type Section interface {
	SName() string
	Load(*context.Context, ConfigDB, *CGRConfig) error
	AsMapInterface(string) interface{}
	CloneSection() Section
	// UpdateDB(*context.Context) // not know
}

func newSections(cfg *CGRConfig) Sections {
	return Sections{
		cfg.generalCfg,
		cfg.rpcConns,
		cfg.dataDbCfg,
		cfg.storDbCfg,
		cfg.listenCfg,
		cfg.tlsCfg,
		cfg.httpCfg,
		cfg.cacheCfg,
		cfg.filterSCfg,
		cfg.templates,
		cfg.attributeSCfg,
		cfg.chargerSCfg,
		cfg.resourceSCfg,
		cfg.statsCfg,
		cfg.thresholdSCfg,
		cfg.routeSCfg,
		cfg.rateSCfg,
		cfg.accountSCfg,
		cfg.actionSCfg,
		cfg.sessionSCfg,
		cfg.cdrsCfg,
		&cfg.loaderCfg,
		cfg.loaderCgrCfg,
		cfg.ersCfg,
		cfg.eesCfg,
		cfg.asteriskAgentCfg,
		cfg.fsAgentCfg,
		cfg.kamAgentCfg,
		cfg.diameterAgentCfg,
		cfg.radiusAgentCfg,
		&cfg.httpAgentCfg,
		cfg.dnsAgentCfg,
		cfg.sipAgentCfg,
		cfg.migratorCgrCfg,
		cfg.dispatcherSCfg,
		cfg.registrarCCfg,
		cfg.analyzerSCfg,
		cfg.admS,
		cfg.coreSCfg,
		cfg.configSCfg,
		cfg.apiBanCfg,
		cfg.configDBCfg,
		cfg.sureTaxCfg,
	}
}

type Sections []Section

func (r Sections) Get(name string) (sec Section, has bool) {
	for _, sec = range r {
		if has = sec.SName() == name; has {
			return
		}
	}
	return
}

func (r Sections) Load(ctx *context.Context, db ConfigDB, cfg *CGRConfig) (err error) {
	for _, f := range r {
		if err = f.Load(ctx, db, cfg); err != nil {
			return
		}
	}
	return
}

func (r Sections) LoadWithout(ctx *context.Context, db ConfigDB, cfg *CGRConfig, sections ...string) (err error) {
	eSec := utils.NewStringSet(sections)
	for _, sec := range r {
		if !eSec.Has(sec.SName()) {
			if err = sec.Load(ctx, db, cfg); err != nil {
				return
			}
		}
	}
	return
}
func (r Sections) AsMapInterface(sep string) (m map[string]interface{}) {
	m = make(map[string]interface{})
	for _, sec := range r {
		m[sec.SName()] = sec.AsMapInterface(sep)
	}
	return
}

func (r Sections) Clone() (c Sections) {
	c = make(Sections, len(r))
	for s, f := range r {
		c[s] = f.CloneSection()
	}
	return
}
