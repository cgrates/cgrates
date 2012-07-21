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
	"bufio"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	_ "github.com/bmizerany/pq"
	"github.com/cgrates/cgrates/timespans"
	"log"
	"net/rpc/jsonrpc"
	"os"
	"time"
)

/*var (
	cdrFile    = flag.String("freeswitchcdr", "Master.csv", "Freeswitch Master CSV CDR file.")
	resultFile = flag.String("resultfile", "out.csv", "Generated file containing CDR and price info.")
	host       = flag.String("host", "localhost", "The host to connect to. Values that start with / are for UNIX domain sockets.")
	port       = flag.String("port", "5432", "The port to bind to.")
	dbName     = flag.String("dbname", "cgrates", "The name of the database to connect to.")
	user       = flag.String("user", "", "The user to sign in as.")
	password   = flag.String("password", "", "The user's password.")
)*/

func readDbRecord(db *sql.DB, searchedUUID string) (cc *timespans.CallCost, timespansText string, err error) {
	row := db.QueryRow(fmt.Sprintf("SELECT * FROM callcosts WHERE uuid='%s'", searchedUUID))
	var uuid string
	cc = &timespans.CallCost{}
	err = row.Scan(&uuid, &cc.Direction, &cc.Tenant, &cc.TOR, &cc.Subject, &cc.Destination, &cc.Cost, &cc.ConnectFee, &timespansText)
	return
}

func maina() {
	flag.Parse()
	useDB := true
	file, err := os.Open(mediator_cdr_file)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", mediator_host, mediator_port, mediator_db, mediator_user, mediator_password))
	defer db.Close()
	if err != nil {
		log.Printf("failed to open the database: %v", err)
		useDB = false
	}
	csvReader := csv.NewReader(bufio.NewReader(file))
	client, err := jsonrpc.Dial("tcp", "localhost:2001")
	useRPC := true
	if err != nil {
		log.Printf("Could not connect to rater server: %v!", err)
		useRPC = false
	}
	for record, err := csvReader.Read(); err == nil; record, err = csvReader.Read() {
		uuid := record[10]
		t, _ := time.Parse("2012-05-21 17:48:20", record[5])
		fmt.Println(t)
		if useDB {
			cc, timespansText, err := readDbRecord(db, uuid)
			if err != nil && useRPC {
				// try getting the price from the rater

				tenant := record[0]
				subject := record[1]
				dest := record[2]
				t1, _ := time.Parse("2012-05-21 17:48:20", record[5])
				t2, _ := time.Parse("2012-05-21 17:48:20", record[6])
				cd := timespans.CallDescriptor{
					Direction:   "OUT",
					Tenant:      tenant,
					TOR:         "0",
					Subject:     subject,
					Destination: dest,
					TimeStart:   t1,
					TimeEnd:     t2}
				client.Call("Responder.GetCost", cd, cc)
			}
			_ = timespansText
			//log.Print(cc, timespansText)
		}
	}
}
