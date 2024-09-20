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
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// FsConnCfg one connection to FreeSWITCH server
type FsConnCfg struct {
	Address              string
	Password             string
	Reconnects           int
	MaxReconnectInterval time.Duration
	Alias                string
}

func (fs *FsConnCfg) loadFromJSONCfg(jsnCfg *FsConnJsonCfg) (err error) {
	if jsnCfg == nil {
		return
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
	if jsnCfg.Max_reconnect_interval != nil {
		if fs.MaxReconnectInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_reconnect_interval); err != nil {
			return
		}
	}
	fs.Alias = fs.Address
	if jsnCfg.Alias != nil && *jsnCfg.Alias != "" {
		fs.Alias = *jsnCfg.Alias
	}

	return
}

// AsMapInterface returns the config as a map[string]any
func (fs FsConnCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.AddressCfg:              fs.Address,
		utils.Password:                fs.Password,
		utils.ReconnectsCfg:           fs.Reconnects,
		utils.MaxReconnectIntervalCfg: fs.MaxReconnectInterval.String(),
		utils.AliasCfg:                fs.Alias,
	}
}

// Clone returns a deep copy of AsteriskAgentCfg
func (fs FsConnCfg) Clone() *FsConnCfg {
	return &FsConnCfg{
		Address:              fs.Address,
		Password:             fs.Password,
		Reconnects:           fs.Reconnects,
		MaxReconnectInterval: fs.MaxReconnectInterval,
		Alias:                fs.Alias,
	}
}

// FsAgentCfg the config section that describes the FreeSWITCH Agent
type FsAgentCfg struct {
	Enabled                bool
	SessionSConns          []string
	SubscribePark          bool
	CreateCDR              bool
	ExtraFields            RSRParsers
	LowBalanceAnnFile      string
	EmptyBalanceContext    string
	EmptyBalanceAnnFile    string
	MaxWaitConnection      time.Duration
	ActiveSessionDelimiter string
	EventSocketConns       []*FsConnCfg
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
	if jsnCfg.SessionSConns != nil {
		fscfg.SessionSConns = updateBiRPCInternalConns(*jsnCfg.SessionSConns, utils.MetaSessionS)
	}
	if jsnCfg.SubscribePark != nil {
		fscfg.SubscribePark = *jsnCfg.SubscribePark
	}
	if jsnCfg.CreateCDR != nil {
		fscfg.CreateCDR = *jsnCfg.CreateCDR
	}
	if jsnCfg.ExtraFields != nil {
		if fscfg.ExtraFields, err = NewRSRParsersFromSlice(*jsnCfg.ExtraFields); err != nil {
			return err
		}
	}
	if jsnCfg.LowBalanceAnnFile != nil {
		fscfg.LowBalanceAnnFile = *jsnCfg.LowBalanceAnnFile
	}
	if jsnCfg.EmptyBalanceContext != nil {
		fscfg.EmptyBalanceContext = *jsnCfg.EmptyBalanceContext
	}

	if jsnCfg.EmptyBalanceAnnFile != nil {
		fscfg.EmptyBalanceAnnFile = *jsnCfg.EmptyBalanceAnnFile
	}
	if jsnCfg.MaxWaitConnection != nil {
		if fscfg.MaxWaitConnection, err = utils.ParseDurationWithNanosecs(*jsnCfg.MaxWaitConnection); err != nil {
			return err
		}
	}
	if jsnCfg.ActiveSessionDelimiter != nil {
		fscfg.ActiveSessionDelimiter = *jsnCfg.ActiveSessionDelimiter
	}
	if jsnCfg.EventSocketConns != nil {
		fscfg.EventSocketConns = make([]*FsConnCfg, len(*jsnCfg.EventSocketConns))
		for idx, jsnConnCfg := range *jsnCfg.EventSocketConns {
			fscfg.EventSocketConns[idx] = getDftFsConnCfg()
			fscfg.EventSocketConns[idx].loadFromJSONCfg(jsnConnCfg)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]any
func (fscfg FsAgentCfg) AsMapInterface(separator string) any {
	mp := map[string]any{
		utils.EnabledCfg:                fscfg.Enabled,
		utils.SubscribeParkCfg:          fscfg.SubscribePark,
		utils.CreateCdrCfg:              fscfg.CreateCDR,
		utils.LowBalanceAnnFileCfg:      fscfg.LowBalanceAnnFile,
		utils.EmptyBalanceContextCfg:    fscfg.EmptyBalanceContext,
		utils.EmptyBalanceAnnFileCfg:    fscfg.EmptyBalanceAnnFile,
		utils.MaxWaitConnectionCfg:      utils.EmptyString,
		utils.ActiveSessionDelimiterCfg: fscfg.ActiveSessionDelimiter,
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
		eventSocketConns := make([]map[string]any, len(fscfg.EventSocketConns))
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
		Enabled:                fscfg.Enabled,
		SubscribePark:          fscfg.SubscribePark,
		CreateCDR:              fscfg.CreateCDR,
		ExtraFields:            fscfg.ExtraFields.Clone(),
		LowBalanceAnnFile:      fscfg.LowBalanceAnnFile,
		EmptyBalanceContext:    fscfg.EmptyBalanceContext,
		EmptyBalanceAnnFile:    fscfg.EmptyBalanceAnnFile,
		MaxWaitConnection:      fscfg.MaxWaitConnection,
		ActiveSessionDelimiter: fscfg.ActiveSessionDelimiter,
		SessionSConns:          slices.Clone(fscfg.SessionSConns),
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
	Enabled                *bool             `json:"enabled"`
	SessionSConns          *[]string         `json:"sessions_conns"`
	SubscribePark          *bool             `json:"subscribe_park"`
	CreateCDR              *bool             `json:"create_cdr"`
	ExtraFields            *[]string         `json:"extra_fields"`
	LowBalanceAnnFile      *string           `json:"low_balance_ann_file"`
	EmptyBalanceContext    *string           `json:"empty_balance_context"`
	EmptyBalanceAnnFile    *string           `json:"empty_balance_ann_file"`
	MaxWaitConnection      *string           `json:"max_wait_connection"`
	ActiveSessionDelimiter *string           `json:"active_session_delimiter"`
	EventSocketConns       *[]*FsConnJsonCfg `json:"event_socket_conns"`
}

// Represents one connection instance towards FreeSWITCH
type FsConnJsonCfg struct {
	Address                *string
	Password               *string
	Reconnects             *int
	Max_reconnect_interval *string
	Alias                  *string
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
	if v1.MaxReconnectInterval != v2.MaxReconnectInterval {
		d.Max_reconnect_interval = utils.StringPointer(v2.MaxReconnectInterval.String())
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
			v1[i].MaxReconnectInterval != v2[i].MaxReconnectInterval ||
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
	if !slices.Equal(v1.SessionSConns, v2.SessionSConns) {
		d.SessionSConns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.SubscribePark != v2.SubscribePark {
		d.SubscribePark = utils.BoolPointer(v2.SubscribePark)
	}
	if v1.CreateCDR != v2.CreateCDR {
		d.CreateCDR = utils.BoolPointer(v2.CreateCDR)
	}
	extra1 := v1.ExtraFields.AsStringSlice()
	extra2 := v2.ExtraFields.AsStringSlice()
	if !slices.Equal(extra1, extra2) {
		d.ExtraFields = &extra2
	}
	if v1.LowBalanceAnnFile != v2.LowBalanceAnnFile {
		d.LowBalanceAnnFile = utils.StringPointer(v2.LowBalanceAnnFile)
	}
	if v1.EmptyBalanceContext != v2.EmptyBalanceContext {
		d.EmptyBalanceContext = utils.StringPointer(v2.EmptyBalanceContext)
	}
	if v1.EmptyBalanceAnnFile != v2.EmptyBalanceAnnFile {
		d.EmptyBalanceAnnFile = utils.StringPointer(v2.EmptyBalanceAnnFile)
	}
	if v1.MaxWaitConnection != v2.MaxWaitConnection {
		d.MaxWaitConnection = utils.StringPointer(v2.MaxWaitConnection.String())
	}
	if v1.ActiveSessionDelimiter != v2.ActiveSessionDelimiter {
		d.ActiveSessionDelimiter = utils.StringPointer(v2.ActiveSessionDelimiter)
	}

	if !equalsFsConnsJsonCfg(v1.EventSocketConns, v2.EventSocketConns) {
		v := make([]*FsConnJsonCfg, len(v2.EventSocketConns))
		dflt := getDftFsConnCfg()
		for i, val := range v2.EventSocketConns {
			v[i] = diffFsConnJsonCfg(dflt, val)
		}
		d.EventSocketConns = &v
	}
	return d
}
