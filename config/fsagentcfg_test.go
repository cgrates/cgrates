// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestFsConnCfgLoadFromJSONCfg(t *testing.T) {
	fs := &FsConnCfg{
		MaxReconnectInterval: time.Duration(4),
	}
	jsnCfg := &FsConnJsonCfg{
		MaxReconnectInterval: utils.StringPointer("invalid time"),
	}
	expErr := `time: invalid duration "invalid time"`
	if err := fs.loadFromJSONCfg(jsnCfg); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err.Error())

	}
}

func TestDiffFsConnJsonCfgMaxReconnInterval(t *testing.T) {
	v1 := &FsConnCfg{MaxReconnectInterval: time.Duration(3)}

	v2 := &FsConnCfg{MaxReconnectInterval: time.Duration(2)}

	expected := &FsConnJsonCfg{MaxReconnectInterval: utils.StringPointer("2ns")}

	rcv := diffFsConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &FsConnJsonCfg{}

	rcv = diffFsConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
