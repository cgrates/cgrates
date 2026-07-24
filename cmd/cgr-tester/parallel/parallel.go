// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

func main() {
	log.Print("Start!")
	var wg sync.WaitGroup
	for i := 1; i < 1002; i++ {
		wg.Add(1)
		go func(index int) {
			resp, err := http.Post("http://localhost:2080/jsonrpc", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"method": "APIerSv1.SetAccount","params": [{"Tenant":"reglo","Account":"100%d","ActionPlanId":"PACKAGE_NEW_FOR795", "ReloadScheduler":false}], "id":%d}`, index, index))))
			if err != nil {
				log.Print("Post error: ", err)
			}
			contents, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Print("Body error: ", err)
			}
			log.Printf("SetAccount(%d): %s", index, string(contents))
			wg.Done()
		}(i)
	}
	wg.Wait()
	for index := 1; index < 1002; index++ {
		resp, err := http.Post("http://localhost:2080/jsonrpc", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"method": "%s","params": [{"Tenant":"reglo","Account":"100%d"}], "id":%d}`, utils.APIerSv1GetAccountActionPlan, index, index))))
		if err != nil {
			log.Print("Post error: ", err)
		}
		contents, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Print("Body error: ", err)
		}
		log.Printf("GetAccountActionPlan(%d): %s", index, string(contents))
	}

	log.Print("Done!")
}
