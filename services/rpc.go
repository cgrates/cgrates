// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"reflect"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func newRPCService(rcvr any, name string) (*birpc.Service, error) {
	srv, err := birpc.NewService(rcvr, name, true)
	if err != nil {
		return nil, err
	}
	srv.Methods[utils.Ping] = pingM
	return srv, nil
}

func ping(_ any, _ *context.Context, _ *utils.CGREvent, reply *string) error {
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
