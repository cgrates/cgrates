package history

import (
	"sort"
)

type Store interface {
	Record(key string, obj interface{}) error
}

type record struct {
	Key    string
	Object interface{}
}

type records []*record

func (rs records) Len() int {
	return len(rs)
}

func (rs records) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs records) Less(i, j int) bool {
	return rs[i].Key < rs[j].Key
}

func (rs records) Sort() {
	sort.Sort(rs)
}

func (rs records) SetOrAdd(key string, obj interface{}) records {
	found := false
	for _, r := range rs {
		if r.Key == key {
			found = true
			r.Object = obj
			return rs
		}
	}
	if !found {
		rs = append(rs, &record{key, obj})
	}
	return rs
}
