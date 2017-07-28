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
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestASRGetStringValue(t *testing.T) {
	asr, _ := NewASR()
	if strVal := asr.GetStringValue(""); strVal != utils.NOT_AVAILABLE {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(
		engine.StatsEvent{
			"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)})
	if strVal := asr.GetStringValue(""); strVal != "100%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.AddEvent(engine.StatsEvent{})
	asr.AddEvent(engine.StatsEvent{})
	if strVal := asr.GetStringValue(""); strVal != "33.33333%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
	asr.RemEvent(engine.StatsEvent{})
	if strVal := asr.GetStringValue(""); strVal != "50%" {
		t.Errorf("wrong asr value: %s", strVal)
	}
}

func TestASRGetValue(t *testing.T) {
	asr, _ := NewASR()
	ev := engine.StatsEvent{
		"AnswerTime": time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
	}
	asr.AddEvent(ev)
	if v := asr.GetValue(); v != 100.0 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.AddEvent(engine.StatsEvent{})
	asr.AddEvent(engine.StatsEvent{})
	if v := asr.GetValue(); v != 33.33333 {
		t.Errorf("wrong asr value: %f", v)
	}
	asr.RemEvent(engine.StatsEvent{})
	if v := asr.GetValue(); v != 50.0 {
		t.Errorf("wrong asr value: %f", v)
	}
}
