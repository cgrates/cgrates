// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"errors"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// V1RunReaderParams contains required parameters for an ErSv1.RunReader request.
type V1RunReaderParams struct {
	Tenant   string
	ID       string // unique identifier of the request
	ReaderID string
	Filters  []string
	APIOpts  map[string]any
}

// V1RunReader processes the configured reader once. Request filters are applied together with the configured filters.
// Only readers with RunDelay set to 0 can be run manually.
//
// Note: This API is not safe to call concurrently for the same reader. Ensure the current input finishes being
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
	filters := slices.Concat(rdrCfg.Filters, params.Filters)
	switch rdr := er.(type) {
	case *CSVFileER:
		processReaderDir(rdr.sourceDir, utils.CSVSuffix,
			func(fileName string) error { return rdr.processFile(fileName, filters) })
	case *XMLFileER:
		processReaderDir(rdr.sourceDir, utils.XMLSuffix,
			func(fileName string) error { return rdr.processFile(fileName, filters) })
	case *FWVFileER:
		processReaderDir(rdr.sourceDir, utils.FWVSuffix,
			func(fileName string) error { return rdr.processFile(fileName, filters) })
	case *JSONFileER:
		processReaderDir(rdr.sourceDir, utils.JSONSuffix,
			func(fileName string) error { return rdr.processFile(fileName, filters) })
	case *CgrCDR:
		if err := rdr.run(filters); err != nil {
			return err
		}
	default:
		return errors.New("reader type does not yet support manual processing")
	}
	*reply = utils.OK
	return nil
}
