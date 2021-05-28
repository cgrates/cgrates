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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestConfigV1SetConfigWithDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	cfg.db = db

	v2 := NewDefaultCGRConfig()
	v2.GeneralCfg().NodeID = "Test"
	v2.GeneralCfg().DefaultCaching = utils.MetaClear
	var reply string
	if err := cfg.V1SetConfig(context.Background(), &SetConfigArgs{
		Config: v2.AsMapInterface(utils.InfieldSep),
	}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestConfigV1StoreCfgInDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	cfg.db = db

	cfg.GeneralCfg().NodeID = "Test"
	cfg.GeneralCfg().DefaultCaching = utils.MetaClear
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), &SectionWithAPIOpts{Sections: []string{utils.MetaAll}}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestConfigV1SetConfigFromJSONWithDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	cfg.db = db

	v2 := NewDefaultCGRConfig()
	v2.GeneralCfg().NodeID = "Test"
	v2.GeneralCfg().DefaultCaching = utils.MetaClear
	var reply string
	if err := cfg.V1SetConfigFromJSON(context.Background(), &SetConfigFromJSONArgs{
		Config: utils.ToJSON(v2.AsMapInterface(utils.InfieldSep)),
	}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}

func TestConfigLoadFromDB(t *testing.T) {
	cfg := NewDefaultCGRConfig()
	for _, section := range sortedCfgSections {
		cfg.rldChans[section] = make(chan struct{}, 10)
	}
	db := make(CgrJsonCfg)
	g1 := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}

	if err := db.SetSection(context.Background(), GeneralJSON, g1); err != nil {
		t.Fatal(err)
	}

	if err := cfg.LoadFromDB(db); err != nil {
		t.Fatal(err)
	}
	expGeneral := &GeneralCfg{
		NodeID:           "Test",
		DefaultCaching:   utils.MetaClear,
		Logger:           utils.MetaSysLog,
		LogLevel:         6,
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
		PosterAttempts:   3,
		FailedPostsDir:   "/var/spool/cgrates/failed_posts",
		DefaultReqType:   utils.MetaRated,
		DefaultCategory:  utils.Call,
		DefaultTenant:    "cgrates.org",
		DefaultTimezone:  "Local",
		ConnectAttempts:  5,
		Reconnects:       -1,
		ConnectTimeout:   time.Second,
		ReplyTimeout:     2 * time.Second,
		DigestSeparator:  ",",
		DigestEqual:      ":",
		MaxParallelConns: 100,
		RSRSep:           ";",
		FailedPostsTTL:   5 * time.Second,
	}
	if !reflect.DeepEqual(expGeneral, cfg.GeneralCfg()) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expGeneral), utils.ToJSON(cfg.GeneralCfg()))
	}

	cfg.GeneralCfg().NodeID = "Test2"
	var reply string
	if err := cfg.V1StoreCfgInDB(context.Background(), &SectionWithAPIOpts{Sections: []string{utils.MetaAll}}, &reply); err != nil {
		t.Fatal(err)
	}

	exp := &GeneralJsonCfg{
		Node_id:         utils.StringPointer("Test2"),
		Default_caching: utils.StringPointer(utils.MetaClear),
	}
	if rpl, err := db.GeneralJsonCfg(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp), utils.ToJSON(rpl))
	}
	exp2 := new(AccountSJsonCfg)
	if rpl, err := db.AccountSCfgJson(); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp2, rpl) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(exp2), utils.ToJSON(rpl))
	}
}
