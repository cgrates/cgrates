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
			dft := NewDefaultCGRConfig()
			for _, section := range sections {
				if err = storeDiff(ctx, section, cfg.db, dft, oldCfg, cfg); err != nil {
					return
				}
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
			dft := NewDefaultCGRConfig()
			for _, section := range sortedCfgSections {
				if err = storeDiff(ctx, section, cfg.db, dft, oldCfg, cfg); err != nil {
					return
				}
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
		dft := NewDefaultCGRConfig()
		for _, section := range args.Sections {
			if err = storeDiff(ctx, section, cfg.db, dft, nil, cfg); err != nil {
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

func storeDiff(ctx *context.Context, section string, db ConfigDB, dft, v1, v2 *CGRConfig) (err error) {
	switch section {
	case GeneralJSON:
		var jsn *GeneralJsonCfg
		if jsn, err = db.GeneralJsonCfg(); err != nil {
			return
		}
		g1 := dft.GeneralCfg()
		if v1 != nil {
			g1 = v1.GeneralCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffGeneralJsonCfg(jsn, g1, v2.GeneralCfg()))
	case RPCConnsJSON:
		var jsn RPCConnsJson
		if jsn, err = db.RPCConnJsonCfg(); err != nil {
			return
		}
		g1 := dft.RPCConns()
		if v1 != nil {
			g1 = v1.RPCConns()
		} else {
			g1.loadFromJSONCfg(jsn)
		}
		return db.SetSection(ctx, section, diffRPCConnsJson(jsn, g1, v2.RPCConns()))
	case CacheJSON:
		var jsn *CacheJsonCfg
		if jsn, err = db.CacheJsonCfg(); err != nil {
			return
		}
		g1 := dft.CacheCfg()
		if v1 != nil {
			g1 = v1.CacheCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCacheJsonCfg(jsn, g1, v2.CacheCfg()))
	case ListenJSON:
		var jsn *ListenJsonCfg
		if jsn, err = db.ListenJsonCfg(); err != nil {
			return
		}
		g1 := dft.ListenCfg()
		if v1 != nil {
			g1 = v1.ListenCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffListenJsonCfg(jsn, g1, v2.ListenCfg()))
	case HTTPJSON:
		var jsn *HTTPJsonCfg
		if jsn, err = db.HttpJsonCfg(); err != nil {
			return
		}
		g1 := dft.HTTPCfg()
		if v1 != nil {
			g1 = v1.HTTPCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffHTTPJsonCfg(jsn, g1, v2.HTTPCfg()))
	case StorDBJSON:
		var jsn *DbJsonCfg
		if jsn, err = db.DbJsonCfg(StorDBJSON); err != nil {
			return
		}
		g1 := dft.StorDbCfg()
		if v1 != nil {
			g1 = v1.StorDbCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffStorDBDbJsonCfg(jsn, g1, v2.StorDbCfg()))
	case DataDBJSON:
		var jsn *DbJsonCfg
		if jsn, err = db.DbJsonCfg(DataDBJSON); err != nil {
			return
		}
		g1 := dft.DataDbCfg()
		if v1 != nil {
			g1 = v1.DataDbCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDataDbJsonCfg(jsn, g1, v2.DataDbCfg()))
	case FilterSJSON:
		var jsn *FilterSJsonCfg
		if jsn, err = db.FilterSJsonCfg(); err != nil {
			return
		}
		g1 := dft.FilterSCfg()
		if v1 != nil {
			g1 = v1.FilterSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFilterSJsonCfg(jsn, g1, v2.FilterSCfg()))
	case CDRsJSON:
		var jsn *CdrsJsonCfg
		if jsn, err = db.CdrsJsonCfg(); err != nil {
			return
		}
		g1 := dft.CdrsCfg()
		if v1 != nil {
			g1 = v1.CdrsCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCdrsJsonCfg(jsn, g1, v2.CdrsCfg()))
	case ERsJSON:
		var jsn *ERsJsonCfg
		if jsn, err = db.ERsJsonCfg(); err != nil {
			return
		}
		g1 := dft.ERsCfg()
		if v1 != nil {
			g1 = v1.ERsCfg()
		} else {
			if err = dft.loadTemplateSCfg(db); err != nil {
				return
			}
			if err = g1.loadFromJSONCfg(jsn, dft.TemplatesCfg(), dft.GeneralCfg().RSRSep, dft.dfltEvRdr); err != nil {
				return
			}
		}
		return db.SetSection(ctx, section, diffERsJsonCfg(jsn, g1, v2.ERsCfg(), v2.GeneralCfg().RSRSep))
	case EEsJSON:
		var jsn *EEsJsonCfg
		if jsn, err = db.EEsJsonCfg(); err != nil {
			return
		}
		g1 := dft.EEsCfg()
		if v1 != nil {
			g1 = v1.EEsCfg()
		} else {
			if err = dft.loadTemplateSCfg(db); err != nil {
				return
			}
			if err = g1.loadFromJSONCfg(jsn, dft.TemplatesCfg(), dft.GeneralCfg().RSRSep, dft.dfltEvExp); err != nil {
				return
			}
		}
		return db.SetSection(ctx, section, diffEEsJsonCfg(jsn, g1, v2.EEsCfg(), v2.GeneralCfg().RSRSep))
	case SessionSJSON:
		var jsn *SessionSJsonCfg
		if jsn, err = db.SessionSJsonCfg(); err != nil {
			return
		}
		g1 := dft.SessionSCfg()
		if v1 != nil {
			g1 = v1.SessionSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSessionSJsonCfg(jsn, g1, v2.SessionSCfg()))
	case FreeSWITCHAgentJSON:
		var jsn *FreeswitchAgentJsonCfg
		if jsn, err = db.FreeswitchAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.FsAgentCfg()
		if v1 != nil {
			g1 = v1.FsAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFreeswitchAgentJsonCfg(jsn, g1, v2.FsAgentCfg()))
	case KamailioAgentJSON:
		var jsn *KamAgentJsonCfg
		if jsn, err = db.KamAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.KamAgentCfg()
		if v1 != nil {
			g1 = v1.KamAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffKamAgentJsonCfg(jsn, g1, v2.KamAgentCfg()))
	case AsteriskAgentJSON:
		var jsn *AsteriskAgentJsonCfg
		if jsn, err = db.AsteriskAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.AsteriskAgentCfg()
		if v1 != nil {
			g1 = v1.AsteriskAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAsteriskAgentJsonCfg(jsn, g1, v2.AsteriskAgentCfg()))
	case DiameterAgentJSON:
		var jsn *DiameterAgentJsonCfg
		if jsn, err = db.DiameterAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.DiameterAgentCfg()
		if v1 != nil {
			g1 = v1.DiameterAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn, dft.GeneralCfg().RSRSep); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDiameterAgentJsonCfg(jsn, g1, v2.DiameterAgentCfg(), v2.GeneralCfg().RSRSep))
	case RadiusAgentJSON:
		var jsn *RadiusAgentJsonCfg
		if jsn, err = db.RadiusAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.RadiusAgentCfg()
		if v1 != nil {
			g1 = v1.RadiusAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn, dft.GeneralCfg().RSRSep); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRadiusAgentJsonCfg(jsn, g1, v2.RadiusAgentCfg(), v2.GeneralCfg().RSRSep))
	case HTTPAgentJSON:
		var jsn *RadiusAgentJsonCfg
		if jsn, err = db.RadiusAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.RadiusAgentCfg()
		if v1 != nil {
			g1 = v1.RadiusAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn, dft.GeneralCfg().RSRSep); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRadiusAgentJsonCfg(jsn, g1, v2.RadiusAgentCfg(), v2.GeneralCfg().RSRSep))
	case DNSAgentJSON:
		var jsn *DNSAgentJsonCfg
		if jsn, err = db.DNSAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.DNSAgentCfg()
		if v1 != nil {
			g1 = v1.DNSAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn, dft.GeneralCfg().RSRSep); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDNSAgentJsonCfg(jsn, g1, v2.DNSAgentCfg(), v2.GeneralCfg().RSRSep))
	case AttributeSJSON:
		var jsn *AttributeSJsonCfg
		if jsn, err = db.AttributeServJsonCfg(); err != nil {
			return
		}
		g1 := dft.AttributeSCfg()
		if v1 != nil {
			g1 = v1.AttributeSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAttributeSJsonCfg(jsn, g1, v2.AttributeSCfg()))
	case ChargerSJSON:
		var jsn *ChargerSJsonCfg
		if jsn, err = db.ChargerServJsonCfg(); err != nil {
			return
		}
		g1 := dft.ChargerSCfg()
		if v1 != nil {
			g1 = v1.ChargerSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffChargerSJsonCfg(jsn, g1, v2.ChargerSCfg()))
	case ResourceSJSON:
		var jsn *ResourceSJsonCfg
		if jsn, err = db.ResourceSJsonCfg(); err != nil {
			return
		}
		g1 := dft.ResourceSCfg()
		if v1 != nil {
			g1 = v1.ResourceSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffResourceSJsonCfg(jsn, g1, v2.ResourceSCfg()))
	case StatSJSON:
		var jsn *StatServJsonCfg
		if jsn, err = db.StatSJsonCfg(); err != nil {
			return
		}
		g1 := dft.StatSCfg()
		if v1 != nil {
			g1 = v1.StatSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffStatServJsonCfg(jsn, g1, v2.StatSCfg()))
	case ThresholdSJSON:
		var jsn *ThresholdSJsonCfg
		if jsn, err = db.ThresholdSJsonCfg(); err != nil {
			return
		}
		g1 := dft.ThresholdSCfg()
		if v1 != nil {
			g1 = v1.ThresholdSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffThresholdSJsonCfg(jsn, g1, v2.ThresholdSCfg()))
	case RouteSJSON:
		var jsn *RouteSJsonCfg
		if jsn, err = db.RouteSJsonCfg(); err != nil {
			return
		}
		g1 := dft.RouteSCfg()
		if v1 != nil {
			g1 = v1.RouteSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRouteSJsonCfg(jsn, g1, v2.RouteSCfg()))
	case LoaderSJSON:
		var jsn []*LoaderJsonCfg
		if jsn, err = db.LoaderJsonCfg(); err != nil {
			return
		}
		var g1 LoaderSCfgs
		if v1 != nil {
			g1 = v1.LoaderCfg()
		} else {
			if err = dft.loadTemplateSCfg(db); err != nil {
				return
			}
			if err = dft.loadLoaderSCfg(db); err != nil {
				return
			}
			g1 = dft.LoaderCfg()
		}
		return db.SetSection(ctx, section, diffLoadersJsonCfg(jsn, g1, v2.LoaderCfg(), v2.GeneralCfg().RSRSep))
	case SureTaxJSON:
		var jsn *SureTaxJsonCfg
		if jsn, err = db.SureTaxJsonCfg(); err != nil {
			return
		}
		g1 := dft.SureTaxCfg()
		if v1 != nil {
			g1 = v1.SureTaxCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSureTaxJsonCfg(jsn, g1, v2.SureTaxCfg(), v2.GeneralCfg().RSRSep))
	case DispatcherSJSON:
		var jsn *DispatcherSJsonCfg
		if jsn, err = db.DispatcherSJsonCfg(); err != nil {
			return
		}
		g1 := dft.DispatcherSCfg()
		if v1 != nil {
			g1 = v1.DispatcherSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDispatcherSJsonCfg(jsn, g1, v2.DispatcherSCfg()))
	case RegistrarCJSON:
		var jsn *RegistrarCJsonCfgs
		if jsn, err = db.RegistrarCJsonCfgs(); err != nil {
			return
		}
		g1 := dft.RegistrarCCfg()
		if v1 != nil {
			g1 = v1.RegistrarCCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRegistrarCJsonCfgs(jsn, g1, v2.RegistrarCCfg()))
	case LoaderJSON:
		var jsn *LoaderCfgJson
		if jsn, err = db.LoaderCfgJson(); err != nil {
			return
		}
		g1 := dft.LoaderCgrCfg()
		if v1 != nil {
			g1 = v1.LoaderCgrCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffLoaderCfgJson(jsn, g1, v2.LoaderCgrCfg()))
	case MigratorJSON:
		var jsn *MigratorCfgJson
		if jsn, err = db.MigratorCfgJson(); err != nil {
			return
		}
		g1 := dft.MigratorCgrCfg()
		if v1 != nil {
			g1 = v1.MigratorCgrCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffMigratorCfgJson(jsn, g1, v2.MigratorCgrCfg()))
	case TlsJSON:
		var jsn *TlsJsonCfg
		if jsn, err = db.TlsCfgJson(); err != nil {
			return
		}
		g1 := dft.TLSCfg()
		if v1 != nil {
			g1 = v1.TLSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffTlsJsonCfg(jsn, g1, v2.TLSCfg()))
	case AnalyzerSJSON:
		var jsn *AnalyzerSJsonCfg
		if jsn, err = db.AnalyzerCfgJson(); err != nil {
			return
		}
		g1 := dft.AnalyzerSCfg()
		if v1 != nil {
			g1 = v1.AnalyzerSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAnalyzerSJsonCfg(jsn, g1, v2.AnalyzerSCfg()))
	case AdminSJSON:
		var jsn *AdminSJsonCfg
		if jsn, err = db.AdminSCfgJson(); err != nil {
			return
		}
		g1 := dft.AdminSCfg()
		if v1 != nil {
			g1 = v1.AdminSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAdminSJsonCfg(jsn, g1, v2.AdminSCfg()))
	case RateSJSON:
		var jsn *RateSJsonCfg
		if jsn, err = db.RateCfgJson(); err != nil {
			return
		}
		g1 := dft.RateSCfg()
		if v1 != nil {
			g1 = v1.RateSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRateSJsonCfg(jsn, g1, v2.RateSCfg()))
	case SIPAgentJSON:
		var jsn *SIPAgentJsonCfg
		if jsn, err = db.SIPAgentJsonCfg(); err != nil {
			return
		}
		g1 := dft.SIPAgentCfg()
		if v1 != nil {
			g1 = v1.SIPAgentCfg()
		} else if err = g1.loadFromJSONCfg(jsn, dft.GeneralCfg().RSRSep); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSIPAgentJsonCfg(jsn, g1, v2.SIPAgentCfg(), v2.GeneralCfg().RSRSep))
	case TemplatesJSON:
		var jsn FcTemplatesJsonCfg
		if jsn, err = db.TemplateSJsonCfg(); err != nil {
			return
		}
		var g1 FCTemplates
		if v1 != nil {
			g1 = v1.TemplatesCfg()
		} else if err = dft.loadTemplateSCfg(db); err != nil {
			return
		} else {
			g1 = dft.TemplatesCfg()
		}
		return db.SetSection(ctx, section, diffFcTemplatesJsonCfg(jsn, g1, v2.TemplatesCfg(), v2.GeneralCfg().RSRSep))
	case ConfigSJSON:
		var jsn *ConfigSCfgJson
		if jsn, err = db.ConfigSJsonCfg(); err != nil {
			return
		}
		g1 := dft.ConfigSCfg()
		if v1 != nil {
			g1 = v1.ConfigSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffConfigSCfgJson(jsn, g1, v2.ConfigSCfg()))
	case APIBanJSON:
		var jsn *APIBanJsonCfg
		if jsn, err = db.ApiBanCfgJson(); err != nil {
			return
		}
		g1 := dft.APIBanCfg()
		if v1 != nil {
			g1 = v1.APIBanCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAPIBanJsonCfg(jsn, g1, v2.APIBanCfg()))
	case CoreSJSON:
		var jsn *CoreSJsonCfg
		if jsn, err = db.CoreSJSON(); err != nil {
			return
		}
		g1 := dft.CoreSCfg()
		if v1 != nil {
			g1 = v1.CoreSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCoreSJsonCfg(jsn, g1, v2.CoreSCfg()))
	case ActionSJSON:
		var jsn *ActionSJsonCfg
		if jsn, err = db.ActionSCfgJson(); err != nil {
			return
		}
		g1 := dft.ActionSCfg()
		if v1 != nil {
			g1 = v1.ActionSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffActionSJsonCfg(jsn, g1, v2.ActionSCfg()))
	case AccountSJSON:
		var jsn *AccountSJsonCfg
		if jsn, err = db.AccountSCfgJson(); err != nil {
			return
		}
		g1 := dft.AccountSCfg()
		if v1 != nil {
			g1 = v1.AccountSCfg()
		} else if err = g1.loadFromJSONCfg(jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAccountSJsonCfg(jsn, g1, v2.AccountSCfg()))
	}
	return
}
