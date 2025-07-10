package audit

import (
	"io"
	"log"
	"os"
	"regexp"
	"sync"
	"time"
)

var wait *sync.WaitGroup

type Audit struct {
	logger log.Logger
	level  int
	writer io.Writer
	format string
	emoji  bool
}

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
)

func NewAudit() *Audit {
	wait = &sync.WaitGroup{}

	a := new(Audit)
	a.logger = *log.New(os.Stderr, "", 0)
	a.level = int(INFO)
	a.format = "[2006-01-02 15:04:05]"
	a.emoji = true // This will probably not be modifiable
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
	wait.Add(1)
	if audit.level <= int(TRACE) {
		go audit.logg("\033[95m🔎TRAC\033[m", msg)
	}
}

func (audit *Audit) Debug(msg string) {
	wait.Add(1)
	if audit.level <= int(DEBUG) {
		go audit.logg("\033[95m🐛DEBU\033[m", msg)
	}
}

func (audit *Audit) Info(msg string) {
	wait.Add(1)
	go audit.logg("\033[92m👋INFO\033[m", msg)
}

func (audit *Audit) Warn(msg string) {
	wait.Add(1)
	go audit.logg("\033[33m⚠WARN\033[m", msg)
}

func (audit *Audit) Error(msg string) {
	/* Send some sort of alert here as well eventually */
	wait.Add(1)
	go audit.logg("\033[31m❌ERRO\033[m", msg)
}

func (audit *Audit) Fatal(msg string) {
	wait.Add(1)
        audit.logg("\033[35m☠FATA\033[m", msg)
	os.Exit(22)
}

func (audit *Audit) logg(step, msg string) {
	defer wait.Done()
	pattern, _ := regexp.Compile("\r?\n") // Not catostrophic if this fails, so ignore it
	msg = pattern.ReplaceAllString(msg, " ")
	audit.logger.Printf("\033[1m%s %s \033[1m%s", time.Now().UTC().Format(audit.format), step, msg)
}
