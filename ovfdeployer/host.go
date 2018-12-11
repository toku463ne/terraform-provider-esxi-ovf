package ovfdeployer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type host struct {
	poolid        string
	ip            string
	user          string
	pass          string
	version       string
	sshExp        sshExpect
	memTotalMB    int
	memActiveMB   int
	cpuCoresCnt   int
	vmMaxCnt      int
	vmCnt         int
	ballooningMB  int
	portGroups    Table
	dsAvailMB     Table
	registeredVMs Table
}

func (h *host) getHostDbName() string {
	return fmt.Sprintf("host_%s", h.ip)
}

func (h *host) sync2Db() error {
	dbw, err := openDb(h.poolid, h.getHostDbName())
	if err != nil {
		return err
	}
	defer dbw.closeDb()
	keytables := make(map[string]Table, 5)
	keytables["DS_AVAIL_MB"] = h.dsAvailMB
	keytables["REGISTERED_VMS"] = h.registeredVMs
	keytables["PORT_GROUPS"] = h.portGroups
	keytables["BASIC"] = Table{
		"memTotalMB":   strconv.Itoa(h.memTotalMB),
		"memActiveMB":  strconv.Itoa(h.memActiveMB),
		"ballooningMB": strconv.Itoa(h.ballooningMB),
		"cpuCoresCnt":  strconv.Itoa(h.cpuCoresCnt),
		"user":         h.user,
		"vmCnt":        strconv.Itoa(h.vmCnt),
		"version":      h.version,
		"testFile":     h.sshExp.testFile,
	}
	for k, v := range keytables {
		if err := dbw.sync2KeyTable(k, v); err != nil {
			return err
		}
	}
	return nil
}

func loadHost(poolid, hostIP, password string) (*host, error) {
	h := new(host)
	h.poolid = poolid
	h.ip = hostIP
	h.pass = password

	dbw, err := openDb(h.poolid, h.getHostDbName())
	if err != nil {
		return nil, err
	}
	defer dbw.closeDb()

	if h.dsAvailMB, err = dbw.syncFromKeyTable("DS_AVAIL_MB"); err != nil {
		return nil, err
	}
	if h.registeredVMs, err = dbw.syncFromKeyTable("REGISTERED_VMS"); err != nil {
		return nil, err
	}
	if h.portGroups, err = dbw.syncFromKeyTable("PORT_GROUPS"); err != nil {
		return nil, err
	}
	basict, err := dbw.syncFromKeyTable("BASIC")
	if err != nil {
		return nil, err
	}
	h.memTotalMB, err = strconv.Atoi(basict["memTotalMB"])
	if err != nil {
		return nil, err
	}
	h.memActiveMB, err = strconv.Atoi(basict["memActiveMB"])
	if err != nil {
		return nil, err
	}
	h.ballooningMB, err = strconv.Atoi(basict["ballooningMB"])
	if err != nil {
		return nil, err
	}
	h.memTotalMB += h.ballooningMB
	h.cpuCoresCnt, err = strconv.Atoi(basict["cpuCoresCnt"])
	if err != nil {
		return nil, err
	}
	h.user = basict["user"]
	h.version = basict["version"]
	cnt, err := strconv.Atoi(basict["vmCnt"])
	if err != nil {
		return nil, errors.Wrapf(err, "Error converting vmCnt=%s", basict["vmCnt"])
	}
	h.vmCnt = cnt

	ma, err := calcVMMaxCnt(h.memTotalMB)
	if err != nil {
		return nil, err
	}
	h.vmMaxCnt = ma
	se, err := getSSHExpectConn(hostIP, h.user, h.pass, basict["testFile"])
	if err != nil {
		return nil, err
	}
	h.sshExp = *se

	return h, nil
}

func (h *host) deleteHost() error {
	return deleteDb(h.poolid, h.getHostDbName())
}

func (h *host) registerVM(vmid, vmname string) error {
	dbw, err := openDb(h.poolid, h.getHostDbName())
	if err != nil {
		return err
	}
	strsql := fmt.Sprintf(`INSERT INTO REGISTERED_VMS 
	VALUES("%s", "%s")`, vmid, vmname)
	err = dbw.updateDbWithRetry(strsql)
	if err != nil {
		return err
	}
	return nil
}

func newHost(poolid, hostIP, user, pass string,
	ballooningMB int, testFile string) (*host, error) {
	h := new(host)
	h.poolid = poolid
	h.ip = hostIP
	h.user = user
	h.pass = pass
	h.ballooningMB = ballooningMB
	se, err := getSSHExpectConn(hostIP, user, pass, testFile)
	if err != nil {
		return nil, err
	}
	h.sshExp = *se

	ver, err := getEsxiVersion(se)
	if err != nil {
		return nil, err
	}
	h.version = ver

	if err := h.setupEsxiHost(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *host) setupEsxiHost() error {
	if err := h.getDsInfo(); err != nil {
		return err
	}
	if err := h.getTotalMem(); err != nil {
		return err
	}
	if err := h.getCPUCores(); err != nil {
		return err
	}
	if err := h.getVMInfo(); err != nil {
		return err
	}
	if err := h.getPortGroups(); err != nil {
		return err
	}
	return nil
}

func (h *host) assertVMDir(vmname, vmdir string) error {
	cmd := fmt.Sprintf(`grep displayName %s/%s.vmx | awk '{print $3}' | sed s/\"//g`, vmdir, vmname)
	res, err := h.sshExp.run(cmdGetDisplayName, cmd)
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return errors.New(`Param diplayName is absent in`)
	}
	displayName := strings.Trim(res[0], " ")
	if vmname == displayName {
		return nil
	}

	return errors.New(fmt.Sprintf("Path of vmx file is not as expected. Expected=%s", vmdir))
}

func (h *host) destroyVM(vmid, vmname, vmdir string) error {
	vmname2, err := h.getVMName(vmid)
	if err != nil {
		return err
	}
	if vmname != vmname2 {
		return errors.New(fmt.Sprintf("Vmid and Name for Vmid %s do not match. Expected=%s Got=%s",
			vmid, vmname, vmname2))
	}
	cmd := fmt.Sprintf("ls %s/%s.vmx &> /dev/null && vim-cmd /vmsvc/destroy %s",
		vmdir, vmname, vmid)
	if isDebug {
		logInfo(cmd)
	} else {
		if err := h.runCmd(cmdDestroyVM, cmd); err != nil {
			return err
		}
	}

	dbw, err := openDb(h.poolid, h.getHostDbName())
	if err != nil {
		return err
	}
	strsql := fmt.Sprintf(`DELETE FROM REGISTERED_VMS 
	WHERE KEY="%s" AND VAL="%s"`, vmid, vmname)
	err = dbw.updateDbWithRetry(strsql)
	if err != nil {
		return err
	}
	return nil
}

func (h *host) runCmd(cmdName, cmd string) error {
	cmd = fmt.Sprintf("(%s); echo $?", cmd)
	res, err := h.sshExp.run(cmdName, cmd)
	if err != nil {
		return err
	}
	status := res[len(res)-1]
	if status != "0" {
		return errors.New(fmt.Sprintf(`Non zero status.
			cmd=%s
			result=%s`, cmd, res))
	}
	return nil
}

func (h *host) checkIfVMExists(vmname string) (bool, error) {
	dbw, err := openDb(h.poolid, h.getHostDbName())
	if err != nil {
		return false, err
	}
	s, err := dbw.getKeysFromKeyTable("REGISTERED_VMS", vmname)
	if err != nil {
		return false, err
	}
	if len(s) == 0 {
		return false, nil
	}

	if id, err := h.getVMId(vmname); err != nil {
		return false, err
	} else if id != "" {
		return true, nil
	}
	return true, nil
}
