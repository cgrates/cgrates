/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"testing"
)

func TestBiRpcNewBiJSONrpcClient(t *testing.T) {
	rcv, err := NewBiJSONrpcClient("test", nil)

	if err != nil {
		if err.Error() != "dial tcp: address test: missing port in address" {
			t.Error(err)
		}
	}

	if rcv != nil {
		t.Error(err)
	}
}
