/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package sessionmanager

import (
	"fmt"
	"log"
	"regexp"
)

type Event struct {
	Fields map[string]string
}

var (
	eventBodyRE = regexp.MustCompile(`"(.*?)":\s+"(.*?)"`)
)

//eventBodyRE *regexp.Regexp

func NewEvent(body string) (ev *Event) {
	ev = &Event{Fields: make(map[string]string)}
	for _, fields := range eventBodyRE.FindAllStringSubmatch(body, -1) {
		if len(fields) == 3 {
			ev.Fields[fields[1]] = fields[2]
		} else {
			log.Printf("malformed event field: %v", fields)
		}
	}
	return
}

func (ev *Event) String() (result string) {
	for k, v := range ev.Fields {
		result += fmt.Sprintf("%s = %s\n", k, v)
	}
	result += "=============================================================="
	return
}
