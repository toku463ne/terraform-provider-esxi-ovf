package ovfdeployer

import "testing"

func Test_init4Test(t *testing.T) {
	type args struct {
		testname string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test with args", args{"testarg"}, false},
		{"test without args", args{""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := init4Test(tt.args.testname); (err != nil) != tt.wantErr {
				t.Errorf("init4Test() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
