/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ees

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

func TestNewNatsEE(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}

	exp := new(NatsEE)
	exp.cfg = cfg
	exp.dc = dc
	exp.subject = utils.DefaultQueueID
	exp.reqs = newConcReq(cfg.ConcurrentRequests)
	// err = exp.parseOpt(cfg.Opts, nodeID, connTimeout)
	// if err != nil {
	// 	t.Error(err)
	// }

	rcv, err := NewNatsEE(cfg, nodeID, connTimeout, dc)
	if err != nil {
		t.Error(err)
	}
	rcv.opts = nil
	exp.opts = nil
	// fmt.Println(rcv)
	// fmt.Println(exp)

	// if exp != rcv {
	// 	t.Errorf("Expected %v \n but received \n %v", exp, rcv)
	// }
}

func TestParseOpt(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
	}
	opts := map[string]interface{}{}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, dc)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpt(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}

}

func TestParseOptJetStream(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
	}
	opts := map[string]interface{}{
		utils.NatsJetStream: true,
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, dc)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpt(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}

	if !pstr.jetStream {
		t.Error("Expected jetStream to be true")
	}

	//test error on converson
	opts = map[string]interface{}{
		utils.NatsJetStream: uint16(2),
	}

	err = pstr.parseOpt(opts, nodeID, connTimeout)

	if err.Error() != "cannot convert field: 2 to bool" {
		t.Error("The conversion shouldn't have been possible")
	}
}

func TestParseOptJetStreamMaxWait(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
	}
	opts := map[string]interface{}{
		utils.NatsJetStream:        true,
		utils.NatsJetStreamMaxWait: "2ns",
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, dc)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpt(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}
	exp := []nats.JSOpt{nats.MaxWait(2 * time.Nanosecond)}
	if !reflect.DeepEqual(pstr.jsOpts, exp) {
		t.Errorf("Expected %v \n but received \n %v", exp, pstr.jsOpts)
	}

	//test conversion error
	opts = map[string]interface{}{
		utils.NatsJetStream:        true,
		utils.NatsJetStreamMaxWait: true,
	}

	err = pstr.parseOpt(opts, nodeID, connTimeout)
	if err.Error() != "cannot convert field: true to time.Duration" {
		t.Errorf("The conversion shouldn't have been possible: %v", err.Error())
	}
}

func TestParseOptSubject(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
	}
	opts := map[string]interface{}{
		utils.NatsSubject: "nats_subject",
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	dc, err := newEEMetrics("Local")
	if err != nil {
		t.Error(err)
	}
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, dc)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpt(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}

	if pstr.subject != opts[utils.NatsSubject] {
		t.Errorf("Expected %v \n but received \n %v", opts[utils.NatsSubject], pstr.subject)
	}
}

func TestGetNatsOptsJWT(t *testing.T) {
	opts := map[string]interface{}{
		utils.NatsJWTFile: "jwtfile",
		// utils.NatsSeedFile: "file",
	}

	nodeID := "node_id1"
	connTimeout := 2 * time.Second

	_, err := GetNatsOpts(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNatsOptsClientCert(t *testing.T) {
	opts := map[string]interface{}{
		utils.NatsClientCertificate: "client_cert",
		utils.NatsClientKey:         "client_key",
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second

	_, err := GetNatsOpts(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}
	// exp := make([]nats.Option, 0, 7)
	// exp = append(exp, nats.Name(utils.CGRateSLwr+nodeID),
	// 	nats.Timeout(connTimeout),
	// 	nats.DrainTimeout(time.Second))
	// exp = append(exp, nats.ClientCert("client_cert", "client_key"))
	// if !reflect.DeepEqual(exp[3], rcv[3]) {
	// 	t.Errorf("Expected %+v \n but received \n %+v", exp, rcv)
	// }

	// no key error
	opts = map[string]interface{}{
		utils.NatsClientCertificate: "client_cert",
	}
	_, err = GetNatsOpts(opts, nodeID, connTimeout)
	if err.Error() != "has certificate but no key" {
		t.Error("There was supposedly no key")
	}

	// no certificate error
	opts = map[string]interface{}{
		utils.NatsClientKey: "client_key",
	}
	_, err = GetNatsOpts(opts, nodeID, connTimeout)
	if err.Error() != "has key but no certificate" {
		t.Error("There was supposedly no certificate")
	}
}
