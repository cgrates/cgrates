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

	"github.com/cgrates/cgrates/utils"
)

// Returns the first cached default value for a FreeSWITCHAgent connection
func NewDfltHaPoolConfig() *HaPoolConfig {
	if dfltHaPoolConfig == nil {
		return new(HaPoolConfig) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltHaPoolConfig // Copy the value instead of it's pointer
	return &dfltVal
}

// One connection to Rater
type HaPoolConfig struct {
	Address     string
	Transport   string
	Synchronous bool
	Tls         bool
}

func (self *HaPoolConfig) loadFromJsonCfg(jsnCfg *HaPoolJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Transport != nil {
		self.Transport = *jsnCfg.Transport
	}
	if jsnCfg.Synchronous != nil {
		self.Synchronous = *jsnCfg.Synchronous
	}
	if jsnCfg.Tls != nil {
		self.Tls = *jsnCfg.Synchronous
	}
	return nil
}

// Returns the first cached default value for a FreeSWITCHAgent connection
func NewDfltFsConnConfig() *FsConnCfg {
	if dfltFsConnConfig == nil {
		return new(FsConnCfg) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltFsConnConfig // Copy the value instead of it's pointer
	return &dfltVal
}

// One connection to FreeSWITCH server
type FsConnCfg struct {
	Address    string
	Password   string
	Reconnects int
	Alias      string
}

func (self *FsConnCfg) loadFromJsonCfg(jsnCfg *FsConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		self.Address = *jsnCfg.Address
	}
	if jsnCfg.Password != nil {
		self.Password = *jsnCfg.Password
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	self.Alias = self.Address
	if jsnCfg.Alias != nil && *jsnCfg.Alias != "" {
		self.Alias = *jsnCfg.Alias
	}

	return nil
}

type SessionSCfg struct {
	Enabled                 bool
	ListenBijson            string
	ChargerSConns           []*HaPoolConfig
	RALsConns               []*HaPoolConfig
	ResSConns               []*HaPoolConfig
	ThreshSConns            []*HaPoolConfig
	StatSConns              []*HaPoolConfig
	SupplSConns             []*HaPoolConfig
	AttrSConns              []*HaPoolConfig
	CDRsConns               []*HaPoolConfig
	SessionReplicationConns []*HaPoolConfig
	DebitInterval           time.Duration
	MinCallDuration         time.Duration
	MaxCallDuration         time.Duration
	SessionTTL              time.Duration
	SessionTTLMaxDelay      *time.Duration
	SessionTTLLastUsed      *time.Duration
	SessionTTLUsage         *time.Duration
	SessionIndexes          utils.StringMap
	ClientProtocol          float64
	ChannelSyncInterval     time.Duration
}

func (self *SessionSCfg) loadFromJsonCfg(jsnCfg *SessionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_bijson != nil {
		self.ListenBijson = *jsnCfg.Listen_bijson
	}
	if jsnCfg.Chargers_conns != nil {
		self.ChargerSConns = make([]*HaPoolConfig, len(*jsnCfg.Chargers_conns))
		for idx, jsnHaCfg := range *jsnCfg.Chargers_conns {
			self.ChargerSConns[idx] = NewDfltHaPoolConfig()
			self.ChargerSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Rals_conns != nil {
		self.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			self.RALsConns[idx] = NewDfltHaPoolConfig()
			self.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Resources_conns != nil {
		self.ResSConns = make([]*HaPoolConfig, len(*jsnCfg.Resources_conns))
		for idx, jsnHaCfg := range *jsnCfg.Resources_conns {
			self.ResSConns[idx] = NewDfltHaPoolConfig()
			self.ResSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		self.ThreshSConns = make([]*HaPoolConfig, len(*jsnCfg.Thresholds_conns))
		for idx, jsnHaCfg := range *jsnCfg.Thresholds_conns {
			self.ThreshSConns[idx] = NewDfltHaPoolConfig()
			self.ThreshSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Stats_conns != nil {
		self.StatSConns = make([]*HaPoolConfig, len(*jsnCfg.Stats_conns))
		for idx, jsnHaCfg := range *jsnCfg.Stats_conns {
			self.StatSConns[idx] = NewDfltHaPoolConfig()
			self.StatSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Suppliers_conns != nil {
		self.SupplSConns = make([]*HaPoolConfig, len(*jsnCfg.Suppliers_conns))
		for idx, jsnHaCfg := range *jsnCfg.Suppliers_conns {
			self.SupplSConns[idx] = NewDfltHaPoolConfig()
			self.SupplSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Attributes_conns != nil {
		self.AttrSConns = make([]*HaPoolConfig, len(*jsnCfg.Attributes_conns))
		for idx, jsnHaCfg := range *jsnCfg.Attributes_conns {
			self.AttrSConns[idx] = NewDfltHaPoolConfig()
			self.AttrSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CDRsConns[idx] = NewDfltHaPoolConfig()
			self.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Session_replication_conns != nil {
		self.SessionReplicationConns = make([]*HaPoolConfig, len(*jsnCfg.Session_replication_conns))
		for idx, jsnHaCfg := range *jsnCfg.Session_replication_conns {
			self.SessionReplicationConns[idx] = NewDfltHaPoolConfig()
			self.SessionReplicationConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl != nil {
		if self.SessionTTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl_max_delay != nil {
		if maxTTLDelay, err := utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_max_delay); err != nil {
			return err
		} else {
			self.SessionTTLMaxDelay = &maxTTLDelay
		}
	}
	if jsnCfg.Session_ttl_last_used != nil {
		if sessionTTLLastUsed, err := utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_last_used); err != nil {
			return err
		} else {
			self.SessionTTLLastUsed = &sessionTTLLastUsed
		}
	}
	if jsnCfg.Session_indexes != nil {
		self.SessionIndexes = utils.StringMapFromSlice(*jsnCfg.Session_indexes)
	}
	if jsnCfg.Client_protocol != nil {
		self.ClientProtocol = *jsnCfg.Client_protocol
	}
	if jsnCfg.Channel_sync_interval != nil {
		if self.ChannelSyncInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Channel_sync_interval); err != nil {
			return err
		}
	}
	return nil
}

type FsAgentCfg struct {
	Enabled       bool
	SessionSConns []*HaPoolConfig
	SubscribePark bool
	CreateCdr     bool
	ExtraFields   RSRParsers
	//MinDurLowBalance    time.Duration
	//LowBalanceAnnFile   string
	EmptyBalanceContext string
	EmptyBalanceAnnFile string
	MaxWaitConnection   time.Duration
	EventSocketConns    []*FsConnCfg
}

func (self *FsAgentCfg) loadFromJsonCfg(jsnCfg *FreeswitchAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		self.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			self.SessionSConns[idx] = NewDfltHaPoolConfig()
			self.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Subscribe_park != nil {
		self.SubscribePark = *jsnCfg.Subscribe_park
	}
	if jsnCfg.Create_cdr != nil {
		self.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Extra_fields != nil {
		if self.ExtraFields, err = NewRSRParsersFromSlice(*jsnCfg.Extra_fields, true); err != nil {
			return err
		}
	}
	if jsnCfg.Empty_balance_context != nil {
		self.EmptyBalanceContext = *jsnCfg.Empty_balance_context
	}

	if jsnCfg.Empty_balance_ann_file != nil {
		self.EmptyBalanceAnnFile = *jsnCfg.Empty_balance_ann_file
	}
	if jsnCfg.Max_wait_connection != nil {
		if self.MaxWaitConnection, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_wait_connection); err != nil {
			return err
		}
	}
	if jsnCfg.Event_socket_conns != nil {
		self.EventSocketConns = make([]*FsConnCfg, len(*jsnCfg.Event_socket_conns))
		for idx, jsnConnCfg := range *jsnCfg.Event_socket_conns {
			self.EventSocketConns[idx] = NewDfltFsConnConfig()
			self.EventSocketConns[idx].loadFromJsonCfg(jsnConnCfg)
		}
	}
	return nil
}

// Returns the first cached default value for a FreeSWITCHAgent connection
func NewDfltKamConnConfig() *KamConnCfg {
	if dfltKamConnConfig == nil {
		return new(KamConnCfg) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltKamConnConfig
	return &dfltVal
}

// Represents one connection instance towards OpenSIPS, not in use for now but planned for future
type OsipsConnConfig struct {
	MiAddr     string
	Reconnects int
}

func (self *OsipsConnConfig) loadFromJsonCfg(jsnCfg *OsipsConnJsonCfg) error {
	if jsnCfg.Mi_addr != nil {
		self.MiAddr = *jsnCfg.Mi_addr
	}
	if jsnCfg.Reconnects != nil {
		self.Reconnects = *jsnCfg.Reconnects
	}
	return nil
}

// SM-OpenSIPS config section
type SmOsipsConfig struct {
	Enabled                 bool
	ListenUdp               string
	RALsConns               []*HaPoolConfig
	CDRsConns               []*HaPoolConfig
	CreateCdr               bool
	DebitInterval           time.Duration
	MinCallDuration         time.Duration
	MaxCallDuration         time.Duration
	EventsSubscribeInterval time.Duration
	MiAddr                  string
}

func (self *SmOsipsConfig) loadFromJsonCfg(jsnCfg *SmOsipsJsonCfg) error {
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_udp != nil {
		self.ListenUdp = *jsnCfg.Listen_udp
	}
	if jsnCfg.Rals_conns != nil {
		self.RALsConns = make([]*HaPoolConfig, len(*jsnCfg.Rals_conns))
		for idx, jsnHaCfg := range *jsnCfg.Rals_conns {
			self.RALsConns[idx] = NewDfltHaPoolConfig()
			self.RALsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CDRsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CDRsConns[idx] = NewDfltHaPoolConfig()
			self.CDRsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Create_cdr != nil {
		self.CreateCdr = *jsnCfg.Create_cdr
	}
	if jsnCfg.Debit_interval != nil {
		if self.DebitInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Min_call_duration != nil {
		if self.MinCallDuration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if self.MaxCallDuration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Events_subscribe_interval != nil {
		if self.EventsSubscribeInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Events_subscribe_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Mi_addr != nil {
		self.MiAddr = *jsnCfg.Mi_addr
	}

	return nil
}

// Uses stored defaults so we can pre-populate by loading from JSON config
func NewDefaultAsteriskConnCfg() *AsteriskConnCfg {
	if dfltAstConnCfg == nil {
		return new(AsteriskConnCfg) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltAstConnCfg // Copy the value instead of it's pointer
	return &dfltVal
}

type AsteriskConnCfg struct {
	Address         string
	User            string
	Password        string
	ConnectAttempts int
	Reconnects      int
}

func (aConnCfg *AsteriskConnCfg) loadFromJsonCfg(jsnCfg *AstConnJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Address != nil {
		aConnCfg.Address = *jsnCfg.Address
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

type AsteriskAgentCfg struct {
	Enabled       bool
	SessionSConns []*HaPoolConfig
	CreateCDR     bool
	AsteriskConns []*AsteriskConnCfg
}

func (aCfg *AsteriskAgentCfg) loadFromJsonCfg(jsnCfg *AsteriskAgentJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		aCfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		aCfg.SessionSConns = make([]*HaPoolConfig, len(*jsnCfg.Sessions_conns))
		for idx, jsnHaCfg := range *jsnCfg.Sessions_conns {
			aCfg.SessionSConns[idx] = NewDfltHaPoolConfig()
			aCfg.SessionSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Create_cdr != nil {
		aCfg.CreateCDR = *jsnCfg.Create_cdr
	}
	if jsnCfg.Asterisk_conns != nil {
		aCfg.AsteriskConns = make([]*AsteriskConnCfg, len(*jsnCfg.Asterisk_conns))
		for i, jsnAConn := range *jsnCfg.Asterisk_conns {
			aCfg.AsteriskConns[i] = NewDefaultAsteriskConnCfg()
			aCfg.AsteriskConns[i].loadFromJsonCfg(jsnAConn)
		}
	}
	return nil
}
