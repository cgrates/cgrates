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

package ees

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestS3GetMetrics(t *testing.T) {
	em := &utils.ExporterMetrics{}
	pstr := &S3EE{
		em: em,
	}
	result := pstr.GetMetrics()
	if result == nil {
		t.Errorf("GetMetrics() returned nil; expected a non-nil SafeMapStorage")
		return
	}
	if result != em {
		t.Errorf("GetMetrics() returned unexpected result; got %v, want %v", result, em)
	}
}

func TestClose(t *testing.T) {
	pstr := &S3EE{}
	err := pstr.Close()
	if err != nil {
		t.Errorf("Close() returned an error: %v; expected nil", err)
	}
}

func TestS3Cfg(t *testing.T) {
	expectedCfg := &config.EventExporterCfg{}
	pstr := &S3EE{
		cfg: expectedCfg,
	}
	actualCfg := pstr.Cfg()
	if actualCfg != expectedCfg {
		t.Errorf("Cfg() = %v; expected %v", actualCfg, expectedCfg)
	}
}

func TestParseOptsAllFieldsSet(t *testing.T) {
	bucketID := "bucket"
	folderPath := "folder"
	region := "region"
	key := "id"
	secret := "key"
	token := "token"

	opts := &config.EventExporterOpts{
		AWS: &config.AWSOpts{
			S3BucketID:   &bucketID,
			S3FolderPath: &folderPath,
			Region:       &region,
			Key:          &key,
			Secret:       &secret,
			Token:        &token,
		},
	}

	expected := S3EE{
		bucket:     bucketID,
		folderPath: folderPath,
		awsRegion:  region,
		awsID:      key,
		awsKey:     secret,
		awsToken:   token,
	}

	s3ee := &S3EE{}
	s3ee.parseOpts(opts)
	if s3ee.bucket != expected.bucket {
		t.Errorf("Expected bucket %s, got %s", expected.bucket, s3ee.bucket)
	}
	if s3ee.folderPath != expected.folderPath {
		t.Errorf("Expected folderPath %s, got %s", expected.folderPath, s3ee.folderPath)
	}
	if s3ee.awsRegion != expected.awsRegion {
		t.Errorf("Expected awsRegion %s, got %s", expected.awsRegion, s3ee.awsRegion)
	}
	if s3ee.awsID != expected.awsID {
		t.Errorf("Expected awsID %s, got %s", expected.awsID, s3ee.awsID)
	}
	if s3ee.awsKey != expected.awsKey {
		t.Errorf("Expected awsKey %s, got %s", expected.awsKey, s3ee.awsKey)
	}
	if s3ee.awsToken != expected.awsToken {
		t.Errorf("Expected awsToken %s, got %s", expected.awsToken, s3ee.awsToken)
	}
}
