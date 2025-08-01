package audit

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

// Cleanup created log files
func cleanup() {
	fmt.Print("\n\nStarting cleanup process\n")
	items, err := os.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	var filesToDelete []string
	for _, item := range items {
		if !item.IsDir() && len(item.Name()) > 4 && item.Name()[0:4] == "logs" {
			filesToDelete = append(filesToDelete, item.Name())
		}
	}

	for _, fileName := range filesToDelete {
		fmt.Printf("Attempting to delete: \033[1m%s\n\033[0m", fileName)
		if err := os.Remove(fileName); err != nil {
			fmt.Printf("Error deleting %s: %s\n", fileName, err)
		} else {
			fmt.Printf("Successfully deleted: \033[1m%s\n\033[0m", fileName)
		}
	}
}

func TestLogDebugOnce(t *testing.T) {
	// Setup
	config := AuditConfig{Level: DEBUG}
	audit, err := NewAudit(config)
	if err != nil {
		t.Fatalf("Failed to create audit: %v", err)
	}
	defer audit.Close()

	// Act
	audit.Debug("Debug!")

	// Assert
	if audit.queue.count == 0 {
		t.Error("Expected at least one log message in the queue")
	}

	// Cleanup
	audit.Close()
	cleanup()
}

func TestLogTraceOnce(t *testing.T) {
	// Setup
	config := AuditConfig{Level: TRACE}
	audit, err := NewAudit(config)
	if err != nil {
		t.Fatalf("Failed to create audit: %v", err)
	}
	defer audit.Close()

	// Act
	audit.Trace("Trace!")

	// Assert
	if audit.queue.count == 0 {
		t.Error("Expected at least one log message in the queue")
	}

	// Cleanup
	audit.Close()
	cleanup()
}

func TestLogInfoOnce(t *testing.T) {
	// Setup
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}

	audit.writer = bufio.NewWriter(os.Stdout) // Hijack the writer to only write to stdout

	// Act
	audit.Info("Info!")

	// Assert
	if audit.queue.count < 1 {
		t.Error("Expected at least one log message in the queue")
	}

	// Cleanup
	audit.Close()
	cleanup()
}

func TestLogWarningOnce(t *testing.T) {
	// Setup
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}

	audit.writer = bufio.NewWriter(os.Stdout) // Hijack the writer to only write to stdout

	// Act
	audit.Warn("Warning!")

	// Assert
	if audit.queue.count < 1 {
		t.Error("Expected at least one log message in the queue")
	}

	// Cleanup
	audit.Close()
	cleanup()
}

func TestLogErrorOnce(t *testing.T) {
	// Setup
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}

	audit.writer = bufio.NewWriter(os.Stdout) // Hijack the writer to only write to stdout

	// Act
	audit.Error("Warning!")

	// Assert
	if audit.queue.count < 1 {
		t.Error("Expected at least one log message in the queue")
	}

	// Cleanup
	audit.Close()
	cleanup()
}
func TestLogLevel(t *testing.T) {
	// Setup
	config := AuditConfig{Level: INFO}
	audit, err := NewAudit(config)
	if err != nil {
		t.Fatalf("Failed to create audit: %v", err)
	}

	// Act
	audit.Trace("Trace")
	audit.Debug("Debug")
	audit.Info("Info")
	audit.Warn("Warn")
	audit.Error("Error")

	// Assert
	if audit.queue.count != 3 {
		t.Errorf("Expected 3 log messages, got %d", audit.queue.count) // Trace and Debug should not be logged
	}

	// Cleanup
	audit.Close()
	cleanup()
}

func TestLogMultipleLevels(t *testing.T) {
	// Setup
	config := AuditConfig{Level: TRACE}
	audit, err := NewAudit(config)
	if err != nil {
		t.Fatalf("Failed to create audit: %v", err)
	}

	// Act
	audit.Trace("Trace")
	audit.Debug("Debug")
	audit.Info("Info")
	audit.Warn("Warn")
	audit.Error("Error")

	// Assert
	if audit.queue.count < 5 {
		t.Errorf("Expected at least 5 log messages, got %d", audit.queue.count)
	}

	// Cleanup
	audit.Close()
	cleanup()
}

func TestLogSimulated(t *testing.T) {
	// Setup
	config := AuditConfig{}
	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}
	defer audit.Close()

	// Act
	LogMessages := 10000
	for i := 0; i < LogMessages; i++ {
		audit.Info(fmt.Sprintf("Info %d", i))
		audit.Warn(fmt.Sprintf("Warn %d", i))
		audit.Error(fmt.Sprintf("Error %d", i))
	}

	// Assert
	expectedFileCount := LogMessages * 3 / audit.config.FileSize
	actualFileCount := 0

	files, _ := os.ReadDir(".")
	for _, file := range files {
		if file.Name()[0:4] == "logs" {
			actualFileCount++
		}
	}

	if expectedFileCount != actualFileCount {
		t.Error("Expected %d files to be created, got %d", expectedFileCount, actualFileCount)
	}
}
