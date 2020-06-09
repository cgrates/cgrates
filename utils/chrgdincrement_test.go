/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package utils

/*
func TestJSONMarshalUnmarshal(t *testing.T) {
	incrmt := &ChargedIncrement{
		Usage:          time.Duration(1 * time.Hour),
		Cost:           NewDecimalFromFloat64(2.13),
		AccountingID:   "abbsjweejrmdhfr",
		CompressFactor: 1,
	}
	jsn, err := json.Marshal(incrmt)
	if err != nil {
		t.Error(err)
	}
	var uIncrmnt ChargedIncrement
	if err := json.Unmarshal(jsn, &uIncrmnt); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(incrmt, &uIncrmnt) {
		t.Errorf("expecting: %+v, received: %+v", incrmt, uIncrmnt)
	}
	incrmt = &ChargedIncrement{
		Usage:          time.Duration(1 * time.Hour),
		AccountingID:   "abbsjweejrmdhfr",
		CompressFactor: 1,
	}
	if jsn, err = json.Marshal(incrmt); err != nil {
		t.Error(err)
	}
	if err := json.Unmarshal(jsn, &uIncrmnt); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(incrmt, &uIncrmnt) {
		t.Errorf("expecting: %+v, received: %+v", incrmt, uIncrmnt)
	}
}
*/
