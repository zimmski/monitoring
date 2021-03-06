package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

func critical(msg string, err error) int {
	fmt.Printf("CRITICAL | %s: %s\n", msg, err)

	return 2
}

var opts struct {
	Repository string `long:"repository" required:"true" description:"The repository URL which that should be tested."`
	User       string `long:"user" required:"true" description:"The user to access the repository."`
	Password   string `long:"password" required:"true" description:"The password for the user."`
	Message    string `long:"message" description:"The commit message which should be used for the commit to the repository. If not set the current Unix time will be used."`
	Verbose    bool   `long:"verbose" description:"Activate log messages and forward STDOUT and STDERR of executed commands."`
}

func runToStd(cmd ...string) (exitStatus int, err error) {
	if opts.Verbose {
		fmt.Printf("%v\n", cmd)
	}

	c := exec.Command(cmd[0], cmd[1:]...)

	if opts.Verbose {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}

	err = c.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), err
			}
		}

		return 0, err
	}

	return 0, nil
}

func cmd() (exitStatus int) {
	_, err := flags.Parse(&opts)
	if err != nil {
		return critical("parsing arguments", err)
	}

	tmp, err := ioutil.TempDir("", "check_svn")
	if err != nil {
		return critical("create temporary directory", err)
	}
	defer func() {
		err = os.RemoveAll(tmp)
		if err != nil {
			exitStatus = critical("remove temporary directory", err)
		}
	}()

	_, err = runToStd("svn", "checkout", "--username", opts.User, "--password", opts.Password, opts.Repository, tmp)
	if err != nil {
		return critical("checkout repository", err)
	}

	err = os.Chdir(tmp)
	if err != nil {
		return critical("change to temporary directory", err)
	}

	if opts.Message == "" {
		opts.Message = fmt.Sprintf("%d", time.Now().Unix())
	}

	err = ioutil.WriteFile(tmp+"/test.txt", []byte(opts.Message), 0700)
	if err != nil {
		return critical("write to test file", err)
	}

	_, err = runToStd("svn", "commit", "--username", opts.User, "--password", opts.Password, "--message", opts.Message)
	if err != nil {
		return critical("commit change", err)
	}

	_, err = runToStd("svn", "up", "--username", opts.User, "--password", opts.Password)
	if err != nil {
		return critical("update repository", err)
	}

	fmt.Printf("OK - can read from and write to repository %s\n", opts.Repository)

	return 0
}

func main() {
	os.Exit(cmd())
}
