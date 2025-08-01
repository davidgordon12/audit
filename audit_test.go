package audit

import (
	"fmt"
	"os"
	"testing"
)

func TestLogInfoOnce(t *testing.T) {
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}
	defer audit.Close()

	audit.Info("Hello world!")
}

func TestLogWarningOnce(t *testing.T) {
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}
	defer audit.Close()

	audit.Info("Hello world!")
}

func TestLogFatalOnce(t *testing.T) {
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}
	defer audit.Close()

	audit.Info("Hello world!")
}

func TestLogInfoSimulated(t *testing.T) {

}
