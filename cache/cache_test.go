/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
package cache

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
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
	Set("mmm_t11", "test", false, transID)
	if t1, ok := Get("mmm_t11"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("mmm_t12", "test", false, transID)
	RemKey("mmm_t11", false, transID)
	if _, hasTransID := transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	CommitTransaction(transID)
	if t1, ok := Get("mmm_t12"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("mmm_t11"); ok || t1 == "test" {
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
	RemPrefixKey(utils.ANY, false, transID)
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
	Set("aaa_t31", "test", false, transID)
	if t1, ok := Get("aaa_t31"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	Set("aaa_t32", "test", false, transID)
	if _, hasTransID := transactionBuffer[transID]; !hasTransID {
		t.Error("Does not have transactionID")
	}
	RollbackTransaction(transID)
	if t1, ok := Get("aaa_t32"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := Get("aaa_t31"); ok || t1 == "test" {
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
	RemPrefixKey(utils.ANY, true, "")
	_, okX := Get("xxx_t1")
	_, okY := Get("yyy_t1")
	if okX || okY {
		t.Error("Error removing prefix: ", okX, okY)
	}
}

func TestCacheCount(t *testing.T) {
	Set("dst_A1", "1", true, "")
	Set("dst_A2", "2", true, "")
	Set("rpf_A3", "3", true, "")
	Set("dst_A4", "4", true, "")
	Set("dst_A5", "5", true, "")
	if cnt := CountEntries(utils.DESTINATION_PREFIX); cnt != 4 {
		t.Error("Error counting entries: ", cnt)
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

// BenchmarkSet 	 5000000	       313 ns/op
func BenchmarkSet(b *testing.B) {
	cacheItems := [][]string{
		[]string{"aaa_1", "1"},
		[]string{"aaa_2", "1"},
		[]string{"aaa_3", "1"},
		[]string{"aaa_4", "1"},
		[]string{"aaa_5", "1"},
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(cacheItems)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		ci := cacheItems[rand.Intn(max-min)+min]
		Set(ci[0], ci[1], false, "")
	}
}

// BenchmarkGet 	10000000	       131 ns/op
func BenchmarkGet(b *testing.B) {
	cacheItems := [][]string{
		[]string{"aaa_1", "1"},
		[]string{"aaa_2", "1"},
		[]string{"aaa_3", "1"},
		[]string{"aaa_4", "1"},
		[]string{"aaa_5", "1"},
	}
	for _, ci := range cacheItems {
		Set(ci[0], ci[1], false, "")
	}
	rand.Seed(time.Now().UTC().UnixNano())
	min, max := 0, len(cacheItems)-1 // so we can have random index
	for n := 0; n < b.N; n++ {
		cache.Get(cacheItems[rand.Intn(max-min)+min][0])
	}
}
