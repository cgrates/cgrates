/*
Real-time Charging System for Telecom & ISP environments
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

package sessionmanager

import (
	"testing"

	"github.com/cgrates/cgrates/utils"
)

var kamEv = KamEvent{"callid": "ODVkMDI2Mzc2MDY5N2EzODhjNTAzNTdlODhiZjRlYWQ", "from_tag": "eb082607", "to_tag": "4ea9687f", "cgr_account": "dan",
	"cgr_reqtype": "prepaid", "cgr_subject": "dan", "cgr_destination": "+4986517174963", "cgr_tenant": "itsyscom.com",
	"cgr_duration": "20", "extra1": "val1", "extra2": "val2"}

func TestKamailioEventInterface(t *testing.T) {
	var _ utils.Event = utils.Event(kamEv)
}
