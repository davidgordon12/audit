package audit

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"time"
	"sync"
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
	BatchSize int
	FilePath string
	FileSize int
	Level LogLevel
}

type Audit struct {
	config AuditConfig

	file *os.File
	writer *bufio.Writer
	queue  *Queue

	wg sync.Wait
	mu sync.Mutex

	ctx context.Context
	cancel context.CancelFunc
}

const (
	DefaultBatchSize = 256 // 256 Messages
	DefaultFileSize = 100 * 1024 * 1024 // 100 MB
	DefaultFlushInterval = 1 * time.Second

	MaxBatchSize = 512
)

func NewAudit(cfg AuditConfig) (*Audit, Error) {
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
	
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(cfg.LogFilePath)
	if logDir != "" && logDir != "." { // Don't try to create "" or "."
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}
	}
	
	now = time.Now().Format("20060102_150405")
	f, err := os.OpenFile(cfg.FilePath + now, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", f, err)
	}

	audit := &Audit {
		config: cfg,
		file: f,
		writer: bufio.NewWriter(f)
		queue: NewQueue(),
	}

	audit.wg.Add(1)
	go func(a *Audit) {
		defer wait.Done()
		startLogWriterService(a)
	}(audit)

	return audit, nil
}

func (audit *Audit) Close() {
	audit.cancel()
	audit.mu.Lock()
	audit.wg.Wait()

	if audit.queue.count != 0 {
		flush(audit) // Manual flush
	}

	audit.file.Close()
	audit.mu.Unlock()
}

func (audit *Audit) Trace(msg string) {
	audit.log("\033[95m🔎TRAC\033[m", msg)
}

func (audit *Audit) Debug(msg string) {
	audit.log("\033[95m🐛DEBU\033[m", msg)
}

func (audit *Audit) Info(msg string) {
	audit.log("\033[92m👋INFO\033[m", msg)
}

func (audit *Audit) Warn(msg string) {
	audit.log("\033[33m⚠WARN\033[m", msg)
}

func (audit *Audit) Error(msg string) {
	/* Send some sort of alert here as well eventually */
	audit.log("\033[31m❌ERRO\033[m", msg)
}

func (audit *Audit) Fatal(msg string) {
	audit.log("\033[35m☠FATA\033[m", msg)
	os.Exit(22)
}

func (audit *Audit) log(step, msg string) {
	structured_msg := fmt.Sprintf("\033[1m%s %s \033[1m%s", time.Now().UTC().Format(audit.format), step, msg)

	audit.queue.Append(structured_msg)

	fmt.Printf(structured_msg)
}
