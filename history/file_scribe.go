/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type FileScribe struct {
	mu          sync.Mutex
	fileRoot    string
	gitCommand  string
	loopChecker chan int
	waitingFile string
	savePeriod  time.Duration
}

func NewFileScribe(fileRoot string, saveInterval time.Duration) (*FileScribe, error) {
	// looking for git
	gitCommand, err := exec.LookPath("git")
	if err != nil {
		return nil, errors.New("Please install git: " + err.Error())
	}
	s := &FileScribe{fileRoot: fileRoot, gitCommand: gitCommand, savePeriod: saveInterval}
	s.loopChecker = make(chan int)
	files := []string{DESTINATIONS_FN, RATING_PLANS_FN, RATING_PROFILES_FN}
	s.gitInit(files)

	for _, fn := range files {
		if err := s.load(fn); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *FileScribe) Record(rec Record, out *int) error {
	s.mu.Lock()
	fileToSave := rec.Filename
	recordsMap[fileToSave] = recordsMap[fileToSave].SetOrAdd(&rec)

	// flood protection for save method (do not save on every loop iteration)
	if s.waitingFile == fileToSave {
		s.loopChecker <- 1
	}
	s.waitingFile = fileToSave
	defer s.mu.Unlock()
	go func() {
		t := time.NewTicker(s.savePeriod)
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

func (s *FileScribe) gitInit(files []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
		log.Print("CREATING FILES")
		for _, fn := range files {
			log.Print("FILE: ", fn)
			if f, err := os.Create(filepath.Join(s.fileRoot, fn)); err != nil {
				return fmt.Errorf("<History> Error writing %s file: %s", fn, err.Error())
			} else {
				f.Close()
			}
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
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(filepath.Join(s.fileRoot, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(f)

	records := recordsMap[filename]
	if err := d.Decode(&records); err != nil && err != io.EOF {
		return fmt.Errorf("<History> Error loading %s: %s", filename, err.Error())
	}
	records.Sort()
	return nil
}

func (s *FileScribe) save(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Create(filepath.Join(s.fileRoot, filename))
	if err != nil {
		return err
	}

	b := bufio.NewWriter(f)
	records := recordsMap[filename]
	if err := format(b, records); err != nil {
		return err
	}
	b.Flush()
	f.Close()
	return s.gitCommit()
}
