package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		OlderThan int `long:"older-than" required:"true" description:"Remove every spam mail older than (in seconds)"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("ERROR: ", err)

		os.Exit(1)
	}

	if opts.OlderThan <= 0 {
		fmt.Println("ERROR: older-than must be at least 1")

		os.Exit(1)
	}

	listRaw, err := exec.Command("/usr/sbin/exiqgrep", "-i", "-o", strconv.Itoa(opts.OlderThan)).CombinedOutput()
	if err != nil {
		fmt.Println("ERROR: cannot execute exiqgrep:", err)

		os.Exit(2)
	}

MESSAGE:
	for _, messageID := range strings.Split(string(listRaw), "\n") {
		if strings.Trim(messageID, " \n\r\t") == "" {
			continue MESSAGE
		}

		if strings.ContainsAny(messageID, " |") {
			fmt.Println("ERROR: Message ID is malicious:", messageID)

			continue MESSAGE
		}

		messageCmd := exec.Command("/usr/sbin/exim", "-Mvh", messageID)

		messageReader, err := messageCmd.StdoutPipe()
		if err != nil {
			fmt.Println("ERROR: cannot create pipe of exim -Mvh", messageID, ":", err)

			os.Exit(3)
		}

		message := bufio.NewReader(messageReader)

		go func() {
			if err := messageCmd.Run(); err != nil {
				fmt.Println("ERROR: cannot run exim -Mvh", messageID, ":", err)

				os.Exit(3)
			}
		}()

	SUBJECT:
		for {
			line, _, err := message.ReadLine()
			if err != nil {
				if err == io.EOF {
					break SUBJECT
				}

				fmt.Println("ERROR: cannot read line of", messageID, ":", err)

				continue MESSAGE
			}

			s := string(line)
			if strings.Contains(s, "Subject:") {
				if strings.Contains(s, "***SPAM***") {
					// fmt.Println(messageID, "will be removed, contains Spam")

					removeMail(messageID)

					continue MESSAGE
				}

				break SUBJECT
			}
		}

		logsRaw, err := exec.Command("/usr/sbin/exim", "-Mvl", messageID).CombinedOutput()
		if err != nil {
			fmt.Println("ERROR: cannot execute exim -Mvl:", err)

			continue MESSAGE
		}
		logs := string(logsRaw)

		if (strings.Contains(logs, "SMTP error from remote mail server after RCPT TO") && (strings.Contains(logs, ": 4") || strings.Contains(logs, ": 5"))) || // we do not care which 4xx or 5xx status this is since it was already established that its a RCPT TO error
			strings.Contains(logs, ": 550-5.7.1") || strings.Contains(logs, ": 550 5.7.1") || strings.Contains(logs, ": 550 5.1.1 User unknown") ||
			strings.Contains(logs, ": 554-5.7.1") || strings.Contains(logs, ": 554 5.7.1") || strings.Contains(logs, ": 554 delivery error") || strings.Contains(logs, ": 554 Denied") ||
			strings.Contains(logs, "550 <> Sender rejected") || strings.Contains(logs, ": 550 5.5.0 Sender domain is empty.") || strings.Contains(logs, ": 550 Error: no third-party DSNs") || strings.Contains(logs, ": 550 permanent failure for one or more recipients") ||
			strings.Contains(logs, "Connection timed out") || strings.Contains(logs, "Connection refused") ||
			strings.Contains(logs, "retry time not reached for any host after a long failure period") {
			removeMail(messageID)

			continue MESSAGE
		}
	}
}

func removeMail(messageID string) {
	messageRm, err := exec.Command("/usr/sbin/exim", "-Mrm", messageID).CombinedOutput()
	if err != nil {
		fmt.Println("ERROR: cannot execute message remove of", messageID, ":", err)
	}

	s := string(messageRm)
	if l := strings.Trim(s, " \n\r\t"); l != "" {
		if !strings.Contains(s, "has been removed") {
			fmt.Println(l)
		}
	}
}
