/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"math/rand"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewDiameterClient(addr, originHost, originRealm string, vendorId int, productName string, firmwareRev int) (*DiameterClient, error) {
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(originHost),
		OriginRealm:      datatype.DiameterIdentity(originRealm),
		VendorID:         datatype.Unsigned32(vendorId),
		ProductName:      datatype.UTF8String(productName),
		FirmwareRevision: datatype.Unsigned32(firmwareRev),
	}
	handlers := sm.New(cfg)
	cli := &sm.Client{
		Handler:            handlers,
		MaxRetransmits:     3,
		RetransmitInterval: time.Second,
		EnableWatchdog:     true,
		WatchdogInterval:   5 * time.Second,
		AcctApplicationID: []*diam.AVP{
			// Advertise that we want support for both
			// Accounting applications 4 and 999.
			diam.NewAVP(avp.AcctApplicationID, avp.Mbit, 0, datatype.Unsigned32(4)), // RFC 4006
		},
	}
	conn, err := cli.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &DiameterClient{conn: conn, handlers: handlers}, nil
}

type DiameterClient struct {
	conn     diam.Conn
	handlers diam.Handler
}
