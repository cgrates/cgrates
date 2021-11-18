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
	"testing"
)

func TestSetHTTPPstrTransport(t *testing.T) {
	tmp := httpPstrTransport
	SetHTTPPstrTransport(nil)
	if httpPstrTransport != nil {
		t.Error("Expected the transport to be nil", httpPstrTransport)
	}
	httpPstrTransport = tmp
}

func TestSetCdrStorage(t *testing.T) {
	tmp := cdrStorage
	SetCdrStorage(nil)
	if cdrStorage != nil {
		t.Error("Expected the cdrStorage to be nil", cdrStorage)
	}
	cdrStorage = tmp
}

func TestSetDataStorage(t *testing.T) {
	tmp := dm
	SetDataStorage(nil)
	if dm != nil {
		t.Error("Expected the dm to be nil", dm)
	}
	dm = tmp
}
