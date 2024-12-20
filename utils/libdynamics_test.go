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

package utils

import (
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
)

func TestCloneDynamicStringsSliceOpt(t *testing.T) {
	in := []*DynamicStringSliceOpt{
		{
			Values:    []string{"VAL_1", "VAL_2"},
			FilterIDs: []string{"fltr1"},
		},
		{
			Values:    []string{"VAL_3", "VAL_4"},
			FilterIDs: []string{"fltr2"},
		},
	}

	clone := CloneDynamicStringSliceOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicStringOpt(t *testing.T) {
	in := []*DynamicStringOpt{
		{
			Value:     "VAL_1",
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     "VAL_2",
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicStringOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicInterfaceOpt(t *testing.T) {
	in := []*DynamicInterfaceOpt{
		{
			Value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     "VAL_2",
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicInterfaceOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicBoolOpt(t *testing.T) {
	in := []*DynamicBoolOpt{
		{
			Value:     true,
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     false,
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicBoolOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicIntOpt(t *testing.T) {
	in := []*DynamicIntOpt{
		{
			Value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicIntOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicFloat64Opt(t *testing.T) {
	in := []*DynamicFloat64Opt{
		{
			Value:     1.3,
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     2.7,
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicFloat64Opt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicDurationOpt(t *testing.T) {
	in := []*DynamicDurationOpt{
		{
			Value:     time.Second * 2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     time.Second * 5,
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicDurationOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicDecimalBigOpt(t *testing.T) {
	in := []*DynamicDecimalBigOpt{
		{
			Value:     decimal.WithContext(DecimalContext).SetUint64(10),
			FilterIDs: []string{"fltr1"},
		},
		{
			Value:     decimal.WithContext(DecimalContext).SetUint64(2),
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicDecimalBigOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

// Equals
func TestDynamicStringSliceOptEqual(t *testing.T) {
	v1 := []*DynamicStringSliceOpt{
		{
			Tenant:    "cgrates.org",
			Values:    []string{"VAL_1", "VAL_2"},
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Values:    []string{"VAL_3", "VAL_4"},
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicStringSliceOpt{
		{
			Tenant:    "cgrates.org",
			Values:    []string{"VAL_1", "VAL_2"},
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Values:    []string{"VAL_3", "VAL_4"},
			FilterIDs: []string{"fltr2"},
		},
	}

	//Test if equal
	if !DynamicStringSliceOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	//Test if different
	v1[0].Values = append(v1[0].Values, "VAL_3")
	if DynamicStringSliceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicStringSliceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicStringSliceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicStringSliceOpt{
		Values:    []string{"VAL_1", "VAL_2"},
		FilterIDs: []string{"fltr1"},
	})
	if DynamicStringSliceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicStringOptEqual(t *testing.T) {
	v1 := []*DynamicStringOpt{
		{
			Tenant:    "cgrates.org",
			Value:     "VAL_1",
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     "VAL_2",
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicStringOpt{
		{
			Tenant:    "cgrates.org",
			Value:     "VAL_1",
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     "VAL_2",
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicStringOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = "VAL_3"
	if DynamicStringOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicStringOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicStringOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicStringOpt{
		Value:     "NEW_VAL",
		FilterIDs: []string{"fltr1"},
	})
	if DynamicStringOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicBoolOptEquals(t *testing.T) {
	v1 := []*DynamicBoolOpt{
		{
			Tenant:    "cgrates.org",
			Value:     true,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     false,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicBoolOpt{
		{
			Tenant:    "cgrates.org",
			Value:     true,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     false,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicBoolOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = false
	if DynamicBoolOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicBoolOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicBoolOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicBoolOpt{
		Value:     true,
		FilterIDs: []string{"fltr1"},
	})
	if DynamicBoolOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicIntOptEqual(t *testing.T) {
	v1 := []*DynamicIntOpt{
		{
			Tenant:    "cgrates.org",
			Value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicIntOpt{
		{
			Tenant:    "cgrates.org",
			Value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicIntOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = 2
	if DynamicIntOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicIntOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicIntOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicIntOpt{
		Value:     2,
		FilterIDs: []string{"fltr1"},
	})
	if DynamicIntOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicFloat64OptEqual(t *testing.T) {
	v1 := []*DynamicFloat64Opt{
		{
			Tenant:    "cgrates.org",
			Value:     1.2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     2.6,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicFloat64Opt{
		{
			Tenant:    "cgrates.org",
			Value:     1.2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     2.6,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicFloat64OptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = 2.8
	if DynamicFloat64OptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicFloat64OptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicFloat64OptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicFloat64Opt{
		Value:     3.5,
		FilterIDs: []string{"fltr1"},
	})
	if DynamicFloat64OptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicDurationOptEquals(t *testing.T) {
	v1 := []*DynamicDurationOpt{
		{
			Tenant:    "cgrates.org",
			Value:     time.Second * 2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     time.Second * 5,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicDurationOpt{
		{
			Tenant:    "cgrates.org",
			Value:     time.Second * 2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     time.Second * 5,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = time.Second * 11
	if DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicDurationOpt{
		Value:     time.Second * 2,
		FilterIDs: []string{"fltr1"},
	})
	if DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicDecimalBigOptEquals(t *testing.T) {
	v1 := []*DynamicDecimalBigOpt{
		{
			Tenant:    "cgrates.org",
			Value:     decimal.WithContext(DecimalContext).SetUint64(10),
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     decimal.WithContext(DecimalContext).SetUint64(2),
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicDecimalBigOpt{
		{
			Tenant:    "cgrates.org",
			Value:     decimal.WithContext(DecimalContext).SetUint64(10),
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     decimal.WithContext(DecimalContext).SetUint64(2),
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicDecimalBigOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = decimal.WithContext(DecimalContext).SetUint64(16)
	if DynamicDecimalBigOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicDecimalBigOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicDecimalBigOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicDecimalBigOpt{
		Value:     decimal.WithContext(DecimalContext).SetUint64(10),
		FilterIDs: []string{"fltr1"},
	})
	if DynamicDecimalBigOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicInterfaceOptEqual(t *testing.T) {
	v1 := []*DynamicInterfaceOpt{
		{
			Tenant:    "cgrates.org",
			Value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicInterfaceOpt{
		{
			Tenant:    "cgrates.org",
			Value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			Value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicInterfaceOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].Value = 2
	if DynamicInterfaceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicInterfaceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicInterfaceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicInterfaceOpt{
		Value:     2,
		FilterIDs: []string{"fltr1"},
	})
	if DynamicInterfaceOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestStringToDecimalBigDynamicOpts(t *testing.T) {
	dsOpt := []*DynamicStringOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     "200",
		},
	}

	exp := []*DynamicDecimalBigOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     decimal.WithContext(DecimalContext).SetUint64(200),
		},
	}

	rcv, err := StringToDecimalBigDynamicOpts(dsOpt)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	//Check conversion error
	errExpect := "can't convert <this_is_definitely_a_decimal_big> to decimal"
	dsOpt[0].Value = "this_is_definitely_a_decimal_big"
	if _, err := StringToDecimalBigDynamicOpts(dsOpt); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}

func TestDynamicIntPointerOptEqual(t *testing.T) {
	v1 := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     IntPointer(200),
		},
	}

	v2 := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     IntPointer(200),
		},
	}

	if !DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected items to match")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different values
	v1[0].Value = IntPointer(500)
	if DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Value = IntPointer(200)

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicIntPointerOpt{
		Value:     IntPointer(2),
		FilterIDs: []string{"fltr1"},
	})
	if DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicDurationPointerOptEqual(t *testing.T) {
	v1 := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"fld3"},
			Tenant:    "cgrates.org",
			Value:     DurationPointer(3 * time.Second),
		},
	}

	v2 := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"fld3"},
			Tenant:    "cgrates.org",
			Value:     DurationPointer(3 * time.Second),
		},
	}

	if !DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected items to match")
	}
	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different values
	v1[0].Value = DurationPointer(4 * time.Second)
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Value = DurationPointer(3 * time.Second)

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicDurationPointerOpt{
		Value:     DurationPointer(2),
		FilterIDs: []string{"fltr1"},
	})
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDecimalBigToStringDynamicOpts(t *testing.T) {
	dbOpt := []*DynamicDecimalBigOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     decimal.WithContext(DecimalContext).SetUint64(300),
		},
	}

	exp := []*DynamicStringOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "300",
		},
	}

	rcv := DecimalBigToStringDynamicOpts(dbOpt)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestStringToDurationDynamicOpts(t *testing.T) {
	sOpts := []*DynamicStringOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "50s",
		},
	}

	exp := []*DynamicDurationOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     50 * time.Second,
		},
	}

	rcv, err := StringToDurationDynamicOpts(sOpts)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	//Check conversion error
	errExpect := `time: unknown unit "c" in duration "50c"`
	sOpts[0].Value = "50c"
	if _, err := StringToDurationDynamicOpts(sOpts); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}

func TestDurationToStringDynamicOpts(t *testing.T) {
	exp := []*DynamicStringOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "50s",
		},
	}

	sOpts := []*DynamicDurationOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     50 * time.Second,
		},
	}

	rcv := DurationToStringDynamicOpts(sOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestIntToIntPointerDynamicOpts(t *testing.T) {
	iOpts := []*DynamicIntOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     50,
		},
	}

	exp := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     IntPointer(50),
		},
	}

	rcv := IntToIntPointerDynamicOpts(iOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestIntPointerToIntDynamicOpts(t *testing.T) {
	exp := []*DynamicIntOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     50,
		},
	}

	iOpts := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     IntPointer(50),
		},
	}

	rcv := IntPointerToIntDynamicOpts(iOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestStringToDurationPointerDynamicOpts(t *testing.T) {
	sOpts := []*DynamicStringOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "50s",
		},
	}

	exp := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     DurationPointer(50 * time.Second),
		},
	}

	rcv, err := StringToDurationPointerDynamicOpts(sOpts)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	errExpect := `time: unknown unit "c" in duration "50c"`
	sOpts[0].Value = "50c"
	if _, err := StringToDurationPointerDynamicOpts(sOpts); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}
func TestDurationPointerToStringDynamicOpts(t *testing.T) {
	exp := []*DynamicStringOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "50s",
		},
	}

	dpOpts := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     DurationPointer(50 * time.Second),
		},
	}

	rcv := DurationPointerToStringDynamicOpts(dpOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestCloneDynamicIntPointerOpt(t *testing.T) {
	ipOpt := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     IntPointer(200),
		},
	}

	exp := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     IntPointer(200),
		},
	}

	rcv := CloneDynamicIntPointerOpt(ipOpt)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestCloneDynamicDurationPointerOpt(t *testing.T) {
	dpOpts := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     DurationPointer(50 * time.Second),
		},
	}

	exp := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     DurationPointer(50 * time.Second),
		},
	}

	rcv := CloneDynamicDurationPointerOpt(dpOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}
