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
	url = fmt.Sprintf("http://%s%s/%d?rid=1714324194137&maxev=10", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID)
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	pullResp := make(chan *http.Response)
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				rsp, err := janhttpC.Do(req)
				if err != nil {
					t.Error(err)
				}
				pullResp <- rsp
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
	data, err = json.Marshal(&messagePlugin)
	if err != nil {
		t.Error(err)
	}

	url = fmt.Sprintf("http://%s%s/%d/%d", janCfg.ListenCfg().HTTPListen, janCfg.JanusAgentCfg().URL, sessionMsg.Data.ID, pluginMsg.Data.ID)
	req2, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		t.Error(err)
	}
	res, err = janhttpC.Do(req2)
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
	var eventMsg []janus.EventMsg
	eventmsg := <-pullResp

	body, err = io.ReadAll(eventmsg.Body)
	if err != nil {
		t.Error(err)
	}

	if err = json.Unmarshal(body, &eventMsg); err != nil {
		t.Error(err)
	}
	eventmsg.Body.Close()
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

	id = utils.GenUUID()
	jsepMessage := janus.HandlerMessageJsep{
		HandlerMessage: janus.HandlerMessage{
			BaseMsg: janus.BaseMsg{
				Type: "message",
				ID:   id,
			},
			Body: map[string]any{
				"audio": true,
				"video": true,
			},
		},
		Jsep: map[string]any{
			"type": "offer",
			"sdp":  "v=0\r\no=- 6988110098816037393 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0 1 2\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS b98fa06f-90d2-4b82-8b51-ade1eefb38fa\r\nm=audio 9 UDP/TLS/RTP/SAVPF 111 63 9 0 8 13 110 126\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:uB6B\r\na=ice-pwd:YmBQuaIhSa85WKcu70uyDuME\r\na=ice-options:trickle\r\na=fingerprint:sha-256 C5:94:C2:D3:FF:DF:BA:B2:67:A3:E9:B6:3A:BE:67:98:EE:61:28:00:00:E3:73:2F:A7:4C:8F:F0:99:5E:D3:77\r\na=setup:actpass\r\na=mid:0\r\na=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=sendrecv\r\na=msid:b98fa06f-90d2-4b82-8b51-ade1eefb38fa f420b74c-3101-4882-805f-da7276f7bcf7\r\na=rtcp-mux\r\na=rtpmap:111 opus/48000/2\r\na=rtcp-fb:111 transport-cc\r\na=fmtp:111 minptime=10;useinbandfec=1\r\na=rtpmap:63 red/48000/2\r\na=fmtp:63 111/111\r\na=rtpmap:9 G722/8000\r\na=rtpmap:0 PCMU/8000\r\na=rtpmap:8 PCMA/8000\r\na=rtpmap:13 CN/8000\r\na=rtpmap:110 telephone-event/48000\r\na=rtpmap:126 telephone-event/8000\r\na=ssrc:58232649 cname:RdHAkKyMp2NLvuvZ\r\na=ssrc:58232649 msid:b98fa06f-90d2-4b82-8b51-ade1eefb38fa f420b74c-3101-4882-805f-da7276f7bcf7\r\nm=video 9 UDP/TLS/RTP/SAVPF 96 97 102 103 104 105 106 107 108 109 127 125 39 40 45 46 98 99 100 101 112 113 114\r\nc=IN IP4 0.0.0.0\r\na=rtcp:9 IN IP4 0.0.0.0\r\na=ice-ufrag:uB6B\r\na=ice-pwd:YmBQuaIhSa85WKcu70uyDuME\r\na=ice-options:trickle\r\na=fingerprint:sha-256 C5:94:C2:D3:FF:DF:BA:B2:67:A3:E9:B6:3A:BE:67:98:EE:61:28:00:00:E3:73:2F:A7:4C:8F:F0:99:5E:D3:77\r\na=setup:actpass\r\na=mid:1\r\na=extmap:14 urn:ietf:params:rtp-hdrext:toffset\r\na=extmap:2 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time\r\na=extmap:13 urn:3gpp:video-orientation\r\na=extmap:3 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01\r\na=extmap:5 http://www.webrtc.org/experiments/rtp-hdrext/playout-delay\r\na=extmap:6 http://www.webrtc.org/experiments/rtp-hdrext/video-content-type\r\na=extmap:7 http://www.webrtc.org/experiments/rtp-hdrext/video-timing\r\na=extmap:8 http://www.webrtc.org/experiments/rtp-hdrext/color-space\r\na=extmap:4 urn:ietf:params:rtp-hdrext:sdes:mid\r\na=extmap:10 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id\r\na=extmap:11 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id\r\na=sendrecv\r\na=msid:b98fa06f-90d2-4b82-8b51-ade1eefb38fa 253b02b7-e51b-4f44-a4f4-195cef6a16bb\r\na=rtcp-mux\r\na=rtcp-rsize\r\na=rtpmap:96 VP8/90000\r\na=rtcp-fb:96 goog-remb\r\na=rtcp-fb:96 transport-cc\r\na=rtcp-fb:96 ccm fir\r\na=rtcp-fb:96 nack\r\na=rtcp-fb:96 nack pli\r\na=rtpmap:97 rtx/90000\r\na=fmtp:97 apt=96\r\na=rtpmap:102 H264/90000\r\na=rtcp-fb:102 goog-remb\r\na=rtcp-fb:102 transport-cc\r\na=rtcp-fb:102 ccm fir\r\na=rtcp-fb:102 nack\r\na=rtcp-fb:102 nack pli\r\na=fmtp:102 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f\r\na=rtpmap:103 rtx/90000\r\na=fmtp:103 apt=102\r\na=rtpmap:104 H264/90000\r\na=rtcp-fb:104 goog-remb\r\na=rtcp-fb:104 transport-cc\r\na=rtcp-fb:104 ccm fir\r\na=rtcp-fb:104 nack\r\na=rtcp-fb:104 nack pli\r\na=fmtp:104 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f\r\na=rtpmap:105 rtx/90000\r\na=fmtp:105 apt=104\r\na=rtpmap:106 H264/90000\r\na=rtcp-fb:106 goog-remb\r\na=rtcp-fb:106 transport-cc\r\na=rtcp-fb:106 ccm fir\r\na=rtcp-fb:106 nack\r\na=rtcp-fb:106 nack pli\r\na=fmtp:106 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f\r\na=rtpmap:107 rtx/90000\r\na=fmtp:107 apt=106\r\na=rtpmap:108 H264/90000\r\na=rtcp-fb:108 goog-remb\r\na=rtcp-fb:108 transport-cc\r\na=rtcp-fb:108 ccm fir\r\na=rtcp-fb:108 nack\r\na=rtcp-fb:108 nack pli\r\na=fmtp:108 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f\r\na=rtpmap:109 rtx/90000\r\na=fmtp:109 apt=108\r\na=rtpmap:127 H264/90000\r\na=rtcp-fb:127 goog-remb\r\na=rtcp-fb:127 transport-cc\r\na=rtcp-fb:127 ccm fir\r\na=rtcp-fb:127 nack\r\na=rtcp-fb:127 nack pli\r\na=fmtp:127 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4d001f\r\na=rtpmap:125 rtx/90000\r\na=fmtp:125 apt=127\r\na=rtpmap:39 H264/90000\r\na=rtcp-fb:39 goog-remb\r\na=rtcp-fb:39 transport-cc\r\na=rtcp-fb:39 ccm fir\r\na=rtcp-fb:39 nack\r\na=rtcp-fb:39 nack pli\r\na=fmtp:39 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=4d001f\r\na=rtpmap:40 rtx/90000\r\na=fmtp:40 apt=39\r\na=rtpmap:45 AV1/90000\r\na=rtcp-fb:45 goog-remb\r\na=rtcp-fb:45 transport-cc\r\na=rtcp-fb:45 ccm fir\r\na=rtcp-fb:45 nack\r\na=rtcp-fb:45 nack pli\r\na=rtpmap:46 rtx/90000\r\na=fmtp:46 apt=45\r\na=rtpmap:98 VP9/90000\r\na=rtcp-fb:98 goog-remb\r\na=rtcp-fb:98 transport-cc\r\na=rtcp-fb:98 ccm fir\r\na=rtcp-fb:98 nack\r\na=rtcp-fb:98 nack pli\r\na=fmtp:98 profile-id=0\r\na=rtpmap:99 rtx/90000\r\na=fmtp:99 apt=98\r\na=rtpmap:100 VP9/90000\r\na=rtcp-fb:100 goog-remb\r\na=rtcp-fb:100 transport-cc\r\na=rtcp-fb:100 ccm fir\r\na=rtcp-fb:100 nack\r\na=rtcp-fb:100 nack pli\r\na=fmtp:100 profile-id=2\r\na=rtpmap:101 rtx/90000\r\na=fmtp:101 apt=100\r\na=rtpmap:112 red/90000\r\na=rtpmap:113 rtx/90000\r\na=fmtp:113 apt=112\r\na=rtpmap:114 ulpfec/90000\r\na=ssrc-group:FID 3369134375 2588693051\r\na=ssrc:3369134375 cname:RdHAkKyMp2NLvuvZ\r\na=ssrc:3369134375 msid:b98fa06f-90d2-4b82-8b51-ade1eefb38fa 253b02b7-e51b-4f44-a4f4-195cef6a16bb\r\na=ssrc:2588693051 cname:RdHAkKyMp2NLvuvZ\r\na=ssrc:2588693051 msid:b98fa06f-90d2-4b82-8b51-ade1eefb38fa 253b02b7-e51b-4f44-a4f4-195cef6a16bb\r\nm=application 9 UDP/DTLS/SCTP webrtc-datachannel\r\nc=IN IP4 0.0.0.0\r\na=ice-ufrag:uB6B\r\na=ice-pwd:YmBQuaIhSa85WKcu70uyDuME\r\na=ice-options:trickle\r\na=fingerprint:sha-256 C5:94:C2:D3:FF:DF:BA:B2:67:A3:E9:B6:3A:BE:67:98:EE:61:28:00:00:E3:73:2F:A7:4C:8F:F0:99:5E:D3:77\r\na=setup:actpass\r\na=mid:2\r\na=sctp-port:5000\r\na=max-message-size:262144\r\n",
		},
	}
	data, err = json.Marshal(jsepMessage)
	if err != nil {
		t.Error(err)
	}
	req2, err = http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		t.Error(err)
	}
	res, err = janhttpC.Do(req2)
	if err != nil {
		t.Error(err)
	}
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}
	res.Body.Close()
	if err := json.Unmarshal(body, &ackMsg); err != nil {
		t.Error(err)
	}

	expAck.ID = id

	if !reflect.DeepEqual(ackMsg, expAck) {
		t.Errorf("Expected <%s> ,got <%s> instead", utils.ToJSON(expAck), utils.ToJSON(ackMsg))
	}

	jsepEventRsp := <-pullResp
	defer jsepEventRsp.Body.Close()
	var jsepMsg []janus.EventMsg
	data, err = io.ReadAll(jsepEventRsp.Body)
	if err != nil {
		t.Error(err)
	}
	if err := json.Unmarshal(data, &jsepMsg); err != nil {
		t.Error(err)
	}

	expJsepEvnt := []janus.EventMsg{
		{
			Type:    "event",
			ID:      id,
			Session: sessionMsg.Data.ID,
			Handle:  pluginMsg.Data.ID,
			Plugindata: janus.PluginData{
				Plugin: "janus.plugin.echotest",
				Data: map[string]interface{}{
					"echotest": "event",
					"result":   "ok",
				},
			},
			Jsep: map[string]interface{}{
				"sdp":  jsepMessage.Jsep["sdp"],
				"type": "answer",
			},
		},
	}

	if !reflect.DeepEqual(jsepMsg[0].Plugindata, expJsepEvnt[0].Plugindata) || jsepMsg[0].Handle != expJsepEvnt[0].Handle || jsepMsg[0].Jsep["type"] != expJsepEvnt[0].Jsep["type"] {
		t.Errorf("Expected <%s> ,\n got <%s> instead\n", utils.ToJSON(expJsepEvnt), utils.ToJSON(jsepMsg))
	}

	close(done)

	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
