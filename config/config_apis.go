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
	case MailerJSON:
		mp = cfg.MailerCfg().AsMapInterface()
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
	Path    string
	Section string
	DryRun  bool
}

// V1ReloadConfig reloads the configuration
func (cfg *CGRConfig) V1ReloadConfig(ctx *context.Context, args *ReloadArgs, reply *string) (err error) {
	updateDB := cfg.db != nil
	if !updateDB &&
		args.Path == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing("Path")
	} else if updateDB &&
		args.Path != utils.EmptyString {
		return fmt.Errorf("Reload from the path is disabled when the configDB is enabled")
	}
	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}
	cfgV.reloadDPCache(args.Section)
	if updateDB {
		sections := []string{args.Section}
		if args.Section == utils.MetaEmpty ||
			args.Section == utils.MetaAll {
			sections = sortedCfgSections[:len(sortedCfgSections)-1] // all exept the configDB section
		}
		err = cfgV.loadCfgFromDB(cfg.db, sections)
	} else {
		err = cfgV.loadCfgWithLocks(args.Path, args.Section)
	}
	if err != nil {
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

	var mp map[string]interface{}
	updateDB := cfg.db != nil
	if !args.DryRun && updateDB { // need to update the DB but only parts
		if err = cfg.V1GetConfig(ctx, &SectionWithAPIOpts{Sections: sections}, &mp); err != nil {
			return
		}
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
			// ToDo: add here the call
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
	var mp map[string]interface{}
	updateDB := cfg.db != nil
	if !args.DryRun && updateDB { // need to update the DB but only parts
		if err = cfg.V1GetConfig(ctx, &SectionWithAPIOpts{Sections: sortedCfgSections}, &mp); err != nil {
			return
		}
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
			// ToDo: add here the call
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
		cfg.loadMailerCfg, cfg.loadSureTaxCfg, cfg.loadDispatcherSCfg,
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

func (cfg *CGRConfig) loadCfgFromDB(db ConfigDB, sections []string) (err error) {
	loadMap := cfg.getLoadFunctions()
	for _, section := range sections {
		if fnct, has := loadMap[section]; !has ||
			section == ConfigDBJSON {
			return fmt.Errorf("Invalid section: <%s> ", section)
		} else {
			cfg.lks[section].Lock()
			err = fnct(db)
			cfg.lks[section].Unlock()
			if err != nil {
				return
			}
		}
	}
	return
}

func (cfg *CGRConfig) V1StoreCfgInDB(ctx *context.Context, args *SectionWithAPIOpts, rply *string) (err error) {
	if cfg.db == nil {
		return errors.New("no DB connection for config")
	}
	if len(args.Sections) != 0 && args.Sections[0] != utils.MetaAll {
		for _, section := range args.Sections {
			var mp interface{}
			if mp, err = cfg.getSectionAsMap(section); err != nil {
				return
			}
			var cfgSec interface{}
			if cfgSec, err = prepareSectionFromMap(section, mp); err != nil {
				return
			}
			if err = cfg.db.SetSection(ctx, section, cfgSec); err != nil {
				return
			}
		}
	}
	cfg.rLockSections()
	mp := cfg.AsMapInterface(cfg.generalCfg.RSRSep)
	cfg.rUnlockSections()
	var data []byte
	if data, err = json.Marshal(mp); err != nil {
		return
	}
	var dp ConfigDB
	if dp, err = NewCgrJsonCfgFromBytes(data); err != nil {
		return
	}
	var sc interface{}
	if sc, err = dp.GeneralJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, GeneralJSON, sc); err != nil {
		return
	}
	if sc, err = dp.RPCConnJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, RPCConnsJSON, sc); err != nil {
		return
	}
	if sc, err = dp.CacheJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, CacheJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ListenJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ListenJSON, sc); err != nil {
		return
	}
	if sc, err = dp.HttpJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, HTTPJSON, sc); err != nil {
		return
	}
	if sc, err = dp.DbJsonCfg(StorDBJSON); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, StorDBJSON, sc); err != nil {
		return
	}
	if sc, err = dp.DbJsonCfg(DataDBJSON); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, DataDBJSON, sc); err != nil {
		return
	}
	if sc, err = dp.FilterSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, FilterSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.CdrsJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, CDRsJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ERsJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ERsJSON, sc); err != nil {
		return
	}
	if sc, err = dp.EEsJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, EEsJSON, sc); err != nil {
		return
	}
	if sc, err = dp.SessionSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, SessionSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.FreeswitchAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, FreeSWITCHAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.KamAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, KamailioAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.AsteriskAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, AsteriskAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.DiameterAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, DiameterAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.RadiusAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, RadiusAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.HttpAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, HTTPAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.DNSAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, DNSAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.AttributeServJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, AttributeSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ChargerServJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ChargerSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ResourceSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ResourceSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.StatSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, StatSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ThresholdSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ThresholdSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.RouteSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, RouteSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.LoaderJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, LoaderSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.MailerJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, MailerJSON, sc); err != nil {
		return
	}
	if sc, err = dp.SureTaxJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, SureTaxJSON, sc); err != nil {
		return
	}
	if sc, err = dp.DispatcherSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, DispatcherSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.RegistrarCJsonCfgs(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, RegistrarCJSON, sc); err != nil {
		return
	}
	if sc, err = dp.LoaderCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, LoaderJSON, sc); err != nil {
		return
	}
	if sc, err = dp.MigratorCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, MigratorJSON, sc); err != nil {
		return
	}
	if sc, err = dp.TlsCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, TlsJSON, sc); err != nil {
		return
	}
	if sc, err = dp.AnalyzerCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, AnalyzerSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.AdminSCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, AdminSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.RateCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, RateSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.SIPAgentJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, SIPAgentJSON, sc); err != nil {
		return
	}
	if sc, err = dp.TemplateSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, TemplatesJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ConfigSJsonCfg(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ConfigSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ApiBanCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, APIBanJSON, sc); err != nil {
		return
	}
	if sc, err = dp.CoreSJSON(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, CoreSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.ActionSCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, ActionSJSON, sc); err != nil {
		return
	}
	if sc, err = dp.AccountSCfgJson(); err != nil {
		return
	} else if err = cfg.db.SetSection(ctx, AccountSJSON, sc); err != nil {
		return
	}
	*rply = utils.OK
	return
}

func prepareSectionFromMap(section string, mp interface{}) (cfgSec interface{}, err error) {
	var data []byte
	if data, err = json.Marshal(mp); err != nil {
		return
	}
	switch section {
	case GeneralJSON:
		cfgSec = new(GeneralJsonCfg)
	case RPCConnsJSON:
		cfgSec = make(RPCConnsJson)
	case CacheJSON:
		cfgSec = new(CacheJsonCfg)
	case ListenJSON:
		cfgSec = new(ListenJsonCfg)
	case HTTPJSON:
		cfgSec = new(HTTPJsonCfg)
	case StorDBJSON:
		cfgSec = new(DbJsonCfg)
	case DataDBJSON:
		cfgSec = new(DbJsonCfg)
	case FilterSJSON:
		cfgSec = new(FilterSJsonCfg)
	case CDRsJSON:
		cfgSec = new(CdrsJsonCfg)
	case ERsJSON:
		cfgSec = new(ERsJsonCfg)
	case EEsJSON:
		cfgSec = new(EEsJsonCfg)
	case SessionSJSON:
		cfgSec = new(SessionSJsonCfg)
	case FreeSWITCHAgentJSON:
		cfgSec = new(FreeswitchAgentJsonCfg)
	case KamailioAgentJSON:
		cfgSec = new(KamAgentJsonCfg)
	case AsteriskAgentJSON:
		cfgSec = new(AsteriskAgentJsonCfg)
	case DiameterAgentJSON:
		cfgSec = new(DiameterAgentJsonCfg)
	case RadiusAgentJSON:
		cfgSec = new(RadiusAgentJsonCfg)
	case HTTPAgentJSON:
		cfgSec = new([]*HttpAgentJsonCfg)
	case DNSAgentJSON:
		cfgSec = new(DNSAgentJsonCfg)
	case AttributeSJSON:
		cfgSec = new(AttributeSJsonCfg)
	case ChargerSJSON:
		cfgSec = new(ChargerSJsonCfg)
	case ResourceSJSON:
		cfgSec = new(ResourceSJsonCfg)
	case StatSJSON:
		cfgSec = new(StatServJsonCfg)
	case ThresholdSJSON:
		cfgSec = new(ThresholdSJsonCfg)
	case RouteSJSON:
		cfgSec = new(RouteSJsonCfg)
	case LoaderSJSON:
		cfgSec = make([]*LoaderJsonCfg, 0)
	case MailerJSON:
		cfgSec = new(MailerJsonCfg)
	case SureTaxJSON:
		cfgSec = new(SureTaxJsonCfg)
	case DispatcherSJSON:
		cfgSec = new(DispatcherSJsonCfg)
	case RegistrarCJSON:
		cfgSec = new(RegistrarCJsonCfgs)
	case LoaderJSON:
		cfgSec = new(LoaderCfgJson)
	case MigratorJSON:
		cfgSec = new(MigratorCfgJson)
	case TlsJSON:
		cfgSec = new(TlsJsonCfg)
	case AnalyzerSJSON:
		cfgSec = new(AnalyzerSJsonCfg)
	case AdminSJSON:
		cfgSec = new(AdminSJsonCfg)
	case RateSJSON:
		cfgSec = new(RateSJsonCfg)
	case SIPAgentJSON:
		cfgSec = new(SIPAgentJsonCfg)
	case TemplatesJSON:
		cfgSec = make(FcTemplatesJsonCfg)
	case ConfigSJSON:
		cfgSec = new(ConfigSCfgJson)
	case APIBanJSON:
		cfgSec = new(APIBanJsonCfg)
	case CoreSJSON:
		cfgSec = new(CoreSJsonCfg)
	case ActionSJSON:
		cfgSec = new(ActionSJsonCfg)
	case AccountSJSON:
		cfgSec = new(AccountSJsonCfg)
	}

	err = json.Unmarshal(data, cfgSec)
	return
}

func updateSections(ctx *context.Context, db ConfigDB, mp map[string]interface{}) (err error) {
	for section, val := range mp {
		var sec interface{}
		if sec, err = prepareSectionFromDB(section, db); err != nil {
			return
		}
		var data []byte
		if data, err = json.Marshal(val); err != nil {
			return
		}
		if err = json.Unmarshal(data, &sec); err != nil {
			return
		}
		if err = db.SetSection(ctx, section, sec); err != nil {
			return
		}
	}
	return
}

func prepareSectionFromDB(section string, db ConfigDB) (cfgSec interface{}, err error) {
	switch section {
	case GeneralJSON:
		cfgSec, err = db.GeneralJsonCfg()
	case RPCConnsJSON:
		cfgSec, err = db.RPCConnJsonCfg()
	case CacheJSON:
		cfgSec, err = db.CacheJsonCfg()
	case ListenJSON:
		cfgSec, err = db.ListenJsonCfg()
	case HTTPJSON:
		cfgSec, err = db.HttpJsonCfg()
	case StorDBJSON:
		cfgSec, err = db.DbJsonCfg(StorDBJSON)
	case DataDBJSON:
		cfgSec, err = db.DbJsonCfg(DataDBJSON)
	case FilterSJSON:
		cfgSec, err = db.FilterSJsonCfg()
	case CDRsJSON:
		cfgSec, err = db.CdrsJsonCfg()
	case ERsJSON:
		cfgSec, err = db.ERsJsonCfg()
	case EEsJSON:
		cfgSec, err = db.EEsJsonCfg()
	case SessionSJSON:
		cfgSec, err = db.SessionSJsonCfg()
	case FreeSWITCHAgentJSON:
		cfgSec, err = db.FreeswitchAgentJsonCfg()
	case KamailioAgentJSON:
		cfgSec, err = db.KamAgentJsonCfg()
	case AsteriskAgentJSON:
		cfgSec, err = db.AsteriskAgentJsonCfg()
	case DiameterAgentJSON:
		cfgSec, err = db.DiameterAgentJsonCfg()
	case RadiusAgentJSON:
		cfgSec, err = db.RadiusAgentJsonCfg()
	case HTTPAgentJSON:
		cfgSec, err = db.HttpAgentJsonCfg()
	case DNSAgentJSON:
		cfgSec, err = db.DNSAgentJsonCfg()
	case AttributeSJSON:
		cfgSec, err = db.AttributeServJsonCfg()
	case ChargerSJSON:
		cfgSec, err = db.ChargerServJsonCfg()
	case ResourceSJSON:
		cfgSec, err = db.ResourceSJsonCfg()
	case StatSJSON:
		cfgSec, err = db.StatSJsonCfg()
	case ThresholdSJSON:
		cfgSec, err = db.ThresholdSJsonCfg()
	case RouteSJSON:
		cfgSec, err = db.RouteSJsonCfg()
	case LoaderSJSON:
		cfgSec, err = db.LoaderJsonCfg()
	case MailerJSON:
		cfgSec, err = db.MailerJsonCfg()
	case SureTaxJSON:
		cfgSec, err = db.SureTaxJsonCfg()
	case DispatcherSJSON:
		cfgSec, err = db.DispatcherSJsonCfg()
	case RegistrarCJSON:
		cfgSec, err = db.RegistrarCJsonCfgs()
	case LoaderJSON:
		cfgSec, err = db.LoaderCfgJson()
	case MigratorJSON:
		cfgSec, err = db.MigratorCfgJson()
	case TlsJSON:
		cfgSec, err = db.TlsCfgJson()
	case AnalyzerSJSON:
		cfgSec, err = db.AnalyzerCfgJson()
	case AdminSJSON:
		cfgSec, err = db.AdminSCfgJson()
	case RateSJSON:
		cfgSec, err = db.RateCfgJson()
	case SIPAgentJSON:
		cfgSec, err = db.SIPAgentJsonCfg()
	case TemplatesJSON:
		cfgSec, err = db.TemplateSJsonCfg()
	case ConfigSJSON:
		cfgSec, err = db.ConfigSJsonCfg()
	case APIBanJSON:
		cfgSec, err = db.ApiBanCfgJson()
	case CoreSJSON:
		cfgSec, err = db.CoreSJSON()
	case ActionSJSON:
		cfgSec, err = db.ActionSCfgJson()
	case AccountSJSON:
		cfgSec, err = db.AccountSCfgJson()
	}
	return
}
