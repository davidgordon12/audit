package audit

import (
	"time"
)

// Starts the background worker to write messages from the queue to a file.
func startLogWriterService(audit *Audit) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()


	for {
		select {
		case <-audit.queue.count > audit.config.BatchSize:
			flush()
		case <-ticker.C:
			flush()
		case <-audit.ctx.Done():
			flush()
			return
		}

		// TODO: Check if the file needs to be rotated
		// Lock the mutex, check file size, unlock then call rotate (locks again)
	}
	flush(audit)
}

// Flush the queue holding all our logger messages
func flush(audit *Audit) {
	audit.mu.Lock()
	defer audit.mu.Unlock()

	// Pop does not actually remove any elements from the slice
	// So it is safe to call it in the loop
	for i := 0; i < audit.queue.count; i++ {
		log_msg, err := audit.queue.Pop()
		if err != nil {
			return;
		}

		if len(msg) > 0 && msg[len(msg)-1] != '\n' {
			_, err := audit.writer.WriteString(log_msg + '\n')
			fmt.Fprintf(os.Stderr, err.Error())
		} else {
			_, err := audit.writer.WriteString(log_msg)
			fmt.Fprintf(os.Stderr, err.Error())
		}
	}

	if err := l.writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}

}

func (audit *Audit) rotateLogFile() error {
	audit.mu.Lock()
	defer audit.mu.Unlock()

	if err := audit.writer.Flush(); err != nil {
		return err
	}
	if err := audit.file.Close(); err != nil {
		return err
	}

	// Switch to a new file
	now = time.Now().Format("20060102_150405")
	f, err := os.OpenFile(cfg.FilePath + now, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", f, err)
	}

	audit.file = f
	audit.writer = bufio.NewWriter(f)
	return nil
}
