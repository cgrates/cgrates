//go:build integration

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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

func TestRadiusULI(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	var testDict = `
ATTRIBUTE   User-Name           1   string
ATTRIBUTE   NAS-IP-Address      4   ipaddr
ATTRIBUTE   NAS-Port            5   integer
ATTRIBUTE   Service-Type        6   integer
ATTRIBUTE   Acct-Status-Type    40  integer
ATTRIBUTE   Acct-Session-Id     44  string
ATTRIBUTE   Event-Timestamp     55  integer

VALUE   Acct-Status-Type    Start   1

VENDOR      3GPP    10415

BEGIN-VENDOR 3GPP
ATTRIBUTE   3GPP-User-Location-Info    22    octets
END-VENDOR 3GPP
`

	dictDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dictDir, "dictionary.test"),
		[]byte(testDict), 0644); err != nil {
		t.Fatal(err)
	}

	cfgJSON := fmt.Sprintf(`{
    "sessions": {
        "enabled": true
    },
    "radius_agent": {
        "enabled": true,
        "client_dictionaries": {
            "*default": [%q]
        },
        "request_processors": [
            {
                "id": "uli_extract",
                "flags": ["*dryRun"],
                "request_fields": [
                    {
                        "tag": "DecodedULI",
                        "path": "*cgreq.DecodedULI",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli}"
                    },
                    {
                        "tag": "TAI-MCC",
                        "path": "*cgreq.TAI-MCC",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli:TAI.MCC}"
                    },
                    {
                        "tag": "TAI-MNC",
                        "path": "*cgreq.TAI-MNC",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli:TAI.MNC}"
                    },
                    {
                        "tag": "TAI-TAC",
                        "path": "*cgreq.TAI-TAC",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli:TAI.TAC}"
                    },
                    {
                        "tag": "ECGI-ECI",
                        "path": "*cgreq.ECGI-ECI",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli:ECGI.ECI}"
                    },
                    {
                        "tag": "MCC-Name",
                        "path": "*cgreq.MCC-Name",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli:TAI.MCC.Name}"
                    },
                    {
                        "tag": "MNC-Name",
                        "path": "*cgreq.MNC-Name",
                        "type": "*variable",
                        "value": "~*req.3GPP.3GPP-User-Location-Info{*3gpp_uli:TAI.MNC.Name}"
                    }
                ],
                "reply_fields": []
            }
        ]
    }
}`, dictDir+"/")

	buf := &bytes.Buffer{}
	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
		LogBuffer:  buf,
	}
	_, cfg := ng.Run(t)

	dictRad := radigo.RFC2865Dictionary()
	dictRad.ParseFromReader(strings.NewReader(testDict))
	secret := cfg.RadiusAgentCfg().ClientSecrets[utils.MetaDefault]
	net := cfg.RadiusAgentCfg().Listeners[0].Network
	acctAddr := cfg.RadiusAgentCfg().Listeners[0].AcctAddr
	client, err := radigo.NewClient(net, acctAddr, secret, dictRad, 1, nil, utils.Logger)
	if err != nil {
		t.Fatal(err)
	}

	// Binary ULI: TAI+ECGI, MCC=262, MNC=01, TAC=1, ECI=257
	uliBytes, err := hex.DecodeString("8262f210000162f21000000101")
	if err != nil {
		t.Fatal(err)
	}

	req := client.NewRequest(radigo.AccountingRequest, 1)
	req.AddAVPWithName("User-Name", "TestULI", "")
	req.AddAVPWithName("Acct-Status-Type", "Start", "")
	req.AddAVPWithName("Acct-Session-Id", "uli-test-session", "")
	req.AddAVPWithName("3GPP-User-Location-Info", string(uliBytes), "3GPP")

	if _, err := client.SendRequest(req); err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"DecodedULI": `{"TAI":{"MCC":"262","MNC":"01","TAC":1},"ECGI":{"MCC":"262","MNC":"01","ECI":257}}`,
		"TAI-MCC":    "262",
		"TAI-MNC":    "01",
		"TAI-TAC":    "1",
		"ECGI-ECI":   "257",
		"MCC-Name":   "Germany",
		"MNC-Name":   "Telekom Deutschland GmbH",
	}

	parts := strings.Split(buf.String(), "CGREvent: ")
	if len(parts) < 2 {
		t.Fatalf("no CGREvent found in dryrun log:\n%s", buf.String())
	}

	var ev utils.CGREvent
	if err := json.NewDecoder(strings.NewReader(parts[len(parts)-1])).Decode(&ev); err != nil {
		t.Fatalf("failed to decode CGREvent: %v", err)
	}

	for field, want := range expected {
		got := utils.IfaceAsString(ev.Event[field])
		if got != want {
			t.Errorf("%s: got %q, want %q", field, got, want)
		}
	}
}
