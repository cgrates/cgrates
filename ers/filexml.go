/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ers

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/xmlquery"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewXMLFileER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	srcPath := cfg.ERsCfg().Readers[cfgIdx].SourcePath
	if strings.HasSuffix(srcPath, utils.Slash) {
		srcPath = srcPath[:len(srcPath)-1]
	}
	xmlER := &XMLFileER{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		sourceDir:     srcPath,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrError:      rdrErr,
		rdrExit:       rdrExit,
		conReqs:       make(chan struct{}, cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs)}
	var processFile struct{}
	for i := 0; i < cfg.ERsCfg().Readers[cfgIdx].ConcurrentReqs; i++ {
		xmlER.conReqs <- processFile // Empty initiate so we do not need to wait later when we pop
	}
	return xmlER, nil
}

// XMLFileER implements EventReader interface for .xml files
type XMLFileER struct {
	sync.RWMutex
	cgrCfg        *config.CGRConfig
	cfgIdx        int // index of config instance within ERsCfg.Readers
	fltrS         *engine.FilterS
	sourceDir     string        // path to the directory monitored by the reader for new events
	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrError      chan error
	rdrExit       chan struct{}
	conReqs       chan struct{} // limit number of opened files
}

func (rdr *XMLFileER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *XMLFileER) Serve() (err error) {
	switch rdr.Config().RunDelay {
	case time.Duration(0): // 0 disables the automatic read, maybe done per API
		return
	case time.Duration(-1):
		go func() {
			time.Sleep(rdr.Config().StartDelay)
			// Ensure that files already existing in the source path are processed
			// before the reader starts listening for filesystem change events.
			processReaderDir(rdr.sourceDir, utils.XMLSuffix, rdr.processFile)

			if err := utils.WatchDir(rdr.sourceDir, rdr.processFile,
				utils.ERs, rdr.rdrExit); err != nil {
				rdr.rdrError <- err
			}
		}()
	default:
		go func() {
			if rdr.Config().StartDelay > 0 {
				select {
				case <-time.After(rdr.Config().StartDelay):
				case <-rdr.rdrExit:
					utils.Logger.Info(
						fmt.Sprintf("<%s> stop monitoring path <%s>",
							utils.ERs, rdr.sourceDir))
					return
				}
			}
			tm := time.NewTimer(0)
			for {
				// Not automated, process and sleep approach
				select {
				case <-rdr.rdrExit:
					tm.Stop()
					utils.Logger.Info(
						fmt.Sprintf("<%s> stop monitoring path <%s>",
							utils.ERs, rdr.sourceDir))
					return
				case <-tm.C:
				}
				processReaderDir(rdr.sourceDir, utils.XMLSuffix, rdr.processFile)
				tm.Reset(rdr.Config().RunDelay)
			}
		}()
	}
	return
}

/*
   `xml_root_path` is a slice that determines which XML nodes to process.
   When used by `xmlquery.QueryAll()`, it behaves as follows:
   ```xml
   <?xml version="1.0" encoding="ISO-8859-1"?>
   <A>
       <B>
           <C>item1</C>
           <D>item2</D>
       </B>
       <B>
           <C>item3</C>
       </B>
   </A>
   ```
   - If the root_path_string is empty or ["A"], it retrieves everything within <A></A>.
   - For ["A", "B"], it retrieves each <B></B> element.
   - For ["A", "B", "C"], it retrieves the text within each <C></C> ("item1" and "item3").
*/

// processFile is called for each file in a directory and dispatches erEvents from it
func (rdr *XMLFileER) processFile(fName string) error {
	if cap(rdr.conReqs) != 0 { // 0 goes for no limit
		processFile := <-rdr.conReqs // Queue here for maxOpenFiles
		defer func() { rdr.conReqs <- processFile }()
	}
	absPath := path.Join(rdr.sourceDir, fName)
	utils.Logger.Info(
		fmt.Sprintf("<%s> parsing <%s>", utils.ERs, absPath))
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()
	doc, err := xmlquery.Parse(file)
	if err != nil {
		return err
	}
	var xmlRootPath utils.HierarchyPath
	if rdr.Config().Opts.XMLRootPath != nil {
		xmlRootPath = utils.ParseHierarchyPath(*rdr.Config().Opts.XMLRootPath, utils.EmptyString)
	}
	xmlElmts, err := xmlquery.QueryAll(doc, xmlRootPath.AsString("/", true))
	if err != nil {
		return err
	}
	rowNr := 0 // This counts the rows in the file, not really number of CDRs
	evsPosted := 0
	timeStart := time.Now()
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaFileName: utils.NewLeafNode(fName), utils.MetaReaderID: utils.NewLeafNode(rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].ID)}}
	for _, xmlElmt := range xmlElmts {
		rowNr++ // increment the rowNr after checking if it's not the end of file
		reqVars.Map[utils.MetaFileLineNumber] = utils.NewLeafNode(rowNr)
		agReq := agents.NewAgentRequest(
			config.NewXMLProvider(xmlElmt, xmlRootPath), reqVars,
			nil, nil, nil, rdr.Config().Tenant,
			rdr.cgrCfg.GeneralCfg().DefaultTenant,
			utils.FirstNonEmpty(rdr.Config().Timezone,
				rdr.cgrCfg.GeneralCfg().DefaultTimezone),
			rdr.fltrS, nil) // create an AgentRequest
		if pass, err := rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
			agReq); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to filter error: <%s>",
					utils.ERs, absPath, rowNr, err.Error()))
			return err
		} else if !pass {
			continue
		}
		if err := agReq.SetFields(rdr.Config().Fields); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> reading file: <%s> row <%d>, ignoring due to error: <%s>",
					utils.ERs, absPath, rowNr, err.Error()))
			continue
		}
		cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
		rdrEv := rdr.rdrEvents
		if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
			rdrEv = rdr.partialEvents
		}
		rdrEv <- &erEvent{
			cgrEvent: cgrEv,
			rdrCfg:   rdr.Config(),
		}
		evsPosted++
	}

	if rdr.Config().ProcessedPath != "" {
		// Finished with file, move it to processed folder
		outPath := path.Join(rdr.Config().ProcessedPath, fName)
		if err = os.Rename(absPath, outPath); err != nil {
			return err
		}
	}

	utils.Logger.Info(
		fmt.Sprintf("%s finished processing file <%s>. Total records processed: %d, events posted: %d, run duration: %s",
			utils.ERs, absPath, rowNr, evsPosted, time.Now().Sub(timeStart)))
	return nil
}
