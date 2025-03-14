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

package ees

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewNatsEE(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
		Opts: &config.EventExporterOpts{
			AMQP:  &config.AMQPOpts{},
			Els:   &config.ElsOpts{},
			AWS:   &config.AWSOpts{},
			NATS:  &config.NATSOpts{},
			Kafka: &config.KafkaOpts{},
			RPC:   &config.RPCOpts{},
		},
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	em := utils.NewExporterMetrics("", time.Local)

	exp := new(NatsEE)
	exp.cfg = cfg
	exp.em = em
	exp.subject = utils.DefaultQueueID
	exp.reqs = newConcReq(cfg.ConcurrentRequests)
	// err = exp.parseOpt(cfg.Opts, nodeID, connTimeout)
	// if err != nil {
	// 	t.Error(err)
	// }

	rcv, err := NewNatsEE(cfg, nodeID, connTimeout, em)
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
		Opts: &config.EventExporterOpts{
			AMQP:  &config.AMQPOpts{},
			Els:   &config.ElsOpts{},
			AWS:   &config.AWSOpts{},
			NATS:  &config.NATSOpts{},
			Kafka: &config.KafkaOpts{},
			RPC:   &config.RPCOpts{},
		},
	}
	opts := &config.EventExporterOpts{}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	em := utils.NewExporterMetrics("", time.Local)
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, em)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpts(opts.NATS, nodeID, connTimeout)
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
		Opts: &config.EventExporterOpts{
			AMQP:  &config.AMQPOpts{},
			Els:   &config.ElsOpts{},
			AWS:   &config.AWSOpts{},
			NATS:  &config.NATSOpts{},
			Kafka: &config.KafkaOpts{},
			RPC:   &config.RPCOpts{},
		},
	}
	opts := &config.EventExporterOpts{
		NATS: &config.NATSOpts{
			JetStream: utils.BoolPointer(true)},
	}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	em := utils.NewExporterMetrics("", time.Local)
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, em)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpts(opts.NATS, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}

	if !pstr.jetStream {
		t.Error("Expected jetStream to be true")
	}
}

func TestParseOptSubject(t *testing.T) {
	cfg := &config.EventExporterCfg{
		ID:                 "nats_exporter",
		Type:               "nats",
		Attempts:           2,
		ConcurrentRequests: 2,
		Opts: &config.EventExporterOpts{
			AMQP:  &config.AMQPOpts{},
			Els:   &config.ElsOpts{},
			AWS:   &config.AWSOpts{},
			NATS:  &config.NATSOpts{},
			Kafka: &config.KafkaOpts{},
			RPC:   &config.RPCOpts{},
		},
	}
	opts := &config.EventExporterOpts{
		NATS: &config.NATSOpts{
			Subject: utils.StringPointer("nats_subject"),
		}}
	nodeID := "node_id1"
	connTimeout := 2 * time.Second
	em := utils.NewExporterMetrics("", time.Local)
	pstr, err := NewNatsEE(cfg, nodeID, connTimeout, em)
	if err != nil {
		t.Error(err)
	}

	err = pstr.parseOpts(opts.NATS, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}

	if opts.NATS.Subject == nil || pstr.subject != *opts.NATS.Subject {
		t.Errorf("Expected %v \n but received \n %v", *opts.NATS.Subject, pstr.subject)
	}
}

func TestGetNatsOptsJWT(t *testing.T) {
	opts := &config.NATSOpts{
		JWTFile: utils.StringPointer("jwtfile"),
	}

	nodeID := "node_id1"
	connTimeout := 2 * time.Second

	_, err := GetNatsOpts(opts, nodeID, connTimeout)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNatsOptsClientCert(t *testing.T) {
	opts := &config.NATSOpts{
		ClientCertificate: utils.StringPointer("client_cert"),
		ClientKey:         utils.StringPointer("client_key"),
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
	opts = &config.NATSOpts{
		ClientCertificate: utils.StringPointer("client_cert"),
	}
	_, err = GetNatsOpts(opts, nodeID, connTimeout)
	if err.Error() != "has certificate but no key" {
		t.Error("There was supposedly no key")
	}

	// no certificate error
	opts = &config.NATSOpts{
		ClientKey: utils.StringPointer("client_key"),
	}
	_, err = GetNatsOpts(opts, nodeID, connTimeout)
	if err.Error() != "has key but no certificate" {
		t.Error("There was supposedly no certificate")
	}
}

func TestNatsEECfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	pstr := &NatsEE{
		cfg: expectedCfg,
	}
	actualCfg := pstr.Cfg()
	if actualCfg != expectedCfg {
		t.Errorf("Cfg() = %v, want %v", actualCfg, expectedCfg)
	}
}

func TestNatsEEGetMetrics(t *testing.T) {
	expectedMetrics := &utils.ExporterMetrics{}
	pstr := &NatsEE{
		em: expectedMetrics,
	}
	actualMetrics := pstr.GetMetrics()
	if actualMetrics != expectedMetrics {
		t.Errorf("GetMetrics() = %v, want %v", actualMetrics, expectedMetrics)
	}
}
