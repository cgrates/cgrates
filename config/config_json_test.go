/*
Real-time Charging System for Telecom & ISP environments
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
	"github.com/cgrates/cgrates/utils"
	"reflect"
	"testing"
)

var cgrJsonCfg CgrJsonCfg

func TestNewCgrJsonCfgFromFile(t *testing.T) {
	var err error
	if cgrJsonCfg, err = NewCgrJsonCfgFromFile("cgrates_sample_cfg.json"); err != nil {
		t.Error(err.Error())
	}
}

func TestGeneralJsonCfg(t *testing.T) {
	eGCfg := &GeneralJsonCfg{
		Http_skip_tls_veify: utils.BoolPointer(false),
		Rounding_decimals:   utils.IntPointer(10),
		Dbdata_encoding:     utils.StringPointer("msgpack"),
		Tpexport_dir:        utils.StringPointer("/var/log/cgrates/tpe"),
		Default_reqtype:     utils.StringPointer("rated"),
		Default_category:    utils.StringPointer("call"),
		Default_tenant:      utils.StringPointer("cgrates.org"),
		Default_subject:     utils.StringPointer("cgrates")}
	if gCfg, err := cgrJsonCfg.GeneralJsonCfg(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eGCfg, gCfg) {
		t.Error("Received: ", gCfg)
	}
}
