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

package engine

import (
	"errors"
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// GetFloat64Opts checks the specified option names in order among the keys in APIOpts returning the first value it finds as float64, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetFloat64Opts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicFloat64Opt,
	optNames ...string) (cfgOpt float64, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsFloat64(opt)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return 0, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetDurationOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as time.Duration, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetDurationOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicDurationOpt,
	optNames ...string) (cfgOpt time.Duration, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsDuration(opt)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return 0, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetDurationPointerOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *time.Duration, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetDurationPointerOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicDurationPointerOpt,
	optNames ...string) (cfgOpt *time.Duration, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		var value time.Duration
		value, err = utils.IfaceAsDuration(opt)
		if err != nil {
			return nil, err
		}
		return utils.DurationPointer(value), nil
	} else if !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetStringOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as string, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetStringOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicStringOpt,
	optNames ...string) (cfgOpt string, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsString(opt), nil
	} else if !errors.Is(err, utils.ErrNotFound) {
		return "", err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return utils.EmptyString, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetTimeOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as time.Time, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetTimeOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicStringOpt,
	tmz string, optNames ...string) (_ time.Time, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsTime(opt, tmz)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return time.Time{}, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		var pass bool
		if pass, err = fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return
		} else if pass {
			var dur string
			dur, err = opt.Value(dP)
			if err != nil {
				return
			}
			return utils.ParseTimeDetectLayout(dur, tmz)
		}
	}
	return
}

// GetStringSliceOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as []string, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetStringSliceOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicStringSliceOpt,
	dftOpt []string, optNames ...string) (cfgOpt []string, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsStringSlice(opt)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt.Values, nil
		}
	}
	return dftOpt, nil // return the default value if there are no options and none of the filters pass
}

// GetIntOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as int, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetIntOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicIntOpt,
	optNames ...string) (cfgOpt int, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsInt(opt)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return 0, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetBoolOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as bool, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetBoolOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicBoolOpt,
	optNames ...string) (cfgOpt bool, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsBool(opt)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return false, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return false, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetDecimalBigOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *decimal.Big, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetDecimalBigOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicDecimalOpt,
	optNames ...string) (cfgOpt *decimal.Big, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return utils.IfaceAsBig(opt)
	} else if !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return
}

// GetInterfaceOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as any, otherwise it
// returns the config option if at least one filter passes or the default value if none of them do
func GetInterfaceOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicInterfaceOpt,
	optNames ...string) (cfgOpt any, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		return opt, nil
	} else if !errors.Is(err, utils.ErrNotFound) {
		return false, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return false, err
		} else if pass {
			return opt.Value, nil
		}
	}
	return
}

// GetIntPointerOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *int, otherwise it
// returns the config option if at least one filter passes or NOT_FOUND if none of them do
func GetIntPointerOpts(ctx *context.Context, tnt string, dP utils.DataProvider, fS *FilterS, dynOpts []*config.DynamicIntPointerOpt,
	optNames ...string) (cfgOpt *int, err error) {
	if opt, err := optIfaceFromDP(dP, optNames); err == nil {
		var value int
		if value, err = utils.IfaceAsInt(opt); err != nil {
			return nil, err
		}
		return utils.IntPointer(value), nil
	} else if !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, dP); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt.Value(dP)
		}
	}
	return nil, nil
}

// GetDurationPointerOptsFromMultipleMaps checks the specified option names in order among the keys in APIOpts, then in startOpts, returning the first value it finds as *time.Duration,
// otherwise it returns the config option if at least one filter passes or NOT_FOUND if none of them do
func GetDurationPointerOptsFromMultipleMaps(ctx *context.Context, tnt string, eventStart, apiOpts, startOpts map[string]any, fS *FilterS,
	dynOpts []*config.DynamicDurationPointerOpt, optName string) (cfgOpt *time.Duration, err error) {
	var value time.Duration
	if opt, has := apiOpts[optName]; has {
		if value, err = utils.IfaceAsDuration(opt); err != nil {
			return nil, err
		}
		return utils.DurationPointer(value), nil
	} else if opt, has = startOpts[optName]; has {
		if value, err = utils.IfaceAsDuration(opt); err != nil {
			return nil, err
		}
		return utils.DurationPointer(value), nil
	}
	evMS := utils.MapStorage{
		utils.MetaOpts: apiOpts,
		utils.MetaReq:  eventStart,
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, evMS); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt.Value(evMS)
		}
	}
	return nil, nil
}

// GetDurationOptsFromMultipleMaps checks the specified option names in order among the keys in APIOpts, then in startOpts, returning the first value it finds as time.Duration,
// otherwise it returns the config option if at least one filter passes or the default one if none of them do
func GetDurationOptsFromMultipleMaps(ctx *context.Context, tnt string, eventStart, apiOpts, startOpts map[string]any, fS *FilterS, dynOpts []*config.DynamicDurationOpt,
	dftOpt time.Duration, optName string) (cfgOpt time.Duration, err error) {
	var value time.Duration
	if opt, has := apiOpts[optName]; has {
		if value, err = utils.IfaceAsDuration(opt); err != nil {
			return 0, err
		}
		return value, nil
	} else if opt, has = startOpts[optName]; has {
		if value, err = utils.IfaceAsDuration(opt); err != nil {
			return 0, err
		}
		return value, nil
	}
	evMS := utils.MapStorage{
		utils.MetaOpts: apiOpts,
		utils.MetaReq:  eventStart,
	}
	for _, opt := range dynOpts { // iterate through the options
		if !slices.Contains([]string{utils.EmptyString, utils.MetaAny, tnt}, opt.Tenant) {
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, evMS); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value(evMS)
		}
	}
	return dftOpt, nil // return the default value if there are no options and none of the filters pass
}

func ConvertOptsToMapStringAny(in any) (map[string]any, error) {
	out := make(map[string]any)
	switch val := in.(type) {
	case MapEvent:
		for k, v := range val {
			out[k] = v
		}
	case utils.MapStorage:
		for k, v := range val {
			out[k] = v
		}
	case map[string]any:
		return val, nil
	default:
		return nil, errors.New("cannot convert to map[string]any")
	}
	return out, nil
}

// getOptIfaceFromDP is a helper that returns the first option (as interface{}) found in the data provider
func optIfaceFromDP(dP utils.DataProvider, optNames []string) (any, error) {
	values, err := dP.FieldAsInterface([]string{utils.MetaOpts})
	if err != nil {
		return nil, err
	}
	opts, err := ConvertOptsToMapStringAny(values)
	if err != nil {
		return nil, err
	}
	for _, optName := range optNames {
		if opt, ok := opts[optName]; ok {
			return opt, nil
		}
	}
	return nil, utils.ErrNotFound
}
