/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package ers

import (
	"fmt"
	"path/filepath"

	"github.com/cgrates/cgrates/config"
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
	defer watcher.Close()
	if err = watcher.Add(dirPath); err != nil {
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> monitoring <%s> for file moves.", sysID, dirPath))
	for {
		select {
		case <-stopWatching:
			utils.Logger.Info(fmt.Sprintf("<%s> stop watching path <%s>", sysID, dirPath))
			return
		case ev := <-watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create {
				go func() { //Enable async processing here
					if err = f(filepath.Dir(ev.Name), filepath.Base(ev.Name)); err != nil {
						utils.Logger.Warning(fmt.Sprintf("<%s> processing path <%s>, error: <%s>",
							sysID, ev.Name, err.Error()))
					}
				}()
			}
		case err = <-watcher.Errors:
			return
		}
	}
}

// erEvent is passed from reader to ERs
type erEvent struct {
	cgrEvent *utils.CGREvent
	rdrCfg   *config.EventReaderCfg
}
