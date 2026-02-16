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
	"testing"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
)

func TestDiamULI(t *testing.T) {
	// t.Skip("configuration reference for *3gpp_uli request_fields; does not verify anything")
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	cfgJSON := `{
    "sessions": {
        "enabled": true
    },
    "diameter_agent": {
        "enabled": true,
        "request_processors": [
            {
                "id": "tgpp_loc_info",
                "flags": [
                    "*dryrun"
                ],
                "request_fields": [
                    {
                        "tag": "ULI",
                        "path": "*cgreq.ULI",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info"
                    },
                    {
                        "tag": "DecodedULI",
                        "path": "*cgreq.DecodedULI",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli}"
                    },
                    {
                        "tag": "TAI",
                        "path": "*cgreq.TAI",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:TAI}"
                    },
                    {
                        "tag": "TAI-MCC",
                        "path": "*cgreq.TAI-MCC",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:TAI.MCC}"
                    },
                    {
                        "tag": "TAI-MNC",
                        "path": "*cgreq.TAI-MNC",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:TAI.MNC}"
                    },
                    {
                        "tag": "TAI-TAC",
                        "path": "*cgreq.TAI-TAC",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:TAI.TAC}"
                    },
                    {
                        "tag": "ECGI",
                        "path": "*cgreq.ECGI",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:ECGI}"
                    },
                    {
                        "tag": "ECGI-MCC",
                        "path": "*cgreq.ECGI-MCC",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:ECGI.MCC}"
                    },
                    {
                        "tag": "ECGI-MNC",
                        "path": "*cgreq.ECGI-MNC",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:ECGI.MNC}"
                    },
                    {
                        "tag": "ECGI-ECI",
                        "path": "*cgreq.ECGI-ECI",
                        "type": "*variable",
                        "value": "~*req.Service-Information.PS-Information.3GPP-User-Location-Info{*3gpp_uli:ECGI.ECI}"
                    }
                ],
                "reply_fields": []
            }
        ]
    }
}`

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		DBCfg:      engine.InternalDBCfg,
		LogBuffer:  &bytes.Buffer{},
	}
	t.Cleanup(func() {
		t.Log(ng.LogBuffer)
	})
	_, cfg := ng.Run(t)

	diamClient, err := agents.NewDiameterClient(cfg.DiameterAgentCfg().Listeners[0].Address, "localhost",
		cfg.DiameterAgentCfg().OriginRealm, cfg.DiameterAgentCfg().VendorID,
		cfg.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		cfg.DiameterAgentCfg().DictionariesPath, cfg.DiameterAgentCfg().Listeners[0].Network)
	if err != nil {
		t.Fatal(err)
	}

	// Binary ULI from Wireshark capture: TAI+ECGI, MCC=547, MNC=05, TAC=1, ECI=257
	uliBytes, err := hex.DecodeString("8245f750000145f75000000101")
	if err != nil {
		t.Fatal(err)
	}
	ccr := diam.NewRequest(diam.CreditControl, 4, nil)
	ccr.NewAVP(avp.ServiceInformation, avp.Mbit, 10415,
		&diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.PSInformation, avp.Mbit, 10415,
					&diam.GroupedAVP{
						AVP: []*diam.AVP{
							diam.NewAVP(avp.TGPPUserLocationInfo, avp.Mbit, 10415, datatype.OctetString(uliBytes)),
						},
					},
				),
			},
		},
	)

	if err := diamClient.SendMessage(ccr); err != nil {
		t.Errorf("failed to send diameter message: %v", err)
	}
	_ = diamClient.ReceivedMessage(2 * time.Second)
}
