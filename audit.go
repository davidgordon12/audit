package audit

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

type Audit struct {
	logger log.Logger
}

type AuditType int

func NewAudit() *Audit {
	a := new(Audit)
	a.logger = *log.New(os.Stderr, "", 0)
	return a
}

func (audit *Audit) addFile(path string) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Couldn't open file " + path)
		return
	}
	wrt := io.MultiWriter(os.Stdout, f)
	audit.logger.SetOutput(wrt)
}

func (audit *Audit) info(msg string) {
	go logg("INFO", msg)
}

func (audit *Audit) warn(msg string) {
	go logg("WARNING", msg)
}

func (audit *Audit) error(msg string) {
	/* Send some sort of alert here as well eventually */
	go logg("ERROR", msg)
}

func logg(step, msg string) {
	pattern, _ := regexp.Compile("\r?\n")
	msg = pattern.ReplaceAllString(msg, " ")
	audit.logger.Printf("%s %s: %s", time.Now().UTC().Format("[2006-01-02 15:04:05] "), step, msg)
}
