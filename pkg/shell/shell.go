package shell

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// Execute the given command.
func Execute(combinedOutput bool, format string, args ...interface{}) (string, error) {
	s := fmt.Sprintf(format, args...)
	// TODO: escape handling
	parts := strings.Split(s, " ")

	var p []string
	for i := 0; i < len(parts); i++ {
		if parts[i] != "" {
			p = append(p, parts[i])
		}
	}

	var argStrings []string
	if len(p) > 0 {
		argStrings = p[1:]
	}
	return ExecuteArgs(nil, combinedOutput, parts[0], argStrings...)
}

func ExecuteArgs(env []string, combinedOutput bool, name string, args ...string) (string, error) {
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		cmd := strings.Join(args, " ")
		cmd = name + " " + cmd
		logrus.Debugf("Executing command: %s", cmd)
	}

	c := exec.Command(name, args...)
	c.Env = env

	var b []byte
	var err error
	if combinedOutput {
		// Combine stderr and stdout in b.
		b, err = c.CombinedOutput()
	} else {
		// Just return stdout in b.
		b, err = c.Output()
	}

	if err != nil || !c.ProcessState.Success() {
		logrus.Debugf("Command[%s] => (FAILED) %s", name, string(b))
	} else {
		logrus.Debugf("Command[%s] => %s", name, string(b))
	}

	return string(b), err
}
