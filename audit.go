package audit

import (
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

type Audit struct {
	logger log.Logger
	level  int
	writer io.Writer
	format string
}

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
)

func NewAudit() *Audit {
	a := new(Audit)
	a.logger = *log.New(os.Stderr, "", 0)
	a.level = int(INFO)
	a.format = "[2006-01-02 15:04:05]"
	return a
}

func (audit *Audit) AddFile(path string) (*Audit, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		audit.Error("Couldn't open file " + path)
		return nil, err
	}
	defer f.Close()

	audit.writer = io.MultiWriter(os.Stdout, f)
	audit.logger.SetOutput(audit.writer)
	return audit, nil
}

func (audit *Audit) Level(level LogLevel) *Audit {
	audit.level = int(level) // Info level is the default log level
	return audit
}

func (audit *Audit) DateFormat(format string) *Audit {
	audit.format = format
	return audit
}

func (audit *Audit) Trace(msg string) {
	if audit.level <= int(TRACE) {
		go audit.logg("\033[95mTRAC\033[m", msg)
	}
}

func (audit *Audit) Debug(msg string) {
	if audit.level <= int(DEBUG) {
		go audit.logg("\033[95mDEBU\033[m", msg)
	}
}

func (audit *Audit) Info(msg string) {
	go audit.logg("\033[92mINFO\033[m", msg)
}

func (audit *Audit) Warn(msg string) {
	go audit.logg("\033[33mWARN\033[m", msg)
}

func (audit *Audit) Error(msg string) {
	/* Send some sort of alert here as well eventually */
	go audit.logg("\033[31mERRO\033[m", msg)
}

func (audit *Audit) logg(step, msg string) {
	pattern, _ := regexp.Compile("\r?\n") // Not catostrophic if this fails, so ignore it
	msg = pattern.ReplaceAllString(msg, " ")
	audit.logger.Printf("\033[1m%s %s \033[1m%s", time.Now().UTC().Format(audit.format), step, msg)
}
