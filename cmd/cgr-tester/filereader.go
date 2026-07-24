// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func NewFileReaderTester(fPath, cgrAddr string, parallel, runs int, reqSep []byte) (frt *FileReaderTester, err error) {
	frt = &FileReaderTester{
		parallel: parallel, runs: runs,
		reqSep: reqSep,
	}
	if frt.rdr, err = os.Open(fPath); err != nil {
		return nil, err
	}
	if frt.conn, err = net.Dial(utils.TCP, cgrAddr); err != nil {
		return nil, err
	}
	return
}

// TesterReader will read requests from file and post them remotely
type FileReaderTester struct {
	parallel int
	runs     int
	reqSep   []byte

	rdr      io.Reader
	conn     net.Conn
	connScnr *bufio.Scanner
}

func (frt *FileReaderTester) connSendReq(req []byte) (err error) {
	frt.conn.SetReadDeadline(time.Now().Add(time.Millisecond)) // will block most of the times on read
	if _, err = frt.conn.Write(req); err != nil {
		return
	}
	io.ReadAll(frt.conn)
	return
}

// Test reads from rdr, split the content based on lineSep and sends individual lines to remote
func (frt *FileReaderTester) Test() (err error) {
	var fContent []byte
	if fContent, err = io.ReadAll(frt.rdr); err != nil {
		return
	}
	reqs := bytes.Split(fContent, frt.reqSep)

	// parallel requests
	if frt.parallel > 0 {
		var wg sync.WaitGroup
		reqLimiter := make(chan struct{}, frt.parallel)
		for i := 0; i < frt.runs; i++ {
			wg.Add(1)
			go func(i int) {
				reqLimiter <- struct{}{} // block till buffer will allow
				if err := frt.connSendReq(reqs[rand.Intn(len(reqs))]); err != nil {
					log.Printf("ERROR: %s", err.Error())
				}
				<-reqLimiter // release one request from buffer
				wg.Done()
			}(i)
		}
		wg.Wait()
		return
	}

	// serial requests
	for i := 0; i < frt.runs; i++ {
		for _, req := range reqs {
			if err := frt.connSendReq(req); err != nil {
				log.Printf("ERROR: %s", err.Error())
			}
		}
	}
	return
}
