// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package loaders

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type lockResult struct {
	unlock func()
	err    error
}

func TestNopLocker(t *testing.T) {
	locker := newLoaderLocker("", "", engine.NewLocker(nil))
	unlock, err := locker.lock()
	if err != nil {
		t.Fatal(err)
	}
	unlock()
	if locked, err := locker.locked(); err != nil {
		t.Fatal(err)
	} else if locked {
		t.Fatal("expected no lock")
	}
	if err := locker.forceUnlock(); err != nil {
		t.Fatal(err)
	}
	if locker.isLockFile("") {
		t.Fatal("expected no lock file")
	}
}

func TestFolderLocker(t *testing.T) {
	filePath := path.Join(t.TempDir(), ".lkr")
	locker := newLoaderLocker(filePath, "", engine.NewLocker(nil))
	unlock, err := locker.lock()
	if err != nil {
		t.Fatal(err)
	}
	if locked, err := locker.locked(); err != nil {
		t.Fatal(err)
	} else if !locked {
		t.Fatal("expected lock")
	}
	if locker.isLockFile("") {
		t.Fatal("expected no match")
	}
	if !locker.isLockFile(filePath) {
		t.Fatal("expected lock file match")
	}
	unlock()
	if locked, err := locker.locked(); err != nil {
		t.Fatal(err)
	} else if locked {
		t.Fatal("expected no lock")
	}
}

func TestMemoryLocker(t *testing.T) {
	memory := engine.NewLocker(nil)
	locker := newLoaderLocker(utils.MetaMemory, "ID", memory)
	unlock, err := locker.lock()
	if err != nil {
		t.Fatal(err)
	}
	if locked, err := locker.locked(); err != nil {
		t.Fatal(err)
	} else if locked {
		t.Fatal("expected memory lock state to be hidden")
	}
	if err := locker.forceUnlock(); err != nil {
		t.Fatal(err)
	}
	acquired := make(chan lockResult, 1)
	started := make(chan struct{})
	go func() {
		close(started)
		unlock, err := locker.lock()
		acquired <- lockResult{unlock: unlock, err: err}
	}()
	<-started
	select {
	case result := <-acquired:
		if result.err == nil {
			result.unlock()
		}
		t.Fatal("memory force unlock released active work")
	case <-time.After(20 * time.Millisecond):
	}
	unlock()
	select {
	case result := <-acquired:
		if result.err != nil {
			t.Fatal(result.err)
		}
		result.unlock()
	case <-time.After(time.Second):
		t.Fatal("memory lock was not released")
	}
	if locker.isLockFile("") {
		t.Fatal("expected no lock file")
	}
}

func TestMemoryLockerCoordinatesIDs(t *testing.T) {
	memory := engine.NewLocker(nil)
	first := newLoaderLocker(utils.MetaMemory, "one", memory)
	same := newLoaderLocker(utils.MetaMemory, "one", memory)
	different := newLoaderLocker(utils.MetaMemory, "two", memory)

	unlockFirst, err := first.lock()
	if err != nil {
		t.Fatal(err)
	}
	sameAcquired := make(chan lockResult, 1)
	sameStarted := make(chan struct{})
	go func() {
		close(sameStarted)
		unlock, err := same.lock()
		sameAcquired <- lockResult{unlock: unlock, err: err}
	}()
	<-sameStarted

	differentAcquired := make(chan lockResult, 1)
	go func() {
		unlock, err := different.lock()
		differentAcquired <- lockResult{unlock: unlock, err: err}
	}()
	select {
	case result := <-differentAcquired:
		if result.err != nil {
			t.Fatal(result.err)
		}
		result.unlock()
	case <-time.After(time.Second):
		t.Fatal("different loader IDs blocked each other")
	}
	select {
	case result := <-sameAcquired:
		if result.err == nil {
			result.unlock()
		}
		t.Fatal("same loader ID did not wait")
	case <-time.After(20 * time.Millisecond):
	}
	unlockFirst()
	select {
	case result := <-sameAcquired:
		if result.err != nil {
			t.Fatal(result.err)
		}
		result.unlock()
	case <-time.After(time.Second):
		t.Fatal("same loader ID remained blocked")
	}
}

func TestMemoryLockerTimeoutKeepsOwnerBoundRelease(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().LockingTimeout = 100 * time.Millisecond
	memory := engine.NewLocker(cfg)
	first := newLoaderLocker(utils.MetaMemory, "ID", memory)
	second := newLoaderLocker(utils.MetaMemory, "ID", memory)
	third := newLoaderLocker(utils.MetaMemory, "ID", memory)

	unlockFirst, err := first.lock()
	if err != nil {
		t.Fatal(err)
	}
	secondAcquired := make(chan lockResult, 1)
	go func() {
		unlock, err := second.lock()
		secondAcquired <- lockResult{unlock: unlock, err: err}
	}()
	var secondResult lockResult
	select {
	case secondResult = <-secondAcquired:
		if secondResult.err != nil {
			t.Fatal(secondResult.err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out owner did not release the lock")
	}

	unlockFirst()
	thirdAcquired := make(chan lockResult, 1)
	go func() {
		unlock, err := third.lock()
		thirdAcquired <- lockResult{unlock: unlock, err: err}
	}()
	select {
	case result := <-thirdAcquired:
		if result.err == nil {
			result.unlock()
		}
		t.Fatal("old owner released its successor")
	case <-time.After(20 * time.Millisecond):
	}
	secondResult.unlock()
	select {
	case result := <-thirdAcquired:
		if result.err != nil {
			t.Fatal(result.err)
		}
		result.unlock()
	case <-time.After(time.Second):
		t.Fatal("successor did not release the lock")
	}
}
