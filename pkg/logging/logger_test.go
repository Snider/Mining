package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output: &buf,
		Level:  LevelInfo,
	})

	// Debug should not appear at Info level
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Debug message should not appear at Info level")
	}

	// Info should appear
	logger.Info("info message")
	if !strings.Contains(buf.String(), "[INFO]") {
		t.Error("Info message should appear")
	}
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info message content should appear")
	}
	buf.Reset()

	// Warn should appear
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "[WARN]") {
		t.Error("Warn message should appear")
	}
	buf.Reset()

	// Error should appear
	logger.Error("error message")
	if !strings.Contains(buf.String(), "[ERROR]") {
		t.Error("Error message should appear")
	}
}

func TestLoggerDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output: &buf,
		Level:  LevelDebug,
	})

	logger.Debug("debug message")
	if !strings.Contains(buf.String(), "[DEBUG]") {
		t.Error("Debug message should appear at Debug level")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output: &buf,
		Level:  LevelInfo,
	})

	logger.Info("test message", Fields{"key": "value", "num": 42})
	output := buf.String()

	if !strings.Contains(output, "key=value") {
		t.Error("Field key=value should appear")
	}
	if !strings.Contains(output, "num=42") {
		t.Error("Field num=42 should appear")
	}
}

func TestLoggerWithComponent(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output:    &buf,
		Level:     LevelInfo,
		Component: "TestComponent",
	})

	logger.Info("test message")
	output := buf.String()

	if !strings.Contains(output, "[TestComponent]") {
		t.Error("Component name should appear in log")
	}
}

func TestLoggerDerivedComponent(t *testing.T) {
	var buf bytes.Buffer
	parent := New(Config{
		Output: &buf,
		Level:  LevelInfo,
	})

	child := parent.WithComponent("ChildComponent")
	child.Info("child message")
	output := buf.String()

	if !strings.Contains(output, "[ChildComponent]") {
		t.Error("Derived component name should appear")
	}
}

func TestLoggerFormatted(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output: &buf,
		Level:  LevelInfo,
	})

	logger.Infof("formatted %s %d", "string", 123)
	output := buf.String()

	if !strings.Contains(output, "formatted string 123") {
		t.Errorf("Formatted message should appear, got: %s", output)
	}
}

func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output: &buf,
		Level:  LevelError,
	})

	// Info should not appear at Error level
	logger.Info("should not appear")
	if buf.Len() > 0 {
		t.Error("Info should not appear at Error level")
	}

	// Change to Info level
	logger.SetLevel(LevelInfo)
	logger.Info("should appear now")
	if !strings.Contains(buf.String(), "should appear now") {
		t.Error("Info should appear after level change")
	}

	// Verify GetLevel
	if logger.GetLevel() != LevelInfo {
		t.Error("GetLevel should return LevelInfo")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
		wantErr  bool
	}{
		{"DEBUG", LevelDebug, false},
		{"debug", LevelDebug, false},
		{"INFO", LevelInfo, false},
		{"info", LevelInfo, false},
		{"WARN", LevelWarn, false},
		{"WARNING", LevelWarn, false},
		{"ERROR", LevelError, false},
		{"error", LevelError, false},
		{"invalid", LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLevel(tt.input)
			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && level != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, level)
			}
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Output: &buf,
		Level:  LevelInfo,
	})

	SetGlobal(logger)

	Info("global test")
	if !strings.Contains(buf.String(), "global test") {
		t.Error("Global logger should write message")
	}

	buf.Reset()
	SetGlobalLevel(LevelError)
	Info("should not appear")
	if buf.Len() > 0 {
		t.Error("Info should not appear at Error level")
	}

	// Reset to default for other tests
	SetGlobal(New(DefaultConfig()))
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

func TestMergeFields(t *testing.T) {
	// Empty fields
	result := mergeFields(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}

	result = mergeFields([]Fields{})
	if result != nil {
		t.Error("empty input should return nil")
	}

	// Single fields
	result = mergeFields([]Fields{{"key": "value"}})
	if result["key"] != "value" {
		t.Error("Single field should be preserved")
	}

	// Multiple fields
	result = mergeFields([]Fields{
		{"key1": "value1"},
		{"key2": "value2"},
	})
	if result["key1"] != "value1" || result["key2"] != "value2" {
		t.Error("Multiple fields should be merged")
	}

	// Override
	result = mergeFields([]Fields{
		{"key": "value1"},
		{"key": "value2"},
	})
	if result["key"] != "value2" {
		t.Error("Later fields should override earlier ones")
	}
}
