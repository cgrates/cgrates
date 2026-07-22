//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestEscapeCharacters(t *testing.T) {

	// Set up config.
	content := `{
"general": {
	"log_level": 7,
},

"data_db": {
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"attributes": {
	"enabled": true,
},

"apiers": {
	"enabled": true
}

}`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	/*
		When escape sequences are written manually, like \u0000 in the csv file, they are not interpreted as
		escape sequences but as literal characters. So, when they are read, what will be returned is the literal
		string \u0000 instead of the null character. The *req.Password field that would be set using the csv
		below will not match "password\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000" but will match
		"password\\u0000\\u0000\\u0000\\u0000\\u0000\\u0000\\u0000\\u0000" instead.

				#Tenant,ID,Contexts,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
				cgrates.org,ATTR_TP,*any,*string:~*req.Password:password\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000,,,,,,false,20
				cgrates.org,ATTR_TP,,,,,*req.Password,*constant,processed,,
	*/

	// One of the workarounds for the issue described above is to set the profile using the Set API:

	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "ATTR_ESCAPE",
			FilterIDs: []string{
				"*string:~*req.Password:password\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Password",
					Value: config.NewRSRParsersMustCompile("processed", utils.InfieldSep),
				},
			},
			Weight: 10,
		},
	}
	var reply string
	if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &reply); err != nil {
		t.Fatal(err)
	}

	// Call AttributeSv1.ProcessEvent to check if filters match.
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Password": "password\u0000\u0000\u0000\u0000\u0000\u0000\u0000\u0000",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if rplyEv.MatchedProfiles[0] != "cgrates.org:ATTR_ESCAPE" ||
		rplyEv.Event["Password"] != "processed" ||
		rplyEv.AlteredFields[0] != "*req.Password" {
		t.Error("unexpected reply:", utils.ToJSON(rplyEv))
	}

}
