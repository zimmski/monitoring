package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mxk/go-imap/imap"
)

const (
	returnOk      = 0
	returnWarning = 1
	returnError   = 2
	returnHelp    = 3
)

type options struct {
	General struct {
		Help    bool `long:"help" description:"Show this help message"`
		Verbose bool `long:"verbose" description:"Verbose log output"`
	} `group:"General options"`

	IMAP struct {
		Server   string `long:"server" description:"IMAP server address with port"`
		User     string `long:"user" description:"IMAP user"`
		Password string `long:"password" description:"IMAP password"`
		TLS      bool   `long:"tls" description:"Use TLS for the IMAP connection"`
		StartTLS bool   `long:"starttls" description:"Use STARTTLS for the IMAP connection"`
		Mailbox  string `long:"mailbox" description:"The mailbox that should be queried" default:"INBOX"`

		ListMailboxes bool `long:"list-mailboxes" description:"List all mail boxes and exit."`
	} `group:"IMAP options"`

	Monitoring struct {
		Warning uint32 `long:"warning" description:"If higher a monitoring warning is issued." default:"40"`
		Error   uint32 `long:"error" description:"If higher a monitoring error is issued." default:"50"`
	} `group:"Monitoring options"`
}

func checkArguments(args []string, opts *options) (bool, int) {
	p := flags.NewNamedParser("check_imap_mailbox_count", flags.None)

	if _, err := p.AddGroup("check_imap_mailbox_count", "check_imap_mailbox_count arguments", opts); err != nil {
		return true, exitError(err.Error())
	}

	_, err := p.ParseArgs(args)
	if opts.General.Help || len(args) == 0 {
		p.WriteHelp(os.Stdout)

		return true, returnHelp
	}

	if err != nil {
		return true, exitError(err.Error())
	}

	return false, 0
}

func exitError(format string, args ...interface{}) int {
	fmt.Fprintf(os.Stderr, format+"\n", args...)

	return returnError
}

func main() {
	var opts = &options{}

	if exit, exitCode := checkArguments(os.Args[1:], opts); exit {
		os.Exit(exitCode)
	}

	var client *imap.Client
	var err error

	if opts.General.Verbose {
		imap.DefaultLogger = log.New(os.Stdout, "", 0)
		imap.DefaultLogMask = imap.LogConn | imap.LogRaw
	}

	if opts.IMAP.TLS {
		client, err = imap.DialTLS(opts.IMAP.Server, nil)
	} else {
		client, err = imap.Dial(opts.IMAP.Server)
	}
	checkIMAPCmdError(err)

	if opts.IMAP.StartTLS {
		checkIMAPCmd(client.StartTLS(nil))
	}

	checkIMAPCmd(client.Login(opts.IMAP.User, opts.IMAP.Password))

	var count uint32 = 0

	if opts.IMAP.ListMailboxes {
		c, err := imap.Wait(client.List("", "*"))
		checkIMAPCmdError(err)

		for _, m := range c.Data {
			fmt.Printf("%s\n", m.MailboxInfo().Name)
		}

		os.Exit(returnError)
	} else {
		_, err := imap.Wait(client.Select(opts.IMAP.Mailbox, true))
		checkIMAPCmdError(err)

		count = client.Mailbox.Messages
	}

	checkIMAPCmd(client.Logout(5 * time.Second))

	if count > opts.Monitoring.Error {
		fmt.Printf("ERROR - Message count in %q is %d\n", opts.IMAP.Mailbox, count)

		os.Exit(returnError)
	} else if count > opts.Monitoring.Error {
		fmt.Printf("WARNING - Message count in %q is %d\n", opts.IMAP.Mailbox, count)

		os.Exit(returnWarning)
	} else {
		fmt.Printf("OK - Message count in %q is %d\n", opts.IMAP.Mailbox, count)

		os.Exit(returnOk)
	}
}

func checkIMAPCmd(cmd *imap.Command, err error) {
	checkIMAPCmdError(err)
}

func checkIMAPCmdError(err error) {
	if err != nil {
		os.Exit(exitError(err.Error()))
	}
}
