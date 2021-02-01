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

package console

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestCmdGetCommands(t *testing.T) {
	expected := commands
	result := GetCommands()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpected <%+v>,\nReceived <%+v>", expected, result)
	}
}

func TestGetAvailableCommandsErr(t *testing.T) {
	err := getAvailabelCommandsErr()
	if err == nil {
		t.Errorf("\nExpected not nil,\nReceived <%+v>", err)
	}
}

func TestGetCommandValueCase1(t *testing.T) {
	expected := &CmdGetChargersForEvent{
		name:      "chargers_for_event",
		rpcMethod: utils.ChargerSv1GetChargersForEvent,
		rpcParams: &utils.CGREvent{},
	}
	result, err := GetCommandValue("chargers_for_event", false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(utils.ToJSON(result), utils.ToJSON(expected)) {
		t.Errorf("\nExpected <%+v>,\nReceived <%+v>", utils.ToJSON(result), utils.ToJSON(expected))
	}
}

func TestGetCommandValueCase2(t *testing.T) {
	_, err := GetCommandValue("", false)
	if err == nil {
		t.Fatal(err)
	}
}

func TestGetCommandValueCase3(t *testing.T) {
	_, err := GetCommandValue("false_command", false)
	if err == nil {
		t.Fatal(err)
	}
}

func TestGetCommandValueCase4(t *testing.T) {
	_, err := GetCommandValue("chargers _for_event", false)
	if err == nil {
		t.Fatal(err)
	}
}

type mockCommander struct {
	Commander
}

func (*mockCommander) FromArgs(args string, verbose bool) error {
	return nil
}

func (*mockCommander) Usage() string {
	return utils.EmptyString
}

func (*mockCommander) RpcMethod() string {
	return utils.EmptyString
}
func (*mockCommander) RpcParams(bool) interface{} {
	return nil
}

func (*mockCommander) PostprocessRpcParams() error {
	return utils.ErrNotImplemented
}

func (*mockCommander) RpcResult() interface{} {
	return nil
}
func (*mockCommander) ClientArgs() []string {
	return []string{}
}
func (*mockCommander) Name() string {
	return utils.EmptyString
}
func (*mockCommander) LocalExecute() string {
	return utils.EmptyString
}
func (*mockCommander) GetFormatedResult(result interface{}) string {
	return utils.EmptyString
}

func TestGetCommandValueCase5(t *testing.T) {
	commands["chargers_for_event"] = new(mockCommander)
	_, err := GetCommandValue("chargers_for_event", true)
	if !reflect.DeepEqual(utils.ErrNotImplemented, err) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.ErrNotImplemented, err)
	}
}

type mockCommander2 struct {
	Commander
}

func (*mockCommander2) FromArgs(args string, verbose bool) error {
	return utils.ErrNotImplemented
}

func (*mockCommander2) Usage() string {
	return utils.EmptyString
}

func (*mockCommander2) RpcMethod() string {
	return utils.EmptyString
}
func (*mockCommander2) RpcParams(bool) interface{} {
	return nil
}

func (*mockCommander2) PostprocessRpcParams() error {
	return nil
}

func (*mockCommander2) RpcResult() interface{} {
	return nil
}

func (*mockCommander2) ClientArgs() []string {
	return []string{}
}

func (*mockCommander2) Name() string {
	return utils.EmptyString
}

func (*mockCommander2) LocalExecute() string {
	return utils.EmptyString
}

func (*mockCommander2) GetFormatedResult(result interface{}) string {
	return utils.EmptyString
}

func TestGetCommandValueCase6(t *testing.T) {
	commands["test"] = new(mockCommander2)
	_, err := GetCommandValue("test", true)
	if !reflect.DeepEqual(utils.ErrNotImplemented, err) {
		t.Errorf("Expected <%+v>, Received <%+v>", utils.ErrNotImplemented, err)
	}
}
