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
	"flag"
	// "github.com/rif/cgrates/timespans"
	"log"
	"os"
	"encoding/csv"
)

var (
	separator   = flag.String("separator", ";", "Default field separator")
	redisserver = flag.String("redisserver", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb     = flag.Int("rdb", 10, "redis database number (10)")
	redispass   = flag.String("pass", "", "redis database password")
	months      = flag.String("month", "Months.csv", "Months file")
)

func main() {
	flag.Parse()
	fp, err := os.Open(*months)
	if err != nil {
		log.Printf("Could not open months file: %v", err)
	}
	csv := csv.NewReader(fp)
	csv.Comma = rune(*separator)
	for record, err := csv.Read(); err == nil; record, err = csv.Read() {
		if record[0] == "Tag" {
			// skip header line
			continue
		}
		log.Print(record)
	}
}
