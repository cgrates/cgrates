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
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// FilterFloat64CfgOpts returns the option as float64 if the filters match
func FilterFloat64CfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, dynOpts []*utils.DynamicFloat64Opt) (float64, error) {
	for _, opt := range dynOpts { // iterate through the options
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value, nil
		}
	}
	return 0, utils.ErrNotFound // return NOT_FOUND if there are no options or none of the filters pass
}

// FilterDurationCfgOpts returns the option as time.Duration if the filters match
func FilterDurationCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, dynOpts []*utils.DynamicDurationOpt) (time.Duration, error) {
	for _, opt := range dynOpts { // iterate through the options
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value, nil
		}
	}
	return 0, utils.ErrNotFound // return NOT_FOUND if there are no options or none of the filters pass
}

// FilterStringCfgOpts returns the option as string if the filters match
func FilterStringCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, dynOpts []*utils.DynamicStringOpt) (string, error) {
	for _, opt := range dynOpts { // iterate through the options
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return utils.EmptyString, err
		} else if pass {
			return opt.Value, nil
		}
	}
	return utils.EmptyString, utils.ErrNotFound // return NOT_FOUND if there are no options or none of the filters pass
}

// FilterStringSliceCfgOpts returns the option as []string if the filters match
func FilterStringSliceCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, dynOpts []*utils.DynamicStringSliceOpt) ([]string, error) {
	for _, opt := range dynOpts { // iterate through the options
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt.Value, nil
		}
	}
	return nil, utils.ErrNotFound // return NOT_FOUND if there are no options or none of the filters pass
}

// FilterIntCfgOpts returns the option as int if the filters match
func FilterIntCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, dynOpts []*utils.DynamicIntOpt) (int, error) {
	for _, opt := range dynOpts { // iterate through the options
		if pass, err := fS.Pass(ctx, tnt, opt.FilterIDs, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt.Value, nil
		}
	}
	return 0, utils.ErrNotFound // return NOT_FOUND if there are no options or none of the filters pass
}
