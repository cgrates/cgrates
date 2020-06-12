/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package rates

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestOrderRatesOnIntervals(t *testing.T) {
	allRts := map[time.Duration][]*engine.Rate{
		time.Duration(0): {
			&engine.Rate{
				ID:            "RATE0",
				IntervalStart: time.Duration(0),
			},
			&engine.Rate{
				ID:            "RATE100",
				IntervalStart: time.Duration(0),
				Weight:        100,
			},
			&engine.Rate{
				ID:            "RATE50",
				IntervalStart: time.Duration(0),
				Weight:        50,
			},
		},
	}
	expOrdered := []*engine.Rate{
		&engine.Rate{
			ID:            "RATE100",
			IntervalStart: time.Duration(0),
			Weight:        100,
		},
	}
	if ordRts := orderRatesOnIntervals(allRts); !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
	allRts = map[time.Duration][]*engine.Rate{
		time.Duration(1 * time.Minute): {
			&engine.Rate{
				ID:            "RATE30",
				IntervalStart: time.Duration(1 * time.Minute),
				Weight:        30,
			},
			&engine.Rate{
				ID:            "RATE70",
				IntervalStart: time.Duration(1 * time.Minute),
				Weight:        70,
			},
		},
		time.Duration(0): {
			&engine.Rate{
				ID:            "RATE0",
				IntervalStart: time.Duration(0),
			},
			&engine.Rate{
				ID:            "RATE100",
				IntervalStart: time.Duration(0),
				Weight:        100,
			},
			&engine.Rate{
				ID:            "RATE50",
				IntervalStart: time.Duration(0),
				Weight:        50,
			},
		},
	}
	expOrdered = []*engine.Rate{
		&engine.Rate{
			ID:            "RATE100",
			IntervalStart: time.Duration(0),
			Weight:        100,
		},
		&engine.Rate{
			ID:            "RATE70",
			IntervalStart: time.Duration(1 * time.Minute),
			Weight:        70,
		},
	}
	if ordRts := orderRatesOnIntervals(allRts); !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
	allRts = map[time.Duration][]*engine.Rate{
		time.Duration(1 * time.Minute): {
			&engine.Rate{
				ID:            "RATE30",
				IntervalStart: time.Duration(1 * time.Minute),
				Weight:        30,
			},
			&engine.Rate{
				ID:            "RATE70",
				IntervalStart: time.Duration(1 * time.Minute),
				Weight:        70,
			},
		},
		time.Duration(2 * time.Minute): {
			&engine.Rate{
				ID:            "RATE0",
				IntervalStart: time.Duration(2 * time.Minute),
			},
		},
		time.Duration(0): {
			&engine.Rate{
				ID:            "RATE0",
				IntervalStart: time.Duration(0),
			},
			&engine.Rate{
				ID:            "RATE100",
				IntervalStart: time.Duration(0),
				Weight:        100,
			},
			&engine.Rate{
				ID:            "RATE50",
				IntervalStart: time.Duration(0),
				Weight:        50,
			},
		},
	}
	expOrdered = []*engine.Rate{
		&engine.Rate{
			ID:            "RATE100",
			IntervalStart: time.Duration(0),
			Weight:        100,
		},
		&engine.Rate{
			ID:            "RATE70",
			IntervalStart: time.Duration(1 * time.Minute),
			Weight:        70,
		},
		&engine.Rate{
			ID:            "RATE0",
			IntervalStart: time.Duration(2 * time.Minute),
		},
	}
	if ordRts := orderRatesOnIntervals(allRts); !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
}
