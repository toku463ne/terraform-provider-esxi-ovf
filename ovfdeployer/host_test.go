package ovfdeployer

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func Test_hostFuncs(t *testing.T) {
	if err := init4Test("Test_syncHost"); err != nil {
		t.Errorf("%+v", err)
	}

	testFiles, err := getTestFiles("hosts")
	if err != nil {
		t.Errorf("%+v", err)
	}

	for _, testFile := range testFiles {

		h, err := getOfflineTestHost(testFile)
		if err != nil {
			t.Errorf("%+v", err)
		}

		if err := h.sync2Db(); err != nil {
			t.Errorf("%+v", err)
		}

		h2, err := loadHost(h.poolid, h.ip, h.pass)
		if err != nil {
			t.Errorf("%+v", err)
		}

		a := [][]string{
			{"poolid", h.poolid, h2.poolid},
			{"ip", h.ip, h2.ip},
			{"user", h.user, h2.user},
			{"version", h.version, h2.version},
			{"memTotalMB", h.memTotalMB, h2.memTotalMB},
		}
		for _, s := range a {
			if s[1] != s[2] || s[1] == "" {
				t.Errorf("%s not match. h:%s, h2:%s", s[0], s[1], s[2])
			}
		}
		type intSSlice struct {
			k string
			i int
			j int
		}
		b := []intSSlice{
			{"vmMaxCnt", h.vmMaxCnt, h2.vmMaxCnt},
			{"vmCnt", h.vmCnt, h2.vmCnt},
		}
		for _, s := range b {
			if s.i != s.j || s.i == 0 {
				t.Errorf("%s not match. h:%d, h2:%d", s.k, s.i, s.j)
			}
		}
		type tableSlice struct {
			k string
			i Table
			j Table
		}
		c := []tableSlice{
			{"portGroups", h.portGroups, h2.portGroups},
			{"dsAvailMB", h.dsAvailMB, h2.dsAvailMB},
			{"registeredVMs", h.registeredVMs, h2.registeredVMs},
		}
		for _, s := range c {
			if !reflect.DeepEqual(s.i, s.j) || len(s.i) == 0 {
				t.Errorf("%s not match. h:%v, h2:%v", s.k, s.i, s.j)
			}
		}

		if err := h.registerVM("testid", "testname"); err != nil {
			t.Errorf("%+v", err)
		}

		h2, err = loadHost(h.poolid, h.ip, h.pass)
		if err != nil {
			t.Errorf("%+v", err)
		}

		if h2.registeredVMs["testid"] != "testname" {
			t.Error("h.registerVM(): VM with id=testid is not registered.")
		}

		dbpath := fmt.Sprintf("%s/%s/host_%s.db", workDir, h.poolid, h.ip)
		if fileExists(dbpath) == false {
			t.Errorf("DB file %s does not exist.", dbpath)
		}
		if err := h.deleteHost(); err != nil {
			t.Errorf("%+v", err)
		}
		if fileExists(dbpath) {
			t.Errorf("DB file %s still exists after deleteHost().", dbpath)
		}
	}
}

func Test_newHostOffline(t *testing.T) {
	if err := init4Test("Test_newHostOffline"); err != nil {
		t.Errorf("%+v", err)
	}
	testFiles, err := getTestFiles("hosts")
	if err != nil {
		t.Errorf("%+v", err)
	}

	for _, testFile := range testFiles {

		h, err := getOfflineTestHost(testFile)
		if err != nil {
			t.Errorf("%+v", err)
		}

		v, err := strconv.Atoi(h.dsAvailMB["Disk3"])
		if err != nil {
			t.Errorf("%+v", err)
		}
		v2 := 299.8 * 1024
		v3 := int(v2)
		if v != v3 {
			t.Errorf("getDsInfo() Non expected result. Got=%d Expected=%d",
				v, v3)
		}

		if h.memTotalMB != "16078" {
			t.Errorf("getTotalMem() Non expected result. Got=%s Expected=%d", h.memTotalMB, 16078)
		}

		if h.vmCnt != 6 {
			t.Errorf("getTotalMem() Non expected h.vmCnt. Got=%d Expected=%d", h.vmCnt, 6)
		}

		if h.vmMaxCnt != int(16078/memPerVMMB) {
			t.Errorf("getTotalMem() Non expected h.vmMaxCnt. Got=%d Expected=%d",
				h.vmMaxCnt, int(16078/memPerVMMB))
		}

		if h.registeredVMs["103"] != "mineubuntu" {
			t.Errorf("getTotalMem() Non expected h.registeredVMs. Got=%s Expected=%s",
				h.registeredVMs["103"], "mineubuntu")
		}

		if _, ok := h.portGroups["Fake"]; !ok {
			t.Errorf("getportGroups() Must include portgroup 'Fake'")
		}
	}
}

func Test_newHostOnline(t *testing.T) {
	if isOffLineTest {
		return
	}
	if err := init4Test("Test_newHostOnline"); err != nil {
		t.Errorf("%+v", err)
	}

	h, err := getOnlineTestHost("host")
	if err != nil {
		t.Errorf("%+v", err)
	}

	if len(h.dsAvailMB) == 0 {
		t.Errorf("getDsInfo() Could not get ds info")
	}

	if h.memTotalMB == "" {
		t.Errorf("getTotalMem() Could not get memTotal.")
	}

	if h.vmCnt == 0 {
		t.Errorf("getTotalMem() Could not get vmCnt.")
	}

	if h.vmMaxCnt == 0 {
		t.Errorf("getTotalMem() Could not get vmMaxCnt")
	}

	if len(h.registeredVMs) == 0 {
		t.Errorf("getTotalMem() Could not get registeredVMs")
	}

	if len(h.portGroups) == 0 {
		t.Errorf("getportGroups() Must include portgroup")
	}
}

func Test_host_assertVMDir(t *testing.T) {
	if err := init4Test("Test_host_assertVMDir"); err != nil {
		t.Errorf("%+v", err)
	}
	testFiles, err := getTestFiles("hosts")
	if err != nil {
		t.Errorf("%+v", err)
	}

	for _, testFile := range testFiles {
		h, err := getOfflineTestHost(testFile)
		if err != nil {
			t.Errorf("%+v", err)
		}
		vmdir := fmt.Sprintf("%s/testtools/esxi5/vmfs/volumes/Disk3/testvm", pwd)
		wrongvmdir := fmt.Sprintf("%s/testtools/esxi5/vmfs/volumes/Disk3/wrongvm", pwd)
		type args struct {
			vmname string
			vmdir  string
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			// TODO: Add test cases.
			{"normal", args{"testvm", vmdir}, false},
			{"non-existing vm", args{"testvm2", vmdir}, true},
			{"wrong displayName", args{"wrongvm", wrongvmdir}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := h.assertVMDir(tt.args.vmname, tt.args.vmdir); (err != nil) != tt.wantErr {
					t.Errorf("host.assertVMDir() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	}
}
