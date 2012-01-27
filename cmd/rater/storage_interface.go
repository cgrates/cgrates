package main

type StorageGetter interface {
	Close()
	Get(key string) (string, error)
}
