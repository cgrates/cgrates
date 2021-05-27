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
)

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
		scfg.ChargerSConns = updateInternalConns(*jsnCfg.Chargers_conns, utils.MetaChargers)
	}
	if jsnCfg.Resources_conns != nil {
		scfg.ResSConns = updateInternalConns(*jsnCfg.Resources_conns, utils.MetaResources)
	}
	if jsnCfg.Thresholds_conns != nil {
		scfg.ThreshSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Stats_conns != nil {
		scfg.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Routes_conns != nil {
		scfg.RouteSConns = updateInternalConns(*jsnCfg.Routes_conns, utils.MetaRoutes)
	}
	if jsnCfg.Attributes_conns != nil {
		scfg.AttrSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	if jsnCfg.Cdrs_conns != nil {
		scfg.CDRsConns = updateInternalConns(*jsnCfg.Cdrs_conns, utils.MetaCDRs)
	}
	if jsnCfg.Actions_conns != nil {
		scfg.ActionSConns = updateInternalConns(*jsnCfg.Actions_conns, utils.MetaActions)
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
		for k, v := range jsnCfg.Default_usage {
			if scfg.DefaultUsage[k], err = utils.ParseDurationWithNanosecs(v); err != nil {
				return
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
		initialMP[utils.ChargerSConnsCfg] = getInternalJSONConns(scfg.ChargerSConns)
	}
	if scfg.ResSConns != nil {
		initialMP[utils.ResourceSConnsCfg] = getInternalJSONConns(scfg.ResSConns)
	}
	if scfg.ThreshSConns != nil {
		initialMP[utils.ThresholdSConnsCfg] = getInternalJSONConns(scfg.ThreshSConns)
	}
	if scfg.StatSConns != nil {
		initialMP[utils.StatSConnsCfg] = getInternalJSONConns(scfg.StatSConns)
	}
	if scfg.RouteSConns != nil {
		initialMP[utils.RouteSConnsCfg] = getInternalJSONConns(scfg.RouteSConns)
	}
	if scfg.AttrSConns != nil {
		initialMP[utils.AttributeSConnsCfg] = getInternalJSONConns(scfg.AttrSConns)
	}
	if scfg.CDRsConns != nil {
		initialMP[utils.CDRsConnsCfg] = getInternalJSONConns(scfg.CDRsConns)
	}
	if scfg.ActionSConns != nil {
		initialMP[utils.ActionSConnsCfg] = getInternalJSONConns(scfg.ActionSConns)
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
		cln.ChargerSConns = utils.CloneStringSlice(scfg.ChargerSConns)
	}
	if scfg.ResSConns != nil {
		cln.ResSConns = utils.CloneStringSlice(scfg.ResSConns)
	}
	if scfg.ThreshSConns != nil {
		cln.ThreshSConns = utils.CloneStringSlice(scfg.ThreshSConns)
	}
	if scfg.StatSConns != nil {
		cln.StatSConns = utils.CloneStringSlice(scfg.StatSConns)
	}
	if scfg.RouteSConns != nil {
		cln.RouteSConns = utils.CloneStringSlice(scfg.RouteSConns)
	}
	if scfg.AttrSConns != nil {
		cln.AttrSConns = utils.CloneStringSlice(scfg.AttrSConns)
	}
	if scfg.CDRsConns != nil {
		cln.CDRsConns = utils.CloneStringSlice(scfg.CDRsConns)
	}
	if scfg.ReplicationConns != nil {
		cln.ReplicationConns = utils.CloneStringSlice(scfg.ReplicationConns)
	}
	if scfg.ActionSConns != nil {
		cln.ActionSConns = utils.CloneStringSlice(scfg.ActionSConns)
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

type STIRJsonCfg struct {
	Allowed_attest      *[]string
	Payload_maxduration *string
	Default_attest      *string
	Publickey_path      *string
	Privatekey_path     *string
}

func diffSTIRJsonCfg(d *STIRJsonCfg, v1, v2 *STIRcfg) *STIRJsonCfg {
	if d == nil {
		d = new(STIRJsonCfg)
	}
	if v1.AllowedAttest.Equals(v2.AllowedAttest) {
		d.Allowed_attest = nil
		if v2.AllowedAttest != nil {
			d.Allowed_attest = utils.SliceStringPointer(v2.AllowedAttest.AsSlice())
		}
	}
	if v1.PayloadMaxduration != v2.PayloadMaxduration {
		d.Payload_maxduration = utils.StringPointer(v2.PayloadMaxduration.String())
	}
	if v1.DefaultAttest != v2.DefaultAttest {
		d.Default_attest = utils.StringPointer(v2.DefaultAttest)
	}
	if v1.PublicKeyPath != v2.PublicKeyPath {
		d.Publickey_path = utils.StringPointer(v2.PublicKeyPath)
	}
	if v1.PrivateKeyPath != v2.PrivateKeyPath {
		d.Privatekey_path = utils.StringPointer(v2.PrivateKeyPath)
	}
	return d
}

// SessionSJsonCfg config section
type SessionSJsonCfg struct {
	Enabled                *bool
	Listen_bijson          *string
	Listen_bigob           *string
	Chargers_conns         *[]string
	Resources_conns        *[]string
	Thresholds_conns       *[]string
	Stats_conns            *[]string
	Routes_conns           *[]string
	Cdrs_conns             *[]string
	Replication_conns      *[]string
	Attributes_conns       *[]string
	Debit_interval         *string
	Store_session_costs    *bool
	Session_ttl            *string
	Session_ttl_max_delay  *string
	Session_ttl_last_used  *string
	Session_ttl_usage      *string
	Session_ttl_last_usage *string
	Session_indexes        *[]string
	Client_protocol        *float64
	Channel_sync_interval  *string
	Terminate_attempts     *int
	Alterable_fields       *[]string
	Min_dur_low_balance    *string
	Actions_conns          *[]string
	Stir                   *STIRJsonCfg
	Default_usage          map[string]string
}

func diffSessionSJsonCfg(d *SessionSJsonCfg, v1, v2 *SessionSCfg) *SessionSJsonCfg {
	if d == nil {
		d = new(SessionSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.ListenBijson != v2.ListenBijson {
		d.Listen_bijson = utils.StringPointer(v2.ListenBijson)
	}
	if v1.ListenBigob != v2.ListenBigob {
		d.Listen_bigob = utils.StringPointer(v2.ListenBigob)
	}
	if !utils.SliceStringEqual(v1.ChargerSConns, v2.ChargerSConns) {
		d.Chargers_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ChargerSConns))
	}
	if !utils.SliceStringEqual(v1.ResSConns, v2.ResSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResSConns))
	}
	if !utils.SliceStringEqual(v1.ThreshSConns, v2.ThreshSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThreshSConns))
	}
	if !utils.SliceStringEqual(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !utils.SliceStringEqual(v1.RouteSConns, v2.RouteSConns) {
		d.Routes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RouteSConns))
	}
	if !utils.SliceStringEqual(v1.AttrSConns, v2.AttrSConns) {
		d.Cdrs_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttrSConns))
	}
	if !utils.SliceStringEqual(v1.CDRsConns, v2.CDRsConns) {
		d.Replication_conns = utils.SliceStringPointer(getInternalJSONConns(v2.CDRsConns))
	}
	if !utils.SliceStringEqual(v1.ReplicationConns, v2.ReplicationConns) {
		d.Attributes_conns = utils.SliceStringPointer(v2.ReplicationConns)
	}
	if v1.DebitInterval != v2.DebitInterval {
		d.Debit_interval = utils.StringPointer(v2.DebitInterval.String())
	}
	if v1.StoreSCosts != v2.StoreSCosts {
		d.Store_session_costs = utils.BoolPointer(v2.StoreSCosts)
	}
	if v1.SessionTTL != v2.SessionTTL {
		d.Session_ttl = utils.StringPointer(v2.SessionTTL.String())
	}
	if v2.SessionTTLMaxDelay != nil {
		if v1.SessionTTLMaxDelay == nil ||
			*v1.SessionTTLMaxDelay != *v2.SessionTTLMaxDelay {
			d.Session_ttl_max_delay = utils.StringPointer(v2.SessionTTLMaxDelay.String())
		}
	} else {
		d.Session_ttl_max_delay = nil
	}
	if v2.SessionTTLLastUsed != nil {
		if v1.SessionTTLLastUsed == nil ||
			*v1.SessionTTLLastUsed != *v2.SessionTTLLastUsed {
			d.Session_ttl_last_used = utils.StringPointer(v2.SessionTTLLastUsed.String())
		}
	} else {
		d.Session_ttl_last_used = nil
	}
	if v2.SessionTTLUsage != nil {
		if v1.SessionTTLUsage == nil ||
			*v1.SessionTTLUsage != *v2.SessionTTLUsage {
			d.Session_ttl_usage = utils.StringPointer(v2.SessionTTLUsage.String())
		}
	} else {
		d.Session_ttl_usage = nil
	}
	if v2.SessionTTLLastUsage != nil {
		if v1.SessionTTLLastUsage == nil ||
			*v1.SessionTTLLastUsage != *v2.SessionTTLLastUsage {
			d.Session_ttl_last_usage = utils.StringPointer(v2.SessionTTLLastUsage.String())
		}
	} else {
		d.Session_ttl_last_usage = nil
	}
	if !v1.SessionIndexes.Equals(v2.SessionIndexes) {
		d.Session_indexes = utils.SliceStringPointer(v2.SessionIndexes.AsSlice())
	}
	if v1.ClientProtocol != v2.ClientProtocol {
		d.Client_protocol = utils.Float64Pointer(v2.ClientProtocol)
	}
	if v1.ChannelSyncInterval != v2.ChannelSyncInterval {
		d.Channel_sync_interval = utils.StringPointer(v2.ChannelSyncInterval.String())
	}
	if v1.TerminateAttempts != v2.TerminateAttempts {
		d.Terminate_attempts = utils.IntPointer(v2.TerminateAttempts)
	}
	if !v1.AlterableFields.Equals(v2.AlterableFields) {
		d.Alterable_fields = utils.SliceStringPointer(v2.AlterableFields.AsSlice())
	}
	if v1.MinDurLowBalance != v2.MinDurLowBalance {
		d.Min_dur_low_balance = utils.StringPointer(v2.MinDurLowBalance.String())
	}
	if !utils.SliceStringEqual(v1.ActionSConns, v2.ActionSConns) {
		d.Actions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ActionSConns))
	}
	d.Stir = diffSTIRJsonCfg(d.Stir, v1.STIRCfg, v2.STIRCfg)
	if d.Default_usage == nil {
		d.Default_usage = make(map[string]string)
	}
	for tor, usage2 := range v2.DefaultUsage {
		if usage1, has := v1.DefaultUsage[tor]; !has || usage1 != usage2 {
			d.Default_usage[tor] = usage2.String()
		}
	}
	return d
}
