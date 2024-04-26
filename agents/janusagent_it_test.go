//go:build integration
// +build integration

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
package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	janus "github.com/cgrates/janusgo"
)

var (
	janCfgPath string
	janCfgDIR  string
	janCfg     *config.CGRConfig
	janhttpC   *http.Client // so we can cache the connection

)

func TestJanusit(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		janCfgDIR = "janus_agent"
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	file, err := exec.LookPath("/opt/janus/bin/janus")
	if err != nil || file == "" {
		t.SkipNow()
	}
	janCfgPath = path.Join(*utils.DataDir, "conf", "samples", janCfgDIR)
	janCfg, err = config.NewCGRConfigFromPath(janCfgPath)
	if err != nil {
		t.Fatal("could not start init ", err.Error())
	}

	cmd := exec.Command(file)

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Wait()

	janhttpC = http.DefaultClient

	if err := engine.InitDataDb(janCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(janCfg); err != nil {
		t.Fatal(err)
	}

	if _, err := engine.StopStartEngine(janCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	var sessionMsg janus.SuccessMsg

	id := utils.GenUUID()

	msg := janus.BaseMsg{
		Type: "create",
		ID:   id,
	}
	data, err := json.Marshal(&msg)
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf("http://%s%s", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
	}
	res, err := janhttpC.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	exp := janus.SuccessMsg{
		Type: "success",
		ID:   id,
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(body, &sessionMsg); err != nil {
		t.Error(err)
	}
	if exp.ID != id || exp.Type != "success" {
		t.Error(err)
	}

	var pluginMsg janus.SuccessMsg

	id = utils.GenUUID()

	msg = janus.BaseMsg{
		Type:   "attach",
		Plugin: "janus.plugin.echotest",
		ID:     id,
	}

	url = fmt.Sprintf("http://%s%s/%d", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID)
	data, err = json.Marshal(&msg)
	if err != nil {
		t.Error(err)
	}
	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
	}
	res, err = janhttpC.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	exp = janus.SuccessMsg{
		Type: "success",
		ID:   id,
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(body, &pluginMsg); err != nil {
		t.Error(err)
	}
	if exp.ID != id || exp.Type != "success" {
		t.Error(err)
	}

	id = utils.GenUUID()
	messagePlugin := janus.HandlerMessage{
		BaseMsg: janus.BaseMsg{
			Type: "message",
			ID:   id,
		},
		Body: map[string]any{
			"audio": true,
			"video": true,
		},
	}
	data, err = json.Marshal(&messagePlugin)
	if err != nil {
		t.Error(err)
	}

	url = fmt.Sprintf("http://%s%s/%d/%d", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID, pluginMsg.Data.ID)
	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
	}
	res, err = janhttpC.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	expAck := janus.AckMsg{
		Type: "ack",
		ID:   id,
		Hint: "I'm taking my time!",
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	var ackMsg janus.AckMsg
	if err := json.Unmarshal(body, &ackMsg); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(ackMsg, expAck) {
		t.Errorf("Expected <%s> ,got <%s> instead", utils.ToJSON(expAck), utils.ToJSON(ackMsg))
	}

	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
