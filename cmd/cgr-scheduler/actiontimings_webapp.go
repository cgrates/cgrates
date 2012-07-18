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
	"html/template"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	actionTimings, err := storage.GetAllActionTimings()
	if err != nil {
		log.Print("Cannot get action timings:", err)
	}
	if t, err := template.ParseFiles("templates/base.html", "templates/actiontimings.html"); err == nil {
		t.Execute(w, actionTimings)
	} else {
		log.Print("Error rendering status: ", err)
	}
}

func startWebApp() {
	http.Handle("/static/", http.FileServer(http.Dir("")))
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(*httpAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
