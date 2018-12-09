package ovfdeployer

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPoolFuncs(t *testing.T) {

	type args struct {
		id      string
		hostIPs []string
		user    string
		pass    string
		testID  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"TestPoolFuncs-normal",
			args{
				id:      "pooltest",
				hostIPs: []string{"host1", "host2"},
				user:    "",
				pass:    "",
				testID:  "pool1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := init4Test(tt.name); err != nil {
				t.Errorf("%+v", err)
			}

			got, err := NewPool(tt.args.id, tt.args.hostIPs,
				tt.args.user, tt.args.pass, tt.args.testID, currentLogLevelStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			p2, err := LoadPool(tt.args.id, tt.args.pass, currentLogLevelStr)
			if err != nil {
				t.Errorf("%+v", err)
			}
			if got.ID != p2.ID {
				t.Errorf("got.ID(=%s) and p2.ID(=%s) is not equal.", got.ID, p2.ID)
			}

			if err := AssertPool(tt.args.id, tt.args.pass, got.hostIPs, currentLogLevelStr); err != nil {
				t.Errorf("AssertPool(): %+v", err)
			}

			dummyIPs := append(got.hostIPs, "1.1.1.1")
			if err := AssertPool(tt.args.id, tt.args.pass,
				dummyIPs, currentLogLevelStr); err == nil {
				t.Errorf("AssertPool(): Host members changed from %v to %v. Must return err",
					got.hostIPs, dummyIPs)
			}

			for _, ip := range got.hostIPs {
				h, ok := got.Hosts[ip]
				if !ok {
					continue
				}
				if res, err := h.sshExp.run(cmdGetHostname, "uname -n"); err != nil {
					t.Errorf("%+v", err)
				} else {
					if len(res) == 0 || res[0] != ip {
						t.Errorf("Wrong host registered.")
					}
				}
				if got.hasHost(ip) == false {
					t.Errorf("hasHost(): strange attribute. Must have ip=%s", ip)
				}

				if p2.hasHost(ip) == false {
					t.Errorf("LoadPool(): New pool must have host with ip=%s", ip)
				}

				if got.Hosts[ip].vmCnt != p2.Hosts[ip].vmCnt {
					t.Errorf("LoadPool(): New pool must have host with same vmCnt")
				}

				if err := got.deleteHost(ip); err != nil {
					t.Errorf("%+v", err)
				}
				if got.hasHost(ip) {
					t.Errorf("hasHost(): strange attribute. Must not have ip=%s", ip)
				}
			}
			dbpath := fmt.Sprintf("%s/%s/%s.db", workDir, got.ID, poolDbName)
			if fileExists(dbpath) == false {
				t.Errorf("DB file %s does not exist.", dbpath)
			}

			if err := DeletePool(tt.args.id, tt.args.pass, currentLogLevelStr); err != nil {
				t.Errorf("%+v", err)
			}
			if fileExists(dbpath) {
				t.Errorf("DB file %s still exists after deleteHost().", dbpath)
			}
		})
	}
}

func TestDeletePool(t *testing.T) {

	type args struct {
		id      string
		hostIPs []string
		user    string
		pass    string
		testID  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"TestDeletePool-normal",
			args{
				id:      "pooltest",
				hostIPs: []string{"host1", "host2"},
				user:    "",
				pass:    "",
				testID:  "pool1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := init4Test(tt.name); err != nil {
				t.Errorf("%+v", err)
			}

			got, err := NewPool(tt.args.id, tt.args.hostIPs,
				tt.args.user, tt.args.pass, tt.args.testID, currentLogLevelStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, ip := range got.hostIPs {
				h := got.Hosts[ip]
				dbpath := fmt.Sprintf("%s/%s/%s.db", workDir, got.ID, h.getHostDbName())
				if fileExists(dbpath) == false {
					t.Errorf("DB file %s does not exist.", dbpath)
				}
			}
			dbpath := fmt.Sprintf("%s/%s/%s.db", workDir, got.ID, poolDbName)
			if fileExists(dbpath) == false {
				t.Errorf("DB file %s does not exist.", dbpath)
			}

			if err := DeletePool(tt.args.id, tt.args.pass, currentLogLevelStr); err != nil {
				t.Errorf("%+v", err)
			}
			if fileExists(dbpath) {
				t.Errorf("DB file %s still exists after deleteHost().", dbpath)
			}
			for _, ip := range got.hostIPs {
				h := got.Hosts[ip]
				dbpath := fmt.Sprintf("%s/%s/%s.db", workDir, got.ID, h.getHostDbName())
				if fileExists(dbpath) {
					t.Errorf("DB file %s exists.", dbpath)
				}
			}

		})
	}
}

func TestChangePool(t *testing.T) {
	type args struct {
		id       string
		user     string
		password string
		hostIPs  []string
		hostIPs2 []string
		testID   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"TestChangePool-normal1",
			args{
				id:       "changepooltest1",
				user:     "",
				password: "",
				hostIPs:  []string{"host1", "host2"},
				hostIPs2: []string{"host2"},
				testID:   "pool1"},
			false,
		},
		{"TestChangePool-normal2",
			args{
				id:       "changepooltest2",
				user:     "",
				password: "",
				hostIPs:  []string{"host1", "host2"},
				hostIPs2: []string{"host1", "host3"},
				testID:   "pool1"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := init4Test(tt.name); err != nil {
				t.Errorf("%+v", err)
			}
			_, err := NewPool(tt.args.id, tt.args.hostIPs,
				tt.args.user, tt.args.password, tt.args.testID, currentLogLevelStr)
			if err != nil {
				t.Errorf("%+v", err)
			}

			got, err := ChangePool(tt.args.id, tt.args.user,
				tt.args.password, tt.args.hostIPs2, tt.args.testID, currentLogLevelStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangePool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.hostIPs, tt.args.hostIPs2) {
				t.Errorf("ChangePool() got=%v expected=%v", got.hostIPs, tt.args.hostIPs2)
			}
			lastip := tt.args.hostIPs2[len(tt.args.hostIPs2)-1]
			if h, ok := got.Hosts[lastip]; !ok {
				t.Errorf("ChangePool() new pool must have host with IP=%s", lastip)
			} else if h.ip != lastip {
				t.Errorf("ChangePool() Wrong IP is set to host. got=%s expected=%s", h.ip, lastip)
			}
		})
	}
}

func TestCreateTestPool(t *testing.T) {
	p := createTestPool("test_pool_test", 1, false, "dummy")
	if p == nil {
		t.Errorf("Error creating test pool")
	}
}

func TestPool_getMostVacantHost(t *testing.T) {
	if err := init4Test("TestPool_getMostVacantHost"); err != nil {
		t.Errorf("%+v", err)
	}

	type args struct {
		memSize  int
		dsSize   int
		cpuCores int
	}
	type fields struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		f       fields
		want    string
		want1   int
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1 host: ok", args{2000, 30000, 2}, fields{1}, "1.2.3.1", 4, false},
		{"1 host: ng", args{4000, 30000, 2}, fields{1}, "", -1, true},
		{"1 host: ng2", args{2000, 60000, 2}, fields{1}, "", -1, true},
		{"1 host: ng3", args{100, 1000, 17}, fields{1}, "", -1, true},
		{"2 hosts: ok", args{2000, 30000, 2}, fields{2}, "1.2.3.2", 6, false},
		{"2 hosts: ng", args{6000, 30000, 2}, fields{2}, "", -1, true},
		{"2 hosts: ng2", args{2000, 60000, 2}, fields{2}, "", -1, true},
		{"3 hosts: ok", args{2000, 30000, 2}, fields{3}, "1.2.3.3", 8, false},
		{"3 hosts: ng", args{9000, 30000, 2}, fields{3}, "", -1, true},
		{"3 hosts: ng2", args{2000, 60000, 2}, fields{3}, "", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createTestPool("TestPool_getMostVacantHost", tt.f.n, false, "dummy")

			got, got1, err := p.getMostVacantHost(tt.args.memSize, tt.args.dsSize, tt.args.cpuCores, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Pool.getMostVacantHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Pool.getMostVacantHost() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Pool.getMostVacantHost() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPool_getMostVacantStorage(t *testing.T) {
	type fields struct {
		n int
	}
	type args struct {
		hostIP string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   string
		want2   int
		wantErr bool
	}{
		// TODO: Add test cases.
		{"ok: 1 host. No IP specified", fields{1}, args{""}, "1.2.3.1", "Disk1", 51400, false},
		{"ng: 1 host. IP specified", fields{1}, args{"1.2.3.1"}, "1.2.3.1", "Disk1", 51400, false},
		{"ng: 1 host. uncorrect IP", fields{1}, args{"1.2.3.2"}, "", "", -1, true},
		{"ok: 3 host. No IP specified", fields{3}, args{""}, "1.2.3.2", "Disk1", 52400, false},
		{"ok: 3 host. IP specified", fields{3}, args{"1.2.3.1"}, "1.2.3.1", "Disk1", 51400, false},
		{"ng: 3 host. uncorrect IP specified", fields{3}, args{"1.2.3.4"}, "", "", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createTestPool("test", tt.fields.n, false, "dummy")
			got, got1, got2, err := p.getMostVacantStorage(tt.args.hostIP)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pool.getMostVacantStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Pool.getMostVacantStorage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Pool.getMostVacantStorage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Pool.getMostVacantStorage() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestPool_appendVMResource(t *testing.T) {
	poolid := "testp"
	p := createTestPool(poolid, 4, false, "dummy")

	type fields struct {
		ID      string
		Hosts   hostgroup
		hostIPs []string
	}
	type args struct {
		hostIP   string
		dsSize   int
		memSize  int
		cpuCores int
		ds       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"Normal pattern", fields{poolid, p.Hosts, p.hostIPs},
			args{"", 30000, 2048, 2, ""},
			"1.2.3.4", "Disk1", false},
		{"Must choose host with largest storage", fields{poolid, p.Hosts, p.hostIPs},
			args{"", 52000, 2048, 2, ""},
			"1.2.3.2", "Disk1", false},
		{"No host with this big storage", fields{poolid, p.Hosts, p.hostIPs},
			args{"", 62000, 2048, 2, ""},
			"", "", true},
		{"Requires too much memory", fields{poolid, p.Hosts, p.hostIPs},
			args{"", 30000, 20000, 2, ""},
			"", "", true},
		{"Too may CPU requirements", fields{poolid, p.Hosts, p.hostIPs},
			args{"", 8000, 100, 17, ""},
			"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				ID:      tt.fields.ID,
				Hosts:   tt.fields.Hosts,
				hostIPs: tt.fields.hostIPs,
			}
			got, got1, err := p.appendVMResource(tt.args.hostIP,
				tt.args.dsSize, tt.args.memSize, tt.args.cpuCores, tt.args.ds, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Pool.appendVMResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Pool.appendVMResource() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Pool.appendVMResource() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
