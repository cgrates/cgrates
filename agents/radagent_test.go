// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

func TestRadAgentSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(RadiusAgent))
}

func TestNewRadiusAgentFailDict(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	cfg.RadiusAgentCfg().ClientDictionaries = map[string][]string{
		"badpath": {"bad/path"},
	}
	exp := "stat bad/path: no such file or directory"
	if _, err := NewRadiusAgent(cfg, nil, nil, nil); err == nil || err.Error() != exp {
		t.Errorf("Expected error <%v>, received <%v>", exp, err)
	}
}

func TestNewRadiusAgentOK(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	exp := &RadiusAgent{
		cgrCfg: cfg,
	}
	if rcv, err := NewRadiusAgent(cfg, nil, nil, nil); err != nil {
		if err.Error() == "stat /usr/share/cgrates/radius/dict/: no such file or directory" {
			t.SkipNow() // skipping if running in gitactions
		}
		t.Error(err)
	} else if exp.RWMutex != rcv.RWMutex || exp.cgrCfg != rcv.cgrCfg || exp.connMgr != rcv.connMgr || exp.filterS != rcv.filterS || exp.dacCfg != rcv.dacCfg || exp.WaitGroup != rcv.WaitGroup {
		t.Errorf("Expected <%+v>,\nreceived\n<%+v>", exp, rcv)
	}
}

func TestNewRadiusDAClientCfgOK(t *testing.T) {

	radAgCfg := &config.RadiusAgentCfg{
		ClientDaAddresses: map[string]config.DAClientOpts{
			"udp://127.0.0.1:1813": {
				Transport: "udp",
				Host:      "127.0.0.1",
				Port:      1813,
			},
		},
	}
	dicts := &radigo.Dictionaries{}
	secrets := &radigo.Secrets{}
	exp := radiusDAClientCfg{}
	expDicts := make(map[string]*radigo.Dictionary, len(radAgCfg.ClientDaAddresses))
	expSecrets := make(map[string]string, len(radAgCfg.ClientDaAddresses))
	for client := range radAgCfg.ClientDaAddresses {
		expDicts[client] = dicts.GetInstance(client)
		exp.dicts = radigo.NewDictionaries(expDicts)
		expSecrets[client] = secrets.GetSecret(client)
		exp.secrets = radigo.NewSecrets(expSecrets)
	}

	if rcv := newRadiusDAClientCfg(dicts, secrets, radAgCfg); !reflect.DeepEqual(exp, rcv) {
		rcvStr := map[string]string{
			"dicts":   fmt.Sprint(rcv.dicts),
			"secrets": fmt.Sprint(rcv.secrets),
		}
		expStr := map[string]any{
			"dicts":   fmt.Sprint(exp.dicts),
			"secrets": fmt.Sprint(exp.secrets),
		}
		t.Errorf("Expected <%+v>,\nreceived\n<%+v>", expStr, rcvStr)
	}

}
func TestRadagentV1DisconnectPeer(t *testing.T) {
	TestAgent := &RadiusAgent{}
	ctx := context.Background()
	args := &utils.DPRArgs{}
	var reply *string
	err := TestAgent.V1DisconnectPeer(ctx, args, reply)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestRadagentV1GetActiveSessionIDs(t *testing.T) {
	TestAgent := &RadiusAgent{}
	ctx := context.Background()
	clientId := "test-client-id"
	var sessionIDs []*sessions.SessionID
	err := TestAgent.V1GetActiveSessionIDs(ctx, clientId, &sessionIDs)
	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error %v, got %v", utils.ErrNotImplemented, err)
	}

}

func TestRadagentV1WarnDisconnect(t *testing.T) {
	agent := &RadiusAgent{}
	err := agent.V1WarnDisconnect(context.Background(), map[string]any{}, nil)
	if err == nil {
		t.Errorf("Expected error 'not implemented', got nil")
	} else if err.Error() != utils.ErrNotImplemented.Error() {
		t.Errorf("Expected error 'not implemented', got '%v'", err.Error())
	}
}

func TestRadagentDaRequestAddress(t *testing.T) {
	type testCase struct {
		name         string
		remoteAddr   string
		dynAuthAddrs map[string]config.DAClientOpts
		expectedAddr string
		expectedHost string
		expectedErr  error
	}
	testCases := []testCase{
		{
			name:         "Empty dynAuthAddresses",
			remoteAddr:   "testhost:1234",
			dynAuthAddrs: map[string]config.DAClientOpts{},
			expectedAddr: "",
			expectedHost: "",
			expectedErr:  utils.ErrNotFound,
		},
		{
			name:       "Matching remote host",
			remoteAddr: "matchinghost:5678",
			dynAuthAddrs: map[string]config.DAClientOpts{
				"matchinghost": {Host: "targethost", Port: 8080},
			},
			expectedAddr: "targethost:8080",
			expectedHost: "matchinghost",
			expectedErr:  nil,
		},
		{
			name:       "Non-matching remote host",
			remoteAddr: "nonmatchinghost:9012",
			dynAuthAddrs: map[string]config.DAClientOpts{
				"otherhost": {Host: "yetanotherhost", Port: 3000},
			},
			expectedAddr: "",
			expectedHost: "",
			expectedErr:  utils.ErrNotFound,
		},
		{
			name:         "Invalid remote address format",
			remoteAddr:   "invalidformat",
			dynAuthAddrs: map[string]config.DAClientOpts{},
			expectedAddr: "",
			expectedHost: "",
			expectedErr:  utils.ErrNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr, host, err := daRequestAddress(tc.remoteAddr, tc.dynAuthAddrs)

			if err != tc.expectedErr {
				t.Errorf("Error does not match: expected %v, got %v", tc.expectedErr, err)
			}
			if addr != tc.expectedAddr {
				t.Errorf("Address does not match: expected %v, got %v", tc.expectedAddr, addr)
			}
			if host != tc.expectedHost {
				t.Errorf("Host does not match: expected %v, got %v", tc.expectedHost, host)
			}
		})
	}
}
