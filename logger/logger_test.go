package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	// Setup: Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	
	logPath := filepath.Join(tempDir, "logs", "test.log")
	
	// Test initialization
	err = Init(logPath)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}
	
	// Verify logger was created
	if Logger == nil {
		t.Error("Logger is nil after initialization")
	}
	
	// Verify log file was created
	if LogFile == nil {
		t.Error("LogFile is nil after initialization")
	}
	
	// Verify the file exists on disk
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logPath)
	}
	
	// Test writing to the logger
	testMessage := "This is a test log message"
	Logger.Println(testMessage)
	
	// Close the logger to release the file
	Close()
	
	// Verify the log message was written to the file
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Could not open log file for verification: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	var found bool
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), testMessage) {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Log message was not written to the file")
	}
	
	// Clean up
	os.RemoveAll(tempDir)
}

func TestInitInvalidPath(t *testing.T) {
	// Test with an invalid path (where we don't have permissions)
	invalidPath := "/root/invalid/path/test.log" // This should fail on most systems without root access
	
	err := Init(invalidPath)
	if err == nil {
		// This is not expected to succeed on most systems without root privileges
		t.Error("Expected error with invalid path, but got none")
		Close() // Clean up if it unexpectedly succeeded
	}
}

func TestClose(t *testing.T) {
	// Setup
	tempDir := filepath.Join(os.TempDir(), "logger_close_test")
	os.MkdirAll(tempDir, 0755)
	logPath := filepath.Join(tempDir, "close_test.log")
	
	// Initialize logger
	err := Init(logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	
	// Get the file info before closing
	_, err = LogFile.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}
	
	// Close the logger
	Close()
	
	// Try to use the logger after closing - this shouldn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic occurred after Close: %v", r)
			}
		}()
		
		// This should be safe and not cause issues after Close
		if LogFile != nil {
			// This would fail if the file is closed, which is expected
			_, err := LogFile.Write([]byte("test"))
			if err == nil {
				t.Error("Expected error writing to closed file, but got none")
			}
		}
	}()
	
	// Clean up
	os.RemoveAll(tempDir)
}

func TestInitAppend(t *testing.T) {
	// Setup
	tempDir := filepath.Join(os.TempDir(), "logger_append_test")
	os.MkdirAll(tempDir, 0755)
	logPath := filepath.Join(tempDir, "append_test.log")
	
	// Create initial content
	initialContent := "Initial log content\n"
	err := os.WriteFile(logPath, []byte(initialContent), 0666)
	if err != nil {
		t.Fatalf("Failed to write initial content: %v", err)
	}
	
	// Initialize logger
	err = Init(logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	
	// Write new content
	newContent := "New log content"
	Logger.Println(newContent)
	
	// Close the logger
	Close()
	
	// Verify both contents exist in the file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	contentStr := string(content)
	if !strings.Contains(contentStr, initialContent) {
		t.Error("Initial content was not preserved in the log file")
	}
	
	if !strings.Contains(contentStr, newContent) {
		t.Error("New content was not appended to the log file")
	}
	
	// Clean up
	os.RemoveAll(tempDir)
}

func TestLoggerOutput(t *testing.T) {
	// Setup
	tempDir := filepath.Join(os.TempDir(), "logger_output_test")
	os.MkdirAll(tempDir, 0755)
	logPath := filepath.Join(tempDir, "output_test.log")
	
	// Initialize logger
	err := Init(logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	
	// Log different types of messages
	Logger.Print("Print message")
	Logger.Printf("Printf %s", "message")
	Logger.Println("Println message")
	
	// Close the logger
	Close()
	
	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	contentStr := string(content)
	
	// Verify all messages were written
	if !strings.Contains(contentStr, "Print message") {
		t.Error("Print message not found in log file")
	}
	
	if !strings.Contains(contentStr, "Printf message") {
		t.Error("Printf message not found in log file")
	}
	
	if !strings.Contains(contentStr, "Println message") {
		t.Error("Println message not found in log file")
	}
	
	// Verify log format (contains date, time, and file info)
	// This is a basic check - you might want more specific verification
	scanner := bufio.NewScanner(strings.NewReader(contentStr))
	if scanner.Scan() {
		line := scanner.Text()
		// Check format: date and time should be present
		hasDate := strings.Contains(line, "/")
		hasTime := strings.Contains(line, ":")
		hasFile := strings.Contains(line, ".go:")
		
		if !hasDate || !hasTime || !hasFile {
			t.Errorf("Log format doesn't match expected: %s", line)
		}
	}
	
	// Clean up
	os.RemoveAll(tempDir)
}