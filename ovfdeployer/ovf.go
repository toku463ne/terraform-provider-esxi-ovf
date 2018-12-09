package ovfdeployer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func getMbInt(valstr, unitstr string) (int, error) {
	valstr = strings.Replace(valstr, "\"", "", -1)
	val, err := strconv.Atoi(valstr)
	if err != nil {
		err = errors.Wrapf(err, "Target value=%s", valstr)
		return -1, err
	}
	switch unitstr {
	case "byte * 2^30":
		val *= 1024
	case "byte * 2^20":
		val *= 1
	case "byte * 2^10":
		val /= 1024
		val = int(val)
	default:
		return -1, errors.New(fmt.Sprintf("Not expected unit. Got=%s", unitstr))
	}
	return val, nil
}

func getWorkOvfPath(vmname, orgOvfPath string) (string, string) {
	ovfDir := ""
	if orgOvfPath != "" {
		logDebug("filepath.Dir(%s)", orgOvfPath)
		ovfDir = filepath.Dir(orgOvfPath)
	} else {
		ovfDir = workSubDir
	}
	setWorkSubDir()
	tmp := strings.Split(orgOvfPath, "/")
	ovfPath := fmt.Sprintf("%s/%s_%s", workSubDir, vmname, tmp[len(tmp)-1])
	return ovfDir, ovfPath
}

func (vm *VM) editOvf() error {
	ovfDir, ovfpath := getWorkOvfPath(vm.name, vm.ovfpath)
	if vm.ovfpath == "" {
		return errors.Errorf("vm.ovfpath is empty")
	}
	in, err := os.Open(vm.ovfpath)
	if err != nil {
		return errors.Wrapf(err, "editOvf() ovfpath=%s", vm.ovfpath)
	}
	defer func() error {
		if err := in.Close(); err != nil {
			return err
		}
		return nil
	}()
	ou, err := os.Create(ovfpath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(ou)
	defer func() error {
		if err := ou.Close(); err != nil {
			return err
		}
		return nil
	}()

	reader := bufio.NewReader(in)
	var line string
	lines := make([]string, 0)
	title := ""
	oktooutput := true
	resourcetype := ""
	eof := false
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			eof = true
			oktooutput = true
		}
		if match, _ := regexp.MatchString(`<References>`, line); match {
			title = getXMLTagVal(line)
		}

		if match, _ := regexp.MatchString(`<.*Section>`, line); match {
			title = getXMLTagVal(line)
		}

		if match, _ := regexp.MatchString(`<Item>`, line); match {
			title = "Item"
		}

		if title == "Item" {
			oktooutput = false
		}

		if title == "Item" && strings.Contains(line, `<rasd:ResourceType>`) {
			resourcetype = getXMLVal(line)
		}

		if title == "References" && strings.Contains(line, `<File`) {
			vmdkFileName := getXMLOvfAttr(line, "href")
			vmdkPath := ""
			if vmdkFileName[:1] == "/" {
				vmdkPath = vmdkFileName
			} else if ovfDir[:1] == "/" {
				vmdkPath = fmt.Sprintf("%s/%s", ovfDir, vmdkFileName)
			} else {
				vmdkPath = fmt.Sprintf("../../%s/%s", ovfDir, vmdkFileName)
			}
			line = strings.Replace(line, vmdkFileName, vmdkPath, 1)
		}

		//get disk size from ovf info
		if title == "DiskSection" && strings.Contains(line, `<Disk`) {
			sizestr := getXMLOvfAttr(line, "capacity")
			unit := getXMLOvfAttr(line, "capacityAllocationUnits")
			if sizestr != "" && unit != "" {
				dsSize, err := getMbInt(sizestr, unit)
				if err != nil {
					return err
				}
				vm.dsSize = dsSize
			}
		}

		if strings.Contains(line, "</Item>") {
			oktooutput = true
		}

		lines = append(lines, line)

		if oktooutput {
			for i, line := range lines {
				if title == "Item" && strings.Contains(line, `<rasd:VirtualQuantity>`) {
					switch resourcetype {
					case "3": //CPU cores
						line = setXMLVal(line, strconv.Itoa(vm.cpuCores))
					case "4": //Memory
						line = setXMLVal(line, strconv.Itoa(vm.memSize))
					}
				}
				if title == "Item" && strings.Contains(line, `<rasd:ElementName>`) {
					switch resourcetype {
					case "3": //CPU cores
						line = setXMLVal(line, fmt.Sprintf("%d virtual CPU(s)", vm.cpuCores))
					case "4": //Memory
						line = setXMLVal(line, fmt.Sprintf("%dMB of memory", vm.memSize))
					}
				}

				if title == "Item" && strings.Contains(line, `vmw:CoresPerSocket`) {
					line = setXMLVal(line, strconv.Itoa(vm.cpuCores))
				}
				lines[i] = line
			}
			resourcetype = ""

			for _, l := range lines {
				if _, err := fmt.Fprint(w, l); err != nil {
					return err
				}
			}
			lines = make([]string, 0)
			oktooutput = true
			if eof {
				break
			}
		}
	}
	vm.ovfpath = ovfpath
	return w.Flush()
}
