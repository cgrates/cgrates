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

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestActionProfileSort(t *testing.T) {
	testStruct := &ActionProfiles{
		{
			Tenant: "test_tenantA",
			ID:     "test_idA",
			Weights: utils.DynamicWeights{
				{
					Weight: 1,
				},
			},
			weight: 1,
		},
		{
			Tenant: "test_tenantB",
			ID:     "test_idB",
			Weights: utils.DynamicWeights{
				{
					Weight: 2,
				},
			},
			weight: 2,
		},
		{
			Tenant: "test_tenantC",
			ID:     "test_idC",
			Weights: utils.DynamicWeights{
				{
					Weight: 3,
				},
			},
			weight: 3,
		},
	}
	expStruct := &ActionProfiles{
		{
			Tenant: "test_tenantC",
			ID:     "test_idC",
			Weights: utils.DynamicWeights{
				{
					Weight: 3,
				},
			},
			weight: 3,
		},

		{
			Tenant: "test_tenantB",
			ID:     "test_idB",
			Weights: utils.DynamicWeights{
				{
					Weight: 2,
				},
			},
			weight: 2,
		},
		{
			Tenant: "test_tenantA",
			ID:     "test_idA",
			Weights: utils.DynamicWeights{
				{
					Weight: 1,
				},
			},
			weight: 1,
		},
	}
	testStruct.Sort()
	if !reflect.DeepEqual(expStruct, testStruct) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expStruct), utils.ToJSON(testStruct))
	}
}

func TestActionAPDiktatRSRValues(t *testing.T) {
	apdDiktat := APDiktat{
		valRSR: config.RSRParsers{
			&config.RSRParser{
				Rules: ">;q=0.7;expires=3600",
			},
			&config.RSRParser{
				Rules: ">;q=0.7;expires=3600",
			},
		},
	}
	rsrPars, err := apdDiktat.RSRValues()
	if err != nil {
		t.Error(err)
	}
	expected := config.RSRParsers{
		&config.RSRParser{
			Rules: ">;q=0.7;expires=3600",
		},
		&config.RSRParser{
			Rules: ">;q=0.7;expires=3600",
		},
	}
	if !reflect.DeepEqual(rsrPars, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rsrPars))

	}
}

func TestActionAPDiktatRSRValuesNil(t *testing.T) {
	apdDiktat := APDiktat{}
	rsrPars, err := apdDiktat.RSRValues()
	if err != nil {
		t.Error(err)
	}
	var expected config.RSRParsers
	if !reflect.DeepEqual(rsrPars, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rsrPars))

	}
}

func TestActionAPDiktatRSRValuesError(t *testing.T) {
	apdDiktat := APDiktat{
		Value: "val`val2val3",
	}
	expErr := "Closed unspilit syntax"
	_, err := apdDiktat.RSRValues()
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}
func TestAPDiktatRSRValues(t *testing.T) {
	dk := &APDiktat{Value: "1001"}
	if rply, err := dk.RSRValues(); err != nil {
		return
	} else if exp := config.NewRSRParsersMustCompile("1001", utils.InfieldSep); !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %+v ,received: %+v", exp, rply)
	}
}

func TestActionProfileSet(t *testing.T) {
	ap := ActionProfile{Targets: make(map[string]utils.StringSet)}
	exp := ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Schedule:  utils.MetaNow,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts:   utils.NewStringSet([]string{"1001", "1002"}),
			utils.MetaThresholds: utils.NewStringSet([]string{"TH1", "TH2"}),
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
				"opt3": utils.MapStorage{"opt4": "val1"},
			},
			Diktats: []*APDiktat{{
				Path:  "path",
				Value: "val1",
			}},
		}},
	}
	if err := ap.Set([]string{}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{"NotAField"}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{"NotAField", "1"}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := ap.Set([]string{utils.Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Schedule}, utils.MetaNow, false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Targets + "[" + utils.MetaAccounts + "]"}, "1001;1002", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Targets, utils.MetaThresholds}, "TH1;TH2", false); err != nil {
		t.Error(err)
	}

	if err := ap.Set([]string{utils.Actions + "[acc1]", utils.ID}, "acc1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", "Wrong", "path", "2"}, "acc1", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", "Wrong", "path"}, "acc1", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", "Wrong"}, "acc1", false); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := ap.Set([]string{utils.Actions, "acc1", utils.Opts}, "opt0:val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", utils.Opts + "[opt1]"}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", utils.Opts, "opt2"}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", utils.Opts, "opt3", "opt4"}, "val1", false); err != nil {
		t.Error(err)
	}

	if err := ap.Set([]string{utils.Actions, "acc1", utils.Type}, "val1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", utils.FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", utils.TTL}, "10", false); err != nil {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Actions, "acc1", utils.Diktats, utils.Path}, "path", false); err != nil {
		t.Error(err)
	}

	if err := ap.Set([]string{utils.Actions, "acc1", utils.Diktats, utils.Value}, "val1", false); err != nil {
		t.Error(err)
	}

	if err := ap.Actions[0].Set(nil, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := ap.Set([]string{utils.Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, ap) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(ap))
	}
}

func TestActionProfileFieldAsInterface(t *testing.T) {
	ap := ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Schedule:  utils.MetaNow,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts:   utils.NewStringSet([]string{"1001", "1002"}),
			utils.MetaThresholds: utils.NewStringSet([]string{"TH1", "TH2"}),
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
				"opt3": utils.MapStorage{"opt4": "val1"},
			},
			Diktats: []*APDiktat{{
				Path:  "path",
				Value: "val1",
			}},
		}},
	}
	if _, err := ap.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Schedule}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Schedule; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Targets}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Targets; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Targets + "[*accounts]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Targets[utils.MetaAccounts]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	expErrMsg := `strconv.Atoi: parsing "a": invalid syntax`
	if _, err := ap.FieldAsInterface([]string{utils.FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[a]", "a"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions, ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[4]", "a"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Targets + "[4]", "a"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{utils.Targets + "[*accounts]", "1001"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Targets[utils.MetaAccounts]["1001"]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", "a"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].ID; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.TTL}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].TTL; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Type}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Type; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Opts}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Opts; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Opts + "[0]"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.FilterIDs + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[a]"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Opts + "0"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Opts + "0", "0"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Opts + "[0]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Opts + "[0]", "", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", "" + "[0]", "", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats, "0"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[4]", "0"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[0]", "0"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[a]", "0"}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}

	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[0]", utils.Path}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats[0].Path; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Actions + "[0]", utils.Diktats + "[0]", utils.Value}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Actions[0].Diktats[0].Value; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := ap.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsString([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := ap.String(), utils.ToJSON(ap); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := ap.Actions[0].FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.Actions[0].FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "acc1"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := ap.Actions[0].String(), utils.ToJSON(ap.Actions[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := ap.Actions[0].Diktats[0].FieldAsString([]string{"", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.Actions[0].Diktats[0].FieldAsString([]string{utils.Path}); err != nil {
		t.Fatal(err)
	} else if exp := "path"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := ap.Actions[0].Diktats[0].String(), utils.ToJSON(ap.Actions[0].Diktats[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestActionProfileMerge(t *testing.T) {
	acc := &ActionProfile{
		Targets: make(map[string]utils.StringSet),
	}
	exp := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*APAction{{}},
	}
	if acc.Merge(&ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
		Actions:  []*APAction{{}},
	}); !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(acc))
	}
}

func TestActionProfileMergeAPActionMerge(t *testing.T) {
	acc := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
		},
		Schedule: "* * * * 1-5",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1002": {}}},
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
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"fltr2"},
				Weight:    40,
			},
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
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
						Path:  "path2",
						Value: "value2",
					},
				},
			},
		},
	}
	if acc.Merge(&ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr2"},
		Weights: utils.DynamicWeights{
			{
				Weight: 65,
			},
		},
		Schedule: "* * * * *",
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: {
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
						Path:  "path2",
						Value: "value2",
					},
				},
			},
		},
	}); !reflect.DeepEqual(exp, acc) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(acc))
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
				Path:  "path",
				Value: "value",
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
				Path:  "path",
				Value: "value",
			},
		},
	}

	apAct.Merge(&APAction{})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expected), utils.ToJSON(apAct))
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
				Path:  "path",
				Value: "value",
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
				Path:  "path",
				Value: "value",
			},
		},
	})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expected), utils.ToJSON(apAct))
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
		Diktats: []*APDiktat{
			{
				Path:  "",
				Value: "",
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
		Diktats: []*APDiktat{},
	}

	apAct.Merge(&APAction{
		ID:        "APAct1",
		FilterIDs: []string{"FLTR2", "FLTR3"},
		TTL:       time.Minute,
		Type:      "type2",
		Opts: map[string]any{
			"key2": "value2",
		},
		Diktats: []*APDiktat{
			{
				Path:  "",
				Value: "value1",
			},
		},
	})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expected), utils.ToJSON(apAct))
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
		Diktats: []*APDiktat{
			{
				Path:  "path2",
				Value: "value2",
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
				Path:  "path2",
				Value: "value2",
			},
		},
	})
	if !reflect.DeepEqual(apAct, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expected), utils.ToJSON(apAct))
	}
}
func TestFilterHelpersGetWeightFromDynamics(t *testing.T) {
	var expected float64 = 64
	ap := &ActionProfile{
		Weights: []*utils.DynamicWeight{
			{
				Weight: 64,
			},
		},
		weight: 64,
	}
	ctx := context.Background()
	fltrs := &FilterS{}
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}
	err := ap.GetWeightFromDynamics(ctx, fltrs, tnt, ev)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(ap.weight, expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, ap.weight)
	}
}

func TestActionProfileGetWeightFromDynamicsErr(t *testing.T) {

	ap := &ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_ID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*stirng"},
				Weight:    20,
			},
		},
		Schedule: utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: utils.NewStringSet([]string{"acc1", "acc2", "acc3"}),
		},
		Actions: []*APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*APDiktat{{
					Path: "~*balance.TestBalance.Value",
				}},
			},
			{
				ID:        "TOPUP_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*APDiktat{{
					Path: "~*balance.TestVoiceBalance.Value",
				}},
			},
		},
	}
	ctx := context.Background()
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, config.CgrConfig().CacheCfg(), nil)

	cM := NewConnManager(cfg)
	fltrs := NewFilterS(cfg, cM, dm)
	tnt := utils.CGRateSorg
	ev := utils.MapStorage{}
	expErr := "inline parse error for string: <*stirng>"
	err := ap.GetWeightFromDynamics(ctx, fltrs, tnt, ev)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v> , Received <%v>", expErr, err)
	}
}
