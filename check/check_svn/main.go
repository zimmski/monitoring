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

func critical(msg string, err error) {
	fmt.Printf("CRITICAL | %s: %s\n", msg, err)

	os.Exit(2)
}

func runToStd(cmd ...string) (exitStatus int, err error) {
	c := exec.Command(cmd[0], cmd[1:]...)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

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

func main() {
	var opts struct {
		Repository string `long:"repository" required:"true"`
		User       string `long:"user" required:"true"`
		Password   string `long:"password" required:"true"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		critical("parsing arguments", err)
	}

	tmp, err := ioutil.TempDir("", "check_svn")
	if err != nil {
		critical("create temporary directory", err)
	}
	defer func() {
		err = os.RemoveAll(tmp)
		if err != nil {
			critical("remove temporary directory", err)
		}
	}()

	_, err = runToStd("svn", "checkout", "--username", opts.User, "--password", opts.Password, opts.Repository, tmp)
	if err != nil {
		critical("checkout repository", err)
	}

	err = os.Chdir(tmp)
	if err != nil {
		critical("change to temporary directory", err)
	}

	now := fmt.Sprintf("%d", time.Now().Unix())

	err = ioutil.WriteFile(tmp+"/test.txt", []byte(now), 0700)
	if err != nil {
		critical("write to test file", err)
	}

	_, err = runToStd("svn", "commit", "--username", opts.User, "--password", opts.Password, "--message", now)
	if err != nil {
		critical("commit change", err)
	}

	_, err = runToStd("svn", "up", "--username", opts.User, "--password", opts.Password)
	if err != nil {
		critical("update repository", err)
	}

	fmt.Printf("OK - can read from and write to repository %s\n", opts.Repository)

	os.Exit(0)
}
