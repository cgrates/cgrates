// +build integration

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

package loaders

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTestItLoaders = []func(t *testing.T){
		testV1LoadResource,
		testV1LoadDefaultIDError,
		testV1LoadUnableToDeleteFile,
		testV1LoadProcessFolderError,
		testV1RemoveResource,
		testV1RemoveDefaultIDError,
		testV1RemoveUnableToDeleteFile,
		testV1RemoveProcessFolderError,
		testV1LoadAndRemoveProcessRemoveFolderError,
		testLoaderServiceReload,
		testLoaderServiceListenAndServe,
	}
)

func TestITLoaders(t *testing.T) {
	for _, test := range sTestItLoaders {
		t.Run("Loaders_IT_Tests", test)
	}
}

func testV1LoadResource(t *testing.T) {
	flPath := "/tmp/testV1LoadResource"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1LoadResource",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	ldrs.ldrs["testV1LoadResource"].dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
		},
	}

	resCsv := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := ioutil.NopCloser(strings.NewReader(resCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldrs.ldrs["testV1LoadResource"].rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{
				fileName: utils.ResourcesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}

	var reply string
	expected := "ANOTHER_LOADER_RUNNING"
	//cannot load when there is another loader running
	if err := ldrs.V1Load(&ArgsProcessFolder{
		LoaderID:  "testV1LoadResource",
		ForceLock: false}, &reply); err == nil || reply != utils.EmptyString || err.Error() != expected {
		t.Errorf("Expected %+v and %+v \n, received %+v and %+v", expected, utils.EmptyString, err, reply)
	}

	if err := ldrs.V1Load(&ArgsProcessFolder{
		LoaderID:  "testV1LoadResource",
		ForceLock: true}, &reply); err != nil && reply != utils.OK {
		t.Error(err)
	}

	expRes := &engine.ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "NewRes1",
	}

	if rcv, err := ldrs.ldrs["testV1LoadResource"].dm.GetResourceProfile(expRes.Tenant, expRes.ID,
		true, true, utils.NonTransactional); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expRes) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRes), utils.ToJSON(rcv))
	}

	if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1LoadDefaultIDError(t *testing.T) {
	flPath := "/tmp/testV1LoadResource"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1LoadDefaultIDError",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}

	var reply string
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	if err := ldrs.V1Load(&ArgsProcessFolder{
		LoaderID: utils.EmptyString}, &reply); err == nil && reply != utils.EmptyString && err.Error() != utils.EmptyString {
		t.Errorf("Expected %+v and %+v \n, received %+v and %+v", utils.EmptyString, utils.EmptyString, err, reply)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1LoadUnableToDeleteFile(t *testing.T) {
	flPath := "testV1LoadUnableToDeleteFile"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	_, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1LoadUnableToDeleteFile",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        "/\x00",
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}
	var reply string
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	expected := "SERVER_ERROR: stat /\x00/Resources.csv: invalid argument"
	if err := ldrs.V1Load(&ArgsProcessFolder{
		LoaderID:  "testV1LoadUnableToDeleteFile",
		ForceLock: true}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v and %+v \n, received %+v and %+v", utils.EmptyString, utils.EmptyString, err, reply)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1LoadProcessFolderError(t *testing.T) {
	flPath := "testV1LoadProcessFolderError"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#PK
NOT_UINT
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1LoadResource",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	ldrs.ldrs["testV1LoadResource"].dataTpls = map[string][]*config.FCTemplate{
		utils.MetaFilters: {
			{Tag: "PK",
				Path:  "PK",
				Type:  utils.META_COMPOSED,
				Value: config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP)},
		},
	}
	resCsv := `
//PK
NOT_UINT
`
	rdr := ioutil.NopCloser(strings.NewReader(resCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldrs.ldrs["testV1LoadResource"].rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"not_a_file": &openedCSVFile{
				fileName: utils.ResourcesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}

	var reply string
	expected := "SERVER_ERROR: open testV1LoadProcessFolderError/not_a_file: no such file or directory"
	//try to load by changing the caching method
	if err := ldrs.V1Load(&ArgsProcessFolder{
		LoaderID:    "testV1LoadResource",
		ForceLock:   true,
		Caching:     utils.StringPointer("not_a_chaching_method"),
		StopOnError: true}, &reply); err == nil || err.Error() != expected || reply != utils.EmptyString {
		t.Errorf("Expected %+q and %+q \n, received %+q and %+q", expected, utils.EmptyString, err, reply)
	}

	if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1RemoveResource(t *testing.T) {
	flPath := "/tmp/testV1RemoveResource"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1RemoveResource",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	ldrs.ldrs["testV1RemoveResource"].dataTpls = map[string][]*config.FCTemplate{
		utils.MetaResources: {
			{Tag: "Tenant",
				Path:      "Tenant",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.0", utils.INFIELD_SEP),
				Mandatory: true},
			{Tag: "ID",
				Path:      "ID",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~*req.1", utils.INFIELD_SEP),
				Mandatory: true},
		},
	}

	resCsv := `
#Tenant[0],ID[1]
cgrates.org,NewRes1
`
	rdr := ioutil.NopCloser(strings.NewReader(resCsv))
	csvRdr := csv.NewReader(rdr)
	csvRdr.Comment = '#'
	ldrs.ldrs["testV1RemoveResource"].rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			utils.ResourcesCsv: &openedCSVFile{
				fileName: utils.ResourcesCsv,
				rdr:      rdr,
				csvRdr:   csvRdr,
			},
		},
	}

	expRes := &engine.ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "NewRes1",
	}
	//To remove a resource, we need to set it first
	if err := ldrs.ldrs["testV1RemoveResource"].dm.SetResourceProfile(expRes, true); err != nil {
		t.Error(expRes)
	}

	var reply string
	expected := "ANOTHER_LOADER_RUNNING"
	//cannot load when there is another loader running
	if err := ldrs.V1Remove(&ArgsProcessFolder{
		LoaderID:  "testV1RemoveResource",
		ForceLock: false}, &reply); err == nil || reply != utils.EmptyString || err.Error() != expected {
		t.Errorf("Expected %+v and %+v \n, received %+v and %+v", expected, utils.EmptyString, err, reply)
	}

	ldrs.ldrs["testV1RemoveResource"].lockFilename = "invalidFile"
	if err := ldrs.V1Remove(&ArgsProcessFolder{
		LoaderID:  "testV1RemoveResource",
		ForceLock: true}, &reply); err != nil && reply != utils.OK {
		t.Error(err)
	}

	//nothing to get from dataBase
	if _, err := ldrs.ldrs["testV1RemoveResource"].dm.GetResourceProfile(expRes.Tenant, expRes.ID,
		true, true, utils.NonTransactional); err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1RemoveDefaultIDError(t *testing.T) {
	flPath := "/tmp/testV1RemoveDefaultIDError"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	file, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}
	file.Write([]byte(`
#Tenant[0],ID[1]
cgrates.org,NewRes1
`))
	file.Close()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1RemoveDefaultIDError",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}

	var reply string
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	expected := "UNKNOWN_LOADER: *default"
	if err := ldrs.V1Remove(&ArgsProcessFolder{
		LoaderID: utils.EmptyString}, &reply); err == nil || reply != utils.EmptyString || err.Error() != expected {
		t.Errorf("Expected %+v and %+v \n, received %+v and %+v", expected, utils.EmptyString, err, reply)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1RemoveUnableToDeleteFile(t *testing.T) {
	flPath := "testV1RemoveUnableToDeleteFile"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	_, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1RemoveUnableToDeleteFile",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        "/\x00",
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}
	var reply string
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	expected := "SERVER_ERROR: stat /\x00/Resources.csv: invalid argument"
	if err := ldrs.V1Remove(&ArgsProcessFolder{
		LoaderID:  "testV1RemoveUnableToDeleteFile",
		ForceLock: true}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v and %+v \n, received %+v and %+v", utils.EmptyString, utils.EmptyString, err, reply)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1LoadAndRemoveProcessRemoveFolderError(t *testing.T) {
	flPath := "/tmp/testV1RemoveProcessFolderError"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	_, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1RemoveProcessFolderError",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		Data:           nil,
	}
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)

	ldrs.ldrs["testV1RemoveProcessFolderError"].rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"not_a_file": &openedCSVFile{
				fileName: utils.ResourcesCsv,
			},
		},
	}

	var reply string
	expected := "SERVER_ERROR: remove /tmp/testV1RemoveProcessFolderError: directory not empty"
	//try to load by changing the caching method, but there is not a lockFileName
	if err := ldrs.V1Load(&ArgsProcessFolder{
		LoaderID:    "testV1RemoveProcessFolderError",
		ForceLock:   true,
		Caching:     utils.StringPointer("not_a_chaching_method"),
		StopOnError: true}, &reply); err == nil || err.Error() != expected || reply != utils.EmptyString {
		t.Errorf("Expected %+q and %+q \n, received %+q and %+q", expected, utils.EmptyString, err, reply)
	}

	//try to remove by changing the caching method
	if err := ldrs.V1Remove(&ArgsProcessFolder{
		LoaderID:    "testV1RemoveProcessFolderError",
		ForceLock:   true,
		Caching:     utils.StringPointer("not_a_chaching_method"),
		StopOnError: true}, &reply); err == nil || err.Error() != expected || reply != utils.EmptyString {
		t.Errorf("Expected %+q and %+q \n, received %+q and %+q", expected, utils.EmptyString, err, reply)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testV1RemoveProcessFolderError(t *testing.T) {
	flPath := "testV1RemoveProcessFolderError"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	_, err := os.Create(path.Join(flPath, utils.ResourcesCsv))
	if err != nil {
		t.Error(err)
	}

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1RemoveProcessFolderError",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   "notResource.csv",
		Data:           nil,
	}
	ldrs := NewLoaderService(dm, cfgLdr, "UTC", nil, nil)
	ldrs.ldrs["testV1RemoveProcessFolderError"].rdrs = map[string]map[string]*openedCSVFile{
		utils.MetaResources: {
			"not_a_file2": &openedCSVFile{
				fileName: utils.ResourcesCsv,
			},
		},
	}

	var reply string
	expected := "SERVER_ERROR: open testV1RemoveProcessFolderError/not_a_file2: no such file or directory"
	//try to load by changing the caching method
	if err := ldrs.V1Remove(&ArgsProcessFolder{
		LoaderID:    "testV1RemoveProcessFolderError",
		ForceLock:   true,
		Caching:     utils.StringPointer("not_a_chaching_method"),
		StopOnError: true}, &reply); err == nil || err.Error() != expected || reply != utils.EmptyString {
		t.Errorf("Expected %+q and %+q \n, received %+q and %+q", expected, utils.EmptyString, err, reply)
	}

	if err := os.Remove(path.Join(flPath, utils.ResourcesCsv)); err != nil {
		t.Error(err)
	} else if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testLoaderServiceListenAndServe(t *testing.T) {
	ldr := &Loader{
		runDelay: -1,
		tpInDir:  "/tmp/TestLoaderServiceListenAndServe",
	}
	ldrs := &LoaderService{
		ldrs: map[string]*Loader{
			"TEST_LOADER": ldr,
		},
	}
	stopChan := make(chan struct{}, 1)
	stopChan <- struct{}{}
	expected := "no such file or directory"
	if err := ldrs.ListenAndServe(stopChan); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	ldrs.ldrs["TEST_LOADER"].tpInDir = utils.EmptyString
	if err := ldrs.ListenAndServe(stopChan); err != nil {
		t.Error(err)
	}
}

func testLoaderServiceReload(t *testing.T) {
	flPath := "/tmp/testLoaderServiceReload"
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfgLdr := config.NewDefaultCGRConfig().LoaderCfg()
	cfgLdr[0] = &config.LoaderSCfg{
		ID:             "testV1LoadResource",
		Enabled:        true,
		FieldSeparator: utils.FIELDS_SEP,
		TpInDir:        flPath,
		TpOutDir:       "/tmp",
		LockFileName:   utils.ResourcesCsv,
		Data:           nil,
	}
	ldrs := &LoaderService{}
	ldrs.Reload(dm, cfgLdr, "UTC", nil, nil)
	if ldrs.ldrs == nil {
		t.Error("Expected to be populated")
	}
}
