# LRU/TTL Cache system written in [Go](http://golang.org/ "Go Website")

Some of it's properties:

 * Both LRU or TTL enforcements are optional and independent of each other.
 * Thread safe through the use of unique lock, making each operation atomic.
 * TTL refresh on get/set optional through the use of static setting.

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
	cache := New(3, time.Duration(10*time.Millisecond), false, 
		func(k key, v interface{}) { fmt.Printf("Evicted key: %v, value: %v", k, v)})
	cache.Set("key1": "val1")
	cache.Get("key1")
	cache.Rem("key1")
	cache.Set(1, 1)
	cache.Rem(1)
}
```

[![Build Status](https://secure.travis-ci.org/cgrates/ltcache.png)](http://travis-ci.org/cgrates/ltcache)
