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
	"slices"
	"strconv"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	SessionsAccountsDftOpt               = false
	SessionsAttributesDftOpt             = false
	SessionsCDRsDftOpt                   = false
	SessionsChargersDftOpt               = false
	SessionsResourcesDftOpt              = false
	SessionsRoutesDftOpt                 = false
	SessionsStatsDftOpt                  = false
	SessionsThresholdsDftOpt             = false
	SessionsInitiateDftOpt               = false
	SessionsUpdateDftOpt                 = false
	SessionsTerminateDftOpt              = false
	SessionsMessageDftOpt                = false
	SessionsAttributesDerivedReplyDftOpt = false
	SessionsBlockerErrorDftOpt           = false
	SessionsCDRsDerivedReplyDftOpt       = false
	SessionsResourcesAuthorizeDftOpt     = false
	SessionsResourcesAllocateDftOpt      = false
	SessionsResourcesReleaseDftOpt       = false
	SessionsResourcesDerivedReplyDftOpt  = false
	SessionsRoutesDerivedReplyDftOpt     = false
	SessionsStatsDerivedReplyDftOpt      = false
	SessionsThresholdsDerivedReplyDftOpt = false
	SessionsMaxUsageDftOpt               = false
	SessionsForceDurationDftOpt          = false
	SessionsTTLDftOpt                    = 0
	SessionsChargeableDftOpt             = true
	SessionsTTLMaxDelayDftOpt            = 0
	SessionsDebitIntervalDftOpt          = 0
)

type SessionsOpts struct {
	Accounts               []*utils.DynamicBoolOpt
	Attributes             []*utils.DynamicBoolOpt
	CDRs                   []*utils.DynamicBoolOpt
	Chargers               []*utils.DynamicBoolOpt
	Resources              []*utils.DynamicBoolOpt
	Routes                 []*utils.DynamicBoolOpt
	Stats                  []*utils.DynamicBoolOpt
	Thresholds             []*utils.DynamicBoolOpt
	Initiate               []*utils.DynamicBoolOpt
	Update                 []*utils.DynamicBoolOpt
	Terminate              []*utils.DynamicBoolOpt
	Message                []*utils.DynamicBoolOpt
	AttributesDerivedReply []*utils.DynamicBoolOpt
	BlockerError           []*utils.DynamicBoolOpt
	CDRsDerivedReply       []*utils.DynamicBoolOpt
	ResourcesAuthorize     []*utils.DynamicBoolOpt
	ResourcesAllocate      []*utils.DynamicBoolOpt
	ResourcesRelease       []*utils.DynamicBoolOpt
	ResourcesDerivedReply  []*utils.DynamicBoolOpt
	RoutesDerivedReply     []*utils.DynamicBoolOpt
	StatsDerivedReply      []*utils.DynamicBoolOpt
	ThresholdsDerivedReply []*utils.DynamicBoolOpt
	MaxUsage               []*utils.DynamicBoolOpt
	ForceDuration          []*utils.DynamicBoolOpt
	TTL                    []*utils.DynamicDurationOpt
	Chargeable             []*utils.DynamicBoolOpt
	TTLLastUsage           []*utils.DynamicDurationPointerOpt
	TTLLastUsed            []*utils.DynamicDurationPointerOpt
	DebitInterval          []*utils.DynamicDurationOpt
	TTLMaxDelay            []*utils.DynamicDurationOpt
	TTLUsage               []*utils.DynamicDurationPointerOpt
}

// SessionSCfg is the config section for SessionS
type SessionSCfg struct {
	Enabled             bool
	ListenBijson        string
	ListenBigob         string
	ChargerSConns       []string
	ResourceSConns      []string
	ThresholdSConns     []string
	StatSConns          []string
	RouteSConns         []string
	AttributeSConns     []string
	CDRsConns           []string
	ReplicationConns    []string
	RateSConns          []string
	AccountSConns       []string
	StoreSCosts         bool
	SessionIndexes      utils.StringSet
	ClientProtocol      float64
	ChannelSyncInterval time.Duration
	TerminateAttempts   int
	AlterableFields     utils.StringSet
	MinDurLowBalance    time.Duration
	ActionSConns        []string
	STIRCfg             *STIRcfg
	DefaultUsage        map[string]time.Duration
	Opts                *SessionsOpts
}

// loadSessionSCfg loads the SessionS section of the configuration
func (scfg *SessionSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnSessionSCfg := new(SessionSJsonCfg)
	if err = jsnCfg.GetSection(ctx, SessionSJSON, jsnSessionSCfg); err != nil {
		return
	}
	return scfg.loadFromJSONCfg(jsnSessionSCfg)
}

func (sesOpts *SessionsOpts) loadFromJSONCfg(jsnCfg *SessionsOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Accounts != nil {
		sesOpts.Accounts = append(sesOpts.Accounts, jsnCfg.Accounts...)
	}
	if jsnCfg.Attributes != nil {
		sesOpts.Attributes = append(sesOpts.Attributes, jsnCfg.Attributes...)
	}
	if jsnCfg.CDRs != nil {
		sesOpts.CDRs = append(sesOpts.CDRs, jsnCfg.CDRs...)
	}
	if jsnCfg.Chargers != nil {
		sesOpts.Chargers = append(sesOpts.Chargers, jsnCfg.Chargers...)
	}
	if jsnCfg.Resources != nil {
		sesOpts.Resources = append(sesOpts.Resources, jsnCfg.Resources...)
	}
	if jsnCfg.Routes != nil {
		sesOpts.Routes = append(sesOpts.Routes, jsnCfg.Routes...)
	}
	if jsnCfg.Stats != nil {
		sesOpts.Stats = append(sesOpts.Stats, jsnCfg.Stats...)
	}
	if jsnCfg.Thresholds != nil {
		sesOpts.Thresholds = append(sesOpts.Thresholds, jsnCfg.Thresholds...)
	}
	if jsnCfg.Initiate != nil {
		sesOpts.Initiate = append(sesOpts.Initiate, jsnCfg.Initiate...)
	}
	if jsnCfg.Update != nil {
		sesOpts.Update = append(sesOpts.Update, jsnCfg.Update...)
	}
	if jsnCfg.Terminate != nil {
		sesOpts.Terminate = append(sesOpts.Terminate, jsnCfg.Terminate...)
	}
	if jsnCfg.Message != nil {
		sesOpts.Message = append(sesOpts.Message, jsnCfg.Message...)
	}
	if jsnCfg.AttributesDerivedReply != nil {
		sesOpts.AttributesDerivedReply = append(sesOpts.AttributesDerivedReply, jsnCfg.AttributesDerivedReply...)
	}
	if jsnCfg.BlockerError != nil {
		sesOpts.BlockerError = append(sesOpts.BlockerError, jsnCfg.BlockerError...)
	}
	if jsnCfg.CDRsDerivedReply != nil {
		sesOpts.CDRsDerivedReply = append(sesOpts.CDRsDerivedReply, jsnCfg.CDRsDerivedReply...)
	}
	if jsnCfg.ResourcesAuthorize != nil {
		sesOpts.ResourcesAuthorize = append(sesOpts.ResourcesAuthorize, jsnCfg.ResourcesAuthorize...)
	}
	if jsnCfg.ResourcesAllocate != nil {
		sesOpts.ResourcesAllocate = append(sesOpts.ResourcesAllocate, jsnCfg.ResourcesAllocate...)
	}
	if jsnCfg.ResourcesRelease != nil {
		sesOpts.ResourcesRelease = append(sesOpts.ResourcesRelease, jsnCfg.ResourcesRelease...)
	}
	if jsnCfg.ResourcesDerivedReply != nil {
		sesOpts.ResourcesDerivedReply = append(sesOpts.ResourcesDerivedReply, jsnCfg.ResourcesDerivedReply...)
	}
	if jsnCfg.RoutesDerivedReply != nil {
		sesOpts.RoutesDerivedReply = append(sesOpts.RoutesDerivedReply, jsnCfg.RoutesDerivedReply...)
	}
	if jsnCfg.StatsDerivedReply != nil {
		sesOpts.StatsDerivedReply = append(sesOpts.StatsDerivedReply, jsnCfg.StatsDerivedReply...)
	}
	if jsnCfg.ThresholdsDerivedReply != nil {
		sesOpts.ThresholdsDerivedReply = append(sesOpts.ThresholdsDerivedReply, jsnCfg.ThresholdsDerivedReply...)
	}
	if jsnCfg.MaxUsage != nil {
		sesOpts.MaxUsage = append(sesOpts.MaxUsage, jsnCfg.MaxUsage...)
	}
	if jsnCfg.ForceDuration != nil {
		sesOpts.ForceDuration = append(sesOpts.ForceDuration, jsnCfg.ForceDuration...)
	}
	if jsnCfg.TTL != nil {
		var ttl []*utils.DynamicDurationOpt
		if ttl, err = utils.StringToDurationDynamicOpts(jsnCfg.TTL); err != nil {
			return
		}
		sesOpts.TTL = append(sesOpts.TTL, ttl...)
	}
	if jsnCfg.Chargeable != nil {
		sesOpts.Chargeable = append(sesOpts.Chargeable, jsnCfg.Chargeable...)
	}
	if jsnCfg.TTLLastUsage != nil {
		var lastUsage []*utils.DynamicDurationPointerOpt
		if lastUsage, err = utils.StringToDurationPointerDynamicOpts(jsnCfg.TTLLastUsage); err != nil {
			return
		}
		sesOpts.TTLLastUsage = append(sesOpts.TTLLastUsage, lastUsage...)
	}
	if jsnCfg.TTLLastUsed != nil {
		var lastUsed []*utils.DynamicDurationPointerOpt
		if lastUsed, err = utils.StringToDurationPointerDynamicOpts(jsnCfg.TTLLastUsed); err != nil {
			return
		}
		sesOpts.TTLLastUsed = append(sesOpts.TTLLastUsed, lastUsed...)
	}
	if jsnCfg.DebitInterval != nil {
		var debitInterval []*utils.DynamicDurationOpt
		if debitInterval, err = utils.StringToDurationDynamicOpts(jsnCfg.DebitInterval); err != nil {
			return
		}
		sesOpts.DebitInterval = append(sesOpts.DebitInterval, debitInterval...)
	}
	if jsnCfg.TTLMaxDelay != nil {
		var maxDelay []*utils.DynamicDurationOpt
		if maxDelay, err = utils.StringToDurationDynamicOpts(jsnCfg.TTLMaxDelay); err != nil {
			return
		}
		sesOpts.TTLMaxDelay = append(sesOpts.TTLMaxDelay, maxDelay...)
	}
	if jsnCfg.TTLUsage != nil {
		var usage []*utils.DynamicDurationPointerOpt
		if usage, err = utils.StringToDurationPointerDynamicOpts(jsnCfg.TTLUsage); err != nil {
			return
		}
		sesOpts.TTLUsage = append(sesOpts.TTLUsage, usage...)
	}
	return
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
		scfg.ResourceSConns = updateInternalConns(*jsnCfg.Resources_conns, utils.MetaResources)
	}
	if jsnCfg.Thresholds_conns != nil {
		scfg.ThresholdSConns = updateInternalConns(*jsnCfg.Thresholds_conns, utils.MetaThresholds)
	}
	if jsnCfg.Stats_conns != nil {
		scfg.StatSConns = updateInternalConns(*jsnCfg.Stats_conns, utils.MetaStats)
	}
	if jsnCfg.Routes_conns != nil {
		scfg.RouteSConns = updateInternalConns(*jsnCfg.Routes_conns, utils.MetaRoutes)
	}
	if jsnCfg.Attributes_conns != nil {
		scfg.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
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
	if jsnCfg.Rates_conns != nil {
		scfg.RateSConns = updateInternalConns(*jsnCfg.Rates_conns, utils.MetaRates)
	}
	if jsnCfg.Accounts_conns != nil {
		scfg.AccountSConns = updateInternalConns(*jsnCfg.Accounts_conns, utils.MetaAccounts)
	}
	if jsnCfg.Store_session_costs != nil {
		scfg.StoreSCosts = *jsnCfg.Store_session_costs
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
	if jsnCfg.Stir != nil {
		if err = scfg.STIRCfg.loadFromJSONCfg(jsnCfg.Stir); err != nil {
			return
		}
	}
	if jsnCfg.Opts != nil {
		if err = scfg.Opts.loadFromJSONCfg(jsnCfg.Opts); err != nil {
			return
		}
	}
	return
}

func (scfg SessionSCfg) GetDefaultUsage(tor string) time.Duration {
	if tor == utils.EmptyString {
		tor = utils.MetaAny
	}
	return scfg.DefaultUsage[tor]
}

// AsMapInterface returns the config as a map[string]any
func (scfg SessionSCfg) AsMapInterface(string) any {
	maxComputed := make(map[string]string)
	for key, item := range scfg.DefaultUsage {
		if key == utils.MetaAny || key == utils.MetaVoice {
			maxComputed[key] = item.String()
		} else {
			maxComputed[key] = strconv.Itoa(int(item))
		}
	}
	opts := map[string]any{
		utils.MetaAccounts:                  scfg.Opts.Accounts,
		utils.MetaAttributes:                scfg.Opts.Attributes,
		utils.MetaCDRs:                      scfg.Opts.CDRs,
		utils.MetaChargers:                  scfg.Opts.Chargers,
		utils.MetaResources:                 scfg.Opts.Resources,
		utils.MetaRoutes:                    scfg.Opts.Routes,
		utils.MetaStats:                     scfg.Opts.Stats,
		utils.MetaThresholds:                scfg.Opts.Thresholds,
		utils.MetaInitiate:                  scfg.Opts.Initiate,
		utils.MetaUpdate:                    scfg.Opts.Update,
		utils.MetaTerminate:                 scfg.Opts.Terminate,
		utils.MetaMessage:                   scfg.Opts.Message,
		utils.MetaAttributesDerivedReplyCfg: scfg.Opts.AttributesDerivedReply,
		utils.MetaBlockerErrorCfg:           scfg.Opts.BlockerError,
		utils.MetaCDRsDerivedReplyCfg:       scfg.Opts.CDRsDerivedReply,
		utils.MetaResourcesAuthorizeCfg:     scfg.Opts.ResourcesAuthorize,
		utils.MetaResourcesAllocateCfg:      scfg.Opts.ResourcesAllocate,
		utils.MetaResourcesReleaseCfg:       scfg.Opts.ResourcesRelease,
		utils.MetaResourcesDerivedReplyCfg:  scfg.Opts.ResourcesDerivedReply,
		utils.MetaRoutesDerivedReplyCfg:     scfg.Opts.RoutesDerivedReply,
		utils.MetaStatsDerivedReplyCfg:      scfg.Opts.StatsDerivedReply,
		utils.MetaThresholdsDerivedReplyCfg: scfg.Opts.ThresholdsDerivedReply,
		utils.MetaMaxUsageCfg:               scfg.Opts.MaxUsage,
		utils.MetaForceDurationCfg:          scfg.Opts.ForceDuration,
		utils.MetaTTLCfg:                    scfg.Opts.TTL,
		utils.MetaChargeableCfg:             scfg.Opts.Chargeable,
		utils.MetaDebitIntervalCfg:          scfg.Opts.DebitInterval,
		utils.MetaTTLLastUsageCfg:           scfg.Opts.TTLLastUsage,
		utils.MetaTTLLastUsedCfg:            scfg.Opts.TTLLastUsed,
		utils.MetaTTLMaxDelayCfg:            scfg.Opts.TTLMaxDelay,
		utils.MetaTTLUsageCfg:               scfg.Opts.TTLUsage,
	}
	mp := map[string]any{
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
		utils.DefaultUsageCfg:        maxComputed,
		utils.OptsCfg:                opts,
	}
	if scfg.ChannelSyncInterval != 0 {
		mp[utils.ChannelSyncIntervalCfg] = scfg.ChannelSyncInterval.String()
	}
	if scfg.MinDurLowBalance != 0 {
		mp[utils.MinDurLowBalanceCfg] = scfg.MinDurLowBalance.String()
	}
	if scfg.ChargerSConns != nil {
		mp[utils.ChargerSConnsCfg] = getInternalJSONConns(scfg.ChargerSConns)
	}
	if scfg.ResourceSConns != nil {
		mp[utils.ResourceSConnsCfg] = getInternalJSONConns(scfg.ResourceSConns)
	}
	if scfg.ThresholdSConns != nil {
		mp[utils.ThresholdSConnsCfg] = getInternalJSONConns(scfg.ThresholdSConns)
	}
	if scfg.StatSConns != nil {
		mp[utils.StatSConnsCfg] = getInternalJSONConns(scfg.StatSConns)
	}
	if scfg.RouteSConns != nil {
		mp[utils.RouteSConnsCfg] = getInternalJSONConns(scfg.RouteSConns)
	}
	if scfg.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(scfg.AttributeSConns)
	}
	if scfg.CDRsConns != nil {
		mp[utils.CDRsConnsCfg] = getInternalJSONConns(scfg.CDRsConns)
	}
	if scfg.ActionSConns != nil {
		mp[utils.ActionSConnsCfg] = getInternalJSONConns(scfg.ActionSConns)
	}
	if scfg.RateSConns != nil {
		mp[utils.RateSConnsCfg] = getInternalJSONConns(scfg.RateSConns)
	}
	if scfg.AccountSConns != nil {
		mp[utils.AccountSConnsCfg] = getInternalJSONConns(scfg.AccountSConns)
	}
	return mp
}

func (SessionSCfg) SName() string              { return SessionSJSON }
func (scfg SessionSCfg) CloneSection() Section { return scfg.Clone() }

func (sesOpts *SessionsOpts) Clone() (cln *SessionsOpts) {
	var acntS []*utils.DynamicBoolOpt
	if sesOpts.Accounts != nil {
		acntS = utils.CloneDynamicBoolOpt(sesOpts.Accounts)
	}
	var attrS []*utils.DynamicBoolOpt
	if sesOpts.Attributes != nil {
		attrS = utils.CloneDynamicBoolOpt(sesOpts.Attributes)
	}
	var cdrS []*utils.DynamicBoolOpt
	if sesOpts.CDRs != nil {
		cdrS = utils.CloneDynamicBoolOpt(sesOpts.CDRs)
	}
	var chrgS []*utils.DynamicBoolOpt
	if sesOpts.Chargers != nil {
		chrgS = utils.CloneDynamicBoolOpt(sesOpts.Chargers)
	}
	var reS []*utils.DynamicBoolOpt
	if sesOpts.Resources != nil {
		reS = utils.CloneDynamicBoolOpt(sesOpts.Resources)
	}
	var rouS []*utils.DynamicBoolOpt
	if sesOpts.Routes != nil {
		rouS = utils.CloneDynamicBoolOpt(sesOpts.Routes)
	}
	var stS []*utils.DynamicBoolOpt
	if sesOpts.Stats != nil {
		stS = utils.CloneDynamicBoolOpt(sesOpts.Stats)
	}
	var thdS []*utils.DynamicBoolOpt
	if sesOpts.Thresholds != nil {
		thdS = utils.CloneDynamicBoolOpt(sesOpts.Thresholds)
	}
	var initS []*utils.DynamicBoolOpt
	if sesOpts.Initiate != nil {
		initS = utils.CloneDynamicBoolOpt(sesOpts.Initiate)
	}
	var updS []*utils.DynamicBoolOpt
	if sesOpts.Update != nil {
		updS = utils.CloneDynamicBoolOpt(sesOpts.Update)
	}
	var termS []*utils.DynamicBoolOpt
	if sesOpts.Terminate != nil {
		termS = utils.CloneDynamicBoolOpt(sesOpts.Terminate)
	}
	var msg []*utils.DynamicBoolOpt
	if sesOpts.Message != nil {
		msg = utils.CloneDynamicBoolOpt(sesOpts.Message)
	}
	var attrDerivedReply []*utils.DynamicBoolOpt
	if sesOpts.AttributesDerivedReply != nil {
		attrDerivedReply = utils.CloneDynamicBoolOpt(sesOpts.AttributesDerivedReply)
	}
	var blockerErr []*utils.DynamicBoolOpt
	if sesOpts.BlockerError != nil {
		blockerErr = utils.CloneDynamicBoolOpt(sesOpts.BlockerError)
	}
	var cdrsDerivedReply []*utils.DynamicBoolOpt
	if sesOpts.CDRsDerivedReply != nil {
		cdrsDerivedReply = utils.CloneDynamicBoolOpt(sesOpts.CDRsDerivedReply)
	}
	var resAuthorize []*utils.DynamicBoolOpt
	if sesOpts.ResourcesAuthorize != nil {
		resAuthorize = utils.CloneDynamicBoolOpt(sesOpts.ResourcesAuthorize)
	}
	var resAllocate []*utils.DynamicBoolOpt
	if sesOpts.ResourcesAllocate != nil {
		resAllocate = utils.CloneDynamicBoolOpt(sesOpts.ResourcesAllocate)
	}
	var resRelease []*utils.DynamicBoolOpt
	if sesOpts.ResourcesRelease != nil {
		resRelease = utils.CloneDynamicBoolOpt(sesOpts.ResourcesRelease)
	}
	var resDerivedReply []*utils.DynamicBoolOpt
	if sesOpts.ResourcesDerivedReply != nil {
		resDerivedReply = utils.CloneDynamicBoolOpt(sesOpts.ResourcesDerivedReply)
	}
	var rouDerivedReply []*utils.DynamicBoolOpt
	if sesOpts.RoutesDerivedReply != nil {
		rouDerivedReply = utils.CloneDynamicBoolOpt(sesOpts.RoutesDerivedReply)
	}
	var stsDerivedReply []*utils.DynamicBoolOpt
	if sesOpts.StatsDerivedReply != nil {
		stsDerivedReply = utils.CloneDynamicBoolOpt(sesOpts.StatsDerivedReply)
	}
	var thdsDerivedReply []*utils.DynamicBoolOpt
	if sesOpts.ThresholdsDerivedReply != nil {
		thdsDerivedReply = utils.CloneDynamicBoolOpt(sesOpts.ThresholdsDerivedReply)
	}
	var maxUsage []*utils.DynamicBoolOpt
	if sesOpts.MaxUsage != nil {
		maxUsage = utils.CloneDynamicBoolOpt(sesOpts.MaxUsage)
	}
	var forceDuration []*utils.DynamicBoolOpt
	if sesOpts.ForceDuration != nil {
		forceDuration = utils.CloneDynamicBoolOpt(sesOpts.ForceDuration)
	}
	var ttl []*utils.DynamicDurationOpt
	if sesOpts.TTL != nil {
		ttl = utils.CloneDynamicDurationOpt(sesOpts.TTL)
	}
	var chargeable []*utils.DynamicBoolOpt
	if sesOpts.Chargeable != nil {
		chargeable = utils.CloneDynamicBoolOpt(sesOpts.Chargeable)
	}
	var debitIvl []*utils.DynamicDurationOpt
	if sesOpts.DebitInterval != nil {
		debitIvl = utils.CloneDynamicDurationOpt(sesOpts.DebitInterval)
	}
	var lastUsg []*utils.DynamicDurationPointerOpt
	if sesOpts.TTLLastUsage != nil {
		lastUsg = utils.CloneDynamicDurationPointerOpt(sesOpts.TTLLastUsage)
	}
	var lastUsed []*utils.DynamicDurationPointerOpt
	if sesOpts.TTLLastUsed != nil {
		lastUsed = utils.CloneDynamicDurationPointerOpt(sesOpts.TTLLastUsed)
	}
	var maxDelay []*utils.DynamicDurationOpt
	if sesOpts.TTLMaxDelay != nil {
		maxDelay = utils.CloneDynamicDurationOpt(sesOpts.TTLMaxDelay)
	}
	var usg []*utils.DynamicDurationPointerOpt
	if sesOpts.TTLUsage != nil {
		usg = utils.CloneDynamicDurationPointerOpt(sesOpts.TTLUsage)
	}
	return &SessionsOpts{
		Accounts:               acntS,
		Attributes:             attrS,
		CDRs:                   cdrS,
		Chargers:               chrgS,
		Resources:              reS,
		Routes:                 rouS,
		Stats:                  stS,
		Thresholds:             thdS,
		Initiate:               initS,
		Update:                 updS,
		Terminate:              termS,
		Message:                msg,
		AttributesDerivedReply: attrDerivedReply,
		BlockerError:           blockerErr,
		CDRsDerivedReply:       cdrsDerivedReply,
		ResourcesAuthorize:     resAuthorize,
		ResourcesAllocate:      resAllocate,
		ResourcesRelease:       resRelease,
		ResourcesDerivedReply:  resDerivedReply,
		RoutesDerivedReply:     rouDerivedReply,
		StatsDerivedReply:      stsDerivedReply,
		ThresholdsDerivedReply: thdsDerivedReply,
		MaxUsage:               maxUsage,
		ForceDuration:          forceDuration,
		TTL:                    ttl,
		Chargeable:             chargeable,
		DebitInterval:          debitIvl,
		TTLLastUsage:           lastUsg,
		TTLLastUsed:            lastUsed,
		TTLMaxDelay:            maxDelay,
		TTLUsage:               usg,
	}
}

// Clone returns a deep copy of SessionSCfg
func (scfg SessionSCfg) Clone() (cln *SessionSCfg) {
	cln = &SessionSCfg{
		Enabled:             scfg.Enabled,
		ListenBijson:        scfg.ListenBijson,
		StoreSCosts:         scfg.StoreSCosts,
		ClientProtocol:      scfg.ClientProtocol,
		ChannelSyncInterval: scfg.ChannelSyncInterval,
		TerminateAttempts:   scfg.TerminateAttempts,
		MinDurLowBalance:    scfg.MinDurLowBalance,

		SessionIndexes:  scfg.SessionIndexes.Clone(),
		AlterableFields: scfg.AlterableFields.Clone(),
		STIRCfg:         scfg.STIRCfg.Clone(),
		DefaultUsage:    make(map[string]time.Duration),
		Opts:            scfg.Opts.Clone(),
	}
	for k, v := range scfg.DefaultUsage {
		cln.DefaultUsage[k] = v
	}
	if scfg.ChargerSConns != nil {
		cln.ChargerSConns = slices.Clone(scfg.ChargerSConns)
	}
	if scfg.ResourceSConns != nil {
		cln.ResourceSConns = slices.Clone(scfg.ResourceSConns)
	}
	if scfg.ThresholdSConns != nil {
		cln.ThresholdSConns = slices.Clone(scfg.ThresholdSConns)
	}
	if scfg.StatSConns != nil {
		cln.StatSConns = slices.Clone(scfg.StatSConns)
	}
	if scfg.RouteSConns != nil {
		cln.RouteSConns = slices.Clone(scfg.RouteSConns)
	}
	if scfg.AttributeSConns != nil {
		cln.AttributeSConns = slices.Clone(scfg.AttributeSConns)
	}
	if scfg.CDRsConns != nil {
		cln.CDRsConns = slices.Clone(scfg.CDRsConns)
	}
	if scfg.ReplicationConns != nil {
		cln.ReplicationConns = slices.Clone(scfg.ReplicationConns)
	}
	if scfg.ActionSConns != nil {
		cln.ActionSConns = slices.Clone(scfg.ActionSConns)
	}
	if scfg.RateSConns != nil {
		cln.RateSConns = slices.Clone(scfg.RateSConns)
	}
	if scfg.AccountSConns != nil {
		cln.AccountSConns = slices.Clone(scfg.AccountSConns)
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

// AsMapInterface returns the config as a map[string]any
func (stirCfg *STIRcfg) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
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

type SessionsOptsJson struct {
	Accounts               []*utils.DynamicBoolOpt   `json:"*accounts"`
	Attributes             []*utils.DynamicBoolOpt   `json:"*attributes"`
	CDRs                   []*utils.DynamicBoolOpt   `json:"*cdrs"`
	Chargers               []*utils.DynamicBoolOpt   `json:"*chargers"`
	Resources              []*utils.DynamicBoolOpt   `json:"*resources"`
	Routes                 []*utils.DynamicBoolOpt   `json:"*routes"`
	Stats                  []*utils.DynamicBoolOpt   `json:"*stats"`
	Thresholds             []*utils.DynamicBoolOpt   `json:"*thresholds"`
	Initiate               []*utils.DynamicBoolOpt   `json:"*initiate"`
	Update                 []*utils.DynamicBoolOpt   `json:"*update"`
	Terminate              []*utils.DynamicBoolOpt   `json:"*terminate"`
	Message                []*utils.DynamicBoolOpt   `json:"*message"`
	AttributesDerivedReply []*utils.DynamicBoolOpt   `json:"*attributesDerivedReply"`
	BlockerError           []*utils.DynamicBoolOpt   `json:"*blockerError"`
	CDRsDerivedReply       []*utils.DynamicBoolOpt   `json:"*cdrsDerivedReply"`
	ResourcesAuthorize     []*utils.DynamicBoolOpt   `json:"*resourcesAuthorize"`
	ResourcesAllocate      []*utils.DynamicBoolOpt   `json:"*resourcesAllocate"`
	ResourcesRelease       []*utils.DynamicBoolOpt   `json:"*resourcesRelease"`
	ResourcesDerivedReply  []*utils.DynamicBoolOpt   `json:"*resourcesDerivedReply"`
	RoutesDerivedReply     []*utils.DynamicBoolOpt   `json:"*routesDerivedReply"`
	StatsDerivedReply      []*utils.DynamicBoolOpt   `json:"*statsDerivedReply"`
	ThresholdsDerivedReply []*utils.DynamicBoolOpt   `json:"*thresholdsDerivedReply"`
	MaxUsage               []*utils.DynamicBoolOpt   `json:"*maxUsage"`
	ForceDuration          []*utils.DynamicBoolOpt   `json:"*forceDuration"`
	TTL                    []*utils.DynamicStringOpt `json:"*ttl"`
	Chargeable             []*utils.DynamicBoolOpt   `json:"*chargeable"`
	DebitInterval          []*utils.DynamicStringOpt `json:"*debitInterval"`
	TTLLastUsage           []*utils.DynamicStringOpt `json:"*ttlLastUsage"`
	TTLLastUsed            []*utils.DynamicStringOpt `json:"*ttlLastUsed"`
	TTLMaxDelay            []*utils.DynamicStringOpt `json:"*ttlMaxDelay"`
	TTLUsage               []*utils.DynamicStringOpt `json:"*ttlUsage"`
}

// SessionSJsonCfg config section
type SessionSJsonCfg struct {
	Enabled               *bool
	Listen_bijson         *string
	Listen_bigob          *string
	Chargers_conns        *[]string
	Resources_conns       *[]string
	Thresholds_conns      *[]string
	Stats_conns           *[]string
	Routes_conns          *[]string
	Cdrs_conns            *[]string
	Replication_conns     *[]string
	Attributes_conns      *[]string
	Actions_conns         *[]string
	Rates_conns           *[]string
	Accounts_conns        *[]string
	Store_session_costs   *bool
	Session_indexes       *[]string
	Client_protocol       *float64
	Channel_sync_interval *string
	Terminate_attempts    *int
	Alterable_fields      *[]string
	Min_dur_low_balance   *string
	Stir                  *STIRJsonCfg
	Default_usage         map[string]string
	Opts                  *SessionsOptsJson
}

func diffSessionsOptsJsonCfg(d *SessionsOptsJson, v1, v2 *SessionsOpts) *SessionsOptsJson {
	if d == nil {
		d = new(SessionsOptsJson)
	}
	if !utils.DynamicBoolOptEqual(v1.Accounts, v2.Accounts) {
		d.Accounts = v2.Accounts
	}
	if !utils.DynamicBoolOptEqual(v1.Attributes, v2.Attributes) {
		d.Attributes = v2.Attributes
	}
	if !utils.DynamicBoolOptEqual(v1.CDRs, v2.CDRs) {
		d.CDRs = v2.CDRs
	}
	if !utils.DynamicBoolOptEqual(v1.Chargers, v2.Chargers) {
		d.Chargers = v2.Chargers
	}
	if !utils.DynamicBoolOptEqual(v1.Resources, v2.Resources) {
		d.Resources = v2.Resources
	}
	if !utils.DynamicBoolOptEqual(v1.Routes, v2.Routes) {
		d.Routes = v2.Routes
	}
	if !utils.DynamicBoolOptEqual(v1.Stats, v2.Stats) {
		d.Stats = v2.Stats
	}
	if !utils.DynamicBoolOptEqual(v1.Thresholds, v2.Thresholds) {
		d.Thresholds = v2.Thresholds
	}
	if !utils.DynamicBoolOptEqual(v1.Initiate, v2.Initiate) {
		d.Initiate = v2.Initiate
	}
	if !utils.DynamicBoolOptEqual(v1.Update, v2.Update) {
		d.Update = v2.Update
	}
	if !utils.DynamicBoolOptEqual(v1.Terminate, v2.Terminate) {
		d.Terminate = v2.Terminate
	}
	if !utils.DynamicBoolOptEqual(v1.Message, v2.Message) {
		d.Message = v2.Message
	}
	if !utils.DynamicBoolOptEqual(v1.AttributesDerivedReply, v2.AttributesDerivedReply) {
		d.AttributesDerivedReply = v2.AttributesDerivedReply
	}
	if !utils.DynamicBoolOptEqual(v1.BlockerError, v2.BlockerError) {
		d.BlockerError = v2.BlockerError
	}
	if !utils.DynamicBoolOptEqual(v1.CDRsDerivedReply, v2.CDRsDerivedReply) {
		d.CDRsDerivedReply = v2.CDRsDerivedReply
	}
	if !utils.DynamicBoolOptEqual(v1.ResourcesAuthorize, v2.ResourcesAuthorize) {
		d.ResourcesAuthorize = v2.ResourcesAuthorize
	}
	if !utils.DynamicBoolOptEqual(v1.ResourcesAllocate, v2.ResourcesAllocate) {
		d.ResourcesAllocate = v2.ResourcesAllocate
	}
	if !utils.DynamicBoolOptEqual(v1.ResourcesRelease, v2.ResourcesRelease) {
		d.ResourcesRelease = v2.ResourcesRelease
	}
	if !utils.DynamicBoolOptEqual(v1.ResourcesDerivedReply, v2.ResourcesDerivedReply) {
		d.ResourcesDerivedReply = v2.ResourcesDerivedReply
	}
	if !utils.DynamicBoolOptEqual(v1.RoutesDerivedReply, v2.RoutesDerivedReply) {
		d.RoutesDerivedReply = v2.RoutesDerivedReply
	}
	if !utils.DynamicBoolOptEqual(v1.StatsDerivedReply, v2.StatsDerivedReply) {
		d.StatsDerivedReply = v2.StatsDerivedReply
	}
	if !utils.DynamicBoolOptEqual(v1.ThresholdsDerivedReply, v2.ThresholdsDerivedReply) {
		d.ThresholdsDerivedReply = v2.ThresholdsDerivedReply
	}
	if !utils.DynamicBoolOptEqual(v1.MaxUsage, v2.MaxUsage) {
		d.MaxUsage = v2.MaxUsage
	}
	if !utils.DynamicBoolOptEqual(v1.ForceDuration, v2.ForceDuration) {
		d.ForceDuration = v2.ForceDuration
	}
	if !utils.DynamicDurationOptEqual(v1.TTL, v2.TTL) {
		d.TTL = utils.DurationToStringDynamicOpts(v2.TTL)
	}
	if !utils.DynamicBoolOptEqual(v1.Chargeable, v2.Chargeable) {
		d.Chargeable = v2.Chargeable
	}
	if !utils.DynamicDurationPointerOptEqual(v1.TTLLastUsage, v2.TTLLastUsage) {
		d.TTLLastUsage = utils.DurationPointerToStringDynamicOpts(v2.TTLLastUsage)
	}
	if !utils.DynamicDurationPointerOptEqual(v1.TTLLastUsed, v2.TTLLastUsed) {
		d.TTLLastUsed = utils.DurationPointerToStringDynamicOpts(v2.TTLLastUsed)
	}
	if !utils.DynamicDurationOptEqual(v1.DebitInterval, v2.DebitInterval) {
		d.DebitInterval = utils.DurationToStringDynamicOpts(v2.DebitInterval)
	}
	if !utils.DynamicDurationOptEqual(v1.TTLMaxDelay, v2.TTLMaxDelay) {
		d.TTLMaxDelay = utils.DurationToStringDynamicOpts(v2.TTLMaxDelay)
	}
	if !utils.DynamicDurationPointerOptEqual(v1.TTLUsage, v2.TTLUsage) {
		d.TTLUsage = utils.DurationPointerToStringDynamicOpts(v2.TTLUsage)
	}
	return d
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
	if !slices.Equal(v1.ChargerSConns, v2.ChargerSConns) {
		d.Chargers_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ChargerSConns))
	}
	if !slices.Equal(v1.ResourceSConns, v2.ResourceSConns) {
		d.Resources_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ResourceSConns))
	}
	if !slices.Equal(v1.ThresholdSConns, v2.ThresholdSConns) {
		d.Thresholds_conns = utils.SliceStringPointer(getInternalJSONConns(v2.ThresholdSConns))
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.Stats_conns = utils.SliceStringPointer(getInternalJSONConns(v2.StatSConns))
	}
	if !slices.Equal(v1.RouteSConns, v2.RouteSConns) {
		d.Routes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RouteSConns))
	}
	if !slices.Equal(v1.AttributeSConns, v2.AttributeSConns) {
		d.Cdrs_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	if !slices.Equal(v1.CDRsConns, v2.CDRsConns) {
		d.Replication_conns = utils.SliceStringPointer(getInternalJSONConns(v2.CDRsConns))
	}
	if !slices.Equal(v1.ReplicationConns, v2.ReplicationConns) {
		d.Attributes_conns = utils.SliceStringPointer(v2.ReplicationConns)
	}
	if !slices.Equal(v1.RateSConns, v2.RateSConns) {
		d.Rates_conns = utils.SliceStringPointer(getInternalJSONConns(v2.RateSConns))
	}
	if !slices.Equal(v1.AccountSConns, v2.AccountSConns) {
		d.Accounts_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AccountSConns))
	}
	if v1.StoreSCosts != v2.StoreSCosts {
		d.Store_session_costs = utils.BoolPointer(v2.StoreSCosts)
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
	if !slices.Equal(v1.ActionSConns, v2.ActionSConns) {
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
	d.Opts = diffSessionsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
