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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewRpcEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage,
	connMgr *engine.ConnManager) (e *RPCee, err error) {
	e = &RPCee{
		cfg:     cfg,
		dc:      dc,
		connMgr: connMgr,
	}
	err = e.parseOpts()
	return
}

type RPCee struct {
	cfg     *config.EventExporterCfg
	dc      *utils.SafeMapStorage
	connMgr *engine.ConnManager

	//opts
	codec         string
	serviceMethod string
	tls           bool
	keyPath       string
	certPath      string
	caPath        string
	connIDs       []string
	connTimeout   time.Duration
	replyTimeout  time.Duration

	sync.RWMutex // protect connection
}

func (e *RPCee) Cfg() (eCfg *config.EventExporterCfg) {
	return e.cfg
}

func (e *RPCee) Connect() (err error) {
	return
}

func (e *RPCee) ExportEvent(args interface{}, _ string) (err error) {
	e.Lock()
	defer e.Unlock()
	var rply string
	return e.connMgr.Call(e.connIDs, nil, e.serviceMethod, args, &rply)
}

func (e *RPCee) Close() (err error) {
	e.Lock()
	defer e.Unlock()
	e.connMgr = nil
	return
}

func (e *RPCee) GetMetrics() (mp *utils.SafeMapStorage) {
	return e.dc
}

func (e *RPCee) PrepareMap(mp *utils.CGREvent) (interface{}, error) {
	if mp == nil {
		return nil, nil
	}
	if mp.APIOpts == nil {
		mp.APIOpts = make(map[string]interface{})
	}
	for i, v := range e.Cfg().Opts.RPCAPIOpts {
		mp.APIOpts[i] = v
	}
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
	if e.cfg.Opts.RPCCodec != nil {
		e.codec = *e.cfg.Opts.RPCCodec
	}
	if e.cfg.Opts.ServiceMethod != nil {
		e.serviceMethod = *e.cfg.Opts.ServiceMethod
	}
	if e.cfg.Opts.KeyPath != nil {
		e.keyPath = *e.cfg.Opts.KeyPath
	}
	if e.cfg.Opts.CertPath != nil {
		e.certPath = *e.cfg.Opts.CertPath
	}
	if e.cfg.Opts.CAPath != nil {
		e.caPath = *e.cfg.Opts.CAPath
	}
	if e.cfg.Opts.TLS != nil {
		e.tls = *e.cfg.Opts.TLS
	}
	if e.cfg.Opts.ConnIDs != nil {
		e.connIDs = *e.cfg.Opts.ConnIDs
	}
	if e.cfg.Opts.RPCConnTimeout != nil {
		e.connTimeout = *e.cfg.Opts.RPCConnTimeout
	}
	if e.cfg.Opts.RPCReplyTimeout != nil {
		e.replyTimeout = *e.cfg.Opts.RPCReplyTimeout
	}
	return
}
