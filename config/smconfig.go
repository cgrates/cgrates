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
	Enabled             bool
	ListenBijson        string
	ChargerSConns       []string
	RALsConns           []string
	ResSConns           []string
	ThreshSConns        []string
	StatSConns          []string
	SupplSConns         []string
	AttrSConns          []string
	CDRsConns           []string
	ReplicationConns    []*RemoteHost
	DebitInterval       time.Duration
	StoreSCosts         bool
	MinCallDuration     time.Duration
	MaxCallDuration     time.Duration
	SessionTTL          time.Duration
	SessionTTLMaxDelay  *time.Duration
	SessionTTLLastUsed  *time.Duration
	SessionTTLUsage     *time.Duration
	SessionIndexes      utils.StringMap
	ClientProtocol      float64
	ChannelSyncInterval time.Duration
	TerminateAttempts   int
}

func (scfg *SessionSCfg) loadFromJsonCfg(jsnCfg *SessionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		scfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_bijson != nil {
		scfg.ListenBijson = *jsnCfg.Listen_bijson
	}
	if jsnCfg.Chargers_conns != nil {
		scfg.ChargerSConns = make([]string, len(*jsnCfg.Chargers_conns))
		for idx, connID := range *jsnCfg.Chargers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.ChargerSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
			} else {
				scfg.ChargerSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Rals_conns != nil {
		scfg.RALsConns = make([]string, len(*jsnCfg.Rals_conns))
		for idx, connID := range *jsnCfg.Rals_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.RALsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)
			} else {
				scfg.RALsConns[idx] = connID
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		scfg.ResSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, connID := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.ResSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			} else {
				scfg.ResSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		scfg.ThreshSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, connID := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.ThreshSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			} else {
				scfg.ThreshSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		scfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, connID := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS)
			} else {
				scfg.StatSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Suppliers_conns != nil {
		scfg.SupplSConns = make([]string, len(*jsnCfg.Suppliers_conns))
		for idx, connID := range *jsnCfg.Suppliers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.SupplSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSuppliers)
			} else {
				scfg.SupplSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Attributes_conns != nil {
		scfg.AttrSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, connID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.AttrSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				scfg.AttrSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		scfg.CDRsConns = make([]string, len(*jsnCfg.Cdrs_conns))
		for idx, connID := range *jsnCfg.Cdrs_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				scfg.CDRsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)
			} else {
				scfg.CDRsConns[idx] = connID
			}
		}
	}
	if jsnCfg.Replication_conns != nil {
		scfg.ReplicationConns = make([]*RemoteHost, len(*jsnCfg.Replication_conns))
		for idx, jsnHaCfg := range *jsnCfg.Replication_conns {
			scfg.ReplicationConns[idx] = NewDfltRemoteHost()
			scfg.ReplicationConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Debit_interval != nil {
		if scfg.DebitInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Debit_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Store_session_costs != nil {
		scfg.StoreSCosts = *jsnCfg.Store_session_costs
	}
	if jsnCfg.Min_call_duration != nil {
		if scfg.MinCallDuration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Min_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Max_call_duration != nil {
		if scfg.MaxCallDuration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Max_call_duration); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl != nil {
		if scfg.SessionTTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl_max_delay != nil {
		if maxTTLDelay, err := utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_max_delay); err != nil {
			return err
		} else {
			scfg.SessionTTLMaxDelay = &maxTTLDelay
		}
	}
	if jsnCfg.Session_ttl_last_used != nil {
		if sessionTTLLastUsed, err := utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_last_used); err != nil {
			return err
		} else {
			scfg.SessionTTLLastUsed = &sessionTTLLastUsed
		}
	}
	if jsnCfg.Session_indexes != nil {
		scfg.SessionIndexes = utils.StringMapFromSlice(*jsnCfg.Session_indexes)
	}
	if jsnCfg.Client_protocol != nil {
		scfg.ClientProtocol = *jsnCfg.Client_protocol
	}
	if jsnCfg.Channel_sync_interval != nil {
		if scfg.ChannelSyncInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Channel_sync_interval); err != nil {
			return err
		}
	}
	if jsnCfg.Terminate_attempts != nil {
		scfg.TerminateAttempts = *jsnCfg.Terminate_attempts
	}
	return nil
}

type FsAgentCfg struct {
	Enabled       bool
	SessionSConns []string
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
		self.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				self.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				self.SessionSConns[idx] = connID
			}
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

// Uses stored defaults so we can pre-populate by loading from JSON config
func NewDefaultAsteriskConnCfg() *AsteriskConnCfg {
	if dfltAstConnCfg == nil {
		return new(AsteriskConnCfg) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltAstConnCfg // Copy the value instead of it's pointer
	return &dfltVal
}

type AsteriskConnCfg struct {
	Alias           string
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

type AsteriskAgentCfg struct {
	Enabled       bool
	SessionSConns []string
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
		aCfg.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if attrConn == utils.MetaInternal {
				aCfg.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				aCfg.SessionSConns[idx] = attrConn
			}
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
