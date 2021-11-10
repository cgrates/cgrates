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
	r := TenantIDFromMap(utils.MapStorage{
		utils.Tenant: exp.Tenant,
		utils.ID:     exp.ID,
	})
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}

func TestRateIDsFromMap(t *testing.T) {
	expErrMsg := "cannot find RateIDs in map"
	if _, err := RateIDsFromMap(utils.MapStorage{}); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	exp := []string{"RT1", "RT2"}
	r, err := RateIDsFromMap(utils.MapStorage{
		utils.RateIDs: "RT1;RT2",
	})
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
	if _, err := newRecord(context.Background(), utils.MapStorage{},
		[]*config.FCTemplate{
			{Filters: []string{"*exists:~*req.NoField:"}},
			{Filters: []string{"*string"}},
		},
		"cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil)); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}
	r, err := newRecord(context.Background(), utils.MapStorage{}, []*config.FCTemplate{}, "cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil))
	if err != nil {
		t.Fatal(err)
	}
	exp := utils.MapStorage{}
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}

}

func TestNewRecordErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	expErrMsg := "unsupported type: <notSupported>"
	if _, err := newRecord(context.Background(), utils.MapStorage{},
		[]*config.FCTemplate{
			{Type: "notSupported"},
		},
		"cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil)); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}

	r := &record{
		data: utils.MapStorage{
			"Tenant": []string{},
		},
		req:   utils.MapStorage{},
		cfg:   cfg.GetDataProvider(),
		cache: ltcache.NewCache(-1, 0, false, nil),
	}
	if exp, rply := "{}", r.String(); exp != rply {
		t.Errorf("Expected %q, received %q", exp, rply)
	}
	fc := []*config.FCTemplate{
		{Type: utils.MetaComposed, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*cfg.general.node_id", ";")},
		{Type: utils.MetaComposed, Path: "Tenant.NewID", Value: config.NewRSRParsersMustCompile("10", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	if err := r.parseTemplates(context.Background(), fc,
		"cgrates.org", fs, 0, "", ";"); err != utils.ErrWrongPath {
		t.Errorf("Expeceted: %q, received: %v", utils.ErrWrongPath, err)
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
	}
	for _, f := range fc {
		f.ComputePath()
	}
	exp := utils.MapStorage{"ID": "Attr1", "Tenant": "cgrates.org", "Value": "0"}
	if r, err := newRecord(context.Background(), config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil),
		fc, "cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}
