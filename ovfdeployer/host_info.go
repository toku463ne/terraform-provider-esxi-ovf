package ovfdeployer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

/***************************
  Get datastore size info
  Tested with ESXi5.5, 6
****************************/
func (h *host) getDsInfo() error {
	cmd := "df -h | grep 'VMFS' | awk '{print $2,$4,$6}'"
	if res, err := h.sshExp.run(cmdGetDatastores, cmd); err == nil {
		h.dsAvailMB = make(Table, len(res))
		for _, line := range res {
			s := strings.Split(line, " ")
			if len(s) < 2 {
				continue
			}
			dsPath := s[2]
			d := strings.Split(dsPath, "/")
			h.dsAvailMB[d[len(d)-1]] = getSizeStr(s[1])
		}
	} else {
		return err
	}
	return nil
}

/***************************
  Get datastore size info
  Tested with ESXi5.5, 6
****************************/
func (h *host) getTotalMem() error {
	cmd := "esxcli hardware memory get|grep Physical|sed 's/Bytes//'|cut -d':' -f2"
	//cmd := "esxcli hardware memory get|grep Physical|sed 's/Bytes//'"
	res, err := h.sshExp.run(cmdGetTotalmem, cmd)
	if err != nil {
		return err
	}
	h.memTotalMB = getSizeStr(removeBlanks(res[0]))
	return nil
}

/***************************
  Get VM info on ESXi host
  Tested with ESXi5.5, 6
****************************/
func (h *host) getVMInfo() error {
	cmd := "ps | grep vmx-svga | awk '{print $3}'|cut -d':' -f2"
	ares, err := h.sshExp.run(cmdGetActiveVMs, cmd)
	if err != nil {
		return err
	}
	ma, err := calcVMMaxCnt(h.memTotalMB)
	if err != nil {
		return err
	}
	h.vmCnt = len(ares)
	h.vmMaxCnt = ma

	cmd = "vim-cmd vmsvc/getallvms | sed '1d' | awk '{if ($1 > 0) print $1,$2}'"
	rres, err := h.sshExp.run(cmdGetAllVMs, cmd)
	if err != nil {
		return err
	}
	h.registeredVMs = make(Table, len(rres))
	activeVMIDs := make([]string, 0)
	for _, l := range rres {
		inf := strings.Split(l, " ")
		if len(inf) != 2 {
			continue
		}
		_, err = strconv.Atoi(inf[0])
		if err != nil { //vmid must be number
			continue
		}
		h.registeredVMs[inf[0]] = inf[1]
		for _, i := range ares {
			if i == inf[1] {
				activeVMIDs = append(activeVMIDs, inf[0])
			}
		}
	}

	cmd = `for i in %s;do vim-cmd vmsvc/get.summary $i|grep memorySizeMB|awk -F'=' '{print $2}'|sed 's/,//g';done`
	mres, err := h.sshExp.run(cmdGetAllocatedMem, cmd, strings.Join(activeVMIDs, " "))
	if err != nil {
		return err
	}
	memActiveMBi := 0
	for _, m := range mres {
		m = strings.Trim(m, " ")
		if im, err := strconv.Atoi(m); err != nil {
			continue
		} else {
			memActiveMBi += im
		}
	}
	h.memActiveMB = strconv.Itoa(memActiveMBi)

	return nil
}

/***************************
  Get VMID from VMname
  Tested with ESXi5.5, 6
****************************/
func (h *host) getVMId(vmname string) (string, error) {
	vmid := ""
	res, err := h.sshExp.run(cmdGetVMIDFromVMName, "vim-cmd vmsvc/getallvms | sed '1d' | awk '{if ($2 == \"%s\") print $1}'", vmname)
	if err != nil {
		return "", err
	}
	if len(res) > 0 {
		vmid = strings.Trim(res[0], " ")
	}
	return vmid, nil
}

/***************************
  Get VMname from VMid
  Tested with ESXi5.5, 6
****************************/
func (h *host) getVMName(vmid string) (string, error) {
	vmname := ""
	res, err := h.sshExp.run(cmdGetVMNameFromVMID, "vim-cmd vmsvc/getallvms | sed '1d' | awk '{if ($1 == \"%s\") print $2}'", vmid)
	if err != nil {
		return "", err
	}
	if len(res) == 0 {
		return "", errors.New(fmt.Sprintf("Vmid %s does not exist.", vmid))
	}
	vmname = strings.Trim(res[0], " ")

	return vmname, nil
}

/***************************
  Get portGroups
  Tested with ESXi5.5, 6
****************************/
func (h *host) getPortGroups() error {
	cmd := "esxcli network vswitch standard list | grep -i portgroups | cut -d':' -f2"
	res, err := h.sshExp.run(cmdGetPortgroupNames, cmd)
	if err != nil {
		return err
	}
	i := 0
	h.portGroups = make(Table, len(res))
	for _, line := range res {
		tmp := strings.Split(line, ",")
		for _, pg := range tmp {
			h.portGroups[strings.Trim(pg, " ")] = "" //= append(h.portGroups, strings.Trim(pg, " "))
			i++
		}
	}

	return nil
}

/***************************
  Get Esxi version
  Tested with ESXi5.5, 6
****************************/
func getEsxiVersion(se *sshExpect) (string, error) {
	res, err := se.run(cmdGetVMwareVersion, "vmware -v | awk '{print $3}'")
	if err == nil && len(res) > 0 {
		return res[0], err
	}
	return "", err
}

func calcVMMaxCnt(totalMemoryMB string) (int, error) {
	m, err := strconv.Atoi(totalMemoryMB)
	if err != nil {
		return 0, errors.Wrapf(err, "Error converting. totalMemoryMB=%s", totalMemoryMB)
	}
	return int(m / memPerVMMB), nil
}
