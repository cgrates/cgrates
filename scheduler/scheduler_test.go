// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package scheduler

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSchedulerUpdateActStats(t *testing.T) {
	sched := &Scheduler{actStatsInterval: 10 * time.Millisecond, actSuccessStats: make(map[string]map[time.Time]bool)}
	sched.updateActStats(&engine.Action{Id: "REMOVE_1", ActionType: utils.REMOVE_ACCOUNT}, false)
	if len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 1 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats[utils.REMOVE_ACCOUNT])
	}
	sched.updateActStats(&engine.Action{Id: "REMOVE_2", ActionType: utils.REMOVE_ACCOUNT}, false)
	if len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 2 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats[utils.REMOVE_ACCOUNT])
	}
	sched.updateActStats(&engine.Action{Id: "LOG1", ActionType: utils.LOG}, false)
	if len(sched.actSuccessStats[utils.LOG]) != 1 ||
		len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 2 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats)
	}
	time.Sleep(sched.actStatsInterval)
	sched.updateActStats(&engine.Action{Id: "REMOVE_3", ActionType: utils.REMOVE_ACCOUNT}, false)
	if len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 1 || len(sched.actSuccessStats) != 1 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats)
	}
}
