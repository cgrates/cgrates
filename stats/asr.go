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

package stats

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewASR() (StatsMetric, error) {
	return new(ASR), nil
}

// ASR implements AverageSuccessRatio metric
type ASR struct {
	Answered float64
	Count    float64
}

func (asr *ASR) GetStringValue(fmtOpts string) (valStr string) {
	if asr.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	val := asr.GetValue().(float64)
	return fmt.Sprintf("%v%%", val) // %v will automatically limit the number of decimals printed
}

func (asr *ASR) GetValue() (v interface{}) {
	if asr.Count == 0 {
		return float64(engine.STATS_NA)
	}
	return utils.Round((asr.Answered / asr.Count * 100),
		config.CgrConfig().RoundingDecimals, utils.ROUNDING_MIDDLE)
}

func (asr *ASR) AddEvent(ev engine.StatsEvent) (err error) {
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		asr.Answered += 1
	}
	asr.Count += 1
	return
}

func (asr *ASR) RemEvent(ev engine.StatsEvent) (err error) {
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		asr.Answered -= 1
	}
	asr.Count -= 1
	return
}

func (asr *ASR) GetMarshaled(ms engine.Marshaler) (vals []byte, err error) {
	return ms.Marshal(asr)
}

func (asr *ASR) SetFromMarshaled(vals []byte, ms engine.Marshaler) (err error) {
	return ms.Unmarshal(vals, asr)
}
