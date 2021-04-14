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
	"fmt"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewDfltFsConnConfig returns the first cached default value for a FreeSWITCHAgent connection
func NewDfltFsConnConfig() *FsConnCfg {
	if dfltFsConnConfig == nil {
		return new(FsConnCfg) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltFsConnConfig // Copy the value instead of it's pointer
	return &dfltVal
}

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
func (fs *FsConnCfg) AsMapInterface() map[string]interface{} {
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

// SessionSCfg is the config section for SessionS
type SessionSCfg struct {
	Enabled             bool
	ListenBijson        string
	ListenBigob         string
	ChargerSConns       []string
	ResSConns           []string
	ThreshSConns        []string
	StatSConns          []string
	RouteSConns         []string
	AttrSConns          []string
	CDRsConns           []string
	ReplicationConns    []string
	DebitInterval       time.Duration
	StoreSCosts         bool
	SessionTTL          time.Duration
	SessionTTLMaxDelay  *time.Duration
	SessionTTLLastUsed  *time.Duration
	SessionTTLUsage     *time.Duration
	SessionTTLLastUsage *time.Duration
	SessionIndexes      utils.StringSet
	ClientProtocol      float64
	ChannelSyncInterval time.Duration
	TerminateAttempts   int
	AlterableFields     utils.StringSet
	MinDurLowBalance    time.Duration
	ActionSConns        []string
	STIRCfg             *STIRcfg
	DefaultUsage        map[string]time.Duration
}

func (scfg *SessionSCfg) loadFromJSONCfg(jsnCfg *SessionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		scfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen_bijson != nil {
		scfg.ListenBijson = *jsnCfg.Listen_bijson
	}
	if jsnCfg.Listen_bigob != nil {
		scfg.ListenBigob = *jsnCfg.Listen_bigob
	}
	if jsnCfg.Chargers_conns != nil {
		scfg.ChargerSConns = make([]string, len(*jsnCfg.Chargers_conns))
		for idx, connID := range *jsnCfg.Chargers_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.ChargerSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.ChargerSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers)
			}
		}
	}
	if jsnCfg.Resources_conns != nil {
		scfg.ResSConns = make([]string, len(*jsnCfg.Resources_conns))
		for idx, connID := range *jsnCfg.Resources_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.ResSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.ResSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)
			}
		}
	}
	if jsnCfg.Thresholds_conns != nil {
		scfg.ThreshSConns = make([]string, len(*jsnCfg.Thresholds_conns))
		for idx, connID := range *jsnCfg.Thresholds_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.ThreshSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.ThreshSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)
			}
		}
	}
	if jsnCfg.Stats_conns != nil {
		scfg.StatSConns = make([]string, len(*jsnCfg.Stats_conns))
		for idx, connID := range *jsnCfg.Stats_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.StatSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
	}
	if jsnCfg.Routes_conns != nil {
		scfg.RouteSConns = make([]string, len(*jsnCfg.Routes_conns))
		for idx, connID := range *jsnCfg.Routes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.RouteSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.RouteSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes)
			}
		}
	}
	if jsnCfg.Attributes_conns != nil {
		scfg.AttrSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for idx, connID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.AttrSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.AttrSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	if jsnCfg.Cdrs_conns != nil {
		scfg.CDRsConns = make([]string, len(*jsnCfg.Cdrs_conns))
		for idx, connID := range *jsnCfg.Cdrs_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.CDRsConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.CDRsConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)
			}
		}
	}
	if jsnCfg.Replication_conns != nil {
		scfg.ReplicationConns = make([]string, len(*jsnCfg.Replication_conns))
		for idx, connID := range *jsnCfg.Replication_conns {
			if connID == utils.MetaInternal {
				return fmt.Errorf("Replication connection ID needs to be different than *internal ")
			}
			scfg.ReplicationConns[idx] = connID
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
	if jsnCfg.Session_ttl != nil {
		if scfg.SessionTTL, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl); err != nil {
			return err
		}
	}
	if jsnCfg.Session_ttl_max_delay != nil {
		var maxTTLDelay time.Duration
		if maxTTLDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_max_delay); err != nil {
			return err
		}
		scfg.SessionTTLMaxDelay = &maxTTLDelay
	}
	if jsnCfg.Session_ttl_last_used != nil {
		var sessionTTLLastUsed time.Duration
		if sessionTTLLastUsed, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_last_used); err != nil {
			return err
		}
		scfg.SessionTTLLastUsed = &sessionTTLLastUsed
	}
	if jsnCfg.Session_ttl_usage != nil {
		var sessionTTLUsage time.Duration
		if sessionTTLUsage, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_usage); err != nil {
			return err
		}
		scfg.SessionTTLUsage = &sessionTTLUsage
	}
	if jsnCfg.Session_ttl_last_usage != nil {
		var sessionTTLLastUsage time.Duration
		if sessionTTLLastUsage, err = utils.ParseDurationWithNanosecs(*jsnCfg.Session_ttl_last_usage); err != nil {
			return err
		}
		scfg.SessionTTLLastUsage = &sessionTTLLastUsage
	}
	if jsnCfg.Session_indexes != nil {
		scfg.SessionIndexes = utils.NewStringSet(*jsnCfg.Session_indexes)
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
	if jsnCfg.Alterable_fields != nil {
		scfg.AlterableFields = utils.NewStringSet(*jsnCfg.Alterable_fields)
	}
	if jsnCfg.Min_dur_low_balance != nil {
		if scfg.MinDurLowBalance, err = utils.ParseDurationWithNanosecs(*jsnCfg.Min_dur_low_balance); err != nil {
			return err
		}
	}
	if jsnCfg.Default_usage != nil {
		for k, v := range *jsnCfg.Default_usage {
			if scfg.DefaultUsage[k], err = utils.ParseDurationWithNanosecs(v); err != nil {
				return
			}
		}
	}
	if jsnCfg.Actions_conns != nil {
		scfg.ActionSConns = make([]string, len(*jsnCfg.Actions_conns))
		for idx, connID := range *jsnCfg.Actions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			scfg.ActionSConns[idx] = connID
			if connID == utils.MetaInternal {
				scfg.ActionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions)
			}
		}
	}
	return scfg.STIRCfg.loadFromJSONCfg(jsnCfg.Stir)
}

func (scfg *SessionSCfg) GetDefaultUsage(tor string) time.Duration {
	if tor == utils.EmptyString {
		tor = utils.MetaAny
	}
	return scfg.DefaultUsage[tor]
}

// AsMapInterface returns the config as a map[string]interface{}
func (scfg *SessionSCfg) AsMapInterface() (initialMP map[string]interface{}) {
	maxComputed := make(map[string]string)
	for key, item := range scfg.DefaultUsage {
		if key == utils.MetaAny || key == utils.MetaVoice {
			maxComputed[key] = item.String()
		} else {
			maxComputed[key] = strconv.Itoa(int(item))
		}
	}
	initialMP = map[string]interface{}{
		utils.EnabledCfg:             scfg.Enabled,
		utils.ListenBijsonCfg:        scfg.ListenBijson,
		utils.ListenBigobCfg:         scfg.ListenBigob,
		utils.ReplicationConnsCfg:    scfg.ReplicationConns,
		utils.StoreSCostsCfg:         scfg.StoreSCosts,
		utils.SessionIndexesCfg:      scfg.SessionIndexes.AsSlice(),
		utils.ClientProtocolCfg:      scfg.ClientProtocol,
		utils.TerminateAttemptsCfg:   scfg.TerminateAttempts,
		utils.AlterableFieldsCfg:     scfg.AlterableFields.AsSlice(),
		utils.STIRCfg:                scfg.STIRCfg.AsMapInterface(),
		utils.MinDurLowBalanceCfg:    "0",
		utils.ChannelSyncIntervalCfg: "0",
		utils.DebitIntervalCfg:       "0",
		utils.SessionTTLCfg:          "0",
		utils.DefaultUsageCfg:        maxComputed,
	}
	if scfg.DebitInterval != 0 {
		initialMP[utils.DebitIntervalCfg] = scfg.DebitInterval.String()
	}
	if scfg.SessionTTL != 0 {
		initialMP[utils.SessionTTLCfg] = scfg.SessionTTL.String()
	}
	if scfg.SessionTTLMaxDelay != nil {
		initialMP[utils.SessionTTLMaxDelayCfg] = scfg.SessionTTLMaxDelay.String()
	}
	if scfg.SessionTTLLastUsed != nil {
		initialMP[utils.SessionTTLLastUsedCfg] = scfg.SessionTTLLastUsed.String()
	}
	if scfg.SessionTTLUsage != nil {
		initialMP[utils.SessionTTLUsageCfg] = scfg.SessionTTLUsage.String()
	}
	if scfg.SessionTTLLastUsage != nil {
		initialMP[utils.SessionTTLLastUsageCfg] = scfg.SessionTTLLastUsage.String()
	}
	if scfg.ChannelSyncInterval != 0 {
		initialMP[utils.ChannelSyncIntervalCfg] = scfg.ChannelSyncInterval.String()
	}
	if scfg.MinDurLowBalance != 0 {
		initialMP[utils.MinDurLowBalanceCfg] = scfg.MinDurLowBalance.String()
	}
	if scfg.ChargerSConns != nil {
		chargerSConns := make([]string, len(scfg.ChargerSConns))
		for i, item := range scfg.ChargerSConns {
			chargerSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers) {
				chargerSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ChargerSConnsCfg] = chargerSConns
	}
	if scfg.ResSConns != nil {
		resSConns := make([]string, len(scfg.ResSConns))
		for i, item := range scfg.ResSConns {
			resSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources) {
				resSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ResourceSConnsCfg] = resSConns
	}
	if scfg.ThreshSConns != nil {
		threshSConns := make([]string, len(scfg.ThreshSConns))
		for i, item := range scfg.ThreshSConns {
			threshSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds) {
				threshSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ThresholdSConnsCfg] = threshSConns
	}
	if scfg.StatSConns != nil {
		statSConns := make([]string, len(scfg.StatSConns))
		for i, item := range scfg.StatSConns {
			statSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats) {
				statSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.StatSConnsCfg] = statSConns
	}
	if scfg.RouteSConns != nil {
		routesConns := make([]string, len(scfg.RouteSConns))
		for i, item := range scfg.RouteSConns {
			routesConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes) {
				routesConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.RouteSConnsCfg] = routesConns
	}
	if scfg.AttrSConns != nil {
		attrSConns := make([]string, len(scfg.AttrSConns))
		for i, item := range scfg.AttrSConns {
			attrSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attrSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attrSConns
	}
	if scfg.CDRsConns != nil {
		CDRsConns := make([]string, len(scfg.CDRsConns))
		for i, item := range scfg.CDRsConns {
			CDRsConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs) {
				CDRsConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.CDRsConnsCfg] = CDRsConns
	}
	if scfg.ActionSConns != nil {
		actionConns := make([]string, len(scfg.ActionSConns))
		for i, item := range scfg.ActionSConns {
			actionConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions) {
				actionConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.ActionSConnsCfg] = actionConns
	}
	return
}

// Clone returns a deep copy of SessionSCfg
func (scfg SessionSCfg) Clone() (cln *SessionSCfg) {
	cln = &SessionSCfg{
		Enabled:             scfg.Enabled,
		ListenBijson:        scfg.ListenBijson,
		DebitInterval:       scfg.DebitInterval,
		StoreSCosts:         scfg.StoreSCosts,
		SessionTTL:          scfg.SessionTTL,
		ClientProtocol:      scfg.ClientProtocol,
		ChannelSyncInterval: scfg.ChannelSyncInterval,
		TerminateAttempts:   scfg.TerminateAttempts,
		MinDurLowBalance:    scfg.MinDurLowBalance,

		SessionIndexes:  scfg.SessionIndexes.Clone(),
		AlterableFields: scfg.AlterableFields.Clone(),
		STIRCfg:         scfg.STIRCfg.Clone(),
		DefaultUsage:    make(map[string]time.Duration),
	}
	for k, v := range scfg.DefaultUsage {
		cln.DefaultUsage[k] = v
	}
	if scfg.SessionTTLMaxDelay != nil {
		cln.SessionTTLMaxDelay = utils.DurationPointer(*scfg.SessionTTLMaxDelay)
	}
	if scfg.SessionTTLLastUsed != nil {
		cln.SessionTTLLastUsed = utils.DurationPointer(*scfg.SessionTTLLastUsed)
	}
	if scfg.SessionTTLUsage != nil {
		cln.SessionTTLUsage = utils.DurationPointer(*scfg.SessionTTLUsage)
	}
	if scfg.SessionTTLLastUsage != nil {
		cln.SessionTTLLastUsage = utils.DurationPointer(*scfg.SessionTTLLastUsage)
	}

	if scfg.ChargerSConns != nil {
		cln.ChargerSConns = make([]string, len(scfg.ChargerSConns))
		for i, con := range scfg.ChargerSConns {
			cln.ChargerSConns[i] = con
		}
	}
	if scfg.ResSConns != nil {
		cln.ResSConns = make([]string, len(scfg.ResSConns))
		for i, con := range scfg.ResSConns {
			cln.ResSConns[i] = con
		}
	}
	if scfg.ThreshSConns != nil {
		cln.ThreshSConns = make([]string, len(scfg.ThreshSConns))
		for i, con := range scfg.ThreshSConns {
			cln.ThreshSConns[i] = con
		}
	}
	if scfg.StatSConns != nil {
		cln.StatSConns = make([]string, len(scfg.StatSConns))
		for i, con := range scfg.StatSConns {
			cln.StatSConns[i] = con
		}
	}
	if scfg.RouteSConns != nil {
		cln.RouteSConns = make([]string, len(scfg.RouteSConns))
		for i, con := range scfg.RouteSConns {
			cln.RouteSConns[i] = con
		}
	}
	if scfg.AttrSConns != nil {
		cln.AttrSConns = make([]string, len(scfg.AttrSConns))
		for i, con := range scfg.AttrSConns {
			cln.AttrSConns[i] = con
		}
	}
	if scfg.CDRsConns != nil {
		cln.CDRsConns = make([]string, len(scfg.CDRsConns))
		for i, con := range scfg.CDRsConns {
			cln.CDRsConns[i] = con
		}
	}
	if scfg.ReplicationConns != nil {
		cln.ReplicationConns = make([]string, len(scfg.ReplicationConns))
		for i, con := range scfg.ReplicationConns {
			cln.ReplicationConns[i] = con
		}
	}
	if scfg.ActionSConns != nil {
		cln.ActionSConns = make([]string, len(scfg.ActionSConns))
		for i, con := range scfg.ActionSConns {
			cln.ActionSConns[i] = con
		}
	}

	return
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

func (fscfg *FsAgentCfg) loadFromJSONCfg(jsnCfg *FreeswitchAgentJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		fscfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		fscfg.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, connID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			fscfg.SessionSConns[idx] = connID
			if connID == utils.MetaInternal ||
				connID == rpcclient.BiRPCInternal {
				fscfg.SessionSConns[idx] = utils.ConcatenatedKey(connID, utils.MetaSessionS)
			}
		}
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
			fscfg.EventSocketConns[idx] = NewDfltFsConnConfig()
			fscfg.EventSocketConns[idx].loadFromJSONCfg(jsnConnCfg)
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (fscfg *FsAgentCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:             fscfg.Enabled,
		utils.SubscribeParkCfg:       fscfg.SubscribePark,
		utils.CreateCdrCfg:           fscfg.CreateCdr,
		utils.LowBalanceAnnFileCfg:   fscfg.LowBalanceAnnFile,
		utils.EmptyBalanceContextCfg: fscfg.EmptyBalanceContext,
		utils.EmptyBalanceAnnFileCfg: fscfg.EmptyBalanceAnnFile,
	}
	if fscfg.SessionSConns != nil {
		sessionSConns := make([]string, len(fscfg.SessionSConns))
		for i, item := range fscfg.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			} else if item == utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS) {
				sessionSConns[i] = rpcclient.BiRPCInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
	}
	if fscfg.ExtraFields != nil {
		extra := make([]string, len(fscfg.ExtraFields))
		for i, rsr := range fscfg.ExtraFields {
			extra[i] = rsr.Rules
		}
		initialMP[utils.ExtraFieldsCfg] = extra
	}

	if fscfg.MaxWaitConnection != 0 {
		initialMP[utils.MaxWaitConnectionCfg] = fscfg.MaxWaitConnection.String()
	} else {
		initialMP[utils.MaxWaitConnectionCfg] = utils.EmptyString
	}
	if fscfg.EventSocketConns != nil {
		eventSocketConns := make([]map[string]interface{}, len(fscfg.EventSocketConns))
		for key, item := range fscfg.EventSocketConns {
			eventSocketConns[key] = item.AsMapInterface()
		}
		initialMP[utils.EventSocketConnsCfg] = eventSocketConns
	}
	return
}

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
	}
	if fscfg.SessionSConns != nil {
		cln.SessionSConns = make([]string, len(fscfg.SessionSConns))
		for i, con := range fscfg.SessionSConns {
			cln.SessionSConns[i] = con
		}
	}
	if fscfg.EventSocketConns != nil {
		cln.EventSocketConns = make([]*FsConnCfg, len(fscfg.EventSocketConns))
		for i, req := range fscfg.EventSocketConns {
			cln.EventSocketConns[i] = req.Clone()
		}
	}
	return
}

// NewDefaultAsteriskConnCfg is uses stored defaults so we can pre-populate by loading from JSON config
func NewDefaultAsteriskConnCfg() *AsteriskConnCfg {
	if dfltAstConnCfg == nil {
		return new(AsteriskConnCfg) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltAstConnCfg // Copy the value instead of it's pointer
	return &dfltVal
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
		aCfg.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			aCfg.SessionSConns[idx] = attrConn
			if attrConn == utils.MetaInternal ||
				attrConn == rpcclient.BiRPCInternal {
				aCfg.SessionSConns[idx] = utils.ConcatenatedKey(attrConn, utils.MetaSessionS)
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
		sessionSConns := make([]string, len(aCfg.SessionSConns))
		for i, item := range aCfg.SessionSConns {
			sessionSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS) {
				sessionSConns[i] = utils.MetaInternal
			} else if item == utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS) {
				sessionSConns[i] = rpcclient.BiRPCInternal
			}
		}
		initialMP[utils.SessionSConnsCfg] = sessionSConns
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
		cln.SessionSConns = make([]string, len(aCfg.SessionSConns))
		for i, con := range aCfg.SessionSConns {
			cln.SessionSConns[i] = con
		}
	}
	if aCfg.AsteriskConns != nil {
		cln.AsteriskConns = make([]*AsteriskConnCfg, len(aCfg.AsteriskConns))
		for i, req := range aCfg.AsteriskConns {
			cln.AsteriskConns[i] = req.Clone()
		}
	}
	return
}

// STIRcfg the confuguration structure for STIR
type STIRcfg struct {
	AllowedAttest      utils.StringSet
	PayloadMaxduration time.Duration
	DefaultAttest      string
	PublicKeyPath      string
	PrivateKeyPath     string
}

func (stirCfg *STIRcfg) loadFromJSONCfg(jsnCfg *STIRJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Allowed_attest != nil {
		stirCfg.AllowedAttest = utils.NewStringSet(*jsnCfg.Allowed_attest)
	}
	if jsnCfg.Payload_maxduration != nil {
		if stirCfg.PayloadMaxduration, err = utils.ParseDurationWithNanosecs(*jsnCfg.Payload_maxduration); err != nil {
			return err
		}
	}
	if jsnCfg.Default_attest != nil {
		stirCfg.DefaultAttest = *jsnCfg.Default_attest
	}
	if jsnCfg.Publickey_path != nil {
		stirCfg.PublicKeyPath = *jsnCfg.Publickey_path
	}
	if jsnCfg.Privatekey_path != nil {
		stirCfg.PrivateKeyPath = *jsnCfg.Privatekey_path
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (stirCfg *STIRcfg) AsMapInterface() (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.DefaultAttestCfg:      stirCfg.DefaultAttest,
		utils.PublicKeyPathCfg:      stirCfg.PublicKeyPath,
		utils.PrivateKeyPathCfg:     stirCfg.PrivateKeyPath,
		utils.AllowedAtestCfg:       stirCfg.AllowedAttest.AsSlice(),
		utils.PayloadMaxdurationCfg: "0",
	}
	if stirCfg.PayloadMaxduration > 0 {
		initialMP[utils.PayloadMaxdurationCfg] = stirCfg.PayloadMaxduration.String()
	} else if stirCfg.PayloadMaxduration < 0 {
		initialMP[utils.PayloadMaxdurationCfg] = "-1"
	}
	return
}

// Clone returns a deep copy of STIRcfg
func (stirCfg STIRcfg) Clone() *STIRcfg {
	return &STIRcfg{
		AllowedAttest:      stirCfg.AllowedAttest.Clone(),
		PayloadMaxduration: stirCfg.PayloadMaxduration,
		DefaultAttest:      stirCfg.DefaultAttest,
		PublicKeyPath:      stirCfg.PublicKeyPath,
		PrivateKeyPath:     stirCfg.PrivateKeyPath,
	}
}
