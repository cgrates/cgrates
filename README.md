# LRU/TTL Cache system written in [Go](http://golang.org/ "Go Website")

Some of it's properties:

 * Both LRU or TTL enforcements are optional and independent of each other.
 * Thread safe through the use of unique lock, making each operation atomic.
 * TTL refresh on get/set optional through the use of static setting.
 * Item groups for groupped remove
 * Transactional if TransCache is used
 * Multiple instances if TransCache is used

## Installation ##

`go get github.com/cgrates/ltcache`

## Support ##
Join [CGRateS](http://www.cgrates.org/ "CGRateS Website") on Google Groups [here](https://groups.google.com/forum/#!forum/cgrates "CGRateS on Google Groups").

## License ##
ltcache.go is released under the [MIT License](http://www.opensource.org/licenses/mit-license.php "MIT License").
Copyright (C) ITsysCOM GmbH. All Rights Reserved.

## Sample usage code ##
```
package main

func main() {
	cache := NewCache(3, time.Duration(10*time.Millisecond), false, 
		func(k key, v interface{}) { fmt.Printf("Evicted key: %v, value: %v", k, v)})
	cache.Set("key1": "val1")
	cache.Get("key1")
	cache.Remove("key1")
	cache.Set(1, 1)
	cache.Remove(1)
	tc := NewTransCache(map[string]*CacheConfig{
		"dst_": &CacheConfig{MaxItems: -1},
		"rpf_": &CacheConfig{MaxItems: -1}})
	transID := tc.BeginTransaction()
	tc.Set("aaa_", "t31", "test", nil, false, transID)
	tc.Set("bbb_", "t32", "test", nil, false, transID)
	tc.CommitTransaction(transID)
	transID2 := tc.BeginTransaction()
	tc.Set("ccc_", "t31", "test", nil, false, transID2)
	tc.RollbackTransaction(transID2)
}
```

[![Build Status](https://secure.travis-ci.org/cgrates/ltcache.png)](http://travis-ci.org/cgrates/ltcache)
