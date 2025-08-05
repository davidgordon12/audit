package audit

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

		if audit.queue.count >= int(float64(audit.queue.capacity+1)/1.5) {
			flush(audit)
		}

		audit.mtx.Lock()
		info, err := audit.file.Stat()
		audit.mtx.Unlock()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't access file: %s", err.Error())
			return
		}

		if info.Size() > int64(audit.config.FileSize) {
			rotateLogFile(audit)
		}
	}
}

// Flush the queue holding all our logger messages
func flush(audit *Audit) {
	audit.mtx.Lock()
	defer audit.mtx.Unlock()

	// Not sure if this should be re-compiled each time we call flush, or on Audit initialization.
	// But for my use-case, flush isn't called often. I can take the performance hit for now
	ansiRegex := regexp.MustCompile("\x1b\\[[0-?]*[ -/]*[@-~]")

	// Pop does not actually remove any elements from the slice
	// So it is safe to call it in the loop
	for range audit.queue.count {
		log_msg, err := audit.queue.Pop()
		log_msg_raw := ansiRegex.ReplaceAllString(log_msg, "")
		if err != nil {
			return
		}

		if len(log_msg_raw) > 0 && log_msg_raw[len(log_msg_raw)-1] != '\n' {
			_, err := audit.writer.WriteString(log_msg_raw + "\n")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
		} else {
			_, err := audit.writer.WriteString(log_msg_raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
			}
		}
	}

	if err := audit.writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}

}

func rotateLogFile(audit *Audit) error {
	audit.mtx.Lock()
	defer audit.mtx.Unlock()

	if err := audit.writer.Flush(); err != nil {
		return err
	}
	if err := audit.file.Close(); err != nil {
		return err
	}

	f, err := openFile(audit.config)
	if err != nil {
		return fmt.Errorf("failed to rename old log file: %w", err)
	}

	audit.file = f
	audit.writer = bufio.NewWriter(f)
	return nil
}
