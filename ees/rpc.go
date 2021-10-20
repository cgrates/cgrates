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
	"strings"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewRpcEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) (e *RPCee, err error) {
	e = &RPCee{
		cfg: cfg,
		dc:  dc,
	}
	err = e.parseOpts()
	return
}

type RPCee struct {
	cfg  *config.EventExporterCfg
	dc   *utils.SafeMapStorage
	conn *rpcclient.RPCClient

	//opts
	codec         string
	serviceMethod string
	tls           bool
	keyPath       string
	certPath      string
	caPath        string
	connTimeout   time.Duration
	replyTimeout  time.Duration

	sync.RWMutex // protect connection
}

func (e *RPCee) Cfg() (eCfg *config.EventExporterCfg) {
	return e.cfg
}

func (e *RPCee) Connect() (err error) {
	e.Lock()
	if e.conn != nil {
		var conn *rpcclient.RPCClient
		conn, err = rpcclient.NewRPCClient(context.TODO(), utils.TCP, e.cfg.ExportPath, e.tls,
			e.keyPath, e.certPath, e.caPath, 1, 1, e.connTimeout, e.replyTimeout, e.codec, nil, false, nil)
		if err == nil {
			e.conn = conn
		}
	}
	e.Unlock()
	return
}

func (e *RPCee) ExportEvent(ctx *context.Context, args interface{}, _ string) (err error) {
	e.Lock()
	defer e.Unlock()
	return e.conn.Call(ctx, e.serviceMethod, args, nil)
}

func (e *RPCee) Close() (err error) {
	e.Lock()
	defer e.Unlock()
	e.conn = nil
	return
}

func (e *RPCee) GetMetrics() (mp *utils.SafeMapStorage) {
	return e.dc
}

func (e *RPCee) PrepareMap(mp map[string]interface{}) (interface{}, error) {
	return mp, nil
}

func (e *RPCee) PrepareOrderMap(oMp *utils.OrderedNavigableMap) (interface{}, error) {
	mP := make(map[string]interface{})
	for i := oMp.GetFirstElement(); i != nil; i = i.Next() {
		path := i.Value
		val, _ := oMp.Field(path)
		if val.AttributeID != utils.EmptyString {
			continue
		}
		path = path[:len(path)-1] // remove the last index
		opath := strings.Join(path, utils.NestingSep)
		if _, has := mP[opath]; !has {
			mP[opath] = val.Data // first item which is not an attribute will become the value
		}
	}
	return mP, nil
}

func (e *RPCee) parseOpts() (err error) {
	e.codec = utils.IfaceAsString(e.cfg.Opts[utils.RpcCodec])
	e.serviceMethod = utils.IfaceAsString(e.cfg.Opts[utils.ServiceMethod])
	e.keyPath = utils.IfaceAsString(e.cfg.Opts[utils.KeyPath])
	e.certPath = utils.IfaceAsString(e.cfg.Opts[utils.CertPath])
	e.caPath = utils.IfaceAsString(e.cfg.Opts[utils.CaPath])
	if e.tls, err = utils.IfaceAsBool(e.cfg.Opts[utils.Tls]); err != nil {
		return
	}
	if e.connTimeout, err = utils.IfaceAsDuration(e.cfg.Opts[utils.RpcConnTimeout]); err != nil {
		return
	}
	if e.replyTimeout, err = utils.IfaceAsDuration(e.cfg.Opts[utils.RpcReplyTimeout]); err != nil {
		return
	}
	return
}
