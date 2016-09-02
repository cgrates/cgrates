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
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

func main() {
	log.Print("Start!")
	var wg sync.WaitGroup
	for i := 1; i < 1002; i++ {
		go func(index int) {
			wg.Add(1)
			resp, err := http.Post("http://localhost:2080/jsonrpc", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"method": "ApierV1.SetAccount","params": [{"Tenant":"reglo","Account":"100%d","ActionPlanId":"PACKAGE_NEW_FOR795", "ReloadScheduler":false}], "id":%d}`, index, index))))
			if err != nil {
				log.Print("Post error: ", err)
			}
			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Print("Body error: ", err)
			}
			log.Printf("SetAccount(%d): %s", index, string(contents))
			wg.Done()
		}(i)
	}
	wg.Wait()
	for index := 1; index < 1002; index++ {
		resp, err := http.Post("http://localhost:2080/jsonrpc", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"method": "ApierV1.GetAccountActionPlan","params": [{"Tenant":"reglo","Account":"100%d"}], "id":%d}`, index, index))))
		if err != nil {
			log.Print("Post error: ", err)
		}
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print("Body error: ", err)
		}
		log.Printf("GetAccountActionPlan(%d): %s", index, string(contents))
	}

	log.Print("Done!")
}
