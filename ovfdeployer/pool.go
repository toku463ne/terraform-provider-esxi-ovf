package ovfdeployer

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

type hostgroup map[string]*host

// Pool Available host info
type Pool struct {
	ID      string
	Hosts   hostgroup
	hostIPs []string
}

// NewPool .. Create pool object and saves to database
func NewPool(id string, hostIPs []string,
	user, pass, testID string) (*Pool, error) {
	logDebug("NewPool(%s, %v, %s, ***, %s)", id, hostIPs,
		user, pass, testID)

	p := new(Pool)
	if id == "" {
		return nil, errors.New("NewPool(): Id cannot be nil")
	}
	p.Hosts = make(hostgroup, len(p.hostIPs))
	p.ID = id
	for _, ip := range hostIPs {
		if err := p.addHost(ip, user, pass, getTestFile(testID, ip)); err != nil {
			return nil, err
		}
	}
	if err := p.sync2Db(); err != nil {
		return nil, err
	}
	return p, nil
}

// LoadPool .. Load pool data from database
func LoadPool(poolid, password string) (*Pool, error) {
	logDebug("LoadPool(%s, ***)", poolid)
	p := new(Pool)
	p.ID = poolid
	dbw, err := openDb(p.ID, poolDbName)
	if err != nil {
		return nil, err
	}
	defer dbw.closeDb()

	hostIPs, err := dbw.syncFromKeyTable("HOST_IPS")
	if err != nil {
		return nil, err
	}
	p.hostIPs = table2slice(hostIPs)
	p.Hosts = make(hostgroup, 0)
	for _, ip := range p.hostIPs {
		if ip == "" {
			continue
		}
		h, err := loadHost(poolid, ip, password)
		//log.Printf("host is %+v", h)
		if err != nil {
			return nil, err
		}
		p.Hosts[ip] = h
	}
	return p, nil
}

// DeletePool .. Removes pool data from database
func DeletePool(id, password string) error {
	logDebug("DeletePool(%s, ***), id")
	p, err := LoadPool(id, password)
	if err != nil {
		return err
	}
	for _, h := range p.Hosts {
		if err := h.deleteHost(); err != nil {
			return err
		}
	}
	if err := deleteDb(id, poolDbName); err != nil {
		return err
	}
	if err := deleteDb(id, vmDbName); err != nil {
		return err
	}
	return err
}

// ChangePool .. Detect change of host ips and change database. Ignore changes at user, password, logLevel
func ChangePool(id, user, password string, hostIPs []string, testID string) (*Pool, error) {
	logDebug("ChangePool(%s, %s, ***, %s, %v, %s)", id, user, hostIPs, testID)

	p, err := LoadPool(id, password)
	if err != nil {
		return nil, err
	}
	removedIPs := getRemovedElesFromSlice(p.hostIPs, hostIPs)
	for _, ip := range removedIPs {
		if err := p.deleteHost(ip); err != nil {
			return nil, err
		}
	}
	addedIPs := getRemovedElesFromSlice(hostIPs, p.hostIPs)
	for _, ip := range addedIPs {
		if err := p.addHost(ip, user, password, getTestFile(testID, ip)); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// AssertPool .. Check if database and tf file matches
func AssertPool(id, password string, hostIPs []string) error {
	p, err := LoadPool(id, password)
	if err != nil {
		return err
	}
	ips1 := sortIPSlice(p.hostIPs)
	ips2 := sortIPSlice(hostIPs)
	if !reflect.DeepEqual(ips1, ips2) {
		return errors.New(fmt.Sprintf("Host IPs do not match. got=%+v expected=%+v",
			hostIPs, p.hostIPs))
	}

	return nil
}

func (p *Pool) hasHost(ip string) bool {
	return hasStr(p.hostIPs, ip)
}

func (p *Pool) addHost(ip, user, pass, testFile string) error {
	if p.hasHost(ip) {
		return errors.Errorf("addHost(): Host with IP=%s already exists in this pool", ip)
	}
	h, err := newHost(p.ID, ip, user, pass, testFile)
	if err != nil {
		return err
	}
	err = h.sync2Db()
	if err != nil {
		return err
	}
	p.hostIPs = append(p.hostIPs, ip)
	p.Hosts[ip] = h
	return nil
}

func (p *Pool) deleteHost(ip string) error {
	if p.hasHost(ip) == false {
		return nil
	}
	p.hostIPs = delSliceElement(p.hostIPs, ip)

	if err := p.Hosts[ip].deleteHost(); err != nil {
		return err
	}
	delete(p.Hosts, ip)
	dbw, err := openDb(p.ID, poolDbName)
	if err != nil {
		return err
	}
	defer dbw.closeDb()

	sql := fmt.Sprintf(`DELETE FROM HOST_IPS WHERE VAL="%s";`, ip)
	if err := dbw.updateDbWithRetry(sql); err != nil {
		return err
	}
	return nil
}

func (p *Pool) sync2Db() error {
	if err := p.clearVMsTable(); err != nil {
		return err
	}

	dbw, err := openDb(p.ID, poolDbName)
	if err != nil {
		return err
	}
	defer dbw.closeDb()

	keytables := make(map[string]Table, 5)
	keytables["HOST_IPS"] = slice2table(p.hostIPs)
	for k, v := range keytables {
		if err := dbw.sync2KeyTable(k, v); err != nil {
			return err
		}
	}
	return err
}

func (p *Pool) clearVMsTable() error {
	dbw, err := openDb(p.ID, vmDbName)
	if err != nil {
		return err
	}
	if err := dbw.createVMsTable(); err != nil {
		return err
	}
	db := dbw.db
	if _, err := db.Exec(`DELETE FROM VMS;`); err != nil {
		return err
	}
	return nil
}

func (p *Pool) getMostVacantHost(memSize, dsSize, cpuCores int, portgroup string) (string, int, error) {
	hosts := p.Hosts
	maxVMAvailableCnt := 0
	maxVMAvailableHost := ""
	hostIPs := sortIPSlice(p.hostIPs)
	for _, ip := range hostIPs {
		if ip == "" {
			continue
		}
		h := hosts[ip]

		if cpuCores > h.cpuCoresCnt {
			logDebug("Host %s has only %d CPU Cores. Want=%d", ip, h.cpuCoresCnt, cpuCores)
			continue
		}

		if portgroup != "" {
			hasPG := false
			for pg := range h.portGroups {
				if pg == portgroup {
					hasPG = true
				}
			}
			if hasPG == false {
				logDebug("Host %s does not have porggroup=%s", ip, portgroup)
				continue
			}
		}

		_, dsMaxName, dsAvailSize, err := p.getMostVacantStorage(ip)
		if err != nil {
			return "", -1, err
		}
		if dsSize >= dsAvailSize {
			logDebug("Most vacant storage in host=%s is %s. avail=%d want=%d", ip,
				dsMaxName, dsAvailSize, dsSize)
			continue
		}

		activeVMMemSize, err := calcVMMaxCnt(h.memActiveMB)
		if err != nil {
			logDebug("Error in calcVMMaxCnt(%d). %+v", h.memActiveMB, err)
			continue
		}
		vmAvailableCnt := h.vmMaxCnt - activeVMMemSize
		vmMemSize, err := calcVMMaxCnt(strconv.Itoa(memSize))
		if err != nil {
			return "", -1, err
		}
		if vmAvailableCnt <= vmMemSize {
			memTotal, _ := strconv.Atoi(h.memTotalMB)
			memActive, _ := strconv.Atoi(h.memActiveMB)
			logDebug("Not enough memory in host=%s memAvailable=%d, need=%d",
				ip, memTotal-memActive, memSize)
			continue
		}
		if vmAvailableCnt > maxVMAvailableCnt {
			maxVMAvailableCnt = vmAvailableCnt
			maxVMAvailableHost = ip
		}
	}
	if maxVMAvailableHost == "" {
		return "", -1, errors.New("Could not get most vacant host, or all hosts may be full")
	}
	return maxVMAvailableHost, maxVMAvailableCnt, nil
}

func (p *Pool) getMostVacantStorage(hostIP string) (string, string, int, error) {
	var hostIPs []string
	hosts := p.Hosts
	if hostIP == "" {
		hostIPs = []string{}
		for k := range hosts {
			hostIPs = append(hostIPs, k)
		}
	} else {
		_, ok := hosts[hostIP]
		if ok {
			hostIPs = []string{hostIP}
		} else {
			return "", "", -1,
				errors.New(fmt.Sprintf("Host %s is not in the host group %s", hostIP, p.ID))
		}
	}
	maxDsSize := 0
	maxDsName := ""
	maxHostIP := ""
	for _, ip := range hostIPs {
		host := hosts[ip]
		for dsName, dsSize := range host.dsAvailMB {
			if fdsSize, err := strconv.Atoi(dsSize); err == nil {
				if fdsSize > maxDsSize {
					maxDsSize = fdsSize
					maxDsName = dsName
					maxHostIP = ip
				}
			} else {
				return "", "", -1, err
			}
		}
	}
	if maxDsSize == 0 {
		return "", "", -1,
			errors.New(fmt.Sprintf("No available storage in %s found!", p.ID))
	}
	return maxHostIP, maxDsName, maxDsSize, nil
}

func (p *Pool) getDsAvailSize(hostIP string, dsPath string) (int, error) {
	h := p.Hosts[hostIP]
	for dsName, dsSize := range h.dsAvailMB {
		if dsName == dsPath {
			return strconv.Atoi(dsSize)
		}
	}
	return 0, errors.New(fmt.Sprintf("Could not find datastore %s", dsPath))
}

func (p *Pool) appendVMResource(hostIP string,
	dsSize, memSize, cpuCores int, ds, portgroup string) (string, string, error) {
	var err error
	vmAvailCnt := 0
	dsAvailSize := 0

	if hostIP == "" {
		hostIP, vmAvailCnt, err = p.getMostVacantHost(memSize, dsSize, cpuCores, portgroup)
		if err != nil {
			return "", "", err
		}
	} else {
		vmAvailCnt = p.Hosts[hostIP].vmMaxCnt - p.Hosts[hostIP].vmCnt
	}

	vmCntToAdd := int(memSize / memPerVMMB)
	if vmAvailCnt < vmCntToAdd {
		return "", "", errors.New("No host with enough memory found")
	}

	h := p.Hosts[hostIP]

	//determine which datastore to deploy
	if ds == "" {
		hostIP, ds, dsAvailSize, err = p.getMostVacantStorage(hostIP)
		if err != nil {
			return "", "", err
		}
	} else {
		dsAvailSize, err = p.getDsAvailSize(hostIP, ds)
		if err != nil {
			return "", "", err
		}
	}
	if dsSize > dsAvailSize {
		return "", "", errors.New(fmt.Sprintf("VM HD size exceeds available storage size. Most available : %d",
			dsAvailSize))
	}

	h.dsAvailMB[ds] = strconv.Itoa(dsAvailSize - dsSize)
	memActiveMBi, err := strconv.Atoi(h.memActiveMB)
	if err != nil {
		return "", "", err
	}
	h.vmCnt += vmCntToAdd
	h.memActiveMB = strconv.Itoa(memSize + memActiveMBi)
	return hostIP, ds, nil
}

func (p *Pool) checkIfVMExists(vmname string) (bool, string, error) {
	for _, ip := range p.hostIPs {
		h := p.Hosts[ip]
		logDebug("h.checkIfVMExists(%s) hostip=%s", vmname, ip)
		existsInHost, err := h.checkIfVMExists(vmname)
		if err != nil {
			return false, "", errors.Wrapf(err, "Checking host %s", ip)
		}
		if existsInHost {
			return false, ip, nil
		}
	}
	return false, "", nil
}
