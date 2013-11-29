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
	"time"
)

const (
	DESTINATIONS_FILE    = "destinations.json"
	RATING_PLANS_FILE    = "rating_plans.json"
	RATING_PROFILES_FILE = "rating_profiles.json"
)

type FileScribe struct {
	sync.Mutex
	fileRoot       string
	gitCommand     string
	destinations   records
	ratingPlans    records
	ratingProfiles records
	loopChecker    chan int
	waitingFile    string
}

func NewFileScribe(fileRoot string) (*FileScribe, error) {
	// looking for git
	gitCommand, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("Please install git: " + err.Error())
	}
	s := &FileScribe{fileRoot: fileRoot, gitCommand: gitCommand}
	s.loopChecker = make(chan int)
	s.gitInit()
	if err := s.load(DESTINATIONS_FILE); err != nil {
		return nil, err
	}
	if err := s.load(RATING_PROFILES_FILE); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileScribe) Record(rec *Record, out *int) error {
	s.Lock()
	defer s.Unlock()
	var fileToSave string
	switch {
	case strings.HasPrefix(rec.Key, DESTINATION_PREFIX):
		s.destinations = s.destinations.SetOrAdd(&Record{rec.Key[len(DESTINATION_PREFIX):], rec.Object})
		fileToSave = DESTINATIONS_FILE
	case strings.HasPrefix(rec.Key, RATING_PLAN_PREFIX):
		s.ratingPlans = s.ratingPlans.SetOrAdd(&Record{rec.Key[len(RATING_PLAN_PREFIX):], rec.Object})
		fileToSave = RATING_PLANS_FILE
	case strings.HasPrefix(rec.Key, RATING_PROFILE_PREFIX):
		s.ratingProfiles = s.ratingProfiles.SetOrAdd(&Record{rec.Key[len(RATING_PROFILE_PREFIX):], rec.Object})
		fileToSave = RATING_PROFILES_FILE
	}

	// flood protection for save method (do not save on every loop iteration)
	if s.waitingFile == fileToSave {
		s.loopChecker <- 1
	}
	s.waitingFile = fileToSave
	go func() {
		t := time.NewTicker(1 * time.Second)
		select {
		case <-s.loopChecker:
			// cancel saving
		case <-t.C:
			if fileToSave != "" {
				s.save(fileToSave)
			}
			t.Stop()
			s.waitingFile = ""
		}
	}()
	// no protection variant
	/*if fileToSave != "" {
		s.save(fileToSave)
	}*/
	*out = 0
	return nil
}

func (s *FileScribe) gitInit() error {
	s.Lock()
	defer s.Unlock()
	if _, err := os.Stat(s.fileRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(s.fileRoot, os.ModeDir|0755); err != nil {
			return errors.New("<History> Error creating history folder: " + err.Error())
		}
	}
	if _, err := os.Stat(filepath.Join(s.fileRoot, ".git")); os.IsNotExist(err) {
		cmd := exec.Command(s.gitCommand, "init")
		cmd.Dir = s.fileRoot
		if out, err := cmd.Output(); err != nil {
			return errors.New(string(out) + " " + err.Error())
		}
		if f, err := os.Create(filepath.Join(s.fileRoot, DESTINATIONS_FILE)); err != nil {
			return errors.New("<History> Error writing destinations file: " + err.Error())
		} else {
			f.Close()
		}
		if f, err := os.Create(filepath.Join(s.fileRoot, RATING_PLANS_FILE)); err != nil {
			return errors.New("<History> Error writing rating plans file: " + err.Error())
		} else {
			f.Close()
		}
		if f, err := os.Create(filepath.Join(s.fileRoot, RATING_PROFILES_FILE)); err != nil {
			return errors.New("<History> Error writing rating profiles file: " + err.Error())
		} else {
			f.Close()
		}
		cmd = exec.Command(s.gitCommand, "add", ".")
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
	s.Lock()
	defer s.Unlock()
	f, err := os.Open(filepath.Join(s.fileRoot, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)

	switch filename {
	case DESTINATIONS_FILE:
		if err := d.Decode(&s.destinations); err != nil && err != io.EOF {
			return errors.New("<History> Error loading destinations: " + err.Error())
		}
		s.destinations.Sort()
	case RATING_PLANS_FILE:
		if err := d.Decode(&s.ratingPlans); err != nil && err != io.EOF {
			return errors.New("<History> Error loading rating plans: " + err.Error())
		}
		s.ratingPlans.Sort()
	case RATING_PROFILES_FILE:
		if err := d.Decode(&s.ratingProfiles); err != nil && err != io.EOF {
			return errors.New("<History> Error loading rating profiles: " + err.Error())
		}
		s.ratingProfiles.Sort()
	}
	return nil
}

func (s *FileScribe) save(filename string) error {
	s.Lock()
	defer s.Unlock()
	f, err := os.Create(filepath.Join(s.fileRoot, filename))
	if err != nil {
		return err
	}

	b := bufio.NewWriter(f)
	switch filename {
	case DESTINATIONS_FILE:
		if err := s.format(b, s.destinations); err != nil {
			return err
		}
	case RATING_PLANS_FILE:
		if err := s.format(b, s.ratingPlans); err != nil {
			return err
		}
	case RATING_PROFILES_FILE:
		if err := s.format(b, s.ratingProfiles); err != nil {
			return err
		}
	}
	b.Flush()
	f.Close()
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
			b.Write([]byte(",\n"))
		}
	}
	b.Write([]byte("]"))
	return nil
}
