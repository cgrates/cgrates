// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"time"

	"github.com/ericlagergren/decimal"
)

// GetFloat64Opts checks the specified option names in order among the keys in APIOpts returning the first value it finds as float64, otherwise it
// returns the default option (usually the value specified in config)
func GetFloat64Opts(ev *CGREvent, dftOpt float64, optNames ...string) (cfgOpt float64, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return IfaceAsFloat64(opt)
		}
	}
	return dftOpt, nil
}

// GetDurationOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as time.Duration, otherwise it
// returns the default option (usually the value specified in config)
func GetDurationOpts(ev *CGREvent, dftOpt time.Duration, optNames ...string) (cfgOpt time.Duration, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return IfaceAsDuration(opt)
		}
	}
	return dftOpt, nil
}

// GetStringOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as string, otherwise it
// returns the default option (usually the value specified in config)
func GetStringOpts(ev *CGREvent, dftOpt string, optNames ...string) (cfgOpt string) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return IfaceAsString(opt)
		}
	}
	return dftOpt
}

// GetStringSliceOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as []string, otherwise it
// returns the default option (usually the value specified in config)
func GetStringSliceOpts(ev *CGREvent, dftOpt []string, optNames ...string) (cfgOpt []string, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return IfaceAsSliceString(opt)
		}
	}
	return dftOpt, nil
}

// GetIntOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as int, otherwise it
// returns the default option (usually the value specified in config)
func GetIntOpts(ev *CGREvent, dftOpt int, optNames ...string) (cfgOpt int, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			var value int64
			if value, err = IfaceAsTInt64(opt); err != nil {
				return
			}
			return int(value), nil
		}
	}
	return dftOpt, nil
}

// GetBoolOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as bool, otherwise it
// returns the default option (usually the value specified in config)
func GetBoolOpts(ev *CGREvent, dftOpt bool, optNames ...string) (cfgOpt bool, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return IfaceAsBool(opt)
		}
	}
	return dftOpt, nil
}

// GetDecimalBigOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *decimal.Big, otherwise it
// returns the default option (usually the value specified in config)
func GetDecimalBigOpts(ev *CGREvent, dftOpt *decimal.Big, optNames ...string) (cfgOpt *decimal.Big, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return IfaceAsBig(opt)
		}
	}
	return dftOpt, nil
}

// GetInterfaceOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as any, otherwise it
// returns the default option (usually the value specified in config)
func GetInterfaceOpts(ev *CGREvent, dftOpt any, optNames ...string) (cfgOpt any) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			return opt
		}
	}
	return dftOpt
}

// GetIntPointerOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *int, otherwise it
// returns the default option (usually the value specified in config)
func GetIntPointerOpts(ev *CGREvent, dftOpt *int, optNames ...string) (cfgOpt *int, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			var value int64
			if value, err = IfaceAsTInt64(opt); err != nil {
				return
			}
			return IntPointer(int(value)), nil
		}
	}
	return dftOpt, nil
}

// GetDurationPointerOpts checks the specified option names in order among the keys in APIOpts returning the first value it finds as *time.Duration, otherwise it
// returns the default option (usually the value specified in config)
func GetDurationPointerOpts(ev *CGREvent, dftOpt *time.Duration, optNames ...string) (cfgOpt *time.Duration, err error) {
	for _, optName := range optNames {
		if opt, has := ev.APIOpts[optName]; has {
			var value time.Duration
			if value, err = IfaceAsDuration(opt); err != nil {
				return
			}
			return DurationPointer(value), nil
		}
	}
	return dftOpt, nil
}
