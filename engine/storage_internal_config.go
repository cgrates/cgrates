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

package engine

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// no config for internal

func (*InternalDB) GeneralJsonCfg() (*config.GeneralJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.GeneralJSON)
	v, _ := cfg.(*config.GeneralJsonCfg)
	return v, nil
}
func (*InternalDB) RPCConnJsonCfg() (config.RPCConnsJson, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.CacheJSON)
	v, _ := cfg.(config.RPCConnsJson)
	return v, nil
}
func (*InternalDB) CacheJsonCfg() (*config.CacheJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ListenJSON)
	v, _ := cfg.(*config.CacheJsonCfg)
	return v, nil
}
func (*InternalDB) ListenJsonCfg() (*config.ListenJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.HTTPJSON)
	v, _ := cfg.(*config.ListenJsonCfg)
	return v, nil
}
func (*InternalDB) HttpJsonCfg() (*config.HTTPJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.StorDBJSON)
	v, _ := cfg.(*config.HTTPJsonCfg)
	return v, nil
}
func (*InternalDB) DbJsonCfg(section string) (*config.DbJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, section)
	v, _ := cfg.(*config.DbJsonCfg)
	return v, nil
}
func (*InternalDB) FilterSJsonCfg() (*config.FilterSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.CDRsJSON)
	v, _ := cfg.(*config.FilterSJsonCfg)
	return v, nil
}
func (*InternalDB) CdrsJsonCfg() (*config.CdrsJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.SessionSJSON)
	v, _ := cfg.(*config.CdrsJsonCfg)
	return v, nil
}
func (*InternalDB) ERsJsonCfg() (*config.ERsJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.FreeSWITCHAgentJSON)
	v, _ := cfg.(*config.ERsJsonCfg)
	return v, nil
}
func (*InternalDB) EEsJsonCfg() (*config.EEsJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.KamailioAgentJSON)
	v, _ := cfg.(*config.EEsJsonCfg)
	return v, nil
}
func (*InternalDB) SessionSJsonCfg() (*config.SessionSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.AsteriskAgentJSON)
	v, _ := cfg.(*config.SessionSJsonCfg)
	return v, nil
}
func (*InternalDB) FreeswitchAgentJsonCfg() (*config.FreeswitchAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.DiameterAgentJSON)
	v, _ := cfg.(*config.FreeswitchAgentJsonCfg)
	return v, nil
}
func (*InternalDB) KamAgentJsonCfg() (*config.KamAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.RadiusAgentJSON)
	v, _ := cfg.(*config.KamAgentJsonCfg)
	return v, nil
}
func (*InternalDB) AsteriskAgentJsonCfg() (*config.AsteriskAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.HTTPAgentJSON)
	v, _ := cfg.(*config.AsteriskAgentJsonCfg)
	return v, nil
}
func (*InternalDB) DiameterAgentJsonCfg() (*config.DiameterAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.AttributeSJSON)
	v, _ := cfg.(*config.DiameterAgentJsonCfg)
	return v, nil
}
func (*InternalDB) RadiusAgentJsonCfg() (*config.RadiusAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ResourceSJSON)
	v, _ := cfg.(*config.RadiusAgentJsonCfg)
	return v, nil
}
func (*InternalDB) HttpAgentJsonCfg() (*[]*config.HttpAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.StatSJSON)
	v, _ := cfg.(*[]*config.HttpAgentJsonCfg)
	return v, nil
}
func (*InternalDB) DNSAgentJsonCfg() (*config.DNSAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ThresholdSJSON)
	v, _ := cfg.(*config.DNSAgentJsonCfg)
	return v, nil
}
func (*InternalDB) AttributeServJsonCfg() (*config.AttributeSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.RouteSJSON)
	v, _ := cfg.(*config.AttributeSJsonCfg)
	return v, nil
}
func (*InternalDB) ChargerServJsonCfg() (*config.ChargerSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.LoaderSJSON)
	v, _ := cfg.(*config.ChargerSJsonCfg)
	return v, nil
}
func (*InternalDB) ResourceSJsonCfg() (*config.ResourceSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.MailerJSON)
	v, _ := cfg.(*config.ResourceSJsonCfg)
	return v, nil
}
func (*InternalDB) StatSJsonCfg() (*config.StatServJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.SureTaxJSON)
	v, _ := cfg.(*config.StatServJsonCfg)
	return v, nil
}
func (*InternalDB) ThresholdSJsonCfg() (*config.ThresholdSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.DispatcherSJSON)
	v, _ := cfg.(*config.ThresholdSJsonCfg)
	return v, nil
}
func (*InternalDB) RouteSJsonCfg() (*config.RouteSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.RegistrarCJSON)
	v, _ := cfg.(*config.RouteSJsonCfg)
	return v, nil
}
func (*InternalDB) LoaderJsonCfg() ([]*config.LoaderJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.LoaderJSON)
	v, _ := cfg.([]*config.LoaderJsonCfg)
	return v, nil
}
func (*InternalDB) MailerJsonCfg() (*config.MailerJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.MigratorJSON)
	v, _ := cfg.(*config.MailerJsonCfg)
	return v, nil
}
func (*InternalDB) SureTaxJsonCfg() (*config.SureTaxJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ChargerSJSON)
	v, _ := cfg.(*config.SureTaxJsonCfg)
	return v, nil
}
func (*InternalDB) DispatcherSJsonCfg() (*config.DispatcherSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.TlsJSON)
	v, _ := cfg.(*config.DispatcherSJsonCfg)
	return v, nil
}
func (*InternalDB) RegistrarCJsonCfgs() (*config.RegistrarCJsonCfgs, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.AnalyzerSJSON)
	v, _ := cfg.(*config.RegistrarCJsonCfgs)
	return v, nil
}
func (*InternalDB) LoaderCfgJson() (*config.LoaderCfgJson, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.AdminSJSON)
	v, _ := cfg.(*config.LoaderCfgJson)
	return v, nil
}
func (*InternalDB) MigratorCfgJson() (*config.MigratorCfgJson, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.DNSAgentJSON)
	v, _ := cfg.(*config.MigratorCfgJson)
	return v, nil
}
func (*InternalDB) TlsCfgJson() (*config.TlsJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ERsJSON)
	v, _ := cfg.(*config.TlsJsonCfg)
	return v, nil
}
func (*InternalDB) AnalyzerCfgJson() (*config.AnalyzerSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.EEsJSON)
	v, _ := cfg.(*config.AnalyzerSJsonCfg)
	return v, nil
}
func (*InternalDB) AdminSCfgJson() (*config.AdminSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.RateSJSON)
	v, _ := cfg.(*config.AdminSJsonCfg)
	return v, nil
}
func (*InternalDB) RateCfgJson() (*config.RateSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ActionSJSON)
	v, _ := cfg.(*config.RateSJsonCfg)
	return v, nil
}
func (*InternalDB) SIPAgentJsonCfg() (*config.SIPAgentJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.RPCConnsJSON)
	v, _ := cfg.(*config.SIPAgentJsonCfg)
	return v, nil
}
func (*InternalDB) TemplateSJsonCfg() (config.FcTemplatesJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.SIPAgentJSON)
	v, _ := cfg.(config.FcTemplatesJsonCfg)
	return v, nil
}
func (*InternalDB) ConfigSJsonCfg() (*config.ConfigSCfgJson, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.TemplatesJSON)
	v, _ := cfg.(*config.ConfigSCfgJson)
	return v, nil
}
func (*InternalDB) ApiBanCfgJson() (*config.APIBanJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.ConfigSJSON)
	v, _ := cfg.(*config.APIBanJsonCfg)
	return v, nil
}
func (*InternalDB) CoreSJSON() (*config.CoreSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.APIBanJSON)
	v, _ := cfg.(*config.CoreSJsonCfg)
	return v, nil
}
func (*InternalDB) ActionSCfgJson() (*config.ActionSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.CoreSJSON)
	v, _ := cfg.(*config.ActionSJsonCfg)
	return v, nil
}
func (*InternalDB) AccountSCfgJson() (*config.AccountSJsonCfg, error) {
	cfg, _ := Cache.Get(utils.MetaConfig, config.AccountSJSON)
	v, _ := cfg.(*config.AccountSJsonCfg)
	return v, nil
}

func (*InternalDB) SetSection(_ *context.Context, section string, jsn interface{}) error {
	Cache.SetWithoutReplicate(utils.MetaConfig, section, jsn, nil, true, utils.NonTransactional)
	return nil
}
