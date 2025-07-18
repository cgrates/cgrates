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
)

func TestActionAPDiktatRSRValues(t *testing.T) {
	apdDiktat := APDiktat{
		valRSR: RSRParsers{
			&RSRParser{
				Rules: ">;q=0.7;expires=3600",
			},
			&RSRParser{
				Rules: ">;q=0.7;expires=3600",
			},
		},
	}
	rsrPars, err := apdDiktat.RSRValues()
	if err != nil {
		t.Error(err)
	}
	expected := RSRParsers{
		&RSRParser{
			Rules: ">;q=0.7;expires=3600",
		},
		&RSRParser{
			Rules: ">;q=0.7;expires=3600",
		},
	}
	if !reflect.DeepEqual(rsrPars, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", ToJSON(expected), ToJSON(rsrPars))

	}
}

func TestActionAPDiktatRSRValuesNil(t *testing.T) {
	apdDiktat := APDiktat{}
	rsrPars, err := apdDiktat.RSRValues()
	if err != nil {
		t.Error(err)
	}
	var expected RSRParsers
	if !reflect.DeepEqual(rsrPars, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", ToJSON(expected), ToJSON(rsrPars))

	}
}

func TestActionAPDiktatRSRValuesError(t *testing.T) {
	apdDiktat := APDiktat{
		Opts: map[string]any{
			"*balanceValue": "val`val2val3",
		},
	}
	expErr := "Closed unspilit syntax"
	_, err := apdDiktat.RSRValues()
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}
func TestAPDiktatRSRValues(t *testing.T) {
	dk := &APDiktat{
		Opts: map[string]any{
			"*balanceValue": "1001",
		},
	}
	if rply, err := dk.RSRValues(); err != nil {
		return
	} else if exp := NewRSRParsersMustCompile("1001", InfieldSep); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %+v ,received: %+v", exp, rply)
	}
}

func TestActionProfileSet(t *testing.T) {
	ap := ActionProfile{Targets: make(map[string]StringSet)}
	exp := ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Schedule:  MetaNow,
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Targets: map[string]StringSet{
			MetaAccounts:   NewStringSet([]string{"1001", "1002"}),
			MetaThresholds: NewStringSet([]string{"TH1", "TH2"}),
		},
		Actions: []*APAction{{
			ID:        "acc1",
			Type:      "val1",
			FilterIDs: []string{"fltr1"},
			TTL:       10,
			Opts: map[string]any{
				"opt0": "val1",
				"opt1": "val1",
				"opt2": "val1",
				"opt3": MapStorage{"opt4": "val1"},
			},
			Weights: DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Diktats: []*APDiktat{{
				ID:        "dID",
				FilterIDs: []string{"fltr1"},
				Opts: map[string]any{
					"*balancePath":  "path",
					"*balanceValue": "val1",
				},
				Weights: DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: DynamicBlockers{
					{
						Blocker: true,
					},
				},
			}},
		}},
	}
	if err := ap.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := ap.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Schedule}, MetaNow, false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Targets + "[" + MetaAccounts + "]"}, "1001;1002", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Targets, MetaThresholds}, "TH1;TH2", false); err != nil {
		t.Error(err)
	}

	if err := ap.Set([]string{Actions + "[acc1]", ID}, "acc1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", "Wrong", "path", "2"}, "acc1", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", "Wrong", "path"}, "acc1", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", "Wrong"}, "acc1", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := ap.Set([]string{Actions, "acc1", Opts}, "opt0:val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Opts + "[opt1]"}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Opts, "opt2"}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Opts, "opt3", "opt4"}, "val1", false); err != nil {
		t.Error(err)
	}

	if err := ap.Set([]string{Actions, "acc1", Type}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", TTL}, "10", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, ID}, "dID", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, Path}, "path", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := ap.Set([]string{Actions, "acc1", Diktats, Value}, "val1", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, Opts}, "*balancePath:path", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, Opts, "*balanceValue"}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{Actions, "acc1", Diktats, Blockers}, ";true", false); err != nil {
		t.Error(err)
	}

	if err := ap.Actions[0].Set(nil, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, ap) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(ap))
	}
}

func TestActionProfileFieldAsInterface(t *testing.T) {
	ap := ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Schedule:  MetaNow,
		Weights: DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Targets: map[string]StringSet{
			MetaAccounts:   NewStringSet([]string{"1001", "1002"}),
			MetaThresholds: NewStringSet([]string{"TH1", "TH2"}),
		},
		Actions: []*APAction{{
			ID:        "acc1",
			Type:      "val1",
			FilterIDs: []string{"fltr1"},
			TTL:       10,
			Opts: map[string]any{
				"opt0": "val1",
				"opt1": "val1",
				"opt2": "val1",
				"opt3": MapStorage{"opt4": "val1"},
			},
			Diktats: []*APDiktat{{
				Opts: map[string]any{
					"*balancePath":  "path",
					"*balanceValue": "val1",
				},
			}},
		}},
	}
	if _, err := ap.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Schedule}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Schedule; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Targets}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Targets; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Targets + "[*accounts]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Targets[MetaAccounts]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	expErrMsg := `strconv.Atoi: parsing "a": invalid syntax`
	if _, err := ap.FieldAsInterface([]string{FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[a]", "a"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions, ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[4]", "a"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Targets + "[4]", "a"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{Targets + "[*accounts]", "1001"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Targets[MetaAccounts]["1001"]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", "a"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", ID}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].ID; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", TTL}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].TTL; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", Type}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Type; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", Opts}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Opts; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Opts + "[0]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Opts + "0"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Opts + "0", "0"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Opts + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Opts + "[0]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", "" + "[0]", "", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats, "0"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[4]", "0"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[0]", "0"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[a]", FilterIDs + "[0]", "Blocker"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[0]", Opts, "*balancePath"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats[0].Opts["*balancePath"]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Actions + "[0]", Diktats + "[0]", Opts, "*balanceValue"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats[0].Opts["*balanceValue"]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := ap.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsString([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := ap.String(), ToJSON(ap); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := ap.Actions[0].FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.Actions[0].FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "acc1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := ap.Actions[0].String(), ToJSON(ap.Actions[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := ap.Actions[0].Diktats[0].FieldAsString([]string{"", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.Actions[0].Diktats[0].FieldAsString([]string{Opts, "*balancePath"}); err != nil {
		t.Fatal(err)
	} else if exp := "path"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := ap.Actions[0].Diktats[0].String(), ToJSON(ap.Actions[0].Diktats[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestActionProfileMerge(t *testing.T) {
	acc := &ActionProfile{
		Targets: make(map[string]StringSet),
	}
	exp := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]StringSet{MetaAccounts: {"1001": {}}},
		Actions:  []*APAction{{}},
	}
	if acc.Merge(&ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]StringSet{MetaAccounts: {"1001": {}}},
		Actions:  []*APAction{{}},
	}); !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(acc))
	}
}

func TestActionProfileMergeAPActionMerge(t *testing.T) {
	acc := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
		},
		Schedule: "* * * * 1-5",
		Targets:  map[string]StringSet{MetaAccounts: {"1002": {}}},
		Actions: []*APAction{
			{
				ID:        "APAct1",
				FilterIDs: []string{"FLTR1", "FLTR2", "FLTR3"},
				TTL:       time.Minute,
				Type:      "type2",
				Opts: map[string]any{
					"key1": "value1",
					"key2": "value2",
				},
				Diktats: []*APDiktat{},
			},
		},
	}
	exp := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "fltr2"},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]StringSet{MetaAccounts: {"1001": {}}},
		Actions: []*APAction{
			{
				ID:        "APAct1",
				FilterIDs: []string{"FLTR1", "FLTR2", "FLTR3", "FLTR4"},
				TTL:       time.Minute,
				Type:      "type2",
				Opts: map[string]any{
					"key1": "value1",
					"key2": "value3",
				},
				Diktats: []*APDiktat{
					{
						Opts: map[string]any{
							"*balancePath":  "path2",
							"*balanceValue": "val2",
						},
					},
				},
			},
		},
	}
	if acc.Merge(&ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr2"},
		Weights: DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets: map[string]StringSet{
			MetaAccounts: {
				"1001": {},
			},
			"": {},
		},
		Actions: []*APAction{
			{
				ID:        "APAct1",
				FilterIDs: []string{"FLTR4"},
				Type:      "",
				Opts: map[string]any{
					"key2": "value3",
				},
				Diktats: []*APDiktat{
					{
						Opts: map[string]any{
							"*balancePath":  "path2",
							"*balanceValue": "val2",
						},
					},
				},
			},
		},
	}); !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(acc))
	}
}

func TestActionProfileAPActionMergeEmptyV2(t *testing.T) {
	apAct := &APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1"},
		TTL:       time.Second,
		Type:      "type",
		Opts: map[string]any{
			"key": "value",
		},
		Diktats: []*APDiktat{
			{
				Opts: map[string]any{
					"*balancePath":  "path",
					"*balanceValue": "value",
				},
			},
		},
	}
	expected := &APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1"},
		TTL:       time.Second,
		Type:      "type",
		Opts: map[string]any{
			"key": "value",
		},
		Diktats: []*APDiktat{
			{
				Opts: map[string]any{
					"*balancePath":  "path",
					"*balanceValue": "value",
				},
			},
		},
	}

	apAct.Merge(&APAction{})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(expected), ToJSON(apAct))
	}
}

func TestActionProfileAPActionMergeEmptyV1(t *testing.T) {
	apAct := &APAction{
		Opts: make(map[string]any),
	}
	expected := &APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1"},
		TTL:       time.Second,
		Type:      "type",
		Opts: map[string]any{
			"key": "value",
		},
		Diktats: []*APDiktat{
			{
				Opts: map[string]any{
					"*balancePath":  "path",
					"*balanceValue": "value",
				},
			},
		},
	}

	apAct.Merge(&APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1"},
		TTL:       time.Second,
		Type:      "type",
		Opts: map[string]any{
			"key": "value",
		},
		Diktats: []*APDiktat{
			{
				Opts: map[string]any{
					"*balancePath":  "path",
					"*balanceValue": "value",
				},
			},
		},
	})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(expected), ToJSON(apAct))
	}
}

func TestActionProfileAPActionMerge(t *testing.T) {
	apAct := &APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1"},
		TTL:       time.Second,
		Type:      "type1",
		Opts: map[string]any{
			"key1": "value1",
		},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
		},
		Blockers: DynamicBlockers{
			{
				FilterIDs: []string{"fltr2"},
				Blocker:   true,
			},
		},
		Diktats: []*APDiktat{
			{
				ID:        "DID1",
				FilterIDs: []string{"fltr1"},
				Opts: map[string]any{
					"*balancePath":  "",
					"*balanceValue": "",
				},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"fltr2"},
						Weight:    40,
					},
				},
				Blockers: DynamicBlockers{
					{
						FilterIDs: []string{"fltr2"},
						Blocker:   true,
					},
				},
			},
		},
	}
	expected := &APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1", "FLTR2", "FLTR3"},
		TTL:       time.Minute,
		Type:      "type2",
		Opts: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
			{
				Weight: 65,
			},
		},
		Blockers: DynamicBlockers{
			{
				FilterIDs: []string{"fltr2"},
				Blocker:   true,
			},
			{
				FilterIDs: []string{"fltr3"},
				Blocker:   true,
			},
		},
		Diktats: []*APDiktat{
			{
				ID:        "DID1",
				FilterIDs: []string{"fltr1", "fltr1"},
				Opts: map[string]any{
					"*balancePath":  "",
					"*balanceValue": "value1",
				},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"fltr2"},
						Weight:    40,
					},
					{
						Weight: 65,
					},
				},
				Blockers: DynamicBlockers{
					{
						FilterIDs: []string{"fltr2"},
						Blocker:   true,
					},
					{
						FilterIDs: []string{"fltr3"},
						Blocker:   true,
					},
				},
			},
		},
	}

	apAct.Merge(&APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR2", "FLTR3"},
		TTL:       time.Minute,
		Type:      "type2",
		Opts: map[string]any{
			"key2": "value2",
		},
		Weights: DynamicWeights{
			{
				Weight: 65,
			},
		},
		Blockers: DynamicBlockers{
			{
				FilterIDs: []string{"fltr3"},
				Blocker:   true,
			},
		},
		Diktats: []*APDiktat{
			{
				ID:        "DID1",
				FilterIDs: []string{"fltr1"},
				Opts: map[string]any{
					"*balancePath":  "",
					"*balanceValue": "value1",
				},
				Weights: DynamicWeights{
					{
						Weight: 65,
					},
				},
				Blockers: DynamicBlockers{
					{
						FilterIDs: []string{"fltr3"},
						Blocker:   true,
					},
				},
			},
		},
	})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(expected), ToJSON(apAct))
	}

	expected = &APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR1", "FLTR2", "FLTR3", "FLTR4"},
		TTL:       time.Minute,
		Type:      "type2",
		Opts: map[string]any{
			"key1": "value1",
			"key2": "value3",
		},
		Weights: DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
			{
				Weight: 65,
			},
		},
		Blockers: DynamicBlockers{
			{
				FilterIDs: []string{"fltr2"},
				Blocker:   true,
			},
			{
				FilterIDs: []string{"fltr3"},
				Blocker:   true,
			},
		},
		Diktats: []*APDiktat{
			{
				ID:        "DID1",
				FilterIDs: []string{"fltr1", "fltr1"},
				Opts: map[string]any{
					"*balancePath":  "",
					"*balanceValue": "value1",
				},
				Weights: DynamicWeights{
					{
						FilterIDs: []string{"fltr2"},
						Weight:    40,
					},
					{
						Weight: 65,
					},
				},
				Blockers: DynamicBlockers{
					{
						FilterIDs: []string{"fltr2"},
						Blocker:   true,
					},
					{
						FilterIDs: []string{"fltr3"},
						Blocker:   true,
					},
				},
			},
			{
				Opts: map[string]any{
					"*balancePath":  "path2",
					"*balanceValue": "value2",
				},
			},
		},
	}

	apAct.Merge(&APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR4"},
		Type:      "",
		Opts: map[string]any{
			"key2": "value3",
		},
		Diktats: []*APDiktat{
			{
				Opts: map[string]any{
					"*balancePath":  "path2",
					"*balanceValue": "value2",
				},
			},
		},
	})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			ToJSON(expected), ToJSON(apAct))
	}
}
