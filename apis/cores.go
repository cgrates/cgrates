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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/utils"
)

func NewCoreSv1(cS *cores.CoreService) *CoreSv1 {
	return &CoreSv1{cS: cS}
}

// CoreSv1 exports RPC from RLs
type CoreSv1 struct {
	cS *cores.CoreService
	ping
}

func (cS *CoreSv1) Status(_ *context.Context, arg *utils.TenantWithAPIOpts, reply *map[string]interface{}) error {
	return cS.cS.Status(arg, reply)
}

// Sleep is used to test the concurrent requests mechanism
func (cS *CoreSv1) Sleep(_ *context.Context, arg *utils.DurationArgs, reply *string) error {
	time.Sleep(arg.Duration)
	*reply = utils.OK
	return nil
}
