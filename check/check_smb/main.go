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
	Share    string `long:"share" required:"true"`
	User     string `long:"user" required:"true"`
	Domain   string `long:"domain" required:"true"`
	Password string `long:"password" required:"true"`
	Verbose  bool   `long:"verbose"`
}

func runToStd(cmd ...string) (exitStatus int, err error) {
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

	tmp, err := ioutil.TempDir("", "check_smb")
	if err != nil {
		return critical("create temporary directory", err)
	}
	defer func() {
		err = os.RemoveAll(tmp)
		if err != nil {
			exitStatus = critical("remove temporary directory", err)
		}
	}()

	_, err = runToStd("sudo", "/sbin/mount.cifs", opts.Share, tmp, "-o", "username="+opts.User+",password="+opts.Password+",domain="+opts.Domain)
	if err != nil {
		return critical("mount share", err)
	}

	err = os.Chdir(tmp)
	if err != nil {
		critical("change to temporary directory", err)
	}

	now := fmt.Sprintf("%d", time.Now().Unix())

	err = ioutil.WriteFile(tmp+"/test.txt", []byte(now), 0700)
	if err != nil {
		return critical("write to test file", err)
	}

	c, err := ioutil.ReadFile(tmp + "/test.txt")
	if err != nil {
		return critical("read from written test file", err)
	}

	if string(c) != now {
		return critical("written content does not match", err)
	}

	_, err = runToStd("sudo", "umount", "-t", "cifs", "-l", tmp)
	if err != nil {
		return critical("umount share", err)
	}

	fmt.Printf("OK - can read from and write to share %s\n", opts.Share)

	return 0
}

func main() {
	os.Exit(cmd())
}
