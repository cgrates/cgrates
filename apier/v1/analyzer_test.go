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

package v1

import (
	"testing"

	"github.com/cgrates/cgrates/analyzers"
)

func TestNewAnalyzerSv1(t *testing.T) {
	analyzerService := &analyzers.AnalyzerService{}
	analyzerSv1 := NewAnalyzerSv1(analyzerService)
	if analyzerSv1 == nil {
		t.Errorf("expected non-nil AnalyzerSv1, got nil")
	}
	if analyzerSv1.aS != analyzerService {
		t.Errorf("expected AnalyzerService to be %v, got %v", analyzerService, analyzerSv1.aS)
	}
}

func TestVerifyFormat(t *testing.T) {
	tests := []struct {
		tStr         string
		expectedBool bool
	}{

		{"12:34:56", true},
		{"23:59:59", true},
		{"12:34", false},
		{"12:34:56:78", false},
		{"12:abc:56", false},
		{"123:456:789", false},
		{"00:00:00", true},
		{"12:34:56", true},
		{"t:01:t", false},
		{"1,1,1", false},
		{"0:0:0", true},
		{"119911", false},
		{"00/01/03", false},
		{"t1:t2:t3", false},
	}

	for _, tt := range tests {
		t.Run(tt.tStr, func(t *testing.T) {
			result := verifyFormat(tt.tStr)
			if result != tt.expectedBool {
				t.Errorf("verifyFormat(%q) = %v; want %v", tt.tStr, result, tt.expectedBool)
			}
		})
	}
}