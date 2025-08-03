package audit

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
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
		if !item.IsDir() && len(item.Name()) >= 4 && item.Name()[0:4] == "logs" {
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

func TestQueue(t *testing.T) {
	// Setup
	q := NewQueue(3)

	// Act
	q.Append("a")
	q.Append("b")
	q.Append("c")

	q.Append("d") // Should overwrite "a", "b" becomes new "first" element

	val, _ := q.Pop()
	if val != "b" {
		t.Errorf("Expected 'b', got %s", val)
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

func TestFlush(t *testing.T) {
	// Setup
	config := AuditConfig{
		FilePath:  "../resources/logs/log.txt",
		BatchSize: 10,
	}

	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}

	// Act
	LogMessages := 10
	for i := 0; i < LogMessages; i++ {
		audit.Info("Info %d", i)
		audit.Warn("Warn %d", i)
		audit.Error("Error %d", i)
	}
	audit.Close()

	// Assert
	file, err := os.Open(audit.config.FilePath)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		t.Log(err)
		t.FailNow()
	}

	if lineCount != LogMessages*3 {
		t.Errorf("Expected %d lines to be written to file, got %d", LogMessages*3, lineCount)
	}

	// Cleanup
	cleanup()
}

func TestRotate(t *testing.T) {
	// Setup
	config := AuditConfig{
		FileSize:  1024, // 1KB approx. 15 messages
		QueueSize: 3,
	}

	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}

	// Act
	LogMessages := 10
	for i := 0; i < LogMessages; i++ {
		audit.Debug("Debug %d", i)
		audit.Trace("Trace %d", i)
		audit.Info("Info %d", i)
		audit.Warn("Warn %d", i)
		audit.Error("Error %d", i)

		time.Sleep(2 * time.Second) // Wait for flush interval
	}

	// Assert
	expectedFileCount := 2
	actualFileCount := 0

	files, _ := os.ReadDir(".")
	for _, file := range files {
		if file.Name()[0:4] == "logs" {
			actualFileCount++
		}
	}

	if expectedFileCount != actualFileCount {
		t.Errorf("Expected %d files to be created, got %d", expectedFileCount, actualFileCount)
	}

	audit.Close()
	cleanup()
}

func TestAuditWithArgs(t *testing.T) {
	// Setup
	config := AuditConfig{}

	audit, err := NewAudit(config)
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create audit: %w", err)
	}

	// Act
	audit.Info("My name is %s", "audit")
	audit.Close()

	f, _ := os.Open("logs")
	b := make([]byte, 44)
	f.Read(b)
	f.Close()

	// Assert
	expectedStr := "INFO My name is audit"
	if !strings.Contains(string(b), expectedStr) {
		t.Errorf("Expected log message %s was not in log file", expectedStr)
	}

	// Cleanup
	cleanup()
}
