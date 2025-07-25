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
	"cmp"
	"fmt"
	"slices"
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
		keyVal := strings.SplitN(tuple, utils.InInFieldSep, 2)
		if len(keyVal) != 2 {
			return fmt.Errorf("invalid key-value pair: %s", tuple)
		}
		targetMap[keyVal[0]] = keyVal[1]
	}
	return nil
}

// actDynamicThreshold processes the `ActionDiktatsOpts` field from the action to construct a Threshold profile
//
// The ActionDiktatsOpts field format is expected as follows:
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
	fltrS   *engine.FilterS
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
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
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetThresholdProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicStats processes the `ActionDiktatsOpts` field from the action to construct a StatQueueProfile
//
// The ActionDiktatsOpts field format is expected as follows:
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
	fltrS   *engine.FilterS
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
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
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetStatQueueProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicAttribute processes the `ActionDiktatsOpts` field from the action to construct a AttributeProfile
//
// The ActionDiktatsOpts field format is expected as follows:
//
//		 0 Tenant: string
//		 1 ID: string
//		 2 FilterIDs: strings separated by "&".
//		 3 Weights: strings separated by "&".
//		 4 Blocker: strings separated by "&".
//	 	 5 AttributeFilterIDs: strings separated by "&".
//	 	 6 AttributeBlockers: strings separated by "&".
//	 	 7 Path: string
//	 	 8 Type: string
//	 	 9 Value: strings separated by "&".
//		10 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicAttribute struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicAttribute) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicAttribute) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicAttribute) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
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
		args := &utils.APIAttributeProfileWithAPIOpts{
			APIAttributeProfile: &utils.APIAttributeProfile{
				Tenant: params[0],
				ID:     params[1],
			},
			APIOpts: make(map[string]any),
		}
		// populate Attribute's FilterIDs
		if params[2] != utils.EmptyString {
			args.FilterIDs = strings.Split(params[2], utils.ANDSep)
		}
		// populate Attribute's Weights
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
		// populate Attribute's Blockers
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
		// populate Attribute's Attributes
		if params[7] != utils.EmptyString {
			var attrFltrIDs []string
			if params[5] != utils.EmptyString {
				attrFltrIDs = strings.Split(params[5], utils.ANDSep)
			}
			var attrFltrBlckrs utils.DynamicBlockers
			if params[6] != utils.EmptyString {
				attrFltrBlckrs = utils.DynamicBlockers{&utils.DynamicBlocker{}}
				blckrSplit := strings.Split(params[6], utils.ANDSep)
				if len(blckrSplit) > 2 {
					return utils.ErrUnsupportedFormat
				}
				if blckrSplit[0] != utils.EmptyString {
					attrFltrBlckrs[0].FilterIDs = []string{blckrSplit[0]}
				}
				if blckrSplit[1] != utils.EmptyString {
					attrFltrBlckrs[0].Blocker, err = strconv.ParseBool(blckrSplit[1])
					if err != nil {
						return err
					}
				}
			}
			args.Attributes = append(args.Attributes, &utils.ExternalAttribute{
				FilterIDs: attrFltrIDs,
				Blockers:  attrFltrBlckrs,
				Path:      params[7],
				Type:      params[8],
				Value:     params[9],
			})
		}
		// populate Attribute's APIOpts
		if params[10] != utils.EmptyString {
			if err := parseParamStringToMap(params[10], args.APIOpts); err != nil {
				return err
			}
		}

		// create the AttributeProfile based on the populated parameters
		var rply string
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetAttributeProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicResource processes the `ActionDiktatsOpts` field from the action to construct a ResourceProfile
//
// The ActionDiktatsOpts field format is expected as follows:
//
//		 0 Tenant: string
//		 1 ID: string
//		 2 FilterIDs: strings separated by "&".
//		 3 Weights: strings separated by "&".
//		 4 TTL: duration
//	 	 5 Limit: float
//	 	 6 AllocationMessage: string
//	 	 7 Blocker: bool
//	 	 8 Stored: bool
//	 	 9 ThresholdIDs: strings separated by "&".
//		10 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicResource struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicResource) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicResource) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicResource) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
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
		args := &utils.ResourceProfileWithAPIOpts{
			ResourceProfile: &utils.ResourceProfile{
				Tenant:            params[0],
				ID:                params[1],
				AllocationMessage: params[6],
			},
			APIOpts: make(map[string]any),
		}
		// populate Resource's FilterIDs
		if params[2] != utils.EmptyString {
			args.FilterIDs = strings.Split(params[2], utils.ANDSep)
		}
		// populate Resource's Weights
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
		// populate Resource's UsageTTL
		if params[4] != utils.EmptyString {
			args.UsageTTL, err = utils.ParseDurationWithNanosecs(params[4])
			if err != nil {
				return err
			}
		}
		// populate Resource's Limit
		if params[5] != utils.EmptyString {
			args.Limit, err = strconv.ParseFloat(params[5], 64)
			if err != nil {
				return err
			}
		}
		// populate Resource's Blocker
		if params[7] != utils.EmptyString {
			args.Blocker, err = strconv.ParseBool(params[7])
			if err != nil {
				return err
			}
		}
		// populate Resource's Stored
		if params[8] != utils.EmptyString {
			args.Stored, err = strconv.ParseBool(params[8])
			if err != nil {
				return err
			}
		}
		// populate Resource's ThresholdIDs
		if params[9] != utils.EmptyString {
			args.ThresholdIDs = strings.Split(params[9], utils.ANDSep)
		}
		// populate Resource's APIOpts
		if params[10] != utils.EmptyString {
			if err := parseParamStringToMap(params[10], args.APIOpts); err != nil {
				return err
			}
		}

		// create the ResourceProfile based on the populated parameters
		var rply string
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetResourceProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicTrend processes the `ActionDiktatsOpts` field from the action to construct a TrendProfile
//
// The ActionDiktatsOpts field format is expected as follows:
//
//		 0 Tenant: string
//		 1 ID: string
//		 2 Schedule: string
//		 3 StatID: string
//		 4 Metrics: strings separated by "&".
//	 	 5 TTL: duration
//	 	 6 QueueLength: integer
//	 	 7 MinItems: integer
//	 	 8 CorrelationType: string
//	 	 9 Tolerance: float
//	 	10 Stored: bool
//	 	11 ThresholdIDs: strings separated by "&".
//		12 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicTrend struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicTrend) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicTrend) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicTrend) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
		if len(params) != 13 {
			return fmt.Errorf("invalid number of parameters <%d> expected 13", len(params))
		}
		// parse dynamic parameters
		for i := range params {
			if params[i], err = utils.ParseParamForDataProvider(params[i], data, false); err != nil {
				return err
			}
		}
		// Prepare request arguments based on provided parameters.
		args := &utils.TrendProfileWithAPIOpts{
			TrendProfile: &utils.TrendProfile{
				Tenant:          params[0],
				ID:              params[1],
				Schedule:        params[2],
				StatID:          params[3],
				CorrelationType: params[8],
			},
			APIOpts: make(map[string]any),
		}
		// populate Trend's Metrics
		if params[4] != utils.EmptyString {
			args.Metrics = strings.Split(params[4], utils.ANDSep)
		}
		// populate Trend's TTL
		if params[5] != utils.EmptyString {
			args.TTL, err = utils.ParseDurationWithNanosecs(params[5])
			if err != nil {
				return err
			}
		}
		// populate Trend's QueueLengh
		if params[6] != utils.EmptyString {
			args.QueueLength, err = strconv.Atoi(params[6])
			if err != nil {
				return err
			}
		}
		// populate Trend's MinItems
		if params[7] != utils.EmptyString {
			args.MinItems, err = strconv.Atoi(params[7])
			if err != nil {
				return err
			}
		}
		// populate Trend's Tolerance
		if params[9] != utils.EmptyString {
			args.Tolerance, err = strconv.ParseFloat(params[9], 64)
			if err != nil {
				return err
			}
		}
		// populate Trend's Stored
		if params[10] != utils.EmptyString {
			args.Stored, err = strconv.ParseBool(params[10])
			if err != nil {
				return err
			}
		}
		// populate Trend's ThresholdIDs
		if params[11] != utils.EmptyString {
			args.ThresholdIDs = strings.Split(params[11], utils.ANDSep)
		}
		// populate Trend's APIOpts
		if params[12] != utils.EmptyString {
			if err := parseParamStringToMap(params[12], args.APIOpts); err != nil {
				return err
			}
		}

		// create the TrendProfile based on the populated parameters
		var rply string
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetTrendProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicRanking processes the `ActionDiktatsOpts` field from the action to construct a RankingProfile
//
// The ActionDiktatsOpts field format is expected as follows:
//
//		 0 Tenant: string
//		 1 ID: string
//		 2 Schedule: string
//		 3 StatIDs: strings separated by "&".
//		 4 MetricIDs: strings separated by "&".
//	 	 5 Sorting: string
//	 	 6 SortingParameters: strings separated by "&".
//	 	 7 Stored: bool
//	 	 8 ThresholdIDs: strings separated by "&".
//		 9 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicRanking struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicRanking) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicRanking) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicRanking) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
		if len(params) != 10 {
			return fmt.Errorf("invalid number of parameters <%d> expected 10", len(params))
		}
		// parse dynamic parameters
		for i := range params {
			if params[i], err = utils.ParseParamForDataProvider(params[i], data, false); err != nil {
				return err
			}
		}
		// Prepare request arguments based on provided parameters.
		args := &utils.RankingProfileWithAPIOpts{
			RankingProfile: &utils.RankingProfile{
				Tenant:   params[0],
				ID:       params[1],
				Schedule: params[2],
				Sorting:  params[5],
			},
			APIOpts: make(map[string]any),
		}
		// populate Ranking's StatIDs
		if params[3] != utils.EmptyString {
			args.StatIDs = strings.Split(params[3], utils.ANDSep)
		}
		// populate Ranking's MetricIDs
		if params[4] != utils.EmptyString {
			args.MetricIDs = strings.Split(params[4], utils.ANDSep)
		}
		// populate Ranking's SortingParameters
		if params[6] != utils.EmptyString {
			args.SortingParameters = strings.Split(params[6], utils.ANDSep)
		}
		// populate Ranking's Stored
		if params[7] != utils.EmptyString {
			args.Stored, err = strconv.ParseBool(params[7])
			if err != nil {
				return err
			}
		}
		// populate Ranking's ThresholdIDs
		if params[8] != utils.EmptyString {
			args.ThresholdIDs = strings.Split(params[8], utils.ANDSep)
		}
		// populate Ranking's APIOpts
		if params[9] != utils.EmptyString {
			if err := parseParamStringToMap(params[9], args.APIOpts); err != nil {
				return err
			}
		}

		// create the RankingProfile based on the populated parameters
		var rply string
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetRankingProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicFilter processes the `ActionDiktatsOpts` field from the action to construct a Filter
//
// The ActionDiktatsOpts field format is expected as follows:
//
//		0 Tenant: string
//		1 ID: string
//		2 Type: string
//		3 Path: string
//		4 Values: strings separated by "&".
//	 	5 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicFilter struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicFilter) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicFilter) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicFilter) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
		if len(params) != 6 {
			return fmt.Errorf("invalid number of parameters <%d> expected 6", len(params))
		}
		// parse dynamic parameters
		for i := range params {
			var onlyEncapsulatead bool
			if i == 3 { // dont parse un-encapsulated "< >" string from Path
				onlyEncapsulatead = true
			}
			if params[i], err = utils.ParseParamForDataProvider(params[i], data, onlyEncapsulatead); err != nil {
				return err
			}
		}
		// Prepare request arguments based on provided parameters.
		args := &engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: params[0],
				ID:     params[1],
				Rules: []*engine.FilterRule{{
					Type:    params[2],
					Element: params[3],
				}},
			},
			APIOpts: make(map[string]any),
		}
		// populate Filter's Values
		if params[4] != utils.EmptyString {
			args.Filter.Rules[0].Values = strings.Split(params[4], utils.ANDSep)
		}
		// populate Filter's APIOpts
		if params[5] != utils.EmptyString {
			if err := parseParamStringToMap(params[5], args.APIOpts); err != nil {
				return err
			}
		}

		// create the Filter based on the populated parameters
		var rply string
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetFilter, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}

// actDynamicRoute processes the `ActionDiktatsOpts` field from the action to construct a RouteProfile
//
// The ActionDiktatsOpts field format is expected as follows:
//
//		 0 Tenant: string
//		 1 ID: string
//		 2 FilterIDs: strings separated by "&".
//		 3 Weights: strings separated by "&".
//		 4 Blockers: strings separated by "&".
//	 	 5 Sorting: string
//	 	 6 SortingParameters: strings separated by "&".
//	 	 7 RouteID: string
//	 	 8 RouteFilterIDs: strings separated by "&".
//	 	 9 RouteAccountIDs: strings separated by "&".
//	 	10 RouteRateProfileIDs: strings separated by "&".
//	 	11 RouteResourceIDs: strings separated by "&".
//	 	12 RouteStatIDs: strings separated by "&".
//	 	13 RouteWeights: strings separated by "&".
//	 	14 RouteBlockers: strings separated by "&".
//	 	15 RouteParameters: string
//		16 APIOpts: set of key-value pairs (separated by "&").
//
// Parameters are separated by ";" and must be provided in the specified order.
type actDynamicRoute struct {
	config  *config.CGRConfig
	connMgr *engine.ConnManager
	dm      *engine.DataManager
	fltrS   *engine.FilterS
	aCfg    *utils.APAction
	tnt     string
	cgrEv   *utils.CGREvent
}

func (aL *actDynamicRoute) id() string {
	return aL.aCfg.ID
}

func (aL *actDynamicRoute) cfg() *utils.APAction {
	return aL.aCfg
}

// execute implements actioner interface
func (aL *actDynamicRoute) execute(ctx *context.Context, data utils.MapStorage, trgID string) (err error) {
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
	weights := make(map[string]float64)   // stores sorting weights by Diktat ID
	diktats := make([]*utils.APDiktat, 0) // list of diktats which have *template in opts, will be weight sorted later
	for _, diktat := range aL.aCfg.Diktats {
		if pass, err := aL.fltrS.Pass(ctx, aL.tnt, diktat.FilterIDs, data); err != nil {
			return err
		} else if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, diktat.Weights, aL.fltrS, aL.tnt, data)
		if err != nil {
			return err
		}
		weights[diktat.ID] = weight
		diktats = append(diktats, diktat)
	}
	// Sort by weight (higher values first).
	slices.SortFunc(diktats, func(a, b *utils.APDiktat) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})
	for _, diktat := range diktats {
		params := strings.Split(utils.IfaceAsString(diktat.Opts[utils.MetaTemplate]),
			utils.InfieldSep)
		if len(params) != 17 {
			return fmt.Errorf("invalid number of parameters <%d> expected 17", len(params))
		}
		// parse dynamic parameters
		for i := range params {
			if params[i], err = utils.ParseParamForDataProvider(params[i], data, false); err != nil {
				return err
			}
		}
		// Take only the string after @, for cases when the RouteProfileID is gotten from a switch agents event
		routeFieldParts := strings.Split(params[1], utils.AtChar)
		routeProfileFound := new(utils.RouteProfile)
		if len(routeFieldParts) > 2 {
			return fmt.Errorf("more than 1 \"@\" character for RouteProfileID: <%s>", params[1])
		} else if len(routeFieldParts) > 1 {
			params[1] = routeFieldParts[1]
			if routeProfileFound, err = aL.dm.GetRouteProfile(context.Background(),
				utils.FirstNonEmpty(aL.cgrEv.Tenant, aL.tnt,
					config.CgrConfig().GeneralCfg().DefaultTenant),
				params[1], true, true, utils.NonTransactional); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				return err
			}
		}
		// Prepare request arguments based on provided parameters.
		args := &utils.RouteProfileWithAPIOpts{
			RouteProfile: &utils.RouteProfile{
				Tenant:  utils.FirstNonEmpty(params[0], routeProfileFound.Tenant),
				ID:      params[1],
				Sorting: utils.FirstNonEmpty(params[5], routeProfileFound.Sorting),
			},
			APIOpts: make(map[string]any),
		}
		// populate RouteProfile's FilterIDs
		if params[2] != utils.EmptyString {
			args.FilterIDs = strings.Split(params[2], utils.ANDSep)
		} else {
			args.FilterIDs = routeProfileFound.FilterIDs
		}
		// populate RouteProfile's Weights
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
		} else {
			args.Weights = routeProfileFound.Weights
		}
		// populate RouteProfile's Blockers
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
		} else {
			args.Blockers = routeProfileFound.Blockers
		}
		// populate RouteProfile's SortingParameters
		if params[6] != utils.EmptyString {
			args.SortingParameters = strings.Split(params[6], utils.ANDSep)
		} else {
			args.SortingParameters = routeProfileFound.SortingParameters
		}
		// populate RouteProfile's Routes
		if params[7] != utils.EmptyString {
			// keep the existing routes if routeProfile already existed, and modify the specified Routes by ID
			var routeModified bool // if route doesnt exist in the found route Profile
			for _, existingRoute := range routeProfileFound.Routes {
				if existingRoute.ID == params[7] { // modify routes with ID
					// populate RouteProfile's RouteFilterIDs
					if params[8] != utils.EmptyString {
						existingRoute.FilterIDs = strings.Split(params[8], utils.ANDSep)
					}
					// populate RouteProfile's RouteAccountIDs
					if params[9] != utils.EmptyString {
						existingRoute.AccountIDs = strings.Split(params[9], utils.ANDSep)
					}
					// populate RouteProfile's RouteRateProfileIDs
					if params[10] != utils.EmptyString {
						existingRoute.RateProfileIDs = strings.Split(params[10], utils.ANDSep)
					}
					// populate RouteProfile's RouteResourceIDs
					if params[11] != utils.EmptyString {
						existingRoute.ResourceIDs = strings.Split(params[11], utils.ANDSep)
					}
					// populate RouteProfile's RouteStatIDs
					if params[12] != utils.EmptyString {
						existingRoute.StatIDs = strings.Split(params[12], utils.ANDSep)
					}
					// populate RouteProfile's RouteWeight
					if params[13] != utils.EmptyString {
						existingRoute.Weights = utils.DynamicWeights{&utils.DynamicWeight{}}
						wghtSplit := strings.Split(params[13], utils.ANDSep)
						if len(wghtSplit) > 2 {
							return utils.ErrUnsupportedFormat
						}
						if wghtSplit[0] != utils.EmptyString {
							existingRoute.Weights[0].FilterIDs = []string{wghtSplit[0]}
						}
						if wghtSplit[1] != utils.EmptyString {
							existingRoute.Weights[0].Weight, err = strconv.ParseFloat(wghtSplit[1], 64)
							if err != nil {
								return err
							}
						}
					}
					// populate RouteProfile's RouteBlocker
					if params[14] != utils.EmptyString {
						existingRoute.Blockers = utils.DynamicBlockers{&utils.DynamicBlocker{}}
						blckrSplit := strings.Split(params[14], utils.ANDSep)
						if len(blckrSplit) > 2 {
							return utils.ErrUnsupportedFormat
						}
						if blckrSplit[0] != utils.EmptyString {
							existingRoute.Blockers[0].FilterIDs = []string{blckrSplit[0]}
						}
						if blckrSplit[1] != utils.EmptyString {
							existingRoute.Blockers[0].Blocker, err = strconv.ParseBool(blckrSplit[1])
							if err != nil {
								return err
							}
						}
					}
					// populate RouteProfile's RouteParameters
					if params[15] != utils.EmptyString {
						existingRoute.RouteParameters = params[15]
					}
					routeModified = true
				}
				args.Routes = append(args.Routes, existingRoute)
			}
			if !routeModified { // if no existing routes were modified, append a new route
				appendRoute := new(utils.Route) // new route to be appended
				// populate RouteProfile's RouteID
				appendRoute.ID = params[7]
				// populate RouteProfile's RouteFilterIDs
				if params[8] != utils.EmptyString {
					appendRoute.FilterIDs = strings.Split(params[8], utils.ANDSep)
				}
				// populate RouteProfile's RouteAccountIDs
				if params[9] != utils.EmptyString {
					appendRoute.AccountIDs = strings.Split(params[9], utils.ANDSep)
				}
				// populate RouteProfile's RouteRateProfileIDs
				if params[10] != utils.EmptyString {
					appendRoute.RateProfileIDs = strings.Split(params[10], utils.ANDSep)
				}
				// populate RouteProfile's RouteResourceIDs
				if params[11] != utils.EmptyString {
					appendRoute.ResourceIDs = strings.Split(params[11], utils.ANDSep)
				}
				// populate RouteProfile's RouteStatIDs
				if params[12] != utils.EmptyString {
					appendRoute.StatIDs = strings.Split(params[12], utils.ANDSep)
				}
				// populate RouteProfile's RouteWeight
				if params[13] != utils.EmptyString {
					appendRoute.Weights = utils.DynamicWeights{&utils.DynamicWeight{}}
					wghtSplit := strings.Split(params[13], utils.ANDSep)
					if len(wghtSplit) > 2 {
						return utils.ErrUnsupportedFormat
					}
					if wghtSplit[0] != utils.EmptyString {
						appendRoute.Weights[0].FilterIDs = []string{wghtSplit[0]}
					}
					if wghtSplit[1] != utils.EmptyString {
						appendRoute.Weights[0].Weight, err = strconv.ParseFloat(wghtSplit[1], 64)
						if err != nil {
							return err
						}
					}
				}
				// populate RouteProfile's RouteBlocker
				if params[14] != utils.EmptyString {
					appendRoute.Blockers = utils.DynamicBlockers{&utils.DynamicBlocker{}}
					blckrSplit := strings.Split(params[14], utils.ANDSep)
					if len(blckrSplit) > 2 {
						return utils.ErrUnsupportedFormat
					}
					if blckrSplit[0] != utils.EmptyString {
						appendRoute.Blockers[0].FilterIDs = []string{blckrSplit[0]}
					}
					if blckrSplit[1] != utils.EmptyString {
						appendRoute.Blockers[0].Blocker, err = strconv.ParseBool(blckrSplit[1])
						if err != nil {
							return err
						}
					}
				}
				// populate RouteProfile's RouteParameters
				appendRoute.RouteParameters = params[15]
				args.Routes = append(args.Routes, appendRoute)
			}
		} else {
			args.Routes = routeProfileFound.Routes
		}
		// populate RouteProfile's APIOpts
		if params[16] != utils.EmptyString {
			if err := parseParamStringToMap(params[16], args.APIOpts); err != nil {
				return err
			}
		}

		// create the RouteProfile based on the populated parameters
		var rply string
		if err = aL.connMgr.Call(ctx, aL.config.ActionSCfg().AdminSConns,
			utils.AdminSv1SetRouteProfile, args, &rply); err != nil {
			return err
		}
		if blocker, err := engine.BlockerFromDynamics(ctx, diktat.Blockers, aL.fltrS, aL.tnt, data); err != nil {
			return err
		} else if blocker {
			break
		}
	}
	return
}
