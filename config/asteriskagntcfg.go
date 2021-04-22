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
	"github.com/cgrates/cgrates/utils"
)

// NewDefaultAsteriskConnCfg is uses stored defaults so we can pre-populate by loading from JSON config
func NewDefaultAsteriskConnCfg() *AsteriskConnCfg {
	if dfltAstConnCfg == nil {
		return new(AsteriskConnCfg) // No defaults, most probably we are building the defaults now
	}
	return dfltAstConnCfg.Clone()
}

// AsteriskConnCfg the config for a Asterisk connection
type AsteriskConnCfg struct {
	Alias           string
	Address         string
	User            string
	Password        string
	ConnectAttempts int
	Reconnects      int
}

func (aConnCfg *AsteriskConnCfg) loadFromJSONCfg(jsnCfg *AstConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		aConnCfg.Address = *jsnCfg.Address
	}
	if jsnCfg.Alias != nil {
		aConnCfg.Alias = *jsnCfg.Alias
	}
	if jsnCfg.User != nil {
		aConnCfg.User = *jsnCfg.User
	}
	if jsnCfg.Password != nil {
		aConnCfg.Password = *jsnCfg.Password
	}
	if jsnCfg.Connect_attempts != nil {
		aConnCfg.ConnectAttempts = *jsnCfg.Connect_attempts
	}
	if jsnCfg.Reconnects != nil {
		aConnCfg.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (aConnCfg *AsteriskConnCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.AliasCfg:           aConnCfg.Alias,
		utils.AddressCfg:         aConnCfg.Address,
		utils.UserCf:             aConnCfg.User,
		utils.Password:           aConnCfg.Password,
		utils.ConnectAttemptsCfg: aConnCfg.ConnectAttempts,
		utils.ReconnectsCfg:      aConnCfg.Reconnects,
	}
}

// Clone returns a deep copy of AsteriskConnCfg
func (aConnCfg AsteriskConnCfg) Clone() *AsteriskConnCfg {
	return &AsteriskConnCfg{
		Alias:           aConnCfg.Alias,
		Address:         aConnCfg.Address,
		User:            aConnCfg.User,
		Password:        aConnCfg.Password,
		ConnectAttempts: aConnCfg.ConnectAttempts,
		Reconnects:      aConnCfg.Reconnects,
	}
}

// AsteriskAgentCfg the config section that describes the Asterisk Agent
type AsteriskAgentCfg struct {
	Enabled       bool
	SessionSConns []string
	CreateCDR     bool
	AsteriskConns []*AsteriskConnCfg
}

func (aCfg *AsteriskAgentCfg) loadFromJSONCfg(jsnCfg *AsteriskAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		aCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		aCfg.SessionSConns = updateBiRPCInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	if jsnCfg.Create_cdr != nil {
		aCfg.CreateCDR = *jsnCfg.Create_cdr
	}

	if jsnCfg.Asterisk_conns != nil {
		aCfg.AsteriskConns = make([]*AsteriskConnCfg, len(*jsnCfg.Asterisk_conns))
		for i, jsnAConn := range *jsnCfg.Asterisk_conns {
			aCfg.AsteriskConns[i] = NewDefaultAsteriskConnCfg()
			aCfg.AsteriskConns[i].loadFromJSONCfg(jsnAConn)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (aCfg *AsteriskAgentCfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:   aCfg.Enabled,
		utils.CreateCDRCfg: aCfg.CreateCDR,
	}
	if aCfg.AsteriskConns != nil {
		conns := make([]map[string]interface{}, len(aCfg.AsteriskConns))
		for i, item := range aCfg.AsteriskConns {
			conns[i] = item.AsMapInterface()
		}
		initialMP[utils.AsteriskConnsCfg] = conns
	}
	if aCfg.SessionSConns != nil {
		initialMP[utils.SessionSConnsCfg] = getBiRPCInternalJSONConns(aCfg.SessionSConns)
	}
	return
}

// Clone returns a deep copy of AsteriskAgentCfg
func (aCfg AsteriskAgentCfg) Clone() (cln *AsteriskAgentCfg) {
	cln = &AsteriskAgentCfg{
		Enabled:   aCfg.Enabled,
		CreateCDR: aCfg.CreateCDR,
	}
	if aCfg.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(aCfg.SessionSConns)
	}
	if aCfg.AsteriskConns != nil {
		cln.AsteriskConns = make([]*AsteriskConnCfg, len(aCfg.AsteriskConns))
		for i, req := range aCfg.AsteriskConns {
			cln.AsteriskConns[i] = req.Clone()
		}
	}
	return
}

type AstConnJsonCfg struct {
	Alias            *string
	Address          *string
	User             *string
	Password         *string
	Connect_attempts *int
	Reconnects       *int
}

type AsteriskAgentJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Create_cdr     *bool
	Asterisk_conns *[]*AstConnJsonCfg
}

func diffAstConnJsonCfg(v1, v2 *AsteriskConnCfg) (d *AstConnJsonCfg) {
	d = new(AstConnJsonCfg)
	if v1.Alias != v2.Alias {
		d.Alias = utils.StringPointer(v2.Alias)
	}
	if v1.Address != v2.Address {
		d.Address = utils.StringPointer(v2.Address)
	}
	if v1.User != v2.User {
		d.User = utils.StringPointer(v2.User)
	}
	if v1.Password != v2.Password {
		d.Password = utils.StringPointer(v2.Password)
	}
	if v1.ConnectAttempts != v2.ConnectAttempts {
		d.Connect_attempts = utils.IntPointer(v2.ConnectAttempts)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.Reconnects)
	}
	return
}

func equalsAstConnJsonCfg(v1, v2 []*AsteriskConnCfg) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].Alias != v2[i].Alias ||
			v1[i].Address != v2[i].Address ||
			v1[i].User != v2[i].User ||
			v1[i].Password != v2[i].Password ||
			v1[i].ConnectAttempts != v2[i].ConnectAttempts ||
			v1[i].Reconnects != v2[i].Reconnects {
			return false
		}
	}
	return true
}

func diffAsteriskAgentJsonCfg(d *AsteriskAgentJsonCfg, v1, v2 *AsteriskAgentCfg) *AsteriskAgentJsonCfg {
	if d == nil {
		d = new(AsteriskAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getBiRPCInternalJSONConns(v2.SessionSConns))
	}
	if v1.CreateCDR != v2.CreateCDR {
		d.Create_cdr = utils.BoolPointer(v2.CreateCDR)
	}

	if !equalsAstConnJsonCfg(v1.AsteriskConns, v2.AsteriskConns) {
		v := make([]*AstConnJsonCfg, len(v2.AsteriskConns))
		dflt := NewDefaultAsteriskConnCfg()
		for i, val := range v2.AsteriskConns {
			v[i] = diffAstConnJsonCfg(dflt, val)
		}
		d.Asterisk_conns = &v
	}
	return d
}
