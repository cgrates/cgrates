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

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
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
			value:     "VAL_1",
			FilterIDs: []string{"fltr1"},
		},
		{
			value:     "VAL_2",
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
			value:     true,
			FilterIDs: []string{"fltr1"},
		},
		{
			value:     false,
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
			value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			value:     2,
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
			value:     1.3,
			FilterIDs: []string{"fltr1"},
		},
		{
			value:     2.7,
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
			value:     time.Second * 2,
			FilterIDs: []string{"fltr1"},
		},
		{
			value:     time.Second * 5,
			FilterIDs: []string{"fltr2"},
		},
	}
	clone := CloneDynamicDurationOpt(in)
	if !reflect.DeepEqual(in, clone) {
		t.Error("Expected objects to match")
	}
}

func TestCloneDynamicDecimalBigOpt(t *testing.T) {
	in := []*DynamicDecimalOpt{
		NewDynamicDecimalOpt([]string{"fltr1"}, "", decimal.WithContext(utils.DecimalContext).SetUint64(10), nil),
		NewDynamicDecimalOpt([]string{"fltr2"}, "", decimal.WithContext(utils.DecimalContext).SetUint64(2), nil),
	}
	clone := CloneDynamicDecimalOpt(in)
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
			value:     "VAL_1",
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     "VAL_2",
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicStringOpt{
		{
			Tenant:    "cgrates.org",
			value:     "VAL_1",
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     "VAL_2",
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicStringOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].value = "VAL_3"
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
		value:     "NEW_VAL",
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
			value:     true,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     false,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicBoolOpt{
		{
			Tenant:    "cgrates.org",
			value:     true,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     false,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicBoolOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].value = false
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
		value:     true,
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
			value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicIntOpt{
		{
			Tenant:    "cgrates.org",
			value:     1,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     2,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicIntOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].value = 2
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
		value:     2,
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
			value:     1.2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     2.6,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicFloat64Opt{
		{
			Tenant:    "cgrates.org",
			value:     1.2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     2.6,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicFloat64OptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].value = 2.8
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
		value:     3.5,
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
			value:     time.Second * 2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     time.Second * 5,
			FilterIDs: []string{"fltr2"},
		},
	}

	v2 := []*DynamicDurationOpt{
		{
			Tenant:    "cgrates.org",
			value:     time.Second * 2,
			FilterIDs: []string{"fltr1"},
		},
		{
			Tenant:    "cgrates.org",
			value:     time.Second * 5,
			FilterIDs: []string{"fltr2"},
		},
	}

	if !DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	v1[0].value = time.Second * 11
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
		value:     time.Second * 2,
		FilterIDs: []string{"fltr1"},
	})
	if DynamicDurationOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDynamicDecimalBigOptEquals(t *testing.T) {
	v1 := []*DynamicDecimalOpt{
		NewDynamicDecimalOpt([]string{"fltr1"}, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(10), nil),
		NewDynamicDecimalOpt([]string{"fltr2"}, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(2), nil),
	}

	v2 := []*DynamicDecimalOpt{
		NewDynamicDecimalOpt([]string{"fltr1"}, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(10), nil),
		NewDynamicDecimalOpt([]string{"fltr2"}, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(2), nil),
	}

	if !DynamicDecimalOptEqual(v1, v2) {
		t.Error("Expected both slices to be the same")
	}

	//Test if different tenants
	v1[0].Tenant = "cgrates.net"
	if DynamicDecimalOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].Tenant = "cgrates.org"

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicDecimalOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, NewDynamicDecimalOpt([]string{"fltr1"}, "", decimal.WithContext(utils.DecimalContext).SetUint64(10), nil))
	if DynamicDecimalOptEqual(v1, v2) {
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

func TestIfaceToDecimalBigDynamicOpts(t *testing.T) {
	dsOpt := []*DynamicInterfaceOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			Value:     "200",
		},
	}

	exp := []*DynamicDecimalOpt{
		NewDynamicDecimalOpt([]string{"fld1", "fld2"}, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(200), nil),
	}

	rcv, err := IfaceToDecimalBigDynamicOpts(dsOpt)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	//Check conversion error
	errExpect := "can't convert <this_is_definitely_a_decimal_big> to decimal"
	dsOpt[0].Value = "this_is_definitely_a_decimal_big"
	if _, err := IfaceToDecimalBigDynamicOpts(dsOpt); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}

func TestDynamicIntPointerOptEqual(t *testing.T) {
	v1 := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			value:     utils.IntPointer(200),
		},
	}

	v2 := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			value:     utils.IntPointer(200),
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
	v1[0].value = utils.IntPointer(500)
	if DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].value = utils.IntPointer(200)

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicIntPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicIntPointerOpt{
		value:     utils.IntPointer(2),
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
			value:     utils.DurationPointer(3 * time.Second),
		},
	}

	v2 := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"fld3"},
			Tenant:    "cgrates.org",
			value:     utils.DurationPointer(3 * time.Second),
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
	v1[0].value = utils.DurationPointer(4 * time.Second)
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
	v1[0].value = utils.DurationPointer(3 * time.Second)

	//Test if different filters
	v1[0].FilterIDs = append(v1[0].FilterIDs, "new_fltr")
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}

	//Test if different lengths
	v1 = append(v1, &DynamicDurationPointerOpt{
		value:     utils.DurationPointer(2),
		FilterIDs: []string{"fltr1"},
	})
	if DynamicDurationPointerOptEqual(v1, v2) {
		t.Error("Expected slices to differ")
	}
}

func TestDecimalBigToStringDynamicOpts(t *testing.T) {
	dbOpt := []*DynamicDecimalOpt{
		NewDynamicDecimalOpt([]string{"test_filter", "test_filter2"}, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(300), nil),
	}

	exp := []*DynamicInterfaceOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "300",
		},
	}

	rcv := DecimalToIfaceDynamicOpts(dbOpt)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestIfaceToDurationDynamicOpts(t *testing.T) {
	sOpts := []*DynamicInterfaceOpt{
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
			value:     50 * time.Second,
		},
	}

	rcv, err := IfaceToDurationDynamicOpts(sOpts)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	//Check conversion error
	errExpect := `time: unknown unit "c" in duration "50c"`
	sOpts[0].Value = "50c"
	if _, err := IfaceToDurationDynamicOpts(sOpts); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}

func TestDurationToIfaceDynamicOpts(t *testing.T) {
	exp := []*DynamicInterfaceOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     50 * time.Second,
		},
	}

	sOpts := []*DynamicDurationOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			value:     50 * time.Second,
		},
	}

	rcv := DurationToIfaceDynamicOpts(sOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestIntToIntPointerDynamicOpts(t *testing.T) {
	iOpts := []*DynamicInterfaceOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     "50",
		},
	}
	exp := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			value:     utils.IntPointer(50),
		},
	}
	rcv, err := IfaceToIntPointerDynamicOpts(iOpts)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %+v \n but received \n %+v", *exp[0], *rcv[0])
	}
}

func TestIntPointerToIntDynamicOpts(t *testing.T) {
	exp := []*DynamicInterfaceOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     utils.IntPointer(50),
		},
	}

	iOpts := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			value:     utils.IntPointer(50),
		},
	}

	rcv := IntPointerToIfaceDynamicOpts(iOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestStringToDurationPointerDynamicOpts(t *testing.T) {
	sOpts := []*DynamicInterfaceOpt{
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
			value:     utils.DurationPointer(50 * time.Second),
		},
	}

	rcv, err := IfaceToDurationPointerDynamicOpts(sOpts)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}

	errExpect := `time: unknown unit "c" in duration "50c"`
	sOpts[0].Value = "50c"
	if _, err := IfaceToDurationPointerDynamicOpts(sOpts); err == nil || err.Error() != errExpect {
		t.Error(err)
	}
}
func TestDurationPointerToIfaceDynamicOpts(t *testing.T) {
	exp := []*DynamicInterfaceOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			Value:     utils.DurationPointer(50 * time.Second),
		},
	}

	dpOpts := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			value:     utils.DurationPointer(50 * time.Second),
		},
	}

	rcv := DurationPointerToIfaceDynamicOpts(dpOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}

func TestCloneDynamicIntPointerOpt(t *testing.T) {
	ipOpt := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			value:     utils.IntPointer(200),
		},
	}

	exp := []*DynamicIntPointerOpt{
		{
			FilterIDs: []string{"fld1", "fld2"},
			Tenant:    "cgrates.org",
			value:     utils.IntPointer(200),
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
			value:     utils.DurationPointer(50 * time.Second),
		},
	}

	exp := []*DynamicDurationPointerOpt{
		{
			FilterIDs: []string{"test_filter", "test_filter2"},
			Tenant:    "cgrates.org",
			value:     utils.DurationPointer(50 * time.Second),
		},
	}

	rcv := CloneDynamicDurationPointerOpt(dpOpts)
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	}
}
