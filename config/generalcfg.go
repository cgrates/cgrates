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

package config

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// General config section
type GeneralCfg struct {
	NodeID            string        // Identifier for this engine instance
	Logger            string        // dictates the way logs are displayed/stored
	LogLevel          int           // system wide log level, nothing higher than this will be logged
	HttpSkipTlsVerify bool          // If enabled Http Client will accept any TLS certificate
	RoundingDecimals  int           // Number of decimals to round end prices at
	DBDataEncoding    string        // The encoding used to store object data in strings: <msgpack|json>
	TpExportPath      string        // Path towards export folder for offline Tariff Plans
	PosterAttempts    int           // Time to wait before writing the failed posts in a single file
	FailedPostsDir    string        // Directory path where we store failed http requests
	FailedPostsTTL    time.Duration // Directory path where we store failed http requests
	DefaultReqType    string        // Use this request type if not defined on top
	DefaultCategory   string        // set default type of record
	DefaultTenant     string        // set default tenant
	DefaultTimezone   string        // default timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	DefaultCaching    string
	ConnectAttempts   int           // number of initial connection attempts before giving up
	Reconnects        int           // number of recconect attempts in case of connection lost <-1 for infinite | nb>
	ConnectTimeout    time.Duration // timeout for RPC connection attempts
	ReplyTimeout      time.Duration // timeout replies if not reaching back
	LockingTimeout    time.Duration // locking mechanism timeout to avoid deadlocks
	DigestSeparator   string        //
	DigestEqual       string        //
	RSRSep            string        // separator used to split RSRParser (by degault is used ";")
	MaxParralelConns  int           // the maximum number of connection used by the *parallel strategy
}

//loadFromJsonCfg loads General config from JsonCfg
func (gencfg *GeneralCfg) loadFromJsonCfg(jsnGeneralCfg *GeneralJsonCfg) (err error) {
	if jsnGeneralCfg == nil {
		return nil
	}
	if jsnGeneralCfg.Node_id != nil && *jsnGeneralCfg.Node_id != "" {
		gencfg.NodeID = *jsnGeneralCfg.Node_id
	}
	if jsnGeneralCfg.Logger != nil {
		gencfg.Logger = *jsnGeneralCfg.Logger
	}
	if jsnGeneralCfg.Log_level != nil {
		gencfg.LogLevel = *jsnGeneralCfg.Log_level
	}

	if jsnGeneralCfg.Dbdata_encoding != nil {
		gencfg.DBDataEncoding = strings.TrimPrefix(*jsnGeneralCfg.Dbdata_encoding, "*")
	}
	if jsnGeneralCfg.Default_request_type != nil {
		gencfg.DefaultReqType = *jsnGeneralCfg.Default_request_type
	}
	if jsnGeneralCfg.Default_category != nil {
		gencfg.DefaultCategory = *jsnGeneralCfg.Default_category
	}
	if jsnGeneralCfg.Default_tenant != nil {
		gencfg.DefaultTenant = *jsnGeneralCfg.Default_tenant
	}
	if jsnGeneralCfg.Connect_attempts != nil {
		gencfg.ConnectAttempts = *jsnGeneralCfg.Connect_attempts
	}
	if jsnGeneralCfg.Reconnects != nil {
		gencfg.Reconnects = *jsnGeneralCfg.Reconnects
	}
	if jsnGeneralCfg.Connect_timeout != nil {
		if gencfg.ConnectTimeout, err = utils.ParseDurationWithNanosecs(*jsnGeneralCfg.Connect_timeout); err != nil {
			return err
		}
	}
	if jsnGeneralCfg.Reply_timeout != nil {
		if gencfg.ReplyTimeout, err = utils.ParseDurationWithNanosecs(*jsnGeneralCfg.Reply_timeout); err != nil {
			return err
		}
	}
	if jsnGeneralCfg.Rounding_decimals != nil {
		gencfg.RoundingDecimals = *jsnGeneralCfg.Rounding_decimals
	}
	if jsnGeneralCfg.Http_skip_tls_verify != nil {
		gencfg.HttpSkipTlsVerify = *jsnGeneralCfg.Http_skip_tls_verify
	}
	if jsnGeneralCfg.Tpexport_dir != nil {
		gencfg.TpExportPath = *jsnGeneralCfg.Tpexport_dir
	}
	if jsnGeneralCfg.Poster_attempts != nil {
		gencfg.PosterAttempts = *jsnGeneralCfg.Poster_attempts
	}
	if jsnGeneralCfg.Failed_posts_dir != nil {
		gencfg.FailedPostsDir = *jsnGeneralCfg.Failed_posts_dir
	}
	if jsnGeneralCfg.Failed_posts_ttl != nil {
		if gencfg.FailedPostsTTL, err = utils.ParseDurationWithNanosecs(*jsnGeneralCfg.Failed_posts_ttl); err != nil {
			return err
		}
	}
	if jsnGeneralCfg.Default_timezone != nil {
		gencfg.DefaultTimezone = *jsnGeneralCfg.Default_timezone
	}
	if jsnGeneralCfg.Default_caching != nil {
		gencfg.DefaultCaching = *jsnGeneralCfg.Default_caching
	}
	if jsnGeneralCfg.Locking_timeout != nil {
		if gencfg.LockingTimeout, err = utils.ParseDurationWithNanosecs(*jsnGeneralCfg.Locking_timeout); err != nil {
			return err
		}
	}
	if jsnGeneralCfg.Digest_separator != nil {
		gencfg.DigestSeparator = *jsnGeneralCfg.Digest_separator
	}
	if jsnGeneralCfg.Digest_equal != nil {
		gencfg.DigestEqual = *jsnGeneralCfg.Digest_equal
	}
	if jsnGeneralCfg.Rsr_separator != nil {
		gencfg.RSRSep = *jsnGeneralCfg.Rsr_separator
	}
	if jsnGeneralCfg.Max_parralel_conns != nil {
		gencfg.MaxParralelConns = *jsnGeneralCfg.Max_parralel_conns
	}

	return nil
}

func (gencfg *GeneralCfg) AsMapInterface() map[string]interface{} {
	return map[string]interface{}{
		utils.NodeIDCfg:            gencfg.NodeID,
		utils.LoggerCfg:            gencfg.Logger,
		utils.LogLevelCfg:          gencfg.LogLevel,
		utils.HttpSkipTlsVerifyCfg: gencfg.HttpSkipTlsVerify,
		utils.RoundingDecimalsCfg:  gencfg.RoundingDecimals,
		utils.DBDataEncodingCfg:    gencfg.DBDataEncoding,
		utils.TpExportPathCfg:      gencfg.TpExportPath,
		utils.PosterAttemptsCfg:    gencfg.PosterAttempts,
		utils.FailedPostsDirCfg:    gencfg.FailedPostsDir,
		utils.FailedPostsTTLCfg:    gencfg.FailedPostsTTL,
		utils.DefaultReqTypeCfg:    gencfg.DefaultReqType,
		utils.DefaultCategoryCfg:   gencfg.DefaultCategory,
		utils.DefaultTenantCfg:     gencfg.DefaultTenant,
		utils.DefaultTimezoneCfg:   gencfg.DefaultTimezone,
		utils.DefaultCachingCfg:    gencfg.DefaultCaching,
		utils.ConnectAttemptsCfg:   gencfg.ConnectAttempts,
		utils.ReconnectsCfg:        gencfg.Reconnects,
		utils.ConnectTimeoutCfg:    gencfg.ConnectTimeout,
		utils.ReplyTimeoutCfg:      gencfg.ReplyTimeout,
		utils.LockingTimeoutCfg:    gencfg.LockingTimeout,
		utils.DigestSeparatorCfg:   gencfg.DigestSeparator,
		utils.DigestEqualCfg:       gencfg.DigestEqual,
		utils.RSRSepCfg:            gencfg.RSRSep,
		utils.MaxParralelConnsCfg:  gencfg.MaxParralelConns,
	}
}
