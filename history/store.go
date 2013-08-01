package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/rpc"
	"os"
	"os/exec"
	"sort"
	"sync"
)

type Store interface {
	Record(key string, obj interface{}) error
}

type HistoryStore struct {
	sync.RWMutex
	filename string
	records  records
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

func NewHistoryStore(filename string) (*HistoryStore, error) {
	// looking for git
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("Please install git: " + err.Error())
	}
	s := &HistoryStore{filename: filename}
	return s, s.load()
}

func (s *HistoryStore) Record(key string, obj interface{}) error {
	s.Lock()
	defer s.Unlock()
	s.records = s.records.SetOrAdd(key, obj)
	return nil
}

func (s *HistoryStore) commit() error {
	out, err := exec.Command("git", "commit", "-a", "-m", "'historic commit'").Output()
	if err != nil {
		return errors.New(string(out) + " " + err.Error())
	}
	return nil
}

func (s *HistoryStore) load() error {
	f, err := os.Open(s.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)

	if err := d.Decode(&s.records); err != nil {
		return err
	}
	s.records.Sort()
	return nil
}

func (s *HistoryStore) save(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	b := bufio.NewWriter(f)
	e := json.NewEncoder(b)
	defer f.Close()
	defer b.Flush()
	s.records.Sort()
	if err := e.Encode(s.records); err != nil {
		return err
	}
	return s.commit()
}

type ProxyStore struct {
	client *rpc.Client
}

func NewProxyStore(addr string) (*ProxyStore, error) {
	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &ProxyStore{client: client}, nil
}

func (ps *ProxyStore) Record(key string, obj interface{}) error {
	if err := ps.client.Call("Store.Record", key, obj); err != nil {
		return err
	}
	return nil
}
