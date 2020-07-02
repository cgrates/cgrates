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
	rt0 := &engine.Rate{
		ID:     "RATE0",
		Weight: 0,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	rt0.Compile()
	rtChristmas := &engine.Rate{
		ID:              "RT_CHRISTMAS",
		ActivationStart: "* * 24 12 *",
		Weight:          50,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: time.Duration(0),
			},
		},
	}
	rtChristmas.Compile()
	allRts := []*engine.Rate{rt0, rtChristmas}
	expOrdered := []*engine.RateSInterval{
		{
			Increments: []*engine.RateSIncrement{
				{
					Rate:              rt0,
					IntervalRateIndex: 0,
				},
			},
		},
	}
	//
	// time.Date(2020, time.December, 23, 23, 59, 05, 0, time.UTC),
	if ordRts := orderRatesOnIntervals(
		allRts, time.Date(2020, time.June, 28, 18, 56, 05, 0, time.UTC),
		time.Duration(2*time.Minute), true); !reflect.DeepEqual(expOrdered, ordRts) {
		t.Errorf("expecting: %s\n, received: %s",
			utils.ToIJSON(expOrdered), utils.ToIJSON(ordRts))
	}
}
