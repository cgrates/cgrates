// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestLoadLoggerCfg(t *testing.T) {
	loggCfg := &LoggerCfg{}
	ctx := &context.Context{}
	jsnCfg := new(mockDb)
	cgrcfg := &CGRConfig{}
	if err := loggCfg.Load(ctx, jsnCfg, cgrcfg); err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}
}

func TestLoadFromJSONCfgLoggerCfg(t *testing.T) {
	loggCfg := &LoggerCfg{}

	jsnLoggerCfg := &LoggerJsonCfg{
		Type: utils.StringPointer("testType"),
	}

	exp := &LoggerCfg{
		Type: "testType",
	}

	if err := loggCfg.loadFromJSONCfg(jsnLoggerCfg); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", nil, err)
	}

	if !reflect.DeepEqual(exp, loggCfg) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(loggCfg))
	}

	//check with empty json config
	jsnLoggerCfg = nil
	if err := loggCfg.loadFromJSONCfg(jsnLoggerCfg); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", nil, err)
	}
}

func TestLoggerCfgCloneSection(t *testing.T) {
	loggCfg := &LoggerCfg{
		Type:     "testType",
		Level:    2,
		EFsConns: []string{"efsConnsTest"},
		Opts: &LoggerOptsCfg{
			KafkaConn: "TestKafkaConn",
		},
	}

	rcv := loggCfg.CloneSection()
	if !reflect.DeepEqual(loggCfg, rcv.(*LoggerCfg)) {
		t.Errorf("Expected %v \n but received \n %v", loggCfg, rcv.(*LoggerCfg))
	}
}

func TestLoadFromJSONCfgLoggerOptsCfg(t *testing.T) {
	loggOpts := &LoggerOptsCfg{}

	jsnCfg := &LoggerOptsJson{
		Kafka_conn: utils.StringPointer("Kafka_connTest"),
	}

	exp := &LoggerOptsCfg{
		KafkaConn: "Kafka_connTest",
	}

	loggOpts.loadFromJSONCfg(jsnCfg)

	if !reflect.DeepEqual(exp, loggOpts) {
		t.Errorf("Expected %v \n but received \n %v",
			utils.ToJSON(exp), utils.ToJSON(loggOpts))
	}

	//check with empty json config
	jsnCfg = nil
	loggOpts.loadFromJSONCfg(jsnCfg)
	if !reflect.DeepEqual(exp, loggOpts) {
		t.Errorf("Expected %v \n but received \n %v",
			utils.ToJSON(exp), utils.ToJSON(loggOpts))
	}
}

func TestLoggerOptsCfgCloneNil(t *testing.T) {

	var loggerOpts *LoggerOptsCfg

	if rcv := loggerOpts.Clone(); rcv != nil {
		t.Errorf("Expected to return <nil>, Received <%v>", rcv)
	}

}

func TestDiffLoggerJsonCfg(t *testing.T) {
	var d *LoggerJsonCfg

	v1 := &LoggerCfg{
		Type:     "testTypev1",
		Level:    1,
		EFsConns: []string{"efsConnsTestv1"},
		Opts: &LoggerOptsCfg{
			KafkaConn: "TestKafkaConnv1",
		},
	}

	v2 := &LoggerCfg{
		Type:     "testType",
		Level:    2,
		EFsConns: []string{"efsConnsTest"},
		Opts: &LoggerOptsCfg{
			KafkaConn: "TestKafkaConn",
		},
	}

	expected := &LoggerJsonCfg{
		Type:      utils.StringPointer("testType"),
		Level:     utils.IntPointer(2),
		Efs_conns: utils.SliceStringPointer([]string{"efsConnsTest"}),
		Opts: &LoggerOptsJson{
			Kafka_conn: utils.StringPointer("TestKafkaConn"),
		},
	}

	rcv := diffLoggerJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &LoggerJsonCfg{
		Opts: &LoggerOptsJson{},
	}
	rcv = diffLoggerJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDiffLoggerOptsJsonCfg(t *testing.T) {
	var d *LoggerOptsJson

	v1 := &LoggerOptsCfg{
		KafkaConn:      "TestKafkaConnv1",
		KafkaTopic:     "testKafkaTopicV1",
		KafkaAttempts:  1,
		FailedPostsDir: "TestKafkaPostsDirV1",
	}

	v2 := &LoggerOptsCfg{

		KafkaConn:      "TestKafkaConnv2",
		KafkaTopic:     "testKafkaTopicV2",
		KafkaAttempts:  2,
		FailedPostsDir: "TestKafkaPostsDirV2",
	}

	expected := &LoggerOptsJson{
		Kafka_conn:       utils.StringPointer("TestKafkaConnv2"),
		Kafka_topic:      utils.StringPointer("testKafkaTopicV2"),
		Kafka_attempts:   utils.IntPointer(2),
		Failed_posts_dir: utils.StringPointer("TestKafkaPostsDirV2"),
	}

	rcv := diffLoggerOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &LoggerOptsJson{}
	rcv = diffLoggerOptsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
