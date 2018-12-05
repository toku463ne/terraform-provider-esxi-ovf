package ovfdeployer

import (
	"testing"
)

func Test_getXMLVal(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"normal1", args{"<rasd:ResourceType>3</rasd:ResourceType>"}, "3"},
		{"normal2", args{"<rasd:ResourceType> 3 </rasd:ResourceType>"}, " 3 "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getXMLVal(tt.args.line); got != tt.want {
				t.Errorf("getXMLVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getXMLOvfAttr(t *testing.T) {
	type args struct {
		line     string
		attrname string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"test1", args{`<Disk ovf:capacity="32" ovf:capacityAllocationUnits="byte * 2^30" ovf:diskId="vmdisk1" ovf:fileRef="file1" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized" ovf:populatedSize="2052325376"/>`,
			"capacity"}, "32"},
		{"test2", args{`<Disk ovf:capacity="32" ovf:capacityAllocationUnits="byte * 2^30" ovf:diskId="vmdisk1" ovf:fileRef="file1" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized" ovf:populatedSize="2052325376"/>`,
			"capacityAllocationUnits"}, "byte * 2^30"},
		{"test3", args{`<Disk ovf:capacity="32" ovf:capacityAllocationUnits="byte * 2^30" ovf:diskId="vmdisk1" ovf:fileRef="file1" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized" ovf:populatedSize="2052325376"/>`,
			"populatedSize"}, "2052325376"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getXMLOvfAttr(tt.args.line, tt.args.attrname); got != tt.want {
				t.Errorf("getXMLOvfAttr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getXMLTagVal(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"test1", args{"<DiskSection>"}, "DiskSection"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getXMLTagVal(tt.args.line); got != tt.want {
				t.Errorf("getXMLTagVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setXMLVal(t *testing.T) {
	type args struct {
		line string
		val  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"test1", args{"<rasd:VirtualQuantity>2</rasd:VirtualQuantity>", "4"},
			"<rasd:VirtualQuantity>4</rasd:VirtualQuantity>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setXMLVal(tt.args.line, tt.args.val); got != tt.want {
				t.Errorf("setXMLVal() = %v, want %v", got, tt.want)
			}
		})
	}
}
