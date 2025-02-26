/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
package utils

import (
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// WatchDir sets up a watcher via inotify to be triggered on new files
// sysID is the subsystem ID, f will be triggered on match
func WatchDir(dirPath string, f func(itmID string) error, sysID string, stopWatching chan struct{}) (err error) {
	return watchDir(dirPath, f, sysID, stopWatching, fsnotify.NewWatcher)
}

func watchDir(dirPath string, f func(itmID string) error, sysID string,
	stopWatching chan struct{}, newWatcher func() (*fsnotify.Watcher, error)) (err error) {
	var watcher *fsnotify.Watcher
	if watcher, err = newWatcher(); err != nil {
		return
	}
	if err = watcher.Add(dirPath); err != nil {
		watcher.Close()
		return
	}
	Logger.Info(fmt.Sprintf("<%s> monitoring <%s> for file moves.", sysID, dirPath))
	go watch(dirPath, sysID, f, watcher, stopWatching) // read async
	return
}

// watch monitors a directory for file creation events and processes them asynchronously.
func watch(dirPath, sysID string, f func(itmID string) error,
	watcher *fsnotify.Watcher, stopWatching chan struct{}) error {
	defer watcher.Close()
	for {
		select {
		case <-stopWatching:
			Logger.Info(fmt.Sprintf("<%s> stop watching path <%s>", sysID, dirPath))
			return nil
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create {
				// Process files asynchronously to prevent blocking the watcher.
				go func() {
					if err := f(filepath.Base(ev.Name)); err != nil {
						Logger.Warning(fmt.Sprintf("<%s> processing path <%s>, error: <%s>",
							sysID, ev.Name, err.Error()))
					}
				}()
			}
		case err := <-watcher.Errors:
			Logger.Err(fmt.Sprintf(
				"<%s> watching path <%s>, error: <%s>, exiting!",
				sysID, dirPath, err.Error()))
			return err
		}
	}
}
