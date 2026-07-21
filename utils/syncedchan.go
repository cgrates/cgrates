// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"sync"
)

func NewSyncedChan() *SyncedChan {
	return &SyncedChan{
		c: make(chan struct{}),
		d: new(sync.Once),
	}
}

type SyncedChan struct {
	c chan struct{}
	d *sync.Once
}

func (s *SyncedChan) CloseOnce() {
	s.d.Do(func() {
		close(s.c)
	})
}

func (s *SyncedChan) Done() <-chan struct{} {
	return s.c
}
