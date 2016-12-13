package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Sudo bool `long:"sudo" description:"Use sudo for all commands"`
}

func main() {
	fmt.Println("<<<sip>>>")

	_, err := flags.Parse(&opts)
	if err != nil {
		exitError(err)
	}

	out, exitStatus, err := run("asterisk", "-rv", "-x", "sip show registry")
	if exitStatus != 0 {
		exitError(err)
	}

	registrations, err := parseSIPRegistryStatus(string(out))
	if exitStatus != 0 {
		exitError(err)
	}

	fmt.Printf("registrations\t%d\t", len(registrations))
	for i, r := range registrations {
		if i != 0 {
			fmt.Print(",")
		}
		fmt.Print(r)
	}
	fmt.Println()
}

func exitError(err error) {
	fmt.Println("ERROR: ", err)

	os.Exit(1)
}

func run(cmd ...string) (out []byte, exitStatus int, err error) {
	if len(cmd) == 0 {
		return nil, 0, fmt.Errorf("No cmd defined")
	}

	if opts.Sudo {
		cmd = append([]string{"sudo"}, cmd...)
	}

	c := exec.Command(cmd[0], cmd[1:]...)

	out, err = c.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return out, status.ExitStatus(), err
			}
		}

		return out, 0, err
	}

	return out, 0, nil
}

var parseSIPRegistryStatusRegistrationCount = regexp.MustCompile(`(?s)(\d+) SIP registrations`)
var parseSIPRegistryStatusRowRe = regexp.MustCompile(`^([^\s]+?)\s+[^\s]+?\s+([^\s]+?)\s+[^\s]+?\s+([^\s]+?)\s+(.+)$`)

func parseSIPRegistryStatus(out string) ([]string, error) {
	rc := parseSIPRegistryStatusRegistrationCount.FindStringSubmatch(out)
	if rc == nil {
		return nil, nil
	}

	registrationCount, err := strconv.Atoi(rc[1])
	if err != nil {
		return nil, err
	}
	if registrationCount == 0 {
		return nil, nil
	}

	var rs []string

	// Remove the first (headers) and last (summary) line.
	lines := strings.Split(out, "\n")
	lines = lines[1 : len(lines)-2]

	for _, l := range lines {
		l = strings.TrimSpace(l)

		if m := parseSIPRegistryStatusRowRe.FindStringSubmatch(l); m != nil {
			if m[3] == "Registered" {
				rs = append(rs, fmt.Sprintf("%s->%s", m[1], m[2]))
			}
		}
	}

	sort.Strings(rs)

	return rs, nil
}
