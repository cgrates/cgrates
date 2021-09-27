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

// filterFloat64CfgOpts returns the option as float64 if the filters match
func filterFloat64CfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, opts map[string]float64) (float64, error) {
	for filter, opt := range opts { // iterate through the option map
		if filter == utils.EmptyString { // if the filter key is empty continue
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, []string{filter}, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt, nil
		}
	}
	if opt, has := opts[utils.EmptyString]; has { // if the empty key exists in the opts map we can assume the filter is passing so we can return the option
		return opt, nil
	}
	return 0, utils.ErrNotFound // return NOT_FOUND if option map is empty or none of the filters pass
}

// filterDurationCfgOpts returns the option as time.Duration if the filters match
func filterDurationCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, opts map[string]time.Duration) (time.Duration, error) {
	for filter, opt := range opts { // iterate through the option map
		if filter == utils.EmptyString { // if the filter key is empty continue
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, []string{filter}, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return 0, err
		} else if pass {
			return opt, nil
		}
	}
	if opt, has := opts[utils.EmptyString]; has { // if the empty key exists in the opts map we can assume the filter is passing so we can return the option
		return opt, nil
	}
	return 0, utils.ErrNotFound // return NOT_FOUND if option map is empty or none of the filters pass
}

// filterStringCfgOpts returns the option as string if the filters match
func filterStringCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, opts map[string]string) (string, error) {
	for filter, opt := range opts { // iterate through the option map
		if filter == utils.EmptyString { // if the filter key is empty continue
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, []string{filter}, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return utils.EmptyString, err
		} else if pass {
			return opt, nil
		}
	}
	if opt, has := opts[utils.EmptyString]; has { // if the empty key exists in the opts map we can assume the filter is passing so we can return the option
		return opt, nil
	}
	return utils.EmptyString, utils.ErrNotFound // return NOT_FOUND if option map is empty or none of the filters pass
}

// filterStringSliceCfgOpts returns the option as []string if the filters match
func filterStringSliceCfgOpts(ctx *context.Context, tnt string, ev utils.DataProvider, fS *FilterS, opts map[string][]string) ([]string, error) {
	for filter, opt := range opts { // iterate through the option map
		if filter == utils.EmptyString { // if the filter key is empty continue
			continue
		}
		if pass, err := fS.Pass(ctx, tnt, []string{filter}, ev); err != nil { // check if the filter is passing for the DataProvider and return the option if it does
			return nil, err
		} else if pass {
			return opt, nil
		}
	}
	if opt, has := opts[utils.EmptyString]; has { // if the empty key exists in the opts map we can assume the filter is passing so we can return the option
		return opt, nil
	}
	return nil, utils.ErrNotFound // return NOT_FOUND if option map is empty or none of the filters pass
}
