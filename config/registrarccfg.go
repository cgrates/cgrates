// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// RegistrarCCfgs is the configuration of registrarc rpc
type RegistrarCCfgs struct {
	RPC *RegistrarCCfg
}

// loadRegistrarCCfg loads the RegistrarC section of the configuration
func (dps *RegistrarCCfgs) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnRegistrarCCfg := new(RegistrarCJsonCfgs)
	if err = jsnCfg.GetSection(ctx, RegistrarCJSON, jsnRegistrarCCfg); err != nil {
		return
	}
	return dps.loadFromJSONCfg(jsnRegistrarCCfg)
}

func (dps *RegistrarCCfgs) loadFromJSONCfg(jsnCfg *RegistrarCJsonCfgs) (err error) {
	if jsnCfg == nil {
		return nil
	}
	return dps.RPC.loadFromJSONCfg(jsnCfg.RPC)
}

// AsMapInterface returns the config as a map[string]any
func (dps RegistrarCCfgs) AsMapInterface() any {
	return map[string]any{
		utils.RPCCfg: dps.RPC.AsMapInterface(),
	}
}

func (RegistrarCCfgs) SName() string             { return RegistrarCJSON }
func (dps RegistrarCCfgs) CloneSection() Section { return dps.Clone() }

// Clone returns a deep copy of DispatcherHCfg
func (dps RegistrarCCfgs) Clone() (cln *RegistrarCCfgs) {
	return &RegistrarCCfgs{
		RPC: dps.RPC.Clone(),
	}
}

// RegistrarCCfg is the configuration of registrarc
type RegistrarCCfg struct {
	RegistrarSConns []string
	Hosts           map[string][]*RemoteHost
	RefreshInterval time.Duration
}

type RemoteHostJsonWithTenant struct {
	*RemoteHostJson
	Tenant *string
}

func (dps *RegistrarCCfg) loadFromJSONCfg(jsnCfg *RegistrarCJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Registrars_conns != nil {
		dps.RegistrarSConns = slices.Clone(*jsnCfg.Registrars_conns)
	}
	if jsnCfg.Hosts != nil {
		for _, hostJSON := range jsnCfg.Hosts {
			conn := new(RemoteHost)
			conn.loadFromJSONCfg(hostJSON.RemoteHostJson)
			if hostJSON.Tenant == nil || *hostJSON.Tenant == "" {
				dps.Hosts[utils.MetaDefault] = append(dps.Hosts[utils.MetaDefault], conn)
			} else {
				dps.Hosts[*hostJSON.Tenant] = append(dps.Hosts[*hostJSON.Tenant], conn)
			}
		}
	}
	if jsnCfg.Refresh_interval != nil {
		if dps.RefreshInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Refresh_interval); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (dps *RegistrarCCfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.RegistrarsConnsCfg: slices.Clone(dps.RegistrarSConns),
		utils.RefreshIntervalCfg: dps.RefreshInterval.String(),
	}
	if dps.RefreshInterval == 0 {
		initialMP[utils.RefreshIntervalCfg] = "0"
	}
	if dps.Hosts != nil {
		hosts := []map[string]any{}
		for tnt, hs := range dps.Hosts {
			for _, h := range hs {
				mp := h.AsMapInterface()
				delete(mp, utils.AddressCfg)
				mp[utils.Tenant] = tnt
				hosts = append(hosts, mp)
			}
		}
		initialMP[utils.HostsCfg] = hosts
	}
	return
}

// Clone returns a deep copy of DispatcherHCfg
func (dps RegistrarCCfg) Clone() (cln *RegistrarCCfg) {
	cln = &RegistrarCCfg{
		RefreshInterval: dps.RefreshInterval,
		Hosts:           make(map[string][]*RemoteHost),
	}
	if dps.RegistrarSConns != nil {
		cln.RegistrarSConns = slices.Clone(dps.RegistrarSConns)
	}
	for tnt, hosts := range dps.Hosts {
		clnH := make([]*RemoteHost, len(hosts))
		for i, host := range hosts {
			clnH[i] = host.Clone()
		}
		cln.Hosts[tnt] = clnH
	}
	return
}

type RegistrarCJsonCfg struct {
	Registrars_conns *[]string `json:"registrarsConns"`
	Hosts            []*RemoteHostJsonWithTenant
	Refresh_interval *string `json:"refreshInterval"`
}

func diffRegistrarCJsonCfg(d *RegistrarCJsonCfg, v1, v2 *RegistrarCCfg) *RegistrarCJsonCfg {
	if d == nil {
		d = new(RegistrarCJsonCfg)
	}
	if !slices.Equal(v1.RegistrarSConns, v2.RegistrarSConns) {
		d.Registrars_conns = utils.SliceStringPointer(slices.Clone(v2.RegistrarSConns))
	}
	if d.Hosts == nil {
		d.Hosts = []*RemoteHostJsonWithTenant{}
	}
	for k, host := range v2.Hosts {
		for _, conn := range host {
			dConn := &RemoteHostJsonWithTenant{
				RemoteHostJson: new(RemoteHostJson),
			}
			if conn.ID != utils.EmptyString {
				dConn.Id = utils.StringPointer(conn.ID)
			}
			if conn.Transport != utils.EmptyString {
				dConn.Transport = utils.StringPointer(conn.Transport)
			}
			if conn.Address != utils.EmptyString {
				dConn.Address = utils.StringPointer(conn.Address)
			}
			if conn.TLS {
				dConn.Tls = utils.BoolPointer(conn.TLS)
			}
			if k != utils.MetaDefault {
				dConn.Tenant = utils.StringPointer(k)
			}
			d.Hosts = append(d.Hosts, dConn)

		}
	}
	if v1.RefreshInterval != v2.RefreshInterval {
		d.Refresh_interval = utils.StringPointer(v2.RefreshInterval.String())
	}
	return d
}

type RegistrarCJsonCfgs struct {
	RPC *RegistrarCJsonCfg
}

func diffRegistrarCJsonCfgs(d *RegistrarCJsonCfgs, v1, v2 *RegistrarCCfgs) *RegistrarCJsonCfgs {
	if d == nil {
		d = new(RegistrarCJsonCfgs)
	}
	d.RPC = diffRegistrarCJsonCfg(d.RPC, v1.RPC, v2.RPC)
	return d
}
