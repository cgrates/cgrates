/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"github.com/cgrates/cgrates/history"
	"testing"
)

func TestHistory(t *testing.T) {
	scribe := historyScribe.(*history.MockScribe)
	expected := `[{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}}]
		[{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}},{"Key":"GERMANY_O2","Object":{"Id":"GERMANY_O2","Prefixes":["41"]}}]
		[{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}},{"Key":"GERMANY_O2","Object":{"Id":"GERMANY_O2","Prefixes":["41"]}},{"Key":"GERMANY_PREMIUM","Object":{"Id":"GERMANY_PREMIUM","Prefixes":["43"]}}]
		[{"Key":"ALL","Object":{"Id":"ALL","Prefixes":["49","41","43"]}},{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}},{"Key":"GERMANY_O2","Object":{"Id":"GERMANY_O2","Prefixes":["41"]}},{"Key":"GERMANY_PREMIUM","Object":{"Id":"GERMANY_PREMIUM","Prefixes":["43"]}}]
		[{"Key":"ALL","Object":{"Id":"ALL","Prefixes":["49","41","43"]}},{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}},{"Key":"GERMANY_O2","Object":{"Id":"GERMANY_O2","Prefixes":["41"]}},{"Key":"GERMANY_PREMIUM","Object":{"Id":"GERMANY_PREMIUM","Prefixes":["43"]}},{"Key":"NAT","Object":{"Id":"NAT","Prefixes":["0256","0257","0723"]}}]
		[{"Key":"ALL","Object":{"Id":"ALL","Prefixes":["49","41","43"]}},{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}},{"Key":"GERMANY_O2","Object":{"Id":"GERMANY_O2","Prefixes":["41"]}},{"Key":"GERMANY_PREMIUM","Object":{"Id":"GERMANY_PREMIUM","Prefixes":["43"]}},{"Key":"NAT","Object":{"Id":"NAT","Prefixes":["0256","0257","0723"]}},{"Key":"RET","Object":{"Id":"RET","Prefixes":["0723","0724"]}}]
		[{"Key":"ALL","Object":{"Id":"ALL","Prefixes":["49","41","43"]}},{"Key":"GERMANY","Object":{"Id":"GERMANY","Prefixes":["49"]}},{"Key":"GERMANY_O2","Object":{"Id":"GERMANY_O2","Prefixes":["41"]}},{"Key":"GERMANY_PREMIUM","Object":{"Id":"GERMANY_PREMIUM","Prefixes":["43"]}},{"Key":"NAT","Object":{"Id":"NAT","Prefixes":["0256","0257","0723"]}},{"Key":"RET","Object":{"Id":"RET","Prefixes":["0723","0724"]}},{"Key":"nat","Object":{"Id":"nat","Prefixes":["0257","0256","0723"]}}]`
	if scribe.Buf.String() != expected {
		t.Error("Error in history content: ", scribe.Buf.String())
	}
}
