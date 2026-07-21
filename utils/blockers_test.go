// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"reflect"
	"testing"
)

func TestNewBlockersFromString(t *testing.T) {
	blkrs := DynamicBlockers{
		{
			FilterIDs: []string{"*string:~*opts.*cost:0"},
			Blocker:   false,
		},
		{
			FilterIDs: []string{"*suffix:~*req.Destination:+4432", "eq:~*opts.*usage:10s"},
			Blocker:   false,
		},
		{
			FilterIDs: []string{"*notstring:~*req.RequestType:*prepaid"},
			Blocker:   true,
		},
		{
			Blocker: false,
		},
	}
	blkrsStr := "*string:~*opts.*cost:0;false;*suffix:~*req.Destination:+4432&eq:~*opts.*usage:10s;false;*notstring:~*req.RequestType:*prepaid;true;;false"
	if rcv, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, blkrs) {
		t.Errorf("Expected %v \n received %v", ToJSON(blkrs), ToJSON(rcv))
	}
	blkrs = DynamicBlockers{
		{
			FilterIDs: nil,
			Blocker:   false,
		},
	}
	if received, err := NewDynamicBlockersFromString("", InfieldSep, ANDSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(received, blkrs) {
		t.Errorf("Expected %v \n received %v", ToJSON(blkrs), ToJSON(received))
	}

}

func TestNewBlockersFromString2(t *testing.T) {
	blkrs := DynamicBlockers{
		{
			FilterIDs: []string{"*string:~*opts.*cost:0"},
			Blocker:   false,
		},
		{},
	}
	blkrsStr := "*string:~*opts.*cost:0;false;;"
	if rcv, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, blkrs) {
		t.Errorf("Expected %+v \n ,received %+v", ToJSON(blkrs), ToJSON(rcv))
	}
}

func TestNewBlockersFromStringErrSeparator(t *testing.T) {
	blkrsStr := "*string:~*opts.*cost:0;false;;;"
	exp := "invalid DynamicBlocker format for string <*string:~*opts.*cost:0;false;;;>"
	if _, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err.Error() != exp {
		t.Errorf("Expected %s \n received %s", exp, err.Error())
	}
}

func TestNewBlockersFromStringFormatBool(t *testing.T) {
	blkrsStr := "*string:~*opts.*cost:0;tttrrruuue"
	exp := "cannot convert bool with value: <tttrrruuue> into Blocker"
	if _, err := NewDynamicBlockersFromString(blkrsStr, InfieldSep, ANDSep); err.Error() != exp {
		t.Errorf("Expected %s \n received %s", exp, err.Error())
	}
}

func TestBlockersClone(t *testing.T) {

	blckrs := DynamicBlockers{
		{
			FilterIDs: []string{"*string:~*opts.*cost:0"},
			Blocker:   false,
		},
	}

	exp := DynamicBlockers{
		{
			FilterIDs: []string{"*string:~*opts.*cost:0"},
			Blocker:   false,
		},
	}
	if rcv := blckrs.Clone(); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n received %v", ToJSON(exp), ToJSON(rcv))

	}
}

func TestBlockersString(t *testing.T) {
	blkrs := DynamicBlockers{}

	exp := ""

	if rcv := blkrs.String("", ""); !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n received %v, length of blksr is %v", ToJSON(exp), ToJSON(rcv), len(blkrs))

	}
	blkrsTrue := DynamicBlocker{

		Blocker: true,
	}
	expTrue := "true"
	if rcvTrue := blkrsTrue.String("", ""); !reflect.DeepEqual(rcvTrue, expTrue) {
		t.Errorf("Expected %v \n received %v", ToJSON(expTrue), ToJSON(rcvTrue))

	}

}
