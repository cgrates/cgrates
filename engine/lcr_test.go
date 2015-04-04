/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"sort"
	"testing"
)

func TestLcrQOSSorter(t *testing.T) {
	s := QOSSorter{
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 3,
				"ACD": 3,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 1,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 2,
				"ACD": 2,
			},
			qosSortParams: []string{ASR, ACD},
		},
	}
	sort.Sort(s)
	if s[0].QOS[ASR] != 1 ||
		s[1].QOS[ASR] != 2 ||
		s[2].QOS[ASR] != 3 {
		t.Error("Lcr qos sort failed: ", s)
	}
}

func TestLcrQOSSorterOACD(t *testing.T) {
	s := QOSSorter{
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 3,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 1,
			},
			qosSortParams: []string{ASR, ACD},
		},
		&LCRSupplierCost{
			QOS: map[string]float64{
				"ASR": 1,
				"ACD": 2,
			},
			qosSortParams: []string{ASR, ACD},
		},
	}
	sort.Sort(s)
	if s[0].QOS[ACD] != 1 ||
		s[1].QOS[ACD] != 2 ||
		s[2].QOS[ACD] != 3 {
		t.Error("Lcr qos sort failed: ", s)
	}
}
