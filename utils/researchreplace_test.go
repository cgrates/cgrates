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

package utils

import (
	"regexp"
	"testing"
)

func TestProcessReSearchReplace(t *testing.T) {
	rsr := &ReSearchReplace{regexp.MustCompile(`sip:\+49(\d+)@(\d*\.\d*\.\d*\.\d*)`), "0$1@$2"}
	source := "<sip:+4986517174963@127.0.0.1;transport=tcp>"
	expectOut := "086517174963@127.0.0.1"
	if outStr := rsr.Process(source); outStr != expectOut {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}

func TestProcessReSearchReplace2(t *testing.T) {
	rsr := &ReSearchReplace{regexp.MustCompile(`(\d+)`), "+$1"}
	source := "4986517174963"
	expectOut := "+4986517174963"
	if outStr := rsr.Process(source); outStr != expectOut {
		t.Error("Unexpected output from SearchReplace: ", outStr)
	}
}
