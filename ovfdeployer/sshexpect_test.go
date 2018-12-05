package ovfdeployer

import (
	"testing"
)

func Test_sshExpect_login(t *testing.T) {
	inf, err := getIniSectionInfo("localhost")
	if err != nil {
		t.Errorf("%+v", err)
	}
	hostIP := "localhost"
	user := inf["user"]
	password := inf["password"]

	tests := []struct {
		name     string
		password string
		testFile string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{"login 1st time", password, "", false},
		{"login 2nd time", password, "", false},
		{"Wrong password", "wrong_password", "", true},
	}
	for _, tt := range tests {
		se := newSSHExpect(hostIP, user, tt.password, tt.testFile)
		t.Run(tt.name, func(t *testing.T) {
			if err := se.login(); (err != nil) != tt.wantErr {
				t.Errorf("sshExpect.login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sshExpect_run(t *testing.T) {
	inf, err := getIniSectionInfo("localhost")
	if err != nil {
		t.Errorf("%+v", err)
	}
	hostIP := "localhost"
	user := inf["user"]
	password := inf["password"]
	se := newSSHExpect(hostIP, user, password, "")
	if err := se.login(); err != nil {
		t.Errorf("%+v", err)
	}

	type args struct {
		cmdName string
		cmd     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{"normal", args{"", "echo test"}, "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := se.run(tt.args.cmdName, tt.args.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("sshExpect.run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got[0] != tt.want {
				t.Errorf("sshExpect.run() = %v, want %v", got, tt.want)
			}
		})
	}
	se.close()
}
