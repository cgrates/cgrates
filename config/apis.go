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
	if sec, has := cfg.sections.Get(section); has {
		mp = sec.AsMapInterface(cfg.GeneralCfg().RSRSep)
		return
	}
	err = errors.New("Invalid section ")
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
	if err = cfgV.loadCfgWithLocks(ctx, cfg.ConfigPath, args.Section); err != nil {
		return
	}
	sections := []string{args.Section}
	allSections := args.Section == utils.MetaEmpty ||
		args.Section == utils.MetaAll
	if allSections {
		sections = cfgV.GetAllSectionIDs()
	}
	if cfg.db != nil {
		if err = cfgV.loadCfgFromDB(ctx, cfg.db, sections, allSections); err != nil {
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
		cfgV.reloadSections(sections...)
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
		args.Sections = cfg.GetAllSectionIDs()
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
	if sections.Size() == len(cfg.sections) {
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

	cfgV := cfg
	if args.DryRun {
		cfgV = cfg.Clone()
	}
	var oldCfg *CGRConfig
	updateDB := cfg.db != nil
	if !args.DryRun && updateDB { // need to update the DB but only parts
		oldCfg = cfg.Clone()
	}
	sectionNms := make([]string, 0, len(args.Config))
	sections := make(Sections, 0, len(args.Config))
	for secNm := range args.Config {
		sec, has := cfgV.sections.Get(secNm)
		if !has {
			return fmt.Errorf("Invalid section <%s> ", secNm)
		}
		sections = append(sections, sec)
		sectionNms = append(sectionNms, secNm)
	}
	var b []byte
	if b, err = json.Marshal(args.Config); err != nil {
		return
	}

	cfgV.reloadDPCache(sectionNms...)
	cfgV.LockSections(sectionNms...)
	err = loadConfigFromReader(ctx, bytes.NewBuffer(b), sections, false, cfgV)
	cfgV.UnlockSections(sectionNms...)
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
		cfgV.reloadSections(sectionNms...)
		if updateDB { // need to update the DB but only parts
			if err = storeDiffSections(ctx, sectionNms, cfgV.db, oldCfg, cfgV); err != nil {
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

	sections := cfg.GetAllSectionIDs()
	cfgV.reloadDPCache(sections...)
	cfg.LockSections(sections...)
	err = loadConfigFromReader(ctx, bytes.NewBufferString(args.Config), cfgV.sections, false, cfgV)
	cfg.UnlockSections(sections...)
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
		cfgV.reloadSections(sections...)
		if updateDB { // need to update the DB but only parts
			if err = storeDiffSections(ctx, sections, cfg.db, oldCfg, cfg); err != nil {
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
func (cfg *CGRConfig) LoadFromDB(ctx *context.Context, jsnCfg ConfigDB) (err error) {
	// Load sections out of JSON config, stop on error
	cfg.lockSections()
	defer cfg.unlockSections()
	cfg.db = jsnCfg
	if err = cfg.sections.LoadWithout(ctx, jsnCfg, cfg, ConfigDBJSON); err != nil {
		return
	}
	return cfg.checkConfigSanity()
}

// LoadFromPath reads all json files out of a folder/subfolders and loads them up in lexical order
func (cfg *CGRConfig) LoadFromPath(ctx *context.Context, path string) (err error) {
	cfg.ConfigPath = path
	if err = loadConfigFromPath(ctx, path, cfg.sections, false, cfg); err != nil {
		return
	}
	return cfg.checkConfigSanity()
}

func (cfg *CGRConfig) loadCfgFromDB(ctx *context.Context, db ConfigDB, sections []string, ignoreConfigDB bool) (err error) {
	for _, section := range sections {
		if section == ConfigDBJSON {
			if ignoreConfigDB {
				continue
			}
			return fmt.Errorf("Invalid section: <%s> ", section)
		}
		sec, has := cfg.sections.Get(section)
		if !has {
			return fmt.Errorf("Invalid section: <%s> ", section)
		}
		cfg.lks[section].Lock()
		err = sec.Load(ctx, db, cfg)
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
	if err = v1.sections.Load(ctx, cfg.db, cfg); err != nil { // load the config from DB
		return
	}
	if len(args.Sections) != 0 && args.Sections[0] == utils.MetaAll {
		args.Sections = cfg.GetAllSectionIDs()
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
		jsn := new(GeneralJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffGeneralJsonCfg(jsn, v1.GeneralCfg(), v2.GeneralCfg()))
	case RPCConnsJSON:
		jsn := make(RPCConnsJson)
		if err = db.GetSection(ctx, section, &jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRPCConnsJson(jsn, v1.RPCConns(), v2.RPCConns()))
	case CacheJSON:
		jsn := new(CacheJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCacheJsonCfg(jsn, v1.CacheCfg(), v2.CacheCfg()))
	case ListenJSON:
		jsn := new(ListenJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffListenJsonCfg(jsn, v1.ListenCfg(), v2.ListenCfg()))
	case HTTPJSON:
		jsn := new(HTTPJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffHTTPJsonCfg(jsn, v1.HTTPCfg(), v2.HTTPCfg()))
	case StorDBJSON:
		jsn := new(DbJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffStorDBDbJsonCfg(jsn, v1.StorDbCfg(), v2.StorDbCfg()))
	case DataDBJSON:
		jsn := new(DbJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDataDbJsonCfg(jsn, v1.DataDbCfg(), v2.DataDbCfg()))
	case FilterSJSON:
		jsn := new(FilterSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFilterSJsonCfg(jsn, v1.FilterSCfg(), v2.FilterSCfg()))
	case CDRsJSON:
		jsn := new(CdrsJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCdrsJsonCfg(jsn, v1.CdrsCfg(), v2.CdrsCfg()))
	case ERsJSON:
		jsn := new(ERsJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffERsJsonCfg(jsn, v1.ERsCfg(), v2.ERsCfg(), v2.GeneralCfg().RSRSep))
	case EEsJSON:
		jsn := new(EEsJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffEEsJsonCfg(jsn, v1.EEsCfg(), v2.EEsCfg(), v2.GeneralCfg().RSRSep))
	case SessionSJSON:
		jsn := new(SessionSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSessionSJsonCfg(jsn, v1.SessionSCfg(), v2.SessionSCfg()))
	case FreeSWITCHAgentJSON:
		jsn := new(FreeswitchAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFreeswitchAgentJsonCfg(jsn, v1.FsAgentCfg(), v2.FsAgentCfg()))
	case KamailioAgentJSON:
		jsn := new(KamAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffKamAgentJsonCfg(jsn, v1.KamAgentCfg(), v2.KamAgentCfg()))
	case AsteriskAgentJSON:
		jsn := new(AsteriskAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAsteriskAgentJsonCfg(jsn, v1.AsteriskAgentCfg(), v2.AsteriskAgentCfg()))
	case DiameterAgentJSON:
		jsn := new(DiameterAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDiameterAgentJsonCfg(jsn, v1.DiameterAgentCfg(), v2.DiameterAgentCfg(), v2.GeneralCfg().RSRSep))
	case RadiusAgentJSON:
		jsn := new(RadiusAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRadiusAgentJsonCfg(jsn, v1.RadiusAgentCfg(), v2.RadiusAgentCfg(), v2.GeneralCfg().RSRSep))
	case HTTPAgentJSON:
		jsn := new([]*HttpAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffHttpAgentsJsonCfg(jsn, v1.HTTPAgentCfg(), v2.HTTPAgentCfg(), v2.GeneralCfg().RSRSep))
	case DNSAgentJSON:
		jsn := new(DNSAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDNSAgentJsonCfg(jsn, v1.DNSAgentCfg(), v2.DNSAgentCfg(), v2.GeneralCfg().RSRSep))
	case AttributeSJSON:
		jsn := new(AttributeSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAttributeSJsonCfg(jsn, v1.AttributeSCfg(), v2.AttributeSCfg()))
	case ChargerSJSON:
		jsn := new(ChargerSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffChargerSJsonCfg(jsn, v1.ChargerSCfg(), v2.ChargerSCfg()))
	case ResourceSJSON:
		jsn := new(ResourceSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffResourceSJsonCfg(jsn, v1.ResourceSCfg(), v2.ResourceSCfg()))
	case StatSJSON:
		jsn := new(StatServJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffStatServJsonCfg(jsn, v1.StatSCfg(), v2.StatSCfg()))
	case ThresholdSJSON:
		jsn := new(ThresholdSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffThresholdSJsonCfg(jsn, v1.ThresholdSCfg(), v2.ThresholdSCfg()))
	case RouteSJSON:
		jsn := new(RouteSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRouteSJsonCfg(jsn, v1.RouteSCfg(), v2.RouteSCfg()))
	case LoaderSJSON:
		jsn := make([]*LoaderJsonCfg, 0)
		if err = db.GetSection(ctx, section, &jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffLoadersJsonCfg(jsn, v1.LoaderCfg(), v2.LoaderCfg(), v2.GeneralCfg().RSRSep))
	case SureTaxJSON:
		jsn := new(SureTaxJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSureTaxJsonCfg(jsn, v1.SureTaxCfg(), v2.SureTaxCfg(), v2.GeneralCfg().RSRSep))
	case DispatcherSJSON:
		jsn := new(DispatcherSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffDispatcherSJsonCfg(jsn, v1.DispatcherSCfg(), v2.DispatcherSCfg()))
	case RegistrarCJSON:
		jsn := new(RegistrarCJsonCfgs)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRegistrarCJsonCfgs(jsn, v1.RegistrarCCfg(), v2.RegistrarCCfg()))
	case LoaderJSON:
		jsn := new(LoaderCfgJson)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffLoaderCfgJson(jsn, v1.LoaderCgrCfg(), v2.LoaderCgrCfg()))
	case MigratorJSON:
		jsn := new(MigratorCfgJson)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffMigratorCfgJson(jsn, v1.MigratorCgrCfg(), v2.MigratorCgrCfg()))
	case TlsJSON:
		jsn := new(TlsJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffTlsJsonCfg(jsn, v1.TLSCfg(), v2.TLSCfg()))
	case AnalyzerSJSON:
		jsn := new(AnalyzerSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAnalyzerSJsonCfg(jsn, v1.AnalyzerSCfg(), v2.AnalyzerSCfg()))
	case AdminSJSON:
		jsn := new(AdminSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAdminSJsonCfg(jsn, v1.AdminSCfg(), v2.AdminSCfg()))
	case RateSJSON:
		jsn := new(RateSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffRateSJsonCfg(jsn, v1.RateSCfg(), v2.RateSCfg()))
	case SIPAgentJSON:
		jsn := new(SIPAgentJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffSIPAgentJsonCfg(jsn, v1.SIPAgentCfg(), v2.SIPAgentCfg(), v2.GeneralCfg().RSRSep))
	case TemplatesJSON:
		jsn := make(FcTemplatesJsonCfg)
		if err = db.GetSection(ctx, section, &jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffFcTemplatesJsonCfg(jsn, v1.TemplatesCfg(), v2.TemplatesCfg(), v2.GeneralCfg().RSRSep))
	case ConfigSJSON:
		jsn := new(ConfigSCfgJson)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffConfigSCfgJson(jsn, v1.ConfigSCfg(), v2.ConfigSCfg()))
	case APIBanJSON:
		jsn := new(APIBanJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAPIBanJsonCfg(jsn, v1.APIBanCfg(), v2.APIBanCfg()))
	case CoreSJSON:
		jsn := new(CoreSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffCoreSJsonCfg(jsn, v1.CoreSCfg(), v2.CoreSCfg()))
	case ActionSJSON:
		jsn := new(ActionSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffActionSJsonCfg(jsn, v1.ActionSCfg(), v2.ActionSCfg()))
	case AccountSJSON:
		jsn := new(AccountSJsonCfg)
		if err = db.GetSection(ctx, section, jsn); err != nil {
			return
		}
		return db.SetSection(ctx, section, diffAccountSJsonCfg(jsn, v1.AccountSCfg(), v2.AccountSCfg()))
	}
	return
}
