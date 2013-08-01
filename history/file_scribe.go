/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"sync"
)

type FileScribe struct {
	sync.RWMutex
	filename string
	records  records
}

func NewFileScribe(filename string) (Scribe, error) {
	// looking for git
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("Please install git: " + err.Error())
	}
	s := &FileScribe{filename: filename}
	return s, s.load()
}

func (s *FileScribe) Record(key string, obj interface{}) error {
	s.Lock()
	defer s.Unlock()
	s.records = s.records.SetOrAdd(key, obj)
	s.save()
	return nil
}

func (s *FileScribe) commit() error {
	out, err := exec.Command("git", "commit", "-a", "-m", "'historic commit'").Output()
	if err != nil {
		return errors.New(string(out) + " " + err.Error())
	}
	return nil
}

func (s *FileScribe) load() error {
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

func (s *FileScribe) save() error {
	f, err := os.Create(s.filename)
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
