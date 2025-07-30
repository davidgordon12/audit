package audit

import (
	"time"
)

func StartLogWriterService(audit *Audit) {
	for {
		time.Sleep(10 * time.Second)
		performWrite(audit)
	}
}

func performWrite(audit *Audit) {
	// Take the queue and read from head -> tail
	// Pop each element from the queue when done
}
