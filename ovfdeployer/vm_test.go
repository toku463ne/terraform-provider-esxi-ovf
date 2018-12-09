package ovfdeployer

import (
	"fmt"
	"testing"
)

func Test_DeployVMOnline(t *testing.T) {
	if err := init4Test("Test_DeployVMOnline"); err != nil {
		t.Errorf("%+v", err)
	}
	isDebug = false

	/*
		inf, err := getIniSectionInfo("host")
		if err != nil {
			t.Errorf("%+v", err)
		}
	*/

}

func TestVm_ReserveVMResource(t *testing.T) {
	if err := init4Test("TestVm_ReserveVMResource"); err != nil {
		t.Errorf("%+v", err)
	}
	poolid := "vmtest"
	p := createTestPool(poolid, 4, true, "dummy")
	if p == nil {
		t.Fatal("Could not create pool")
	}

	type fields struct {
		poolid     string
		name       string
		ovfpath    string
		dsSize     int
		memSize    int
		cpuCores   int
		hostIP     string
		datastore  string
		portgroups []string
		guestinfos []string
		pool       Pool
	}

	vm1 := fields{
		poolid:     poolid,
		name:       "testvm1",
		ovfpath:    "",
		dsSize:     30000,
		memSize:    1500,
		cpuCores:   2,
		hostIP:     "",
		datastore:  "",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test1"},
		pool:       *p,
	}

	vm2 := fields{
		poolid:     poolid,
		name:       "testvm2",
		ovfpath:    "",
		dsSize:     30000,
		memSize:    2000,
		cpuCores:   2,
		hostIP:     "",
		datastore:  "",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test2"},
		pool:       *p,
	}

	vm3 := fields{
		poolid:     poolid,
		name:       "testvm3",
		ovfpath:    "",
		dsSize:     30000,
		memSize:    1000,
		cpuCores:   2,
		hostIP:     "",
		datastore:  "",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test3"},
		pool:       *p,
	}

	vm4 := fields{
		poolid:     poolid,
		name:       "testvm4",
		ovfpath:    "",
		dsSize:     300000,
		memSize:    1024,
		cpuCores:   2,
		hostIP:     "",
		datastore:  "",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test4"},
		pool:       *p,
	}
	vm5 := fields{
		poolid:     poolid,
		name:       "testvm5",
		ovfpath:    "",
		dsSize:     20000,
		memSize:    1024,
		cpuCores:   2,
		hostIP:     "1.2.3.1",
		datastore:  "",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test5"},
		pool:       *p,
	}
	vm6 := fields{
		poolid:     poolid,
		name:       "testvm6",
		ovfpath:    "1.2.3.2",
		dsSize:     20000,
		memSize:    1024,
		cpuCores:   2,
		hostIP:     "1.2.3.2",
		datastore:  "Disk4",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test6"},
		pool:       *p,
	}
	vm7 := fields{
		poolid:     poolid,
		name:       "testvm7",
		ovfpath:    "1.2.3.2",
		dsSize:     20000,
		memSize:    4000,
		cpuCores:   2,
		hostIP:     "",
		datastore:  "",
		portgroups: []string{"1.2.3.0"},
		guestinfos: []string{"guestinfo.test7"},
		pool:       *p,
	}
	vm8 := fields{
		poolid:     poolid,
		name:       "testvm8",
		ovfpath:    "",
		dsSize:     3000,
		memSize:    500,
		cpuCores:   1,
		hostIP:     "",
		datastore:  "",
		portgroups: []string{"1.2.3.1"},
		guestinfos: []string{"guestinfo.test8"},
		pool:       *p,
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"vm1", vm1, map[string]string{"hostIP": "1.2.3.4", "ds": "Disk1"}, false},
		{"vm2", vm2, map[string]string{"hostIP": "1.2.3.3", "ds": "Disk1"}, false},
		{"vm3", vm3, map[string]string{"hostIP": "1.2.3.4", "ds": "Disk2"}, false},
		{"vm4", vm4, map[string]string{"hostIP": "", "ds": ""}, true},
		{"vm5", vm5, map[string]string{"hostIP": "1.2.3.1", "ds": "Disk1"}, false},
		{"vm6", vm6, map[string]string{"hostIP": "1.2.3.2", "ds": "Disk4"}, false},
		{"vm7", vm7, map[string]string{"hostIP": "", "ds": ""}, true},
		{"vm8", vm8, map[string]string{"hostIP": "", "ds": ""}, true},
	}
	for _, tt := range tests {
		vm, err := NewVM(poolid,
			tt.fields.name,
			"",
			tt.fields.ovfpath,
			tt.fields.memSize, tt.fields.cpuCores,
			tt.fields.hostIP,
			tt.fields.datastore,
			tt.fields.portgroups,
			tt.fields.guestinfos, currentLogLevelStr)
		if err != nil {
			t.Errorf("%+v", err)
			return
		}
		vm.dsSize = tt.fields.dsSize

		if err := vm.reserveVMResource(); (err != nil) != tt.wantErr {
			t.Errorf("%s Vm.ReserveVMResource() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
		hostIP := vm.hostIP
		ds := vm.datastore
		if hostIP != tt.want["hostIP"] {
			t.Errorf("%s Vm.ReserveVMResource() hostIP not match. got=%s want=%s", tt.name, hostIP, tt.want["hostIP"])
		}
		if ds != tt.want["ds"] {
			t.Errorf("%s Vm.ReserveVMResource() ds not match. got=%s want=%s", tt.name, ds, tt.want["ds"])
		}

	}
}

func TestDeployVM(t *testing.T) {
	if err := init4Test("TestDeployVM"); err != nil {
		t.Errorf("%+v", err)
	}
	poolid := "vmtest"
	testFile := getTestFile("hosts", "esxi5")
	p := createTestPool(poolid, 4, true, testFile)
	if p == nil {
		t.Fatal("Could not create pool")
	}

	type args struct {
		name      string
		ovfname   string
		memSize   int
		cpuCores  int
		hostIP    string
		datastore string
	}
	tests := []struct {
		name       string
		args       args
		wantHostIP string
		wantDs     string
		wantDsSize int
		wantErr    bool
	}{
		// TODO: Add test cases.
		{"test1", args{"testvm", "test1.ovf", 4096, 4, "", ""},
			"1.2.3.4", "Disk1", 32768, false},
		{"test2", args{"testvm", "test1.ovf", 10000, 4, "", ""},
			"1.2.3.4", "Disk1", 32768, true},
	}
	for _, tt := range tests {
		ovfpath := fmt.Sprintf("%s/ovf/%s", getTestDir(), tt.args.ovfname)
		got, err := DeployVM(poolid,
			tt.args.name, "",
			ovfpath, tt.args.memSize,
			tt.args.cpuCores, tt.args.hostIP,
			tt.args.datastore, []string{"1.2.3.0"},
			[]string{"guestinfo.test"}, false, currentLogLevelStr)
		if (err != nil) != tt.wantErr {
			t.Errorf("DeployVM() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if err != nil {
			return
		}
		if got == "" {
			t.Error("DeployVM() VMID must not be nil")
		}

		//id := fmt.Sprintf("%s:%s:%s", poolid, tt.wantHostIP, got)
		_, _, vmname2, _, dsSize2, memPerVMMB2, ds2, err := getVMDBInfo(got)
		if err != nil {
			t.Errorf("%+v", err)
			return
		}
		if tt.args.name != vmname2 {
			t.Errorf("DeployVM() vmname got=%s want=%s", vmname2, tt.args.name)
		}
		if tt.wantDs != ds2 {
			t.Errorf("DeployVM() datastore got=%s want=%s", ds2, tt.wantDs)
		}
		if tt.wantDsSize != dsSize2 {
			t.Errorf("DeployVM() ds size got=%d want=%d", dsSize2, tt.wantDsSize)
		}
		if tt.args.memSize != memPerVMMB2 {
			t.Errorf("DeployVM() mem size got=%d want=%d", memPerVMMB2, tt.args.memSize)
		}

		if err := CheckVMID(got, "", currentLogLevelStr); err != nil {
			t.Errorf("CheckVMID() %+v", err)
		}

		if err := DestroyVM(got, "", currentLogLevelStr); err != nil {
			t.Errorf("DestroyVM() %+v", err)
			return
		}

		if err := CheckVMID(got, "", currentLogLevelStr); err == nil {
			t.Errorf("CheckVMID() VM exists after destroy")
		}
	}
}
