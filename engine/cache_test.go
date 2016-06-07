package engine

import "testing"

func TestRemKey(t *testing.T) {
	CacheSet("t11_mm", "test")
	if t1, err := CacheGet("t11_mm"); err != nil || t1 != "test" {
		t.Error("Error setting cache: ", err, t1)
	}
	CacheRemKey("t11_mm")
	if t1, err := CacheGet("t11_mm"); err == nil || t1 == "test" {
		t.Error("Error removing cached key")
	}
}

func TestTransaction(t *testing.T) {
	CacheBeginTransaction()
	CacheSet("t11_mm", "test")
	if t1, err := CacheGet("t11_mm"); err == nil || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	CacheSet("t12_mm", "test")
	CacheRemKey("t11_mm")
	CacheCommitTransaction()
	if t1, err := CacheGet("t12_mm"); err != nil || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, err := CacheGet("t11_mm"); err == nil || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRem(t *testing.T) {
	CacheBeginTransaction()
	CacheSet("t21_mm", "test")
	CacheSet("t21_nn", "test")
	CacheRemPrefixKey("t21_")
	CacheCommitTransaction()
	if t1, err := CacheGet("t21_mm"); err == nil || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, err := CacheGet("t21_nn"); err == nil || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRollback(t *testing.T) {
	CacheBeginTransaction()
	CacheSet("t31_mm", "test")
	if t1, err := CacheGet("t31_mm"); err == nil || t1 == "test" {
		t.Error("Error in transaction cache")
	}
	CacheSet("t32_mm", "test")
	CacheRollbackTransaction()
	if t1, err := CacheGet("t32_mm"); err == nil || t1 == "test" {
		t.Error("Error commiting transaction")
	}
	if t1, err := CacheGet("t31_mm"); err == nil || t1 == "test" {
		t.Error("Error in transaction cache")
	}
}

func TestTransactionRemBefore(t *testing.T) {
	CacheBeginTransaction()
	CacheRemPrefixKey("t41_")
	CacheSet("t41_mm", "test")
	CacheSet("t41_nn", "test")
	CacheCommitTransaction()
	if t1, err := CacheGet("t41_mm"); err != nil || t1 != "test" {
		t.Error("Error commiting transaction")
	}
	if t1, err := CacheGet("t41_nn"); err != nil || t1 != "test" {
		t.Error("Error in transaction cache")
	}
}

func TestCacheRemPrefixKey(t *testing.T) {
	CacheSet("xxx_t1", "test")
	CacheSet("yyy_t1", "test")
	CacheRemPrefixKey("xxx_")
	_, errX := CacheGet("xxx_t1")
	_, errY := CacheGet("yyy_t1")
	if errX == nil || errY != nil {
		t.Error("Error removing prefix: ", errX, errY)
	}
}

func TestCachePush(t *testing.T) {
	CachePush("ccc_t1", "1")
	CachePush("ccc_t1", "2")
	v, err := CacheGet("ccc_t1")
	if err != nil || len(v.(map[string]struct{})) != 2 {
		t.Error("Error in cache push: ", v)
	}
}

func TestCachePop(t *testing.T) {
	CachePush("ccc_t1", "1")
	CachePush("ccc_t1", "2")
	v, err := CacheGet("ccc_t1")
	if err != nil || len(v.(map[string]struct{})) != 2 {
		t.Error("Error in cache push: ", v)
	}
	CachePop("ccc_t1", "1")
	v, err = CacheGet("ccc_t1")
	if err != nil || len(v.(map[string]struct{})) != 1 {
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
