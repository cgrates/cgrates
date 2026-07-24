// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/utils"
)

// NewAnalyzerSv1 initializes AnalyzerSv1
func NewAnalyzerSv1(aS *analyzers.AnalyzerService) *AnalyzerSv1 {
	return &AnalyzerSv1{aS: aS}
}

// Exports RPC from RLs
type AnalyzerSv1 struct {
	aS *analyzers.AnalyzerService
}

// Call implements birpc.ClientConnector interface for internal RPC
func (aSv1 *AnalyzerSv1) Call(ctx *context.Context, serviceMethod string,
	args any, reply any) error {
	return utils.APIerRPCCall(aSv1, serviceMethod, args, reply)
}

// Ping return pong if the service is active
func (alSv1 *AnalyzerSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
