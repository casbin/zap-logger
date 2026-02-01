// Copyright 2020 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zaplogger

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/casbin/casbin/v3/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testNewLogger(t *testing.T) *Logger {
	logger := NewLogger(false, true)

	if logger == nil {
		t.Error("initialize logger failed")
	}

	return logger
}

func testNewLoggerByZap(t *testing.T) *Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "event",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	// Use a buffer instead of stdout for testing
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zap.NewAtomicLevel(),
	)
	logger := NewLoggerByZap(zap.New(core), false)

	if logger == nil {
		t.Error("initialize logger failed")
	}

	return logger
}

func TestLoggerCreation(t *testing.T) {
	loggerByDefault := testNewLogger(t)
	loggerByZap := testNewLoggerByZap(t)

	if loggerByDefault.IsEnabled() || loggerByZap.IsEnabled() {
		t.Error("IsEnabled should be false by default")
	}

	loggerByDefault.EnableLog(true)
	loggerByZap.EnableLog(true)

	if !loggerByDefault.IsEnabled() || !loggerByZap.IsEnabled() {
		t.Error("EnableLog(true) should enable logging")
	}

	loggerByDefault.EnableLog(false)
	loggerByZap.EnableLog(false)

	if loggerByDefault.IsEnabled() || loggerByZap.IsEnabled() {
		t.Error("EnableLog(false) should disable logging")
	}
}

func TestSetEventTypes(t *testing.T) {
	logger := NewLogger(true, true)

	// Set specific event types
	eventTypes := []log.EventType{log.EventEnforce, log.EventAddPolicy}
	err := logger.SetEventTypes(eventTypes)
	if err != nil {
		t.Fatalf("SetEventTypes failed: %v", err)
	}

	// Test that enforce event is active
	enforceEntry := &log.LogEntry{
		EventType: log.EventEnforce,
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Allowed:   true,
	}

	err = logger.OnBeforeEvent(enforceEntry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	if !enforceEntry.IsActive {
		t.Error("Enforce event should be active")
	}

	// Test that loadPolicy event is not active
	loadEntry := &log.LogEntry{
		EventType: log.EventLoadPolicy,
		RuleCount: 5,
	}

	err = logger.OnBeforeEvent(loadEntry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	if loadEntry.IsActive {
		t.Error("LoadPolicy event should not be active")
	}
}

func TestOnBeforeAndAfterEvent(t *testing.T) {
	logger := NewLogger(true, true)

	entry := &log.LogEntry{
		EventType: log.EventEnforce,
		Subject:   "bob",
		Object:    "data2",
		Action:    "write",
		Domain:    "tenant1",
		Allowed:   true,
	}

	// Test OnBeforeEvent
	err := logger.OnBeforeEvent(entry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	if entry.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}

	if !entry.IsActive {
		t.Error("Entry should be active when no event types are set")
	}

	// Small delay to ensure duration > 0
	time.Sleep(1 * time.Millisecond)

	// Test OnAfterEvent
	err = logger.OnAfterEvent(entry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	if entry.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}

	if entry.Duration <= 0 {
		t.Errorf("Duration should be > 0, got %v", entry.Duration)
	}
}

func TestEventLogging(t *testing.T) {
	// Create a logger that outputs to a buffer
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	zapLogger := zap.New(core)
	logger := NewLoggerByZap(zapLogger, true)

	// Test enforce event
	enforceEntry := &log.LogEntry{
		EventType: log.EventEnforce,
		Subject:   "alice",
		Object:    "data1",
		Action:    "read",
		Domain:    "tenant1",
		Allowed:   true,
	}

	err := logger.OnBeforeEvent(enforceEntry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	err = logger.OnAfterEvent(enforceEntry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "enforce") {
		t.Error("Log output should contain 'enforce'")
	}
	if !strings.Contains(output, "alice") {
		t.Error("Log output should contain subject 'alice'")
	}
	if !strings.Contains(output, "data1") {
		t.Error("Log output should contain object 'data1'")
	}
	if !strings.Contains(output, "read") {
		t.Error("Log output should contain action 'read'")
	}
	if !strings.Contains(output, "tenant1") {
		t.Error("Log output should contain domain 'tenant1'")
	}
	if !strings.Contains(output, "true") {
		t.Error("Log output should contain allowed=true")
	}

	// Clear buffer for next test
	buf.Reset()

	// Test addPolicy event with rules
	addPolicyEntry := &log.LogEntry{
		EventType: log.EventAddPolicy,
		RuleCount: 2,
		Rules: [][]string{
			{"alice", "data1", "read"},
			{"bob", "data2", "write"},
		},
	}

	err = logger.OnBeforeEvent(addPolicyEntry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	err = logger.OnAfterEvent(addPolicyEntry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "addPolicy") {
		t.Error("Log output should contain 'addPolicy'")
	}
	if !strings.Contains(output, "rule_count") {
		t.Error("Log output should contain 'rule_count'")
	}
	if !strings.Contains(output, "2") {
		t.Error("Log output should contain rule_count=2")
	}
}

func TestErrorLogging(t *testing.T) {
	// Create a logger that outputs to a buffer
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zap.NewAtomicLevelAt(zap.ErrorLevel), // Only capture error level
	)
	zapLogger := zap.New(core)
	logger := NewLoggerByZap(zapLogger, true)

	// Test enforce event with error
	enforceEntry := &log.LogEntry{
		EventType: log.EventEnforce,
		Subject:   "charlie",
		Object:    "data3",
		Action:    "read",
		Allowed:   false,
		Error:     fmt.Errorf("permission denied"),
	}

	err := logger.OnBeforeEvent(enforceEntry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	err = logger.OnAfterEvent(enforceEntry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "error") {
		t.Error("Error log should be at error level")
	}
	if !strings.Contains(output, "permission denied") {
		t.Error("Log output should contain error message")
	}
}

func TestLogCallback(t *testing.T) {
	logger := NewLogger(true, true)

	var callbackCalled bool
	var callbackEntry *log.LogEntry

	err := logger.SetLogCallback(func(entry *log.LogEntry) error {
		callbackCalled = true
		callbackEntry = entry
		return nil
	})
	if err != nil {
		t.Fatalf("SetLogCallback failed: %v", err)
	}

	entry := &log.LogEntry{
		EventType: log.EventSavePolicy,
		RuleCount: 10,
	}

	err = logger.OnBeforeEvent(entry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	err = logger.OnAfterEvent(entry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	if !callbackCalled {
		t.Error("Log callback should have been called")
	}

	if callbackEntry != entry {
		t.Error("Callback should receive the same entry")
	}
}

func TestDisabledLogging(t *testing.T) {
	// Create a logger that outputs to a buffer
	var buf bytes.Buffer
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&buf),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	zapLogger := zap.New(core)
	logger := NewLoggerByZap(zapLogger, false) // Disabled by default

	// Enable and test
	logger.EnableLog(true)
	entry := &log.LogEntry{
		EventType: log.EventEnforce,
		Subject:   "test",
		Object:    "test",
		Action:    "test",
		Allowed:   true,
	}

	err := logger.OnBeforeEvent(entry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	err = logger.OnAfterEvent(entry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Log output should not be empty when enabled")
	}

	// Clear buffer and disable logging
	buf.Reset()
	logger.EnableLog(false)

	err = logger.OnBeforeEvent(entry)
	if err != nil {
		t.Fatalf("OnBeforeEvent failed: %v", err)
	}

	time.Sleep(1 * time.Millisecond)

	err = logger.OnAfterEvent(entry)
	if err != nil {
		t.Fatalf("OnAfterEvent failed: %v", err)
	}

	output = buf.String()
	if output != "" {
		t.Error("Log output should be empty when disabled")
	}
}