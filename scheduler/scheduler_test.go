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
	sched := &Scheduler{actStatsInterval: time.Millisecond, actSuccessStats: make(map[string]map[time.Time]bool)}
	sched.updateActStats(&engine.Action{Id: "REMOVE_1", ActionType: utils.MetaRemoveAccount}, false)
	if len(sched.actSuccessStats[utils.MetaRemoveAccount]) != 1 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats[utils.MetaRemoveAccount])
	}
	sched.updateActStats(&engine.Action{Id: "REMOVE_2", ActionType: utils.MetaRemoveAccount}, false)
	if len(sched.actSuccessStats[utils.MetaRemoveAccount]) != 2 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats[utils.MetaRemoveAccount])
	}
	sched.updateActStats(&engine.Action{Id: "LOG1", ActionType: utils.MetaLog}, false)
	if len(sched.actSuccessStats[utils.MetaLog]) != 1 ||
		len(sched.actSuccessStats[utils.MetaRemoveAccount]) != 2 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats)
	}
	time.Sleep(sched.actStatsInterval)
	sched.updateActStats(&engine.Action{Id: "REMOVE_3", ActionType: utils.MetaRemoveAccount}, false)
	if len(sched.actSuccessStats[utils.MetaRemoveAccount]) != 1 || len(sched.actSuccessStats) != 1 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats)
	}
}
