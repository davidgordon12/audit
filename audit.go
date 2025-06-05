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

func (audit *Audit) AddFile(path string) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Couldn't open file " + path)
		return
	}
	wrt := io.MultiWriter(os.Stdout, f)
	audit.logger.SetOutput(wrt)
}

func (audit *Audit) Info(msg string) {
	go audit.logg("INFO", msg)
}

func (audit *Audit) Warn(msg string) {
	go audit.logg("WARNING", msg)
}

func (audit *Audit) Error(msg string) {
	/* Send some sort of alert here as well eventually */
	go audit.logg("ERROR", msg)
}

func (audit *Audit) logg(step, msg string) {
	pattern, _ := regexp.Compile("\r?\n")
	msg = pattern.ReplaceAllString(msg, " ")
	audit.logger.Printf("%s %s: %s", time.Now().UTC().Format("[2006-01-02 15:04:05] "), step, msg)
}
