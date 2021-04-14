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

func (iDB *InternalDB) GeneralJsonCfg() (*config.GeneralJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) RPCConnJsonCfg() (config.RPCConnsJson, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) CacheJsonCfg() (*config.CacheJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ListenJsonCfg() (*config.ListenJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) HttpJsonCfg() (*config.HTTPJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) DbJsonCfg(section string) (*config.DbJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) FilterSJsonCfg() (*config.FilterSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) CdrsJsonCfg() (*config.CdrsJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ERsJsonCfg() (*config.ERsJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) EEsJsonCfg() (*config.EEsJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) SessionSJsonCfg() (*config.SessionSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) FreeswitchAgentJsonCfg() (*config.FreeswitchAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) KamAgentJsonCfg() (*config.KamAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) AsteriskAgentJsonCfg() (*config.AsteriskAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) DiameterAgentJsonCfg() (*config.DiameterAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) RadiusAgentJsonCfg() (*config.RadiusAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) HttpAgentJsonCfg() (*[]*config.HttpAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) DNSAgentJsonCfg() (*config.DNSAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) AttributeServJsonCfg() (*config.AttributeSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ChargerServJsonCfg() (*config.ChargerSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ResourceSJsonCfg() (*config.ResourceSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) StatSJsonCfg() (*config.StatServJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ThresholdSJsonCfg() (*config.ThresholdSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) RouteSJsonCfg() (*config.RouteSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) LoaderJsonCfg() ([]*config.LoaderJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) MailerJsonCfg() (*config.MailerJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) SureTaxJsonCfg() (*config.SureTaxJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) DispatcherSJsonCfg() (*config.DispatcherSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) RegistrarCJsonCfgs() (*config.RegistrarCJsonCfgs, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) LoaderCfgJson() (*config.LoaderCfgJson, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) MigratorCfgJson() (*config.MigratorCfgJson, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) TlsCfgJson() (*config.TlsJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) AnalyzerCfgJson() (*config.AnalyzerSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) AdminSCfgJson() (*config.AdminSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) RateCfgJson() (*config.RateSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) SIPAgentJsonCfg() (*config.SIPAgentJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) TemplateSJsonCfg() (config.FcTemplatesJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ConfigSJsonCfg() (*config.ConfigSCfgJson, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ApiBanCfgJson() (*config.APIBanJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) CoreSJSON() (*config.CoreSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) ActionSCfgJson() (*config.ActionSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}
func (iDB *InternalDB) AccountSCfgJson() (*config.AccountSJsonCfg, error) {
	return nil, utils.ErrNotImplemented
}

func (iDB *InternalDB) SetSection(_ *context.Context, section string, jsn interface{}) error {
	return utils.ErrNotImplemented

}
