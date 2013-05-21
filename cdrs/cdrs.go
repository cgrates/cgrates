/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package cdrs

import (
	"fmt"
	"github.com/cgrates/cgrates/mediator"
	"github.com/cgrates/cgrates/rater"
	"io/ioutil"
	"net/http"
)

var (
	Logger = rater.Logger
)

type CDRS struct {
	loggerDb rater.DataStorage
	medi     *mediator.Mediator
}

func (cdrs *CDRS) cdrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if fsCdr, err := new(FSCdr).New(body); err == nil {
		cdrs.loggerDb.SetCdr(fsCdr)
		cdrs.medi.MediateCdrFromDB(fsCdr.GetAccount(), cdrs.loggerDb)
	} else {
		Logger.Err(fmt.Sprintf("Could not create CDR entry: %v", err))
	}
}

func New(storage rater.DataStorage, mediator *mediator.Mediator) *CDRS {
	return &CDRS{storage, mediator}
}

func (cdrs *CDRS) StartCaptiuringCDRs() {
	http.HandleFunc("/cdr", cdrs.cdrHandler)
	http.ListenAndServe(":8080", nil)
}
