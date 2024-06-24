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

	"github.com/cgrates/cgrates/utils"
)

type SagProfileWithAPIOpts struct {
	*SagProfile
	APIOpts map[string]any
}

type SagProfile struct {
	Tenant            string
	ID                string
	QueryInterval     time.Duration
	StatIDs           []string
	MetricIDs         []string
	Sorting           string
	SortingParameters []string
	ThresholdIDs      []string
}

func (sgp *SagProfile) TenantID() string {
	return utils.ConcatenatedKey(sgp.Tenant, sgp.ID)
}
