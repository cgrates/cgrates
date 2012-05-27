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
package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"runtime"
)

/*
Handler for the statistics web client
*/
func statusHandler(w http.ResponseWriter, r *http.Request) {
	if t, err := template.ParseFiles("templates/status.html"); err == nil {
		t.Execute(w, bal.GetClientAddresses())
	} else {
		log.Print("Error rendering status: ", err)
	}
}

/*
Ajax Handler for the connected raters
*/
func ratersHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.Encode(bal.GetClientAddresses())
}

/*
Ajax Handler for current used memory value
*/
func memoryHandler(w http.ResponseWriter, r *http.Request) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	enc := json.NewEncoder(w)
	enc.Encode([]uint64{memstats.HeapAlloc / 1024, memstats.Sys / 1024})
}
