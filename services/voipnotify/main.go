package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
)

const revision = 15

var opts struct {
	AllowedRecipients       string `long:"allowed-recipients" required:"true" default:"allowed_recipients.csv" description:"CSV file with identifier and number colums for each allowed recipient"`
	MonitoringMessageFile   string `long:"monitoring-message-file" required:"true" default:"/var/lib/asterisk/sounds/monitoring-message.gsm" description:"Defines where the monitoring message should be generate to"`
	MonitoringCallDirectory string `long:"monitoring-call-directory" required:"true" default:"/var/spool/asterisk/outgoing/" description:"Defines where the monitoring calls should be generate to"`
	Trunk                   string `long:"trunk" required:"true" default:"trunk_1" description:"The SIP trunk"`
	ServerPort              uint   `long:"port" required:"true" default:"8080" description:"The HTTP server port"`
}

var allowedRecipients map[string]string
var allowedRecipientsLock sync.Mutex

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		exitError(err)
	}

	allowedRecipients, err = parseAllowedRecipients(opts.AllowedRecipients)
	if err != nil {
		exitError(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/hello", HandleHello)
	r.HandleFunc("/notification", HandleNotification)
	r.HandleFunc("/update_recipients", HandleUpdateRecipients)

	log.Printf("Start server on port %d.", opts.ServerPort)
	log.Printf(`Use "kill %d" to stop the server.`, os.Getpid())

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", opts.ServerPort), r))
}

func HandleHello(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		err = writeError(w, err)
		if err != nil {
			log.Printf("ERROR while writing parse form error for HandleHello: %v", err)
		}

		return
	}

	name := r.Form.Get("name")
	if name == "" {
		name = "no-name-defined"
	}

	err = writeResponse(w, fmt.Sprintf("Hello %s!", name), fmt.Sprintf("<h1>Hello %s! We are running revision %d.</h1>", name, revision))
	if err != nil {
		log.Printf("ERROR while writing response for HandleHello: %v", err)
	}
}

func HandleNotification(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		err = writeError(w, err)
		if err != nil {
			log.Printf("ERROR while writing parse form error for HandleNotification: %v", err)
		}

		return
	}

	msg := r.Form.Get("message")
	if msg == "" {
		err = writeError(w, fmt.Errorf("No message defined"))
		if err != nil {
			log.Printf("ERROR while writing no message error for HandleNotification: %v", err)
		}

		return
	}

	var numbers []string
	rs := r.Form["recipient"]

	var content bytes.Buffer
	content.WriteString("<h1>Notified</h1>")
	content.WriteString("<ul>")
	for _, i := range rs {
		n, ok := allowedRecipients[i]
		if ok {
			content.WriteString(fmt.Sprintf("<li>%s: %s</li>", i, n))

			numbers = append(numbers, n)
		}
	}
	content.WriteString("</ul>")

	if len(numbers) == 0 {
		err = writeResponse(w, "Sent no notifications", "<h1>Notified NOBODY! There is something wrong with your receivers.</h1>")
		if err != nil {
			log.Printf("ERROR while writing response for HandleNotification: %v", err)
		}

		return
	}

	// TODO make the creation of the monitoring message race free. The problem is, that we need a cleanup mechanism for the message.
	err = convertTextToSpeech(msg, opts.MonitoringMessageFile)
	if err != nil {
		err = writeError(w, err)
		if err != nil {
			log.Printf("ERROR while writing convertTextToSpeech error for HandleNotification: %v", err)
		}

		return
	}

	err = generateCalls(opts.MonitoringCallDirectory, opts.MonitoringMessageFile, opts.Trunk, numbers)
	if err != nil {
		err = writeError(w, err)
		if err != nil {
			log.Printf("ERROR while writing generateCalls error for HandleNotification: %v", err)
		}

		return
	}

	err = writeResponse(w, "Sent notifications", content.String())
	if err != nil {
		log.Printf("ERROR while writing response for HandleNotification: %v", err)
	}
}

func convertTextToSpeech(msg string, filename string) error {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			log.Printf("ERROR cannot remove %s: %v", f.Name(), err)
		}
	}()

	_, err = f.WriteString(msg)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	out, _, err := run("espeak", "-f", f.Name(), "-s", "130", "-a", "150", "-w", filename+".wav")
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		return err
	}

	out, _, err = run("sox", filename+".wav", "-r", "8000", "-c", "1", filename)
	if len(out) > 0 {
		log.Println(string(out))
	}
	if err != nil {
		return err
	}

	return nil
}

func generateCalls(directory string, messageFile string, trunk string, numbers []string) error {
	notificationID := time.Now().UnixNano()

	ext := filepath.Ext(messageFile)
	dataName := strings.TrimSuffix(filepath.Base(messageFile), ext)

	for i, n := range numbers {
		call := fmt.Sprintf("%s/%d-%d.call", directory, notificationID, i)

		err := ioutil.WriteFile(call, []byte(fmt.Sprintf(`Channel: SIP/%s/%s
Application: Playback
Data: %s
`, trunk, n, dataName)), 0640)
		if err != nil {
			return err
		}

		log.Printf("Called with %s", call)
	}

	return nil
}

func HandleUpdateRecipients(w http.ResponseWriter, r *http.Request) {
	allowedRecipientsLock.Lock()
	defer allowedRecipientsLock.Unlock()

	allowedRecipients, err := parseAllowedRecipients(opts.AllowedRecipients)
	if err != nil {
		err = writeError(w, err)
		if err != nil {
			log.Printf("ERROR while reading allowed recipients for HandleUpdateRecipients: %v", err)
		}

		return
	}

	var content bytes.Buffer
	content.WriteString("<h1>Updated recipients</h1>")
	content.WriteString("<ul>")
	for i, n := range allowedRecipients {
		content.WriteString(fmt.Sprintf("<li>%s: %s</li>", i, n))
	}
	content.WriteString("</ul>")

	err = writeResponse(w, "Updated recipients", content.String())
	if err != nil {
		log.Printf("ERROR while writing response for HandleUpdateRecipients: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) error {
	w.WriteHeader(http.StatusInternalServerError)

	log.Printf("ERROR: %s", err.Error())

	return writeResponse(w, "ERROR", err.Error())
}

func writeResponse(w http.ResponseWriter, header string, content string) error {
	_, err := w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>` + header + `</title>
	</head>
	<body>
		` + content + `
	</body>
</html>`))

	return err
}

func parseAllowedRecipients(filename string) (map[string]string, error) {
	out, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	rsr, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rsr) < 2 {
		return nil, fmt.Errorf("No header or no recipients defined")
	}
	if len(rsr[0]) != 2 {
		return nil, fmt.Errorf("The header row must have exactly two columns")
	}

	if rsr[0][0] != "identifier" || rsr[0][1] != "number" {
		return nil, fmt.Errorf(`The header row must consist of the columns "identifier" and "number"`)
	}

	rsr = rsr[1:]

	rs := map[string]string{}
	for _, r := range rsr {
		if len(r) != 2 {
			return nil, fmt.Errorf("Data rows must have exactly two columns")
		}

		r[0] = strings.TrimSpace(r[0])
		r[1] = strings.TrimSpace(r[1])

		if r[0] == "" || r[1] == "" {
			return nil, fmt.Errorf("Data columns cannot be empty")
		}

		rs[r[0]] = strings.Replace(r[1], "\n", "", -1)
	}

	return rs, nil
}

func exitError(err error) {
	fmt.Println("ERROR: ", err)

	os.Exit(1)
}

func run(cmd ...string) (out []byte, exitStatus int, err error) {
	if len(cmd) == 0 {
		return nil, 0, fmt.Errorf("No cmd defined")
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
