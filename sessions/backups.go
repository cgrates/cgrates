// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

type sessionBackup struct {
	OriginCGREvent *utils.CGREvent
	ClientConnID   string
	Passive        bool
	Runs           map[string]*sessionBackupRun
}

type sessionBackupRun struct {
	CGREvent        *utils.CGREvent
	InterimUsage    *utils.Decimal
	UsageAdjustment *utils.Decimal
	TotalUsage      *utils.Decimal
	Charges         *utils.EventCharges
}

type sessionBackups struct {
	nodeID   string
	interval time.Duration
	locker   *guardian.Locker

	updatesMu sync.RWMutex
	stopped   bool
	stop      chan struct{}
	done      chan struct{}

	pendingMu sync.Mutex
	pending   map[string]uint64
	version   uint64
}

// RestoreBackups loads the sessions saved by this node.
func (sS *SessionS) RestoreBackups(ctx context.Context) error {
	b := sS.backups
	if b == nil {
		return nil
	}
	backups, err := sS.dm.GetSessionBackups(ctx, b.nodeID)
	if err != nil {
		return err
	}
	for id, data := range backups {
		var backup sessionBackup
		if err := (utils.GOBMarshaler{}).Unmarshal(data, &backup); err != nil {
			return fmt.Errorf("restore session backup %q: %w", id, err)
		}
		sS.registerSession(sessionFromBackup(id, &backup), backup.Passive)
	}
	return nil
}

// StartBackups starts saving at the configured interval.
func (sS *SessionS) StartBackups() {
	if b := sS.backups; b != nil && b.interval > 0 {
		b.stop, b.done = make(chan struct{}), make(chan struct{})
		go sS.runBackups()
	}
}

func (sS *SessionS) runBackups() {
	b := sS.backups
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()
	defer close(b.done)
	for {
		select {
		case <-b.stop:
			return
		case <-ticker.C:
			if err := sS.storeBackups(context.TODO()); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> storing backups: %s", utils.SessionS, err))
			}
		}
	}
}

func (sS *SessionS) stopBackups() error {
	b := sS.backups
	b.updatesMu.Lock()
	if b.stopped {
		b.updatesMu.Unlock()
		return nil
	}
	b.stopped = true
	b.updatesMu.Unlock()
	if b.stop != nil {
		close(b.stop)
		<-b.done
	}
	for _, s := range sS.getSessions("", false) {
		s.lk.Lock()
		s.stopSTerminator()
		s.lk.Unlock()
	}
	return sS.storeBackups(context.TODO())
}

func (sS *SessionS) beginSessionUpdate() (func(), error) {
	b := sS.backups
	if b == nil {
		return func() {}, nil
	}
	b.updatesMu.RLock()
	if b.stopped {
		b.updatesMu.RUnlock()
		return nil, errors.New("SessionS is shutting down")
	}
	return b.updatesMu.RUnlock, nil
}

func (sS *SessionS) lockSessionID(id string) func() {
	if sS.backups == nil {
		return func() {}
	}
	return sS.backups.locker.Lock(id)
}

func (sS *SessionS) backupSession(ctx context.Context, id string) {
	if b := sS.backups; b != nil {
		b.pendingMu.Lock()
		b.version++
		b.pending[id] = b.version
		b.pendingMu.Unlock()
		if b.interval == -1 {
			if err := sS.storeBackupBatch(ctx, []string{id}); err != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> storing backups: %s", utils.SessionS, err))
			}
		}
	}
}

func (sS *SessionS) dropPendingBackup(id string) {
	b := sS.backups
	b.pendingMu.Lock()
	delete(b.pending, id)
	b.pendingMu.Unlock()
}

func (sS *SessionS) storeBackups(ctx context.Context) error {
	b := sS.backups
	b.pendingMu.Lock()
	ids := slices.Collect(maps.Keys(b.pending))
	b.pendingMu.Unlock()
	return sS.storeBackupBatch(ctx, ids)
}

// storeBackupBatch expects the caller to lock the CGRID or stop updates when the interval is -1.
func (sS *SessionS) storeBackupBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	b := sS.backups
	backups := make(map[string][]byte, len(ids))
	versions := make(map[string]uint64, len(ids))
	var errs []error
	for _, id := range ids {
		var unlock func()
		if b.interval != -1 {
			// Wait for the debit to finish.
			unlock = sS.lockSessionID(id)
		}
		b.pendingMu.Lock()
		version := b.pending[id]
		b.pendingMu.Unlock()
		if version == 0 {
			if unlock != nil {
				unlock()
			}
			continue
		}
		s := sS.sessionByID(id, false)
		passive := s == nil
		if passive {
			s = sS.sessionByID(id, true)
		}
		var backup []byte
		var err error
		if s != nil {
			backup, err = s.backup(passive)
		}
		if unlock != nil {
			unlock()
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("session %q: %w", id, err))
			continue
		}
		backups[id], versions[id] = backup, version
	}
	if len(backups) == 0 {
		return errors.Join(errs...)
	}
	if err := sS.dm.SetSessionBackups(ctx, b.nodeID, backups); err != nil {
		return errors.Join(append(errs, err)...)
	}
	// Keep any changes queued while the batch was being written.
	b.pendingMu.Lock()
	for id, version := range versions {
		if b.pending[id] == version {
			delete(b.pending, id)
		}
	}
	b.pendingMu.Unlock()
	return errors.Join(errs...)
}
