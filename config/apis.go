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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// getSectionAsMap returns a section from config as a map[string]interface
func (cfg *CGRConfig) getSectionAsMap(section string) (mp interface{}, err error) {
	switch section {
	case GeneralJSON:
		mp = cfg.GeneralCfg().AsMapInterface()
	case DataDBJSON:
		mp = cfg.DataDbCfg().AsMapInterface()
	case StorDBJSON:
		mp = cfg.StorDbCfg().AsMapInterface()
	case TlsJSON:
		mp = cfg.TLSCfg().AsMapInterface()
	case CacheJSON:
		mp = cfg.CacheCfg().AsMapInterface()
	case ListenJSON:
		mp = cfg.ListenCfg().AsMapInterface()
	case HTTPJSON:
		mp = cfg.HTTPCfg().AsMapInterface()
	case FilterSJSON:
		mp = cfg.FilterSCfg().AsMapInterface()
	case CDRsJSON:
		mp = cfg.CdrsCfg().AsMapInterface()
	case SessionSJSON:
		mp = cfg.SessionSCfg().AsMapInterface()
	case FreeSWITCHAgentJSON:
		mp = cfg.FsAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case KamailioAgentJSON:
		mp = cfg.KamAgentCfg().AsMapInterface()
	case AsteriskAgentJSON:
		mp = cfg.AsteriskAgentCfg().AsMapInterface()
	case DiameterAgentJSON:
		mp = cfg.DiameterAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case RadiusAgentJSON:
		mp = cfg.RadiusAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case DNSAgentJSON:
		mp = cfg.DNSAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case AttributeSJSON:
		mp = cfg.AttributeSCfg().AsMapInterface()
	case ChargerSJSON:
		mp = cfg.ChargerSCfg().AsMapInterface()
	case ResourceSJSON:
		mp = cfg.ResourceSCfg().AsMapInterface()
	case StatSJSON:
		mp = cfg.StatSCfg().AsMapInterface()
	case ThresholdSJSON:
		mp = cfg.ThresholdSCfg().AsMapInterface()
	case RouteSJSON:
		mp = cfg.RouteSCfg().AsMapInterface()
	case SureTaxJSON:
		mp = cfg.SureTaxCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case DispatcherSJSON:
		mp = cfg.DispatcherSCfg().AsMapInterface()
	case RegistrarCJSON:
		mp = cfg.RegistrarCCfg().AsMapInterface()
	case LoaderSJSON:
		mp = cfg.LoaderCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case LoaderJSON:
		mp = cfg.LoaderCgrCfg().AsMapInterface()
	case MigratorJSON:
		mp = cfg.MigratorCgrCfg().AsMapInterface()
	case AdminSJSON:
		mp = cfg.AdminSCfg().AsMapInterface()
	case EEsJSON:
		mp = cfg.EEsCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ERsJSON:
		mp = cfg.ERsCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case RPCConnsJSON:
		mp = cfg.RPCConns().AsMapInterface()
	case SIPAgentJSON:
		mp = cfg.SIPAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case TemplatesJSON:
		mp = cfg.TemplatesCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case ConfigSJSON:
		mp = cfg.ConfigSCfg().AsMapInterface()
	case APIBanJSON:
		mp = cfg.APIBanCfg().AsMapInterface()
	case HTTPAgentJSON:
		mp = cfg.HTTPAgentCfg().AsMapInterface(cfg.GeneralCfg().RSRSep)
	case AnalyzerSJSON:
		mp = cfg.AnalyzerSCfg().AsMapInterface()
	case RateSJSON:
		mp = cfg.RateSCfg().AsMapInterface()
	case CoreSJSON:
		mp = cfg.CoreSCfg().AsMapInterface()
	case ActionSJSON:
		mp = cfg.ActionSCfg().AsMapInterface()
	case AccountSJSON:
		mp = cfg.AccountSCfg().AsMapInterface()
	case ConfigDBJSON:
		mp = cfg.ConfigDBCfg().AsMapInterface()
	default:
		err = errors.New("Invalid section ")
	}
	return
}

// ReloadArgs the API params for V1ReloadConfig
type ReloadArgs struct {
	APIOpts map[string]interface{}
	Tenant  string
	Section string
	DryRun  bool
}

// V1ReloadConfig reloads the configuration
func (cfg *CGRConfig) V1ReloadConfig(ctx *context.Context, args *ReloadArgs, reply *string) (err error) {
	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}
	cfgV.reloadDPCache(args.Section)
	if err = cfgV.loadCfgWithLocks(cfg.ConfigPath, args.Section); err != nil {
		return
	}
	if cfg.db != nil {
		sections := []string{args.Section}
		allSections := args.Section == utils.MetaEmpty ||
			args.Section == utils.MetaAll
		if allSections {
			sections = sortedCfgSections
		}
		if err = cfgV.loadCfgFromDB(cfg.db, sections, allSections); err != nil {
			return
		}
	}

	//  lock all sections
	cfgV.rLockSections()

	err = cfgV.checkConfigSanity()

	cfgV.rUnlockSections() // unlock before checking the error

	if err != nil {
		return
	}
	if !args.DryRun {
		if args.Section == utils.EmptyString || args.Section == utils.MetaAll {
			cfgV.reloadSections(sortedCfgSections...)
		} else {
			cfgV.reloadSections(args.Section)
		}
	}
	*reply = utils.OK
	return
}

// SectionWithAPIOpts the API params for GetConfig
type SectionWithAPIOpts struct {
	APIOpts  map[string]interface{}
	Tenant   string
	Sections []string
}

// V1GetConfig will retrieve from CGRConfig a section
func (cfg *CGRConfig) V1GetConfig(ctx *context.Context, args *SectionWithAPIOpts, reply *map[string]interface{}) (err error) {
	if len(args.Sections) == 0 ||
		args.Sections[0] == utils.MetaAll {
		args.Sections = sortedCfgSections
	}
	mp := make(map[string]interface{})
	sections := utils.StringSet{}
	cfg.cacheDPMux.RLock()
	for _, section := range args.Sections {
		if val, has := cfg.cacheDP[section]; has && val != nil {
			mp[section] = val
		} else {
			sections.Add(section)
		}
	}
	cfg.cacheDPMux.RUnlock()
	if sections.Size() == 0 { // all sections were cached
		*reply = mp
		return
	}
	if sections.Size() == len(sortedCfgSections) {
		mp = cfg.AsMapInterface(cfg.GeneralCfg().RSRSep)
	} else {
		for section := range sections {
			var val interface{}
			if val, err = cfg.getSectionAsMap(section); err != nil {
				return
			}
			mp[section] = val
			cfg.cacheDPMux.Lock()
			cfg.cacheDP[section] = val
			cfg.cacheDPMux.Unlock()
		}
	}
	*reply = mp
	return
}

// SetConfigArgs the API params for V1SetConfig
type SetConfigArgs struct {
	APIOpts map[string]interface{}
	Tenant  string
	Config  map[string]interface{}
	DryRun  bool
}

// V1SetConfig reloads the sections of config
func (cfg *CGRConfig) V1SetConfig(ctx *context.Context, args *SetConfigArgs, reply *string) (err error) {
	if len(args.Config) == 0 {
		*reply = utils.OK
		return
	}
	sections := make([]string, 0, len(args.Config))
	for section := range args.Config {
		if !sortedSectionsSet.Has(section) {
			return fmt.Errorf("Invalid section <%s> ", section)
		}
		sections = append(sections, section)
	}
	var b []byte
	if b, err = json.Marshal(args.Config); err != nil {
		return
	}

	var oldCfg *CGRConfig
	updateDB := cfg.db != nil
	if !args.DryRun && updateDB { // need to update the DB but only parts
		oldCfg = cfg.Clone()
	}

	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}

	cfgV.reloadDPCache(sections...)
	if err = cfgV.loadCfgFromJSONWithLocks(bytes.NewBuffer(b), sections); err != nil {
		return
	}

	//  lock all sections
	cfgV.rLockSections()

	err = cfgV.checkConfigSanity()

	cfgV.rUnlockSections() // unlock before checking the error
	if err != nil {
		return
	}
	if !args.DryRun {
		cfgV.reloadSections(sections...)
		if updateDB { // need to update the DB but only parts
			if err = storeDiffSections(ctx, sections, cfgV.db, oldCfg, cfgV); err != nil {
				return
			}
		}
	}
	*reply = utils.OK
	return
}

//V1GetConfigAsJSON will retrieve from CGRConfig a section as a string
func (cfg *CGRConfig) V1GetConfigAsJSON(ctx *context.Context, args *SectionWithAPIOpts, reply *string) (err error) {
	var mp map[string]interface{}
	if err = cfg.V1GetConfig(ctx, args, &mp); err != nil {
		return
	}
	*reply = utils.ToJSON(mp)
	return
}

// SetConfigFromJSONArgs the API params for V1SetConfigFromJSON
type SetConfigFromJSONArgs struct {
	APIOpts map[string]interface{}
	Tenant  string
	Config  string
	DryRun  bool
}

// V1SetConfigFromJSON reloads the sections of config
func (cfg *CGRConfig) V1SetConfigFromJSON(ctx *context.Context, args *SetConfigFromJSONArgs, reply *string) (err error) {
	if len(args.Config) == 0 {
		*reply = utils.OK
		return
	}
	var oldCfg *CGRConfig
	updateDB := cfg.db != nil
	if !args.DryRun && updateDB { // need to update the DB but only parts
		oldCfg = cfg.Clone()
	}
	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}

	cfgV.reloadDPCache(sortedCfgSections...)
	if err = cfgV.loadCfgFromJSONWithLocks(bytes.NewBufferString(args.Config), sortedCfgSections); err != nil {
		return
	}

	//  lock all sections
	cfgV.rLockSections()
	err = cfgV.checkConfigSanity()
	cfgV.rUnlockSections() // unlock before checking the error
	if err != nil {
		return
	}
	if !args.DryRun {
		cfgV.reloadSections(sortedCfgSections...)
		if updateDB { // need to update the DB but only parts
			if err = storeDiffSections(ctx, sortedCfgSections, cfg.db, oldCfg, cfg); err != nil {
				return
			}
		}
	}
	*reply = utils.OK
	return
}

func (cfg *CGRConfig) reloadDPCache(sections ...string) {
	cfg.cacheDPMux.Lock()
	delete(cfg.cacheDP, utils.MetaAll)
	for _, sec := range sections {
		delete(cfg.cacheDP, sec)
	}
	cfg.cacheDPMux.Unlock()
}

// loadFromJSONDB Loads from json configuration object, will be used for defaults, config from file and reload
// this function ignores the config_db section
func (cfg *CGRConfig) LoadFromDB(jsnCfg ConfigDB) (err error) {
	// Load sections out of JSON config, stop on error
	cfg.lockSections()
	defer cfg.unlockSections()
	cfg.db = jsnCfg
	for _, loadFunc := range []func(ConfigDB) error{
		cfg.loadRPCConns,
		cfg.loadGeneralCfg, cfg.loadTemplateSCfg, cfg.loadCacheCfg, cfg.loadListenCfg,
		cfg.loadHTTPCfg, cfg.loadDataDBCfg, cfg.loadStorDBCfg,
		cfg.loadFilterSCfg,
		cfg.loadCdrsCfg, cfg.loadSessionSCfg,
		cfg.loadFreeswitchAgentCfg, cfg.loadKamAgentCfg,
		cfg.loadAsteriskAgentCfg, cfg.loadDiameterAgentCfg, cfg.loadRadiusAgentCfg,
		cfg.loadDNSAgentCfg, cfg.loadHTTPAgentCfg, cfg.loadAttributeSCfg,
		cfg.loadChargerSCfg, cfg.loadResourceSCfg, cfg.loadStatSCfg,
		cfg.loadThresholdSCfg, cfg.loadRouteSCfg, cfg.loadLoaderSCfg,
		cfg.loadSureTaxCfg, cfg.loadDispatcherSCfg,
		cfg.loadLoaderCgrCfg, cfg.loadMigratorCgrCfg, cfg.loadTLSCgrCfg,
		cfg.loadAnalyzerCgrCfg, cfg.loadApierCfg, cfg.loadErsCfg, cfg.loadEesCfg,
		cfg.loadRateSCfg, cfg.loadSIPAgentCfg, cfg.loadRegistrarCCfg,
		cfg.loadConfigSCfg, cfg.loadAPIBanCgrCfg, cfg.loadCoreSCfg, cfg.loadActionSCfg,
		cfg.loadAccountSCfg} {
		if err = loadFunc(jsnCfg); err != nil {
			return
		}
	}
	return cfg.checkConfigSanity()
}

func (cfg *CGRConfig) loadCfgFromDB(db ConfigDB, sections []string, ignoreConfigDB bool) (err error) {
	loadMap := cfg.getLoadFunctions()
	for _, section := range sections {
		if section == ConfigDBJSON {
			if ignoreConfigDB {
				continue
			}
			return fmt.Errorf("Invalid section: <%s> ", section)
		}
		fnct, has := loadMap[section]
		if !has {
			return fmt.Errorf("Invalid section: <%s> ", section)
		}
		cfg.lks[section].Lock()
		err = fnct(db)
		cfg.lks[section].Unlock()
		if err != nil {
			return
		}
	}
	return
}

func (cfg *CGRConfig) V1StoreCfgInDB(ctx *context.Context, args *SectionWithAPIOpts, rply *string) (err error) {
	if cfg.db == nil {
		return errors.New("no DB connection for config")
	}
	v1 := NewDefaultCGRConfig()
	if err = v1.loadFromJSONCfg(cfg.db); err != nil { // load the config from DB
		return
	}
	if len(args.Sections) != 0 && args.Sections[0] == utils.MetaAll {
		args.Sections = sortedCfgSections
	}
	err = storeDiffSections(ctx, args.Sections, cfg.db, v1, cfg)
	if err != nil {
		return
	}
	*rply = utils.OK
	return
}

func storeDiffSections(ctx *context.Context, sections []string, db ConfigDB, v1, v2 *CGRConfig) (err error) {
	for _, section := range sections {
		if err = storeDiffSection(ctx, section, db, v1, v2); err != nil {
			return
		}
	}
	return
}

func storeDiffSection(ctx *context.Context, section string, db ConfigDB, v1, v2 *CGRConfig) (err error) {
	switch section {
	case GeneralJSON:
		var jsn *GeneralJsonCfg
		if jsn, err = db.GeneralJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffGeneralJsonCfg(jsn, v1.GeneralCfg(), v2.GeneralCfg()))
	case RPCConnsJSON:
		var jsn RPCConnsJson
		if jsn, err = db.RPCConnJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRPCConnsJson(jsn, v1.RPCConns(), v2.RPCConns()))
	case CacheJSON:
		var jsn *CacheJsonCfg
		if jsn, err = db.CacheJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCacheJsonCfg(jsn, v1.CacheCfg(), v2.CacheCfg()))
	case ListenJSON:
		var jsn *ListenJsonCfg
		if jsn, err = db.ListenJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffListenJsonCfg(jsn, v1.ListenCfg(), v2.ListenCfg()))
	case HTTPJSON:
		var jsn *HTTPJsonCfg
		if jsn, err = db.HttpJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffHTTPJsonCfg(jsn, v1.HTTPCfg(), v2.HTTPCfg()))
	case StorDBJSON:
		var jsn *DbJsonCfg
		if jsn, err = db.DbJsonCfg(StorDBJSON); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffStorDBDbJsonCfg(jsn, v1.StorDbCfg(), v2.StorDbCfg()))
	case DataDBJSON:
		var jsn *DbJsonCfg
		if jsn, err = db.DbJsonCfg(DataDBJSON); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDataDbJsonCfg(jsn, v1.DataDbCfg(), v2.DataDbCfg()))
	case FilterSJSON:
		var jsn *FilterSJsonCfg
		if jsn, err = db.FilterSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFilterSJsonCfg(jsn, v1.FilterSCfg(), v2.FilterSCfg()))
	case CDRsJSON:
		var jsn *CdrsJsonCfg
		if jsn, err = db.CdrsJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCdrsJsonCfg(jsn, v1.CdrsCfg(), v2.CdrsCfg()))
	case ERsJSON:
		var jsn *ERsJsonCfg
		if jsn, err = db.ERsJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffERsJsonCfg(jsn, v1.ERsCfg(), v2.ERsCfg(), v2.GeneralCfg().RSRSep))
	case EEsJSON:
		var jsn *EEsJsonCfg
		if jsn, err = db.EEsJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffEEsJsonCfg(jsn, v1.EEsCfg(), v2.EEsCfg(), v2.GeneralCfg().RSRSep))
	case SessionSJSON:
		var jsn *SessionSJsonCfg
		if jsn, err = db.SessionSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSessionSJsonCfg(jsn, v1.SessionSCfg(), v2.SessionSCfg()))
	case FreeSWITCHAgentJSON:
		var jsn *FreeswitchAgentJsonCfg
		if jsn, err = db.FreeswitchAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFreeswitchAgentJsonCfg(jsn, v1.FsAgentCfg(), v2.FsAgentCfg()))
	case KamailioAgentJSON:
		var jsn *KamAgentJsonCfg
		if jsn, err = db.KamAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffKamAgentJsonCfg(jsn, v1.KamAgentCfg(), v2.KamAgentCfg()))
	case AsteriskAgentJSON:
		var jsn *AsteriskAgentJsonCfg
		if jsn, err = db.AsteriskAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAsteriskAgentJsonCfg(jsn, v1.AsteriskAgentCfg(), v2.AsteriskAgentCfg()))
	case DiameterAgentJSON:
		var jsn *DiameterAgentJsonCfg
		if jsn, err = db.DiameterAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDiameterAgentJsonCfg(jsn, v1.DiameterAgentCfg(), v2.DiameterAgentCfg(), v2.GeneralCfg().RSRSep))
	case RadiusAgentJSON:
		var jsn *RadiusAgentJsonCfg
		if jsn, err = db.RadiusAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRadiusAgentJsonCfg(jsn, v1.RadiusAgentCfg(), v2.RadiusAgentCfg(), v2.GeneralCfg().RSRSep))
	case HTTPAgentJSON:
		var jsn *[]*HttpAgentJsonCfg
		if jsn, err = db.HttpAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffHttpAgentsJsonCfg(jsn, v1.HTTPAgentCfg(), v2.HTTPAgentCfg(), v2.GeneralCfg().RSRSep))
	case DNSAgentJSON:
		var jsn *DNSAgentJsonCfg
		if jsn, err = db.DNSAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDNSAgentJsonCfg(jsn, v1.DNSAgentCfg(), v2.DNSAgentCfg(), v2.GeneralCfg().RSRSep))
	case AttributeSJSON:
		var jsn *AttributeSJsonCfg
		if jsn, err = db.AttributeServJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAttributeSJsonCfg(jsn, v1.AttributeSCfg(), v2.AttributeSCfg()))
	case ChargerSJSON:
		var jsn *ChargerSJsonCfg
		if jsn, err = db.ChargerServJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffChargerSJsonCfg(jsn, v1.ChargerSCfg(), v2.ChargerSCfg()))
	case ResourceSJSON:
		var jsn *ResourceSJsonCfg
		if jsn, err = db.ResourceSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffResourceSJsonCfg(jsn, v1.ResourceSCfg(), v2.ResourceSCfg()))
	case StatSJSON:
		var jsn *StatServJsonCfg
		if jsn, err = db.StatSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffStatServJsonCfg(jsn, v1.StatSCfg(), v2.StatSCfg()))
	case ThresholdSJSON:
		var jsn *ThresholdSJsonCfg
		if jsn, err = db.ThresholdSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffThresholdSJsonCfg(jsn, v1.ThresholdSCfg(), v2.ThresholdSCfg()))
	case RouteSJSON:
		var jsn *RouteSJsonCfg
		if jsn, err = db.RouteSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRouteSJsonCfg(jsn, v1.RouteSCfg(), v2.RouteSCfg()))
	case LoaderSJSON:
		var jsn []*LoaderJsonCfg
		if jsn, err = db.LoaderJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffLoadersJsonCfg(jsn, v1.LoaderCfg(), v2.LoaderCfg(), v2.GeneralCfg().RSRSep))
	case SureTaxJSON:
		var jsn *SureTaxJsonCfg
		if jsn, err = db.SureTaxJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSureTaxJsonCfg(jsn, v1.SureTaxCfg(), v2.SureTaxCfg(), v2.GeneralCfg().RSRSep))
	case DispatcherSJSON:
		var jsn *DispatcherSJsonCfg
		if jsn, err = db.DispatcherSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDispatcherSJsonCfg(jsn, v1.DispatcherSCfg(), v2.DispatcherSCfg()))
	case RegistrarCJSON:
		var jsn *RegistrarCJsonCfgs
		if jsn, err = db.RegistrarCJsonCfgs(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRegistrarCJsonCfgs(jsn, v1.RegistrarCCfg(), v2.RegistrarCCfg()))
	case LoaderJSON:
		var jsn *LoaderCfgJson
		if jsn, err = db.LoaderCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffLoaderCfgJson(jsn, v1.LoaderCgrCfg(), v2.LoaderCgrCfg()))
	case MigratorJSON:
		var jsn *MigratorCfgJson
		if jsn, err = db.MigratorCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffMigratorCfgJson(jsn, v1.MigratorCgrCfg(), v2.MigratorCgrCfg()))
	case TlsJSON:
		var jsn *TlsJsonCfg
		if jsn, err = db.TlsCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffTlsJsonCfg(jsn, v1.TLSCfg(), v2.TLSCfg()))
	case AnalyzerSJSON:
		var jsn *AnalyzerSJsonCfg
		if jsn, err = db.AnalyzerCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAnalyzerSJsonCfg(jsn, v1.AnalyzerSCfg(), v2.AnalyzerSCfg()))
	case AdminSJSON:
		var jsn *AdminSJsonCfg
		if jsn, err = db.AdminSCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAdminSJsonCfg(jsn, v1.AdminSCfg(), v2.AdminSCfg()))
	case RateSJSON:
		var jsn *RateSJsonCfg
		if jsn, err = db.RateCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRateSJsonCfg(jsn, v1.RateSCfg(), v2.RateSCfg()))
	case SIPAgentJSON:
		var jsn *SIPAgentJsonCfg
		if jsn, err = db.SIPAgentJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSIPAgentJsonCfg(jsn, v1.SIPAgentCfg(), v2.SIPAgentCfg(), v2.GeneralCfg().RSRSep))
	case TemplatesJSON:
		var jsn FcTemplatesJsonCfg
		if jsn, err = db.TemplateSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFcTemplatesJsonCfg(jsn, v1.TemplatesCfg(), v2.TemplatesCfg(), v2.GeneralCfg().RSRSep))
	case ConfigSJSON:
		var jsn *ConfigSCfgJson
		if jsn, err = db.ConfigSJsonCfg(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffConfigSCfgJson(jsn, v1.ConfigSCfg(), v2.ConfigSCfg()))
	case APIBanJSON:
		var jsn *APIBanJsonCfg
		if jsn, err = db.ApiBanCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAPIBanJsonCfg(jsn, v1.APIBanCfg(), v2.APIBanCfg()))
	case CoreSJSON:
		var jsn *CoreSJsonCfg
		if jsn, err = db.CoreSJSON(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCoreSJsonCfg(jsn, v1.CoreSCfg(), v2.CoreSCfg()))
	case ActionSJSON:
		var jsn *ActionSJsonCfg
		if jsn, err = db.ActionSCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffActionSJsonCfg(jsn, v1.ActionSCfg(), v2.ActionSCfg()))
	case AccountSJSON:
		var jsn *AccountSJsonCfg
		if jsn, err = db.AccountSCfgJson(); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAccountSJsonCfg(jsn, v1.AccountSCfg(), v2.AccountSCfg()))
	}
	return
}
