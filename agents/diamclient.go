// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/go-diameter/diam"
	"github.com/cgrates/go-diameter/diam/avp"
	"github.com/cgrates/go-diameter/diam/datatype"
	"github.com/cgrates/go-diameter/diam/dict"
	"github.com/cgrates/go-diameter/diam/sm"
)

var dictOnce sync.Once

func NewDiameterClient(addr, originHost, originRealm string, vendorId int, productName string,
	firmwareRev int, dictsDir string, dictsAppendDefaults bool, network string) (dc *DiameterClient, err error) {
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity(originHost),
		OriginRealm:      datatype.DiameterIdentity(originRealm),
		VendorID:         datatype.Unsigned32(vendorId),
		ProductName:      datatype.UTF8String(productName),
		FirmwareRevision: datatype.Unsigned32(firmwareRev),
		Dict:             dict.Default,
	}
	if len(dictsDir) != 0 {
		if !dictsAppendDefaults {
			if cfg.Dict, err = dict.NewParser(); err != nil {
				return nil, err
			}
		}
		dictOnce.Do(func() { err = loadDictionaries(cfg.Dict, dictsDir, "DiameterClient") })
		if err != nil {
			return nil, err
		}
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
				net.ParseIP(strings.Split(iAddr.String(), utils.HDRValSep)[0]))) // address came in form x.y.z.t/24
		}
	}
	dSM := sm.New(cfg)
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
	conn, err := cli.DialNetwork(network, addr)
	if err != nil {
		return nil, err
	}
	dc = &DiameterClient{
		conn:     conn,
		handlers: dSM,
		received: make(chan *diam.Message, 1),
		pending:  make(map[uint32]chan *diam.Message),
	}
	var closed <-chan struct{}
	if notifier, ok := conn.(diam.CloseNotifier); ok {
		closed = notifier.CloseNotify()
	}
	go func() {
		for {
			select {
			case err := <-dSM.ErrorReports():
				utils.Logger.Err(fmt.Sprintf("<DiameterClient> StateMachine error: %+v", err))
			case <-closed:
				return
			}
		}
	}()
	dSM.HandleFunc("ALL", dc.handleALL)
	return dc, nil
}

type DiameterClient struct {
	conn     diam.Conn
	handlers diam.Handler
	received chan *diam.Message

	pendingMu sync.Mutex
	pending   map[uint32]chan *diam.Message
}

func (dc *DiameterClient) SendMessage(m *diam.Message) error {
	_, err := m.WriteTo(dc.conn)
	return err
}

// RoundTrip sends m and waits for the answer with the same Hop-by-Hop ID.
func (dc *DiameterClient) RoundTrip(m *diam.Message, replyTimeout time.Duration) (*diam.Message, error) {
	hopID := m.Header.HopByHopID
	waiter := make(chan *diam.Message, 1)

	dc.pendingMu.Lock()
	if _, found := dc.pending[hopID]; found {
		dc.pendingMu.Unlock()
		return nil, fmt.Errorf("Diameter Hop-by-Hop ID %d is already pending", hopID)
	}
	dc.pending[hopID] = waiter
	dc.pendingMu.Unlock()

	if err := dc.SendMessage(m); err != nil {
		dc.removePending(hopID, waiter)
		return nil, err
	}
	timer := time.NewTimer(replyTimeout)
	defer timer.Stop()
	select {
	case reply := <-waiter:
		return reply, nil
	case <-timer.C:
		if dc.removePending(hopID, waiter) {
			return nil, fmt.Errorf("timeout waiting for Diameter Hop-by-Hop ID %d", hopID)
		}
		return <-waiter, nil
	}
}

func (dc *DiameterClient) removePending(hopID uint32, waiter chan *diam.Message) bool {
	dc.pendingMu.Lock()
	defer dc.pendingMu.Unlock()
	if dc.pending[hopID] != waiter {
		return false
	}
	delete(dc.pending, hopID)
	return true
}

func (dc *DiameterClient) handleALL(c diam.Conn, m *diam.Message) {
	if m.Header.CommandFlags&diam.RequestFlag == 0 {
		dc.pendingMu.Lock()
		waiter, found := dc.pending[m.Header.HopByHopID]
		if found {
			delete(dc.pending, m.Header.HopByHopID)
		}
		dc.pendingMu.Unlock()
		if found {
			waiter <- m
			return
		}
	}
	utils.Logger.Warning(fmt.Sprintf("<DiameterClient> Received unexpected message from %s:\n%s", c.RemoteAddr(), m))
	dc.received <- m
}

// ReceivedMessage returns the next unmatched message, or nil after the timeout.
func (dc *DiameterClient) ReceivedMessage(rplyTimeout time.Duration) *diam.Message {
	select {
	case rcv := <-dc.received:
		return rcv
	case <-time.After(rplyTimeout): // Timeout reading
		return nil
	}
}

// Close disconnects the DiameterClient. Implements io.Closer.
func (dc *DiameterClient) Close() error {
	dc.conn.Close()
	return nil
}
