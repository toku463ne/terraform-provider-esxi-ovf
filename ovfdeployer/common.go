package ovfdeployer

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var isDebug bool
var pwd string
var currentlogLevel int
var currentLogLevelStr string
var workSubDir string
var poolID string

const (
	// FlagLocked Queue table is locked
	FlagLocked = 1
	// FlacUnlocked Queue table is unlocked
	FlacUnlocked = 0
	// SqliteMsgUniqErr test
	SqliteMsgUniqErr = "UNIQUE constraint failed"
	// SqliteMsgDbLocked All database is locked
	SqliteMsgDbLocked = "database is locked"
)

// Table is a struct to save in DB
type Table map[string]string

var logfd *os.File

func setWorkSubDir() {
	if poolID == "" {
		workSubDir = fmt.Sprintf("./%s", workDir)
	} else {
		workSubDir = fmt.Sprintf("%s/%s", workDir, poolID)
	}
	ensureDir(workSubDir)
}

// Init .. Initialize env
func initApp(id, appname, logLevelStr string) error {
	//Do not init in cases log level is not specified
	if logLevelStr == "" {
		return nil
	}
	setPwd()
	poolID = id
	setWorkSubDir()

	for _, s := range []string{id, appname} {
		if checkID(s) == false {
			return errors.Errorf(`May only contain lowercase alphanumeric characters & underscores. got=%s`, s)
		}
	}
	setLogLevel(logLevelStr)
	if err := openLog(appname); err != nil {
		return err
	}
	return nil
}

func setLogLevel(logLevelStr string) {
	switch logLevelStr {
	case logLevelDebugStr:
		currentlogLevel = logLevelDebug
	case logLevelInfoStr:
		currentlogLevel = logLevelInfo
	case logLevelWarningStr:
		currentlogLevel = logLevelWarning
	case logLevelErrorStr:
		currentlogLevel = logLevelError
	}
	currentLogLevelStr = logLevelStr
}

func closeApp(logLevelStr string) {
	if logLevelStr != "" {
		closeLog()
	}
}

func ensureDir(dirName string) {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		os.MkdirAll(dirName, 0755)
	}
}

// OpenLog open log file
func openLog(name string) error {
	var err error
	setWorkSubDir()
	fname := fmt.Sprintf("%s/%s.log", workSubDir, name)
	logfd, err = os.Create(fname)
	if err != nil {
		return err
	}
	log.SetOutput(logfd)
	return nil
}

// CloseLog .. close the log file
func closeLog() {
	logfd.Close()
}

func logDebug(msg string, args ...interface{}) {
	writeLog(logLevelDebug, msg, args...)
}
func logInfo(msg string, args ...interface{}) {
	writeLog(logLevelInfo, msg, args...)
}
func logError(msg string, args ...interface{}) {
	writeLog(logLevelError, msg, args...)
}

func writeLog(logLevel int, msg string, args ...interface{}) {
	if currentlogLevel > logLevel {
		return
	}
	severity := ""
	switch logLevel {
	case logLevelDebug:
		severity = "DEBUG "
	case logLevelInfo:
		severity = "INFO "
	case logLevelError:
		severity = "ERROR"
	}

	msg = fmt.Sprintf(msg, args...)
	log.Printf("%d %s | %s", os.Getpid(), severity, msg)
}

func convSizeStrToMB(s string) int {
	n, err := strconv.Atoi(s)
	if err == nil {
		return n / 1024 / 1024
	}
	size := -1.0
	unit := s[len(s)-1 : len(s)]
	if f, err := strconv.ParseFloat(s[0:len(s)-1], 64); err == nil {
		switch unit {
		case "K":
			size = f / 1024
		case "M":
			size = f
		case "G":
			size = f * 1024
		}
	}
	return int(size)
}

func getSizeStr(s string) string {
	return strconv.Itoa(convSizeStrToMB(s))
}

func slice2table(s []string) Table {
	t := make(Table, len(s))
	for i, v := range s {
		t[strconv.Itoa(i)] = v
	}
	return t
}

func table2slice(t Table) []string {
	sl := make([]string, 0)
	for _, v := range t {
		sl = append(sl, v)
	}
	return sl
}

func assertError(err error, msg string) bool {
	if err == nil {
		return false
	}
	errmsg := err.Error()
	if len(errmsg) >= len(msg) && errmsg[:len(msg)] == msg {
		return true
	}
	return false
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	for v := range elements {
		encountered[elements[v]] = true
	}
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func pickCommonElementsBetweenSlices(s1 []string, s2 []string) []string {
	s3 := make([]string, 0)
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				s3 = append(s3, v1)
				break
			}
		}
	}
	return s3
}

func runSed(targetfile, orgstr, replacestr string) error {
	cmd := exec.Command("sed", "-i", "-E", fmt.Sprintf(`s/%s/%s/`, orgstr, replacestr), targetfile)
	return cmd.Run()
}

func getEpochNanos() int64 {
	rand.Seed(time.Now().UnixNano())
	msecs := time.Duration(rand.Intn(100) * 1)
	time.Sleep(msecs * time.Microsecond)
	return time.Now().UnixNano()
}

func removeBlanks(s string) string {
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
}

func setPwd() {
	if pwd == "" {
		pwd, _ = os.Getwd()
	}
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func hasStr(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func delSliceElement(a []string, x string) []string {
	b := make([]string, 0)
	for _, y := range a {
		if y != x {
			b = append(b, y)
		}
	}
	return b
}

func sortIPSlice(ips []string) []string {
	ipmap := map[string]string{}
	ipns := []string{}
	for _, ip := range ips {
		if ip == "" {
			continue
		}
		ipstrs := strings.Split(ip, ".")
		for i, ipstr := range ipstrs {
			ipstrs[i] = fmt.Sprintf("%03s", ipstr)
		}
		ipstr := strings.Join(ipstrs, "")
		ipmap[ipstr] = ip
		ipns = append(ipns, ipstr)
	}
	sort.Strings(ipns)
	newIPs := make([]string, len(ipns))
	for i, ipn := range ipns {
		newIPs[i] = ipmap[ipn]
	}
	return newIPs
}

func checkID(id string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9\.\_\-]+$`).Match([]byte(id))
}

func getRemovedElesFromSlice(a, b []string) []string {
	c := make([]string, 0)
	for _, ae := range a {
		if hasStr(b, ae) == false {
			c = append(c, ae)
		}
	}
	return c
}
