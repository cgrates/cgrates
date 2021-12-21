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

package loaders

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestTenantIDFromMap(t *testing.T) {
	exp := utils.NewTenantID("cgrates.org:ATTR1")
	r := TenantIDFromOrderedNavigableMap(
		newOrderNavMap(utils.MapStorage{
			utils.Tenant: exp.Tenant,
			utils.ID:     exp.ID,
		}))
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}

func TestRateIDsFromMap(t *testing.T) {
	expErrMsg := "cannot find RateIDs in map"
	if _, err := RateIDsFromOrderedNavigableMap(newOrderNavMap(utils.MapStorage{})); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	exp := []string{"RT1", "RT2"}
	r, err := RateIDsFromOrderedNavigableMap(newOrderNavMap(utils.MapStorage{
		utils.RateIDs: "RT1;RT2",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}

func TestNewRecord(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	expErrMsg := "inline parse error for string: <*string>"
	if err := newRecord(utils.MapStorage{}, nil, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), []*config.FCTemplate{
			{Filters: []string{"*exists:~*req.NoField:"}},
			{Filters: []string{"*string"}},
		}, fs, 0, "", utils.InfieldSep); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}
	pt := profileTest{}
	if err := newRecord(utils.MapStorage{}, pt, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), []*config.FCTemplate{}, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	}
	if len(pt) != 0 {
		t.Fatal("Expected empty map")
	}
	exp := `{}`
	if rply := newRecord(utils.MapStorage{}, pt, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).String(); exp != rply {
		t.Errorf("Expeceted: %q, received: %v", exp, rply)
	}
}

func TestNewRecordWithCahe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*req.0", ";")},
		{Type: utils.MetaVariable, Path: "ID", Value: config.NewRSRParsersMustCompile("~*req.1", ";")},
		{Type: utils.MetaComposed, Path: "*uch.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaVariable, Path: "Value", Value: config.NewRSRParsersMustCompile("~*uch.*tntID.Value", ";")},
		{Type: utils.MetaRemove, Path: "Value", Value: config.NewRSRParsersMustCompile("~*uch.*tntID.Value", ";")}, //ignored
		{Type: utils.MetaComposed, Path: "*uch.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaRemove, Path: "*uch.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaRemoveAll, Path: "*uch", Value: config.NewRSRParsersMustCompile("0", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	exp := profileTest{
		utils.Tenant: "cgrates.org",
		utils.ID:     "Attr1",
		utils.Value:  "0",
	}
	r := profileTest{}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+v", exp, r)
	}
}

func TestNewRecordWithTmp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*req.0", ";")},
		{Type: utils.MetaVariable, Path: "ID", Value: config.NewRSRParsersMustCompile("~*req.1", ";")},
		{Type: utils.MetaComposed, Path: "*tmp.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaVariable, Path: "Value", Value: config.NewRSRParsersMustCompile("~*tmp.*tntID.Value", ";")},
		{Type: utils.MetaComposed, Path: "*tmp.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaRemove, Path: "*tmp.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaRemoveAll, Path: "*tmp", Value: config.NewRSRParsersMustCompile("0", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	exp := profileTest{
		utils.Tenant: "cgrates.org",
		utils.ID:     "Attr1",
		utils.Value:  "0",
	}
	r := profileTest{}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+v", exp, r)
	}
}

type profileTest utils.MapStorage

func (p profileTest) Set(path []string, val interface{}, _ bool, _ string) error {
	return utils.MapStorage(p).Set(path, val)
}
func (p profileTest) Merge(v2 interface{}) {
	var vi map[string]interface{}
	json.Unmarshal([]byte(utils.ToJSON(v2)), &vi)
	for k, v := range vi {
		(map[string]interface{}(p))[k] = v
	}
}
func (p profileTest) TenantID() string {
	return utils.ConcatenatedKey(utils.IfaceAsString(map[string]interface{}(p)[utils.Tenant]), utils.IfaceAsString(map[string]interface{}(p)[utils.ID]))
}

func (p profileTest) String() string { return utils.MapStorage(p).String() }
func (p profileTest) FieldAsString(fldPath []string) (string, error) {
	return utils.MapStorage(p).FieldAsString(fldPath)
}
func (p profileTest) FieldAsInterface(fldPath []string) (interface{}, error) {
	return utils.MapStorage(p).FieldAsInterface(fldPath)
}

func TestNewProfileFunc(t *testing.T) {
	tf := newProfileFunc(utils.EmptyString)
	if v := tf(); v != nil {
		t.Fatal("Expected emoty reply")
	}
}

func TestNewRecordWithTmp2(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*req.0", ";")},
		{Type: utils.MetaVariable, Path: "ID", Value: config.NewRSRParsersMustCompile("~*req.1", ";")},
		{Type: utils.MetaVariable, Path: "*tmp.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaVariable, Path: "*uch.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaComposed, Path: "Value", Value: config.NewRSRParsersMustCompile("~*tmp.*tntID.Value", ";"), Blocker: true},
		{Type: utils.MetaComposed, Path: "Value", Value: config.NewRSRParsersMustCompile("~*tmp.*tntID.Value", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	exp := profileTest{
		utils.Tenant: "cgrates.org",
		utils.ID:     "Attr1",
		utils.Value:  "0",
	}
	r := profileTest{}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+v", exp, r)
	}
}

func TestNewRecordWithComposeError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*req.0", ";")},
		{Type: utils.MetaVariable, Path: "ID", Value: config.NewRSRParsersMustCompile("~*req.1", ";")},
		{Type: utils.MetaVariable, Path: "*tmp.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaComposed, Path: "Value.NotVal", Value: config.NewRSRParsersMustCompile("~*tmp.*tntID.Value", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	r := profileTest{
		"Value": []string{},
	}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != utils.ErrWrongPath {
		t.Error(err)
	}
}

func TestNewRecordWithRemoveError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "*tmp.Value.0.Field", Value: config.NewRSRParsersMustCompile("tmp", ";")},
		{Type: utils.MetaRemove, Path: "*tmp.Value.NotVal", Value: config.NewRSRParsersMustCompile("tmp", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	r := profileTest{
		"Value": []string{},
	}
	expErrMsg := `strconv.Atoi: parsing "NotVal": invalid syntax`
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func TestNewRecordSetFieldsError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "*tmp.Value<0.Field", Value: config.NewRSRParsersMustCompile("~*cfg.tmp", ";")},
		{Type: utils.MetaVariable, Path: "*tmp.Value<0.Field", Value: config.NewRSRParsersMustCompile("tmp", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	r := profileTest{
		"Value": []string{},
	}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != utils.ErrWrongPath {
		t.Error(err)
	}
}

func TestNewRecordSetFieldsMandatoryError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "path", Value: config.NewRSRParsersMustCompile("~*cfg.tmp", ";"), Mandatory: true},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	r := profileTest{
		"Value": []string{},
	}
	expErrMsg := `NOT_FOUND:`
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
}

func TestRecordFieldAsInterface(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dp := config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil)
	r := newRecord(dp, profileTest{}, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil))
	if val, err := r.FieldAsInterface([]string{utils.MetaReq}); err != nil {
		t.Fatal(err)
	} else if exp := dp; !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %+v, received %+v", exp, val)
	}
	if val, err := r.FieldAsInterface([]string{utils.MetaTmp}); err != nil {
		t.Fatal(err)
	} else if exp := (&utils.DataNode{Type: utils.NMMapType, Map: make(map[string]*utils.DataNode)}); !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %+v, received %+v", exp, val)
	}
	if val, err := r.FieldAsInterface([]string{utils.MetaCfg}); err != nil {
		t.Fatal(err)
	} else if exp := cfg.GetDataProvider(); !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %+v, received %+v", exp, val)
	}
	if val, err := r.FieldAsInterface([]string{utils.MetaTenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; !reflect.DeepEqual(val, exp) {
		t.Errorf("Expected %+v, received %+v", exp, val)
	}
	if _, err := r.FieldAsInterface([]string{utils.MetaUCH}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
}
