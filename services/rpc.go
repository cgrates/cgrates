// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"context"
	"reflect"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/utils"
)

// NewRPCService builds a named RPC service with the standard CGRateS Ping method.
func NewRPCService(rcvr any, name string) (*birpc.Service, error) {
	srv, err := birpc.NewService(rcvr, name, true)
	if err != nil {
		return nil, err
	}
	srv.Methods[utils.Ping] = pingM
	return srv, nil
}

// RegisterRPCService builds a named RPC service with Ping and registers it on cl.
func RegisterRPCService(cl *commonlisteners.CommonListenerS, rcvr any, name string) (*birpc.Service, error) {
	srv, err := NewRPCService(rcvr, name)
	if err != nil {
		return nil, err
	}
	if err := cl.RpcRegister(srv); err != nil {
		return nil, err
	}
	return srv, nil
}

func ping(_ any, _ context.Context, _ *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

var pingM = &birpc.MethodType{
	Method: reflect.Method{
		Name: utils.Ping,
		Type: reflect.TypeOf(ping),
		Func: reflect.ValueOf(ping),
	},
	ArgType:   reflect.TypeFor[*utils.CGREvent](),
	ReplyType: reflect.TypeFor[*string](),
}
