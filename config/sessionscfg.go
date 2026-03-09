/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"strconv"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

const (
	SessionsAccountsDftOpt               = false
	SessionsRatesDftOpt                  = false
	SessionsAttributesDftOpt             = false
	SessionsCDRsDftOpt                   = false
	SessionsChargersDftOpt               = false
	SessionsResourcesDftOpt              = false
	SessionsIPsDftOpt                    = false
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
	SessionsIPsAuthorizeDftOpt           = false
	SessionsIPsAllocateDftOpt            = false
	SessionsIPsReleaseDftOpt             = false
	SessionsRoutesDerivedReplyDftOpt     = false
	SessionsStatsDerivedReplyDftOpt      = false
	SessionsThresholdsDerivedReplyDftOpt = false
	SessionsMaxUsageDftOpt               = false
	SessionsTTLDftOpt                    = 0
	SessionsChargeableDftOpt             = true
	SessionsTTLMaxDelayDftOpt            = 0
	SessionsDebitIntervalDftOpt          = 0
)

type SessionsOpts struct {
	Accounts               []*DynamicBoolOpt
	Rates                  []*DynamicBoolOpt
	Attributes             []*DynamicBoolOpt
	CDRs                   []*DynamicBoolOpt
	Chargers               []*DynamicBoolOpt
	Resources              []*DynamicBoolOpt
	IPs                    []*DynamicBoolOpt
	Routes                 []*DynamicBoolOpt
	Stats                  []*DynamicBoolOpt
	Thresholds             []*DynamicBoolOpt
	Authorize              []*DynamicBoolOpt
	Initiate               []*DynamicBoolOpt
	Update                 []*DynamicBoolOpt
	Terminate              []*DynamicBoolOpt
	Message                []*DynamicBoolOpt
	AttributesDerivedReply []*DynamicBoolOpt
	BlockerError           []*DynamicBoolOpt
	CDRsDerivedReply       []*DynamicBoolOpt
	ResourcesAuthorize     []*DynamicBoolOpt
	ResourcesAllocate      []*DynamicBoolOpt
	ResourcesRelease       []*DynamicBoolOpt
	ResourcesDerivedReply  []*DynamicBoolOpt
	ResourcesUsageID       []*DynamicStringOpt
	ResourcesUnits         []*DynamicIntOpt
	IPsAuthorize           []*DynamicBoolOpt
	IPsAllocate            []*DynamicBoolOpt
	IPsRelease             []*DynamicBoolOpt
	RoutesDerivedReply     []*DynamicBoolOpt
	StatsDerivedReply      []*DynamicBoolOpt
	ThresholdsDerivedReply []*DynamicBoolOpt
	MaxUsage               []*DynamicBoolOpt
	ForceUsage             []*DynamicBoolOpt
	TTL                    []*DynamicDurationOpt
	Chargeable             []*DynamicBoolOpt
	TTLLastUsage           []*DynamicDurationPointerOpt
	TTLLastUsed            []*DynamicDurationPointerOpt
	DebitInterval          []*DynamicDurationOpt
	TTLMaxDelay            []*DynamicDurationOpt
	TTLUsage               []*DynamicDurationPointerOpt
	OriginID               []*DynamicStringOpt
	AccountsForceUsage     []*DynamicBoolOpt
	AccountsAuthorize      []*DynamicBoolOpt
	AccountsInitialize     []*DynamicBoolOpt
	AccountsUpdate         []*DynamicBoolOpt
	AccountsTerminate      []*DynamicBoolOpt
}

// SessionSCfg is the config section for SessionS
type SessionSCfg struct {
	Enabled             bool
	ListenBiJSON        string
	ListenBiGob         string
	StoreSCosts         bool
	SessionIndexes      utils.StringSet
	ClientProtocol      float64
	ChannelSyncInterval time.Duration
	TerminateAttempts   int
	AlterableFields     utils.StringSet
	MinDurLowBalance    time.Duration
	Conns               map[string][]*DynamicStringSliceOpt
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

func (sesOpts *SessionsOpts) loadFromJSONCfg(jsnCfg *SessionsOptsJson) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Accounts != nil {
		opt, err := IfaceToBoolDynamicOpts(jsnCfg.Accounts)
		if err != nil {
			return err
		}
		sesOpts.Accounts = append(opt, sesOpts.Accounts...)
	}
	if jsnCfg.Rates != nil {
		opt, err := IfaceToBoolDynamicOpts(jsnCfg.Rates)
		if err != nil {
			return err
		}
		sesOpts.Rates = append(opt, sesOpts.Rates...)
	}
	if jsnCfg.Attributes != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Attributes)
		if err != nil {
			return err
		}
		sesOpts.Attributes = append(opts, sesOpts.Attributes...)
	}
	if jsnCfg.CDRs != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.CDRs)
		if err != nil {
			return err
		}
		sesOpts.CDRs = append(opts, sesOpts.CDRs...)
	}
	if jsnCfg.Chargers != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Chargers)
		if err != nil {
			return err
		}
		sesOpts.Chargers = append(opts, sesOpts.Chargers...)
	}
	if jsnCfg.Resources != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Resources)
		if err != nil {
			return err
		}
		sesOpts.Resources = append(opts, sesOpts.Resources...)
	}
	if jsnCfg.IPs != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.IPs)
		if err != nil {
			return err
		}
		sesOpts.IPs = append(opts, sesOpts.IPs...)
	}
	if jsnCfg.Routes != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Routes)
		if err != nil {
			return err
		}
		sesOpts.Routes = append(opts, sesOpts.Routes...)
	}
	if jsnCfg.Stats != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Stats)
		if err != nil {
			return err
		}
		sesOpts.Stats = append(opts, sesOpts.Stats...)
	}
	if jsnCfg.Thresholds != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Thresholds)
		if err != nil {
			return err
		}
		sesOpts.Thresholds = append(opts, sesOpts.Thresholds...)
	}
	if jsnCfg.Initiate != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Initiate)
		if err != nil {
			return err
		}
		sesOpts.Initiate = append(opts, sesOpts.Initiate...)
	}
	if jsnCfg.Update != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Update)
		if err != nil {
			return err
		}
		sesOpts.Update = append(opts, sesOpts.Update...)
	}
	if jsnCfg.Terminate != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Terminate)
		if err != nil {
			return err
		}
		sesOpts.Terminate = append(opts, sesOpts.Terminate...)
	}
	if jsnCfg.Message != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Message)
		if err != nil {
			return err
		}
		sesOpts.Message = append(opts, sesOpts.Message...)
	}
	if jsnCfg.AttributesDerivedReply != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.AttributesDerivedReply)
		if err != nil {
			return err
		}
		sesOpts.AttributesDerivedReply = append(opts, sesOpts.AttributesDerivedReply...)
	}
	if jsnCfg.BlockerError != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.BlockerError)
		if err != nil {
			return err
		}
		sesOpts.BlockerError = append(opts, sesOpts.BlockerError...)
	}
	if jsnCfg.CDRsDerivedReply != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.CDRsDerivedReply)
		if err != nil {
			return err
		}
		sesOpts.CDRsDerivedReply = append(opts, sesOpts.CDRsDerivedReply...)
	}
	if jsnCfg.ResourcesAuthorize != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.ResourcesAuthorize)
		if err != nil {
			return err
		}
		sesOpts.ResourcesAuthorize = append(opts, sesOpts.ResourcesAuthorize...)
	}
	if jsnCfg.ResourcesAllocate != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.ResourcesAllocate)
		if err != nil {
			return err
		}
		sesOpts.ResourcesAllocate = append(opts, sesOpts.ResourcesAllocate...)
	}
	if jsnCfg.ResourcesRelease != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.ResourcesRelease)
		if err != nil {
			return err
		}
		sesOpts.ResourcesRelease = append(opts, sesOpts.ResourcesRelease...)
	}
	if jsnCfg.ResourcesDerivedReply != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.ResourcesDerivedReply)
		if err != nil {
			return err
		}
		sesOpts.ResourcesDerivedReply = append(opts, sesOpts.ResourcesDerivedReply...)
	}
	if jsnCfg.IPsAuthorize != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.IPsAuthorize)
		if err != nil {
			return err
		}
		sesOpts.IPsAuthorize = append(opts, sesOpts.IPsAuthorize...)
	}
	if jsnCfg.IPsAllocate != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.IPsAllocate)
		if err != nil {
			return err
		}
		sesOpts.IPsAllocate = append(opts, sesOpts.IPsAllocate...)
	}
	if jsnCfg.IPsRelease != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.IPsRelease)
		if err != nil {
			return err
		}
		sesOpts.IPsRelease = append(opts, sesOpts.IPsRelease...)
	}
	if jsnCfg.RoutesDerivedReply != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.RoutesDerivedReply)
		if err != nil {
			return err
		}
		sesOpts.RoutesDerivedReply = append(opts, sesOpts.RoutesDerivedReply...)
	}
	if jsnCfg.StatsDerivedReply != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.StatsDerivedReply)
		if err != nil {
			return err
		}
		sesOpts.StatsDerivedReply = append(opts, sesOpts.StatsDerivedReply...)
	}
	if jsnCfg.ThresholdsDerivedReply != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.ThresholdsDerivedReply)
		if err != nil {
			return err
		}
		sesOpts.ThresholdsDerivedReply = append(opts, sesOpts.ThresholdsDerivedReply...)
	}
	if jsnCfg.MaxUsage != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.MaxUsage)
		if err != nil {
			return err
		}
		sesOpts.MaxUsage = append(opts, sesOpts.MaxUsage...)
	}
	if jsnCfg.ForceUsage != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.ForceUsage)
		if err != nil {
			return err
		}
		sesOpts.ForceUsage = append(opts, sesOpts.ForceUsage...)
	}
	if jsnCfg.TTL != nil {
		ttl, err := IfaceToDurationDynamicOpts(jsnCfg.TTL)
		if err != nil {
			return err
		}
		sesOpts.TTL = append(ttl, sesOpts.TTL...)
	}
	if jsnCfg.Chargeable != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.Chargeable)
		if err != nil {
			return err
		}
		sesOpts.Chargeable = append(opts, sesOpts.Chargeable...)
	}
	if jsnCfg.TTLLastUsage != nil {
		lastUsage, err := IfaceToDurationPointerDynamicOpts(jsnCfg.TTLLastUsage)
		if err != nil {
			return err
		}
		sesOpts.TTLLastUsage = append(lastUsage, sesOpts.TTLLastUsage...)
	}
	if jsnCfg.TTLLastUsed != nil {
		lastUsed, err := IfaceToDurationPointerDynamicOpts(jsnCfg.TTLLastUsed)
		if err != nil {
			return err
		}
		sesOpts.TTLLastUsed = append(lastUsed, sesOpts.TTLLastUsed...)
	}
	if jsnCfg.DebitInterval != nil {
		debitInterval, err := IfaceToDurationDynamicOpts(jsnCfg.DebitInterval)
		if err != nil {
			return err
		}
		sesOpts.DebitInterval = append(debitInterval, sesOpts.DebitInterval...)
	}
	if jsnCfg.TTLMaxDelay != nil {
		maxDelay, err := IfaceToDurationDynamicOpts(jsnCfg.TTLMaxDelay)
		if err != nil {
			return err
		}
		sesOpts.TTLMaxDelay = append(maxDelay, sesOpts.TTLMaxDelay...)
	}
	if jsnCfg.TTLUsage != nil {
		usage, err := IfaceToDurationPointerDynamicOpts(jsnCfg.TTLUsage)
		if err != nil {
			return err
		}
		sesOpts.TTLUsage = append(usage, sesOpts.TTLUsage...)
	}
	if jsnCfg.OriginID != nil {
		originID, err := InterfaceToDynamicStringOpts(jsnCfg.OriginID)
		if err != nil {
			return err
		}
		sesOpts.OriginID = append(originID, sesOpts.OriginID...)
	}
	if jsnCfg.AccountsForceUsage != nil {
		opts, err := IfaceToBoolDynamicOpts(jsnCfg.AccountsForceUsage)
		if err != nil {
			return err
		}
		sesOpts.AccountsForceUsage = append(opts, sesOpts.AccountsForceUsage...)
	}
	return nil
}

func (scfg *SessionSCfg) loadFromJSONCfg(jsnCfg *SessionSJsonCfg) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		scfg.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.ListenBiJSON != nil {
		scfg.ListenBiJSON = *jsnCfg.ListenBiJSON
	}
	if jsnCfg.ListenBiGob != nil {
		scfg.ListenBiGob = *jsnCfg.ListenBiGob
	}
	if jsnCfg.StoreSCosts != nil {
		scfg.StoreSCosts = *jsnCfg.StoreSCosts
	}
	if jsnCfg.SessionIndexes != nil {
		scfg.SessionIndexes = utils.NewStringSet(*jsnCfg.SessionIndexes)
	}
	if jsnCfg.ClientProtocol != nil {
		scfg.ClientProtocol = *jsnCfg.ClientProtocol
	}
	if jsnCfg.ChannelSyncInterval != nil {
		if scfg.ChannelSyncInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.ChannelSyncInterval); err != nil {
			return err
		}
	}
	if jsnCfg.TerminateAttempts != nil {
		scfg.TerminateAttempts = *jsnCfg.TerminateAttempts
	}
	if jsnCfg.AlterableFields != nil {
		scfg.AlterableFields = utils.NewStringSet(*jsnCfg.AlterableFields)
	}
	if jsnCfg.MinDurLowBalance != nil {
		if scfg.MinDurLowBalance, err = utils.ParseDurationWithNanosecs(*jsnCfg.MinDurLowBalance); err != nil {
			return err
		}
	}
	if jsnCfg.DefaultUsage != nil {
		for k, v := range jsnCfg.DefaultUsage {
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
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			scfg.Conns[connType] = opts
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
func (scfg SessionSCfg) AsMapInterface() any {
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
		utils.MetaRates:                     scfg.Opts.Rates,
		utils.MetaAttributes:                scfg.Opts.Attributes,
		utils.MetaCDRs:                      scfg.Opts.CDRs,
		utils.MetaChargers:                  scfg.Opts.Chargers,
		utils.MetaResources:                 scfg.Opts.Resources,
		utils.MetaIPs:                       scfg.Opts.IPs,
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
		utils.MetaIPsAuthorizeCfg:           scfg.Opts.IPsAuthorize,
		utils.MetaIPsAllocateCfg:            scfg.Opts.IPsAllocate,
		utils.MetaIPsReleaseCfg:             scfg.Opts.IPsRelease,
		utils.MetaRoutesDerivedReplyCfg:     scfg.Opts.RoutesDerivedReply,
		utils.MetaStatsDerivedReplyCfg:      scfg.Opts.StatsDerivedReply,
		utils.MetaThresholdsDerivedReplyCfg: scfg.Opts.ThresholdsDerivedReply,
		utils.MetaMaxUsageCfg:               scfg.Opts.MaxUsage,
		utils.MetaForceUsageCfg:             scfg.Opts.ForceUsage,
		utils.MetaTTLCfg:                    scfg.Opts.TTL,
		utils.MetaChargeableCfg:             scfg.Opts.Chargeable,
		utils.MetaDebitIntervalCfg:          scfg.Opts.DebitInterval,
		utils.MetaTTLLastUsageCfg:           scfg.Opts.TTLLastUsage,
		utils.MetaTTLLastUsedCfg:            scfg.Opts.TTLLastUsed,
		utils.MetaTTLMaxDelayCfg:            scfg.Opts.TTLMaxDelay,
		utils.MetaTTLUsageCfg:               scfg.Opts.TTLUsage,
		utils.MetaOriginID:                  scfg.Opts.OriginID,
		utils.MetaAccountsForceUsage:        scfg.Opts.AccountsForceUsage,
	}
	mp := map[string]any{
		utils.EnabledCfg:             scfg.Enabled,
		utils.ListenBijsonCfg:        scfg.ListenBiJSON,
		utils.ListenBigobCfg:         scfg.ListenBiGob,
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
		utils.ConnsCfg:               stripConns(scfg.Conns),
	}
	if scfg.ChannelSyncInterval != 0 {
		mp[utils.ChannelSyncIntervalCfg] = scfg.ChannelSyncInterval.String()
	}
	if scfg.MinDurLowBalance != 0 {
		mp[utils.MinDurLowBalanceCfg] = scfg.MinDurLowBalance.String()
	}
	return mp
}

func (SessionSCfg) SName() string              { return SessionSJSON }
func (scfg SessionSCfg) CloneSection() Section { return scfg.Clone() }

func (o *SessionsOpts) Clone() *SessionsOpts {
	return &SessionsOpts{
		Accounts:               CloneDynamicBoolOpt(o.Accounts),
		Rates:                  CloneDynamicBoolOpt(o.Rates),
		Attributes:             CloneDynamicBoolOpt(o.Attributes),
		CDRs:                   CloneDynamicBoolOpt(o.CDRs),
		Chargers:               CloneDynamicBoolOpt(o.Chargers),
		Resources:              CloneDynamicBoolOpt(o.Resources),
		IPs:                    CloneDynamicBoolOpt(o.IPs),
		Routes:                 CloneDynamicBoolOpt(o.Routes),
		Stats:                  CloneDynamicBoolOpt(o.Stats),
		Thresholds:             CloneDynamicBoolOpt(o.Thresholds),
		Initiate:               CloneDynamicBoolOpt(o.Initiate),
		Update:                 CloneDynamicBoolOpt(o.Update),
		Terminate:              CloneDynamicBoolOpt(o.Terminate),
		Message:                CloneDynamicBoolOpt(o.Message),
		AttributesDerivedReply: CloneDynamicBoolOpt(o.AttributesDerivedReply),
		BlockerError:           CloneDynamicBoolOpt(o.BlockerError),
		CDRsDerivedReply:       CloneDynamicBoolOpt(o.CDRsDerivedReply),
		ResourcesAuthorize:     CloneDynamicBoolOpt(o.ResourcesAuthorize),
		ResourcesAllocate:      CloneDynamicBoolOpt(o.ResourcesAllocate),
		ResourcesRelease:       CloneDynamicBoolOpt(o.ResourcesRelease),
		ResourcesDerivedReply:  CloneDynamicBoolOpt(o.ResourcesDerivedReply),
		IPsAuthorize:           CloneDynamicBoolOpt(o.IPsAuthorize),
		IPsAllocate:            CloneDynamicBoolOpt(o.IPsAllocate),
		IPsRelease:             CloneDynamicBoolOpt(o.IPsRelease),
		RoutesDerivedReply:     CloneDynamicBoolOpt(o.RoutesDerivedReply),
		StatsDerivedReply:      CloneDynamicBoolOpt(o.StatsDerivedReply),
		ThresholdsDerivedReply: CloneDynamicBoolOpt(o.ThresholdsDerivedReply),
		MaxUsage:               CloneDynamicBoolOpt(o.MaxUsage),
		ForceUsage:             CloneDynamicBoolOpt(o.ForceUsage),
		TTL:                    CloneDynamicDurationOpt(o.TTL),
		Chargeable:             CloneDynamicBoolOpt(o.Chargeable),
		DebitInterval:          CloneDynamicDurationOpt(o.DebitInterval),
		TTLLastUsage:           CloneDynamicDurationPointerOpt(o.TTLLastUsage),
		TTLLastUsed:            CloneDynamicDurationPointerOpt(o.TTLLastUsed),
		TTLMaxDelay:            CloneDynamicDurationOpt(o.TTLMaxDelay),
		TTLUsage:               CloneDynamicDurationPointerOpt(o.TTLUsage),
		OriginID:               CloneDynamicStringOpt(o.OriginID),
		AccountsForceUsage:     CloneDynamicBoolOpt(o.AccountsForceUsage),
	}
}

// Clone returns a deep copy of SessionSCfg
func (scfg SessionSCfg) Clone() (cln *SessionSCfg) {
	cln = &SessionSCfg{
		Enabled:             scfg.Enabled,
		ListenBiJSON:        scfg.ListenBiJSON,
		StoreSCosts:         scfg.StoreSCosts,
		ClientProtocol:      scfg.ClientProtocol,
		ChannelSyncInterval: scfg.ChannelSyncInterval,
		TerminateAttempts:   scfg.TerminateAttempts,
		MinDurLowBalance:    scfg.MinDurLowBalance,
		Conns:               CloneConnsOpt(scfg.Conns),
		SessionIndexes:      scfg.SessionIndexes.Clone(),
		AlterableFields:     scfg.AlterableFields.Clone(),
		STIRCfg:             scfg.STIRCfg.Clone(),
		DefaultUsage:        make(map[string]time.Duration),
		Opts:                scfg.Opts.Clone(),
	}
	for k, v := range scfg.DefaultUsage {
		cln.DefaultUsage[k] = v
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
	Accounts               []*DynamicInterfaceOpt `json:"*accounts"`
	Rates                  []*DynamicInterfaceOpt `json:"*rates"`
	Attributes             []*DynamicInterfaceOpt `json:"*attributes"`
	CDRs                   []*DynamicInterfaceOpt `json:"*cdrs"`
	Chargers               []*DynamicInterfaceOpt `json:"*chargers"`
	Resources              []*DynamicInterfaceOpt `json:"*resources"`
	IPs                    []*DynamicInterfaceOpt `json:"*ips"`
	Routes                 []*DynamicInterfaceOpt `json:"*routes"`
	Stats                  []*DynamicInterfaceOpt `json:"*stats"`
	Thresholds             []*DynamicInterfaceOpt `json:"*thresholds"`
	Initiate               []*DynamicInterfaceOpt `json:"*initiate"`
	Update                 []*DynamicInterfaceOpt `json:"*update"`
	Terminate              []*DynamicInterfaceOpt `json:"*terminate"`
	Message                []*DynamicInterfaceOpt `json:"*message"`
	AttributesDerivedReply []*DynamicInterfaceOpt `json:"*attributesDerivedReply"`
	BlockerError           []*DynamicInterfaceOpt `json:"*blockerError"`
	CDRsDerivedReply       []*DynamicInterfaceOpt `json:"*cdrsDerivedReply"`
	ResourcesAuthorize     []*DynamicInterfaceOpt `json:"*resourcesAuthorize"`
	ResourcesAllocate      []*DynamicInterfaceOpt `json:"*resourcesAllocate"`
	ResourcesRelease       []*DynamicInterfaceOpt `json:"*resourcesRelease"`
	ResourcesDerivedReply  []*DynamicInterfaceOpt `json:"*resourcesDerivedReply"`
	IPsAuthorize           []*DynamicInterfaceOpt `json:"*ipsAuthorize"`
	IPsAllocate            []*DynamicInterfaceOpt `json:"*ipsAllocate"`
	IPsRelease             []*DynamicInterfaceOpt `json:"*ipsRelease"`
	RoutesDerivedReply     []*DynamicInterfaceOpt `json:"*routesDerivedReply"`
	StatsDerivedReply      []*DynamicInterfaceOpt `json:"*statsDerivedReply"`
	ThresholdsDerivedReply []*DynamicInterfaceOpt `json:"*thresholdsDerivedReply"`
	MaxUsage               []*DynamicInterfaceOpt `json:"*maxUsage"`
	ForceUsage             []*DynamicInterfaceOpt `json:"*forceUsage"`
	TTL                    []*DynamicInterfaceOpt `json:"*ttl"`
	Chargeable             []*DynamicInterfaceOpt `json:"*chargeable"`
	DebitInterval          []*DynamicInterfaceOpt `json:"*debitInterval"`
	TTLLastUsage           []*DynamicInterfaceOpt `json:"*ttlLastUsage"`
	TTLLastUsed            []*DynamicInterfaceOpt `json:"*ttlLastUsed"`
	TTLMaxDelay            []*DynamicInterfaceOpt `json:"*ttlMaxDelay"`
	TTLUsage               []*DynamicInterfaceOpt `json:"*ttlUsage"`
	OriginID               []*DynamicInterfaceOpt `json:"*originID"`
	AccountsForceUsage     []*DynamicInterfaceOpt `json:"*accountsForceUsage"`
}

// SessionSJsonCfg config section
type SessionSJsonCfg struct {
	Enabled             *bool                               `json:"enabled"`
	ListenBiJSON        *string                             `json:"listen_bijson"`
	ListenBiGob         *string                             `json:"listen_bigob"`
	StoreSCosts         *bool                               `json:"store_session_costs"`
	SessionIndexes      *[]string                           `json:"session_indexes"`
	ClientProtocol      *float64                            `json:"client_protocol"`
	ChannelSyncInterval *string                             `json:"channel_sync_interval"`
	TerminateAttempts   *int                                `json:"terminate_attempts"`
	AlterableFields     *[]string                           `json:"alterable_fields"`
	MinDurLowBalance    *string                             `json:"min_dur_low_balance"`
	Stir                *STIRJsonCfg                        `json:"stir"`
	DefaultUsage        map[string]string                   `json:"default_usage"`
	Conns               map[string][]*DynamicStringSliceOpt `json:"conns,omitempty"`
	Opts                *SessionsOptsJson                   `json:"opts"`
}

func diffSessionsOptsJsonCfg(d *SessionsOptsJson, v1, v2 *SessionsOpts) *SessionsOptsJson {
	if d == nil {
		d = new(SessionsOptsJson)
	}
	if !DynamicBoolOptEqual(v1.Accounts, v2.Accounts) {
		d.Accounts = BoolToIfaceDynamicOpts(v2.Accounts)
	}
	if !DynamicBoolOptEqual(v1.Rates, v2.Rates) {
		d.Rates = BoolToIfaceDynamicOpts(v2.Rates)
	}
	if !DynamicBoolOptEqual(v1.Attributes, v2.Attributes) {
		d.Attributes = BoolToIfaceDynamicOpts(v2.Attributes)
	}
	if !DynamicBoolOptEqual(v1.CDRs, v2.CDRs) {
		d.CDRs = BoolToIfaceDynamicOpts(v2.CDRs)
	}
	if !DynamicBoolOptEqual(v1.Chargers, v2.Chargers) {
		d.Chargers = BoolToIfaceDynamicOpts(v2.Chargers)
	}
	if !DynamicBoolOptEqual(v1.Resources, v2.Resources) {
		d.Resources = BoolToIfaceDynamicOpts(v2.Resources)
	}
	if !DynamicBoolOptEqual(v1.IPs, v2.IPs) {
		d.IPs = BoolToIfaceDynamicOpts(v2.IPs)
	}
	if !DynamicBoolOptEqual(v1.Routes, v2.Routes) {
		d.Routes = BoolToIfaceDynamicOpts(v2.Routes)
	}
	if !DynamicBoolOptEqual(v1.Stats, v2.Stats) {
		d.Stats = BoolToIfaceDynamicOpts(v2.Stats)
	}
	if !DynamicBoolOptEqual(v1.Thresholds, v2.Thresholds) {
		d.Thresholds = BoolToIfaceDynamicOpts(v2.Thresholds)
	}
	if !DynamicBoolOptEqual(v1.Initiate, v2.Initiate) {
		d.Initiate = BoolToIfaceDynamicOpts(v2.Initiate)
	}
	if !DynamicBoolOptEqual(v1.Update, v2.Update) {
		d.Update = BoolToIfaceDynamicOpts(v2.Update)
	}
	if !DynamicBoolOptEqual(v1.Terminate, v2.Terminate) {
		d.Terminate = BoolToIfaceDynamicOpts(v2.Terminate)
	}
	if !DynamicBoolOptEqual(v1.Message, v2.Message) {
		d.Message = BoolToIfaceDynamicOpts(v2.Message)
	}
	if !DynamicBoolOptEqual(v1.AttributesDerivedReply, v2.AttributesDerivedReply) {
		d.AttributesDerivedReply = BoolToIfaceDynamicOpts(v2.AttributesDerivedReply)
	}
	if !DynamicBoolOptEqual(v1.BlockerError, v2.BlockerError) {
		d.BlockerError = BoolToIfaceDynamicOpts(v2.BlockerError)
	}
	if !DynamicBoolOptEqual(v1.CDRsDerivedReply, v2.CDRsDerivedReply) {
		d.CDRsDerivedReply = BoolToIfaceDynamicOpts(v2.CDRsDerivedReply)
	}
	if !DynamicBoolOptEqual(v1.ResourcesAuthorize, v2.ResourcesAuthorize) {
		d.ResourcesAuthorize = BoolToIfaceDynamicOpts(v2.ResourcesAuthorize)
	}
	if !DynamicBoolOptEqual(v1.ResourcesAllocate, v2.ResourcesAllocate) {
		d.ResourcesAllocate = BoolToIfaceDynamicOpts(v2.ResourcesAllocate)
	}
	if !DynamicBoolOptEqual(v1.ResourcesRelease, v2.ResourcesRelease) {
		d.ResourcesRelease = BoolToIfaceDynamicOpts(v2.ResourcesRelease)
	}
	if !DynamicBoolOptEqual(v1.ResourcesDerivedReply, v2.ResourcesDerivedReply) {
		d.ResourcesDerivedReply = BoolToIfaceDynamicOpts(v2.ResourcesDerivedReply)
	}
	if !DynamicBoolOptEqual(v1.IPsAuthorize, v2.IPsAuthorize) {
		d.IPsAuthorize = BoolToIfaceDynamicOpts(v2.IPsAuthorize)
	}
	if !DynamicBoolOptEqual(v1.IPsAllocate, v2.IPsAllocate) {
		d.IPsAllocate = BoolToIfaceDynamicOpts(v2.IPsAllocate)
	}
	if !DynamicBoolOptEqual(v1.IPsRelease, v2.IPsRelease) {
		d.IPsRelease = BoolToIfaceDynamicOpts(v2.IPsRelease)
	}
	if !DynamicBoolOptEqual(v1.RoutesDerivedReply, v2.RoutesDerivedReply) {
		d.RoutesDerivedReply = BoolToIfaceDynamicOpts(v2.RoutesDerivedReply)
	}
	if !DynamicBoolOptEqual(v1.StatsDerivedReply, v2.StatsDerivedReply) {
		d.StatsDerivedReply = BoolToIfaceDynamicOpts(v2.StatsDerivedReply)
	}
	if !DynamicBoolOptEqual(v1.ThresholdsDerivedReply, v2.ThresholdsDerivedReply) {
		d.ThresholdsDerivedReply = BoolToIfaceDynamicOpts(v2.ThresholdsDerivedReply)
	}
	if !DynamicBoolOptEqual(v1.MaxUsage, v2.MaxUsage) {
		d.MaxUsage = BoolToIfaceDynamicOpts(v2.MaxUsage)
	}
	if !DynamicBoolOptEqual(v1.ForceUsage, v2.ForceUsage) {
		d.ForceUsage = BoolToIfaceDynamicOpts(v2.ForceUsage)
	}
	if !DynamicDurationOptEqual(v1.TTL, v2.TTL) {
		d.TTL = DurationToIfaceDynamicOpts(v2.TTL)
	}
	if !DynamicBoolOptEqual(v1.Chargeable, v2.Chargeable) {
		d.Chargeable = BoolToIfaceDynamicOpts(v2.Chargeable)
	}
	if !DynamicDurationPointerOptEqual(v1.TTLLastUsage, v2.TTLLastUsage) {
		d.TTLLastUsage = DurationPointerToIfaceDynamicOpts(v2.TTLLastUsage)
	}
	if !DynamicDurationPointerOptEqual(v1.TTLLastUsed, v2.TTLLastUsed) {
		d.TTLLastUsed = DurationPointerToIfaceDynamicOpts(v2.TTLLastUsed)
	}
	if !DynamicDurationOptEqual(v1.DebitInterval, v2.DebitInterval) {
		d.DebitInterval = DurationToIfaceDynamicOpts(v2.DebitInterval)
	}
	if !DynamicDurationOptEqual(v1.TTLMaxDelay, v2.TTLMaxDelay) {
		d.TTLMaxDelay = DurationToIfaceDynamicOpts(v2.TTLMaxDelay)
	}
	if !DynamicDurationPointerOptEqual(v1.TTLUsage, v2.TTLUsage) {
		d.TTLUsage = DurationPointerToIfaceDynamicOpts(v2.TTLUsage)
	}
	if !DynamicStringOptEqual(v1.OriginID, v2.OriginID) {
		d.OriginID = DynamicStringToInterfaceOpts(v2.OriginID)
	}
	if !DynamicBoolOptEqual(v1.AccountsForceUsage, v2.AccountsForceUsage) {
		d.AccountsForceUsage = BoolToIfaceDynamicOpts(v2.AccountsForceUsage)
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
	if v1.ListenBiJSON != v2.ListenBiJSON {
		d.ListenBiJSON = utils.StringPointer(v2.ListenBiJSON)
	}
	if v1.ListenBiGob != v2.ListenBiGob {
		d.ListenBiGob = utils.StringPointer(v2.ListenBiGob)
	}
	if v1.StoreSCosts != v2.StoreSCosts {
		d.StoreSCosts = utils.BoolPointer(v2.StoreSCosts)
	}
	if !v1.SessionIndexes.Equals(v2.SessionIndexes) {
		d.SessionIndexes = utils.SliceStringPointer(v2.SessionIndexes.AsSlice())
	}
	if v1.ClientProtocol != v2.ClientProtocol {
		d.ClientProtocol = utils.Float64Pointer(v2.ClientProtocol)
	}
	if v1.ChannelSyncInterval != v2.ChannelSyncInterval {
		d.ChannelSyncInterval = utils.StringPointer(v2.ChannelSyncInterval.String())
	}
	if v1.TerminateAttempts != v2.TerminateAttempts {
		d.TerminateAttempts = utils.IntPointer(v2.TerminateAttempts)
	}
	if !v1.AlterableFields.Equals(v2.AlterableFields) {
		d.AlterableFields = utils.SliceStringPointer(v2.AlterableFields.AsSlice())
	}
	if v1.MinDurLowBalance != v2.MinDurLowBalance {
		d.MinDurLowBalance = utils.StringPointer(v2.MinDurLowBalance.String())
	}
	d.Stir = diffSTIRJsonCfg(d.Stir, v1.STIRCfg, v2.STIRCfg)
	if d.DefaultUsage == nil {
		d.DefaultUsage = make(map[string]string)
	}
	for tor, usage2 := range v2.DefaultUsage {
		if usage1, has := v1.DefaultUsage[tor]; !has || usage1 != usage2 {
			d.DefaultUsage[tor] = usage2.String()
		}
	}
	if !ConnsEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	d.Opts = diffSessionsOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
