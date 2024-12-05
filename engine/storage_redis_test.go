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

package engine

import (
	"testing"
	"time"

	"github.com/mediocregopher/radix/v3"
)

func TestNewRedisStorage(t *testing.T) {
	address := "localhost:6379"
	db := 0
	user := ""
	pass := ""
	mrshlerStr := "json"
	maxConns := 10
	attempts := 3
	sentinelName := ""
	isCluster := false
	clusterSync := 1 * time.Second
	clusterOnDownDelay := 1 * time.Second
	connTimeout := 2 * time.Second
	readTimeout := 2 * time.Second
	writeTimeout := 2 * time.Second
	pipelineWindow := 1 * time.Second
	pipelineLimit := 100
	tlsConn := false
	tlsClientCert := ""
	tlsClientKey := ""
	tlsCACert := ""

	redisStorage, err := NewRedisStorage(
		address, db, user, pass, mrshlerStr,
		maxConns, attempts, sentinelName, isCluster,
		clusterSync, clusterOnDownDelay, connTimeout, readTimeout, writeTimeout,
		pipelineWindow, pipelineLimit, tlsConn, tlsClientCert, tlsClientKey, tlsCACert,
	)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if redisStorage == nil {
		t.Errorf("Expected a valid RedisStorage instance, but got nil")
	}

	if redisStorage.client == nil {
		t.Errorf("Expected client to be initialized, but got nil")
	}

	if redisStorage.ms == nil {
		t.Errorf("Expected Marshaler to be initialized, but got nil")
	}
}

func TestRedisStorageIsDBEmpty(t *testing.T) {
	client, err := radix.NewPool("tcp", "localhost:6379", 10)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer client.Close()

	err = client.Do(radix.Cmd(nil, "FLUSHDB"))
	if err != nil {
		t.Fatalf("Failed to flush Redis database: %v", err)
	}

	rs := &RedisStorage{
		client: client,
		ms:     nil,
	}

	resp, err := rs.IsDBEmpty()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if !resp {
		t.Errorf("Expected DB to be empty, but got: %v", resp)
	}

	err = client.Do(radix.Cmd(nil, "SET", "key1", "value1"))
	if err != nil {
		t.Fatalf("Failed to add key to Redis: %v", err)
	}

	resp, err = rs.IsDBEmpty()
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if resp {
		t.Errorf("Expected DB to be non-empty, but got: %v", resp)
	}

	err = client.Do(radix.Cmd(nil, "DEL", "key1"))
	if err != nil {
		t.Fatalf("Failed to delete key from Redis: %v", err)
	}
}

func TestRedisStorageCmd(t *testing.T) {

	address := "localhost:6379"
	db := 0
	user := ""
	pass := ""
	mrshlerStr := "json"
	maxConns := 10
	attempts := 3
	sentinelName := ""
	isCluster := false
	clusterSync := 1 * time.Second
	clusterOnDownDelay := 1 * time.Second
	connTimeout := 2 * time.Second
	readTimeout := 2 * time.Second
	writeTimeout := 2 * time.Second
	pipelineWindow := 1 * time.Second
	pipelineLimit := 100
	tlsConn := false
	tlsClientCert := ""
	tlsClientKey := ""
	tlsCACert := ""

	redisStorage, err := NewRedisStorage(
		address, db, user, pass, mrshlerStr,
		maxConns, attempts, sentinelName, isCluster,
		clusterSync, clusterOnDownDelay, connTimeout, readTimeout, writeTimeout,
		pipelineWindow, pipelineLimit, tlsConn, tlsClientCert, tlsClientKey, tlsCACert,
	)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if redisStorage.client == nil {
		t.Fatalf("Expected client to be initialized, but got nil")
	}

	err = redisStorage.Cmd(nil, "SET", "test_key", "test_value")
	if err != nil {
		t.Errorf("Failed to execute SET command: %v", err)
	}

	var result string
	err = redisStorage.Cmd(&result, "GET", "test_key")
	if err != nil {
		t.Errorf("Failed to execute GET command: %v", err)
	}

	if result != "test_value" {
		t.Errorf("Expected 'test_value', but got '%s'", result)
	}

	var invalidResult string
	err = redisStorage.Cmd(&invalidResult, "INVALID_CMD", "test_key")
	if err == nil {
		t.Error("Expected error for invalid command, but got nil")
	}
}

func TestRedisStorageFlatCmd(t *testing.T) {
	address := "localhost:6379"
	db := 0
	user := ""
	pass := ""
	mrshlerStr := "json"
	maxConns := 10
	attempts := 3
	sentinelName := ""
	isCluster := false
	clusterSync := 1 * time.Second
	clusterOnDownDelay := 1 * time.Second
	connTimeout := 2 * time.Second
	readTimeout := 2 * time.Second
	writeTimeout := 2 * time.Second
	pipelineWindow := 1 * time.Second
	pipelineLimit := 100
	tlsConn := false
	tlsClientCert := ""
	tlsClientKey := ""
	tlsCACert := ""

	redisStorage, err := NewRedisStorage(
		address, db, user, pass, mrshlerStr,
		maxConns, attempts, sentinelName, isCluster,
		clusterSync, clusterOnDownDelay, connTimeout, readTimeout, writeTimeout,
		pipelineWindow, pipelineLimit, tlsConn, tlsClientCert, tlsClientKey, tlsCACert,
	)

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	if redisStorage.client == nil {
		t.Fatalf("Expected client to be initialized, but got nil")
	}

	err = redisStorage.FlatCmd(nil, "SET", "test_key", "test_value")
	if err != nil {
		t.Errorf("Failed to execute SET command: %v", err)
	}

	var result string
	err = redisStorage.FlatCmd(&result, "GET", "test_key")
	if err != nil {
		t.Errorf("Failed to execute GET command: %v", err)
	}

	if result != "test_value" {
		t.Errorf("Expected 'test_value', but got '%s'", result)
	}

	var invalidResult string
	err = redisStorage.FlatCmd(&invalidResult, "INVALID_CMD", "test_key")
	if err == nil {
		t.Error("Expected error for invalid command, but got nil")
	}
}
