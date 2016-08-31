/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package cache2go

import (
	"sync"
	"testing"
)

func TestRemKey(t *testing.T) {
	Set("t11_mm", "test", true, "")
	if t1, ok := Get("t11_mm"); !ok || t1 != "test" {
		t.Error("Error setting cache: ", ok, t1)
	}
	RemKey("t11_mm", true, "")
	if t1, ok := Get("t11_mm"); ok || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestTransaction(t *testing.T) {
	transID := BeginTransaction()
	Set("t11_mm", "test", false, transID)
	if t1, ok := Get("t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t12_mm", "test", false, transID)
	RemKey("t11_mm", false, transID)
	if _, hasTransID := transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	CommitTransaction(transID)
	if t1, ok := Get("t12_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}
}

func TestTransactionRem(t *testing.T) {
	transID := BeginTransaction()
	Set("t21_mm", "test", false, transID)
	Set("t21_nn", "test", false, transID)
	RemPrefixKey("t21_", false, transID)
	if _, hasTransID := transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	CommitTransaction(transID)
	if t1, ok := Get("t21_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t21_nn"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}
}

func TestTransactionRollback(t *testing.T) {
	transID := BeginTransaction()
	Set("t31_mm", "test", false, transID)
	if t1, ok := Get("t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("t32_mm", "test", false, transID)
	if _, hasTransID := transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	RollbackTransaction(transID)
	if t1, ok := Get("t32_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	if _, hasTransID := transactionBuffer[transID]; hasTransID {
		t.Error("Should not longer have transactionID")
	}
}

func TestTransactionRemBefore(t *testing.T) {
	transID := BeginTransaction()
	RemPrefixKey("t41_", false, transID)
	Set("t41_mm", "test", false, transID)
	Set("t41_nn", "test", false, transID)
	CommitTransaction(transID)
	if t1, ok := Get("t41_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("t41_nn"); !ok || t1 != "test" {
		t.Error("Error in transaction cache")
	}
}

func TestRemPrefixKey(t *testing.T) {
	Set("xxx_t1", "test", true, "")
	Set("yyy_t1", "test", true, "")
	RemPrefixKey("xxx_", true, "")
	_, okX := Get("xxx_t1")
	_, okY := Get("yyy_t1")
	if okX || !okY {
		t.Error("Error removing prefix: ", okX, okY)
	}
}

func TestCacheCount(t *testing.T) {
	Set("dst_A1", "1", true, "")
	Set("dst_A2", "2", true, "")
	Set("rpf_A3", "3", true, "")
	Set("dst_A4", "4", true, "")
	Set("dst_A5", "5", true, "")
	if CountEntries("dst_") != 4 {
		t.Error("Error countiong entries: ", CountEntries("dst_"))
	}
}

// Try concurrent read/write of the cache
func TestCacheConcurrent(t *testing.T) {
	s := &struct{ Prefix string }{Prefix: "+49"}
	Set("dst_DE", s, true, "")
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			Get("dst_DE")
			wg.Done()
		}()
	}
	s.Prefix = "+491"
	wg.Wait()
}
