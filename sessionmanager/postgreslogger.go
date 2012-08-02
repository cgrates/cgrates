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
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/bmizerany/pq"
	"github.com/cgrates/cgrates/timespans"
	"log"
)

type PostgresLogger struct {
	db *sql.DB
}

func (psl *PostgresLogger) Close() {
	psl.db.Close()
}

func (psl *PostgresLogger) Log(uuid string, cc *timespans.CallCost) {
	if psl.db == nil {
		timespans.Logger.Warning("Cannot write log to database.")
		return
	}
	tss, err := json.Marshal(cc.Timespans)
	if err != nil {
		log.Printf("Error marshalling timespans to json: %v", err)
	}
	_, err = psl.db.Exec(fmt.Sprintf("INSERT INTO callcosts VALUES ('%s','%s', '%s', '%s', '%s', '%s', '%s', %v, %v, '%s')",
		uuid,
		cc.Destination,
		cc.Tenant,
		cc.TOR,
		cc.Subject,
		cc.Account,
		cc.Destination,
		cc.Cost,
		cc.ConnectFee,
		tss))
	if err != nil {
		log.Printf("failed to execute insert statement: %v", err)
	}
}
