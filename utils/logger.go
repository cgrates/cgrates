/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2014 ITsysCOM

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
	"log"
)

type LoggerInterface interface {
	Alert(m string) error
	Close() error
	Crit(m string) error
	Debug(m string) error
	Emerg(m string) error
	Err(m string) error
	Info(m string) error
	Notice(m string) error
	Warning(m string) error
}

// Logs to standard output
type StdLogger struct{}

func (sl *StdLogger) Alert(m string) (err error) {
	log.Print("[ALERT]" + m)
	return
}
func (sl *StdLogger) Close() (err error) {
	return
}
func (sl *StdLogger) Crit(m string) (err error) {
	log.Print("[CRITICAL]" + m)
	return
}
func (sl *StdLogger) Debug(m string) (err error) {
	log.Print("[DEBUG]" + m)
	return
}
func (sl *StdLogger) Emerg(m string) (err error) {
	log.Print("[EMERGENCY]" + m)
	return
}
func (sl *StdLogger) Err(m string) (err error) {
	log.Print("[ERROR]" + m)
	return
}
func (sl *StdLogger) Info(m string) (err error) {
	log.Print("[INFO]" + m)
	return
}
func (sl *StdLogger) Notice(m string) (err error) {
	log.Print("[NOTICE]" + m)
	return
}
func (sl *StdLogger) Warning(m string) (err error) {
	log.Print("[WARNING]" + m)
	return
}
