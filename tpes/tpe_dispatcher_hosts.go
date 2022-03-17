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

package tpes

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type TPDispatcherHosts struct {
	dm *engine.DataManager
}

// newTPDispatcherHosts is the constructor for TPDispatcherHosts
func newTPDispatcherHosts(dm *engine.DataManager) *TPDispatcherHosts {
	return &TPDispatcherHosts{
		dm: dm,
	}
}

// exportItems for TPDispatcherHosts will implement the method for tpExporter interface
func (tpDspHst TPDispatcherHosts) exportItems(ctx *context.Context, wrtr io.Writer, tnt string, itmIDs []string) (err error) {
	csvWriter := csv.NewWriter(wrtr)
	csvWriter.Comma = utils.CSVSep
	// before writing the profiles, we must write the headers
	if err = csvWriter.Write([]string{"#Tenant", "ID", "Address", "Transport", "ConnectAttempts", "Reconnects", "ConnectTimeout", "ReplyTimeout", "Tls", "ClientKey", "ClientCertificate", "CaCertificate"}); err != nil {
		return
	}
	for _, dspHostID := range itmIDs {
		var dspHostPrf *engine.DispatcherHost
		dspHostPrf, err = tpDspHst.dm.GetDispatcherHost(ctx, tnt, dspHostID, true, true, utils.NonTransactional)
		if err != nil {
			if err.Error() == utils.ErrNotFound.Error() {
				return fmt.Errorf("<%s> cannot find DispatcherHost with id: <%v>", err, dspHostID)
			}
			return err
		}
		dspHstsMdl := engine.APItoModelTPDispatcherHost(engine.DispatcherHostToAPI(dspHostPrf))
		// for every profile, convert it into model to be compatible in csv format
		// transform every record into a []string
		record, err := engine.CsvDump(dspHstsMdl)
		if err != nil {
			return err
		}
		// record is a line of a csv file
		if err := csvWriter.Write(record); err != nil {
			return err
		}

	}
	csvWriter.Flush()
	return
}
