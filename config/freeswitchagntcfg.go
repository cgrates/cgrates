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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// FsConnCfg one connection to FreeSWITCH server
type FsConnCfg struct {
	Address    string
	Password   string
	Reconnects int
	Alias      string
}

func (fs *FsConnCfg) loadFromJSONCfg(jsnCfg *FsConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		fs.Address = *jsnCfg.Address
	}
	if jsnCfg.Password != nil {
		fs.Password = *jsnCfg.Password
	}
	if jsnCfg.Reconnects != nil {
		fs.Reconnects = *jsnCfg.Reconnects
	}
	fs.Alias = fs.Address
	if jsnCfg.Alias != nil && *jsnCfg.Alias != "" {
		fs.Alias = *jsnCfg.Alias
	}

	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (fs FsConnCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.AddressCfg:    fs.Address,
		utils.Password:      fs.Password,
		utils.ReconnectsCfg: fs.Reconnects,
		utils.AliasCfg:      fs.Alias,
	}
}

// Clone returns a deep copy of AsteriskAgentCfg
func (fs FsConnCfg) Clone() *FsConnCfg {
	return &FsConnCfg{
		Address:    fs.Address,
		Password:   fs.Password,
		Reconnects: fs.Reconnects,
		Alias:      fs.Alias,
	}
}

// FsAgentCfg the config section that describes the FreeSWITCH Agent
type FsAgentCfg struct {
	Enabled             bool
	SessionSConns       []string
	SubscribePark       bool
	CreateCdr           bool
	ExtraFields         RSRParsers
	LowBalanceAnnFile   string
	EmptyBalanceContext string
	EmptyBalanceAnnFile string
	MaxWaitConnection   time.Duration
	EventSocketConns    []*FsConnCfg
}

// loadFreeswitchAgentCfg loads the FreeswitchAgent section of the configuration
func (fscfg *FsAgentCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnSmFsCfg := new(FreeswitchAgentJsonCfg)
	if err = jsnCfg.GetSection(ctx, FreeSWITCHAgentJSON, jsnSmFsCfg); err != nil {
		return
	}
	return fscfg.loadFromJSONCfg(jsnSmFsCfg)
}

func (fscfg *FsAgentCfg) loadFromJSONCfg(jsnCfg *FreeswitchAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		fscfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		fscfg.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Subscribe_park != nil {
		fscfg.SubscribePark = *jsnCfg.Subscribe_park
	}
	if jsnCfg.Create_cdr != nil {
		fscfg.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Extra_fields != nil {
		if fscfg.ExtraFields, err = NewRSRParsersFromSlice(*jsnCfg.Extra_fields); err != nil {
			return err
		}
	}
	if jsnCfg.Low_balance_ann_file != nil {
		fscfg.LowBalanceAnnFile = *jsnCfg.Low_balance_ann_file
	}
	if jsnCfg.Empty_balance_context != nil {
		fscfg.EmptyBalanceContext = *jsnCfg.Empty_balance_context
	}

	if jsnCfg.Empty_balance_ann_file != nil {
		fscfg.EmptyBalanceAnnFile = *jsnCfg.Empty_balance_ann_file
	}
	if jsnCfg.Max_wait_connection != nil {
		if fscfg.MaxWaitConnection, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_wait_connection); err != nil {
			return err
		}
	}
	if jsnCfg.Event_socket_conns != nil {
		fscfg.EventSocketConns = make([]*FsConnCfg, len(*jsnCfg.Event_socket_conns))
		for idx, jsnConnCfg := range *jsnCfg.Event_socket_conns {
			fscfg.EventSocketConns[idx] = getDftFsConnCfg()
			fscfg.EventSocketConns[idx].loadFromJSONCfg(jsnConnCfg)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (fscfg FsAgentCfg) AsMapInterface(separator string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg:             fscfg.Enabled,
		utils.SubscribeParkCfg:       fscfg.SubscribePark,
		utils.CreateCdrCfg:           fscfg.CreateCdr,
		utils.LowBalanceAnnFileCfg:   fscfg.LowBalanceAnnFile,
		utils.EmptyBalanceContextCfg: fscfg.EmptyBalanceContext,
		utils.EmptyBalanceAnnFileCfg: fscfg.EmptyBalanceAnnFile,
		utils.MaxWaitConnectionCfg:   utils.EmptyString,
	}
	if fscfg.SessionSConns != nil {
		mp[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(fscfg.SessionSConns)
	}
	if fscfg.ExtraFields != nil {
		mp[utils.ExtraFieldsCfg] = fscfg.ExtraFields.AsStringSlice()
	}

	if fscfg.MaxWaitConnection != 0 {
		mp[utils.MaxWaitConnectionCfg] = fscfg.MaxWaitConnection.String()
	}
	if fscfg.EventSocketConns != nil {
		eventSocketConns := make([]map[string]interface{}, len(fscfg.EventSocketConns))
		for key, item := range fscfg.EventSocketConns {
			eventSocketConns[key] = item.AsMapInterface()
		}
		mp[utils.EventSocketConnsCfg] = eventSocketConns
	}
	return mp
}

func (FsAgentCfg) SName() string               { return FreeSWITCHAgentJSON }
func (fscfg FsAgentCfg) CloneSection() Section { return fscfg.Clone() }

// Clone returns a deep copy of FsAgentCfg
func (fscfg FsAgentCfg) Clone() (cln *FsAgentCfg) {
	cln = &FsAgentCfg{
		Enabled:             fscfg.Enabled,
		SubscribePark:       fscfg.SubscribePark,
		CreateCdr:           fscfg.CreateCdr,
		ExtraFields:         fscfg.ExtraFields.Clone(),
		LowBalanceAnnFile:   fscfg.LowBalanceAnnFile,
		EmptyBalanceContext: fscfg.EmptyBalanceContext,
		EmptyBalanceAnnFile: fscfg.EmptyBalanceAnnFile,
		MaxWaitConnection:   fscfg.MaxWaitConnection,
		SessionSConns:       utils.CloneStringSlice(fscfg.SessionSConns),
	}
	if fscfg.EventSocketConns != nil {
		cln.EventSocketConns = make([]*FsConnCfg, len(fscfg.EventSocketConns))
		for i, req := range fscfg.EventSocketConns {
			cln.EventSocketConns[i] = req.Clone()
		}
	}
	return
}

// FreeSWITCHAgent config section
type FreeswitchAgentJsonCfg struct {
	Enabled                *bool
	Sessions_conns         *[]string
	Subscribe_park         *bool
	Create_cdr             *bool
	Extra_fields           *[]string
	Low_balance_ann_file   *string
	Empty_balance_context  *string
	Empty_balance_ann_file *string
	Max_wait_connection    *string
	Event_socket_conns     *[]*FsConnJsonCfg
}

// Represents one connection instance towards FreeSWITCH
type FsConnJsonCfg struct {
	Address    *string
	Password   *string
	Reconnects *int
	Alias      *string
}

func diffFsConnJsonCfg(v1, v2 *FsConnCfg) (d *FsConnJsonCfg) {
	d = new(FsConnJsonCfg)
	if v1.Address != v2.Address {
		d.Address = utils.StringPointer(v2.Address)
	}
	if v1.Password != v2.Password {
		d.Password = utils.StringPointer(v2.Password)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.Reconnects)
	}
	if v1.Alias != v2.Alias {
		d.Alias = utils.StringPointer(v2.Alias)
	}
	return
}

func equalsFsConnsJsonCfg(v1, v2 []*FsConnCfg) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].Address != v2[i].Address ||
			v1[i].Password != v2[i].Password ||
			v1[i].Reconnects != v2[i].Reconnects ||
			v1[i].Alias != v2[i].Alias {
			return false
		}
	}
	return true
}

func diffFreeswitchAgentJsonCfg(d *FreeswitchAgentJsonCfg, v1, v2 *FsAgentCfg) *FreeswitchAgentJsonCfg {
	if d == nil {
		d = new(FreeswitchAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.SubscribePark != v2.SubscribePark {
		d.Subscribe_park = utils.BoolPointer(v2.SubscribePark)
	}
	if v1.CreateCdr != v2.CreateCdr {
		d.Create_cdr = utils.BoolPointer(v2.CreateCdr)
	}
	extra1 := v1.ExtraFields.AsStringSlice()
	extra2 := v2.ExtraFields.AsStringSlice()
	if !utils.SliceStringEqual(extra1, extra2) {
		d.Extra_fields = &extra2
	}
	if v1.LowBalanceAnnFile != v2.LowBalanceAnnFile {
		d.Low_balance_ann_file = utils.StringPointer(v2.LowBalanceAnnFile)
	}
	if v1.EmptyBalanceContext != v2.EmptyBalanceContext {
		d.Empty_balance_context = utils.StringPointer(v2.EmptyBalanceContext)
	}
	if v1.EmptyBalanceAnnFile != v2.EmptyBalanceAnnFile {
		d.Empty_balance_ann_file = utils.StringPointer(v2.EmptyBalanceAnnFile)
	}
	if v1.MaxWaitConnection != v2.MaxWaitConnection {
		d.Max_wait_connection = utils.StringPointer(v2.MaxWaitConnection.String())
	}

	if !equalsFsConnsJsonCfg(v1.EventSocketConns, v2.EventSocketConns) {
		v := make([]*FsConnJsonCfg, len(v2.EventSocketConns))
		dflt := getDftFsConnCfg()
		for i, val := range v2.EventSocketConns {
			v[i] = diffFsConnJsonCfg(dflt, val)
		}
		d.Event_socket_conns = &v
	}
	return d
}
