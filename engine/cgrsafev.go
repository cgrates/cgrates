// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

func NewCGRSafEventFromCGREvent(cgrEv *utils.CGREvent) *CGRSafEvent {
	return &CGRSafEvent{
		Tenant: cgrEv.Tenant,
		ID:     cgrEv.ID,
		Time:   cgrEv.Time,
		Event:  NewSafEvent(cgrEv.Event),
	}
}

// CGRSafEvent is a safe CGREvent
type CGRSafEvent struct {
	Tenant string
	ID     string
	Time   *time.Time // event time
	Event  *SafEvent
}

func (cgrSafEv *CGRSafEvent) AsCGREvent() *utils.CGREvent {
	return &utils.CGREvent{
		Tenant: cgrSafEv.Tenant,
		ID:     cgrSafEv.ID,
		Time:   cgrSafEv.Time,
		Event:  cgrSafEv.Event.AsMapInterface(),
	}
}
