// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"fmt"
	"path/filepath"

	"github.com/cgrates/cgrates/utils"
	"github.com/fsnotify/fsnotify"
)

// watchDir sets up a watcher via inotify to be triggered on new files
// sysID is the subsystem ID, f will be triggered on match
func watchDir(dirPath string, f func(itmPath, itmID string) error,
	sysID string, stopWatching chan struct{}) (err error) {
	var watcher *fsnotify.Watcher
	if watcher, err = fsnotify.NewWatcher(); err != nil {
		return
	}
	if err = watcher.Add(dirPath); err != nil {
		watcher.Close()
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> monitoring <%s> for file moves.", sysID, dirPath))
	go func() { // read async
		defer watcher.Close()
		for {
			select {
			case <-stopWatching:
				utils.Logger.Info(fmt.Sprintf("<%s> stop watching path <%s>", sysID, dirPath))
				return
			case ev := <-watcher.Events:
				if ev.Op&fsnotify.Create == fsnotify.Create {
					go func() { //Enable async processing here so we can simultaneously process files
						if err := f(filepath.Dir(ev.Name), filepath.Base(ev.Name)); err != nil {
							utils.Logger.Warning(fmt.Sprintf("<%s> processing path <%s>, error: <%s>",
								sysID, ev.Name, err.Error()))
						}
					}()
				}
			case err = <-watcher.Errors:
				utils.Logger.Err(
					fmt.Sprintf("<%s> watching path <%s>, error: <%s>, exiting!",
						sysID, dirPath, err.Error()))
				return
			}
		}
	}()
	return
}
