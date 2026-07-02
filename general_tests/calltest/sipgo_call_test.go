//go:build call

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

package calltest

import (
	"testing"
	"time"
)

func TestSipgoUACBasicCall(t *testing.T) {
	SipgoUAS{Port: 5094}.Start(t)
	SipgoUAC{Addr: "127.0.0.1:5094"}.Call(t, CallParams{
		To:       "test",
		From:     "1001",
		HoldTime: time.Second,
	})
}
