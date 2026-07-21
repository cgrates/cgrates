// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"errors"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// V1RunReaderParams contains required parameters for an ErSv1.RunReader request.
type V1RunReaderParams struct {
	Tenant   string
	ID       string // unique identifier of the request
	ReaderID string
	APIOpts  map[string]any
}

// V1RunReader processes files in the configured directory for the given reader. This function handles files
// based on the reader's type and configuration. Only available for readers that are not processing files
// automatically (RunDelay should equal 0).
//
// Note: This API is not safe to call concurrently for the same reader. Ensure the current files finish being
// processed before calling again.
func (erS *ERService) V1RunReader(ctx *context.Context, params V1RunReaderParams, reply *string) error {
	rdrCfg := erS.cfg.ERsCfg().ReaderCfg(params.ReaderID)
	er, has := erS.rdrs[params.ReaderID]
	if !has || rdrCfg == nil {
		return utils.ErrNotFound
	}
	if rdrCfg.RunDelay != 0 {
		return errors.New("readers with RunDelay different from 0 are not supported")
	}
	switch rdr := er.(type) {
	case *CSVFileER:
		processReaderDir(rdr.sourceDir, utils.CSVSuffix, rdr.processFile)
	case *XMLFileER:
		processReaderDir(rdr.sourceDir, utils.XMLSuffix, rdr.processFile)
	case *FWVFileER:
		processReaderDir(rdr.sourceDir, utils.FWVSuffix, rdr.processFile)
	case *JSONFileER:
		processReaderDir(rdr.sourceDir, utils.JSONSuffix, rdr.processFile)
	default:
		return errors.New("reader type does not yet support manual processing")
	}
	*reply = utils.OK
	return nil
}
