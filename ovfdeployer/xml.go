package ovfdeployer

import (
	"fmt"
	"strings"
)

func getXMLVal(line string) string {
	pos := strings.Index(line, ">") + 1
	line = line[pos:]
	pos = strings.Index(line, "<")
	return line[:pos]
}

func getXMLOvfAttr(line string, attrname string) string {
	pos := strings.Index(line, "<") + 1
	line = line[pos:]
	pos = strings.Index(line, ">")
	line = line[:pos]
	if line[len(line)-1:] == "/" {
		line = line[:len(line)-1]
	}
	attrs := strings.Split(line, "ovf:")
	for _, attr := range attrs {
		dattr := strings.Split(attr, "=")
		if len(dattr) != 2 {
			continue
		}
		k := dattr[0]
		k = strings.Trim(k, "\" \t\r")
		v := dattr[1]
		v = strings.Trim(v, "\" \t\r")
		if k == attrname {
			return v
		}
	}
	return ""
}

func getXMLAttr(line string, attrname string) string {
	pos := strings.Index(line, "<") + 1
	line = line[pos:]
	pos = strings.Index(line, ">")
	line = line[:pos]
	attrs := strings.Split(line, " ")
	for _, attr := range attrs {
		if pos := strings.Index(attr, "="); pos > 0 {
			a := strings.Split(attr, "=")
			if len(a) != 2 {
				continue
			}
			if a[0] == attrname {
				return strings.Replace(a[1], "\"", "", -1)
			}
		}
	}
	return ""
}

func setXMLVal(line string, val string) string {
	pos := strings.Index(line, ">") + 1
	pre := line[:pos]
	post := line[pos:]
	pos = strings.Index(post, "<")
	post = post[pos:]
	return fmt.Sprintf("%s%s%s", pre, val, post)
}

func getXMLTagVal(line string) string {
	pos := strings.Index(line, "<") + 1
	line = line[pos:]
	pos = strings.Index(line, ">")
	line = line[:pos]
	return line
}
