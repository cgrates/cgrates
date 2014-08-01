/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package apier

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Interact with Stats server
type CDRStatsV1 struct {
	CdrStats *engine.Stats
}

type AttrGetMetrics struct {
	StatsQueueId string // Id of the stats instance queried
}

func (sts *CDRStatsV1) GetMetrics(attr AttrGetMetrics, reply *map[string]float64) error {
	if len(attr.StatsQueueId) == 0 {
		return fmt.Errorf("%s:StatsInstanceId", utils.ERR_MANDATORY_IE_MISSING)
	}
	return sts.CdrStats.GetValues(attr.StatsQueueId, reply)
}

func (sts *CDRStatsV1) GetQueueIds(empty string, reply *[]string) error {
	return sts.CdrStats.GetQueueIds(0, reply)
}

type AttrReloadStatsQueues struct {
	StatsQueueIds []string
}

func (sts *CDRStatsV1) ReloadStatsQueues(attr AttrReloadStatsQueues, reply *string) error {
	*reply = utils.OK
	return nil
}
