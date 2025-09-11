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
	TRACE LogLevel = iota + 1
	DEBUG
	INFO
	WARN
	ERROR
)

type AuditConfig struct {
	// How often queue messages will be written to file
	FlushInterval time.Duration

	// How many messages will sit in the queue before they are all written at once
	BatchSize int

	// The file path with the name (e.g. resources/logs/log.txt)
	FilePath string

	// Max file size in bytes
	FileSize int

	// LogLevel. Default is INFO
	Level LogLevel

	// How many messages (string) the queue can hold at any given moment
	QueueSize int
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
	DefaultBatchSize     = 256                // 256 Messages
	DefaultFileSize      = 1024 * 1024 * 1024 // 1 GB
	DefaultFlushInterval = 250 * time.Millisecond
	DefaultQueueSize     = 1024

	MaxBatchSize = 512
)

// Create a new Audit. Starts a background thread that periodically writes to file.
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
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = DefaultQueueSize
	}
	if cfg.Level <= 0 {
		cfg.Level = INFO
	}

	var audit *Audit

	if cfg.Level <= DEBUG {
		ctx, cancel := context.WithCancel(context.Background())
		audit = &Audit{
			config: cfg,
			file:   nil,
			writer: bufio.NewWriter(os.Stdout),
			queue:  NewQueue(cfg.QueueSize),
			ctx:    ctx,
			cancel: cancel,
		}
	} else {
		f, err := openFile(cfg)
		if err != nil {
			return nil, fmt.Errorf("couldn't open file: %w", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		audit = &Audit{
			config: cfg,
			file:   f,
			writer: bufio.NewWriter(f),
			queue:  NewQueue(cfg.QueueSize),
			ctx:    ctx,
			cancel: cancel,
		}
		go startLogWriterService(audit)
	}

	return audit, nil
}

func openFile(cfg AuditConfig) (*os.File, error) {
	logDir := filepath.Dir(cfg.FilePath)
	if logDir != "" && logDir != "." {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}

	// If a file already exists, rename it, instead of opening and appending to it
	_, err := os.Stat(cfg.FilePath)
	if !os.IsNotExist(err) {
		now := time.Now().Format("20060102_150405")
		if err := os.Rename(cfg.FilePath, cfg.FilePath+now); err != nil {
			return nil, fmt.Errorf("failed to rename old log file: %w", err)
		}
	}

	f, err := os.OpenFile(cfg.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", f.Name(), err)
	}

	return f, nil
}

// Gracefully shuts down the auditer and flushes all messages from the queue to file before exiting.
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

func (audit *Audit) Trace(msg string, a ...any) {
	if audit.config.Level <= TRACE {
		audit.log("\033[35mTRAC\033[m", msg, a...)
	}
}

func (audit *Audit) Debug(msg string, a ...any) {
	if audit.config.Level <= DEBUG {
		audit.log("\033[34mDEBU\033[m", msg, a...)
	}
}

func (audit *Audit) Info(msg string, a ...any) {
	if audit.config.Level <= INFO {
		audit.log("\033[92mINFO\033[m", msg, a...)
	}
}

func (audit *Audit) Warn(msg string, a ...any) {
	if audit.config.Level <= WARN {
		audit.log("\033[33mWARN\033[m", msg, a...)
	}
}

func (audit *Audit) Error(msg string, a ...any) {
	if audit.config.Level <= ERROR {
		audit.log("\033[31mERRO\033[m", msg, a...)
	}
}

func (audit *Audit) Fatal(msg string, a ...any) {
	audit.log("\033[35mFATA\033[m", msg, a...)
	os.Exit(22)
}

func (audit *Audit) log(step string, msg string, a ...any) {
	log_prefix := fmt.Sprintf("\033[1m[%s] %s \033[0m", time.Now().UTC().Format("2006-01-02 15:04:05"), step)
	log_data := fmt.Sprintf(msg, a...)

	log_message_parsed := log_prefix + log_data

	audit.queue.Append(log_message_parsed)

	fmt.Println(log_message_parsed)
}
