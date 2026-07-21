// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"encoding/json"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	GeneralJSON         = "general"
	LoggerJSON          = "logger"
	CacheJSON           = "caches"
	ListenJSON          = "listen"
	HTTPJSON            = "http"
	DBJSON              = "db"
	StorDBJSON          = "storDB"
	FilterSJSON         = "filters"
	CDRsJSON            = "cdrs"
	SessionSJSON        = "sessions"
	FreeSWITCHAgentJSON = "freeswitchAgent"
	KamailioAgentJSON   = "kamailioAgent"
	AsteriskAgentJSON   = "asteriskAgent"
	DiameterAgentJSON   = "diameterAgent"
	RadiusAgentJSON     = "radiusAgent"
	HTTPAgentJSON       = "httpAgent"
	PrometheusAgentJSON = "prometheusAgent"
	AttributeSJSON      = "attributes"
	ResourceSJSON       = "resources"
	IPsJSON             = "ips"
	JanusAgentJSON      = "janusAgent"
	StatSJSON           = "stats"
	ThresholdSJSON      = "thresholds"
	TrendSJSON          = "trends"
	RankingSJSON        = "rankings"
	TPeSJSON            = "tpes"
	EFsJSON             = "efs"
	RouteSJSON          = "routes"
	LoaderSJSON         = "loaders"
	SureTaxJSON         = "suretax"
	RegistrarCJSON      = "registrarc"
	LoaderJSON          = "loader"
	MigratorJSON        = "migrator"
	ChargerSJSON        = "chargers"
	TlsJSON             = "tls"
	AnalyzerSJSON       = "analyzers"
	AdminSJSON          = "admins"
	DNSAgentJSON        = "dnsAgent"
	ERsJSON             = "ers"
	EEsJSON             = "ees"
	RateSJSON           = "rates"
	ActionSJSON         = "actions"
	RPCConnsJSON        = "rpcConns"
	SIPAgentJSON        = "sipAgent"
	TemplatesJSON       = "templates"
	ConfigSJSON         = "configs"
	APIBanJSON          = "apiban"
	SentryPeerJSON      = "sentrypeer"
	CoreSJSON           = "cores"
	AccountSJSON        = "accounts"
	ConfigDBJSON        = "configDB"
)

var (
	SectionToService = map[string]string{
		GeneralJSON:         utils.GlobalVarS,
		AttributeSJSON:      utils.AttributeS,
		ChargerSJSON:        utils.ChargerS,
		ThresholdSJSON:      utils.ThresholdS,
		TrendSJSON:          utils.TrendS,
		RankingSJSON:        utils.RankingS,
		StatSJSON:           utils.StatS,
		ResourceSJSON:       utils.ResourceS,
		IPsJSON:             utils.IPs,
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
		PrometheusAgentJSON: utils.PrometheusAgent,
		LoaderSJSON:         utils.LoaderS,
		AnalyzerSJSON:       utils.AnalyzerS,
		DBJSON:              utils.DB,
		EEsJSON:             utils.EEs,
		EFsJSON:             utils.EFs,
		RateSJSON:           utils.RateS,
		SIPAgentJSON:        utils.SIPAgent,
		RegistrarCJSON:      utils.RegistrarC,
		HTTPJSON:            utils.GlobalVarS,
		AccountSJSON:        utils.AccountS,
		ActionSJSON:         utils.ActionS,
		CoreSJSON:           utils.CoreS,
		TPeSJSON:            utils.TPeS,
		RPCConnsJSON:        utils.ConnManager,
	}
)

type ConfigDB interface {
	GetSection(ctx *context.Context, section string, val any) error // in this case value must be a not nil pointer
	SetSection(ctx *context.Context, section string, val any) error
	DumpConfigDB() error
	RewriteConfigDB() error
	BackupConfigDB(string, bool) error
}

// Loads the json config out of io.Reader, eg other sources than file, maybe over http
func NewCgrJsonCfgFromBytes(buf []byte) (cgrJsonCfg *CgrJsonCfg, err error) {
	cgrJsonCfg = new(CgrJsonCfg)
	err = NewRjReaderFromBytes(buf).Decode(cgrJsonCfg)
	return
}

// Main object holding the loaded config as section raw messages
type CgrJsonCfg map[string]json.RawMessage

func (jsnCfg CgrJsonCfg) GetSection(ctx *context.Context, section string, val any) (err error) {
	if rawCfg, hasKey := jsnCfg[section]; hasKey {
		err = json.Unmarshal(rawCfg, val)
	}
	return
}

func (jsnCfg CgrJsonCfg) SetSection(_ *context.Context, section string, jsn any) (_ error) {
	data, err := json.Marshal(jsn)
	if err != nil {
		return err
	}
	jsnCfg[section] = json.RawMessage(data)
	return
}

// Only intended for InternalDB
func (jsnCfg CgrJsonCfg) BackupConfigDB(string, bool) error {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (jsnCfg CgrJsonCfg) DumpConfigDB() error {
	return utils.ErrNotImplemented
}

// Only intended for InternalDB
func (jsnCfg CgrJsonCfg) RewriteConfigDB() error {
	return utils.ErrNotImplemented
}

type Section interface {
	SName() string
	Load(*context.Context, ConfigDB, *CGRConfig) error
	AsMapInterface() any
	CloneSection() Section
	// UpdateDB(*context.Context) // not know
}

func newSections(cfg *CGRConfig) Sections {
	return Sections{
		cfg.generalCfg,
		cfg.loggerCfg,
		cfg.efsCfg,
		cfg.rpcConns,
		cfg.dbCfg,
		cfg.listenCfg,
		cfg.tlsCfg,
		cfg.httpCfg,
		cfg.cacheCfg,
		cfg.filterSCfg,
		cfg.templates,
		cfg.attributeSCfg,
		cfg.chargerSCfg,
		cfg.resourceSCfg,
		cfg.ipsCfg,
		cfg.statsCfg,
		cfg.thresholdSCfg,
		cfg.rankingSCfg,
		cfg.trendSCfg,
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
		cfg.prometheusAgentCfg,
		cfg.radiusAgentCfg,
		cfg.janusAgentCfg,
		&cfg.httpAgentCfg,
		cfg.dnsAgentCfg,
		cfg.sipAgentCfg,
		cfg.migratorCgrCfg,
		cfg.registrarCCfg,
		cfg.analyzerSCfg,
		cfg.admS,
		cfg.coreSCfg,
		cfg.configSCfg,
		cfg.apiBanCfg,
		cfg.sentryPeerCfg,
		cfg.configDBCfg,
		cfg.sureTaxCfg,
		cfg.tpeSCfg,
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
func (r Sections) AsMapInterface() (m map[string]any) {
	m = make(map[string]any)
	for _, sec := range r {
		m[sec.SName()] = sec.AsMapInterface()
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
