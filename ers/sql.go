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

package ers

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	// libs for sql DBs
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

const (
	dbName         = "db_name"
	tableName      = "db_name"
	sslMode        = "sslmode"
	defaultSSLMode = "disable"
	defaultDBName  = "cgrates"
)

// NewSQLEventReader return a new kafka event reader
func NewSQLEventReader(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {

	rdr := &SQLEventReader{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrEvents: rdrEvents,
		rdrExit:   rdrExit,
		rdrErr:    rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	if err = rdr.setURL(rdr.Config().SourcePath, rdr.Config().ProcessedPath); err != nil {
		return
	}
	er = rdr

	return
}

// SQLEventReader implements EventReader interface for sql
type SQLEventReader struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	connString string
	connType   string
	tableName  string

	expConnString string
	expConnType   string
	expTableName  string

	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrExit   chan struct{}
	rdrErr    chan error
	cap       chan struct{}
}

// Config returns the curent configuration
func (rdr *SQLEventReader) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the kafka topic
func (rdr *SQLEventReader) Serve() (err error) {
	// connectString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'", user, password, host, port, name)
	var db *gorm.DB
	if db, err = gorm.Open(strings.TrimPrefix(rdr.connType, utils.Meta), rdr.connString); err != nil {
		return
	}
	if err = db.DB().Ping(); err != nil {
		return
	}
	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}
	go rdr.readLoop(db) // read until the connection is closed
	return
}

func (rdr *SQLEventReader) readLoop(db *gorm.DB) {
	for {
		if db = db.Table(rdr.tableName).Select("*"); db.Error != nil {
			rdr.rdrErr <- db.Error
			return
		}
		rows, err := db.Rows()
		if err != nil {
			rdr.rdrErr <- err
			return
		}
		colNames, err := rows.Columns()
		if err != nil {
			rdr.rdrErr <- err
			return
		}
		for rows.Next() {
			select {
			case <-rdr.rdrExit:
				utils.Logger.Info(
					fmt.Sprintf("<%s> stop monitoring sql DB <%s>",
						utils.ERs, rdr.Config().SourcePath))
				db.Close()
				return
			default:
			}
			if err := rows.Err(); err != nil {
				rdr.rdrErr <- err
				return
			}
			if rdr.Config().ConcurrentReqs != -1 {
				<-rdr.cap // do not try to read if the limit is reached
			}
			columns := make([]interface{}, len(colNames))
			columnPointers := make([]interface{}, len(colNames))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}
			if err = rows.Scan(columnPointers...); err != nil {
				rdr.rdrErr <- err
				return
			}
			go func(columns []interface{}, colNames []string) {
				msg := make(map[string]interface{})
				for i, colName := range colNames {
					msg[colName] = columns[i]
				}
				if err := rdr.processMessage(msg); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, utils.ToJSON(msg), err.Error()))
				}
				db.Delete(msg) // to ensure we don't read it again
				if rdr.Config().ProcessedPath != utils.EmptyString {
					// post it
				}
				if rdr.Config().ConcurrentReqs != -1 {
					rdr.cap <- struct{}{}
				}
			}(columns, colNames)
		}
		if rdr.Config().RunDelay < 0 {
			return
		}
		time.Sleep(rdr.Config().RunDelay)
	}
}

func (rdr *SQLEventReader) processMessage(msg map[string]interface{}) (err error) {
	reqVars := make(map[string]interface{})
	agReq := agents.NewAgentRequest(
		config.NewNavigableMap(msg), reqVars,
		nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	var navMp *config.NavigableMap
	if navMp, err = agReq.AsNavigableMap(rdr.Config().ContentFields); err != nil {
		return
	}
	rdr.rdrEvents <- &erEvent{
		cgrEvent: navMp.AsCGREvent(agReq.Tenant, utils.NestingSep),
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *SQLEventReader) setURL(inURL, outURL string) (err error) {
	// *dbtype:user:password@host:port?options
	split := strings.SplitN(inURL, utils.InInFieldSep, 2)
	if len(split) != 2 {
		return utils.NewErrMandatoryIeMissing("db_type")
	}
	rdr.connType = split[0]
	inURL = split[1]

	//outhpath if no meta is op[tions only
	var u *url.URL
	if u, err = url.Parse(inURL); err != nil {
		return
	}
	password, _ := u.User.Password()
	qry := u.Query()

	dbname := defaultDBName
	if vals, has := qry[dbName]; has && len(vals) != 0 {
		dbname = vals[0]
	}
	ssl := defaultSSLMode
	if vals, has := qry[sslMode]; has && len(vals) != 0 {
		ssl = vals[0]
	}

	rdr.tableName = utils.CDRsTBL
	if vals, has := qry[tableName]; has && len(vals) != 0 {
		rdr.tableName = vals[0]
	}

	switch rdr.connType {
	case utils.MetaMySQL:
		rdr.connString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			u.User.Username(), password, u.Hostname(), u.Port(), dbname)
	case utils.MetaPostgres:
		rdr.connString = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", u.Hostname(), u.Port(), dbname, u.User.Username(), password, ssl)
	default:
		return fmt.Errorf("unknown db_type %s", rdr.connType)
	}

	// outURL
	if len(outURL) == 0 {
		return
	}
	var outUser, outPassword, outDBname, outSSL, outHost, outPort string
	var oqry url.Values
	if !strings.HasPrefix(outURL, utils.Meta) {
		rdr.expConnType = rdr.connType
		outUser = u.User.Username()
		outPassword = password
		outHost = u.Hostname()
		outPort = u.Port()
		if oqry, err = url.ParseQuery(outURL); err != nil {
			return
		}
	} else {
		split := strings.SplitN(inURL, utils.ConcatenatedKey(), 2)
		if len(split) != 2 {
			return utils.NewErrMandatoryIeMissing("db_type")
		}
		rdr.expConnType = split[0]
		inURL = split[1]

		//outhpath if no meta is op[tions only
		var outURL *url.URL
		if outURL, err = url.Parse(inURL); err != nil {
			return
		}
		outPassword, _ = outURL.User.Password()
		outUser = outURL.User.Username()
		outHost = outURL.Hostname()
		outPort = outURL.Port()
		oqry = outURL.Query()
	}

	outDBname = defaultDBName
	if vals, has := oqry[dbName]; has && len(vals) != 0 {
		outDBname = vals[0]
	}
	outSSL = defaultSSLMode
	if vals, has := oqry[sslMode]; has && len(vals) != 0 {
		outSSL = vals[0]
	}
	rdr.expTableName = utils.CDRsTBL
	if vals, has := qry[tableName]; has && len(vals) != 0 {
		rdr.expTableName = vals[0]
	}

	switch rdr.connType {
	case utils.MetaMySQL:
		rdr.expConnString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&loc=Local&parseTime=true&sql_mode='ALLOW_INVALID_DATES'",
			outUser, outPassword, outHost, outPort, outDBname)
	case utils.MetaPostgres:
		rdr.expConnString = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			outHost, outPort, outDBname, outUser, outPassword, outSSL)
	default:
		return fmt.Errorf("unknown db type")
	}
	return
}
