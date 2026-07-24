// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

// RawCDR is the original CDR received from external sources (eg: FreeSWITCH)
type RawCdr interface {
	AsCDR(string) *CDR // Convert the inbound Cdr into internally used one, CgrCdr
}
