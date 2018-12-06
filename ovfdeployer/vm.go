package ovfdeployer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// VM vm to be deployed
type VM struct {
	poolid, name, password    string
	ovfpath                   string
	dsSize, memSize, cpuCores int
	hostIP                    string
	datastore                 string
	portgroups                []string
	guestinfos                []string
	vmid                      string
	hostUser                  string
	hostPass                  string
	seq                       int64
	pool                      Pool
}

// NewVM set each infos, load pool data from Sqlite
func NewVM(
	poolid string,
	name string,
	password string,
	ovfpath string,
	memSize, cpuCores int,
	hostIP string,
	datastore string,
	portgroups []string,
	guestinfos []string) (*VM, error) {
	vm := new(VM)
	vm.poolid = poolid
	vm.name = name
	vm.password = password
	vm.ovfpath = ovfpath
	vm.dsSize = -1
	vm.memSize = memSize
	vm.cpuCores = cpuCores
	vm.hostIP = hostIP
	vm.datastore = datastore
	vm.portgroups = portgroups
	vm.guestinfos = guestinfos
	p, err := LoadPool(poolid, password)
	if err != nil {
		return nil, err
	}
	vm.pool = *p
	return vm, nil
}

// CheckVMID Check if the id equals to the one in the database
func CheckVMID(id, password string) error {
	poolid, vmid, vmname, hostIP, _, _, _, err := getVMDBInfo(id)
	if err != nil {
		return err
	}
	p, err := LoadPool(poolid, password)
	if err != nil {
		return err
	}
	h := p.Hosts[hostIP]
	vmid2, err := h.getVMId(vmname)
	if err != nil {
		return err
	} else if vmid != vmid2 {
		return errors.New(fmt.Sprintf("VmId not match. Expected=%s Got=%s vmname=%s",
			vmid, vmid2, vmname))
	}
	return nil
}

// DestroyVM Deletes VM from ESXi.
func DestroyVM(id, password string) error {
	logInfo("DestroyVM(%s)", id)

	poolid, vmid, vmname, hostIP, _, _, datastore, err := getVMDBInfo(id)
	if err != nil {
		return err
	}

	vm, err := NewVM(poolid, vmname, "", password,
		-1, -1, hostIP, datastore, nil, nil)
	if err != nil {
		return err
	}
	p := vm.pool
	h := p.Hosts[hostIP]

	isPowerON, err := vm.isPowerON("OFF")
	if isPowerON {
		return errors.Errorf("VM %s is powered on. Please power off before destroying.", vmname)
	}

	vmdir := getVMDir(vmname, datastore)
	if err := h.assertVMDir(vmname, vmdir); err != nil {
		return err
	}

	if err := h.destroyVM(vmid, vmname, vmdir); err != nil {
		return err
	}
	if err := deleteVMDB(poolid, vmname); err != nil {
		return err
	}
	_, ovfpath := getWorkOvfPath(vmname, "")
	if ovfpath == vm.ovfpath {
		if err := os.Remove(vm.ovfpath); err != nil {
			return err
		}
	}
	return nil
}

func deleteVMDB(poolID, vmname string) error {
	dbw, err := openDb(poolID, vmDbName)
	if err != nil {
		return err
	}
	defer dbw.closeDb()

	sql := fmt.Sprintf(`DELETE FROM VMS 
			WHERE  NAME="%s";`,
		vmname)
	err = dbw.updateDbWithRetry(sql)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) getID() string {
	return fmt.Sprintf("%s:%s:%s", vm.poolid, vm.hostIP, vm.vmid)
}

func expandID(id string) (string, string, string, error) {
	eid := strings.Split(id, ":")
	if len(eid) != 3 {
		return "", "", "", errors.Errorf("id is not in correct format")
	}
	return eid[0], eid[1], eid[2], nil
}

func getVMDir(vmname, datastore string) string {
	return fmt.Sprintf("%s/%s/%s", vmVolumesPath, datastore, vmname)
}

func getVMDBInfo(id string) (string, string, string, string, int, int, string, error) {
	poolid, hostIP, vmid, err := expandID(id)
	if err != nil {
		return "", "", "", "", -1, -1, "", err
	}

	dbw, err := openDb(poolid, vmDbName)
	if err != nil {
		return "", "", "", "", -1, -1, "", err
	}

	db := dbw.db
	row := db.QueryRow(fmt.Sprintf(`SELECT 
		NAME, 
		DS_SIZE_MB, 
		MEM_SIZE_MB, 
		DATASTORE FROM VMS 
		WHERE VMID = "%s" AND HOST_IP = "%s"`, vmid, hostIP))
	if err != nil {
		return "", "", "", "", -1, -1, "", err
	}
	var vmname, datastore string
	var dsSize, memPerVMMB int
	err = row.Scan(&vmname, &dsSize, &memPerVMMB, &datastore)
	//no_rows_str := "sql: no rows"
	if err != nil {
		return "", "", "", "", -1, -1, "", errors.Wrap(err, "DB access error")
	}

	return poolid, vmid, vmname, hostIP, dsSize, memPerVMMB, datastore, nil
}

func (vm *VM) reserveVMResource() error {
	dbw, err := openDb(vm.poolid, vmDbName)
	if err != nil {
		return err
	}
	seq, err := vm.insert2VMDB()
	if err != nil {
		return err
	}
	vm.seq = seq
	dbw.closeDb()
	dbw, err = dbw.openDb()
	if err != nil {
		return err
	}

	db := dbw.db
	rows, err := db.Query(fmt.Sprintf(`SELECT 
		SEQ, 
		NAME,
		HOST_IP, 
		DS_SIZE_MB, 
		MEM_SIZE_MB, 
		DATASTORE FROM VMS 
		WHERE SEQ <= %d
		ORDER BY SEQ`, seq))
	if err != nil {
		return err
	}
	p, err := LoadPool(vm.poolid, vm.password)
	if err != nil {
		return err
	}
	var seq2 int64
	var dsSize, memSize int
	var name, hostIP, ds string
	if rows == nil {
		return errors.New("no rows in VMS table")
	}
	defer rows.Close()
	for rows.Next() {
		hostIP2 := ""
		ds2 := ""
		rows.Scan(&seq2, &name, &hostIP2, &dsSize, &memSize, &ds2)
		portgroup := ""
		if len(vm.portgroups) > 0 {
			portgroup = vm.portgroups[0]
		}
		hostIP, ds, err = p.appendVMResource(hostIP2, dsSize, memSize, ds2, portgroup)
		if err != nil {
			continue
		}
	}
	if err != nil {
		return err
	}
	h := p.Hosts[hostIP]
	vm.name = name
	vm.hostIP = hostIP
	vm.datastore = ds
	vm.hostUser = h.user
	vm.hostPass = h.pass
	vm.pool = *p

	return nil
}

func (vm *VM) insert2VMDB() (int64, error) {
	dbw, err := openDb(vm.poolid, vmDbName)
	if err != nil {
		return -1, err
	}
	seq := getEpochNanos()
	sql := fmt.Sprintf(`INSERT INTO VMS 
			(SEQ, NAME, HOST_IP, DS_SIZE_MB, MEM_SIZE_MB, DATASTORE) 
			VALUES (%d, "%s", "%s", %d, %d, "%s");`,
		seq, vm.name, vm.hostIP, vm.dsSize, vm.memSize, vm.datastore)
	//err = dbw.updateDb(sql)
	err = dbw.updateDbWithRetry(sql)
	defer dbw.closeDb()
	time.Sleep(1000 * time.Millisecond)
	return seq, err
}

func (vm *VM) registerVM() error {
	dbw, err := openDb(vm.poolid, vmDbName)
	if err != nil {
		return err
	}
	defer dbw.closeDb()
	strsql := fmt.Sprintf(`UPDATE VMS 
	SET VMID=%s, HOST_IP="%s", DATASTORE="%s", IS_REGISTERED=1 
	WHERE SEQ=%d`, vm.vmid, vm.hostIP, vm.datastore, vm.seq)
	err = dbw.updateDbWithRetry(strsql)
	if err != nil {
		return err
	}
	h := vm.pool.Hosts[vm.hostIP]
	err = h.registerVM(vm.vmid, vm.name)
	if err != nil {
		return err
	}
	return nil
}

func (vm *VM) configureVMGuestInfo() error {
	h := vm.pool.Hosts[vm.hostIP]
	vmxPath := fmt.Sprintf("/vmfs/volumes/%s/%s/%s.vmx", vm.datastore, vm.name, vm.name)
	cmd := fmt.Sprintf("ls %s 2> /dev/null; echo $?", vmxPath)
	res, err := h.sshExp.run(cmdAssertVmxPath, cmd)
	if err != nil {
		return err
	}
	if res[0] != "0" {
		return errors.New(fmt.Sprintf("File %s.vmx does not exist", vm.name))
	}
	for _, guestinfo := range vm.guestinfos {
		cmd := fmt.Sprintf("grep %s %s &> /dev/null || echo %s >> %s; echo $?", guestinfo, vmxPath,
			guestinfo, vmxPath)
		res, err = h.sshExp.run(cmdAddGuestInfo, cmd)
		if err != nil {
			return err
		} else if len(res) > 0 && res[0] != "0" {
			return errors.Errorf("Non zero status. %s", cmd)
		}
	}
	return nil
}

func (vm *VM) isPowerON(testMode string) (bool, error) {
	h := vm.pool.Hosts[vm.hostIP]
	if vm.vmid == "" {
		return false, errors.New("Vmid is null")
	}
	cmd := fmt.Sprintf("vim-cmd vmsvc/power.getstate %s", vm.vmid)
	testCmdName := ""
	if testMode == "ON" {
		testCmdName = cmdGetVMPowerStateON
	} else if testMode == "OFF" {
		testCmdName = cmdGetVMPowerStateOFF
	}
	res, err := h.sshExp.run(testCmdName, cmd)
	if err != nil {
		return false, err
	}
	if len(res) < 2 {
		return false, errors.New(fmt.Sprintf("Got unexpected responce from power.getstate. Got=%s",
			strings.Join(res, ",")))
	}
	if res[1] == "Powered on" {
		return true, nil
	}
	return false, nil
}

func (vm *VM) powerOnVM() error {
	h := vm.pool.Hosts[vm.hostIP]
	if vm.vmid == "" {
		return errors.New("Vmid is null")
	}
	cmd := fmt.Sprintf("vim-cmd vmsvc/power.on %s &> /dev/null; echo $?", vm.vmid)
	res, err := h.sshExp.run(cmdPowerONVM, cmd)
	if err != nil {
		return err
	}
	log.Printf("%s", strings.Join(res, "\n"))
	if res[0] != "0" {
		return errors.New(fmt.Sprintf("Cannot power on VM. %s. \n%s", vm.name, strings.Join(res, "\n")))
	}
	n := 0
	var lasterr error
	for n < vmCheckPowerStateCnt {
		isPowerON, err := vm.isPowerON("ON")
		if err != nil {
			return err
		}
		if isPowerON {
			return nil
		}
		lasterr = errors.New(fmt.Sprintf("Got unexpected responce. %s", strings.Join(res, "\n")))
		n++
	}
	return lasterr
}

// DeployVM Run ovftool commands and deploy the VM
func DeployVM(poolid, name, password, ovfpath string,
	memSize, cpuCores int,
	hostIP string,
	datastore string,
	portgroups []string,
	guestinfos []string) (string, error) {
	vm, err := NewVM(poolid, name, password, ovfpath,
		memSize, cpuCores, hostIP,
		datastore, portgroups, guestinfos)
	if err != nil {
		return "", err
	}

	vmexists, hostIP, err := vm.pool.checkIfVMExists(name)
	if err != nil {
		return "", err
	}
	if vmexists {
		return "", errors.Errorf("VM with name %s already exists in %s", name, hostIP)
	}

	//ver := vm.pool.Esxi_ver

	if err := vm.editOvf(); err != nil {
		return "", err
	}
	if err := vm.reserveVMResource(); err != nil {
		deleteVMDB(poolid, vm.name)
		return "", err
	}
	h, err := loadHost(poolid, vm.hostIP, vm.password)
	if err != nil {
		return "", err
	}
	if err := vm.deployVM(); err != nil {
		return "", err
	}

	vmid, err := h.getVMId(vm.name)
	if err != nil {
		return "", err
	}
	vm.vmid = vmid
	if err := vm.registerVM(); err != nil {
		return "", err
	}

	if err := vm.configureVMGuestInfo(); err != nil {
		return "", err
	}
	if err := vm.powerOnVM(); err != nil {
		return "", err
	}

	return vm.getID(), nil
}

func (vm *VM) deployVM() error {
	cmdstr := fmt.Sprintf(`%s --disableVerification --noSSLVerify 
		--diskMode=thin --name=%s --datastore=%s --network=%s %s vi://%s:%s@%s`,
		ovfBin, vm.name, vm.datastore, vm.portgroups[0],
		vm.ovfpath, vm.hostUser, "***", vm.hostIP)

	log.Printf(cmdstr)
	cmd := exec.Command(ovfBin, "--disableVerification", "--noSSLVerify", "--diskMode=thin",
		fmt.Sprintf("--name=%s", vm.name), fmt.Sprintf("--datastore=%s", vm.datastore),
		fmt.Sprintf("--network=%s", vm.portgroups[0]), vm.ovfpath,
		fmt.Sprintf("vi://%s:%s@%s", vm.hostUser, vm.hostPass, vm.hostIP))
	if isDebug == false {
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
