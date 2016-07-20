package engine

import "testing"

func TestRemKey(t *testing.T) {
	CacheSet("t11_mm", "test")
	if t1, ok := CacheGet("t11_mm"); !ok || t1 != "test" {
		t.Error("Error setting cache: ", err, t1)
	}
	CacheRemKey("t11_mm")
	if t1, ok := CacheGet("t11_mm"); ok || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestTransaction(t *testing.T) {
	CacheBeginTransaction()
	CacheSet("t11_mm", "test")
	if t1, ok := CacheGet("t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	CacheSet("t12_mm", "test")
	CacheRemKey("t11_mm")
	CacheCommitTransaction()
	if t1, ok := CacheGet("t12_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := CacheGet("t11_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRem(t *testing.T) {
	CacheBeginTransaction()
	CacheSet("t21_mm", "test")
	CacheSet("t21_nn", "test")
	CacheRemPrefixKey("t21_")
	CacheCommitTransaction()
	if t1, ok := CacheGet("t21_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := CacheGet("t21_nn"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRollback(t *testing.T) {
	CacheBeginTransaction()
	CacheSet("t31_mm", "test")
	if t1, ok := CacheGet("t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	CacheSet("t32_mm", "test")
	CacheRollbackTransaction()
	if t1, ok := CacheGet("t32_mm"); ok || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := CacheGet("t31_mm"); ok || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRemBefore(t *testing.T) {
	CacheBeginTransaction()
	CacheRemPrefixKey("t41_")
	CacheSet("t41_mm", "test")
	CacheSet("t41_nn", "test")
	CacheCommitTransaction()
	if t1, ok := CacheGet("t41_mm"); !ok || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, ok := CacheGet("t41_nn"); !ok || t1 != "test" {
		t.Error("Error in transaction cache")
	}
}

func TestCacheRemPrefixKey(t *testing.T) {
	CacheSet("xxx_t1", "test")
	CacheSet("yyy_t1", "test")
	CacheRemPrefixKey("xxx_")
	_, okX := CacheGet("xxx_t1")
	_, okY := CacheGet("yyy_t1")
	if okX || !okY {
		t.Error("Error removing prefix: ", okX, okY)
	}
}

func TestCachePush(t *testing.T) {
	CachePush("ccc_t1", "1")
	CachePush("ccc_t1", "2")
	v, ok := CacheGet("ccc_t1")
	if !ok || len(v.(map[string]struct{})) != 2 {
		t.Error("Error in cache push: ", v)
	}
}

func TestCachePop(t *testing.T) {
	CachePush("ccc_t1", "1")
	CachePush("ccc_t1", "2")
	v, ok := CacheGet("ccc_t1")
	if !ok || len(v.(map[string]struct{})) != 2 {
		t.Error("Error in cache push: ", v)
	}
	CachePop("ccc_t1", "1")
	v, ok = CacheGet("ccc_t1")
	if !ok || len(v.(map[string]struct{})) != 1 {
		t.Error("Error in cache pop: ", v)
	}
}

/*func TestCount(t *testing.T) {
	CacheSet("dst_A1", "1")
	CacheSet("dst_A2", "2")
	CacheSet("rpf_A3", "3")
	CacheSet("dst_A4", "4")
	CacheSet("dst_A5", "5")
	if CacheCountEntries("dst_") != 4 {
		t.Error("Error countiong entries: ", CacheCountEntries("dst_"))
	}
}*/
