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
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	janus "github.com/cgrates/janusgo"
)

var (
	janCfgPath  string
	janCfgDIR   string
	janCfg      *config.CGRConfig
	janBin      string
	janRPC      *birpc.Client
	sTestsJanus = []func(t *testing.T){
		testJanitInitCfg,
		testJanitResetDB,
		testJanCmd,
		testJanitStartEngine,
		testJanitApierRpcConn,
		testJanitTPFromFolder,
		testJanitAgent,
		testJanitStopEngine,
	}
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
	janBin, err = exec.LookPath("/opt/janus/bin/janus")
	if err != nil || janBin == "" {
		t.SkipNow()
	}
	for _, stests := range sTestsJanus {
		t.Run(janCfgDIR, stests)
	}
}

func testJanitInitCfg(t *testing.T) {
	var err error
	janCfgPath = path.Join(*utils.DataDir, "conf", "samples", janCfgDIR)
	janCfg, err = config.NewCGRConfigFromPath(janCfgPath)
	if err != nil {
		t.Fatal(err)
	}
}

func testJanitResetDB(t *testing.T) {
	if err := engine.InitDataDb(janCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(janCfg); err != nil {
		t.Fatal(err)
	}
}

func testJanitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(janCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testJanitApierRpcConn(t *testing.T) {
	var err error
	janRPC, err = newRPCClient(janCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testJanCmd(t *testing.T) {
	cmd := exec.Command(janBin)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
}

func testJanitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := janRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testJanitAgent(t *testing.T) {

	var sessionMsg janus.SuccessMsg

	id := utils.GenUUID()

	msg := janus.BaseMsg{
		Type: "create",
		ID:   id,
	}

	sessionurl := fmt.Sprintf("http://%s%s", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL)
	body, err := sendReq(true, sessionurl, msg)
	if err != nil {
		t.Error(err)
	}

	exp := janus.SuccessMsg{
		Type: "success",
		ID:   id,
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

	url := fmt.Sprintf("http://%s%s/%d", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID)
	body, err = sendReq(true, url, msg)
	if err != nil {
		t.Error(err)
	}

	exp = janus.SuccessMsg{
		Type: "success",
		ID:   id,
	}

	if err := json.Unmarshal(body, &pluginMsg); err != nil {
		t.Error(err)
	}
	if exp.ID != id || exp.Type != "success" {
		t.Error(err)
	}

	pullResp := make(chan []byte)
	done := make(chan struct{})
	go func() {
		url := fmt.Sprintf("http://localhost%s%s/%d?rid=1714324194137&maxev=10", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID)
		ticker := time.NewTicker(50 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				body, err := sendReq(false, url, nil)
				if err != nil {
					continue
				}
				pullResp <- body
			case <-done:
				return
			}
		}

	}()

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
	url = fmt.Sprintf("http://%s%s/%d/%d", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID, pluginMsg.Data.ID)
	body, err = sendReq(true, url, messagePlugin)
	if err != nil {
		t.Error(err)
	}

	expAck := janus.AckMsg{
		Type: "ack",
		ID:   id,
		Hint: "I'm taking my time!",
	}

	var ackMsg janus.AckMsg
	if err := json.Unmarshal(body, &ackMsg); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(ackMsg, expAck) {
		t.Errorf("Expected <%s> ,got <%s> instead", utils.ToJSON(expAck), utils.ToJSON(ackMsg))
	}
	var eventMsg []janus.EventMsg
	eventmsg := <-pullResp

	if err = json.Unmarshal(eventmsg, &eventMsg); err != nil {
		t.Error(err)
	}

	expEvent := []janus.EventMsg{
		{
			ID:      id,
			Session: sessionMsg.Data.ID,
			Handle:  pluginMsg.Data.ID,
			Type:    "event",
			Plugindata: janus.PluginData{
				Plugin: "janus.plugin.echotest",
				Data: map[string]interface{}{
					"echotest": "event",
					"result":   "ok",
				},
			},
		},
	}

	if !reflect.DeepEqual(eventMsg, expEvent) {
		t.Errorf("expected <%+v>,got <%+v instead", expEvent, eventMsg)
	}
	close(done)
}

func sendReq(isPost bool, url string, msg any) (body []byte, err error) {
	var (
		req  *http.Request
		data []byte
	)
	if isPost {
		data, err = json.Marshal(msg)
		if err != nil {
			return
		}
		req, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))

	} else {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	}
	if err != nil {
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}

	return
}

func testJanitStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
	if err := exec.Command("pkill", "janus").Run(); err != nil {
		t.Error(err)
	}
}
