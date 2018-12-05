package ovfdeployer

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestVM_editOvf(t *testing.T) {
	if err := init4Test("TestVM_editOvf"); err != nil {
		t.Errorf("%+v", err)
	}

	poolid := "ovftest"
	vm, err := createTestVM(poolid, "testvm", true)
	if err != nil {
		t.Errorf("%+v", err)
	}

	type fields struct {
		name     string
		ovffile  string
		memSize  int
		cpuCores int
	}
	tests := []struct {
		name     string
		fields   fields
		wantFile string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{"test1", fields{"test1vm", "test1.ovf", 2048, 4}, "test1want.ovf", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm.name = tt.fields.name
			vm.ovfpath = fmt.Sprintf("%s/ovf/%s", getTestDir(), tt.fields.ovffile)
			vm.memSize = tt.fields.memSize
			vm.cpuCores = tt.fields.cpuCores

			if err := vm.editOvf(); (err != nil) != tt.wantErr {
				t.Errorf("VM.editOvf() error = %v, wantErr %v", err, tt.wantErr)
			}

			wantOvf := fmt.Sprintf("%s/ovf/%s", getTestDir(), tt.wantFile)
			out, _ := exec.Command("diff", vm.ovfpath, wantOvf).Output()
			//if err != nil {
			//	t.Errorf("%+v", err)
			//}
			diffs := strings.Split(string(out), "\n")
			if len(diffs) != 5 {
				t.Errorf("Diff in ovf created(%s) and ovf wanted(%s): %+v %d",
					vm.ovfpath, wantOvf, string(out), len(diffs))
			}
		})
	}
}
