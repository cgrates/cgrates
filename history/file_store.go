package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"sync"
)

type FileStore struct {
	sync.RWMutex
	filename string
	records  records
}

func NewFileStore(filename string) (*FileStore, error) {
	// looking for git
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("Please install git: " + err.Error())
	}
	s := &FileStore{filename: filename}
	return s, s.load()
}

func (s *FileStore) Record(key string, obj interface{}) error {
	s.Lock()
	defer s.Unlock()
	s.records = s.records.SetOrAdd(key, obj)
	return nil
}

func (s *FileStore) commit() error {
	out, err := exec.Command("git", "commit", "-a", "-m", "'historic commit'").Output()
	if err != nil {
		return errors.New(string(out) + " " + err.Error())
	}
	return nil
}

func (s *FileStore) load() error {
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

func (s *FileStore) save(filename string) error {
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
