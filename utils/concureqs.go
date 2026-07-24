// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
)

var ConReqs *ConcReqs

type ConcReqs struct {
	limit    int
	strategy string
	aReqs    chan struct{}
}

func NewConReqs(reqs int, strategy string) *ConcReqs {
	cR := &ConcReqs{
		limit:    reqs,
		strategy: strategy,
		aReqs:    make(chan struct{}, reqs),
	}
	for i := 0; i < reqs; i++ {
		cR.aReqs <- struct{}{}
	}
	return cR
}

var errDeny = fmt.Errorf("denying request due to maximum active requests reached")

func (cR *ConcReqs) Allocate() (err error) {
	if cR.limit == 0 {
		return
	}
	switch cR.strategy {
	case MetaBusy:
		if len(cR.aReqs) == 0 {
			return errDeny
		}
		fallthrough
	case MetaQueue:
		<-cR.aReqs // get from channel
	}
	return
}

func (cR *ConcReqs) Deallocate() {
	if cR.limit == 0 {
		return
	}
	cR.aReqs <- struct{}{}
}
