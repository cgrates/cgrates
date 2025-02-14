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
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type LoggerCfg struct {
	Type     string
	Level    int
	EFsConns []string
	Opts     *LoggerOptsCfg
}

// Load loads the Logger section of the configuration
func (loggCfg *LoggerCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnLoggerCfg := new(LoggerJsonCfg)
	if err = jsnCfg.GetSection(ctx, LoggerJSON, jsnLoggerCfg); err != nil {
		return
	}
	return loggCfg.loadFromJSONCfg(jsnLoggerCfg)
}

// loadFromJSONCfg loads Logger config from JsonCfg
func (loggCfg *LoggerCfg) loadFromJSONCfg(jsnLoggerCfg *LoggerJsonCfg) (err error) {
	if jsnLoggerCfg == nil {
		return nil
	}
	if jsnLoggerCfg.Type != nil && *jsnLoggerCfg.Type != utils.EmptyString {
		loggCfg.Type = *jsnLoggerCfg.Type
	}
	if jsnLoggerCfg.Level != nil {
		loggCfg.Level = *jsnLoggerCfg.Level
	}
	if jsnLoggerCfg.Efs_conns != nil {
		loggCfg.EFsConns = updateInternalConns(*jsnLoggerCfg.Efs_conns, utils.MetaEFs)
	}
	if jsnLoggerCfg.Opts != nil {
		loggCfg.Opts.loadFromJSONCfg(jsnLoggerCfg.Opts)
	}
	return
}

// AsMapInterface returns the config of logger as a map[string]any
func (loggCfg *LoggerCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.TypeCfg:  loggCfg.Type,
		utils.LevelCfg: loggCfg.Level,
		utils.OptsCfg:  loggCfg.Opts.AsMapInterface(),
	}
	if loggCfg.EFsConns != nil {
		mp[utils.EFsConnsCfg] = getInternalJSONConns(loggCfg.EFsConns)
	}
	return mp
}

type LoggerOptsCfg struct {
	KafkaConn      string
	KafkaTopic     string
	KafkaAttempts  int
	FailedPostsDir string
}

func (LoggerCfg) SName() string                 { return LoggerJSON }
func (loggCfg LoggerCfg) CloneSection() Section { return loggCfg.Clone() }

// Clone returns a deep copy of LoggerCfg
func (loggCfg LoggerCfg) Clone() *LoggerCfg {
	cln := &LoggerCfg{
		Type:  loggCfg.Type,
		Level: loggCfg.Level,
		Opts:  loggCfg.Opts.Clone(),
	}
	if loggCfg.EFsConns != nil {
		cln.EFsConns = *utils.SliceStringPointer(loggCfg.EFsConns)
	}
	return cln
}

// loadFromJSONCfg loads Logger opts config from JsonCfg
func (loggOpts *LoggerOptsCfg) loadFromJSONCfg(jsnCfg *LoggerOptsJson) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Kafka_conn != nil && *jsnCfg.Kafka_conn != utils.EmptyString {
		loggOpts.KafkaConn = *jsnCfg.Kafka_conn
	}
	if jsnCfg.Kafka_topic != nil {
		loggOpts.KafkaTopic = *jsnCfg.Kafka_topic
	}
	if jsnCfg.Kafka_attempts != nil {
		loggOpts.KafkaAttempts = *jsnCfg.Kafka_attempts
	}
	if jsnCfg.Failed_posts_dir != nil {
		loggOpts.FailedPostsDir = *jsnCfg.Failed_posts_dir
	}
}

// AsMapInterface returns the config of logger OPTS as a map[string]any
func (loggOpts *LoggerOptsCfg) AsMapInterface() any {
	return map[string]any{
		utils.KafkaConnCfg:      loggOpts.KafkaConn,
		utils.KafkaTopicCfg:     loggOpts.KafkaTopic,
		utils.KafkaAttemptsCfg:  loggOpts.KafkaAttempts,
		utils.FailedPostsDirCfg: loggOpts.FailedPostsDir,
	}
}

// Clone returns a deep copy of LoggerOpts
func (loggerOpts *LoggerOptsCfg) Clone() *LoggerOptsCfg {
	if loggerOpts == nil {
		return nil
	}
	return &LoggerOptsCfg{
		KafkaConn:      loggerOpts.KafkaConn,
		KafkaTopic:     loggerOpts.KafkaTopic,
		KafkaAttempts:  loggerOpts.KafkaAttempts,
		FailedPostsDir: loggerOpts.FailedPostsDir,
	}
}

type LoggerJsonCfg struct {
	Type      *string
	Level     *int
	Efs_conns *[]string
	Opts      *LoggerOptsJson
}

type LoggerOptsJson struct {
	Kafka_conn       *string `json:"kafka_conn"`
	Kafka_topic      *string `json:"kafka_topic"`
	Kafka_attempts   *int    `json:"kafka_attempts"`
	Failed_posts_dir *string `json:"failed_posts_dir"`
}

func diffLoggerJsonCfg(d *LoggerJsonCfg, v1, v2 *LoggerCfg) *LoggerJsonCfg {
	if d == nil {
		d = new(LoggerJsonCfg)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.Level != v2.Level {
		d.Level = utils.IntPointer(v2.Level)
	}
	if !slices.Equal(v1.EFsConns, v2.EFsConns) {
		d.Efs_conns = utils.SliceStringPointer(getInternalJSONConns(v2.EFsConns))
	}
	d.Opts = diffLoggerOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	return d
}

func diffLoggerOptsJsonCfg(d *LoggerOptsJson, v1, v2 *LoggerOptsCfg) *LoggerOptsJson {
	if d == nil {
		d = new(LoggerOptsJson)
	}
	if v1.KafkaConn != v2.KafkaConn {
		d.Kafka_conn = utils.StringPointer(v2.KafkaConn)
	}
	if v1.KafkaTopic != v2.KafkaTopic {
		d.Kafka_topic = utils.StringPointer(v2.KafkaTopic)
	}
	if v1.KafkaAttempts != v2.KafkaAttempts {
		d.Kafka_attempts = utils.IntPointer(v2.KafkaAttempts)
	}
	if v1.FailedPostsDir != v2.FailedPostsDir {
		d.Failed_posts_dir = utils.StringPointer(v2.FailedPostsDir)
	}
	return d
}
