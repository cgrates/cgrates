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
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

func NewDiameterClient(addr, originHost, originRealm string, vendorId int, productName string,
	firmwareRev int, dictsDir string, network string) (*DiameterClient, error) {
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(originHost),
		OriginRealm:      datatype.DiameterIdentity(originRealm),
		VendorID:         datatype.Unsigned32(vendorId),
		ProductName:      datatype.UTF8String(productName),
		FirmwareRevision: datatype.Unsigned32(firmwareRev),
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, inter := range interfaces {
		addrs, err := inter.Addrs()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<DiameterClient> error: %+v, when taking address from interface: %+v",
				err, inter.Name))
			continue
		}
		for _, iAddr := range addrs {
			cfg.HostIPAddresses = append(cfg.HostIPAddresses, datatype.Address(
				strings.Split(iAddr.String(), utils.HDR_VAL_SEP)[0])) // address came in form x.y.z.t/24
		}
	}
	dSM := sm.New(cfg)
	go func() {
		for err := range dSM.ErrorReports() {
			utils.Logger.Err(fmt.Sprintf("<DiameterClient> StateMachine error: %+v", err))
		}
	}()
	cli := &sm.Client{
		Handler:            dSM,
		MaxRetransmits:     3,
		RetransmitInterval: time.Second,
		EnableWatchdog:     true,
		WatchdogInterval:   5 * time.Second,
		AuthApplicationID: []*diam.AVP{
			// Advertise support for credit control application
			diam.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4)), // RFC 4006
		},
	}
	if len(dictsDir) != 0 {
		if err := loadDictionaries(dictsDir, "DiameterClient"); err != nil {
			return nil, err
		}
	}
	conn, err := cli.DialNetwork(network, addr)
	if err != nil {
		return nil, err
	}
	dc := &DiameterClient{conn: conn, handlers: dSM, received: make(chan *diam.Message)}
	dSM.HandleFunc("ALL", dc.handleALL)
	return dc, nil
}

type DiameterClient struct {
	conn     diam.Conn
	handlers diam.Handler
	received chan *diam.Message
}

func (dc *DiameterClient) SendMessage(m *diam.Message) error {
	_, err := m.WriteTo(dc.conn)
	return err
}

func (dc *DiameterClient) handleALL(c diam.Conn, m *diam.Message) {
	utils.Logger.Warning(fmt.Sprintf("<DiameterClient> Received unexpected message from %s:\n%s", c.RemoteAddr(), m))
	dc.received <- m
}

// Returns the message out of received buffer
func (dc *DiameterClient) ReceivedMessage(rplyTimeout time.Duration) *diam.Message {
	select {
	case rcv := <-dc.received:
		return rcv
	case <-time.After(rplyTimeout): // Timeout reading
		return nil
	}
}
