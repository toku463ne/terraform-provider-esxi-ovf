package ovfdeployer

import (
	"strconv"
	"testing"
)

func Test_HostInfoOffLine(t *testing.T) {
	if err := init4Test("Test_hostInfoOffLine"); err != nil {
		t.Errorf("%+v", err)
	}

	poolid := "testHostInfo"
	h := new(host)
	h.poolid = poolid

	testFiles, err := getTestFiles("hosts")
	if err != nil {
		t.Errorf("%+v", err)
	}

	for _, testFile := range testFiles {
		se, err := getSSHExpectConn("", "", "", testFile)
		if err != nil {
			t.Errorf("%+v", err)
		}
		h.sshExp = *se

		err = h.getDsInfo()
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

		err = h.getTotalMem()
		if err != nil {
			t.Errorf("%+v", err)
		}
		if h.memTotalMB != 16078 {
			t.Errorf("getTotalMem() Non expected result. Got=%d Expected=%d", h.memTotalMB, 16078)
		}
		err = h.getCPUCores()
		if err != nil {
			t.Errorf("%+v", err)
		}
		if h.cpuCoresCnt != 8 {
			t.Errorf("getCPUCores() Non expected result. Got=%d Expected=%d", h.cpuCoresCnt, 8)
		}

		err = h.getVMInfo()
		if err != nil {
			t.Errorf("%+v", err)
		}

		if h.vmCnt != 6 {
			t.Errorf("getVMInfo() Non expected h.vmCnt. Got=%d Expected=%d", h.vmCnt, 6)
		}

		if h.vmMaxCnt != int(16078/memPerVMMB) {
			t.Errorf("getVMInfo() Non expected h.vmMaxCnt. Got=%d Expected=%d",
				h.vmMaxCnt, int(16078/memPerVMMB))
		}

		if h.registeredVMs["103"] != "mineubuntu" {
			t.Errorf("getVMInfo() Non expected h.registeredVMs. Got=%s Expected=%s",
				h.registeredVMs["103"], "mineubuntu")
		}

		if h.memActiveMB != 27328 {
			t.Errorf("getVMInfo() Non expected h.memActiveMB. Got=%d Expected=%d",
				h.memActiveMB, 27328)
		}

		vmid, err := h.getVMId("mineubuntu")
		if err != nil {
			t.Errorf("%+v", err)
		}
		if vmid != "103" {
			t.Errorf("getVMId() Non expected vmid. Got=%s Expected=103", vmid)
		}

		vmname, err := h.getVMName(vmid)
		if err != nil {
			t.Errorf("%+v", err)
		}
		if vmname != "testvm" {
			t.Errorf("getVMName() Non expected vmname. Got=%s Expected=mineubuntu", vmname)
		}

		if err := h.getPortGroups(); err != nil {
			t.Errorf("%+v", err)
		}
		if _, ok := h.portGroups["Fake"]; !ok {
			t.Errorf("getPortGroups() Must include portgroup 'Fake'")
		}
	}
}

func Test_hostInfoOnLine(t *testing.T) {
	if isOffLineTest {
		return
	}

	if err := init4Test("Test_hostInfoOnLine"); err != nil {
		t.Errorf("%+v", err)
	}

	inf, err := getIniSectionInfo("host")
	if err != nil {
		t.Errorf("%+v", err)
	}
	poolid := "testHostInfo"
	hostIP := inf["ipaddr"]
	user := inf["user"]
	pass := inf["password"]
	h := new(host)
	h.poolid = poolid
	h.ip = hostIP
	h.user = user
	h.pass = pass
	se, err := getSSHExpectConn(hostIP, user, pass, "")
	if err != nil {
		t.Errorf("%+v", err)
	}
	h.sshExp = *se

	err = h.getDsInfo()
	if err != nil {
		t.Errorf("%+v", err)
	}
	if len(h.dsAvailMB) == 0 {
		t.Errorf("getDsInfo() Could not get ds info")
	}

	err = h.getTotalMem()
	if err != nil {
		t.Errorf("%+v", err)
	}
	if h.memTotalMB == 0 {
		t.Errorf("getTotalMem() Could not get memTotal.")
	}

	err = h.getVMInfo()
	if err != nil {
		t.Errorf("%+v", err)
	}

	if h.vmCnt == 0 {
		t.Errorf("getVMInfo() Could not get vmCnt.")
	}

	if h.vmMaxCnt == 0 {
		t.Errorf("getVMInfo() Could not get vmMaxCnt")
	}

	if len(h.registeredVMs) == 0 {
		t.Errorf("getVMInfo() Could not get registeredVMs")
	}

	if h.memActiveMB == 0 {
		t.Errorf("getVMInfo() Non expected h.memActiveMB. Got=%d Expected>0",
			h.memActiveMB)
	}

	if err := h.getPortGroups(); err != nil {
		t.Errorf("%+v", err)
	}
	if len(h.portGroups) == 0 {
		t.Errorf("getportGroups() Must include portgroup")
	}
}
