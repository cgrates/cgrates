//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

func TestRadiusStat(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	var testRadiusDict = `
	ATTRIBUTE	Message-Authenticator	80		octets
	ATTRIBUTE   User-Name       1   string
	ATTRIBUTE   User-Password       2   string
	ATTRIBUTE   CHAP-Password       3   string
	ATTRIBUTE   NAS-IP-Address      4   ipaddr
	ATTRIBUTE   NAS-Port        5   integer
	ATTRIBUTE   Service-Type        6   integer
	ATTRIBUTE   Framed-Protocol     7   integer
	ATTRIBUTE   Framed-IP-Address   8   ipaddr
	ATTRIBUTE   Framed-IP-Netmask   9   ipaddr
	ATTRIBUTE   Framed-Routing      10  integer
	ATTRIBUTE   Filter-Id       11  string
	ATTRIBUTE   Framed-MTU      12  integer
	ATTRIBUTE   Framed-Compression  13  integer
	ATTRIBUTE   Login-IP-Host       14  ipaddr
	ATTRIBUTE   Login-Service       15  integer
	ATTRIBUTE   Login-TCP-Port      16  integer
	ATTRIBUTE   Reply-Message       18  string
	ATTRIBUTE   Callback-Number     19  string
	ATTRIBUTE   Callback-Id     20  string
	ATTRIBUTE   Framed-Route        22  string
	ATTRIBUTE   Framed-IPX-Network  23  ipaddr
	ATTRIBUTE   State           24  string
	ATTRIBUTE   Class           25  string
	ATTRIBUTE   Vendor-Specific     26  string
	ATTRIBUTE   Session-Timeout     27  integer
	ATTRIBUTE   Idle-Timeout        28  integer
	ATTRIBUTE   Termination-Action  29  integer
	ATTRIBUTE   Called-Station-Id   30  string
	ATTRIBUTE   Calling-Station-Id  31  string
	ATTRIBUTE   NAS-Identifier      32  string
	ATTRIBUTE   Proxy-State     33  string
	ATTRIBUTE   Login-LAT-Service   34  string
	ATTRIBUTE   Login-LAT-Node      35  string
	ATTRIBUTE   Login-LAT-Group     36  string
	ATTRIBUTE   Framed-AppleTalk-Link   37  integer
	ATTRIBUTE   Framed-AppleTalk-Network    38  integer
	ATTRIBUTE   Framed-AppleTalk-Zone   39  string
	ATTRIBUTE   Acct-Status-Type    40  integer
	ATTRIBUTE   Acct-Delay-Time     41  integer
	ATTRIBUTE   Acct-Input-Octets   42  integer
	ATTRIBUTE   Acct-Output-Octets  43  integer
	ATTRIBUTE   Acct-Session-Id     44  string
	ATTRIBUTE   Acct-Authentic      45  integer
	ATTRIBUTE   Acct-Session-Time   46  integer
	ATTRIBUTE   Acct-Input-Packets  47  integer
	ATTRIBUTE   Acct-Output-Packets 48  integer
	ATTRIBUTE   Acct-Terminate-Cause    49  integer
	ATTRIBUTE   Acct-Multi-Session-Id   50  string
	ATTRIBUTE   Acct-Link-Count     51  integer
	ATTRIBUTE   Acct-Input-Gigawords    52  integer
	ATTRIBUTE   Acct-Output-Gigawords   53  integer
	ATTRIBUTE   Event-Timestamp     55  integer
	ATTRIBUTE   Egress-VLANID       56  string
	ATTRIBUTE   Ingress-Filters     57  integer
	ATTRIBUTE   Egress-VLAN-Name    58  string
	ATTRIBUTE   User-Priority-Table 59  string
	ATTRIBUTE   CHAP-Challenge      60  string
	ATTRIBUTE   NAS-Port-Type       61  integer
	ATTRIBUTE   Port-Limit      62  integer
	ATTRIBUTE   Login-LAT-Port      63  integer
	`

	dictDir := t.TempDir()
	dictPath := filepath.Join(dictDir, "dictionary.test")
	if err := os.WriteFile(dictPath, []byte(testRadiusDict), 0644); err != nil {
		t.Fatal(err)
	}

	ng := engine.TestEngine{
		ConfigPath: filepath.Join(*utils.DataDir, "conf", "samples", "radius_status"),
		ConfigJSON: fmt.Sprintf(`{
			"radius_agent": {
				"client_dictionaries": {
					"*default": [
						%q
					]
				}
			}
			}`, dictDir+"/"),
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	_, cfg := ng.Run(t)

	dictRad := radigo.RFC2865Dictionary()
	dictRad.ParseFromReader(strings.NewReader(testRadiusDict))
	secret := cfg.RadiusAgentCfg().ClientSecrets[utils.MetaDefault]
	net := cfg.RadiusAgentCfg().Listeners[0].Network
	authAddr := cfg.RadiusAgentCfg().Listeners[0].AuthAddr
	acctAddr := cfg.RadiusAgentCfg().Listeners[0].AcctAddr
	clientAuth, err := radigo.NewClient(net, authAddr, secret, dictRad, 1, nil, utils.Logger)
	if err != nil {
		t.Fatal(err)
	}

	req := clientAuth.NewRequest(radigo.StatusServer, 71)

	if err := req.AddAVPWithName("NAS-Identifier", "Status Check 1806. Are you alive?", ""); err != nil {
		t.Fatal(err)
	}

	if err := req.AddAVPWithName("Message-Authenticator", "A7kLm29qXtP4vWcE0uYdRgHsJnFbZxQ3", ""); err != nil {
		t.Fatal(err)
	}

	replyPacket, err := clientAuth.SendRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	if len(replyPacket.AVPs) > 1 {
		t.Errorf("Expected 1 AVP, received %v AVPS", len(replyPacket.AVPs))
	}

	if replyPacket.AVPs[0].Number != 18 {
		t.Errorf("Expected 18 , received %v", replyPacket.AVPs[0].Number)
	}

	if string(replyPacket.AVPs[0].RawValue) != "OK" {
		t.Errorf("Expected Reply-Message 'OK', received %v", string(replyPacket.AVPs[0].RawValue))
	}

	clientAcct, err := radigo.NewClient(net, acctAddr, secret, dictRad, 1, nil, utils.Logger)
	if err != nil {
		t.Fatal(err)
	}

	req = clientAcct.NewRequest(radigo.StatusServer, 71)

	if err := req.AddAVPWithName("NAS-Identifier", "Status Check 1806. Are you alive?", ""); err != nil {
		t.Fatal(err)
	}

	if err := req.AddAVPWithName("Message-Authenticator", "A7kLm29qXtP4vWcE0uYdRgHsJnFbZxQ3", ""); err != nil {
		t.Fatal(err)
	}

	replyPacket, err = clientAcct.SendRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	if len(replyPacket.AVPs) > 1 {
		t.Errorf("Expected 1 AVP, received %v AVPS", len(replyPacket.AVPs))
	}

	if replyPacket.AVPs[0].Number != 18 {
		t.Errorf("Expected 18 , received %v", replyPacket.AVPs[0].Number)
	}

	if string(replyPacket.AVPs[0].RawValue) != "OK" {
		t.Errorf("Expected Reply-Message 'OK', received %v", string(replyPacket.AVPs[0].RawValue))
	}

}
