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
)

const (
	configPrefix = "cfg_"
)

func (rs *RedisStorage) GeneralJsonCfg() (r *config.GeneralJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.GeneralJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) RPCConnJsonCfg() (r config.RPCConnsJson, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.RPCConnsJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) CacheJsonCfg() (r *config.CacheJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.CacheJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ListenJsonCfg() (r *config.ListenJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ListenJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) HttpJsonCfg() (r *config.HTTPJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.HTTPJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) DbJsonCfg(section string) (r *config.DbJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+section); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) FilterSJsonCfg() (r *config.FilterSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.FilterSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) CdrsJsonCfg() (r *config.CdrsJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.CDRsJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ERsJsonCfg() (r *config.ERsJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ERsJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) EEsJsonCfg() (r *config.EEsJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.EEsJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) SessionSJsonCfg() (r *config.SessionSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.SessionSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) FreeswitchAgentJsonCfg() (r *config.FreeswitchAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.FreeSWITCHAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) KamAgentJsonCfg() (r *config.KamAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.KamailioAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) AsteriskAgentJsonCfg() (r *config.AsteriskAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.AsteriskAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) DiameterAgentJsonCfg() (r *config.DiameterAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.DiameterAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) RadiusAgentJsonCfg() (r *config.RadiusAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.RadiusAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) HttpAgentJsonCfg() (r *[]*config.HttpAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.HTTPAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) DNSAgentJsonCfg() (r *config.DNSAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.DNSAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) AttributeServJsonCfg() (r *config.AttributeSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.AttributeSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ChargerServJsonCfg() (r *config.ChargerSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ChargerSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ResourceSJsonCfg() (r *config.ResourceSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ResourceSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) StatSJsonCfg() (r *config.StatServJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.StatSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ThresholdSJsonCfg() (r *config.ThresholdSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ThresholdSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) RouteSJsonCfg() (r *config.RouteSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.RouteSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) LoaderJsonCfg() (r []*config.LoaderJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.LoaderSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) MailerJsonCfg() (r *config.MailerJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.MailerJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) SureTaxJsonCfg() (r *config.SureTaxJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.SureTaxJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) DispatcherSJsonCfg() (r *config.DispatcherSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.DispatcherSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) RegistrarCJsonCfgs() (r *config.RegistrarCJsonCfgs, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.RegistrarCJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) LoaderCfgJson() (r *config.LoaderCfgJson, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.LoaderJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) MigratorCfgJson() (r *config.MigratorCfgJson, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.MigratorJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) TlsCfgJson() (r *config.TlsJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.TlsJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) AnalyzerCfgJson() (r *config.AnalyzerSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.AnalyzerSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) AdminSCfgJson() (r *config.AdminSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.AdminSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) RateCfgJson() (r *config.RateSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.RateSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) SIPAgentJsonCfg() (r *config.SIPAgentJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.SIPAgentJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) TemplateSJsonCfg() (r config.FcTemplatesJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.TemplatesJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ConfigSJsonCfg() (r *config.ConfigSCfgJson, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ConfigSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ApiBanCfgJson() (r *config.APIBanJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.APIBanJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) CoreSJSON() (r *config.CoreSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.CoreSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) ActionSCfgJson() (r *config.ActionSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.ActionSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}
func (rs *RedisStorage) AccountSCfgJson() (r *config.AccountSJsonCfg, err error) {
	var values []byte
	if err = rs.Cmd(&values, redisGET, configPrefix+config.AccountSJSON); err != nil {
		return
	} else if len(values) == 0 {
		return
	}
	err = rs.ms.Unmarshal(values, &r)
	return
}

func (rs *RedisStorage) SetSection(_ *context.Context, section string, jsn interface{}) (err error) {
	var result []byte
	if result, err = rs.ms.Marshal(jsn); err != nil {
		return
	}
	return rs.Cmd(nil, redisSET, configPrefix+section, string(result))
}
