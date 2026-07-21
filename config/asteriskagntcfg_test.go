// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestAsteriskConnCfgloadFromJSONCfg(t *testing.T) {
	aConnCfg := &AsteriskConnCfg{
		MaxReconnectInterval: time.Duration(5),
	}

	jsnCfg := &AstConnJsonCfg{

		Max_reconnect_interval: utils.StringPointer("return error"),
	}
	expErr := "time: invalid duration \"return error\""
	if err := aConnCfg.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err)
	}

}

func TestDiffAstConnJsonCfgMaxReconn(t *testing.T) {
	v1 := &AsteriskConnCfg{
		MaxReconnectInterval: time.Duration(4),
		AriWebSocket:         true,
	}

	v2 := &AsteriskConnCfg{
		MaxReconnectInterval: time.Duration(5),
		AriWebSocket:         false,
	}

	expected := &AstConnJsonCfg{
		Max_reconnect_interval: utils.StringPointer("5ns"),
		Ari_websocket:          utils.BoolPointer(false),
	}

	rcv := diffAstConnJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}
