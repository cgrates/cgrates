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
