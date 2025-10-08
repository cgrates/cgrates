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

package config

import (
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

type GeneralOpts struct {
	ExporterIDs []*DynamicStringSliceOpt
}

// GeneralCfg is the general config section
type GeneralCfg struct {
	NodeID               string // Identifier for this engine instance
	RoundingDecimals     int    // Number of decimals to round end prices at
	DBDataEncoding       string // The encoding used to store object data in strings: <msgpack|json>
	TpExportPath         string // Path towards export folder for offline Tariff Plans
	DefaultReqType       string // Use this request type if not defined on top
	DefaultCategory      string // set default type of record
	DefaultTenant        string // set default tenant
	DefaultTimezone      string // default timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	DefaultCaching       string
	CachingDelay         time.Duration // use to add delay before cache reload
	ConnectAttempts      int           // number of initial connection attempts before giving up
	Reconnects           int           // number of recconect attempts in case of connection lost <-1 for infinite | nb>
	MaxReconnectInterval time.Duration // time to wait in between reconnect attempts
	ConnectTimeout       time.Duration // timeout for RPC connection attempts
	ReplyTimeout         time.Duration // timeout replies if not reaching back
	LockingTimeout       time.Duration // locking mechanism timeout to avoid deadlocks
	DigestSeparator      string        //
	DigestEqual          string        //
	MaxParallelConns     int           // the maximum number of connections used by the *parallel strategy

	DecimalMaxScale     int
	DecimalMinScale     int
	DecimalPrecision    int
	DecimalRoundingMode decimal.RoundingMode
	Opts                *GeneralOpts
}

func (gencfg *GeneralCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnGeneralCfg := new(GeneralJsonCfg)
	if err = jsnCfg.GetSection(ctx, GeneralJSON, jsnGeneralCfg); err != nil {
		return
	}
	return gencfg.loadFromJSONCfg(jsnGeneralCfg)
}

// loadGeneralCfg loads the General opts section of the configuration
func (generalOpts *GeneralOpts) loadFromJSONCfg(jsnCfg *GeneralOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.ExporterIDs != nil {
		generalOpts.ExporterIDs = append(generalOpts.ExporterIDs, jsnCfg.ExporterIDs...)
	}
}

// loadFromJSONCfg loads General config from JsonCfg
func (gencfg *GeneralCfg) loadFromJSONCfg(jsnGeneralCfg *GeneralJsonCfg) (err error) {
	if jsnGeneralCfg == nil {
		return nil
	}
	if jsnGeneralCfg.Node_id != nil && *jsnGeneralCfg.Node_id != "" {
		gencfg.NodeID = *jsnGeneralCfg.Node_id
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
	if jsnGeneralCfg.Max_reconnect_interval != nil {
		if gencfg.MaxReconnectInterval, err = utils.ParseDurationWithNanosecs(*jsnGeneralCfg.Max_reconnect_interval); err != nil {
			return err
		}
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
	if jsnGeneralCfg.Tpexport_dir != nil {
		gencfg.TpExportPath = *jsnGeneralCfg.Tpexport_dir
	}
	if jsnGeneralCfg.Default_timezone != nil {
		gencfg.DefaultTimezone = *jsnGeneralCfg.Default_timezone
	}
	if jsnGeneralCfg.Default_caching != nil {
		gencfg.DefaultCaching = *jsnGeneralCfg.Default_caching
	}
	if jsnGeneralCfg.Caching_delay != nil {
		if gencfg.CachingDelay, err = utils.ParseDurationWithNanosecs(*jsnGeneralCfg.Caching_delay); err != nil {
			return err
		}
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
	if jsnGeneralCfg.Max_parallel_conns != nil {
		gencfg.MaxParallelConns = *jsnGeneralCfg.Max_parallel_conns
	}

	if jsnGeneralCfg.Decimal_max_scale != nil {
		gencfg.DecimalMaxScale = *jsnGeneralCfg.Decimal_max_scale
	}
	if jsnGeneralCfg.Decimal_min_scale != nil {
		gencfg.DecimalMinScale = *jsnGeneralCfg.Decimal_min_scale
	}
	if jsnGeneralCfg.Decimal_precision != nil {
		gencfg.DecimalPrecision = *jsnGeneralCfg.Decimal_precision
	}
	if jsnGeneralCfg.Decimal_rounding_mode != nil {
		gencfg.DecimalRoundingMode, err = utils.NewRoundingMode(*jsnGeneralCfg.Decimal_rounding_mode)
	}
	if jsnGeneralCfg.Opts != nil {
		gencfg.Opts.loadFromJSONCfg(jsnGeneralCfg.Opts)
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (gencfg GeneralCfg) AsMapInterface() any {
	opts := map[string]any{
		utils.MetaExporterIDs: gencfg.Opts.ExporterIDs,
	}
	mp := map[string]any{
		utils.NodeIDCfg:               gencfg.NodeID,
		utils.RoundingDecimalsCfg:     gencfg.RoundingDecimals,
		utils.DBDataEncodingCfg:       utils.Meta + gencfg.DBDataEncoding,
		utils.TpExportPathCfg:         gencfg.TpExportPath,
		utils.DefaultReqTypeCfg:       gencfg.DefaultReqType,
		utils.DefaultCategoryCfg:      gencfg.DefaultCategory,
		utils.DefaultTenantCfg:        gencfg.DefaultTenant,
		utils.DefaultTimezoneCfg:      gencfg.DefaultTimezone,
		utils.DefaultCachingCfg:       gencfg.DefaultCaching,
		utils.CachingDlayCfg:          "0",
		utils.ConnectAttemptsCfg:      gencfg.ConnectAttempts,
		utils.ReconnectsCfg:           gencfg.Reconnects,
		utils.MaxReconnectIntervalCfg: "0",
		utils.DigestSeparatorCfg:      gencfg.DigestSeparator,
		utils.DigestEqualCfg:          gencfg.DigestEqual,
		utils.MaxParallelConnsCfg:     gencfg.MaxParallelConns,
		utils.LockingTimeoutCfg:       "0",
		utils.ConnectTimeoutCfg:       "0",
		utils.ReplyTimeoutCfg:         "0",
		utils.DecimalMaxScaleCfg:      gencfg.DecimalMaxScale,
		utils.DecimalMinScaleCfg:      gencfg.DecimalMinScale,
		utils.DecimalPrecisionCfg:     gencfg.DecimalPrecision,
		utils.DecimalRoundingModeCfg:  utils.RoundingModeToString(gencfg.DecimalRoundingMode),
		utils.OptsCfg:                 opts,
	}

	if gencfg.CachingDelay != 0 {
		mp[utils.CachingDlayCfg] = gencfg.CachingDelay.String()
	}

	if gencfg.MaxReconnectInterval != 0 {
		mp[utils.MaxReconnectIntervalCfg] = gencfg.MaxReconnectInterval.String()
	}

	if gencfg.LockingTimeout != 0 {
		mp[utils.LockingTimeoutCfg] = gencfg.LockingTimeout.String()
	}

	if gencfg.ConnectTimeout != 0 {
		mp[utils.ConnectTimeoutCfg] = gencfg.ConnectTimeout.String()
	}

	if gencfg.ReplyTimeout != 0 {
		mp[utils.ReplyTimeoutCfg] = gencfg.ReplyTimeout.String()
	}

	return mp
}

func (GeneralCfg) SName() string                { return GeneralJSON }
func (gencfg GeneralCfg) CloneSection() Section { return gencfg.Clone() }

func (generalOpts *GeneralOpts) Clone() *GeneralOpts {
	if generalOpts == nil {
		return nil
	}
	return &GeneralOpts{
		ExporterIDs: CloneDynamicStringSliceOpt(generalOpts.ExporterIDs),
	}
}

// Clone returns a deep copy of GeneralCfg
func (gencfg GeneralCfg) Clone() *GeneralCfg {
	return &GeneralCfg{
		NodeID:               gencfg.NodeID,
		RoundingDecimals:     gencfg.RoundingDecimals,
		DBDataEncoding:       gencfg.DBDataEncoding,
		TpExportPath:         gencfg.TpExportPath,
		DefaultReqType:       gencfg.DefaultReqType,
		DefaultCategory:      gencfg.DefaultCategory,
		DefaultTenant:        gencfg.DefaultTenant,
		DefaultTimezone:      gencfg.DefaultTimezone,
		DefaultCaching:       gencfg.DefaultCaching,
		ConnectAttempts:      gencfg.ConnectAttempts,
		Reconnects:           gencfg.Reconnects,
		MaxReconnectInterval: gencfg.MaxReconnectInterval,
		ConnectTimeout:       gencfg.ConnectTimeout,
		ReplyTimeout:         gencfg.ReplyTimeout,
		LockingTimeout:       gencfg.LockingTimeout,
		DigestSeparator:      gencfg.DigestSeparator,
		DigestEqual:          gencfg.DigestEqual,
		MaxParallelConns:     gencfg.MaxParallelConns,
		DecimalMaxScale:      gencfg.DecimalMaxScale,
		DecimalMinScale:      gencfg.DecimalMinScale,
		DecimalPrecision:     gencfg.DecimalPrecision,
		DecimalRoundingMode:  gencfg.DecimalRoundingMode,
		Opts:                 gencfg.Opts.Clone(),
	}
}

type GeneralOptsJson struct {
	ExporterIDs []*DynamicStringSliceOpt `json:"*exporterIDs"`
}

// General config section
type GeneralJsonCfg struct {
	Node_id                *string
	Rounding_decimals      *int
	Dbdata_encoding        *string
	Tpexport_dir           *string
	Default_request_type   *string
	Default_category       *string
	Default_tenant         *string
	Default_timezone       *string
	Default_caching        *string
	Caching_delay          *string
	Connect_attempts       *int
	Reconnects             *int
	Max_reconnect_interval *string
	Connect_timeout        *string
	Reply_timeout          *string
	Locking_timeout        *string
	Digest_separator       *string
	Digest_equal           *string
	Max_parallel_conns     *int

	Decimal_max_scale     *int
	Decimal_min_scale     *int
	Decimal_precision     *int
	Decimal_rounding_mode *string
	Opts                  *GeneralOptsJson
}

func diffGeneralOptsJsonCfg(d *GeneralOptsJson, v1, v2 *GeneralOpts) *GeneralOptsJson {
	if d == nil {
		d = new(GeneralOptsJson)
	}
	if !DynamicStringSliceOptEqual(v1.ExporterIDs, v2.ExporterIDs) {
		d.ExporterIDs = v2.ExporterIDs
	}
	return d
}

func diffGeneralJsonCfg(d *GeneralJsonCfg, v1, v2 *GeneralCfg) *GeneralJsonCfg {
	if d == nil {
		d = new(GeneralJsonCfg)
	}

	if v1.NodeID != v2.NodeID {
		d.Node_id = utils.StringPointer(v2.NodeID)
	}
	if v1.RoundingDecimals != v2.RoundingDecimals {
		d.Rounding_decimals = utils.IntPointer(v2.RoundingDecimals)
	}
	if v1.DBDataEncoding != v2.DBDataEncoding {
		d.Dbdata_encoding = utils.StringPointer(v2.DBDataEncoding)
	}
	if v1.TpExportPath != v2.TpExportPath {
		d.Tpexport_dir = utils.StringPointer(v2.TpExportPath)
	}
	if v1.DefaultReqType != v2.DefaultReqType {
		d.Default_request_type = utils.StringPointer(v2.DefaultReqType)
	}
	if v1.DefaultCategory != v2.DefaultCategory {
		d.Default_category = utils.StringPointer(v2.DefaultCategory)
	}
	if v1.DefaultTenant != v2.DefaultTenant {
		d.Default_tenant = utils.StringPointer(v2.DefaultTenant)
	}
	if v1.DefaultTimezone != v2.DefaultTimezone {
		d.Default_timezone = utils.StringPointer(v2.DefaultTimezone)
	}
	if v1.DefaultCaching != v2.DefaultCaching {
		d.Default_caching = utils.StringPointer(v2.DefaultCaching)
	}
	if v1.ConnectAttempts != v2.ConnectAttempts {
		d.Connect_attempts = utils.IntPointer(v2.ConnectAttempts)
	}
	if v1.Reconnects != v2.Reconnects {
		d.Reconnects = utils.IntPointer(v2.Reconnects)
	}
	if v1.MaxReconnectInterval != v2.MaxReconnectInterval {
		d.Max_reconnect_interval = utils.StringPointer(v2.MaxReconnectInterval.String())
	}
	if v1.ConnectTimeout != v2.ConnectTimeout {
		d.Connect_timeout = utils.StringPointer(v2.ConnectTimeout.String())
	}
	if v1.ReplyTimeout != v2.ReplyTimeout {
		d.Reply_timeout = utils.StringPointer(v2.ReplyTimeout.String())
	}
	if v1.LockingTimeout != v2.LockingTimeout {
		d.Locking_timeout = utils.StringPointer(v2.LockingTimeout.String())
	}
	if v1.DigestSeparator != v2.DigestSeparator {
		d.Digest_separator = utils.StringPointer(v2.DigestSeparator)
	}
	if v1.DigestEqual != v2.DigestEqual {
		d.Digest_equal = utils.StringPointer(v2.DigestEqual)
	}
	if v1.MaxParallelConns != v2.MaxParallelConns {
		d.Max_parallel_conns = utils.IntPointer(v2.MaxParallelConns)
	}
	if v1.DecimalMaxScale != v2.DecimalMaxScale {
		d.Decimal_max_scale = utils.IntPointer(v2.DecimalMaxScale)
	}
	if v1.DecimalMinScale != v2.DecimalMinScale {
		d.Decimal_min_scale = utils.IntPointer(v2.DecimalMinScale)
	}
	if v1.DecimalPrecision != v2.DecimalPrecision {
		d.Decimal_precision = utils.IntPointer(v2.DecimalPrecision)
	}
	if v1.DecimalRoundingMode != v2.DecimalRoundingMode {
		d.Decimal_rounding_mode = utils.StringPointer(v2.DecimalRoundingMode.String())
	}
	d.Opts = diffGeneralOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}
