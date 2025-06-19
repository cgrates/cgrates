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

package actions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// parseParamStringToMap parses a string containing key-value pairs separated by "&" and assigns
// these pairs to a given map. Each pair is expected to be in the format "key:value".
func parseParamStringToMap(paramStr string, targetMap map[string]any) error {
	for tuple := range strings.SplitSeq(paramStr, utils.ANDSep) {
		// Use strings.Cut to split 'tuple' into key-value pairs at the first occurrence of ':'.
		// This ensures that additional ':' characters within the value do not affect parsing.
		key, value, found := strings.Cut(tuple, utils.InInFieldSep)
		if !found {
			return fmt.Errorf("invalid key-value pair: %s", tuple)
		}
		targetMap[key] = value
	}
	return nil
}

// actDynamicThreshold processes the `ActionValue` field from the action to construct a Threshold profile
//
// The ActionValue field format is expected as follows:
//
//	 0 Tenant: string
//	 1 ID: string
//	 2 FilterIDs: strings separated by "&".
//	 3 Weight: strings separated by "&". Should be higher than the threshold weight that
//			   triggers this action
//	 4 MaxHits: integer
//	 5 MinHits: integer
//	 6 MinSleep: duration
//	 7 Blocker: bool, should always be true
//	 8 ActionProfileIDs: strings separated by "&".
//	 9 Async: bool
//	10 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicThreshold struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicThreshold) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicThreshold) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicThreshold) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
	if len(aL.config.ActionSCfg().AdminSConns) == 0 {
		return fmt.Errorf("no connection with AdminS")
	}
	data[utils.MetaNow] = time.Now()
	data[utils.MetaTenant] = utils.FirstNonEmpty(aL.cgrEv.Tenant, aL.tnt,
		config.CgrConfig().GeneralCfg().DefaultTenant)
	// Parse action parameters based on the predefined format.
	if len(aL.aCfg.Diktats) == 0 {
		return fmt.Errorf("No diktats were speified for action <%v>", aL.aCfg.ID)
	}
	params := strings.Split(aL.aCfg.Diktats[0].Value, utils.InfieldSep)
	if len(params) != 11 {
		return fmt.Errorf("invalid number of parameters <%d> expected 11", len(params))
	}
	// parse dynamic parameters
	for i := range params {
		if params[i], err = utils.ParseParamForDataProvider(params[i], data, false); err != nil {
			return err
		}
	}
	// Prepare request arguments based on provided parameters.
	args := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: params[0],
			ID:     params[1],
		},
		APIOpts: make(map[string]any),
	}
	// populate Threshold's FilterIDs
	if params[2] != utils.EmptyString {
		args.FilterIDs = strings.Split(params[2], utils.ANDSep)
	}
	// populate Threshold's Weight
	if params[3] != utils.EmptyString {
		args.Weights = utils.DynamicWeights{&utils.DynamicWeight{}}
		wghtSplit := strings.Split(params[3], utils.ANDSep)
		if len(wghtSplit) > 2 {
			return utils.ErrUnsupportedFormat
		}
		if wghtSplit[0] != utils.EmptyString {
			args.Weights[0].FilterIDs = []string{wghtSplit[0]}
		}
		if wghtSplit[1] != utils.EmptyString {
			args.Weights[0].Weight, err = strconv.ParseFloat(wghtSplit[1], 64)
			if err != nil {
				return err
			}
		}
	}
	// populate Threshold's MaxHits
	if params[4] != utils.EmptyString {
		args.MaxHits, err = strconv.Atoi(params[4])
		if err != nil {
			return err
		}
	}
	// populate Threshold's MinHits
	if params[5] != utils.EmptyString {
		args.MinHits, err = strconv.Atoi(params[5])
		if err != nil {
			return err
		}
	}
	// populate Threshold's MinSleep
	if params[6] != utils.EmptyString {
		args.MinSleep, err = utils.ParseDurationWithNanosecs(params[6])
		if err != nil {
			return err
		}
	}
	// populate Threshold's Blocker
	if params[7] != utils.EmptyString {
		args.Blocker, err = strconv.ParseBool(params[7])
		if err != nil {
			return err
		}
	}
	// populate Threshold's ActionProfileIDs
	if params[8] != utils.EmptyString {
		args.ActionProfileIDs = strings.Split(params[8], utils.ANDSep)
	}
	// populate Threshold's Async bool
	if params[9] != utils.EmptyString {
		args.Async, err = strconv.ParseBool(params[9])
		if err != nil {
			return err
		}
	}
	// populate Threshold's APIOpts
	if params[10] != utils.EmptyString {
		if err := parseParamStringToMap(params[10], args.APIOpts); err != nil {
			return err
		}
	}
	// create the ThresholdProfile based on the populated parameters
	var rply string
	return aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
		utils.AdminSv1SetThresholdProfile, args, &rply)
}

// actDynamicStats processes the `ActionValue` field from the action to construct a StatQueueProfile
//
// The ActionValue field format is expected as follows:
//
//	 0 Tenant: string
//	 1 ID: string
//	 2 FilterIDs: strings separated by "&".
//	 3 Weights: strings separated by "&".
//	 4 Blockers: strings separated by "&".
//	 5 QueueLength: integer
//	 6 TTL: duration
//	 7 MinItems: integer
//	 8 Stored: bool
//	 9 ThresholdIDs: strings separated by "&".
//	10 MetricIDs: strings separated by "&".
//	11 MetricFilterIDs: strings separated by "&".
//	12 MetricBlockers: strings separated by "&".
//	13 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicStats struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicStats) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicStats) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicStats) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
	if len(aL.config.ActionSCfg().AdminSConns) == 0 {
		return fmt.Errorf("no connection with AdminS")
	}
	data[utils.MetaNow] = time.Now()
	data[utils.MetaTenant] = utils.FirstNonEmpty(aL.cgrEv.Tenant, aL.tnt,
		config.CgrConfig().GeneralCfg().DefaultTenant)
	// Parse action parameters based on the predefined format.
	if len(aL.aCfg.Diktats) == 0 {
		return fmt.Errorf("No diktats were speified for action <%v>", aL.aCfg.ID)
	}
	params := strings.Split(aL.aCfg.Diktats[0].Value, utils.InfieldSep)
	if len(params) != 14 {
		return fmt.Errorf("invalid number of parameters <%d> expected 14", len(params))
	}
	// parse dynamic parameters
	for i := range params {
		if params[i], err = utils.ParseParamForDataProvider(params[i], data, false); err != nil {
			return err
		}
	}
	// Prepare request arguments based on provided parameters.
	args := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: params[0],
			ID:     params[1],
		},
		APIOpts: make(map[string]any),
	}
	// populate Stat's FilterIDs
	if params[2] != utils.EmptyString {
		args.FilterIDs = strings.Split(params[2], utils.ANDSep)
	}
	// populate Stat's Weights
	if params[3] != utils.EmptyString {
		args.Weights = utils.DynamicWeights{&utils.DynamicWeight{}}
		wghtSplit := strings.Split(params[3], utils.ANDSep)
		if len(wghtSplit) > 2 {
			return utils.ErrUnsupportedFormat
		}
		if wghtSplit[0] != utils.EmptyString {
			args.Weights[0].FilterIDs = []string{wghtSplit[0]}
		}
		if wghtSplit[1] != utils.EmptyString {
			args.Weights[0].Weight, err = strconv.ParseFloat(wghtSplit[1], 64)
			if err != nil {
				return err
			}
		}
	}
	// populate Stat's Blockers
	if params[4] != utils.EmptyString {
		args.Blockers = utils.DynamicBlockers{&utils.DynamicBlocker{}}
		blckrSplit := strings.Split(params[4], utils.ANDSep)
		if len(blckrSplit) > 2 {
			return utils.ErrUnsupportedFormat
		}
		if blckrSplit[0] != utils.EmptyString {
			args.Blockers[0].FilterIDs = []string{blckrSplit[0]}
		}
		if blckrSplit[1] != utils.EmptyString {
			args.Blockers[0].Blocker, err = strconv.ParseBool(blckrSplit[1])
			if err != nil {
				return err
			}
		}
	}
	// populate Stat's QueueLengh
	if params[5] != utils.EmptyString {
		args.QueueLength, err = strconv.Atoi(params[5])
		if err != nil {
			return err
		}
	}
	// populate Stat's TTL
	if params[6] != utils.EmptyString {
		args.TTL, err = utils.ParseDurationWithNanosecs(params[6])
		if err != nil {
			return err
		}
	}
	// populate Stat's MinItems
	if params[7] != utils.EmptyString {
		args.MinItems, err = strconv.Atoi(params[7])
		if err != nil {
			return err
		}
	}
	// populate Stat's Stored
	if params[8] != utils.EmptyString {
		args.Stored, err = strconv.ParseBool(params[8])
		if err != nil {
			return err
		}
	}
	// populate Stat's ThresholdIDs
	if params[9] != utils.EmptyString {
		args.ThresholdIDs = strings.Split(params[9], utils.ANDSep)
	}
	// populate Stat's MetricID
	if params[10] != utils.EmptyString {
		metrics := strings.Split(params[10], utils.ANDSep)
		args.Metrics = make([]*engine.MetricWithFilters, len(metrics))
		for i, strM := range metrics {
			args.Metrics[i] = &engine.MetricWithFilters{MetricID: strM}
		}
	}
	// populate Stat's metricFliterIDs
	if params[11] != utils.EmptyString {
		metricFliters := strings.Split(params[11], utils.ANDSep)
		for i := range args.Metrics {
			args.Metrics[i].FilterIDs = metricFliters
		}
	}
	// populate Stat's metricBlockers
	if params[12] != utils.EmptyString {
		blckrSplit := strings.Split(params[12], utils.ANDSep)
		if len(blckrSplit) > 2 {
			return utils.ErrUnsupportedFormat
		}
		for i := range args.Metrics {
			args.Metrics[i].Blockers = utils.DynamicBlockers{&utils.DynamicBlocker{}}
			if blckrSplit[0] != utils.EmptyString {
				args.Metrics[i].Blockers[0].FilterIDs = []string{blckrSplit[0]}
			}
			if blckrSplit[1] != utils.EmptyString {
				args.Metrics[i].Blockers[0].Blocker, err = strconv.ParseBool(blckrSplit[1])
				if err != nil {
					return err
				}
			}
		}
	}
	// populate Stat's APIOpts
	if params[13] != utils.EmptyString {
		if err := parseParamStringToMap(params[13], args.APIOpts); err != nil {
			return err
		}
	}

	// create the StatQueueProfile based on the populated parameters
	var rply string
	return aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
		utils.AdminSv1SetStatQueueProfile, args, &rply)
}
