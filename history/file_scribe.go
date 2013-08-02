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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	DESTINATIONS_FILE    = "destinations.json"
	RATING_PROFILES_FILE = "rating_profiles.json"
)

type FileScribe struct {
	sync.RWMutex
	fileRoot       string
	gitCommand     string
	destinations   records
	ratingProfiles records
}

func NewFileScribe(fileRoot string) (Scribe, error) {
	// looking for git
	gitCommand, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("Please install git: " + err.Error())
	}
	s := &FileScribe{fileRoot: fileRoot, gitCommand: gitCommand}
	s.gitInit()
	if err := s.load(DESTINATIONS_FILE); err != nil {
		return nil, err
	}
	if err := s.load(RATING_PROFILES_FILE); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileScribe) Record(key string, obj interface{}) error {
	s.Lock()
	defer s.Unlock()
	switch {
	case strings.HasPrefix(key, DESTINATION_PREFIX):
		s.destinations = s.destinations.SetOrAdd(key, obj)
		s.save(DESTINATIONS_FILE)
	case strings.HasPrefix(key, RATING_PROFILE_PREFIX):
		s.ratingProfiles = s.ratingProfiles.SetOrAdd(key, obj)
		s.save(RATING_PROFILES_FILE)
	}
	return nil
}

func (s *FileScribe) gitInit() error {
	if _, err := os.Stat(filepath.Join(s.fileRoot, ".git")); os.IsNotExist(err) {
		cmd := exec.Command(s.gitCommand, "init")
		cmd.Dir = s.fileRoot
		if out, err := cmd.Output(); err != nil {
			return errors.New(string(out) + " " + err.Error())
		}
		if f, err := os.Create(filepath.Join(s.fileRoot, DESTINATIONS_FILE)); err != nil {
			return err
		} else {
			f.Close()
		}
		if f, err := os.Create(filepath.Join(s.fileRoot, RATING_PROFILES_FILE)); err != nil {
			return err
		} else {
			f.Close()
		}
		cmd = exec.Command(s.gitCommand, "add")
		cmd.Dir = s.fileRoot
		if out, err := cmd.Output(); err != nil {
			return errors.New(string(out) + " " + err.Error())
		}
	}
	return nil
}

func (s *FileScribe) gitCommit() error {
	cmd := exec.Command(s.gitCommand, "commit", "-a", "-m", "'historic commit'")
	cmd.Dir = s.fileRoot
	if out, err := cmd.Output(); err != nil {
		return errors.New(string(out) + " " + err.Error())
	}
	return nil
}

func (s *FileScribe) load(filename string) error {
	f, err := os.Open(filepath.Join(s.fileRoot, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)

	switch {
	case filename == DESTINATIONS_FILE:
		if err := d.Decode(&s.destinations); err != nil {
			return err
		}
		s.destinations.Sort()
	case filename == RATING_PROFILES_FILE:
		if err := d.Decode(&s.ratingProfiles); err != nil {
			return err
		}
		s.ratingProfiles.Sort()
	}
	return nil
}

func (s *FileScribe) save(filename string) error {
	f, err := os.Create(filepath.Join(s.fileRoot, filename))
	if err != nil {
		return err
	}

	b := bufio.NewWriter(f)
	defer b.Flush()
	switch {
	case filename == DESTINATIONS_FILE:
		if err := s.format(b, s.destinations); err != nil {
			return err
		}
	case filename == RATING_PROFILES_FILE:
		if err := s.format(b, s.ratingProfiles); err != nil {
			return err
		}
	}

	return s.gitCommit()
}

func (s *FileScribe) format(b io.Writer, recs records) error {
	recs.Sort()
	b.Write([]byte("["))
	for i, r := range recs {
		src, err := json.Marshal(r)
		if err != nil {
			return err
		}
		b.Write(src)
		if i < len(recs)-1 {
			b.Write([]byte("\n"))
		}
	}
	b.Write([]byte("]"))
	return nil
}
