// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ers

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
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
		return errors.New("readers with a non-zero RunDelay are not supported")
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
	case *CgrUR:
		if err := rdr.run(filters); err != nil {
			return err
		}
	default:
		return errors.New("reader type does not yet support manual processing")
	}
	*reply = utils.OK
	return nil
}

// V1ProcessReaderEvents will find all events on the specified reader, based on filters, and process them while returning any replies it gets back
func (erS *ERService) V1ProcessReaderEvents(ctx *context.Context, params V1RunReaderParams, reply *[]*sessions.V1ProcessEventReply) (err error) {
	rdrCfg := erS.cfg.ERsCfg().ReaderCfg(params.ReaderID)
	er, has := erS.rdrs[params.ReaderID]
	if !has || rdrCfg == nil {
		return utils.ErrNotFound
	}
	if rdrCfg.RunDelay != 0 {
		return errors.New("readers with a non-zero RunDelay are not supported")
	}
	filters := slices.Concat(rdrCfg.Filters, params.Filters)
	var rdrEvents []*erEvent
	switch rdr := er.(type) {
	case *CSVFileER:
		rdrEvents = getReaderEventsFromDir(rdr.sourceDir, utils.CSVSuffix,
			func(fileName string) ([]*erEvent, error) { return rdr.getReaderEvents(fileName, filters) })
	case *XMLFileER:
		rdrEvents = getReaderEventsFromDir(rdr.sourceDir, utils.XMLSuffix,
			func(fileName string) ([]*erEvent, error) { return rdr.getReaderEvents(fileName, filters) })
	case *FWVFileER:
		rdrEvents = getReaderEventsFromDir(rdr.sourceDir, utils.FWVSuffix,
			func(fileName string) ([]*erEvent, error) { return rdr.getReaderEvents(fileName, filters) })
	case *JSONFileER:
		rdrEvents = getReaderEventsFromDir(rdr.sourceDir, utils.JSONSuffix,
			func(fileName string) ([]*erEvent, error) {
				ev, err := rdr.getReaderEvents(fileName, filters)
				return []*erEvent{ev}, err
			})
	case *CgrUR:
		rdrEvents, err = rdr.getReaderEvents(filters)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("reader type <%v> not yet supported", rdr.Config().Type)
	}

	sessReplies := make([]*sessions.V1ProcessEventReply, 0) // gather all replies from processEvent (when event is passed through readers)
	for _, rdrEv := range rdrEvents {
		// if event is partial, proccess it, if it becomes complete after processing, process normaly the completed event
		if _, isPartial := rdrEv.cgrEvent.APIOpts[utils.PartialOpt]; isPartial {
			rdrEv.cgrEvent, err = erS.processPartialEvent(rdrEv.cgrEvent, rdrEv.rdrCfg)
			if err != nil {
				utils.Logger.Warning(fmt.Sprintf(
					"<%s> reading partial event: <%s> from reader: <%s> got error: <%v>",
					utils.ERs, utils.ToJSON(rdrEv.cgrEvent), rdrEv.rdrCfg.ID, err))
				return err
			}
			if rdrEv.cgrEvent == nil {
				continue
			}
		}
		// process either a normal full event, or the completed event (from partials)
		sessReply, err := erS.processEvent(rdrEv.cgrEvent, rdrEv.rdrCfg)
		if err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> reading event: <%s> from reader: <%s> got error: <%v>",
				utils.ERs, utils.ToJSON(rdrEv.cgrEvent), rdrEv.rdrCfg.ID, err))
			return err
		}
		sessReplies = append(sessReplies, sessReply)
	}
	*reply = sessReplies
	return
}
