// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"testing"
)

func TestDynamicDataProviderProccesFieldPath(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": MapStorage{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": MapStorage{
				"SortedRoutes": []MapStorage{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	newpath, err := ProcessFieldPath("~*cgrep.Stir.<CHRG_;~*cgrep.Routes.SortedRoutes[1].ID;_END>", InfieldSep, dp)
	if err != nil {
		t.Fatal(err)
	}
	expectedPath := "~*cgrep.Stir.CHRG_ROUTE2_END"
	if newpath != expectedPath {
		t.Errorf("Expected: %q,received %q", expectedPath, newpath)
	}
	_, err = ProcessFieldPath("~*cgrep.Stir.<CHRG_;~*cgrep.Routes.SortedRoutes[1].ID;_END", InfieldSep, dp)
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = ProcessFieldPath("<*now>", InfieldSep, dp)
	if err != ErrNotFound {
		t.Errorf("Expected error %s received %v", ErrNotFound, err)
	}

	_, err = ProcessFieldPath("~*cgrep.Stir<CHRG_", InfieldSep, dp)
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = ProcessFieldPath("~*cgrep.Stir<CHRG_;~*cgrep.Routes.SortedRoutes[1].ID2;_END>", InfieldSep, dp)
	if err != ErrNotFound {
		t.Errorf("Expected error %s received %v", ErrNotFound, err)
	}
	newpath, err = ProcessFieldPath("~*cgrep.Stir[1]", InfieldSep, dp)
	if err != nil {
		t.Fatal(err)
	}
	if newpath != EmptyString {
		t.Errorf("Expected: %q,received %q", EmptyString, newpath)
	}
}

func TestDynamicDataProviderGetFullFieldPath(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": MapStorage{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": MapStorage{
				"SortedRoutes": []MapStorage{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	newpath, err := GetFullFieldPath("~*cgrep.Stir.<CHRG_;~*cgrep.Routes.SortedRoutes[1].ID;_END>.Something", dp)
	if err != nil {
		t.Fatal(err)
	}
	expectedPath := "~*cgrep.Stir.CHRG_ROUTE2_END.Something"
	if newpath.Path != expectedPath {
		t.Errorf("Expected: %q,received %q", expectedPath, newpath.Path)
	}
	_, err = GetFullFieldPath("~*cgrep.Stir<CHRG_;~*cgrep.Routes.SortedRoutes[1].ID;_END.Something", dp)
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = GetFullFieldPath("~*cgrep.Stir<CHRG_", dp)
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = GetFullFieldPath("~*cgrep.Stir<CHRG_;~*cgrep.Routes.SortedRoutes[1].ID2;_END>", dp)
	if err != ErrNotFound {
		t.Errorf("Expected error %s received %v", ErrNotFound, err)
	}
	newpath, err = GetFullFieldPath("~*cgrep.Stir[1]", dp)
	if err != nil {
		t.Fatal(err)
	}
	if newpath != nil {
		t.Errorf("Expected: %v,received %q", nil, newpath)
	}

	newpath, err = GetFullFieldPath("~*cgrep.Stir", dp)
	if err != nil {
		t.Fatal(err)
	}
	if newpath != nil {
		t.Errorf("Expected: %v,received %q", nil, newpath)
	}

}

func TestParseParamForDataProvider(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": MapStorage{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": MapStorage{
				"SortedRoutes": []MapStorage{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"AccountField": "Account",
			"Account":      "1001",
			"HalfAccount":  "01",
			"BestRoute":    0,
		},
	}

	tests := []struct {
		name              string
		param             string
		dP                DataProvider
		onlyEncapsulatead bool
		want              string
		expErr            error
	}{
		{
			name:              "With parameters",
			param:             "*string:~*req.+~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "*string:~*req.+~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
		},
		{
			name:              "With param that contains ANDsep",
			param:             "~*cgrep.AccountField+:10" + ANDSep + "~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "~*cgrep.AccountField+:10" + ANDSep + "~*cgrep.HalfAccount",
		},
		{
			name:              "With param that is not valid",
			param:             "<~*cgrep.AccountField+:10" + ANDSep + "~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "",
			expErr:            ErrWrongPath,
		},
		{
			name:              "Empty string ",
			param:             "",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "",
		},
		{
			name:              "With MetaDynReq as prefix and onlyEncapsulatead true",
			param:             MetaDynReq + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: true,
			want:              "~*req~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
		},
		{
			name:              "With DynamicDataPrefix as prefix and onlyEncapsulatead true",
			param:             DynamicDataPrefix + MetaOpts + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: true,
			want:              "~*opts~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
		},
		{
			name:              "With MetaDynReq as prefix and onlyEncapsulatead false",
			param:             MetaDynReq + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "",
			expErr:            ErrNotFound,
		},
		{
			name:              "With DynamicDataPrefix as prefix and onlyEncapsulatead false",
			param:             DynamicDataPrefix + MetaOpts + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "",
			expErr:            ErrNotFound,
		},
		{
			name:              "With MetaTenant as prefix and onlyEncapsulatead false",
			param:             MetaTenant + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "",
			expErr:            ErrNotFound,
		},
		{
			name:              "With MetaTenant as prefix and onlyEncapsulatead true",
			param:             MetaTenant + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
			dP:                dp,
			onlyEncapsulatead: true,
			want:              MetaTenant + "~*cgrep.AccountField+:10+~*cgrep.HalfAccount",
		},
		{
			name:              "With  PlusChar",
			param:             "<*string:~*req.+~*cgrep.AccountField" + PlusChar + ":10+~*cgrep.HalfAccount+>",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "*string:~*req.Account:1001",
		},
		{
			name:              "error case",
			param:             "<*string:~*req.+~*cgrep.AccountField+:10+~*cgrep.HalfAccount+",
			dP:                dp,
			onlyEncapsulatead: false,
			want:              "",
			expErr:            ErrWrongPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ParseParamForDataProvider(tt.param, tt.dP, tt.onlyEncapsulatead)
			if gotErr != nil && gotErr != tt.expErr {
				t.Errorf("Expected: %v,received %q", tt.expErr, gotErr)
			}
			if got != tt.want {
				t.Errorf("Expected: %v,received %q", tt.want, got)
			}
		})
	}
}
