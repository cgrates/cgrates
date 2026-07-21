// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// tagInternalConns adds subsystem to internal connections.
func tagInternalConns(conns []string, subsystem string) []string {
	if len(conns) == 0 {
		return conns
	}
	suffix := utils.ConcatenatedKeySep + subsystem
	result := make([]string, len(conns))
	for i, conn := range conns {
		switch conn {
		case utils.MetaInternal, rpcclient.BiRPCInternal:
			result[i] = conn + suffix
		default:
			result[i] = conn
		}
	}
	return result
}

// stripInternalConns resets all internal connection variants to base type (by
// removing the subsystem suffix).
func stripInternalConns(conns []string) []string {
	if len(conns) == 0 {
		return conns
	}
	result := make([]string, len(conns))
	for i, conn := range conns {
		switch {
		case strings.HasPrefix(conn, utils.MetaInternal):
			result[i] = utils.MetaInternal
		case strings.HasPrefix(conn, rpcclient.BiRPCInternal):
			result[i] = rpcclient.BiRPCInternal
		default:
			result[i] = conn
		}
	}
	return result
}

func tagInternalConnsOpt(opts []*DynamicConns, subsystem string) []*DynamicConns {
	if len(opts) == 0 {
		return opts
	}
	result := make([]*DynamicConns, len(opts))
	for i, opt := range opts {
		result[i] = &DynamicConns{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
			ConnIDs:   tagInternalConns(opt.ConnIDs, subsystem),
		}
	}
	return result
}

func stripInternalConnsOpt(opts []*DynamicConns) []*DynamicConns {
	if len(opts) == 0 {
		return opts
	}
	result := make([]*DynamicConns, len(opts))
	for i, opt := range opts {
		result[i] = &DynamicConns{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
			ConnIDs:   stripInternalConns(opt.ConnIDs),
		}
	}
	return result
}

func tagConns(conns map[string][]*DynamicConns) map[string][]*DynamicConns {
	if conns == nil {
		return nil
	}
	result := make(map[string][]*DynamicConns, len(conns))
	for connType, opts := range conns {
		result[connType] = tagInternalConnsOpt(opts, connType)
	}
	return result
}

func stripConns(conns map[string][]*DynamicConns) map[string][]*DynamicConns {
	if conns == nil {
		return nil
	}
	result := make(map[string][]*DynamicConns, len(conns))
	for connType, opts := range conns {
		result[connType] = stripInternalConnsOpt(opts)
	}
	return result
}
