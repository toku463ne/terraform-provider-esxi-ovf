package ovfdeployer

import (
	"reflect"
	"testing"
)

func Test_sortIPSlice(t *testing.T) {
	type args struct {
		ips []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{"pattern1",
			args{
				[]string{"1.2.3.4", "1.2.3.10", "1.2.3.2", "1.2.3.1"},
			},
			[]string{"1.2.3.1", "1.2.3.2", "1.2.3.4", "1.2.3.10"},
		},
		{"pattern2",
			args{
				[]string{"1.2.13.4", "1.2.14.10", "1.2.12.2", "1.2.3.1"},
			},
			[]string{"1.2.3.1", "1.2.12.2", "1.2.13.4", "1.2.14.10"},
		},
		{"pattern3",
			args{
				[]string{"1.6.1.1", "1.2.1.2", "1.2.1.3", "1.1.1.4"},
			},
			[]string{"1.1.1.4", "1.2.1.2", "1.2.1.3", "1.6.1.1"},
		},
		{"pattern4",
			args{
				[]string{"254.255.255.254", "254.254.254.255"},
			},
			[]string{"254.254.254.255", "254.255.255.254"},
		},
		{"pattern5",
			args{
				[]string{"host2", "host1"},
			},
			[]string{"host1", "host2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortIPSlice(tt.args.ips)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortIPSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkID(t *testing.T) {
	type args struct {
		poolid string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{"normal1", args{"test"}, true},
		{"normal2", args{"Test1"}, true},
		{"normal3", args{"test-1"}, true},
		{"normal4", args{"test_1"}, true},
		{"normal5", args{"test.1"}, true},
		{"normal5", args{"test 1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkID(tt.args.poolid); got != tt.want {
				t.Errorf("checkID() = %v, want %v", got, tt.want)
			}
		})
	}
}
