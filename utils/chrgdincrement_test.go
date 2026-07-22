// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

/*
func TestJSONMarshalUnmarshal(t *testing.T) {
	incrmt := &ChargedIncrement{
		Usage:          time.Hour,
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
		Usage:          time.Hour,
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
