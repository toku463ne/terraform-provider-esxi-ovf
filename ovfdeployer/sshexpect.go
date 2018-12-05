package ovfdeployer

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jamesharr/expect"
	"github.com/pkg/errors"
)

type sshExpect struct {
	hostIP, user, password string
	UserS, PromptS, PassS  string
	YesnoS                 string
	e                      expect.Expect
	testFile               string
	Result                 string
}

func newSSHExpect(hostIP, user, pass, testFile string) *sshExpect {
	si := new(sshExpect)
	si.PromptS = `[\$\#\]] `
	si.UserS = `[Uu]sername:`
	si.PassS = `[Pp]assword:|[Yy]es/[Nn]o`
	si.hostIP = hostIP
	si.user = user
	si.password = pass
	if testFile != "" {
		si.testFile = testFile
	}
	return si
}

func (se *sshExpect) login() error {
	if se.testFile != "" {
		return nil
	}

	cmd := fmt.Sprintf("%s@%s", se.user, se.hostIP)
	e, err := expect.Spawn(ssh, cmd)
	if err != nil {
		return errors.Wrapf(err, cmd)
	}
	e.SetTimeout(sshTimeout)
	se.e = *e
	match, err := se.e.Expect(se.PassS)
	if err != nil {
		return errors.Wrapf(err, "Error when sending password to login %s.", se.hostIP)
	}
	if matched, err := regexp.MatchString("[Yy]es/[Nn]o", match.Groups[0]); matched {
		se.e.SendLn("yes")
		match, err = se.e.Expect(se.PassS)
		if err != nil {
			return errors.Wrapf(err, "Error when sending password to login %s.", se.hostIP)
		}
	} else if err != nil {
		return errors.Wrapf(err, "Error when sending 'yes' before login")
	}
	if matched, err := regexp.MatchString(".*[Pp]assword:.*", match.Groups[0]); err != nil {
		return errors.Wrapf(err, "Error getting password prompt")
	} else if matched == false {
		return errors.Errorf("Not expected responce. Got=%s, Expected=*[Pp]assword:", match.Before)
	}

	se.e.SendLn(se.password)
	_, err = e.Expect(se.PromptS)
	if err != nil {
		return errors.Wrapf(err, "EXPECT could not get prompt")
	}

	cmds := []string{
		"export LC_ALL=en_US.UTF-8",
		"export LANG=en_US.UTF-8",
		"export LANGUAGE=en_US.UTF-8",
	}
	for _, cmd := range cmds {
		se.e.SendLn(cmd)
		_, err = e.Expect(se.PromptS)
		if err != nil {
			return errors.Wrapf(err, "EXPECT could not get prompt sending %s", cmd)
		}
	}

	return nil
}

func getSSHExpectConn(hostIP string, user string, pass, testFile string) (*sshExpect, error) {
	se := newSSHExpect(hostIP, user, pass, testFile)
	err := se.login()
	return se, err
}

func (se *sshExpect) run(cmdName, cmdFormat string, args ...interface{}) ([]string, error) {
	cmdstr := fmt.Sprintf(cmdFormat, args...)
	logDebug("%s # %s", se.hostIP, cmdstr)
	if se.testFile != "" {
		sh := fmt.Sprintf("%s/siml.sh", getTestDir())
		cmd := exec.Command(sh, se.testFile, cmdName)
		cmd.Env = append(os.Environ(),
			"LC_ALL=en_US.UTF-8",
			"LANG=en_US.UTF-8",
			"LANGUAGE=en_US.UTF-8")
		out, err := cmd.Output()
		if err != nil {
			return nil, errors.Wrapf(err, "Error executing: %s %s %s", sh, se.testFile, cmdFormat)
		}
		s := strings.Split(string(out), "\n")
		t := make([]string, 0)
		for _, v := range s {
			if v != "" {
				t = append(t, v)
			}
		}
		if len(t) < 1 {
			return nil, errors.Errorf("Non expected result from %s %s %s", sh, se.testFile, cmdName)
		}
		return t[1:], nil
	}
	se.e.SendLn(cmdstr)
	match, err := se.e.Expect(se.PromptS)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get prompt after command: %s", cmdstr)
	}
	s := strings.Split(match.Before, "\n")
	if len(s) == 0 {
		return nil, errors.New("Could not get value from command line")
	}
	/*
		a := removeBlanks(s[0])
		i := 1
		scmd := removeBlanks(cmd)
		for a != scmd && i < len(s) {
			a += removeBlanks(s[i])
			i++
		}
		if a != scmd {
			return nil, errors.New(fmt.Sprintf("Unexpected res from expect. got=%+v", match.Before))
		}
	*/

	logDebug(">%+v", strings.Join(s, "\n"))
	return s[1 : len(s)-1], nil

}

func (se *sshExpect) close() error {
	se.e.SendLn("exit")
	return se.e.ExpectEOF()
}
