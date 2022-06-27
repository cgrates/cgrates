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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

type LoggerCfg struct {
	Type  string
	Level int
	Opts  *LoggerOptsCfg
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
	if loggCfg == nil {
		return nil
	}
	if jsnLoggerCfg.Type != nil && *jsnLoggerCfg.Type != utils.EmptyString {
		loggCfg.Type = *jsnLoggerCfg.Type
	}
	if jsnLoggerCfg.Level != nil {
		loggCfg.Level = *jsnLoggerCfg.Level
	}
	if jsnLoggerCfg.Opts != nil {
		loggCfg.Opts.loadFromJSONCfg(jsnLoggerCfg.Opts)
	}
	return
}

// AsMapInterface returns the config of logger as a map[string]interface{}
func (loggCfg *LoggerCfg) AsMapInterface(string) interface{} {
	return map[string]interface{}{
		utils.TypeCfg:  loggCfg.Type,
		utils.LevelCfg: loggCfg.Level,
		utils.OptsCfg:  loggCfg.Opts.AsMapInterface(),
	}
}

type LoggerOptsCfg struct {
	KafkaConn  string `json:"*kakfa_conn"`
	KafkaTopic string `json:"*kakfa_topic"`
	Attempts   int    `json:"*attempts"`
}

func (LoggerCfg) SName() string                 { return LoggerJSON }
func (loggCfg LoggerCfg) CloneSection() Section { return loggCfg.Clone() }

// Clone returns a deep copy of LoggerCfg
func (loggCfg LoggerCfg) Clone() *LoggerCfg {
	return &LoggerCfg{
		Type:  loggCfg.Type,
		Level: loggCfg.Level,
		Opts:  loggCfg.Opts.Clone(),
	}
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
	if jsnCfg.Attempts != nil {
		loggOpts.Attempts = *jsnCfg.Attempts
	}
}

// AsMapInterface returns the config of logger OPTS as a map[string]interface{}
func (loggOpts *LoggerOptsCfg) AsMapInterface() interface{} {
	return map[string]interface{}{
		utils.KafkaConnCfg:  loggOpts.KafkaConn,
		utils.KafkaTopicCfg: loggOpts.KafkaTopic,
		utils.AttemptsCfg:   loggOpts.Attempts,
	}
}

// Clone returns a deep copy of LoggerOpts
func (loggerOpts *LoggerOptsCfg) Clone() *LoggerOptsCfg {
	if loggerOpts == nil {
		return nil
	}
	return &LoggerOptsCfg{
		KafkaConn:  loggerOpts.KafkaConn,
		KafkaTopic: loggerOpts.KafkaTopic,
		Attempts:   loggerOpts.Attempts,
	}
}

type LoggerJsonCfg struct {
	Type  *string
	Level *int
	Opts  *LoggerOptsJson
}

type LoggerOptsJson struct {
	Kafka_conn  *string
	Kafka_topic *string
	Attempts    *int
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
	if v1.Attempts != v2.Attempts {
		d.Attempts = utils.IntPointer(v2.Attempts)
	}
	return d
}
