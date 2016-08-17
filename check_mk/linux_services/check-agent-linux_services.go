package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Sudo bool `long:"sudo" description:"Use sudo for all commands"`
}

func main() {
	fmt.Println("<<<linux_services>>>")

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("ERROR: ", err)

		os.Exit(1)
	}

	out, exitStatus, err := run("systemctl", "list-units", "--type", "service", "--no-legend", "--all")
	if exitStatus != 0 {
		fmt.Println("ERROR: ", err)

		os.Exit(1)
	}

	services := parseServicesStatus(string(out))

	for n, s := range services {
		fmt.Printf("%s\t%s\n", n, s)
	}
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

var parseServicesStatusLine = regexp.MustCompile(`^([^\s]+?)\s+([^\s]+?)\s+([^\s]+?)\s+([^\s]+?)\s+(.+)$`)

func parseServicesStatus(out string) map[string]string {
	services := map[string]string{}

	for _, l := range strings.Split(string(out), "\n") {
		l = strings.TrimSpace(l)

		if l == "" {
			continue
		}

		if m := parseServicesStatusLine.FindStringSubmatch(l); m != nil {
			services[m[1]] = m[4]
		}
	}

	return services
}
