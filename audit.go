package audit

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
)

type AuditConfig struct {
	FlushInterval time.Duration
	BatchSize     int
	FilePath      string
	FileSize      int
	Level         LogLevel
}

type Audit struct {
	config AuditConfig

	file   *os.File
	writer *bufio.Writer
	queue  *Queue

	wg  sync.WaitGroup
	mtx sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

const (
	DefaultBatchSize     = 256               // 256 Messages
	DefaultFileSize      = 100 * 1024 * 1024 // 100 MB
	DefaultFlushInterval = 1 * time.Second

	MaxBatchSize = 512
)

func NewAudit(cfg AuditConfig) (*Audit, error) {
	// Set defaults
	if cfg.FilePath == "" {
		cfg.FilePath = "logs"
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultBatchSize
	} else if cfg.BatchSize > MaxBatchSize {
		cfg.BatchSize = MaxBatchSize
	}
	if cfg.FileSize <= 0 {
		cfg.FileSize = DefaultFileSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = DefaultFlushInterval
	}

	logDir := filepath.Dir(cfg.FilePath)
	if logDir != "" && logDir != "." {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}

	now := time.Now().Format("20060102_150405")
	f, err := os.OpenFile(cfg.FilePath+now, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", f.Name(), err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	audit := &Audit{
		config: cfg,
		file:   f,
		writer: bufio.NewWriter(f),
		queue:  NewQueue(),
		ctx:    ctx,
		cancel: cancel,
	}

	go startLogWriterService(audit)

	return audit, nil
}

func (audit *Audit) Close() {
	audit.cancel()
	audit.wg.Wait()

	if audit.queue.count != 0 {
		flush(audit) // Manual flush
	}

	audit.mtx.Lock()
	audit.file.Close()
	audit.mtx.Unlock()
}

func (audit *Audit) Trace(msg string) {
	if audit.config.Level <= TRACE {
		audit.log("\033[35mTRAC\033[m", msg)
	}
}

func (audit *Audit) Debug(msg string) {
	if audit.config.Level <= DEBUG {
		audit.log("\033[34mDEBU\033[m", msg)
	}
}

func (audit *Audit) Info(msg string) {
	if audit.config.Level <= INFO {
		audit.log("\033[92mINFO\033[m", msg)
	}
}

func (audit *Audit) Warn(msg string) {
	if audit.config.Level <= WARN {
		audit.log("\033[33mWARN\033[m", msg)
	}
}

func (audit *Audit) Error(msg string) {
	if audit.config.Level <= ERROR {
		audit.log("\033[31mERRO\033[m", msg)
	}
}

func (audit *Audit) Fatal(msg string) {
	audit.log("\033[35mFATA\033[m", msg)
	os.Exit(22)
}

func (audit *Audit) log(step string, msg string) {
	structured_msg := fmt.Sprintf("\033[1m[%s] %s \033[0m%s", time.Now().UTC().Format("2006-01-02 15:04:05"), step, msg)

	audit.queue.Append(structured_msg)

	fmt.Fprintf(os.Stdout, "%s \n", structured_msg)
}
