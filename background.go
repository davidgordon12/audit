package audit

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

// Starts the background worker to write messages from the queue to a file.
func startLogWriterService(audit *Audit) {
	audit.wg.Add(1)
	defer audit.wg.Done()

	ticker := time.NewTicker(audit.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			flush(audit)
		case <-audit.ctx.Done():
			flush(audit)
			return
		}

		// TODO: Check if the file needs to be rotated
		// Lock the mutex, check file size, unlock then call rotate (locks again)
	}
}

// Flush the queue holding all our logger messages
func flush(audit *Audit) {
	audit.mtx.Lock()
	defer audit.mtx.Unlock()

	// Pop does not actually remove any elements from the slice
	// So it is safe to call it in the loop
	for range audit.queue.count {
		log_msg, err := audit.queue.Pop()
		if err != nil {
			return
		}

		if len(log_msg) > 0 && log_msg[len(log_msg)-1] != '\n' {
			_, err := audit.writer.WriteString(log_msg + "\n")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
		} else {
			_, err := audit.writer.WriteString(log_msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
		}
	}

	if err := audit.writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

}

func (audit *Audit) rotateLogFile() error {
	audit.mtx.Lock()
	defer audit.mtx.Unlock()

	if err := audit.writer.Flush(); err != nil {
		return err
	}
	if err := audit.file.Close(); err != nil {
		return err
	}

	// Switch to a new file
	now := time.Now().Format("20060102_150405")
	f, err := os.OpenFile(audit.config.FilePath+now, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", f.Name(), err)
	}

	audit.file = f
	audit.writer = bufio.NewWriter(f)
	return nil
}
