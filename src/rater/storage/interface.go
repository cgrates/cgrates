package main

type StorageGetter interface {
	Open(string) error
	Close()
	Get(key string) (string, error)
}
