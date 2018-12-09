package ovfdeployer

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	ini "gopkg.in/ini.v1"
)

const (
	testBaseDir     = "/tmp/odtests"
	testFilesFolder = "testFiles"
	testFolder      = "te"
)

func getTestDir() string {
	return fmt.Sprintf("%s/%s", pwd, testFolder)
}

func getTestFilesDir() string {
	return fmt.Sprintf("%s/%s", getTestDir(), testFilesFolder)
}

func init4Test(testname string) error {
	isDebug = true
	setPwd()
	setLogLevel(logLevel4Test)
	//poolID = testname
	testdir := ""
	if testname == "" {
		testdir = testBaseDir
	} else {
		testdir = fmt.Sprintf("%s/%s", testBaseDir, testname)
	}
	os.RemoveAll(testdir)

	if _, err := os.Stat(testdir); os.IsNotExist(err) {
		os.MkdirAll(testdir, 0755)
	}
	err := os.Chdir(testdir)
	if err != nil {
		return err
	}
	//if err := openLog("test"); err != nil {
	//	return err
	//}
	return nil
}

func getTestFile(targetFolder, targetName string) string {
	if targetFolder != "" && targetName != "" {
		return fmt.Sprintf("%s/%s/%s.test", getTestFilesDir(), targetFolder, targetName)
	}
	return ""
}

func getTestFiles(targetFolder string) ([]string, error) {
	return filepath.Glob(fmt.Sprintf("%s/%s/*.test", getTestFilesDir(), targetFolder))
}

func getOfflineTestHost(testFile string) (*host, error) {
	h, err := newHost("test", "testhost", "offlineuser", "", testFile)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func getOnlineTestHost(hosttype string) (*host, error) {
	inf, err := getIniSectionInfo(hosttype)
	if err != nil {
		return nil, err
	}
	poolid := fmt.Sprintf("test%s", hosttype)
	hostIP := inf["ipaddr"]
	user := inf["user"]
	pass := inf["password"]
	h, err := newHost(poolid, hostIP, user, pass, "")
	if err != nil {
		return nil, err
	}
	return h, nil
}

func getIniSectionInfo(section string) (map[string]string, error) {
	usr, _ := user.Current()
	cfg, err := ini.Load(fmt.Sprintf("%s/%s", usr.HomeDir, configName))
	if err != nil {
		return nil, err
	}
	keys := cfg.Section(section).Keys()
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		m[k.Name()] = k.Value()
	}
	return m, nil
}

func createTestPool(poolid string, hostCnt int, doSync bool, testFile string) *Pool {
	p := new(Pool)
	p.ID = poolid
	p.Hosts = make(hostgroup)
	for n := 1; n <= hostCnt; n++ {
		h := createTestHost(poolid, n, doSync, testFile)
		p.Hosts[h.ip] = h
	}
	for k := range p.Hosts {
		p.hostIPs = append(p.hostIPs, k)
	}
	hostIPs := sortIPSlice(p.hostIPs)
	p.hostIPs = hostIPs
	if doSync {
		if err := p.sync2Db(); err != nil {
			return nil
		}
	}
	return p
}

func createTestHost(poolid string, n int, doSync bool, testFile string) *host {
	h := new(host)
	h.poolid = poolid
	h.ip = fmt.Sprintf("1.2.3.%d", n)
	h.user = "root"
	h.pass = "hogehoge"
	h.version = "x"
	h.memTotalMB = fmt.Sprintf("%d000", n*2+2)
	h.memActiveMB = fmt.Sprintf("%d000", n+1)
	h.cpuCoresCnt = 16
	ma, err := calcVMMaxCnt(h.memTotalMB)
	h.sshExp.testFile = testFile

	if err != nil {
		return nil
	}
	h.vmMaxCnt = ma
	h.vmCnt = n * 2
	h.portGroups = Table{
		"1.2.3.0": "",
	}
	h.dsAvailMB = Table{
		"Disk1": fmt.Sprintf("5%d400", n%3),
		"Disk2": fmt.Sprintf("5%d300", n%3),
		"Disk3": fmt.Sprintf("5%d200", n%3),
		"Disk4": fmt.Sprintf("5%d100", n%3),
	}
	if doSync {
		if err := h.sync2Db(); err != nil {
			return nil
		}
	}
	return h
}

func createTestVM(poolid, name string, createPoolFirst bool) (*VM, error) {
	if createPoolFirst {
		_ = createTestPool(poolid, 4, true, "dummy")
	}

	return NewVM(
		poolid,
		name,
		"",
		"",
		4000, 2,
		"",
		"",
		[]string{"1.2.3.0"},
		[]string{""}, currentLogLevelStr)
}
